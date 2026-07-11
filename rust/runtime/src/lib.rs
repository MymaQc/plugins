use dragonfly_plugin_sys::{
    DF_ABI_VERSION, DF_COMMAND_PARAMETER_BOOL, DF_COMMAND_PARAMETER_DYNAMIC_ENUM,
    DF_COMMAND_PARAMETER_ENUM, DF_COMMAND_PARAMETER_FLOAT, DF_COMMAND_PARAMETER_INTEGER,
    DF_COMMAND_PARAMETER_PLAYER, DF_COMMAND_PARAMETER_RAW_TEXT, DF_COMMAND_PARAMETER_STRING,
    DF_COMMAND_PARAMETER_SUBCOMMAND, DF_EVENT_PLAYER_CHAT, DF_EVENT_PLAYER_MOVE, DF_STATUS_ERROR,
    DF_STATUS_OK, DF_SUBSCRIPTION_PLAYER_CHAT, DF_SUBSCRIPTION_PLAYER_MOVE, DfCommandDescriptor,
    DfCommandInput, DfCommandState, DfPlayerChatInput, DfPlayerChatState, DfPlayerMoveInput,
    DfPlayerMoveState, DfPluginApiV1, DfPluginEntryV1Fn, DfStatus, DfStringView,
};
use libloading::{Library, Symbol};
use std::ffi::{OsStr, c_void};
use std::fs;
use std::mem::size_of;
use std::path::{Path, PathBuf};
use std::ptr;
use std::slice;

#[repr(C)]
pub struct DfRuntimeConfig {
    pub plugin_directory: DfStringView,
}

pub struct DfRuntime {
    plugins: Vec<LoadedPlugin>,
    commands: Vec<RuntimeCommand>,
    subscriptions: u64,
}

struct RuntimeCommand {
    plugin: usize,
    local: u64,
    descriptor: DfCommandDescriptor,
}

struct LoadedPlugin {
    api: &'static DfPluginApiV1,
    instance: *mut c_void,
    id: String,
    enabled: bool,
    _library: Library,
}

// Plugin ABI declares callbacks thread-safe. Runtime never mutates these fields during dispatch.
unsafe impl Send for LoadedPlugin {}
unsafe impl Sync for LoadedPlugin {}

impl Drop for LoadedPlugin {
    fn drop(&mut self) {
        if let Some(destroy) = self.api.destroy {
            // SAFETY: instance came from this API's create callback and library is still loaded.
            unsafe { destroy(self.instance) };
        }
    }
}

impl DfRuntime {
    fn load(plugin_directory: &Path) -> Result<Self, String> {
        let mut paths = native_libraries(plugin_directory)?;
        paths.sort();

        let mut plugins = Vec::with_capacity(paths.len());
        let mut subscriptions = 0;
        for path in paths {
            // SAFETY: symbols and returned API are validated before use. Library stays owned by LoadedPlugin.
            let plugin = unsafe { LoadedPlugin::open(&path)? };
            if plugins
                .iter()
                .any(|loaded: &LoadedPlugin| loaded.id == plugin.id)
            {
                return Err(format!("duplicate plugin ID {:?}", plugin.id));
            }
            subscriptions |= plugin.api.header.subscriptions;
            plugins.push(plugin);
        }
        Ok(Self {
            plugins,
            commands: Vec::new(),
            subscriptions,
        })
    }

    fn handle_move(&self, input: &DfPlayerMoveInput, state: &mut DfPlayerMoveState) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_MOVE == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            // SAFETY: pointers refer to ABI-compatible values for this synchronous callback.
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_MOVE,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            // Cancellation is monotonic even if a plugin writes raw ABI state incorrectly.
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn enable(&mut self) -> DfStatus {
        for index in 0..self.plugins.len() {
            let plugin = &mut self.plugins[index];
            if plugin.enabled {
                continue;
            }
            let status = plugin.api.enable.map_or(DF_STATUS_OK, |enable| {
                // SAFETY: instance belongs to this plugin and callback is ABI-validated.
                unsafe { enable(plugin.instance) }
            });
            if status != DF_STATUS_OK {
                for previous in self.plugins[..index].iter_mut().rev() {
                    previous.disable();
                }
                return status;
            }
            plugin.enabled = true;
        }
        let status = self.rebuild_commands();
        if status != DF_STATUS_OK {
            self.disable();
        }
        status
    }

    fn disable(&mut self) {
        self.commands.clear();
        for plugin in self.plugins.iter_mut().rev() {
            plugin.disable();
        }
    }

