.PHONY: generate check-generated build-native build-server build stage-examples run test benchmark clean

UNAME_S := $(shell uname -s)
MOVEMENT_MANIFEST := examples/plugins/movement-guard/Cargo.toml
CHAT_MANIFEST := examples/plugins/chat-filter/Cargo.toml
LIFECYCLE_MANIFEST := examples/plugins/lifecycle-logger/Cargo.toml
COMMAND_MANIFEST := examples/plugins/hello-command/Cargo.toml
PLAYER_COMMAND_MANIFEST := examples/plugins/player-command/Cargo.toml
ITEMS_MANIFEST := examples/plugins/items-command/Cargo.toml
PING_MANIFEST := examples/plugins/ping-command/Cargo.toml
SCOREBOARD_MANIFEST := examples/plugins/scoreboard/Cargo.toml
FORMS_MANIFEST := examples/plugins/forms/Cargo.toml
WORLD_MANIFEST := examples/plugins/world-command/Cargo.toml
ENTITY_MANIFEST := examples/plugins/entity-command/Cargo.toml
PARTICLE_MANIFEST := examples/plugins/particle-command/Cargo.toml
SOUND_MANIFEST := examples/plugins/sound-command/Cargo.toml
ifeq ($(UNAME_S),Darwin)
RUNTIME_LIBRARY := libdragonfly_plugin_runtime.dylib
PLUGIN_LIBRARY := libmovement_guard_plugin.dylib
CHAT_PLUGIN_LIBRARY := libchat_filter_plugin.dylib
LIFECYCLE_PLUGIN_LIBRARY := liblifecycle_logger_plugin.dylib
COMMAND_PLUGIN_LIBRARY := libhello_command_plugin.dylib
PLAYER_COMMAND_PLUGIN_LIBRARY := libplayer_command_plugin.dylib
ITEMS_PLUGIN_LIBRARY := libitems_command_plugin.dylib
PING_PLUGIN_LIBRARY := libping_command_plugin.dylib
SCOREBOARD_PLUGIN_LIBRARY := libscoreboard_plugin.dylib
FORMS_PLUGIN_LIBRARY := libforms_plugin.dylib
WORLD_PLUGIN_LIBRARY := libworld_command_plugin.dylib
ENTITY_PLUGIN_LIBRARY := libentity_command_plugin.dylib
PARTICLE_PLUGIN_LIBRARY := libparticle_command_plugin.dylib
SOUND_PLUGIN_LIBRARY := libsound_command_plugin.dylib
else
RUNTIME_LIBRARY := libdragonfly_plugin_runtime.so
PLUGIN_LIBRARY := libmovement_guard_plugin.so
CHAT_PLUGIN_LIBRARY := libchat_filter_plugin.so
LIFECYCLE_PLUGIN_LIBRARY := liblifecycle_logger_plugin.so
COMMAND_PLUGIN_LIBRARY := libhello_command_plugin.so
PLAYER_COMMAND_PLUGIN_LIBRARY := libplayer_command_plugin.so
ITEMS_PLUGIN_LIBRARY := libitems_command_plugin.so
PING_PLUGIN_LIBRARY := libping_command_plugin.so
SCOREBOARD_PLUGIN_LIBRARY := libscoreboard_plugin.so
FORMS_PLUGIN_LIBRARY := libforms_plugin.so
WORLD_PLUGIN_LIBRARY := libworld_command_plugin.so
ENTITY_PLUGIN_LIBRARY := libentity_command_plugin.so
PARTICLE_PLUGIN_LIBRARY := libparticle_command_plugin.so
SOUND_PLUGIN_LIBRARY := libsound_command_plugin.so
endif

generate:
	go run ./cmd/abi-gen -root .
	cargo fmt --all
	cargo fmt --manifest-path $(MOVEMENT_MANIFEST)
	cargo fmt --manifest-path $(CHAT_MANIFEST)
	cargo fmt --manifest-path $(LIFECYCLE_MANIFEST)
	cargo fmt --manifest-path $(COMMAND_MANIFEST)
	cargo fmt --manifest-path $(PLAYER_COMMAND_MANIFEST)
	cargo fmt --manifest-path $(ITEMS_MANIFEST)
	cargo fmt --manifest-path $(PING_MANIFEST)
	cargo fmt --manifest-path $(SCOREBOARD_MANIFEST)
	cargo fmt --manifest-path $(FORMS_MANIFEST)
	cargo fmt --manifest-path $(WORLD_MANIFEST)
	cargo fmt --manifest-path $(ENTITY_MANIFEST)
	cargo fmt --manifest-path $(PARTICLE_MANIFEST)
	cargo fmt --manifest-path $(SOUND_MANIFEST)

