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
3. Player methods and commands.
4. Worlds, items, blocks, forms, entities, particles, and sounds.
5. Convert practice-core and expand parity tests against Dragonfly.

Each slice removes the replaced legacy implementation. Unsupported API remains absent rather than gaining a parallel abstraction.

`examples/plugins/kitchen-sink` must use every exposed API. Its NativeAOT build is the compile-time parity canary for handlers, commands, worlds, items, blocks, forms, entities, particles, and sounds as those slices land.