    fn rebuild_commands(&mut self) -> DfStatus {
        self.commands.clear();
        for (plugin_index, plugin) in self.plugins.iter().enumerate() {
            let Some(commands) = plugin.api.commands else {
                continue;
            };
            let mut count = 0;
            let descriptors = unsafe { commands(plugin.instance, &mut count) };
            if count != 0 && descriptors.is_null() {
                return DF_STATUS_ERROR;
            }
            let Ok(descriptors) = (unsafe { abi_slice(descriptors, count) }) else {
                return DF_STATUS_ERROR;
            };
            for (local, descriptor) in descriptors.iter().copied().enumerate() {
                let Ok(name) = (unsafe { string_view(descriptor.name) }) else {
                    return DF_STATUS_ERROR;
                };
                if name.is_empty()
                    || self.commands.iter().any(|command| {
                        unsafe { string_view(command.descriptor.name) }
                            .is_ok_and(|existing| existing == name)
                    })
                {
                    return DF_STATUS_ERROR;
                }
                if unsafe { string_view(descriptor.description) }.is_err() {
                    return DF_STATUS_ERROR;
                }
                if !valid_command_descriptor(&descriptor) {
                    return DF_STATUS_ERROR;
                }
                self.commands.push(RuntimeCommand {
                    plugin: plugin_index,
                    local: local as u64,
                    descriptor,
                });
            }
        }
        DF_STATUS_OK
    }

    fn handle_command(
        &self,
        index: usize,
        input: &DfCommandInput,
        state: &mut DfCommandState,
    ) -> DfStatus {
        let Some(command) = self.commands.get(index) else {
            return DF_STATUS_ERROR;
        };
        let plugin = &self.plugins[command.plugin];
        if !plugin.enabled {
            return DF_STATUS_ERROR;
        }
        let Some(handle) = plugin.api.handle_command else {
            return DF_STATUS_ERROR;
        };
        let status = unsafe { handle(plugin.instance, command.local, input, state) };
        if status != DF_STATUS_OK || !valid_command_state(state) {
            return DF_STATUS_ERROR;
        }
        DF_STATUS_OK
    }

    fn command_enum_options(
        &self,
        index: usize,
        overload: u64,
        parameter: u64,
        context: &dragonfly_plugin_sys::DfCommandEnumContext,
        output: &mut dragonfly_plugin_sys::DfStringBuffer,
    ) -> DfStatus {
        let Some(command) = self.commands.get(index) else {
            return DF_STATUS_ERROR;
        };
        let plugin = &self.plugins[command.plugin];
        if !plugin.enabled {
            return DF_STATUS_ERROR;
        }
        let Some(options) = plugin.api.command_enum_options else {
            return DF_STATUS_ERROR;
        };
        let status = unsafe {
            options(
                plugin.instance,
                command.local,
                overload,
                parameter,
                context,
                output,
            )
        };
        if status != DF_STATUS_OK || output.len > output.capacity {
            return DF_STATUS_ERROR;
        }
        if output.len != 0 {
            if output.data.is_null() {
                return DF_STATUS_ERROR;
            }
            let bytes = unsafe { slice::from_raw_parts(output.data, output.len as usize) };
            if std::str::from_utf8(bytes).is_err() {
                return DF_STATUS_ERROR;
            }
        }
        DF_STATUS_OK
    }

    fn handle_chat(&self, input: &DfPlayerChatInput, state: &mut DfPlayerChatState) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_CHAT == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            // SAFETY: pointers refer to ABI-compatible values for this synchronous callback.
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_CHAT,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK || !valid_chat_state(state) {
                return DF_STATUS_ERROR;
            }
        }
        DF_STATUS_OK
    }
}

impl Drop for DfRuntime {
    fn drop(&mut self) {
        self.disable();
    }
}

fn valid_chat_state(state: &DfPlayerChatState) -> bool {
    if state.has_replacement == 0 {
        return true;
    }
    if state.replacement.len > state.replacement.capacity {
        return false;
    }
    if state.replacement.len == 0 {
        return true;
    }
    if state.replacement.data.is_null() {
        return false;
    }
    // SAFETY: bounds and pointer were validated above for the caller-owned buffer.
    let bytes = unsafe {
        slice::from_raw_parts(
            state.replacement.data.cast_const(),
            state.replacement.len as usize,
        )
    };
    std::str::from_utf8(bytes).is_ok()
}

