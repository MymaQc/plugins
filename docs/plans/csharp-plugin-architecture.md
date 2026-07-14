# C# plugin architecture

## Direction

- C# NativeAOT is the only plugin language.
- The public namespace mirrors Dragonfly's packages and exported types as closely as C# permits.
- Plugins subclass `Plugin`; generated build plumbing supplies the native entry point and project-name identity.
- The Go host owns Dragonfly and exposes a private flat C ABI. Plugins never use ABI types.
- Code generation reads the pinned Dragonfly Go source with `go/ast` and emits C#; there is no second public API schema.

## Shape

```text
Dragonfly Go API -> Go AST generator -> C# Dragonfly API
                                         |
                                         v
plugin source -> NativeAOT .so -> private C ABI -> Go host -> Dragonfly
```

The ABI is transport, not the API. C# names, interfaces, constructors, and behavior should come from Dragonfly. Hand-written code is limited to marshalling semantics that cannot be inferred from Go types.

## Order

1. NativeAOT loading and `OnEnable`/`OnDisable`.
2. `player.Handler` events. Movement, chat, food loss, jump, teleport, sprint/sneak toggles, punch-air, and quit are implemented.
3. Player methods and commands. Command interfaces and the implemented `Player` method surface are generated from Dragonfly's Go AST. C# uses `Cmd.New`/`Cmd.Register`, one `Cmd.Runnable` per overload, and reflected public fields as Dragonfly uses reflected Go struct fields. Supported command fields include subcommands, native enums, dynamic `Cmd.Enum` values, players, vectors, optional values, and `Cmd.Varargs`. The generator roots runnable fields and field types for NativeAOT; runnable types use `internal` visibility and require no linker annotations. Bedrock-facing subcommands and enum/player suggestions are always lowercase. `Player.SetGameMode(World.GameMode)` and Dragonfly's four built-in game-mode values are available. The text slice includes `Message`, `SendPopup`, `SendTip`, `SendJukeboxPopup`, `SetNameTag`, and `Disconnect`. `Messagef` remains absent until Go `fmt.Sprintf` semantics can be preserved honestly.
4. World and block parity. The first landed slice generates `Cube.Pos`, `Cube.Range`,
   `Cube.Face`, `World.Block`, `World.SetOpts`, `World.Tx.Range`, `World.Tx.Block`,
   `World.Tx.BlockLoaded`, `World.Tx.BlocksWithin`, `World.Tx.SetBlock`,
   `World.Tx.ScheduleBlockUpdate`,
   `World.Tx.HighestLightBlocker`, `World.Tx.HighestBlock`, `World.Tx.Light`,
   `World.Tx.SkyLight`, `World.Liquid`, `World.Tx.Liquid`, `World.Tx.SetLiquid`, 79 stateless
   block types, `Block.Sand`, `Block.Water`, and `Block.Lava`. The biome slice adds
   `World.Biome`, all 88 registered vanilla biome types, `SetBiome`, `Biome`, `Temperature`,
   `RainingAt`, `SnowingAt`, `ThunderingAt`, `Raining`, `Thundering`, and `CurrentTick`.
   Transaction method signatures and parameter names come from Dragonfly's `world.Tx` Go AST;
   `BlockLoaded` preserves Dragonfly's non-loading query through a C# tuple. `BlocksWithin` maps
   `iter.Seq[cube.Pos]` to lazy `IEnumerable<Cube.Pos>` backed by a transaction-scoped native
   iterator; it is not materialised into a snapshot. `Player.Context : World.Context : World.Tx`,
   so event handlers use world operations directly just like embedded Dragonfly contexts.
   Registered block/liquid states and registered biomes supply canonical private codecs; public
   plugins never handle identifiers, NBT, numeric biome IDs, world handles, iterator handles, or
   host errors. `Liquid` preserves Dragonfly's `(Liquid, bool)` result, and passing `null` to
   `SetLiquid` removes the liquid. Host
   ABI 27 transports that distinction, signed `time.Duration` nanoseconds, private biome IDs, and
   the transaction owner's current tick without exposing them publicly. `ScheduleBlockUpdate`
   maps Go `time.Duration` to C# `TimeSpan`
   and preserves the transaction-owned call. `BuildStructure` remains absent until its synchronous `At` and
   `blockAt` callbacks can be implemented without materialising or changing Dragonfly semantics.
   Remaining stateful blocks, structures, world methods, custom blocks, and block models land
   incrementally.
5. Items, forms, entities, particles, sounds, and remaining world/block methods.
6. Convert practice-core and expand parity tests against Dragonfly.

Each slice removes the replaced legacy implementation. Unsupported API remains absent rather than gaining a parallel abstraction.

Block access belongs on `World.Tx`. Do not revive `WorldSync`, expose the framework's
`WorldManager` or namespaced world IDs, or restore schema/block-gen as a second public model. Go
AST owns exported shape; Dragonfly's live registry supplies encoded state bytes; the C ABI only
transports calls.

`examples/plugins/kitchen-sink` must use every exposed API. Its `/kitchen block` overload reads the
block below the source through `World.Tx`, writes typed `Block.Sand`, and exercises all three
`World.SetOpts` flags, reads the world range, performs a non-loading block lookup, lazily searches
nearby blocks, reads height/light data, inspects typed water, and writes/removes typed liquid.
It also leaves typed water present and schedules its update with an exact 250 ms delay.
Its separate `/kitchen biome` overload changes and restores a typed biome while exercising every
temperature and weather query in this slice.
`/kitchen tick` reads Dragonfly's transaction-owned current tick; it does not alias world day-time.
NativeAOT and host-call tests verify the public shape, lazy iterator cleanup, and transaction-safe
transport.

`examples/plugins/vanilla-commands` keeps its plugin entry tiny and one command per file. It currently exercises `/gamemode`, `/help`, `/ping`, and `/position`, and expands with each gameplay parity slice.

Practice remains out of the parity loop until the framework API needed by practice exists. Feature work lands and is tested in this repository first.
