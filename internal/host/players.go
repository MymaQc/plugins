package host

import (
	"slices"
	"sync"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/player"
)

// Players owns stable native IDs for the lifetime of connected Dragonfly players.
type Players struct {
	mu      sync.RWMutex
	entries map[*player.Player]native.PlayerID
}

func NewPlayers() *Players {
	return &Players{entries: map[*player.Player]native.PlayerID{}}
}

func (p *Players) Register(player *player.Player, generation uint64) native.PlayerID {
	id := native.PlayerID{Generation: generation}
	uuid := player.UUID()
	copy(id.UUID[:], uuid[:])
	p.mu.Lock()
	p.entries[player] = id
	p.mu.Unlock()
	return id
}

func (p *Players) Unregister(player *player.Player) {
	p.mu.Lock()
	delete(p.entries, player)
	p.mu.Unlock()
}

func (p *Players) ID(player *player.Player) (native.PlayerID, bool) {
	p.mu.RLock()
	id, ok := p.entries[player]
	p.mu.RUnlock()
	return id, ok
}

func (p *Players) Names() []string {
	p.mu.RLock()
	names := make([]string, 0, len(p.entries))
	for connected := range p.entries {
		names = append(names, connected.Name())
	}
	p.mu.RUnlock()
	slices.Sort(names)
	return names
}
