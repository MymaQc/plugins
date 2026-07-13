use super::*;
use std::sync::atomic::{AtomicI32, AtomicUsize, Ordering};

#[derive(Clone, Debug, Eq, PartialEq)]
struct RecordedSpec {
    provider_path: String,
    struct_size: u32,
    dimension: u32,
    save_milliseconds: u64,
    unload_milliseconds: u64,
    fixed_time: i64,
    open_mode: u32,
    save_policy: u32,
    random_tick_policy: u32,
    random_tick_rate: u32,
    time_policy: u32,
    weather_policy: u32,
    unload_policy: u32,
    read_only: u8,
    reserved: [u8; 3],
}

static CALLS: AtomicUsize = AtomicUsize::new(0);
static RESULT: AtomicI32 = AtomicI32::new(dragonfly_plugin_sys::DF_STATUS_OK);
static RECORDED: std::sync::Mutex<Option<RecordedSpec>> = std::sync::Mutex::new(None);

unsafe extern "C" fn record_open(
    _context: u64,
    _invocation: dragonfly_plugin_sys::DfInvocationId,
    _name: dragonfly_plugin_sys::DfStringView,
    spec: *const dragonfly_plugin_sys::DfWorldOpenSpecV1,
    output: *mut dragonfly_plugin_sys::DfWorldId,
) -> dragonfly_plugin_sys::DfStatus {
    CALLS.fetch_add(1, Ordering::Relaxed);
    let Some(spec) = (unsafe { spec.as_ref() }) else {
        return dragonfly_plugin_sys::DF_STATUS_ERROR;
    };
    let provider_path = unsafe {
        core::slice::from_raw_parts(
            spec.provider_path.data,
            usize::try_from(spec.provider_path.len).unwrap_or(0),
        )
    };
    let Ok(provider_path) = core::str::from_utf8(provider_path) else {
        return dragonfly_plugin_sys::DF_STATUS_ERROR;
    };
    *RECORDED.lock().unwrap() = Some(RecordedSpec {
        provider_path: provider_path.to_owned(),
        struct_size: spec.struct_size,
        dimension: spec.dimension,
        save_milliseconds: spec.save_interval_milliseconds,
        unload_milliseconds: spec.chunk_unload_interval_milliseconds,
        fixed_time: spec.fixed_time,
        open_mode: spec.open_mode,
        save_policy: spec.save_policy,
        random_tick_policy: spec.random_tick_policy,
        random_tick_rate: spec.random_tick_rate,
        time_policy: spec.time_policy,
        weather_policy: spec.weather_policy,
        unload_policy: spec.chunk_unload_policy,
        read_only: spec.read_only,
        reserved: spec.reserved,
    });
    let status = RESULT.load(Ordering::Relaxed);
    if status == dragonfly_plugin_sys::DF_STATUS_OK
        && let Some(output) = unsafe { output.as_mut() }
    {
        output.value = 41;
    }
    status
}

fn install_recording_host(
    result: dragonfly_plugin_sys::DfStatus,
) -> dragonfly_plugin_sys::DfHostApiV16 {
    CALLS.store(0, Ordering::Relaxed);
    RESULT.store(result, Ordering::Relaxed);
    *RECORDED.lock().unwrap() = None;
    let mut host: dragonfly_plugin_sys::DfHostApiV16 = unsafe { core::mem::zeroed() };
    host.world_open_spec = Some(record_open);
    host
}

#[test]
fn defaults_encode_explicit_dragonfly_values() {
    let spec = WorldSpec::persistent("arenas/default");
    let raw = spec.encode().unwrap();
    assert_eq!(raw.struct_size, 80);
    assert_eq!(
        raw.dimension,
        dragonfly_plugin_sys::DF_WORLD_DIMENSION_OVERWORLD
    );
    assert_eq!(raw.open_mode, dragonfly_plugin_sys::DF_WORLD_OPEN_OR_CREATE);
    assert_eq!(
        raw.save_policy,
        dragonfly_plugin_sys::DF_WORLD_SAVE_AUTOMATIC
    );
    assert_eq!(raw.save_interval_milliseconds, 600_000);
    assert_eq!(
        raw.random_tick_policy,
        dragonfly_plugin_sys::DF_WORLD_RANDOM_TICKS_PER_SUBCHUNK
    );
    assert_eq!(raw.random_tick_rate, 3);
    assert_eq!(
        raw.time_policy,
        dragonfly_plugin_sys::DF_WORLD_TIME_PRESERVE
    );
    assert_eq!(
        raw.weather_policy,
        dragonfly_plugin_sys::DF_WORLD_WEATHER_PRESERVE
    );
    assert_eq!(raw.chunk_unload_interval_milliseconds, 120_000);
    assert_eq!(raw.read_only, 0);
    assert_eq!(raw.reserved, [0; 3]);
}