fn valid_command_state(state: &DfCommandState) -> bool {
    if state.output.len > state.output.capacity {
        return false;
    }
    if state.output.len == 0 {
        return true;
    }
    if state.output.data.is_null() {
        return false;
    }
    // SAFETY: bounds and pointer were validated above for the caller-owned buffer.
    let bytes =
        unsafe { slice::from_raw_parts(state.output.data.cast_const(), state.output.len as usize) };
    std::str::from_utf8(bytes).is_ok()
}

impl LoadedPlugin {
    fn disable(&mut self) {
        if !self.enabled {
            return;
        }
        if let Some(disable) = self.api.disable {
            // SAFETY: instance belongs to this plugin and callback is ABI-validated.
            let _ = unsafe { disable(self.instance) };
        }
        self.enabled = false;
    }

    unsafe fn open(path: &Path) -> Result<Self, String> {
        // SAFETY: loading native plugins is the purpose of this trusted plugin runtime.
        let library = unsafe { Library::new(path) }
            .map_err(|err| format!("load {}: {err}", path.display()))?;
        // SAFETY: symbol name and function signature are fixed by ABI v1.
        let entry: Symbol<DfPluginEntryV1Fn> = unsafe { library.get(b"df_plugin_entry_v1\0") }
            .map_err(|err| format!("load entry from {}: {err}", path.display()))?;
        // SAFETY: entry has no arguments and returns a static API descriptor.
        let api_ptr = unsafe { entry() };
        let Some(api) = (unsafe { api_ptr.as_ref() }) else {
            return Err(format!("{} returned a null plugin API", path.display()));
        };
        if api.header.abi_version != DF_ABI_VERSION {
            return Err(format!(
                "{} uses ABI {}, expected {}",
                path.display(),
                api.header.abi_version,
                DF_ABI_VERSION
            ));
        }
        if api.header.struct_size < size_of::<DfPluginApiV1>() as u32 {
            return Err(format!(
                "{} returned a truncated plugin API",
                path.display()
            ));
        }
        let id = unsafe { string_view(api.plugin_id) }?.to_owned();
        if id.is_empty() {
            return Err(format!("{} has an empty plugin ID", path.display()));
        }
        if api.header.subscriptions != 0 && api.handle_event.is_none() {
            return Err(format!("plugin {id:?} subscribes without a handler"));
        }
        let instance = match api.create {
            Some(create) => unsafe { create() },
            None => ptr::null_mut(),
        };
        if api.create.is_some() && instance.is_null() {
            return Err(format!("plugin {id:?} failed to create its instance"));
        }
        Ok(Self {
            api,
            instance,
            id,
            enabled: false,
            _library: library,
        })
    }
}

fn native_libraries(directory: &Path) -> Result<Vec<PathBuf>, String> {
    let entries = fs::read_dir(directory)
        .map_err(|err| format!("read plugin directory {}: {err}", directory.display()))?;
    let extension = if cfg!(target_os = "macos") {
        "dylib"
    } else if cfg!(target_os = "windows") {
        "dll"
    } else {
        "so"
    };
    let mut paths = Vec::new();
    for entry in entries {
        let entry = entry.map_err(|err| format!("read plugin directory entry: {err}"))?;
        let path = entry.path();
        if path.is_file() && path.extension() == Some(OsStr::new(extension)) {
            paths.push(path);
        }
    }
    Ok(paths)
}

unsafe fn string_view<'a>(view: DfStringView) -> Result<&'a str, String> {
    if view.len == 0 {
        return Ok("");
    }
    if view.data.is_null() {
        return Err("non-empty string view has null data".to_owned());
    }
    // SAFETY: caller guarantees view points to readable memory for view.len bytes.
    let bytes = unsafe { slice::from_raw_parts(view.data, view.len as usize) };
    std::str::from_utf8(bytes).map_err(|err| format!("invalid UTF-8: {err}"))
}

unsafe fn abi_slice<'a, T>(data: *const T, len: u64) -> Result<&'a [T], ()> {
    if len == 0 {
        return Ok(&[]);
    }
    if data.is_null() || len > 1024 {
        return Err(());
    }
    Ok(unsafe { slice::from_raw_parts(data, len as usize) })
}

