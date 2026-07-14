package host

import (
	"context"
	"testing"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
)

func TestEntitiesKeepStableGenerationAndResolveExactTransaction(t *testing.T) {
	other := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = other.Close() })
	withPlayerTx(t, func(tx *world.Tx, connected *player.Player) {
		entities := NewEntities()
		first := entities.Register(connected)
		if first.Generation == 0 {
			t.Fatal("zero entity generation")
		}
		if same := entities.Register(connected); same != first {
			t.Fatalf("same handle changed identity: %+v to %+v", first, same)
		}
		resolved, ok := entities.Resolve(first, tx)
		if !ok || resolved.H() != connected.H() {
			t.Fatalf("entity did not resolve in owner transaction: %T %v", resolved, ok)
		}
		if err := other.Do(func(otherTx *world.Tx) {
			if _, ok := entities.Resolve(first, otherTx); ok {
				t.Fatal("entity resolved in a different world's transaction")
			}
		}).Wait(context.Background()); err != nil {
			t.Fatal(err)
		}

		entities.deactivateHandle(connected.H())
		if _, ok := entities.Handle(first); ok {
			t.Fatal("worldless entity handle remained active")
		}
		if reactivated := entities.Register(connected); reactivated != first {
			t.Fatalf("world transfer changed identity: %+v to %+v", first, reactivated)
		}

		entities.Unregister(connected)
		if _, ok := entities.Resolve(first, tx); ok {
			t.Fatal("expired entity ID resolved")
		}
		second := entities.Register(connected)
		if second.UUID != first.UUID || second.Generation <= first.Generation {
			t.Fatalf("re-registration did not advance generation: %+v to %+v", first, second)
		}
		if _, ok := entities.Resolve(first, tx); ok {
			t.Fatal("old generation aliased re-registered entity")
		}
		if _, ok := entities.Resolve(second, tx); !ok {
			t.Fatal("new generation did not resolve")
		}
	})
}

func TestPlayersShareIdentityWithEntityRegistry(t *testing.T) {
	withPlayerTx(t, func(tx *world.Tx, connected *player.Player) {
		players := NewPlayers()
		playerID := players.Register(connected, 41)
		entityID, ok := players.EntityRegistry().ID(connected)
		if !ok || entityID.UUID != playerID.UUID || entityID.Generation != playerID.Generation {
			t.Fatalf("player ID %+v and entity ID %+v differ (ok=%v)", playerID, entityID, ok)
		}
		invocation, end := players.BeginInvocation(tx)
		defer end()
		resolved, ok := players.ResolveEntityID(entityID, invocation)
		if !ok || resolved.H() != connected.H() {
			t.Fatalf("shared entity ID did not resolve: %T %v", resolved, ok)
		}
		players.Unregister(connected)
		if _, ok := players.EntityRegistry().Handle(entityID); ok {
			t.Fatal("unregistered player remained in entity registry")
		}
	})
}

func TestEntitiesKeepStableHandleIdentityAcrossTransfer(t *testing.T) {
	withPlayerTx(t, func(tx *world.Tx, connected *player.Player) {
		entities := NewEntities()
		entityID := entities.Register(connected)
		handleID, ok := entities.EntityHandleID(entityID)
		if !ok || !handleID.Valid() {
			t.Fatalf("handle ID = %#v, ok=%v", handleID, ok)
		}
		if handle, ok := entities.HandleByID(handleID); !ok || handle != connected.H() {
			t.Fatalf("handle = %p, ok=%v", handle, ok)
		}
		if resolved, ok := entities.ResolveHandle(handleID, tx); !ok || resolved.H() != connected.H() {
			t.Fatalf("resolved handle = %T, ok=%v", resolved, ok)
		}

		entities.deactivateHandle(connected.H())
		if _, ok := entities.Handle(entityID); ok {
			t.Fatal("transferring entity ID remained active")
		}
		if transferringID, ok := entities.EntityHandleID(entityID); !ok || transferringID != handleID {
			t.Fatalf("transferring handle ID = %#v, ok=%v", transferringID, ok)
		}
		if reactivated := entities.Register(connected); reactivated != entityID {
			t.Fatalf("transfer changed entity ID: %#v to %#v", entityID, reactivated)
		}
		if reactivatedHandleID, ok := entities.EntityHandleID(entityID); !ok || reactivatedHandleID != handleID {
			t.Fatalf("transfer changed handle ID: %#v to %#v", handleID, reactivatedHandleID)
		}
	})
}

