use crate::{BlockPos, Entity, Player, Vec3, block, entity, particle, sound};

const MAX_WORLD_NAME_BYTES: usize = 256;
const MAX_BLOCK_IDENTIFIER_BYTES: usize = 256;
const MAX_BLOCK_PROPERTIES_BYTES: usize = 64 << 10;

#[repr(u32)]
#[derive(Clone, Copy, Debug, Default, Eq, Hash, PartialEq)]
pub enum Dimension {
    #[default]
    Overworld = dragonfly_plugin_sys::DF_WORLD_DIMENSION_OVERWORLD,
    Nether = dragonfly_plugin_sys::DF_WORLD_DIMENSION_NETHER,
    End = dragonfly_plugin_sys::DF_WORLD_DIMENSION_END,
}

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct World {
    raw: u64,
}

impl World {
    pub(crate) const fn from_raw(raw: u64) -> Option<Self> {
        if raw == 0 { None } else { Some(Self { raw }) }
    }

    pub(crate) const fn from_valid_raw(raw: u64) -> Self {
        Self { raw }
    }

    pub fn get(name: &str) -> Option<Self> {
        let host = crate::host_api()?;
        let lookup = host.world_lookup?;
        let mut raw = dragonfly_plugin_sys::DfWorldId::default();
        let status = unsafe {
            lookup(
                host.context,
                crate::current_invocation(),
                crate::string_view_from_str(name),
                &mut raw,
            )
        };
        (status == dragonfly_plugin_sys::DF_STATUS_OK && raw.value != 0)
            .then_some(Self { raw: raw.value })
    }

    /// Opens an existing persistent custom world or creates it when absent.
    pub fn open(name: &str, dimension: Dimension) -> Option<Self> {
        let host = crate::host_api()?;
        let open = host.world_open?;
        let mut raw = dragonfly_plugin_sys::DfWorldId::default();
        let status = unsafe {
            open(
                host.context,
                crate::current_invocation(),
                crate::string_view_from_str(name),
                dimension as u32,
                &mut raw,
            )
        };
        (status == dragonfly_plugin_sys::DF_STATUS_OK && raw.value != 0)
            .then_some(Self { raw: raw.value })
    }

    pub fn overworld() -> Option<Self> {
        Self::get("minecraft:overworld")
    }

    pub fn nether() -> Option<Self> {
        Self::get("minecraft:nether")
    }

    pub fn end() -> Option<Self> {
        Self::get("minecraft:end")
    }

    pub fn name(&self) -> Option<String> {
        let host = crate::host_api()?;
        let name = host.world_name?;
        let mut bytes = [0; MAX_WORLD_NAME_BYTES];
        let mut output = crate::bytes_buffer(&mut bytes);
        let status = unsafe {
            name(
                host.context,
                crate::current_invocation(),
                self.raw_id(),
                &mut output,
            )
        };
        if status != dragonfly_plugin_sys::DF_STATUS_OK {
            return None;
        }
        let length = usize::try_from(output.len).ok()?;
        String::from_utf8(bytes.get(..length)?.to_vec()).ok()
    }

    /// Unloads a custom world. Returns false for core worlds, stale handles, or occupied worlds.
    pub fn unload(self) -> bool {
        let Some(host) = crate::host_api() else {
            return false;
        };
        let Some(unload) = host.world_unload else {
            return false;
        };
        (unsafe { unload(host.context, crate::current_invocation(), self.raw_id()) })
            == dragonfly_plugin_sys::DF_STATUS_OK
    }

    pub fn save(&self) {
        let Some(host) = crate::host_api() else {
            return;
        };
        let Some(save) = host.world_save else { return };
        let _ = unsafe { save(host.context, crate::current_invocation(), self.raw_id()) };
    }

