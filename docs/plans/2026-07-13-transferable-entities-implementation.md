# Transferable Entities Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expose safe, move-only non-player entity removal and re-addition across managed Dragonfly worlds.

**Architecture:** Host ABI v16 carries generation-checked detached tokens. `WorldManager` owns detached and queued transfer state; active entity registry reserves fresh inactive IDs before asynchronous cross-world adds. Rust `DetachedEntity` owns one token and releases it on drop, while public `Result` values represent attached-versus-detached ownership rather than transport errors.

**Tech Stack:** Go 1.26, Dragonfly v0.11, cgo/C ABI, Rust 2024 workspace, generated bindings, `world.Tx`, `world.EntityHandle`.

## Global Constraints

- Plugin ABI remains v3; host ABI becomes v16.
- Never retain `world.Entity` or `world.Tx` outside its transaction.
- Never block one world owner while waiting on another world.
- Reject player removal; player sessions require a separate world-transfer API.
- Old active entity IDs never reactivate after removal.
- Host transport status stays private to Rust SDK.
- `DetachedEntity` is move-only and cleans up exactly once.
- Every production change follows red-green-refactor.
- Push coherent milestones; every pushed master commit must be pinned, built, committed, and pushed on `minimal`.

---

### Task 1: Define host ABI v16

**Files:**
- Modify: `cmd/abi-gen/main.go`
- Modify generated: `abi/include/dragonfly_plugin.h`
- Modify generated: `rust/dragonfly-plugin-sys/src/generated.rs`
- Test: `rust/dragonfly-plugin-sys/src/lib.rs`
- Modify: `internal/native/bridge.h`
- Modify: `internal/native/bridge.c`
- Modify: `rust/runtime/src/lib.rs`

**Interfaces:**
- Produces: `DfDetachedEntityId`, `DfHostWorldEntityRemoveFn`, `DfHostWorldEntityAddFn`, `DfHostDetachedEntityDropFn`, `DfHostApiV16`.
- Preserves: `DF_ABI_VERSION == 3`.

- [ ] **Step 1: Write failing Rust layout test**

Replace `host_v15_layout_is_stable` with v16 assertions and add detached-ID layout:

```rust
#[test]
#[cfg(target_pointer_width = "64")]
fn host_v16_layout_is_stable() {
    assert_eq!(DF_ABI_VERSION, 3);
    assert_eq!(DF_HOST_ABI_VERSION, 16);
    assert_eq!(size_of::<DfDetachedEntityId>(), 16);
    assert_eq!(align_of::<DfDetachedEntityId>(), 8);
    assert_eq!(offset_of!(DfDetachedEntityId, generation), 8);
    assert_eq!(size_of::<DfHostApiV16>(), 472);
    assert_eq!(offset_of!(DfHostApiV16, world_entity_remove), 448);
    assert_eq!(offset_of!(DfHostApiV16, world_entity_add), 456);
    assert_eq!(offset_of!(DfHostApiV16, detached_entity_drop), 464);
}
```

- [ ] **Step 2: Run test and confirm RED**

Run: `cargo test -p dragonfly-plugin-sys host_v16_layout_is_stable`

Expected: compile failure because v16 types do not exist.

- [ ] **Step 3: Add generator source definitions**

Generate these exact C signatures and Rust equivalents:

```c
typedef struct DfDetachedEntityId {
    uint64_t value;
    uint64_t generation;
} DfDetachedEntityId;

typedef DfStatus (*DfHostWorldEntityRemoveFn)(
    uint64_t context, DfInvocationId invocation, DfWorldId world,
    DfEntityId entity, DfDetachedEntityId *detached);
typedef DfStatus (*DfHostWorldEntityAddFn)(
    uint64_t context, DfInvocationId invocation, DfWorldId world,
    DfDetachedEntityId detached, const DfVec3 *position,
    DfEntityId *entity);
typedef void (*DfHostDetachedEntityDropFn)(
    uint64_t context, DfDetachedEntityId detached);
```

