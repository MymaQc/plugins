use super::*;

#[derive(Clone, Copy, Debug, PartialEq)]
struct RecordedTransfer {
    context: u64,
    invocation: u64,
    player_bytes: [u8; 16],
    player_generation: u64,
    world: u64,
    position: Vec3,
}

static RECORDED_TRANSFER: std::sync::Mutex<Option<RecordedTransfer>> = std::sync::Mutex::new(None);

unsafe extern "C" fn record_transfer(
    context: u64,
    invocation: dragonfly_plugin_sys::DfInvocationId,
    player: dragonfly_plugin_sys::DfPlayerId,
    world: dragonfly_plugin_sys::DfWorldId,
    position: dragonfly_plugin_sys::DfVec3,
) -> dragonfly_plugin_sys::DfStatus {
    *RECORDED_TRANSFER.lock().unwrap() = Some(RecordedTransfer {
        context,
        invocation,
        player_bytes: player.bytes,
        player_generation: player.generation,
        world: world.value,
        position: Vec3 {
            x: position.x,
            y: position.y,
            z: position.z,
        },
    });
    dragonfly_plugin_sys::DF_STATUS_ERROR
}

#[test]
fn player_transfer_sends_typed_handles_and_hides_host_status() {
    let _host_guard = TEST_HOST_LOCK.lock().unwrap();
    *RECORDED_TRANSFER.lock().unwrap() = None;
    let mut host: dragonfly_plugin_sys::DfHostApiV18 = unsafe { core::mem::zeroed() };
    host.context = 27;
    host.player_transfer = Some(record_transfer);
    unsafe { install_host(&host) };

    let raw_player = dragonfly_plugin_sys::DfPlayerId {
        bytes: [9; 16],
        generation: 14,
    };
    let player = Player::from_id(raw_player);
    with_invocation(31, || {
        player.transfer(
            World::from_valid_raw(42),
            Vec3 {
                x: 1.5,
                y: 72.0,
                z: -8.25,
            },
        );
    });
    unsafe { install_host(core::ptr::null()) };

    assert_eq!(
        RECORDED_TRANSFER.lock().unwrap().take(),
        Some(RecordedTransfer {
            context: 27,
            invocation: 31,
            player_bytes: raw_player.bytes,
            player_generation: raw_player.generation,
            world: 42,
            position: Vec3 {
                x: 1.5,
                y: 72.0,
                z: -8.25,
            },
        })
    );
}
