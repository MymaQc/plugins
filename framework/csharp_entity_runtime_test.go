package framework

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
)

func TestCSharpCustomEntityRoundTrip(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	extension := ".so"
	if runtime.GOOS == "darwin" {
		extension = ".dylib"
	} else if runtime.GOOS == "windows" {
		extension = ".dll"
	}
	library := filepath.Join(root, "build", "lib", "libdragonfly_plugin_runtime"+extension)
	plugins := filepath.Join(root, "build", "plugins")
	if _, err := os.Stat(library); err != nil {
		t.Skipf("C# runtime not built: run make build-native (%v)", err)
	}

	players := host.NewPlayers()
	serverHost := host.NewServer(players)
	worlds := newWorldManager("", nil, players)
	pluginRuntime, err := native.OpenWithHost(library, plugins, struct {
		*host.Players
		*host.Server
		*WorldManager
	}{players, serverHost, worlds})
	if err != nil {
		t.Fatal(err)
	}
	worlds.attachRuntime(pluginRuntime)
	t.Cleanup(func() {
		pluginRuntime.BeginDisable()
		worlds.DrainDetachedEntities()
		pluginRuntime.FinishDisable()
		serverHost.Close()
		pluginRuntime.Close()
	})

	definitions, err := pluginRuntime.EntityTypes()
	if err != nil {
		t.Fatal(err)
	}
	registry, err := buildEntityRegistry(entity.DefaultRegistry, definitions, foreignEntityServices{
		runtime: pluginRuntime, players: players, entities: worlds.entityHandles,
	})
	if err != nil {
		t.Fatal(err)
	}
	openedWorld := world.Config{Synchronous: true, Entities: registry}.New()
	t.Cleanup(func() { _ = openedWorld.Close() })
	if err := worlds.RegisterCore(OverworldID, openedWorld); err != nil {
		t.Fatal(err)
	}
	if err := pluginRuntime.Enable(); err != nil {
		t.Fatal(err)
	}

	commands, err := pluginRuntime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	var command native.Command
	foundCommand := false
	for _, candidate := range commands {
		if candidate.Name == "kitchen" {
			command, foundCommand = candidate, true
			break
		}
	}
	if !foundCommand {
		t.Fatal("kitchen command not published")
	}
	var overload uint64
	foundOverload := false
	for index, candidate := range command.Overloads {
		if len(candidate.Parameters) == 1 && candidate.Parameters[0].Name == "custom-entity" {
			overload, foundOverload = uint64(index), true
			break
		}
	}
	if !foundOverload {
		t.Fatalf("custom-entity overload missing: %#v", command.Overloads)
	}

	var output native.CommandOutput
	if err := openedWorld.Do(func(tx *world.Tx) {
		invocation, end := players.BeginInvocation(tx)
		defer end()
		output, err = pluginRuntime.HandleCommand(command.Index, native.CommandInput{
			Invocation: invocation,
			Source:     "Console", SourceKind: native.CommandSourceConsole,
			SourcePosition: native.Vec3{X: 1, Y: 64, Z: 2},
			Overload:       overload, Arguments: []string{"custom-entity"},
		})
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	const expected = "type=bedrock_gophers:kitchen_marker, uuid=true, detached=true, added=true, removed=true, readded=true, closed=true"
	if output.Failed || output.Message != expected {
		t.Fatalf("custom entity output = %#v, want %q", output, expected)
	}
}
