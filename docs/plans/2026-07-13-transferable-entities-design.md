# Transferable entities design

Status: Draft approved in conversation; awaiting written-spec review

## Goal

Expose Dragonfly's `Tx.RemoveEntity` and `Tx.AddEntity` ownership transition
without retaining `world.Entity` or `world.Tx` values across FFI. This fills the
requested world entity-management gap and permits moving non-player entities
between managed worlds.

This work does not expose `Tx.Viewers`. Dragonfly returns opaque
`world.Viewer` implementations with no public player identity, so converting
them to `Player` handles would be incorrect.

## Rust API

```rust
let Ok(detached) = source.remove_entity(entity) else {
    return;
};
let Ok(entity) = destination.add_entity(detached) else {
    return;
};
```

`World::remove_entity(Entity) -> Result<DetachedEntity, Entity>` removes a
non-player entity from that exact world. Success invalidates every old `Entity`
handle for that entity. Failure returns the still-valid entity handle.

`World::add_entity(DetachedEntity) -> Result<Entity, DetachedEntity>` consumes
the detached owner and returns a fresh-generation `Entity` handle. A rejected
destination returns the still-owned detached value, not a host error. Addition
happens immediately when the current invocation owns the destination
transaction. Otherwise the host queues the add on the destination world. Until
that task runs, state reads return `None` and actions are ignored rather than
blocking a world owner.

`World::add_entity_at(DetachedEntity, Vec3) -> Result<Entity, DetachedEntity>`
uses Dragonfly's `Tx.AddEntityAt`. `add_entity` preserves the position stored
in the handle.

`DetachedEntity` is neither `Clone` nor `Copy`. Its fields are private. Dropping
it releases the detached host entry exactly once. `Result` represents domain
ownership: attached or detached. No native status or transport error appears in
the public API.

Players are rejected. Dragonfly player world changes also move a private
session loader and must use a dedicated player-transfer operation; raw
`Tx.RemoveEntity`/`Tx.AddEntity` would desynchronise the session.

## Host ownership model

World manager owns a detached-entity registry separate from active entity IDs.
Each entry contains:

- a never-reused detached token;
- Dragonfly `*world.EntityHandle`;
- cleanup needed by framework-owned advanced entities;
- lifecycle state: detached, queued, consumed, or released.

Removing performs one source-world transaction:

1. Resolve active entity ID in source transaction.
2. Reject players and entities not present in source world.
3. Capture framework-specific cleanup while entity interface is still valid.
4. Call `Tx.RemoveEntity`.
5. Expire active ID and allocate detached token. Existing world-handler
   deactivation is not allowed to reactivate that generation later.

Adding atomically consumes detached token. Same-world-owner calls use current
transaction. Cross-world calls enqueue `World.Do`, reserve a fresh active ID,
and return it immediately. Reserved IDs remain inactive until `Tx.AddEntity` or
`Tx.AddEntityAt` completes. Queue failure expires reserved ID and releases
entity. Spawn-handler registration adopts reserved fresh generation instead of
reactivating ID expired during removal.

Dropping detached token closes handle and invokes framework-owned state cleanup.
Runtime shutdown drains all detached and queued entries before unloading Rust
libraries. Cleanup is idempotent and never calls plugin code after unload.

## ABI

Host ABI becomes v16. Plugin ABI remains v3.

Add language-neutral types and calls:

```c
typedef struct DfDetachedEntityId {
    uint64_t value;
    uint64_t generation;
} DfDetachedEntityId;

DfStatus world_entity_remove(
    context, invocation, world, entity, detached_out);
DfStatus world_entity_add(
    context, invocation, world, detached, position_or_null, entity_out);
void detached_entity_drop(
    context, detached);
```

Remove/add return ABI status only to SDK internals. `DfDetachedEntityId` uses a
generation so stale or repeated tokens cannot consume another entry. Null
position selects `Tx.AddEntity`; non-null selects `Tx.AddEntityAt` after finite
coordinate validation.

## Failure and concurrency rules

- No call waits on another world owner from an active invocation.
- Remove failure performs no mutation.
- Add consumes token once; repeated or stale tokens fail closed.
- Old active IDs never revive. Successful add gets a new generation.
- Drop racing add has one winner through registry state transition.
- World unload rejects new queued additions and drains already queued work.
- Panics from Dragonfly add/remove are recovered at host boundary and converted
  to stale/released state.
- Detached or queued entities are not returned by `World::entities()`.

## Tests

Test-first coverage:

1. Rust compile/API test proves `DetachedEntity` is move-only and Drop calls
   host release once.
2. Native bridge tests reject null outputs, zero generations, stale tokens,
   repeated add, and non-finite positions.
3. Framework test removes then re-adds entity, proving old ID stale, new ID
   live, position preserved, and world lists updated.
4. `add_entity_at` test proves destination position and viewer spawn behavior.
5. Cross-world invocation test proves source callback never deadlocks and
   destination add completes asynchronously.
6. Player removal test proves no world/session mutation.
7. Race test proves add versus drop consumes and cleans exactly once.
8. Advanced custom-entity test proves transfer preserves Rust instance and
   dropping detached state destroys it once.
9. Shutdown test proves detached/queued state drains before runtime unload.
10. Full generated ABI checks, Rust workspace tests, Go race tests, examples,
    and minimal-branch build remain green.

## Documentation and example

Update architecture parity status and entity example. Example spawns an entity,
removes it from overworld, and adds it to a managed custom world without raw
identifiers or adapter code.
