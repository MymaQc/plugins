use core::ops::{BitOr, BitOrAssign};

use crate::{Enchantment, Entity, block};

#[derive(Clone, Copy, Debug, Default, Eq, Hash, PartialEq)]
pub struct Traits {
    reduced_by_armour: bool,
    reduced_by_resistance: bool,
    fire: bool,
    ignores_totem: bool,
}

impl Traits {
    pub const fn new(
        reduced_by_armour: bool,
        reduced_by_resistance: bool,
        fire: bool,
        ignores_totem: bool,
    ) -> Self {
        Self {
            reduced_by_armour,
            reduced_by_resistance,
            fire,
            ignores_totem,
        }
    }

    pub const fn reduced_by_armour(self) -> bool {
        self.reduced_by_armour
    }

    pub const fn reduced_by_resistance(self) -> bool {
        self.reduced_by_resistance
    }

    pub const fn is_fire(self) -> bool {
        self.fire
    }

    pub const fn ignores_totem(self) -> bool {
        self.ignores_totem
    }

    const fn flags(self) -> u32 {
        (if self.reduced_by_armour {
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_REDUCED_BY_ARMOUR
        } else {
            0
        }) | (if self.reduced_by_resistance {
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_REDUCED_BY_RESISTANCE
        } else {
            0
        }) | (if self.fire {
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_FIRE
        } else {
            0
        }) | (if self.ignores_totem {
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_IGNORES_TOTEM
        } else {
            0
        })
    }

    const fn from_flags(flags: u32) -> Self {
        Self::new(
            flags & dragonfly_plugin_sys::DF_DAMAGE_SOURCE_REDUCED_BY_ARMOUR != 0,
            flags & dragonfly_plugin_sys::DF_DAMAGE_SOURCE_REDUCED_BY_RESISTANCE != 0,
            flags & dragonfly_plugin_sys::DF_DAMAGE_SOURCE_FIRE != 0,
            flags & dragonfly_plugin_sys::DF_DAMAGE_SOURCE_IGNORES_TOTEM != 0,
        )
    }
}

#[derive(Clone, Copy, Debug, Default, Eq, Hash, PartialEq)]
pub struct AffectedProtections(u32);

impl AffectedProtections {
    pub const NONE: Self = Self(0);
    pub const FIRE: Self = Self(dragonfly_plugin_sys::DF_DAMAGE_SOURCE_FIRE_PROTECTION);
    pub const FEATHER_FALLING: Self = Self(dragonfly_plugin_sys::DF_DAMAGE_SOURCE_FEATHER_FALLING);
    pub const BLAST: Self = Self(dragonfly_plugin_sys::DF_DAMAGE_SOURCE_BLAST_PROTECTION);
    pub const PROJECTILE: Self = Self(dragonfly_plugin_sys::DF_DAMAGE_SOURCE_PROJECTILE_PROTECTION);

    pub const fn contains(self, other: Self) -> bool {
        self.0 & other.0 == other.0
    }

    const fn flags(self) -> u32 {
        self.0
    }

    const fn from_flags(flags: u32) -> Self {
        Self(flags & (Self::FIRE.0 | Self::FEATHER_FALLING.0 | Self::BLAST.0 | Self::PROJECTILE.0))
    }
}

impl BitOr for AffectedProtections {
    type Output = Self;

    fn bitor(self, rhs: Self) -> Self::Output {
        Self(self.0 | rhs.0)
    }
}

impl BitOrAssign for AffectedProtections {
    fn bitor_assign(&mut self, rhs: Self) {
        self.0 |= rhs.0;
    }
}

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct Attack {
    attacker: Option<Entity>,
}

impl Attack {
    pub const fn new(attacker: Entity) -> Self {
        Self {
            attacker: Some(attacker),
        }
    }

    pub const fn without_attacker() -> Self {
        Self { attacker: None }
    }

