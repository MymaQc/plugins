// Package host adapts Dragonfly lifecycle and handlers to the native runtime.
package host

import (
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/player"
	playerskin "github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
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
	HandlePlayerJump(native.InvocationID, native.PlayerSnapshot) error
	HandlePlayerTeleport(native.InvocationID, native.PlayerTeleportInput, bool) (bool, error)
	HandlePlayerExperienceGain(native.InvocationID, native.PlayerSnapshot, int, bool) (native.PlayerExperienceGainOutput, error)
	HandlePlayerPunchAir(native.InvocationID, native.PlayerSnapshot, bool) (bool, error)
	HandlePlayerHeldSlotChange(native.InvocationID, native.PlayerHeldSlotChangeInput, bool) (bool, error)
	HandlePlayerSleep(native.InvocationID, native.PlayerSnapshot, bool, bool) (native.PlayerSleepOutput, error)
	HandlePlayerBlockPick(native.InvocationID, native.PlayerBlockPickInput, bool) (bool, error)
	HandlePlayerLecternPageTurn(native.InvocationID, native.PlayerLecternPageTurnInput, bool) (native.PlayerLecternPageTurnOutput, error)
	HandlePlayerSignEdit(native.InvocationID, native.PlayerSignEditInput, bool) (bool, error)
	HandlePlayerItemUse(native.InvocationID, native.PlayerSnapshot, bool) (bool, error)
	HandlePlayerItemUseOnBlock(native.InvocationID, native.PlayerItemUseOnBlockInput, bool) (bool, error)
	HandlePlayerItemConsume(native.InvocationID, native.PlayerSnapshot, native.ItemStack, bool) (bool, error)
	HandlePlayerItemRelease(native.InvocationID, native.PlayerSnapshot, native.ItemStack, time.Duration, bool) (bool, error)
	HandlePlayerItemDamage(native.InvocationID, native.PlayerSnapshot, native.ItemStack, int, bool) (native.PlayerItemDamageOutput, error)
	HandlePlayerItemDrop(native.InvocationID, native.PlayerSnapshot, native.ItemStack, bool) (bool, error)
	HandlePlayerItemPickup(native.InvocationID, native.PlayerItemPickupInput, bool) (native.PlayerItemPickupOutput, error)
	HandlePlayerAttackEntity(native.InvocationID, native.PlayerAttackEntityInput, float64, float64, bool, bool) (native.PlayerAttackEntityOutput, error)
	HandlePlayerItemUseOnEntity(native.InvocationID, native.PlayerItemUseOnEntityInput, bool) (bool, error)
	HandlePlayerChangeWorld(native.InvocationID, native.PlayerChangeWorldInput) error
	HandlePlayerRespawn(native.InvocationID, native.PlayerRespawnInput, native.Vec3, native.WorldID) (native.PlayerRespawnOutput, error)
	HandlePlayerSkinChange(native.InvocationID, native.PlayerSkinChangeInput, native.PlayerSkin, bool) (native.PlayerSkinChangeOutput, error)
	HandlePlayerTransfer(native.InvocationID, native.PlayerTransferInput, bool) (native.PlayerTransferOutput, error)
	HandlePlayerCommandExecution(native.InvocationID, native.PlayerCommandExecutionInput, bool) (native.PlayerCommandExecutionOutput, error)
	HandlePlayerDiagnostics(native.InvocationID, native.PlayerDiagnosticsInput) error
}

type playerWorldResolver interface {
	WorldHandle(*world.World) (native.WorldID, bool)
	WorldByHandle(native.WorldID) (*world.World, bool)
}