Append three pointers to `DfHostApiV16`, change all host API/config/set-host references from `DfHostApiV15` to `DfHostApiV16`, and set `DF_HOST_ABI_VERSION` to `16`.

- [ ] **Step 4: Regenerate bindings**

Run: `go run ./cmd/abi-gen -root .`

Expected: header and Rust generated file change; `go run ./cmd/abi-gen -root . -check` exits 0.

- [ ] **Step 5: Wire C bridge declarations and wrappers**

Add externs, wrappers, v16 static assertions, and initializer fields:

```c
extern DfStatus bg_go_world_entity_remove(
    uint64_t, DfInvocationId, DfWorldId, DfEntityId,
    DfDetachedEntityId *);
extern DfStatus bg_go_world_entity_add(
    uint64_t, DfInvocationId, DfWorldId, DfDetachedEntityId,
    const DfVec3 *, DfEntityId *);
extern void bg_go_detached_entity_drop(uint64_t, DfDetachedEntityId);
```

Rename runtime-side host pointer types to `DfHostApiV16`; do not add plugin ABI fields.

- [ ] **Step 6: Verify GREEN**

Run:

```bash
cargo test -p dragonfly-plugin-sys
cargo test -p dragonfly-plugin-runtime
go run ./cmd/abi-gen -root . -check
```

Expected: all pass; v16 size is 472 bytes.

- [ ] **Step 7: Commit locally**

```bash
git add cmd/abi-gen abi/include rust/dragonfly-plugin-sys rust/runtime internal/native/bridge.c internal/native/bridge.h
git commit -m "feat(abi)!: add detached entity handles"
```

Commit body must state host ABI v15 clients must rebuild for v16.

---

### Task 2: Add active and detached ownership registries

**Files:**
- Modify: `internal/native/host.go`
- Modify: `internal/host/entities.go`
- Test: `internal/host/entities_test.go`
- Create: `framework/detached_entities.go`
- Test: `framework/detached_entities_test.go`

**Interfaces:**
- Consumes: `native.DetachedEntityID` matching ABI token fields.
- Produces: inactive ID reservation, fresh-generation activation, atomic detached consume/drop.

- [ ] **Step 1: Write failing active-ID lifecycle tests**

Add tests proving an expired ID never revives and a reserved ID cannot resolve:

```go
func TestEntitiesReserveFreshGeneration(t *testing.T) {
    withPlayerTx(t, func(_ *world.Tx, connected *player.Player) {
        entities := NewEntities()
        handle := connected.H()
        first := entities.registerHandle(handle, 0)
        entities.expire(first)
        reserved := entities.reserve(handle)
        if first == reserved || reserved.Generation == 0 {
            t.Fatalf("reserved ID = %#v after %#v", reserved, first)
        }
        if _, ok := entities.Handle(reserved); ok {
            t.Fatal("reserved ID resolved before activation")
        }
        entities.activate(handle)
        if _, ok := entities.Handle(reserved); !ok {
            t.Fatal("activated ID did not resolve")
        }
    })
}
```

- [ ] **Step 2: Run test and confirm RED**

Run: `go test ./internal/host -run TestEntitiesReserveFreshGeneration -count=1`

Expected: compile failure for missing `expire`, `reserve`, and `activate`.

- [ ] **Step 3: Implement entity registry lifecycle**

Replace boolean entry state with explicit state:

```go
type entityEntryState uint8
const (
    entityInactive entityEntryState = iota
    entityActive
)
type entityEntry struct {
    id native.EntityID
    state entityEntryState
}
```

Implement `expire(id)`, `reserve(handle)`, and `activate(handle)`. `Register`
activates a reserved handle; it never reuses an expired entry. `Handle` returns
only active entries.

Add native token type used by later layers:

```go
type DetachedEntityID struct {
    Value      uint64
    Generation uint64
}

func (id DetachedEntityID) Valid() bool {
    return id.Value != 0 && id.Generation != 0
}
```

- [ ] **Step 4: Write failing detached registry race tests**

Test `take` versus `drop` with two goroutines and assert exactly one cleanup:

