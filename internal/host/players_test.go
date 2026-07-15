package host

import (
	"context"
	"image/color"
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
	"github.com/df-mc/dragonfly/server/player/hud"
	"github.com/df-mc/dragonfly/server/player/input"
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
	creativeDescriptor, ok := encodeGameModeDescriptor(world.GameModeCreative)
	if !ok {
		t.Fatal("encode creative game mode")
	}
	var id native.PlayerID
	if err := w.Do(func(tx *world.Tx) {
		id = players.Register(tx.AddEntity(handle).(*player.Player), 2)
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	changed := false
	if err := w.Do(func(tx *world.Tx) {
		players.WithInvocation(tx, func(invocation native.InvocationID) {
			changed = players.SetPlayerState(invocation, id, native.PlayerStateGameMode, native.PlayerStateValue{Integer: creativeDescriptor})
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
			cleanups := 0
			if !players.OnInvocationEnd(secondID, func() { cleanups++ }) ||
				!players.OnInvocationEnd(secondID, func() { cleanups++ }) {
				t.Fatal("invocation cleanup registration failed")
			}
			endSecond()
			endSecond()
			if cleanups != 2 {
				t.Fatalf("invocation cleanups = %d", cleanups)
			}
			if players.OnInvocationEnd(secondID, func() {}) {
				t.Fatal("cleanup registered after invocation ended")
			}
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

type testCustomGameMode struct {
	flags uint8
	data  []byte
}

func (m testCustomGameMode) AllowsEditing() bool       { return m.flags&(1<<0) != 0 }
func (m testCustomGameMode) AllowsTakingDamage() bool  { return m.flags&(1<<1) != 0 }
func (m testCustomGameMode) CreativeInventory() bool   { return m.flags&(1<<2) != 0 }
func (m testCustomGameMode) HasCollision() bool        { return m.flags&(1<<3) != 0 }
func (m testCustomGameMode) AllowsFlying() bool        { return m.flags&(1<<4) != 0 }
func (m testCustomGameMode) AllowsInteraction() bool   { return m.flags&(1<<5) != 0 }
func (m testCustomGameMode) Visible() bool             { return m.flags&(1<<6) != 0 }
func (m testCustomGameMode) InstantPortalTravel() bool { return m.flags&(1<<7) != 0 }

func TestGameModeDescriptorsRoundTrip(t *testing.T) {
	builtins := []world.GameMode{
		world.GameModeSurvival,
		world.GameModeCreative,
		world.GameModeAdventure,
		world.GameModeSpectator,
	}
	for id, want := range builtins {
		descriptorBits := gameModeBuiltinFlag | uint64(id)
		descriptor := int64(descriptorBits)
		decoded, ok := decodeGameModeDescriptor(descriptor)
		if !ok {
			t.Fatalf("decode builtin %d", id)
		}
		decodedID, registered := world.GameModeID(decoded)
		if !registered || decodedID != id {
			t.Fatalf("decoded builtin %d = %T, id=%d, registered=%v", id, decoded, decodedID, registered)
		}
		encoded, ok := encodeGameModeDescriptor(want)
		if !ok || encoded != descriptor {
			t.Fatalf("encode builtin %d = %#x, %v, want %#x", id, uint64(encoded), ok, descriptorBits)
		}
	}

	for bits := uint64(0); bits <= gameModeCustomMask; bits++ {
		decoded, ok := decodeGameModeDescriptor(int64(bits))
		if !ok {
			t.Fatalf("decode custom %#x", bits)
		}
		if _, registered := world.GameModeID(decoded); registered {
			t.Fatalf("custom %#x resolved as a registered game mode", bits)
		}
		encoded, ok := encodeGameModeDescriptor(decoded)
		if !ok || uint64(encoded) != bits {
			t.Fatalf("encode custom %#x = %#x, %v", bits, uint64(encoded), ok)
		}
	}

	// A custom implementation need not be comparable. Dragonfly's registry uses
	// interface keys, so the host must fall back to capability encoding without
	// leaking the resulting map-key panic.
	custom := testCustomGameMode{flags: 0xa5, data: []byte{1}}
	encoded, ok := encodeGameModeDescriptor(custom)
	if !ok || uint64(encoded) != 0xa5 {
		t.Fatalf("encode unregistered game mode = %#x, %v", uint64(encoded), ok)
	}

	for _, descriptorBits := range []uint64{
		1 << 8,
		1 << 62,
		gameModeBuiltinFlag | 999,
	} {
		if mode, ok := decodeGameModeDescriptor(int64(descriptorBits)); ok || mode != nil {
			t.Fatalf("invalid descriptor %#x decoded as %T, %v", descriptorBits, mode, ok)
		}
	}
	if _, ok := encodeGameModeDescriptor(nil); ok {
		t.Fatal("nil game mode encoded")
	}
}

func TestPlayersReadAndSetCustomGameModeDescriptor(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()

		player.SetGameMode(testCustomGameMode{flags: 0xa5, data: []byte{1}})
		state, ok := players.PlayerState(invocation, id, native.PlayerStateGameMode)
		if !ok || uint64(state.Integer) != 0xa5 {
			t.Fatalf("custom game-mode state = %#x, %v", uint64(state.Integer), ok)
		}
		if !players.SetPlayerState(invocation, id, native.PlayerStateGameMode, native.PlayerStateValue{Integer: 0x5a}) {
			t.Fatal("set custom game mode")
		}
		mode := player.GameMode()
		if mode.AllowsEditing() || !mode.AllowsTakingDamage() || mode.CreativeInventory() || !mode.HasCollision() ||
			!mode.AllowsFlying() || mode.AllowsInteraction() || !mode.Visible() || mode.InstantPortalTravel() {
			t.Fatalf("custom game-mode capabilities do not match descriptor: %T", mode)
		}

		before := player.GameMode()
		for _, descriptorBits := range []uint64{1 << 8, 1 << 62, gameModeBuiltinFlag | 999} {
			if players.SetPlayerState(invocation, id, native.PlayerStateGameMode, native.PlayerStateValue{Integer: int64(descriptorBits)}) {
				t.Fatalf("accepted invalid game-mode descriptor %#x", descriptorBits)
			}
			if player.GameMode() != before {
				t.Fatalf("invalid descriptor %#x changed the game mode", descriptorBits)
			}
		}
	})
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
		kinematics, ok := players.PlayerKinematics(invocation, id)
		if !ok || kinematics.Position != (native.Vec3{X: 5, Y: 5, Z: 6}) ||
			kinematics.Rotation.Yaw != 20 || kinematics.Rotation.Pitch != 5 {
			t.Fatalf("kinematics = %+v ok=%v", kinematics, ok)
		}
		if !players.TransformPlayer(invocation, id, native.PlayerTransformDisplace, native.Vec3{X: 0.25}, 0, 0) || player.Position().X() != 5.25 {
			t.Fatalf("displaced position = %v", player.Position())
		}
		if !players.TransformPlayer(invocation, id, native.PlayerTransformVelocity, native.Vec3{Y: 1}, 0, 0) || player.Velocity().Y() != 1 {
			t.Fatalf("velocity = %v", player.Velocity())
		}
		if players.TransformPlayer(invocation, id, native.PlayerTransformDisplace+1, native.Vec3{}, 0, 0) {
			t.Fatal("unknown transform kind accepted")
		}
	})
}

func TestPlayersTransformKeepsDragonflyNumericSemantics(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()
		if !players.TransformPlayer(invocation, id, native.PlayerTransformVelocity, native.Vec3{Y: math.NaN()}, 0, 0) {
			t.Fatal("non-finite velocity rejected unlike Dragonfly")
		}
		if !math.IsNaN(player.Velocity().Y()) {
			t.Fatalf("velocity = %v", player.Velocity())
		}
	})
}

func TestPlayersKnockBackUsesDragonflySemantics(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()
		player.Teleport(mgl64.Vec3{4, 5, 6})
		if !players.KnockBackPlayer(invocation, id, native.Vec3{X: 3, Y: 99, Z: 6}, 0.4, 0.25) {
			t.Fatal("knockback failed")
		}
		if velocity := player.Velocity(); velocity != (mgl64.Vec3{0.4, 0.25, 0}) {
			t.Fatalf("velocity = %v", velocity)
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

func TestPlayersRunPresentationActions(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()

		for _, action := range []struct {
			kind  native.PlayerActionKind
			value native.PlayerStateValue
		}{
			{native.PlayerActionEnableInstantRespawn, native.PlayerStateValue{}},
			{native.PlayerActionDisableInstantRespawn, native.PlayerStateValue{}},
			{native.PlayerActionShowCoordinates, native.PlayerStateValue{}},
			{native.PlayerActionHideCoordinates, native.PlayerStateValue{}},
			{native.PlayerActionSendSleepingIndicator, native.PlayerStateValue{Integer: 1, Number: 2}},
		} {
			if _, ok := players.PlayerAction(invocation, id, action.kind, action.value); !ok {
				t.Fatalf("action %d failed", action.kind)
			}
		}
		for _, invalid := range []native.PlayerStateValue{
			{Integer: math.MaxInt64, Number: 1},
			{Integer: 1, Number: 1.5},
		} {
			if _, ok := players.PlayerAction(invocation, id, native.PlayerActionSendSleepingIndicator, invalid); ok {
				t.Fatalf("invalid sleeping indicator accepted: %+v", invalid)
			}
		}
	})
}

func TestPlayersRunZeroArgumentActions(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()

		for _, action := range []native.PlayerActionKind{
			native.PlayerActionAbortBreaking,
			native.PlayerActionClearInputLocks,
			native.PlayerActionFinishBreaking,
			native.PlayerActionJump,
			native.PlayerActionMoveItemsToInventory,
			native.PlayerActionPunchAir,
			native.PlayerActionReleaseItem,
			native.PlayerActionRemoveAllDebugShapes,
			native.PlayerActionSwingArm,
			native.PlayerActionUseItem,
			native.PlayerActionWake,
		} {
			if _, ok := players.PlayerAction(invocation, id, action, native.PlayerStateValue{}); !ok {
				t.Fatalf("action %d failed", action)
			}
		}
		if _, ok := players.PlayerAction(invocation, id, native.PlayerActionKind(99), native.PlayerStateValue{}); ok {
			t.Fatal("unknown player action accepted")
		}
	})
}

