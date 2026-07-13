use dragonfly_plugin_sys::{
    DF_ABI_VERSION, DF_COMMAND_PARAMETER_BOOL, DF_COMMAND_PARAMETER_DYNAMIC_ENUM,
    DF_COMMAND_PARAMETER_ENUM, DF_COMMAND_PARAMETER_FLOAT, DF_COMMAND_PARAMETER_INTEGER,
    DF_COMMAND_PARAMETER_PLAYER, DF_COMMAND_PARAMETER_RAW_TEXT, DF_COMMAND_PARAMETER_STRING,
    DF_COMMAND_PARAMETER_SUBCOMMAND, DF_EVENT_PLAYER_ATTACK_ENTITY, DF_EVENT_PLAYER_BLOCK_BREAK,
    DF_EVENT_PLAYER_BLOCK_PICK, DF_EVENT_PLAYER_BLOCK_PLACE, DF_EVENT_PLAYER_CHANGE_WORLD,
    DF_EVENT_PLAYER_CHAT, DF_EVENT_PLAYER_DEATH, DF_EVENT_PLAYER_EXPERIENCE_GAIN,
    DF_EVENT_PLAYER_FIRE_EXTINGUISH, DF_EVENT_PLAYER_FOOD_LOSS, DF_EVENT_PLAYER_HEAL,
    DF_EVENT_PLAYER_HELD_SLOT_CHANGE, DF_EVENT_PLAYER_HURT, DF_EVENT_PLAYER_ITEM_CONSUME,
    DF_EVENT_PLAYER_ITEM_DAMAGE, DF_EVENT_PLAYER_ITEM_DROP, DF_EVENT_PLAYER_ITEM_RELEASE,
    DF_EVENT_PLAYER_ITEM_USE, DF_EVENT_PLAYER_ITEM_USE_ON_BLOCK,
    DF_EVENT_PLAYER_ITEM_USE_ON_ENTITY, DF_EVENT_PLAYER_JOIN, DF_EVENT_PLAYER_JUMP,
    DF_EVENT_PLAYER_LECTERN_PAGE_TURN, DF_EVENT_PLAYER_MOVE, DF_EVENT_PLAYER_PUNCH_AIR,
    DF_EVENT_PLAYER_QUIT, DF_EVENT_PLAYER_RESPAWN, DF_EVENT_PLAYER_SIGN_EDIT,
    DF_EVENT_PLAYER_SKIN_CHANGE, DF_EVENT_PLAYER_SLEEP, DF_EVENT_PLAYER_START_BREAK,
    DF_EVENT_PLAYER_TELEPORT, DF_EVENT_PLAYER_TOGGLE_SNEAK, DF_EVENT_PLAYER_TOGGLE_SPRINT,
    DF_HOST_ABI_VERSION, DF_STATUS_ERROR, DF_STATUS_OK, DF_SUBSCRIPTION_PLAYER_ATTACK_ENTITY,
    DF_SUBSCRIPTION_PLAYER_BLOCK_BREAK, DF_SUBSCRIPTION_PLAYER_BLOCK_PICK,
    DF_SUBSCRIPTION_PLAYER_BLOCK_PLACE, DF_SUBSCRIPTION_PLAYER_CHANGE_WORLD,
    DF_SUBSCRIPTION_PLAYER_CHAT, DF_SUBSCRIPTION_PLAYER_DEATH,
    DF_SUBSCRIPTION_PLAYER_EXPERIENCE_GAIN, DF_SUBSCRIPTION_PLAYER_FIRE_EXTINGUISH,
    DF_SUBSCRIPTION_PLAYER_FOOD_LOSS, DF_SUBSCRIPTION_PLAYER_HEAL,
    DF_SUBSCRIPTION_PLAYER_HELD_SLOT_CHANGE, DF_SUBSCRIPTION_PLAYER_HURT,
    DF_SUBSCRIPTION_PLAYER_ITEM_CONSUME, DF_SUBSCRIPTION_PLAYER_ITEM_DAMAGE,
    DF_SUBSCRIPTION_PLAYER_ITEM_DROP, DF_SUBSCRIPTION_PLAYER_ITEM_RELEASE,
    DF_SUBSCRIPTION_PLAYER_ITEM_USE, DF_SUBSCRIPTION_PLAYER_ITEM_USE_ON_BLOCK,
    DF_SUBSCRIPTION_PLAYER_ITEM_USE_ON_ENTITY, DF_SUBSCRIPTION_PLAYER_JOIN,
    DF_SUBSCRIPTION_PLAYER_JUMP, DF_SUBSCRIPTION_PLAYER_LECTERN_PAGE_TURN,
    DF_SUBSCRIPTION_PLAYER_MOVE, DF_SUBSCRIPTION_PLAYER_PUNCH_AIR, DF_SUBSCRIPTION_PLAYER_QUIT,
    DF_SUBSCRIPTION_PLAYER_RESPAWN, DF_SUBSCRIPTION_PLAYER_SIGN_EDIT,
    DF_SUBSCRIPTION_PLAYER_SKIN_CHANGE, DF_SUBSCRIPTION_PLAYER_SLEEP,
    DF_SUBSCRIPTION_PLAYER_START_BREAK, DF_SUBSCRIPTION_PLAYER_TELEPORT,
    DF_SUBSCRIPTION_PLAYER_TOGGLE_SNEAK, DF_SUBSCRIPTION_PLAYER_TOGGLE_SPRINT, DfCommandDescriptor,
    DfCommandInput, DfCommandState, DfEntityTypeDescriptorV1, DfHostApiV14, DfItemStackSnapshot,
    DfPlayerAttackEntityInput, DfPlayerAttackEntityState, DfPlayerBlockBreakInput,
    DfPlayerBlockBreakState, DfPlayerBlockPickInput, DfPlayerBlockPickState,
    DfPlayerBlockPlaceInput, DfPlayerBlockPlaceState, DfPlayerChangeWorldInput,
    DfPlayerChangeWorldState, DfPlayerChatInput, DfPlayerChatState, DfPlayerDeathInput,
    DfPlayerDeathState, DfPlayerExperienceGainInput, DfPlayerExperienceGainState,
    DfPlayerFireExtinguishInput, DfPlayerFireExtinguishState, DfPlayerFoodLossInput,
    DfPlayerFoodLossState, DfPlayerHealInput, DfPlayerHealState, DfPlayerHeldSlotChangeInput,
    DfPlayerHeldSlotChangeState, DfPlayerHurtInput, DfPlayerHurtState, DfPlayerItemConsumeInput,
    DfPlayerItemConsumeState, DfPlayerItemDamageInput, DfPlayerItemDamageState,
    DfPlayerItemDropInput, DfPlayerItemDropState, DfPlayerItemReleaseInput,
    DfPlayerItemReleaseState, DfPlayerItemUseInput, DfPlayerItemUseOnBlockInput,
    DfPlayerItemUseOnBlockState, DfPlayerItemUseOnEntityInput, DfPlayerItemUseOnEntityState,
    DfPlayerItemUseState, DfPlayerJoinInput, DfPlayerJoinState, DfPlayerJumpInput,
    DfPlayerJumpState, DfPlayerLecternPageTurnInput, DfPlayerLecternPageTurnState,
    DfPlayerMoveInput, DfPlayerMoveState, DfPlayerPunchAirInput, DfPlayerPunchAirState,
    DfPlayerQuitInput, DfPlayerQuitState, DfPlayerRespawnInput, DfPlayerRespawnState,
    DfPlayerSignEditInput, DfPlayerSignEditState, DfPlayerSkinChangeInput, DfPlayerSkinChangeState,
    DfPlayerSleepInput, DfPlayerSleepState, DfPlayerStartBreakInput, DfPlayerStartBreakState,
    DfPlayerTeleportInput, DfPlayerTeleportState, DfPlayerToggleSneakInput,
    DfPlayerToggleSneakState, DfPlayerToggleSprintInput, DfPlayerToggleSprintState, DfPluginApiV2,
    DfPluginEntryV2Fn, DfStatus, DfStringView,
};
use libloading::{Library, Symbol};
use std::ffi::{OsStr, c_void};
use std::fs;
use std::mem::size_of;
use std::path::{Path, PathBuf};
use std::ptr;
use std::slice;

