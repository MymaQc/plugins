package framework

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/mcdb"
	"github.com/google/uuid"
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
	spec      *normalizedWorldSpec
	provider  *recordingProvider
	core      bool
	unloading bool
	closed    bool
}

type worldOpening struct {
	spec normalizedWorldSpec
	done chan struct{}
	id   native.WorldID
	err  error
}

// recordingProvider preserves Dragonfly's provider close error, which
// World.Close logs but does not return.
type recordingProvider struct {
	world.Provider
	mu  sync.Mutex
	err error
}

func (p *recordingProvider) Close() error {
	err := p.Provider.Close()
	p.mu.Lock()
	p.err = err
	p.mu.Unlock()
	return err
}

func (p *recordingProvider) Err() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

func (p *recordingProvider) Retry() error { return p.Close() }

type blockPositionIterator struct {
	mu         sync.Mutex
	invocation native.InvocationID
	next       func() (cube.Pos, bool)
	stop       func()
	stopped    bool
}

func (i *blockPositionIterator) advance() (position cube.Pos, found bool, valid bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.stopped {
		return cube.Pos{}, false, false
	}
	defer func() {
		if recover() != nil {
			position, found, valid = cube.Pos{}, false, false
		}
	}()
	position, found = i.next()
	return position, found, true
}

func (i *blockPositionIterator) close() {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.stopped {
		return
	}
	i.stopped = true
	defer func() { _ = recover() }()
	i.stop()
}

// WorldManager owns every world exposed through the plugin framework.
type WorldManager struct {
	mu               sync.RWMutex
	worlds           map[WorldID]*managedWorld
	handles          map[native.WorldID]*managedWorld
	byWorld          map[*world.World]*managedWorld
	next             native.WorldID
	root             string
	log              *slog.Logger
	players          *host.Players
	entityHandles    *host.Entities
	runtime          host.WorldRuntime
	blocks           world.BlockRegistry
	entityTypes      world.EntityRegistry
	registriesReady  bool
	openings         map[WorldID]*worldOpening
	providerPaths    map[string]WorldID
	createWG         sync.WaitGroup
	closing          bool
	closeInProgress  bool
	closeDone        chan struct{}
	closeErr         error
	blockIteratorMu  sync.Mutex
	blockIterators   map[native.BlockIteratorID]*blockPositionIterator
	nextBlockIter    native.BlockIteratorID
	entityIteratorMu sync.Mutex
	entityIterators  map[native.EntityIteratorID]*worldEntityIterator
	nextEntityIter   native.EntityIteratorID
	detachedEntityMu sync.Mutex
	detachedEntities map[native.EntityHandleID]func()
	scheduleMu       sync.Mutex
	scheduleClosing  bool
	scheduledTasks   map[worldTaskKey]*scheduledWorldTask
	scheduleWG       sync.WaitGroup
}

type worldTaskKey struct {
	plugin, callback uint64
}

type scheduledWorldTask struct {
	ready chan struct{}
	task  *world.Task
}

func (t *scheduledWorldTask) cancel() bool {
	if t == nil {
		return false
	}
	<-t.ready
	return t.task != nil && t.task.Cancel()
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
		openings: make(map[WorldID]*worldOpening), providerPaths: make(map[string]WorldID),
		blockIterators:   make(map[native.BlockIteratorID]*blockPositionIterator),
		entityIterators:  make(map[native.EntityIteratorID]*worldEntityIterator),
		detachedEntities: make(map[native.EntityHandleID]func()),
		scheduledTasks:   make(map[worldTaskKey]*scheduledWorldTask),
		root:             root, log: log, players: players, entityHandles: entityHandles,
	}
}

