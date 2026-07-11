# Native Plugin Architecture Plan

Status: Proposed

Initial language: Rust

Primary goals:

- Keep hot event handling close to raw Dragonfly performance.
- Make plugin code idiomatic and small.
- Make adding another plugin language require one thin SDK/runtime shim, not a rewrite of every event adapter.
- Own the Dragonfly server lifecycle so users never attach handlers or worlds manually.
- Let plugins configure, create, inspect, and retire worlds through the generated SDK.
- Reach behavioral parity with Dragonfly's public server extension points, including custom blocks, items, entities, providers, generators, and registries.
- Keep Dragonfly transaction and object lifetime rules intact.
- Keep the binary contract independent from Go and Rust ABI changes.

## Decision

Build an opinionated Dragonfly server runtime around one schema-driven C ABI:

```text
bedrock-gophers server
   |
Dragonfly lifecycle and world manager
   |
Go host adapter
   |
generated C ABI
   |
Rust runtime
   |
Rust plugin libraries
```

The framework owns `Config.New`, `Listen`, `Accept`, player attachment, every managed world, and shutdown. Server operators run the provided executable; they do not write integration Go code.

The Go host makes one FFI call per event. The Rust runtime dispatches that event to every subscribed Rust plugin using cheap native function calls. This avoids one cgo crossing per plugin.

C is only the stable ABI shape. Plugin authors use an idiomatic Rust SDK and write no C.

## Dragonfly constraints

This design is based on Dragonfly `master` at commit `a3fb52159c77d9f2ea5b6fe754ae525e1ac359f4`.

Relevant upstream behavior:

- Dragonfly is a Go library, not a standalone plugin host.
- Players are yielded through `Server.Accept()` and are valid only inside that iterator callback.
- A player has one `player.Handler`; a world has one `world.Handler`.
- Many handlers are synchronous, cancellable, or mutate pointer arguments.
- Most world/entity work occurs inside `world.Tx`.
- `world.Tx` is not safe for concurrent use.
- Awaiting nested `World.Exec` can deadlock.
- `EntityHandle` is the persistent, transaction-safe identity for an entity.

Sources:

