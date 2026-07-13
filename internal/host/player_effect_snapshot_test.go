package host

import (
	"context"
	"image/color"
	"strconv"
	"testing"
	"time"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
)

type snapshotCustomEffect struct{}

func (snapshotCustomEffect) RGBA() color.RGBA                  { return color.RGBA{} }
func (snapshotCustomEffect) Apply(world.Entity, effect.Effect) {}
func (snapshotCustomEffect) Start(world.Entity, int)           {}
func (snapshotCustomEffect) End(world.Entity, int)             {}

func TestPlayersSnapshotAndClearEffects(t *testing.T) {
	const customID = -31_997
	effect.Register(customID, snapshotCustomEffect{})
	withPlayer(t, func(connected *player.Player) {
		players := NewPlayers()
		id := players.Register(connected, 81)
		invocation, leave := players.BeginInvocation(connected.Tx())
		defer leave()

		connected.AddEffect(effect.New(effect.Speed, 2, 5*time.Second))
		connected.AddEffect(effect.NewAmbient(effect.Regeneration, 1, 2*time.Second).WithoutParticles())
		connected.AddEffect(effect.NewInfinite(effect.FireResistance, 3).WithoutParticles())
		connected.AddEffect(effect.New(snapshotCustomEffect{}, 4, time.Second))

		values, ok := players.PlayerEffects(invocation, id)
		if !ok || len(values) != 4 {
			t.Fatalf("effects = %#v ok=%v", values, ok)
		}
		byType := map[native.EffectType]native.PlayerEffect{}
		for _, value := range values {
			byType[value.Type] = value
		}
		if value := byType[native.EffectSpeed]; value.Level != 2 || value.Duration != 5*time.Second || value.Potency != 1 || value.Mode != native.PlayerEffectTimed || value.ParticlesHidden {
			t.Fatalf("speed = %#v", value)
		}
		if value := byType[native.EffectRegeneration]; value.Mode != native.PlayerEffectAmbient || !value.ParticlesHidden {
			t.Fatalf("ambient = %#v", value)
		}
		if value := byType[native.EffectFireResistance]; value.Mode != native.PlayerEffectInfinite || value.Duration != 0 || !value.ParticlesHidden {
			t.Fatalf("infinite = %#v", value)
		}
		if value := byType[native.EffectType(customID)]; value.Level != 4 || value.Mode != native.PlayerEffectTimed {
			t.Fatalf("custom = %#v", value)
		}

		if !players.ClearPlayerEffects(invocation, id) || len(connected.Effects()) != 0 {
			t.Fatalf("effects after clear = %#v", connected.Effects())
		}

		players.Unregister(connected)
		if values, ok := players.PlayerEffects(invocation, id); ok || values != nil {
			t.Fatalf("stale effects = %#v ok=%v", values, ok)
		}
		if players.ClearPlayerEffects(invocation, id) {
			t.Fatal("cleared effects for stale player")
		}
	})
}

func TestPlayersSnapshotEmptyEffects(t *testing.T) {
	withPlayer(t, func(connected *player.Player) {
		players := NewPlayers()
		id := players.Register(connected, 82)
		values, ok := players.PlayerEffects(0, id)
		if !ok || len(values) != 0 {
			t.Fatalf("effects = %#v ok=%v", values, ok)
		}
	})
}

func TestPlayersSnapshotFiltersInitialInstantEffects(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = w.Close() })
	id := uuid.MustParse("56b1b0c8-d899-43ae-a683-ef8714980289")
	handle := world.EntitySpawnOpts{ID: id}.New(player.Type, player.Config{
		UUID: id, Name: "InitialEffects", Effects: []effect.Effect{
			effect.New(effect.Speed, 2, time.Second),
			effect.NewInstant(effect.InstantHealth, 1),
		},
	})
	if err := w.Do(func(tx *world.Tx) {
		connected := tx.AddEntity(handle).(*player.Player)
		players := NewPlayers()
		playerID := players.Register(connected, 83)
		invocation, leave := players.BeginInvocation(tx)
		defer leave()
		values, ok := players.PlayerEffects(invocation, playerID)
		if !ok || len(values) != 1 || values[0].Type != native.EffectSpeed {
			t.Fatalf("effects = %#v ok=%v", values, ok)
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestSnapshotPlayerEffectClampsNegativeRemainingDuration(t *testing.T) {
	current := effect.New(effect.Speed, 1, time.Millisecond).TickDuration()
	value, ok := snapshotPlayerEffect(current)
	if !ok || current.Duration() >= 0 || value.Duration != 0 {
		t.Fatalf("current duration=%v snapshot=%#v ok=%v", current.Duration(), value, ok)
	}
}

type snapshotLargeIDEffect struct{}

func (snapshotLargeIDEffect) RGBA() color.RGBA                  { return color.RGBA{} }
func (snapshotLargeIDEffect) Apply(world.Entity, effect.Effect) {}
func (snapshotLargeIDEffect) Start(world.Entity, int)           {}
func (snapshotLargeIDEffect) End(world.Entity, int)             {}

func TestSnapshotPlayerEffectRejectsOutOfRangeID(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("requires a 64-bit effect registry ID")
	}
	effect.Register(1<<40, snapshotLargeIDEffect{})
	if _, ok := snapshotPlayerEffect(effect.New(snapshotLargeIDEffect{}, 1, time.Second)); ok {
		t.Fatal("snapshot accepted an effect ID outside i32")
	}
}

func TestSnapshotPlayerEffectRejectsOutOfRangeLevel(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("requires a 64-bit effect level")
	}
	if _, ok := snapshotPlayerEffect(effect.New(effect.Speed, 1<<40, time.Second)); ok {
		t.Fatal("snapshot accepted an effect level outside i32")
	}
}

type snapshotUnregisteredEffect struct{}

func (snapshotUnregisteredEffect) RGBA() color.RGBA                  { return color.RGBA{} }
func (snapshotUnregisteredEffect) Apply(world.Entity, effect.Effect) {}
func (snapshotUnregisteredEffect) Start(world.Entity, int)           {}
func (snapshotUnregisteredEffect) End(world.Entity, int)             {}

func TestPlayersClearEffectsRejectsUnregisteredBeforeRemoval(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = w.Close() })
	id := uuid.MustParse("83e59098-d2ce-4421-8a51-9d62838fb026")
	handle := world.EntitySpawnOpts{ID: id}.New(player.Type, player.Config{
		UUID: id, Name: "UnregisteredEffect", Effects: []effect.Effect{
			effect.New(effect.Speed, 1, time.Second),
			effect.New(snapshotUnregisteredEffect{}, 1, time.Second),
		},
	})
	if err := w.Do(func(tx *world.Tx) {
		connected := tx.AddEntity(handle).(*player.Player)
		players := NewPlayers()
		playerID := players.Register(connected, 84)
		invocation, leave := players.BeginInvocation(tx)
		defer leave()
		if players.ClearPlayerEffects(invocation, playerID) {
			t.Fatal("clear accepted an unregistered effect type")
		}
		if effects := connected.Effects(); len(effects) != 2 {
			t.Fatalf("effects were partially cleared: %#v", effects)
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}