func (m *WorldManager) attachRuntime(runtime host.WorldRuntime) {
	m.mu.Lock()
	m.runtime = runtime
	m.mu.Unlock()
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

// Open opens or creates a persistent custom world using the default policies.
func (m *WorldManager) Open(name WorldID, dimension native.WorldDimension) (native.WorldID, error) {
	dim, ok := worldSpecDimensionFromNative(dimension)
	if !ok {
		return 0, fmt.Errorf("invalid world dimension %d", dimension)
	}
	parts := strings.SplitN(string(name), ":", 2)
	providerPath := ""
	if len(parts) == 2 {
		providerPath = parts[0] + "/" + parts[1]
	}
	return m.OpenSpec(name, WorldSpec{
		ProviderPath: providerPath, Dimension: dim, OpenMode: WorldOpenOrCreate,
		Save: WorldSaveAutomatic, SaveInterval: 10 * time.Minute,
		RandomTicks: WorldRandomTicksPerSubchunk, RandomTickRate: 3,
		Time: WorldTimePreserve, Weather: WorldWeatherPreserve,
		ChunkUnload: WorldChunkUnloadAfter, ChunkUnloadAfter: 2 * time.Minute,
	})
}

// OpenSpec synchronously opens or creates a persistent world with immutable policies.
func (m *WorldManager) OpenSpec(name WorldID, value WorldSpec) (native.WorldID, error) {
	if err := validateCustomWorldID(name); err != nil {
		return 0, err
	}
	spec, err := normalizeWorldSpec(m.root, value)
	if err != nil {
		return 0, err
	}

	m.mu.Lock()
	if m.closing {
		m.mu.Unlock()
		return 0, errors.New("world manager is closing")
	}
	if entry, exists := m.worlds[name]; exists {
		if entry.unloading || entry.spec == nil || *entry.spec != spec {
			m.mu.Unlock()
			return 0, fmt.Errorf("world %q is already open with another specification", name)
		}
		id := entry.id
		m.mu.Unlock()
		return id, nil
	}
	if opening, exists := m.openings[name]; exists {
		if opening.spec != spec {
			m.mu.Unlock()
			return 0, fmt.Errorf("world %q is already opening with another specification", name)
		}
		done := opening.done
		m.mu.Unlock()
		<-done
		return opening.id, opening.err
	}
	if owner, exists := m.providerPaths[spec.absoluteProviderPath]; exists {
		m.mu.Unlock()
		return 0, fmt.Errorf("world provider path %q is already owned by %q", spec.providerPath, owner)
	}
	if !m.registriesReady {
		m.mu.Unlock()
		return 0, errors.New("core world registries are not ready")
	}
	opening := &worldOpening{spec: spec, done: make(chan struct{})}
	m.openings[name] = opening
	m.providerPaths[spec.absoluteProviderPath] = name
	m.mu.Unlock()

	id, openErr := m.openNormalized(name, spec)
	m.mu.Lock()
	opening.id, opening.err = id, openErr
	delete(m.openings, name)
	if openErr != nil {
		delete(m.providerPaths, spec.absoluteProviderPath)
	}
	close(opening.done)
	m.mu.Unlock()
	return id, openErr
}

func (m *WorldManager) openNormalized(name WorldID, spec normalizedWorldSpec) (native.WorldID, error) {
	if err := preflightProvider(spec.OpenMode, spec.absoluteProviderPath); err != nil {
		return 0, fmt.Errorf("open world %q: %w", name, err)
	}
	m.mu.RLock()
	blocks, entities := m.blocks, m.entityTypes
	m.mu.RUnlock()
	provider, err := (mcdb.Config{Log: m.log, Blocks: blocks}).Open(spec.absoluteProviderPath)
	if err != nil {
		return 0, fmt.Errorf("open world %q: %w", name, err)
	}
	spec.applySettings(provider.Settings())
	config := spec.config(m.log, provider, blocks, entities)
	config.PortalDestination = m.portalDestination
	w := config.New()

	m.mu.Lock()
	if m.next == native.WorldID(math.MaxUint64) {
		m.mu.Unlock()
		_ = w.Close()
		return 0, errors.New("world handle space exhausted")
	}
	m.next++
	stored := spec
	entry := &managedWorld{id: m.next, name: name, world: w, spec: &stored}
	handler := host.NewWorldHandler(m.entityHandles, m.players, entry.id)
	handler.AttachRuntime(m.runtime, m.log)
	w.Handle(handler)
	m.worlds[name], m.handles[entry.id], m.byWorld[w] = entry, entry, entry
	m.mu.Unlock()
	return entry.id, nil
}

func preflightProvider(mode WorldOpenMode, path string) error {
	info, err := os.Lstat(path)
	switch mode {
	case WorldOpenOrCreate:
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("inspect provider: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("provider path is a symlink")
		}
		return nil
	case WorldOpenExisting:
		if err != nil {
			return fmt.Errorf("inspect existing provider: %w", err)
		}
		if !info.IsDir() {
			return errors.New("existing provider path is not a directory")
		}
		for _, relative := range []string{"level.dat", filepath.Join("db", "CURRENT")} {
			artifact, statErr := os.Lstat(filepath.Join(path, relative))
			if statErr != nil {
				return fmt.Errorf("inspect existing provider artifact %q: %w", relative, statErr)
			}
			if !artifact.Mode().IsRegular() {
				return fmt.Errorf("existing provider artifact %q is not a regular file", relative)
			}
		}
		return nil
	case WorldCreateNew:
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("inspect new provider: %w", err)
		}
		return errors.New("new provider path already exists")
	default:
		return fmt.Errorf("invalid world open mode %d", mode)
	}
}

func (m *WorldManager) register(name WorldID, w *world.World, core bool) (native.WorldID, error) {
	return m.registerConfigured(name, w, core, nil)
}

