# bedrock-gophers/plugins

C# NativeAOT plugins for [df-mc/dragonfly](https://github.com/df-mc/dragonfly).

The public C# API follows Dragonfly. The native ABI is internal plumbing, and generated C# API code is derived from Dragonfly's Go AST rather than a separate public schema.

## Run

Requirements: Go 1.26+, .NET 10 SDK, and a C compiler/linker.

```shell
make run
```

`make run` always regenerates the AST-derived C# API, republishes the runtime and every
`examples/plugins/*/*.csproj` as NativeAOT, replaces staged example `.so` files, then starts the
server. It does not reuse an older plugin build.

Consumer repositories keep plugins in `plugins/`; see the `minimal` branch for the local and
Docker setup. The loader also accepts compatible precompiled `.so` files. Source-owned build
commands replace their staged output, so precompiled files are staged after that build or run
directly with the server.

## Plugin

Plugin code does not declare an ID, shared-library path, or native entry point:

```csharp
using Dragonfly;

public sealed class Example : Plugin
{
    public override void OnEnable() => Console.WriteLine("enabled");
    public override void OnDisable() => Console.WriteLine("disabled");

    public override void HandleQuit(Player player)
    {
        Console.WriteLine($"{player.Name()} quit");
    }
}
```

The project name is the plugin ID. A compile-time generator emits the hidden native entry point.

Current C# slice: loading, lifecycle, reflected commands, player text actions, game mode, typed
forms, movement, chat, food loss, jump, teleport, sprint/sneak toggles, punch-air, and quit
handlers. Player handler, command-interface, player-text, game-mode, and form surfaces are
generated from Dragonfly's Go AST.
`World.GameMode` includes Dragonfly's four registered values and exact `GameModeByID`/`GameModeID`
lookups. `Player.SetGameMode` accepts custom implementations just like Dragonfly, and
`Player.GameMode` returns their capabilities without exposing the transport descriptor.
Forms use Dragonfly's reflected public-field model through `Form.New`, `NewMenu`, and `NewModal`,
with typed elements, submitted values, `Closer`, and callback-owned `World.Tx`. `Form.Value`
remains open for custom implementations, matching Dragonfly's public `form.Form` interface.
`examples/plugins/kitchen-sink` compiles against every exposed C# API and grows with each parity
slice.

World callbacks now carry Dragonfly-shaped transactions. `Player.Context` inherits
`World.Context`, which inherits `World.Tx`; commands receive the same `World.Tx`. `Cube.Pos`,
`World.Block`, `World.Liquid`, `World.Biome`, `World.SetOpts`, and the current `World.Tx` block and
biome surface are generated from Dragonfly source. This includes `Range`, `Block`, `BlockLoaded`,
`Liquid`, `SetLiquid`, `ScheduleBlockUpdate`, `HighestLightBlocker`, `HighestBlock`, `Light`, and
`SkyLight`, `CurrentTick`, `AddParticle`, plus 79 stateless concrete block types, `Block.Sand`,
`Block.Water`, `Block.Lava`, all 88 registered vanilla biome types, and all 20 Dragonfly particle
types with typed colours, blocks, faces, positions, and note instruments:

```csharp
var pos = Cube.PosFromVec3(source.Position()).Side(Cube.Face.Down);
var (block, loaded) = tx.BlockLoaded(pos);
World.Block? previous = loaded ? block : tx.Block(pos);
Cube.Range bounds = tx.Range();
var nearby = tx.BlocksWithin(pos, 8, new Block.Sand());
tx.SetBlock(pos, new Block.Sand());
var (liquid, found) = tx.Liquid(pos);
tx.SetLiquid(pos, new Block.Water(Still: true, Depth: 8, Falling: false));
tx.SetLiquid(pos, null); // Remove the liquid.
var water = new Block.Water(Still: true, Depth: 8, Falling: false);
tx.SetLiquid(pos, water);
tx.ScheduleBlockUpdate(pos, water, TimeSpan.FromMilliseconds(250));
var previousBiome = tx.Biome(pos);
tx.SetBiome(pos, new Biome.Desert());
var temperature = tx.Temperature(pos);
var rainingHere = tx.RainingAt(pos);
tx.SetBiome(pos, previousBiome);
var tick = tx.CurrentTick();
tx.AddParticle(source.Position(), new Particle.Flame(new Color.RGBA(255, 96, 32, 255)));
tx.AddParticle(source.Position(), new Particle.Note(Sound.Piano(), 12));
```

`BlocksWithin` stays lazy across the private ABI: each C# enumerator owns a transaction-scoped
Dragonfly iterator and closes it on exhaustion, early exit, or callback completion.

Public block, liquid, biome, particle, colour, and instrument types come from Dragonfly's Go AST.
Live registries feed internal generated codecs, so Minecraft identifiers, state NBT, numeric biome
IDs, particle kinds, and instrument IDs never enter plugin code. Private host ABI 29 preserves the
separate “no liquid” result, nullable liquid removal, signed nanosecond scheduling delays,
biome/weather queries, particle payloads, registered/custom game-mode capabilities, and full
callback-scoped player snapshots for form responses. Structurally valid form contexts receive
exactly one response or drop callback, including synchronous send failures. World handles,
capability descriptors, and ABI errors also remain private transport details.
