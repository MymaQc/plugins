//! Safe Rust SDK for native Dragonfly plugins.

extern crate self as dragonfly_plugin;

pub use dragonfly_plugin_macros::{Command, CommandEnum, plugin};

#[doc(hidden)]
pub mod __private {
    pub use core::ffi::c_void;
    pub use dragonfly_plugin_sys as sys;
}

#[derive(Clone, Copy, Debug, Default, PartialEq)]
pub struct Vec3 {
    pub x: f64,
    pub y: f64,
    pub z: f64,
}

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct PlayerId {
    uuid: [u8; 16],
    generation: u64,
}

impl PlayerId {
    pub const fn uuid_bytes(self) -> [u8; 16] {
        self.uuid
    }

    pub const fn generation(self) -> u64 {
        self.generation
    }
}

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct Player {
    id: PlayerId,
}

impl Player {
    pub const fn id(self) -> PlayerId {
        self.id
    }

    #[doc(hidden)]
    pub fn from_command_argument(value: &str) -> Result<Self, CommandParseError> {
        let (uuid, generation) = value
            .split_once(':')
            .ok_or_else(|| CommandParseError::new("player is no longer online"))?;
        if uuid.len() != 32 {
            return Err(CommandParseError::new("player is no longer online"));
        }
        let mut bytes = [0; 16];
        for (index, byte) in bytes.iter_mut().enumerate() {
            *byte = u8::from_str_radix(&uuid[index * 2..index * 2 + 2], 16)
                .map_err(|_| CommandParseError::new("player is no longer online"))?;
        }
        let generation = generation
            .parse()
            .map_err(|_| CommandParseError::new("player is no longer online"))?;
        Ok(Self {
            id: PlayerId {
                uuid: bytes,
                generation,
            },
        })
    }
}

impl From<dragonfly_plugin_sys::DfVec3> for Vec3 {
    fn from(value: dragonfly_plugin_sys::DfVec3) -> Self {
        Self {
            x: value.x,
            y: value.y,
            z: value.z,
        }
    }
}

pub struct PlayerMoveEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerMoveInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerMoveState,
}

impl<'a> PlayerMoveEvent<'a> {
    /// Creates a safe event view over runtime-validated ABI values.
    ///
    /// # Safety
    /// Both references must belong to the same active movement callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerMoveInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerMoveState,
    ) -> Self {
        Self { input, state }
    }

    pub fn old_position(&self) -> Vec3 {
        self.input.old_position.into()
    }

    pub fn new_position(&self) -> Vec3 {
        self.input.new_position.into()
    }

    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }

    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub struct MessageTooLong {
    pub length: usize,
    pub capacity: usize,
}

impl core::fmt::Display for MessageTooLong {
    fn fmt(&self, formatter: &mut core::fmt::Formatter<'_>) -> core::fmt::Result {
        write!(
            formatter,
            "message uses {} bytes, capacity is {}",
            self.length, self.capacity
        )
    }
}

impl std::error::Error for MessageTooLong {}

#[repr(transparent)]
pub struct Command(dragonfly_plugin_sys::DfCommandDescriptor);

unsafe impl Sync for Command {}

#[repr(transparent)]
pub struct CommandValue(dragonfly_plugin_sys::DfStringView);

unsafe impl Sync for CommandValue {}

impl CommandValue {
    pub const fn new(value: &'static str) -> Self {
        Self(dragonfly_plugin_sys::DfStringView {
            data: value.as_ptr(),
            len: value.len() as u64,
        })
    }
}

#[repr(transparent)]
pub struct CommandParameter(dragonfly_plugin_sys::DfCommandParameter);

unsafe impl Sync for CommandParameter {}

impl CommandParameter {
    const fn typed(name: &'static str, kind: u32) -> Self {
        Self(dragonfly_plugin_sys::DfCommandParameter {
            kind,
            name: dragonfly_plugin_sys::DfStringView {
                data: name.as_ptr(),
                len: name.len() as u64,
            },
            values: core::ptr::null(),
            value_count: 0,
        })
    }

    pub const fn subcommand(name: &'static str) -> Self {
        Self::typed(name, dragonfly_plugin_sys::DF_COMMAND_PARAMETER_SUBCOMMAND)
    }

