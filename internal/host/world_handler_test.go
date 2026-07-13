package host

import (
	"context"
	"testing"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
)

func TestWorldHandlerRecordsPlayerDeparture(t *testing.T) {
	players := NewPlayers()
	w := world.Config{Synchronous: true}.New()
	w.Handle(NewWorldHandler(players.EntityRegistry(), players, native.WorldID(44)))
	t.Cleanup(func() { _ = w.Close() })
	id := uuid.MustParse("65c2514b-f068-4da5-a403-1d5890c8f2a7")
	handle := world.EntitySpawnOpts{ID: id, Position: mgl64.Vec3{1, 2, 3}}.New(
		player.Type,
		player.Config{UUID: id, Name: "Traveller", Position: mgl64.Vec3{1, 2, 3}},
	)
	if err := w.Do(func(tx *world.Tx) {
		connected := tx.AddEntity(handle).(*player.Player)
		players.Register(connected, 7)
		if tx.RemoveEntity(connected) == nil {
			t.Fatal("remove player returned no handle")
		}
		departure, ok := players.takeWorldDeparture(connected)
		if !ok || departure != 44 {
			t.Fatalf("departure = %d, %v; want 44, true", departure, ok)
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}