func TestPlayerControlMappings(t *testing.T) {
	element, elementOK := playerHudElement(0)
	lock, lockOK := playerInputLock(4096)
	if !elementOK || element != hud.PaperDoll() || !lockOK || lock != input.MoveRight() {
		t.Fatalf("decoded controls = %#v/%v %#v/%v", element, elementOK, lock, lockOK)
	}
	if _, ok := playerHudElement(-1); ok {
		t.Fatal("invalid HUD element decoded")
	}
	if _, ok := playerInputLock(8); ok {
		t.Fatal("reserved input lock decoded")
	}
}

func TestPlayersReadPresentationStrings(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()

		player.SetNameTag("Name tag")
		player.SetScoreTag("Score tag")
		nameTag, nameOK := players.PlayerString(invocation, id, native.PlayerStringNameTag)
		scoreTag, scoreOK := players.PlayerString(invocation, id, native.PlayerStringScoreTag)
		deviceID, deviceIDOK := players.PlayerString(invocation, id, native.PlayerStringDeviceID)
		deviceModel, deviceModelOK := players.PlayerString(invocation, id, native.PlayerStringDeviceModel)
		selfSignedID, selfSignedIDOK := players.PlayerString(invocation, id, native.PlayerStringSelfSignedID)
		locale, localeOK := players.PlayerString(invocation, id, native.PlayerStringLocale)
		if !nameOK || !scoreOK || nameTag != "Name tag" || scoreTag != "Score tag" {
			t.Fatalf("name=%q/%v score=%q/%v", nameTag, nameOK, scoreTag, scoreOK)
		}
		if !deviceIDOK || !deviceModelOK || !selfSignedIDOK || !localeOK ||
			deviceID != "" || deviceModel != "" || selfSignedID != "" || locale != player.Locale().String() {
			t.Fatalf("connection strings=%q/%v %q/%v %q/%v %q/%v", deviceID, deviceIDOK,
				deviceModel, deviceModelOK, selfSignedID, selfSignedIDOK, locale, localeOK)
		}
		if _, addrOK := players.PlayerString(invocation, id, native.PlayerStringAddrNetwork); addrOK {
			t.Fatal("nop player returned an address")
		}
		if !players.SendPlayerToast(invocation, id, "Title", "Message") {
			t.Fatal("send toast failed")
		}
		if _, ok := players.PlayerString(invocation, id, native.PlayerStringKind(99)); ok {
			t.Fatal("unknown player string kind accepted")
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
			{native.PlayerStateSpeed, native.PlayerStateValue{Number: 0.2}},
			{native.PlayerStateFlightSpeed, native.PlayerStateValue{Number: 0.1}},
			{native.PlayerStateVerticalFlightSpeed, native.PlayerStateValue{Number: 1.5}},
			{native.PlayerStateFallDistance, native.PlayerStateValue{}},
			{native.PlayerStateAbsorption, native.PlayerStateValue{Number: 5}},
			{native.PlayerStateSprinting, native.PlayerStateValue{}},
			{native.PlayerStateSneaking, native.PlayerStateValue{}},
			{native.PlayerStateSwimming, native.PlayerStateValue{}},
			{native.PlayerStateCrawling, native.PlayerStateValue{}},
			{native.PlayerStateGliding, native.PlayerStateValue{}},
			{native.PlayerStateFlying, native.PlayerStateValue{}},
			{native.PlayerStateOnFireDuration, native.PlayerStateValue{Integer: int64(2 * time.Second)}},
			{native.PlayerStateAirSupply, native.PlayerStateValue{Integer: int64(10 * time.Second)}},
			{native.PlayerStateMaxAirSupply, native.PlayerStateValue{Integer: int64(15 * time.Second)}},
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
		finalDamage, ok := players.FinalPlayerDamage(invocation, id, -4, native.DamageSource{Kind: native.DamageSourceInstant})
		if !ok || finalDamage != 0 {
			t.Fatalf("final damage = %v ok=%v", finalDamage, ok)
		}
		healed, ok := players.HealPlayer(invocation, id, 3, native.HealingSource{Kind: native.HealingSourceInstant})
		if !ok || healed != 3 {
			t.Fatalf("healed = %v ok=%v", healed, ok)
		}
		creativeDescriptor, ok := encodeGameModeDescriptor(world.GameModeCreative)
		if !ok {
			t.Fatal("encode creative game mode")
		}
		if !players.SetPlayerState(invocation, id, native.PlayerStateGameMode, native.PlayerStateValue{Integer: creativeDescriptor}) {
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
		speed, _ := players.PlayerState(invocation, id, native.PlayerStateSpeed)
		flightSpeed, _ := players.PlayerState(invocation, id, native.PlayerStateFlightSpeed)
		verticalFlightSpeed, _ := players.PlayerState(invocation, id, native.PlayerStateVerticalFlightSpeed)
		fallDistance, fallDistanceOK := players.PlayerState(invocation, id, native.PlayerStateFallDistance)
		absorption, absorptionOK := players.PlayerState(invocation, id, native.PlayerStateAbsorption)
		dead, deadOK := players.PlayerState(invocation, id, native.PlayerStateDead)
		onGround, onGroundOK := players.PlayerState(invocation, id, native.PlayerStateOnGround)
		eyeHeight, eyeHeightOK := players.PlayerState(invocation, id, native.PlayerStateEyeHeight)
		torsoHeight, torsoHeightOK := players.PlayerState(invocation, id, native.PlayerStateTorsoHeight)
		breathing, breathingOK := players.PlayerState(invocation, id, native.PlayerStateBreathing)
		onFire, onFireOK := players.PlayerState(invocation, id, native.PlayerStateOnFireDuration)
		fireProof, fireProofOK := players.PlayerState(invocation, id, native.PlayerStateFireProof)
		air, airOK := players.PlayerState(invocation, id, native.PlayerStateAirSupply)
		maxAir, maxAirOK := players.PlayerState(invocation, id, native.PlayerStateMaxAirSupply)
		if gameMode.Integer != creativeDescriptor || food.Integer != 12 || maxHealth.Number != 40 || health.Number != 23 || level.Integer != 12 || math.Abs(progress.Number-0.5) > 0.02 || scale.Number != 1.5 || invisible.Integer != 1 || immobile.Integer != 1 || speed.Number != 0.2 || flightSpeed.Number != 0.1 || verticalFlightSpeed.Number != 1.5 || fallDistance.Number != 0 || absorption.Number != 1 || dead.Integer != 0 || (onGround.Integer != 0 && onGround.Integer != 1) || eyeHeight.Number != 1.62 || torsoHeight.Number != 1.52 || (breathing.Integer != 0 && breathing.Integer != 1) || onFire.Integer != int64(2*time.Second) || fireProof.Integer != 1 || air.Integer != int64(10*time.Second) || maxAir.Integer != int64(15*time.Second) || !fallDistanceOK || !absorptionOK || !deadOK || !onGroundOK || !eyeHeightOK || !torsoHeightOK || !breathingOK || !onFireOK || !fireProofOK || !airOK || !maxAirOK {
			t.Fatalf("state mismatch: game=%+v food=%+v max=%+v health=%+v level=%+v progress=%+v scale=%+v invisible=%+v immobile=%+v speed=%+v flight=%+v vertical=%+v fall=%+v absorption=%+v dead=%+v ground=%+v eye=%+v torso=%+v breathing=%+v", gameMode, food, maxHealth, health, level, progress, scale, invisible, immobile, speed, flightSpeed, verticalFlightSpeed, fallDistance, absorption, dead, onGround, eyeHeight, torsoHeight, breathing)
		}
		for _, kind := range []native.PlayerStateKind{
			native.PlayerStateSprinting, native.PlayerStateSneaking, native.PlayerStateSwimming,
			native.PlayerStateCrawling, native.PlayerStateGliding, native.PlayerStateFlying,
		} {
			value, ok := players.PlayerState(invocation, id, kind)
			if !ok || value.Integer != 0 {
				t.Fatalf("activity state %d = %+v, %v", kind, value, ok)
			}
		}
		if !players.SendPlayerText(invocation, id, native.PlayerTextNameTag, "C# Player") || player.NameTag() != "C# Player" {
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
			Potency: 1,
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
		player.SetGameMode(world.GameModeSurvival)
		before := player.Health()
		if !players.ChangePlayerEffect(invocation, id, native.PlayerEffectAdd, native.PlayerEffect{
			Type: native.EffectInstantHealth, Level: 1, Potency: 0.5,
		}) {
			t.Fatal("instant effect failed")
		}
		if got := player.Health(); got != before+2 {
			t.Fatalf("health after half-potency instant health = %v, want %v", got, before+2)
		}

		const customID = 32_000
		effect.Register(customID, testLastingEffect{})
		if !players.ChangePlayerEffect(invocation, id, native.PlayerEffectAdd, native.PlayerEffect{
			Type: customID, Level: 1, Duration: time.Second, Potency: 1,
		}) {
			t.Fatal("registered effect failed")
		}
		if applied, ok := player.Effect(testLastingEffect{}); !ok || applied.Level() != 1 {
			t.Fatalf("registered effect = %+v ok=%v", applied, ok)
		}
		if !players.ChangePlayerEffect(invocation, id, native.PlayerEffectAdd, native.PlayerEffect{
			Type: native.EffectRegeneration, Level: 1, Duration: time.Second,
			Potency: 1, Ambient: true, ParticlesHidden: true,
		}) {
			t.Fatal("ambient effect failed")
		}
		if applied, ok := player.Effect(effect.Regeneration); !ok || !applied.Ambient() || !applied.ParticlesHidden() {
			t.Fatalf("ambient effect = %+v ok=%v", applied, ok)
		}
		if !players.ChangePlayerEffect(invocation, id, native.PlayerEffectAdd, native.PlayerEffect{
			Type: native.EffectFireResistance, Level: 1, Potency: 1, Infinite: true,
		}) {
			t.Fatal("infinite effect failed")
		}
		if applied, ok := player.Effect(effect.FireResistance); !ok || !applied.Infinite() {
			t.Fatalf("infinite effect = %+v ok=%v", applied, ok)
		}
		if !players.ChangePlayerEffect(invocation, id, native.PlayerEffectAdd, native.PlayerEffect{
			Type: customID, Level: 1, Potency: 0.5,
		}) {
			t.Fatal("registered lasting effect rejected with instant potency")
		}
		if players.ChangePlayerEffect(invocation, id, native.PlayerEffectAdd, native.PlayerEffect{
			Type: native.EffectSpeed, Level: 0, Potency: 1,
		}) {
			t.Fatal("zero-level effect accepted")
		}
		if !players.SetPlayerState(invocation, id, native.PlayerStateMaxHealth, native.PlayerStateValue{Number: -5}) {
			t.Fatal("negative max-health change failed")
		}
		if players.SetPlayerState(invocation, id, native.PlayerStateSprinting, native.PlayerStateValue{Integer: 2}) {
			t.Fatal("invalid sprinting state accepted")
		}
		clampedMaxHealth, _ := players.PlayerState(invocation, id, native.PlayerStateMaxHealth)
		if clampedMaxHealth.Number != 1 {
			t.Fatalf("clamped max health=%v, want 1", clampedMaxHealth.Number)
		}
		if !players.SetPlayerState(invocation, id, native.PlayerStateExperienceProgress, native.PlayerStateValue{Number: math.NaN()}) {
			t.Fatal("NaN experience progress rejected unlike Dragonfly")
		}
	})
}

func TestPlayersRunHungerAndExperienceActions(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := players.Register(player, 1)
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()

		player.SetFood(10)
		for _, action := range []struct {
			kind  native.PlayerActionKind
			value native.PlayerStateValue
		}{
			{native.PlayerActionAddFood, native.PlayerStateValue{Integer: 2}},
			{native.PlayerActionSaturate, native.PlayerStateValue{Integer: 1, Number: 2}},
			{native.PlayerActionExhaust, native.PlayerStateValue{}},
			{native.PlayerActionResetEnchantmentSeed, native.PlayerStateValue{}},
		} {
			if _, ok := players.PlayerAction(invocation, id, action.kind, action.value); !ok {
				t.Fatalf("action %d failed", action.kind)
			}
		}
		food, _ := players.PlayerState(invocation, id, native.PlayerStateFood)
		seed, seedOK := players.PlayerState(invocation, id, native.PlayerStateEnchantmentSeed)
		if food.Integer != 13 || !seedOK {
			t.Fatalf("food=%+v seed=%+v ok=%v", food, seed, seedOK)
		}

		added, ok := players.PlayerAction(invocation, id, native.PlayerActionAddExperience, native.PlayerStateValue{Integer: 7})
		if !ok || added.Integer != 7 {
			t.Fatalf("added=%+v ok=%v", added, ok)
		}
		if _, ok = players.PlayerAction(invocation, id, native.PlayerActionRemoveExperience, native.PlayerStateValue{Integer: 2}); !ok {
			t.Fatal("remove experience failed")
		}
		experience, _ := players.PlayerState(invocation, id, native.PlayerStateExperience)
		canCollect, _ := players.PlayerState(invocation, id, native.PlayerStateCanCollectExperience)
		collected, ok := players.PlayerAction(invocation, id, native.PlayerActionCollectExperience, native.PlayerStateValue{})
		cannotCollect, _ := players.PlayerState(invocation, id, native.PlayerStateCanCollectExperience)
		if experience.Integer != 5 || canCollect.Integer != 1 || !ok || collected.Integer != 1 || cannotCollect.Integer != 0 {
			t.Fatalf("experience=%+v can=%+v collected=%+v ok=%v cannot=%+v", experience, canCollect, collected, ok, cannotCollect)
		}
	})
}