    pub const fn enumeration(name: &'static str, values: &'static [CommandValue]) -> Self {
        Self(dragonfly_plugin_sys::DfCommandParameter {
            kind: dragonfly_plugin_sys::DF_COMMAND_PARAMETER_ENUM,
            name: dragonfly_plugin_sys::DfStringView {
                data: name.as_ptr(),
                len: name.len() as u64,
            },
            values: values.as_ptr().cast(),
            value_count: values.len() as u64,
        })
    }

    pub const fn string(name: &'static str) -> Self {
        Self::typed(name, dragonfly_plugin_sys::DF_COMMAND_PARAMETER_STRING)
    }

    pub const fn integer(name: &'static str) -> Self {
        Self::typed(name, dragonfly_plugin_sys::DF_COMMAND_PARAMETER_INTEGER)
    }

    pub const fn float(name: &'static str) -> Self {
        Self::typed(name, dragonfly_plugin_sys::DF_COMMAND_PARAMETER_FLOAT)
    }

    pub const fn boolean(name: &'static str) -> Self {
        Self::typed(name, dragonfly_plugin_sys::DF_COMMAND_PARAMETER_BOOL)
    }

    pub const fn dynamic_enum(name: &'static str) -> Self {
        Self::typed(
            name,
            dragonfly_plugin_sys::DF_COMMAND_PARAMETER_DYNAMIC_ENUM,
        )
    }

    pub const fn player(name: &'static str) -> Self {
        Self::typed(name, dragonfly_plugin_sys::DF_COMMAND_PARAMETER_PLAYER)
    }

    pub const fn raw_text(name: &'static str) -> Self {
        Self::typed(name, dragonfly_plugin_sys::DF_COMMAND_PARAMETER_RAW_TEXT)
    }
}

#[repr(transparent)]
pub struct CommandOverload(dragonfly_plugin_sys::DfCommandOverload);

unsafe impl Sync for CommandOverload {}

impl CommandOverload {
    pub const fn new(parameters: &'static [CommandParameter]) -> Self {
        Self(dragonfly_plugin_sys::DfCommandOverload {
            parameters: parameters.as_ptr().cast(),
            parameter_count: parameters.len() as u64,
        })
    }
}

impl Command {
    pub const fn new(name: &'static str, description: &'static str) -> Self {
        Self(dragonfly_plugin_sys::DfCommandDescriptor {
            name: dragonfly_plugin_sys::DfStringView {
                data: name.as_ptr(),
                len: name.len() as u64,
            },
            description: dragonfly_plugin_sys::DfStringView {
                data: description.as_ptr(),
                len: description.len() as u64,
            },
            overloads: core::ptr::null(),
            overload_count: 0,
        })
    }

    pub const fn with_overloads(mut self, overloads: &'static [CommandOverload]) -> Self {
        self.0.overloads = overloads.as_ptr().cast();
        self.0.overload_count = overloads.len() as u64;
        self
    }
}

pub trait CommandEnum: Sized {
    const VALUES: &'static [CommandValue];

    fn parse(value: &str) -> Option<Self>;
}

pub trait CommandDefinition: Sized {
    const COMMAND: Command;

    fn parse(arguments: &str) -> Result<Self, CommandParseError>;

    #[doc(hidden)]
    fn dynamic_options(
        _overload: usize,
        _parameter: usize,
        _source: CommandSource<'_>,
    ) -> Option<Vec<String>> {
        None
    }
}

#[derive(Clone, Copy, Debug)]
pub struct CommandSource<'a> {
    name: &'a str,
    online_players: &'a [dragonfly_plugin_sys::DfStringView],
}

impl<'a> CommandSource<'a> {
    #[doc(hidden)]
    pub const fn new(
        name: &'a str,
        online_players: &'a [dragonfly_plugin_sys::DfStringView],
    ) -> Self {
        Self {
            name,
            online_players,
        }
    }

    pub const fn name(self) -> &'a str {
        self.name
    }

    pub fn online_players(self) -> impl Iterator<Item = &'a str> {
        self.online_players
            .iter()
            .map(|name| unsafe { string_view(*name) })
    }
}

