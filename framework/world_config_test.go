package framework

import (
	"context"
	"testing"
	"time"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

func TestWorldConfigCreatesOwnedInMemoryWorld(t *testing.T) {
	manager := newWorldManager("", nil, host.NewPlayers())
	core := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = core.Close() })
	if err := manager.RegisterCore(OverworldID, core); err != nil {
		t.Fatal(err)
	}
	id, ok := manager.CreateWorld(native.WorldConfig{
		Dimension:       native.WorldDimensionNether,
		Provider:        native.WorldProviderNop,
		SaveInterval:    -1,
		RandomTickSpeed: -1,
	})
	if !ok || id == 0 {
		t.Fatal("CreateWorld rejected an in-memory config")
	}
	created, ok := manager.WorldByHandle(id)
	if !ok || created.Dimension() != world.Nether || created.Name() != "World" {
		t.Fatalf("created world = %v, %v", created, ok)
	}
	if !manager.UnloadWorld(0, id) {
		t.Fatal("in-memory world was not owned by the manager")
	}
	if _, ok := manager.WorldByHandle(id); ok {
		t.Fatal("closed in-memory world remains registered")
	}
}

func TestWorldConfigCreatesCustomDimension(t *testing.T) {
	manager := newWorldManager("", nil, host.NewPlayers())
	core := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = core.Close() })
	if err := manager.RegisterCore(OverworldID, core); err != nil {
		t.Fatal(err)
	}
	custom := &native.CustomWorldDimension{
		Range: native.BlockRange{Min: -32, Max: 191}, WaterEvaporates: true,
		LavaSpreadDuration: 750 * time.Millisecond, TimeCycle: true,
	}
	id, ok := manager.CreateWorld(native.WorldConfig{CustomDimension: custom})
	if !ok {
		t.Fatal("CreateWorld rejected a custom dimension")
	}
	created, ok := manager.WorldByHandle(id)
	if !ok || created.Range() != (cube.Range{-32, 191}) || !created.Dimension().WaterEvaporates() ||
		created.Dimension().LavaSpreadDuration() != 750*time.Millisecond ||
		created.Dimension().WeatherCycle() || !created.Dimension().TimeCycle() {
		t.Fatalf("custom world = %v, %v", created, ok)
	}
	view, ok := manager.WorldDimension(0, id)
	if !ok || view.Custom == nil || *view.Custom != *custom {
		t.Fatalf("custom dimension view = %+v, %v", view, ok)
	}
}

func TestWorldConfigAnonymousNameCannotReplaceNamedWorld(t *testing.T) {
	manager := newWorldManager("", nil, host.NewPlayers())
	core := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = core.Close() })
	if err := manager.RegisterCore(OverworldID, core); err != nil {
		t.Fatal(err)
	}
	named, err := manager.Create("bedrock-gophers:world/3", world.Config{Synchronous: true})
	if err != nil {
		t.Fatal(err)
	}
	id, ok := manager.CreateWorld(native.WorldConfig{Provider: native.WorldProviderNop})
	if !ok {
		t.Fatal("CreateWorld rejected an in-memory config")
	}
	entry, ok := manager.entryByHandle(id)
	if !ok || entry.name != "bedrock-gophers:world/4" {
		t.Fatalf("anonymous entry = %+v, %v", entry, ok)
	}
	if current, ok := manager.World("bedrock-gophers:world/3"); !ok || current != named {
		t.Fatal("anonymous world replaced the named world")
	}
	if err := manager.CloseCustom(); err != nil {
		t.Fatal(err)
	}
}

func TestWorldConfigMCDBPersistsAndReopens(t *testing.T) {
	manager, err := NewPersistentWorldManager(t.TempDir(), nil, host.NewPlayers())
	if err != nil {
		t.Fatal(err)
	}
	core := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = core.Close() })
	if err := manager.RegisterCore(OverworldID, core); err != nil {
		t.Fatal(err)
	}
	config := native.WorldConfig{
		Dimension:           native.WorldDimensionOverworld,
		Provider:            native.WorldProviderMCDB,
		ProviderPath:        "arenas/nodebuff",
		SaveInterval:        -1,
		ChunkUnloadInterval: 2 * time.Minute,
		RandomTickSpeed:     -1,
	}
	firstID, ok := manager.CreateWorld(config)
	if !ok {
		t.Fatal("CreateWorld rejected an MCDB config")
	}
	if _, duplicate := manager.CreateWorld(config); duplicate {
		t.Fatal("CreateWorld accepted a live duplicate MCDB provider")
	}
	first, _ := manager.WorldByHandle(firstID)
	position := cube.Pos{3, 64, 5}
	if err := first.Do(func(tx *world.Tx) {
		tx.SetBlock(position, block.Diamond{}, nil)
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	first.Save()
	if !manager.UnloadWorld(0, firstID) {
		t.Fatal("close first MCDB world")
	}

	secondID, ok := manager.CreateWorld(config)
	if !ok {
		t.Fatal("reopen MCDB world")
	}
	second, _ := manager.WorldByHandle(secondID)
	if err := second.Do(func(tx *world.Tx) {
		if _, ok := tx.Block(position).(block.Diamond); !ok {
			t.Fatalf("persisted block = %T, want block.Diamond", tx.Block(position))
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !manager.UnloadWorld(0, secondID) {
		t.Fatal("close reopened MCDB world")
	}
}