func (m *WorldManager) registerConfigured(name WorldID, w *world.World, core bool, spec *normalizedWorldSpec) (native.WorldID, error) {
	if w == nil {
		return 0, errors.New("world is nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closing {
		return 0, errors.New("world manager is closing")
	}
	if name != "" {
		if _, exists := m.worlds[name]; exists {
			return 0, fmt.Errorf("world %q already exists", name)
		}
		if _, exists := m.openings[name]; exists {
			return 0, fmt.Errorf("world %q is opening", name)
		}
	}
	if spec != nil {
		if _, exists := m.providerPaths[spec.absoluteProviderPath]; !exists {
			return 0, errors.New("world provider path is not reserved")
		}
	}
	if m.next == native.WorldID(math.MaxUint64) {
		return 0, errors.New("world handle space exhausted")
	}
	m.next++
	for name == "" {
		candidate := WorldID(fmt.Sprintf("bedrock-gophers:world/%d", m.next))
		_, worldExists := m.worlds[candidate]
		_, openingExists := m.openings[candidate]
		if !worldExists && !openingExists {
			name = candidate
			break
		}
		if m.next == native.WorldID(math.MaxUint64) {
			return 0, errors.New("world handle space exhausted")
		}
		m.next++
	}
	if core && !m.registriesReady {
		m.blocks, m.entityTypes, m.registriesReady = w.BlockRegistry(), w.EntityRegistry(), true
	}
	entry := &managedWorld{id: m.next, name: name, world: w, spec: spec, core: core}
	handler := host.NewWorldHandler(m.entityHandles, m.players, entry.id)
	handler.AttachRuntime(m.runtime, m.log)
	w.Handle(handler)
	m.worlds[name], m.handles[entry.id], m.byWorld[w] = entry, entry, entry
	if spec != nil {
		m.providerPaths[spec.absoluteProviderPath] = name
	}
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
	if entry.closed {
		entry.lifecycle.Unlock()
		err := m.retryProviderClose(entry)
		if err != nil {
			m.mu.Lock()
			entry.unloading = false
			m.mu.Unlock()
		}
		return err
	}
	// Marking the entry as unloading prevents new off-owner operations. Drain
	// operations that already resolved it, but do not hold this lock while
	// waiting on the world owner: an active callback may still use its Tx.
	entry.lifecycle.Unlock()
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
	return m.finishWorldClose(entry, entry.world.Close)
}

// CloseCustom closes all custom worlds. Dragonfly closes core dimensions itself.
func (m *WorldManager) CloseCustom() error {
	m.mu.Lock()
	if m.closing {
		if m.closeInProgress || m.closeErr == nil {
			done := m.closeDone
			m.mu.Unlock()
			<-done
			m.mu.RLock()
			defer m.mu.RUnlock()
			return m.closeErr
		}
		// A prior close finished with retained provider ownership. Start a
		// fresh pass so transient provider failures can be retried.
		m.closeInProgress = true
		m.closeDone = make(chan struct{})
		m.closeErr = nil
	} else {
		m.closing = true
		m.closeInProgress = true
		m.closeDone = make(chan struct{})
	}
	openings := make([]<-chan struct{}, 0, len(m.openings))
	for _, opening := range m.openings {
		openings = append(openings, opening.done)
	}
	m.mu.Unlock()
	for _, done := range openings {
		<-done
	}
	m.createWG.Wait()

	m.mu.Lock()
	custom := make([]*managedWorld, 0, len(m.worlds))
	for _, entry := range m.worlds {
		if !entry.core {
			entry.unloading = true
			custom = append(custom, entry)
		}
	}
	m.mu.Unlock()

	var failures []error
	for _, entry := range custom {
		entry.lifecycle.Lock()
		closed := entry.closed
		entry.lifecycle.Unlock()
		var err error
		if closed {
			err = m.retryProviderClose(entry)
		} else {
			err = m.finishWorldClose(entry, entry.world.Close)
		}
		if err != nil {
			m.mu.Lock()
			entry.unloading = false
			m.mu.Unlock()
			failures = append(failures, err)
		}
	}
	err := errors.Join(failures...)
	m.mu.Lock()
	m.closeErr = err
	m.closeInProgress = false
	close(m.closeDone)
	m.mu.Unlock()
	return err
}

func (m *WorldManager) finishWorldClose(entry *managedWorld, closeWorld func() error) error {
	err := closeWorld()
	if err == nil && entry.provider != nil {
		err = entry.provider.Err()
	}
	entry.lifecycle.Lock()
	entry.closed = true
	entry.lifecycle.Unlock()
	if err != nil {
		m.mu.Lock()
		entry.unloading = false
		m.mu.Unlock()
		return err
	}
	m.forgetWorld(entry)
	return nil
}

func (m *WorldManager) retryProviderClose(entry *managedWorld) error {
	if entry.provider == nil {
		m.forgetWorld(entry)
		return nil
	}
	if err := entry.provider.Retry(); err != nil {
		return err
	}
	m.forgetWorld(entry)
	return nil
}

func (m *WorldManager) forgetWorld(entry *managedWorld) {
	m.mu.Lock()
	if m.worlds[entry.name] == entry {
		delete(m.worlds, entry.name)
		delete(m.handles, entry.id)
		delete(m.byWorld, entry.world)
	}
	if entry.spec != nil {
		delete(m.providerPaths, entry.spec.absoluteProviderPath)
	}
	m.mu.Unlock()
}

// Native Host implementation.
func (m *WorldManager) ServerWorld(dimension native.WorldDimension) (native.WorldID, bool) {
	var name WorldID
	switch dimension {
	case native.WorldDimensionOverworld:
		name = OverworldID
	case native.WorldDimensionNether:
		name = NetherID
	case native.WorldDimensionEnd:
		name = EndID
	default:
		return 0, false
	}
	return m.WorldByName(0, string(name))
}

// CreateWorld constructs and owns a Dragonfly world.Config. A Nop provider is
// in-memory; an MCDB provider persists below the configured world root.
func (m *WorldManager) CreateWorld(value native.WorldConfig) (native.WorldID, bool) {
	dimension, ok := pluginWorldDimension(value)
	if !ok || value.Provider > native.WorldProviderMCDB {
		return 0, false
	}
	var spec *normalizedWorldSpec
	if value.Provider == native.WorldProviderMCDB {
		providerPath, absolutePath, err := normalizeProviderPath(m.root, value.ProviderPath)
		if err != nil || preflightProvider(WorldOpenOrCreate, absolutePath) != nil {
			return 0, false
		}
		spec = &normalizedWorldSpec{providerPath: providerPath, absoluteProviderPath: absolutePath}
	} else if value.ProviderPath != "" {
		return 0, false
	}

	m.mu.Lock()
	if m.closing || !m.registriesReady {
		m.mu.Unlock()
		return 0, false
	}
	if spec != nil {
		if _, exists := m.providerPaths[spec.absoluteProviderPath]; exists {
			m.mu.Unlock()
			return 0, false
		}
		m.providerPaths[spec.absoluteProviderPath] = "bedrock-gophers:opening"
	}
	blocks, entities := m.blocks, m.entityTypes
	m.createWG.Add(1)
	m.mu.Unlock()
	defer m.createWG.Done()

	reserved := spec != nil
	defer func() {
		if !reserved {
			return
		}
		m.mu.Lock()
		if m.providerPaths[spec.absoluteProviderPath] == "bedrock-gophers:opening" {
			delete(m.providerPaths, spec.absoluteProviderPath)
		}
		m.mu.Unlock()
	}()

	config := world.Config{
		Log: m.log, Dim: dimension, ReadOnly: value.ReadOnly, PortalDestination: m.portalDestination,
		SaveInterval: value.SaveInterval, ChunkUnloadInterval: value.ChunkUnloadInterval,
		RandomTickSpeed: value.RandomTickSpeed, Blocks: blocks, Entities: entities,
	}
	if spec != nil {
		provider, err := (mcdb.Config{Log: m.log, Blocks: blocks}).Open(spec.absoluteProviderPath)
		if err != nil {
			m.log.Error("open plugin world provider", "error", err)
			return 0, false
		}
		recorder := &recordingProvider{Provider: provider}
		config.Provider = recorder
	}
	w, err := constructWorld(config)
	if err != nil {
		if config.Provider != nil {
			_ = config.Provider.Close()
		}
		m.log.Error("construct plugin world", "error", err)
		return 0, false
	}
	id, err := m.registerConfigured("", w, false, spec)
	if err != nil {
		_ = w.Close()
		return 0, false
	}
	if recorder, ok := config.Provider.(*recordingProvider); ok {
		m.mu.Lock()
		if entry := m.handles[id]; entry != nil {
			entry.provider = recorder
		}
		m.mu.Unlock()
	}
	reserved = false
	return id, true
}

type customDimension struct{ value native.CustomWorldDimension }

func (d customDimension) Range() cube.Range {
	return cube.Range{int(d.value.Range.Min), int(d.value.Range.Max)}
}
func (d customDimension) WaterEvaporates() bool { return d.value.WaterEvaporates }
func (d customDimension) LavaSpreadDuration() time.Duration {
	return d.value.LavaSpreadDuration
}
func (d customDimension) WeatherCycle() bool { return d.value.WeatherCycle }
func (d customDimension) TimeCycle() bool    { return d.value.TimeCycle }

func pluginWorldDimension(value native.WorldConfig) (world.Dimension, bool) {
	if value.CustomDimension != nil {
		return customDimension{value: *value.CustomDimension}, true
	}
	return worldSpecDimension(WorldDimension(value.Dimension))
}

func constructWorld(config world.Config) (created *world.World, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			created = nil
			err = fmt.Errorf("world.Config.New panicked: %v", recovered)
		}
	}()
	return config.New(), nil
}