    pub const fn attacker(self) -> Option<Entity> {
        self.attacker
    }
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct Block {
    block: block::Block,
}

impl Block {
    pub fn new(block: block::Block) -> Self {
        Self { block }
    }

    pub const fn block(&self) -> &block::Block {
        &self.block
    }

    pub fn into_block(self) -> block::Block {
        self.block
    }
}

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct Poison {
    fatal: bool,
}

impl Poison {
    pub const fn new(fatal: bool) -> Self {
        Self { fatal }
    }

    pub const fn fatal(self) -> bool {
        self.fatal
    }
}

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct Projectile {
    projectile: Option<Entity>,
    owner: Option<Entity>,
}

impl Projectile {
    pub const fn new(projectile: Entity, owner: Option<Entity>) -> Self {
        Self {
            projectile: Some(projectile),
            owner,
        }
    }

    pub const fn without_projectile(owner: Option<Entity>) -> Self {
        Self {
            projectile: None,
            owner,
        }
    }

    pub const fn projectile(self) -> Option<Entity> {
        self.projectile
    }

    pub const fn owner(self) -> Option<Entity> {
        self.owner
    }
}

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct Thorns {
    owner: Option<Entity>,
}

impl Thorns {
    pub const fn new(owner: Entity) -> Self {
        Self { owner: Some(owner) }
    }

    pub const fn without_owner() -> Self {
        Self { owner: None }
    }

    pub const fn owner(self) -> Option<Entity> {
        self.owner
    }
}

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct Custom<'a> {
    name: &'a str,
    traits: Traits,
    affected_protections: AffectedProtections,
}

impl<'a> Custom<'a> {
    pub const fn new(
        name: &'a str,
        traits: Traits,
        affected_protections: AffectedProtections,
    ) -> Self {
        Self {
            name,
            traits,
            affected_protections,
        }
    }

    pub const fn name(self) -> &'a str {
        self.name
    }

    pub const fn traits(self) -> Traits {
        self.traits
    }

    pub const fn affected_protections(self) -> AffectedProtections {
        self.affected_protections
    }
}

macro_rules! unit_sources {
    ($($name:ident),+ $(,)?) => {$ (
        #[derive(Clone, Copy, Debug, Default, Eq, Hash, PartialEq)]
        pub struct $name;
    )+ };
}

unit_sources! {
    Drowning,
    Explosion,
    Fall,
    Fire,
    Glide,
    Instant,
    Lava,
    Lightning,
    Magma,
    Starvation,
    Suffocation,
    Void,
    Wither,
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub enum Source<'a> {
    Attack(Attack),
    Block(Block),
    Drowning,
    Explosion,
    Fall,
    Fire,
    Glide,
    Instant,
    Lava,
    Lightning,
    Magma,
    Poison(Poison),
    Projectile(Projectile),
    Starvation,
    Suffocation,
    Thorns(Thorns),
    Void,
    Wither,
    Custom(Custom<'a>),
}

macro_rules! source_from_value {
    ($type:ident, $variant:ident) => {
        impl<'a> From<$type> for Source<'a> {
            fn from(value: $type) -> Self {
                Self::$variant(value)
            }
        }
    };
}

source_from_value!(Attack, Attack);
source_from_value!(Block, Block);
source_from_value!(Poison, Poison);
source_from_value!(Projectile, Projectile);
source_from_value!(Thorns, Thorns);

impl<'a> From<Custom<'a>> for Source<'a> {
    fn from(value: Custom<'a>) -> Self {
        Self::Custom(value)
    }
}

macro_rules! source_from_unit {
    ($($type:ident => $variant:ident),+ $(,)?) => {$ (
        impl<'a> From<$type> for Source<'a> {
            fn from(_: $type) -> Self {
                Self::$variant
            }
        }
    )+ };
}

source_from_unit! {
    Drowning => Drowning,
    Explosion => Explosion,
    Fall => Fall,
    Fire => Fire,
    Glide => Glide,
    Instant => Instant,
    Lava => Lava,
    Lightning => Lightning,
    Magma => Magma,
    Starvation => Starvation,
    Suffocation => Suffocation,
    Void => Void,
    Wither => Wither,
}

