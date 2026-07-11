package host

import (
	"testing"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
)

type runtimeStub struct {
	subscriptions uint64
	moveCancelled bool
	chatOutput    native.PlayerChatOutput
	moveInput     native.PlayerMoveInput
	chatInput     native.PlayerChatInput
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
