package framework

import (
	"context"
	"errors"
	"iter"
	"math"
	"sync"
	"time"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type worldEntityIterator struct {
	mu         sync.Mutex
	invocation native.InvocationID
	next       func() (world.Entity, bool)
	stop       func()
	stopped    bool
}

func (i *worldEntityIterator) advance() (entity world.Entity, found bool, valid bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.stopped {
		return nil, false, false
	}
	defer func() {
		if recover() != nil {
			entity, found, valid = nil, false, false
		}
	}()
	entity, found = i.next()
	return entity, found, true
}

func (i *worldEntityIterator) close() {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.stopped {
		return
	}
	i.stopped = true
	defer func() { _ = recover() }()
	i.stop()
}

type velocityEntity interface {
	world.Entity
	Velocity() mgl64.Vec3
	SetVelocity(mgl64.Vec3)
}

type teleportEntity interface {
	world.Entity
	Teleport(mgl64.Vec3)
}

type nameTagEntity interface {
	world.Entity
	NameTag() string
	SetNameTag(string)
}

func (m *WorldManager) SpawnWorldEntity(invocation native.InvocationID, id native.WorldID, value native.EntitySpawn) (native.EntityID, bool) {
	entry, ok := m.entryByHandle(id)
	if !ok || !validEntitySpawn(value) {
		return native.EntityID{}, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return native.EntityID{}, false
	}
	return readManagedWorld(m, invocation, entry, func(tx *world.Tx) (native.EntityID, bool) {
		handle, ok := m.newEntityHandle(tx, value)
		if !ok || handle == nil {
			return native.EntityID{}, false
		}
		spawned := tx.AddEntity(handle)
		entityID := m.entityHandles.Register(spawned)
		return entityID, entityID.Generation != 0
	})
}

// CurrentWorld resolves the exact managed world owned by invocation's active
// Dragonfly transaction. It deliberately rejects the context-free invocation.
func (m *WorldManager) CurrentWorld(invocation native.InvocationID) (native.WorldID, bool) {
	if invocation == 0 {
		return 0, false
	}
	tx, ok := m.invocationTx(invocation)
	if !ok {
		return 0, false
	}
	w, ok := transactionWorld(tx)
	if !ok {
		return 0, false
	}
	return m.handleByWorld(w)
}

// OpenWorldEntityIterator opens a lazy pull iterator over the exact current
// transaction. Passing any world other than CurrentWorld is rejected.
func (m *WorldManager) OpenWorldEntityIterator(invocation native.InvocationID, id native.WorldID, playersOnly bool) (native.EntityIteratorID, bool) {
	currentID, ok := m.CurrentWorld(invocation)
	if !ok || id == 0 || id != currentID || m.players == nil {
		return 0, false
	}
	tx, ok := m.invocationTx(invocation)
	if !ok {
		return 0, false
	}
	sequence := tx.Entities()
	if playersOnly {
		sequence = tx.Players()
	}
	next, stop := iter.Pull(sequence)
	iterator := &worldEntityIterator{invocation: invocation, next: next, stop: stop}
	m.entityIteratorMu.Lock()
	if m.nextEntityIter == native.EntityIteratorID(^uint64(0)) {
		m.entityIteratorMu.Unlock()
		iterator.close()
		return 0, false
	}
	m.nextEntityIter++
	iteratorID := m.nextEntityIter
	m.entityIterators[iteratorID] = iterator
	m.entityIteratorMu.Unlock()
	if !m.players.OnInvocationEnd(invocation, func() { m.closeWorldEntities(invocation, iteratorID) }) {
		m.closeWorldEntities(invocation, iteratorID)
		return 0, false
	}
	return iteratorID, true
}

// NextWorldEntity advances one invocation-scoped iterator. The second result
// reports whether a value was yielded and the third whether the call was valid.
func (m *WorldManager) NextWorldEntity(invocation native.InvocationID, id native.EntityIteratorID) (native.EntityID, bool, bool) {
	m.entityIteratorMu.Lock()
	iterator, ok := m.entityIterators[id]
	m.entityIteratorMu.Unlock()
	if !ok || iterator.invocation != invocation {
		return native.EntityID{}, false, false
	}
	current, found, valid := iterator.advance()
	if !valid || !found {
		m.closeWorldEntities(invocation, id)
	}
	if !valid {
		return native.EntityID{}, false, false
	}
	if !found {
		return native.EntityID{}, false, true
	}
	entityID := m.entityHandles.Register(current)
	if entityID.Generation == 0 {
		m.closeWorldEntities(invocation, id)
		return native.EntityID{}, false, false
	}
	return entityID, true, true
}

func (m *WorldManager) CloseWorldEntities(invocation native.InvocationID, id native.EntityIteratorID) {
	m.closeWorldEntities(invocation, id)
}

func (m *WorldManager) closeWorldEntities(invocation native.InvocationID, id native.EntityIteratorID) {
	m.entityIteratorMu.Lock()
	iterator, ok := m.entityIterators[id]
	if ok && iterator.invocation == invocation {
		delete(m.entityIterators, id)
	} else {
		iterator = nil
	}
	m.entityIteratorMu.Unlock()
	if iterator != nil {
		iterator.close()
	}
}

// EntityHandle returns the stable Dragonfly handle for an entity in the
// invocation's exact transaction.
func (m *WorldManager) EntityHandle(invocation native.InvocationID, id native.EntityID) (native.EntityHandleID, bool) {
	tx, ok := m.invocationTx(invocation)
	if !ok {
		return native.EntityHandleID{}, false
	}
	if _, ok := m.entityHandles.Resolve(id, tx); !ok {
		return native.EntityHandleID{}, false
	}
	return m.entityHandles.EntityHandleID(id)
}

// EntityHandleEntity resolves a stable handle in the invocation's exact
// transaction, matching world.EntityHandle.Entity.
func (m *WorldManager) EntityHandleEntity(invocation native.InvocationID, id native.EntityHandleID) (native.EntityID, bool, bool) {
	tx, ok := m.invocationTx(invocation)
	if !ok {
		return native.EntityID{}, false, false
	}
	current, found := m.entityHandles.ResolveHandle(id, tx)
	if !found {
		return native.EntityID{}, false, true
	}
	entityID := m.entityHandles.Register(current)
	return entityID, entityID.Generation != 0, entityID.Generation != 0
}

func (m *WorldManager) EntityHandleUUID(id native.EntityHandleID) ([16]byte, bool) {
	return m.entityHandles.HandleUUID(id)
}

func (m *WorldManager) EntityHandleClosed(id native.EntityHandleID) (bool, bool) {
	return m.entityHandles.HandleClosed(id)
}

func (m *WorldManager) CloseEntityHandle(id native.EntityHandleID) bool {
	m.detachedEntityMu.Lock()
	defer m.detachedEntityMu.Unlock()
	cleanup := m.detachedEntities[id]
	delete(m.detachedEntities, id)
	if cleanup != nil {
		cleanup()
	}
	return m.entityHandles.CloseHandle(id)
}

// RemoveEntity removes an entity through the invocation's exact transaction.
// Player handles remain owned by their session; Player.ChangeWorld is the safe
// player transfer path until Dragonfly exposes that session operation directly.
func (m *WorldManager) RemoveEntity(invocation native.InvocationID, id native.EntityID) (result native.EntityHandleID, ok bool) {
	tx, ok := m.invocationTx(invocation)
	if !ok {
		return native.EntityHandleID{}, false
	}
	current, ok := m.entityHandles.Resolve(id, tx)
	if !ok {
		return native.EntityHandleID{}, false
	}
	if _, playerEntity := current.(*player.Player); playerEntity {
		return native.EntityHandleID{}, false
	}
	cleanup := advancedEntityCleanup(current)
	m.detachedEntityMu.Lock()
	defer m.detachedEntityMu.Unlock()
	defer func() {
		if recover() != nil {
			result, ok = native.EntityHandleID{}, false
		}
	}()
	handle := tx.RemoveEntity(current)
	if handle == nil {
		return native.EntityHandleID{}, false
	}
	result, detachedHandle, ok := m.entityHandles.Detach(id)
	if !ok || detachedHandle != handle {
		_ = handle.Close()
		if cleanup != nil {
			cleanup()
		}
		return native.EntityHandleID{}, false
	}
	m.detachedEntities[result] = cleanup
	return result, true
}

// AddEntity adds a worldless Dragonfly handle through the invocation's exact
// transaction. A non-nil position selects Tx.AddEntityAt.
func (m *WorldManager) AddEntity(invocation native.InvocationID, id native.EntityHandleID, position *native.Vec3) (result native.EntityID, ok bool) {
	tx, ok := m.invocationTx(invocation)
	if !ok {
		return native.EntityID{}, false
	}
	m.detachedEntityMu.Lock()
	defer m.detachedEntityMu.Unlock()
	handle, ok := m.entityHandles.DetachedHandle(id)
	if !ok {
		return native.EntityID{}, false
	}
	defer func() {
		if recover() != nil {
			result, ok = native.EntityID{}, false
		}
	}()
	var current world.Entity
	if position == nil {
		current = tx.AddEntity(handle)
	} else {
		current = tx.AddEntityAt(handle, vec3(*position))
	}
	result = m.entityHandles.Register(current)
	if result.Generation == 0 {
		return native.EntityID{}, false
	}
	delete(m.detachedEntities, id)
	return result, true
}

func (m *WorldManager) DrainDetachedEntities() {
	m.detachedEntityMu.Lock()
	detached := m.detachedEntities
	m.detachedEntities = make(map[native.EntityHandleID]func())
	m.detachedEntityMu.Unlock()
	for id, cleanup := range detached {
		if cleanup != nil {
			cleanup()
		}
		m.entityHandles.CloseHandle(id)
	}
}

func (m *WorldManager) EntityState(invocation native.InvocationID, id native.EntityID) (native.EntityState, bool) {
	handle, ok := m.entityHandles.Handle(id)
	if !ok {
		return native.EntityState{}, false
	}
	read := func(tx *world.Tx, current world.Entity) native.EntityState {
		position, rotation := current.Position(), current.Rotation()
		worldID, _ := m.handleByWorld(tx.World())
		state := native.EntityState{
			Position: nativeVec3(position), Rotation: native.Rotation{Yaw: rotation.Yaw(), Pitch: rotation.Pitch()},
			World: worldID, Type: handle.Type().EncodeEntity(),
		}
		if moving, ok := current.(velocityEntity); ok {
			state.Velocity, state.HasVelocity = nativeVec3(moving.Velocity()), true
		}
		if named, ok := current.(nameTagEntity); ok {
			state.NameTag, state.HasNameTag = named.NameTag(), true
		}
		_, state.CanTeleport = current.(teleportEntity)
		return state
	}
	if tx, ok := m.invocationTx(invocation); ok {
		current, ok := m.entityHandles.Resolve(id, tx)
		if !ok {
			return native.EntityState{}, false
		}
		return read(tx, current), true
	}
	if invocation != 0 {
		return native.EntityState{}, false
	}
	state, err := world.CallRef(context.Background(), world.NewEntityRef[world.Entity](handle), func(tx *world.Tx, current world.Entity) (native.EntityState, error) {
		return read(tx, current), nil
	})
	return state, err == nil
}

func (m *WorldManager) TeleportEntity(invocation native.InvocationID, id native.EntityID, value native.Vec3) bool {
	if !finiteVec3(value) {
		return false
	}
	return m.mutateEntity(invocation, id, func(current world.Entity) {
		if moving, ok := current.(teleportEntity); ok {
			moving.Teleport(vec3(value))
		}
	})
}

func (m *WorldManager) SetEntityVelocity(invocation native.InvocationID, id native.EntityID, value native.Vec3) bool {
	if !finiteVec3(value) {
		return false
	}
	return m.mutateEntity(invocation, id, func(current world.Entity) {
		if moving, ok := current.(velocityEntity); ok {
			moving.SetVelocity(vec3(value))
		}
	})
}

func (m *WorldManager) SetEntityNameTag(invocation native.InvocationID, id native.EntityID, value string) bool {
	return m.mutateEntity(invocation, id, func(current world.Entity) {
		if named, ok := current.(nameTagEntity); ok {
			named.SetNameTag(value)
		}
	})
}

func (m *WorldManager) DespawnEntity(invocation native.InvocationID, id native.EntityID) bool {
	return m.mutateEntity(invocation, id, func(current world.Entity) { _ = current.Close() })
}

func (m *WorldManager) mutateEntity(invocation native.InvocationID, id native.EntityID, mutate func(world.Entity)) bool {
	handle, ok := m.entityHandles.Handle(id)
	if !ok {
		return false
	}
	if tx, ok := m.invocationTx(invocation); ok {
		if current, sameWorld := m.entityHandles.Resolve(id, tx); sameWorld {
			mutate(current)
			return true
		}
	} else if invocation != 0 {
		return false
	}
	world.NewEntityRef[world.Entity](handle).Do(func(_ *world.Tx, current world.Entity) { mutate(current) })
	return true
}

func (m *WorldManager) invocationTx(invocation native.InvocationID) (*world.Tx, bool) {
	if m.players == nil {
		return nil, false
	}
	return m.players.InvocationTx(invocation)
}

func readManagedWorld[T any](m *WorldManager, invocation native.InvocationID, entry *managedWorld, read func(*world.Tx) (T, bool)) (T, bool) {
	if tx := m.currentTx(invocation, entry.world); tx != nil {
		return read(tx)
	}
	var zero T
	if invocation != 0 {
		return zero, false
	}
	value, err := world.Call(context.Background(), entry.world, func(tx *world.Tx) (T, error) {
		value, ok := read(tx)
		if !ok {
			return zero, errors.New("world entity operation failed")
		}
		return value, nil
	})
	return value, err == nil
}

func (m *WorldManager) newEntityHandle(tx *world.Tx, value native.EntitySpawn) (*world.EntityHandle, bool) {
	fuse, ok := entityDuration(value.FuseMilliseconds)
	if !ok {
		return nil, false
	}
	opts := world.EntitySpawnOpts{
		Position: vec3(value.Position), Rotation: rotation(value.Rotation), Velocity: vec3(value.Velocity), NameTag: value.NameTag,
	}
	registry := tx.World().EntityRegistry().Config()
	var handle *world.EntityHandle
	switch value.Kind {
	case native.EntityText:
		handle = entity.NewText(value.Text, opts.Position)
	case native.EntityLightning:
		handle = entity.NewLightningWithDamage(opts, value.Damage, value.Flags&native.EntityLightningBlockFire != 0, fuse)
	case native.EntityTNT:
		if registry.TNT != nil {
			handle = registry.TNT(opts, fuse)
		}
	case native.EntityExperienceOrb:
		handle = entity.NewExperienceOrb(opts, int(value.Experience))
	case native.EntityItem:
		if registry.Item != nil && value.Item != nil {
			stack, ok := host.ItemStackFromNative(*value.Item)
			if ok {
				if value.Flags&native.EntityItemHasPickupDelay != 0 {
					handle = entity.NewItemPickupDelay(opts, stack, fuse)
				} else {
					handle = registry.Item(opts, stack)
				}
			}
		}
	case native.EntityFallingBlock:
		if registry.FallingBlock != nil && value.Block != nil {
			properties, ok := decodeBlockProperties(value.Block.PropertiesNBT)
			if ok {
				block, ok := tx.World().BlockRegistry().BlockByName(value.Block.Identifier, properties)
				if ok {
					handle = registry.FallingBlock(opts, block)
				}
			}
		}
	case native.EntityCustom:
		entityType, found := tx.World().EntityRegistry().Lookup(value.Type)
		if found && isForeignEntityType(entityType) {
			instance := native.EntityInstanceID(0)
			if advanced, ok := foreignAdvancedType(entityType); ok && advanced.services.runtime != nil {
				if value.CustomInstance == 0 {
					return nil, false
				}
				adopted, adoptErr := advanced.services.runtime.EntityAdopt(advanced.definition.TypeKey, value.CustomInstance)
				if adoptErr != nil || adopted == 0 {
					return nil, false
				}
				instance = adopted
			}
			handle = opts.New(entityType, foreignEntityConfigFor(entityType, instance))
			if handle == nil && instance != 0 {
				advanced, _ := foreignAdvancedType(entityType)
				advanced.services.runtime.EntityDestroy(instance)
			}
		}
	default:
		owner, ok := m.entityHandles.Resolve(value.Owner, tx)
		if !ok {
			return nil, false
		}
		switch value.Kind {
		case native.EntityArrow:
			if registry.Arrow != nil {
				handle = registry.Arrow(opts, world.ArrowSpawnConfig{
					Damage: value.Damage, Owner: owner, Critical: value.Flags&native.EntityArrowCritical != 0,
					DisablePickup:       value.Flags&native.EntityArrowDisablePickup != 0,
					ObtainArrowOnPickup: value.Flags&native.EntityArrowObtainOnPickup != 0,
					PunchLevel:          int(value.Punch), PiercingLevel: int(value.Piercing), Tip: potion.From(int32(value.Potion)),
				})
			}
		case native.EntityEgg:
			if registry.Egg != nil {
				handle = registry.Egg(opts, owner)
			}
		case native.EntitySnowball:
			if registry.Snowball != nil {
				handle = registry.Snowball(opts, owner)
			}
		case native.EntityEnderPearl:
			if registry.EnderPearl != nil {
				handle = registry.EnderPearl(opts, owner)
			}
		case native.EntityBottleOfEnchanting:
			if registry.BottleOfEnchanting != nil {
				handle = registry.BottleOfEnchanting(opts, owner)
			}
		case native.EntitySplashPotion:
			if registry.SplashPotion != nil {
				handle = registry.SplashPotion(opts, potion.From(int32(value.Potion)), owner)
			}
		case native.EntityLingeringPotion:
			if registry.LingeringPotion != nil {
				handle = registry.LingeringPotion(opts, potion.From(int32(value.Potion)), owner)
			}
		}
	}
	return handle, handle != nil
}

func validEntitySpawn(value native.EntitySpawn) bool {
	return value.Kind <= native.EntityCustom && finiteVec3(value.Position) && finiteVec3(value.Velocity) &&
		finiteFloat(value.Rotation.Yaw) && finiteFloat(value.Rotation.Pitch) && finiteFloat(value.Damage) &&
		value.Experience >= 0 && value.Punch >= 0 && value.Piercing >= 0 && value.Potion <= math.MaxUint8 &&
		(value.Kind != native.EntityCustom || value.Type != "")
}

func finiteVec3(value native.Vec3) bool {
	return finiteFloat(value.X) && finiteFloat(value.Y) && finiteFloat(value.Z)
}

func finiteFloat(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func entityDuration(milliseconds uint64) (time.Duration, bool) {
	const maximum = uint64(math.MaxInt64 / int64(time.Millisecond))
	if milliseconds > maximum {
		return 0, false
	}
	return time.Duration(milliseconds) * time.Millisecond, true
}

func vec3(value native.Vec3) mgl64.Vec3 { return mgl64.Vec3{value.X, value.Y, value.Z} }
func nativeVec3(value mgl64.Vec3) native.Vec3 {
	return native.Vec3{X: value.X(), Y: value.Y(), Z: value.Z()}
}
func rotation(value native.Rotation) cube.Rotation {
	return cube.Rotation{value.Yaw, value.Pitch}
}