    pub fn block(&self, position: BlockPos) -> Option<block::Block> {
        let host = crate::host_api()?;
        let get = host.world_block_get?;
        let mut identifier = [0; MAX_BLOCK_IDENTIFIER_BYTES];
        let mut properties = [0; 1024];
        let mut data = dragonfly_plugin_sys::DfBlockData {
            identifier: crate::bytes_buffer(&mut identifier),
            properties_nbt: crate::bytes_buffer(&mut properties),
        };
        let mut status = unsafe {
            get(
                host.context,
                crate::current_invocation(),
                self.raw_id(),
                position.into(),
                &mut data,
            )
        };
        let required_identifier = usize::try_from(data.identifier.len).ok()?;
        let required_properties = usize::try_from(data.properties_nbt.len).ok()?;
        if required_identifier > MAX_BLOCK_IDENTIFIER_BYTES
            || required_properties > MAX_BLOCK_PROPERTIES_BYTES
        {
            return None;
        }
        let mut large_properties = Vec::new();
        if status != dragonfly_plugin_sys::DF_STATUS_OK
            && (required_identifier > identifier.len() || required_properties > properties.len())
        {
            large_properties
                .try_reserve_exact(required_properties)
                .ok()?;
            large_properties.resize(required_properties, 0);
            data = dragonfly_plugin_sys::DfBlockData {
                identifier: crate::bytes_buffer(&mut identifier),
                properties_nbt: crate::bytes_buffer(&mut large_properties),
            };
            status = unsafe {
                get(
                    host.context,
                    crate::current_invocation(),
                    self.raw_id(),
                    position.into(),
                    &mut data,
                )
            };
        }
        if status != dragonfly_plugin_sys::DF_STATUS_OK {
            return None;
        }
        let identifier_length = usize::try_from(data.identifier.len).ok()?;
        let properties_length = usize::try_from(data.properties_nbt.len).ok()?;
        let identifier = String::from_utf8(identifier.get(..identifier_length)?.to_vec()).ok()?;
        let properties = if large_properties.is_empty() {
            properties.get(..properties_length)?
        } else {
            large_properties.get(..properties_length)?
        };
        block::Block::from_nbt(identifier, properties)
    }

    pub fn set_block(&self, position: BlockPos, block: &block::Block) {
        let Some(host) = crate::host_api() else {
            return;
        };
        let Some(set) = host.world_block_set else {
            return;
        };
        let Some(properties) = block.properties_nbt() else {
            return;
        };
        let view = dragonfly_plugin_sys::DfBlockView {
            identifier: crate::string_view_from_str(block.identifier()),
            properties_nbt: crate::bytes_view(&properties),
        };
        let _ = unsafe {
            set(
                host.context,
                crate::current_invocation(),
                self.raw_id(),
                position.into(),
                &view,
            )
        };
    }

    pub fn time(&self) -> Option<i64> {
        let host = crate::host_api()?;
        let get = host.world_time_get?;
        let mut value = 0;
        let status = unsafe {
            get(
                host.context,
                crate::current_invocation(),
                self.raw_id(),
                &mut value,
            )
        };
        (status == dragonfly_plugin_sys::DF_STATUS_OK).then_some(value)
    }

    pub fn set_time(&self, time: i64) {
        let Some(host) = crate::host_api() else {
            return;
        };
        let Some(set) = host.world_time_set else {
            return;
        };
        let _ = unsafe {
            set(
                host.context,
                crate::current_invocation(),
                self.raw_id(),
                time,
            )
        };
    }

    pub fn spawn(&self) -> Option<BlockPos> {
        let host = crate::host_api()?;
        let get = host.world_spawn_get?;
        let mut position = dragonfly_plugin_sys::DfBlockPos::default();
        let status = unsafe {
            get(
                host.context,
                crate::current_invocation(),
                self.raw_id(),
                &mut position,
            )
        };
        (status == dragonfly_plugin_sys::DF_STATUS_OK).then_some(position.into())
    }

    pub fn set_spawn(&self, position: BlockPos) {
        let Some(host) = crate::host_api() else {
            return;
        };
        let Some(set) = host.world_spawn_set else {
            return;
        };
        let _ = unsafe {
            set(
                host.context,
                crate::current_invocation(),
                self.raw_id(),
                position.into(),
            )
        };
    }

