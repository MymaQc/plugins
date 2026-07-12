package host

import (
	"testing"
	"time"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
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
}

func (r *runtimeStub) HandlePlayerJoin(input native.PlayerJoinInput, _ bool) (bool, error) {
	r.joinInput = input
	return r.joinCancelled, nil
}
func (r *runtimeStub) HandlePlayerQuit(input native.PlayerQuitInput) error {
	r.quitInput = input
	return nil
}
func (r *runtimeStub) HandlePlayerHurt(input native.PlayerHurtInput, _ bool) (native.PlayerHurtOutput, error) {
	r.hurtInput = input
	return r.hurtOutput, nil
}
func (r *runtimeStub) HandlePlayerHeal(input native.PlayerHealInput, _ bool) (native.PlayerHealOutput, error) {
	r.healInput = input
	return r.healOutput, nil
}
func (r *runtimeStub) HandlePlayerBlockBreak(input native.PlayerBlockBreakInput, _ bool) (native.PlayerBlockBreakOutput, error) {
	r.blockBreakInput = input
	return r.blockBreakOutput, nil
}
func (r *runtimeStub) HandlePlayerBlockPlace(input native.PlayerBlockPlaceInput, _ bool) (bool, error) {
	r.blockPlaceInput = input
	return r.blockPlaceCancelled, nil
}
func (r *runtimeStub) HandlePlayerFoodLoss(input native.PlayerFoodLossInput, _ bool) (native.PlayerFoodLossOutput, error) {
	r.foodLossInput = input
	return r.foodLossOutput, nil
}
func (r *runtimeStub) HandlePlayerDeath(input native.PlayerDeathInput, _ bool) (bool, error) {
	r.deathInput = input
	return r.keepInventory, nil
}
func (r *runtimeStub) HandlePlayerStartBreak(_ native.PlayerPositionInput, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerFireExtinguish(_ native.PlayerPositionInput, cancelled bool) (bool, error) {
	return cancelled, nil
}

func (r *runtimeStub) Subscriptions() uint64 { return r.subscriptions }
func (r *runtimeStub) HandlePlayerMove(input native.PlayerMoveInput, _ bool) (bool, error) {
	r.moveInput = input
	return r.moveCancelled, nil
}
func (r *runtimeStub) HandlePlayerChat(input native.PlayerChatInput, _ bool) (native.PlayerChatOutput, error) {
	r.chatInput = input
	return r.chatOutput, nil
}

func withPlayer(t *testing.T, function func(*player.Player)) {
	t.Helper()
	w := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = w.Close() })
	id := uuid.MustParse("4f62ee78-9519-4f1c-b0bd-69f57b578daf")
	handle := world.EntitySpawnOpts{ID: id, Position: mgl64.Vec3{1, 2, 3}}.New(
		player.Type,
		player.Config{UUID: id, Name: "TestPlayer", Position: mgl64.Vec3{1, 2, 3}},
	)
	<-w.Exec(func(tx *world.Tx) {
		function(tx.AddEntity(handle).(*player.Player))
	})
}

