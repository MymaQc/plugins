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
- plugin lifecycle.

New APIs belong here only when the example can do something real with them. This plugin is not a
compile-time API dumping ground.