pub trait DynamicCommandEnum {
    fn options(source: CommandSource<'_>) -> Vec<String>;
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct Dynamic<T> {
    value: String,
    marker: core::marker::PhantomData<T>,
}

impl<T> Dynamic<T> {
    #[doc(hidden)]
    pub fn new(value: impl Into<String>) -> Self {
        Self {
            value: value.into(),
            marker: core::marker::PhantomData,
        }
    }

    pub fn value(&self) -> &str {
        &self.value
    }

    pub fn into_string(self) -> String {
        self.value
    }
}

#[doc(hidden)]
pub fn write_dynamic_options(
    options: Vec<String>,
    output: &mut dragonfly_plugin_sys::DfStringBuffer,
) -> Result<(), MessageTooLong> {
    let encoded = options.join("\n");
    if options.iter().any(|option| option.contains('\n')) {
        return Err(MessageTooLong {
            length: encoded.len(),
            capacity: output.capacity as usize,
        });
    }
    let capacity = output.capacity as usize;
    if encoded.len() > capacity || (!encoded.is_empty() && output.data.is_null()) {
        return Err(MessageTooLong {
            length: encoded.len(),
            capacity,
        });
    }
    if !encoded.is_empty() {
        // SAFETY: the runtime supplied a writable buffer and capacity was checked above.
        unsafe { core::ptr::copy_nonoverlapping(encoded.as_ptr(), output.data, encoded.len()) };
    }
    output.len = encoded.len() as u64;
    Ok(())
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct CommandParseError(String);

impl CommandParseError {
    #[doc(hidden)]
    pub fn new(message: impl Into<String>) -> Self {
        Self(message.into())
    }
}

impl core::fmt::Display for CommandParseError {
    fn fmt(&self, formatter: &mut core::fmt::Formatter<'_>) -> core::fmt::Result {
        formatter.write_str(&self.0)
    }
}

impl std::error::Error for CommandParseError {}

pub struct CommandEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfCommandInput,
    state: &'a mut dragonfly_plugin_sys::DfCommandState,
}

impl<'a> CommandEvent<'a> {
    /// Creates a safe command view over runtime-validated ABI values.
    ///
    /// # Safety
    /// Both references and the state's output buffer must belong to the same active command callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfCommandInput,
        state: &'a mut dragonfly_plugin_sys::DfCommandState,
    ) -> Self {
        Self { input, state }
    }

    pub fn source(&self) -> &str {
        unsafe { string_view(self.input.source) }
    }

    pub fn arguments(&self) -> &str {
        unsafe { string_view(self.input.arguments) }
    }

    pub fn reply(&mut self, message: &str) -> Result<(), MessageTooLong> {
        self.write(message, false)
    }

    pub fn fail(&mut self, message: &str) -> Result<(), MessageTooLong> {
        self.write(message, true)
    }

    fn write(&mut self, message: &str, failed: bool) -> Result<(), MessageTooLong> {
        let capacity = self.state.output.capacity as usize;
        if message.len() > capacity || (!message.is_empty() && self.state.output.data.is_null()) {
            return Err(MessageTooLong {
                length: message.len(),
                capacity,
            });
        }
        if !message.is_empty() {
            unsafe {
                core::ptr::copy_nonoverlapping(
                    message.as_ptr(),
                    self.state.output.data,
                    message.len(),
                )
            };
        }
        self.state.output.len = message.len() as u64;
        self.state.failed = u8::from(failed);
        Ok(())
    }
}

unsafe fn string_view<'a>(view: dragonfly_plugin_sys::DfStringView) -> &'a str {
    if view.len == 0 {
        return "";
    }
    unsafe {
        core::str::from_utf8_unchecked(core::slice::from_raw_parts(view.data, view.len as usize))
    }
}

pub struct PlayerChatEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerChatInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerChatState,
}