// ScheduleWorld mirrors World.Do and World.DoAfter with a borrowed transaction.
func (m *WorldManager) ScheduleWorld(id native.WorldID, plugin, callback uint64, delayNanoseconds int64) bool {
	if id == 0 || plugin == 0 || callback == 0 || m.players == nil {
		return false
	}
	entry, ok := m.entryByHandle(id)
	if !ok {
		return false
	}
	m.mu.RLock()
	runtime := m.runtime
	m.mu.RUnlock()
	if runtime == nil {
		return false
	}
	key := worldTaskKey{plugin: plugin, callback: callback}
	pending := &scheduledWorldTask{ready: make(chan struct{})}
	m.scheduleMu.Lock()
	if m.scheduleClosing || m.scheduledTasks[key] != nil {
		m.scheduleMu.Unlock()
		return false
	}
	m.scheduledTasks[key] = pending
	m.scheduleWG.Add(1)
	m.scheduleMu.Unlock()
	callbackFunction := func(tx *world.Tx) {
		invocation, end := m.players.BeginInvocation(tx)
		defer end()
		if invocation == 0 {
			panic("scheduled world callback has no invocation")
		}
		if err := runtime.HandleWorldScheduled(plugin, callback, invocation, native.WorldTaskExecute, native.WorldTaskSuccess); err != nil {
			panic(err)
		}
	}
	var task *world.Task
	if delayNanoseconds <= 0 {
		task = entry.world.Do(callbackFunction)
	} else {
		task = entry.world.DoAfter(time.Duration(delayNanoseconds), callbackFunction)
	}
	pending.task = task
	close(pending.ready)
	task.OnDone(func(err error) {
		result := native.WorldTaskSuccess
		switch {
		case errors.Is(err, world.ErrTaskCancelled):
			result = native.WorldTaskCancelled
		case errors.Is(err, world.ErrWorldClosed):
			result = native.WorldTaskWorldClosed
		case errors.Is(err, world.ErrTaskPanicked):
			result = native.WorldTaskPanicked
		case err != nil:
			result = native.WorldTaskFailed
		}
		if callbackErr := runtime.HandleWorldScheduled(plugin, callback, 0, native.WorldTaskComplete, result); callbackErr != nil {
			m.log.Error("complete native world task", "world", entry.name, "error", callbackErr)
		}
		m.scheduleMu.Lock()
		if m.scheduledTasks[key] == pending {
			delete(m.scheduledTasks, key)
		}
		m.scheduleMu.Unlock()
		m.scheduleWG.Done()
	})
	return true
}

