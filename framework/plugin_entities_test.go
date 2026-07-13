package framework

import (
	"context"
	"math"
	"path/filepath"
	"testing"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block/cube"
	dfentity "github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/mcdb"
	"github.com/go-gl/mathgl/mgl64"
)

func TestBuildEntityRegistryPreservesDragonflyAndAddsForeignTypes(t *testing.T) {
	definition := native.EntityTypeDefinition{
		SaveID: "example:marker", NetworkID: "minecraft:armor_stand",
		Min: native.Vec3{X: -0.25, Z: -0.25}, Max: native.Vec3{X: 0.25, Y: 1.975, Z: 0.25},
	}
	registry, err := buildEntityRegistry(dfentity.DefaultRegistry, []native.EntityTypeDefinition{definition})
	if err != nil {
		t.Fatal(err)
	}
	if registry.Config().TNT == nil || registry.Config().Arrow == nil || registry.Config().Item == nil {
		t.Fatal("Dragonfly entity factories were not preserved")
	}
	if _, ok := registry.Lookup("minecraft:tnt"); !ok {
		t.Fatal("Dragonfly entity types were not preserved")
	}
	entityType, ok := registry.Lookup("example:marker")
	if !ok {
		t.Fatal("custom entity type is missing")
	}
	foreign, ok := entityType.(*foreignBaseEntityType)
	if !ok || foreign.NetworkEncodeEntity() != "minecraft:armor_stand" {
		t.Fatalf("custom entity type = %#v", entityType)
	}
	wantBounds := cube.Box(-0.25, 0, -0.25, 0.25, 1.975, 0.25)
	if got := foreign.BBox(nil); got != wantBounds {
		t.Fatalf("BBox() = %#v, want %#v", got, wantBounds)
	}
}

func TestBuildEntityRegistryRejectsInvalidAndDuplicateTypes(t *testing.T) {
	tests := []native.EntityTypeDefinition{
		{SaveID: "minecraft:tnt", NetworkID: "minecraft:tnt", Max: native.Vec3{X: 1, Y: 1, Z: 1}},
		{SaveID: "missing-namespace", NetworkID: "minecraft:pig", Max: native.Vec3{X: 1, Y: 1, Z: 1}},
		{SaveID: "example:nan", NetworkID: "minecraft:pig", Min: native.Vec3{X: math.NaN()}, Max: native.Vec3{X: 1, Y: 1, Z: 1}},
		{SaveID: "example:inverted", NetworkID: "minecraft:pig", Min: native.Vec3{X: 2}, Max: native.Vec3{X: 1, Y: 1, Z: 1}},
	}
	for _, definition := range tests {
		if _, err := buildEntityRegistry(dfentity.DefaultRegistry, []native.EntityTypeDefinition{definition}); err == nil {
			t.Fatalf("accepted invalid definition %#v", definition)
		}
	}
	duplicate := native.EntityTypeDefinition{
		SaveID: "example:duplicate", NetworkID: "minecraft:pig", Max: native.Vec3{X: 1, Y: 1, Z: 1},
	}
	if _, err := buildEntityRegistry(dfentity.DefaultRegistry, []native.EntityTypeDefinition{duplicate, duplicate}); err == nil {
		t.Fatal("accepted duplicate custom entity definitions")
	}
}

