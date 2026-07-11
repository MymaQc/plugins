package framework

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"sync"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/df-mc/dragonfly/server/world"
)

type WorldID string

const (
	OverworldID WorldID = "minecraft:overworld"
	NetherID    WorldID = "minecraft:nether"
	EndID       WorldID = "minecraft:end"
)

var worldIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*:[a-z0-9][a-z0-9_./-]*$`)

type managedWorld struct {
	world     *world.World
	core      bool
	unloading bool
}

// WorldManager owns every world exposed through the plugin framework.
type WorldManager struct {
	mu     sync.RWMutex
	worlds map[WorldID]*managedWorld
}

func NewWorldManager() *WorldManager {
	return &WorldManager{worlds: make(map[WorldID]*managedWorld)}
}

// RegisterCore registers a Dragonfly-owned dimension. Its handler is installed before publication.
func (m *WorldManager) RegisterCore(id WorldID, w *world.World) error {
	if id != OverworldID && id != NetherID && id != EndID {
		return fmt.Errorf("invalid core world ID %q", id)
	}
	return m.register(id, w, true)
}

// Create constructs and publishes a namespaced custom world.
func (m *WorldManager) Create(id WorldID, config world.Config) (*world.World, error) {
	if err := validateCustomWorldID(id); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.worlds[id]; exists {
		return nil, fmt.Errorf("world %q already exists", id)
	}
	w := config.New()
	w.Handle(host.NewWorldHandler())
	m.worlds[id] = &managedWorld{world: w}
	return w, nil
}

func (m *WorldManager) register(id WorldID, w *world.World, core bool) error {
	if w == nil {
		return errors.New("world is nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.worlds[id]; exists {
		return fmt.Errorf("world %q already exists", id)
	}
	w.Handle(host.NewWorldHandler())
	m.worlds[id] = &managedWorld{world: w, core: core}
	return nil
}

// World returns a managed world unless it is being unloaded.
func (m *WorldManager) World(id WorldID) (*world.World, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.worlds[id]
	return worldValue(entry, ok)
}

func worldValue(entry *managedWorld, ok bool) (*world.World, bool) {
	if !ok || entry.unloading {
		return nil, false
	}
	return entry.world, true
}

// IDs returns stable sorted world IDs.
func (m *WorldManager) IDs() []WorldID {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]WorldID, 0, len(m.worlds))
	for id, entry := range m.worlds {
		if !entry.unloading {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (m *WorldManager) Save(id WorldID) error {
	w, ok := m.World(id)
	if !ok {
		return fmt.Errorf("world %q not found", id)
	}
	w.Save()
	return nil
}

// Unload closes a custom world. Worlds containing players must be evacuated first.
func (m *WorldManager) Unload(id WorldID) error {
	m.mu.Lock()
	entry, ok := m.worlds[id]
	if !ok || entry.unloading {
		m.mu.Unlock()
		return fmt.Errorf("world %q not found", id)
	}
	if entry.core {
		m.mu.Unlock()
		return fmt.Errorf("core world %q cannot be unloaded", id)
	}
	entry.unloading = true
	m.mu.Unlock()

	players := 0
	<-entry.world.Exec(func(tx *world.Tx) {
		for range tx.Players() {
			players++
		}
	})
	if players != 0 {
		m.mu.Lock()
		entry.unloading = false
		m.mu.Unlock()
		return fmt.Errorf("world %q contains %d player(s)", id, players)
	}

	m.mu.Lock()
	delete(m.worlds, id)
	m.mu.Unlock()
	return entry.world.Close()
}

// CloseCustom closes all custom worlds. Dragonfly closes core dimensions itself.
func (m *WorldManager) CloseCustom() error {
	m.mu.Lock()
	custom := make([]*world.World, 0, len(m.worlds))
	for id, entry := range m.worlds {
		if !entry.core {
			entry.unloading = true
			custom = append(custom, entry.world)
			delete(m.worlds, id)
		}
	}
	m.mu.Unlock()

	var failures []error
	for _, w := range custom {
		if err := w.Close(); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func validateCustomWorldID(id WorldID) error {
	if !worldIDPattern.MatchString(string(id)) {
		return fmt.Errorf("invalid world ID %q", id)
	}
	if len(id) >= len("minecraft:") && string(id[:len("minecraft:")]) == "minecraft:" {
		return fmt.Errorf("namespace minecraft is reserved")
	}
	return nil
}