impl<'a> Source<'a> {
    pub const fn traits(&self) -> Traits {
        match self {
            Self::Attack(_) | Self::Block(_) | Self::Lightning => {
                Traits::new(true, true, false, false)
            }
            Self::Explosion | Self::Projectile(_) => Traits::new(true, true, false, false),
            Self::Fire | Self::Lava | Self::Magma => Traits::new(true, true, true, false),
            Self::Fall
            | Self::Glide
            | Self::Instant
            | Self::Poison(_)
            | Self::Thorns(_)
            | Self::Wither => Traits::new(false, true, false, false),
            Self::Drowning | Self::Starvation | Self::Suffocation => {
                Traits::new(false, false, false, false)
            }
            Self::Void => Traits::new(false, false, false, true),
            Self::Custom(custom) => custom.traits(),
        }
    }

    pub const fn affected_protections(&self) -> AffectedProtections {
        match self {
            Self::Explosion => AffectedProtections::BLAST,
            Self::Fall => AffectedProtections::FEATHER_FALLING,
            Self::Fire | Self::Magma => AffectedProtections::FIRE,
            Self::Projectile(_) => AffectedProtections::PROJECTILE,
            Self::Custom(custom) => custom.affected_protections(),
            _ => AffectedProtections::NONE,
        }
    }

    pub const fn reduced_by_armour(&self) -> bool {
        self.traits().reduced_by_armour()
    }

    pub const fn reduced_by_resistance(&self) -> bool {
        self.traits().reduced_by_resistance()
    }

    pub const fn is_fire(&self) -> bool {
        self.traits().is_fire()
    }

    pub const fn ignores_totem(&self) -> bool {
        self.traits().ignores_totem()
    }

    pub const fn affected_by(&self, enchantment: Enchantment) -> bool {
        match enchantment {
            Enchantment::Protection => self.reduced_by_resistance(),
            Enchantment::FireProtection => self
                .affected_protections()
                .contains(AffectedProtections::FIRE),
            Enchantment::FeatherFalling => self
                .affected_protections()
                .contains(AffectedProtections::FEATHER_FALLING),
            Enchantment::BlastProtection => self
                .affected_protections()
                .contains(AffectedProtections::BLAST),
            Enchantment::ProjectileProtection => self
                .affected_protections()
                .contains(AffectedProtections::PROJECTILE),
            _ => false,
        }
    }

    pub const fn name(&self) -> &str {
        match self {
            Self::Attack(_) => "attack",
            Self::Block(_) => "block",
            Self::Drowning => "drowning",
            Self::Explosion => "explosion",
            Self::Fall => "fall",
            Self::Fire => "fire",
            Self::Glide => "glide",
            Self::Instant => "instant",
            Self::Lava => "lava",
            Self::Lightning => "lightning",
            Self::Magma => "magma",
            Self::Poison(_) => "poison",
            Self::Projectile(_) => "projectile",
            Self::Starvation => "starvation",
            Self::Suffocation => "suffocation",
            Self::Thorns(_) => "thorns",
            Self::Void => "void",
            Self::Wither => "wither",
            Self::Custom(custom) => custom.name(),
        }
    }

