package host

import (
	"context"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
)

func TestPlayersResolveFreshTransaction(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = w.Close() })
	playerUUID := uuid.New()
	handle := world.EntitySpawnOpts{ID: playerUUID}.New(
		player.Type,
		player.Config{UUID: playerUUID, Name: "Transactional", Position: mgl64.Vec3{}},
	)
	players := NewPlayers()
	var id native.PlayerID
	if err := w.Do(func(tx *world.Tx) {
		id = players.Register(tx.AddEntity(handle).(*player.Player), 2)
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	changed := false
	if err := w.Do(func(tx *world.Tx) {
		players.WithInvocation(tx, func(invocation native.InvocationID) {
			changed = players.SetPlayerState(invocation, id, native.PlayerStateGameMode, native.PlayerStateValue{Integer: 1})
		})
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("player action did not resolve through fresh transaction")
	}
}

func TestPlayersInvocationRegistryIsExactAndExpires(t *testing.T) {
	first := world.Config{Synchronous: true}.New()
	second := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = first.Close(); _ = second.Close() })
	players := NewPlayers()
	if err := first.Do(func(firstTx *world.Tx) {
		firstID, endFirst := players.BeginInvocation(firstTx)
		defer endFirst()
		if firstID == 0 {
			t.Fatal("zero invocation ID")
		}
		if got, ok := players.InvocationTx(firstID); !ok || got != firstTx {
			t.Fatalf("first invocation = %p, %v", got, ok)
		}
		if err := second.Do(func(secondTx *world.Tx) {
			secondID, endSecond := players.BeginInvocation(secondTx)
			if secondID <= firstID {
				t.Fatalf("invocation IDs not monotonic: %d then %d", firstID, secondID)
			}
			if got, ok := players.InvocationTx(firstID); !ok || got != firstTx {
				t.Fatalf("second invocation aliased first: %p, %v", got, ok)
			}
			endSecond()
			endSecond()
			if _, ok := players.InvocationTx(secondID); ok {
				t.Fatal("ended invocation still resolves")
			}
		}).Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := players.InvocationTx(0); ok {
		t.Fatal("zero invocation resolved")
	}
}

func TestPlayersTransformsPlayer(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()
		if !players.TransformPlayer(invocation, id, native.PlayerTransformTeleport, native.Vec3{X: 4, Y: 5, Z: 6}, 0, 0) {
			t.Fatal("teleport failed")
		}
		if player.Position() != ([3]float64{4, 5, 6}) {
			t.Fatalf("position = %v", player.Position())
		}
		if !players.TransformPlayer(invocation, id, native.PlayerTransformMove, native.Vec3{X: 1}, 20, 5) {
			t.Fatal("move failed")
		}
		rotation, ok := players.PlayerRotation(invocation, id)
		if !ok || rotation.Yaw != 20 || rotation.Pitch != 5 {
			t.Fatalf("rotation = %+v ok=%v", rotation, ok)
		}
		if !players.TransformPlayer(invocation, id, native.PlayerTransformVelocity, native.Vec3{Y: 1}, 0, 0) || player.Velocity().Y() != 1 {
			t.Fatalf("velocity = %v", player.Velocity())
		}
	})
}

func TestPlayersSendsAndRemovesScoreboard(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()
		if !players.SendPlayerScoreboard(invocation, id, native.PlayerScoreboard{
			Name: "Stats", Lines: []string{"Wins: 3", "Losses: 1"}, Padding: false, Descending: true,
		}) {
			t.Fatal("send scoreboard failed")
		}
		if players.SendPlayerScoreboard(invocation, id, native.PlayerScoreboard{Lines: make([]string, 16)}) {
			t.Fatal("accepted a scoreboard with more than 15 lines")
		}
		players.Unregister(player)
		if players.RemovePlayerScoreboard(invocation, id) {
			t.Fatal("removed scoreboard for a stale player")
		}
	})
}

