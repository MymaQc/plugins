# Configurable Persistent World Specs Implementation Plan

**Goal:** Add immutable typed specifications for framework-owned persistent worlds, including explicit provider and runtime policies, across Go, C, and Rust.

**Architecture:** Normalize and reserve world specifications in `WorldManager` before opening an `mcdb` provider, then pass exact creation-time values into Dragonfly. A versioned `DfWorldOpenSpecV1` is appended to standalone host ABI v16, while the Rust SDK exposes policy enums and keeps native failures private behind `Option<World>`.

**Tech Stack:** Go 1.26, Dragonfly v0.11, cgo/C11, Rust 1.96 edition 2024, generated native ABI, `mcdb`, standard-library filesystem and concurrency primitives.

## Global Constraints

- Work from `master@b5d33d97fec35ff59ea1843d9266edbed8953631` unless the ABI coordination gate requires a rebase.
- Keep this a generic framework API; no application names, paths, modes, or policies enter production code.
- Do not retain `world.Tx`, `world.Entity`, or `*player.Player` across callbacks.
- Do not expose host transport errors in the Rust public interface.
- Existing `World::open(name, dimension)` remains as the default convenience API.
- Tests live in dedicated test files, not inline in production modules.
- Provider paths are root-relative UTF-8 slash paths and are not an OS sandbox for trusted native plugins.
- Core-world creation policy remains out of scope.
- Commit each task after its focused verification passes.
- Before merging to master, resolve the hard host ABI v16 conflict with `feature/transferable-entities`; never publish both independent v16 layouts.

---

### Task 1: Define and Normalize the Go World Policy Model

**Files:**
- Create: `framework/world_spec.go`
- Create: `framework/world_spec_test.go`

**Interfaces:**
- Consumes: Dragonfly `world.Config`, `world.Settings`, `world.Dimension`, and existing `WorldID` validation.
- Produces: `WorldSpec`, policy enums, `normalizedWorldSpec`, `normalizeWorldSpec`, `normalizedWorldSpec.config`, and `normalizedWorldSpec.applySettings` for Task 2.

- [ ] **Step 1: Write failing default and validation tests**

Create table-driven tests in `framework/world_spec_test.go`. The default mapping test must inspect the config before `Config.New`, so Dragonfly cannot rewrite an accidental zero:

```go
func TestNormalizeWorldSpecUsesExplicitDragonflyDefaults(t *testing.T) {
    spec, err := normalizeWorldSpec(t.TempDir(), WorldSpec{
        ProviderPath: "arenas/one",
        Dimension: WorldDimensionOverworld,
        OpenMode: WorldOpenOrCreate,
        Save: WorldSaveAutomatic,
        SaveInterval: 10 * time.Minute,
        RandomTicks: WorldRandomTicksPerSubchunk,
        RandomTickRate: 3,
        Time: WorldTimePreserve,
        Weather: WorldWeatherPreserve,
        ChunkUnload: WorldChunkUnloadAfter,
        ChunkUnloadAfter: 2 * time.Minute,
    })
    if err != nil {
        t.Fatal(err)
    }
    config := spec.config(nil, nil, nil, nil)
    if config.SaveInterval != 10*time.Minute ||
        config.ChunkUnloadInterval != 2*time.Minute ||
        config.RandomTickSpeed != 3 {
        t.Fatalf("config = %#v", config)
    }
}
```

Cover empty/absolute/backslash/dot/dot-dot paths, paths over 4096 bytes,
automatic save with zero duration, manual save with non-zero duration,
read-only canonicalization regardless of builder order, disabled random ticks with non-zero rate, rate zero
or above `math.MaxInt32`, non-fixed time with non-zero fixed value, and
non-positive chunk-unload durations.

- [ ] **Step 2: Run the focused tests and confirm the API is absent**

Run:

```bash
go test ./framework -run 'TestNormalizeWorldSpec' -count=1
```

Expected: compile failure because `WorldSpec` and `normalizeWorldSpec` do not exist.

