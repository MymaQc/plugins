package framework

import (
	"fmt"
	"sort"
	"sync"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type foreignBaseEntityType struct {
	saveID, networkID string
	bbox              cube.BBox
}

func (t *foreignBaseEntityType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &foreignBaseEntity{tx: tx, handle: handle, data: data}
}

func (t *foreignBaseEntityType) EncodeEntity() string                      { return t.saveID }
func (t *foreignBaseEntityType) NetworkEncodeEntity() string               { return t.networkID }
func (t *foreignBaseEntityType) BBox(world.Entity) cube.BBox               { return t.bbox }
func (*foreignBaseEntityType) DecodeNBT(map[string]any, *world.EntityData) {}
func (*foreignBaseEntityType) EncodeNBT(*world.EntityData) map[string]any  { return nil }

type foreignBaseEntityConfig struct{}

func (foreignBaseEntityConfig) Apply(*world.EntityData) {}

type foreignBaseEntity struct {
	tx     *world.Tx
	handle *world.EntityHandle
	data   *world.EntityData
	once   sync.Once
}

func (e *foreignBaseEntity) H() *world.EntityHandle  { return e.handle }
func (e *foreignBaseEntity) Position() mgl64.Vec3    { return e.data.Pos }
func (e *foreignBaseEntity) Rotation() cube.Rotation { return e.data.Rot }
func (e *foreignBaseEntity) NameTag() string         { return e.data.Name }

func (e *foreignBaseEntity) SetNameTag(name string) {
	e.data.Name = name
	for _, viewer := range e.tx.Viewers(e.Position()) {
		viewer.ViewEntityState(e)
	}
}

func (e *foreignBaseEntity) Teleport(position mgl64.Vec3) {
	viewers := e.tx.Viewers(e.Position())
	e.data.Pos = position
	for _, viewer := range viewers {
		viewer.ViewEntityTeleport(e, position)
	}
}

func (e *foreignBaseEntity) Close() error {
	e.once.Do(func() {
		e.tx.RemoveEntity(e)
		_ = e.handle.Close()
	})
	return nil
}

func buildEntityRegistry(base world.EntityRegistry, definitions []native.EntityTypeDefinition) (world.EntityRegistry, error) {
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
		if !worldIDPattern.MatchString(definition.SaveID) {
			return world.EntityRegistry{}, fmt.Errorf("invalid custom entity save identifier %q", definition.SaveID)
		}
		if !worldIDPattern.MatchString(definition.NetworkID) {
			return world.EntityRegistry{}, fmt.Errorf("invalid custom entity network identifier %q", definition.NetworkID)
		}
		if !validEntityBounds(definition.Min, definition.Max) {
			return world.EntityRegistry{}, fmt.Errorf("invalid bounds for custom entity %q", definition.SaveID)
		}
		if _, duplicate := seen[definition.SaveID]; duplicate {
			return world.EntityRegistry{}, fmt.Errorf("duplicate entity type %q", definition.SaveID)
		}
		seen[definition.SaveID] = struct{}{}
		types = append(types, &foreignBaseEntityType{
			saveID: definition.SaveID, networkID: definition.NetworkID,
			bbox: cube.Box(definition.Min.X, definition.Min.Y, definition.Min.Z, definition.Max.X, definition.Max.Y, definition.Max.Z),
		})
	}
	return base.Config().New(types), nil
}

func validEntityBounds(minimum, maximum native.Vec3) bool {
	return finiteVec3(minimum) && finiteVec3(maximum) &&
		minimum.X <= maximum.X && minimum.Y <= maximum.Y && minimum.Z <= maximum.Z
}
