package host

import (
	"math"
	"slices"
	"strings"
	"sync"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/title"
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

func (p *Players) ResolveEntityID(id native.EntityID) (*player.Player, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for connected, candidate := range p.entries {
		if candidate.UUID == id.UUID && candidate.Generation == id.Generation {
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
	return sendPlayerText(connected, kind, message)
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
	return setPlayerState(connected, kind, value)
}

func (p *Players) PlayerState(id native.PlayerID, kind native.PlayerStateKind) (native.PlayerStateValue, bool) {
	connected, ok := p.ResolveID(id)
	if !ok {
		return native.PlayerStateValue{}, false
	}
	return readPlayerState(connected, kind)
}

func (p *Players) ChangePlayerEffect(id native.PlayerID, operation native.PlayerEffectOperation, value native.PlayerEffect) bool {
	connected, ok := p.ResolveID(id)
	if !ok {
		return false
	}
	effectType, ok := effect.ByID(int(value.Type))
	if !ok {
		return false
	}
	if operation == native.PlayerEffectRemove {
		connected.RemoveEffect(effectType)
		return true
	}
	if operation != native.PlayerEffectAdd || value.Level < 0 || value.Duration < 0 {
		return false
	}
	var applied effect.Effect
	if lasting, ok := effectType.(effect.LastingType); ok {
		switch {
		case value.Infinite:
			applied = effect.NewInfinite(lasting, int(value.Level))
		case value.Ambient:
			applied = effect.NewAmbient(lasting, int(value.Level), value.Duration)
		default:
			applied = effect.New(lasting, int(value.Level), value.Duration)
		}
	} else {
		applied = effect.NewInstant(effectType, int(value.Level))
	}
	if value.ParticlesHidden {
		applied = applied.WithoutParticles()
	}
	connected.AddEffect(applied)
	return true
}

func (p *Players) SetPlayerEntityVisible(viewerID native.PlayerID, entityID native.EntityID, visible bool) bool {
	viewer, ok := p.ResolveID(viewerID)
	if !ok {
		return false
	}
	entity, ok := p.ResolveEntityID(entityID)
	if !ok {
		return false
	}
	if visible {
		viewer.ShowEntity(entity)
	} else {
		viewer.HideEntity(entity)
	}
	return true
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
