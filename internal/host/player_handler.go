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
	HandlePlayerMove(native.InvocationID, native.PlayerMoveInput, bool) (bool, error)
	HandlePlayerChat(native.InvocationID, native.PlayerChatInput, bool) (native.PlayerChatOutput, error)
	HandlePlayerJoin(native.InvocationID, native.PlayerJoinInput, bool) (bool, error)
	HandlePlayerQuit(native.InvocationID, native.PlayerQuitInput) error
	HandlePlayerHurt(native.InvocationID, native.PlayerHurtInput, bool) (native.PlayerHurtOutput, error)
	HandlePlayerHeal(native.InvocationID, native.PlayerHealInput, bool) (native.PlayerHealOutput, error)
	HandlePlayerBlockBreak(native.InvocationID, native.PlayerBlockBreakInput, bool) (native.PlayerBlockBreakOutput, error)
	HandlePlayerBlockPlace(native.InvocationID, native.PlayerBlockPlaceInput, bool) (bool, error)
	HandlePlayerFoodLoss(native.InvocationID, native.PlayerFoodLossInput, bool) (native.PlayerFoodLossOutput, error)
	HandlePlayerDeath(native.InvocationID, native.PlayerDeathInput, bool) (bool, error)
	HandlePlayerStartBreak(native.InvocationID, native.PlayerPositionInput, bool) (bool, error)
	HandlePlayerFireExtinguish(native.InvocationID, native.PlayerPositionInput, bool) (bool, error)
	HandlePlayerToggleSprint(native.InvocationID, native.PlayerToggleInput, bool) (bool, error)
	HandlePlayerToggleSneak(native.InvocationID, native.PlayerToggleInput, bool) (bool, error)
	HandlePlayerJump(native.InvocationID, native.PlayerID) error
	HandlePlayerTeleport(native.InvocationID, native.PlayerTeleportInput, bool) (bool, error)
	HandlePlayerExperienceGain(native.InvocationID, native.PlayerID, int, bool) (native.PlayerExperienceGainOutput, error)
	HandlePlayerPunchAir(native.InvocationID, native.PlayerID, bool) (bool, error)
	HandlePlayerHeldSlotChange(native.InvocationID, native.PlayerHeldSlotChangeInput, bool) (bool, error)
	HandlePlayerSleep(native.InvocationID, native.PlayerID, bool, bool) (native.PlayerSleepOutput, error)
	HandlePlayerBlockPick(native.InvocationID, native.PlayerBlockPickInput, bool) (bool, error)
	HandlePlayerLecternPageTurn(native.InvocationID, native.PlayerLecternPageTurnInput, bool) (native.PlayerLecternPageTurnOutput, error)
	HandlePlayerSignEdit(native.InvocationID, native.PlayerSignEditInput, bool) (bool, error)
	HandlePlayerItemUse(native.InvocationID, native.PlayerID, bool) (bool, error)
	HandlePlayerItemUseOnBlock(native.InvocationID, native.PlayerItemUseOnBlockInput, bool) (bool, error)
	HandlePlayerItemConsume(native.InvocationID, native.PlayerID, native.ItemStack, bool) (bool, error)
	HandlePlayerItemRelease(native.InvocationID, native.PlayerID, native.ItemStack, time.Duration, bool) (bool, error)
	HandlePlayerItemDamage(native.InvocationID, native.PlayerID, native.ItemStack, int, bool) (native.PlayerItemDamageOutput, error)
	HandlePlayerItemDrop(native.InvocationID, native.PlayerID, native.ItemStack, bool) (bool, error)
	HandlePlayerAttackEntity(native.InvocationID, native.PlayerAttackEntityInput, float64, float64, bool, bool) (native.PlayerAttackEntityOutput, error)
	HandlePlayerItemUseOnEntity(native.InvocationID, native.PlayerItemUseOnEntityInput, bool) (bool, error)
}

func (h *PlayerHandler) HandleJump(p *player.Player) {
	if h.runtime.Subscriptions()&native.PlayerJumpSubscription == 0 {
		return
	}
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	if err := h.runtime.HandlePlayerJump(invocation, h.playerID(p)); err != nil {
		h.log.Error("native plugin jump handler failed", "player", p.Name(), "error", err)
	}
}