func TestPlayersTracksStableGenerationAndNames(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 42)
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()
		if id.Generation != 42 || len(players.Names()) != 1 || players.Names()[0] != "TestPlayer" {
			t.Fatalf("id=%+v names=%v", id, players.Names())
		}
		resolved, ok := players.ResolveName("testplayer")
		if !ok || resolved != id {
			t.Fatalf("resolved=%+v ok=%v", resolved, ok)
		}
		connected, ok := players.ResolveID(id, invocation)
		if !ok || connected.UUID() != player.UUID() {
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
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()
		changes := []struct {
			kind  native.PlayerStateKind
			value native.PlayerStateValue
		}{
			{native.PlayerStateFood, native.PlayerStateValue{Integer: 12}},
			{native.PlayerStateMaxHealth, native.PlayerStateValue{Number: 40}},
			{native.PlayerStateExperienceLevel, native.PlayerStateValue{Integer: 12}},
			{native.PlayerStateExperienceProgress, native.PlayerStateValue{Number: 0.5}},
			{native.PlayerStateScale, native.PlayerStateValue{Number: 1.5}},
			{native.PlayerStateInvisible, native.PlayerStateValue{Integer: 1}},
			{native.PlayerStateImmobile, native.PlayerStateValue{Integer: 1}},
		}
		for _, change := range changes {
			if !players.SetPlayerState(invocation, id, change.kind, change.value) {
				t.Fatalf("state change %d failed", change.kind)
			}
		}
		hurt, ok := players.HurtPlayer(invocation, id, 4, native.DamageSource{Kind: native.DamageSourceInstant})
		if !ok || hurt.Damage != 4 || !hurt.Vulnerable {
			t.Fatalf("hurt = %+v ok=%v", hurt, ok)
		}
		healed, ok := players.HealPlayer(invocation, id, 3, native.HealingSource{Kind: native.HealingSourceInstant})
		if !ok || healed != 3 {
			t.Fatalf("healed = %v ok=%v", healed, ok)
		}
		if !players.SetPlayerState(invocation, id, native.PlayerStateGameMode, native.PlayerStateValue{Integer: 1}) {
			t.Fatal("game-mode change failed")
		}
		gameMode, _ := players.PlayerState(invocation, id, native.PlayerStateGameMode)
		food, _ := players.PlayerState(invocation, id, native.PlayerStateFood)
		maxHealth, _ := players.PlayerState(invocation, id, native.PlayerStateMaxHealth)
		health, _ := players.PlayerState(invocation, id, native.PlayerStateHealth)
		level, _ := players.PlayerState(invocation, id, native.PlayerStateExperienceLevel)
		progress, _ := players.PlayerState(invocation, id, native.PlayerStateExperienceProgress)
		scale, _ := players.PlayerState(invocation, id, native.PlayerStateScale)
		invisible, _ := players.PlayerState(invocation, id, native.PlayerStateInvisible)
		immobile, _ := players.PlayerState(invocation, id, native.PlayerStateImmobile)
		if gameMode.Integer != 1 || food.Integer != 12 || maxHealth.Number != 40 || health.Number != 19 || level.Integer != 12 || math.Abs(progress.Number-0.5) > 0.02 || scale.Number != 1.5 || invisible.Integer != 1 || immobile.Integer != 1 {
			t.Fatalf("game mode=%+v food=%+v max=%+v health=%+v level=%+v progress=%+v scale=%+v invisible=%+v immobile=%+v", gameMode, food, maxHealth, health, level, progress, scale, invisible, immobile)
		}
		if !players.SendPlayerText(invocation, id, native.PlayerTextNameTag, "Rust Player") || player.NameTag() != "Rust Player" {
			t.Fatalf("name tag = %q", player.NameTag())
		}
		if !players.PlayPlayerSound(invocation, id, native.WorldSound{Kind: native.SoundLevelUp}) {
			t.Fatal("play sound failed")
		}
		entityID := native.EntityID{UUID: id.UUID, Generation: id.Generation}
		if !players.SetPlayerEntityVisible(invocation, id, entityID, false) || !players.SetPlayerEntityVisible(invocation, id, entityID, true) {
			t.Fatal("entity visibility failed")
		}
		if !players.ChangePlayerEffect(invocation, id, native.PlayerEffectAdd, native.PlayerEffect{
			Type: native.EffectSpeed, Level: 2, Duration: 30 * time.Second,
		}) {
			t.Fatal("add effect failed")
		}
		applied, ok := player.Effect(effect.Speed)
		if !ok || applied.Level() != 2 {
			t.Fatalf("effect = %+v ok=%v", applied, ok)
		}
		if !players.ChangePlayerEffect(invocation, id, native.PlayerEffectRemove, native.PlayerEffect{Type: native.EffectSpeed}) {
			t.Fatal("remove effect failed")
		}
		if _, ok := player.Effect(effect.Speed); ok {
			t.Fatal("effect still present")
		}
	})
}

