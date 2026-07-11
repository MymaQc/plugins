# Movement guard

Minimal cancellable movement handler. Movement continues by default; positions below Y=0 are cancelled.

```shell
cargo build --release -p movement-guard-plugin
```

Output: `target/release/libmovement_guard_plugin.so` on Linux.
