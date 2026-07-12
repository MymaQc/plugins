// Package host adapts Dragonfly lifecycle and handlers to the native runtime.
package host

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
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
	HandlePlayerBlockBreak(native.PlayerBlockBreakInput, bool) (native.PlayerBlockBreakOutput, error)
	HandlePlayerBlockPlace(native.PlayerBlockPlaceInput, bool) (bool, error)
	HandlePlayerFoodLoss(native.PlayerFoodLossInput, bool) (native.PlayerFoodLossOutput, error)
	HandlePlayerDeath(native.PlayerDeathInput, bool) (bool, error)
	HandlePlayerStartBreak(native.PlayerPositionInput, bool) (bool, error)
	HandlePlayerFireExtinguish(native.PlayerPositionInput, bool) (bool, error)
	HandlePlayerToggleSprint(native.PlayerToggleInput, bool) (bool, error)
	HandlePlayerToggleSneak(native.PlayerToggleInput, bool) (bool, error)
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

func (h *PlayerHandler) HandleBlockBreak(ctx *player.Context, position cube.Pos, _ *[]item.Stack, experience *int) {
	if h.runtime.Subscriptions()&native.PlayerBlockBreakSubscription == 0 {
		return
	}
	p := ctx.Val()
	output, err := h.runtime.HandlePlayerBlockBreak(native.PlayerBlockBreakInput{
		Player:     h.playerID(p),
		Position:   nativeBlockPos(position),
		Block:      blockName(p.Tx().Block(position)),
		Experience: int32(*experience),
	}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin block-break handler failed", "player", p.Name(), "error", err)
		return
	}
	*experience = int(output.Experience)
	if output.Cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleBlockPlace(ctx *player.Context, position cube.Pos, block world.Block) {
	if h.runtime.Subscriptions()&native.PlayerBlockPlaceSubscription == 0 {
		return
	}
	p := ctx.Val()
	cancelled, err := h.runtime.HandlePlayerBlockPlace(native.PlayerBlockPlaceInput{
		Player:   h.playerID(p),
		Position: nativeBlockPos(position),
		Block:    blockName(block),
	}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin block-place handler failed", "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleFoodLoss(ctx *player.Context, from int, to *int) {
	if h.runtime.Subscriptions()&native.PlayerFoodLossSubscription == 0 {
		return
	}
	p := ctx.Val()
	output, err := h.runtime.HandlePlayerFoodLoss(native.PlayerFoodLossInput{
		Player: h.playerID(p),
		From:   int32(from),
		To:     int32(*to),
	}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin food-loss handler failed", "player", p.Name(), "error", err)
		return
	}
	*to = int(output.To)
	if output.Cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleDeath(p *player.Player, source world.DamageSource, keepInventory *bool) {
	if h.runtime.Subscriptions()&native.PlayerDeathSubscription == 0 {
		return
	}
	keep, err := h.runtime.HandlePlayerDeath(native.PlayerDeathInput{
		Player: h.playerID(p),
		Source: fmt.Sprintf("%T", source),
	}, *keepInventory)
	if err != nil {
		h.log.Error("native plugin death handler failed", "player", p.Name(), "error", err)
		return
	}
	*keepInventory = keep
}

func (h *PlayerHandler) HandleStartBreak(ctx *player.Context, position cube.Pos) {
	h.handlePositionEvent(ctx, position, native.PlayerStartBreakSubscription, h.runtime.HandlePlayerStartBreak, "start-break")
}

func (h *PlayerHandler) HandleFireExtinguish(ctx *player.Context, position cube.Pos) {
	h.handlePositionEvent(ctx, position, native.PlayerFireExtinguishSubscription, h.runtime.HandlePlayerFireExtinguish, "fire-extinguish")
}

func (h *PlayerHandler) HandleToggleSprint(ctx *player.Context, after bool) {
	h.handleToggleEvent(ctx, after, native.PlayerToggleSprintSubscription, h.runtime.HandlePlayerToggleSprint, "toggle-sprint")
}
func (h *PlayerHandler) HandleToggleSneak(ctx *player.Context, after bool) {
	h.handleToggleEvent(ctx, after, native.PlayerToggleSneakSubscription, h.runtime.HandlePlayerToggleSneak, "toggle-sneak")
}
func (h *PlayerHandler) handleToggleEvent(ctx *player.Context, after bool, subscription uint64, handle func(native.PlayerToggleInput, bool) (bool, error), name string) {
	if h.runtime.Subscriptions()&subscription == 0 {
		return
	}
	p := ctx.Val()
	cancelled, err := handle(native.PlayerToggleInput{Player: h.playerID(p), After: after}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin toggle handler failed", "event", name, "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) handlePositionEvent(ctx *player.Context, position cube.Pos, subscription uint64, handle func(native.PlayerPositionInput, bool) (bool, error), name string) {
	if h.runtime.Subscriptions()&subscription == 0 {
		return
	}
	p := ctx.Val()
	cancelled, err := handle(native.PlayerPositionInput{Player: h.playerID(p), Position: nativeBlockPos(position)}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin position handler failed", "event", name, "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func nativeBlockPos(position cube.Pos) native.BlockPos {
	return native.BlockPos{X: int32(position.X()), Y: int32(position.Y()), Z: int32(position.Z())}
}

func blockName(block world.Block) string {
	if block == nil {
		return "minecraft:air"
	}
	name, _ := block.EncodeBlock()
	return name
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