```go
func TestDetachedRegistryConsumesExactlyOnce(t *testing.T) {
    var cleaned atomic.Int32
    registry := newDetachedEntities()
    handle := new(world.EntityHandle)
    id := registry.put(handle, func() { cleaned.Add(1) })
    var winners atomic.Int32
    var wait sync.WaitGroup
    wait.Add(2)
    go func() { defer wait.Done(); if _, ok := registry.take(id); ok { winners.Add(1) } }()
    go func() { defer wait.Done(); registry.drop(id) }()
    wait.Wait()
    if winners.Load()+cleaned.Load() != 1 {
        t.Fatalf("take winners=%d cleanup=%d", winners.Load(), cleaned.Load())
    }
}
```

- [ ] **Step 5: Run race test and confirm RED**

Run: `go test -race ./framework -run TestDetachedRegistryConsumesExactlyOnce -count=1`

Expected: compile failure because detached registry does not exist.

- [ ] **Step 6: Implement detached registry**

Use mutex-protected entries and monotonic value/generation allocation:

```go
type detachedEntityEntry struct {
    handle *world.EntityHandle
    cleanup func()
}
type detachedEntities struct {
    mu sync.Mutex
    next uint64
    entries map[native.DetachedEntityID]detachedEntityEntry
}
```

`put`, `take`, `drop`, and `drain` each remove/clean at most once. Allocation
fails on `math.MaxUint64`; zero token is always invalid.

- [ ] **Step 7: Verify GREEN**

Run:

```bash
go test -race ./internal/host ./framework -run 'TestEntitiesReserveFreshGeneration|TestDetachedRegistry' -count=1
```

Expected: all focused tests pass.

- [ ] **Step 8: Commit locally**

```bash
git add internal/native/host.go internal/host/entities.go internal/host/entities_test.go framework/detached_entities.go framework/detached_entities_test.go
git commit -m "feat(entities): track detached ownership"
```

---

### Task 3: Implement Dragonfly remove/add transactions

**Files:**
- Modify: `internal/native/host.go`
- Modify: `framework/entities.go`
- Modify: `framework/detached_entities.go`
- Modify: `framework/plugin_entities_advanced.go`
- Modify: `framework/worlds.go`
- Modify: `framework/run.go`
- Test: `framework/entities_test.go`
- Test: `framework/plugin_entities_test.go`

**Interfaces:**
- Produces: `RemoveWorldEntity`, `AddWorldEntity`, `DropDetachedEntity`, `DrainDetachedEntities` host methods.
- Requires: destination add accepts optional finite `*native.Vec3`.

- [ ] **Step 1: Write failing transfer behavior tests**

Cover same-world remove/add, `add_at`, player rejection, and fresh ID:

```go
detached, ok := manager.RemoveWorldEntity(0, sourceID, originalID)
if !ok { t.Fatal("remove failed") }
if _, ok := manager.EntityState(0, originalID); ok { t.Fatal("old ID remained live") }
added, ok := manager.AddWorldEntity(0, targetID, detached, nil)
if !ok || added.Generation == originalID.Generation { t.Fatalf("added=%#v", added) }
```

Add a viewer to destination chunk and assert `ViewEntity` runs once for
`AddWorldEntity(..., &native.Vec3{X: 8, Y: 70, Z: 9})`.

- [ ] **Step 2: Run focused tests and confirm RED**

Run: `go test ./framework -run 'TestWorldEntityTransfer|TestWorldEntityAddAt|TestWorldEntityTransferRejectsPlayer' -count=1`

Expected: compile failure for missing manager methods.

- [ ] **Step 3: Implement remove transaction**

Resolve source world under lifecycle read lock. Use current invocation
transaction only when it belongs to source; reject cross-world active
invocations. Before `tx.RemoveEntity`, reject `*player.Player` and capture
advanced cleanup with a type switch. After removal, expire old ID and put
handle into detached registry.

- [ ] **Step 4: Implement add and add-at transactions**

Validate destination and optional finite position before taking token. Reserve
fresh inactive ID, then:

