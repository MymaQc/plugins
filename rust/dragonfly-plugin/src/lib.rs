//! Safe Rust SDK for native Dragonfly plugins.

extern crate self as dragonfly_plugin;

use core::sync::atomic::{AtomicPtr, Ordering};

pub use dragonfly_plugin_macros::{Command, CommandEnum, plugin};

#[doc(hidden)]
pub mod __private {
    pub use core::ffi::c_void;
    pub use dragonfly_plugin_sys as sys;
}

static HOST_API: AtomicPtr<dragonfly_plugin_sys::DfHostApiV1> =
    AtomicPtr::new(core::ptr::null_mut());

#[doc(hidden)]
/// # Safety
/// `host` must remain valid while plugin callbacks may execute.
pub unsafe fn install_host(host: *const dragonfly_plugin_sys::DfHostApiV1) {
    HOST_API.store(host.cast_mut(), Ordering::Release);
}

#[derive(Clone, Copy, Debug, Default, PartialEq)]
pub struct Vec3 {
    pub x: f64,
    pub y: f64,
    pub z: f64,
}

#[derive(Clone, Copy, Debug, Default, PartialEq)]
pub struct Rotation {
    pub yaw: f64,
    pub pitch: f64,
}

#[derive(Clone, Copy, Debug, Default, Eq, Hash, PartialEq)]
pub struct BlockPos {
    pub x: i32,
    pub y: i32,
    pub z: i32,
}

#[repr(i32)]
#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub enum BlockFace {
    Down = 0,
    Up = 1,
    North = 2,
    South = 3,
    West = 4,
    East = 5,
}

#[derive(Clone, Copy, Debug)]
pub struct ItemStack<'a> {
    raw: &'a dragonfly_plugin_sys::DfItemStackView,
}

impl<'a> ItemStack<'a> {
    fn from_raw(raw: &'a dragonfly_plugin_sys::DfItemStackView) -> Self {
        Self { raw }
    }
    pub fn identifier(&self) -> &'a str {
        unsafe { string_view(self.raw.identifier) }
    }
    pub fn metadata(&self) -> i32 {
        self.raw.metadata
    }
    pub fn count(&self) -> i32 {
        self.raw.count
    }
    pub fn damage(&self) -> i32 {
        self.raw.damage
    }
}

impl From<dragonfly_plugin_sys::DfBlockPos> for BlockPos {
    fn from(value: dragonfly_plugin_sys::DfBlockPos) -> Self {
        Self {
            x: value.x,
            y: value.y,
            z: value.z,
        }
    }
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

impl From<dragonfly_plugin_sys::DfPlayerId> for PlayerId {
    fn from(value: dragonfly_plugin_sys::DfPlayerId) -> Self {
        Self {
            uuid: value.bytes,
            generation: value.generation,
        }
    }
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct Title {
    text: String,
    subtitle: String,
    action_text: String,
    fade_in: std::time::Duration,
    duration: std::time::Duration,
    fade_out: std::time::Duration,
}

impl Title {
    pub fn new(text: impl Into<String>) -> Self {
        Self {
            text: text.into(),
            subtitle: String::new(),
            action_text: String::new(),
            fade_in: std::time::Duration::from_millis(50),
            duration: std::time::Duration::from_secs(2),
            fade_out: std::time::Duration::from_millis(50),
        }
    }
    pub fn subtitle(mut self, text: impl Into<String>) -> Self {
        self.subtitle = text.into();
        self
    }
    pub fn action_text(mut self, text: impl Into<String>) -> Self {
        self.action_text = text.into();
        self
    }
    pub fn fade_in(mut self, duration: std::time::Duration) -> Self {
        self.fade_in = duration;
        self
    }
    pub fn duration(mut self, duration: std::time::Duration) -> Self {
        self.duration = duration;
        self
    }
    pub fn fade_out(mut self, duration: std::time::Duration) -> Self {
        self.fade_out = duration;
        self
    }
}

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct Player {
    id: PlayerId,
    latency_milliseconds: Option<u64>,
    name: Option<PlayerName>,
}

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
struct PlayerName {
    bytes: [u8; 64],
    len: u8,
}

impl Player {
    pub const fn id(self) -> PlayerId {
        self.id
    }