- [ ] **Step 3: Implement policy types and canonical normalization**

Define these exact public Go values in `framework/world_spec.go`:

```go
type WorldDimension uint32

const (
    WorldDimensionOverworld WorldDimension = iota
    WorldDimensionNether
    WorldDimensionEnd
)

type WorldOpenMode uint32

const (
    WorldOpenOrCreate WorldOpenMode = iota
    WorldOpenExisting
    WorldCreateNew
)

type WorldSavePolicy uint32

const (
    WorldSaveAutomatic WorldSavePolicy = iota
    WorldSaveManual
)

type WorldRandomTickPolicy uint32

const (
    WorldRandomTicksDisabled WorldRandomTickPolicy = iota
    WorldRandomTicksPerSubchunk
)

type WorldTimePolicy uint32

const (
    WorldTimePreserve WorldTimePolicy = iota
    WorldTimeCycle
    WorldTimeFixed
)

type WorldWeatherPolicy uint32

const (
    WorldWeatherPreserve WorldWeatherPolicy = iota
    WorldWeatherCycle
    WorldWeatherClear
)

type WorldChunkUnloadPolicy uint32

const WorldChunkUnloadAfter WorldChunkUnloadPolicy = 0

type WorldSpec struct {
    ProviderPath string
    Dimension WorldDimension
    OpenMode WorldOpenMode
    ReadOnly bool
    Save WorldSavePolicy
    SaveInterval time.Duration
    RandomTicks WorldRandomTickPolicy
    RandomTickRate uint32
    Time WorldTimePolicy
    FixedTime int64
    Weather WorldWeatherPolicy
    ChunkUnload WorldChunkUnloadPolicy
    ChunkUnloadAfter time.Duration
}
```

`normalizedWorldSpec` stores the cleaned slash path, absolute provider path,
and fully explicit policy values. Map manual save to `SaveInterval: -1`,
map every read-only spec to manual save, map disabled ticks to
`RandomTickSpeed: -1`, and never pass zero for the three Dragonfly-defaulted
config fields. Reject existing symlink components below the configured root.

`applySettings` locks the provider settings and applies fixed/cycling time and
clear/cycling weather before returning.

- [ ] **Step 4: Run normalization tests**

Run:

```bash
go test ./framework -run 'TestNormalizeWorldSpec|TestWorldSpecSettings|TestWorldSpecPath' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit the policy model**

```bash
git add framework/world_spec.go framework/world_spec_test.go
git commit -m "feat(worlds): define persistent world policies"
```

---

### Task 2: Make WorldManager Own Immutable Open Specifications

**Files:**
- Modify: `framework/worlds.go`
- Create: `framework/world_spec_open_test.go`

**Interfaces:**
- Consumes: `WorldSpec` and `normalizedWorldSpec` from Task 1.
- Produces: `WorldManager.OpenSpec(name WorldID, spec WorldSpec)`, exact duplicate semantics, one provider owner per path, and the legacy `Open` wrapper used by the native layer.

- [ ] **Step 1: Write failing open-mode and duplicate tests**

Add separate tests with these exact responsibilities:

```go
func TestWorldManagerConcurrentIdenticalSpecsShareOneWorld(t *testing.T)
func TestWorldManagerRejectsMismatchedDuplicateSpec(t *testing.T)
func TestWorldManagerRejectsProviderPathAlias(t *testing.T)
func TestWorldManagerRejectsProviderPathSymlink(t *testing.T)
func TestWorldManagerOpenExistingRequiresMCDBArtifacts(t *testing.T)
func TestWorldManagerCreateNewRejectsExistingPath(t *testing.T)
func TestWorldManagerFailedOpenReleasesReservations(t *testing.T)
func TestWorldManagerUnloadReleasesProviderPathAfterClose(t *testing.T)
```

For `OpenExisting`, require regular `level.dat` and `db/CURRENT` files before
Dragonfly opens the provider. Build a valid fixture by opening and closing an
`mcdb` world once, not by inventing database bytes.

- [ ] **Step 2: Run the focused tests and verify red state**

Run:

```bash
go test ./framework -run 'TestWorldManager.*Spec|TestWorldManagerOpenExisting|TestWorldManagerCreateNew|TestWorldManagerFailedOpen|TestWorldManagerUnloadReleases' -count=1
```

Expected: compile failure because `OpenSpec` does not exist.

- [ ] **Step 3: Add opening and provider-path ownership state**

Extend `managedWorld` and `WorldManager`:

```go
type managedWorld struct {
    // Existing fields remain.
    spec *normalizedWorldSpec
}

