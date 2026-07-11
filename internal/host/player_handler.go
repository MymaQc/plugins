// Package host adapts Dragonfly lifecycle and handlers to the native runtime.
package host

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type playerRuntime interface {
	Subscriptions() uint64
	HandlePlayerMove(native.PlayerMoveInput, bool) (bool, error)
	HandlePlayerChat(native.PlayerChatInput, bool) (native.PlayerChatOutput, error)
	HandlePlayerJoin(native.PlayerJoinInput, bool) (bool, error)
	HandlePlayerQuit(native.PlayerQuitInput) error
	HandlePlayerHurt(native.PlayerHurtInput, bool) (native.PlayerHurtOutput, error)
	HandlePlayerHeal(native.PlayerHealInput, bool) (native.PlayerHealOutput, error)
}

// PlayerHandler forwards supported Dragonfly player events into the native runtime.
// Unsupported events keep Dragonfly's default behavior through NopHandler.
type PlayerHandler struct {
	player.NopHandler
	runtime playerRuntime
	log     *slog.Logger
	players *Players
}

var _ player.Handler = (*PlayerHandler)(nil)

func NewPlayerHandler(runtime playerRuntime, log *slog.Logger, players *Players) *PlayerHandler {
	if log == nil {
		log = slog.Default()
	}
	return &PlayerHandler{runtime: runtime, log: log, players: players}
}

func (h *PlayerHandler) HandleMove(ctx *player.Context, newPosition mgl64.Vec3, newRotation cube.Rotation) {
	if h.runtime.Subscriptions()&native.PlayerMoveSubscription == 0 {
		return
	}
	p := ctx.Val()
	oldPosition := p.Position()
	cancelled, err := h.runtime.HandlePlayerMove(native.PlayerMoveInput{
		Player:      h.playerID(p),
		OldPosition: native.Vec3{X: oldPosition.X(), Y: oldPosition.Y(), Z: oldPosition.Z()},
		NewPosition: native.Vec3{X: newPosition.X(), Y: newPosition.Y(), Z: newPosition.Z()},
		Rotation:    native.Rotation{Yaw: float32(newRotation.Yaw()), Pitch: float32(newRotation.Pitch())},
	}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin movement handler failed", "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleChat(ctx *player.Context, message *string) {
	if h.runtime.Subscriptions()&native.PlayerChatSubscription == 0 {
		return
	}
	p := ctx.Val()
	output, err := h.runtime.HandlePlayerChat(native.PlayerChatInput{
		Player:  h.playerID(p),
		Message: *message,
	}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin chat handler failed", "player", p.Name(), "error", err)
		return
	}
	if output.Replacement != nil {
		*message = *output.Replacement
	}
	if output.Cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleHurt(ctx *player.Context, damage *float64, immune bool, attackImmunity *time.Duration, source world.DamageSource) {
	if h.runtime.Subscriptions()&native.PlayerHurtSubscription == 0 {
		return
	}
	p := ctx.Val()
	output, err := h.runtime.HandlePlayerHurt(native.PlayerHurtInput{
		Player:         h.playerID(p),
		Damage:         *damage,
		Immune:         immune,
		AttackImmunity: *attackImmunity,
		Source:         fmt.Sprintf("%T", source),
	}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin hurt handler failed", "player", p.Name(), "error", err)
		return
	}
	*damage = output.Damage
	*attackImmunity = output.AttackImmunity
	if output.Cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleHeal(ctx *player.Context, health *float64, source world.HealingSource) {
	if h.runtime.Subscriptions()&native.PlayerHealSubscription == 0 {
		return
	}
	p := ctx.Val()
	output, err := h.runtime.HandlePlayerHeal(native.PlayerHealInput{
		Player: h.playerID(p),
		Health: *health,
		Source: fmt.Sprintf("%T", source),
	}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin heal handler failed", "player", p.Name(), "error", err)
		return
	}
	*health = output.Health
	if output.Cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) playerID(p *player.Player) native.PlayerID {
	id, _ := h.players.ID(p)
	return id
}

func (h *PlayerHandler) Join(p *player.Player) bool {
	if h.runtime.Subscriptions()&native.PlayerJoinSubscription == 0 {
		return false
	}
	cancelled, err := h.runtime.HandlePlayerJoin(native.PlayerJoinInput{
		Player: h.playerID(p),
		Name:   p.Name(),
	}, false)
	if err != nil {
		h.log.Error("native plugin join handler failed", "player", p.Name(), "error", err)
		return false
	}
	return cancelled
}

func (h *PlayerHandler) HandleQuit(p *player.Player) {
	if h.runtime.Subscriptions()&native.PlayerQuitSubscription != 0 {
		if err := h.runtime.HandlePlayerQuit(native.PlayerQuitInput{
			Player: h.playerID(p),
			Name:   p.Name(),
		}); err != nil {
			h.log.Error("native plugin quit handler failed", "player", p.Name(), "error", err)
		}
	}
	h.players.Unregister(p)
}