#[test]
fn read_only_switches_to_manual_save_independent_of_order() {
    let first = WorldSpec::persistent("arenas/readonly")
        .read_only(true)
        .save(SavePolicy::Automatic(Duration::from_secs(5)))
        .encode()
        .unwrap();
    let second = WorldSpec::persistent("arenas/readonly")
        .save(SavePolicy::Automatic(Duration::from_secs(30)))
        .read_only(true)
        .encode()
        .unwrap();
    assert_eq!(
        first.save_policy,
        dragonfly_plugin_sys::DF_WORLD_SAVE_MANUAL
    );
    assert_eq!(first.save_interval_milliseconds, 0);
    assert_eq!(second.save_policy, first.save_policy);
    assert_eq!(
        second.save_interval_milliseconds,
        first.save_interval_milliseconds
    );
}

#[test]
fn every_policy_encodes_its_exact_tag() {
    let raw = WorldSpec::persistent("arenas/policies")
        .dimension(Dimension::End)
        .open_mode(OpenMode::CreateNew)
        .save(SavePolicy::Manual)
        .random_ticks(RandomTicks::Disabled)
        .time(TimePolicy::Fixed(i64::MIN))
        .weather(WeatherPolicy::Clear)
        .chunk_unload(ChunkUnloadPolicy::After(Duration::from_millis(1)))
        .encode()
        .unwrap();
    assert_eq!(raw.dimension, dragonfly_plugin_sys::DF_WORLD_DIMENSION_END);
    assert_eq!(raw.open_mode, dragonfly_plugin_sys::DF_WORLD_CREATE_NEW);
    assert_eq!(raw.save_policy, dragonfly_plugin_sys::DF_WORLD_SAVE_MANUAL);
    assert_eq!(
        raw.random_tick_policy,
        dragonfly_plugin_sys::DF_WORLD_RANDOM_TICKS_DISABLED
    );
    assert_eq!(raw.random_tick_rate, 0);
    assert_eq!(raw.time_policy, dragonfly_plugin_sys::DF_WORLD_TIME_FIXED);
    assert_eq!(raw.fixed_time, i64::MIN);
    assert_eq!(
        raw.weather_policy,
        dragonfly_plugin_sys::DF_WORLD_WEATHER_CLEAR
    );
    assert_eq!(
        raw.chunk_unload_policy,
        dragonfly_plugin_sys::DF_WORLD_CHUNK_UNLOAD_AFTER
    );

    let alternate = WorldSpec::persistent("arenas/policies")
        .dimension(Dimension::Nether)
        .open_mode(OpenMode::OpenExisting)
        .random_ticks(RandomTicks::PerSubchunk(1))
        .time(TimePolicy::Cycle)
        .weather(WeatherPolicy::Cycle)
        .encode()
        .unwrap();
    assert_eq!(
        alternate.open_mode,
        dragonfly_plugin_sys::DF_WORLD_OPEN_EXISTING
    );
    assert_eq!(
        alternate.time_policy,
        dragonfly_plugin_sys::DF_WORLD_TIME_CYCLE
    );
    assert_eq!(
        alternate.weather_policy,
        dragonfly_plugin_sys::DF_WORLD_WEATHER_CYCLE
    );
}

#[test]
fn invalid_rate_or_duration_does_not_call_host() {
    let _guard = crate::TEST_HOST_LOCK.lock().unwrap();
    let host = install_recording_host(dragonfly_plugin_sys::DF_STATUS_OK);
    unsafe { crate::install_host(&host) };
    assert!(
        World::open_with(
            "example:invalid",
            &WorldSpec::persistent("arenas/invalid").random_ticks(RandomTicks::PerSubchunk(0)),
        )
        .is_none()
    );
    assert!(
        World::open_with(
            "example:invalid",
            &WorldSpec::persistent("arenas/invalid")
                .save(SavePolicy::Automatic(Duration::from_nanos(1))),
        )
        .is_none()
    );
    unsafe { crate::install_host(core::ptr::null()) };
    assert_eq!(CALLS.load(Ordering::Relaxed), 0);
}

#[test]
fn open_with_returns_none_for_host_rejection() {
    let _guard = crate::TEST_HOST_LOCK.lock().unwrap();
    let host = install_recording_host(dragonfly_plugin_sys::DF_STATUS_ERROR);
    unsafe { crate::install_host(&host) };
    let world = World::open_with(
        "example:rejected",
        &WorldSpec::persistent("arenas/rejected"),
    );
    unsafe { crate::install_host(core::ptr::null()) };
    assert!(world.is_none());
    assert_eq!(CALLS.load(Ordering::Relaxed), 1);
}

#[test]
fn open_with_passes_owned_path_and_returns_handle() {
    let _guard = crate::TEST_HOST_LOCK.lock().unwrap();
    let host = install_recording_host(dragonfly_plugin_sys::DF_STATUS_OK);
    unsafe { crate::install_host(&host) };
    let world = World::open_with("example:arena", &WorldSpec::persistent("arenas/one"));
    unsafe { crate::install_host(core::ptr::null()) };
    assert_eq!(world.map(|world| world.raw), Some(41));
    assert_eq!(
        RECORDED.lock().unwrap().as_ref().unwrap().provider_path,
        "arenas/one"
    );
}