check-generated:
	go run ./cmd/abi-gen -root . -check
	cargo fmt --all -- --check
	cargo fmt --manifest-path $(MOVEMENT_MANIFEST) -- --check
	cargo fmt --manifest-path $(CHAT_MANIFEST) -- --check
	cargo fmt --manifest-path $(LIFECYCLE_MANIFEST) -- --check
	cargo fmt --manifest-path $(COMMAND_MANIFEST) -- --check
	cargo fmt --manifest-path $(PLAYER_COMMAND_MANIFEST) -- --check
	cargo fmt --manifest-path $(ITEMS_MANIFEST) -- --check
	cargo fmt --manifest-path $(PING_MANIFEST) -- --check
	cargo fmt --manifest-path $(SCOREBOARD_MANIFEST) -- --check
	cargo fmt --manifest-path $(FORMS_MANIFEST) -- --check
	cargo fmt --manifest-path $(WORLD_MANIFEST) -- --check
	cargo fmt --manifest-path $(ENTITY_MANIFEST) -- --check
	cargo fmt --manifest-path $(PARTICLE_MANIFEST) -- --check
	cargo fmt --manifest-path $(SOUND_MANIFEST) -- --check

build-native: generate
	cargo build --release -p dragonfly-plugin-runtime
	cargo build --release --manifest-path $(MOVEMENT_MANIFEST)
	cargo build --release --manifest-path $(CHAT_MANIFEST)
	cargo build --release --manifest-path $(LIFECYCLE_MANIFEST)
	cargo build --release --manifest-path $(COMMAND_MANIFEST)
	cargo build --release --manifest-path $(PLAYER_COMMAND_MANIFEST)
	cargo build --release --manifest-path $(ITEMS_MANIFEST)
	cargo build --release --manifest-path $(PING_MANIFEST)
	cargo build --release --manifest-path $(SCOREBOARD_MANIFEST)
	cargo build --release --manifest-path $(FORMS_MANIFEST)
	cargo build --release --manifest-path $(WORLD_MANIFEST)
	cargo build --release --manifest-path $(ENTITY_MANIFEST)
	cargo build --release --manifest-path $(PARTICLE_MANIFEST)
	cargo build --release --manifest-path $(SOUND_MANIFEST)
	mkdir -p build/lib build/plugins
	cp target/release/$(RUNTIME_LIBRARY) build/lib/
	cp examples/plugins/movement-guard/target/release/$(PLUGIN_LIBRARY) build/plugins/
	cp examples/plugins/chat-filter/target/release/$(CHAT_PLUGIN_LIBRARY) build/plugins/
	cp examples/plugins/lifecycle-logger/target/release/$(LIFECYCLE_PLUGIN_LIBRARY) build/plugins/
	cp examples/plugins/hello-command/target/release/$(COMMAND_PLUGIN_LIBRARY) build/plugins/
	cp examples/plugins/player-command/target/release/$(PLAYER_COMMAND_PLUGIN_LIBRARY) build/plugins/
	cp examples/plugins/items-command/target/release/$(ITEMS_PLUGIN_LIBRARY) build/plugins/
	cp examples/plugins/ping-command/target/release/$(PING_PLUGIN_LIBRARY) build/plugins/
	cp examples/plugins/scoreboard/target/release/$(SCOREBOARD_PLUGIN_LIBRARY) build/plugins/
	cp examples/plugins/forms/target/release/$(FORMS_PLUGIN_LIBRARY) build/plugins/
	cp examples/plugins/world-command/target/release/$(WORLD_PLUGIN_LIBRARY) build/plugins/
	cp examples/plugins/entity-command/target/release/$(ENTITY_PLUGIN_LIBRARY) build/plugins/
	cp examples/plugins/particle-command/target/release/$(PARTICLE_PLUGIN_LIBRARY) build/plugins/
	cp examples/plugins/sound-command/target/release/$(SOUND_PLUGIN_LIBRARY) build/plugins/

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
	cp build/plugins/$(LIFECYCLE_PLUGIN_LIBRARY) examples/plugins/
	cp build/plugins/$(COMMAND_PLUGIN_LIBRARY) examples/plugins/
	cp build/plugins/$(PLAYER_COMMAND_PLUGIN_LIBRARY) examples/plugins/
	cp build/plugins/$(ITEMS_PLUGIN_LIBRARY) examples/plugins/
	cp build/plugins/$(PING_PLUGIN_LIBRARY) examples/plugins/
	cp build/plugins/$(SCOREBOARD_PLUGIN_LIBRARY) examples/plugins/
	cp build/plugins/$(FORMS_PLUGIN_LIBRARY) examples/plugins/
	cp build/plugins/$(WORLD_PLUGIN_LIBRARY) examples/plugins/
	cp build/plugins/$(ENTITY_PLUGIN_LIBRARY) examples/plugins/
	cp build/plugins/$(PARTICLE_PLUGIN_LIBRARY) examples/plugins/
	cp build/plugins/$(SOUND_PLUGIN_LIBRARY) examples/plugins/

