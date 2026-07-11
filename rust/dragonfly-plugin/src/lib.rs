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

pub trait Plugin: Default + Send + Sync + 'static {
    fn on_move(&self, _event: &mut PlayerMoveEvent<'_>) {}
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
