# bedrock-gophers/plugins

Native multi-language plugin runtime for [df-mc/dragonfly](https://github.com/df-mc/dragonfly). Rust is the first supported plugin language.

Current status: architecture spike. One generated movement event travels through Go, the native Rust runtime, and a dynamically loaded Rust plugin.

## Build and test

Requirements:

- Go 1.26+
- Rust 1.96+
- C compiler and `dlopen` support

```shell
make test
make benchmark
```

Regenerate ABI files after changing `schema/`:

```shell
make generate
```

## Rust plugin example

```rust
use dragonfly_plugin::{PlayerMoveEvent, Plugin, plugin};

#[derive(Default)]
struct MovementGuard;

#[plugin(id = "example:movement-guard")]
impl Plugin for MovementGuard {
    fn on_move(&self, event: &mut PlayerMoveEvent<'_>) {
        if event.new_position().y < 0.0 {
            event.cancel();
        }
    }
}
```

Events continue by default. Cancellation is monotonic; no `allow()` API exists.

See [native plugin architecture](docs/plans/rust-plugin-architecture.md).

