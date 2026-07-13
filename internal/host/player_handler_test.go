package host

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
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
	attackInput         native.PlayerAttackEntityInput
	attackOutput        native.PlayerAttackEntityOutput
	itemUseEntityInput  native.PlayerItemUseOnEntityInput
	itemUseEntityCancel bool
}

type testDamageSource struct{}

func (testDamageSource) ReducedByArmour() bool     { return true }
func (testDamageSource) ReducedByResistance() bool { return false }
func (testDamageSource) Fire() bool                { return true }
func (testDamageSource) IgnoreTotem() bool         { return true }

type testHealingSource struct{}

func (testHealingSource) HealingSource() {}

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
func (r *runtimeStub) HandlePlayerJump(_ native.InvocationID, _ native.PlayerID) error { return nil }
func (r *runtimeStub) HandlePlayerTeleport(_ native.InvocationID, _ native.PlayerTeleportInput, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerExperienceGain(_ native.InvocationID, _ native.PlayerID, amount int, cancelled bool) (native.PlayerExperienceGainOutput, error) {
	return native.PlayerExperienceGainOutput{Cancelled: cancelled, Amount: amount}, nil
}
func (r *runtimeStub) HandlePlayerPunchAir(_ native.InvocationID, _ native.PlayerID, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerHeldSlotChange(_ native.InvocationID, _ native.PlayerHeldSlotChangeInput, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerSleep(_ native.InvocationID, _ native.PlayerID, sendReminder, cancelled bool) (native.PlayerSleepOutput, error) {
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
func (r *runtimeStub) HandlePlayerItemUse(_ native.InvocationID, _ native.PlayerID, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerItemUseOnBlock(_ native.InvocationID, _ native.PlayerItemUseOnBlockInput, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerItemConsume(_ native.InvocationID, _ native.PlayerID, _ native.ItemStack, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerItemRelease(_ native.InvocationID, _ native.PlayerID, _ native.ItemStack, _ time.Duration, cancelled bool) (bool, error) {
	return cancelled, nil
}
func (r *runtimeStub) HandlePlayerItemDamage(_ native.InvocationID, _ native.PlayerID, _ native.ItemStack, damage int, cancelled bool) (native.PlayerItemDamageOutput, error) {
	return native.PlayerItemDamageOutput{Cancelled: cancelled, Damage: damage}, nil
}
func (r *runtimeStub) HandlePlayerItemDrop(_ native.InvocationID, _ native.PlayerID, _ native.ItemStack, cancelled bool) (bool, error) {
	return cancelled, nil
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

func TestPlayerHandlerMove(t *testing.T) {
	runtime := &runtimeStub{subscriptions: native.PlayerMoveSubscription, moveCancelled: true}
	withPlayer(t, func(p *player.Player) {
		players := NewPlayers()
		players.Register(p, 7)
		handler := NewPlayerHandler(runtime, nil, players)
		p.Handle(handler)
		p.Move(mgl64.Vec3{3, 3, 3}, 90, 10)
		if p.Position() != (mgl64.Vec3{1, 2, 3}) {
			t.Fatalf("cancelled movement changed position: %v", p.Position())
		}
		if runtime.moveInput.Player.Generation != 7 || runtime.moveInput.OldPosition != (native.Vec3{X: 1, Y: 2, Z: 3}) {
			t.Fatalf("unexpected movement input: %+v", runtime.moveInput)
		}
	})
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
		handler := NewPlayerHandler(runtime, nil, players)
		p.Handle(handler)
		target := tx.AddEntity(entity.NewText("target", p.Position().Add(mgl64.Vec3{1, 0, 0})))
		if p.AttackEntity(target) {
			t.Fatal("cancelled attack succeeded")
		}
		targetID, ok := players.EntityRegistry().ID(target)
		if !ok || runtime.attackInput.Player != playerID || runtime.attackInput.Target != targetID {
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
		handler := NewPlayerHandler(runtime, nil, players)
		p.Handle(handler)
		target := tx.AddEntity(entity.NewText("target", p.Position().Add(mgl64.Vec3{1, 0, 0})))
		if p.UseItemOnEntity(target) {
			t.Fatal("cancelled item use succeeded")
		}
		targetID, ok := players.EntityRegistry().ID(target)
		if !ok || runtime.itemUseEntityInput.Player != playerID || runtime.itemUseEntityInput.Target != targetID {
			t.Fatalf("item-use-on-entity input = %#v, player=%#v target=%#v", runtime.itemUseEntityInput, playerID, targetID)
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
		p.Handle(handler)
		p.Chat("original")
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
		if runtime.hurtInput.Player.Generation != 13 || runtime.healInput.Player.Generation != 13 {
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
		handler := NewPlayerHandler(runtime, nil, players)
		breakPos, placePos := cube.Pos{1, 2, 4}, cube.Pos{1, 2, 5}
		tx.SetBlock(breakPos, block.Stone{}, nil)
		p.Handle(handler)
		p.BreakBlock(breakPos)
		p.PlaceBlock(placePos, block.Dirt{}, nil)
		if runtime.blockBreakInput.Position != (native.BlockPos{X: 1, Y: 2, Z: 4}) || runtime.blockPlaceInput.Position != (native.BlockPos{X: 1, Y: 2, Z: 5}) {
			t.Fatalf("break=%+v place=%+v", runtime.blockBreakInput, runtime.blockPlaceInput)
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
		handler := NewPlayerHandler(runtime, nil, players)
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
		if runtime.foodLossInput.Player.Generation != 15 || runtime.deathInput.Player.Generation != 15 {
			t.Fatalf("food=%+v death=%+v", runtime.foodLossInput, runtime.deathInput)
		}
	})
}
