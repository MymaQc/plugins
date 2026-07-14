package host

import (
	"context"
	"errors"
	"math"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/player"
	playerskin "github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
)

type runtimeStub struct {
	subscriptions       uint64
	moveCancelled       bool
	chatOutput          native.PlayerChatOutput
	moveInput           native.PlayerMoveInput
	chatInput           native.PlayerChatInput
	joinInput           native.PlayerJoinInput
	quitInput           native.PlayerQuitInput
	joinCancelled       bool
	hurtInput           native.PlayerHurtInput
	hurtOutput          native.PlayerHurtOutput
	healInput           native.PlayerHealInput
	healOutput          native.PlayerHealOutput
	blockBreakInput     native.PlayerBlockBreakInput
	blockBreakOutput    native.PlayerBlockBreakOutput
	blockPlaceInput     native.PlayerBlockPlaceInput
	blockPlaceCancelled bool
	foodLossInput       native.PlayerFoodLossInput
	foodLossOutput      native.PlayerFoodLossOutput
	deathInput          native.PlayerDeathInput
	keepInventory       bool
	attackInput         native.PlayerAttackEntityInput
	attackOutput        native.PlayerAttackEntityOutput
	itemUseEntityInput  native.PlayerItemUseOnEntityInput
	itemUseEntityCancel bool
	changeWorldInput    native.PlayerChangeWorldInput
	changeWorldTx       *world.World
	changeWorldCalls    int
	respawnInput        native.PlayerRespawnInput
	respawnOutput       native.PlayerRespawnOutput
	respawnTx           *world.World
	skinChangeInput     native.PlayerSkinChangeInput
	skinChangeOutput    native.PlayerSkinChangeOutput
	skinChangeCalled    bool
	skinChangeErr       error
	transferInput       native.PlayerTransferInput
	transferOutput      native.PlayerTransferOutput
	transferErr         error
	transferTx          *world.World
	commandInput        native.PlayerCommandExecutionInput
	commandOutput       native.PlayerCommandExecutionOutput
	commandErr          error
	commandTx           *world.World
	diagnosticsInput    native.PlayerDiagnosticsInput
	diagnosticsCalled   bool
	diagnosticsErr      error
	diagnosticsTx       *world.World
	players             *Players
}

type testDamageSource struct{}

func (testDamageSource) ReducedByArmour() bool     { return true }
func (testDamageSource) ReducedByResistance() bool { return false }
func (testDamageSource) Fire() bool                { return true }
func (testDamageSource) IgnoreTotem() bool         { return true }

type testHealingSource struct{}

func (testHealingSource) HealingSource() {}

type commandExecutionTestRunnable struct{}

func (commandExecutionTestRunnable) Run(cmd.Source, *cmd.Output, *world.Tx) {}

