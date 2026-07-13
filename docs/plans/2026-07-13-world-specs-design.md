# Configurable Persistent World Specs

Status: Implementation-ready

## Goal

Expose a generic, typed world specification that lets a plugin open a managed
persistent Dragonfly world at an explicit provider path with deterministic
open/create, read-only, save, random-tick, time, weather, and chunk-unload
policies.

The specification is immutable after the world is published. Reopening the
same world ID succeeds only when the normalized specification is identical.
All host and filesystem failures remain below the public Rust interface.

## Scope

This feature covers framework-owned custom persistent worlds backed by
Dragonfly's `mcdb` provider. It does not:

- replace the three core worlds after `server.Config.New`;
- change a live world's provider, read-only flag, random-tick rate, save
  interval, or chunk-unload interval;
- add memory, flat, custom provider, or custom generator adapters;
- keep chunks loaded forever, which Dragonfly requires a loader/viewer for;
- turn provider-path validation into a security sandbox for native plugins.

Core-world policies that Dragonfly only accepts during construction remain in
the generic Go server configuration until a pre-`Config.New` plugin bootstrap
phase exists.

## Chosen Approach

Three approaches were considered:

1. Add individual setters to `World`. This cannot correctly configure
   read-only, random ticks, save intervals, or chunk unloading because those
   values are captured by `world.Config.New`.
2. Put named world definitions in `server.toml`. This is useful operationally,
   but prevents a plugin from owning its world topology and requires Go-side
   adapter configuration for every application.
3. Pass one immutable creation specification through the host ABI. This keeps
   application policy in the plugin, maps directly to Dragonfly creation
   inputs, and makes duplicate-open behavior deterministic.

The third approach is selected. The existing `World::open(name, dimension)`
remains as a convenience wrapper using framework defaults and the provider path
derived from the world ID. New code that needs exact behavior uses
`World::open_with`.

## Rust API

The SDK exposes the following types from `dragonfly::world` and re-exports
`WorldSpec` beside `World`:

```rust
use std::time::Duration;

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum OpenMode {
    OpenOrCreate,
    OpenExisting,
    CreateNew,
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum SavePolicy {
    Automatic(Duration),
    Manual,
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum RandomTicks {
    Disabled,
    PerSubchunk(u32),
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum TimePolicy {
    Preserve,
    Cycle,
    Fixed(i64),
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum WeatherPolicy {
    Preserve,
    Cycle,
    Clear,
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum ChunkUnloadPolicy {
    After(Duration),
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct WorldSpec {
    // Private fields.
}

impl WorldSpec {
    pub fn persistent(provider_path: impl Into<String>) -> Self;
    pub fn dimension(self, dimension: Dimension) -> Self;
    pub fn open_mode(self, mode: OpenMode) -> Self;
    pub fn read_only(self, read_only: bool) -> Self;
    pub fn save(self, policy: SavePolicy) -> Self;
    pub fn random_ticks(self, policy: RandomTicks) -> Self;
    pub fn time(self, policy: TimePolicy) -> Self;
    pub fn weather(self, policy: WeatherPolicy) -> Self;
    pub fn chunk_unload(self, policy: ChunkUnloadPolicy) -> Self;
}

impl World {
    pub fn open(name: &str, dimension: Dimension) -> Option<Self>;
    pub fn open_with(name: &str, spec: &WorldSpec) -> Option<Self>;
}
```

Example:

```rust
let arena = World::open_with(
    "example:arena",
    &WorldSpec::persistent("arenas/example")
        .open_mode(OpenMode::OpenExisting)
        .read_only(true)
        .save(SavePolicy::Manual)
        .random_ticks(RandomTicks::Disabled)
        .time(TimePolicy::Fixed(6000))
        .weather(WeatherPolicy::Clear)
        .chunk_unload(ChunkUnloadPolicy::After(Duration::from_secs(120))),
)?;
```

