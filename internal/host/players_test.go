package host

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/player"
)

func TestPlayersTransformsPlayer(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		if !players.TransformPlayer(id, native.PlayerTransformTeleport, native.Vec3{X: 4, Y: 5, Z: 6}, 0, 0) {
			t.Fatal("teleport failed")
		}
		if player.Position() != ([3]float64{4, 5, 6}) {
			t.Fatalf("position = %v", player.Position())
		}
		if !players.TransformPlayer(id, native.PlayerTransformMove, native.Vec3{X: 1}, 20, 5) {
			t.Fatal("move failed")
		}
		rotation, ok := players.PlayerRotation(id)
		if !ok || rotation.Yaw != 20 || rotation.Pitch != 5 {
			t.Fatalf("rotation = %+v ok=%v", rotation, ok)
		}
		if !players.TransformPlayer(id, native.PlayerTransformVelocity, native.Vec3{Y: 1}, 0, 0) || player.Velocity().Y() != 1 {
			t.Fatalf("velocity = %v", player.Velocity())
		}
	})
}

func TestPlayersTracksStableGenerationAndNames(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 42)
		if id.Generation != 42 || len(players.Names()) != 1 || players.Names()[0] != "TestPlayer" {
			t.Fatalf("id=%+v names=%v", id, players.Names())
		}
		resolved, ok := players.ResolveName("testplayer")
		if !ok || resolved != id {
			t.Fatalf("resolved=%+v ok=%v", resolved, ok)
		}
		connected, ok := players.ResolveID(id)
		if !ok || connected != player {
			t.Fatalf("connected=%p ok=%v", connected, ok)
		}
		players.Unregister(player)
		if len(players.Names()) != 0 {
			t.Fatalf("names after unregister = %v", players.Names())
		}
	})
}

func TestPlayersReadsAndChangesState(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		changes := []struct {
			kind  native.PlayerStateKind
			value native.PlayerStateValue
		}{
			{native.PlayerStateFood, native.PlayerStateValue{Integer: 12}},
			{native.PlayerStateMaxHealth, native.PlayerStateValue{Number: 40}},
			{native.PlayerStateHurt, native.PlayerStateValue{Number: 4}},
			{native.PlayerStateHeal, native.PlayerStateValue{Number: 3}},
			{native.PlayerStateExperienceLevel, native.PlayerStateValue{Integer: 12}},
			{native.PlayerStateExperienceProgress, native.PlayerStateValue{Number: 0.5}},
			{native.PlayerStateScale, native.PlayerStateValue{Number: 1.5}},
			{native.PlayerStateInvisible, native.PlayerStateValue{Integer: 1}},
			{native.PlayerStateImmobile, native.PlayerStateValue{Integer: 1}},
			{native.PlayerStateGameMode, native.PlayerStateValue{Integer: 1}},
		}
		for _, change := range changes {
			if !players.SetPlayerState(id, change.kind, change.value) {
				t.Fatalf("state change %d failed", change.kind)
			}
		}
		gameMode, _ := players.PlayerState(id, native.PlayerStateGameMode)
		food, _ := players.PlayerState(id, native.PlayerStateFood)
		maxHealth, _ := players.PlayerState(id, native.PlayerStateMaxHealth)
		health, _ := players.PlayerState(id, native.PlayerStateHealth)
		level, _ := players.PlayerState(id, native.PlayerStateExperienceLevel)
		progress, _ := players.PlayerState(id, native.PlayerStateExperienceProgress)
		scale, _ := players.PlayerState(id, native.PlayerStateScale)
		invisible, _ := players.PlayerState(id, native.PlayerStateInvisible)
		immobile, _ := players.PlayerState(id, native.PlayerStateImmobile)
		if gameMode.Integer != 1 || food.Integer != 12 || maxHealth.Number != 40 || health.Number != 19 || level.Integer != 12 || math.Abs(progress.Number-0.5) > 0.02 || scale.Number != 1.5 || invisible.Integer != 1 || immobile.Integer != 1 {
			t.Fatalf("game mode=%+v food=%+v max=%+v health=%+v level=%+v progress=%+v scale=%+v invisible=%+v immobile=%+v", gameMode, food, maxHealth, health, level, progress, scale, invisible, immobile)
		}
		if !players.SendPlayerText(id, native.PlayerTextNameTag, "Rust Player") || player.NameTag() != "Rust Player" {
			t.Fatalf("name tag = %q", player.NameTag())
		}
		if !players.SetPlayerState(id, native.PlayerStateSound, native.PlayerStateValue{Integer: int64(native.SoundLevelUp)}) {
			t.Fatal("play sound failed")
		}
		entityID := native.EntityID{UUID: id.UUID, Generation: id.Generation}
		if !players.SetPlayerEntityVisible(id, entityID, false) || !players.SetPlayerEntityVisible(id, entityID, true) {
			t.Fatal("entity visibility failed")
		}
		if !players.ChangePlayerEffect(id, native.PlayerEffectAdd, native.PlayerEffect{
			Type: native.EffectSpeed, Level: 2, Duration: 30 * time.Second,
		}) {
			t.Fatal("add effect failed")
		}
		applied, ok := player.Effect(effect.Speed)
		if !ok || applied.Level() != 2 {
			t.Fatalf("effect = %+v ok=%v", applied, ok)
		}
		if !players.ChangePlayerEffect(id, native.PlayerEffectRemove, native.PlayerEffect{Type: native.EffectSpeed}) {
			t.Fatal("remove effect failed")
		}
		if _, ok := player.Effect(effect.Speed); ok {
			t.Fatal("effect still present")
		}
	})
}

