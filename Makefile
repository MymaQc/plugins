.PHONY: generate check-generated build-native build-server build stage-examples run test clean

RUNTIME_PROJECT := csharp/Dragonfly.Runtime/Dragonfly.Runtime.csproj
EXAMPLE_PROJECTS := $(sort $(wildcard examples/plugins/*/*.csproj))
DOTNET_RID ?= linux-x64

generate:
	go run ./cmd/csharp-gen -root .

check-generated:
	go run ./cmd/csharp-gen -root . -check

build-native: generate
	dotnet publish $(RUNTIME_PROJECT) -c Release -r $(DOTNET_RID) --self-contained true -o build/dotnet/runtime
	@set -eu; for project in $(EXAMPLE_PROJECTS); do \
		name=$$(basename "$$(dirname "$$project")"); \
		dotnet publish "$$project" -c Release -r $(DOTNET_RID) --self-contained true -o "build/dotnet/examples/$$name"; \
	done
	mkdir -p build/lib build/plugins
	rm -f build/lib/*.so build/plugins/*.so
	cp build/dotnet/runtime/Dragonfly.Runtime.so build/lib/libdragonfly_plugin_runtime.so
	@set -eu; for project in $(EXAMPLE_PROJECTS); do \
		name=$$(basename "$$(dirname "$$project")"); \
		find "build/dotnet/examples/$$name" -maxdepth 1 -type f -name '*.so' -exec cp {} build/plugins/ \;; \
	done

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
	@set -eu; for project in $(EXAMPLE_PROJECTS); do dotnet clean "$$project"; done
	rm -rf build examples/lib
	rm -f examples/plugins/*.so
