package framework

import (
	"math"
	"time"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/world"
)

type descriptorDifficulty struct {
	foodRegenerates       bool
	starvationHealthLimit float64
	fireSpreadIncrease    int
}

func (d descriptorDifficulty) FoodRegenerates() bool          { return d.foodRegenerates }
func (d descriptorDifficulty) StarvationHealthLimit() float64 { return d.starvationHealthLimit }
func (d descriptorDifficulty) FireSpreadIncrease() int        { return d.fireSpreadIncrease }

func difficultyFromView(value native.DifficultyView) (world.Difficulty, bool) {
	if value.Builtin {
		if value.ID > 3 {
			return nil, false
		}
		return world.DifficultyByID(int(value.ID))
	}
	if value.ID != 0 {
		return nil, false
	}
	return descriptorDifficulty{
		foodRegenerates:       value.FoodRegenerates,
		starvationHealthLimit: value.StarvationHealthLimit,
		fireSpreadIncrease:    int(value.FireSpreadIncrease),
	}, true
}

func difficultyView(value world.Difficulty) (view native.DifficultyView, ok bool) {
	if value == nil {
		return native.DifficultyView{}, false
	}
	defer func() {
		if recover() != nil {
			view, ok = native.DifficultyView{}, false
		}
	}()
	fireSpreadIncrease := value.FireSpreadIncrease()
	if fireSpreadIncrease < math.MinInt32 || fireSpreadIncrease > math.MaxInt32 {
		return native.DifficultyView{}, false
	}
	view = native.DifficultyView{
		FoodRegenerates:       value.FoodRegenerates(),
		StarvationHealthLimit: value.StarvationHealthLimit(),
		FireSpreadIncrease:    int32(fireSpreadIncrease),
	}
	if id, registered := registeredDifficultyID(value); registered {
		view.ID, view.Builtin = uint32(id), true
	}
	return view, true
}

func registeredDifficultyID(value world.Difficulty) (id int, ok bool) {
	defer func() {
		if recover() != nil {
			id, ok = 0, false
		}
	}()
	return world.DifficultyID(value)
}

func (m *WorldManager) WorldDimension(invocation native.InvocationID, id native.WorldID) (native.WorldDimensionView, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return native.WorldDimensionView{}, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return native.WorldDimensionView{}, false
	}
	return host.WorldDimensionView(entry.world.Dimension())
}

func (m *WorldManager) WorldTimeCycle(invocation native.InvocationID, id native.WorldID) (bool, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return false, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return false, false
	}
	return entry.world.TimeCycle(), true
}

func (m *WorldManager) SetWorldTimeCycle(invocation native.InvocationID, id native.WorldID, enabled bool) bool {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return false
	}
	if enabled {
		entry.world.StartTime()
	} else {
		entry.world.StopTime()
	}
	return true
}

func (m *WorldManager) SetWorldRequiredSleepDuration(invocation native.InvocationID, id native.WorldID, duration time.Duration) bool {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return false
	}
	entry.world.SetRequiredSleepDuration(duration)
	return true
}

func (m *WorldManager) WorldDefaultGameMode(invocation native.InvocationID, id native.WorldID) (int64, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return 0, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return 0, false
	}
	return host.EncodeGameModeDescriptor(entry.world.DefaultGameMode())
}

func (m *WorldManager) SetWorldDefaultGameMode(invocation native.InvocationID, id native.WorldID, descriptor int64) bool {
	mode, ok := host.DecodeGameModeDescriptor(descriptor)
	if !ok {
		return false
	}
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return false
	}
	entry.world.SetDefaultGameMode(mode)
	return true
}

func (m *WorldManager) SetWorldTickRange(invocation native.InvocationID, id native.WorldID, value int32) bool {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return false
	}
	entry.world.SetTickRange(int(value))
	return true
}

func (m *WorldManager) WorldDifficulty(invocation native.InvocationID, id native.WorldID) (native.DifficultyView, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return native.DifficultyView{}, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return native.DifficultyView{}, false
	}
	return difficultyView(entry.world.Difficulty())
}

func (m *WorldManager) SetWorldDifficulty(invocation native.InvocationID, id native.WorldID, value native.DifficultyView) bool {
	difficulty, ok := difficultyFromView(value)
	if !ok {
		return false
	}
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return false
	}
	entry.world.SetDifficulty(difficulty)
	return true
}