type worldOpening struct {
    spec normalizedWorldSpec
    done chan struct{}
    id native.WorldID
    err error
}

type WorldManager struct {
    // Existing fields remain.
    openings map[WorldID]*worldOpening
    providerPaths map[string]WorldID
}
```

Initialize both maps in `newWorldManager`. Reserve world ID and provider path
under `m.mu`, perform provider work outside the lock, publish once, and close
`done` on every path. Identical concurrent callers wait for the same opening;
mismatches reject immediately.

- [ ] **Step 4: Implement provider preflight and construction mapping**

Add:

```go
func (m *WorldManager) OpenSpec(name WorldID, spec WorldSpec) (native.WorldID, error)
func (m *WorldManager) openNormalized(name WorldID, spec normalizedWorldSpec) (native.WorldID, error)
func preflightProvider(mode WorldOpenMode, path string) error
```

`OpenSpec` may return an already published identical spec from any caller.
Concurrent identical opens wait on the opening's `done` channel without
holding `m.mu`. Provider opening stays synchronous, matching the existing API;
no invocation token is retained or used to add callback-only restrictions.

Open `mcdb.Config{Log: m.log, Blocks: blocks}`, call `spec.applySettings` on
`provider.Settings()`, then construct:

```go
config := spec.config(m.log, provider, blocks, entities)
config.PortalDestination = m.portalDestination
managed := config.New()
```

Make legacy `Open(name, dimension)` derive `namespace/path` and construct the
documented defaults. Store the normalized spec on successful persistent
worlds. Release the provider path only after `entry.world.Close()` returns
during unload.

- [ ] **Step 5: Run framework tests including races**

Run:

```bash
go test ./framework -count=1
go test -race ./framework -run 'TestWorldManagerConcurrentIdenticalSpecs|TestWorldManagerRejectsMismatchedDuplicateSpec' -count=1
```

Expected: PASS with one world handle returned to every identical caller.

- [ ] **Step 6: Commit manager ownership**

```bash
git add framework/worlds.go framework/world_spec_open_test.go
git commit -m "feat(worlds): open immutable persistent specs"
```

---

### Task 3: Expose WorldSpec Through Host ABI v16

**Files:**
- Modify: `cmd/abi-gen/main.go`
- Modify generated: `abi/include/dragonfly_plugin.h`
- Modify generated: `rust/dragonfly-plugin-sys/src/generated.rs`
- Modify: `rust/dragonfly-plugin-sys/src/lib.rs`
- Modify: `internal/native/bridge.c`
- Modify: `internal/native/host.go`
- Create: `internal/native/world_spec.go`
- Modify: `internal/native/world_exports.go`
- Create: `internal/native/world_spec_test.go`
- Modify: `rust/runtime/src/lib.rs`
- Modify: `rust/dragonfly-plugin-macros/src/lib.rs`

**Interfaces:**
- Consumes: `WorldManager.OpenSpec` and the standalone ABI v16 layout from the design.
- Produces: `DfWorldOpenSpecV1`, `DfHostWorldOpenSpecFn`, `DfHostApiV16.world_open_spec`, `native.WorldOpenSpec`, and `Host.OpenWorldSpec` for Task 4.

- [ ] **Step 1: Apply the ABI coordination gate**

Run:

```bash
git fetch origin
git log --oneline --all -- abi/include/dragonfly_plugin.h | head -20
```

Expected on the specified base: master is host ABI v15. If another host ABI
v16 layout has landed, stop and revise this design and plan to v17 before
editing. The current transferable-entities v16 uses offset 448 for another
function signature and cannot coexist with the standalone layout here.

- [ ] **Step 2: Write failing layout and native validation tests**

Add these Rust layout assertions:

```rust
assert_eq!(size_of::<DfWorldOpenSpecV1>(), 80);
assert_eq!(offset_of!(DfWorldOpenSpecV1, provider_path), 8);
assert_eq!(offset_of!(DfWorldOpenSpecV1, fixed_time), 40);
assert_eq!(offset_of!(DfWorldOpenSpecV1, open_mode), 48);
assert_eq!(offset_of!(DfWorldOpenSpecV1, read_only), 76);
assert_eq!(size_of::<DfHostApiV16>(), 456);
assert_eq!(offset_of!(DfHostApiV16, world_open_spec), 448);
```

In `internal/native/world_spec_test.go`, table-test null-equivalent values,
unknown tags, non-zero reserved bytes, invalid duration/rate combinations,
read-only automatic save, invalid UTF-8, and provider paths over 4096 bytes.

- [ ] **Step 3: Run tests and confirm missing ABI types**

Run:

```bash
cargo test -p dragonfly-plugin-sys
go test ./internal/native -run 'TestCopyWorldOpenSpec' -count=1
```

Expected: compile failures for the missing v16 types and converter.

- [ ] **Step 4: Generate the exact host ABI v16 contract**

Update the generator, not generated files by hand, with the constants and
80-byte `DfWorldOpenSpecV1` from the design. Keep `world_open`; append:

```c
DfHostWorldOpenSpecFn world_open_spec;
```

to `DfHostApiV16`. Change every generated/runtime/macro host pointer from
`DfHostApiV15` to `DfHostApiV16`, set `DF_HOST_ABI_VERSION` to 16, and run:

```bash
go run ./cmd/abi-gen -root .
cargo fmt --all
```

- [ ] **Step 5: Implement bounded native conversion and export**

Define `native.WorldOpenSpec` with exact policy tags and durations in
`internal/native/world_spec.go`. Implement:

```go
func copyWorldOpenSpec(view *C.DfWorldOpenSpecV1) (WorldOpenSpec, bool)
```

using the design's validation table. Add to `Host`:

```go
OpenWorldSpec(InvocationID, string, WorldOpenSpec) (WorldID, bool)
```

The framework adapter converts the wire value to `framework.WorldSpec` and
calls `WorldManager.OpenSpec(name, spec)`. The invocation remains part of the
generic host signature but does not affect synchronous provider ownership.

Export:

```go
//export bg_go_world_open_spec
func bg_go_world_open_spec(
    context C.uint64_t,
    invocation C.DfInvocationId,
    name C.DfStringView,
    spec *C.DfWorldOpenSpecV1,
    output *C.DfWorldId,
) C.DfStatus
```

Zero `output` before validation. Resolve the host, copy bounded strings, call
`OpenWorldSpec`, and expose only `DF_STATUS_OK` or `DF_STATUS_ERROR`.

- [ ] **Step 6: Wire and assert the C bridge**

Add the extern, wrapper, and host table field. Add `_Static_assert` entries for
the spec and host offsets. Set the new field to the wrapper in
`bg_runtime_open`.

- [ ] **Step 7: Regenerate and run ABI/native verification**

Run:

```bash
make check-generated
cargo test -p dragonfly-plugin-sys
cargo test -p dragonfly-plugin-runtime
go test ./internal/native -count=1
```

Expected: PASS; generated check reports no stale files.

- [ ] **Step 8: Commit ABI v16 atomically**

```bash
git add cmd/abi-gen/main.go abi/include/dragonfly_plugin.h rust/dragonfly-plugin-sys/src/generated.rs rust/dragonfly-plugin-sys/src/lib.rs internal/native/bridge.c internal/native/host.go internal/native/world_spec.go internal/native/world_exports.go internal/native/world_spec_test.go rust/runtime/src/lib.rs rust/dragonfly-plugin-macros/src/lib.rs
git commit -m "feat(abi)!: expose persistent world specs"
```

---

### Task 4: Add the Typed Rust WorldSpec API

**Files:**
- Modify: `rust/dragonfly-plugin/src/world.rs`
- Create: `rust/dragonfly-plugin/src/world_spec_test.rs`
- Modify: `rust/dragonfly-plugin/src/lib.rs`

**Interfaces:**
- Consumes: `DfWorldOpenSpecV1` and `DfHostApiV16.world_open_spec` from Task 3.
- Produces: the public Rust enums, `WorldSpec` builder, and `World::open_with` described in the design.

- [ ] **Step 1: Write failing encoding and host-call tests in a separate file**

Attach `world_spec_test.rs` under `#[cfg(test)]` from `world.rs`. Add:

