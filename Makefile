.PHONY: generate check-generated build-native build-server build stage-examples run test benchmark clean

UNAME_S := $(shell uname -s)
MOVEMENT_MANIFEST := examples/plugins/movement-guard/Cargo.toml
CHAT_MANIFEST := examples/plugins/chat-filter/Cargo.toml
ifeq ($(UNAME_S),Darwin)
RUNTIME_LIBRARY := libdragonfly_plugin_runtime.dylib
PLUGIN_LIBRARY := libmovement_guard_plugin.dylib
CHAT_PLUGIN_LIBRARY := libchat_filter_plugin.dylib
else
RUNTIME_LIBRARY := libdragonfly_plugin_runtime.so
PLUGIN_LIBRARY := libmovement_guard_plugin.so
CHAT_PLUGIN_LIBRARY := libchat_filter_plugin.so
endif

generate:
	go run ./tools/abi-gen -root .
	cargo fmt --all
	cargo fmt --manifest-path $(MOVEMENT_MANIFEST)
	cargo fmt --manifest-path $(CHAT_MANIFEST)

check-generated:
	go run ./tools/abi-gen -root . -check
	cargo fmt --all -- --check
	cargo fmt --manifest-path $(MOVEMENT_MANIFEST) -- --check
	cargo fmt --manifest-path $(CHAT_MANIFEST) -- --check

build-native: generate
	cargo build --release -p dragonfly-plugin-runtime
	cargo build --release --manifest-path $(MOVEMENT_MANIFEST)
	cargo build --release --manifest-path $(CHAT_MANIFEST)
	mkdir -p build/lib build/plugins
	cp target/release/$(RUNTIME_LIBRARY) build/lib/
	cp examples/plugins/movement-guard/target/release/$(PLUGIN_LIBRARY) build/plugins/
	cp examples/plugins/chat-filter/target/release/$(CHAT_PLUGIN_LIBRARY) build/plugins/

build-server:
	mkdir -p build
	go build -o build/bedrock-gophers ./cmd/bedrock-gophers

build: build-native build-server

stage-examples: build-native
	mkdir -p examples/lib examples/plugins
	rm -f examples/lib/*.so examples/lib/*.dylib examples/lib/*.dll
	cp build/lib/$(RUNTIME_LIBRARY) examples/lib/
	cp build/plugins/$(PLUGIN_LIBRARY) examples/plugins/
	cp build/plugins/$(CHAT_PLUGIN_LIBRARY) examples/plugins/

run: stage-examples
	go run ./cmd/bedrock-gophers -config examples/server.toml

test: build-native check-generated
	cargo test --workspace
	cargo test --manifest-path $(MOVEMENT_MANIFEST)
	cargo test --manifest-path $(CHAT_MANIFEST)
	go test ./...

benchmark: build-native
	go test ./internal/native -run '^$$' -bench . -benchmem

clean:
	cargo clean
	cargo clean --manifest-path $(MOVEMENT_MANIFEST)
	cargo clean --manifest-path $(CHAT_MANIFEST)
	rm -rf build
	rm -rf examples/lib
	rm -f examples/plugins/$(PLUGIN_LIBRARY) examples/plugins/$(CHAT_PLUGIN_LIBRARY)
