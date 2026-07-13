use std::time::Duration;

use crate::{ItemStack, Potion, Rotation, Vec3, World, block};

const MAX_ENTITY_TYPE_BYTES: usize = 256;
const MAX_ENTITY_TAG_BYTES: usize = 4 << 10;

/// A stable identity for an entity managed by the host.
#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct EntityId {
    uuid: [u8; 16],
    generation: u64,
}

impl EntityId {
    pub const fn uuid_bytes(self) -> [u8; 16] {
        self.uuid
    }

    pub const fn generation(self) -> u64 {
        self.generation
    }

    pub(crate) const fn from_parts(uuid: [u8; 16], generation: u64) -> Self {
        Self { uuid, generation }
    }

    pub(crate) const fn raw(self) -> dragonfly_plugin_sys::DfEntityId {
        dragonfly_plugin_sys::DfEntityId {
            bytes: self.uuid,
            generation: self.generation,
        }
    }
}

impl From<dragonfly_plugin_sys::DfEntityId> for EntityId {
    fn from(value: dragonfly_plugin_sys::DfEntityId) -> Self {
        Self::from_parts(value.bytes, value.generation)
    }
}

/// A persistent handle to an entity.
///
/// The handle is safe to retain. Entity state is resolved by the host when an
/// operation is performed, so a handle may become stale after its entity is
/// removed.
#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct Entity {
    id: EntityId,
}

impl Entity {
    pub const fn id(self) -> EntityId {
        self.id
    }

    pub(crate) const fn from_id(id: EntityId) -> Self {
        Self { id }
    }

    pub(crate) const fn raw_id(self) -> dragonfly_plugin_sys::DfEntityId {
        self.id.raw()
    }

    pub fn state(&self) -> Option<EntityState> {
        let host = crate::host_api()?;
        let read = host.entity_state?;
        let mut entity_type = [0; MAX_ENTITY_TYPE_BYTES];
        let mut name_tag = vec![0; MAX_ENTITY_TAG_BYTES];
        let mut raw = dragonfly_plugin_sys::DfEntityState {
            position: dragonfly_plugin_sys::DfVec3::default(),
            rotation: dragonfly_plugin_sys::DfRotation::default(),
            velocity: dragonfly_plugin_sys::DfVec3::default(),
            capabilities: 0,
            world: dragonfly_plugin_sys::DfWorldId::default(),
            entity_type: crate::bytes_buffer(&mut entity_type),
            name_tag: crate::bytes_buffer(&mut name_tag),
        };
        let status = unsafe {
            read(
                host.context,
                crate::current_invocation(),
                self.raw_id(),
                &mut raw,
            )
        };
        if status != dragonfly_plugin_sys::DF_STATUS_OK {
            return None;
        }
        let type_length = usize::try_from(raw.entity_type.len).ok()?;
        let name_length = usize::try_from(raw.name_tag.len).ok()?;
        Some(EntityState {
            entity_type: String::from_utf8(entity_type.get(..type_length)?.to_vec()).ok()?,
            position: raw.position.into(),
            rotation: raw.rotation.into(),
            world: World::from_raw(raw.world.value),
            velocity: (raw.capabilities & dragonfly_plugin_sys::DF_ENTITY_HAS_VELOCITY != 0)
                .then(|| raw.velocity.into()),
            name_tag: if raw.capabilities & dragonfly_plugin_sys::DF_ENTITY_HAS_NAME_TAG != 0 {
                Some(String::from_utf8(name_tag.get(..name_length)?.to_vec()).ok()?)
            } else {
                None
            },
            can_teleport: raw.capabilities & dragonfly_plugin_sys::DF_ENTITY_CAN_TELEPORT != 0,
        })
    }

    pub fn entity_type(&self) -> Option<String> {
        self.state().map(|state| state.entity_type)
    }
    pub fn position(&self) -> Option<Vec3> {
        self.state().map(|state| state.position)
    }
    pub fn rotation(&self) -> Option<Rotation> {
        self.state().map(|state| state.rotation)
    }
    pub fn world(&self) -> Option<World> {
        self.state()?.world
    }
    pub fn velocity(&self) -> Option<Vec3> {
        self.state()?.velocity
    }
    pub fn name_tag(&self) -> Option<String> {
        self.state()?.name_tag
    }

