# Phase 0 Results

Status: Initial movement slice complete

Date: 2026-07-11

Implemented path:

```text
Go wrapper
  -> C bridge loaded with dlopen
  -> Rust runtime
  -> dynamically loaded Rust plugin
  -> mutable cancellation state
  -> Go wrapper
```

Implemented:

- YAML event schema.
- Generator producing canonical C ABI and raw Rust types.
- ABI layout test.
- Rust runtime and dynamic plugin discovery.
- ABI version, struct-size, plugin-ID, and subscription validation.
- Duplicate plugin-ID rejection.
- Safe Rust movement event.
- Attribute macro generating subscription and unsafe ABI entry code.
- Default-continue and monotonic-cancellation behavior.
- End-to-end Go test using release native libraries.
- Zero-allocation scalar movement bridge.

Benchmark command:

```shell
make benchmark
```

Observed on AMD Ryzen 5 7640U, Linux amd64:

```text
BenchmarkRawGoMovement       2.736 ns/op   0 B/op   0 allocs/op
BenchmarkNativeRustMovement 56.54 ns/op   0 B/op   0 allocs/op
```

Native benchmark includes Go-to-C crossing, Rust runtime dispatch, safe SDK handler, and return to Go. It does not yet include construction of a real Dragonfly `player.Handler` context.

Next slice:

1. Integrate movement dispatcher with Dragonfly's `player.Handler`.
2. Add generated chat input and mutable message state.
3. Move generated subscription detection beyond the single movement event.
4. Start framework-owned server lifecycle and automatic player/world management.