    pub fn latency(&self) -> Option<std::time::Duration> {
        self.latency_milliseconds
            .map(std::time::Duration::from_millis)
    }

    pub fn name(&self) -> Option<&str> {
        self.name.as_ref().map(|name| {
            // Names are copied from runtime-validated UTF-8 or parsed Rust strings.
            unsafe { core::str::from_utf8_unchecked(&name.bytes[..name.len as usize]) }
        })
    }

    pub fn message(&self, message: &str) {
        self.send_text(dragonfly_plugin_sys::DF_PLAYER_TEXT_MESSAGE, message);
    }

    pub fn send_tip(&self, message: &str) {
        self.send_text(dragonfly_plugin_sys::DF_PLAYER_TEXT_TIP, message);
    }

    pub fn send_popup(&self, message: &str) {
        self.send_text(dragonfly_plugin_sys::DF_PLAYER_TEXT_POPUP, message);
    }

    pub fn send_jukebox_popup(&self, message: &str) {
        self.send_text(dragonfly_plugin_sys::DF_PLAYER_TEXT_JUKEBOX_POPUP, message);
    }

    pub fn send_title(&self, title: &Title) {
        let host = HOST_API.load(Ordering::Acquire);
        let Some(host) = (unsafe { host.as_ref() }) else {
            return;
        };
        let Some(send) = host.player_title else {
            return;
        };
        let view = dragonfly_plugin_sys::DfTitleView {
            text: string_view_from_str(&title.text),
            subtitle: string_view_from_str(&title.subtitle),
            action_text: string_view_from_str(&title.action_text),
            fade_in_milliseconds: duration_milliseconds(title.fade_in),
            duration_milliseconds: duration_milliseconds(title.duration),
            fade_out_milliseconds: duration_milliseconds(title.fade_out),
        };
        let _ = unsafe { send(host.context, self.raw_id(), view) };
    }

    pub fn teleport(&self, position: Vec3) {
        self.transform(
            dragonfly_plugin_sys::DF_PLAYER_TRANSFORM_TELEPORT,
            position,
            0.0,
            0.0,
        );
    }

    pub fn move_by(&self, delta: Vec3, delta_yaw: f64, delta_pitch: f64) {
        self.transform(
            dragonfly_plugin_sys::DF_PLAYER_TRANSFORM_MOVE,
            delta,
            delta_yaw,
            delta_pitch,
        );
    }

    pub fn set_velocity(&self, velocity: Vec3) {
        self.transform(
            dragonfly_plugin_sys::DF_PLAYER_TRANSFORM_VELOCITY,
            velocity,
            0.0,
            0.0,
        );
    }

    pub fn rotation(&self) -> Rotation {
        let host = HOST_API.load(Ordering::Acquire);
        let Some(host) = (unsafe { host.as_ref() }) else {
            return Rotation::default();
        };
        let Some(read) = host.player_rotation else {
            return Rotation::default();
        };
        let mut raw = dragonfly_plugin_sys::DfRotation::default();
        let _ = unsafe { read(host.context, self.raw_id(), &mut raw) };
        Rotation {
            yaw: raw.yaw,
            pitch: raw.pitch,
        }
    }

    pub fn face(&self, rotation: Rotation) {
        let current = self.rotation();
        self.move_by(
            Vec3::default(),
            rotation.yaw - current.yaw,
            rotation.pitch - current.pitch,
        );
    }