    pub fn teleport(&self, position: Vec3) {
        let Some(host) = crate::host_api() else {
            return;
        };
        let Some(call) = host.entity_teleport else {
            return;
        };
        let _ = unsafe {
            call(
                host.context,
                crate::current_invocation(),
                self.raw_id(),
                position.into(),
            )
        };
    }

    pub fn set_velocity(&self, velocity: Vec3) {
        let Some(host) = crate::host_api() else {
            return;
        };
        let Some(call) = host.entity_velocity_set else {
            return;
        };
        let _ = unsafe {
            call(
                host.context,
                crate::current_invocation(),
                self.raw_id(),
                velocity.into(),
            )
        };
    }

    pub fn set_name_tag(&self, name_tag: &str) {
        let Some(host) = crate::host_api() else {
            return;
        };
        let Some(call) = host.entity_name_tag_set else {
            return;
        };
        let _ = unsafe {
            call(
                host.context,
                crate::current_invocation(),
                self.raw_id(),
                crate::string_view_from_str(name_tag),
            )
        };
    }

    pub fn despawn(self) {
        let Some(host) = crate::host_api() else {
            return;
        };
        let Some(call) = host.entity_despawn else {
            return;
        };
        let _ = unsafe { call(host.context, crate::current_invocation(), self.raw_id()) };
    }
}

impl From<dragonfly_plugin_sys::DfEntityId> for Entity {
    fn from(value: dragonfly_plugin_sys::DfEntityId) -> Self {
        Self::from_id(value.into())
    }
}

/// A snapshot of the common state exposed by an entity handle.
#[derive(Clone, Debug, Default, PartialEq)]
pub struct EntityState {
    pub entity_type: String,
    pub position: Vec3,
    pub rotation: Rotation,
    pub world: Option<World>,
    pub velocity: Option<Vec3>,
    pub name_tag: Option<String>,
    pub can_teleport: bool,
}

/// Common options applied when an entity is spawned.
#[derive(Clone, Debug, PartialEq)]
pub struct SpawnOptions {
    pub position: Vec3,
    pub rotation: Option<Rotation>,
    pub velocity: Option<Vec3>,
    pub name_tag: Option<String>,
}

impl SpawnOptions {
    pub const fn new(position: Vec3) -> Self {
        Self {
            position,
            rotation: None,
            velocity: None,
            name_tag: None,
        }
    }

    pub const fn rotation(mut self, rotation: Rotation) -> Self {
        self.rotation = Some(rotation);
        self
    }

    pub const fn velocity(mut self, velocity: Vec3) -> Self {
        self.velocity = Some(velocity);
        self
    }

    pub fn name_tag(mut self, name_tag: impl Into<String>) -> Self {
        self.name_tag = Some(name_tag.into());
        self
    }
}

/// A stationary text label.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct Text {
    text: String,
}

impl Text {
    pub fn new(text: impl Into<String>) -> Self {
        Self { text: text.into() }
    }

    pub fn text(&self) -> &str {
        &self.text
    }
}

/// A lightning bolt.
#[derive(Clone, Copy, Debug, PartialEq)]
pub struct Lightning {
    damage: f64,
    block_fire: bool,
    entity_fire_duration: Duration,
}

impl Lightning {
    pub const fn new() -> Self {
        Self {
            damage: 5.0,
            block_fire: true,
            entity_fire_duration: Duration::from_secs(8),
        }
    }

    pub const fn damage(mut self, damage: f64) -> Self {
        self.damage = damage;
        self
    }

    pub const fn block_fire(mut self, block_fire: bool) -> Self {
        self.block_fire = block_fire;
        self
    }

    pub const fn entity_fire_duration(mut self, duration: Duration) -> Self {
        self.entity_fire_duration = duration;
        self
    }

    pub const fn damage_value(self) -> f64 {
        self.damage
    }

    pub const fn creates_block_fire(self) -> bool {
        self.block_fire
    }

    pub const fn fire_duration(self) -> Duration {
        self.entity_fire_duration
    }
}

impl Default for Lightning {
    fn default() -> Self {
        Self::new()
    }
}

/// Primed TNT that explodes after its fuse expires.
#[allow(clippy::upper_case_acronyms)]
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub struct TNT {
    fuse: Duration,
}

impl TNT {
    pub const fn new(fuse: Duration) -> Self {
        Self { fuse }
    }

    pub const fn fuse(self) -> Duration {
        self.fuse
    }
}