run: stage-examples
	go run ./cmd/bedrock-gophers -config examples/server.toml

test: build-native check-generated
	cargo test --workspace
	cargo test --manifest-path $(MOVEMENT_MANIFEST)
	cargo test --manifest-path $(CHAT_MANIFEST)
	cargo test --manifest-path $(LIFECYCLE_MANIFEST)
	cargo test --manifest-path $(COMMAND_MANIFEST)
	cargo test --manifest-path $(PLAYER_COMMAND_MANIFEST)
	cargo test --manifest-path $(ITEMS_MANIFEST)
	cargo test --manifest-path $(PING_MANIFEST)
	cargo test --manifest-path $(SCOREBOARD_MANIFEST)
	cargo test --manifest-path $(FORMS_MANIFEST)
	cargo test --manifest-path $(WORLD_MANIFEST)
	cargo test --manifest-path $(ENTITY_MANIFEST)
	cargo test --manifest-path $(PARTICLE_MANIFEST)
	cargo test --manifest-path $(SOUND_MANIFEST)
	go test ./...

benchmark: build-native
	go test ./internal/native -run '^$$' -bench . -benchmem

clean:
	cargo clean
	cargo clean --manifest-path $(MOVEMENT_MANIFEST)
	cargo clean --manifest-path $(CHAT_MANIFEST)
	cargo clean --manifest-path $(LIFECYCLE_MANIFEST)
	cargo clean --manifest-path $(COMMAND_MANIFEST)
	cargo clean --manifest-path $(PLAYER_COMMAND_MANIFEST)
	cargo clean --manifest-path $(ITEMS_MANIFEST)
	cargo clean --manifest-path $(PING_MANIFEST)
	cargo clean --manifest-path $(SCOREBOARD_MANIFEST)
	cargo clean --manifest-path $(FORMS_MANIFEST)
	cargo clean --manifest-path $(WORLD_MANIFEST)
	cargo clean --manifest-path $(ENTITY_MANIFEST)
	cargo clean --manifest-path $(PARTICLE_MANIFEST)
	cargo clean --manifest-path $(SOUND_MANIFEST)
	rm -rf build
	rm -rf examples/lib
	rm -f examples/plugins/$(PLUGIN_LIBRARY) examples/plugins/$(CHAT_PLUGIN_LIBRARY) examples/plugins/$(LIFECYCLE_PLUGIN_LIBRARY) examples/plugins/$(COMMAND_PLUGIN_LIBRARY) examples/plugins/$(PLAYER_COMMAND_PLUGIN_LIBRARY) examples/plugins/$(ITEMS_PLUGIN_LIBRARY) examples/plugins/$(PING_PLUGIN_LIBRARY) examples/plugins/$(SCOREBOARD_PLUGIN_LIBRARY) examples/plugins/$(FORMS_PLUGIN_LIBRARY) examples/plugins/$(WORLD_PLUGIN_LIBRARY) examples/plugins/$(ENTITY_PLUGIN_LIBRARY) examples/plugins/$(PARTICLE_PLUGIN_LIBRARY) examples/plugins/$(SOUND_PLUGIN_LIBRARY)