func TestPlayerHandlerMove(t *testing.T) {
	runtime := &runtimeStub{subscriptions: native.PlayerMoveSubscription, moveCancelled: true}
	withPlayer(t, func(p *player.Player) {
		players := NewPlayers()
		players.Register(p, 7)
		handler := NewPlayerHandler(runtime, nil, players)
		ctx := event.C(p)
		handler.HandleMove(ctx, mgl64.Vec3{4, 5, 6}, [2]float64{90, 10})
		if !ctx.Cancelled() {
			t.Fatal("movement was not cancelled")
		}
		if runtime.moveInput.Player.Generation != 7 || runtime.moveInput.OldPosition != (native.Vec3{X: 1, Y: 2, Z: 3}) {
			t.Fatalf("unexpected movement input: %+v", runtime.moveInput)
		}
	})
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
		handler := NewPlayerHandler(runtime, nil, players)
		ctx := event.C(p)
		message := "original"
		handler.HandleChat(ctx, &message)
		if !ctx.Cancelled() || message != replacement {
			t.Fatalf("cancelled=%v message=%q", ctx.Cancelled(), message)
		}
		if runtime.chatInput.Message != "original" || runtime.chatInput.Player.Generation != 9 {
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
		handler := NewPlayerHandler(runtime, nil, players)
		if !handler.Join(p) {
			t.Fatal("join was not cancelled")
		}
		handler.HandleQuit(p)
		if runtime.joinInput.Player.Generation != 11 || runtime.quitInput.Player.Generation != 11 {
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
		handler := NewPlayerHandler(runtime, nil, players)
		hurtContext := event.C(p)
		damage, immunity := 8.0, 500*time.Millisecond
		handler.HandleHurt(hurtContext, &damage, false, &immunity, nil)
		if !hurtContext.Cancelled() || damage != 2.5 || immunity != 750*time.Millisecond {
			t.Fatalf("cancelled=%v damage=%v immunity=%v", hurtContext.Cancelled(), damage, immunity)
		}
		healContext := event.C(p)
		health := 1.0
		handler.HandleHeal(healContext, &health, nil)
		if healContext.Cancelled() || health != 3.5 {
			t.Fatalf("cancelled=%v health=%v", healContext.Cancelled(), health)
		}
		if runtime.hurtInput.Player.Generation != 13 || runtime.healInput.Player.Generation != 13 {
			t.Fatalf("hurt=%+v heal=%+v", runtime.hurtInput, runtime.healInput)
		}
	})
}

func TestPlayerHandlerBlockBreakAndPlace(t *testing.T) {
	runtime := &runtimeStub{
		subscriptions:       native.PlayerBlockBreakSubscription | native.PlayerBlockPlaceSubscription,
		blockBreakOutput:    native.PlayerBlockBreakOutput{Experience: 7},
		blockPlaceCancelled: true,
	}
	withPlayer(t, func(p *player.Player) {
		players := NewPlayers()
		players.Register(p, 14)
		handler := NewPlayerHandler(runtime, nil, players)
		breakContext := event.C(p)
		drops := []item.Stack{}
		experience := 2
		handler.HandleBlockBreak(breakContext, cube.Pos{1, 2, 3}, &drops, &experience)
		if breakContext.Cancelled() || experience != 7 {
			t.Fatalf("cancelled=%v experience=%d", breakContext.Cancelled(), experience)
		}
		placeContext := event.C(p)
		handler.HandleBlockPlace(placeContext, cube.Pos{4, 5, 6}, nil)
		if !placeContext.Cancelled() {
			t.Fatal("block place was not cancelled")
		}
		if runtime.blockBreakInput.Position != (native.BlockPos{X: 1, Y: 2, Z: 3}) || runtime.blockPlaceInput.Position != (native.BlockPos{X: 4, Y: 5, Z: 6}) {
			t.Fatalf("break=%+v place=%+v", runtime.blockBreakInput, runtime.blockPlaceInput)
		}
	})
}

func TestPlayerHandlerFoodLossAndDeath(t *testing.T) {
	runtime := &runtimeStub{
		subscriptions:  native.PlayerFoodLossSubscription | native.PlayerDeathSubscription,
		foodLossOutput: native.PlayerFoodLossOutput{Cancelled: true, To: 8},
		keepInventory:  true,
	}
	withPlayer(t, func(p *player.Player) {
		players := NewPlayers()
		players.Register(p, 15)
		handler := NewPlayerHandler(runtime, nil, players)
		foodContext := event.C(p)
		food := 9
		handler.HandleFoodLoss(foodContext, 10, &food)
		if !foodContext.Cancelled() || food != 8 {
			t.Fatalf("cancelled=%v food=%d", foodContext.Cancelled(), food)
		}
		keepInventory := false
		handler.HandleDeath(p, nil, &keepInventory)
		if !keepInventory {
			t.Fatal("keep inventory was not applied")
		}
		if runtime.foodLossInput.Player.Generation != 15 || runtime.deathInput.Player.Generation != 15 {
			t.Fatalf("food=%+v death=%+v", runtime.foodLossInput, runtime.deathInput)
		}
	})
}