/// A single experience orb.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub struct ExperienceOrb {
    experience: u32,
}

impl ExperienceOrb {
    pub const fn new(experience: u32) -> Self {
        Self { experience }
    }

    pub const fn experience(self) -> u32 {
        self.experience
    }
}

/// An item stack dropped into a world.
#[derive(Clone, Debug, PartialEq)]
pub struct DroppedItem {
    stack: ItemStack,
    pickup_delay: Option<Duration>,
}

impl DroppedItem {
    pub const fn new(stack: ItemStack) -> Self {
        Self {
            stack,
            pickup_delay: None,
        }
    }

    pub const fn pickup_delay(mut self, delay: Duration) -> Self {
        self.pickup_delay = Some(delay);
        self
    }

    pub const fn stack(&self) -> &ItemStack {
        &self.stack
    }

    pub const fn delay(&self) -> Option<Duration> {
        self.pickup_delay
    }
}

/// A block state falling as an entity.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct FallingBlock {
    block: block::Block,
}

impl FallingBlock {
    pub const fn new(block: block::Block) -> Self {
        Self { block }
    }

    pub const fn block(&self) -> &block::Block {
        &self.block
    }
}

#[derive(Clone, Copy, Debug, PartialEq)]
pub struct Arrow {
    owner: Entity,
    damage: f64,
    tip: Potion,
    critical: bool,
    disable_pickup: bool,
    obtain_on_pickup: bool,
    punch_level: u32,
    piercing_level: u32,
}

impl Arrow {
    pub const fn new(owner: Entity) -> Self {
        Self {
            owner,
            damage: 2.0,
            tip: Potion::Water,
            critical: false,
            disable_pickup: false,
            obtain_on_pickup: false,
            punch_level: 0,
            piercing_level: 0,
        }
    }

    pub const fn damage(mut self, damage: f64) -> Self {
        self.damage = damage;
        self
    }
    pub const fn tip(mut self, tip: Potion) -> Self {
        self.tip = tip;
        self
    }
    pub const fn critical(mut self, critical: bool) -> Self {
        self.critical = critical;
        self
    }
    pub const fn disable_pickup(mut self, disabled: bool) -> Self {
        self.disable_pickup = disabled;
        self
    }
    pub const fn obtain_on_pickup(mut self, obtain: bool) -> Self {
        self.obtain_on_pickup = obtain;
        self
    }
    pub const fn punch_level(mut self, level: u32) -> Self {
        self.punch_level = level;
        self
    }
    pub const fn piercing_level(mut self, level: u32) -> Self {
        self.piercing_level = level;
        self
    }
}

macro_rules! owner_projectile {
    ($name:ident) => {
        #[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
        pub struct $name {
            owner: Entity,
        }
        impl $name {
            pub const fn new(owner: Entity) -> Self {
                Self { owner }
            }
            pub const fn owner(self) -> Entity {
                self.owner
            }
        }
    };
}

owner_projectile!(Egg);
owner_projectile!(Snowball);
owner_projectile!(EnderPearl);
owner_projectile!(BottleOfEnchanting);

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct SplashPotion {
    owner: Entity,
    potion: Potion,
}

impl SplashPotion {
    pub const fn new(owner: Entity, potion: Potion) -> Self {
        Self { owner, potion }
    }
}

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct LingeringPotion {
    owner: Entity,
    potion: Potion,
}

impl LingeringPotion {
    pub const fn new(owner: Entity, potion: Potion) -> Self {
        Self { owner, potion }
    }
}

/// Static definition of a plugin-owned base entity type.
#[derive(Clone, Copy, Debug, PartialEq)]
pub struct Definition {
    identifier: &'static str,
    network_identifier: &'static str,
    bounds: [f64; 6],
}

impl Definition {
    #[doc(hidden)]
    pub const fn new(
        identifier: &'static str,
        network_identifier: &'static str,
        bounds: [f64; 6],
    ) -> Self {
        Self {
            identifier,
            network_identifier,
            bounds,
        }
    }

    pub const fn identifier(self) -> &'static str {
        self.identifier
    }

    pub const fn network_identifier(self) -> &'static str {
        self.network_identifier
    }

    /// Returns `[min_x, min_y, min_z, max_x, max_y, max_z]`.
    pub const fn bounds(self) -> [f64; 6] {
        self.bounds
    }
}

