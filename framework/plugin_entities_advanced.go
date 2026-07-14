package framework

import (
	"math"
	"sync"
	"time"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type foreignEntityRuntime interface {
	EntityAdopt(uint64, uint64) (native.EntityInstanceID, error)
	EntityLoad(uint64, native.EntityLoadInput) (native.EntityInstanceID, error)
	EntitySave(native.EntityInstanceID) (native.EntitySaveOutput, error)
	EntityTick(native.EntityInstanceID, native.EntityTickInput) (native.EntityTickOutput, error)
	EntityHurt(native.EntityInstanceID, native.EntityHurtInput) (native.EntityHurtOutput, error)
	EntityHeal(native.EntityInstanceID, native.EntityHealInput) (native.EntityHealOutput, error)
	EntityDeath(native.EntityInstanceID, native.EntityDeathInput) (native.EntityDeathOutput, error)
	EntityDestroy(native.EntityInstanceID)
}

type foreignEntityServices struct {
	runtime  foreignEntityRuntime
	players  *host.Players
	entities *host.Entities
}

type foreignAdvancedEntityType struct {
	definition native.EntityTypeDefinition
	bbox       cube.BBox
	services   foreignEntityServices
}

const (
	foreignEntityStateDataKey    = "bedrock_gophers:state"
	foreignEntityStateVersionKey = "bedrock_gophers:state_version"
	foreignEntityHealthKey       = "bedrock_gophers:health"
	foreignEntityMaxHealthKey    = "bedrock_gophers:max_health"
	foreignEntitySpeedKey        = "bedrock_gophers:speed"
	foreignEntityEffectsKey      = "bedrock_gophers:effects"
	foreignEntityInvalidStateKey = "bedrock_gophers:invalid_state"
)

func (t *foreignAdvancedEntityType) EncodeEntity() string        { return t.definition.SaveID }
func (t *foreignAdvancedEntityType) NetworkEncodeEntity() string { return t.definition.NetworkID }
func (t *foreignAdvancedEntityType) BBox(world.Entity) cube.BBox { return t.bbox }

func (t *foreignAdvancedEntityType) DecodeNBT(nbt map[string]any, data *world.EntityData) {
	state := t.newState(0)
	if t.definition.Family == native.EntityFamilyLiving {
		health := nbtFloat64(nbt[foreignEntityHealthKey], t.definition.InitialHealth)
		maximum := nbtFloat64(nbt[foreignEntityMaxHealthKey], t.definition.MaxHealth)
		if maximum <= 0 || math.IsNaN(maximum) || math.IsInf(maximum, 0) {
			maximum = t.definition.MaxHealth
		}
		if health < 0 || health > maximum || math.IsNaN(health) || math.IsInf(health, 0) {
			health = t.definition.InitialHealth
		}
		state.health = entity.NewHealthManager(health, maximum)
		state.baseSpeed = nbtFloat64(nbt[foreignEntitySpeedKey], t.definition.Speed)
		if state.baseSpeed < 0 || math.IsNaN(state.baseSpeed) || math.IsInf(state.baseSpeed, 0) {
			state.baseSpeed = t.definition.Speed
		}
		state.speed = state.baseSpeed
		state.effects = entity.NewEffectManager(decodeForeignEffects(nbt[foreignEntityEffectsKey])...)
	}
	state.invalidPersistence = nbtBool(nbt[foreignEntityInvalidStateKey])
	if t.definition.CallbackFlags&native.EntityCallbackState != 0 {
		if encoded, ok := nbt[foreignEntityStateDataKey].([]byte); ok {
			state.pluginState = append([]byte(nil), encoded...)
			state.pluginStateVersion = uint32(nbtUint64(nbt[foreignEntityStateVersionKey]))
			state.hasPluginState = true
		}
	}
	if t.services.runtime != nil && !state.invalidPersistence {
		instance, err := t.services.runtime.EntityLoad(t.definition.TypeKey, native.EntityLoadInput{
			Data: append([]byte(nil), state.pluginState...), Version: state.pluginStateVersion,
		})
		if err == nil {
			state.instance = instance
		}
	}
	data.Data = state
}

func (t *foreignAdvancedEntityType) EncodeNBT(data *world.EntityData) map[string]any {
	state, ok := data.Data.(*foreignEntityState)
	if !ok {
		return nil
	}
	nbt := map[string]any{}
	if t.definition.Family == native.EntityFamilyLiving {
		nbt[foreignEntityHealthKey] = state.health.Health()
		nbt[foreignEntityMaxHealthKey] = state.health.MaxHealth()
		nbt[foreignEntitySpeedKey] = state.baseSpeed
		nbt[foreignEntityEffectsKey] = encodeForeignEffects(state.effects.Effects())
	}
	if t.definition.CallbackFlags&native.EntityCallbackState != 0 {
		if t.services.runtime != nil && state.instance != 0 {
			if encoded, err := t.services.runtime.EntitySave(state.instance); err == nil {
				state.pluginState = append(state.pluginState[:0], encoded.Data...)
				state.pluginStateVersion = encoded.Version
				state.hasPluginState = true
			}
		}
		if state.hasPluginState {
			nbt[foreignEntityStateDataKey] = append([]byte(nil), state.pluginState...)
			nbt[foreignEntityStateVersionKey] = int64(state.pluginStateVersion)
		} else {
			nbt[foreignEntityInvalidStateKey] = uint8(1)
		}
	}
	return nbt
}

func nbtFloat64(value any, fallback float64) float64 {
	switch value := value.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int64:
		return float64(value)
	case int32:
		return float64(value)
	default:
		return fallback
	}
}

