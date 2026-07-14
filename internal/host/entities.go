package host

import (
	"sync"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/world"
)

// Entities owns stable native IDs for Dragonfly entity handles. An ID remains
// valid until its handle is unregistered. Re-registering a handle receives a
// fresh generation, so an old plugin reference cannot resolve to a new entity.
type Entities struct {
	mu                   sync.RWMutex
	byHandle             map[*world.EntityHandle]entityEntry
	byID                 map[native.EntityID]*world.EntityHandle
	byHandleID           map[native.EntityHandleID]*world.EntityHandle
	handleTombstones     map[native.EntityHandleID][16]byte
	nextGeneration       uint64
	nextHandleGeneration uint64
}

type entityEntryState uint8

const (
	entityActive entityEntryState = iota
	entityTransferInactive
	entityDetached
)

type entityEntry struct {
	id       native.EntityID
	handleID native.EntityHandleID
	state    entityEntryState
	cleanup  func()
}

func NewEntities() *Entities {
	return &Entities{
		byHandle:         map[*world.EntityHandle]entityEntry{},
		byID:             map[native.EntityID]*world.EntityHandle{},
		byHandleID:       map[native.EntityHandleID]*world.EntityHandle{},
		handleTombstones: map[native.EntityHandleID][16]byte{},
	}
}

// Register returns the stable ID of entity. Registering the same live handle
// more than once returns the same ID.
func (e *Entities) Register(entity world.Entity) native.EntityID {
	if entity == nil {
		return native.EntityID{}
	}
	return e.registerHandle(entity.H(), 0)
}

// Unregister expires the ID associated with entity.
func (e *Entities) Unregister(entity world.Entity) {
	if entity != nil {
		e.unregisterHandle(entity.H())
	}
}

// ID returns the live stable ID associated with entity.
func (e *Entities) ID(entity world.Entity) (native.EntityID, bool) {
	if entity == nil {
		return native.EntityID{}, false
	}
	e.mu.RLock()
	entry, ok := e.byHandle[entity.H()]
	e.mu.RUnlock()
	return entry.id, ok && entry.state == entityActive
}

// Handle returns the live Dragonfly handle associated with id.
func (e *Entities) Handle(id native.EntityID) (*world.EntityHandle, bool) {
	if id.Generation == 0 {
		return nil, false
	}
	e.mu.RLock()
	handle, ok := e.byID[id]
	entry := e.byHandle[handle]
	e.mu.RUnlock()
	return handle, ok && entry.state == entityActive && handle != nil && !handle.Closed()
}