func TestPlayersSetPlayerStateSchedulesAcrossWorldInvocation(t *testing.T) {
	source := world.Config{Synchronous: true}.New()
	destination := world.Config{Synchronous: true}.New()
	t.Cleanup(func() {
		_ = source.Close()
		_ = destination.Close()
	})
	players := NewPlayers()
	spawn := func(id uuid.UUID, name string) *world.EntityHandle {
		position := mgl64.Vec3{1, 64, 1}
		return world.EntitySpawnOpts{ID: id, Position: position}.New(
			player.Type,
			player.Config{UUID: id, Name: name, Position: position},
		)
	}
	sourceHandle := spawn(uuid.MustParse("a6423d6a-cb19-4cff-8d30-362027f20f00"), "Source")
	destinationHandle := spawn(uuid.MustParse("581b4619-d662-4514-bd27-d2c30a94bb08"), "Destination")
	var destinationID native.PlayerID
	if err := destination.Do(func(tx *world.Tx) {
		destinationID = players.Register(tx.AddEntity(destinationHandle).(*player.Player), 75)
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := source.Do(func(tx *world.Tx) {
		players.Register(tx.AddEntity(sourceHandle).(*player.Player), 74)
		invocation, leave := players.BeginInvocation(tx)
		defer leave()
		if !players.SetPlayerState(invocation, destinationID, native.PlayerStateExperienceLevel, native.PlayerStateValue{Integer: 20}) {
			t.Fatal("cross-world experience level write was not scheduled")
		}
		if !players.SetPlayerState(invocation, destinationID, native.PlayerStateExperienceProgress, native.PlayerStateValue{Number: 0.5}) {
			t.Fatal("cross-world experience progress write was not scheduled")
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	level, ok := players.PlayerState(0, destinationID, native.PlayerStateExperienceLevel)
	if !ok || level.Integer != 20 {
		t.Fatalf("level = %+v, %v", level, ok)
	}
	progress, ok := players.PlayerState(0, destinationID, native.PlayerStateExperienceProgress)
	if !ok || math.Abs(progress.Number-0.5) > 0.02 {
		t.Fatalf("progress = %+v, %v", progress, ok)
	}
}

type testLastingEffect struct{}

func (testLastingEffect) RGBA() color.RGBA                  { return color.RGBA{} }
func (testLastingEffect) Apply(world.Entity, effect.Effect) {}
func (testLastingEffect) Start(world.Entity, int)           {}
func (testLastingEffect) End(world.Entity, int)             {}

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