    fn send_text(&self, kind: u32, message: &str) {
        let host = HOST_API.load(Ordering::Acquire);
        let Some(host) = (unsafe { host.as_ref() }) else {
            return;
        };
        let Some(send) = host.player_text else {
            return;
        };
        let _ = unsafe {
            send(
                host.context,
                self.raw_id(),
                kind,
                string_view_from_str(message),
            )
        };
    }

    fn transform(&self, kind: u32, vector: Vec3, yaw: f64, pitch: f64) {
        let host = HOST_API.load(Ordering::Acquire);
        let Some(host) = (unsafe { host.as_ref() }) else {
            return;
        };
        let Some(transform) = host.player_transform else {
            return;
        };
        let raw = dragonfly_plugin_sys::DfVec3 {
            x: vector.x,
            y: vector.y,
            z: vector.z,
        };
        let _ = unsafe { transform(host.context, self.raw_id(), kind, raw, yaw, pitch) };
    }

    fn raw_id(&self) -> dragonfly_plugin_sys::DfPlayerId {
        dragonfly_plugin_sys::DfPlayerId {
            bytes: self.id.uuid,
            generation: self.id.generation,
        }
    }

    fn from_id(id: dragonfly_plugin_sys::DfPlayerId) -> Self {
        Self {
            id: id.into(),
            latency_milliseconds: None,
            name: None,
        }
    }

    fn from_snapshot(id: dragonfly_plugin_sys::DfPlayerId, latency: u64, name: &str) -> Self {
        let mut player = Self::from_id(id);
        player.latency_milliseconds = Some(latency);
        if name.len() <= 64 {
            let mut bytes = [0; 64];
            bytes[..name.len()].copy_from_slice(name.as_bytes());
            player.name = Some(PlayerName {
                bytes,
                len: name.len() as u8,
            });
        }
        player
    }

    #[doc(hidden)]
    pub fn from_command_argument(value: &str) -> Result<Self, CommandParseError> {
        let mut parts = value.split(':');
        let uuid = parts
            .next()
            .ok_or_else(|| CommandParseError::new("player is no longer online"))?;
        let generation = parts
            .next()
            .ok_or_else(|| CommandParseError::new("player is no longer online"))?;
        let latency_milliseconds = parts.next().and_then(|value| value.parse().ok());
        let name = parts.next();
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
        let mut player = Self {
            id: PlayerId {
                uuid: bytes,
                generation,
            },
            latency_milliseconds,
            name: None,
        };
        if let Some(name) = name
            && name.len() <= 64
        {
            let mut bytes = [0; 64];
            bytes[..name.len()].copy_from_slice(name.as_bytes());
            player.name = Some(PlayerName {
                bytes,
                len: name.len() as u8,
            });
        }
        Ok(player)
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

pub struct PlayerJumpEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerJumpInput,
}

impl<'a> PlayerJumpEvent<'a> {
    /// # Safety
    /// The input must belong to the active jump callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(input: &'a dragonfly_plugin_sys::DfPlayerJumpInput) -> Self {
        Self { input }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
}

pub struct PlayerTeleportEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerTeleportInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerTeleportState,
}

impl<'a> PlayerTeleportEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active teleport callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerTeleportInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerTeleportState,
    ) -> Self {
        Self { input, state }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn position(&self) -> Vec3 {
        self.input.position.into()
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
            optional: 0,
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
            optional: 0,
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

    pub const fn optional(mut self) -> Self {
        self.0.optional = 1;
        self
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

#[derive(Clone, Debug, Default, Eq, PartialEq)]
pub struct Varargs(String);

impl Varargs {
    #[doc(hidden)]
    pub fn new(value: String) -> Self {
        Self(value)
    }

    pub fn value(&self) -> &str {
        &self.0
    }

    pub fn into_string(self) -> String {
        self.0
    }
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

pub struct Any;
pub struct Console;

pub struct Context<'a, S = Any> {
    input: &'a dragonfly_plugin_sys::DfCommandInput,
    state: &'a mut dragonfly_plugin_sys::DfCommandState,
    source: S,
}

impl<'a> Context<'a, Any> {
    /// Creates a safe command view over runtime-validated ABI values.
    ///
    /// # Safety
    /// Both references and the state's output buffer must belong to the same active command callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfCommandInput,
        state: &'a mut dragonfly_plugin_sys::DfCommandState,
    ) -> Self {
        Self {
            input,
            state,
            source: Any,
        }
    }

    pub fn player_context(&mut self) -> Option<Context<'_, Player>> {
        let source = self.source_player()?;
        Some(Context {
            input: self.input,
            state: &mut *self.state,
            source,
        })
    }

    pub fn console_context(&mut self) -> Option<Context<'_, Console>> {
        (self.input.source_kind == dragonfly_plugin_sys::DF_COMMAND_SOURCE_CONSOLE).then_some({
            Context {
                input: self.input,
                state: &mut *self.state,
                source: Console,
            }
        })
    }
}

impl<S> Context<'_, S> {
    pub fn source_name(&self) -> &str {
        unsafe { string_view(self.input.source) }
    }