func (m *WorldManager) CancelWorldTask(plugin, callback uint64) (bool, bool) {
	key := worldTaskKey{plugin: plugin, callback: callback}
	m.scheduleMu.Lock()
	pending, ok := m.scheduledTasks[key]
	m.scheduleMu.Unlock()
	if !ok {
		return false, false
	}
	return pending.cancel(), true
}

// StopScheduling rejects new managed world callbacks. DrainScheduled must be
// called before the native runtime is closed.
func (m *WorldManager) StopScheduling() {
	m.scheduleMu.Lock()
	m.scheduleClosing = true
	pending := make([]*scheduledWorldTask, 0, len(m.scheduledTasks))
	for _, task := range m.scheduledTasks {
		pending = append(pending, task)
	}
	m.scheduleMu.Unlock()
	for _, task := range pending {
		task.cancel()
	}
}

// DrainScheduled waits until every accepted callback has executed or dropped.
func (m *WorldManager) DrainScheduled() { m.scheduleWG.Wait() }

func (m *WorldManager) WorldByName(_ native.InvocationID, name string) (native.WorldID, bool) {
	entry, ok := m.entryByName(WorldID(name))
	if !ok {
		return 0, false
	}
	return entry.id, true
}

func (m *WorldManager) WorldName(invocation native.InvocationID, id native.WorldID) (string, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return "", false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return "", false
	}
	return entry.world.Name(), true
}

func (m *WorldManager) UnloadWorld(invocation native.InvocationID, id native.WorldID) bool {
	entry, ok := m.entryByHandle(id)
	return ok && m.unload(invocation, entry) == nil
}

// BlockByName resolves a block using the registry shared by the managed server worlds.
func (m *WorldManager) BlockByName(name string, encodedProperties []byte) (result native.WorldBlock, ok bool) {
	defer func() {
		if recover() != nil {
			result, ok = native.WorldBlock{}, false
		}
	}()
	properties, ok := decodeBlockProperties(encodedProperties)
	if !ok || name == "" {
		return native.WorldBlock{}, false
	}
	block, ok := world.BlockByName(name, properties)
	if !ok {
		return native.WorldBlock{}, false
	}
	identifier, state := block.EncodeBlock()
	encodedState, ok := encodeBlockProperties(state)
	if !ok {
		return native.WorldBlock{}, false
	}
	return native.WorldBlock{Identifier: identifier, PropertiesNBT: encodedState}, true
}