func (r *runtimeStub) HandlePlayerJoin(_ native.InvocationID, input native.PlayerJoinInput, _ bool) (bool, error) {
	r.joinInput = input
	return r.joinCancelled, nil
}
func (r *runtimeStub) HandlePlayerQuit(_ native.InvocationID, input native.PlayerQuitInput) error {
	r.quitInput = input
	return nil
}
func (r *runtimeStub) HandlePlayerHurt(_ native.InvocationID, input native.PlayerHurtInput, _ bool) (native.PlayerHurtOutput, error) {
	r.hurtInput = input
	return r.hurtOutput, nil
}
func (r *runtimeStub) HandlePlayerHeal(_ native.InvocationID, input native.PlayerHealInput, _ bool) (native.PlayerHealOutput, error) {
	r.healInput = input
	return r.healOutput, nil
}
func (r *runtimeStub) HandlePlayerBlockBreak(_ native.InvocationID, input native.PlayerBlockBreakInput, _ bool) (native.PlayerBlockBreakOutput, error) {
	r.blockBreakInput = input
	return r.blockBreakOutput, nil
}
func (r *runtimeStub) HandlePlayerBlockPlace(_ native.InvocationID, input native.PlayerBlockPlaceInput, _ bool) (bool, error) {
	r.blockPlaceInput = input
	return r.blockPlaceCancelled, nil
}
func (r *runtimeStub) HandlePlayerFoodLoss(_ native.InvocationID, input native.PlayerFoodLossInput, _ bool) (native.PlayerFoodLossOutput, error) {
	r.foodLossInput = input
	return r.foodLossOutput, nil
}
func (r *runtimeStub) HandlePlayerDeath(_ native.InvocationID, input native.PlayerDeathInput, _ bool) (bool, error) {
	r.deathInput = input
	return r.keepInventory, nil
}
func (r *runtimeStub) HandlePlayerStartBreak(_ native.InvocationID, _ native.PlayerPositionInput, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerFireExtinguish(_ native.InvocationID, _ native.PlayerPositionInput, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerToggleSprint(_ native.InvocationID, _ native.PlayerToggleInput, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerToggleSneak(_ native.InvocationID, _ native.PlayerToggleInput, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerJump(_ native.InvocationID, _ native.PlayerSnapshot) error {
	return nil
}
func (r *runtimeStub) HandlePlayerTeleport(_ native.InvocationID, _ native.PlayerTeleportInput, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerExperienceGain(_ native.InvocationID, _ native.PlayerSnapshot, amount int, cancelled bool) (native.PlayerExperienceGainOutput, error) {
	return native.PlayerExperienceGainOutput{Cancelled: cancelled, Amount: amount}, nil
}
func (r *runtimeStub) HandlePlayerPunchAir(_ native.InvocationID, _ native.PlayerSnapshot, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerHeldSlotChange(_ native.InvocationID, _ native.PlayerHeldSlotChangeInput, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerSleep(_ native.InvocationID, _ native.PlayerSnapshot, sendReminder, cancelled bool) (native.PlayerSleepOutput, error) {
	return native.PlayerSleepOutput{Cancelled: cancelled, SendReminder: sendReminder}, nil
}
func (r *runtimeStub) HandlePlayerBlockPick(_ native.InvocationID, _ native.PlayerBlockPickInput, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerLecternPageTurn(_ native.InvocationID, input native.PlayerLecternPageTurnInput, cancelled bool) (native.PlayerLecternPageTurnOutput, error) {
	return native.PlayerLecternPageTurnOutput{Cancelled: cancelled, NewPage: input.NewPage}, nil
}
func (r *runtimeStub) HandlePlayerSignEdit(_ native.InvocationID, _ native.PlayerSignEditInput, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerItemUse(_ native.InvocationID, _ native.PlayerSnapshot, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerItemUseOnBlock(_ native.InvocationID, _ native.PlayerItemUseOnBlockInput, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerItemConsume(_ native.InvocationID, _ native.PlayerSnapshot, _ native.ItemStack, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerItemRelease(_ native.InvocationID, _ native.PlayerSnapshot, _ native.ItemStack, _ time.Duration, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerItemDamage(_ native.InvocationID, _ native.PlayerSnapshot, _ native.ItemStack, damage int, cancelled bool) (native.PlayerItemDamageOutput, error) {
	return native.PlayerItemDamageOutput{Cancelled: cancelled, Damage: damage}, nil
}
func (r *runtimeStub) HandlePlayerItemDrop(_ native.InvocationID, _ native.PlayerSnapshot, _ native.ItemStack, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerItemPickup(_ native.InvocationID, input native.PlayerItemPickupInput, cancelled bool) (native.PlayerItemPickupOutput, error) {
	return native.PlayerItemPickupOutput{Cancelled: cancelled, Item: input.Item}, nil
}
func (r *runtimeStub) HandlePlayerAttackEntity(_ native.InvocationID, input native.PlayerAttackEntityInput, force, height float64, critical, cancelled bool) (native.PlayerAttackEntityOutput, error) {
	r.attackInput = input
	output := r.attackOutput
	if output == (native.PlayerAttackEntityOutput{}) {
		output = native.PlayerAttackEntityOutput{Cancelled: cancelled, KnockbackForce: force, KnockbackHeight: height, Critical: critical}
	}
	return output, nil
}
func (r *runtimeStub) HandlePlayerItemUseOnEntity(_ native.InvocationID, input native.PlayerItemUseOnEntityInput, cancelled bool) (bool, error) {
	r.itemUseEntityInput = input
	return cancelled || r.itemUseEntityCancel, nil
}
func (r *runtimeStub) HandlePlayerChangeWorld(invocation native.InvocationID, input native.PlayerChangeWorldInput) error {
	r.changeWorldCalls++
	r.changeWorldInput = input
	if r.players != nil {
		if tx, ok := r.players.InvocationTx(invocation); ok {
			r.changeWorldTx = tx.World()
		}
	}
	return nil
}

func (r *runtimeStub) HandlePlayerRespawn(invocation native.InvocationID, input native.PlayerRespawnInput, position native.Vec3, worldID native.WorldID) (native.PlayerRespawnOutput, error) {
	r.respawnInput = input
	if r.players != nil {
		if tx, ok := r.players.InvocationTx(invocation); ok {
			r.respawnTx = tx.World()
		}
	}
	if r.respawnOutput == (native.PlayerRespawnOutput{}) {
		return native.PlayerRespawnOutput{Position: position, World: worldID}, nil
	}
	return r.respawnOutput, nil
}

func (r *runtimeStub) HandlePlayerSkinChange(_ native.InvocationID, input native.PlayerSkinChangeInput, skin native.PlayerSkin, cancelled bool) (native.PlayerSkinChangeOutput, error) {
	r.skinChangeCalled = true
	r.skinChangeInput = input
	if r.skinChangeErr != nil {
		return r.skinChangeOutput, r.skinChangeErr
	}
	if r.skinChangeOutput.Skin.Width == 0 && r.skinChangeOutput.Skin.Height == 0 {
		return native.PlayerSkinChangeOutput{Cancelled: cancelled, Skin: skin}, nil
	}
	return r.skinChangeOutput, nil
}

func (r *runtimeStub) HandlePlayerTransfer(invocation native.InvocationID, input native.PlayerTransferInput, cancelled bool) (native.PlayerTransferOutput, error) {
	r.transferInput = input
	if r.players != nil {
		if tx, ok := r.players.InvocationTx(invocation); ok {
			r.transferTx = tx.World()
		}
	}
	output := r.transferOutput
	output.Cancelled = output.Cancelled || cancelled
	if output.Address.IP == nil {
		output.Address = input.Address
	}
	return output, r.transferErr
}

func (r *runtimeStub) HandlePlayerCommandExecution(invocation native.InvocationID, input native.PlayerCommandExecutionInput, cancelled bool) (native.PlayerCommandExecutionOutput, error) {
	r.commandInput = input
	if r.players != nil {
		if tx, ok := r.players.InvocationTx(invocation); ok {
			r.commandTx = tx.World()
		}
	}
	output := r.commandOutput
	output.Cancelled = output.Cancelled || cancelled
	if output.Arguments == nil {
		output.Arguments = append([]string(nil), input.Arguments...)
	}
	return output, r.commandErr
}

func (r *runtimeStub) HandlePlayerDiagnostics(invocation native.InvocationID, input native.PlayerDiagnosticsInput) error {
	r.diagnosticsCalled = true
	r.diagnosticsInput = input
	if r.players != nil {
		if tx, ok := r.players.InvocationTx(invocation); ok {
			r.diagnosticsTx = tx.World()
		}
	}
	return r.diagnosticsErr
}

func TestPlayerHandlerChangeWorldFiresOnFirstDestinationTick(t *testing.T) {
	before := world.Config{Synchronous: true}.New()
	after := world.Config{Synchronous: true}.New()
	t.Cleanup(func() {
		_ = before.Close()
		_ = after.Close()
	})
	players := NewPlayers()
	runtime := &runtimeStub{subscriptions: native.PlayerChangeWorldSubscription, players: players}
	resolver := worldResolverStub{before: 11, after: 12}
	before.Handle(NewWorldHandler(players.EntityRegistry(), players, 11))
	after.Handle(NewWorldHandler(players.EntityRegistry(), players, 12))
	handler := NewPlayerHandler(runtime, nil, players, resolver)
	id := uuid.MustParse("05ed0094-34ea-442d-8292-a337bd677bef")
	handle := world.EntitySpawnOpts{ID: id, Position: mgl64.Vec3{1, 2, 3}}.New(
		player.Type,
		player.Config{UUID: id, Name: "Traveller", Position: mgl64.Vec3{1, 2, 3}},
	)
	var transferred *world.EntityHandle
	if err := before.Do(func(tx *world.Tx) {
		connected := tx.AddEntity(handle).(*player.Player)
		players.Register(connected, 94)
		connected.Handle(handler)
		connected.Tick(tx, 1)
		if runtime.changeWorldCalls != 0 {
			t.Fatalf("initial tick emitted %d change-world calls", runtime.changeWorldCalls)
		}
		sameWorld := tx.RemoveEntity(connected)
		if sameWorld == nil {
			t.Fatal("same-world remove returned no handle")
		}
		connected = tx.AddEntity(sameWorld).(*player.Player)
		connected.Tick(tx, 2)
		if runtime.changeWorldCalls != 0 {
			t.Fatalf("same-world re-add emitted %d change-world calls", runtime.changeWorldCalls)
		}
		transferred = tx.RemoveEntity(connected)
		if transferred == nil {
			t.Fatal("remove player returned no handle")
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := after.Do(func(tx *world.Tx) {
		connected := tx.AddEntity(transferred).(*player.Player)
		connected.Tick(tx, 3)
		if runtime.changeWorldCalls != 1 || runtime.changeWorldInput.Before == nil ||
			*runtime.changeWorldInput.Before != 11 || runtime.changeWorldInput.After != 12 {
			t.Fatalf("destination change-world result: calls=%d input=%#v", runtime.changeWorldCalls, runtime.changeWorldInput)
		}
		connected.Tick(tx, 4)
		if runtime.changeWorldCalls != 1 {
			t.Fatalf("second destination tick emitted %d calls", runtime.changeWorldCalls)
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

type worldResolverStub map[*world.World]native.WorldID

func (r worldResolverStub) WorldHandle(w *world.World) (native.WorldID, bool) {
	id, ok := r[w]
	return id, ok
}

func (r worldResolverStub) WorldByHandle(id native.WorldID) (*world.World, bool) {
	for w, candidate := range r {
		if candidate == id {
			return w, true
		}
	}
	return nil, false
}

func TestPlayerHandlerRespawnMutatesPositionAndWorld(t *testing.T) {
	destination := world.Config{Synchronous: true}.New()
	redirected := world.Config{Synchronous: true}.New()
	t.Cleanup(func() {
		_ = destination.Close()
		_ = redirected.Close()
	})
	withPlayerTx(t, func(tx *world.Tx, p *player.Player) {
		players := NewPlayers()
		players.Register(p, 41)
		runtime := &runtimeStub{
			subscriptions: native.PlayerRespawnSubscription,
			players:       players,
			respawnOutput: native.PlayerRespawnOutput{
				Position: native.Vec3{X: 10, Y: 72, Z: -4},
				World:    13,
			},
		}
		handler := NewPlayerHandler(runtime, nil, players, worldResolverStub{destination: 12, redirected: 13})
		position := mgl64.Vec3{1, 64, 2}
		chosenWorld := destination

		handler.HandleRespawn(p, &position, &chosenWorld)

		if runtime.respawnInput.Player.Player.Generation != 41 {
			t.Fatalf("unexpected respawn input: %#v", runtime.respawnInput)
		}
		if runtime.respawnTx != tx.World() {
			t.Fatal("respawn invocation did not use the player's source transaction")
		}
		if position != (mgl64.Vec3{10, 72, -4}) {
			t.Fatalf("unexpected respawn position: %v", position)
		}
		if chosenWorld != redirected {
			t.Fatal("respawn world was not redirected")
		}
	})
}

func TestPlayerHandlerRespawnRejectsUnknownWorldHandle(t *testing.T) {
	destination := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = destination.Close() })
	withPlayerTx(t, func(_ *world.Tx, p *player.Player) {
		players := NewPlayers()
		players.Register(p, 42)
		runtime := &runtimeStub{
			subscriptions: native.PlayerRespawnSubscription,
			respawnOutput: native.PlayerRespawnOutput{
				Position: native.Vec3{X: 10, Y: 72, Z: -4},
				World:    999,
			},
		}
		handler := NewPlayerHandler(runtime, nil, players, worldResolverStub{destination: 12})
		position := mgl64.Vec3{1, 64, 2}
		chosenWorld := destination

		handler.HandleRespawn(p, &position, &chosenWorld)

		if position != (mgl64.Vec3{1, 64, 2}) {
			t.Fatalf("invalid world changed respawn position: %v", position)
		}
		if chosenWorld != destination {
			t.Fatal("invalid world changed respawn destination")
		}
	})
}

func TestPlayerHandlerRespawnRejectsNonFinitePosition(t *testing.T) {
	destination := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = destination.Close() })
	withPlayerTx(t, func(_ *world.Tx, p *player.Player) {
		players := NewPlayers()
		players.Register(p, 43)
		runtime := &runtimeStub{
			subscriptions: native.PlayerRespawnSubscription,
			respawnOutput: native.PlayerRespawnOutput{
				Position: native.Vec3{X: math.NaN(), Y: 72, Z: -4},
				World:    12,
			},
		}
		handler := NewPlayerHandler(runtime, nil, players, worldResolverStub{destination: 12})
		position := mgl64.Vec3{1, 64, 2}
		chosenWorld := destination

		handler.HandleRespawn(p, &position, &chosenWorld)

		if position != (mgl64.Vec3{1, 64, 2}) || chosenWorld != destination {
			t.Fatal("invalid position changed respawn state")
		}
	})
}

func (r *runtimeStub) Subscriptions() uint64 { return r.subscriptions }
func (r *runtimeStub) HandlePlayerMove(_ native.InvocationID, input native.PlayerMoveInput, _ bool) (bool, error) {
	r.moveInput = input
	return r.moveCancelled, nil
}
func (r *runtimeStub) HandlePlayerChat(_ native.InvocationID, input native.PlayerChatInput, _ bool) (native.PlayerChatOutput, error) {
	r.chatInput = input
	return r.chatOutput, nil
}

func withPlayer(t *testing.T, function func(*player.Player)) {
	withPlayerTx(t, func(_ *world.Tx, player *player.Player) { function(player) })
}

func withPlayerTx(t *testing.T, function func(*world.Tx, *player.Player)) {
	t.Helper()
	w := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = w.Close() })
	id := uuid.MustParse("4f62ee78-9519-4f1c-b0bd-69f57b578daf")
	handle := world.EntitySpawnOpts{ID: id, Position: mgl64.Vec3{1, 2, 3}}.New(
		player.Type,
		player.Config{UUID: id, Name: "TestPlayer", Position: mgl64.Vec3{1, 2, 3}},
	)
	if err := w.Do(func(tx *world.Tx) {
		function(tx, tx.AddEntity(handle).(*player.Player))
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestPlayerHandlerTransferForwardsAddress(t *testing.T) {
	withPlayerTx(t, func(tx *world.Tx, p *player.Player) {
		players := NewPlayers()
		runtime := &runtimeStub{
			subscriptions: native.PlayerTransferSubscription,
			players:       players,
			transferOutput: native.PlayerTransferOutput{
				Cancelled: true,
				Address:   native.UDPAddress{IP: net.ParseIP("2001:db8::1").To16(), Port: 19133, Zone: "eth0"},
			},
		}
		players.Register(p, 7)
		p.Handle(NewPlayerHandler(runtime, nil, players, nil))
		if err := p.Transfer("127.0.0.1:19132"); err != nil {
			t.Fatal(err)
		}
		if runtime.transferInput.Player.Player.Generation != 7 || runtime.transferInput.Player.Name != "TestPlayer" {
			t.Fatalf("unexpected player snapshot: %+v", runtime.transferInput.Player)
		}
		if !net.IP(runtime.transferInput.Address.IP).Equal(net.ParseIP("127.0.0.1")) || runtime.transferInput.Address.Port != 19132 {
			t.Fatalf("unexpected transfer address: %+v", runtime.transferInput.Address)
		}
		if runtime.transferTx != tx.World() {
			t.Fatal("transfer invocation did not use the player's transaction")
		}
	})
}

func TestApplyTransferAddress(t *testing.T) {
	destination := &net.UDPAddr{IP: net.IPv4zero, Port: 1}
	source := native.UDPAddress{IP: net.ParseIP("2001:db8::2").To16(), Port: 19133, Zone: "eth0"}
	if !applyTransferAddress(destination, source) {
		t.Fatal("valid transfer address was rejected")
	}
	if !destination.IP.Equal(net.ParseIP("2001:db8::2")) || destination.Port != 19133 || destination.Zone != "eth0" {
		t.Fatalf("unexpected destination: %+v", destination)
	}
	source.IP[0] = 0
	if destination.IP[0] == 0 {
		t.Fatal("destination aliases the native address")
	}
	if applyTransferAddress(destination, native.UDPAddress{IP: []byte{1, 2, 3}, Port: 19132}) {
		t.Fatal("invalid transfer IP was accepted")
	}
}

func TestPlayerHandlerCommandExecutionForwardsIdentityAndArguments(t *testing.T) {
	const name = "native-event-command-test"
	command := cmd.New(name, "command event test", []string{"nect"}, commandExecutionTestRunnable{})
	cmd.Register(command)
	withPlayerTx(t, func(tx *world.Tx, p *player.Player) {
		players := NewPlayers()
		runtime := &runtimeStub{
			subscriptions: native.PlayerCommandExecutionSubscription,
			players:       players,
			commandOutput: native.PlayerCommandExecutionOutput{
				Cancelled: true,
				Arguments: []string{"changed", "arguments"},
			},
		}
		players.Register(p, 9)
		p.Handle(NewPlayerHandler(runtime, nil, players, nil))
		p.ExecuteCommand("/" + name + " first second")
		if runtime.commandInput.Player.Player.Generation != 9 || runtime.commandInput.Command.Name != name ||
			runtime.commandInput.Command.Description != "command event test" || runtime.commandInput.Command.Usage != command.Usage() ||
			!reflect.DeepEqual(runtime.commandInput.Command.Aliases, command.Aliases()) ||
			!reflect.DeepEqual(runtime.commandInput.Arguments, []string{"first", "second"}) {
			t.Fatalf("unexpected command input: %+v", runtime.commandInput)
		}
		if runtime.commandTx != tx.World() {
			t.Fatal("command-execution invocation did not use the player's transaction")
		}
	})
}

func TestApplyCommandArguments(t *testing.T) {
	destination := []string{"first", "second"}
	if !applyCommandArguments(destination, []string{"changed", "arguments"}) || !reflect.DeepEqual(destination, []string{"changed", "arguments"}) {
		t.Fatalf("arguments were not replaced: %v", destination)
	}
	if applyCommandArguments(destination, []string{"wrong-count"}) {
		t.Fatal("argument count change was accepted")
	}
}

func TestPlayerHandlerDiagnosticsForwardsAllFields(t *testing.T) {
	diagnostics := session.Diagnostics{
		AverageFramesPerSecond:        1,
		AverageServerSimTickTime:      2,
		AverageClientSimTickTime:      3,
		AverageBeginFrameTime:         4,
		AverageInputTime:              5,
		AverageRenderTime:             6,
		AverageEndFrameTime:           7,
		AverageRemainderTimePercent:   8,
		AverageUnaccountedTimePercent: 9,
	}
	withPlayerTx(t, func(tx *world.Tx, p *player.Player) {
		players := NewPlayers()
		runtime := &runtimeStub{
			subscriptions: native.PlayerDiagnosticsSubscription,
			players:       players,
		}
		players.Register(p, 11)
		p.Handle(NewPlayerHandler(runtime, nil, players, nil))
		p.UpdateDiagnostics(diagnostics)
		input := runtime.diagnosticsInput
		if !runtime.diagnosticsCalled || input.Player.Player.Generation != 11 || input.AverageFramesPerSecond != 1 ||
			input.AverageServerSimTickTime != 2 || input.AverageClientSimTickTime != 3 || input.AverageBeginFrameTime != 4 ||
			input.AverageInputTime != 5 || input.AverageRenderTime != 6 || input.AverageEndFrameTime != 7 ||
			input.AverageRemainderTimePercent != 8 || input.AverageUnaccountedTimePercent != 9 {
			t.Fatalf("unexpected diagnostics input: %+v", input)
		}
		if runtime.diagnosticsTx != tx.World() {
			t.Fatal("diagnostics invocation did not use the player's transaction")
		}
	})
}

func TestPlayerHandlerMove(t *testing.T) {
	runtime := &runtimeStub{subscriptions: native.PlayerMoveSubscription, moveCancelled: true}
	withPlayer(t, func(p *player.Player) {
		players := NewPlayers()
		players.Register(p, 7)
		handler := NewPlayerHandler(runtime, nil, players, nil)
		p.Handle(handler)
		p.Move(mgl64.Vec3{3, 3, 3}, 90, 10)
		if p.Position() != (mgl64.Vec3{1, 2, 3}) {
			t.Fatalf("cancelled movement changed position: %v", p.Position())
		}
		if runtime.moveInput.Player.Player.Generation != 7 || runtime.moveInput.Player.Name != "TestPlayer" ||
			runtime.moveInput.Player.Position != (native.Vec3{X: 1, Y: 2, Z: 3}) ||
			runtime.moveInput.OldPosition != (native.Vec3{X: 1, Y: 2, Z: 3}) {
			t.Fatalf("unexpected movement input: %+v", runtime.moveInput)
		}
	})
}

func TestPlayerHandlerSkinChangeMutatesAndCancelsCandidate(t *testing.T) {
	for _, test := range []struct {
		name      string
		cancelled bool
		fail      bool
		invalid   bool
		wantModel string
	}{
		{name: "allowed", wantModel: "geometry.changed"},
		{name: "cancelled", cancelled: true, wantModel: "geometry.old"},
		{name: "failure after cancellation", cancelled: true, fail: true, wantModel: "geometry.old"},
		{name: "failure fails open", fail: true, wantModel: "geometry.proposed"},
		{name: "invalid skin keeps cancellation", cancelled: true, invalid: true, wantModel: "geometry.old"},
	} {
		t.Run(test.name, func(t *testing.T) {
			withPlayer(t, func(p *player.Player) {
				old := playerskin.New(64, 64)
				old.ModelConfig.Default = "geometry.old"
				p.SetSkin(old)

				changed := playerskin.New(64, 64)
				changed.ModelConfig.Default = "geometry.changed"
				output, ok := playerSkinToNative(changed)
				if !ok {
					t.Fatal("test skin did not convert")
				}
				runtime := &runtimeStub{
					subscriptions: native.PlayerSkinChangeSubscription,
					skinChangeOutput: native.PlayerSkinChangeOutput{
						Cancelled: test.cancelled,
						Skin:      output,
					},
				}
				if test.fail {
					runtime.skinChangeErr = errors.New("plugin failed")
				}
				if test.invalid {
					runtime.skinChangeOutput.Skin = native.PlayerSkin{Width: 1, Height: 1}
				}
				players := NewPlayers()
				playerID := players.Register(p, 95)
				p.Handle(NewPlayerHandler(runtime, nil, players, nil))

				proposed := playerskin.New(64, 64)
				proposed.FullID = "proposed"
				proposed.ModelConfig.Default = "geometry.proposed"
				p.SetSkin(proposed)
				if !runtime.skinChangeCalled || runtime.skinChangeInput.Player.Player != playerID {
					t.Fatalf("skin-change input = %#v", runtime.skinChangeInput)
				}
				if got := p.Skin().ModelConfig.Default; got != test.wantModel {
					t.Fatalf("committed model = %q, want %q", got, test.wantModel)
				}
			})
		})
	}
}

func TestPlayerHandlerAttackEntityUsesStableTargetID(t *testing.T) {
	runtime := &runtimeStub{
		subscriptions: native.PlayerAttackEntitySubscription,
		attackOutput: native.PlayerAttackEntityOutput{
			Cancelled: true, KnockbackForce: -0.25, KnockbackHeight: 0.9, Critical: true,
		},
	}
	withPlayerTx(t, func(tx *world.Tx, p *player.Player) {
		players := NewPlayers()
		playerID := players.Register(p, 91)
		handler := NewPlayerHandler(runtime, nil, players, nil)
		p.Handle(handler)
		target := tx.AddEntity(entity.NewText("target", p.Position().Add(mgl64.Vec3{1, 0, 0})))
		if p.AttackEntity(target) {
			t.Fatal("cancelled attack succeeded")
		}
		targetID, ok := players.EntityRegistry().ID(target)
		if !ok || runtime.attackInput.Player.Player != playerID || runtime.attackInput.Target != targetID {
			t.Fatalf("attack input = %#v, player=%#v target=%#v", runtime.attackInput, playerID, targetID)
		}
	})
}

func TestPlayerHandlerItemUseOnEntityUsesStableTargetID(t *testing.T) {
	runtime := &runtimeStub{
		subscriptions:       native.PlayerItemUseOnEntitySubscription,
		itemUseEntityCancel: true,
	}
	withPlayerTx(t, func(tx *world.Tx, p *player.Player) {
		players := NewPlayers()
		playerID := players.Register(p, 92)
		handler := NewPlayerHandler(runtime, nil, players, nil)
		p.Handle(handler)
		target := tx.AddEntity(entity.NewText("target", p.Position().Add(mgl64.Vec3{1, 0, 0})))
		if p.UseItemOnEntity(target) {
			t.Fatal("cancelled item use succeeded")
		}
		targetID, ok := players.EntityRegistry().ID(target)
		if !ok || runtime.itemUseEntityInput.Player.Player != playerID || runtime.itemUseEntityInput.Target != targetID {
			t.Fatalf("item-use-on-entity input = %#v, player=%#v target=%#v", runtime.itemUseEntityInput, playerID, targetID)
		}
	})
}

func TestPlayerHandlerChangeWorldUsesManagedHandlesAndNewWorldTransaction(t *testing.T) {
	before := world.Config{Synchronous: true}.New()
	after := world.Config{Synchronous: true}.New()
	t.Cleanup(func() {
		_ = before.Close()
		_ = after.Close()
	})
	players := NewPlayers()
	runtime := &runtimeStub{subscriptions: native.PlayerChangeWorldSubscription, players: players}
	resolver := worldResolverStub{before: 11, after: 12}
	id := uuid.MustParse("d06ad8ec-a3c0-4c91-95e0-dbd3643778d5")
	handle := world.EntitySpawnOpts{ID: id, Position: mgl64.Vec3{1, 2, 3}}.New(
		player.Type,
		player.Config{UUID: id, Name: "Traveller", Position: mgl64.Vec3{1, 2, 3}},
	)
	if err := after.Do(func(tx *world.Tx) {
		p := tx.AddEntity(handle).(*player.Player)
		playerID := players.Register(p, 93)
		handler := NewPlayerHandler(runtime, nil, players, resolver)
		handler.HandleChangeWorld(p, before, after)
		if runtime.changeWorldInput.Player.Player != playerID || runtime.changeWorldInput.Before == nil ||
			*runtime.changeWorldInput.Before != 11 || runtime.changeWorldInput.After != 12 {
			t.Fatalf("change-world input = %#v", runtime.changeWorldInput)
		}
		if runtime.changeWorldTx != after {
			t.Fatalf("invocation world = %p, want %p", runtime.changeWorldTx, after)
		}
		players.recordWorldDeparture(p, 11)
		delete(resolver, before)
		handler.HandleChangeWorld(p, before, after)
		if runtime.changeWorldInput.Before == nil || *runtime.changeWorldInput.Before != 11 {
			t.Fatalf("unloaded source handle = %#v", runtime.changeWorldInput.Before)
		}
		handler.HandleChangeWorld(p, nil, after)
		if runtime.changeWorldInput.Before != nil || runtime.changeWorldInput.After != 12 {
			t.Fatalf("initial change-world input = %#v", runtime.changeWorldInput)
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestPlayerHandlerChat(t *testing.T) {
	replacement := "filtered"
	runtime := &runtimeStub{
		subscriptions: native.PlayerChatSubscription,
		chatOutput:    native.PlayerChatOutput{Cancelled: true, Replacement: &replacement},
	}
	withPlayer(t, func(p *player.Player) {
		players := NewPlayers()
		players.Register(p, 9)
		handler := NewPlayerHandler(runtime, nil, players, nil)
		p.Handle(handler)
		p.Chat("original")
		if runtime.chatInput.Message != "original" || runtime.chatInput.Player.Player.Generation != 9 {
			t.Fatalf("unexpected chat input: %+v", runtime.chatInput)
		}
	})
}

func TestPlayerHandlerJoinAndQuit(t *testing.T) {
	runtime := &runtimeStub{
		subscriptions: native.PlayerJoinSubscription | native.PlayerQuitSubscription,
		joinCancelled: true,
	}
	withPlayer(t, func(p *player.Player) {
		players := NewPlayers()
		players.Register(p, 11)
		handler := NewPlayerHandler(runtime, nil, players, nil)
		if !handler.Join(p) {
			t.Fatal("join was not cancelled")
		}
		handler.HandleQuit(p)
		if runtime.joinInput.Player.Player.Generation != 11 || runtime.quitInput.Player.Player.Generation != 11 {
			t.Fatalf("join=%+v quit=%+v", runtime.joinInput, runtime.quitInput)
		}
		if _, ok := players.ID(p); ok {
			t.Fatal("player remained registered after quit")
		}
	})
}

func TestPlayerHandlerHurtAndHeal(t *testing.T) {
	runtime := &runtimeStub{
		subscriptions: native.PlayerHurtSubscription | native.PlayerHealSubscription,
		hurtOutput: native.PlayerHurtOutput{
			Cancelled:      true,
			Damage:         2.5,
			AttackImmunity: 750 * time.Millisecond,
		},
		healOutput: native.PlayerHealOutput{Health: 3.5},
	}
	withPlayer(t, func(p *player.Player) {
		players := NewPlayers()
		players.Register(p, 13)
		handler := NewPlayerHandler(runtime, nil, players, nil)
		p.Hurt(5, testDamageSource{})
		p.Handle(handler)
		before := p.Health()
		p.Hurt(8, testDamageSource{})
		if p.Health() != before {
			t.Fatalf("cancelled hurt changed health: %v -> %v", before, p.Health())
		}
		p.Heal(1, testHealingSource{})
		if p.Health() != before+3.5 {
			t.Fatalf("modified heal = %v, want %v", p.Health(), before+3.5)
		}
		if runtime.hurtInput.Player.Player.Generation != 13 || runtime.healInput.Player.Player.Generation != 13 {
			t.Fatalf("hurt=%+v heal=%+v", runtime.hurtInput, runtime.healInput)
		}
		if !strings.Contains(runtime.hurtInput.Source.Name, "testDamageSource") ||
			!runtime.hurtInput.Source.ReducedByArmour || runtime.hurtInput.Source.ReducedByResistance ||
			!runtime.hurtInput.Source.Fire || !runtime.hurtInput.Source.IgnoresTotem {
			t.Fatalf("damage source = %+v", runtime.hurtInput.Source)
		}
		if !strings.Contains(runtime.healInput.Source.Name, "testHealingSource") {
			t.Fatalf("healing source = %+v", runtime.healInput.Source)
		}
		attack := handler.nativeDamageSource(entity.AttackDamageSource{Attacker: p})
		if attack.Kind != native.DamageSourceAttack || attack.Entity.Generation != 13 {
			t.Fatalf("attack source = %+v", attack)
		}
		attack = handler.nativeDamageSource(entity.AttackDamageSource{})
		if attack.Kind != native.DamageSourceAttack || attack.Entity.Generation != 0 {
			t.Fatalf("attack source without attacker = %+v", attack)
		}
		pointerAttack := entity.AttackDamageSource{Attacker: p}
		attack = handler.nativeDamageSource(&pointerAttack)
		if attack.Kind != native.DamageSourceAttack || attack.Entity.Generation != 13 {
			t.Fatalf("pointer attack source = %+v", attack)
		}
		food := handler.nativeHealingSource(entity.FoodHealingSource{QuickRegeneration: true})
		if food.Kind != native.HealingSourceFood || !food.Data {
			t.Fatalf("food source = %+v", food)
		}
		pointerFood := entity.FoodHealingSource{QuickRegeneration: true}
		food = handler.nativeHealingSource(&pointerFood)
		if food.Kind != native.HealingSourceFood || !food.Data {
			t.Fatalf("pointer food source = %+v", food)
		}
	})
}

func TestPlayerHandlerBlockBreakAndPlace(t *testing.T) {
	runtime := &runtimeStub{
		subscriptions:       native.PlayerBlockBreakSubscription | native.PlayerBlockPlaceSubscription,
		blockBreakOutput:    native.PlayerBlockBreakOutput{Experience: 7},
		blockPlaceCancelled: true,
	}
	withPlayerTx(t, func(tx *world.Tx, p *player.Player) {
		players := NewPlayers()
		players.Register(p, 14)
		handler := NewPlayerHandler(runtime, nil, players, nil)
		breakPos, placePos := cube.Pos{1, 2, 4}, cube.Pos{1, 2, 5}
		tx.SetBlock(breakPos, block.Stone{}, nil)
		p.Handle(handler)
		p.BreakBlock(breakPos)
		p.PlaceBlock(placePos, block.Dirt{}, nil)
		if runtime.blockBreakInput.Position != (native.BlockPos{X: 1, Y: 2, Z: 4}) || runtime.blockPlaceInput.Position != (native.BlockPos{X: 1, Y: 2, Z: 5}) {
			t.Fatalf("break=%+v place=%+v", runtime.blockBreakInput, runtime.blockPlaceInput)
		}
		if runtime.blockBreakInput.Block.Identifier != "minecraft:stone" || len(runtime.blockBreakInput.Block.PropertiesNBT) == 0 ||
			runtime.blockPlaceInput.Block.Identifier != "minecraft:dirt" || len(runtime.blockPlaceInput.Block.PropertiesNBT) == 0 {
			t.Fatalf("full block payloads: break=%+v place=%+v", runtime.blockBreakInput.Block, runtime.blockPlaceInput.Block)
		}
		if _, ok := tx.Block(placePos).(block.Air); !ok {
			t.Fatalf("cancelled placement wrote %T", tx.Block(placePos))
		}
	})
}

func TestPlayerHandlerFoodLossAndDeath(t *testing.T) {
	runtime := &runtimeStub{
		subscriptions:  native.PlayerFoodLossSubscription | native.PlayerDeathSubscription,
		foodLossOutput: native.PlayerFoodLossOutput{Cancelled: true, To: 8},
		keepInventory:  true,
	}
	withPlayerTx(t, func(tx *world.Tx, p *player.Player) {
		players := NewPlayers()
		players.Register(p, 15)
		handler := NewPlayerHandler(runtime, nil, players, nil)
		tx.World().SetDifficulty(world.DifficultyHard)
		p.SetFood(10)
		p.Saturate(0, -20)
		p.Handle(handler)
		p.Exhaust(4)
		if p.Food() != 10 {
			t.Fatalf("cancelled food loss changed food to %d", p.Food())
		}
		keepInventory := false
		handler.HandleDeath(p, nil, &keepInventory)
		if !keepInventory {
			t.Fatal("keep inventory was not applied")
		}
		if runtime.foodLossInput.Player.Player.Generation != 15 || runtime.deathInput.Player.Player.Generation != 15 {
			t.Fatalf("food=%+v death=%+v", runtime.foodLossInput, runtime.deathInput)
		}
	})
}