`WorldSpec::persistent` defaults to:

| Policy | Default |
| --- | --- |
| dimension | `Dimension::Overworld` |
| open mode | `OpenMode::OpenOrCreate` |
| read-only | `false` |
| save | `SavePolicy::Automatic(Duration::from_secs(600))` |
| random ticks | `RandomTicks::PerSubchunk(3)` |
| time | `TimePolicy::Preserve` |
| weather | `WeatherPolicy::Preserve` |
| chunk unload | `ChunkUnloadPolicy::After(Duration::from_secs(120))` |

Read-only dominates saving during normalization. Calling `.read_only(true)`
always encodes and compares as manual save, regardless of builder call order;
any automatic-save interval is discarded. Calling `.read_only(false)` leaves
the selected save policy unchanged. This keeps the infallible builder
order-independent without exposing a configuration error type.

`World::open_with` returns `None` for invalid input, a missing/corrupt required
provider, a duplicate mismatch, or a host failure. It never exposes a native
status or host error. `World::open` derives `namespace/path` from a world ID and
uses the defaults above.

## Open/Create Semantics

`OpenMode` is explicit because `mcdb.Config.Open` creates directories and a new
database when the path is absent:

- `OpenOrCreate` accepts an absent path or opens an existing provider.
- `OpenExisting` requires the provider directory, `level.dat`, and
  `db/CURRENT` to be regular files before calling `mcdb.Config.Open`.
  Dragonfly then validates the level data and LevelDB.
- `CreateNew` requires that the provider path does not exist. Any existing
  file, directory, or symlink rejects the request.

The preflight checks provide deterministic intent; they do not replace
Dragonfly's provider validation and cannot eliminate all filesystem races.

## Provider Paths

Provider paths are UTF-8, slash-separated, relative to `worlds.directory`.
They are independent from the namespaced world ID.

Normalization:

1. Reject empty strings, NUL, backslashes, absolute paths, volume names, and
   paths longer than 4096 UTF-8 bytes.
2. Split on `/` and reject empty, `.`, and `..` components.
3. Join the components below the manager's absolute root with
   `filepath.FromSlash`.
4. Reject any existing symlink component below the configured root before
   opening or creating the provider.
5. Use `filepath.Rel` to prove the resulting path remains below the root.
6. Store the cleaned slash form as the immutable comparison key.

Two live or opening world IDs may not own the same normalized provider path.
The reservation is released after a failed open and only after an unloaded
world has closed its provider.

Native plugins are trusted code, so this containment rule prevents accidental
path escape and aliasing; it is not presented as an operating-system sandbox.

## Policy Mapping to Dragonfly

The Go host normalizes every policy before opening the provider. It never
passes zero values whose meaning Dragonfly rewrites implicitly.

| WorldSpec policy | Dragonfly mapping |
| --- | --- |
| writable + automatic save | `ReadOnly: false`, positive `SaveInterval` |
| writable + manual save | `ReadOnly: false`, `SaveInterval: -1` |
| read-only | `ReadOnly: true`, `SaveInterval: -1` |
| random ticks disabled | `RandomTickSpeed: -1` |
| random tick rate | positive `RandomTickSpeed` |
| chunk unload after | positive `ChunkUnloadInterval` |
| preserve time | leave provider settings unchanged |
| cycle time | set provider `Settings.TimeCycle = true` |
| fixed time | set `Settings.Time`, then `TimeCycle = false` |
| preserve weather | leave provider settings unchanged |
| cycle weather | set `Settings.WeatherCycle = true` |
| clear weather | zero rain/thunder timers and flags, then set `WeatherCycle = false` |

Dragonfly rewrites `SaveInterval == 0` to ten minutes,
`ChunkUnloadInterval <= 0` to two minutes, and `RandomTickSpeed == 0` to three.
The normalized host representation therefore uses positive durations/rates or
the explicit disabled/manual mappings above.

