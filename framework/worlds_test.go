package framework

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
)

func TestWorldManagerRegistersCoreWorlds(t *testing.T) {
	manager := NewWorldManager()
	overworld := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = overworld.Close() })
	if err := manager.RegisterCore(OverworldID, overworld); err != nil {
		t.Fatal(err)
	}
	if _, ok := overworld.Handler().(*host.WorldHandler); !ok {
		t.Fatalf("handler type = %T", overworld.Handler())
	}
	if err := manager.Unload(OverworldID); err == nil {
		t.Fatal("core world unload succeeded")
	}
}

func TestWorldManagerCreatesAndUnloadsCustomWorld(t *testing.T) {
	manager := NewWorldManager()
	w, err := manager.Create("example:lobby", world.Config{Synchronous: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := w.Handler().(*host.WorldHandler); !ok {
		t.Fatalf("handler type = %T", w.Handler())
	}
	if got := manager.IDs(); !slices.Equal(got, []WorldID{"example:lobby"}) {
		t.Fatalf("IDs = %v", got)
	}
	if err := manager.Save("example:lobby"); err != nil {
		t.Fatal(err)
	}
	if err := manager.Unload("example:lobby"); err != nil {
		t.Fatal(err)
	}
	if _, ok := manager.World("example:lobby"); ok {
		t.Fatal("unloaded world remains registered")
	}
}

func TestWorldManagerRejectsInvalidOrDuplicateIDs(t *testing.T) {
	manager := NewWorldManager()
	if _, err := manager.Create("minecraft:custom", world.Config{Synchronous: true}); err == nil {
		t.Fatal("reserved namespace accepted")
	}
	if _, err := manager.Create("missing_namespace", world.Config{Synchronous: true}); err == nil {
		t.Fatal("invalid ID accepted")
	}
	if _, err := manager.Create("example:arena", world.Config{Synchronous: true}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	if _, err := manager.Create("example:arena", world.Config{Synchronous: true}); err == nil {
		t.Fatal("duplicate ID accepted")
	}
}

func TestWorldManagerNativeHandlesAreStableAndNeverReused(t *testing.T) {
	manager := NewWorldManager()
	first, err := manager.Create("example:first", world.Config{Synchronous: true})
	if err != nil {
		t.Fatal(err)
	}
	firstID, ok := manager.WorldByName(0, "example:first")
	if !ok {
		t.Fatal("first world missing")
	}
	if name, ok := manager.WorldName(0, firstID); !ok || name != "example:first" {
		t.Fatalf("WorldName() = %q, %v", name, ok)
	}
	if err := manager.Unload("example:first"); err != nil {
		t.Fatal(err)
	}
	_ = first
	if _, ok := manager.WorldName(0, firstID); ok {
		t.Fatal("stale handle still resolves")
	}
	if _, err := manager.Create("example:second", world.Config{Synchronous: true}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	secondID, _ := manager.WorldByName(0, "example:second")
	if secondID <= firstID {
		t.Fatalf("handle reused: first=%d second=%d", firstID, secondID)
	}
}

func TestWorldManagerBlockAndStateOperationsUseActiveTx(t *testing.T) {
	players := host.NewPlayers()
	manager := newWorldManager("", nil, players)
	w, err := manager.Create("example:blocks", world.Config{Synchronous: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	id, _ := manager.WorldByName(0, "example:blocks")
	if err := w.Do(func(tx *world.Tx) {
		invocation, leave := players.BeginInvocation(tx)
		defer leave()
		properties, ok := encodeBlockProperties(map[string]any{
			"bool": true, "byte": uint8(4), "int": int32(-9), "string": "north",
		})
		if !ok {
			t.Fatal("encode block properties failed")
		}
		decoded, ok := decodeBlockProperties(properties)
		if !ok || decoded["bool"] != true || decoded["byte"] != uint8(4) || decoded["int"] != int32(-9) || decoded["string"] != "north" {
			t.Fatalf("decoded properties = %#v", decoded)
		}
		gold, ok := w.BlockRegistry().BlockByName("minecraft:gold_block", nil)
		if !ok {
			t.Fatal("gold block is not registered")
		}
		name, state := gold.EncodeBlock()
		stateNBT, ok := encodeBlockProperties(state)
		if !ok || !manager.SetWorldBlock(invocation, id, native.BlockPos{X: 2, Y: 3, Z: 4}, native.WorldBlock{Identifier: name, PropertiesNBT: stateNBT}) {
			t.Fatal("SetWorldBlock() failed")
		}
		got, ok := manager.WorldBlock(invocation, id, native.BlockPos{X: 2, Y: 3, Z: 4})
		if !ok || got.Identifier != name {
			t.Fatalf("WorldBlock() = %#v, %v", got, ok)
		}
		if manager.SaveWorld(invocation, id) {
			t.Fatal("SaveWorld succeeded from owner transaction")
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !manager.SetWorldTime(0, id, 6000) {
		t.Fatal("SetWorldTime failed")
	}
	if got, ok := manager.WorldTime(0, id); !ok || got != 6000 {
		t.Fatalf("WorldTime() = %d, %v", got, ok)
	}
	spawn := native.BlockPos{X: 8, Y: 70, Z: -3}
	if !manager.SetWorldSpawn(0, id, spawn) {
		t.Fatal("SetWorldSpawn failed")
	}
	if got, ok := manager.WorldSpawn(0, id); !ok || got != spawn {
		t.Fatalf("WorldSpawn() = %#v, %v", got, ok)
	}
}

func TestWorldManagerInvocationCannotBorrowAnotherWorldTransaction(t *testing.T) {
	players := host.NewPlayers()
	manager := newWorldManager("", nil, players)
	first, err := manager.Create("example:first", world.Config{Synchronous: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create("example:second", world.Config{Synchronous: true}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	firstID, _ := manager.WorldByName(0, "example:first")
	secondID, _ := manager.WorldByName(0, "example:second")
	position := native.BlockPos{X: 1, Y: 2, Z: 3}
	var stale native.InvocationID
	if err := first.Do(func(tx *world.Tx) {
		invocation, leave := players.BeginInvocation(tx)
		stale = invocation
		if _, ok := manager.WorldBlock(invocation, secondID, position); ok {
			t.Fatal("cross-world synchronous read succeeded")
		}
		if _, ok := manager.WorldBlock(invocation, firstID, position); !ok {
			t.Fatal("same-world read did not use invocation transaction")
		}
		leave()
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := manager.WorldBlock(stale, firstID, position); ok {
		t.Fatal("stale invocation still resolves")
	}
}

func TestWorldManagerUnloadRejectsPlayersAndCurrentOwner(t *testing.T) {
	players := host.NewPlayers()
	manager := newWorldManager("", nil, players)
	w, err := manager.Create("example:occupied", world.Config{Synchronous: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	id, _ := manager.WorldByName(0, "example:occupied")
	if err := w.Do(func(tx *world.Tx) {
		invocation, leave := players.BeginInvocation(tx)
		defer leave()
		if manager.UnloadWorld(invocation, id) {
			t.Fatal("owner unload succeeded")
		}
		playerID := uuid.MustParse("4f62ee78-9519-4f1c-b0bd-69f57b578daf")
		handle := world.EntitySpawnOpts{ID: playerID, Position: mgl64.Vec3{}}.New(
			player.Type, player.Config{UUID: playerID, Name: "Occupant"},
		)
		tx.AddEntity(handle)
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Unload("example:occupied"); err == nil || !strings.Contains(err.Error(), "contains 1 player") {
		t.Fatalf("occupied unload error = %v", err)
	}
}

func TestPersistentWorldManagerKeepsWorldsBelowRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "custom-worlds")
	manager, err := NewPersistentWorldManager(root, nil, host.NewPlayers())
	if err != nil {
		t.Fatal(err)
	}
	core := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = core.Close() })
	if err := manager.RegisterCore(OverworldID, core); err != nil {
		t.Fatal(err)
	}
	if _, ok := manager.OpenWorld(0, "example:arenas/one", native.WorldDimensionOverworld); !ok {
		t.Fatal("OpenWorld failed")
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	if _, err := manager.Open("example:../escape", native.WorldDimensionOverworld); err == nil {
		t.Fatal("unsafe world path accepted")
	}
	if _, err := os.Stat(filepath.Join(root, "example", "arenas", "one", "db")); err != nil {
		t.Fatalf("persistent DB missing: %v", err)
	}
}