func (m *WorldManager) WorldBlock(invocation native.InvocationID, id native.WorldID, position native.BlockPos) (native.WorldBlock, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
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

func (m *WorldManager) WorldBlockLoaded(invocation native.InvocationID, id native.WorldID, position native.BlockPos) (native.WorldBlock, bool, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return native.WorldBlock{}, false, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return native.WorldBlock{}, false, false
	}
	read := func(tx *world.Tx) (native.WorldBlock, bool, bool) {
		value, ok := tx.BlockLoaded(blockPosition(position))
		if !ok {
			return native.WorldBlock{}, false, true
		}
		name, properties := value.EncodeBlock()
		encoded, ok := encodeBlockProperties(properties)
		return native.WorldBlock{Identifier: name, PropertiesNBT: encoded}, true, ok
	}
	if tx := m.currentTx(invocation, entry.world); tx != nil {
		return read(tx)
	}
	if invocation != 0 {
		return native.WorldBlock{}, false, false
	}
	type loadedResult struct {
		block  native.WorldBlock
		loaded bool
	}
	result, err := world.Call(context.Background(), entry.world, func(tx *world.Tx) (loadedResult, error) {
		block, loaded, ok := read(tx)
		if !ok {
			return loadedResult{}, errors.New("world block-loaded operation failed")
		}
		return loadedResult{block: block, loaded: loaded}, nil
	})
	return result.block, result.loaded, err == nil
}

func (m *WorldManager) OpenWorldBlocksWithin(invocation native.InvocationID, id native.WorldID, position native.BlockPos, radius int32, values []native.WorldBlock) (native.BlockIteratorID, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok || m.players == nil {
		return 0, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return 0, false
	}
	tx := m.currentTx(invocation, entry.world)
	if tx == nil {
		return 0, false
	}
	blocks := make([]world.Block, len(values))
	for index, value := range values {
		properties, valid := decodeBlockProperties(value.PropertiesNBT)
		if !valid || value.Identifier == "" {
			return 0, false
		}
		block, valid := entry.world.BlockRegistry().BlockByName(value.Identifier, properties)
		if !valid {
			return 0, false
		}
		blocks[index] = block
	}
	next, stop := iter.Pull(tx.BlocksWithin(blockPosition(position), int(radius), blocks...))
	iterator := &blockPositionIterator{invocation: invocation, next: next, stop: stop}
	m.blockIteratorMu.Lock()
	if m.nextBlockIter == native.BlockIteratorID(^uint64(0)) {
		m.blockIteratorMu.Unlock()
		iterator.close()
		return 0, false
	}
	m.nextBlockIter++
	iteratorID := m.nextBlockIter
	m.blockIterators[iteratorID] = iterator
	m.blockIteratorMu.Unlock()
	if !m.players.OnInvocationEnd(invocation, func() { m.closeWorldBlocks(invocation, iteratorID) }) {
		m.closeWorldBlocks(invocation, iteratorID)
		return 0, false
	}
	return iteratorID, true
}

func (m *WorldManager) NextWorldBlock(invocation native.InvocationID, id native.BlockIteratorID) (native.BlockPos, bool, bool) {
	m.blockIteratorMu.Lock()
	iterator, ok := m.blockIterators[id]
	m.blockIteratorMu.Unlock()
	if !ok || iterator.invocation != invocation {
		return native.BlockPos{}, false, false
	}
	position, found, valid := iterator.advance()
	if !valid || !found {
		m.closeWorldBlocks(invocation, id)
	}
	if !valid {
		return native.BlockPos{}, false, false
	}
	if !found {
		return native.BlockPos{}, false, true
	}
	if position.X() < math.MinInt32 || position.X() > math.MaxInt32 || position.Y() < math.MinInt32 || position.Y() > math.MaxInt32 || position.Z() < math.MinInt32 || position.Z() > math.MaxInt32 {
		m.closeWorldBlocks(invocation, id)
		return native.BlockPos{}, false, false
	}
	return native.BlockPos{X: int32(position.X()), Y: int32(position.Y()), Z: int32(position.Z())}, true, true
}

func (m *WorldManager) CloseWorldBlocks(invocation native.InvocationID, id native.BlockIteratorID) {
	m.closeWorldBlocks(invocation, id)
}

func (m *WorldManager) closeWorldBlocks(invocation native.InvocationID, id native.BlockIteratorID) {
	m.blockIteratorMu.Lock()
	iterator, ok := m.blockIterators[id]
	if ok && iterator.invocation == invocation {
		delete(m.blockIterators, id)
	} else {
		iterator = nil
	}
	m.blockIteratorMu.Unlock()
	if iterator != nil {
		iterator.close()
	}
}

func (m *WorldManager) WorldRange(invocation native.InvocationID, id native.WorldID) (native.BlockRange, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return native.BlockRange{}, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return native.BlockRange{}, false
	}
	value := entry.world.Range()
	if value.Min() < math.MinInt32 || value.Min() > math.MaxInt32 || value.Max() < math.MinInt32 || value.Max() > math.MaxInt32 {
		return native.BlockRange{}, false
	}
	return native.BlockRange{Min: int32(value.Min()), Max: int32(value.Max())}, true
}

func (m *WorldManager) WorldHighestLightBlocker(invocation native.InvocationID, id native.WorldID, x, z int32) (int32, bool) {
	return m.worldHeight(invocation, id, func(tx *world.Tx) int { return tx.HighestLightBlocker(int(x), int(z)) })
}