/// Implemented by unit structs annotated with `#[dragonfly::entity]`.
pub trait CustomType {
    const DEFINITION: Definition;
}

#[doc(hidden)]
#[repr(transparent)]
pub struct RegisteredType(dragonfly_plugin_sys::DfEntityTypeDescriptorV1);

impl RegisteredType {
    pub const fn new(definition: Definition) -> Self {
        let bounds = definition.bounds();
        Self(dragonfly_plugin_sys::DfEntityTypeDescriptorV1 {
            save_id: dragonfly_plugin_sys::DfStringView {
                data: definition.identifier().as_ptr(),
                len: definition.identifier().len() as u64,
            },
            network_id: dragonfly_plugin_sys::DfStringView {
                data: definition.network_identifier().as_ptr(),
                len: definition.network_identifier().len() as u64,
            },
            min: dragonfly_plugin_sys::DfVec3 {
                x: bounds[0],
                y: bounds[1],
                z: bounds[2],
            },
            max: dragonfly_plugin_sys::DfVec3 {
                x: bounds[3],
                y: bounds[4],
                z: bounds[5],
            },
        })
    }
}

// The descriptor only contains pointers to static UTF-8 strings.
unsafe impl Sync for RegisteredType {}

#[linkme::distributed_slice]
#[doc(hidden)]
pub static REGISTERED_TYPES: [RegisteredType] = [..];

mod private {
    pub trait Sealed {}
}

