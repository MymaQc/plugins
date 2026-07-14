package framework

import (
	"fmt"
	"sort"
	"sync"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type managedEntityRuntime interface {
	EntityAdoptLocal(uint64, uint64, uint64) (native.EntityInstanceID, error)
	EntityDecodeNBT(uint64, native.EntityCommonData, []byte) (native.EntityInstanceID, native.EntityCommonData, error)
	EntityEncodeNBT(native.EntityInstanceID, native.EntityCommonData) ([]byte, native.EntityCommonData, error)
	EntityOpen(native.EntityInstanceID, native.InvocationID, native.EntityHandleID, native.EntityCommonData) (native.EntityOpenID, uint32, native.EntityCommonData, error)
	EntityBBox(native.EntityOpenID, native.InvocationID, native.EntityCommonData) (native.BBox, native.EntityCommonData, error)
	EntityClose(native.EntityOpenID, native.InvocationID, native.EntityCommonData) (native.EntityCommonData, error)
	EntityH(native.EntityOpenID, native.InvocationID, native.EntityCommonData) (native.EntityHandleID, native.EntityCommonData, error)
	EntityPosition(native.EntityOpenID, native.InvocationID, native.EntityCommonData) (native.Vec3, native.EntityCommonData, error)
	EntityRotation(native.EntityOpenID, native.InvocationID, native.EntityCommonData) (native.Rotation, native.EntityCommonData, error)
	EntityTickExact(native.EntityOpenID, native.InvocationID, int64, native.EntityCommonData) (native.EntityCommonData, error)
	EntityReleaseOpen(native.EntityOpenID)
	EntityDestroy(native.EntityInstanceID)
}

type foreignEntityServices struct {
	runtime  managedEntityRuntime
	players  *host.Players
	entities *host.Entities
}

type managedEntityType struct {
	definition native.EntityTypeDefinition
	services   foreignEntityServices
}

func (t *managedEntityType) EncodeEntity() string        { return t.definition.SaveID }
func (t *managedEntityType) NetworkEncodeEntity() string { return t.definition.NetworkID }

func (t *managedEntityType) DecodeNBT(values map[string]any, data *world.EntityData) {
	encoded, ok := host.MarshalNBT(values)
	if !ok {
		panic("encode custom entity NBT")
	}
	instance, common, err := t.services.runtime.EntityDecodeNBT(t.definition.TypeKey, entityCommon(data), encoded)
	if err != nil || instance == 0 {
		panic(fmt.Sprintf("decode custom entity NBT: %v", err))
	}
	applyEntityCommon(data, common)
	data.Data = &managedEntityState{runtime: t.services.runtime, instance: instance}
}

func (t *managedEntityType) EncodeNBT(data *world.EntityData) map[string]any {
	state := managedState(data)
	encoded, common, err := state.runtime.EntityEncodeNBT(state.instance, entityCommon(data))
	if err != nil {
		panic(fmt.Sprintf("encode custom entity NBT: %v", err))
	}
	applyEntityCommon(data, common)
	values, ok := host.UnmarshalNBT(encoded)
	if !ok {
		panic("decode custom entity NBT result")
	}
	return values
}

func (t *managedEntityType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	state := managedState(data)
	stable, ok := t.services.entities.EnsureHandle(handle, state.destroy)
	if !ok {
		panic("register custom entity handle")
	}
	invocation, end := t.services.players.BeginInvocation(tx)
	opened, capabilities, common, err := state.runtime.EntityOpen(state.instance, invocation, stable, entityCommon(data))
	if err != nil || opened == 0 {
		end()
		panic(fmt.Sprintf("open custom entity: %v", err))
	}
	applyEntityCommon(data, common)
	tx.Defer(func(*world.Tx) {
		state.runtime.EntityReleaseOpen(opened)
		end()
	})
	view := &managedEntityView{
		tx: tx, handle: handle, data: data, state: state,
		services: t.services, invocation: invocation, opened: opened,
	}
	if capabilities&native.EntityCapabilityTicker != 0 {
		return &managedTickerEntityView{managedEntityView: view}
	}
	return view
}

func (t *managedEntityType) BBox(entity world.Entity) cube.BBox {
	view := managedView(entity)
	box, common, err := view.state.runtime.EntityBBox(view.opened, view.invocation, entityCommon(view.data))
	if err != nil || !finiteVec3(box.Min) || !finiteVec3(box.Max) ||
		box.Min.X > box.Max.X || box.Min.Y > box.Max.Y || box.Min.Z > box.Max.Z {
		panic(fmt.Sprintf("custom entity bounding box: %v", err))
	}
	applyEntityCommon(view.data, common)
	return cube.Box(box.Min.X, box.Min.Y, box.Min.Z, box.Max.X, box.Max.Y, box.Max.Z)
}

type managedEntityConfig struct {
	runtime  managedEntityRuntime
	instance native.EntityInstanceID
	common   *native.EntitySpawnOptions
	cleanup  func()
}

func (c managedEntityConfig) Apply(data *world.EntityData) {
	if c.common != nil {
		data.FireDuration, data.Age = c.common.FireDuration, c.common.Age
	}
	data.Data = &managedEntityState{runtime: c.runtime, instance: c.instance, cleanup: c.cleanup}
}

type managedEntityState struct {
	runtime  managedEntityRuntime
	instance native.EntityInstanceID
	cleanup  func()
	once     sync.Once
}

func (s *managedEntityState) destroy() {
	if s == nil {
		return
	}
	s.once.Do(func() {
		if s.cleanup != nil {
			s.cleanup()
		} else if s.runtime != nil && s.instance != 0 {
			s.runtime.EntityDestroy(s.instance)
		}
	})
}

type managedEntityView struct {
	tx         *world.Tx
	handle     *world.EntityHandle
	data       *world.EntityData
	state      *managedEntityState
	services   foreignEntityServices
	invocation native.InvocationID
	opened     native.EntityOpenID
}

func (e *managedEntityView) Close() error {
	common, err := e.state.runtime.EntityClose(e.opened, e.invocation, entityCommon(e.data))
	if err == nil {
		applyEntityCommon(e.data, common)
	}
	return err
}

func (e *managedEntityView) H() *world.EntityHandle {
	handle, common, err := e.state.runtime.EntityH(e.opened, e.invocation, entityCommon(e.data))
	if err != nil {
		panic(fmt.Sprintf("custom entity handle: %v", err))
	}
	applyEntityCommon(e.data, common)
	resolved, ok := e.services.entities.HandleByID(handle)
	if !ok {
		panic("custom entity returned unknown handle")
	}
	return resolved
}

func (e *managedEntityView) Position() mgl64.Vec3 {
	position, common, err := e.state.runtime.EntityPosition(e.opened, e.invocation, entityCommon(e.data))
	if err != nil || !finiteVec3(position) {
		panic(fmt.Sprintf("custom entity position: %v", err))
	}
	applyEntityCommon(e.data, common)
	return vec3(position)
}

func (e *managedEntityView) Rotation() cube.Rotation {
	value, common, err := e.state.runtime.EntityRotation(e.opened, e.invocation, entityCommon(e.data))
	if err != nil {
		panic(fmt.Sprintf("custom entity rotation: %v", err))
	}
	applyEntityCommon(e.data, common)
	return rotation(value)
}

type managedTickerEntityView struct{ *managedEntityView }

func (e *managedTickerEntityView) Tick(_ *world.Tx, current int64) {
	common, err := e.state.runtime.EntityTickExact(e.opened, e.invocation, current, entityCommon(e.data))
	if err != nil {
		panic(fmt.Sprintf("tick custom entity: %v", err))
	}
	applyEntityCommon(e.data, common)
}

func managedState(data *world.EntityData) *managedEntityState {
	state, ok := data.Data.(*managedEntityState)
	if !ok || state.runtime == nil || state.instance == 0 {
		panic("invalid custom entity state")
	}
	return state
}

func managedView(entity world.Entity) *managedEntityView {
	switch entity := entity.(type) {
	case *managedEntityView:
		return entity
	case *managedTickerEntityView:
		return entity.managedEntityView
	default:
		panic("custom EntityType.BBox called with foreign entity")
	}
}

func entityCommon(data *world.EntityData) native.EntityCommonData {
	return native.EntityCommonData{
		Position: nativeVec3(data.Pos), Velocity: nativeVec3(data.Vel),
		Rotation: native.Rotation{Yaw: data.Rot.Yaw(), Pitch: data.Rot.Pitch()},
		Name:     data.Name, FireDuration: data.FireDuration, Age: data.Age,
	}
}

func applyEntityCommon(data *world.EntityData, value native.EntityCommonData) {
	data.Pos, data.Vel = vec3(value.Position), vec3(value.Velocity)
	data.Rot = rotation(value.Rotation)
	data.Name, data.FireDuration, data.Age = value.Name, value.FireDuration, value.Age
}

func buildEntityRegistry(base world.EntityRegistry, definitions []native.EntityTypeDefinition, configured ...foreignEntityServices) (world.EntityRegistry, error) {
	var services foreignEntityServices
	if len(configured) != 0 {
		services = configured[0]
	}
	if services.entities == nil && services.players != nil {
		services.entities = services.players.EntityRegistry()
	}
	if len(base.Types()) == 0 {
		base = entity.DefaultRegistry
	}
	types := base.Types()
	seen := make(map[string]struct{}, len(types)+len(definitions))
	for _, entityType := range types {
		seen[entityType.EncodeEntity()] = struct{}{}
	}
	definitions = append([]native.EntityTypeDefinition(nil), definitions...)
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].SaveID < definitions[j].SaveID })
	for _, definition := range definitions {
		if !worldIDPattern.MatchString(definition.SaveID) || !worldIDPattern.MatchString(definition.NetworkID) || definition.TypeKey == 0 {
			return world.EntityRegistry{}, fmt.Errorf("invalid custom entity type %q", definition.SaveID)
		}
		if _, duplicate := seen[definition.SaveID]; duplicate {
			return world.EntityRegistry{}, fmt.Errorf("duplicate entity type %q", definition.SaveID)
		}
		seen[definition.SaveID] = struct{}{}
		types = append(types, &managedEntityType{definition: definition, services: services})
	}
	return base.Config().New(types), nil
}

func managedEntityTypeInfo(entityType world.EntityType) (*managedEntityType, bool) {
	managed, ok := entityType.(*managedEntityType)
	return managed, ok
}

func managedEntityConfigFor(entityType *managedEntityType, instance native.EntityInstanceID, common *native.EntitySpawnOptions, cleanup func()) world.EntityConfig {
	return managedEntityConfig{runtime: entityType.services.runtime, instance: instance, common: common, cleanup: cleanup}
}
