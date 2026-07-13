//! Safe Rust SDK for native Dragonfly plugins.

extern crate self as dragonfly_plugin;

use core::sync::atomic::{AtomicPtr, Ordering};
use std::cell::Cell;

pub mod block;
pub mod damage;
pub mod effect;
pub mod entity;
pub mod form;
pub mod healing;
mod item_nbt;
pub mod particle;
pub mod sound;
pub mod world;

pub use entity::{Entity, EntityId};
pub use world::{Dimension, World};

pub use dragonfly_plugin_macros::{Command, CommandEnum, entity, plugin};

#[doc(hidden)]
pub mod __private {
    pub use crate::entity::{REGISTERED_TYPES, RegisteredType};
    pub use core::ffi::c_void;
    pub use dragonfly_plugin_sys as sys;
    pub use linkme;
}

#[allow(non_snake_case)]
pub mod Event {
    pub use super::PlayerAttackEntityEventData as PlayerAttackEntity;
    pub use super::PlayerBlockBreakEventData as PlayerBlockBreak;
    pub use super::PlayerBlockPickEventData as PlayerBlockPick;
    pub use super::PlayerBlockPlaceEventData as PlayerBlockPlace;
    pub use super::PlayerChangeWorldEventData as PlayerChangeWorld;
    pub use super::PlayerChatEventData as PlayerChat;
    pub use super::PlayerDeathEventData as PlayerDeath;
    pub use super::PlayerExperienceGainEventData as PlayerExperienceGain;
    pub use super::PlayerFireExtinguishEventData as PlayerFireExtinguish;
    pub use super::PlayerFoodLossEventData as PlayerFoodLoss;
    pub use super::PlayerHealEventData as PlayerHeal;
    pub use super::PlayerHeldSlotChangeEventData as PlayerHeldSlotChange;
    pub use super::PlayerHurtEventData as PlayerHurt;
    pub use super::PlayerItemConsumeEventData as PlayerItemConsume;
    pub use super::PlayerItemDamageEventData as PlayerItemDamage;
    pub use super::PlayerItemDropEventData as PlayerItemDrop;
    pub use super::PlayerItemReleaseEventData as PlayerItemRelease;
    pub use super::PlayerItemUseEventData as PlayerItemUse;
    pub use super::PlayerItemUseOnBlockEventData as PlayerItemUseOnBlock;
    pub use super::PlayerItemUseOnEntityEventData as PlayerItemUseOnEntity;
    pub use super::PlayerJoinEventData as PlayerJoin;
    pub use super::PlayerJumpEventData as PlayerJump;
    pub use super::PlayerLecternPageTurnEventData as PlayerLecternPageTurn;
    pub use super::PlayerMoveEventData as PlayerMove;
    pub use super::PlayerPunchAirEventData as PlayerPunchAir;
    pub use super::PlayerQuitEventData as PlayerQuit;
    pub use super::PlayerRespawnEventData as PlayerRespawn;
    pub use super::PlayerSignEditEventData as PlayerSignEdit;
    pub use super::PlayerSkinChangeEventData as PlayerSkinChange;
    pub use super::PlayerSleepEventData as PlayerSleep;
    pub use super::PlayerStartBreakEventData as PlayerStartBreak;
    pub use super::PlayerTeleportEventData as PlayerTeleport;
    pub use super::PlayerToggleSneakEventData as PlayerToggleSneak;
    pub use super::PlayerToggleSprintEventData as PlayerToggleSprint;
    pub use crate::entity::Death as EntityDeath;
    pub use crate::entity::Heal as EntityHeal;
    pub use crate::entity::Hurt as EntityHurt;
}

static HOST_API: AtomicPtr<dragonfly_plugin_sys::DfHostApiV15> =
    AtomicPtr::new(core::ptr::null_mut());

std::thread_local! {
    static CURRENT_INVOCATION: Cell<dragonfly_plugin_sys::DfInvocationId> = const { Cell::new(0) };
}

#[doc(hidden)]
pub fn current_invocation() -> dragonfly_plugin_sys::DfInvocationId {
    CURRENT_INVOCATION.get()
}

#[doc(hidden)]
pub fn with_invocation<R>(
    invocation: dragonfly_plugin_sys::DfInvocationId,
    function: impl FnOnce() -> R,
) -> R {
    struct Restore(dragonfly_plugin_sys::DfInvocationId);
    impl Drop for Restore {
        fn drop(&mut self) {
            CURRENT_INVOCATION.set(self.0);
        }
    }

    let previous = CURRENT_INVOCATION.replace(invocation);
    let _restore = Restore(previous);
    function()
}

const MAX_SKIN_DATA_BYTES: u64 = 64 << 20;
const MAX_SKIN_ANIMATIONS: usize = 64;
pub const MAX_SCOREBOARD_LINES: usize = 15;

struct SkinSnapshot {
    context: u64,
    invocation: dragonfly_plugin_sys::DfInvocationId,
    id: u64,
    close: dragonfly_plugin_sys::DfHostPlayerSkinCloseFn,
}

impl Drop for SkinSnapshot {
    fn drop(&mut self) {
        unsafe { (self.close)(self.context, self.invocation, self.id) }
    }
}