/// A typed descriptor that the host can turn into a Dragonfly entity.
///
/// This trait is sealed so every descriptor has a stable host encoding.
pub trait Spawnable: private::Sealed {
    #[doc(hidden)]
    fn encode(&self) -> EncodedSpawnable;
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub(crate) enum SpawnKind {
    Text,
    Lightning,
    Tnt,
    ExperienceOrb,
    DroppedItem,
    FallingBlock,
    Arrow,
    Egg,
    Snowball,
    EnderPearl,
    BottleOfEnchanting,
    SplashPotion,
    LingeringPotion,
    Custom,
}

#[derive(Clone, Debug, PartialEq)]
pub(crate) enum SpawnPayload {
    Text(String),
    Lightning {
        damage: f64,
        block_fire: bool,
        entity_fire_duration: Duration,
    },
    Tnt {
        fuse: Duration,
    },
    ExperienceOrb {
        experience: u32,
    },
    DroppedItem {
        stack: ItemStack,
        pickup_delay: Option<Duration>,
    },
    FallingBlock(block::Block),
    Arrow(Arrow),
    Egg(Entity),
    Snowball(Entity),
    EnderPearl(Entity),
    BottleOfEnchanting(Entity),
    SplashPotion {
        owner: Entity,
        potion: Potion,
    },
    LingeringPotion {
        owner: Entity,
        potion: Potion,
    },
    Custom(&'static str),
}

/// Internal, owned representation passed from a typed descriptor to the SDK's
/// host adapter.
#[doc(hidden)]
#[derive(Clone, Debug, PartialEq)]
pub struct EncodedSpawnable {
    pub(crate) kind: SpawnKind,
    pub(crate) payload: SpawnPayload,
}

impl EncodedSpawnable {
    pub(crate) fn with_raw<R>(
        &self,
        options: &SpawnOptions,
        callback: impl FnOnce(&dragonfly_plugin_sys::DfEntitySpawnViewV2) -> R,
    ) -> Option<R> {
        let mut callback = Some(callback);
        let mut view = dragonfly_plugin_sys::DfEntitySpawnViewV2 {
            kind: 0,
            flags: 0,
            options: dragonfly_plugin_sys::DfEntitySpawnOptions {
                position: options.position.into(),
                rotation: options.rotation.unwrap_or_default().into(),
                velocity: options.velocity.unwrap_or_default().into(),
                name_tag: crate::string_view_from_str(options.name_tag.as_deref().unwrap_or("")),
            },
            owner: dragonfly_plugin_sys::DfEntityId::default(),
            damage: 0.0,
            fuse_milliseconds: 0,
            experience: 0,
            potion: 0,
            punch_level: 0,
            piercing_level: 0,
            text: dragonfly_plugin_sys::DfStringView::default(),
            item: core::ptr::null(),
            block: core::ptr::null(),
            custom_type: dragonfly_plugin_sys::DfStringView::default(),
        };
        match &self.payload {
            SpawnPayload::Text(text) => {
                view.kind = dragonfly_plugin_sys::DF_ENTITY_TEXT;
                view.text = crate::string_view_from_str(text);
                Some(callback.take()?(&view))
            }
            SpawnPayload::Lightning {
                damage,
                block_fire,
                entity_fire_duration,
            } => {
                view.kind = dragonfly_plugin_sys::DF_ENTITY_LIGHTNING;
                view.damage = *damage;
                view.flags =
                    u32::from(*block_fire) * dragonfly_plugin_sys::DF_ENTITY_LIGHTNING_BLOCK_FIRE;
                view.fuse_milliseconds = entity_fire_duration.as_millis().try_into().ok()?;
                Some(callback.take()?(&view))
            }
            SpawnPayload::Tnt { fuse } => {
                view.kind = dragonfly_plugin_sys::DF_ENTITY_TNT;
                view.fuse_milliseconds = fuse.as_millis().try_into().ok()?;
                Some(callback.take()?(&view))
            }
            SpawnPayload::ExperienceOrb { experience } => {
                view.kind = dragonfly_plugin_sys::DF_ENTITY_EXPERIENCE_ORB;
                view.experience = (*experience).try_into().ok()?;
                Some(callback.take()?(&view))
            }
            SpawnPayload::DroppedItem {
                stack,
                pickup_delay,
            } => stack
                .with_raw(|item| {
                    view.kind = dragonfly_plugin_sys::DF_ENTITY_ITEM;
                    view.item = item;
                    if let Some(delay) = pickup_delay {
                        view.flags = dragonfly_plugin_sys::DF_ENTITY_ITEM_HAS_PICKUP_DELAY;
                        view.fuse_milliseconds = delay.as_millis().try_into().unwrap_or(u64::MAX);
                    }
                    callback.take().map(|callback| callback(&view))
                })
                .flatten(),
            SpawnPayload::FallingBlock(block) => {
                let properties = block.properties_nbt()?;
                let raw_block = dragonfly_plugin_sys::DfBlockView {
                    identifier: crate::string_view_from_str(block.identifier()),
                    properties_nbt: crate::bytes_view(&properties),
                };
                view.kind = dragonfly_plugin_sys::DF_ENTITY_FALLING_BLOCK;
                view.block = &raw_block;
                Some(callback.take()?(&view))
            }
            SpawnPayload::Arrow(arrow) => {
                view.kind = dragonfly_plugin_sys::DF_ENTITY_ARROW;
                view.owner = arrow.owner.raw_id();
                view.damage = arrow.damage;
                view.potion = u32::from(arrow.tip.id());
                view.punch_level = arrow.punch_level.try_into().ok()?;
                view.piercing_level = arrow.piercing_level.try_into().ok()?;
                if arrow.critical {
                    view.flags |= dragonfly_plugin_sys::DF_ENTITY_ARROW_CRITICAL;
                }
                if arrow.disable_pickup {
                    view.flags |= dragonfly_plugin_sys::DF_ENTITY_ARROW_DISABLE_PICKUP;
                }
                if arrow.obtain_on_pickup {
                    view.flags |= dragonfly_plugin_sys::DF_ENTITY_ARROW_OBTAIN_ON_PICKUP;
                }
                Some(callback.take()?(&view))
            }
            SpawnPayload::Egg(owner) => encode_owner_projectile(
                &mut view,
                dragonfly_plugin_sys::DF_ENTITY_EGG,
                *owner,
                callback.take()?,
            ),
            SpawnPayload::Snowball(owner) => encode_owner_projectile(
                &mut view,
                dragonfly_plugin_sys::DF_ENTITY_SNOWBALL,
                *owner,
                callback.take()?,
            ),
            SpawnPayload::EnderPearl(owner) => encode_owner_projectile(
                &mut view,
                dragonfly_plugin_sys::DF_ENTITY_ENDER_PEARL,
                *owner,
                callback.take()?,
            ),
            SpawnPayload::BottleOfEnchanting(owner) => encode_owner_projectile(
                &mut view,
                dragonfly_plugin_sys::DF_ENTITY_BOTTLE_OF_ENCHANTING,
                *owner,
                callback.take()?,
            ),
            SpawnPayload::SplashPotion { owner, potion } => {
                view.potion = u32::from(potion.id());
                encode_owner_projectile(
                    &mut view,
                    dragonfly_plugin_sys::DF_ENTITY_SPLASH_POTION,
                    *owner,
                    callback.take()?,
                )
            }
            SpawnPayload::LingeringPotion { owner, potion } => {
                view.potion = u32::from(potion.id());
                encode_owner_projectile(
                    &mut view,
                    dragonfly_plugin_sys::DF_ENTITY_LINGERING_POTION,
                    *owner,
                    callback.take()?,
                )
            }
            SpawnPayload::Custom(identifier) => {
                view.kind = dragonfly_plugin_sys::DF_ENTITY_CUSTOM;
                view.custom_type = crate::string_view_from_str(identifier);
                Some(callback.take()?(&view))
            }
        }
    }
}

impl<T: CustomType> private::Sealed for T {}

impl<T: CustomType> Spawnable for T {
    fn encode(&self) -> EncodedSpawnable {
        EncodedSpawnable {
            kind: SpawnKind::Custom,
            payload: SpawnPayload::Custom(T::DEFINITION.identifier()),
        }
    }
}

fn encode_owner_projectile<R>(
    view: &mut dragonfly_plugin_sys::DfEntitySpawnViewV2,
    kind: u32,
    owner: Entity,
    callback: impl FnOnce(&dragonfly_plugin_sys::DfEntitySpawnViewV2) -> R,
) -> Option<R> {
    view.kind = kind;
    view.owner = owner.raw_id();
    Some(callback(view))
}

impl From<Vec3> for dragonfly_plugin_sys::DfVec3 {
    fn from(value: Vec3) -> Self {
        Self {
            x: value.x,
            y: value.y,
            z: value.z,
        }
    }
}

impl From<Rotation> for dragonfly_plugin_sys::DfRotation {
    fn from(value: Rotation) -> Self {
        Self {
            yaw: value.yaw,
            pitch: value.pitch,
        }
    }
}

impl From<dragonfly_plugin_sys::DfRotation> for Rotation {
    fn from(value: dragonfly_plugin_sys::DfRotation) -> Self {
        Self {
            yaw: value.yaw,
            pitch: value.pitch,
        }
    }
}

macro_rules! impl_spawnable {
    ($type:ty, $kind:ident, $encode:expr) => {
        impl private::Sealed for $type {}

        impl Spawnable for $type {
            fn encode(&self) -> EncodedSpawnable {
                EncodedSpawnable {
                    kind: SpawnKind::$kind,
                    payload: ($encode)(self),
                }
            }
        }
    };
}

impl_spawnable!(Text, Text, |value: &Text| SpawnPayload::Text(
    value.text.clone()
));
impl_spawnable!(Lightning, Lightning, |value: &Lightning| {
    SpawnPayload::Lightning {
        damage: value.damage,
        block_fire: value.block_fire,
        entity_fire_duration: value.entity_fire_duration,
    }
});
impl_spawnable!(TNT, Tnt, |value: &TNT| SpawnPayload::Tnt {
    fuse: value.fuse,
});
impl_spawnable!(ExperienceOrb, ExperienceOrb, |value: &ExperienceOrb| {
    SpawnPayload::ExperienceOrb {
        experience: value.experience,
    }
});
impl_spawnable!(DroppedItem, DroppedItem, |value: &DroppedItem| {
    SpawnPayload::DroppedItem {
        stack: value.stack.clone(),
        pickup_delay: value.pickup_delay,
    }
});
impl_spawnable!(FallingBlock, FallingBlock, |value: &FallingBlock| {
    SpawnPayload::FallingBlock(value.block.clone())
});
impl_spawnable!(Arrow, Arrow, |value: &Arrow| SpawnPayload::Arrow(*value));
impl_spawnable!(Egg, Egg, |value: &Egg| SpawnPayload::Egg(value.owner));
impl_spawnable!(Snowball, Snowball, |value: &Snowball| {
    SpawnPayload::Snowball(value.owner)
});
impl_spawnable!(EnderPearl, EnderPearl, |value: &EnderPearl| {
    SpawnPayload::EnderPearl(value.owner)
});
impl_spawnable!(
    BottleOfEnchanting,
    BottleOfEnchanting,
    |value: &BottleOfEnchanting| SpawnPayload::BottleOfEnchanting(value.owner)
);
impl_spawnable!(SplashPotion, SplashPotion, |value: &SplashPotion| {
    SpawnPayload::SplashPotion {
        owner: value.owner,
        potion: value.potion,
    }
});
impl_spawnable!(
    LingeringPotion,
    LingeringPotion,
    |value: &LingeringPotion| SpawnPayload::LingeringPotion {
        owner: value.owner,
        potion: value.potion
    }
);

#[cfg(test)]
mod tests {
    use super::*;
    use crate::item;