    pub(crate) fn with_raw<R>(
        &self,
        callback: impl FnOnce(&dragonfly_plugin_sys::DfDamageSourceView) -> R,
    ) -> Option<R> {
        let block_owner = match self {
            Self::Block(source) => Some(BlockViewOwner::new(source.block())?),
            _ => None,
        };
        let raw_block = block_owner.as_ref().map(BlockViewOwner::view);
        let mut entity = dragonfly_plugin_sys::DfEntityId::default();
        let mut secondary_entity = dragonfly_plugin_sys::DfEntityId::default();
        let mut data = 0;
        let kind = match self {
            Self::Attack(source) => {
                if let Some(attacker) = source.attacker() {
                    entity = attacker.raw_id();
                }
                dragonfly_plugin_sys::DF_DAMAGE_SOURCE_ATTACK
            }
            Self::Block(_) => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_BLOCK,
            Self::Drowning => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_DROWNING,
            Self::Explosion => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_EXPLOSION,
            Self::Fall => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_FALL,
            Self::Fire => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_FIRE_KIND,
            Self::Glide => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_GLIDE,
            Self::Instant => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_INSTANT,
            Self::Lava => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_LAVA,
            Self::Lightning => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_LIGHTNING,
            Self::Magma => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_MAGMA,
            Self::Poison(source) => {
                data = u8::from(source.fatal());
                dragonfly_plugin_sys::DF_DAMAGE_SOURCE_POISON
            }
            Self::Projectile(source) => {
                if let Some(projectile) = source.projectile() {
                    entity = projectile.raw_id();
                }
                if let Some(owner) = source.owner() {
                    secondary_entity = owner.raw_id();
                }
                dragonfly_plugin_sys::DF_DAMAGE_SOURCE_PROJECTILE
            }
            Self::Starvation => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_STARVATION,
            Self::Suffocation => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_SUFFOCATION,
            Self::Thorns(source) => {
                if let Some(owner) = source.owner() {
                    entity = owner.raw_id();
                }
                dragonfly_plugin_sys::DF_DAMAGE_SOURCE_THORNS
            }
            Self::Void => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_VOID,
            Self::Wither => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_WITHER,
            Self::Custom(_) => dragonfly_plugin_sys::DF_DAMAGE_SOURCE_CUSTOM,
        };
        let view = dragonfly_plugin_sys::DfDamageSourceView {
            name: crate::string_view_from_str(self.name()),
            kind,
            flags: self.traits().flags() | self.affected_protections().flags(),
            entity,
            secondary_entity,
            block: raw_block
                .as_ref()
                .map_or(core::ptr::null(), core::ptr::from_ref),
            data,
        };
        Some(callback(&view))
    }

    pub(crate) unsafe fn from_raw(
        raw: &'a dragonfly_plugin_sys::DfDamageSourceView,
    ) -> Option<Self> {
        Some(match raw.kind {
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_ATTACK => {
                Self::Attack(if raw.entity.generation == 0 {
                    Attack::without_attacker()
                } else {
                    Attack::new(raw.entity.into())
                })
            }
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_BLOCK => {
                let raw_block = unsafe { raw.block.as_ref() }?;
                let identifier = unsafe { crate::string_view(raw_block.identifier) }.to_owned();
                let properties = unsafe { bytes(raw_block.properties_nbt) };
                Self::Block(Block::new(block::Block::from_nbt(identifier, properties)?))
            }
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_DROWNING => Self::Drowning,
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_EXPLOSION => Self::Explosion,
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_FALL => Self::Fall,
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_FIRE_KIND => Self::Fire,
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_GLIDE => Self::Glide,
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_INSTANT => Self::Instant,
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_LAVA => Self::Lava,
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_LIGHTNING => Self::Lightning,
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_MAGMA => Self::Magma,
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_POISON => {
                Self::Poison(Poison::new(raw.data != 0))
            }
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_PROJECTILE => {
                let owner = (raw.secondary_entity.generation != 0)
                    .then(|| Entity::from(raw.secondary_entity));
                Self::Projectile(if raw.entity.generation == 0 {
                    Projectile::without_projectile(owner)
                } else {
                    Projectile::new(raw.entity.into(), owner)
                })
            }
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_STARVATION => Self::Starvation,
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_SUFFOCATION => Self::Suffocation,
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_THORNS => {
                Self::Thorns(if raw.entity.generation == 0 {
                    Thorns::without_owner()
                } else {
                    Thorns::new(raw.entity.into())
                })
            }
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_VOID => Self::Void,
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_WITHER => Self::Wither,
            dragonfly_plugin_sys::DF_DAMAGE_SOURCE_CUSTOM => Self::Custom(Custom::new(
                unsafe { crate::string_view(raw.name) },
                Traits::from_flags(raw.flags),
                AffectedProtections::from_flags(raw.flags),
            )),
            _ => return None,
        })
    }
}