func nbtUint64(value any) uint64 {
	switch value := value.(type) {
	case uint64:
		return value
	case int64:
		return uint64(max(value, 0))
	case int32:
		return uint64(max(value, 0))
	default:
		return 0
	}
}

func encodeForeignEffects(values []effect.Effect) []map[string]any {
	encoded := make([]map[string]any, 0, len(values))
	for _, value := range values {
		id, ok := effect.ID(value.Type())
		if !ok {
			continue
		}
		encoded = append(encoded, map[string]any{
			"id": int32(id), "level": int32(value.Level()), "duration": int64(value.Duration()),
			"ambient": boolByte(value.Ambient()), "hidden": boolByte(value.ParticlesHidden()), "infinite": boolByte(value.Infinite()),
		})
	}
	return encoded
}

func decodeForeignEffects(value any) []effect.Effect {
	var compounds []map[string]any
	switch value := value.(type) {
	case []map[string]any:
		compounds = value
	case []any:
		for _, entry := range value {
			if compound, ok := entry.(map[string]any); ok {
				compounds = append(compounds, compound)
			}
		}
	}
	decoded := make([]effect.Effect, 0, len(compounds))
	for _, compound := range compounds {
		typeValue, ok := effect.ByID(int(nbtInt64(compound["id"])))
		lasting, lastingOK := typeValue.(effect.LastingType)
		level := int(nbtInt64(compound["level"]))
		if !ok || !lastingOK || level <= 0 {
			continue
		}
		var value effect.Effect
		switch {
		case nbtBool(compound["ambient"]):
			value = effect.NewAmbient(lasting, level, max(time.Duration(nbtInt64(compound["duration"])), 0))
		case nbtBool(compound["infinite"]):
			value = effect.NewInfinite(lasting, level)
		default:
			value = effect.New(lasting, level, max(time.Duration(nbtInt64(compound["duration"])), 0))
		}
		if nbtBool(compound["hidden"]) {
			value = value.WithoutParticles()
		}
		decoded = append(decoded, value)
	}
	return decoded
}

func boolByte(value bool) uint8 {
	if value {
		return 1
	}
	return 0
}

func nbtBool(value any) bool { return nbtInt64(value) != 0 }