```rust
#[test]
fn defaults_encode_explicit_dragonfly_values() {}

#[test]
fn read_only_switches_to_manual_save() {}

#[test]
fn every_policy_encodes_its_exact_tag() {}

#[test]
fn invalid_rate_or_duration_does_not_call_host() {}

#[test]
fn open_with_returns_none_for_host_rejection() {}
```

Use a fake `DfHostApiV16.world_open_spec` callback to capture the view and
assert `struct_size == 80`, positive default durations, and zero unused fields.

- [ ] **Step 2: Run focused Rust tests and verify red state**

Run:

```bash
cargo test -p dragonfly -- world_spec
```

Expected: compile failure because `WorldSpec` and policy enums do not exist.

- [ ] **Step 3: Implement typed policies and builder validation**

Implement the exact public API from the design in `world.rs`. Keep fields
private. `WorldSpec::persistent` stores the owned provider path and exact
defaults. Duration conversion uses checked milliseconds and rejects zero or
values above `i64::MAX / 1_000_000` milliseconds before FFI.

Re-export the complete typed API at the crate root so plugins may use either
`dragonfly::world::WorldSpec` or the concise imports used by the example:

```rust
pub use world::{
    ChunkUnloadPolicy, Dimension, OpenMode, RandomTicks, SavePolicy, TimePolicy,
    WeatherPolicy, World, WorldSpec,
};
```