func TestWorldManagerSpawnsForeignBaseEntity(t *testing.T) {
	players := host.NewPlayers()
	manager := newWorldManager("", nil, players)
	registry, err := buildEntityRegistry(dfentity.DefaultRegistry, []native.EntityTypeDefinition{{
		SaveID: "example:marker", NetworkID: "minecraft:armor_stand",
		Min: native.Vec3{X: -0.25, Z: -0.25}, Max: native.Vec3{X: 0.25, Y: 1.975, Z: 0.25},
	}})
	if err != nil {
		t.Fatal(err)
	}
	w, err := manager.Create("example:entities", world.Config{Synchronous: true, Entities: registry})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	worldID, _ := manager.WorldByName(0, "example:entities")
	if err := w.Do(func(tx *world.Tx) {
		invocation, end := players.BeginInvocation(tx)
		defer end()
		id, ok := manager.SpawnWorldEntity(invocation, worldID, native.EntitySpawn{
			Kind: native.EntityCustom, Type: "example:marker", Position: native.Vec3{X: 1, Y: 64, Z: 2}, NameTag: "Marker",
		})
		if !ok {
			t.Fatal("custom entity spawn failed")
		}
		state, ok := manager.EntityState(invocation, id)
		if !ok || state.Type != "example:marker" || state.NameTag != "Marker" || !state.CanTeleport || state.HasVelocity {
			t.Fatalf("state = %#v, %v", state, ok)
		}
		current, ok := manager.entityHandles.Resolve(id, tx)
		if !ok {
			t.Fatal("custom entity did not resolve")
		}
		if _, ok := current.(world.TickerEntity); ok {
			t.Fatal("base custom entity accidentally implements world.TickerEntity")
		}
		if _, ok := current.(dfentity.Living); ok {
			t.Fatal("base custom entity accidentally implements entity.Living")
		}
		if _, ok := current.(velocityEntity); ok {
			t.Fatal("base custom entity accidentally implements velocity")
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestWorldManagerPropagatesPluginEntityRegistryToCustomWorlds(t *testing.T) {
	registry, err := buildEntityRegistry(dfentity.DefaultRegistry, []native.EntityTypeDefinition{{
		SaveID: "example:marker", NetworkID: "minecraft:armor_stand",
		Min: native.Vec3{X: -0.25, Z: -0.25}, Max: native.Vec3{X: 0.25, Y: 1.975, Z: 0.25},
	}})
	if err != nil {
		t.Fatal(err)
	}
	manager := NewWorldManager()
	core := world.Config{Synchronous: true, Entities: registry}.New()
	t.Cleanup(func() { _ = core.Close() })
	if err := manager.RegisterCore(OverworldID, core); err != nil {
		t.Fatal(err)
	}
	custom, err := manager.Create("example:custom", world.Config{Synchronous: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	if _, ok := custom.EntityRegistry().Lookup("example:marker"); !ok {
		t.Fatal("custom world did not inherit plugin entity registry")
	}
}

func TestForeignBaseEntityPersistsThroughDragonflyProvider(t *testing.T) {
	registry, err := buildEntityRegistry(dfentity.DefaultRegistry, []native.EntityTypeDefinition{{
		SaveID: "example:marker", NetworkID: "minecraft:armor_stand",
		Min: native.Vec3{X: -0.25, Z: -0.25}, Max: native.Vec3{X: 0.25, Y: 1.975, Z: 0.25},
	}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "world")
	provider, err := (mcdb.Config{}).Open(path)
	if err != nil {
		t.Fatal(err)
	}
	first := world.Config{Synchronous: true, Provider: provider, Entities: registry}.New()
	if err := first.Do(func(tx *world.Tx) {
		entityType, _ := registry.Lookup("example:marker")
		handle := (world.EntitySpawnOpts{
			Position: mgl64.Vec3{1, 64, 2}, NameTag: "Persistent marker",
		}).New(entityType, foreignBaseEntityConfig{})
		tx.AddEntity(handle)
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	provider, err = (mcdb.Config{}).Open(path)
	if err != nil {
		t.Fatal(err)
	}
	second := world.Config{Synchronous: true, Provider: provider, Entities: registry}.New()
	t.Cleanup(func() { _ = second.Close() })
	if err := second.Do(func(tx *world.Tx) {
		_ = tx.Block(cube.Pos{1, 64, 2})
		for current := range tx.Entities() {
			if current.H().Type().EncodeEntity() != "example:marker" {
				continue
			}
			named, ok := current.(nameTagEntity)
			if !ok {
				t.Fatalf("persisted entity type = %T", current)
			}
			if named.NameTag() != "Persistent marker" {
				t.Fatalf("persisted entity name = %q", named.NameTag())
			}
			return
		}
		t.Fatal("persisted custom entity was not loaded")
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}
