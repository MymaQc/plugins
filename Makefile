.PHONY: generate check-generated build-native build-server build stage-examples run test clean

RUNTIME_PROJECT := csharp/Dragonfly.Runtime/Dragonfly.Runtime.csproj
EXAMPLE_PROJECT := examples/plugins/lifecycle-logger/LifecycleLogger.csproj
MOVEMENT_PROJECT := examples/plugins/movement-guard/MovementGuard.csproj
CHAT_PROJECT := examples/plugins/chat-filter/ChatFilter.csproj
KITCHEN_SINK_PROJECT := examples/plugins/kitchen-sink/KitchenSink.csproj
DOTNET_RID ?= linux-x64

generate:
	go run ./cmd/csharp-gen -root .

check-generated:
	go run ./cmd/csharp-gen -root . -check

build-native: generate
	dotnet publish $(RUNTIME_PROJECT) -c Release -r $(DOTNET_RID) --self-contained true -o build/dotnet/runtime
	dotnet publish $(EXAMPLE_PROJECT) -c Release -r $(DOTNET_RID) --self-contained true -o build/dotnet/lifecycle-logger
	dotnet publish $(MOVEMENT_PROJECT) -c Release -r $(DOTNET_RID) --self-contained true -o build/dotnet/movement-guard
	dotnet publish $(CHAT_PROJECT) -c Release -r $(DOTNET_RID) --self-contained true -o build/dotnet/chat-filter
	dotnet publish $(KITCHEN_SINK_PROJECT) -c Release -r $(DOTNET_RID) --self-contained true -o build/dotnet/kitchen-sink
	mkdir -p build/lib build/plugins
	rm -f build/lib/*.so build/plugins/*.so
	cp build/dotnet/runtime/Dragonfly.Runtime.so build/lib/libdragonfly_plugin_runtime.so
	cp build/dotnet/lifecycle-logger/LifecycleLogger.so build/plugins/
	cp build/dotnet/movement-guard/MovementGuard.so build/plugins/
	cp build/dotnet/chat-filter/ChatFilter.so build/plugins/
	cp build/dotnet/kitchen-sink/KitchenSink.so build/plugins/

build-server:
	mkdir -p build
	go build -o build/bedrock-gophers ./cmd/bedrock-gophers

build: build-native build-server

stage-examples: build-native
	mkdir -p examples/lib
	rm -f examples/lib/*.so examples/plugins/*.so
	cp build/lib/libdragonfly_plugin_runtime.so examples/lib/
	cp build/plugins/*.so examples/plugins/

run: stage-examples
	go run ./cmd/bedrock-gophers -config examples/server.toml

test: build-native check-generated
	dotnet build csharp/Dragonfly.Generator/Dragonfly.Generator.csproj -c Release
	go test ./...

clean:
	dotnet clean $(RUNTIME_PROJECT)
	dotnet clean $(EXAMPLE_PROJECT)
	dotnet clean $(MOVEMENT_PROJECT)
	dotnet clean $(CHAT_PROJECT)
	dotnet clean $(KITCHEN_SINK_PROJECT)
	rm -rf build examples/lib
	rm -f examples/plugins/*.so