struct BlockViewOwner<'a> {
    block: &'a block::Block,
    properties: Vec<u8>,
}

impl<'a> BlockViewOwner<'a> {
    fn new(block: &'a block::Block) -> Option<Self> {
        Some(Self {
            block,
            properties: block.properties_nbt()?,
        })
    }

    fn view(&self) -> dragonfly_plugin_sys::DfBlockView {
        dragonfly_plugin_sys::DfBlockView {
            identifier: crate::string_view_from_str(self.block.identifier()),
            properties_nbt: crate::bytes_view(&self.properties),
        }
    }
}

unsafe fn bytes<'a>(view: dragonfly_plugin_sys::DfStringView) -> &'a [u8] {
    if view.len == 0 {
        return &[];
    }
    unsafe { core::slice::from_raw_parts(view.data, view.len as usize) }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn entity(generation: u64) -> Entity {
        dragonfly_plugin_sys::DfEntityId {
            bytes: [generation as u8; 16],
            generation,
        }
        .into()
    }

    #[test]
    fn concrete_sources_are_matchable_and_keep_handles() {
        let attacker = entity(7);
        let source: Source<'_> = Attack::new(attacker).into();
        assert!(matches!(source, Source::Attack(value) if value.attacker() == Some(attacker)));

        let projectile = entity(9);
        let source: Source<'_> = Projectile::new(projectile, Some(attacker)).into();
        assert!(
            matches!(source, Source::Projectile(value) if value.projectile() == Some(projectile) && value.owner() == Some(attacker))
        );
    }

    #[test]
    fn concrete_sources_preserve_absent_entities() {
        for source in [
            Source::from(Attack::without_attacker()),
            Source::from(Projectile::without_projectile(None)),
            Source::from(Thorns::without_owner()),
        ] {
            source.with_raw(|raw| {
                assert_eq!(raw.entity.generation, 0);
                assert_eq!(unsafe { Source::from_raw(raw) }.unwrap(), source);
            });
        }
    }

    #[test]
    fn traits_and_protection_rules_match_dragonfly() {
        let fall: Source<'_> = Fall.into();
        assert!(!fall.reduced_by_armour());
        assert!(fall.reduced_by_resistance());
        assert!(fall.affected_by(Enchantment::Protection));
        assert!(fall.affected_by(Enchantment::FeatherFalling));
        assert!(!fall.affected_by(Enchantment::FireProtection));

        let void: Source<'_> = Void.into();
        assert!(void.ignores_totem());
        assert!(!void.affected_by(Enchantment::Protection));
    }

    #[test]
    fn custom_round_trips_traits_and_flags() {
        let traits = Traits::new(true, false, true, true);
        let affected = AffectedProtections::FIRE | AffectedProtections::BLAST;
        let source: Source<'_> = Custom::new("example.Custom", traits, affected).into();
        source
            .with_raw(|raw| {
                let decoded = unsafe { Source::from_raw(raw) }.unwrap();
                assert_eq!(decoded.name(), "example.Custom");
                assert_eq!(decoded.traits(), traits);
                assert_eq!(decoded.affected_protections(), affected);
            })
            .unwrap();
    }

    #[test]
    fn block_view_keeps_property_bytes_alive() {
        let block = block::new("minecraft:oak_door").with_property("open_bit", true);
        let source: Source<'_> = Block::new(block.clone()).into();
        source
            .with_raw(|raw| {
                assert!(!raw.block.is_null());
                let decoded = unsafe { Source::from_raw(raw) }.unwrap();
                assert!(matches!(decoded, Source::Block(value) if value.block() == &block));
            })
            .unwrap();
    }
}
