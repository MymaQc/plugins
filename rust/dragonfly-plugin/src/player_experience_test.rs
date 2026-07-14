use super::*;

#[derive(Clone, Copy, Debug)]
struct ExperienceCall {
    context: u64,
    invocation: u64,
    player: dragonfly_plugin_sys::DfPlayerId,
    level: i32,
    progress: f64,
}

static CALLS: std::sync::Mutex<Vec<ExperienceCall>> = std::sync::Mutex::new(Vec::new());

unsafe extern "C" fn set_experience(
    context: u64,
    invocation: dragonfly_plugin_sys::DfInvocationId,
    player: dragonfly_plugin_sys::DfPlayerId,
    level: i32,
    progress: f64,
) -> dragonfly_plugin_sys::DfStatus {
    CALLS.lock().unwrap().push(ExperienceCall {
        context,
        invocation,
        player,
        level,
        progress,
    });
    dragonfly_plugin_sys::DF_STATUS_ERROR
}

#[test]
fn batches_valid_experience_and_ignores_invalid_values() {
    let _host_guard = crate::TEST_HOST_LOCK.lock().unwrap();
    CALLS.lock().unwrap().clear();
    let mut host: dragonfly_plugin_sys::DfHostApiV20 = unsafe { core::mem::zeroed() };
    host.context = 41;
    host.player_experience_set = Some(set_experience);
    unsafe { crate::install_host(&host) };

    let raw_player = dragonfly_plugin_sys::DfPlayerId {
        bytes: [7; 16],
        generation: 13,
    };
    let player = Player::from_id(raw_player);
    with_invocation(17, || {
        player.set_experience(-1, 0.5);
        player.set_experience(1, f64::NAN);
        player.set_experience(1, f64::INFINITY);
        player.set_experience(1, -0.1);
        player.set_experience(1, 1.1);
        player.set_experience(30, 0.75);
    });
    unsafe { crate::install_host(core::ptr::null()) };

    let calls = CALLS.lock().unwrap();
    assert_eq!(calls.len(), 1);
    let call = calls[0];
    assert_eq!(call.context, 41);
    assert_eq!(call.invocation, 17);
    assert_eq!(call.player.bytes, raw_player.bytes);
    assert_eq!(call.player.generation, raw_player.generation);
    assert_eq!(call.level, 30);
    assert_eq!(call.progress, 0.75);
}
