# bedrock-gophers/plugins

Native multi-language plugin runtime for [df-mc/dragonfly](https://github.com/df-mc/dragonfly). Rust is the first supported plugin language.

Current status: native runtime foundation plus player actions, typed items, and player inventory handles. Generated events, lifecycle hooks, Dragonfly commands, item snapshots, and synchronous host actions travel through Go, the native Rust runtime, and dynamically loaded Rust plugins.

## Build and test

Requirements:

- Go 1.26+
- Rust 1.96+
- C compiler and `dlopen` support

```shell
make test
make benchmark
```

Build and run owned server:

```shell
make run
```

`make run` compiles example Rust plugins, stages them under `examples/plugins`, then runs root Go command using `examples/server.toml`. Framework derives native runtime library path for current platform; config contains no `.so` path.

Framework creates Dragonfly, installs world/player handlers, owns accept loop, and closes cleanly on `SIGINT`/`SIGTERM`.

Framework world manager protects `minecraft:overworld`, `minecraft:nether`, and `minecraft:end`; custom worlds use namespaced IDs such as `example:lobby`. It installs handlers before publication and owns save/unload cleanup.

Regenerate ABI files after changing `schema/`:

```shell
make generate
```

## Rust plugin example

```rust
use dragonfly::{Event, Plugin, plugin};

#[derive(Default)]
struct MovementGuard;

#[plugin]
impl Plugin for MovementGuard {
    fn on_move(&self, event: &mut Event::PlayerMove<'_>) {
        if event.new_position().y < 0.0 {
            event.cancel();
        }
    }
}
```

Events continue by default. Cancellation is monotonic; no `allow()` API exists.
Plugin identity defaults to Cargo's package name; handler code does not repeat it.

Event types live only under `Event::Player*`. Damage and healing sources are typed values: hurt/death expose `damage_source()`, while heal exposes `healing_source()`.

Items are owned Rust values. Inventory handles stay attached to the generation-tagged player:

```rust
use dragonfly::{Enchantment, Player, item};

fn give_sword(player: Player) {
    let sword = item::new(item::Sword::new(item::ToolTier::Diamond), 1)
        .with_custom_name("Rust Sword")
        .with_lore(["Created by a native plugin"])
        .with_value("plugin", "example")
        .with_enchantment(Enchantment::Sharpness, 5);
    player.inventory().add_item(&sword);
}
```

`Player::inventory()`, `armour()`, and `offhand()` expose get/set/add/clear operations. Mutating setters are fire-and-forget; host transport statuses stay internal. `add_item()` returns only the useful domain result: the count added. `held_items()`, `set_held_items()`, and `set_held_slot()` mirror Dragonfly. Inventory reads and item events preserve count, metadata, damage, unbreakable state, anvil cost, custom name, lore, NBT, `WithValue` data, and enchantments through bounded snapshots. NBT uses standard little-endian encoding, not Go `gob` bytes.

Built-in item identities are typed and mirror Dragonfly's item model: `item::new(item, count)` creates the stack, and metadata belongs to the item type. `item::Custom` is the explicit escape hatch for plugin-registered identifiers.

Commands use compile-time macros in place of Go runtime reflection. `#[command("root")]` declares the command, and each `#[subcommand("name")]` method becomes a Dragonfly runnable with generated native metadata and parsing. See the hello-command example for general command arguments and the items-command example for inventory operations.

See [native plugin architecture](docs/plans/rust-plugin-architecture.md).

## Examples

- [Movement guard](examples/plugins/movement-guard): cancels movement below Y=0.
- [Chat filter](examples/plugins/chat-filter): replaces text and cancels a blocked message.
- [Lifecycle logger](examples/plugins/lifecycle-logger): demonstrates enable and disable hooks.
- [Hello command](examples/plugins/hello-command): demonstrates Dragonfly subcommands and enum parameters.
- [Items command](examples/plugins/items-command): demonstrates typed items and inventory reads/writes.

The examples compile as native plugin libraries through `make stage-examples`. Precompiled `.so`, `.dylib`, or `.dll` plugins may also be placed directly in `examples/plugins`.