    pub fn arguments(&self) -> &str {
        unsafe { string_view(self.input.arguments) }
    }

    pub fn source_player(&self) -> Option<Player> {
        if self.input.source_kind != dragonfly_plugin_sys::DF_COMMAND_SOURCE_PLAYER {
            return None;
        }
        let source = PlayerId::from(self.input.source_player);
        let snapshot = self
            .online_players()
            .iter()
            .find(|candidate| PlayerId::from(candidate.player) == source)?;
        Some(Player::from_snapshot(
            snapshot.player,
            snapshot.latency_milliseconds,
            unsafe { string_view(snapshot.name) },
        ))
    }

    fn online_players(&self) -> &[dragonfly_plugin_sys::DfCommandPlayer] {
        if self.input.online_player_count == 0 {
            &[][..]
        } else {
            // SAFETY: runtime validates this array before invoking plugin command handlers.
            unsafe {
                core::slice::from_raw_parts(
                    self.input.online_players,
                    self.input.online_player_count as usize,
                )
            }
        }
    }
}

impl Context<'_, Player> {
    pub fn source(&self) -> Player {
        self.source
    }
}

impl<S> Context<'_, S> {
    pub fn reply(&mut self, message: &str) {
        if self.try_reply(message).is_err() {
            self.state.failed = 1;
            self.state.output.len = 0;
            let _ = self.write("command output exceeded the runtime buffer", true);
        }
    }

    pub fn fail(&mut self, message: &str) {
        if self.try_fail(message).is_err() {
            self.state.failed = 1;
            self.state.output.len = 0;
            let _ = self.write("command error exceeded the runtime buffer", true);
        }
    }

    pub fn try_reply(&mut self, message: &str) -> Result<(), MessageTooLong> {
        self.write(message, false)
    }

    pub fn try_fail(&mut self, message: &str) -> Result<(), MessageTooLong> {
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

fn string_view_from_str(value: &str) -> dragonfly_plugin_sys::DfStringView {
    dragonfly_plugin_sys::DfStringView {
        data: value.as_ptr(),
        len: value.len() as u64,
    }
}

fn duration_milliseconds(duration: std::time::Duration) -> u64 {
    duration.as_millis().min(u128::from(u64::MAX)) as u64
}

pub struct PlayerChatEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerChatInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerChatState,
}

pub struct PlayerJoinEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerJoinInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerJoinState,
}

impl<'a> PlayerJoinEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active join callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerJoinInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerJoinState,
    ) -> Self {
        Self { input, state }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }

    pub fn name(&self) -> &str {
        unsafe { string_view(self.input.name) }
    }

    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }

    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerQuitEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerQuitInput,
}

