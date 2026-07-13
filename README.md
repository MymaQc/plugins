# bedrock-gophers/plugins

Native multi-language plugin runtime for [df-mc/dragonfly](https://github.com/df-mc/dragonfly). Rust is the first supported plugin language.

Current status: native runtime foundation plus player actions, typed items and blocks, inventories, scoreboards, typed asynchronous forms, managed worlds, stable entity handles, built-in entity/projectile spawning, persistent plugin-owned base/ticking/living entities, typed world particles and sounds, and entity events. Generated events, lifecycle hooks, Dragonfly commands, bounded snapshots, and host actions travel through Go, the native Rust runtime, and dynamically loaded Rust plugins.

Rust mirrors Dragonfly's `Messagef` convenience with Rust formatting arguments:

```rust
player.messagef(format_args!("Welcome, {}!", player.name().unwrap_or("player")));
```

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

Rust can look up core worlds, open persistent custom worlds with typed creation policies, read/write typed block states, and manage time/spawn:

```rust
use dragonfly::{
    BlockPos, OpenMode, RandomTicks, SavePolicy, TimePolicy, WeatherPolicy, World, WorldSpec,
    block,
};

let spec = WorldSpec::persistent("arenas/example")
    .open_mode(OpenMode::OpenOrCreate)
    .save(SavePolicy::Manual)
    .random_ticks(RandomTicks::Disabled)
    .time(TimePolicy::Fixed(6000))
    .weather(WeatherPolicy::Clear);
let world = World::open_with("example:arena", &spec);
if let Some(world) = world {
    let pillar = block::OakLog::new(block::PillarAxis::Y);
    world.set_block(BlockPos { x: 0, y: 64, z: 0 }, pillar);
}
```

`WorldSpec::persistent` defaults to overworld, open-or-create, writable, automatic saves every ten minutes, three random ticks per subchunk, preserved time/weather, and two-minute chunk unloading. Provider paths are slash-separated and relative to `worlds.directory`; absolute, escaping, or existing symlink paths are rejected. This organization is not a security sandbox—native plugins already run with the server process's filesystem access.

Specifications are immutable after publication. The same world name and normalized specification returns the same handle, a mismatch fails, and two names cannot own one provider path. `OpenExisting` requires an existing MCDB provider, while `CreateNew` rejects anything already at the path. Read-only always canonicalizes to manual saving regardless of builder order. Opens are synchronous and valid from lifecycle or event callbacks; no callback transaction is retained. Invalid input and host/provider failures remain private as `None`.

World handles are opaque and never reused. `World::open(name, dimension)` remains the default convenience API and derives its path from the namespaced ID. Every callback carries an opaque invocation ID, so same-world operations reuse exactly that callback's Dragonfly `world.Tx`, never another concurrent callback's transaction. Off-owner writes use `World.Do`. Synchronous cross-world block reads are unavailable inside callbacks until the task API lands, preventing reciprocal world-owner deadlocks.

Entities use generation-tagged `world.EntityHandle` identities. Typed descriptors keep Go adapter code shared:

```rust
use dragonfly::{Vec3, World, entity, item};

#[dragonfly::entity(
    network = "minecraft:armor_stand",
    width = 0.5,
    height = 1.975,
)]
struct Marker;

if let Some(world) = World::overworld() {
    let options = entity::SpawnOptions::new(Vec3 { x: 0.0, y: 65.0, z: 0.0 });
    let text = world.spawn_entity(entity::Text::new("Hello"), options.clone());
    let sword = item::new(item::Sword::new(item::ToolTier::Diamond), 1);
    let dropped = world.spawn_entity(entity::DroppedItem::new(sword), options.clone());
    let marker = world.spawn_entity(Marker, options);
    if let Some(text) = text {
        text.set_name_tag("Updated");
    }
}
```

`#[dragonfly::entity]` registers the type before Dragonfly constructs any world. Put it on a unit struct for a base entity, or on an `impl entity::Ticking` / `impl entity::Living` block for an owned stateful entity. Its save ID defaults to `<cargo-package>:<snake_case_type>`; use `id = "namespace:name"` only when that persisted identity must survive crate or type renames. Advanced entities persist versioned Rust state plus Dragonfly health, effects, speed, movement, and common entity data. See the entity-command training dummy for a complete minimal implementation.

`World::entities()` and `World::players()` resolve within the current transaction. `Entity` exposes its managed world, type, position, rotation, optional velocity/name tag, teleport, velocity/name-tag mutation, and despawn. Dragonfly v0.11 has no exported generic rotation setter, so rotation mutation is deliberately absent rather than implemented with reflection. Projectiles use typed owner handles (`Arrow`, `Snowball`, `Egg`, `EnderPearl`, bottles, and potions). Dragonfly has no global pre-impact projectile hook; a correct cancellable projectile-hit event needs an upstream hook and is not faked by rebuilding private projectile behaviour.

Dragonfly v0.11 has no generic living-entity hurt/death handler. Framework-owned custom living entities therefore emit exact cancellable `Event::EntityHurt`, `Event::EntityHeal`, and `Event::EntityDeath` callbacks from their own implementation; arbitrary third-party living types still require an upstream Dragonfly hook. Despawn is never treated as death.

Synchronous entity spawning inside a callback must target that callback's current world. Cross-world spawning will return an asynchronous handle task when the task API lands; off-callback code may already spawn in any managed world.

Particles mirror Dragonfly's concrete types rather than inventing string identifiers:

```rust
use dragonfly::{Vec3, World, block, particle};

if let Some(world) = World::overworld() {
    world.add_particle(Vec3 { x: 0.0, y: 65.0, z: 0.0 }, particle::HugeExplosion);
    world.add_particle(
        Vec3 { x: 0.0, y: 65.0, z: 0.0 },
        particle::BlockBreak::new(block::DiamondBlock),
    );
}
```

Sounds use the same typed model for private player playback and transaction-aware world playback:

```rust
use dragonfly::{Vec3, World, block, item, sound};

player.play_sound(sound::LevelUp);
if let Some(world) = World::overworld() {
    world.play_sound(
        Vec3 { x: 0.0, y: 65.0, z: 0.0 },
        sound::DoorOpen::new(
            block::WoodenDoor::new()
                .with_cardinal_direction(block::CardinalDirection::North),
        ),
    );
    world.play_sound(
        Vec3 { x: 0.0, y: 65.0, z: 0.0 },
        sound::EquipItem::new(item::Sword::new(item::ToolTier::Diamond)),
    );
}
```

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

Event types live only under `Event::Player*`. Damage and healing sources are typed values: hurt/death expose matchable `damage::Source` values, while heal exposes `healing::Source`. Attack, projectile, thorns, block, poison, food, and every other Dragonfly source retain their concrete payloads. `Event::PlayerChangeWorld` is emitted after transfer on the first destination tick. `Event::PlayerRespawn` runs before transfer and may replace both the spawn position and managed destination world. `Event::PlayerSkinChange` exposes the proposed skin through `skin()`, `set_skin()`, and `edit_skin()` and remains cancellable/default-allowed. `event.player().skin()` still returns the old committed skin, matching Dragonfly.

```rust
fn on_skin_change(&self, event: &mut Event::PlayerSkinChange<'_>) {
    if event.skin().width > 128 {
        event.cancel();
    }
    event.edit_skin(|skin| skin.model_config.default = "geometry.humanoid.custom".into());
}
```

Player healing and damage use the same typed sources and return Dragonfly's domain results:

```rust
use dragonfly::{Entity, Player, damage, healing};

fn combat(player: Player, attacker: Entity) {
    let healed = player.heal(4.0, healing::Food::new(false));
    let (damage, vulnerable) = player.hurt(6.0, damage::Attack::new(attacker));
    eprintln!("healed={healed} damage={damage} vulnerable={vulnerable}");
}
```

Transport failures remain private. `heal` returns actual health gained. `hurt` returns Dragonfly's final reduced damage and vulnerability result.

Effects keep Dragonfly's lasting/instant type split and constructor shape:

```rust
use dragonfly::{Player, effect};

fn apply_effects(player: Player) {
    player.add_effect(effect::new(
        effect::Speed,
        2,
        std::time::Duration::from_secs(30),
    ));
    player.add_effect(effect::instant_with_potency(
        effect::InstantHealth,
        1,
        0.5,
    ));
    player.remove_effect(effect::Speed);
}
```

`effect::RegisteredLasting` and `effect::RegisteredInstant` reference custom IDs already registered in Dragonfly. The host checks their actual Go kind. Invalid levels or mismatched kinds silently no-op instead of exposing transport errors or panicking the server.

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

Scoreboards are owned snapshots with Dragonfly's 15-line limit, padding, and display-order controls:

```rust
use dragonfly::{Player, Scoreboard, ScoreboardLineOutOfBounds};

fn show_scoreboard(player: Player) -> Result<(), ScoreboardLineOutOfBounds> {
    let mut board = Scoreboard::new("Match");
    board.push_line("Red: 3")?.push_line("Blue: 2")?;
    player.send_scoreboard(&board);
    Ok(())
}
```

`send_scoreboard()` and `remove_scoreboard()` are fire-and-forget; native host transport failures remain internal.

Forms are concrete Dragonfly-style `Menu`, `Modal`, and `Custom` types, including every v0.11 element. Their responses are typed and use owned one-shot callbacks because Dragonfly answers asynchronously. `form::Raw` is the explicit JSON escape hatch for experimental Bedrock form shapes:

```rust
use dragonfly::{Player, form};

fn choose(player: Player) {
    let mut menu = form::Menu::new("Choose").body("Pick an option.");
    let first = menu.button(form::Button::new("First"));
    player.send_form(menu, move |player, response| {
        if response.is_some_and(|response| response.selected() == first) {
            player.message("First selected");
        }
    });
}
```

`send_form()` and `close_form()` are fire-and-forget. Submit and close callbacks run inside Dragonfly's response transaction; disconnect, disable, and shutdown safely discard pending callbacks before plugin libraries unload.

Built-in item identities are typed and mirror Dragonfly's item model: `item::new(item, count)` creates the stack, and metadata belongs to the item type. `item::Custom` is the explicit escape hatch for plugin-registered identifiers.

Commands use compile-time macros in place of Go runtime reflection. `#[command("root")]` declares the command, and each `#[subcommand("name")]` method becomes a Dragonfly runnable with generated native metadata and parsing. See the hello-command example for general command arguments and the items-command example for inventory operations.

See [native plugin architecture](docs/plans/rust-plugin-architecture.md).

## Examples

- [Movement guard](examples/plugins/movement-guard): cancels movement below Y=0.
- [Chat filter](examples/plugins/chat-filter): replaces text and cancels a blocked message.
- [Lifecycle logger](examples/plugins/lifecycle-logger): demonstrates enable/disable hooks and the complete bridged player-event set.
- [Hello command](examples/plugins/hello-command): demonstrates Dragonfly subcommands and enum parameters.
- [Player command](examples/plugins/player-command): demonstrates player movement, state, effects, damage/healing, visibility, skin, sound, and disconnect actions.
- [Items command](examples/plugins/items-command): demonstrates typed items and inventory reads/writes.
- [Scoreboard](examples/plugins/scoreboard): sends and removes a sidebar scoreboard.
- [Forms](examples/plugins/forms): demonstrates menu, modal, and typed custom-form responses.
- [World command](examples/plugins/world-command): demonstrates world lookup/open, blocks, time, and spawn.
- [Entity command](examples/plugins/entity-command): demonstrates typed built-in/projectile spawning, persistent base and living entities, plugin state, handles, and world lists.
- [Particle command](examples/plugins/particle-command): demonstrates every typed built-in particle family.
- [Sound command](examples/plugins/sound-command): demonstrates private player playback and typed world sounds.

The examples compile as native plugin libraries through `make stage-examples`. Precompiled `.so`, `.dylib`, or `.dll` plugins may also be placed directly in `examples/plugins`.