    #[test]
    fn spawn_options_have_dragonfly_style_builders() {
        let options = SpawnOptions::new(Vec3 {
            x: 1.0,
            y: 64.0,
            z: -2.0,
        })
        .rotation(Rotation {
            yaw: 90.0,
            pitch: 10.0,
        })
        .velocity(Vec3 {
            x: 0.1,
            y: 0.2,
            z: 0.3,
        })
        .name_tag("Example");

        assert_eq!(options.position.y, 64.0);
        assert_eq!(options.rotation.map(|rotation| rotation.yaw), Some(90.0));
        assert_eq!(options.velocity.map(|velocity| velocity.y), Some(0.2));
        assert_eq!(options.name_tag.as_deref(), Some("Example"));
    }

    #[test]
    fn descriptors_encode_owned_payloads() {
        assert_eq!(
            Text::new("floating text").encode().payload,
            SpawnPayload::Text("floating text".to_owned())
        );
        assert_eq!(
            Lightning::new()
                .damage(8.5)
                .block_fire(false)
                .entity_fire_duration(Duration::from_secs(3))
                .encode()
                .payload,
            SpawnPayload::Lightning {
                damage: 8.5,
                block_fire: false,
                entity_fire_duration: Duration::from_secs(3),
            }
        );
        assert_eq!(
            TNT::new(Duration::from_secs(4)).encode().payload,
            SpawnPayload::Tnt {
                fuse: Duration::from_secs(4)
            }
        );
        assert_eq!(
            ExperienceOrb::new(37).encode().payload,
            SpawnPayload::ExperienceOrb { experience: 37 }
        );

        let stack = item::new(item::Diamond, 3);
        assert_eq!(
            DroppedItem::new(stack.clone())
                .pickup_delay(Duration::from_millis(500))
                .encode()
                .payload,
            SpawnPayload::DroppedItem {
                stack,
                pickup_delay: Some(Duration::from_millis(500)),
            }
        );

        let block = block::new("minecraft:sand").with_property("sand_type", "normal");
        assert_eq!(
            FallingBlock::new(block.clone()).encode().payload,
            SpawnPayload::FallingBlock(block)
        );
    }