fn valid_command_descriptor(descriptor: &DfCommandDescriptor) -> bool {
    let Ok(overloads) = (unsafe { abi_slice(descriptor.overloads, descriptor.overload_count) })
    else {
        return false;
    };
    for overload in overloads {
        if overload.parameter_count > 4 {
            return false;
        }
        let Ok(parameters) = (unsafe { abi_slice(overload.parameters, overload.parameter_count) })
        else {
            return false;
        };
        for parameter in parameters {
            if parameter.optional > 1 {
                return false;
            }
            let Ok(name) = (unsafe { string_view(parameter.name) }) else {
                return false;
            };
            if name.is_empty() {
                return false;
            }
            let Ok(values) = (unsafe { abi_slice(parameter.values, parameter.value_count) }) else {
                return false;
            };
            match parameter.kind {
                DF_COMMAND_PARAMETER_SUBCOMMAND if values.is_empty() => {}
                DF_COMMAND_PARAMETER_ENUM if !values.is_empty() => {
                    for value in values {
                        if unsafe { string_view(*value) }.is_err() {
                            return false;
                        }
                    }
                }
                DF_COMMAND_PARAMETER_STRING
                | DF_COMMAND_PARAMETER_INTEGER
                | DF_COMMAND_PARAMETER_FLOAT
                | DF_COMMAND_PARAMETER_BOOL
                    if values.is_empty() => {}
                DF_COMMAND_PARAMETER_DYNAMIC_ENUM if values.is_empty() => {}
                DF_COMMAND_PARAMETER_PLAYER if values.is_empty() => {}
                DF_COMMAND_PARAMETER_RAW_TEXT if values.is_empty() => {}
                _ => return false,
            }
        }
    }
    true
}

fn write_error(buffer: *mut u8, capacity: u64, message: &str) {
    if buffer.is_null() || capacity == 0 {
        return;
    }
    let count = message.len().min(capacity.saturating_sub(1) as usize);
    // SAFETY: caller provides writable buffer of capacity bytes.
    unsafe {
        ptr::copy_nonoverlapping(message.as_ptr(), buffer, count);
        *buffer.add(count) = 0;
    }
}

#[unsafe(no_mangle)]
/// Creates a runtime and loads plugins from the configured directory.
///
/// # Safety
/// `config`, `out`, and any provided error buffer must be valid for this call.
pub unsafe extern "C" fn df_runtime_create(
    config: *const DfRuntimeConfig,
    out: *mut *mut DfRuntime,
    error: *mut u8,
    error_capacity: u64,
) -> DfStatus {
    if config.is_null() || out.is_null() {
        write_error(error, error_capacity, "null runtime config or output");
        return DF_STATUS_ERROR;
    }
    // SAFETY: pointers were checked and caller owns both for this call.
    unsafe { *out = ptr::null_mut() };
    let result = std::panic::catch_unwind(|| {
        // SAFETY: config is readable for this call.
        let directory = unsafe { string_view((*config).plugin_directory) }?;
        DfRuntime::load(Path::new(directory))
    });
    match result {
        Ok(Ok(runtime)) => {
            // SAFETY: out is writable and ownership transfers to caller.
            unsafe { *out = Box::into_raw(Box::new(runtime)) };
            DF_STATUS_OK
        }
        Ok(Err(message)) => {
            write_error(error, error_capacity, &message);
            DF_STATUS_ERROR
        }
        Err(_) => {
            write_error(error, error_capacity, "panic while creating runtime");
            DF_STATUS_ERROR
        }
    }
}

#[unsafe(no_mangle)]
/// Enables loaded plugins in deterministic order.
///
/// # Safety
/// `runtime` must point to a live runtime and must not be used concurrently during this call.
pub unsafe extern "C" fn df_runtime_enable(runtime: *mut DfRuntime) -> DfStatus {
    let Some(runtime) = (unsafe { runtime.as_mut() }) else {
        return DF_STATUS_ERROR;
    };
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| runtime.enable()))
        .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Disables enabled plugins in reverse order.
///
/// # Safety
/// `runtime` must be null or point to a live runtime and must not be used concurrently during this call.
pub unsafe extern "C" fn df_runtime_disable(runtime: *mut DfRuntime) {
    if let Some(runtime) = unsafe { runtime.as_mut() } {
        let _ = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| runtime.disable()));
    }
}

#[unsafe(no_mangle)]
/// Destroys a runtime returned by [`df_runtime_create`].
///
/// # Safety
/// `runtime` must be null or a live pointer returned by [`df_runtime_create`], and consumed once.
pub unsafe extern "C" fn df_runtime_destroy(runtime: *mut DfRuntime) {
    if !runtime.is_null() {
        // SAFETY: pointer came from df_runtime_create and is consumed exactly once.
        drop(unsafe { Box::from_raw(runtime) });
    }
}

