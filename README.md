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

Current C# slice: loading, lifecycle, reflected commands, player text actions, game mode, movement, chat, food loss, jump, teleport, sprint/sneak toggles, punch-air, and quit handlers. Player handler, command-interface, and player-text surfaces are generated from Dragonfly's Go AST. `examples/plugins/kitchen-sink` compiles against every exposed C# API and grows with each parity slice.

World callbacks now carry Dragonfly-shaped transactions. `Player.Context` inherits
`World.Context`, which inherits `World.Tx`; commands receive the same `World.Tx`. `Cube.Pos`,
`World.Block`, `World.SetOpts`, `World.Tx.Range`, `World.Tx.Block`, `World.Tx.BlockLoaded`,
`World.Tx.SetBlock`, 79 stateless concrete block types, and `Block.Sand` are generated from
Dragonfly source and its registered states:

```csharp
var pos = Cube.PosFromVec3(source.Position()).Side(Cube.Face.Down);
var (block, loaded) = tx.BlockLoaded(pos);
World.Block? previous = loaded ? block : tx.Block(pos);
Cube.Range bounds = tx.Range();
tx.SetBlock(pos, new Block.Sand());
```

Minecraft identifiers, state NBT, world handles, and ABI errors remain private transport details.
