use super::*;
use std::sync::atomic::{AtomicU8, AtomicUsize, Ordering as AtomicOrdering};
use std::time::Duration;

static SNAPSHOT_CALLS: AtomicUsize = AtomicUsize::new(0);
static MALFORMED_MODE: AtomicU8 = AtomicU8::new(0);
static CLEAR_CALL: std::sync::Mutex<Option<RecordedClear>> = std::sync::Mutex::new(None);

#[derive(Debug, PartialEq)]
struct RecordedClear {
    context: u64,
    invocation: u64,
    player_bytes: [u8; 16],
    player_generation: u64,
}

unsafe extern "C" fn snapshot_effects(
    context: u64,
    invocation: dragonfly_plugin_sys::DfInvocationId,
    player: dragonfly_plugin_sys::DfPlayerId,
    output: *mut dragonfly_plugin_sys::DfEffectBuffer,
) -> dragonfly_plugin_sys::DfStatus {
    assert_eq!(context, 27);
    assert_eq!(invocation, 31);
    assert_eq!(player.bytes, [9; 16]);
    assert_eq!(player.generation, 14);
    SNAPSHOT_CALLS.fetch_add(1, AtomicOrdering::Relaxed);
    let output = unsafe { &mut *output };
    output.len = 2;
    if output.capacity < 2 {
        return dragonfly_plugin_sys::DF_STATUS_ERROR;
    }
    assert!(!output.data.is_null());
    let values = unsafe { std::slice::from_raw_parts_mut(output.data, 2) };
    values[0] = dragonfly_plugin_sys::DfEffectView {
        effect_type: -7,
        level: 2,
        duration_milliseconds: 1_500,
        potency: 1.0,
        mode: dragonfly_plugin_sys::DF_EFFECT_MODE_AMBIENT,
        particles_hidden: 1,
    };
    values[1] = dragonfly_plugin_sys::DfEffectView {
        effect_type: 30,
        level: 1,
        duration_milliseconds: 0,
        potency: 1.0,
        mode: dragonfly_plugin_sys::DF_EFFECT_MODE_INFINITE,
        particles_hidden: 0,
    };
    dragonfly_plugin_sys::DF_STATUS_OK
}

unsafe extern "C" fn clear_effects(
    context: u64,
    invocation: dragonfly_plugin_sys::DfInvocationId,
    player: dragonfly_plugin_sys::DfPlayerId,
) -> dragonfly_plugin_sys::DfStatus {
    *CLEAR_CALL.lock().unwrap() = Some(RecordedClear {
        context,
        invocation,
        player_bytes: player.bytes,
        player_generation: player.generation,
    });
    dragonfly_plugin_sys::DF_STATUS_ERROR
}

unsafe extern "C" fn malformed_snapshot(
    _context: u64,
    _invocation: dragonfly_plugin_sys::DfInvocationId,
    _player: dragonfly_plugin_sys::DfPlayerId,
    output: *mut dragonfly_plugin_sys::DfEffectBuffer,
) -> dragonfly_plugin_sys::DfStatus {
    let output = unsafe { &mut *output };
    SNAPSHOT_CALLS.fetch_add(1, AtomicOrdering::Relaxed);
    match MALFORMED_MODE.load(AtomicOrdering::Relaxed) {
        0 => {
            output.len = 1;
            dragonfly_plugin_sys::DF_STATUS_OK
        }
        1 => {
            output.len = 0;
            dragonfly_plugin_sys::DF_STATUS_ERROR
        }
        2 => {
            output.len = output.capacity + 1;
            dragonfly_plugin_sys::DF_STATUS_ERROR
        }
        3 | 5..=9 => {
            output.len = 1;
            if output.capacity == 0 {
                return dragonfly_plugin_sys::DF_STATUS_ERROR;
            }
            let mut value = dragonfly_plugin_sys::DfEffectView {
                effect_type: 1,
                level: 1,
                duration_milliseconds: 1,
                potency: 1.0,
                mode: dragonfly_plugin_sys::DF_EFFECT_MODE_TIMED,
                particles_hidden: 0,
            };
            match MALFORMED_MODE.load(AtomicOrdering::Relaxed) {
                3 => value.potency = 0.5,
                5 => value.level = 0,
                6 => value.particles_hidden = 2,
                7 => value.mode = dragonfly_plugin_sys::DF_EFFECT_MODE_INSTANT,
                8 => {
                    value.mode = dragonfly_plugin_sys::DF_EFFECT_MODE_INFINITE;
                    value.duration_milliseconds = 1;
                }
                9 => value.mode = 99,
                _ => unreachable!(),
            }
            unsafe { *output.data = value };
            dragonfly_plugin_sys::DF_STATUS_OK
        }
        4 => {
            output.len = 257;
            dragonfly_plugin_sys::DF_STATUS_ERROR
        }
        _ => unreachable!(),
    }
}

