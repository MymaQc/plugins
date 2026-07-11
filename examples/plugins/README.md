# Example plugin output

Each plugin owns its `Cargo.toml`, `Cargo.lock`, source, and target directory. `make run` builds each plugin manifest and stages native libraries in this directory.

Precompiled native plugin libraries may also be placed in this directory. Runtime loads both compiled examples and compatible precompiled plugins.