Time and weather policy is applied while holding `provider.Settings()` before
`world.Config.New`. This prevents the first viewer from observing provider
state for one tick before a post-creation correction.

## C Host ABI v16

The standalone world-spec branch advances host ABI v15 to v16. Plugin ABI
remains v3.

```c
#define DF_WORLD_OPEN_OR_CREATE 0u
#define DF_WORLD_OPEN_EXISTING 1u
#define DF_WORLD_CREATE_NEW 2u
#define DF_WORLD_SAVE_AUTOMATIC 0u
#define DF_WORLD_SAVE_MANUAL 1u
#define DF_WORLD_RANDOM_TICKS_DISABLED 0u
#define DF_WORLD_RANDOM_TICKS_PER_SUBCHUNK 1u
#define DF_WORLD_TIME_PRESERVE 0u
#define DF_WORLD_TIME_CYCLE 1u
#define DF_WORLD_TIME_FIXED 2u
#define DF_WORLD_WEATHER_PRESERVE 0u
#define DF_WORLD_WEATHER_CYCLE 1u
#define DF_WORLD_WEATHER_CLEAR 2u
#define DF_WORLD_CHUNK_UNLOAD_AFTER 0u

typedef struct DfWorldOpenSpecV1 {
    uint32_t struct_size;
    uint32_t dimension;
    DfStringView provider_path;
    uint64_t save_interval_milliseconds;
    uint64_t chunk_unload_interval_milliseconds;
    int64_t fixed_time;
    uint32_t open_mode;
    uint32_t save_policy;
    uint32_t random_tick_policy;
    uint32_t random_tick_rate;
    uint32_t time_policy;
    uint32_t weather_policy;
    uint32_t chunk_unload_policy;
    uint8_t read_only;
    uint8_t reserved[3];
} DfWorldOpenSpecV1;

typedef DfStatus (*DfHostWorldOpenSpecFn)(
    uint64_t context,
    DfInvocationId invocation,
    DfStringView name,
    const DfWorldOpenSpecV1 *spec,
    DfWorldId *world);
```

`DfWorldOpenSpecV1` is 80 bytes and aligned to 8 bytes. Important offsets are
`provider_path = 8`, `save_interval_milliseconds = 24`, `fixed_time = 40`,
`open_mode = 48`, and `read_only = 76`.

`world_open_spec` is appended to standalone `DfHostApiV16` at offset 448,
making that structure 456 bytes. The old `world_open` field remains and backs
the convenience API.

The host accepts `spec->struct_size >= sizeof(DfWorldOpenSpecV1)` and validates
the known reserved bytes as zero. Tags make every otherwise-zero field
unambiguous:

- null `spec` or output pointers fail;
- unknown dimension or policy tags fail;
- `read_only` must be 0 or 1 and reserved bytes must be zero;
- automatic save requires a positive representable duration;
- manual save requires a zero wire duration;
- read-only canonicalizes to manual save with a zero wire duration;
- disabled random ticks require rate zero;
- per-subchunk random ticks require `1..=math.MaxInt32`;
- fixed time accepts every `int64`; non-fixed policies require wire time zero;
- chunk unload requires a positive representable duration;
- provider path and world name must pass their independent bounds and UTF-8
  validation.

Milliseconds must fit Go's positive `time.Duration`:
`value <= uint64(math.MaxInt64 / int64(time.Millisecond))`.

### ABI conflict with transferable entities

The existing `feature/transferable-entities` worktree independently names its
layout host ABI v16 and places `world_entity_remove` at offset 448,
`world_entity_add` at 456, and `detached_entity_drop` at 464. Its v16 structure
is 472 bytes. That is a hard binary conflict with this standalone v16: the same
offset would be interpreted as different function signatures.