func TestEntitiesDetachStalesEntityIDAndPreservesHandleID(t *testing.T) {
	withPlayerTx(t, func(_ *world.Tx, connected *player.Player) {
		entities := NewEntities()
		first := entities.Register(connected)
		before, ok := entities.EntityHandleID(first)
		if !ok {
			t.Fatal("registered entity has no handle ID")
		}
		detached, handle, ok := entities.Detach(first)
		if !ok || detached != before || handle != connected.H() {
			t.Fatalf("detach = %#v, %p, %v; want %#v, %p, true", detached, handle, ok, before, connected.H())
		}
		if _, ok := entities.Handle(first); ok {
			t.Fatal("detached entity ID remained live")
		}
		if _, ok := entities.EntityHandleID(first); ok {
			t.Fatal("detached entity ID still addressed handle")
		}
		if detachedHandle, ok := entities.HandleByID(detached); !ok || detachedHandle != connected.H() {
			t.Fatalf("detached handle = %p, ok=%v", detachedHandle, ok)
		}
		if detachedHandle, ok := entities.DetachedHandle(detached); !ok || detachedHandle != connected.H() {
			t.Fatalf("worldless handle = %p, ok=%v", detachedHandle, ok)
		}

		second := entities.Register(connected)
		if second == first || second.Generation <= first.Generation {
			t.Fatalf("detached entity ID revived: %#v to %#v", first, second)
		}
		if after, ok := entities.EntityHandleID(second); !ok || after != before {
			t.Fatalf("stable handle ID changed: %#v to %#v, ok=%v", before, after, ok)
		}
		if _, ok := entities.Handle(first); ok {
			t.Fatal("old entity ID revived after activation")
		}
		if _, ok := entities.DetachedHandle(detached); ok {
			t.Fatal("active handle remained detached")
		}
	})
}

func TestEntitiesUnregisterExpiresHandleWithoutRetainingDragonflyHandle(t *testing.T) {
	withPlayerTx(t, func(_ *world.Tx, connected *player.Player) {
		entities := NewEntities()
		entityID := entities.Register(connected)
		handleID, ok := entities.EntityHandleID(entityID)
		if !ok {
			t.Fatal("registered entity has no handle ID")
		}
		wantUUID := [16]byte(connected.H().UUID())
		entities.Unregister(connected)
		if handle, ok := entities.HandleByID(handleID); ok || handle != nil {
			t.Fatalf("expired handle = %p, ok=%v after unregister", handle, ok)
		}
		if got, ok := entities.HandleUUID(handleID); !ok || got != wantUUID {
			t.Fatalf("expired UUID = %v, %v; want %v, true", got, ok, wantUUID)
		}
		if closed, ok := entities.HandleClosed(handleID); !ok || !closed {
			t.Fatalf("expired closed = %v, %v; want true, true", closed, ok)
		}
		if _, _, ok := entities.Detach(entityID); ok {
			t.Fatal("unregistered entity ID detached")
		}
		second := entities.Register(connected)
		secondHandleID, ok := entities.EntityHandleID(second)
		if !ok || secondHandleID == handleID {
			t.Fatalf("re-registration reused stale handle ID: %#v to %#v, ok=%v", handleID, secondHandleID, ok)
		}
		if _, ok := entities.HandleByID(handleID); ok {
			t.Fatal("old stable handle aliased re-registration")
		}
	})
}

func TestEntitiesKeepClosedHandleInspectable(t *testing.T) {
	entities := NewEntities()
	id := uuid.MustParse("c8d26bc5-0f05-48d0-b596-dc5c0bf0b808")
	handle := world.EntitySpawnOpts{ID: id}.New(
		player.Type,
		player.Config{UUID: id, Name: "Closed"},
	)
	entityID := entities.registerHandle(handle, 0)
	handleID, ok := entities.EntityHandleID(entityID)
	if !ok {
		t.Fatal("registered handle has no stable ID")
	}
	_ = handle.Close()
	entities.unregisterHandle(handle)
	if resolved, ok := entities.HandleByID(handleID); ok || resolved != nil {
		t.Fatalf("closed handle retained = %p, ok=%v", resolved, ok)
	}
	if closed, ok := entities.HandleClosed(handleID); !ok || !closed {
		t.Fatalf("closed=%v, ok=%v", closed, ok)
	}
	if got, ok := entities.HandleUUID(handleID); !ok || got != [16]byte(id) {
		t.Fatalf("UUID=%v, ok=%v", got, ok)
	}
}

func TestEntitiesDoNotAllocateHandlesUntilExposed(t *testing.T) {
	withPlayerTx(t, func(_ *world.Tx, connected *player.Player) {
		entities := NewEntities()
		if id := entities.Register(connected); id.Generation == 0 {
			t.Fatal("entity registration failed")
		}
		entities.Unregister(connected)
		if len(entities.byHandleID) != 0 || len(entities.handleTombstones) != 0 {
			t.Fatalf("unexposed entity retained: live=%d tombstones=%d", len(entities.byHandleID), len(entities.handleTombstones))
		}
	})
}

func TestEntitiesOldHandleTokenCannotResolveReregisteredHandle(t *testing.T) {
	withPlayerTx(t, func(tx *world.Tx, connected *player.Player) {
		entities := NewEntities()
		first := entities.Register(connected)
		oldHandle, ok := entities.EntityHandleID(first)
		if !ok {
			t.Fatal("first stable handle allocation failed")
		}
		entities.Unregister(connected)
		second := entities.Register(connected)
		newHandle, ok := entities.EntityHandleID(second)
		if !ok || newHandle == oldHandle {
			t.Fatalf("new stable handle = %#v, ok=%v; old=%#v", newHandle, ok, oldHandle)
		}
		if _, ok := entities.ResolveHandle(oldHandle, tx); ok {
			t.Fatal("old stable handle resolved re-registered entity")
		}
		if !entities.CloseHandle(oldHandle) {
			t.Fatal("closing expired stable handle was not idempotent")
		}
		if _, ok := entities.Resolve(second, tx); !ok {
			t.Fatal("closing old stable handle closed new registration")
		}
	})
}