func (h *PlayerHandler) HandleTeleport(ctx *player.Context, position mgl64.Vec3) {
	if h.runtime.Subscriptions()&native.PlayerTeleportSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	cancelled, err := h.runtime.HandlePlayerTeleport(invocation, native.PlayerTeleportInput{Player: h.playerID(p), Position: native.Vec3{X: position.X(), Y: position.Y(), Z: position.Z()}}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin teleport handler failed", "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleExperienceGain(ctx *player.Context, amount *int) {
	if h.runtime.Subscriptions()&native.PlayerExperienceGainSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	output, err := h.runtime.HandlePlayerExperienceGain(invocation, h.playerID(p), *amount, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin experience-gain handler failed", "player", p.Name(), "error", err)
		return
	}
	*amount = output.Amount
	if output.Cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandlePunchAir(ctx *player.Context) {
	if h.runtime.Subscriptions()&native.PlayerPunchAirSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	cancelled, err := h.runtime.HandlePlayerPunchAir(invocation, h.playerID(p), ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin punch-air handler failed", "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleHeldSlotChange(ctx *player.Context, from, to int) {
	if h.runtime.Subscriptions()&native.PlayerHeldSlotChangeSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	cancelled, err := h.runtime.HandlePlayerHeldSlotChange(invocation, native.PlayerHeldSlotChangeInput{Player: h.playerID(p), From: from, To: to}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin held-slot handler failed", "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleSleep(ctx *player.Context, sendReminder *bool) {
	if h.runtime.Subscriptions()&native.PlayerSleepSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	output, err := h.runtime.HandlePlayerSleep(invocation, h.playerID(p), *sendReminder, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin sleep handler failed", "player", p.Name(), "error", err)
		return
	}
	*sendReminder = output.SendReminder
	if output.Cancelled {
		ctx.Cancel()
	}
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

func (h *PlayerHandler) HandleAttackEntity(ctx *player.Context, target world.Entity, force, height *float64, critical *bool) {
	if h.runtime.Subscriptions()&native.PlayerAttackEntitySubscription == 0 {
		return
	}
	p := ctx.Player()
	targetID := h.players.EntityRegistry().Register(target)
	if targetID.Generation == 0 {
		return
	}
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	output, err := h.runtime.HandlePlayerAttackEntity(invocation, native.PlayerAttackEntityInput{
		Player: h.playerID(p), Target: targetID,
	}, *force, *height, *critical, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin attack-entity handler failed", "player", p.Name(), "error", err)
		return
	}
	*force, *height, *critical = output.KnockbackForce, output.KnockbackHeight, output.Critical
	if output.Cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleItemUseOnEntity(ctx *player.Context, target world.Entity) {
	if h.runtime.Subscriptions()&native.PlayerItemUseOnEntitySubscription == 0 {
		return
	}
	p := ctx.Player()
	targetID := h.players.EntityRegistry().Register(target)
	if targetID.Generation == 0 {
		return
	}
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	cancelled, err := h.runtime.HandlePlayerItemUseOnEntity(invocation, native.PlayerItemUseOnEntityInput{
		Player: h.playerID(p), Target: targetID,
	}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin item-use-on-entity handler failed", "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleMove(ctx *player.Context, newPosition mgl64.Vec3, newRotation cube.Rotation) {
	if h.runtime.Subscriptions()&native.PlayerMoveSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	oldPosition := p.Position()
	cancelled, err := h.runtime.HandlePlayerMove(invocation, native.PlayerMoveInput{
		Player:      h.playerID(p),
		OldPosition: native.Vec3{X: oldPosition.X(), Y: oldPosition.Y(), Z: oldPosition.Z()},
		NewPosition: native.Vec3{X: newPosition.X(), Y: newPosition.Y(), Z: newPosition.Z()},
		Rotation:    native.Rotation{Yaw: newRotation.Yaw(), Pitch: newRotation.Pitch()},
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
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	output, err := h.runtime.HandlePlayerChat(invocation, native.PlayerChatInput{
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
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	output, err := h.runtime.HandlePlayerHurt(invocation, native.PlayerHurtInput{
		Player:         h.playerID(p),
		Damage:         *damage,
		Immune:         immune,
		AttackImmunity: *attackImmunity,
		Source:         nativeDamageSource(source),
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
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	output, err := h.runtime.HandlePlayerHeal(invocation, native.PlayerHealInput{
		Player: h.playerID(p),
		Health: *health,
		Source: native.HealingSource{Name: fmt.Sprintf("%T", source)},
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
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	output, err := h.runtime.HandlePlayerBlockBreak(invocation, native.PlayerBlockBreakInput{
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
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	cancelled, err := h.runtime.HandlePlayerBlockPlace(invocation, native.PlayerBlockPlaceInput{
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

func (h *PlayerHandler) HandleBlockPick(ctx *player.Context, position cube.Pos, block world.Block) {
	if h.runtime.Subscriptions()&native.PlayerBlockPickSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	cancelled, err := h.runtime.HandlePlayerBlockPick(invocation, native.PlayerBlockPickInput{Player: h.playerID(p), Position: nativeBlockPos(position), Block: blockName(block)}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin block-pick handler failed", "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleLecternPageTurn(ctx *player.Context, position cube.Pos, oldPage int, newPage *int) {
	if h.runtime.Subscriptions()&native.PlayerLecternPageTurnSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	output, err := h.runtime.HandlePlayerLecternPageTurn(invocation, native.PlayerLecternPageTurnInput{Player: h.playerID(p), Position: nativeBlockPos(position), OldPage: oldPage, NewPage: *newPage}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin lectern-page handler failed", "player", p.Name(), "error", err)
		return
	}
	*newPage = output.NewPage
	if output.Cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleSignEdit(ctx *player.Context, position cube.Pos, frontSide bool, oldText, newText string) {
	if h.runtime.Subscriptions()&native.PlayerSignEditSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	cancelled, err := h.runtime.HandlePlayerSignEdit(invocation, native.PlayerSignEditInput{Player: h.playerID(p), Position: nativeBlockPos(position), FrontSide: frontSide, OldText: oldText, NewText: newText}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin sign-edit handler failed", "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleItemUse(ctx *player.Context) {
	if h.runtime.Subscriptions()&native.PlayerItemUseSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	cancelled, err := h.runtime.HandlePlayerItemUse(invocation, h.playerID(p), ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin item-use handler failed", "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleItemUseOnBlock(ctx *player.Context, position cube.Pos, face cube.Face, clickPosition mgl64.Vec3) {
	if h.runtime.Subscriptions()&native.PlayerItemUseOnBlockSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	cancelled, err := h.runtime.HandlePlayerItemUseOnBlock(invocation, native.PlayerItemUseOnBlockInput{
		Player: h.playerID(p), Position: nativeBlockPos(position), Face: int(face),
		ClickPosition: native.Vec3{X: clickPosition.X(), Y: clickPosition.Y(), Z: clickPosition.Z()},
	}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin item-use-on-block handler failed", "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleItemConsume(ctx *player.Context, stack item.Stack) {
	if h.runtime.Subscriptions()&native.PlayerItemConsumeSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	value, ok := itemStackToNative(stack)
	if !ok {
		h.log.Error("convert item consume stack", "player", p.Name())
		return
	}
	cancelled, err := h.runtime.HandlePlayerItemConsume(invocation, h.playerID(p), value, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin item-consume handler failed", "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleItemRelease(ctx *player.Context, stack item.Stack, duration time.Duration) {
	if h.runtime.Subscriptions()&native.PlayerItemReleaseSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	value, ok := itemStackToNative(stack)
	if !ok {
		h.log.Error("convert item release stack", "player", p.Name())
		return
	}
	cancelled, err := h.runtime.HandlePlayerItemRelease(invocation, h.playerID(p), value, duration, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin item-release handler failed", "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleItemDamage(ctx *player.Context, stack item.Stack, damage *int) {
	if h.runtime.Subscriptions()&native.PlayerItemDamageSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	value, ok := itemStackToNative(stack)
	if !ok {
		h.log.Error("convert item damage stack", "player", p.Name())
		return
	}
	output, err := h.runtime.HandlePlayerItemDamage(invocation, h.playerID(p), value, *damage, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin item-damage handler failed", "player", p.Name(), "error", err)
		return
	}
	*damage = output.Damage
	if output.Cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleItemDrop(ctx *player.Context, stack item.Stack) {
	if h.runtime.Subscriptions()&native.PlayerItemDropSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	value, ok := itemStackToNative(stack)
	if !ok {
		h.log.Error("convert item drop stack", "player", p.Name())
		return
	}
	cancelled, err := h.runtime.HandlePlayerItemDrop(invocation, h.playerID(p), value, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin item-drop handler failed", "player", p.Name(), "error", err)
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
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	output, err := h.runtime.HandlePlayerFoodLoss(invocation, native.PlayerFoodLossInput{
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
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	keep, err := h.runtime.HandlePlayerDeath(invocation, native.PlayerDeathInput{
		Player: h.playerID(p),
		Source: nativeDamageSource(source),
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
func (h *PlayerHandler) handleToggleEvent(ctx *player.Context, after bool, subscription uint64, handle func(native.InvocationID, native.PlayerToggleInput, bool) (bool, error), name string) {
	if h.runtime.Subscriptions()&subscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	cancelled, err := handle(invocation, native.PlayerToggleInput{Player: h.playerID(p), After: after}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin toggle handler failed", "event", name, "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) handlePositionEvent(ctx *player.Context, position cube.Pos, subscription uint64, handle func(native.InvocationID, native.PlayerPositionInput, bool) (bool, error), name string) {
	if h.runtime.Subscriptions()&subscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	cancelled, err := handle(invocation, native.PlayerPositionInput{Player: h.playerID(p), Position: nativeBlockPos(position)}, ctx.Cancelled())
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

func nativeDamageSource(source world.DamageSource) native.DamageSource {
	if source == nil {
		return native.DamageSource{Name: "<nil>"}
	}
	return native.DamageSource{
		Name: fmt.Sprintf("%T", source), ReducedByArmour: source.ReducedByArmour(),
		ReducedByResistance: source.ReducedByResistance(), Fire: source.Fire(),
		IgnoresTotem: source.IgnoreTotem(),
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
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	cancelled, err := h.runtime.HandlePlayerJoin(invocation, native.PlayerJoinInput{
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
		invocation, leave := h.players.BeginInvocation(p.Tx())
		defer leave()
		if err := h.runtime.HandlePlayerQuit(invocation, native.PlayerQuitInput{
			Player: h.playerID(p),
			Name:   p.Name(),
		}); err != nil {
			h.log.Error("native plugin quit handler failed", "player", p.Name(), "error", err)
		}
	}
	h.players.Unregister(p)
}
