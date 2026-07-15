# Kitchen sink

Runnable parity showcase. Its single plugin class demonstrates:

- movement validation and cancellation;
- chat and food-state mutation;
- all 37 Dragonfly player handlers, including typed damage/healing sources, worlds, entities,
  mutable skins and respawns, transfer addresses, command arguments, diagnostics, typed block/item
  payloads, mutable break drops, item-pickup replacement, signed durations, and full snapshots;
- every package-shaped Dragonfly damage/healing source plus direct `Player.Hurt` and `Player.Heal`;
- concrete `Player` targets and `Player.Name()` in attack and entity-use handlers;
- all 13 Dragonfly world handlers, including typed liquids/sounds/blocks/entities, entity
  spawn/despawn, mutable explosion lists and scalars, full redstone updates, and world close;
- decoded gophertunnel client/server packet interception, callback-scoped typed packet properties,
  incoming mutation and cancellation, and safe outgoing inspection;
- the cancellable host `OnJoin` lifecycle extension emitted after the framework installs its
  Dragonfly player handler;
- every zero-argument player action plus exact typed block interactions in `/kitchen actions`;
- exact `UseItemOnEntity` and `AttackEntity` results against a selected player in
  `/kitchen entity-actions`;
- reflected command arguments and overloads;
- direct `Player.Message` output;
- transaction-safe range, lazy block/entity/player/bounding-box iteration, current-world lookup, height/light
  queries, and block access through `World.Tx`;
- direct lazy server-wide player iteration, global messaging, stable handle retention, and UUID/name
  lookup in `/kitchen server`;
- generated `Cube.Pos`, `Block.Sand`, all `World.SetOpts` flags, and registry-backed
  `World.BlockByName` lookups covering bool/byte/int/string state properties in `/kitchen block`;
- promoted Dragonfly crop state through typed `Block.WheatSeeds(Growth: 7)` in `/kitchen crop`;
- typed `World.Liquid`, `Block.Water`, liquid inspection, placement, and nullable removal;
- typed `ScheduleBlockUpdate` with a matching water ticker and C# `TimeSpan` delay;
- all 88 generated vanilla biome types plus biome, temperature, and weather queries in
  `/kitchen biome`;
- transaction-owned `CurrentTick` in `/kitchen tick`;
- all 20 generated particle types and all 16 typed note instruments in `/kitchen particle`;
- exact world-broadcast and player-only sound playback across every payload family in `/kitchen sound`;
- registered lookup, player reads, and custom `World.GameMode` in `/kitchen game-mode`;
- exact hide/show and all six per-viewer name-tag, score-tag, and visibility overrides in
  `/kitchen visibility`;
- food, health, experience, scale, visibility, and mobility state round-trips in `/kitchen state`;
- direct typed healing in `/kitchen heal`;
- in-memory `World.New()`, writable MCDB-backed `World.Config.New()`, owner scheduling,
  AST-generated world name/dimension/range/height/time/time-cycle/spawn/sleep/game-mode/tick-range/
  difficulty/save methods, custom world game mode and difficulty transport, and safe cross-world
  player movement in `/kitchen world`; the command is repeatable after entering its arena and only
  performs chunk-backed height reads while it owns that arena's transaction;
- generated typed items (including finite stateful families, NBT-backed books, typed fireworks,
  seven armour tiers, four armour pieces, and 11 trim materials), firework explosions and shapes,
  armour defence/durability/repair/smelting behavior, all 28 armour round-trips, private dyed/trim
  NBT, stack values, typed Protection/Unbreaking enchantments, normal compatibility filtering,
  main/ender-chest inventories, armour slots, and held items in `/kitchen item`;
- the full reflected form API in `/kitchen form`: menu, custom, and modal callbacks; every
  element type; URL and resource-pack button images; tooltips; closers; and submitted values;
- open `Form.Value`, `Element.MarshalJSON`, and `MenuElement.MarshalJSON` implementations in
  `/kitchen raw-form`;
- all 28 registered effect types, constructors, value methods, potion/stew effect lists, and a
  player effect round trip in `/kitchen effect`;
- plugin lifecycle.

New APIs belong here only when the example can do something real with them. This plugin is not a
compile-time API dumping ground.
