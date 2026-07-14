# Kitchen sink

Runnable parity showcase. Its single plugin class demonstrates:

- movement validation and cancellation;
- chat and food-state mutation;
- player-action telemetry;
- reflected command arguments and overloads;
- direct `Player.Message` output;
- transaction-safe range, lazy block search, height/light queries, and block access through `World.Tx`;
- generated `Cube.Pos`, `Block.Sand`, and all `World.SetOpts` flags in `/kitchen block`;
- typed `World.Liquid`, `Block.Water`, liquid inspection, placement, and nullable removal;
- typed `ScheduleBlockUpdate` with a matching water ticker and C# `TimeSpan` delay;
- all 88 generated vanilla biome types plus biome, temperature, and weather queries in
  `/kitchen biome`;
- transaction-owned `CurrentTick` in `/kitchen tick`;
- all 20 generated particle types and all 16 typed note instruments in `/kitchen particle`;
- registered lookup, player reads, and custom `World.GameMode` in `/kitchen game-mode`;
- generated typed items (including finite stateful families, NBT-backed books, and typed fireworks),
  firework explosions and shapes, stack metadata, inventories, armour, and held items in
  `/kitchen item`;
- the full reflected form API in `/kitchen form`: menu, custom, and modal callbacks; every
  element type; URL and resource-pack button images; tooltips; closers; and submitted values;
- open `Form.Value`, `Element.MarshalJSON`, and `MenuElement.MarshalJSON` implementations in
  `/kitchen raw-form`;
- plugin lifecycle.

New APIs belong here only when the example can do something real with them. This plugin is not a
compile-time API dumping ground.