func (m *WorldManager) WorldHighestBlock(invocation native.InvocationID, id native.WorldID, x, z int32) (int32, bool) {
	return m.worldHeight(invocation, id, func(tx *world.Tx) int { return tx.HighestBlock(int(x), int(z)) })
}

func (m *WorldManager) worldHeight(invocation native.InvocationID, id native.WorldID, read func(*world.Tx) int) (int32, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return 0, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return 0, false
	}
	return readManagedWorld(m, invocation, entry, func(tx *world.Tx) (int32, bool) {
		value := read(tx)
		if value < math.MinInt32 || value > math.MaxInt32 {
			return 0, false
		}
		return int32(value), true
	})
}

func (m *WorldManager) WorldLight(invocation native.InvocationID, id native.WorldID, position native.BlockPos) (uint8, bool) {
	return m.worldLight(invocation, id, position, (*world.Tx).Light)
}

func (m *WorldManager) WorldSkyLight(invocation native.InvocationID, id native.WorldID, position native.BlockPos) (uint8, bool) {
	return m.worldLight(invocation, id, position, (*world.Tx).SkyLight)
}

func (m *WorldManager) worldLight(invocation native.InvocationID, id native.WorldID, position native.BlockPos, read func(*world.Tx, cube.Pos) uint8) (uint8, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return 0, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return 0, false
	}
	return readManagedWorld(m, invocation, entry, func(tx *world.Tx) (uint8, bool) {
		return read(tx, blockPosition(position)), true
	})
}

func (m *WorldManager) WorldLiquid(invocation native.InvocationID, id native.WorldID, position native.BlockPos) (native.WorldBlock, bool, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return native.WorldBlock{}, false, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return native.WorldBlock{}, false, false
	}
	read := func(tx *world.Tx) (native.WorldBlock, bool, bool) {
		liquid, ok := tx.Liquid(blockPosition(position))
		if !ok {
			return native.WorldBlock{}, false, true
		}
		name, properties := liquid.EncodeBlock()
		encoded, ok := encodeBlockProperties(properties)
		return native.WorldBlock{Identifier: name, PropertiesNBT: encoded}, true, ok
	}
	if tx := m.currentTx(invocation, entry.world); tx != nil {
		return read(tx)
	}
	if invocation != 0 {
		return native.WorldBlock{}, false, false
	}
	type liquidResult struct {
		block native.WorldBlock
		found bool
	}
	result, err := world.Call(context.Background(), entry.world, func(tx *world.Tx) (liquidResult, error) {
		block, found, valid := read(tx)
		if !valid {
			return liquidResult{}, errors.New("world liquid operation failed")
		}
		return liquidResult{block: block, found: found}, nil
	})
	return result.block, result.found, err == nil
}

func (m *WorldManager) SetWorldLiquid(invocation native.InvocationID, id native.WorldID, position native.BlockPos, value *native.WorldBlock) bool {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return false
	}
	var liquid world.Liquid
	if value != nil {
		if value.Identifier == "" {
			return false
		}
		properties, valid := decodeBlockProperties(value.PropertiesNBT)
		if !valid {
			return false
		}
		block, valid := entry.world.BlockRegistry().BlockByName(value.Identifier, properties)
		if !valid {
			return false
		}
		liquid, valid = block.(world.Liquid)
		if !valid {
			return false
		}
	}
	return m.writeTx(invocation, entry, func(tx *world.Tx) {
		tx.SetLiquid(blockPosition(position), liquid)
	})
}

func (m *WorldManager) SetWorldBlock(invocation native.InvocationID, id native.WorldID, position native.BlockPos, value native.WorldBlock, options native.WorldSetOpts) bool {
	entry, ok := m.entryForInvocation(invocation, id)
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
	return m.writeTx(invocation, entry, func(tx *world.Tx) {
		tx.SetBlock(blockPosition(position), block, &world.SetOpts{
			DisableBlockUpdates:       options.DisableBlockUpdates,
			DisableLiquidDisplacement: options.DisableLiquidDisplacement,
			DisableRedstoneUpdates:    options.DisableRedstoneUpdates,
		})
	})
}

func (m *WorldManager) ScheduleWorldBlockUpdate(invocation native.InvocationID, id native.WorldID, position native.BlockPos, value native.WorldBlock, delayNanoseconds int64) bool {
	entry, ok := m.entryForInvocation(invocation, id)
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
	return m.writeTx(invocation, entry, func(tx *world.Tx) {
		tx.ScheduleBlockUpdate(blockPosition(position), block, time.Duration(delayNanoseconds))
	})
}

func (m *WorldManager) WorldBiome(invocation native.InvocationID, id native.WorldID, position native.BlockPos) (int32, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return 0, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return 0, false
	}
	return readManagedWorld(m, invocation, entry, func(tx *world.Tx) (int32, bool) {
		biome := tx.Biome(blockPosition(position))
		if biome == nil {
			return 0, false
		}
		value := biome.EncodeBiome()
		if value < math.MinInt32 || value > math.MaxInt32 {
			return 0, false
		}
		return int32(value), true
	})
}

