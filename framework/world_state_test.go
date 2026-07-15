package framework

import (
	"math"
	"testing"
	"time"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/world"
)

type oversizedDifficulty struct{}

func (oversizedDifficulty) FoodRegenerates() bool          { return false }
func (oversizedDifficulty) StarvationHealthLimit() float64 { return 2 }
func (oversizedDifficulty) FireSpreadIncrease() int        { return math.MaxInt32 + 1 }

func TestWorldManagerExactStateMethods(t *testing.T) {
	manager := NewWorldManager()
	created, err := manager.Create("example:state", world.Config{Synchronous: true, Dim: world.Nether})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	id, ok := manager.WorldByName(0, "example:state")
	if !ok {
		t.Fatal("created world is not registered")
	}

	if dimension, ok := manager.WorldDimension(0, id); !ok || dimension.ID != native.WorldDimensionNether || dimension.Custom != nil {
		t.Fatalf("WorldDimension() = %+v, %v", dimension, ok)
	}
	if !manager.SetWorldTimeCycle(0, id, false) || created.TimeCycle() {
		t.Fatal("StopTime equivalent failed")
	}
	if cycle, ok := manager.WorldTimeCycle(0, id); !ok || cycle {
		t.Fatalf("WorldTimeCycle() = %v, %v", cycle, ok)
	}
	if !manager.SetWorldTimeCycle(0, id, true) || !created.TimeCycle() {
		t.Fatal("StartTime equivalent failed")
	}
	if !manager.SetWorldRequiredSleepDuration(0, id, 1500*time.Millisecond) {
		t.Fatal("SetWorldRequiredSleepDuration failed")
	}

	creative, ok := host.EncodeGameModeDescriptor(world.GameModeCreative)
	if !ok || !manager.SetWorldDefaultGameMode(0, id, creative) || created.DefaultGameMode() != world.GameModeCreative {
		t.Fatal("SetWorldDefaultGameMode failed")
	}
	if got, ok := manager.WorldDefaultGameMode(0, id); !ok || got != creative {
		t.Fatalf("WorldDefaultGameMode() = %d, %v; want %d", got, ok, creative)
	}
	if !manager.SetWorldTickRange(0, id, -17) {
		t.Fatal("SetWorldTickRange failed")
	}

	custom := native.DifficultyView{
		FoodRegenerates: true, StarvationHealthLimit: 7.5, FireSpreadIncrease: -4,
	}
	if !manager.SetWorldDifficulty(0, id, custom) {
		t.Fatal("SetWorldDifficulty custom failed")
	}
	current := created.Difficulty()
	if !current.FoodRegenerates() || current.StarvationHealthLimit() != 7.5 || current.FireSpreadIncrease() != -4 {
		t.Fatalf("custom difficulty = %#v", current)
	}
	if got, ok := manager.WorldDifficulty(0, id); !ok || got != custom {
		t.Fatalf("WorldDifficulty() = %#v, %v", got, ok)
	}

	hard := native.DifficultyView{ID: 3, Builtin: true, StarvationHealthLimit: -1, FireSpreadIncrease: 21}
	if !manager.SetWorldDifficulty(0, id, hard) || created.Difficulty() != world.DifficultyHard {
		t.Fatal("SetWorldDifficulty builtin failed")
	}
	if got, ok := manager.WorldDifficulty(0, id); !ok || got != hard {
		t.Fatalf("builtin WorldDifficulty() = %#v, %v", got, ok)
	}
}

func TestWorldManagerRejectsInvalidStateHandlesAndDescriptors(t *testing.T) {
	manager := NewWorldManager()
	if _, ok := manager.WorldDimension(0, 99); ok {
		t.Fatal("unknown world dimension succeeded")
	}
	if manager.SetWorldDefaultGameMode(0, 99, 0) {
		t.Fatal("unknown world game mode write succeeded")
	}
	if manager.SetWorldDifficulty(0, 99, native.DifficultyView{}) {
		t.Fatal("unknown world difficulty write succeeded")
	}
	if _, ok := difficultyFromView(native.DifficultyView{Builtin: true, ID: 4}); ok {
		t.Fatal("unknown builtin difficulty succeeded")
	}
	if _, ok := difficultyFromView(native.DifficultyView{ID: 1}); ok {
		t.Fatal("custom difficulty with builtin ID succeeded")
	}
	if _, ok := difficultyView(oversizedDifficulty{}); ok {
		t.Fatal("out-of-range custom difficulty succeeded")
	}
}