impl<'a> PlayerQuitEvent<'a> {
    /// # Safety
    /// The reference must belong to an active quit callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(input: &'a dragonfly_plugin_sys::DfPlayerQuitInput) -> Self {
        Self { input }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }

    pub fn name(&self) -> &str {
        unsafe { string_view(self.input.name) }
    }
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

pub struct PlayerHurtEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerHurtInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerHurtState,
}

impl<'a> PlayerHurtEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active hurt callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerHurtInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerHurtState,
    ) -> Self {
        Self { input, state }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }

    pub fn source(&self) -> &str {
        unsafe { string_view(self.input.source) }
    }

    pub fn immune(&self) -> bool {
        self.input.immune != 0
    }

    pub fn damage(&self) -> f64 {
        self.state.damage
    }

    pub fn set_damage(&mut self, damage: f64) {
        self.state.damage = damage.max(0.0);
    }

    pub fn attack_immunity(&self) -> std::time::Duration {
        std::time::Duration::from_millis(self.state.attack_immunity_milliseconds)
    }

    pub fn set_attack_immunity(&mut self, duration: std::time::Duration) {
        self.state.attack_immunity_milliseconds =
            duration.as_millis().min(u128::from(u64::MAX)) as u64;
    }

    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }

    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerHealEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerHealInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerHealState,
}

pub struct PlayerBlockBreakEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerBlockBreakInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerBlockBreakState,
}

impl<'a> PlayerBlockBreakEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active block-break callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerBlockBreakInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerBlockBreakState,
    ) -> Self {
        Self { input, state }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }

    pub fn position(&self) -> BlockPos {
        self.input.position.into()
    }

    pub fn block(&self) -> &str {
        unsafe { string_view(self.input.block) }
    }

    pub fn experience(&self) -> i32 {
        self.state.experience
    }

    pub fn set_experience(&mut self, experience: i32) {
        self.state.experience = experience.max(0);
    }

    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }

    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerBlockPlaceEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerBlockPlaceInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerBlockPlaceState,
}

impl<'a> PlayerBlockPlaceEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active block-place callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerBlockPlaceInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerBlockPlaceState,
    ) -> Self {
        Self { input, state }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }

    pub fn position(&self) -> BlockPos {
        self.input.position.into()
    }

    pub fn block(&self) -> &str {
        unsafe { string_view(self.input.block) }
    }

    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }

    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

impl<'a> PlayerHealEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active heal callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerHealInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerHealState,
    ) -> Self {
        Self { input, state }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }

    pub fn source(&self) -> &str {
        unsafe { string_view(self.input.source) }
    }

    pub fn health(&self) -> f64 {
        self.state.health
    }

    pub fn set_health(&mut self, health: f64) {
        self.state.health = health.max(0.0);
    }

    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }

    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerFoodLossEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerFoodLossInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerFoodLossState,
}

impl<'a> PlayerFoodLossEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active food-loss callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerFoodLossInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerFoodLossState,
    ) -> Self {
        Self { input, state }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }

    pub fn from(&self) -> i32 {
        self.input.from
    }

    pub fn to(&self) -> i32 {
        self.state.to
    }

    pub fn set_to(&mut self, food: i32) {
        self.state.to = food.clamp(0, 20);
    }

    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }

    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerDeathEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerDeathInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerDeathState,
}

impl<'a> PlayerDeathEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active death callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerDeathInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerDeathState,
    ) -> Self {
        Self { input, state }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }

    pub fn source(&self) -> &str {
        unsafe { string_view(self.input.source) }
    }

    pub fn keep_inventory(&self) -> bool {
        self.state.keep_inventory != 0
    }

    pub fn set_keep_inventory(&mut self, keep: bool) {
        self.state.keep_inventory = u8::from(keep);
    }
}