func nbtInt64(value any) int64 {
	switch value := value.(type) {
	case int64:
		return value
	case int32:
		return int64(value)
	case int16:
		return int64(value)
	case int8:
		return int64(value)
	case uint8:
		return int64(value)
	default:
		return 0
	}
}

func (t *foreignAdvancedEntityType) newState(instance native.EntityInstanceID) *foreignEntityState {
	definition := t.definition
	state := &foreignEntityState{
		typeDefinition: definition,
		services:       t.services,
		instance:       instance,
		health:         entity.NewHealthManager(definition.InitialHealth, definition.MaxHealth),
		effects:        entity.NewEffectManager(),
		speed:          definition.Speed,
		baseSpeed:      definition.Speed,
	}
	if physics := definition.Physics; physics != nil {
		state.movement = &entity.MovementComputer{
			Gravity: physics.Gravity, Drag: physics.Drag, DragBeforeGravity: physics.DragBeforeGravity,
		}
	}
	return state
}

type foreignTickingEntityType struct{ *foreignAdvancedEntityType }

func (t *foreignTickingEntityType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	data.Data.(*foreignEntityState).data = data
	opened := &foreignTickingEntity{Ent: entity.Open(tx, handle, data), tx: tx}
	if opened.state().invalidPersistence {
		_ = opened.Close()
	}
	return opened
}

type foreignLivingEntityType struct{ *foreignAdvancedEntityType }

func (t *foreignLivingEntityType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	data.Data.(*foreignEntityState).data = data
	opened := &foreignLivingEntity{foreignTickingEntity: foreignTickingEntity{Ent: entity.Open(tx, handle, data), tx: tx}}
	if opened.state().invalidPersistence {
		_ = opened.Close()
	}
	return opened
}

type foreignAdvancedEntityConfig struct {
	typeDefinition *foreignAdvancedEntityType
	instance       native.EntityInstanceID
}

func (c foreignAdvancedEntityConfig) Apply(data *world.EntityData) {
	data.Data = c.typeDefinition.newState(c.instance)
}

func foreignEntityConfigFor(entityType world.EntityType, instance native.EntityInstanceID) world.EntityConfig {
	switch entityType := entityType.(type) {
	case *foreignBaseEntityType:
		return foreignBaseEntityConfig{}
	case *foreignTickingEntityType:
		return foreignAdvancedEntityConfig{typeDefinition: entityType.foreignAdvancedEntityType, instance: instance}
	case *foreignLivingEntityType:
		return foreignAdvancedEntityConfig{typeDefinition: entityType.foreignAdvancedEntityType, instance: instance}
	default:
		panic("foreignEntityConfigFor called with an unowned entity type")
	}
}

func isForeignEntityType(entityType world.EntityType) bool {
	switch entityType.(type) {
	case *foreignBaseEntityType, *foreignTickingEntityType, *foreignLivingEntityType:
		return true
	default:
		return false
	}
}

func foreignAdvancedType(entityType world.EntityType) (*foreignAdvancedEntityType, bool) {
	switch entityType := entityType.(type) {
	case *foreignTickingEntityType:
		return entityType.foreignAdvancedEntityType, true
	case *foreignLivingEntityType:
		return entityType.foreignAdvancedEntityType, true
	default:
		return nil, false
	}
}

type foreignEntityState struct {
	typeDefinition     native.EntityTypeDefinition
	services           foreignEntityServices
	instance           native.EntityInstanceID
	health             *entity.HealthManager
	effects            *entity.EffectManager
	speed              float64
	baseSpeed          float64
	effectMutation     bool
	movement           *entity.MovementComputer
	data               *world.EntityData
	pluginState        []byte
	pluginStateVersion uint32
	hasPluginState     bool
	invalidPersistence bool
	destroyOnce        sync.Once
}

