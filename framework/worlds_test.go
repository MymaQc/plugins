package framework

import (
	"slices"
	"testing"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/df-mc/dragonfly/server/world"
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