macro_rules! cancellable_position_event {
    ($name:ident, $input:ty, $state:ty) => {
        pub struct $name<'a> {
            input: &'a $input,
            state: &'a mut $state,
        }

        impl<'a> $name<'a> {
            /// # Safety
            /// Both references must belong to the same active event callback.
            #[doc(hidden)]
            pub unsafe fn from_raw(input: &'a $input, state: &'a mut $state) -> Self {
                Self { input, state }
            }

            pub fn player(&self) -> Player {
                Player::from_id(self.input.player)
            }

            pub fn position(&self) -> BlockPos {
                self.input.position.into()
            }

            pub fn cancelled(&self) -> bool {
                self.state.cancelled != 0
            }

            pub fn cancel(&mut self) {
                self.state.cancelled = 1;
            }
        }
    };
}

cancellable_position_event!(
    PlayerStartBreakEvent,
    dragonfly_plugin_sys::DfPlayerStartBreakInput,
    dragonfly_plugin_sys::DfPlayerStartBreakState
);

macro_rules! cancellable_toggle_event {
    ($name:ident, $input:ty, $state:ty) => {
        pub struct $name<'a> {
            input: &'a $input,
            state: &'a mut $state,
        }

        impl<'a> $name<'a> {
            /// # Safety
            /// Both references must belong to the same active toggle callback.
            #[doc(hidden)]
            pub unsafe fn from_raw(input: &'a $input, state: &'a mut $state) -> Self {
                Self { input, state }
            }

            pub fn player(&self) -> Player {
                Player::from_id(self.input.player)
            }

            pub fn after(&self) -> bool {
                self.input.after != 0
            }

            pub fn cancelled(&self) -> bool {
                self.state.cancelled != 0
            }

            pub fn cancel(&mut self) {
                self.state.cancelled = 1;
            }
        }
    };
}

cancellable_toggle_event!(
    PlayerToggleSprintEvent,
    dragonfly_plugin_sys::DfPlayerToggleSprintInput,
    dragonfly_plugin_sys::DfPlayerToggleSprintState
);
cancellable_toggle_event!(
    PlayerToggleSneakEvent,
    dragonfly_plugin_sys::DfPlayerToggleSneakInput,
    dragonfly_plugin_sys::DfPlayerToggleSneakState
);
cancellable_position_event!(
    PlayerFireExtinguishEvent,
    dragonfly_plugin_sys::DfPlayerFireExtinguishInput,
    dragonfly_plugin_sys::DfPlayerFireExtinguishState
);

pub struct PlayerExperienceGainEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerExperienceGainInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerExperienceGainState,
}

impl<'a> PlayerExperienceGainEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active experience-gain callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerExperienceGainInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerExperienceGainState,
    ) -> Self {
        Self { input, state }
    }
    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn amount(&self) -> i32 {
        self.state.amount
    }
    pub fn set_amount(&mut self, amount: i32) {
        self.state.amount = amount.max(0);
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerPunchAirEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerPunchAirInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerPunchAirState,
}

pub struct PlayerHeldSlotChangeEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerHeldSlotChangeInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerHeldSlotChangeState,
}

impl<'a> PlayerHeldSlotChangeEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active held-slot callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerHeldSlotChangeInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerHeldSlotChangeState,
    ) -> Self {
        Self { input, state }
    }
    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn from(&self) -> i32 {
        self.input.from
    }
    pub fn to(&self) -> i32 {
        self.input.to
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerSleepEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerSleepInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerSleepState,
}

impl<'a> PlayerSleepEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active sleep callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerSleepInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerSleepState,
    ) -> Self {
        Self { input, state }
    }
    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn send_reminder(&self) -> bool {
        self.state.send_reminder != 0
    }
    pub fn set_send_reminder(&mut self, send: bool) {
        self.state.send_reminder = u8::from(send);
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

impl<'a> PlayerPunchAirEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active punch-air callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerPunchAirInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerPunchAirState,
    ) -> Self {
        Self { input, state }
    }
    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerBlockPickEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerBlockPickInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerBlockPickState,
}

