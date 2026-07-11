//! Safe Rust SDK for native Dragonfly plugins.

pub use dragonfly_plugin_macros::plugin;

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
}

#[cfg(test)]
mod tests {
    use super::*;

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
}