#[unsafe(no_mangle)]
/// Returns loaded plugin count.
///
/// # Safety
/// `runtime` must be null or point to a live runtime for this call.
pub unsafe extern "C" fn df_runtime_plugin_count(runtime: *const DfRuntime) -> u64 {
    // SAFETY: null is handled; non-null pointer is owned by caller.
    unsafe { runtime.as_ref() }.map_or(0, |runtime| runtime.plugins.len() as u64)
}

#[unsafe(no_mangle)]
/// Returns combined event subscription bits.
///
/// # Safety
/// `runtime` must be null or point to a live runtime for this call.
pub unsafe extern "C" fn df_runtime_subscriptions(runtime: *const DfRuntime) -> u64 {
    // SAFETY: null is handled; non-null pointer is owned by caller.
    unsafe { runtime.as_ref() }.map_or(0, |runtime| runtime.subscriptions)
}

#[unsafe(no_mangle)]
/// Returns the number of commands exposed by enabled plugins.
///
/// # Safety
/// `runtime` must be null or point to a live runtime for this call.
pub unsafe extern "C" fn df_runtime_command_count(runtime: *const DfRuntime) -> u64 {
    // SAFETY: null is handled; non-null pointer is owned by caller.
    unsafe { runtime.as_ref() }.map_or(0, |runtime| runtime.commands.len() as u64)
}

#[unsafe(no_mangle)]
/// Copies a command descriptor by global runtime index.
///
/// # Safety
/// `runtime` and `out` must point to live ABI-compatible values for this call.
pub unsafe extern "C" fn df_runtime_command_at(
    runtime: *const DfRuntime,
    index: u64,
    out: *mut DfCommandDescriptor,
) -> DfStatus {
    let (Some(runtime), Some(out)) = (unsafe { runtime.as_ref() }, unsafe { out.as_mut() }) else {
        return DF_STATUS_ERROR;
    };
    let Some(command) = runtime.commands.get(index as usize) else {
        return DF_STATUS_ERROR;
    };
    *out = command.descriptor;
    DF_STATUS_OK
}

#[unsafe(no_mangle)]
/// Dispatches a registered command to its owning plugin.
///
/// # Safety
/// All pointers must reference live ABI-compatible values for this synchronous call. The output
/// buffer in `state` must remain writable for its declared capacity.
pub unsafe extern "C" fn df_runtime_handle_command(
    runtime: *mut DfRuntime,
    index: u64,
    input: *const DfCommandInput,
    state: *mut DfCommandState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    if unsafe { string_view(input.source) }.is_err()
        || unsafe { string_view(input.arguments) }.is_err()
    {
        return DF_STATUS_ERROR;
    }
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_command(index as usize, input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Resolves source-aware options for a dynamic command enum.
///
/// # Safety
/// All pointers and string views must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_command_enum_options(
    runtime: *mut DfRuntime,
    index: u64,
    overload: u64,
    parameter: u64,
    context: *const dragonfly_plugin_sys::DfCommandEnumContext,
    output: *mut dragonfly_plugin_sys::DfStringBuffer,
) -> DfStatus {
    let (Some(runtime), Some(context), Some(output)) = (
        unsafe { runtime.as_ref() },
        unsafe { context.as_ref() },
        unsafe { output.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    if unsafe { string_view(context.source) }.is_err() {
        return DF_STATUS_ERROR;
    }
    let Ok(players) = (unsafe { abi_slice(context.online_players, context.online_player_count) })
    else {
        return DF_STATUS_ERROR;
    };
    if players
        .iter()
        .any(|player| unsafe { string_view(*player) }.is_err())
    {
        return DF_STATUS_ERROR;
    }
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.command_enum_options(index as usize, overload, parameter, context, output)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player movement event.
///
/// # Safety
/// All pointers must reference live, ABI-compatible values for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_move(
    runtime: *mut DfRuntime,
    input: *const DfPlayerMoveInput,
    state: *mut DfPlayerMoveState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_move(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player chat event.
///
/// # Safety
/// All pointers must reference live, ABI-compatible values for this synchronous call. The replacement
/// buffer in `state` must remain writable for its declared capacity.
pub unsafe extern "C" fn df_runtime_handle_player_chat(
    runtime: *mut DfRuntime,
    input: *const DfPlayerChatInput,
    state: *mut DfPlayerChatState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_chat(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn empty_directory_loads() {
        let directory =
            std::env::temp_dir().join(format!("dragonfly-runtime-{}", std::process::id()));
        let _ = fs::remove_dir_all(&directory);
        fs::create_dir_all(&directory).unwrap();
        let runtime = DfRuntime::load(&directory).unwrap();
        assert!(runtime.plugins.is_empty());
        fs::remove_dir_all(directory).unwrap();
    }
}