impl<'a> PlayerBlockPickEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active block-pick callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerBlockPickInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerBlockPickState,
    ) -> Self {
        Self { input, state }
    }
    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn position(&self) -> BlockPos {
        self.input.position.into()
    }
    pub fn block(&self) -> &str {
        unsafe { string_view(self.input.block) }
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerLecternPageTurnEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerLecternPageTurnInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerLecternPageTurnState,
}

pub struct PlayerSignEditEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerSignEditInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerSignEditState,
}

impl<'a> PlayerSignEditEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active sign-edit callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerSignEditInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerSignEditState,
    ) -> Self {
        Self { input, state }
    }
    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn position(&self) -> BlockPos {
        self.input.position.into()
    }
    pub fn front_side(&self) -> bool {
        self.input.front_side != 0
    }
    pub fn old_text(&self) -> &str {
        unsafe { string_view(self.input.old_text) }
    }
    pub fn new_text(&self) -> &str {
        unsafe { string_view(self.input.new_text) }
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerItemUseEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerItemUseInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerItemUseState,
}

pub struct PlayerItemUseOnBlockEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerItemUseOnBlockInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerItemUseOnBlockState,
}

pub struct PlayerItemConsumeEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerItemConsumeInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerItemConsumeState,
}

impl<'a> PlayerItemConsumeEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active item-consume callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerItemConsumeInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerItemConsumeState,
    ) -> Self {
        Self { input, state }
    }
    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn item(&self) -> ItemStack<'_> {
        ItemStack::from_raw(&self.input.item)
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerItemReleaseEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerItemReleaseInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerItemReleaseState,
}

pub struct PlayerItemDamageEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerItemDamageInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerItemDamageState,
}

impl<'a> PlayerItemDamageEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active item-damage callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerItemDamageInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerItemDamageState,
    ) -> Self {
        Self { input, state }
    }
    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn item(&self) -> ItemStack<'_> {
        ItemStack::from_raw(&self.input.item)
    }
    pub fn damage(&self) -> i32 {
        self.state.damage
    }
    pub fn set_damage(&mut self, damage: i32) {
        self.state.damage = damage.max(0);
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerItemDropEvent<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerItemDropInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerItemDropState,
}

impl<'a> PlayerItemDropEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active item-drop callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerItemDropInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerItemDropState,
    ) -> Self {
        Self { input, state }
    }
    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn item(&self) -> ItemStack<'_> {
        ItemStack::from_raw(&self.input.item)
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

impl<'a> PlayerItemReleaseEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active item-release callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerItemReleaseInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerItemReleaseState,
    ) -> Self {
        Self { input, state }
    }
    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn item(&self) -> ItemStack<'_> {
        ItemStack::from_raw(&self.input.item)
    }
    pub fn duration(&self) -> std::time::Duration {
        std::time::Duration::from_millis(self.input.duration_milliseconds)
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

impl<'a> PlayerItemUseOnBlockEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active item-use-on-block callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerItemUseOnBlockInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerItemUseOnBlockState,
    ) -> Self {
        Self { input, state }
    }
    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn position(&self) -> BlockPos {
        self.input.position.into()
    }
    pub fn face(&self) -> BlockFace {
        // Runtime validates the generated integer before constructing this event.
        unsafe { core::mem::transmute(self.input.face) }
    }
    pub fn click_position(&self) -> Vec3 {
        self.input.click_position.into()
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

impl<'a> PlayerItemUseEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active item-use callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerItemUseInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerItemUseState,
    ) -> Self {
        Self { input, state }
    }
    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