func TestPlayersReconstructsConcreteHealAndHurtSources(t *testing.T) {
	withPlayerTx(t, func(tx *world.Tx, connected *player.Player) {
		players := NewPlayers()
		playerID := players.Register(connected, 17)
		entityID := native.EntityID{UUID: playerID.UUID, Generation: playerID.Generation}

		attack, ok := players.damageSource(tx, native.DamageSource{Kind: native.DamageSourceAttack, Entity: entityID})
		attackSource, typed := attack.(entity.AttackDamageSource)
		if !ok || !typed || attackSource.Attacker.H() != connected.H() {
			t.Fatalf("attack = %#v ok=%v", attack, ok)
		}
		attack, ok = players.damageSource(tx, native.DamageSource{Kind: native.DamageSourceAttack})
		attackSource, typed = attack.(entity.AttackDamageSource)
		if !ok || !typed || attackSource.Attacker != nil {
			t.Fatalf("attack without attacker = %#v ok=%v", attack, ok)
		}

		projectile, ok := players.damageSource(tx, native.DamageSource{
			Kind: native.DamageSourceProjectile, Entity: entityID, SecondaryEntity: entityID,
		})
		projectileSource, typed := projectile.(entity.ProjectileDamageSource)
		if !ok || !typed || projectileSource.Projectile.H() != connected.H() || projectileSource.Owner.H() != connected.H() {
			t.Fatalf("projectile = %#v ok=%v", projectile, ok)
		}
		projectile, ok = players.damageSource(tx, native.DamageSource{Kind: native.DamageSourceProjectile})
		projectileSource, typed = projectile.(entity.ProjectileDamageSource)
		if !ok || !typed || projectileSource.Projectile != nil || projectileSource.Owner != nil {
			t.Fatalf("projectile without entities = %#v ok=%v", projectile, ok)
		}

		properties, valid := EncodeBlockProperties(map[string]any{"age": int32(4)})
		if !valid {
			t.Fatal("encode cactus properties")
		}
		blockSource, ok := players.damageSource(tx, native.DamageSource{
			Kind:  native.DamageSourceBlock,
			Block: &native.WorldBlock{Identifier: "minecraft:cactus", PropertiesNBT: properties},
		})
		resolvedBlock, typed := blockSource.(block.DamageSource)
		cactus, cactusOK := resolvedBlock.Block.(block.Cactus)
		if !ok || !typed || !cactusOK || cactus.Age != 4 {
			t.Fatalf("block = %#v ok=%v", blockSource, ok)
		}

		custom, ok := players.damageSource(tx, native.DamageSource{
			Name: "example.Custom", ReducedByArmour: true, FireProtection: true,
		})
		customSource, typed := custom.(pluginDamageSource)
		if !ok || !typed || !customSource.ReducedByArmour() || !customSource.AffectedByEnchantment(enchantment.FireProtection) || customSource.Name() != "example.Custom" {
			t.Fatalf("custom = %#v ok=%v", custom, ok)
		}

		food, typed := healingSource(native.HealingSource{Kind: native.HealingSourceFood, Data: true}).(entity.FoodHealingSource)
		if !typed || !food.QuickRegeneration {
			t.Fatalf("food = %#v", food)
		}
	})
}

func TestPlayersSkinRoundTrip(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()
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
		if !players.SetPlayerSkin(invocation, id, want) {
			t.Fatal("set skin failed")
		}
		got, ok := players.PlayerSkin(invocation, id)
		if !ok {
			t.Fatal("get skin failed")
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("skin mismatch\ngot:  %+v\nwant: %+v", got, want)
		}
		want.Pixels[0] ^= 0xff
		want.CapePixels[0] ^= 0xff
		want.Animations[0].Pixels[0] ^= 0xff
		gotAgain, ok := players.PlayerSkin(invocation, id)
		if !ok || !reflect.DeepEqual(gotAgain, got) {
			t.Fatal("host retained caller-owned skin buffers")
		}
		got.Pixels[0] ^= 0xff
		got.CapePixels[0] ^= 0xff
		got.Animations[0].Pixels[0] ^= 0xff
		gotAgain, ok = players.PlayerSkin(invocation, id)
		if !ok || reflect.DeepEqual(gotAgain, got) {
			t.Fatal("host returned player-owned skin buffers")
		}
	})
}

func TestPlayersRejectsInvalidSkinData(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()
		original, ok := players.PlayerSkin(invocation, id)
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
			if players.SetPlayerSkin(invocation, id, value) {
				t.Fatalf("invalid skin %d accepted", i)
			}
		}
		got, ok := players.PlayerSkin(invocation, id)
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
