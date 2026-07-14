package framework

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
)

func TestWorldManagerEntityLifecycleUsesExactTransaction(t *testing.T) {
	players := host.NewPlayers()
	manager := newWorldManager("", nil, players)
	w, err := manager.Create("example:entities", world.Config{Synchronous: true, Entities: entity.DefaultRegistry})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	worldID, _ := manager.WorldByName(0, "example:entities")

	var staleInvocation native.InvocationID
	var spawned native.EntityID
	if err := w.Do(func(tx *world.Tx) {
		invocation, end := players.BeginInvocation(tx)
		defer end()
		staleInvocation = invocation
		id, ok := manager.SpawnWorldEntity(invocation, worldID, native.EntitySpawn{
			Kind: native.EntityText, Position: native.Vec3{X: 1, Y: 64, Z: 2}, Text: "hello",
		})
		if !ok {
			t.Fatal("spawn text failed")
		}
		spawned = id
		state, ok := manager.EntityState(invocation, id)
		if !ok || state.World != worldID || state.Type != "dragonfly:text" || state.Position != (native.Vec3{X: 1, Y: 64, Z: 2}) || !state.CanTeleport {
			t.Fatalf("state = %#v, %v", state, ok)
		}
		if !manager.TeleportEntity(invocation, id, native.Vec3{X: 4, Y: 70, Z: 5}) ||
			!manager.SetEntityVelocity(invocation, id, native.Vec3{Y: 0.5}) ||
			!manager.SetEntityNameTag(invocation, id, "renamed") {
			t.Fatal("entity mutation failed")
		}
		state, ok = manager.EntityState(invocation, id)
		if !ok || state.Position != (native.Vec3{X: 4, Y: 70, Z: 5}) || state.Velocity.Y != 0.5 || state.NameTag != "renamed" {
			t.Fatalf("mutated state = %#v, %v", state, ok)
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if manager.SetEntityNameTag(staleInvocation, spawned, "late") {
		t.Fatal("expired invocation mutated an entity")
	}
	if !manager.DespawnEntity(0, spawned) {
		t.Fatal("off-callback despawn failed")
	}
}

func TestWorldManagerEntityIteratorsStayInsideInvocation(t *testing.T) {
	players := host.NewPlayers()
	manager := newWorldManager("", nil, players)
	w, err := manager.Create("example:iterators", world.Config{Synchronous: true, Entities: entity.DefaultRegistry})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	worldID, _ := manager.WorldByName(0, "example:iterators")

	if err := w.Do(func(tx *world.Tx) {
		playerUUID := uuid.MustParse("125bbd56-6bed-41f3-bff1-dc206b534ace")
		handle := world.EntitySpawnOpts{ID: playerUUID}.New(player.Type, player.Config{UUID: playerUUID, Name: "Iterator"})
		connected := tx.AddEntity(handle).(*player.Player)
		players.Register(connected, 90)
		playerEntityID, ok := manager.entityHandles.ID(connected)
		if !ok {
			t.Fatal("player entity ID missing")
		}

		invocation, end := players.BeginInvocation(tx)
		if got, ok := manager.CurrentWorld(invocation); !ok || got != worldID {
			t.Fatalf("CurrentWorld() = %d, %v, want %d", got, ok, worldID)
		}
		if _, ok := manager.CurrentWorld(0); ok {
			t.Fatal("CurrentWorld(0) succeeded")
		}
		textID, ok := manager.SpawnWorldEntity(invocation, worldID, native.EntitySpawn{
			Kind: native.EntityText, Position: native.Vec3{Y: 64}, Text: "lazy",
		})
		if !ok {
			t.Fatal("spawn text failed")
		}
		if _, ok := manager.OpenWorldEntityIterator(0, worldID, false); ok {
			t.Fatal("context-free entity iterator opened")
		}
		if _, ok := manager.OpenWorldEntityIterator(invocation, worldID+1, false); ok {
			t.Fatal("entity iterator opened for a non-current world")
		}

		all, ok := manager.OpenWorldEntityIterator(invocation, worldID, false)
		if !ok || all == 0 {
			t.Fatal("entity iterator open failed")
		}
		if _, _, valid := manager.NextWorldEntity(invocation+1, all); valid {
			t.Fatal("wrong invocation advanced entity iterator")
		}
		manager.CloseWorldEntities(invocation+1, all)
		var ids []native.EntityID
		for {
			id, found, valid := manager.NextWorldEntity(invocation, all)
			if !valid {
				t.Fatal("entity iterator became invalid")
			}
			if !found {
				break
			}
			ids = append(ids, id)
		}
		if !slices.Contains(ids, textID) || !slices.Contains(ids, playerEntityID) {
			t.Fatalf("entity iterator IDs = %#v", ids)
		}
		if _, _, valid := manager.NextWorldEntity(invocation, all); valid {
			t.Fatal("finished entity iterator remained live")
		}

		onlyPlayers, ok := manager.OpenWorldEntityIterator(invocation, worldID, true)
		if !ok {
			t.Fatal("player iterator open failed")
		}
		gotPlayer, found, valid := manager.NextWorldEntity(invocation, onlyPlayers)
		if !valid || !found || gotPlayer != playerEntityID {
			t.Fatalf("player iterator = %#v, %v, %v", gotPlayer, found, valid)
		}
		manager.CloseWorldEntities(invocation, onlyPlayers)
		if _, _, valid := manager.NextWorldEntity(invocation, onlyPlayers); valid {
			t.Fatal("closed player iterator remained live")
		}

		leaked, ok := manager.OpenWorldEntityIterator(invocation, worldID, false)
		if !ok {
			t.Fatal("leaked iterator open failed")
		}
		end()
		if _, _, valid := manager.NextWorldEntity(invocation, leaked); valid {
			t.Fatal("entity iterator survived invocation cleanup")
		}
		manager.entityIteratorMu.Lock()
		remaining := len(manager.entityIterators)
		manager.entityIteratorMu.Unlock()
		if remaining != 0 {
			t.Fatalf("entity iterators after invocation = %d", remaining)
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestWorldManagerRejectsWorldlessEntityWithoutBlocking(t *testing.T) {
	players := host.NewPlayers()
	manager := newWorldManager("", nil, players)
	w, err := manager.Create("example:worldless", world.Config{Synchronous: true, Entities: entity.DefaultRegistry})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	worldID, _ := manager.WorldByName(0, "example:worldless")
	var id native.EntityID
	if err := w.Do(func(tx *world.Tx) {
		invocation, end := players.BeginInvocation(tx)
		defer end()
		var ok bool
		id, ok = manager.SpawnWorldEntity(invocation, worldID, native.EntitySpawn{
			Kind: native.EntityText, Position: native.Vec3{Y: 64}, Text: "temporary",
		})
		if !ok {
			t.Fatal("spawn text failed")
		}
		current, ok := manager.entityHandles.Resolve(id, tx)
		if !ok || tx.RemoveEntity(current) == nil {
			t.Fatal("remove entity failed")
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	done := make(chan bool, 1)
	go func() {
		_, ok := manager.EntityState(0, id)
		done <- ok
	}()
	select {
	case ok := <-done:
		if ok {
			t.Fatal("worldless entity returned state")
		}
	case <-time.After(time.Second):
		t.Fatal("worldless entity state blocked")
	}
}

func TestWorldManagerRejectsOverflowingEntityDuration(t *testing.T) {
	players := host.NewPlayers()
	manager := newWorldManager("", nil, players)
	w, err := manager.Create("example:duration", world.Config{Synchronous: true, Entities: entity.DefaultRegistry})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	worldID, _ := manager.WorldByName(0, "example:duration")
	if err := w.Do(func(tx *world.Tx) {
		invocation, end := players.BeginInvocation(tx)
		defer end()
		if _, ok := manager.SpawnWorldEntity(invocation, worldID, native.EntitySpawn{
			Kind: native.EntityTNT, Position: native.Vec3{Y: 64}, FuseMilliseconds: ^uint64(0),
		}); ok {
			t.Fatal("overflowing duration was accepted")
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestWorldManagerSpawnsTypedProjectiles(t *testing.T) {
	players := host.NewPlayers()
	manager := newWorldManager("", nil, players)
	w, err := manager.Create("example:projectiles", world.Config{Synchronous: true, Entities: entity.DefaultRegistry})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	worldID, _ := manager.WorldByName(0, "example:projectiles")

	if err := w.Do(func(tx *world.Tx) {
		playerUUID := uuid.MustParse("f898b46d-5ad3-40d6-b48b-40345b9622be")
		handle := world.EntitySpawnOpts{ID: playerUUID}.New(player.Type, player.Config{UUID: playerUUID, Name: "Owner"})
		connected := tx.AddEntity(handle).(*player.Player)
		players.Register(connected, 80)
		owner, ok := players.EntityRegistry().ID(connected)
		if !ok {
			t.Fatal("owner ID missing")
		}
		invocation, end := players.BeginInvocation(tx)
		defer end()
		for _, spawn := range []native.EntitySpawn{
			{Kind: native.EntitySnowball, Owner: owner, Position: native.Vec3{Y: 65}, Velocity: native.Vec3{Z: 1}},
			{Kind: native.EntityArrow, Owner: owner, Position: native.Vec3{Y: 65}, Velocity: native.Vec3{Z: 1}, Damage: 3, Flags: native.EntityArrowCritical, Potion: 25},
			{Kind: native.EntityTNT, Position: native.Vec3{Y: 64}, FuseMilliseconds: uint64((10 * time.Second).Milliseconds())},
			{Kind: native.EntityExperienceOrb, Position: native.Vec3{Y: 64}, Experience: 7},
		} {
			id, ok := manager.SpawnWorldEntity(invocation, worldID, spawn)
			if !ok || id.Generation == 0 {
				t.Fatalf("spawn kind %d failed", spawn.Kind)
			}
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}