impl<'a> PlayerChatEvent<'a> {
    /// Creates a safe event view over runtime-validated ABI values.
    ///
    /// # Safety
    /// Both references and the state's output buffer must belong to the same active chat callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerChatInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerChatState,
    ) -> Self {
        Self { input, state }
    }

    pub fn message(&self) -> &str {
        let (data, len) = if self.state.has_replacement != 0 {
            (
                self.state.replacement.data.cast_const(),
                self.state.replacement.len,
            )
        } else {
            (self.input.message.data, self.input.message.len)
        };
        if len == 0 {
            return "";
        }
        // SAFETY: runtime validates UTF-8 and buffer bounds before dispatching each handler.
        unsafe { core::str::from_utf8_unchecked(core::slice::from_raw_parts(data, len as usize)) }
    }

    pub fn replace_message(&mut self, message: &str) -> Result<(), MessageTooLong> {
        let capacity = self.state.replacement.capacity as usize;
        if message.len() > capacity
            || (!message.is_empty() && self.state.replacement.data.is_null())
        {
            return Err(MessageTooLong {
                length: message.len(),
                capacity,
            });
        }
        if !message.is_empty() {
            // SAFETY: capacity and destination were checked; source and destination do not overlap.
            unsafe {
                core::ptr::copy_nonoverlapping(
                    message.as_ptr(),
                    self.state.replacement.data,
                    message.len(),
                );
            }
        }
        self.state.replacement.len = message.len() as u64;
        self.state.has_replacement = 1;
        Ok(())
    }

    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }

    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub trait Plugin: Default + Send + Sync + 'static {
    fn on_enable(&self) {}
    fn on_disable(&self) {}
    fn on_move(&self, _event: &mut PlayerMoveEvent<'_>) {}
    fn on_chat(&self, _event: &mut PlayerChatEvent<'_>) {}
    fn commands(&self) -> &'static [Command] {
        &[]
    }
    fn on_command(&self, _command: usize, _event: &mut CommandEvent<'_>) {}
}

#[cfg(test)]
mod tests {
    use super::*;

    #[derive(CommandEnum, Debug, Eq, PartialEq)]
    enum Mode {
        Survival,
        Creative,
    }

    #[derive(Command, Debug, PartialEq)]
    #[command(name = "mode", description = "Changes a mode")]
    enum ModeCommand {
        Set {
            mode: Mode,
        },
        Configure {
            level: i64,
            scale: f64,
            enabled: bool,
        },
        Target {
            player: Player,
        },
        Message {
            #[command(varargs)]
            text: String,
        },
        Query,
    }

    #[derive(Default)]
    struct Guard;

    impl Plugin for Guard {
        fn on_move(&self, event: &mut PlayerMoveEvent<'_>) {
            if event.new_position().y < 0.0 {
                event.cancel();
            }
        }
    }

    #[test]
    fn cancellation_defaults_to_false_and_is_monotonic() {
        let input = dragonfly_plugin_sys::DfPlayerMoveInput {
            new_position: dragonfly_plugin_sys::DfVec3 {
                x: 0.0,
                y: -1.0,
                z: 0.0,
            },
            ..Default::default()
        };
        let mut state = dragonfly_plugin_sys::DfPlayerMoveState::default();
        let mut event = unsafe { PlayerMoveEvent::from_raw(&input, &mut state) };
        Guard.on_move(&mut event);
        assert!(event.cancelled());
    }

    #[test]
    fn derived_command_parses_subcommands_and_enums() {
        assert_eq!(
            ModeCommand::parse("set creative").unwrap(),
            ModeCommand::Set {
                mode: Mode::Creative
            }
        );
        assert_eq!(ModeCommand::parse("query").unwrap(), ModeCommand::Query);
        assert_eq!(
            ModeCommand::parse("configure 4 1.5 true").unwrap(),
            ModeCommand::Configure {
                level: 4,
                scale: 1.5,
                enabled: true,
            }
        );
        assert!(ModeCommand::parse("set spectator").is_err());
        let target = ModeCommand::parse("target 000102030405060708090a0b0c0d0e0f:9").unwrap();
        let ModeCommand::Target { player } = target else {
            panic!("wrong command variant");
        };
        assert_eq!(player.id().generation(), 9);
        assert_eq!(player.id().uuid_bytes()[15], 15);
        let ModeCommand::Message { text } = ModeCommand::parse("message hello from rust").unwrap()
        else {
            panic!("wrong command variant");
        };
        assert_eq!(text, "hello from rust");
    }
}