const MAX_ABI_SLICE_ITEMS: u64 = 1024;
const MAX_ENTITY_TYPES: u64 = MAX_ABI_SLICE_ITEMS;

#[repr(C)]
pub struct DfRuntimeConfig {
    pub plugin_directory: DfStringView,
    pub host: *const DfHostApiV14,
}

pub struct DfRuntime {
    plugins: Vec<LoadedPlugin>,
    entity_types: Vec<DfEntityTypeDescriptorV1>,
    commands: Vec<RuntimeCommand>,
    subscriptions: u64,
}

struct RuntimeCommand {
    plugin: usize,
    local: u64,
    descriptor: DfCommandDescriptor,
}

struct LoadedPlugin {
    api: &'static DfPluginApiV2,
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
    fn load(plugin_directory: &Path, host: *const DfHostApiV14) -> Result<Self, String> {
        let mut paths = native_libraries(plugin_directory)?;
        paths.sort();

        let mut plugins = Vec::with_capacity(paths.len());
        let mut entity_types: Vec<DfEntityTypeDescriptorV1> = Vec::new();
        let mut subscriptions = 0;
        for path in paths {
            // SAFETY: symbols and returned API are validated before use. Library stays owned by LoadedPlugin.
            let plugin = unsafe { LoadedPlugin::open(&path, host)? };
            if plugins
                .iter()
                .any(|loaded: &LoadedPlugin| loaded.id == plugin.id)
            {
                return Err(format!("duplicate plugin ID {:?}", plugin.id));
            }
            if let Some(read) = plugin.api.entity_types {
                let mut count = 0;
                let descriptors = unsafe { read(plugin.instance, &mut count) };
                if count > MAX_ENTITY_TYPES || (count != 0 && descriptors.is_null()) {
                    return Err(format!(
                        "plugin {:?} returned invalid entity types",
                        plugin.id
                    ));
                }
                let descriptors = unsafe { abi_slice(descriptors, count) }.map_err(|()| {
                    format!("plugin {:?} returned too many entity types", plugin.id)
                })?;
                for descriptor in descriptors.iter().copied() {
                    valid_entity_type_descriptor(&descriptor)?;
                    let save_id = unsafe { string_view(descriptor.save_id) }?;
                    if entity_types.iter().any(|existing| {
                        unsafe { string_view(existing.save_id) }.is_ok_and(|id| id == save_id)
                    }) {
                        return Err(format!("duplicate custom entity type {save_id:?}"));
                    }
                    entity_types.push(descriptor);
                }
            }
            subscriptions |= plugin.api.header.subscriptions;
            plugins.push(plugin);
        }
        entity_types.sort_by(|left, right| {
            // SAFETY: every descriptor was validated above and plugin libraries remain loaded.
            let left = unsafe { string_view(left.save_id) }.unwrap_or_default();
            // SAFETY: every descriptor was validated above and plugin libraries remain loaded.
            let right = unsafe { string_view(right.save_id) }.unwrap_or_default();
            left.cmp(right)
        });
        Ok(Self {
            plugins,
            entity_types,
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

    fn handle_join(&self, input: &DfPlayerJoinInput, state: &mut DfPlayerJoinState) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_JOIN == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_JOIN,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_quit(&self, input: &DfPlayerQuitInput, state: &mut DfPlayerQuitState) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_QUIT == 0
            {
                continue;
            }
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_QUIT,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_hurt(&self, input: &DfPlayerHurtInput, state: &mut DfPlayerHurtState) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_HURT == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_HURT,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK || !state.damage.is_finite() {
                return DF_STATUS_ERROR;
            }
            state.damage = state.damage.max(0.0);
        }
        DF_STATUS_OK
    }

    fn handle_heal(&self, input: &DfPlayerHealInput, state: &mut DfPlayerHealState) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_HEAL == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_HEAL,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK || !state.health.is_finite() {
                return DF_STATUS_ERROR;
            }
            state.health = state.health.max(0.0);
        }
        DF_STATUS_OK
    }

    fn handle_block_break(
        &self,
        input: &DfPlayerBlockBreakInput,
        state: &mut DfPlayerBlockBreakState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_BLOCK_BREAK == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_BLOCK_BREAK,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
            state.experience = state.experience.max(0);
        }
        DF_STATUS_OK
    }

    fn handle_block_place(
        &self,
        input: &DfPlayerBlockPlaceInput,
        state: &mut DfPlayerBlockPlaceState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_BLOCK_PLACE == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_BLOCK_PLACE,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_food_loss(
        &self,
        input: &DfPlayerFoodLossInput,
        state: &mut DfPlayerFoodLossState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_FOOD_LOSS == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_FOOD_LOSS,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
            state.to = state.to.clamp(0, 20);
        }
        DF_STATUS_OK
    }

    fn handle_death(&self, input: &DfPlayerDeathInput, state: &mut DfPlayerDeathState) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_DEATH == 0
            {
                continue;
            }
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_DEATH,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_start_break(
        &self,
        input: &DfPlayerStartBreakInput,
        state: &mut DfPlayerStartBreakState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_START_BREAK == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_START_BREAK,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_fire_extinguish(
        &self,
        input: &DfPlayerFireExtinguishInput,
        state: &mut DfPlayerFireExtinguishState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_FIRE_EXTINGUISH == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_FIRE_EXTINGUISH,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_toggle_sprint(
        &self,
        input: &DfPlayerToggleSprintInput,
        state: &mut DfPlayerToggleSprintState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_TOGGLE_SPRINT == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_TOGGLE_SPRINT,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_toggle_sneak(
        &self,
        input: &DfPlayerToggleSneakInput,
        state: &mut DfPlayerToggleSneakState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_TOGGLE_SNEAK == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_TOGGLE_SNEAK,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_jump(&self, input: &DfPlayerJumpInput, state: &mut DfPlayerJumpState) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_JUMP == 0
            {
                continue;
            }
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_JUMP,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_teleport(
        &self,
        input: &DfPlayerTeleportInput,
        state: &mut DfPlayerTeleportState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_TELEPORT == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_TELEPORT,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_experience_gain(
        &self,
        input: &DfPlayerExperienceGainInput,
        state: &mut DfPlayerExperienceGainState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_EXPERIENCE_GAIN == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_EXPERIENCE_GAIN,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            state.amount = state.amount.max(0);
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_punch_air(
        &self,
        input: &DfPlayerPunchAirInput,
        state: &mut DfPlayerPunchAirState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_PUNCH_AIR == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_PUNCH_AIR,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_held_slot_change(
        &self,
        input: &DfPlayerHeldSlotChangeInput,
        state: &mut DfPlayerHeldSlotChangeState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_HELD_SLOT_CHANGE == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_HELD_SLOT_CHANGE,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_sleep(&self, input: &DfPlayerSleepInput, state: &mut DfPlayerSleepState) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_SLEEP == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_SLEEP,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_block_pick(
        &self,
        input: &DfPlayerBlockPickInput,
        state: &mut DfPlayerBlockPickState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_BLOCK_PICK == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_BLOCK_PICK,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_lectern_page_turn(
        &self,
        input: &DfPlayerLecternPageTurnInput,
        state: &mut DfPlayerLecternPageTurnState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_LECTERN_PAGE_TURN == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_LECTERN_PAGE_TURN,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            state.new_page = state.new_page.max(0);
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_sign_edit(
        &self,
        input: &DfPlayerSignEditInput,
        state: &mut DfPlayerSignEditState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_SIGN_EDIT == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_SIGN_EDIT,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_item_use(
        &self,
        input: &DfPlayerItemUseInput,
        state: &mut DfPlayerItemUseState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_ITEM_USE == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_ITEM_USE,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_item_use_on_block(
        &self,
        input: &DfPlayerItemUseOnBlockInput,
        state: &mut DfPlayerItemUseOnBlockState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_ITEM_USE_ON_BLOCK == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_ITEM_USE_ON_BLOCK,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_item_consume(
        &self,
        input: &DfPlayerItemConsumeInput,
        state: &mut DfPlayerItemConsumeState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_ITEM_CONSUME == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_ITEM_CONSUME,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_item_release(
        &self,
        input: &DfPlayerItemReleaseInput,
        state: &mut DfPlayerItemReleaseState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_ITEM_RELEASE == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_ITEM_RELEASE,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_item_damage(
        &self,
        input: &DfPlayerItemDamageInput,
        state: &mut DfPlayerItemDamageState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_ITEM_DAMAGE == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_ITEM_DAMAGE,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            state.damage = state.damage.max(0);
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_item_drop(
        &self,
        input: &DfPlayerItemDropInput,
        state: &mut DfPlayerItemDropState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_ITEM_DROP == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_ITEM_DROP,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_attack_entity(
        &self,
        input: &DfPlayerAttackEntityInput,
        state: &mut DfPlayerAttackEntityState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_ATTACK_ENTITY == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let previous_force = state.knockback_force;
            let previous_height = state.knockback_height;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_ATTACK_ENTITY,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if !state.knockback_force.is_finite() {
                state.knockback_force = previous_force;
                return DF_STATUS_ERROR;
            }
            if !state.knockback_height.is_finite() {
                state.knockback_height = previous_height;
                return DF_STATUS_ERROR;
            }
            state.critical = u8::from(state.critical != 0);
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_item_use_on_entity(
        &self,
        input: &DfPlayerItemUseOnEntityInput,
        state: &mut DfPlayerItemUseOnEntityState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_ITEM_USE_ON_ENTITY == 0
            {
                continue;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_ITEM_USE_ON_ENTITY,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_change_world(
        &self,
        input: &DfPlayerChangeWorldInput,
        state: &mut DfPlayerChangeWorldState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_CHANGE_WORLD == 0
            {
                continue;
            }
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_CHANGE_WORLD,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if status != DF_STATUS_OK {
                return status;
            }
        }
        DF_STATUS_OK
    }

    fn handle_respawn(
        &self,
        input: &DfPlayerRespawnInput,
        state: &mut DfPlayerRespawnState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_RESPAWN == 0
            {
                continue;
            }
            if !valid_respawn(input, state) {
                return DF_STATUS_ERROR;
            }
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_RESPAWN,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if status != DF_STATUS_OK || !valid_respawn(input, state) {
                return DF_STATUS_ERROR;
            }
        }
        DF_STATUS_OK
    }

    fn handle_skin_change(
        &self,
        input: &DfPlayerSkinChangeInput,
        state: &mut DfPlayerSkinChangeState,
    ) -> DfStatus {
        for plugin in &self.plugins {
            if !plugin.enabled
                || plugin.api.header.subscriptions & DF_SUBSCRIPTION_PLAYER_SKIN_CHANGE == 0
            {
                continue;
            }
            if !valid_skin_change(input, state) {
                return DF_STATUS_ERROR;
            }
            let was_cancelled = state.cancelled != 0;
            let Some(handle) = plugin.api.handle_event else {
                return DF_STATUS_ERROR;
            };
            let status = unsafe {
                handle(
                    plugin.instance,
                    DF_EVENT_PLAYER_SKIN_CHANGE,
                    ptr::from_ref(input).cast(),
                    ptr::from_mut(state).cast(),
                )
            };
            if was_cancelled {
                state.cancelled = 1;
            }
            if status != DF_STATUS_OK || !valid_skin_change(input, state) {
                return DF_STATUS_ERROR;
            }
        }
        DF_STATUS_OK
    }
}

fn valid_entity_type_descriptor(descriptor: &DfEntityTypeDescriptorV1) -> Result<(), String> {
    let save_id = unsafe { string_view(descriptor.save_id) }?;
    let network_id = unsafe { string_view(descriptor.network_id) }?;
    if save_id.is_empty() || save_id.len() > 256 || network_id.is_empty() || network_id.len() > 256
    {
        return Err("invalid custom entity identifier".to_owned());
    }
    let min = [descriptor.min.x, descriptor.min.y, descriptor.min.z];
    let max = [descriptor.max.x, descriptor.max.y, descriptor.max.z];
    if min.into_iter().chain(max).any(|value| !value.is_finite())
        || min.into_iter().zip(max).any(|(min, max)| min > max)
    {
        return Err(format!("invalid bounds for custom entity {save_id:?}"));
    }
    Ok(())
}

impl Drop for DfRuntime {
    fn drop(&mut self) {
        self.disable();
    }
}

fn valid_change_world_input(input: &DfPlayerChangeWorldInput) -> bool {
    input.player.generation != 0
        && input.after.value != 0
        && (input.before.value == 0 || input.before.value != input.after.value)
}

fn valid_respawn(input: &DfPlayerRespawnInput, state: &DfPlayerRespawnState) -> bool {
    input.player.generation != 0
        && state.position.x.is_finite()
        && state.position.y.is_finite()
        && state.position.z.is_finite()
        && state.world.value != 0
}

fn valid_skin_change(input: &DfPlayerSkinChangeInput, state: &DfPlayerSkinChangeState) -> bool {
    input.player.generation != 0 && input.snapshot != 0 && state.cancelled <= 1
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

    unsafe fn open(path: &Path, host: *const DfHostApiV14) -> Result<Self, String> {
        // SAFETY: loading native plugins is the purpose of this trusted plugin runtime.
        let library = unsafe { Library::new(path) }
            .map_err(|err| format!("load {}: {err}", path.display()))?;
        // SAFETY: symbol name and function signature are fixed by ABI v2.
        let entry: Symbol<DfPluginEntryV2Fn> = unsafe { library.get(b"df_plugin_entry_v2\0") }
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
        if api.header.struct_size < size_of::<DfPluginApiV2>() as u32 {
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
        let plugin = Self {
            api,
            instance,
            id,
            enabled: false,
            _library: library,
        };
        let Some(set_host) = api.set_host else {
            return Err(format!(
                "plugin {:?} does not accept the host API",
                plugin.id
            ));
        };
        if unsafe { set_host(plugin.instance, host) } != DF_STATUS_OK {
            return Err(format!("plugin {:?} rejected the host API", plugin.id));
        }
        Ok(plugin)
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

unsafe fn valid_damage_source(source: &dragonfly_plugin_sys::DfDamageSourceView) -> bool {
    use dragonfly_plugin_sys::*;

    let known_flags = DF_DAMAGE_SOURCE_REDUCED_BY_ARMOUR
        | DF_DAMAGE_SOURCE_REDUCED_BY_RESISTANCE
        | DF_DAMAGE_SOURCE_FIRE
        | DF_DAMAGE_SOURCE_IGNORES_TOTEM
        | DF_DAMAGE_SOURCE_FIRE_PROTECTION
        | DF_DAMAGE_SOURCE_FEATHER_FALLING
        | DF_DAMAGE_SOURCE_BLAST_PROTECTION
        | DF_DAMAGE_SOURCE_PROJECTILE_PROTECTION;
    let Ok(name) = (unsafe { string_view(source.name) }) else {
        return false;
    };
    if source.kind > DF_DAMAGE_SOURCE_WITHER
        || source.flags & !known_flags != 0
        || source.data > 1
        || source.data != 0 && source.kind != DF_DAMAGE_SOURCE_POISON
        || name.len() > 64 << 10
    {
        return false;
    }
    if source.kind != DF_DAMAGE_SOURCE_BLOCK {
        return source.block.is_null();
    }
    let Some(block) = (unsafe { source.block.as_ref() }) else {
        return false;
    };
    let Ok(identifier) = (unsafe { string_view(block.identifier) }) else {
        return false;
    };
    identifier.len() <= 256
        && !identifier.is_empty()
        && block.properties_nbt.len <= 64 << 10
        && (block.properties_nbt.len == 0 || !block.properties_nbt.data.is_null())
}

unsafe fn valid_healing_source(source: &dragonfly_plugin_sys::DfHealingSourceView) -> bool {
    let Ok(name) = (unsafe { string_view(source.name) }) else {
        return false;
    };
    source.kind <= dragonfly_plugin_sys::DF_HEALING_SOURCE_REGENERATION
        && source.data <= 1
        && (source.data == 0 || source.kind == dragonfly_plugin_sys::DF_HEALING_SOURCE_FOOD)
        && name.len() <= 64 << 10
}

fn valid_item_snapshot(item: &DfItemStackSnapshot) -> bool {
    const MAX_ITEM_BYTES: u64 = 16 << 20;
    item.snapshot != 0
        && item.info.identifier_len <= 256
        && item.info.custom_name_len <= 4096
        && item.info.lore_count <= 256
        && item.info.enchantment_count <= 256
        && [
            item.info.identifier_len,
            item.info.custom_name_len,
            item.info.lore_bytes_len,
            item.info.nbt_len,
            item.info.values_nbt_len,
        ]
        .into_iter()
        .try_fold(0u64, u64::checked_add)
        .is_some_and(|total| total <= MAX_ITEM_BYTES)
}

unsafe fn abi_slice<'a, T>(data: *const T, len: u64) -> Result<&'a [T], ()> {
    if len == 0 {
        return Ok(&[]);
    }
    if data.is_null() || len > MAX_ABI_SLICE_ITEMS {
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
        let host = unsafe { (*config).host };
        let Some(host_api) = (unsafe { host.as_ref() }) else {
            return Err("null host API".to_owned());
        };
        if host_api.abi_version != DF_HOST_ABI_VERSION
            || host_api.struct_size < size_of::<DfHostApiV14>() as u32
        {
            return Err("incompatible host API".to_owned());
        }
        DfRuntime::load(Path::new(directory), host)
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
/// Returns the number of custom entity types declared by loaded plugins.
///
/// # Safety
/// `runtime` must be null or point to a live runtime for this call.
pub unsafe extern "C" fn df_runtime_entity_type_count(runtime: *const DfRuntime) -> u64 {
    unsafe { runtime.as_ref() }.map_or(0, |runtime| runtime.entity_types.len() as u64)
}

#[unsafe(no_mangle)]
/// Copies a custom entity type descriptor by global runtime index.
///
/// # Safety
/// `runtime` and `out` must point to live ABI-compatible values for this call.
pub unsafe extern "C" fn df_runtime_entity_type_at(
    runtime: *const DfRuntime,
    index: u64,
    out: *mut DfEntityTypeDescriptorV1,
) -> DfStatus {
    let (Some(runtime), Some(out)) = (unsafe { runtime.as_ref() }, unsafe { out.as_mut() }) else {
        return DF_STATUS_ERROR;
    };
    let Some(descriptor) = runtime.entity_types.get(index as usize) else {
        return DF_STATUS_ERROR;
    };
    *out = *descriptor;
    DF_STATUS_OK
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
    if input.online_player_count != 0 && input.online_players.is_null() {
        return DF_STATUS_ERROR;
    }
    let players = if input.online_player_count == 0 {
        &[][..]
    } else {
        unsafe { slice::from_raw_parts(input.online_players, input.online_player_count as usize) }
    };
    if players
        .iter()
        .any(|player| unsafe { string_view(player.name) }.is_err())
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
/// Dispatches an event using its generated ABI event ID.
///
/// # Safety
/// `input` and `state` must point to the generated types matching `event_id` and remain valid for this call.
pub unsafe extern "C" fn df_runtime_handle_event(
    runtime: *mut DfRuntime,
    event_id: u32,
    input: *const c_void,
    state: *mut c_void,
) -> DfStatus {
    if event_id == DF_EVENT_PLAYER_MOVE {
        let (Some(runtime), Some(input), Some(state)) = (
            unsafe { runtime.as_ref() },
            unsafe { input.cast::<DfPlayerMoveInput>().as_ref() },
            unsafe { state.cast::<DfPlayerMoveState>().as_mut() },
        ) else {
            return DF_STATUS_ERROR;
        };
        return std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
            runtime.handle_move(input, state)
        }))
        .unwrap_or(DF_STATUS_ERROR);
    }
    match event_id {
        DF_EVENT_PLAYER_CHAT => unsafe {
            df_runtime_handle_player_chat(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_JOIN => unsafe {
            df_runtime_handle_player_join(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_QUIT => unsafe {
            df_runtime_handle_player_quit(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_HURT => unsafe {
            df_runtime_handle_player_hurt(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_HEAL => unsafe {
            df_runtime_handle_player_heal(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_BLOCK_BREAK => unsafe {
            df_runtime_handle_player_block_break(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_BLOCK_PLACE => unsafe {
            df_runtime_handle_player_block_place(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_FOOD_LOSS => unsafe {
            df_runtime_handle_player_food_loss(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_DEATH => unsafe {
            df_runtime_handle_player_death(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_START_BREAK => unsafe {
            df_runtime_handle_player_start_break(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_FIRE_EXTINGUISH => unsafe {
            df_runtime_handle_player_fire_extinguish(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_TOGGLE_SPRINT => unsafe {
            df_runtime_handle_player_toggle_sprint(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_TOGGLE_SNEAK => unsafe {
            df_runtime_handle_player_toggle_sneak(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_JUMP => unsafe {
            df_runtime_handle_player_jump(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_TELEPORT => unsafe {
            df_runtime_handle_player_teleport(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_EXPERIENCE_GAIN => unsafe {
            df_runtime_handle_player_experience_gain(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_PUNCH_AIR => unsafe {
            df_runtime_handle_player_punch_air(runtime, input.cast(), state.cast())
        },
        DF_EVENT_PLAYER_HELD_SLOT_CHANGE => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerHeldSlotChangeInput>().as_ref() },
                unsafe { state.cast::<DfPlayerHeldSlotChangeState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_held_slot_change(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR)
        }
        DF_EVENT_PLAYER_SLEEP => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerSleepInput>().as_ref() },
                unsafe { state.cast::<DfPlayerSleepState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_sleep(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR)
        }
        DF_EVENT_PLAYER_BLOCK_PICK => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerBlockPickInput>().as_ref() },
                unsafe { state.cast::<DfPlayerBlockPickState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            if unsafe { string_view(input.block) }.is_err() {
                return DF_STATUS_ERROR;
            }
            std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_block_pick(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR)
        }
        DF_EVENT_PLAYER_LECTERN_PAGE_TURN => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerLecternPageTurnInput>().as_ref() },
                unsafe { state.cast::<DfPlayerLecternPageTurnState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_lectern_page_turn(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR)
        }
        DF_EVENT_PLAYER_SIGN_EDIT => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerSignEditInput>().as_ref() },
                unsafe { state.cast::<DfPlayerSignEditState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            if unsafe { string_view(input.old_text) }.is_err()
                || unsafe { string_view(input.new_text) }.is_err()
            {
                return DF_STATUS_ERROR;
            }
            std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_sign_edit(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR)
        }
        DF_EVENT_PLAYER_ITEM_USE => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerItemUseInput>().as_ref() },
                unsafe { state.cast::<DfPlayerItemUseState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_item_use(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR)
        }
        DF_EVENT_PLAYER_ITEM_USE_ON_BLOCK => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerItemUseOnBlockInput>().as_ref() },
                unsafe { state.cast::<DfPlayerItemUseOnBlockState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            if !(0..=5).contains(&input.face)
                || !input.click_position.x.is_finite()
                || !input.click_position.y.is_finite()
                || !input.click_position.z.is_finite()
            {
                return DF_STATUS_ERROR;
            }
            std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_item_use_on_block(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR)
        }
        DF_EVENT_PLAYER_ITEM_CONSUME => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerItemConsumeInput>().as_ref() },
                unsafe { state.cast::<DfPlayerItemConsumeState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            if !valid_item_snapshot(&input.item) {
                return DF_STATUS_ERROR;
            }
            std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_item_consume(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR)
        }
        DF_EVENT_PLAYER_ITEM_RELEASE => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerItemReleaseInput>().as_ref() },
                unsafe { state.cast::<DfPlayerItemReleaseState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            if !valid_item_snapshot(&input.item) {
                return DF_STATUS_ERROR;
            }
            std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_item_release(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR)
        }
        DF_EVENT_PLAYER_ITEM_DAMAGE => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerItemDamageInput>().as_ref() },
                unsafe { state.cast::<DfPlayerItemDamageState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            if !valid_item_snapshot(&input.item) || state.damage < 0 {
                return DF_STATUS_ERROR;
            }
            std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_item_damage(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR)
        }
        DF_EVENT_PLAYER_ITEM_DROP => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerItemDropInput>().as_ref() },
                unsafe { state.cast::<DfPlayerItemDropState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            if !valid_item_snapshot(&input.item) {
                return DF_STATUS_ERROR;
            }
            std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_item_drop(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR)
        }
        DF_EVENT_PLAYER_ATTACK_ENTITY => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerAttackEntityInput>().as_ref() },
                unsafe { state.cast::<DfPlayerAttackEntityState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            if !state.knockback_force.is_finite() || !state.knockback_height.is_finite() {
                return DF_STATUS_ERROR;
            }
            std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_attack_entity(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR)
        }
        DF_EVENT_PLAYER_ITEM_USE_ON_ENTITY => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerItemUseOnEntityInput>().as_ref() },
                unsafe { state.cast::<DfPlayerItemUseOnEntityState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_item_use_on_entity(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR)
        }
        DF_EVENT_PLAYER_CHANGE_WORLD => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerChangeWorldInput>().as_ref() },
                unsafe { state.cast::<DfPlayerChangeWorldState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            if !valid_change_world_input(input) || state._reserved != 0 {
                return DF_STATUS_ERROR;
            }
            let status = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_change_world(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR);
            if status == DF_STATUS_OK && state._reserved == 0 {
                DF_STATUS_OK
            } else {
                DF_STATUS_ERROR
            }
        }
        DF_EVENT_PLAYER_RESPAWN => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerRespawnInput>().as_ref() },
                unsafe { state.cast::<DfPlayerRespawnState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            if !valid_respawn(input, state) {
                return DF_STATUS_ERROR;
            }
            let status = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_respawn(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR);
            if status == DF_STATUS_OK && valid_respawn(input, state) {
                DF_STATUS_OK
            } else {
                DF_STATUS_ERROR
            }
        }
        DF_EVENT_PLAYER_SKIN_CHANGE => {
            let (Some(runtime), Some(input), Some(state)) = (
                unsafe { runtime.as_ref() },
                unsafe { input.cast::<DfPlayerSkinChangeInput>().as_ref() },
                unsafe { state.cast::<DfPlayerSkinChangeState>().as_mut() },
            ) else {
                return DF_STATUS_ERROR;
            };
            if !valid_skin_change(input, state) {
                return DF_STATUS_ERROR;
            }
            let status = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                runtime.handle_skin_change(input, state)
            }))
            .unwrap_or(DF_STATUS_ERROR);
            if status == DF_STATUS_OK && valid_skin_change(input, state) {
                DF_STATUS_OK
            } else {
                DF_STATUS_ERROR
            }
        }
        _ => DF_STATUS_ERROR,
    }
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

#[unsafe(no_mangle)]
/// Dispatches a player join event.
///
/// # Safety
/// All pointers and string views must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_join(
    runtime: *mut DfRuntime,
    input: *const DfPlayerJoinInput,
    state: *mut DfPlayerJoinState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    if unsafe { string_view(input.name) }.is_err() {
        return DF_STATUS_ERROR;
    }
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_join(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player quit event.
///
/// # Safety
/// All pointers and string views must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_quit(
    runtime: *mut DfRuntime,
    input: *const DfPlayerQuitInput,
    state: *mut DfPlayerQuitState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    if unsafe { string_view(input.name) }.is_err() {
        return DF_STATUS_ERROR;
    }
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_quit(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player hurt event.
///
/// # Safety
/// All pointers and string views must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_hurt(
    runtime: *mut DfRuntime,
    input: *const DfPlayerHurtInput,
    state: *mut DfPlayerHurtState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    if !unsafe { valid_damage_source(&input.source) } || !state.damage.is_finite() {
        return DF_STATUS_ERROR;
    }
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_hurt(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player heal event.
///
/// # Safety
/// All pointers and string views must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_heal(
    runtime: *mut DfRuntime,
    input: *const DfPlayerHealInput,
    state: *mut DfPlayerHealState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    if !unsafe { valid_healing_source(&input.source) } || !state.health.is_finite() {
        return DF_STATUS_ERROR;
    }
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_heal(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player block-break event.
///
/// # Safety
/// All pointers and string views must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_block_break(
    runtime: *mut DfRuntime,
    input: *const DfPlayerBlockBreakInput,
    state: *mut DfPlayerBlockBreakState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    if unsafe { string_view(input.block) }.is_err() {
        return DF_STATUS_ERROR;
    }
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_block_break(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player block-place event.
///
/// # Safety
/// All pointers and string views must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_block_place(
    runtime: *mut DfRuntime,
    input: *const DfPlayerBlockPlaceInput,
    state: *mut DfPlayerBlockPlaceState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    if unsafe { string_view(input.block) }.is_err() {
        return DF_STATUS_ERROR;
    }
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_block_place(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player food-loss event.
///
/// # Safety
/// All pointers must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_food_loss(
    runtime: *mut DfRuntime,
    input: *const DfPlayerFoodLossInput,
    state: *mut DfPlayerFoodLossState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_food_loss(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player death event.
///
/// # Safety
/// All pointers and string views must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_death(
    runtime: *mut DfRuntime,
    input: *const DfPlayerDeathInput,
    state: *mut DfPlayerDeathState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    if !unsafe { valid_damage_source(&input.source) } {
        return DF_STATUS_ERROR;
    }
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_death(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player start-break event.
///
/// # Safety
/// All pointers must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_start_break(
    runtime: *mut DfRuntime,
    input: *const DfPlayerStartBreakInput,
    state: *mut DfPlayerStartBreakState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_start_break(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player fire-extinguish event.
///
/// # Safety
/// All pointers must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_fire_extinguish(
    runtime: *mut DfRuntime,
    input: *const DfPlayerFireExtinguishInput,
    state: *mut DfPlayerFireExtinguishState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_fire_extinguish(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player sprint-toggle event.
///
/// # Safety
/// All pointers must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_toggle_sprint(
    runtime: *mut DfRuntime,
    input: *const DfPlayerToggleSprintInput,
    state: *mut DfPlayerToggleSprintState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_toggle_sprint(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player sneak-toggle event.
///
/// # Safety
/// All pointers must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_toggle_sneak(
    runtime: *mut DfRuntime,
    input: *const DfPlayerToggleSneakInput,
    state: *mut DfPlayerToggleSneakState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_toggle_sneak(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player jump event.
///
/// # Safety
/// All pointers must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_jump(
    runtime: *mut DfRuntime,
    input: *const DfPlayerJumpInput,
    state: *mut DfPlayerJumpState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_jump(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player teleport event.
///
/// # Safety
/// All pointers must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_teleport(
    runtime: *mut DfRuntime,
    input: *const DfPlayerTeleportInput,
    state: *mut DfPlayerTeleportState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_teleport(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player experience-gain event.
///
/// # Safety
/// All pointers must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_experience_gain(
    runtime: *mut DfRuntime,
    input: *const DfPlayerExperienceGainInput,
    state: *mut DfPlayerExperienceGainState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_experience_gain(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[unsafe(no_mangle)]
/// Dispatches a player punch-air event.
///
/// # Safety
/// All pointers must remain valid for this synchronous call.
pub unsafe extern "C" fn df_runtime_handle_player_punch_air(
    runtime: *mut DfRuntime,
    input: *const DfPlayerPunchAirInput,
    state: *mut DfPlayerPunchAirState,
) -> DfStatus {
    let (Some(runtime), Some(input), Some(state)) = (
        unsafe { runtime.as_ref() },
        unsafe { input.as_ref() },
        unsafe { state.as_mut() },
    ) else {
        return DF_STATUS_ERROR;
    };
    std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        runtime.handle_punch_air(input, state)
    }))
    .unwrap_or(DF_STATUS_ERROR)
}

#[cfg(test)]
mod tests {
    use super::*;
    use dragonfly_plugin_sys::*;

    #[test]
    fn empty_directory_loads() {
        let directory =
            std::env::temp_dir().join(format!("dragonfly-runtime-{}", std::process::id()));
        let _ = fs::remove_dir_all(&directory);
        fs::create_dir_all(&directory).unwrap();
        let host = DfHostApiV14 {
            abi_version: DF_HOST_ABI_VERSION,
            struct_size: size_of::<DfHostApiV14>() as u32,
            context: 0,
            player_text: None,
            player_title: None,
            player_transform: None,
            player_rotation: None,
            player_state_set: None,
            player_state_get: None,
            player_effect: None,
            player_entity_visibility: None,
            player_skin_open: None,
            player_skin_animation_info: None,
            player_skin_read: None,
            player_skin_close: None,
            player_skin_set: None,
            inventory_size: None,
            inventory_item_open: None,
            player_held_item_open: None,
            item_stack_read: None,
            item_stack_close: None,
            inventory_item_set: None,
            inventory_item_add: None,
            inventory_clear_slot: None,
            inventory_clear: None,
            player_held_items_set: None,
            player_held_slot_set: None,
            player_scoreboard: None,
            player_scoreboard_remove: None,
            player_form_send: None,
            player_form_close: None,
            world_lookup: None,
            world_open: None,
            world_name: None,
            world_unload: None,
            world_save: None,
            world_block_get: None,
            world_block_set: None,
            world_time_get: None,
            world_time_set: None,
            world_spawn_get: None,
            world_spawn_set: None,
            world_entity_spawn: None,
            world_entities: None,
            world_players: None,
            entity_state: None,
            entity_teleport: None,
            entity_velocity_set: None,
            entity_name_tag_set: None,
            entity_despawn: None,
            world_particle_add: None,
            world_sound_play: None,
            player_sound_play: None,
            player_heal: None,
            player_hurt: None,
            skin_snapshot_info: None,
            skin_snapshot_set: None,
        };
        let runtime = DfRuntime::load(&directory, ptr::from_ref(&host)).unwrap();
        assert!(runtime.plugins.is_empty());
        fs::remove_dir_all(directory).unwrap();
    }

    #[test]
    fn generic_dispatch_rejects_unknown_events() {
        let status = unsafe {
            df_runtime_handle_event(ptr::null_mut(), u32::MAX, ptr::null(), ptr::null_mut())
        };
        assert_eq!(status, DF_STATUS_ERROR);
    }

    #[test]
    fn generic_dispatch_accepts_item_use_on_entity() {
        let mut runtime = DfRuntime {
            plugins: Vec::new(),
            entity_types: Vec::new(),
            commands: Vec::new(),
            subscriptions: 0,
        };
        let input = DfPlayerItemUseOnEntityInput::default();
        let mut state = DfPlayerItemUseOnEntityState::default();
        let status = unsafe {
            df_runtime_handle_event(
                ptr::from_mut(&mut runtime),
                DF_EVENT_PLAYER_ITEM_USE_ON_ENTITY,
                ptr::from_ref(&input).cast(),
                ptr::from_mut(&mut state).cast(),
            )
        };
        assert_eq!(status, DF_STATUS_OK);

        let null_input = unsafe {
            df_runtime_handle_event(
                ptr::from_mut(&mut runtime),
                DF_EVENT_PLAYER_ITEM_USE_ON_ENTITY,
                ptr::null(),
                ptr::from_mut(&mut state).cast(),
            )
        };
        assert_eq!(null_input, DF_STATUS_ERROR);
    }

    #[test]
    fn generic_dispatch_validates_change_world_handles() {
        let mut runtime = DfRuntime {
            plugins: Vec::new(),
            entity_types: Vec::new(),
            commands: Vec::new(),
            subscriptions: 0,
        };
        let mut input = DfPlayerChangeWorldInput {
            player: dragonfly_plugin_sys::DfPlayerId {
                generation: 1,
                ..Default::default()
            },
            before: dragonfly_plugin_sys::DfWorldId { value: 0 },
            after: dragonfly_plugin_sys::DfWorldId { value: 1 },
            ..Default::default()
        };
        let mut state = DfPlayerChangeWorldState::default();
        let mut dispatch = |input: &DfPlayerChangeWorldInput,
                            state: &mut DfPlayerChangeWorldState| unsafe {
            df_runtime_handle_event(
                ptr::from_mut(&mut runtime),
                DF_EVENT_PLAYER_CHANGE_WORLD,
                ptr::from_ref(input).cast(),
                ptr::from_mut(state).cast(),
            )
        };

        assert_eq!(dispatch(&input, &mut state), DF_STATUS_OK);

        input.before.value = 2;
        assert_eq!(dispatch(&input, &mut state), DF_STATUS_OK);

        input.after.value = 2;
        assert_eq!(dispatch(&input, &mut state), DF_STATUS_ERROR);

        input.after.value = 1;
        input.player.generation = 0;
        assert_eq!(dispatch(&input, &mut state), DF_STATUS_ERROR);
        input.player.generation = 1;

        input.before.value = 0;
        input.after.value = 0;
        assert_eq!(dispatch(&input, &mut state), DF_STATUS_ERROR);

        input.after.value = 1;
        state._reserved = 1;
        assert_eq!(dispatch(&input, &mut state), DF_STATUS_ERROR);
    }

    #[test]
    fn generic_dispatch_validates_respawn_state() {
        let mut runtime = DfRuntime {
            plugins: Vec::new(),
            entity_types: Vec::new(),
            commands: Vec::new(),
            subscriptions: 0,
        };
        let mut input = DfPlayerRespawnInput {
            player: dragonfly_plugin_sys::DfPlayerId {
                generation: 1,
                ..Default::default()
            },
            ..Default::default()
        };
        let mut state = DfPlayerRespawnState {
            position: dragonfly_plugin_sys::DfVec3 {
                x: 1.0,
                y: 64.0,
                z: 2.0,
            },
            world: dragonfly_plugin_sys::DfWorldId { value: 1 },
        };
        let mut dispatch = |input: &DfPlayerRespawnInput, state: &mut DfPlayerRespawnState| unsafe {
            df_runtime_handle_event(
                ptr::from_mut(&mut runtime),
                DF_EVENT_PLAYER_RESPAWN,
                ptr::from_ref(input).cast(),
                ptr::from_mut(state).cast(),
            )
        };

        assert_eq!(dispatch(&input, &mut state), DF_STATUS_OK);

        input.player.generation = 0;
        assert_eq!(dispatch(&input, &mut state), DF_STATUS_ERROR);
        input.player.generation = 1;

        state.position.x = f64::NAN;
        assert_eq!(dispatch(&input, &mut state), DF_STATUS_ERROR);
        state.position.x = 1.0;

        state.position.y = f64::INFINITY;
        assert_eq!(dispatch(&input, &mut state), DF_STATUS_ERROR);
        state.position.y = 64.0;

        state.world.value = 0;
        assert_eq!(dispatch(&input, &mut state), DF_STATUS_ERROR);
    }

    #[test]
    fn generic_dispatch_validates_skin_change_state() {
        let mut runtime = DfRuntime {
            plugins: Vec::new(),
            entity_types: Vec::new(),
            commands: Vec::new(),
            subscriptions: 0,
        };
        let mut input = DfPlayerSkinChangeInput {
            invocation: 7,
            player: DfPlayerId {
                generation: 1,
                ..Default::default()
            },
            snapshot: 9,
        };
        let mut state = DfPlayerSkinChangeState { cancelled: 0 };
        let mut dispatch = |input: &DfPlayerSkinChangeInput,
                            state: &mut DfPlayerSkinChangeState| unsafe {
            df_runtime_handle_event(
                ptr::from_mut(&mut runtime),
                DF_EVENT_PLAYER_SKIN_CHANGE,
                ptr::from_ref(input).cast(),
                ptr::from_mut(state).cast(),
            )
        };

        assert_eq!(dispatch(&input, &mut state), DF_STATUS_OK);
        input.snapshot = 0;
        assert_eq!(dispatch(&input, &mut state), DF_STATUS_ERROR);
        input.snapshot = 9;
        input.player.generation = 0;
        assert_eq!(dispatch(&input, &mut state), DF_STATUS_ERROR);
        input.player.generation = 1;
        state.cancelled = 2;
        assert_eq!(dispatch(&input, &mut state), DF_STATUS_ERROR);
    }

    #[test]
    fn validates_damage_source_payloads() {
        let name = b"attack";
        let mut source = DfDamageSourceView {
            name: DfStringView {
                data: name.as_ptr(),
                len: name.len() as u64,
            },
            kind: DF_DAMAGE_SOURCE_ATTACK,
            flags: DF_DAMAGE_SOURCE_REDUCED_BY_ARMOUR | DF_DAMAGE_SOURCE_REDUCED_BY_RESISTANCE,
            entity: DfEntityId {
                generation: 1,
                ..Default::default()
            },
            ..Default::default()
        };
        assert!(unsafe { valid_damage_source(&source) });

        source.entity.generation = 0;
        assert!(unsafe { valid_damage_source(&source) });
        source.entity.generation = 1;

        source.flags |= 1 << 31;
        assert!(!unsafe { valid_damage_source(&source) });
        source.flags &= !(1 << 31);

        source.data = 1;
        assert!(!unsafe { valid_damage_source(&source) });
        source.data = 0;

        let block_name = b"minecraft:stone";
        let block = DfBlockView {
            identifier: DfStringView {
                data: block_name.as_ptr(),
                len: block_name.len() as u64,
            },
            properties_nbt: DfStringView::default(),
        };
        source.kind = DF_DAMAGE_SOURCE_BLOCK;
        source.entity = DfEntityId::default();
        source.block = ptr::from_ref(&block);
        assert!(unsafe { valid_damage_source(&source) });

        source.kind = DF_DAMAGE_SOURCE_FALL;
        assert!(!unsafe { valid_damage_source(&source) });
    }

    #[test]
    fn validates_healing_source_payloads() {
        let name = b"food";
        let mut source = DfHealingSourceView {
            name: DfStringView {
                data: name.as_ptr(),
                len: name.len() as u64,
            },
            kind: DF_HEALING_SOURCE_FOOD,
            data: 1,
        };
        assert!(unsafe { valid_healing_source(&source) });

        source.kind = DF_HEALING_SOURCE_INSTANT;
        assert!(!unsafe { valid_healing_source(&source) });
        source.data = 0;
        assert!(unsafe { valid_healing_source(&source) });

        source.kind = DF_HEALING_SOURCE_CUSTOM;
        source.name = DfStringView::default();
        assert!(unsafe { valid_healing_source(&source) });

        let oversized = vec![b'x'; (64 << 10) + 1];
        source.name = DfStringView {
            data: oversized.as_ptr(),
            len: oversized.len() as u64,
        };
        assert!(!unsafe { valid_healing_source(&source) });
    }

    #[test]
    fn validates_entity_type_descriptors() {
        let save_id = b"example:marker";
        let network_id = b"minecraft:armor_stand";
        let mut descriptor = DfEntityTypeDescriptorV1 {
            save_id: DfStringView {
                data: save_id.as_ptr(),
                len: save_id.len() as u64,
            },
            network_id: DfStringView {
                data: network_id.as_ptr(),
                len: network_id.len() as u64,
            },
            min: DfVec3 {
                x: -0.25,
                y: 0.0,
                z: -0.25,
            },
            max: DfVec3 {
                x: 0.25,
                y: 1.975,
                z: 0.25,
            },
        };
        assert!(valid_entity_type_descriptor(&descriptor).is_ok());

        descriptor.max.y = f64::NAN;
        assert!(valid_entity_type_descriptor(&descriptor).is_err());
        descriptor.max.y = 1.975;
        descriptor.min.x = 1.0;
        assert!(valid_entity_type_descriptor(&descriptor).is_err());
        descriptor.min.x = -0.25;

        let invalid_utf8 = [0xff];
        descriptor.save_id = DfStringView {
            data: invalid_utf8.as_ptr(),
            len: invalid_utf8.len() as u64,
        };
        assert!(valid_entity_type_descriptor(&descriptor).is_err());
        descriptor.save_id = DfStringView::default();
        assert!(valid_entity_type_descriptor(&descriptor).is_err());
    }

    #[test]
    fn abi_slice_limit_matches_entity_type_limit() {
        let values = vec![0u8; MAX_ENTITY_TYPES as usize];
        assert_eq!(
            unsafe { abi_slice(values.as_ptr(), MAX_ENTITY_TYPES) }
                .unwrap()
                .len(),
            MAX_ENTITY_TYPES as usize
        );
        assert!(unsafe { abi_slice(values.as_ptr(), MAX_ENTITY_TYPES + 1) }.is_err());
        assert!(unsafe { abi_slice::<u8>(ptr::null(), 1) }.is_err());
        assert!(unsafe { abi_slice::<u8>(ptr::null(), 0) }.is_ok());
    }
}