`read_only(true)` canonicalizes save to `Manual`, independent of call order. Encoding returns `None` for
contradictory or invalid values and produces a fully tagged
`DfWorldOpenSpecV1` for valid values.

- [ ] **Step 4: Implement `World::open_with`**

Use this control flow:

```rust
pub fn open_with(name: &str, spec: &WorldSpec) -> Option<Self> {
    let host = crate::host_api()?;
    let open = host.world_open_spec?;
    let raw_spec = spec.encode()?;
    let mut world = dragonfly_plugin_sys::DfWorldId::default();
    let status = unsafe {
        open(
            host.context,
            crate::current_invocation(),
            crate::string_view_from_str(name),
            &raw_spec,
            &mut world,
        )
    };
    (status == dragonfly_plugin_sys::DF_STATUS_OK && world.value != 0)
        .then_some(Self { raw: world.value })
}
```

Keep old `World::open` backed by `host.world_open` so its wire behavior remains
explicit and covered.

- [ ] **Step 5: Run SDK tests and lints**

Run:

```bash
cargo test -p dragonfly
cargo clippy -p dragonfly --all-targets -- -D warnings
```

Expected: PASS with tests located in `world_spec_test.rs`.

- [ ] **Step 6: Commit the SDK API**

```bash
git add rust/dragonfly-plugin/src/world.rs rust/dragonfly-plugin/src/world_spec_test.rs rust/dragonfly-plugin/src/lib.rs
git commit -m "feat(rust): add typed persistent world specs"
```

---

### Task 5: Prove End-to-End Behavior and Document the API

**Files:**
- Create: `internal/native/world_spec_integration_test.go`
- Modify: `examples/plugins/world-command/src/lib.rs`
- Modify: `examples/plugins/world-command/README.md`
- Modify: `README.md`
- Modify: `docs/plans/rust-plugin-architecture.md`

