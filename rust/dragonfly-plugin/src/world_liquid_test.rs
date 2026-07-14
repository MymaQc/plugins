use super::*;

static RECORDED_READ: std::sync::Mutex<Option<(u64, u64, BlockPos)>> = std::sync::Mutex::new(None);

unsafe extern "C" fn read_liquid(
    context: u64,
    _invocation: dragonfly_plugin_sys::DfInvocationId,
    world: dragonfly_plugin_sys::DfWorldId,
    position: dragonfly_plugin_sys::DfBlockPos,
    output: *mut dragonfly_plugin_sys::DfBlockData,
) -> dragonfly_plugin_sys::DfStatus {
    let Some(output) = (unsafe { output.as_mut() }) else {
        return dragonfly_plugin_sys::DF_STATUS_ERROR;
    };
    let position = BlockPos {
        x: position.x,
        y: position.y,
        z: position.z,
    };
    *RECORDED_READ.lock().unwrap() = Some((context, world.value, position));

    let liquid: block::Block = block::Water::new(block::LiquidDepth::Value0).into();
    let identifier = liquid.identifier().as_bytes();
    let properties = liquid.properties_nbt().unwrap();
    output.identifier.len = identifier.len() as u64;
    output.properties_nbt.len = properties.len() as u64;
    if output.identifier.capacity < identifier.len() as u64
        || output.properties_nbt.capacity < properties.len() as u64
        || output.identifier.data.is_null()
        || output.properties_nbt.data.is_null()
    {
        return dragonfly_plugin_sys::DF_STATUS_ERROR;
    }
    unsafe {
        core::ptr::copy_nonoverlapping(
            identifier.as_ptr(),
            output.identifier.data,
            identifier.len(),
        );
        core::ptr::copy_nonoverlapping(
            properties.as_ptr(),
            output.properties_nbt.data,
            properties.len(),
        );
    }
    dragonfly_plugin_sys::DF_STATUS_OK
}

#[test]
fn reads_and_decodes_typed_water_from_appended_host_function() {
    let _host_guard = crate::TEST_HOST_LOCK.lock().unwrap();
    *RECORDED_READ.lock().unwrap() = None;
    let mut host: dragonfly_plugin_sys::DfHostApiV20 = unsafe { core::mem::zeroed() };
    host.context = 29;
    host.world_liquid_get = Some(read_liquid);
    unsafe { crate::install_host(&host) };

    let position = BlockPos { x: -4, y: 63, z: 9 };
    let liquid = World::from_valid_raw(31).liquid(position);
    unsafe { crate::install_host(core::ptr::null()) };

    let liquid = liquid.expect("liquid read should decode");
    assert_eq!(liquid.identifier(), "minecraft:water");
    assert_eq!(
        liquid.property("liquid_depth"),
        Some(&block::Property::Int32(0))
    );
    assert_eq!(
        RECORDED_READ.lock().unwrap().take(),
        Some((29, 31, position))
    );
}