func (h *PlayerHandler) HandleJump(p *player.Player) {
	if h.runtime.Subscriptions()&native.PlayerJumpSubscription == 0 {
		return
	}
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	if err := h.runtime.HandlePlayerJump(invocation, h.playerSnapshot(p)); err != nil {
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
	cancelled, err := h.runtime.HandlePlayerTeleport(invocation, native.PlayerTeleportInput{Player: h.playerSnapshot(p), Position: native.Vec3{X: position.X(), Y: position.Y(), Z: position.Z()}}, ctx.Cancelled())
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
	output, err := h.runtime.HandlePlayerExperienceGain(invocation, h.playerSnapshot(p), *amount, ctx.Cancelled())
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
	cancelled, err := h.runtime.HandlePlayerPunchAir(invocation, h.playerSnapshot(p), ctx.Cancelled())
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
	cancelled, err := h.runtime.HandlePlayerHeldSlotChange(invocation, native.PlayerHeldSlotChangeInput{Player: h.playerSnapshot(p), From: from, To: to}, ctx.Cancelled())
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
	output, err := h.runtime.HandlePlayerSleep(invocation, h.playerSnapshot(p), *sendReminder, ctx.Cancelled())
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
	worlds  playerWorldResolver
}

var _ player.Handler = (*PlayerHandler)(nil)

func NewPlayerHandler(runtime playerRuntime, log *slog.Logger, players *Players, worlds playerWorldResolver) *PlayerHandler {
	if log == nil {
		log = slog.Default()
	}
	return &PlayerHandler{runtime: runtime, log: log, players: players, worlds: worlds}
}

func (h *PlayerHandler) HandleChangeWorld(p *player.Player, before, after *world.World) {
	if h.runtime.Subscriptions()&native.PlayerChangeWorldSubscription == 0 || h.worlds == nil || after == nil {
		return
	}
	afterID, ok := h.worlds.WorldHandle(after)
	if !ok {
		return
	}
	input := native.PlayerChangeWorldInput{Player: h.playerSnapshot(p), After: afterID}
	departedID, departed := h.players.takeWorldDeparture(p)
	if before != nil {
		beforeID, ok := h.worlds.WorldHandle(before)
		if ok {
			input.Before = &beforeID
		} else if departed {
			input.Before = &departedID
		} else {
			return
		}
	}
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	if err := h.runtime.HandlePlayerChangeWorld(invocation, input); err != nil {
		h.log.Error("native plugin change-world handler failed", "player", p.Name(), "error", err)
	}
}

func (h *PlayerHandler) HandleRespawn(p *player.Player, position *mgl64.Vec3, destination **world.World) {
	if h.runtime.Subscriptions()&native.PlayerRespawnSubscription == 0 || h.worlds == nil || position == nil || destination == nil || *destination == nil {
		return
	}
	destinationID, ok := h.worlds.WorldHandle(*destination)
	if !ok {
		return
	}
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	output, err := h.runtime.HandlePlayerRespawn(invocation, native.PlayerRespawnInput{Player: h.playerSnapshot(p)}, native.Vec3{
		X: position.X(), Y: position.Y(), Z: position.Z(),
	}, destinationID)
	if err != nil {
		h.log.Error("native plugin respawn handler failed", "player", p.Name(), "error", err)
		return
	}
	if !finite(output.Position.X, output.Position.Y, output.Position.Z) {
		h.log.Error("native plugin respawn handler returned an invalid position", "player", p.Name())
		return
	}
	returnedWorld, ok := h.worlds.WorldByHandle(output.World)
	if !ok {
		h.log.Error("native plugin respawn handler returned an unknown world", "player", p.Name(), "world", output.World)
		return
	}
	*position = mgl64.Vec3{output.Position.X, output.Position.Y, output.Position.Z}
	*destination = returnedWorld
}

func (h *PlayerHandler) HandleSkinChange(ctx *player.Context, candidate *playerskin.Skin) {
	if h.runtime.Subscriptions()&native.PlayerSkinChangeSubscription == 0 || candidate == nil {
		return
	}
	p := ctx.Player()
	value, ok := playerSkinToNative(*candidate)
	if !ok {
		return
	}
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	output, err := h.runtime.HandlePlayerSkinChange(invocation, native.PlayerSkinChangeInput{Player: h.playerSnapshot(p)}, value, ctx.Cancelled())
	if err != nil {
		if output.Cancelled {
			ctx.Cancel()
		}
		h.log.Error("native plugin skin-change handler failed", "player", p.Name(), "error", err)
		return
	}
	if output.Cancelled {
		ctx.Cancel()
	}
	converted, ok := playerSkinFromNative(output.Skin)
	if !ok {
		h.log.Error("native plugin skin-change handler returned an invalid skin", "player", p.Name())
		return
	}
	*candidate = converted
}

func (h *PlayerHandler) HandleTransfer(ctx *player.Context, address *net.UDPAddr) {
	if h.runtime.Subscriptions()&native.PlayerTransferSubscription == 0 || address == nil {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	output, err := h.runtime.HandlePlayerTransfer(invocation, native.PlayerTransferInput{
		Player: h.playerSnapshot(p),
		Address: native.UDPAddress{
			IP:   append([]byte(nil), address.IP...),
			Port: address.Port,
			Zone: address.Zone,
		},
	}, ctx.Cancelled())
	if output.Cancelled {
		ctx.Cancel()
	}
	if err != nil {
		h.log.Error("native plugin transfer handler failed", "player", p.Name(), "error", err)
		return
	}
	if !applyTransferAddress(address, output.Address) {
		h.log.Error("native plugin transfer handler returned an invalid address", "player", p.Name())
	}
}

func (h *PlayerHandler) HandleCommandExecution(ctx *player.Context, command cmd.Command, arguments []string) {
	if h.runtime.Subscriptions()&native.PlayerCommandExecutionSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	output, err := h.runtime.HandlePlayerCommandExecution(invocation, native.PlayerCommandExecutionInput{
		Player: h.playerSnapshot(p),
		Command: native.CommandInfo{
			Name:        command.Name(),
			Description: command.Description(),
			Usage:       command.Usage(),
			Aliases:     append([]string(nil), command.Aliases()...),
		},
		Arguments: append([]string(nil), arguments...),
	}, ctx.Cancelled())
	if output.Cancelled {
		ctx.Cancel()
	}
	if err != nil {
		h.log.Error("native plugin command-execution handler failed", "player", p.Name(), "command", command.Name(), "error", err)
		return
	}
	if !applyCommandArguments(arguments, output.Arguments) {
		h.log.Error("native plugin command-execution handler returned an invalid argument count", "player", p.Name(), "command", command.Name())
	}
}

func (h *PlayerHandler) HandleDiagnostics(p *player.Player, diagnostics session.Diagnostics) {
	if h.runtime.Subscriptions()&native.PlayerDiagnosticsSubscription == 0 {
		return
	}
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	if err := h.runtime.HandlePlayerDiagnostics(invocation, native.PlayerDiagnosticsInput{
		Player:                        h.playerSnapshot(p),
		AverageFramesPerSecond:        diagnostics.AverageFramesPerSecond,
		AverageServerSimTickTime:      diagnostics.AverageServerSimTickTime,
		AverageClientSimTickTime:      diagnostics.AverageClientSimTickTime,
		AverageBeginFrameTime:         diagnostics.AverageBeginFrameTime,
		AverageInputTime:              diagnostics.AverageInputTime,
		AverageRenderTime:             diagnostics.AverageRenderTime,
		AverageEndFrameTime:           diagnostics.AverageEndFrameTime,
		AverageRemainderTimePercent:   diagnostics.AverageRemainderTimePercent,
		AverageUnaccountedTimePercent: diagnostics.AverageUnaccountedTimePercent,
	}); err != nil {
		h.log.Error("native plugin diagnostics handler failed", "player", p.Name(), "error", err)
	}
}

func applyTransferAddress(destination *net.UDPAddr, source native.UDPAddress) bool {
	if destination == nil || (len(source.IP) != 0 && len(source.IP) != net.IPv4len && len(source.IP) != net.IPv6len) {
		return false
	}
	destination.IP = append(net.IP(nil), source.IP...)
	destination.Port = source.Port
	destination.Zone = source.Zone
	return true
}

func applyCommandArguments(destination, source []string) bool {
	if len(destination) != len(source) {
		return false
	}
	copy(destination, source)
	return true
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
		Player: h.playerSnapshot(p), Target: targetID,
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
		Player: h.playerSnapshot(p), Target: targetID,
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
		Player:      h.playerSnapshotAt(p, oldPosition),
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
		Player:  h.playerSnapshot(p),
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
		Player:         h.playerSnapshot(p),
		Damage:         *damage,
		Immune:         immune,
		AttackImmunity: *attackImmunity,
		Source:         h.nativeDamageSource(source),
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
		Player: h.playerSnapshot(p),
		Health: *health,
		Source: h.nativeHealingSource(source),
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

func (h *PlayerHandler) HandleBlockBreak(ctx *player.Context, position cube.Pos, drops *[]item.Stack, experience *int) {
	if h.runtime.Subscriptions()&native.PlayerBlockBreakSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	blockValue, ok := nativeEventBlock(p.Tx().Block(position))
	if !ok {
		h.log.Error("convert block-break block", "player", p.Name())
		return
	}
	dropValues := make([]native.ItemStack, len(*drops))
	for index, stack := range *drops {
		value, valid := itemStackToNative(stack)
		if !valid {
			h.log.Error("convert block-break drop", "player", p.Name(), "drop", index)
			return
		}
		dropValues[index] = value
	}
	output, err := h.runtime.HandlePlayerBlockBreak(invocation, native.PlayerBlockBreakInput{
		Player:     h.playerSnapshot(p),
		Position:   nativeBlockPos(position),
		Block:      blockValue,
		Drops:      dropValues,
		Experience: int32(*experience),
	}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin block-break handler failed", "player", p.Name(), "error", err)
		return
	}
	replacementDrops := make([]item.Stack, len(output.Drops))
	for index, value := range output.Drops {
		stack, valid := itemStackFromNative(value)
		if !valid {
			h.log.Error("convert native block-break drop", "player", p.Name(), "drop", index)
			return
		}
		replacementDrops[index] = stack
	}
	*drops, *experience = replacementDrops, int(output.Experience)
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
	blockValue, ok := nativeEventBlock(block)
	if !ok {
		h.log.Error("convert block-place block", "player", p.Name())
		return
	}
	cancelled, err := h.runtime.HandlePlayerBlockPlace(invocation, native.PlayerBlockPlaceInput{
		Player:   h.playerSnapshot(p),
		Position: nativeBlockPos(position),
		Block:    blockValue,
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
	blockValue, ok := nativeEventBlock(block)
	if !ok {
		h.log.Error("convert block-pick block", "player", p.Name())
		return
	}
	cancelled, err := h.runtime.HandlePlayerBlockPick(invocation, native.PlayerBlockPickInput{Player: h.playerSnapshot(p), Position: nativeBlockPos(position), Block: blockValue}, ctx.Cancelled())
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
	output, err := h.runtime.HandlePlayerLecternPageTurn(invocation, native.PlayerLecternPageTurnInput{Player: h.playerSnapshot(p), Position: nativeBlockPos(position), OldPage: oldPage, NewPage: *newPage}, ctx.Cancelled())
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
	cancelled, err := h.runtime.HandlePlayerSignEdit(invocation, native.PlayerSignEditInput{Player: h.playerSnapshot(p), Position: nativeBlockPos(position), FrontSide: frontSide, OldText: oldText, NewText: newText}, ctx.Cancelled())
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
	cancelled, err := h.runtime.HandlePlayerItemUse(invocation, h.playerSnapshot(p), ctx.Cancelled())
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
		Player: h.playerSnapshot(p), Position: nativeBlockPos(position), Face: int(face),
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
	cancelled, err := h.runtime.HandlePlayerItemConsume(invocation, h.playerSnapshot(p), value, ctx.Cancelled())
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
	cancelled, err := h.runtime.HandlePlayerItemRelease(invocation, h.playerSnapshot(p), value, duration, ctx.Cancelled())
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
	output, err := h.runtime.HandlePlayerItemDamage(invocation, h.playerSnapshot(p), value, *damage, ctx.Cancelled())
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
	cancelled, err := h.runtime.HandlePlayerItemDrop(invocation, h.playerSnapshot(p), value, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin item-drop handler failed", "player", p.Name(), "error", err)
		return
	}
	if cancelled {
		ctx.Cancel()
	}
}

func (h *PlayerHandler) HandleItemPickup(ctx *player.Context, stack *item.Stack) {
	if h.runtime.Subscriptions()&native.PlayerItemPickupSubscription == 0 {
		return
	}
	p := ctx.Player()
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	value, ok := itemStackToNative(*stack)
	if !ok {
		h.log.Error("convert item pickup stack", "player", p.Name())
		return
	}
	output, err := h.runtime.HandlePlayerItemPickup(invocation, native.PlayerItemPickupInput{
		Player: h.playerSnapshot(p), Item: value,
	}, ctx.Cancelled())
	if err != nil {
		h.log.Error("native plugin item-pickup handler failed", "player", p.Name(), "error", err)
		return
	}
	replacement, ok := itemStackFromNative(output.Item)
	if !ok {
		h.log.Error("convert native item pickup stack", "player", p.Name())
		return
	}
	*stack = replacement
	if output.Cancelled {
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
		Player: h.playerSnapshot(p),
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
		Player: h.playerSnapshot(p),
		Source: h.nativeDamageSource(source),
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
	cancelled, err := handle(invocation, native.PlayerToggleInput{Player: h.playerSnapshot(p), After: after}, ctx.Cancelled())
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
	cancelled, err := handle(invocation, native.PlayerPositionInput{Player: h.playerSnapshot(p), Position: nativeBlockPos(position)}, ctx.Cancelled())
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

func nativeEventBlock(block world.Block) (native.WorldBlock, bool) {
	if block == nil {
		properties, ok := EncodeBlockProperties(map[string]any{})
		return native.WorldBlock{Identifier: "minecraft:air", PropertiesNBT: properties}, ok
	}
	identifier, properties := block.EncodeBlock()
	encoded, ok := EncodeBlockProperties(properties)
	return native.WorldBlock{Identifier: identifier, PropertiesNBT: encoded}, ok
}

func (h *PlayerHandler) nativeDamageSource(source world.DamageSource) native.DamageSource {
	return NativeDamageSource(source, h.players.EntityRegistry())
}

// NativeDamageSource converts a Dragonfly damage source without losing its
// concrete attacker, projectile, block, or enchantment semantics. Entities may
// be nil when no stable entity attribution is available.
func NativeDamageSource(source world.DamageSource, entities *Entities) native.DamageSource {
	source = concreteDamageSource(source)
	if source == nil {
		return native.DamageSource{Name: "<nil>"}
	}
	value := native.DamageSource{
		Name: fmt.Sprintf("%T", source), ReducedByArmour: source.ReducedByArmour(),
		ReducedByResistance: source.ReducedByResistance(), Fire: source.Fire(),
		IgnoresTotem: source.IgnoreTotem(),
	}
	if named, ok := source.(interface{ Name() string }); ok && named.Name() != "" {
		value.Name = named.Name()
	}
	if affected, ok := source.(enchantment.AffectedDamageSource); ok {
		value.FireProtection = affected.AffectedByEnchantment(enchantment.FireProtection)
		value.FeatherFalling = affected.AffectedByEnchantment(enchantment.FeatherFalling)
		value.BlastProtection = affected.AffectedByEnchantment(enchantment.BlastProtection)
		value.ProjectileProtection = affected.AffectedByEnchantment(enchantment.ProjectileProtection)
	}
	switch source := source.(type) {
	case entity.AttackDamageSource:
		value.Kind = native.DamageSourceAttack
		if source.Attacker != nil && entities != nil {
			value.Entity = entities.Register(source.Attacker)
		}
	case block.DamageSource:
		if source.Block == nil {
			break
		}
		identifier, properties := source.Block.EncodeBlock()
		encoded, ok := EncodeBlockProperties(properties)
		if !ok {
			break
		}
		value.Kind, value.Block = native.DamageSourceBlock, &native.WorldBlock{Identifier: identifier, PropertiesNBT: encoded}
	case entity.DrowningDamageSource:
		value.Kind = native.DamageSourceDrowning
	case entity.ExplosionDamageSource:
		value.Kind, value.BlastProtection = native.DamageSourceExplosion, true
	case entity.FallDamageSource:
		value.Kind, value.FeatherFalling = native.DamageSourceFall, true
	case block.FireDamageSource:
		value.Kind, value.FireProtection = native.DamageSourceFire, true
	case entity.GlideDamageSource:
		value.Kind = native.DamageSourceGlide
	case effect.InstantDamageSource:
		value.Kind = native.DamageSourceInstant
	case block.LavaDamageSource:
		value.Kind = native.DamageSourceLava
	case entity.LightningDamageSource:
		value.Kind = native.DamageSourceLightning
	case block.MagmaDamageSource:
		value.Kind, value.FireProtection = native.DamageSourceMagma, true
	case effect.PoisonDamageSource:
		value.Kind, value.Data = native.DamageSourcePoison, source.Fatal
	case entity.ProjectileDamageSource:
		value.Kind = native.DamageSourceProjectile
		if source.Projectile != nil && entities != nil {
			value.Entity = entities.Register(source.Projectile)
		}
		if source.Owner != nil && entities != nil {
			value.SecondaryEntity = entities.Register(source.Owner)
		}
		value.ProjectileProtection = true
	case player.StarvationDamageSource:
		value.Kind = native.DamageSourceStarvation
	case entity.SuffocationDamageSource:
		value.Kind = native.DamageSourceSuffocation
	case enchantment.ThornsDamageSource:
		value.Kind = native.DamageSourceThorns
		if source.Owner != nil && entities != nil {
			value.Entity = entities.Register(source.Owner)
		}
	case entity.VoidDamageSource:
		value.Kind = native.DamageSourceVoid
	case effect.WitherDamageSource:
		value.Kind = native.DamageSourceWither
	}
	return value
}

func (h *PlayerHandler) nativeHealingSource(source world.HealingSource) native.HealingSource {
	return NativeHealingSource(source)
}

// NativeHealingSource converts a Dragonfly healing source to the stable ABI
// representation shared by player and custom living entity callbacks.
func NativeHealingSource(source world.HealingSource) native.HealingSource {
	source = concreteHealingSource(source)
	if source == nil {
		return native.HealingSource{Name: "<nil>"}
	}
	value := native.HealingSource{Name: fmt.Sprintf("%T", source)}
	if named, ok := source.(interface{ Name() string }); ok && named.Name() != "" {
		value.Name = named.Name()
	}
	switch source := source.(type) {
	case entity.FoodHealingSource:
		value.Kind, value.Data = native.HealingSourceFood, source.QuickRegeneration
	case effect.InstantHealingSource:
		value.Kind = native.HealingSourceInstant
	case effect.RegenerationHealingSource:
		value.Kind = native.HealingSourceRegeneration
	}
	return value
}

func concreteDamageSource(source world.DamageSource) world.DamageSource {
	switch source := source.(type) {
	case *entity.AttackDamageSource:
		if source != nil {
			return *source
		}
	case *block.DamageSource:
		if source != nil {
			return *source
		}
	case *entity.DrowningDamageSource:
		if source != nil {
			return *source
		}
	case *entity.ExplosionDamageSource:
		if source != nil {
			return *source
		}
	case *entity.FallDamageSource:
		if source != nil {
			return *source
		}
	case *block.FireDamageSource:
		if source != nil {
			return *source
		}
	case *entity.GlideDamageSource:
		if source != nil {
			return *source
		}
	case *effect.InstantDamageSource:
		if source != nil {
			return *source
		}
	case *block.LavaDamageSource:
		if source != nil {
			return *source
		}
	case *entity.LightningDamageSource:
		if source != nil {
			return *source
		}
	case *block.MagmaDamageSource:
		if source != nil {
			return *source
		}
	case *effect.PoisonDamageSource:
		if source != nil {
			return *source
		}
	case *entity.ProjectileDamageSource:
		if source != nil {
			return *source
		}
	case *player.StarvationDamageSource:
		if source != nil {
			return *source
		}
	case *entity.SuffocationDamageSource:
		if source != nil {
			return *source
		}
	case *enchantment.ThornsDamageSource:
		if source != nil {
			return *source
		}
	case *entity.VoidDamageSource:
		if source != nil {
			return *source
		}
	case *effect.WitherDamageSource:
		if source != nil {
			return *source
		}
	default:
		return source
	}
	return nil
}

func concreteHealingSource(source world.HealingSource) world.HealingSource {
	switch source := source.(type) {
	case *entity.FoodHealingSource:
		if source != nil {
			return *source
		}
	case *effect.InstantHealingSource:
		if source != nil {
			return *source
		}
	case *effect.RegenerationHealingSource:
		if source != nil {
			return *source
		}
	default:
		return source
	}
	return nil
}

func (h *PlayerHandler) playerSnapshot(p *player.Player) native.PlayerSnapshot {
	position := p.Position()
	return h.playerSnapshotAt(p, position)
}

func (h *PlayerHandler) playerSnapshotAt(p *player.Player, position mgl64.Vec3) native.PlayerSnapshot {
	id, _ := h.players.ID(p)
	return native.PlayerSnapshot{
		Player:              id,
		Name:                p.Name(),
		LatencyMilliseconds: uint64(max(p.Latency().Milliseconds(), 0)),
		Position:            native.Vec3{X: position.X(), Y: position.Y(), Z: position.Z()},
	}
}

func (h *PlayerHandler) Join(p *player.Player) bool {
	if h.runtime.Subscriptions()&native.PlayerJoinSubscription == 0 {
		return false
	}
	invocation, leave := h.players.BeginInvocation(p.Tx())
	defer leave()
	cancelled, err := h.runtime.HandlePlayerJoin(invocation, native.PlayerJoinInput{
		Player: h.playerSnapshot(p),
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
			Player: h.playerSnapshot(p),
			Name:   p.Name(),
		}); err != nil {
			h.log.Error("native plugin quit handler failed", "player", p.Name(), "error", err)
		}
	}
	h.players.Unregister(p)
}