- use current destination transaction immediately;
- use `world.Call` when invocation is zero;
- queue `World.Do` for a different active invocation.

Call `Tx.AddEntity` for nil position and `Tx.AddEntityAt` otherwise. Existing
world spawn handler activates reserved ID. Recover Dragonfly panics and expire
reserved ID plus cleanup on failure.

- [ ] **Step 5: Add cross-world no-deadlock test**

Begin invocation in source transaction, remove entity, add to target, end
invocation, and await target observation with a channel—not a sleep. Assert add
call returns before target task completes and target entity eventually resolves.

- [ ] **Step 6: Add advanced cleanup test**

Transfer a plugin-owned advanced entity and assert runtime instance survives
remove/add. In separate case drop detached token and assert `EntityDestroy`
runs exactly once.

- [ ] **Step 7: Drain before runtime unload**

Add `WorldManager.DrainDetachedEntities()` and invoke it after managed worlds
close but before `pluginRuntime.Disable()`/native runtime destruction. Await or
cancel queued additions, expire pending IDs, destroy advanced state, and close
handles.

- [ ] **Step 8: Verify GREEN with race detector**

Run:

```bash
go test -race ./framework ./internal/host -run 'TestWorldEntity|TestDetached|TestEntitiesReserve' -count=1
```

Expected: all pass; no race reports or transaction-after-finish panic.

- [ ] **Step 9: Commit locally**

```bash
git add framework internal/host internal/native/host.go
git commit -m "feat(worlds): transfer entity handles"
```

---

### Task 4: Bridge native ABI and validate tokens

**Files:**
- Modify: `internal/native/entity_exports.go`
- Modify: `internal/native/host.go`
- Modify: `internal/native/native_test.go`
- Modify: `internal/native/bridge.c`

**Interfaces:**
- Consumes: host methods from Task 3.
- Produces: `bg_go_world_entity_remove`, `bg_go_world_entity_add`, `bg_go_detached_entity_drop`.

- [ ] **Step 1: Write failing native boundary tests**

Extend `recordingHost` and test:

- null remove/add outputs return `DF_STATUS_ERROR`;
- zero detached token is rejected;
- non-finite add-at position is rejected before host call;
- repeated add fails;
- drop forwards valid token once and ignores zero token.

- [ ] **Step 2: Run focused tests and confirm RED**

Run: `go test ./internal/native -run TestDetachedEntityHostBridge -count=1`

Expected: compile/link failure because exports are missing.

- [ ] **Step 3: Implement native conversions and exports**

Add exact conversions:

```go
func detachedEntityID(value C.DfDetachedEntityId) DetachedEntityID {
    return DetachedEntityID{Value: uint64(value.value), Generation: uint64(value.generation)}
}
func cDetachedEntityID(value DetachedEntityID) C.DfDetachedEntityId {
    return C.DfDetachedEntityId{value: C.uint64_t(value.Value), generation: C.uint64_t(value.Generation)}
}
```

Validate world/entity generations, nullable position, finite coordinates, and
outputs before host dispatch. Drop has no public status and is idempotent.

- [ ] **Step 4: Verify GREEN**

Run:

```bash
make build-native
go test ./internal/native -run TestDetachedEntityHostBridge -count=1
```

Expected: native fixture rebuild succeeds and focused tests pass.

- [ ] **Step 5: Commit locally**

```bash
git add internal/native
git commit -m "feat(native): bridge detached entities"
```

---

### Task 5: Add move-only Rust SDK API

**Files:**
- Modify: `rust/dragonfly-plugin/src/entity.rs`
- Modify: `rust/dragonfly-plugin/src/world.rs`
- Test: `rust/dragonfly-plugin/src/world.rs`
- Test: doc test in `rust/dragonfly-plugin/src/entity.rs`

**Interfaces:**
- Produces: `entity::DetachedEntity`, `World::remove_entity`, `World::add_entity`, `World::add_entity_at`.
- Consumes: host ABI v16 fields from Task 1.

- [ ] **Step 1: Write failing SDK behavior tests**

