package host

import (
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

// WorldHandler is installed on every framework-owned world before the server starts.
// Event callbacks will be added as their generated ABI definitions land.
type WorldHandler struct {
	world.NopHandler
	entities *Entities
	players  *Players
	world    native.WorldID
}

var _ world.Handler = (*WorldHandler)(nil)

func NewWorldHandler(entities *Entities, players *Players, worldID native.WorldID) *WorldHandler {
	return &WorldHandler{entities: entities, players: players, world: worldID}
}

func (h *WorldHandler) HandleEntitySpawn(_ *world.Tx, entity world.Entity) {
	if h.entities != nil {
		h.entities.Register(entity)
	}
}

func (h *WorldHandler) HandleEntityDespawn(tx *world.Tx, entity world.Entity) {
	if h.entities == nil || entity == nil {
		return
	}
	handle := entity.H()
	if connected, ok := entity.(*player.Player); ok {
		h.players.recordWorldDeparture(connected, h.world)
	}
	h.entities.deactivateHandle(handle)
	tx.Defer(func(*world.Tx) {
		if handle.Closed() {
			h.entities.unregisterHandle(handle)
			h.players.forgetWorldDeparture(handle)
		}
	})
}
