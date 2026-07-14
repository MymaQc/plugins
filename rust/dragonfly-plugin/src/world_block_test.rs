use super::*;

#[derive(Debug)]
struct RecordedBlock {
    context: u64,
    world: u64,
    position: BlockPos,
    identifier: String,
    properties: Vec<u8>,
}

static RECORDED_BLOCK: std::sync::Mutex<Option<RecordedBlock>> = std::sync::Mutex::new(None);

unsafe extern "C" fn record_block(
    context: u64,
    _invocation: dragonfly_plugin_sys::DfInvocationId,
    world: dragonfly_plugin_sys::DfWorldId,
    position: dragonfly_plugin_sys::DfBlockPos,
    block: *const dragonfly_plugin_sys::DfBlockView,
) -> dragonfly_plugin_sys::DfStatus {
    let Some(block) = (unsafe { block.as_ref() }) else {
        return dragonfly_plugin_sys::DF_STATUS_ERROR;
    };
    let identifier = unsafe {
        core::slice::from_raw_parts(block.identifier.data, block.identifier.len as usize)
    };
    let properties = unsafe {
        core::slice::from_raw_parts(block.properties_nbt.data, block.properties_nbt.len as usize)
    };
    let Ok(identifier) = core::str::from_utf8(identifier) else {
        return dragonfly_plugin_sys::DF_STATUS_ERROR;
    };
    let Ok(mut recorded) = RECORDED_BLOCK.lock() else {
        return dragonfly_plugin_sys::DF_STATUS_ERROR;
    };
    *recorded = Some(RecordedBlock {
        context,
        world: world.value,
        position: BlockPos {
            x: position.x,
            y: position.y,
            z: position.z,
        },
        identifier: identifier.to_owned(),
        properties: properties.to_vec(),
    });
    dragonfly_plugin_sys::DF_STATUS_OK
}

#[test]
fn set_block_sends_typed_identifier_and_properties_to_host() {
    let _host_guard = crate::TEST_HOST_LOCK.lock().unwrap();
    *RECORDED_BLOCK.lock().unwrap() = None;
    let mut host: dragonfly_plugin_sys::DfHostApiV20 = unsafe { core::mem::zeroed() };
    host.context = 17;
    host.world_block_set = Some(record_block);
    unsafe { crate::install_host(&host) };

    let position = BlockPos { x: 3, y: 64, z: -8 };
    World::from_valid_raw(23).set_block(position, block::OakLog::new(block::PillarAxis::Y));
    unsafe { crate::install_host(core::ptr::null()) };

    let recorded = RECORDED_BLOCK.lock().unwrap().take().unwrap();
    assert_eq!(recorded.context, 17);
    assert_eq!(recorded.world, 23);
    assert_eq!(recorded.position, position);
    assert_eq!(recorded.identifier, "minecraft:oak_log");
    let decoded = block::Block::from_nbt(recorded.identifier, &recorded.properties).unwrap();
    assert_eq!(
        decoded.property("pillar_axis"),
        Some(&block::Property::String("y".to_owned()))
    );
}