- [Dragonfly README](https://github.com/df-mc/dragonfly/blob/a3fb52159c77d9f2ea5b6fe754ae525e1ac359f4/README.md)
- [Server lifecycle](https://github.com/df-mc/dragonfly/blob/a3fb52159c77d9f2ea5b6fe754ae525e1ac359f4/server/server.go)
- [Player handlers](https://github.com/df-mc/dragonfly/blob/a3fb52159c77d9f2ea5b6fe754ae525e1ac359f4/server/player/handler.go)
- [World handlers](https://github.com/df-mc/dragonfly/blob/a3fb52159c77d9f2ea5b6fe754ae525e1ac359f4/server/world/handler.go)
- [World transactions](https://github.com/df-mc/dragonfly/blob/a3fb52159c77d9f2ea5b6fe754ae525e1ac359f4/server/world/tx.go)
- [Entity handles](https://github.com/df-mc/dragonfly/blob/a3fb52159c77d9f2ea5b6fe754ae525e1ac359f4/server/world/entity.go)

The ABI must not expose Go pointers, Go interfaces, `player.Player`, `world.Entity`, or `world.Tx`.

## Minimise adapter code

Maintain one language-neutral event schema:

```text
schema/
  types.yaml
  events/
    player.yaml
    world.yaml
    inventory.yaml
  capabilities/
    blocks.yaml
    items.yaml
    entities.yaml
    worlds.yaml
    server.yaml
```

Example schema:

```yaml
events:
  player_move:
    id: 1
    input:
      player: player_id
      old_position: vec3
      new_position: vec3
      rotation: rotation
    output:
      cancel: bool

  player_chat:
    id: 2
    input:
      player: player_id
      message: string_view
    output:
      cancel: bool
      replacement: optional_string
```

A generator produces:

- C structs, event IDs, ABI tables, and layout assertions.
- Go ABI types and conversion skeletons.
- Rust raw `#[repr(C)]` bindings.
- Rust safe mutable event types.
- Rust handler dispatch code.
- Event reference documentation.
- Test fixtures containing known ABI layouts.

Adding an event then requires:

1. Add schema entry.
2. Implement Dragonfly-to-schema mapping in Go.
3. Regenerate everything else.

No Rust adapter change should be required. Future language SDKs consume the same generated ABI metadata.

Each language needs only:

- Generated ABI types.
- One small runtime shim that exports the entry point.
- Idiomatic wrappers around borrowed strings, slices, IDs, mutable event state, and host errors.

Do not generate plugin business logic or force plugin authors to handle generic byte buffers.

Add a `dragonfly-api-scan` tool using Go type information. It inventories Dragonfly's exported extension interfaces and compares them with the schema. CI fails when the pinned Dragonfly version adds or changes an interface without an explicit supported, deferred, or intentionally unsupported classification. This prevents silent parity drift.

## ABI shape

Use a generic dispatch entry point with generated typed payloads:

```c
typedef uint32_t DfEventId;

typedef struct {
    uint32_t abi_version;
    uint32_t struct_size;
    uint64_t feature_bits;
} DfAbiHeader;

typedef struct {
    DfAbiHeader header;
    void *instance;
    DfStatus (*handle_event)(
        void *instance,
        DfEventId event_id,
        const void *input,
        void *output
    );
    void (*destroy)(void *instance);
} DfPluginApiV1;

const DfPluginApiV1 *df_plugin_entry_v1(void);
```

Generated Rust dispatch converts `event_id` and pointers into typed SDK calls. Plugin authors never match raw IDs or dereference pointers.

Every ABI structure must:

- Use fixed-width integer types.
- Use explicit discriminants for enums.
- Use pointer-plus-length views for borrowed data.
- State ownership and lifetime.
- Avoid platform-sized `long`, `size_t` in persisted layouts, and C bitfields.
- Have generated size, alignment, and offset tests.

ABI v1 evolves append-only where `struct_size` permits it. Breaking changes get a new entry symbol such as `df_plugin_entry_v2`.

## Runtime

Go loads one Rust runtime library. Rust runtime loads plugin `.so`, `.dylib`, or later `.dll` files with local symbol visibility.

Runtime responsibilities:

- Discover manifests and native libraries.
- Validate ABI version and target architecture.
- Resolve dependencies with a topological sort.
- Reject duplicate plugin IDs.
- Build per-event subscriber arrays.
- Sort handlers by priority, then plugin ID.
- Dispatch events and carry mutable event state between handlers.
- Catch Rust panics before they cross FFI.
- Track plugin lifecycle and disabled state.
- Prefix plugin logs.
- Expose event subscription bitsets to Go.

The runtime must not unload plugin libraries. Rust threads, TLS, function pointers, and static destructors make hot `dlclose` unsafe. Reload means server restart. A plugin may be logically disabled by removing it from subscriber arrays.

## Server ownership

This project provides the server executable. Normal operation is:

```shell
bedrock-gophers serve --config server.toml
```

Internal startup sequence:

1. Read base server configuration.
2. Discover and validate native plugins.
3. Run plugin bootstrap callbacks and collect a declarative `ServerPlan`.
4. Resolve configuration and world-definition conflicts deterministically.
5. Build Dragonfly `server.Config` and call `Config.New`.
6. Register the three Dragonfly core worlds with the world manager.
7. Create plugin-defined worlds.
8. Install all player, world, and inventory dispatchers automatically.
9. Run plugin enable callbacks.
10. Call `Listen` and own the `Accept` loop.
11. On shutdown, stop accepting players, disable plugins in reverse order, close worlds, then destroy the Rust runtime.

No public `AttachWorld` or `AttachPlayer` API exists. A world cannot enter the registry without its dispatcher already installed. A player receives its dispatcher before plugin-visible join processing.

The Go host provides the authoritative `player.Handler`, `world.Handler`, and inventory handlers. Internal framework handlers run first, then Rust plugins by priority and ID, then Dragonfly consumes final context and pointer values.

Cancellation is monotonic. Later handlers cannot uncancel an event.

The host does not retain `*player.Player`. It retains `EntityHandle` and generation-tagged IDs where persistent lookup is required.

Go conversion code is the unavoidable adapter because Dragonfly uses Go-specific interfaces and concrete types. Keep conversion next to each handler and test it against Dragonfly semantics.

## World manager

Worlds are first-class managed resources, not objects server owners attach manually.

Reserved core world IDs:

```text
minecraft:overworld
minecraft:nether
minecraft:end
```

Plugin worlds use namespaced IDs:

```text
example:minigame_lobby
example:arena_1
```

World manager owns:

- World registry and opaque ID generations.
- Creation through `world.Config.New`.
- Handler installation before publication.
- Provider and generator preset resolution.
- Portal destination mapping.
- Player evacuation before unload.
- Save, close, and removal lifecycle.
- Ownership and access checks between plugins.
- World load/unload events.

Rust SDK exposes:

```rust
fn bootstrap(&self, server: &mut ServerBuilder) {
    server.worlds().configure("minecraft:overworld", WorldSpec::persistent("world"));
    server.worlds().create(
        "example:lobby",
        WorldSpec::flat().autosave(true),
    );
}

fn on_enable(&self, server: &Server) {
    let lobby = server.worlds().get("example:lobby").unwrap();
    lobby.set_time(6000);
}
```

Bootstrap methods build declarations locally in Rust. Runtime returns one generated `ServerPlan` to Go. This avoids repeated reverse FFI calls during startup.

Runtime world operations use generated commands such as:

- `worlds.list` and `worlds.get`.
- `worlds.create` and `worlds.unload`.
- `world.save` and `world.settings`.
- `world.block` and `world.set_block`.
- `world.players`.
- `player.teleport_to_world`.

Core worlds may be configured during bootstrap but cannot be removed while the server runs. Unloading another world requires an evacuation destination and fails if evacuation cannot complete.

Initial world specs support framework-owned providers and generators: persistent, memory-only, flat, empty, overworld, nether, and end. Rust-defined generators and providers use capability adapters described below.

## Dragonfly capability bridge

Full plugin capability requires more than events. Dragonfly exposes behavior through Go interfaces such as `world.Block`, `world.CustomBlock`, `world.EntityType`, `world.Entity`, `world.Generator`, `world.Provider`, item behavior interfaces, recipes, effects, and commands.

The framework mirrors these public extension points as generated capability descriptors and callbacks.

Two kinds of capability exist:

- Static capability: identity, block states, model, collision box, light level, item category, textures, entity bounding box, command schema. Rust declares it during bootstrap; Go caches it and answers Dragonfly without FFI.
- Dynamic capability: activation, ticking, entity interaction, damage, item use, chunk generation, provider I/O. Go proxy calls the Rust runtime with transaction-scoped input and receives mutations/actions.

Bootstrap must finish registration before `server.Config.New`, because Dragonfly finalizes registries and builds custom resource packs during server construction.

### Go proxies

Go proxy types implement Dragonfly interfaces and delegate behavior to plugin definitions:

```text
foreignBlock       implements world.Block
foreignCustomBlock implements world.CustomBlock and resource metadata
foreignItem        implements world.Item plus item behaviors
foreignEntityType  implements world.EntityType
foreignEntity      implements world.Entity/TickerEntity
foreignGenerator   implements world.Generator
foreignProvider    implements world.Provider
```

Dragonfly frequently uses Go interface presence as a capability check. A universal proxy that implements every optional interface would be incorrect: merely implementing `Liquid`, `Replaceable`, `Living`, or similar interfaces changes behavior.

Bridge rules:

- Implement behavior-neutral optional interfaces on a shared proxy only when defaults exactly match interface absence.
- Use separate proxy families for presence-sensitive roles such as liquid, conductor, custom block, ticking entity, living entity, generator, and provider.
- Generate capability-family selection from bootstrap descriptors.
- Add parity tests comparing an equivalent raw Go implementation with each proxy family.
- If a Dragonfly interface cannot be represented without semantic loss or combinatorial proxy explosion, add a narrow adapter hook to the maintained Dragonfly integration layer and propose it upstream.

Do not claim support for a capability until its proxy passes behavioral parity tests.

### Custom blocks

Rust SDK declares block identity, every state/permutation, model, collision/selection boxes, light, friction, hardness, resistance, materials, geometry, textures, and optional behavior callbacks.

```rust
server.blocks().register(
    Block::custom("example:launcher")
        .display_name("Launcher")
        .texture("all", assets.texture("launcher.png"))
        .strength(3.0)
        .on_activate(Self::activate_launcher),
);
```

Go registers all block states before registry finalization. Framework feeds plugin assets into Dragonfly's resource-pack build path.

### Custom items

Rust SDK declares identity, texture, category, stack size, durability, combat properties, food/consumption, cooldowns, and optional use/release callbacks. Go proxies implement relevant item behavior families and register items before server creation.

### Custom entities

Rust SDK supports base, ticking, and living entity families first, followed by specialized behavior capabilities. Entity definition contains network identity, bounding box, persistent state codec, spawn defaults, and callbacks.

Go owns `EntityHandle`, `EntityData`, world membership, viewer updates, and transaction validity. Rust owns plugin-defined state identified by an opaque instance ID. Tick callbacks receive one borrowed entity context and return state changes/actions. Entity NBT encode/decode routes through a versioned plugin state codec so plugin entities survive restarts.

### Custom generators and providers

A Rust generator receives chunk position, dimension metadata, seed/context, and a host-owned mutable chunk view. One FFI call fills a chunk; per-block FFI calls are forbidden.

A Rust provider implements settings, spawn positions, column load/store, and close through generated callbacks. Provider calls may perform I/O, so they use a separate blocking-capability pool and explicit error values. They never run on the hot event dispatcher lock.

### Other extension points

The same descriptor/proxy model covers:

- Dimensions and portal routing.
- Recipes and creative inventory.
- Enchantments and effects.
- Damage and healing sources.
- Game modes and permissions/allower logic.
- Player and world persistence providers.
- Resource packs.
- Typed commands and command permissions.
- Forms, scoreboards, sounds, particles, structures, and entity actions.

Each family belongs in schema. Language SDKs generate from it; only Go-to-Dragonfly proxy semantics remain handwritten.

### Meaning of parity

Parity means a plugin language can express every supported public Dragonfly extension interface and server operation for the pinned Dragonfly release with equivalent gameplay behavior.

Parity does not mean importing arbitrary Go packages, using unexported Dragonfly internals, passing Go objects through FFI, or monkey-patching concrete Go code. Those are not stable cross-language contracts.

## Rust SDK

Plugin code should contain only requested handlers. Events continue by default; handlers only cancel or mutate them:

```rust
use dragonfly_plugin::{ChatEvent, MoveEvent, Plugin};

struct Example;

#[dragonfly_plugin::plugin]
impl Plugin for Example {
    fn on_move(&self, event: &mut MoveEvent) {
        if event.new_position.y < 0.0 {
            event.cancel();
        }
    }

    fn on_chat(&self, event: &mut ChatEvent) {
        event.replace_message(event.message().replace("foo", "bar"));
    }
}
```

All trait handlers have default no-op implementations. The `plugin` attribute sees which methods are implemented and generates the subscription bitmap and entry point, so Go skips unused events.

There is no `allow()` method or `allow` field. Zero/default ABI event state means continue. Cancellation is monotonic: `cancel()` sets the flag, and no API clears it. Each plugin sees mutations made by earlier plugins.

Rust SDK layers:

```text
dragonfly-plugin-sys     generated unsafe ABI
dragonfly-plugin        safe types and Plugin trait
dragonfly-plugin-macros entry point and dispatch generation
```

Plugins must be `Send + Sync`. Different Dragonfly worlds may invoke the runtime concurrently. Runtime-wide serialization would harm performance.

## Event processing

Use direct generated C structs, not JSON, Protobuf, or sockets.

Hot path:

```text
Dragonfly handler
  -> check subscription bit
  -> construct C-compatible input
  -> one cgo call
  -> Rust runtime dispatches subscribers
  -> return final event state
  -> Go updates Dragonfly context/pointer arguments
```

Keep common mutable state inline and allocation-free:

```c
typedef struct {
    uint8_t cancel;
} DfMoveState;
```

Go zero-initialises state before dispatch. `cancel == 0` means continue; `cancel != 0` means cancelled. Never generate a separate allow flag.

Variable output, such as chat replacement or block drops, uses an output arena owned by the Rust runtime and valid until the FFI call completes. The C bridge copies or applies it before returning to ordinary Go code.

For each event, mirror Dragonfly capability exactly. Example: `HandleMove` permits cancellation but does not expose a mutable target position, so Rust `MoveEvent` must not claim position replacement support.

## Transaction-safe host actions

Separate event mutations from host actions.

Event mutations return directly:

- Cancel event.
- Replace chat text.
- Change damage, XP, drops, knockback, or other mutable arguments.

Host actions are recorded into an event-scoped action list:

- Send message.
- Teleport player.
- Set block.
- Play sound.
- Add particle.

Go applies these actions after Rust returns while the current Dragonfly transaction remains valid. This avoids repeated Rust-to-Go callbacks and nested `World.Exec` deadlocks.

Persistent background operations are not part of v1. Later support must enqueue operations into Go and resolve `EntityHandle` inside a new transaction. Event transaction handles must never escape a callback.

## Object identity

Use opaque, generation-tagged IDs:

```c
typedef struct {
    uint8_t uuid[16];
    uint64_t generation;
} DfPlayerId;

typedef struct {
    uint64_t value;
    uint64_t generation;
} DfWorldId;
```

Generation prevents stale references from affecting a reconnected player or replaced world. IDs never contain Go addresses.

## Commands

Dragonfly commands use reflection over concrete Go `Runnable` structs. Dynamic native plugins cannot create new Go struct methods.

Initial command support uses one generic Go runnable with `cmd.Varargs`:

```text
/plugin-command <raw arguments...>
```

Rust SDK provides parsing and validation. Rich Bedrock command parameter metadata can be generated later from command schema.

## Scope and parity target

Initial ABI foundation includes:

- Framework-owned Dragonfly startup, accept loop, and shutdown.
- Runtime and plugin lifecycle.
- Declarative server bootstrap plan.
- Core-world configuration.
- Plugin-world creation, lookup, save, unload, and player transfer.
- Built-in world provider and generator presets.
- Player join and quit.
- Player movement.
- Chat.
- Hurt and heal.
- Block break and place.
- Basic event-scoped messaging and cancellation.
- Raw-argument commands.
- Player/world opaque IDs.

Temporarily deferred from the first implementation milestone:

- Background plugin threads calling Dragonfly directly.
- Hot unload.
- Untrusted plugin sandboxing.
- Full inventory mutation.
- Direct NBT exposure.

Before product 1.0, parity work includes custom blocks, items, entities, generators, providers, registries, recipes, effects, enchantments, typed commands, persistence, and remaining public Dragonfly operations. Bootstrap is deliberately placed before `Config.New` because Dragonfly registers and finalizes many of these types during server construction.

Native safety limitations are permanent: no hot unload, no sandboxing, and a native crash may terminate the server. Capability limitations are not permanent exclusions.

## Repository layout

```text
.
├── schema/
│   ├── types.yaml
│   └── events/
├── cmd/
│   ├── abi-gen/
│   ├── bedrock-gophers/
│   └── dragonfly-api-scan/
├── abi/
│   ├── include/dragonfly_plugin.h
│   └── ABI.md
├── cmd/bedrock-gophers/
├── framework/
├── internal/
│   ├── host/
│   └── native/
├── rust/
│   ├── runtime/
│   ├── dragonfly-plugin-sys/
│   ├── dragonfly-plugin/
│   └── dragonfly-plugin-macros/
├── examples/
│   ├── chat-filter/
│   └── movement-guard/
├── tests/
│   └── fixture-plugin/
├── Cargo.toml
├── go.mod
└── Makefile
```

Generated files must contain a warning and generator version. CI fails if regeneration changes committed output.

## Safety model

Native plugins are trusted code:

- Segfaults kill the server.
- Infinite loops block the calling Dragonfly transaction.
- Rust panics must be caught inside generated entry points.
- No panic or exception may cross the C ABI.
- Plugins may not retain borrowed pointers.
- Plugins may not use event objects after callbacks return.
- All FFI functions document thread-safety.

Process isolation and WebAssembly are deliberately excluded because this project chooses one native mechanism and prioritises raw-handler performance.

## Verification

Required tests:

- C/Rust size, alignment, and field-offset agreement.
- ABI version and `struct_size` compatibility.
- Unknown event handling.
- Borrowed string and slice lifetime tests.
- Plugin panic containment.
- Duplicate ID and dependency-cycle rejection.
- Concurrent callbacks from multiple worlds.
- Cancellation and mutation parity with raw Go handlers.
- Stale player/world ID rejection.
- Automatic dispatcher installation for core and plugin worlds.
- World creation, evacuation, unload, save, and shutdown ordering.
- Bootstrap conflict resolution.
- API scanner coverage for every pinned Dragonfly extension interface.
- Static custom block/item metadata without hot-path FFI.
- Dynamic block/item callback parity against equivalent Go types.
- Custom entity tick, persistence, despawn, and reload.
- Chunk generator correctness and calls-per-chunk.
- Provider error and concurrency behavior.
- Fixture plugin loaded through the complete Go-to-Rust path.

Required benchmarks:

- Raw Go `NopHandler` baseline.
- Runtime loaded with zero subscribers.
- One no-op Rust movement subscriber.
- Ten no-op Rust movement subscribers.
- Rust movement handler with real validation.
- 2,000, 20,000, and 100,000 movement events per second.
- Go allocations per event and latency distribution.

Target hot path properties:

- Zero FFI calls when no plugin subscribes.
- One cgo call per subscribed event, independent of plugin count.
- No heap allocation for fixed-size events such as movement.
- No global runtime lock during event dispatch.
- Measured overhead documented against raw Go baseline.

Exact equality with raw Go is impossible because FFI has a boundary. Acceptance should be based on measured server impact, not a claim of zero overhead.

## Delivery plan

### Phase 0: performance spike

- Build minimal Go-to-Rust runtime call.
- Load one and ten fixture plugins.
- Implement no-op and validating movement callbacks.
- Run required benchmarks.
- Confirm one-call runtime design before expanding API.

### Phase 1: schema and ABI foundation

- Define schema format and stable event IDs.
- Build ABI generator.
- Generate C header, Rust `sys` types, safe Rust types, dispatch, and layout tests.
- Add ABI version negotiation and feature bits.
- Add Dragonfly public-interface scanner and coverage report.

### Phase 2: runtime and SDK

- Add manifest discovery and dependency ordering.
- Add native plugin loading without unload.
- Add safe `Plugin` trait and export macro.
- Add panic guards, logging, priorities, and subscription tables.
- Add generated `ServerBuilder`, `ServerPlan`, and world SDK types.

### Phase 3: owned server lifecycle

- Implement server executable and configuration.
- Own `Config.New`, `Listen`, `Accept`, and shutdown.
- Implement bootstrap-plan collection and validation.
- Implement host manager and automatic player attachment.
- Implement composite player and world handlers.
- Never retain transaction-scoped objects.

### Phase 4: managed worlds

- Register and configure core worlds automatically.
- Add namespaced plugin-world registry.
- Add built-in provider and generator presets.
- Add create, lookup, save, unload, evacuation, and player-transfer operations.
- Install dispatchers before any world becomes plugin-visible.

### Phase 5: first useful event API

- Add join, quit, movement, chat, hurt, heal, block break, and block place.
- Add exact cancellation/mutation mapping.
- Add event-scoped message action.
- Ship chat-filter and movement-guard examples.

### Phase 6: commands and actions

- Add raw-argument commands.
- Add generated action list representation.
- Add transaction-safe teleport, block, sound, and particle actions.

### Phase 7: Dragonfly capability parity

- Implement custom block and item descriptors, assets, proxy families, and behavior callbacks.
- Implement custom entity base/ticking/living families and persistent state codec.
- Implement chunk-level Rust generators and provider callbacks.
- Implement remaining registry and capability families from scanner report.
- Reach complete public-extension coverage for pinned Dragonfly version.

### Phase 8: hardening and packaging

- Add Linux and macOS CI.
- Publish Rust SDK crates.
- Document supported targets and toolchains.
- Add ABI inspection CLI.
- Pin supported Dragonfly commit/version and test upgrades explicitly.

## First implementation milestone

Deliver only:

1. Schema generator.
2. One Rust runtime.
3. Server executable owning Dragonfly lifecycle.
4. World manager owning the three core worlds and one plugin-created flat world.
5. One FFI movement event.
6. One FFI chat event.
7. One plugin that declares its world during bootstrap and teleports players into it.
8. Benchmarks against equivalent raw Go handlers.

Do not wrap the entire Dragonfly API before this milestone proves performance, generated bindings, and transaction semantics.