    pub fn spawn_entity<E: entity::Spawnable>(
        &self,
        descriptor: E,
        options: entity::SpawnOptions,
    ) -> Option<Entity> {
        let host = crate::host_api()?;
        let spawn = host.world_entity_spawn?;
        let encoded = descriptor.encode();
        let mut output = dragonfly_plugin_sys::DfEntityId::default();
        let status = encoded.with_raw(&options, |view| unsafe {
            spawn(
                host.context,
                crate::current_invocation(),
                self.raw_id(),
                view,
                &mut output,
            )
        })?;
        (status == dragonfly_plugin_sys::DF_STATUS_OK && output.generation != 0)
            .then(|| output.into())
    }

    pub fn add_particle(&self, position: Vec3, value: impl particle::Particle) {
        let Some(host) = crate::host_api() else {
            return;
        };
        let Some(add) = host.world_particle_add else {
            return;
        };
        let encoded = value.encode();
        let _ = encoded.with_raw(|view| unsafe {
            add(
                host.context,
                crate::current_invocation(),
                self.raw_id(),
                position.into(),
                view,
            )
        });
    }

    pub fn play_sound(&self, position: Vec3, value: impl sound::Sound) {
        let Some(host) = crate::host_api() else {
            return;
        };
        let Some(play) = host.world_sound_play else {
            return;
        };
        let encoded = value.encode();
        let _ = encoded.with_raw(|view| unsafe {
            play(
                host.context,
                crate::current_invocation(),
                self.raw_id(),
                position.into(),
                view,
            )
        });
    }

    pub fn entities(&self) -> Option<Vec<Entity>> {
        let host = crate::host_api()?;
        let list = host.world_entities?;
        let mut capacity = 0;
        for _ in 0..3 {
            let mut values = vec![dragonfly_plugin_sys::DfEntityId::default(); capacity];
            let mut raw = dragonfly_plugin_sys::DfEntityIdBuffer {
                data: if values.is_empty() {
                    core::ptr::null_mut()
                } else {
                    values.as_mut_ptr()
                },
                len: 0,
                capacity: values.len() as u64,
            };
            let status = unsafe {
                list(
                    host.context,
                    crate::current_invocation(),
                    self.raw_id(),
                    &mut raw,
                )
            };
            let length = usize::try_from(raw.len).ok()?;
            if length > 1 << 20 {
                return None;
            }
            if status == dragonfly_plugin_sys::DF_STATUS_OK {
                if length > values.len() {
                    return None;
                }
                values.truncate(length);
                return Some(values.into_iter().map(Into::into).collect());
            }
            if length <= values.len() {
                return None;
            }
            capacity = length;
        }
        None
    }

    pub fn players(&self) -> Option<Vec<Player>> {
        let host = crate::host_api()?;
        let list = host.world_players?;
        let mut capacity = 0;
        for _ in 0..3 {
            let mut values = vec![dragonfly_plugin_sys::DfPlayerId::default(); capacity];
            let mut raw = dragonfly_plugin_sys::DfPlayerIdBuffer {
                data: if values.is_empty() {
                    core::ptr::null_mut()
                } else {
                    values.as_mut_ptr()
                },
                len: 0,
                capacity: values.len() as u64,
            };
            let status = unsafe {
                list(
                    host.context,
                    crate::current_invocation(),
                    self.raw_id(),
                    &mut raw,
                )
            };
            let length = usize::try_from(raw.len).ok()?;
            if length > 1 << 20 {
                return None;
            }
            if status == dragonfly_plugin_sys::DF_STATUS_OK {
                if length > values.len() {
                    return None;
                }
                values.truncate(length);
                return Some(values.into_iter().map(Player::from_id).collect());
            }
            if length <= values.len() {
                return None;
            }
            capacity = length;
        }
        None
    }

    pub(crate) const fn raw_id(self) -> dragonfly_plugin_sys::DfWorldId {
        dragonfly_plugin_sys::DfWorldId { value: self.raw }
    }
}

impl From<BlockPos> for dragonfly_plugin_sys::DfBlockPos {
    fn from(value: BlockPos) -> Self {
        Self {
            x: value.x,
            y: value.y,
            z: value.z,
        }
    }
}