fn player() -> Player {
    Player::from_id(dragonfly_plugin_sys::DfPlayerId {
        bytes: [9; 16],
        generation: 14,
    })
}

#[test]
fn player_effects_retry_and_preserve_typed_snapshot_values() {
    let _host_guard = TEST_HOST_LOCK.lock().unwrap();
    SNAPSHOT_CALLS.store(0, AtomicOrdering::Relaxed);
    let mut host: dragonfly_plugin_sys::DfHostApiV20 = unsafe { core::mem::zeroed() };
    host.context = 27;
    host.player_effects = Some(snapshot_effects);
    unsafe { install_host(&host) };

    let effects = with_invocation(31, || player().effects());
    unsafe { install_host(core::ptr::null()) };

    assert_eq!(SNAPSHOT_CALLS.load(AtomicOrdering::Relaxed), 2);
    assert_eq!(effects.len(), 2);
    assert_eq!(effects[0].type_id(), -7);
    assert_eq!(effects[0].level(), 2);
    assert_eq!(effects[0].duration(), Duration::from_millis(1_500));
    assert!(effects[0].ambient());
    assert!(!effects[0].infinite());
    assert!(effects[0].particles_hidden());
    assert_eq!(effects[1].type_id(), 30);
    assert!(effects[1].infinite());
    assert_eq!(effects[1].duration(), Duration::ZERO);
}

#[test]
fn player_clear_effects_forwards_identity_and_hides_host_status() {
    let _host_guard = TEST_HOST_LOCK.lock().unwrap();
    *CLEAR_CALL.lock().unwrap() = None;
    let mut host: dragonfly_plugin_sys::DfHostApiV20 = unsafe { core::mem::zeroed() };
    host.context = 41;
    host.player_effects_clear = Some(clear_effects);
    unsafe { install_host(&host) };

    with_invocation(43, || player().clear_effects());
    unsafe { install_host(core::ptr::null()) };

    assert_eq!(
        CLEAR_CALL.lock().unwrap().take(),
        Some(RecordedClear {
            context: 41,
            invocation: 43,
            player_bytes: [9; 16],
            player_generation: 14,
        })
    );
}

#[test]
fn player_effects_fail_closed_for_invalid_protocol_and_values() {
    let _host_guard = TEST_HOST_LOCK.lock().unwrap();
    let mut host: dragonfly_plugin_sys::DfHostApiV20 = unsafe { core::mem::zeroed() };
    host.player_effects = Some(malformed_snapshot);
    unsafe { install_host(&host) };

    for (mode, expected_calls) in [
        (0, 1),
        (1, 1),
        (2, 3),
        (3, 2),
        (4, 1),
        (5, 2),
        (6, 2),
        (7, 2),
        (8, 2),
        (9, 2),
    ] {
        MALFORMED_MODE.store(mode, AtomicOrdering::Relaxed);
        SNAPSHOT_CALLS.store(0, AtomicOrdering::Relaxed);
        assert!(player().effects().is_empty(), "mode {mode}");
        assert_eq!(
            SNAPSHOT_CALLS.load(AtomicOrdering::Relaxed),
            expected_calls,
            "mode {mode}"
        );
    }
    unsafe { install_host(core::ptr::null()) };
}

#[test]
fn missing_effect_host_calls_fail_closed() {
    let _host_guard = TEST_HOST_LOCK.lock().unwrap();
    unsafe { install_host(core::ptr::null()) };
    assert!(player().effects().is_empty());
    player().clear_effects();
}