impl<'a> PlayerLecternPageTurnEvent<'a> {
    /// # Safety
    /// Both references must belong to the same active lectern callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerLecternPageTurnInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerLecternPageTurnState,
    ) -> Self {
        Self { input, state }
    }
    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn position(&self) -> BlockPos {
        self.input.position.into()
    }
    pub fn old_page(&self) -> i32 {
        self.input.old_page
    }
    pub fn new_page(&self) -> i32 {
        self.state.new_page
    }
    pub fn set_new_page(&mut self, page: i32) {
        self.state.new_page = page.max(0);
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
    fn on_join(&self, _event: &mut PlayerJoinEvent<'_>) {}
    fn on_quit(&self, _event: &PlayerQuitEvent<'_>) {}
    fn on_hurt(&self, _event: &mut PlayerHurtEvent<'_>) {}
    fn on_heal(&self, _event: &mut PlayerHealEvent<'_>) {}
    fn on_block_break(&self, _event: &mut PlayerBlockBreakEvent<'_>) {}
    fn on_block_place(&self, _event: &mut PlayerBlockPlaceEvent<'_>) {}
    fn on_food_loss(&self, _event: &mut PlayerFoodLossEvent<'_>) {}
    fn on_death(&self, _event: &mut PlayerDeathEvent<'_>) {}
    fn on_start_break(&self, _event: &mut PlayerStartBreakEvent<'_>) {}
    fn on_fire_extinguish(&self, _event: &mut PlayerFireExtinguishEvent<'_>) {}
    fn on_toggle_sprint(&self, _event: &mut PlayerToggleSprintEvent<'_>) {}
    fn on_toggle_sneak(&self, _event: &mut PlayerToggleSneakEvent<'_>) {}
    fn on_jump(&self, _event: &PlayerJumpEvent<'_>) {}
    fn on_teleport(&self, _event: &mut PlayerTeleportEvent<'_>) {}
    fn on_experience_gain(&self, _event: &mut PlayerExperienceGainEvent<'_>) {}
    fn on_punch_air(&self, _event: &mut PlayerPunchAirEvent<'_>) {}
    fn on_held_slot_change(&self, _event: &mut PlayerHeldSlotChangeEvent<'_>) {}
    fn on_sleep(&self, _event: &mut PlayerSleepEvent<'_>) {}
    fn on_block_pick(&self, _event: &mut PlayerBlockPickEvent<'_>) {}
    fn on_lectern_page_turn(&self, _event: &mut PlayerLecternPageTurnEvent<'_>) {}
    fn on_sign_edit(&self, _event: &mut PlayerSignEditEvent<'_>) {}
    fn on_item_use(&self, _event: &mut PlayerItemUseEvent<'_>) {}
    fn on_item_use_on_block(&self, _event: &mut PlayerItemUseOnBlockEvent<'_>) {}
    fn on_item_consume(&self, _event: &mut PlayerItemConsumeEvent<'_>) {}
    fn on_item_release(&self, _event: &mut PlayerItemReleaseEvent<'_>) {}
    fn on_item_damage(&self, _event: &mut PlayerItemDamageEvent<'_>) {}
    fn on_item_drop(&self, _event: &mut PlayerItemDropEvent<'_>) {}
    fn commands(&self) -> &'static [Command] {
        &[]
    }
    fn on_command(&self, _command: usize, _context: &mut Context<'_>) {}
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
            text: Varargs,
        },
        Maybe {
            value: Option<i64>,
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
        assert_eq!(player.latency(), None);
        let player =
            Player::from_command_argument("000102030405060708090a0b0c0d0e0f:9:42:Danick").unwrap();
        assert_eq!(player.latency(), Some(std::time::Duration::from_millis(42)));
        assert_eq!(player.name(), Some("Danick"));
        let ModeCommand::Message { text } = ModeCommand::parse("message hello from rust").unwrap()
        else {
            panic!("wrong command variant");
        };
        assert_eq!(text.value(), "hello from rust");
        assert_eq!(
            ModeCommand::parse("maybe").unwrap(),
            ModeCommand::Maybe { value: None }
        );
        assert_eq!(
            ModeCommand::parse("maybe 12").unwrap(),
            ModeCommand::Maybe { value: Some(12) }
        );
    }
}