**Interfaces:**
- Consumes: completed manager, ABI v16, native export, and Rust SDK.
- Produces: executable example, cross-boundary proof, and current architecture documentation.

- [ ] **Step 1: Update the world example with one generic specification**

Add `/world open-spec <name>` alongside the existing legacy `/world open`
command. It opens an example world with:

```rust
WorldSpec::persistent("examples/managed")
    .open_mode(OpenMode::OpenOrCreate)
    .save(SavePolicy::Manual)
    .random_ticks(RandomTicks::Disabled)
    .time(TimePolicy::Fixed(6000))
    .weather(WeatherPolicy::Clear)
```

Do not add application-specific logic or absolute paths.

- [ ] **Step 2: Add a native integration test**

Define a purpose-built recording host in `world_spec_integration_test.go`; do
not add WorldSpec test state to the broad `recordingHost` in `native_test.go`.
Open the built runtime with that host, enable the updated world-command plugin,
and dispatch its `open-spec` runnable. This exercises the Rust builder, runtime
host table, C bridge, Go export, and Host call in one path. Assert exact name,
provider path, tags, durations, and output world ID. Add direct malformed-view
cases proving the host method is not called before command dispatch.

If the native artifacts are missing, follow the existing `nativeArtifacts`
test convention and skip the cross-boundary case with an actionable
`make build-native` message; the complete matrix below builds artifacts before
running the test.

- [ ] **Step 3: Update public and architecture documentation**

Document defaults, root-relative paths, immutable duplicate behavior,
open/create semantics, callback restrictions, and `Option` failure semantics.
Record host ABI v16 and the explicit transferable-entities conflict rule.

- [ ] **Step 4: Run the complete verification matrix**

Run:

```bash
make check-generated
go test ./...
go test -race ./framework ./internal/host ./internal/native
cargo test --workspace
cargo clippy --workspace --all-targets -- -D warnings
cargo build --manifest-path examples/plugins/world-command/Cargo.toml
git diff --check
```

Expected: every command exits 0.

- [ ] **Step 5: Commit documentation and end-to-end evidence**

```bash
git add internal/native/world_spec_integration_test.go examples/plugins/world-command/src/lib.rs examples/plugins/world-command/README.md README.md docs/plans/rust-plugin-architecture.md
git commit -m "docs(worlds): demonstrate persistent world specs"
```

---

### Task 6: Integration and Consumer Pinning Gate

**Files:**
- Modify only if required by selected ABI integration: generated ABI/runtime/macro files listed in Task 3.
- Modify after a master push: the `minimal` branch framework dependency pin.

**Interfaces:**
- Consumes: complete feature branch and current `feature/transferable-entities` state.
- Produces: one non-conflicting host ABI layout and an exact verified minimal pin.

- [ ] **Step 1: Compare final ABI layouts before merge**

Run:

```bash
git fetch origin
git diff origin/master...HEAD -- abi/include/dragonfly_plugin.h rust/dragonfly-plugin-sys/src/generated.rs internal/native/bridge.c
git show feature/transferable-entities:abi/include/dragonfly_plugin.h | grep -n 'DfHostApiV16\|world_entity_remove\|world_open_spec'
```

Expected: identify exactly one selected v16 layout. If transferable v16 landed
first, rebase and advance WorldSpec to v17. If WorldSpec v16 lands first,
transferable entities later advances to v17. Never merge offset-448 functions
with different signatures under v16.

- [ ] **Step 2: Re-run the complete matrix after an ABI rebase**

Run the Task 5 verification matrix again. Expected: every command exits 0 and
all layout assertions match the selected version.

- [ ] **Step 3: Merge/push through repository workflow**

Commit required ABI reconciliation with:

```bash
git commit -m "fix(abi)!: reconcile host world capabilities"
```

After the project commit is pushed from the main development branch, update
`minimal` to consume that exact project commit, run its strict-minimum build,
then commit and push `minimal` as required by `AGENTS.md`.
