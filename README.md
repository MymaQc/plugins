# bedrock-gophers/plugins

C# NativeAOT plugins for [df-mc/dragonfly](https://github.com/df-mc/dragonfly).

The public C# API follows Dragonfly. The native ABI is internal plumbing, and generated C# API code is derived from Dragonfly's Go AST rather than a separate public schema.

## Run

Requirements: Go 1.26+, .NET 10 SDK, and a C compiler/linker.

```shell
make run
```

Plugins live in `plugins/`. Source projects are compiled to NativeAOT shared libraries; compatible precompiled `.so` files can be placed there too.

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
        Console.WriteLine($"{player.Name} quit");
    }
}
```

The project name is the plugin ID. A compile-time generator emits the hidden native entry point.

Current C# slice: loading, lifecycle, and Dragonfly's `player.Handler.HandleQuit`. Remaining Dragonfly surface is being ported directly rather than preserved through the removed Rust abstraction.