#[doc(hidden)]
/// # Safety
/// `host` must remain valid while plugin callbacks may execute.
pub unsafe fn install_host(host: *const dragonfly_plugin_sys::DfHostApiV15) {
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

#[repr(i64)]
#[derive(Clone, Copy, Debug, Default, Eq, Hash, PartialEq)]
pub enum GameMode {
    #[default]
    Survival = 0,
    Creative = 1,
    Adventure = 2,
    Spectator = 3,
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

#[derive(Clone, Debug, PartialEq)]
pub enum Value {
    Byte(i8),
    Short(i16),
    Int(i32),
    Long(i64),
    Float(f32),
    Double(f64),
    String(String),
    ByteArray(Vec<u8>),
    IntArray(Vec<i32>),
    LongArray(Vec<i64>),
    List(Vec<Value>),
    Compound(std::collections::BTreeMap<String, Value>),
}

macro_rules! value_from {
    ($type:ty, $variant:ident) => {
        impl From<$type> for Value {
            fn from(value: $type) -> Self {
                Self::$variant(value)
            }
        }
    };
}

value_from!(i8, Byte);
value_from!(i16, Short);
value_from!(i32, Int);
value_from!(i64, Long);
value_from!(f32, Float);
value_from!(f64, Double);
value_from!(String, String);
value_from!(Vec<Value>, List);
value_from!(std::collections::BTreeMap<String, Value>, Compound);

impl From<&str> for Value {
    fn from(value: &str) -> Self {
        Self::String(value.to_owned())
    }
}

impl From<bool> for Value {
    fn from(value: bool) -> Self {
        Self::Byte(i8::from(value))
    }
}

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct AppliedEnchantment {
    enchantment: Enchantment,
    level: u32,
}

impl AppliedEnchantment {
    pub const fn new(enchantment: Enchantment, level: u32) -> Self {
        Self { enchantment, level }
    }

    pub const fn enchantment(self) -> Enchantment {
        self.enchantment
    }

    pub const fn level(self) -> u32 {
        self.level
    }
}

#[derive(Clone, Debug, PartialEq)]
pub struct ItemStack {
    identifier: String,
    metadata: i16,
    count: u32,
    damage: u32,
    unbreakable: bool,
    anvil_cost: i32,
    custom_name: String,
    lore: Vec<String>,
    nbt: std::collections::BTreeMap<String, Value>,
    values: std::collections::BTreeMap<String, Value>,
    enchantments: Vec<AppliedEnchantment>,
}

impl ItemStack {
    pub fn empty() -> Self {
        Self::new(item::Custom::new(""), 0)
    }

    pub(crate) fn new(item: impl Item, count: u32) -> Self {
        Self {
            identifier: item.identifier().to_owned(),
            metadata: item.metadata(),
            count,
            damage: 0,
            unbreakable: false,
            anvil_cost: 0,
            custom_name: String::new(),
            lore: Vec::new(),
            nbt: std::collections::BTreeMap::new(),
            values: std::collections::BTreeMap::new(),
            enchantments: Vec::new(),
        }
    }

    fn from_snapshot(raw: &dragonfly_plugin_sys::DfItemStackSnapshot) -> Option<Self> {
        let host = host_api()?;
        read_item_stack_snapshot(host, current_invocation(), raw.snapshot, raw.info)
    }

    pub fn potion(potion: Potion, count: u32) -> Self {
        Self::new(
            item::Custom::new("minecraft:potion").with_metadata(i16::from(potion.id())),
            count,
        )
    }

    pub fn splash_potion(potion: Potion, count: u32) -> Self {
        Self::new(
            item::Custom::new("minecraft:splash_potion").with_metadata(i16::from(potion.id())),
            count,
        )
    }

    pub fn lingering_potion(potion: Potion, count: u32) -> Self {
        Self::new(
            item::Custom::new("minecraft:lingering_potion").with_metadata(i16::from(potion.id())),
            count,
        )
    }

    pub fn identifier(&self) -> &str {
        &self.identifier
    }

    pub fn metadata(&self) -> i16 {
        self.metadata
    }

    pub fn count(&self) -> u32 {
        self.count
    }

    pub fn damage(&self) -> u32 {
        self.damage
    }

    pub fn unbreakable(&self) -> bool {
        self.unbreakable
    }

    pub fn anvil_cost(&self) -> i32 {
        self.anvil_cost
    }

    pub fn custom_name(&self) -> &str {
        &self.custom_name
    }

    pub fn lore(&self) -> &[String] {
        &self.lore
    }

    pub fn nbt(&self) -> &std::collections::BTreeMap<String, Value> {
        &self.nbt
    }

    pub fn value(&self, key: &str) -> Option<&Value> {
        self.values.get(key)
    }

    pub fn values(&self) -> &std::collections::BTreeMap<String, Value> {
        &self.values
    }

    pub fn enchantment(&self, enchantment: Enchantment) -> Option<AppliedEnchantment> {
        self.enchantments
            .iter()
            .copied()
            .find(|applied| applied.enchantment == enchantment)
    }

    pub fn enchantments(&self) -> &[AppliedEnchantment] {
        &self.enchantments
    }

    pub fn with_count(mut self, count: u32) -> Self {
        self.count = count;
        self
    }

    pub fn with_damage(mut self, damage: u32) -> Self {
        self.damage = damage;
        self
    }

    pub fn as_unbreakable(mut self) -> Self {
        self.unbreakable = true;
        self
    }

    pub fn as_breakable(mut self) -> Self {
        self.unbreakable = false;
        self
    }

    pub fn with_anvil_cost(mut self, anvil_cost: i32) -> Self {
        self.anvil_cost = anvil_cost;
        self
    }

    pub fn with_custom_name(mut self, name: impl Into<String>) -> Self {
        self.custom_name = name.into();
        self
    }

    pub fn with_lore<I, S>(mut self, lines: I) -> Self
    where
        I: IntoIterator<Item = S>,
        S: Into<String>,
    {
        self.lore = lines.into_iter().map(Into::into).collect();
        self
    }

    pub fn with_value(mut self, key: impl Into<String>, value: impl Into<Value>) -> Self {
        self.values.insert(key.into(), value.into());
        self
    }

    pub fn without_value(mut self, key: &str) -> Self {
        self.values.remove(key);
        self
    }

    pub fn with_nbt(mut self, key: impl Into<String>, value: impl Into<Value>) -> Self {
        self.nbt.insert(key.into(), value.into());
        self
    }

    pub fn without_nbt(mut self, key: &str) -> Self {
        self.nbt.remove(key);
        self
    }

    pub fn with_enchantment(mut self, enchantment: Enchantment, level: u32) -> Self {
        self.enchantments
            .retain(|applied| applied.enchantment != enchantment);
        self.enchantments
            .push(AppliedEnchantment::new(enchantment, level));
        self.enchantments
            .sort_unstable_by_key(|applied| applied.enchantment.id());
        self
    }

    pub fn with_enchantments<I>(mut self, enchantments: I) -> Self
    where
        I: IntoIterator<Item = AppliedEnchantment>,
    {
        for applied in enchantments {
            self = self.with_enchantment(applied.enchantment, applied.level);
        }
        self
    }

    pub fn without_enchantment(mut self, enchantment: Enchantment) -> Self {
        self.enchantments
            .retain(|applied| applied.enchantment != enchantment);
        self
    }

    pub(crate) fn with_raw<R>(
        &self,
        function: impl FnOnce(&dragonfly_plugin_sys::DfItemStackViewV3) -> R,
    ) -> Option<R> {
        let nbt = if self.nbt.is_empty() {
            Vec::new()
        } else {
            item_nbt::encode_values(&self.nbt).ok()?
        };
        let values_nbt = if self.values.is_empty() {
            Vec::new()
        } else {
            item_nbt::encode_values(&self.values).ok()?
        };
        let lore: Vec<_> = self
            .lore
            .iter()
            .map(|line| string_view_from_str(line))
            .collect();
        let enchantments: Vec<_> = self
            .enchantments
            .iter()
            .map(|enchantment| dragonfly_plugin_sys::DfItemEnchantment {
                id: enchantment.enchantment.id(),
                level: enchantment.level,
            })
            .collect();
        let view = dragonfly_plugin_sys::DfItemStackViewV3 {
            identifier: string_view_from_str(&self.identifier),
            metadata: i32::from(self.metadata),
            count: self.count,
            damage: self.damage,
            unbreakable: u8::from(self.unbreakable),
            anvil_cost: self.anvil_cost,
            custom_name: string_view_from_str(&self.custom_name),
            lore: slice_pointer(&lore),
            lore_count: lore.len() as u64,
            nbt: bytes_view(&nbt),
            values_nbt: bytes_view(&values_nbt),
            enchantments: slice_pointer(&enchantments),
            enchantment_count: enchantments.len() as u64,
        };
        Some(function(&view))
    }
}

impl Default for ItemStack {
    fn default() -> Self {
        Self::empty()
    }
}

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct Inventory {
    player: Player,
    kind: u32,
}

impl Inventory {
    fn raw_id(self) -> dragonfly_plugin_sys::DfInventoryId {
        dragonfly_plugin_sys::DfInventoryId {
            player: self.player.raw_id(),
            kind: self.kind,
            reserved: 0,
        }
    }

    pub fn len(&self) -> usize {
        let Some(host) = host_api() else { return 0 };
        let Some(size) = host.inventory_size else {
            return 0;
        };
        let mut value = 0;
        if unsafe {
            size(
                host.context,
                current_invocation(),
                self.raw_id(),
                &mut value,
            )
        } != dragonfly_plugin_sys::DF_STATUS_OK
        {
            return 0;
        }
        value as usize
    }

    pub fn is_empty(&self) -> bool {
        self.len() == 0
    }

    pub fn item(&self, slot: usize) -> Option<ItemStack> {
        let slot = u32::try_from(slot).ok()?;
        read_item_stack(|host, snapshot, info| {
            let open = host.inventory_item_open?;
            Some(unsafe {
                open(
                    host.context,
                    current_invocation(),
                    self.raw_id(),
                    slot,
                    snapshot,
                    info,
                )
            })
        })
    }

    pub fn set_item(&self, slot: usize, item: &ItemStack) {
        let Ok(slot) = u32::try_from(slot) else {
            return;
        };
        let Some(host) = host_api() else { return };
        let Some(set) = host.inventory_item_set else {
            return;
        };
        let _ = item.with_raw(|item| unsafe {
            set(
                host.context,
                current_invocation(),
                self.raw_id(),
                slot,
                item,
            )
        });
    }

    pub fn add_item(&self, item: &ItemStack) -> u32 {
        let Some(host) = host_api() else { return 0 };
        let Some(add) = host.inventory_item_add else {
            return 0;
        };
        item.with_raw(|item| {
            let mut added = 0;
            let status = unsafe {
                add(
                    host.context,
                    current_invocation(),
                    self.raw_id(),
                    item,
                    &mut added,
                )
            };
            if status == dragonfly_plugin_sys::DF_STATUS_OK {
                added
            } else {
                0
            }
        })
        .unwrap_or(0)
    }

    pub fn clear_slot(&self, slot: usize) {
        let Ok(slot) = u32::try_from(slot) else {
            return;
        };
        let Some(host) = host_api() else { return };
        let Some(clear) = host.inventory_clear_slot else {
            return;
        };
        let _ = unsafe { clear(host.context, current_invocation(), self.raw_id(), slot) };
    }

    pub fn clear(&self) {
        let Some(host) = host_api() else { return };
        let Some(clear) = host.inventory_clear else {
            return;
        };
        let _ = unsafe { clear(host.context, current_invocation(), self.raw_id()) };
    }
}

struct ItemSnapshot {
    context: u64,
    invocation: dragonfly_plugin_sys::DfInvocationId,
    id: u64,
    close: dragonfly_plugin_sys::DfHostItemStackCloseFn,
}

impl Drop for ItemSnapshot {
    fn drop(&mut self) {
        unsafe { (self.close)(self.context, self.invocation, self.id) }
    }
}

impl Player {
    /// Sends a form and owns its one-shot callback until it is submitted, closed, or discarded.
    pub fn send_form<F, C>(&self, form: F, callback: C)
    where
        F: form::Form,
        C: FnOnce(Player, Option<F::Response>) + Send + 'static,
    {
        form::send(self, form, callback);
    }

    /// Requests that the client's currently visible form be closed.
    pub fn close_form(&self) {
        let Some(host) = host_api() else { return };
        let Some(close) = host.player_form_close else {
            return;
        };
        let _ = unsafe { close(host.context, current_invocation(), self.raw_id()) };
    }

    pub const fn inventory(&self) -> Inventory {
        Inventory {
            player: *self,
            kind: dragonfly_plugin_sys::DF_INVENTORY_MAIN,
        }
    }

    pub const fn armour(&self) -> Inventory {
        Inventory {
            player: *self,
            kind: dragonfly_plugin_sys::DF_INVENTORY_ARMOUR,
        }
    }

    pub const fn offhand(&self) -> Inventory {
        Inventory {
            player: *self,
            kind: dragonfly_plugin_sys::DF_INVENTORY_OFFHAND,
        }
    }

    pub fn held_items(&self) -> (ItemStack, ItemStack) {
        (
            self.held_item(0).unwrap_or_default(),
            self.held_item(1).unwrap_or_default(),
        )
    }

    pub fn set_held_items(&self, main_hand: &ItemStack, off_hand: &ItemStack) {
        let Some(host) = host_api() else { return };
        let Some(set) = host.player_held_items_set else {
            return;
        };
        let _ = main_hand.with_raw(|main_hand| {
            off_hand.with_raw(|off_hand| unsafe {
                set(
                    host.context,
                    current_invocation(),
                    self.raw_id(),
                    main_hand,
                    off_hand,
                )
            })
        });
    }

    pub fn set_held_slot(&self, slot: u32) {
        let Some(host) = host_api() else { return };
        let Some(set) = host.player_held_slot_set else {
            return;
        };
        let _ = unsafe { set(host.context, current_invocation(), self.raw_id(), slot) };
    }

    fn held_item(&self, hand: u32) -> Option<ItemStack> {
        read_item_stack(|host, snapshot, info| {
            let open = host.player_held_item_open?;
            Some(unsafe {
                open(
                    host.context,
                    current_invocation(),
                    self.raw_id(),
                    hand,
                    snapshot,
                    info,
                )
            })
        })
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

impl From<Player> for Entity {
    fn from(player: Player) -> Self {
        player.entity()
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

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct Scoreboard {
    name: String,
    lines: Vec<String>,
    padding: bool,
    descending: bool,
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub struct ScoreboardLineOutOfBounds {
    pub index: usize,
}

impl core::fmt::Display for ScoreboardLineOutOfBounds {
    fn fmt(&self, formatter: &mut core::fmt::Formatter<'_>) -> core::fmt::Result {
        write!(
            formatter,
            "scoreboard line index {} exceeds the maximum of {} lines",
            self.index, MAX_SCOREBOARD_LINES
        )
    }
}

impl std::error::Error for ScoreboardLineOutOfBounds {}

#[derive(Clone, Debug, Default, Eq, PartialEq)]
pub struct Skin {
    pub width: u32,
    pub height: u32,
    pub persona: bool,
    pub play_fab_id: String,
    pub full_id: String,
    pub pixels: Vec<u8>,
    pub model_config: SkinModelConfig,
    pub model: Vec<u8>,
    pub cape: Cape,
    pub animations: Vec<SkinAnimation>,
}

#[derive(Clone, Debug, Default, Eq, PartialEq)]
pub struct SkinModelConfig {
    pub default: String,
    pub animated_face: String,
}

#[derive(Clone, Debug, Default, Eq, PartialEq)]
pub struct Cape {
    pub width: u32,
    pub height: u32,
    pub pixels: Vec<u8>,
}

#[repr(u32)]
#[derive(Clone, Copy, Debug, Default, Eq, Hash, PartialEq)]
pub enum SkinAnimationType {
    #[default]
    Head = 0,
    Body32x32 = 1,
    Body128x128 = 2,
}

impl SkinAnimationType {
    fn from_raw(value: u32) -> Option<Self> {
        match value {
            0 => Some(Self::Head),
            1 => Some(Self::Body32x32),
            2 => Some(Self::Body128x128),
            _ => None,
        }
    }
}

#[derive(Clone, Debug, Default, Eq, PartialEq)]
pub struct SkinAnimation {
    pub width: u32,
    pub height: u32,
    pub animation_type: SkinAnimationType,
    pub pixels: Vec<u8>,
    pub frame_count: i64,
    pub expression: i64,
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

impl Scoreboard {
    pub fn new(name: impl Into<String>) -> Self {
        Self {
            name: name.into(),
            lines: Vec::new(),
            padding: true,
            descending: false,
        }
    }

    pub fn name(&self) -> &str {
        &self.name
    }

    pub fn lines(&self) -> &[String] {
        &self.lines
    }

    pub const fn padding(&self) -> bool {
        self.padding
    }

    pub const fn is_descending(&self) -> bool {
        self.descending
    }

    pub fn push_line(
        &mut self,
        line: impl Into<String>,
    ) -> Result<&mut Self, ScoreboardLineOutOfBounds> {
        let index = self.lines.len();
        if index >= MAX_SCOREBOARD_LINES {
            return Err(ScoreboardLineOutOfBounds { index });
        }
        self.lines.push(normalize_scoreboard_line(line.into()));
        Ok(self)
    }

    pub fn set_line(
        &mut self,
        index: usize,
        line: impl Into<String>,
    ) -> Result<&mut Self, ScoreboardLineOutOfBounds> {
        if index >= MAX_SCOREBOARD_LINES {
            return Err(ScoreboardLineOutOfBounds { index });
        }
        if self.lines.len() <= index {
            self.lines.resize(index + 1, String::new());
        }
        self.lines[index] = normalize_scoreboard_line(line.into());
        Ok(self)
    }

    pub fn remove_line(&mut self, index: usize) -> Option<String> {
        (index < self.lines.len()).then(|| self.lines.remove(index))
    }

    pub fn clear(&mut self) {
        self.lines.clear();
    }

    pub fn remove_padding(&mut self) -> &mut Self {
        self.padding = false;
        self
    }

    pub fn set_descending(&mut self, descending: bool) -> &mut Self {
        self.descending = descending;
        self
    }
}

fn normalize_scoreboard_line(mut line: String) -> String {
    for _ in 0..2 {
        if line.ends_with('\n') {
            line.pop();
        }
    }
    line
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

    pub const fn entity(self) -> Entity {
        Entity::from_id(EntityId::from_parts(self.id.uuid, self.id.generation))
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

    /// Formats and sends a message without making the caller allocate a `String`.
    pub fn messagef(&self, arguments: core::fmt::Arguments<'_>) {
        if let Some(message) = arguments.as_str() {
            self.message(message);
        } else {
            self.message(&std::fmt::format(arguments));
        }
    }

    pub fn heal<'a>(&self, health: f64, source: impl Into<healing::Source<'a>>) -> f64 {
        let Some(host) = host_api() else { return 0.0 };
        let Some(heal) = host.player_heal else {
            return 0.0;
        };
        let source = source.into();
        source.with_raw(|view| {
            let mut result = dragonfly_plugin_sys::DfPlayerHealResult::default();
            let status = unsafe {
                heal(
                    host.context,
                    current_invocation(),
                    self.raw_id(),
                    health,
                    view,
                    &mut result,
                )
            };
            if status == dragonfly_plugin_sys::DF_STATUS_OK {
                result.healed
            } else {
                0.0
            }
        })
    }

    pub fn hurt<'a>(&self, damage: f64, source: impl Into<damage::Source<'a>>) -> (f64, bool) {
        let Some(host) = host_api() else {
            return (0.0, false);
        };
        let Some(hurt) = host.player_hurt else {
            return (0.0, false);
        };
        let source = source.into();
        source
            .with_raw(|view| {
                let mut result = dragonfly_plugin_sys::DfPlayerHurtResult::default();
                let status = unsafe {
                    hurt(
                        host.context,
                        current_invocation(),
                        self.raw_id(),
                        damage,
                        view,
                        &mut result,
                    )
                };
                if status == dragonfly_plugin_sys::DF_STATUS_OK {
                    (result.damage, result.vulnerable != 0)
                } else {
                    (0.0, false)
                }
            })
            .unwrap_or((0.0, false))
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
        let _ = unsafe { send(host.context, current_invocation(), self.raw_id(), view) };
    }

    pub fn play_sound(&self, value: impl sound::Sound) {
        let Some(host) = host_api() else { return };
        let Some(play) = host.player_sound_play else {
            return;
        };
        let encoded = value.encode();
        let _ = encoded.with_raw(|view| unsafe {
            play(host.context, current_invocation(), self.raw_id(), view)
        });
    }

    pub fn send_scoreboard(&self, scoreboard: &Scoreboard) {
        let Some(host) = host_api() else { return };
        let Some(send) = host.player_scoreboard else {
            return;
        };
        let lines: Vec<_> = scoreboard
            .lines
            .iter()
            .map(|line| string_view_from_str(line))
            .collect();
        let view = dragonfly_plugin_sys::DfScoreboardView {
            name: string_view_from_str(&scoreboard.name),
            lines: lines.as_ptr(),
            line_count: lines.len() as u64,
            padding: u8::from(scoreboard.padding),
            descending: u8::from(scoreboard.descending),
        };
        let _ = unsafe { send(host.context, current_invocation(), self.raw_id(), view) };
    }

    pub fn remove_scoreboard(&self) {
        let Some(host) = host_api() else { return };
        let Some(remove) = host.player_scoreboard_remove else {
            return;
        };
        let _ = unsafe { remove(host.context, current_invocation(), self.raw_id()) };
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
        let _ = unsafe { read(host.context, current_invocation(), self.raw_id(), &mut raw) };
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

    pub fn hide_entity(&self, entity: Entity) {
        self.set_entity_visible(entity, false);
    }

    pub fn show_entity(&self, entity: Entity) {
        self.set_entity_visible(entity, true);
    }

    pub fn skin(&self) -> Skin {
        self.read_skin().unwrap_or_default()
    }

    pub fn set_skin(&self, skin: &Skin) {
        let host = HOST_API.load(Ordering::Acquire);
        let Some(host) = (unsafe { host.as_ref() }) else {
            return;
        };
        let Some(set) = host.player_skin_set else {
            return;
        };
        with_skin_view(skin, |view| {
            let _ = unsafe { set(host.context, current_invocation(), self.raw_id(), view) };
        });
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
                current_invocation(),
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
        let _ = unsafe {
            transform(
                host.context,
                current_invocation(),
                self.raw_id(),
                kind,
                raw,
                yaw,
                pitch,
            )
        };
    }

    fn set_state(&self, kind: u32, number: f64, integer: i64) {
        let host = HOST_API.load(Ordering::Acquire);
        let Some(host) = (unsafe { host.as_ref() }) else {
            return;
        };
        let Some(set) = host.player_state_set else {
            return;
        };
        let value = dragonfly_plugin_sys::DfPlayerStateValue { number, integer };
        let _ = unsafe {
            set(
                host.context,
                current_invocation(),
                self.raw_id(),
                kind,
                value,
            )
        };
    }

    fn set_entity_visible(&self, entity: Entity, visible: bool) {
        let host = HOST_API.load(Ordering::Acquire);
        let Some(host) = (unsafe { host.as_ref() }) else {
            return;
        };
        let Some(change) = host.player_entity_visibility else {
            return;
        };
        let _ = unsafe {
            change(
                host.context,
                current_invocation(),
                self.raw_id(),
                entity.raw_id(),
                u8::from(visible),
            )
        };
    }

    fn read_skin(&self) -> Option<Skin> {
        let host = HOST_API.load(Ordering::Acquire);
        let host = unsafe { host.as_ref() }?;
        let open = host.player_skin_open?;
        let close = host.player_skin_close?;

        let mut snapshot_id = 0;
        let mut info = dragonfly_plugin_sys::DfSkinInfo::default();
        let invocation = current_invocation();
        let status = unsafe {
            open(
                host.context,
                invocation,
                self.raw_id(),
                &mut snapshot_id,
                &mut info,
            )
        };
        if status != dragonfly_plugin_sys::DF_STATUS_OK {
            return None;
        }
        let _snapshot = SkinSnapshot {
            context: host.context,
            invocation,
            id: snapshot_id,
            close,
        };
        read_skin_snapshot(host, invocation, snapshot_id, info)
    }

    fn state(&self, kind: u32) -> dragonfly_plugin_sys::DfPlayerStateValue {
        let host = HOST_API.load(Ordering::Acquire);
        let Some(host) = (unsafe { host.as_ref() }) else {
            return dragonfly_plugin_sys::DfPlayerStateValue::default();
        };
        let Some(get) = host.player_state_get else {
            return dragonfly_plugin_sys::DfPlayerStateValue::default();
        };
        let mut value = dragonfly_plugin_sys::DfPlayerStateValue::default();
        let _ = unsafe {
            get(
                host.context,
                current_invocation(),
                self.raw_id(),
                kind,
                &mut value,
            )
        };
        value
    }

    fn raw_id(&self) -> dragonfly_plugin_sys::DfPlayerId {
        dragonfly_plugin_sys::DfPlayerId {
            bytes: self.id.uuid,
            generation: self.id.generation,
        }
    }

    pub(crate) fn from_id(id: dragonfly_plugin_sys::DfPlayerId) -> Self {
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
    pub fn __from_command_argument(value: &str) -> Result<Self, CommandParseError> {
        let mut parts = value.split(':');
        let uuid = parts
            .next()
            .ok_or_else(|| CommandParseError::__new("player is no longer online"))?;
        let generation = parts
            .next()
            .ok_or_else(|| CommandParseError::__new("player is no longer online"))?;
        let latency_milliseconds = parts.next().and_then(|value| value.parse().ok());
        let name = parts.next();
        if uuid.len() != 32 {
            return Err(CommandParseError::__new("player is no longer online"));
        }
        let mut bytes = [0; 16];
        for (index, byte) in bytes.iter_mut().enumerate() {
            *byte = u8::from_str_radix(&uuid[index * 2..index * 2 + 2], 16)
                .map_err(|_| CommandParseError::__new("player is no longer online"))?;
        }
        let generation = generation
            .parse()
            .map_err(|_| CommandParseError::__new("player is no longer online"))?;
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

fn with_skin_view<R>(
    skin: &Skin,
    function: impl FnOnce(&dragonfly_plugin_sys::DfSkinView) -> R,
) -> R {
    let animations: Vec<_> = skin
        .animations
        .iter()
        .map(|animation| dragonfly_plugin_sys::DfSkinAnimationView {
            width: animation.width,
            height: animation.height,
            animation_type: animation.animation_type as u32,
            frame_count: animation.frame_count,
            expression: animation.expression,
            pixels: bytes_view(&animation.pixels),
        })
        .collect();
    let view = dragonfly_plugin_sys::DfSkinView {
        width: skin.width,
        height: skin.height,
        persona: u8::from(skin.persona),
        play_fab_id: string_view_from_str(&skin.play_fab_id),
        full_id: string_view_from_str(&skin.full_id),
        pixels: bytes_view(&skin.pixels),
        model_default: string_view_from_str(&skin.model_config.default),
        model_animated_face: string_view_from_str(&skin.model_config.animated_face),
        model: bytes_view(&skin.model),
        cape_width: skin.cape.width,
        cape_height: skin.cape.height,
        cape_pixels: bytes_view(&skin.cape.pixels),
        animations: if animations.is_empty() {
            core::ptr::null()
        } else {
            animations.as_ptr()
        },
        animation_count: animations.len() as u64,
    };
    function(&view)
}

fn read_skin_snapshot(
    host: &dragonfly_plugin_sys::DfHostApiV15,
    invocation: dragonfly_plugin_sys::DfInvocationId,
    snapshot: u64,
    info: dragonfly_plugin_sys::DfSkinInfo,
) -> Option<Skin> {
    let animation_info_fn = host.player_skin_animation_info?;
    let read = host.player_skin_read?;
    let animation_count = usize::try_from(info.animation_count).ok()?;
    if animation_count > MAX_SKIN_ANIMATIONS {
        return None;
    }
    let mut animation_info = Vec::new();
    animation_info.try_reserve_exact(animation_count).ok()?;
    for index in 0..animation_count {
        let mut animation = dragonfly_plugin_sys::DfSkinAnimationInfo::default();
        let status = unsafe {
            animation_info_fn(
                host.context,
                invocation,
                snapshot,
                index as u64,
                &mut animation,
            )
        };
        if status != dragonfly_plugin_sys::DF_STATUS_OK {
            return None;
        }
        animation_info.push(animation);
    }

    let mut total_bytes = 0u64;
    for length in [
        info.play_fab_id_len,
        info.full_id_len,
        info.pixels_len,
        info.model_default_len,
        info.model_animated_face_len,
        info.model_len,
        info.cape_pixels_len,
    ]
    .into_iter()
    .chain(animation_info.iter().map(|animation| animation.pixels_len))
    {
        total_bytes = total_bytes.checked_add(length)?;
        if total_bytes > MAX_SKIN_DATA_BYTES {
            return None;
        }
    }

    let mut play_fab_id = allocate_bytes(info.play_fab_id_len)?;
    let mut full_id = allocate_bytes(info.full_id_len)?;
    let mut pixels = allocate_bytes(info.pixels_len)?;
    let mut model_default = allocate_bytes(info.model_default_len)?;
    let mut model_animated_face = allocate_bytes(info.model_animated_face_len)?;
    let mut model = allocate_bytes(info.model_len)?;
    let mut cape_pixels = allocate_bytes(info.cape_pixels_len)?;
    let mut animation_pixels = animation_info
        .iter()
        .map(|animation| allocate_bytes(animation.pixels_len))
        .collect::<Option<Vec<_>>>()?;
    let mut animation_buffers: Vec<_> = animation_pixels
        .iter_mut()
        .map(|pixels| bytes_buffer(pixels))
        .collect();
    let mut data = dragonfly_plugin_sys::DfSkinData {
        play_fab_id: bytes_buffer(&mut play_fab_id),
        full_id: bytes_buffer(&mut full_id),
        pixels: bytes_buffer(&mut pixels),
        model_default: bytes_buffer(&mut model_default),
        model_animated_face: bytes_buffer(&mut model_animated_face),
        model: bytes_buffer(&mut model),
        cape_pixels: bytes_buffer(&mut cape_pixels),
        animation_pixels: if animation_buffers.is_empty() {
            core::ptr::null_mut()
        } else {
            animation_buffers.as_mut_ptr()
        },
        animation_capacity: animation_buffers.len() as u64,
    };
    let status = unsafe { read(host.context, invocation, snapshot, &mut data) };
    if status != dragonfly_plugin_sys::DF_STATUS_OK {
        return None;
    }

    finish_buffer(&mut play_fab_id, data.play_fab_id)?;
    finish_buffer(&mut full_id, data.full_id)?;
    finish_buffer(&mut pixels, data.pixels)?;
    finish_buffer(&mut model_default, data.model_default)?;
    finish_buffer(&mut model_animated_face, data.model_animated_face)?;
    finish_buffer(&mut model, data.model)?;
    finish_buffer(&mut cape_pixels, data.cape_pixels)?;
    for (pixels, buffer) in animation_pixels.iter_mut().zip(animation_buffers) {
        finish_buffer(pixels, buffer)?;
    }

    Some(Skin {
        width: info.width,
        height: info.height,
        persona: info.persona != 0,
        play_fab_id: String::from_utf8(play_fab_id).ok()?,
        full_id: String::from_utf8(full_id).ok()?,
        pixels,
        model_config: SkinModelConfig {
            default: String::from_utf8(model_default).ok()?,
            animated_face: String::from_utf8(model_animated_face).ok()?,
        },
        model,
        cape: Cape {
            width: info.cape_width,
            height: info.cape_height,
            pixels: cape_pixels,
        },
        animations: animation_info
            .into_iter()
            .zip(animation_pixels)
            .map(|(animation, pixels)| {
                Some(SkinAnimation {
                    width: animation.width,
                    height: animation.height,
                    animation_type: SkinAnimationType::from_raw(animation.animation_type)?,
                    pixels,
                    frame_count: animation.frame_count,
                    expression: animation.expression,
                })
            })
            .collect::<Option<Vec<_>>>()?,
    })
}

include!("player_state_generated.rs");
include!("items_generated.rs");

impl Player {
    pub fn add_effect(&self, effect: effect::Effect) {
        self.change_effect(dragonfly_plugin_sys::DF_PLAYER_EFFECT_ADD, effect);
    }

    pub fn remove_effect(&self, effect_type: impl effect::Type) {
        self.change_effect(
            dragonfly_plugin_sys::DF_PLAYER_EFFECT_REMOVE,
            effect::Effect {
                effect_type: effect_type.id(),
                level: 0,
                duration: std::time::Duration::ZERO,
                potency: 1.0,
                mode: dragonfly_plugin_sys::DF_EFFECT_MODE_TIMED,
                particles_hidden: false,
            },
        );
    }

    fn change_effect(&self, operation: u32, effect: effect::Effect) {
        let host = HOST_API.load(Ordering::Acquire);
        let Some(host) = (unsafe { host.as_ref() }) else {
            return;
        };
        let Some(change) = host.player_effect else {
            return;
        };
        let raw = dragonfly_plugin_sys::DfEffectView {
            effect_type: effect.effect_type,
            level: effect.level,
            duration_milliseconds: duration_milliseconds(effect.duration),
            potency: effect.potency,
            mode: effect.mode,
            particles_hidden: u8::from(effect.particles_hidden),
        };
        let _ = unsafe {
            change(
                host.context,
                current_invocation(),
                self.raw_id(),
                operation,
                raw,
            )
        };
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

pub struct PlayerMoveEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerMoveInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerMoveState,
}

impl<'a> PlayerMoveEventData<'a> {
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

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }

    pub fn rotation(&self) -> Rotation {
        self.input.rotation.into()
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

pub struct PlayerJumpEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerJumpInput,
}

impl<'a> PlayerJumpEventData<'a> {
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

pub struct PlayerTeleportEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerTeleportInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerTeleportState,
}

impl<'a> PlayerTeleportEventData<'a> {
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
    pub const fn __new(
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
    pub fn __new(value: String) -> Self {
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
    pub fn __new(value: impl Into<String>) -> Self {
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
pub fn __write_dynamic_options(
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
    pub fn __new(message: impl Into<String>) -> Self {
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

/// A command sender that can be named and receive messages.
pub trait Source {
    fn name(&self) -> &str;

    fn message(&mut self, message: &str);
}

impl Source for Player {
    fn name(&self) -> &str {
        Player::name(self).unwrap_or("")
    }

    fn message(&mut self, message: &str) {
        Player::message(self, message);
    }
}

/// A console command source tied to the active command invocation.
pub struct ConsoleSource<'a> {
    name: &'a str,
    state: &'a mut dragonfly_plugin_sys::DfCommandState,
}

impl Source for ConsoleSource<'_> {
    fn name(&self) -> &str {
        self.name
    }

    fn message(&mut self, message: &str) {
        if write_command_output(self.state, message, false).is_err() {
            write_command_overflow(self.state, "command output exceeded the runtime buffer");
        }
    }
}

/// A command source whose concrete kind is determined at runtime.
pub enum AnySource<'a> {
    Player(Player),
    Console(ConsoleSource<'a>),
}

impl Source for AnySource<'_> {
    fn name(&self) -> &str {
        match self {
            Self::Player(player) => Source::name(player),
            Self::Console(console) => console.name(),
        }
    }

    fn message(&mut self, message: &str) {
        match self {
            Self::Player(player) => Source::message(player, message),
            Self::Console(console) => console.message(message),
        }
    }
}

/// Runtime context passed to command handlers.
///
/// Command output is sent through [`Context::source`] rather than through the context itself.
/// The former response and source-conversion methods are intentionally unavailable:
///
/// ```compile_fail
/// fn removed(context: &mut dragonfly::Context<'_>) {
///     context.reply("hello");
/// }
/// ```
///
/// ```compile_fail
/// fn removed(context: &mut dragonfly::Context<'_>) {
///     context.fail("no");
/// }
/// ```
///
/// ```compile_fail
/// fn removed(context: &mut dragonfly::Context<'_>) {
///     let _ = context.player_context();
/// }
/// ```
///
/// ```compile_fail
/// fn removed(context: &mut dragonfly::Context<'_>) {
///     let _ = context.console_context();
/// }
/// ```
///
/// ```compile_fail
/// fn removed(context: &dragonfly::Context<'_>) {
///     let _ = context.source_name();
/// }
/// ```
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
    pub unsafe fn __from_raw(
        input: &'a dragonfly_plugin_sys::DfCommandInput,
        state: &'a mut dragonfly_plugin_sys::DfCommandState,
    ) -> Self {
        Self {
            input,
            state,
            source: Any,
        }
    }

    pub fn source(&mut self) -> AnySource<'_> {
        if let Some(player) = self.source_player() {
            AnySource::Player(player)
        } else {
            AnySource::Console(ConsoleSource {
                name: unsafe { string_view(self.input.source) },
                state: &mut *self.state,
            })
        }
    }

    #[doc(hidden)]
    pub fn __player_context(&mut self) -> Option<Context<'_, Player>> {
        let source = self.source_player()?;
        Some(Context {
            input: self.input,
            state: &mut *self.state,
            source,
        })
    }

    #[doc(hidden)]
    pub fn __console_context(&mut self) -> Option<Context<'_, Console>> {
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
    pub fn arguments(&self) -> &str {
        unsafe { string_view(self.input.arguments) }
    }

    fn source_player(&self) -> Option<Player> {
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

impl Context<'_, Console> {
    pub fn source(&mut self) -> ConsoleSource<'_> {
        ConsoleSource {
            name: unsafe { string_view(self.input.source) },
            state: &mut *self.state,
        }
    }
}

impl<S> Context<'_, S> {
    #[doc(hidden)]
    pub fn __fail(&mut self, message: &str) {
        if write_command_output(self.state, message, true).is_err() {
            write_command_overflow(self.state, "command error exceeded the runtime buffer");
        }
    }
}

fn write_command_overflow(state: &mut dragonfly_plugin_sys::DfCommandState, fallback: &str) {
    state.failed = 1;
    state.output.len = 0;
    let _ = write_command_output(state, fallback, true);
}

fn write_command_output(
    state: &mut dragonfly_plugin_sys::DfCommandState,
    message: &str,
    failed: bool,
) -> Result<(), MessageTooLong> {
    let capacity = state.output.capacity as usize;
    if message.len() > capacity || (!message.is_empty() && state.output.data.is_null()) {
        return Err(MessageTooLong {
            length: message.len(),
            capacity,
        });
    }
    if !message.is_empty() {
        unsafe {
            core::ptr::copy_nonoverlapping(message.as_ptr(), state.output.data, message.len())
        };
    }
    state.output.len = message.len() as u64;
    state.failed = u8::from(failed);
    Ok(())
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

fn bytes_view(value: &[u8]) -> dragonfly_plugin_sys::DfStringView {
    dragonfly_plugin_sys::DfStringView {
        data: value.as_ptr(),
        len: value.len() as u64,
    }
}

fn slice_pointer<T>(value: &[T]) -> *const T {
    if value.is_empty() {
        core::ptr::null()
    } else {
        value.as_ptr()
    }
}

fn host_api() -> Option<&'static dragonfly_plugin_sys::DfHostApiV15> {
    unsafe { HOST_API.load(Ordering::Acquire).as_ref() }
}

fn read_item_stack(
    open: impl FnOnce(
        &dragonfly_plugin_sys::DfHostApiV15,
        *mut u64,
        *mut dragonfly_plugin_sys::DfItemStackInfo,
    ) -> Option<dragonfly_plugin_sys::DfStatus>,
) -> Option<ItemStack> {
    let host = host_api()?;
    let close = host.item_stack_close?;
    let mut snapshot_id = 0;
    let mut info = dragonfly_plugin_sys::DfItemStackInfo::default();
    if open(host, &mut snapshot_id, &mut info)? != dragonfly_plugin_sys::DF_STATUS_OK {
        return None;
    }
    let snapshot = ItemSnapshot {
        context: host.context,
        invocation: current_invocation(),
        id: snapshot_id,
        close,
    };
    read_item_stack_snapshot(host, snapshot.invocation, snapshot.id, info)
}

fn read_item_stack_snapshot(
    host: &dragonfly_plugin_sys::DfHostApiV15,
    invocation: dragonfly_plugin_sys::DfInvocationId,
    snapshot_id: u64,
    info: dragonfly_plugin_sys::DfItemStackInfo,
) -> Option<ItemStack> {
    const MAX_ITEM_BYTES: u64 = 16 << 20;
    const MAX_ITEM_LIST: usize = 256;

    if snapshot_id == 0 {
        return None;
    }
    let read = host.item_stack_read?;
    let lore_count = usize::try_from(info.lore_count).ok()?;
    let enchantment_count = usize::try_from(info.enchantment_count).ok()?;
    if lore_count > MAX_ITEM_LIST || enchantment_count > MAX_ITEM_LIST {
        return None;
    }
    let total = [
        info.identifier_len,
        info.custom_name_len,
        info.lore_bytes_len,
        info.nbt_len,
        info.values_nbt_len,
    ]
    .into_iter()
    .try_fold(0u64, |total, length| total.checked_add(length))?;
    if total > MAX_ITEM_BYTES {
        return None;
    }
    let mut identifier = allocate_bytes(info.identifier_len)?;
    let mut custom_name = allocate_bytes(info.custom_name_len)?;
    let mut lore_bytes = allocate_bytes(info.lore_bytes_len)?;
    let mut nbt = allocate_bytes(info.nbt_len)?;
    let mut values_nbt = allocate_bytes(info.values_nbt_len)?;
    let mut lore = vec![dragonfly_plugin_sys::DfByteSpan::default(); lore_count];
    let mut enchantments =
        vec![dragonfly_plugin_sys::DfItemEnchantment::default(); enchantment_count];
    let mut data = dragonfly_plugin_sys::DfItemStackData {
        identifier: bytes_buffer(&mut identifier),
        custom_name: bytes_buffer(&mut custom_name),
        lore_bytes: bytes_buffer(&mut lore_bytes),
        nbt: bytes_buffer(&mut nbt),
        values_nbt: bytes_buffer(&mut values_nbt),
        lore: if lore.is_empty() {
            core::ptr::null_mut()
        } else {
            lore.as_mut_ptr()
        },
        lore_capacity: lore.len() as u64,
        enchantments: if enchantments.is_empty() {
            core::ptr::null_mut()
        } else {
            enchantments.as_mut_ptr()
        },
        enchantment_capacity: enchantments.len() as u64,
    };
    if unsafe { read(host.context, invocation, snapshot_id, &mut data) }
        != dragonfly_plugin_sys::DF_STATUS_OK
    {
        return None;
    }
    finish_buffer(&mut identifier, data.identifier)?;
    finish_buffer(&mut custom_name, data.custom_name)?;
    finish_buffer(&mut lore_bytes, data.lore_bytes)?;
    finish_buffer(&mut nbt, data.nbt)?;
    finish_buffer(&mut values_nbt, data.values_nbt)?;

    let lore = lore
        .into_iter()
        .map(|span| {
            let start = usize::try_from(span.offset).ok()?;
            let length = usize::try_from(span.len).ok()?;
            let end = start.checked_add(length)?;
            Some(
                std::str::from_utf8(lore_bytes.get(start..end)?)
                    .ok()?
                    .to_owned(),
            )
        })
        .collect::<Option<Vec<_>>>()?;
    let nbt = if nbt.is_empty() {
        std::collections::BTreeMap::new()
    } else {
        item_nbt::decode_values(&nbt).ok()?
    };
    let values = if values_nbt.is_empty() {
        std::collections::BTreeMap::new()
    } else {
        item_nbt::decode_values(&values_nbt).ok()?
    };
    let enchantments = enchantments
        .into_iter()
        .map(|enchantment| {
            AppliedEnchantment::new(Enchantment::from_id(enchantment.id), enchantment.level)
        })
        .collect();
    Some(ItemStack {
        identifier: String::from_utf8(identifier).ok()?,
        metadata: i16::try_from(info.metadata).ok()?,
        count: info.count,
        damage: info.damage,
        unbreakable: info.unbreakable != 0,
        anvil_cost: info.anvil_cost,
        custom_name: String::from_utf8(custom_name).ok()?,
        lore,
        nbt,
        values,
        enchantments,
    })
}

fn allocate_bytes(length: u64) -> Option<Vec<u8>> {
    let length = usize::try_from(length).ok()?;
    let mut bytes = Vec::new();
    bytes.try_reserve_exact(length).ok()?;
    bytes.resize(length, 0);
    Some(bytes)
}

fn bytes_buffer(value: &mut [u8]) -> dragonfly_plugin_sys::DfStringBuffer {
    dragonfly_plugin_sys::DfStringBuffer {
        data: value.as_mut_ptr(),
        len: 0,
        capacity: value.len() as u64,
    }
}

fn finish_buffer(value: &mut Vec<u8>, buffer: dragonfly_plugin_sys::DfStringBuffer) -> Option<()> {
    let length = usize::try_from(buffer.len).ok()?;
    if length > value.len() {
        return None;
    }
    value.truncate(length);
    Some(())
}

fn duration_milliseconds(duration: std::time::Duration) -> u64 {
    duration.as_millis().min(u128::from(u64::MAX)) as u64
}

pub struct PlayerChatEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerChatInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerChatState,
}

pub struct PlayerJoinEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerJoinInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerJoinState,
}

impl<'a> PlayerJoinEventData<'a> {
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

pub struct PlayerQuitEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerQuitInput,
}

impl<'a> PlayerQuitEventData<'a> {
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

impl<'a> PlayerChatEventData<'a> {
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

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
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

pub struct PlayerHurtEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerHurtInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerHurtState,
}

impl<'a> PlayerHurtEventData<'a> {
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

    pub fn damage_source(&self) -> damage::Source<'_> {
        unsafe { damage::Source::from_raw(&self.input.source) }.unwrap_or_else(|| {
            damage::Custom::new(
                "unknown",
                damage::Traits::default(),
                damage::AffectedProtections::NONE,
            )
            .into()
        })
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

pub struct PlayerHealEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerHealInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerHealState,
}

pub struct PlayerBlockBreakEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerBlockBreakInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerBlockBreakState,
}

impl<'a> PlayerBlockBreakEventData<'a> {
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

pub struct PlayerBlockPlaceEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerBlockPlaceInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerBlockPlaceState,
}

impl<'a> PlayerBlockPlaceEventData<'a> {
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

impl<'a> PlayerHealEventData<'a> {
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

    pub fn healing_source(&self) -> healing::Source<'_> {
        unsafe { healing::Source::from_raw(&self.input.source) }
            .unwrap_or_else(|| healing::Custom::new("unknown").into())
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

pub struct PlayerFoodLossEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerFoodLossInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerFoodLossState,
}

impl<'a> PlayerFoodLossEventData<'a> {
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

pub struct PlayerDeathEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerDeathInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerDeathState,
}

impl<'a> PlayerDeathEventData<'a> {
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

    pub fn damage_source(&self) -> damage::Source<'_> {
        unsafe { damage::Source::from_raw(&self.input.source) }.unwrap_or_else(|| {
            damage::Custom::new(
                "unknown",
                damage::Traits::default(),
                damage::AffectedProtections::NONE,
            )
            .into()
        })
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
    PlayerStartBreakEventData,
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
    PlayerToggleSprintEventData,
    dragonfly_plugin_sys::DfPlayerToggleSprintInput,
    dragonfly_plugin_sys::DfPlayerToggleSprintState
);
cancellable_toggle_event!(
    PlayerToggleSneakEventData,
    dragonfly_plugin_sys::DfPlayerToggleSneakInput,
    dragonfly_plugin_sys::DfPlayerToggleSneakState
);
cancellable_position_event!(
    PlayerFireExtinguishEventData,
    dragonfly_plugin_sys::DfPlayerFireExtinguishInput,
    dragonfly_plugin_sys::DfPlayerFireExtinguishState
);

pub struct PlayerExperienceGainEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerExperienceGainInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerExperienceGainState,
}

impl<'a> PlayerExperienceGainEventData<'a> {
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

pub struct PlayerPunchAirEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerPunchAirInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerPunchAirState,
}

pub struct PlayerHeldSlotChangeEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerHeldSlotChangeInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerHeldSlotChangeState,
}

impl<'a> PlayerHeldSlotChangeEventData<'a> {
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

pub struct PlayerSleepEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerSleepInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerSleepState,
}

impl<'a> PlayerSleepEventData<'a> {
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

impl<'a> PlayerPunchAirEventData<'a> {
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

pub struct PlayerBlockPickEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerBlockPickInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerBlockPickState,
}

impl<'a> PlayerBlockPickEventData<'a> {
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

pub struct PlayerLecternPageTurnEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerLecternPageTurnInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerLecternPageTurnState,
}

pub struct PlayerSignEditEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerSignEditInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerSignEditState,
}

impl<'a> PlayerSignEditEventData<'a> {
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

pub struct PlayerItemUseEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerItemUseInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerItemUseState,
}

pub struct PlayerItemUseOnBlockEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerItemUseOnBlockInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerItemUseOnBlockState,
}

pub struct PlayerItemConsumeEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerItemConsumeInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerItemConsumeState,
}

impl<'a> PlayerItemConsumeEventData<'a> {
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
    pub fn item(&self) -> ItemStack {
        ItemStack::from_snapshot(&self.input.item).unwrap_or_else(ItemStack::empty)
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerItemReleaseEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerItemReleaseInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerItemReleaseState,
}

pub struct PlayerItemDamageEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerItemDamageInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerItemDamageState,
}

impl<'a> PlayerItemDamageEventData<'a> {
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
    pub fn item(&self) -> ItemStack {
        ItemStack::from_snapshot(&self.input.item).unwrap_or_else(ItemStack::empty)
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

pub struct PlayerItemDropEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerItemDropInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerItemDropState,
}

impl<'a> PlayerItemDropEventData<'a> {
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
    pub fn item(&self) -> ItemStack {
        ItemStack::from_snapshot(&self.input.item).unwrap_or_else(ItemStack::empty)
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

impl<'a> PlayerItemReleaseEventData<'a> {
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
    pub fn item(&self) -> ItemStack {
        ItemStack::from_snapshot(&self.input.item).unwrap_or_else(ItemStack::empty)
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

impl<'a> PlayerItemUseOnBlockEventData<'a> {
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

impl<'a> PlayerItemUseEventData<'a> {
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

impl<'a> PlayerLecternPageTurnEventData<'a> {
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

pub struct PlayerAttackEntityEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerAttackEntityInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerAttackEntityState,
}

impl<'a> PlayerAttackEntityEventData<'a> {
    /// # Safety
    /// Both references must belong to the same active attack callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerAttackEntityInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerAttackEntityState,
    ) -> Self {
        Self { input, state }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }
    pub fn target(&self) -> Entity {
        self.input.target.into()
    }
    pub fn knockback_force(&self) -> f64 {
        self.state.knockback_force
    }
    pub fn set_knockback_force(&mut self, force: f64) {
        if force.is_finite() {
            self.state.knockback_force = force;
        }
    }
    pub fn knockback_height(&self) -> f64 {
        self.state.knockback_height
    }
    pub fn set_knockback_height(&mut self, height: f64) {
        if height.is_finite() {
            self.state.knockback_height = height;
        }
    }
    pub fn critical(&self) -> bool {
        self.state.critical != 0
    }
    pub fn set_critical(&mut self, critical: bool) {
        self.state.critical = u8::from(critical);
    }
    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }
    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerItemUseOnEntityEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerItemUseOnEntityInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerItemUseOnEntityState,
}

impl<'a> PlayerItemUseOnEntityEventData<'a> {
    /// # Safety
    /// Both references must belong to the same active item-use callback.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerItemUseOnEntityInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerItemUseOnEntityState,
    ) -> Self {
        Self { input, state }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }

    pub fn target(&self) -> Entity {
        self.input.target.into()
    }

    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }

    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }
}

pub struct PlayerChangeWorldEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerChangeWorldInput,
}

impl<'a> PlayerChangeWorldEventData<'a> {
    /// # Safety
    /// The reference must belong to an active change-world callback whose
    /// world handles were validated by the runtime.
    #[doc(hidden)]
    pub unsafe fn from_raw(input: &'a dragonfly_plugin_sys::DfPlayerChangeWorldInput) -> Self {
        Self { input }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }

    pub fn before(&self) -> Option<World> {
        (self.input.before.value != 0).then(|| World::from_valid_raw(self.input.before.value))
    }

    pub fn after(&self) -> World {
        World::from_valid_raw(self.input.after.value)
    }
}

pub struct PlayerRespawnEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerRespawnInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerRespawnState,
}

impl<'a> PlayerRespawnEventData<'a> {
    /// # Safety
    /// Both references must belong to the same active respawn callback. The
    /// runtime must have validated the player, position, and world handles.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerRespawnInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerRespawnState,
    ) -> Self {
        Self { input, state }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }

    pub fn position(&self) -> Vec3 {
        self.state.position.into()
    }

    pub fn set_position(&mut self, position: Vec3) {
        self.state.position = position.into();
    }

    pub fn world(&self) -> World {
        World::from_valid_raw(self.state.world.value)
    }

    pub fn set_world(&mut self, world: World) {
        self.state.world = world.raw_id();
    }
}

pub struct PlayerSkinChangeEventData<'a> {
    input: &'a dragonfly_plugin_sys::DfPlayerSkinChangeInput,
    state: &'a mut dragonfly_plugin_sys::DfPlayerSkinChangeState,
}

impl<'a> PlayerSkinChangeEventData<'a> {
    /// # Safety
    /// Both references must belong to the same active skin-change callback.
    /// The runtime owns the snapshot until the complete plugin chain returns.
    #[doc(hidden)]
    pub unsafe fn from_raw(
        input: &'a dragonfly_plugin_sys::DfPlayerSkinChangeInput,
        state: &'a mut dragonfly_plugin_sys::DfPlayerSkinChangeState,
    ) -> Self {
        Self { input, state }
    }

    pub fn player(&self) -> Player {
        Player::from_id(self.input.player)
    }

    pub fn skin(&self) -> Skin {
        self.read_skin().unwrap_or_default()
    }

    pub fn set_skin(&mut self, skin: &Skin) {
        let Some(host) = host_api() else {
            return;
        };
        let Some(set) = host.skin_snapshot_set else {
            return;
        };
        with_skin_view(skin, |view| {
            let _ = unsafe {
                set(
                    host.context,
                    self.input.invocation,
                    self.input.snapshot,
                    view,
                )
            };
        });
    }

    pub fn edit_skin(&mut self, edit: impl FnOnce(&mut Skin)) {
        let Some(mut skin) = self.read_skin() else {
            return;
        };
        edit(&mut skin);
        self.set_skin(&skin);
    }

    pub fn cancelled(&self) -> bool {
        self.state.cancelled != 0
    }

    pub fn cancel(&mut self) {
        self.state.cancelled = 1;
    }

    fn read_skin(&self) -> Option<Skin> {
        let host = host_api()?;
        let info_fn = host.skin_snapshot_info?;
        let mut info = dragonfly_plugin_sys::DfSkinInfo::default();
        let status = unsafe {
            info_fn(
                host.context,
                self.input.invocation,
                self.input.snapshot,
                &mut info,
            )
        };
        if status != dragonfly_plugin_sys::DF_STATUS_OK {
            return None;
        }
        read_skin_snapshot(host, self.input.invocation, self.input.snapshot, info)
    }
}

pub trait Plugin: Default + Send + Sync + 'static {
    fn on_enable(&self) {}
    fn on_disable(&self) {}
    fn on_move(&self, _event: &mut Event::PlayerMove<'_>) {}
    fn on_chat(&self, _event: &mut Event::PlayerChat<'_>) {}
    fn on_join(&self, _event: &mut Event::PlayerJoin<'_>) {}
    fn on_quit(&self, _event: &Event::PlayerQuit<'_>) {}
    fn on_hurt(&self, _event: &mut Event::PlayerHurt<'_>) {}
    fn on_heal(&self, _event: &mut Event::PlayerHeal<'_>) {}
    fn on_block_break(&self, _event: &mut Event::PlayerBlockBreak<'_>) {}
    fn on_block_place(&self, _event: &mut Event::PlayerBlockPlace<'_>) {}
    fn on_food_loss(&self, _event: &mut Event::PlayerFoodLoss<'_>) {}
    fn on_death(&self, _event: &mut Event::PlayerDeath<'_>) {}
    fn on_start_break(&self, _event: &mut Event::PlayerStartBreak<'_>) {}
    fn on_fire_extinguish(&self, _event: &mut Event::PlayerFireExtinguish<'_>) {}
    fn on_toggle_sprint(&self, _event: &mut Event::PlayerToggleSprint<'_>) {}
    fn on_toggle_sneak(&self, _event: &mut Event::PlayerToggleSneak<'_>) {}
    fn on_jump(&self, _event: &Event::PlayerJump<'_>) {}
    fn on_teleport(&self, _event: &mut Event::PlayerTeleport<'_>) {}
    fn on_experience_gain(&self, _event: &mut Event::PlayerExperienceGain<'_>) {}
    fn on_punch_air(&self, _event: &mut Event::PlayerPunchAir<'_>) {}
    fn on_held_slot_change(&self, _event: &mut Event::PlayerHeldSlotChange<'_>) {}
    fn on_sleep(&self, _event: &mut Event::PlayerSleep<'_>) {}
    fn on_block_pick(&self, _event: &mut Event::PlayerBlockPick<'_>) {}
    fn on_lectern_page_turn(&self, _event: &mut Event::PlayerLecternPageTurn<'_>) {}
    fn on_sign_edit(&self, _event: &mut Event::PlayerSignEdit<'_>) {}
    fn on_item_use(&self, _event: &mut Event::PlayerItemUse<'_>) {}
    fn on_item_use_on_block(&self, _event: &mut Event::PlayerItemUseOnBlock<'_>) {}
    fn on_item_consume(&self, _event: &mut Event::PlayerItemConsume<'_>) {}
    fn on_item_release(&self, _event: &mut Event::PlayerItemRelease<'_>) {}
    fn on_item_damage(&self, _event: &mut Event::PlayerItemDamage<'_>) {}
    fn on_item_drop(&self, _event: &mut Event::PlayerItemDrop<'_>) {}
    fn on_attack_entity(&self, _event: &mut Event::PlayerAttackEntity<'_>) {}
    fn on_item_use_on_entity(&self, _event: &mut Event::PlayerItemUseOnEntity<'_>) {}
    fn on_change_world(&self, _event: &Event::PlayerChangeWorld<'_>) {}
    fn on_respawn(&self, _event: &mut Event::PlayerRespawn<'_>) {}
    fn on_skin_change(&self, _event: &mut Event::PlayerSkinChange<'_>) {}
    fn on_entity_hurt(&self, _event: &mut Event::EntityHurt<'_>) {}
    fn on_entity_heal(&self, _event: &mut Event::EntityHeal<'_>) {}
    fn on_entity_death(&self, _event: &mut Event::EntityDeath<'_>) {}
    fn commands(&self) -> &'static [Command] {
        &[]
    }
    fn on_command(&self, _command: usize, _context: &mut Context<'_>) {}
}

#[cfg(test)]
mod tests {
    use super::*;

    fn command_input(
        source: &str,
        source_kind: u32,
        source_player: dragonfly_plugin_sys::DfPlayerId,
        online_players: &[dragonfly_plugin_sys::DfCommandPlayer],
    ) -> dragonfly_plugin_sys::DfCommandInput {
        dragonfly_plugin_sys::DfCommandInput {
            invocation: 0,
            source: string_view_from_str(source),
            arguments: string_view_from_str(""),
            source_kind,
            source_player,
            online_players: online_players.as_ptr(),
            online_player_count: online_players.len() as u64,
        }
    }

    fn command_state(output: &mut [u8]) -> dragonfly_plugin_sys::DfCommandState {
        dragonfly_plugin_sys::DfCommandState {
            failed: 0,
            output: dragonfly_plugin_sys::DfStringBuffer {
                data: output.as_mut_ptr(),
                len: 0,
                capacity: output.len() as u64,
            },
        }
    }

    fn message_from_common_source(source: &mut impl Source, expected_name: &str) {
        assert_eq!(source.name(), expected_name);
        source.message("hello from source");
    }

    #[test]
    fn player_context_source_is_the_player_value() {
        let raw_player = dragonfly_plugin_sys::DfPlayerId {
            bytes: [7; 16],
            generation: 9,
        };
        let player_name = "Gopher";
        let online_players = [dragonfly_plugin_sys::DfCommandPlayer {
            player: raw_player,
            latency_milliseconds: 42,
            name: string_view_from_str(player_name),
        }];
        let input = command_input(
            player_name,
            dragonfly_plugin_sys::DF_COMMAND_SOURCE_PLAYER,
            raw_player,
            &online_players,
        );
        let mut output = [0; 64];
        let mut state = command_state(&mut output);
        let mut context = unsafe { Context::__from_raw(&input, &mut state) };
        let context = context.__player_context().unwrap();

        let player: Player = context.source();
        assert_eq!(player.name(), Some(player_name));
        assert_eq!(player.id().generation(), 9);
        player.message("");
    }

    #[test]
    fn console_context_source_names_and_messages_through_command_state() {
        let input = command_input(
            "Console",
            dragonfly_plugin_sys::DF_COMMAND_SOURCE_CONSOLE,
            dragonfly_plugin_sys::DfPlayerId::default(),
            &[],
        );
        let mut output = [0; 64];
        let mut state = command_state(&mut output);
        let mut context = unsafe { Context::__from_raw(&input, &mut state) };
        {
            let mut context = context.__console_context().unwrap();
            let mut source: ConsoleSource<'_> = context.source();
            message_from_common_source(&mut source, "Console");
        }

        assert_eq!(state.failed, 0);
        assert_eq!(state.output.len, 17);
        assert_eq!(&output[..17], b"hello from source");
    }

    #[test]
    fn any_context_source_is_an_explicit_common_source_enum() {
        let input = command_input(
            "Console",
            dragonfly_plugin_sys::DF_COMMAND_SOURCE_CONSOLE,
            dragonfly_plugin_sys::DfPlayerId::default(),
            &[],
        );
        let mut output = [0; 64];
        let mut state = command_state(&mut output);
        let mut context = unsafe { Context::__from_raw(&input, &mut state) };
        let mut source: AnySource<'_> = context.source();
        assert!(matches!(source, AnySource::Console(_)));
        message_from_common_source(&mut source, "Console");

        assert_eq!(state.output.len, 17);
        assert_eq!(&output[..17], b"hello from source");
    }

    #[test]
    fn invocation_scope_restores_nested_value() {
        assert_eq!(current_invocation(), 0);
        with_invocation(11, || {
            assert_eq!(current_invocation(), 11);
            with_invocation(22, || assert_eq!(current_invocation(), 22));
            assert_eq!(current_invocation(), 11);
        });
        assert_eq!(current_invocation(), 0);
    }

    #[test]
    fn invocation_scope_restores_value_after_panic() {
        let panic = std::panic::catch_unwind(|| {
            with_invocation(33, || panic!("test panic"));
        });
        assert!(panic.is_err());
        assert_eq!(current_invocation(), 0);
    }

    #[test]
    fn invocation_scope_is_thread_local() {
        with_invocation(44, || {
            let child = std::thread::spawn(|| {
                assert_eq!(current_invocation(), 0);
                with_invocation(55, current_invocation)
            });
            assert_eq!(child.join().unwrap(), 55);
            assert_eq!(current_invocation(), 44);
        });
        assert_eq!(current_invocation(), 0);
    }

    #[test]
    fn scoreboard_enforces_protocol_limit_and_matches_line_semantics() {
        let mut board = Scoreboard::new("Stats");
        board.push_line("first\n\n\n").unwrap();
        board.set_line(2, "third").unwrap();
        board.remove_padding().set_descending(true);
        assert_eq!(board.name(), "Stats");
        assert_eq!(board.lines(), ["first\n", "", "third"]);
        assert!(!board.padding());
        assert!(board.is_descending());
        for index in 3..MAX_SCOREBOARD_LINES {
            board.set_line(index, index.to_string()).unwrap();
        }
        assert_eq!(
            board.push_line("too many"),
            Err(ScoreboardLineOutOfBounds {
                index: MAX_SCOREBOARD_LINES
            })
        );
        assert_eq!(board.remove_line(1), Some(String::new()));
    }

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
        fn on_move(&self, event: &mut Event::PlayerMove<'_>) {
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
        let mut event = unsafe { PlayerMoveEventData::from_raw(&input, &mut state) };
        Guard.on_move(&mut event);
        assert!(event.cancelled());
    }

    #[test]
    fn movement_exposes_player_and_rotation() {
        let input = dragonfly_plugin_sys::DfPlayerMoveInput {
            player: dragonfly_plugin_sys::DfPlayerId {
                bytes: [3; 16],
                generation: 11,
            },
            rotation: dragonfly_plugin_sys::DfRotation {
                yaw: 92.5,
                pitch: -14.25,
            },
            ..Default::default()
        };
        let mut state = dragonfly_plugin_sys::DfPlayerMoveState::default();
        let event = unsafe { PlayerMoveEventData::from_raw(&input, &mut state) };

        assert_eq!(event.player().id().uuid_bytes(), [3; 16]);
        assert_eq!(event.player().id().generation(), 11);
        assert_eq!(
            event.rotation(),
            Rotation {
                yaw: 92.5,
                pitch: -14.25
            }
        );
    }

    #[test]
    fn chat_exposes_player() {
        let input = dragonfly_plugin_sys::DfPlayerChatInput {
            player: dragonfly_plugin_sys::DfPlayerId {
                bytes: [5; 16],
                generation: 13,
            },
            ..Default::default()
        };
        let mut state = dragonfly_plugin_sys::DfPlayerChatState::default();
        let event = unsafe { PlayerChatEventData::from_raw(&input, &mut state) };

        assert_eq!(event.player().id().uuid_bytes(), [5; 16]);
        assert_eq!(event.player().id().generation(), 13);
    }

    #[test]
    fn item_use_on_entity_exposes_stable_participants_and_cancellation() {
        let input = dragonfly_plugin_sys::DfPlayerItemUseOnEntityInput {
            invocation: 7,
            player: dragonfly_plugin_sys::DfPlayerId {
                bytes: [3; 16],
                generation: 11,
            },
            target: dragonfly_plugin_sys::DfEntityId {
                bytes: [5; 16],
                generation: 13,
            },
        };
        let mut state = dragonfly_plugin_sys::DfPlayerItemUseOnEntityState::default();
        let mut event = unsafe { PlayerItemUseOnEntityEventData::from_raw(&input, &mut state) };

        assert_eq!(event.player().id().uuid_bytes(), [3; 16]);
        assert_eq!(event.player().id().generation(), 11);
        assert_eq!(event.target().id().uuid_bytes(), [5; 16]);
        assert_eq!(event.target().id().generation(), 13);
        assert!(!event.cancelled());
        event.cancel();
        assert!(event.cancelled());
    }

    #[test]
    fn change_world_exposes_optional_before_without_panicking() {
        let mut input = dragonfly_plugin_sys::DfPlayerChangeWorldInput {
            invocation: 17,
            player: dragonfly_plugin_sys::DfPlayerId {
                bytes: [7; 16],
                generation: 19,
            },
            before: dragonfly_plugin_sys::DfWorldId { value: 0 },
            after: dragonfly_plugin_sys::DfWorldId { value: 23 },
        };
        let event = unsafe { PlayerChangeWorldEventData::from_raw(&input) };
        assert_eq!(event.player().id().generation(), 19);
        assert_eq!(event.before(), None);
        assert_eq!(event.after().raw_id().value, 23);

        input.before.value = 29;
        let event = unsafe { PlayerChangeWorldEventData::from_raw(&input) };
        assert_eq!(event.before().map(|world| world.raw_id().value), Some(29));
        assert_eq!(event.after().raw_id().value, 23);
    }

    #[test]
    fn respawn_exposes_mutable_position_and_world() {
        let input = dragonfly_plugin_sys::DfPlayerRespawnInput {
            invocation: 31,
            player: dragonfly_plugin_sys::DfPlayerId {
                bytes: [11; 16],
                generation: 37,
            },
        };
        let mut state = dragonfly_plugin_sys::DfPlayerRespawnState {
            position: dragonfly_plugin_sys::DfVec3 {
                x: 1.0,
                y: 64.0,
                z: 2.0,
            },
            world: dragonfly_plugin_sys::DfWorldId { value: 41 },
        };
        let mut event = unsafe { PlayerRespawnEventData::from_raw(&input, &mut state) };

        assert_eq!(event.player().id().generation(), 37);
        assert_eq!(
            event.position(),
            Vec3 {
                x: 1.0,
                y: 64.0,
                z: 2.0
            }
        );
        assert_eq!(event.world().raw_id().value, 41);

        event.set_position(Vec3 {
            x: -3.0,
            y: 80.0,
            z: 9.0,
        });
        event.set_world(World::from_valid_raw(43));
        assert_eq!(
            event.position(),
            Vec3 {
                x: -3.0,
                y: 80.0,
                z: 9.0
            }
        );
        assert_eq!(event.world().raw_id().value, 43);
    }

    #[test]
    fn skin_change_exposes_player_and_monotonic_cancellation() {
        let input = dragonfly_plugin_sys::DfPlayerSkinChangeInput {
            invocation: 31,
            player: dragonfly_plugin_sys::DfPlayerId {
                bytes: [11; 16],
                generation: 37,
            },
            snapshot: 41,
        };
        let mut state = dragonfly_plugin_sys::DfPlayerSkinChangeState { cancelled: 0 };
        let mut event = unsafe { PlayerSkinChangeEventData::from_raw(&input, &mut state) };

        assert_eq!(event.player().id().generation(), 37);
        assert!(!event.cancelled());
        event.cancel();
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
            Player::__from_command_argument("000102030405060708090a0b0c0d0e0f:9:42:Danick")
                .unwrap();
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

    #[test]
    fn item_stack_builders_keep_typed_data() {
        let nested = Value::Compound(std::collections::BTreeMap::from([
            ("enabled".to_owned(), Value::from(true)),
            (
                "levels".to_owned(),
                Value::List(vec![Value::Int(1), Value::Int(2)]),
            ),
        ]));
        let stack = item::new(item::Sword::new(item::ToolTier::Diamond), 1)
            .with_damage(7)
            .as_unbreakable()
            .with_anvil_cost(4)
            .with_custom_name("Guard blade")
            .with_lore(["First line", "Second line"])
            .with_value("example:data", nested.clone())
            .with_enchantment(Enchantment::Sharpness, 5)
            .with_enchantment(Enchantment::Unbreaking, 3);

        assert_eq!(stack.identifier(), "minecraft:diamond_sword");
        assert_eq!(stack.count(), 1);
        assert_eq!(stack.damage(), 7);
        assert!(stack.unbreakable());
        assert_eq!(stack.anvil_cost(), 4);
        assert_eq!(stack.custom_name(), "Guard blade");
        assert_eq!(stack.lore(), ["First line", "Second line"]);
        assert_eq!(stack.value("example:data"), Some(&nested));
        assert_eq!(
            stack.enchantment(Enchantment::Sharpness),
            Some(AppliedEnchantment::new(Enchantment::Sharpness, 5))
        );
        assert_eq!(Enchantment::Sharpness.max_level(), Some(5));
        assert_eq!(Enchantment::Custom(512).id(), 512);
    }

    #[test]
    fn generated_items_and_custom_metadata_are_typed() {
        let diamond = item::new(item::Diamond, 4);
        assert_eq!(diamond.identifier(), "minecraft:diamond");
        assert_eq!(diamond.metadata(), 0);

        let metadata: i16 = i16::MAX;
        let custom = item::new(
            item::Custom::new("example:variant").with_metadata(metadata),
            2,
        );
        assert_eq!(custom.identifier(), "example:variant");
        assert_eq!(custom.metadata(), metadata);

        let pickaxe = item::Pickaxe::new(item::ToolTier::Copper);
        assert_eq!(pickaxe.tier(), item::ToolTier::Copper);
        assert_eq!(
            item::new(pickaxe, 1).identifier(),
            "minecraft:copper_pickaxe"
        );
    }

    #[test]
    fn potion_helpers_use_dragonfly_metadata() {
        let potion = ItemStack::potion(Potion::StrongHealing, 2);
        assert_eq!(potion.identifier(), "minecraft:potion");
        assert_eq!(potion.count(), 2);
        assert_eq!(potion.metadata(), 22);
        assert_eq!(Potion::from_id(22), Potion::StrongHealing);
        assert_eq!(Potion::from_id(255), Potion::Custom(255));
    }
}