// EntityHandleID returns the stable handle identity associated with an active
// or transferring entity ID.
func (e *Entities) EntityHandleID(id native.EntityID) (native.EntityHandleID, bool) {
	if id.Generation == 0 {
		return native.EntityHandleID{}, false
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	handle, ok := e.byID[id]
	entry, current := e.byHandle[handle]
	if !ok || !current || entry.id != id || handle == nil || handle.Closed() {
		return native.EntityHandleID{}, false
	}
	if !entry.handleID.Valid() {
		var allocated bool
		entry.handleID, allocated = e.nextHandleIDLocked()
		if !allocated {
			return native.EntityHandleID{}, false
		}
		e.byHandle[handle] = entry
		e.byHandleID[entry.handleID] = handle
	}
	return entry.handleID, true
}

// HandleByID returns the Dragonfly handle associated with a stable handle ID.
// Unlike Handle, it also returns live handles that are between worlds or
// detached. Closed and expired handles remain inspectable through HandleUUID
// and HandleClosed without retaining the Dragonfly handle itself.
func (e *Entities) HandleByID(id native.EntityHandleID) (*world.EntityHandle, bool) {
	if !id.Valid() {
		return nil, false
	}
	e.mu.RLock()
	handle, ok := e.byHandleID[id]
	entry, current := e.byHandle[handle]
	e.mu.RUnlock()
	if !ok || !current || entry.handleID != id || handle == nil {
		return nil, false
	}
	return handle, true
}

// HandleUUID returns the UUID of a live or expired stable handle.
func (e *Entities) HandleUUID(id native.EntityHandleID) ([16]byte, bool) {
	if !id.Valid() {
		return [16]byte{}, false
	}
	e.mu.RLock()
	handle, ok := e.byHandleID[id]
	entry, current := e.byHandle[handle]
	if ok && current && entry.handleID == id && handle != nil {
		uuid := [16]byte(handle.UUID())
		e.mu.RUnlock()
		return uuid, true
	}
	uuid, ok := e.handleTombstones[id]
	e.mu.RUnlock()
	return uuid, ok
}

// HandleClosed reports whether a live or expired stable handle is closed.
// An expired token is closed from the plugin's perspective: it can never
// resolve or alias a later registration of the same Dragonfly handle.
func (e *Entities) HandleClosed(id native.EntityHandleID) (bool, bool) {
	if !id.Valid() {
		return false, false
	}
	e.mu.RLock()
	handle, ok := e.byHandleID[id]
	entry, current := e.byHandle[handle]
	if ok && current && entry.handleID == id && handle != nil {
		closed := handle.Closed()
		e.mu.RUnlock()
		return closed, true
	}
	_, ok = e.handleTombstones[id]
	e.mu.RUnlock()
	return ok, ok
}

// CloseHandle closes and expires one stable handle. Repeated closes of an
// expired token are harmless, matching EntityHandle.Close's idempotence.
func (e *Entities) CloseHandle(id native.EntityHandleID) bool {
	if !id.Valid() {
		return false
	}
	e.mu.Lock()
	handle, ok := e.byHandleID[id]
	entry, current := e.byHandle[handle]
	if !ok || !current || entry.handleID != id || handle == nil {
		_, known := e.handleTombstones[id]
		e.mu.Unlock()
		return known
	}
	cleanup := e.expireHandleLocked(handle, entry)
	e.mu.Unlock()
	_ = handle.Close()
	if cleanup != nil {
		cleanup()
	}
	return true
}

// DetachedHandle returns a live worldless handle that was produced by Detach.
func (e *Entities) DetachedHandle(id native.EntityHandleID) (*world.EntityHandle, bool) {
	handle, ok := e.HandleByID(id)
	if !ok {
		return nil, false
	}
	e.mu.RLock()
	entry, current := e.byHandle[handle]
	e.mu.RUnlock()
	if !current || entry.handleID != id || entry.state != entityDetached || handle.Closed() {
		return nil, false
	}
	return handle, true
}

// RegisterDetached publishes a newly-created worldless handle. Unlike Detach,
// no world-bound entity ID exists yet.
func (e *Entities) RegisterDetached(handle *world.EntityHandle, cleanups ...func()) (native.EntityHandleID, bool) {
	if handle == nil || handle.Closed() {
		return native.EntityHandleID{}, false
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, exists := e.byHandle[handle]; exists {
		return native.EntityHandleID{}, false
	}
	id, ok := e.nextHandleIDLocked()
	if !ok {
		return native.EntityHandleID{}, false
	}
	var cleanup func()
	if len(cleanups) != 0 {
		cleanup = cleanups[0]
	}
	e.byHandle[handle] = entityEntry{handleID: id, state: entityDetached, cleanup: cleanup}
	e.byHandleID[id] = handle
	return id, true
}

// EnsureHandle returns stable handle identity before an entity has reached
// world handler registration. Provider-loaded custom entities need this from
// EntityType.Open. Later Register promotes the same entry to active.
func (e *Entities) EnsureHandle(handle *world.EntityHandle, cleanup func()) (native.EntityHandleID, bool) {
	if handle == nil || handle.Closed() {
		return native.EntityHandleID{}, false
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if entry, ok := e.byHandle[handle]; ok {
		if !entry.handleID.Valid() {
			var allocated bool
			entry.handleID, allocated = e.nextHandleIDLocked()
			if !allocated {
				return native.EntityHandleID{}, false
			}
			e.byHandleID[entry.handleID] = handle
		}
		if entry.cleanup == nil {
			entry.cleanup = cleanup
		}
		e.byHandle[handle] = entry
		return entry.handleID, true
	}
	id, ok := e.nextHandleIDLocked()
	if !ok {
		return native.EntityHandleID{}, false
	}
	e.byHandle[handle] = entityEntry{handleID: id, state: entityDetached, cleanup: cleanup}
	e.byHandleID[id] = handle
	return id, true
}

// ResolveHandle opens a stable handle in the exact transaction passed.
func (e *Entities) ResolveHandle(id native.EntityHandleID, tx *world.Tx) (entity world.Entity, ok bool) {
	handle, ok := e.HandleByID(id)
	if !ok || tx == nil {
		return nil, false
	}
	defer func() {
		if recover() != nil {
			entity, ok = nil, false
		}
	}()
	return handle.Entity(tx)
}

// Detach expires an entity's world-bound ID while preserving its stable
// EntityHandleID for a later AddEntity or AddEntityAt call.
func (e *Entities) Detach(id native.EntityID) (native.EntityHandleID, *world.EntityHandle, bool) {
	if id.Generation == 0 {
		return native.EntityHandleID{}, nil, false
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	handle, ok := e.byID[id]
	entry, current := e.byHandle[handle]
	if !ok || !current || entry.id != id || entry.state == entityDetached || handle == nil || handle.Closed() {
		return native.EntityHandleID{}, nil, false
	}
	if !entry.handleID.Valid() {
		var allocated bool
		entry.handleID, allocated = e.nextHandleIDLocked()
		if !allocated {
			return native.EntityHandleID{}, nil, false
		}
		e.byHandleID[entry.handleID] = handle
	}
	delete(e.byID, id)
	entry.id = native.EntityID{}
	entry.state = entityDetached
	e.byHandle[handle] = entry
	return entry.handleID, handle, true
}

// Resolve opens id in the exact transaction passed. It never scans worlds or
// schedules work on another owner.
func (e *Entities) Resolve(id native.EntityID, tx *world.Tx) (entity world.Entity, ok bool) {
	handle, ok := e.Handle(id)
	if !ok || tx == nil {
		return nil, false
	}
	defer func() {
		if recover() != nil {
			entity, ok = nil, false
		}
	}()
	return handle.Entity(tx)
}

func (e *Entities) registerHandle(handle *world.EntityHandle, generation uint64) native.EntityID {
	if handle == nil || handle.Closed() {
		return native.EntityID{}
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if entry, ok := e.byHandle[handle]; ok {
		if entry.state == entityDetached {
			entry.id = e.newEntityIDLocked(handle, 0)
			if entry.id.Generation == 0 {
				return native.EntityID{}
			}
			e.byID[entry.id] = handle
		}
		entry.state = entityActive
		e.byHandle[handle] = entry
		return entry.id
	}
	id := e.newEntityIDLocked(handle, generation)
	if id.Generation == 0 {
		return native.EntityID{}
	}
	e.byHandle[handle] = entityEntry{id: id, state: entityActive}
	e.byID[id] = handle
	return id
}

func (e *Entities) newEntityIDLocked(handle *world.EntityHandle, generation uint64) native.EntityID {
	if generation == 0 {
		var ok bool
		generation, ok = e.nextAvailableGenerationLocked()
		if !ok {
			return native.EntityID{}
		}
	} else if generation > e.nextGeneration {
		e.nextGeneration = generation
	}
	id := native.EntityID{Generation: generation}
	uuid := handle.UUID()
	copy(id.UUID[:], uuid[:])
	for {
		if existing, found := e.byID[id]; !found || existing == handle {
			break
		}
		generation, ok := e.nextAvailableGenerationLocked()
		if !ok {
			return native.EntityID{}
		}
		id.Generation = generation
	}
	return id
}

// deactivateHandle keeps identity across a Dragonfly world transfer while
// preventing operations from waiting on a handle that is not in any world.
func (e *Entities) deactivateHandle(handle *world.EntityHandle) {
	if handle == nil {
		return
	}
	e.mu.Lock()
	if entry, ok := e.byHandle[handle]; ok {
		if entry.state == entityActive {
			entry.state = entityTransferInactive
			e.byHandle[handle] = entry
		}
	}
	e.mu.Unlock()
}

func (e *Entities) nextHandleIDLocked() (native.EntityHandleID, bool) {
	if e.nextHandleGeneration == ^uint64(0) {
		return native.EntityHandleID{}, false
	}
	e.nextHandleGeneration++
	return native.EntityHandleID{
		Value:      e.nextHandleGeneration,
		Generation: e.nextHandleGeneration,
	}, true
}

func (e *Entities) nextAvailableGenerationLocked() (uint64, bool) {
	if e.nextGeneration == ^uint64(0) {
		return 0, false
	}
	e.nextGeneration++
	return e.nextGeneration, true
}

func (e *Entities) unregisterHandle(handle *world.EntityHandle) {
	if handle == nil {
		return
	}
	e.mu.Lock()
	var cleanup func()
	if entry, ok := e.byHandle[handle]; ok {
		cleanup = e.expireHandleLocked(handle, entry)
	}
	e.mu.Unlock()
	if cleanup != nil {
		cleanup()
	}
}

func (e *Entities) expireHandleLocked(handle *world.EntityHandle, entry entityEntry) func() {
	delete(e.byHandle, handle)
	if entry.id.Generation != 0 {
		delete(e.byID, entry.id)
	}
	if entry.handleID.Valid() {
		delete(e.byHandleID, entry.handleID)
		e.handleTombstones[entry.handleID] = [16]byte(handle.UUID())
	}
	return entry.cleanup
}
