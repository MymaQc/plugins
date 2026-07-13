//! Entity effects exposed using Dragonfly's lasting and instant type split.

use std::time::Duration;

mod private {
    pub trait Sealed {}
}

/// A Dragonfly effect type registered under a numeric ID.
pub trait Type: private::Sealed + Copy {
    /// Returns the registered effect ID.
    fn id(self) -> i32;
}

/// An effect type that may be applied for a duration.
pub trait LastingType: Type {}

/// An effect type that is applied once.
pub trait InstantType: Type {}

/// A lasting effect type registered by the server or another bootstrap component.
#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct RegisteredLasting {
    id: i32,
}

impl RegisteredLasting {
    pub const fn new(id: i32) -> Self {
        Self { id }
    }
}

impl private::Sealed for RegisteredLasting {}
impl Type for RegisteredLasting {
    fn id(self) -> i32 {
        self.id
    }
}
impl LastingType for RegisteredLasting {}

/// An instant effect type registered by the server or another bootstrap component.
#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct RegisteredInstant {
    id: i32,
}

impl RegisteredInstant {
    pub const fn new(id: i32) -> Self {
        Self { id }
    }
}

impl private::Sealed for RegisteredInstant {}
impl Type for RegisteredInstant {
    fn id(self) -> i32 {
        self.id
    }
}
impl InstantType for RegisteredInstant {}

#[derive(Clone, Copy, Debug, PartialEq)]
pub struct Effect {
    pub(crate) effect_type: i32,
    pub(crate) level: i32,
    pub(crate) duration: Duration,
    pub(crate) potency: f64,
    pub(crate) mode: u32,
    pub(crate) particles_hidden: bool,
}

fn lasting(effect_type: impl LastingType, level: i32, duration: Duration, mode: u32) -> Effect {
    Effect {
        effect_type: effect_type.id(),
        level,
        duration,
        potency: 1.0,
        mode,
        particles_hidden: false,
    }
}

/// Creates a lasting effect.
pub fn new(effect_type: impl LastingType, level: i32, duration: Duration) -> Effect {
    lasting(
        effect_type,
        level,
        duration,
        dragonfly_plugin_sys::DF_EFFECT_MODE_TIMED,
    )
}

/// Creates an ambient lasting effect.
pub fn ambient(effect_type: impl LastingType, level: i32, duration: Duration) -> Effect {
    lasting(
        effect_type,
        level,
        duration,
        dragonfly_plugin_sys::DF_EFFECT_MODE_AMBIENT,
    )
}

/// Creates an infinite lasting effect.
pub fn infinite(effect_type: impl LastingType, level: i32) -> Effect {
    lasting(
        effect_type,
        level,
        Duration::ZERO,
        dragonfly_plugin_sys::DF_EFFECT_MODE_INFINITE,
    )
}

/// Creates an instant effect with Dragonfly's default potency of `1.0`.
pub fn instant(effect_type: impl InstantType, level: i32) -> Effect {
    instant_with_potency(effect_type, level, 1.0)
}

/// Creates an instant effect with an explicit potency.
pub fn instant_with_potency(effect_type: impl InstantType, level: i32, potency: f64) -> Effect {
    Effect {
        effect_type: effect_type.id(),
        level,
        duration: Duration::ZERO,
        potency,
        mode: dragonfly_plugin_sys::DF_EFFECT_MODE_INSTANT,
        particles_hidden: false,
    }
}

impl Effect {
    /// Returns this effect with client particles hidden.
    pub fn without_particles(mut self) -> Self {
        self.particles_hidden = true;
        self
    }
}

include!("effects_generated.rs");

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn constructors_preserve_dragonfly_effect_semantics() {
        let lasting = new(Speed, 2, Duration::from_secs(5));
        assert_eq!(lasting.effect_type, 1);
        assert_eq!(lasting.mode, dragonfly_plugin_sys::DF_EFFECT_MODE_TIMED);
        assert_eq!(lasting.potency, 1.0);

        let ambient = ambient(Regeneration, 1, Duration::from_secs(2));
        assert_eq!(ambient.mode, dragonfly_plugin_sys::DF_EFFECT_MODE_AMBIENT);
        let infinite = infinite(FireResistance, 1);
        assert_eq!(infinite.mode, dragonfly_plugin_sys::DF_EFFECT_MODE_INFINITE);
        assert_eq!(infinite.duration, Duration::ZERO);

        let instant = instant_with_potency(InstantHealth, 1, 0.5).without_particles();
        assert_eq!(instant.effect_type, 6);
        assert_eq!(instant.mode, dragonfly_plugin_sys::DF_EFFECT_MODE_INSTANT);
        assert_eq!(instant.potency, 0.5);
        assert!(instant.particles_hidden);
    }

    #[test]
    fn registered_effect_types_keep_their_signed_ids() {
        assert_eq!(RegisteredLasting::new(-7).id(), -7);
        assert_eq!(RegisteredInstant::new(42_000).id(), 42_000);
    }
}
