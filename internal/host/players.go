package host

import (
	"math"
	"slices"
	"strings"
	"sync"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
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

const (
	maxSkinDimension  = 4096
	maxSkinAnimations = 64
	maxSkinDataBytes  = 64 << 20
	maxSkinIDBytes    = 4096
)

func (p *Players) PlayerSkin(id native.PlayerID) (native.PlayerSkin, bool) {
	connected, ok := p.ResolveID(id)
	if !ok {
		return native.PlayerSkin{}, false
	}
	return playerSkinToNative(connected.Skin())
}

func (p *Players) SetPlayerSkin(id native.PlayerID, value native.PlayerSkin) bool {
	connected, ok := p.ResolveID(id)
	if !ok {
		return false
	}
	converted, ok := playerSkinFromNative(value)
	if !ok {
		return false
	}
	connected.SetSkin(converted)
	return true
}

func playerSkinToNative(value skin.Skin) (native.PlayerSkin, bool) {
	bounds := value.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	capeBounds := value.Cape.Bounds()
	if !validSkinDimensions(width, height, value.Pix, true) ||
		!validSkinDimensions(capeBounds.Dx(), capeBounds.Dy(), value.Cape.Pix, true) ||
		len(value.Animations) > maxSkinAnimations ||
		!validSkinStrings(value.PlayFabID, value.FullID, value.ModelConfig.Default, value.ModelConfig.AnimatedFace) {
		return native.PlayerSkin{}, false
	}
	total := uint64(len(value.Pix)) + uint64(len(value.Model)) + uint64(len(value.Cape.Pix)) +
		uint64(len(value.PlayFabID)) + uint64(len(value.FullID)) +
		uint64(len(value.ModelConfig.Default)) + uint64(len(value.ModelConfig.AnimatedFace))
	for _, animation := range value.Animations {
		animationBounds := animation.Bounds()
		if !validSkinDimensions(animationBounds.Dx(), animationBounds.Dy(), animation.Pix, false) ||
			animation.FrameCount <= 0 ||
			animation.Type() < skin.AnimationHead || animation.Type() > skin.AnimationBody128x128 {
			return native.PlayerSkin{}, false
		}
		total += uint64(len(animation.Pix))
		if total > maxSkinDataBytes {
			return native.PlayerSkin{}, false
		}
	}
	if total > maxSkinDataBytes {
		return native.PlayerSkin{}, false
	}
	animations := make([]native.SkinAnimation, len(value.Animations))
	for i, animation := range value.Animations {
		animationBounds := animation.Bounds()
		animations[i] = native.SkinAnimation{
			Width:      uint32(animationBounds.Dx()),
			Height:     uint32(animationBounds.Dy()),
			Type:       uint32(animation.Type()),
			FrameCount: int64(animation.FrameCount),
			Expression: int64(animation.AnimationExpression),
			Pixels:     slices.Clone(animation.Pix),
		}
	}
	return native.PlayerSkin{
		Width:             uint32(width),
		Height:            uint32(height),
		Persona:           value.Persona,
		PlayFabID:         value.PlayFabID,
		FullID:            value.FullID,
		Pixels:            slices.Clone(value.Pix),
		ModelDefault:      value.ModelConfig.Default,
		ModelAnimatedFace: value.ModelConfig.AnimatedFace,
		Model:             slices.Clone(value.Model),
		CapeWidth:         uint32(capeBounds.Dx()),
		CapeHeight:        uint32(capeBounds.Dy()),
		CapePixels:        slices.Clone(value.Cape.Pix),
		Animations:        animations,
	}, true
}

func playerSkinFromNative(value native.PlayerSkin) (skin.Skin, bool) {
	if !validNativeSkinDimensions(value.Width, value.Height, value.Pixels, true) ||
		!validNativeSkinDimensions(value.CapeWidth, value.CapeHeight, value.CapePixels, true) ||
		len(value.Animations) > maxSkinAnimations ||
		!validSkinStrings(value.PlayFabID, value.FullID, value.ModelDefault, value.ModelAnimatedFace) {
		return skin.Skin{}, false
	}
	total := uint64(len(value.Pixels)) + uint64(len(value.Model)) + uint64(len(value.CapePixels)) +
		uint64(len(value.PlayFabID)) + uint64(len(value.FullID)) +
		uint64(len(value.ModelDefault)) + uint64(len(value.ModelAnimatedFace))
	for _, animation := range value.Animations {
		if !validNativeSkinDimensions(animation.Width, animation.Height, animation.Pixels, false) ||
			animation.Type > uint32(skin.AnimationBody128x128) || animation.FrameCount <= 0 ||
			animation.FrameCount > int64(math.MaxInt) ||
			animation.Expression < int64(math.MinInt) || animation.Expression > int64(math.MaxInt) {
			return skin.Skin{}, false
		}
		total += uint64(len(animation.Pixels))
		if total > maxSkinDataBytes {
			return skin.Skin{}, false
		}
	}
	if total > maxSkinDataBytes {
		return skin.Skin{}, false
	}
	converted := skin.New(int(value.Width), int(value.Height))
	converted.Persona = value.Persona
	converted.PlayFabID = value.PlayFabID
	converted.FullID = value.FullID
	copy(converted.Pix, value.Pixels)
	converted.ModelConfig = skin.ModelConfig{Default: value.ModelDefault, AnimatedFace: value.ModelAnimatedFace}
	converted.Model = slices.Clone(value.Model)
	converted.Cape = skin.NewCape(int(value.CapeWidth), int(value.CapeHeight))
	copy(converted.Cape.Pix, value.CapePixels)
	converted.Animations = make([]skin.Animation, len(value.Animations))
	for i, animation := range value.Animations {
		convertedAnimation := skin.NewAnimation(int(animation.Width), int(animation.Height), int(animation.Expression), skin.AnimationType(animation.Type))
		copy(convertedAnimation.Pix, animation.Pixels)
		convertedAnimation.FrameCount = int(animation.FrameCount)
		converted.Animations[i] = convertedAnimation
	}
	return converted, true
}

func validSkinStrings(values ...string) bool {
	for _, value := range values {
		if len(value) > maxSkinIDBytes {
			return false
		}
	}
	return true
}

func validNativeSkinDimensions(width, height uint32, pixels []byte, empty bool) bool {
	if width > maxSkinDimension || height > maxSkinDimension {
		return false
	}
	return validSkinDimensions(int(width), int(height), pixels, empty)
}

func validSkinDimensions(width, height int, pixels []byte, empty bool) bool {
	if width < 0 || height < 0 || width > maxSkinDimension || height > maxSkinDimension {
		return false
	}
	if width == 0 || height == 0 {
		return empty && width == 0 && height == 0 && len(pixels) == 0
	}
	return uint64(width)*uint64(height)*4 == uint64(len(pixels))
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
