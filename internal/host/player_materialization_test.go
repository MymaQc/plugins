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

func TestPlayersEntityPlayerReturnsFreshTransactionSnapshot(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = w.Close() })
	playerUUID := uuid.New()
	handle := world.EntitySpawnOpts{ID: playerUUID}.New(
		player.Type,
		player.Config{UUID: playerUUID, Name: "Materialized", Position: mgl64.Vec3{1, 64, 2}},
	)
	players := NewPlayers()
	var playerID native.PlayerID
	if err := w.Do(func(tx *world.Tx) {
		playerID = players.Register(tx.AddEntity(handle).(*player.Player), 7)
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	entityID := native.EntityID{UUID: playerID.UUID, Generation: playerID.Generation}
	var invocation native.InvocationID
	if err := w.Do(func(tx *world.Tx) {
		connected, ok := playerInTransaction(handle, tx)
		if !ok {
			t.Fatal("resolve player in transaction")
		}
		connected.Teleport(mgl64.Vec3{9, 70, -4})
		var end func()
		invocation, end = players.BeginInvocation(tx)
		defer end()
		snapshot, ok := players.EntityPlayer(invocation, entityID)
		if !ok {
			t.Fatal("materialize registered player")
		}
		if snapshot.Player != playerID || snapshot.Name != "Materialized" ||
			snapshot.Position != (native.Vec3{X: 9, Y: 70, Z: -4}) {
			t.Fatalf("snapshot = %+v", snapshot)
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := players.EntityPlayer(invocation, entityID); ok {
		t.Fatal("expired invocation materialized player")
	}
	if _, ok := players.EntityPlayer(0, native.EntityID{UUID: entityID.UUID, Generation: entityID.Generation + 1}); ok {
		t.Fatal("stale entity generation materialized player")
	}
	snapshot, ok := players.EntityPlayer(0, entityID)
	if !ok || snapshot.Position != (native.Vec3{X: 9, Y: 70, Z: -4}) {
		t.Fatalf("fresh owner-world snapshot = %+v, %v", snapshot, ok)
	}
}
