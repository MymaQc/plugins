package host

import (
	"math"
	"slices"
	"strings"
	"sync"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/title"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
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

func (p *Players) CommandSnapshots() []native.CommandPlayer {
	p.mu.RLock()
	snapshots := make([]native.CommandPlayer, 0, len(p.entries))
	for connected, id := range p.entries {
		snapshots = append(snapshots, native.CommandPlayer{
			Player:              id,
			Name:                connected.Name(),
			LatencyMilliseconds: uint64(connected.Latency().Milliseconds()),
		})
	}
	p.mu.RUnlock()
	slices.SortFunc(snapshots, func(left, right native.CommandPlayer) int {
		return strings.Compare(left.Name, right.Name)
	})
	return snapshots
}

func (p *Players) ResolveName(name string) (native.PlayerID, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for connected, id := range p.entries {
		if strings.EqualFold(connected.Name(), name) {
			return id, true
		}
	}
	return native.PlayerID{}, false
}

func (p *Players) ResolveUUID(uuid [16]byte) (native.PlayerID, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, id := range p.entries {
		if id.UUID == uuid {
			return id, true
		}
	}
	return native.PlayerID{}, false
}

func (p *Players) ResolveID(id native.PlayerID) (*player.Player, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for connected, candidate := range p.entries {
		if candidate == id {
			return connected, true
		}
	}
	return nil, false
}

func (p *Players) SendPlayerText(id native.PlayerID, kind native.PlayerTextKind, message string) bool {
	connected, ok := p.ResolveID(id)
	if !ok {
		return false
	}
	switch kind {
	case native.PlayerTextMessage:
		connected.Message(message)
	case native.PlayerTextTip:
		connected.SendTip(message)
	case native.PlayerTextPopup:
		connected.SendPopup(message)
	case native.PlayerTextJukeboxPopup:
		connected.SendJukeboxPopup(message)
	default:
		return false
	}
	return true
}

func (p *Players) SendPlayerTitle(id native.PlayerID, value native.PlayerTitle) bool {
	connected, ok := p.ResolveID(id)
	if !ok {
		return false
	}
	t := title.New(value.Text).
		WithSubtitle(value.Subtitle).
		WithActionText(value.ActionText).
		WithFadeInDuration(value.FadeIn).
		WithDuration(value.Duration).
		WithFadeOutDuration(value.FadeOut)
	connected.SendTitle(t)
	return true
}

func (p *Players) TransformPlayer(id native.PlayerID, kind native.PlayerTransformKind, vector native.Vec3, yaw, pitch float64) bool {
	connected, ok := p.ResolveID(id)
	if !ok || !finite(vector.X, vector.Y, vector.Z, yaw, pitch) {
		return false
	}
	v := mgl64.Vec3{vector.X, vector.Y, vector.Z}
	switch kind {
	case native.PlayerTransformTeleport:
		connected.Teleport(v)
	case native.PlayerTransformMove:
		connected.Move(v, yaw, pitch)
	case native.PlayerTransformVelocity:
		connected.SetVelocity(v)
	default:
		return false
	}
	return true
}

func (p *Players) PlayerRotation(id native.PlayerID) (native.Rotation, bool) {
	connected, ok := p.ResolveID(id)
	if !ok {
		return native.Rotation{}, false
	}
	rotation := connected.Rotation()
	return native.Rotation{Yaw: rotation.Yaw(), Pitch: rotation.Pitch()}, true
}

func (p *Players) SetPlayerState(id native.PlayerID, kind native.PlayerStateKind, value native.PlayerStateValue) bool {
	connected, ok := p.ResolveID(id)
	if !ok {
		return false
	}
	switch kind {
	case native.PlayerStateGameMode:
		mode, found := world.GameModeByID(int(value.Integer))
		if !found {
			return false
		}
		connected.SetGameMode(mode)
	case native.PlayerStateHeal:
		if !finite(value.Number) || value.Number < 0 {
			return false
		}
		connected.Heal(value.Number, pluginHealingSource{})
	case native.PlayerStateHurt:
		if !finite(value.Number) || value.Number < 0 {
			return false
		}
		connected.Hurt(value.Number, pluginDamageSource{})
	case native.PlayerStateFood:
		connected.SetFood(int(value.Integer))
	case native.PlayerStateMaxHealth:
		if !finite(value.Number) || value.Number <= 0 {
			return false
		}
		connected.SetMaxHealth(value.Number)
	default:
		return false
	}
	return true
}

func (p *Players) PlayerState(id native.PlayerID, kind native.PlayerStateKind) (native.PlayerStateValue, bool) {
	connected, ok := p.ResolveID(id)
	if !ok {
		return native.PlayerStateValue{}, false
	}
	switch kind {
	case native.PlayerStateGameMode:
		mode, found := world.GameModeID(connected.GameMode())
		return native.PlayerStateValue{Integer: int64(mode)}, found
	case native.PlayerStateFood:
		return native.PlayerStateValue{Integer: int64(connected.Food())}, true
	case native.PlayerStateMaxHealth:
		return native.PlayerStateValue{Number: connected.MaxHealth()}, true
	case native.PlayerStateHealth:
		return native.PlayerStateValue{Number: connected.Health()}, true
	default:
		return native.PlayerStateValue{}, false
	}
}

type pluginHealingSource struct{}

func (pluginHealingSource) HealingSource() {}

type pluginDamageSource struct{}

func (pluginDamageSource) ReducedByArmour() bool     { return true }
func (pluginDamageSource) ReducedByResistance() bool { return true }
func (pluginDamageSource) Fire() bool                { return false }
func (pluginDamageSource) IgnoreTotem() bool         { return false }

func finite(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}
