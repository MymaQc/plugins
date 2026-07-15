package framework

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
)

func TestCSharpKitchenWorldCommandCanRunInsideCreatedWorld(t *testing.T) {
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
	worlds, err := NewPersistentWorldManager(t.TempDir(), nil, players)
	if err != nil {
		t.Fatal(err)
	}
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
		worlds.StopScheduling()
		worlds.DrainScheduled()
		worlds.DrainDetachedEntities()
		_ = worlds.CloseCustom()
		pluginRuntime.BeginDisable()
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
	openedWorld := world.Config{Entities: registry}.New()
	t.Cleanup(func() { _ = openedWorld.Close() })
	if err := worlds.RegisterCore(OverworldID, openedWorld); err != nil {
		t.Fatal(err)
	}
	if err := pluginRuntime.Enable(); err != nil {
		t.Fatal(err)
	}

	command, overload := kitchenWorldCommand(t, pluginRuntime)
	playerUUID := uuid.MustParse("ca2ce6b2-f3c0-4e16-8f0b-f5ef9a8fb182")
	handle := world.EntitySpawnOpts{ID: playerUUID, Position: mgl64.Vec3{0, 64, 0}}.New(
		player.Type,
		player.Config{UUID: playerUUID, Name: "Kitchen", Position: mgl64.Vec3{0, 64, 0}},
	)
	var playerID native.PlayerID
	if err := openedWorld.Do(func(tx *world.Tx) {
		connected := tx.AddEntity(handle).(*player.Player)
		connected.Handle(host.NewPlayerHandler(pluginRuntime, nil, players, worlds))
		playerID = players.Register(connected, 1)
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	run := func() <-chan native.CommandOutput {
		result := make(chan native.CommandOutput, 1)
		go func() {
			_ = handle.Do(func(tx *world.Tx, _ world.Entity) {
				invocation, end := players.BeginInvocation(tx)
				defer end()
				output, commandErr := pluginRuntime.HandleCommand(command.Index, native.CommandInput{
					Invocation: invocation, Overload: overload, Source: "Kitchen",
					SourceKind: native.CommandSourcePlayer, SourcePlayer: &playerID,
					SourcePosition: native.Vec3{Y: 64}, Arguments: []string{"world"},
				})
				if commandErr != nil {
					output = native.CommandOutput{Failed: true, Message: commandErr.Error()}
				}
				result <- output
			})
		}()
		return result
	}

	for attempt := 1; attempt <= 2; attempt++ {
		select {
		case output := <-run():
			if output.Failed {
				t.Fatalf("world command %d failed: %s", attempt, output.Message)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("world command %d deadlocked", attempt)
		}
	}
}

func kitchenWorldCommand(t *testing.T, pluginRuntime *native.Runtime) (native.Command, uint64) {
	t.Helper()
	commands, err := pluginRuntime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range commands {
		if command.Name != "kitchen" {
			continue
		}
		for index, overload := range command.Overloads {
			if len(overload.Parameters) == 1 && overload.Parameters[0].Name == "world" {
				return command, uint64(index)
			}
		}
	}
	t.Fatal("kitchen world command not published")
	return native.Command{}, 0
}