func (s *foreignEntityState) Tick(ent *entity.Ent, tx *world.Tx) *entity.Movement {
	if s.movement == nil {
		return nil
	}
	movement := s.movement.TickMovement(ent, ent.Position(), ent.Velocity(), ent.Rotation(), tx)
	s.data.Pos, s.data.Vel = movement.Position(), movement.Velocity()
	return movement
}

type foreignTickingEntity struct {
	*entity.Ent
	tx *world.Tx
}

func (e *foreignTickingEntity) state() *foreignEntityState {
	return e.Ent.Behaviour().(*foreignEntityState)
}

func (e *foreignTickingEntity) Tick(tx *world.Tx, current int64) {
	e.tickPlugin(tx, current)
	if e.H().Closed() {
		e.state().destroy()
		return
	}
	e.tickCommon(tx, current)
}

func (e *foreignTickingEntity) tickCommon(tx *world.Tx, current int64) {
	e.Ent.Tick(tx, current)
}

func (e *foreignTickingEntity) tickPlugin(tx *world.Tx, current int64) {
	state := e.state()
	if state.services.runtime == nil || state.services.players == nil || state.services.entities == nil || state.instance == 0 ||
		state.typeDefinition.CallbackFlags&native.EntityCallbackTick == 0 {
		return
	}
	entityID := state.services.entities.Register(e)
	invocation, end := state.services.players.BeginInvocation(tx)
	output, err := state.services.runtime.EntityTick(state.instance, native.EntityTickInput{
		Invocation: invocation, Entity: entityID, Current: current, Age: e.Age(),
	})
	end()
	if err == nil && output.Despawn {
		ref := world.NewEntityRef[world.Entity](e.H())
		ref.Do(func(_ *world.Tx, current world.Entity) { _ = current.Close() })
	}
}

func (e *foreignTickingEntity) Close() error {
	e.state().destroy()
	return e.Ent.Close()
}

func (s *foreignEntityState) destroy() {
	if s.services.runtime != nil && s.instance != 0 {
		s.destroyOnce.Do(func() { s.services.runtime.EntityDestroy(s.instance) })
	}
}

type foreignLivingEntity struct{ foreignTickingEntity }

func advancedEntityCleanup(current world.Entity) func() {
	switch current := current.(type) {
	case *foreignTickingEntity:
		return current.state().destroy
	case *foreignLivingEntity:
		return current.state().destroy
	default:
		return nil
	}
}

func (e *foreignLivingEntity) state() *foreignEntityState {
	return e.foreignTickingEntity.state()
}

func (e *foreignLivingEntity) Tick(tx *world.Tx, current int64) {
	e.state().effectMutation = true
	e.state().effects.Tick(e, tx)
	e.state().effectMutation = false
	e.tickPlugin(tx, current)
	if e.H().Closed() {
		e.state().destroy()
		return
	}
	e.tickCommon(tx, current)
}

func (e *foreignLivingEntity) Health() float64    { return e.state().health.Health() }
func (e *foreignLivingEntity) MaxHealth() float64 { return e.state().health.MaxHealth() }
func (e *foreignLivingEntity) SetMaxHealth(value float64) {
	e.state().health.SetMaxHealth(value)
}
func (e *foreignLivingEntity) Dead() bool { return e.Health() <= mgl64.Epsilon }

