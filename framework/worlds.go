package framework

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/mcdb"
)

type WorldID string

const (
	OverworldID WorldID = "minecraft:overworld"
	NetherID    WorldID = "minecraft:nether"
	EndID       WorldID = "minecraft:end"
)

var worldIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*:[a-z0-9][a-z0-9_./-]*$`)

type managedWorld struct {
	lifecycle sync.RWMutex
	id        native.WorldID
	name      WorldID
	world     *world.World
	core      bool
	unloading bool
	closed    bool
}

// WorldManager owns every world exposed through the plugin framework.
type WorldManager struct {
	mu              sync.RWMutex
	worlds          map[WorldID]*managedWorld
	handles         map[native.WorldID]*managedWorld
	byWorld         map[*world.World]*managedWorld
	next            native.WorldID
	root            string
	log             *slog.Logger
	players         *host.Players
	entityHandles   *host.Entities
	blocks          world.BlockRegistry
	entityTypes     world.EntityRegistry
	registriesReady bool
}

// NewWorldManager constructs an in-memory manager, primarily useful to embedders and tests.
func NewWorldManager() *WorldManager {
	return newWorldManager("", nil, nil)
}

// NewPersistentWorldManager constructs a manager that opens custom worlds below root.
func NewPersistentWorldManager(root string, log *slog.Logger, players *host.Players) (*WorldManager, error) {
	if root == "" {
		return nil, errors.New("worlds.directory is required")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve world root: %w", err)
	}
	if err := os.MkdirAll(absolute, 0o755); err != nil {
		return nil, fmt.Errorf("create world root: %w", err)
	}
	return newWorldManager(absolute, log, players), nil
}

func newWorldManager(root string, log *slog.Logger, players *host.Players) *WorldManager {
	if log == nil {
		log = slog.Default()
	}
	entityHandles := host.NewEntities()
	if players != nil {
		entityHandles = players.EntityRegistry()
	}
	return &WorldManager{
		worlds: make(map[WorldID]*managedWorld), handles: make(map[native.WorldID]*managedWorld), byWorld: make(map[*world.World]*managedWorld),
		root: root, log: log, players: players, entityHandles: entityHandles,
	}
}

// RegisterCore registers a Dragonfly-owned dimension. Its handler is installed before publication.
func (m *WorldManager) RegisterCore(name WorldID, w *world.World) error {
	if name != OverworldID && name != NetherID && name != EndID {
		return fmt.Errorf("invalid core world ID %q", name)
	}
	_, err := m.register(name, w, true)
	return err
}

// Create constructs and publishes an ephemeral namespaced custom world.
func (m *WorldManager) Create(name WorldID, config world.Config) (*world.World, error) {
	if err := validateCustomWorldID(name); err != nil {
		return nil, err
	}
	m.mu.RLock()
	entities, ready := m.entityTypes, m.registriesReady
	m.mu.RUnlock()
	if ready && len(config.Entities.Types()) == 0 {
		config.Entities = entities
	}
	w := config.New()
	if _, err := m.register(name, w, false); err != nil {
		_ = w.Close()
		return nil, err
	}
	return w, nil
}

// Open opens or creates a persistent custom world using an mcdb provider.
func (m *WorldManager) Open(name WorldID, dimension native.WorldDimension) (native.WorldID, error) {
	if err := validateCustomWorldID(name); err != nil {
		return 0, err
	}
	if m.root == "" {
		return 0, errors.New("persistent world root is not configured")
	}
	dim, ok := dragonflyDimension(dimension)
	if !ok {
		return 0, fmt.Errorf("invalid world dimension %d", dimension)
	}
	if entry, ok := m.entryByName(name); ok {
		if entry.world.Dimension() != dim {
			return 0, fmt.Errorf("world %q is already open with another dimension", name)
		}
		return entry.id, nil
	}
	path, err := m.worldPath(name)
	if err != nil {
		return 0, err
	}
	m.mu.RLock()
	blocks, entities, ready := m.blocks, m.entityTypes, m.registriesReady
	m.mu.RUnlock()
	if !ready {
		return 0, errors.New("core world registries are not ready")
	}
	provider, err := (mcdb.Config{Log: m.log, Blocks: blocks}).Open(path)
	if err != nil {
		return 0, fmt.Errorf("open world %q: %w", name, err)
	}
	w := world.Config{
		Log: m.log, Dim: dim, Provider: provider, Blocks: blocks, Entities: entities,
		PortalDestination: m.portalDestination,
	}.New()
	id, err := m.register(name, w, false)
	if err != nil {
		_ = w.Close()
		if existing, ok := m.entryByName(name); ok {
			return existing.id, nil
		}
		return 0, err
	}
	return id, nil
}

func (m *WorldManager) register(name WorldID, w *world.World, core bool) (native.WorldID, error) {
	if w == nil {
		return 0, errors.New("world is nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.worlds[name]; exists {
		return 0, fmt.Errorf("world %q already exists", name)
	}
	if m.next == native.WorldID(math.MaxUint64) {
		return 0, errors.New("world handle space exhausted")
	}
	m.next++
	if core && !m.registriesReady {
		m.blocks, m.entityTypes, m.registriesReady = w.BlockRegistry(), w.EntityRegistry(), true
	}
	entry := &managedWorld{id: m.next, name: name, world: w, core: core}
	w.Handle(host.NewWorldHandler(m.entityHandles, m.players, entry.id))
	m.worlds[name], m.handles[entry.id], m.byWorld[w] = entry, entry, entry
	return entry.id, nil
}

func (m *WorldManager) portalDestination(dimension world.Dimension) *world.World {
	var name WorldID
	switch dimension {
	case world.Nether:
		name = NetherID
	case world.End:
		name = EndID
	default:
		name = OverworldID
	}
	destination, _ := m.World(name)
	return destination
}

func (m *WorldManager) entryByName(name WorldID) (*managedWorld, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.worlds[name]
	return liveWorld(entry, ok)
}

func (m *WorldManager) entryByHandle(id native.WorldID) (*managedWorld, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.handles[id]
	return liveWorld(entry, ok)
}

func (m *WorldManager) handleByWorld(w *world.World) (native.WorldID, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := liveWorld(m.byWorld[w], m.byWorld[w] != nil)
	if !ok {
		return 0, false
	}
	return entry.id, true
}

// WorldHandle resolves a managed Dragonfly world to its stable native handle.
func (m *WorldManager) WorldHandle(w *world.World) (native.WorldID, bool) {
	if w == nil {
		return 0, false
	}
	return m.handleByWorld(w)
}

// WorldByHandle resolves a stable native handle to a live managed Dragonfly world.
func (m *WorldManager) WorldByHandle(id native.WorldID) (*world.World, bool) {
	entry, ok := m.entryByHandle(id)
	if !ok {
		return nil, false
	}
	return entry.world, true
}

func liveWorld(entry *managedWorld, ok bool) (*managedWorld, bool) {
	if !ok || entry.unloading {
		return nil, false
	}
	return entry, true
}

// World returns a managed world unless it is being unloaded.
func (m *WorldManager) World(name WorldID) (*world.World, bool) {
	entry, ok := m.entryByName(name)
	if !ok {
		return nil, false
	}
	return entry.world, true
}

// IDs returns stable sorted world names.
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

func (m *WorldManager) Save(name WorldID) error {
	entry, ok := m.entryByName(name)
	if !ok {
		return fmt.Errorf("world %q not found", name)
	}
	return m.save(0, entry)
}

func (m *WorldManager) save(invocation native.InvocationID, entry *managedWorld) error {
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return fmt.Errorf("world %q is closed", entry.name)
	}
	if invocation != 0 {
		return fmt.Errorf("cannot save world %q from a world callback", entry.name)
	}
	entry.world.Save()
	return nil
}

// Unload closes a custom world. Worlds containing players must be evacuated first.
func (m *WorldManager) Unload(name WorldID) error {
	entry, ok := m.entryByName(name)
	if !ok {
		return fmt.Errorf("world %q not found", name)
	}
	return m.unload(0, entry)
}

func (m *WorldManager) unload(invocation native.InvocationID, entry *managedWorld) error {
	m.mu.Lock()
	current, ok := m.handles[entry.id]
	if !ok || current != entry || entry.unloading {
		m.mu.Unlock()
		return fmt.Errorf("world %q not found", entry.name)
	}
	if entry.core {
		m.mu.Unlock()
		return fmt.Errorf("core world %q cannot be unloaded", entry.name)
	}
	if invocation != 0 {
		m.mu.Unlock()
		return fmt.Errorf("cannot unload world %q from a world callback", entry.name)
	}
	entry.unloading = true
	m.mu.Unlock()
	entry.lifecycle.Lock()
	defer entry.lifecycle.Unlock()
	if entry.closed {
		return fmt.Errorf("world %q is closed", entry.name)
	}
	players, err := world.Call(context.Background(), entry.world, func(tx *world.Tx) (int, error) {
		count := 0
		for range tx.Players() {
			count++
		}
		return count, nil
	})
	if err != nil || players != 0 {
		m.mu.Lock()
		entry.unloading = false
		m.mu.Unlock()
		if err != nil {
			return fmt.Errorf("inspect world %q: %w", entry.name, err)
		}
		return fmt.Errorf("world %q contains %d player(s)", entry.name, players)
	}
	m.mu.Lock()
	delete(m.worlds, entry.name)
	delete(m.handles, entry.id)
	delete(m.byWorld, entry.world)
	entry.closed = true
	m.mu.Unlock()
	return entry.world.Close()
}

// CloseCustom closes all custom worlds. Dragonfly closes core dimensions itself.
func (m *WorldManager) CloseCustom() error {
	m.mu.Lock()
	custom := make([]*managedWorld, 0, len(m.worlds))
	for name, entry := range m.worlds {
		if !entry.core {
			entry.unloading = true
			custom = append(custom, entry)
			delete(m.worlds, name)
			delete(m.handles, entry.id)
			delete(m.byWorld, entry.world)
		}
	}
	m.mu.Unlock()

	var failures []error
	for _, entry := range custom {
		entry.lifecycle.Lock()
		if entry.closed {
			entry.lifecycle.Unlock()
			continue
		}
		entry.closed = true
		if err := entry.world.Close(); err != nil {
			failures = append(failures, err)
		}
		entry.lifecycle.Unlock()
	}
	return errors.Join(failures...)
}

// Native Host implementation.
func (m *WorldManager) WorldByName(_ native.InvocationID, name string) (native.WorldID, bool) {
	entry, ok := m.entryByName(WorldID(name))
	if !ok {
		return 0, false
	}
	return entry.id, true
}

func (m *WorldManager) WorldName(_ native.InvocationID, id native.WorldID) (string, bool) {
	entry, ok := m.entryByHandle(id)
	if !ok {
		return "", false
	}
	return string(entry.name), true
}

func (m *WorldManager) OpenWorld(_ native.InvocationID, name string, dimension native.WorldDimension) (native.WorldID, bool) {
	id, err := m.Open(WorldID(name), dimension)
	return id, err == nil
}

func (m *WorldManager) UnloadWorld(invocation native.InvocationID, id native.WorldID) bool {
	entry, ok := m.entryByHandle(id)
	return ok && m.unload(invocation, entry) == nil
}

func (m *WorldManager) WorldBlock(invocation native.InvocationID, id native.WorldID, position native.BlockPos) (native.WorldBlock, bool) {
	entry, ok := m.entryByHandle(id)
	if !ok {
		return native.WorldBlock{}, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return native.WorldBlock{}, false
	}
	return m.readTx(invocation, entry, func(tx *world.Tx) (native.WorldBlock, bool) {
		block := tx.Block(blockPosition(position))
		name, properties := block.EncodeBlock()
		encoded, ok := encodeBlockProperties(properties)
		return native.WorldBlock{Identifier: name, PropertiesNBT: encoded}, ok
	})
}

func (m *WorldManager) SetWorldBlock(invocation native.InvocationID, id native.WorldID, position native.BlockPos, value native.WorldBlock) bool {
	entry, ok := m.entryByHandle(id)
	if !ok || value.Identifier == "" {
		return false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return false
	}
	properties, ok := decodeBlockProperties(value.PropertiesNBT)
	if !ok {
		return false
	}
	block, ok := entry.world.BlockRegistry().BlockByName(value.Identifier, properties)
	if !ok {
		return false
	}
	return m.writeTx(invocation, entry, func(tx *world.Tx) { tx.SetBlock(blockPosition(position), block, nil) })
}

func (m *WorldManager) WorldTime(_ native.InvocationID, id native.WorldID) (int64, bool) {
	entry, ok := m.entryByHandle(id)
	if !ok {
		return 0, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return 0, false
	}
	return int64(entry.world.Time()), true
}

func (m *WorldManager) SetWorldTime(_ native.InvocationID, id native.WorldID, value int64) bool {
	entry, ok := m.entryByHandle(id)
	if !ok || value < math.MinInt || value > math.MaxInt {
		return false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return false
	}
	entry.world.SetTime(int(value))
	return true
}

func (m *WorldManager) WorldSpawn(_ native.InvocationID, id native.WorldID) (native.BlockPos, bool) {
	entry, ok := m.entryByHandle(id)
	if !ok {
		return native.BlockPos{}, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return native.BlockPos{}, false
	}
	spawn := entry.world.Spawn()
	return native.BlockPos{X: int32(spawn.X()), Y: int32(spawn.Y()), Z: int32(spawn.Z())}, true
}

func (m *WorldManager) SetWorldSpawn(_ native.InvocationID, id native.WorldID, value native.BlockPos) bool {
	entry, ok := m.entryByHandle(id)
	if !ok {
		return false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return false
	}
	entry.world.SetSpawn(blockPosition(value))
	return true
}

func (m *WorldManager) SaveWorld(invocation native.InvocationID, id native.WorldID) bool {
	entry, ok := m.entryByHandle(id)
	return ok && m.save(invocation, entry) == nil
}

func (m *WorldManager) readTx(invocation native.InvocationID, entry *managedWorld, function func(*world.Tx) (native.WorldBlock, bool)) (native.WorldBlock, bool) {
	if tx := m.currentTx(invocation, entry.world); tx != nil {
		return function(tx)
	}
	// A synchronous cross-world read could deadlock two world owners waiting on
	// each other. Plugin callbacks must schedule cross-world work instead.
	if invocation != 0 {
		return native.WorldBlock{}, false
	}
	value, err := world.Call(context.Background(), entry.world, func(tx *world.Tx) (native.WorldBlock, error) {
		value, ok := function(tx)
		if !ok {
			return native.WorldBlock{}, errors.New("world read failed")
		}
		return value, nil
	})
	return value, err == nil
}

func (m *WorldManager) writeTx(invocation native.InvocationID, entry *managedWorld, function func(*world.Tx)) bool {
	if tx := m.currentTx(invocation, entry.world); tx != nil {
		function(tx)
		return true
	}
	if invocation != 0 {
		if _, ok := m.invocationTx(invocation); !ok {
			return false
		}
	}
	entry.world.Do(function)
	return true
}

func (m *WorldManager) currentTx(invocation native.InvocationID, w *world.World) *world.Tx {
	if m.players == nil {
		return nil
	}
	tx, ok := m.players.InvocationTx(invocation)
	if !ok {
		return nil
	}
	current, ok := transactionWorld(tx)
	if !ok || current != w {
		return nil
	}
	return tx
}

func transactionWorld(tx *world.Tx) (value *world.World, ok bool) {
	if tx == nil {
		return nil, false
	}
	defer func() {
		if recover() != nil {
			value, ok = nil, false
		}
	}()
	return tx.World(), true
}

func (m *WorldManager) worldPath(name WorldID) (string, error) {
	parts := strings.SplitN(string(name), ":", 2)
	relative := filepath.Join(parts[0], filepath.FromSlash(parts[1]))
	path := filepath.Join(m.root, relative)
	rel, err := filepath.Rel(m.root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("world ID %q escapes configured root", name)
	}
	return path, nil
}

func dragonflyDimension(value native.WorldDimension) (world.Dimension, bool) {
	switch value {
	case native.WorldDimensionOverworld:
		return world.Overworld, true
	case native.WorldDimensionNether:
		return world.Nether, true
	case native.WorldDimensionEnd:
		return world.End, true
	default:
		return nil, false
	}
}

func blockPosition(value native.BlockPos) cube.Pos {
	return cube.Pos{int(value.X), int(value.Y), int(value.Z)}
}

func encodeBlockProperties(properties map[string]any) ([]byte, bool) {
	return host.EncodeBlockProperties(properties)
}

func decodeBlockProperties(data []byte) (map[string]any, bool) {
	return host.DecodeBlockProperties(data)
}

func validateCustomWorldID(id WorldID) error {
	if !worldIDPattern.MatchString(string(id)) {
		return fmt.Errorf("invalid world ID %q", id)
	}
	if strings.HasPrefix(string(id), "minecraft:") {
		return errors.New("namespace minecraft is reserved")
	}
	for _, segment := range strings.Split(strings.SplitN(string(id), ":", 2)[1], "/") {
		if segment == "." || segment == ".." {
			return fmt.Errorf("invalid world ID %q", id)
		}
	}
	return nil
}