func TestPlayersSkinRoundTrip(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		want := native.PlayerSkin{
			Width: 64, Height: 32,
			Persona: true, PlayFabID: "playfab-id", FullID: "full-id",
			Pixels:            patternedBytes(64 * 32 * 4),
			ModelDefault:      "geometry.humanoid.custom",
			ModelAnimatedFace: "geometry.face.custom",
			Model:             []byte(`{"format_version":"1.12.0"}`),
			CapeWidth:         32, CapeHeight: 64, CapePixels: patternedBytes(32 * 64 * 4),
			Animations: []native.SkinAnimation{
				{
					Width: 32, Height: 32, Type: 1, FrameCount: 2, Expression: 3,
					Pixels: patternedBytes(32 * 32 * 4),
				},
				{
					Width: 16, Height: 16, Type: 0, FrameCount: 1, Expression: -2,
					Pixels: patternedBytes(16 * 16 * 4),
				},
			},
		}
		if !players.SetPlayerSkin(id, want) {
			t.Fatal("set skin failed")
		}
		got, ok := players.PlayerSkin(id)
		if !ok {
			t.Fatal("get skin failed")
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("skin mismatch\ngot:  %+v\nwant: %+v", got, want)
		}
		want.Pixels[0] ^= 0xff
		want.CapePixels[0] ^= 0xff
		want.Animations[0].Pixels[0] ^= 0xff
		gotAgain, ok := players.PlayerSkin(id)
		if !ok || !reflect.DeepEqual(gotAgain, got) {
			t.Fatal("host retained caller-owned skin buffers")
		}
		got.Pixels[0] ^= 0xff
		got.CapePixels[0] ^= 0xff
		got.Animations[0].Pixels[0] ^= 0xff
		gotAgain, ok = players.PlayerSkin(id)
		if !ok || reflect.DeepEqual(gotAgain, got) {
			t.Fatal("host returned player-owned skin buffers")
		}
	})
}

func TestPlayersRejectsInvalidSkinData(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		original, ok := players.PlayerSkin(id)
		if !ok {
			t.Fatal("get original skin failed")
		}
		invalid := []native.PlayerSkin{
			{Width: maxSkinDimension + 1, Height: 1, Pixels: make([]byte, 4)},
			{Width: 64, Height: 64, Pixels: make([]byte, 4)},
			{Width: 64, Height: 64, Pixels: make([]byte, 64*64*4), CapeWidth: 1},
			{Width: 64, Height: 64, Pixels: make([]byte, 64*64*4), PlayFabID: string(make([]byte, maxSkinIDBytes+1))},
			{Width: 64, Height: 64, Pixels: make([]byte, 64*64*4), Animations: make([]native.SkinAnimation, maxSkinAnimations+1)},
			{Width: 64, Height: 64, Pixels: make([]byte, 64*64*4), Animations: []native.SkinAnimation{{Width: 1, Height: 1, Type: 3, FrameCount: 1, Pixels: make([]byte, 4)}}},
			{Width: 64, Height: 64, Pixels: make([]byte, 64*64*4), Animations: []native.SkinAnimation{{Width: 1, Height: 1, FrameCount: 0, Pixels: make([]byte, 4)}}},
		}
		for i, value := range invalid {
			if players.SetPlayerSkin(id, value) {
				t.Fatalf("invalid skin %d accepted", i)
			}
		}
		got, ok := players.PlayerSkin(id)
		if !ok || !reflect.DeepEqual(got, original) {
			t.Fatal("rejected skin changed player skin")
		}
	})
}

func patternedBytes(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i * 31)
	}
	return data
}
