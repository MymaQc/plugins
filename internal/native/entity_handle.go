package native

// EntityHandleID identifies one Dragonfly EntityHandle independently from its
// current world-bound EntityID. It remains stable while the handle is detached
// and added to another world.
type EntityHandleID struct {
	Value      uint64
	Generation uint64
}

// Valid reports whether id identifies an entity handle.
func (id EntityHandleID) Valid() bool {
	return id.Value != 0 && id.Generation != 0
}