func (m *WorldManager) SetWorldBiome(invocation native.InvocationID, id native.WorldID, position native.BlockPos, biomeID int32) bool {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return false
	}
	biome, ok := world.BiomeByID(int(biomeID))
	if !ok || biome == nil {
		return false
	}
	return m.writeTx(invocation, entry, func(tx *world.Tx) {
		tx.SetBiome(blockPosition(position), biome)
	})
}

func (m *WorldManager) WorldTemperature(invocation native.InvocationID, id native.WorldID, position native.BlockPos) (float64, bool) {
	return worldPositionValue(m, invocation, id, position, (*world.Tx).Temperature)
}

func (m *WorldManager) WorldRainingAt(invocation native.InvocationID, id native.WorldID, position native.BlockPos) (bool, bool) {
	return worldPositionValue(m, invocation, id, position, (*world.Tx).RainingAt)
}

func (m *WorldManager) WorldSnowingAt(invocation native.InvocationID, id native.WorldID, position native.BlockPos) (bool, bool) {
	return worldPositionValue(m, invocation, id, position, (*world.Tx).SnowingAt)
}

func (m *WorldManager) WorldThunderingAt(invocation native.InvocationID, id native.WorldID, position native.BlockPos) (bool, bool) {
	return worldPositionValue(m, invocation, id, position, (*world.Tx).ThunderingAt)
}

func worldPositionValue[T any](m *WorldManager, invocation native.InvocationID, id native.WorldID, position native.BlockPos, read func(*world.Tx, cube.Pos) T) (T, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		var zero T
		return zero, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		var zero T
		return zero, false
	}
	return readManagedWorld(m, invocation, entry, func(tx *world.Tx) (T, bool) {
		pos := blockPosition(position)
		if tx.Biome(pos) == nil {
			var zero T
			return zero, false
		}
		return read(tx, pos), true
	})
}

func (m *WorldManager) WorldRaining(invocation native.InvocationID, id native.WorldID) (bool, bool) {
	return m.worldWeather(invocation, id, (*world.Tx).Raining)
}

func (m *WorldManager) WorldThundering(invocation native.InvocationID, id native.WorldID) (bool, bool) {
	return m.worldWeather(invocation, id, (*world.Tx).Thundering)
}

func (m *WorldManager) WorldCurrentTick(invocation native.InvocationID, id native.WorldID) (int64, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return 0, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return 0, false
	}
	return readManagedWorld(m, invocation, entry, func(tx *world.Tx) (int64, bool) {
		return tx.CurrentTick(), true
	})
}

func (m *WorldManager) worldWeather(invocation native.InvocationID, id native.WorldID, read func(*world.Tx) bool) (bool, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return false, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return false, false
	}
	return readManagedWorld(m, invocation, entry, func(tx *world.Tx) (bool, bool) {
		return read(tx), true
	})
}

func (m *WorldManager) entryForInvocation(invocation native.InvocationID, id native.WorldID) (*managedWorld, bool) {
	if id != 0 {
		if entry, ok := m.entryByHandle(id); ok {
			return entry, true
		}
	}
	if invocation == 0 || m.players == nil {
		return nil, false
	}
	tx, ok := m.players.InvocationTx(invocation)
	if !ok {
		return nil, false
	}
	w, ok := transactionWorld(tx)
	if !ok {
		return nil, false
	}
	m.mu.RLock()
	entry, ok := m.byWorld[w]
	m.mu.RUnlock()
	if id != 0 && (!ok || entry.id != id) {
		return nil, false
	}
	return entry, ok
}

func (m *WorldManager) WorldTime(invocation native.InvocationID, id native.WorldID) (int64, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
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

func (m *WorldManager) SetWorldTime(invocation native.InvocationID, id native.WorldID, value int64) bool {
	entry, ok := m.entryForInvocation(invocation, id)
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

func (m *WorldManager) WorldSpawn(invocation native.InvocationID, id native.WorldID) (native.BlockPos, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
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

func (m *WorldManager) SetWorldSpawn(invocation native.InvocationID, id native.WorldID, value native.BlockPos) bool {
	entry, ok := m.entryForInvocation(invocation, id)
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

func (m *WorldManager) WorldPlayerSpawn(invocation native.InvocationID, id native.WorldID, player [16]byte) (native.BlockPos, bool) {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return native.BlockPos{}, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return native.BlockPos{}, false
	}
	spawn := entry.world.PlayerSpawn(uuid.UUID(player))
	return native.BlockPos{X: int32(spawn.X()), Y: int32(spawn.Y()), Z: int32(spawn.Z())}, true
}

func (m *WorldManager) SetWorldPlayerSpawn(invocation native.InvocationID, id native.WorldID, player [16]byte, value native.BlockPos) bool {
	entry, ok := m.entryForInvocation(invocation, id)
	if !ok {
		return false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return false
	}
	entry.world.SetPlayerSpawn(uuid.UUID(player), blockPosition(value))
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