The two layouts must never be merged or published independently under the same
version. Whichever feature lands second must rebase and advance to host ABI
v17. If both are integrated before either v16 is published, the canonical
combined layout may preserve the transferable fields at 448/456/464 and append
`world_open_spec` at 472, for a 480-byte v16. The generator, C bridge, Rust sys
crate, runtime, macros, examples, and pinned consumers must all use one chosen
layout in the same integration milestone.

## Go Architecture

New focused files keep policy logic out of the existing world operation file:

- `framework/world_spec.go`: public Go policy values, normalization, path
  resolution, provider preflight, settings mutation, and `world.Config`
  construction.
- `framework/world_spec_test.go`: pure normalization and mapping tests.
- `framework/world_spec_open_test.go`: provider/open/create/duplicate and
  concurrency tests.
- `internal/native/world_spec.go`: wire enums/value and bounded conversion.
- `internal/native/world_spec_test.go`: native validation tests.

`managedWorld` stores its normalized persistent spec. `WorldManager` adds:

```go
type worldOpening struct {
    spec normalizedWorldSpec
    done chan struct{}
    id   native.WorldID
    err  error
}

openings      map[WorldID]*worldOpening
providerPaths map[string]WorldID
```

Open flow:

1. Validate the world ID and normalize the requested spec without mutation.
2. Under `WorldManager.mu`, return an existing ID only for exact normalized
   equality; reject a mismatch.
3. If an identical open is in progress, wait for its `done` channel without
   holding the manager lock and return the same result.
4. Reserve both the world ID and provider path, then release the manager lock.
5. Perform open-mode preflight, open `mcdb`, mutate settings, construct the
   Dragonfly world, and install its handler before publication.
6. Publish one stable world ID, complete the opening, and wake waiters.
7. On every failure, close owned resources and release both reservations.

Opening a provider is synchronous, matching both the existing `World::open`
surface and direct Dragonfly use. Plugins should normally do lifecycle I/O in
`on_enable`, but the host does not add an artificial callback-only rejection.
The invocation token is irrelevant to provider ownership and is not retained.

## Duplicate and Unload Semantics

- Same world ID + identical normalized spec: same stable `World` handle.
- Same world ID + different spec: reject without changing the live world.
- Different world ID + same provider path: reject.
- Concurrent identical opens: exactly one provider/world, same handle for all
  callers.
- Concurrent mismatched opens: the first reservation wins; mismatches reject.
- Failed open: no published world and no leaked ID/path reservation.
- Unload: reject core or occupied worlds as today, close the provider, then
  release the provider path. Reopen receives a fresh monotonic world handle.

## Failure Handling

Detailed errors remain in Go for logs and tests. The native export converts
them to `DF_STATUS_ERROR`; Rust converts that to `None`. No transport error or
host error type becomes public SDK surface.

Provider/world cleanup follows exactly-once ownership:

- failure before `mcdb.Open`: no provider to close;
- failure after provider open but before publication: close the provider/world;
- duplicate publication race: close the losing world before returning the
  winner only when the normalized spec is identical;
- unload releases its path only after `World.Close` has closed the provider.

## Verification

Required evidence:

1. Rust constructor tests prove every policy encodes exact tags/values and
   invalid zero durations/rates never call the host.
2. C/Rust layout assertions cover all specified sizes and offsets.
3. Native tests reject nulls, malformed tags, non-zero reserved bytes,
   contradictory policies, oversized durations, invalid UTF-8, and path
   overflow.
4. Framework tests prove default normalization avoids Dragonfly's zero-value
   rewrites.
5. Provider tests prove `OpenExisting`, `CreateNew`, and `OpenOrCreate` differ.
6. Settings tests prove fixed time and clear weather exist before
   `world.Config.New`.
7. Duplicate and race tests prove one provider owner and immutable specs.
8. Read-only/manual/automatic and random-tick mappings are asserted on the
   exact `world.Config` passed to construction.
9. Full Go race tests, generated checks, Rust workspace tests, examples, and a
   minimal consumer build pass.