    #[test]
    fn entity_state_keeps_optional_capabilities_explicit() {
        let state = EntityState {
            entity_type: "dragonfly:text".to_owned(),
            position: Vec3 {
                x: 2.0,
                y: 3.0,
                z: 4.0,
            },
            rotation: Rotation {
                yaw: 20.0,
                pitch: -5.0,
            },
            world: None,
            velocity: None,
            name_tag: Some("Label".to_owned()),
            can_teleport: false,
        };

        assert_eq!(state.name_tag.as_deref(), Some("Label"));
        assert!(state.velocity.is_none());
        assert!(!state.can_teleport);
    }

    #[test]
    fn descriptors_are_spawnable_without_adapter_types() {
        fn assert_spawnable(_: &impl Spawnable) {}

        assert_spawnable(&Text::new("text"));
        assert_spawnable(&Lightning::new());
        assert_spawnable(&TNT::new(Duration::from_secs(4)));
        assert_spawnable(&ExperienceOrb::new(7));
        assert_spawnable(&DroppedItem::new(item::new(item::Apple, 1)));
        assert_spawnable(&FallingBlock::new(block::new("minecraft:sand")));
        let owner = Entity::from_id(EntityId::from_parts([7; 16], 9));
        assert_spawnable(&Arrow::new(owner).tip(Potion::Poison).critical(true));
        assert_spawnable(&Egg::new(owner));
        assert_spawnable(&Snowball::new(owner));
        assert_spawnable(&EnderPearl::new(owner));
        assert_spawnable(&BottleOfEnchanting::new(owner));
        assert_spawnable(&SplashPotion::new(owner, Potion::Healing));
        assert_spawnable(&LingeringPotion::new(owner, Potion::Poison));
    }
}
