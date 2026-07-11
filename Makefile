.PHONY: generate check-generated build-native test benchmark clean

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
RUNTIME_LIBRARY := libdragonfly_plugin_runtime.dylib
PLUGIN_LIBRARY := libmovement_guard_plugin.dylib
else
RUNTIME_LIBRARY := libdragonfly_plugin_runtime.so
PLUGIN_LIBRARY := libmovement_guard_plugin.so
endif

generate:
	go run ./tools/abi-gen -root .
	cargo fmt --all

check-generated:
	go run ./tools/abi-gen -root . -check
	cargo fmt --all -- --check

build-native: generate
	cargo build --release -p dragonfly-plugin-runtime -p movement-guard-plugin
	mkdir -p build/native build/plugins
	cp target/release/$(RUNTIME_LIBRARY) build/native/
	cp target/release/$(PLUGIN_LIBRARY) build/plugins/

test: build-native check-generated
	cargo test --workspace
	go test ./...

benchmark: build-native
	go test ./go/internal/native -run '^$$' -bench . -benchmem

clean:
	cargo clean
	rm -rf build