func (e *foreignLivingEntity) Hurt(damage float64, source world.DamageSource) (float64, bool) {
	if damage <= 0 || math.IsNaN(damage) || math.IsInf(damage, 0) || e.Dead() {
		return 0, false
	}
	state := e.state()
	if invocation, entityID, end, ok := e.beginCallback(); ok {
		defer end()
		nativeSource := host.NativeDamageSource(source, state.services.entities)
		output, err := state.services.runtime.EntityHurt(state.instance, native.EntityHurtInput{
			Invocation: invocation, Entity: entityID, Source: nativeSource,
			Health: e.Health(), MaxHealth: e.MaxHealth(), Damage: damage,
		})
		if err == nil {
			if output.Cancelled {
				return 0, false
			}
			if output.Damage >= 0 && !math.IsNaN(output.Damage) && !math.IsInf(output.Damage, 0) {
				damage = output.Damage
			}
		}
		if damage >= e.Health() {
			output, err := state.services.runtime.EntityDeath(state.instance, native.EntityDeathInput{
				Invocation: invocation, Entity: entityID, Source: nativeSource,
				Health: e.Health(), Damage: damage,
			})
			if err == nil && output.Cancelled {
				return 0, false
			}
		}
	}
	if e.H().Closed() || damage <= 0 {
		return 0, false
	}
	damage = min(damage, e.Health())
	e.state().health.AddHealth(-damage)
	e.viewAction(entity.HurtAction{})
	if e.Dead() {
		e.viewAction(entity.DeathAction{})
		ref := world.NewEntityRef[world.Entity](e.H())
		ref.Do(func(_ *world.Tx, current world.Entity) { _ = current.Close() })
	}
	return damage, true
}

func (e *foreignLivingEntity) viewAction(action world.EntityAction) {
	if e.tx == nil {
		return
	}
	for _, viewer := range e.tx.Viewers(e.Position()) {
		viewer.ViewEntityAction(e, action)
	}
}

func (e *foreignLivingEntity) Heal(health float64, source world.HealingSource) float64 {
	if health <= 0 || math.IsNaN(health) || math.IsInf(health, 0) || e.Dead() {
		return 0
	}
	state := e.state()
	if invocation, entityID, end, ok := e.beginCallback(); ok {
		output, err := state.services.runtime.EntityHeal(state.instance, native.EntityHealInput{
			Invocation: invocation, Entity: entityID, Source: host.NativeHealingSource(source),
			Health: e.Health(), MaxHealth: e.MaxHealth(), Amount: health,
		})
		end()
		if err == nil {
			if output.Cancelled {
				return 0
			}
			if output.Amount >= 0 && !math.IsNaN(output.Amount) && !math.IsInf(output.Amount, 0) {
				health = output.Amount
			}
		}
	}
	if e.H().Closed() || health <= 0 {
		return 0
	}
	before := e.Health()
	e.state().health.AddHealth(health)
	return e.Health() - before
}

func (e *foreignLivingEntity) beginCallback() (native.InvocationID, native.EntityID, func(), bool) {
	state := e.state()
	if state.services.runtime == nil || state.services.players == nil || state.services.entities == nil || state.instance == 0 || e.tx == nil {
		return 0, native.EntityID{}, func() {}, false
	}
	entityID := state.services.entities.Register(e)
	invocation, end := state.services.players.BeginInvocation(e.tx)
	if entityID.Generation == 0 || invocation == 0 {
		end()
		return 0, native.EntityID{}, func() {}, false
	}
	return invocation, entityID, end, true
}

func (e *foreignLivingEntity) KnockBack(source mgl64.Vec3, force, height float64) {
	direction := e.Position().Sub(source)
	direction[1] = 0
	if direction.LenSqr() == 0 {
		return
	}
	direction = direction.Normalize().Mul(force)
	direction[1] = height
	e.SetVelocity(e.Velocity().Add(direction))
}

func (e *foreignLivingEntity) AddEffect(value effect.Effect) {
	e.state().effectMutation = true
	e.state().effects.Add(value, e)
	e.state().effectMutation = false
}

func (e *foreignLivingEntity) RemoveEffect(value effect.Type) {
	e.state().effectMutation = true
	e.state().effects.Remove(value, e)
	e.state().effectMutation = false
}

func (e *foreignLivingEntity) Effects() []effect.Effect { return e.state().effects.Effects() }
func (e *foreignLivingEntity) Speed() float64           { return e.state().speed }
func (e *foreignLivingEntity) SetSpeed(value float64) {
	state := e.state()
	state.speed = value
	if !state.effectMutation {
		state.baseSpeed = value
	}
}
