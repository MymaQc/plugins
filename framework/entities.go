package framework

import (
	"context"
	"errors"
	"math"
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

func (m *WorldManager) WorldEntities(invocation native.InvocationID, id native.WorldID) ([]native.EntityID, bool) {
	entry, ok := m.entryByHandle(id)
	if !ok {
		return nil, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return nil, false
	}
	return readManagedWorld(m, invocation, entry, func(tx *world.Tx) ([]native.EntityID, bool) {
		ids := make([]native.EntityID, 0)
		for current := range tx.Entities() {
			id := m.entityHandles.Register(current)
			if id.Generation != 0 {
				ids = append(ids, id)
			}
		}
		return ids, true
	})
}

func (m *WorldManager) WorldPlayers(invocation native.InvocationID, id native.WorldID) ([]native.PlayerID, bool) {
	entry, ok := m.entryByHandle(id)
	if !ok || m.players == nil {
		return nil, false
	}
	entry.lifecycle.RLock()
	defer entry.lifecycle.RUnlock()
	if entry.closed {
		return nil, false
	}
	return readManagedWorld(m, invocation, entry, func(tx *world.Tx) ([]native.PlayerID, bool) {
		ids := make([]native.PlayerID, 0)
		for current := range tx.Players() {
			connected, ok := current.(*player.Player)
			if !ok {
				continue
			}
			id, ok := m.players.ID(connected)
			if ok {
				ids = append(ids, id)
			}
		}
		return ids, true
	})
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
		foreign, owned := entityType.(*foreignBaseEntityType)
		if found && owned {
			handle = opts.New(foreign, foreignBaseEntityConfig{})
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