Install mock host functions and assert remove/add raw IDs, ownership return, and
drop count. Include compile-fail doc test proving moved detached value cannot be
used twice:

```rust
/// ```compile_fail
/// use dragonfly::entity::DetachedEntity;
/// fn consume(_: DetachedEntity) {}
/// fn cannot_copy(value: DetachedEntity) {
///     consume(value);
///     consume(value);
/// }
/// ```
```

- [ ] **Step 2: Run tests and confirm RED**

Run: `cargo test -p dragonfly-plugin world::tests::detached_entity`

Expected: compile failure because public type and methods are missing.

- [ ] **Step 3: Implement move-only token owner**

```rust
pub struct DetachedEntity {
    raw: dragonfly_plugin_sys::DfDetachedEntityId,
}

impl Drop for DetachedEntity {
    fn drop(&mut self) {
        if self.raw.generation == 0 { return; }
        if let Some(drop_entity) = crate::host_api().and_then(|host| host.detached_entity_drop) {
            unsafe { drop_entity(host.context, self.raw) };
        }
        self.raw = dragonfly_plugin_sys::DfDetachedEntityId::default();
    }
}
```

Do not implement `Clone` or `Copy`.

- [ ] **Step 4: Implement world ownership methods**

Remove returns `Err(entity)` on native failure. Add passes current token without
disarming; only after `DF_STATUS_OK` and non-zero output generation does it zero
the token and return `Ok(Entity)`. Failure returns `Err(detached)` still armed.
`add_entity_at` rejects non-finite coordinates before host call.

- [ ] **Step 5: Verify GREEN**

Run:

```bash
cargo test -p dragonfly-plugin
cargo test --doc -p dragonfly-plugin
cargo fmt --all -- --check
```

Expected: all unit and compile-fail doc tests pass.

- [ ] **Step 6: Commit locally**

```bash
git add rust/dragonfly-plugin
git commit -m "feat(rust): expose transferable entities"
```

---

### Task 6: Example, docs, and milestone verification

**Files:**
- Modify: `examples/plugins/entity-command/src/lib.rs`
- Modify: `README.md`
- Modify: `docs/plans/rust-plugin-architecture.md`
- Modify: `docs/plans/2026-07-13-transferable-entities-design.md`

**Interfaces:**
- Consumes: final Rust SDK API.
- Produces: runnable transfer command and current parity documentation.

- [ ] **Step 1: Add example command**

Add one command runnable that spawns typed entity, removes it, opens
`example:entity_transfer`, and re-adds it with `add_entity_at`. Use no raw entity
identifier and message only domain outcomes.

- [ ] **Step 2: Update docs**

Document:

- host ABI v16 and plugin ABI v3;
- move-only detached ownership;
- queued cross-world state reads returning `None`;
- player rejection;
- `Tx.Viewers` upstream identity limitation;
- AddEntity/RemoveEntity priority item complete.

Mark design status `Implemented` only after verification.

- [ ] **Step 3: Run full verification**

Run:

```bash
make test
go test -race ./framework ./internal/host ./internal/native -count=1
git diff --check
git status --short
```

Expected: generated checks, Rust workspace tests, every plugin example, Go tests,
race tests, and diff checks all pass.

- [ ] **Step 4: Review implementation diff**

Dispatch read-only review focused on ownership races, shutdown ordering,
transaction deadlocks, stale-ID revival, player rejection, and ABI layout. Fix
every release blocker and rerun Step 3.

- [ ] **Step 5: Commit fixes/docs locally**

```bash
git add README.md docs examples framework internal rust abi cmd
git commit -m "docs(entities): document handle transfers"
```

- [ ] **Step 6: Push master milestone**

Run: `git push origin master`

Expected: remote master matches local HEAD.

- [ ] **Step 7: Refresh minimal branch**

Pin `Makefile`, Go module, every plugin `Cargo.toml`, and every lockfile to exact
master hash. Sync changed entity example, then run:

```bash
make build
go test ./... -count=1
```

Commit `chore(minimal): pin entity transfers` and push `minimal`. Verify both
worktrees clean and both remote tracking heads equal local heads.
