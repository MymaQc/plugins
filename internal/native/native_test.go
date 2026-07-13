package native

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

func nativeArtifacts(t testing.TB) (string, string) {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	extension := ".so"
	if runtime.GOOS == "darwin" {
		extension = ".dylib"
	} else if runtime.GOOS == "windows" {
		extension = ".dll"
	}
	library := filepath.Join(root, "build", "lib", "libdragonfly_plugin_runtime"+extension)
	plugins := filepath.Join(root, "build", "plugins")
	if _, err := os.Stat(library); err != nil {
		t.Skipf("native runtime not built: run make build-native (%v)", err)
	}
	return library, plugins
}

func openTestRuntime(t testing.TB) *Runtime {
	t.Helper()
	library, plugins := nativeArtifacts(t)
	runtime, err := Open(library, plugins)
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	t.Cleanup(runtime.Disable)
	return runtime
}

type recordingHost struct {
	noopHost
	player            PlayerID
	texts             []string
	kinds             []PlayerTextKind
	title             PlayerTitle
	scoreboard        PlayerScoreboard
	scoreboardRemoved bool
	rotation          Rotation
	transforms        []PlayerTransformKind
	vectors           []Vec3
	yaws              []float64
	pitches           []float64
	states            []PlayerStateKind
	values            []PlayerStateValue
	state             PlayerStateValue
	reads             []PlayerStateKind
	heals             []struct {
		Health float64
		Source HealingSource
	}
	healed float64
	hurts  []struct {
		Damage float64
		Source DamageSource
	}
	hurtResult    PlayerHurtResult
	effectOps     []PlayerEffectOperation
	effects       []PlayerEffect
	entities      []EntityID
	visible       []bool
	skin          PlayerSkin
	setSkins      []PlayerSkin
	inventoryItem ItemStack
	inventorySets []struct {
		Inventory InventoryID
		Slot      uint32
		Item      ItemStack
	}
	inventoryAdds     []ItemStack
	forms             []PlayerForm
	formClosed        bool
	worldOpened       string
	worldDimension    WorldDimension
	worldID           WorldID
	worldLookup       string
	worldLookupOK     bool
	worldName         string
	worldBlock        WorldBlock
	worldBlockOK      bool
	worldBlockPos     BlockPos
	worldBlockSet     WorldBlock
	worldSaved        bool
	worldUnloaded     bool
	worldTime         int64
	worldSpawn        BlockPos
	entityStateID     EntityID
	entityState       EntityState
	entitySpawns      []EntitySpawn
	spawnedEntity     EntityID
	worldEntityIDs    []EntityID
	worldPlayerIDs    []PlayerID
	particleWorldID   WorldID
	particlePositions []Vec3
	particles         []WorldParticle
	worldSoundID      WorldID
	worldSoundPos     []Vec3
	worldSounds       []WorldSound
	playerSounds      []WorldSound
}

func (h *recordingHost) SendPlayerText(_ InvocationID, player PlayerID, kind PlayerTextKind, message string) bool {
	h.player = player
	h.kinds = append(h.kinds, kind)
	h.texts = append(h.texts, message)
	return true
}

func (h *recordingHost) SendPlayerTitle(_ InvocationID, player PlayerID, title PlayerTitle) bool {
	h.player, h.title = player, title
	return true
}

func (h *recordingHost) SendPlayerScoreboard(_ InvocationID, player PlayerID, scoreboard PlayerScoreboard) bool {
	h.player, h.scoreboard = player, scoreboard
	return true
}

func (h *recordingHost) RemovePlayerScoreboard(_ InvocationID, player PlayerID) bool {
	h.player, h.scoreboardRemoved = player, true
	return true
}
func (h *recordingHost) SendPlayerForm(_ InvocationID, _ PlayerID, form PlayerForm) bool {
	h.forms = append(h.forms, form)
	return true
}
func (h *recordingHost) ClosePlayerForm(InvocationID, PlayerID) bool {
	h.formClosed = true
	return true
}

func (h *recordingHost) TransformPlayer(_ InvocationID, _ PlayerID, kind PlayerTransformKind, vector Vec3, yaw, pitch float64) bool {
	h.transforms = append(h.transforms, kind)
	h.vectors = append(h.vectors, vector)
	h.yaws = append(h.yaws, yaw)
	h.pitches = append(h.pitches, pitch)
	return true
}

func (h *recordingHost) PlayerRotation(InvocationID, PlayerID) (Rotation, bool) {
	return h.rotation, true
}

func (h *recordingHost) SetPlayerState(_ InvocationID, _ PlayerID, kind PlayerStateKind, value PlayerStateValue) bool {
	h.states = append(h.states, kind)
	h.values = append(h.values, value)
	return true
}

func (h *recordingHost) PlayerState(_ InvocationID, _ PlayerID, kind PlayerStateKind) (PlayerStateValue, bool) {
	h.reads = append(h.reads, kind)
	return h.state, true
}

func (h *recordingHost) HealPlayer(_ InvocationID, _ PlayerID, health float64, source HealingSource) (float64, bool) {
	h.heals = append(h.heals, struct {
		Health float64
		Source HealingSource
	}{health, source})
	return h.healed, true
}

func (h *recordingHost) HurtPlayer(_ InvocationID, _ PlayerID, damage float64, source DamageSource) (PlayerHurtResult, bool) {
	h.hurts = append(h.hurts, struct {
		Damage float64
		Source DamageSource
	}{damage, source})
	return h.hurtResult, true
}

func (h *recordingHost) ChangePlayerEffect(_ InvocationID, _ PlayerID, operation PlayerEffectOperation, effect PlayerEffect) bool {
	h.effectOps = append(h.effectOps, operation)
	h.effects = append(h.effects, effect)
	return true
}

func (h *recordingHost) SetPlayerEntityVisible(_ InvocationID, _ PlayerID, entity EntityID, visible bool) bool {
	h.entities = append(h.entities, entity)
	h.visible = append(h.visible, visible)
	return true
}

func (h *recordingHost) PlayerSkin(InvocationID, PlayerID) (PlayerSkin, bool) {
	return h.skin, true
}

func (h *recordingHost) SetPlayerSkin(_ InvocationID, _ PlayerID, skin PlayerSkin) bool {
	h.setSkins = append(h.setSkins, skin)
	return true
}

func (h *recordingHost) InventorySize(InvocationID, InventoryID) (uint32, bool) { return 36, true }
func (h *recordingHost) InventoryItem(InvocationID, InventoryID, uint32) (ItemStack, bool) {
	return h.inventoryItem, true
}
func (h *recordingHost) SetInventoryItem(_ InvocationID, inventory InventoryID, slot uint32, item ItemStack) bool {
	h.inventorySets = append(h.inventorySets, struct {
		Inventory InventoryID
		Slot      uint32
		Item      ItemStack
	}{inventory, slot, item})
	return true
}
func (h *recordingHost) AddInventoryItem(_ InvocationID, _ InventoryID, item ItemStack) (uint32, bool) {
	h.inventoryAdds = append(h.inventoryAdds, item)
	return item.Count, true
}
func (h *recordingHost) ClearInventory(InvocationID, InventoryID) bool { return true }
func (h *recordingHost) HeldItem(InvocationID, PlayerID, uint32) (ItemStack, bool) {
	return h.inventoryItem, true
}
func (h *recordingHost) SetHeldItems(InvocationID, PlayerID, ItemStack, ItemStack) bool {
	return true
}
func (h *recordingHost) SetHeldSlot(InvocationID, PlayerID, uint32) bool { return true }
func (h *recordingHost) OpenWorld(_ InvocationID, name string, dimension WorldDimension) (WorldID, bool) {
	h.worldOpened, h.worldDimension = name, dimension
	return h.worldID, h.worldID != 0
}
func (h *recordingHost) WorldByName(_ InvocationID, name string) (WorldID, bool) {
	h.worldLookup = name
	return h.worldID, h.worldLookupOK && h.worldID != 0
}
func (h *recordingHost) WorldName(_ InvocationID, id WorldID) (string, bool) {
	return h.worldName, id == h.worldID && h.worldName != ""
}
func (h *recordingHost) WorldBlock(_ InvocationID, id WorldID, position BlockPos) (WorldBlock, bool) {
	h.worldBlockPos = position
	return h.worldBlock, id == h.worldID && h.worldBlockOK
}
func (h *recordingHost) SetWorldBlock(_ InvocationID, id WorldID, position BlockPos, value WorldBlock) bool {
	h.worldBlockPos, h.worldBlockSet = position, value
	return id == h.worldID
}
func (h *recordingHost) SaveWorld(_ InvocationID, id WorldID) bool {
	h.worldSaved = id == h.worldID
	return h.worldSaved
}
func (h *recordingHost) UnloadWorld(_ InvocationID, id WorldID) bool {
	h.worldUnloaded = id == h.worldID
	return h.worldUnloaded
}
func (h *recordingHost) SetWorldTime(_ InvocationID, _ WorldID, value int64) bool {
	h.worldTime = value
	return true
}
func (h *recordingHost) SetWorldSpawn(_ InvocationID, _ WorldID, position BlockPos) bool {
	h.worldSpawn = position
	return true
}
func (h *recordingHost) EntityState(_ InvocationID, id EntityID) (EntityState, bool) {
	h.entityStateID = id
	return h.entityState, h.entityState.Type != ""
}
func (h *recordingHost) SpawnWorldEntity(_ InvocationID, id WorldID, value EntitySpawn) (EntityID, bool) {
	h.entitySpawns = append(h.entitySpawns, value)
	return h.spawnedEntity, id == h.worldID && h.spawnedEntity.Generation != 0
}
func (h *recordingHost) WorldEntities(_ InvocationID, id WorldID) ([]EntityID, bool) {
	return append([]EntityID(nil), h.worldEntityIDs...), id == h.worldID
}
func (h *recordingHost) WorldPlayers(_ InvocationID, id WorldID) ([]PlayerID, bool) {
	return append([]PlayerID(nil), h.worldPlayerIDs...), id == h.worldID
}
func (h *recordingHost) AddWorldParticle(_ InvocationID, id WorldID, position Vec3, value WorldParticle) bool {
	h.particleWorldID = id
	h.particlePositions = append(h.particlePositions, position)
	h.particles = append(h.particles, value)
	return id == h.worldID
}
func (h *recordingHost) PlayWorldSound(_ InvocationID, id WorldID, position Vec3, value WorldSound) bool {
	h.worldSoundID = id
	h.worldSoundPos = append(h.worldSoundPos, position)
	h.worldSounds = append(h.worldSounds, value)
	return id == h.worldID
}
func (h *recordingHost) PlayPlayerSound(_ InvocationID, player PlayerID, value WorldSound) bool {
	h.player = player
	h.playerSounds = append(h.playerSounds, value)
	return true
}

func TestPluginCanMessagePlayer(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	host := &recordingHost{}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	id := PlayerID{Generation: 42}
	if _, err := runtime.HandlePlayerJoin(0, PlayerJoinInput{Player: id, Name: "TestPlayer"}, false); err != nil {
		t.Fatal(err)
	}
	wantTexts := []string{"Welcome from a Rust plugin.", "Rust tip", "Rust popup", "Rust jukebox popup"}
	wantKinds := []PlayerTextKind{PlayerTextMessage, PlayerTextTip, PlayerTextPopup, PlayerTextJukeboxPopup}
	if host.player != id || !slices.Equal(host.texts, wantTexts) || !slices.Equal(host.kinds, wantKinds) {
		t.Fatalf("host calls = player %+v kinds %v texts %q", host.player, host.kinds, host.texts)
	}
	if host.title.Text != "Rust plugin" || host.title.Subtitle != "Native Dragonfly" || host.title.Duration != 2*time.Second {
		t.Fatalf("title = %+v", host.title)
	}
	wantScoreboard := PlayerScoreboard{
		Name: "Rust scoreboard", Lines: []string{"Welcome, TestPlayer", "Native plugins"}, Padding: false, Descending: true,
	}
	if !reflect.DeepEqual(host.scoreboard, wantScoreboard) {
		t.Fatalf("scoreboard = %+v, want %+v", host.scoreboard, wantScoreboard)
	}
	if err := runtime.HandlePlayerQuit(0, PlayerQuitInput{Player: id, Name: "TestPlayer"}); err != nil {
		t.Fatal(err)
	}
	if !host.scoreboardRemoved {
		t.Fatal("scoreboard was not removed on quit")
	}
}

func TestFormResponseRoundTrip(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	host := &recordingHost{}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	id := PlayerID{Generation: 77}
	if _, err := runtime.HandlePlayerJoin(0, PlayerJoinInput{Player: id, Name: "FormPlayer"}, false); err != nil {
		t.Fatal(err)
	}
	if len(host.forms) == 0 || !bytes.Contains(host.forms[len(host.forms)-1].RequestJSON, []byte(`"type":"form"`)) {
		t.Fatalf("menu form not sent: %+v", host.forms)
	}
	menu := host.forms[len(host.forms)-1]
	if !CompletePlayerForm(menu.ID, 0, id, false, []byte("0")) {
		t.Fatal("menu response rejected")
	}
	if len(host.forms) < 2 || !bytes.Contains(host.forms[len(host.forms)-1].RequestJSON, []byte(`"type":"custom_form"`)) {
		t.Fatalf("custom form not sent after menu response: %+v", host.forms)
	}
	custom := host.forms[len(host.forms)-1]
	response := []byte(`[null,null,null,"Alex",true,5,1,2]`)
	if !CompletePlayerForm(custom.ID, 0, id, false, response) {
		t.Fatal("custom response rejected")
	}
	if !slices.Contains(host.texts, "Hello Alex: volume 5, colour #1, speed #2") {
		t.Fatalf("texts = %v", host.texts)
	}
}

func TestFormRegistryIsBoundedAndDrained(t *testing.T) {
	host := registerHost(noopHost{})
	player := PlayerID{Generation: 9}
	dropped := 0
	for index := 0; index < maxFormsPerPlayer; index++ {
		if _, ok := registerForm(host, player, func(InvocationID, PlayerID, bool, []byte) bool { return true }, func() { dropped++ }); !ok {
			t.Fatalf("registration %d rejected", index)
		}
	}
	if _, ok := registerForm(host, player, func(InvocationID, PlayerID, bool, []byte) bool { return true }, func() { dropped++ }); ok {
		t.Fatal("registration beyond per-player bound accepted")
	}
	drainHostForms(host, false)
	if dropped != maxFormsPerPlayer {
		t.Fatalf("dropped = %d, want %d", dropped, maxFormsPerPlayer)
	}
	if _, ok := registerForm(host, player, func(InvocationID, PlayerID, bool, []byte) bool { return true }, func() { dropped++ }); !ok {
		t.Fatal("registry did not reopen after non-closing drain")
	}
	unregisterHost(host)
	if dropped != maxFormsPerPlayer+1 {
		t.Fatalf("dropped after close = %d", dropped)
	}
}

func TestFormDrainWaitsForConcurrentDrop(t *testing.T) {
	host := registerHost(noopHost{})
	player := PlayerID{Generation: 10}
	started, release := make(chan struct{}), make(chan struct{})
	id, ok := registerForm(host, player, func(InvocationID, PlayerID, bool, []byte) bool { return true }, func() { close(started); <-release })
	if !ok {
		t.Fatal("form registration rejected")
	}
	go CancelPlayerForm(id)
	<-started
	drained := make(chan struct{})
	go func() { drainHostForms(host, false); close(drained) }()
	select {
	case <-drained:
		t.Fatal("drain returned while drop callback was in flight")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	select {
	case <-drained:
	case <-time.After(time.Second):
		t.Fatal("drain did not finish")
	}
	unregisterHost(host)
}

func TestClosingFormDrainKeepsRegistrationGateClosed(t *testing.T) {
	host := registerHost(noopHost{})
	drainHostForms(host, true)
	if _, ok := registerForm(host, PlayerID{Generation: 11}, func(InvocationID, PlayerID, bool, []byte) bool {
		return true
	}, func() {}); ok {
		t.Fatal("form registered after closing drain")
	}
	unregisterHost(host)
}

func TestFormRejectsWrongPlayerAndOversizedResponse(t *testing.T) {
	host := registerHost(noopHost{})
	t.Cleanup(func() { unregisterHost(host) })
	player := PlayerID{Generation: 11}
	dropped := 0
	register := func() uint64 {
		id, ok := registerForm(host, player, func(InvocationID, PlayerID, bool, []byte) bool {
			t.Fatal("invalid response reached Rust callback")
			return true
		}, func() { dropped++ })
		if !ok {
			t.Fatal("form registration rejected")
		}
		return id
	}
	if CompletePlayerForm(register(), 0, PlayerID{Generation: 12}, false, []byte("0")) {
		t.Fatal("response from wrong player accepted")
	}
	if CompletePlayerForm(register(), 0, player, false, make([]byte, maxFormJSONBytes+1)) {
		t.Fatal("oversized response accepted")
	}
	if dropped != 2 {
		t.Fatalf("dropped = %d, want 2", dropped)
	}
}

func TestMovementGuard(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.PluginCount() != 13 {
		t.Fatalf("plugin count = %d, want 13", runtime.PluginCount())
	}
	if runtime.Subscriptions()&PlayerMoveSubscription == 0 {
		t.Fatal("movement subscription missing")
	}

	input := PlayerMoveInput{NewPosition: Vec3{X: 10, Y: 64, Z: 10}}
	cancelled, err := runtime.HandlePlayerMove(0, input, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("valid movement cancelled")
	}

	input.NewPosition.Y = -65
	cancelled, err = runtime.HandlePlayerMove(0, input, false)
	if err != nil {
		t.Fatal(err)
	}
	if !cancelled {
		t.Fatal("movement below world was not cancelled")
	}
}

func TestPlayerTransformHostCalls(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	host := &recordingHost{rotation: Rotation{Yaw: 10, Pitch: 5}}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	id := PlayerID{Generation: 7}
	for _, arguments := range []string{"velocity 1 2 3", "face 90 20", "teleport 10 64 20", "move 1 0 -1"} {
		output, err := runtime.HandleCommand(commandNamed(t, commands, "player").Index, CommandInput{
			Source: "TestPlayer", SourceKind: CommandSourcePlayer, SourcePlayer: &id,
			OnlinePlayers: []CommandPlayer{{Player: id, Name: "TestPlayer"}}, Arguments: arguments,
		})
		if err != nil || output.Failed {
			t.Fatalf("%s: output=%+v error=%v", arguments, output, err)
		}
	}
	want := []PlayerTransformKind{PlayerTransformVelocity, PlayerTransformMove, PlayerTransformTeleport, PlayerTransformMove}
	if !slices.Equal(host.transforms, want) {
		t.Fatalf("transforms = %v", host.transforms)
	}
	if host.yaws[1] != 80 || host.pitches[1] != 15 {
		t.Fatalf("face delta = %v/%v", host.yaws[1], host.pitches[1])
	}
	if host.vectors[2] != (Vec3{X: 10, Y: 64, Z: 20}) {
		t.Fatalf("teleport = %+v", host.vectors[2])
	}
}

func TestPlayerStateHostCalls(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	host := &recordingHost{
		state: PlayerStateValue{Number: 12, Integer: 15}, healed: 4,
		hurtResult: PlayerHurtResult{Damage: 3, Vulnerable: true},
	}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	id := PlayerID{Generation: 8}
	for _, arguments := range []string{"gamemode creative", "heal 4", "hurt 3", "food 15", "max-health 40", "experience-level 12", "experience-progress 0.5"} {
		output, err := runtime.HandleCommand(commandNamed(t, commands, "player").Index, CommandInput{
			Source: "TestPlayer", SourceKind: CommandSourcePlayer, SourcePlayer: &id,
			OnlinePlayers: []CommandPlayer{{Player: id, Name: "TestPlayer"}}, Arguments: arguments,
		})
		if err != nil || output.Failed {
			t.Fatalf("%s: output=%+v error=%v", arguments, output, err)
		}
	}
	target := PlayerID{UUID: [16]byte{1, 2, 3}, Generation: 18}
	encoded := "01020300000000000000000000000000:18:0:Target"
	for _, arguments := range []string{
		"heal-food 2 true",
		"hurt-attack 2 " + encoded,
		"hurt-projectile 2 " + encoded,
		"hurt-block 2",
		"hurt-custom 2",
	} {
		output, err := runtime.HandleCommand(commandNamed(t, commands, "player").Index, CommandInput{
			Source: "TestPlayer", SourceKind: CommandSourcePlayer, SourcePlayer: &id,
			OnlinePlayers: []CommandPlayer{{Player: id, Name: "TestPlayer"}, {Player: target, Name: "Target"}}, Arguments: arguments,
		})
		if err != nil || output.Failed {
			t.Fatalf("%s: output=%+v error=%v", arguments, output, err)
		}
	}
	want := []PlayerStateKind{PlayerStateGameMode, PlayerStateFood, PlayerStateMaxHealth, PlayerStateExperienceLevel, PlayerStateExperienceProgress}
	if !slices.Equal(host.states, want) {
		t.Fatalf("states = %v, want %v", host.states, want)
	}
	if host.values[0].Integer != 1 || host.values[1].Integer != 15 || host.values[2].Number != 40 {
		t.Fatalf("values = %+v", host.values)
	}
	if host.values[3].Integer != 12 || host.values[4].Number != 0.5 {
		t.Fatalf("experience values = %+v", host.values[3:])
	}
	if len(host.heals) != 2 || host.heals[0].Health != 4 || host.heals[0].Source.Kind != HealingSourceInstant ||
		host.heals[1].Health != 2 || host.heals[1].Source.Kind != HealingSourceFood || !host.heals[1].Source.Data {
		t.Fatalf("heals = %+v", host.heals)
	}
	if len(host.hurts) != 5 || host.hurts[0].Damage != 3 || host.hurts[0].Source.Kind != DamageSourceInstant {
		t.Fatalf("hurts = %+v", host.hurts)
	}
	if source := host.hurts[1].Source; source.Kind != DamageSourceAttack || source.Entity != (EntityID{UUID: target.UUID, Generation: target.Generation}) {
		t.Fatalf("attack source = %+v", source)
	}
	if source := host.hurts[2].Source; source.Kind != DamageSourceProjectile || source.Entity != (EntityID{UUID: target.UUID, Generation: target.Generation}) || source.SecondaryEntity.Generation != 0 {
		t.Fatalf("projectile source = %+v", source)
	}
	blockSource := host.hurts[3].Source
	if blockSource.Block == nil {
		t.Fatalf("block source = %+v", blockSource)
	}
	if blockSource.Kind != DamageSourceBlock || blockSource.Block.Identifier != "minecraft:cactus" || len(blockSource.Block.PropertiesNBT) == 0 {
		t.Fatalf("block source = %+v", blockSource)
	}
	if source := host.hurts[4].Source; source.Kind != DamageSourceCustom || source.Name != "example:magic" || !source.ReducedByArmour || !source.IgnoresTotem || !source.FireProtection || !source.BlastProtection {
		t.Fatalf("custom source = %+v", source)
	}
	wantReads := []PlayerStateKind{PlayerStateHealth, PlayerStateHealth, PlayerStateFood, PlayerStateMaxHealth, PlayerStateExperienceLevel, PlayerStateExperienceProgress}
	if !slices.Equal(host.reads, wantReads) {
		t.Fatalf("reads = %v, want %v", host.reads, wantReads)
	}
}

func TestPlayerEffectHostCalls(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	host := &recordingHost{}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	id := PlayerID{Generation: 9}
	for _, arguments := range []string{"speed 2 30", "instant-health 1 0.5", "clear-speed"} {
		output, err := runtime.HandleCommand(commandNamed(t, commands, "player").Index, CommandInput{
			Source: "TestPlayer", SourceKind: CommandSourcePlayer, SourcePlayer: &id,
			OnlinePlayers: []CommandPlayer{{Player: id, Name: "TestPlayer"}}, Arguments: arguments,
		})
		if err != nil || output.Failed {
			t.Fatalf("%s: output=%+v error=%v", arguments, output, err)
		}
	}
	if !slices.Equal(host.effectOps, []PlayerEffectOperation{PlayerEffectAdd, PlayerEffectAdd, PlayerEffectRemove}) {
		t.Fatalf("operations = %v", host.effectOps)
	}
	if host.effects[0].Type != EffectSpeed || host.effects[0].Level != 2 || host.effects[0].Duration != 30*time.Second || host.effects[0].Potency != 1 || host.effects[0].Mode != PlayerEffectTimed {
		t.Fatalf("effect = %+v", host.effects[0])
	}
	if host.effects[1].Type != EffectInstantHealth || host.effects[1].Level != 1 || host.effects[1].Potency != 0.5 || host.effects[1].Mode != PlayerEffectInstant {
		t.Fatalf("instant effect = %+v", host.effects[1])
	}
}

func TestPlayerIdentityHostCalls(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	host := &recordingHost{state: PlayerStateValue{Number: 1.5, Integer: 1}}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	id := PlayerID{Generation: 10}
	for _, arguments := range []string{"name-tag Rust Player", "scale 1.5", "invisible true", "immobile true"} {
		output, err := runtime.HandleCommand(commandNamed(t, commands, "player").Index, CommandInput{
			Source: "TestPlayer", SourceKind: CommandSourcePlayer, SourcePlayer: &id,
			OnlinePlayers: []CommandPlayer{{Player: id, Name: "TestPlayer"}}, Arguments: arguments,
		})
		if err != nil || output.Failed {
			t.Fatalf("%s: output=%+v error=%v", arguments, output, err)
		}
	}
	if !slices.Equal(host.kinds, []PlayerTextKind{PlayerTextNameTag}) || !slices.Equal(host.texts, []string{"Rust Player"}) {
		t.Fatalf("text kinds=%v values=%v", host.kinds, host.texts)
	}
	want := []PlayerStateKind{PlayerStateScale, PlayerStateInvisible, PlayerStateImmobile}
	if !slices.Equal(host.states, want) || host.values[0].Number != 1.5 || host.values[1].Integer != 1 || host.values[2].Integer != 1 {
		t.Fatalf("states=%v values=%+v", host.states, host.values)
	}
	if !slices.Equal(host.reads, want) {
		t.Fatalf("reads=%v", host.reads)
	}
}

func TestPlayerSoundAndDisconnectHostCalls(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	host := &recordingHost{}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	id := PlayerID{Generation: 11}
	for _, arguments := range []string{"sound", "disconnect", "kick"} {
		output, err := runtime.HandleCommand(commandNamed(t, commands, "player").Index, CommandInput{
			Source: "TestPlayer", SourceKind: CommandSourcePlayer, SourcePlayer: &id,
			OnlinePlayers: []CommandPlayer{{Player: id, Name: "TestPlayer"}}, Arguments: arguments,
		})
		if err != nil || output.Failed {
			t.Fatalf("%s: output=%+v error=%v", arguments, output, err)
		}
	}
	if len(host.playerSounds) != 1 || host.playerSounds[0].Kind != SoundLevelUp {
		t.Fatalf("sounds=%+v", host.playerSounds)
	}
	wantKinds := []PlayerTextKind{PlayerTextDisconnect, PlayerTextKick}
	wantTexts := []string{"Disconnected by Rust plugin.", "Kicked by Rust plugin."}
	if !slices.Equal(host.kinds, wantKinds) || !slices.Equal(host.texts, wantTexts) {
		t.Fatalf("kinds=%v texts=%v", host.kinds, host.texts)
	}
}

func TestPlayerEntityVisibilityHostCalls(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	host := &recordingHost{}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	viewer := PlayerID{Generation: 12}
	target := PlayerID{UUID: [16]byte{1, 2, 3}, Generation: 13}
	encoded := "01020300000000000000000000000000:13:0:Target"
	for _, arguments := range []string{"hide " + encoded, "show " + encoded} {
		output, err := runtime.HandleCommand(commandNamed(t, commands, "player").Index, CommandInput{
			Source: "Viewer", SourceKind: CommandSourcePlayer, SourcePlayer: &viewer,
			OnlinePlayers: []CommandPlayer{{Player: viewer, Name: "Viewer"}, {Player: target, Name: "Target"}}, Arguments: arguments,
		})
		if err != nil || output.Failed {
			t.Fatalf("%s: output=%+v error=%v", arguments, output, err)
		}
	}
	want := EntityID{UUID: target.UUID, Generation: target.Generation}
	if len(host.entities) != 2 || host.entities[0] != want || host.entities[1] != want || !slices.Equal(host.visible, []bool{false, true}) {
		t.Fatalf("entities=%+v visible=%v", host.entities, host.visible)
	}
}

func TestPlayerSkinRoundTrip(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	want := PlayerSkin{
		Width: 64, Height: 64, Persona: true,
		PlayFabID: "playfab-id", FullID: "full-skin-id",
		Pixels:       []byte{0, 1, 2, 127, 128, 254, 255},
		ModelDefault: "geometry.humanoid.custom", ModelAnimatedFace: "geometry.animated_face",
		Model:     []byte(`{"geometry":{"description":{"identifier":"geometry.test"}}}`),
		CapeWidth: 64, CapeHeight: 32, CapePixels: []byte{9, 8, 7, 0, 255},
		Animations: []SkinAnimation{
			{Width: 32, Height: 32, Type: 0, FrameCount: 7, Expression: -3, Pixels: []byte{1, 3, 5, 7}},
			{Width: 128, Height: 128, Type: 2, FrameCount: 1 << 33, Expression: 1 << 34, Pixels: []byte{2, 4, 6, 8}},
		},
	}
	host := &recordingHost{skin: want}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	id := PlayerID{Generation: 14}
	output, err := runtime.HandleCommand(commandNamed(t, commands, "player").Index, CommandInput{
		Source: "TestPlayer", SourceKind: CommandSourcePlayer, SourcePlayer: &id,
		OnlinePlayers: []CommandPlayer{{Player: id, Name: "TestPlayer"}}, Arguments: "skin-copy",
	})
	if err != nil || output.Failed {
		t.Fatalf("output=%+v error=%v", output, err)
	}
	if len(host.setSkins) != 1 || !reflect.DeepEqual(host.setSkins[0], want) {
		t.Fatalf("round-tripped skin = %#v, want %#v", host.setSkins, want)
	}
}

func TestPlayerInventoryItemRoundTrip(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	valuesNBT := []byte{10, 0, 0, 8, 5, 0, 'o', 'w', 'n', 'e', 'r', 4, 0, 'r', 'u', 's', 't', 0}
	want := ItemStack{
		Identifier: "minecraft:diamond_sword", Count: 1, Damage: 7,
		Unbreakable: true, AnvilCost: 4,
		CustomName: "Bridge Sword", Lore: []string{"one", "two"}, ValuesNBT: valuesNBT,
		Enchantments: []ItemEnchantment{{ID: 9, Level: 5}, {ID: 17, Level: 3}},
	}
	host := &recordingHost{inventoryItem: want}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	items := commandNamed(t, commands, "items")
	id := PlayerID{Generation: 15}
	input := CommandInput{
		Source: "TestPlayer", SourceKind: CommandSourcePlayer, SourcePlayer: &id,
		OnlinePlayers: []CommandPlayer{{Player: id, Name: "TestPlayer"}}, Arguments: "copy 3 4",
	}
	output, err := runtime.HandleCommand(items.Index, input)
	if err != nil || output.Failed {
		t.Fatalf("output=%+v error=%v", output, err)
	}
	if len(host.inventorySets) != 1 || host.inventorySets[0].Slot != 4 || !reflect.DeepEqual(host.inventorySets[0].Item, want) {
		t.Fatalf("set items=%#v want=%#v", host.inventorySets, want)
	}
	input.Arguments = "give-sword"
	output, err = runtime.HandleCommand(items.Index, input)
	if err != nil || output.Failed {
		t.Fatalf("give output=%+v error=%v", output, err)
	}
	if len(host.inventoryAdds) != 1 {
		t.Fatalf("inventory adds=%#v", host.inventoryAdds)
	}
	added := host.inventoryAdds[0]
	if added.Identifier != "minecraft:diamond_sword" || added.CustomName != "Rust Sword" ||
		len(added.Lore) != 2 || len(added.ValuesNBT) == 0 ||
		!reflect.DeepEqual(added.Enchantments, []ItemEnchantment{{ID: 9, Level: 5}}) {
		t.Fatalf("added item=%#v", added)
	}
}

func TestCommand(t *testing.T) {
	runtime := openTestRuntime(t)
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	if len(commands) != 8 {
		t.Fatalf("commands = %#v, want entity, hello, items, particle, ping, player, sound, and world", commands)
	}
	hello := commandNamed(t, commands, "hello")
	_ = commandNamed(t, commands, "items")
	_ = commandNamed(t, commands, "ping")
	_ = commandNamed(t, commands, "world")
	_ = commandNamed(t, commands, "entity")
	_ = commandNamed(t, commands, "particle")
	_ = commandNamed(t, commands, "player")
	_ = commandNamed(t, commands, "sound")
	optionalFound := false
	for _, overload := range hello.Overloads {
		for _, parameter := range overload.Parameters {
			optionalFound = optionalFound || parameter.Optional
		}
	}
	if !optionalFound {
		t.Fatal("hello command lost its optional argument")
	}
	output, err := runtime.HandleCommand(hello.Index, CommandInput{
		Source:    "Danick",
		Arguments: "say excited dragonfly plugins rock",
	})
	if err != nil {
		t.Fatal(err)
	}
	if output.Failed || output.Message != "HELLO, DANICK! DRAGONFLY PLUGINS ROCK" {
		t.Fatalf("output = %#v", output)
	}
}

func TestWorldCommandHostCalls(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	properties, err := nbt.MarshalEncoding(map[string]any{}, nbt.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	host := &recordingHost{
		worldID: 42, worldLookupOK: true, worldName: "minecraft:overworld",
		worldBlock: WorldBlock{Identifier: "minecraft:gold_block", PropertiesNBT: properties}, worldBlockOK: true,
	}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	world := commandNamed(t, commands, "world")
	player := PlayerID{Generation: 23}
	output, err := runtime.HandleCommand(world.Index, CommandInput{
		Source: "WorldTester", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
		OnlinePlayers: []CommandPlayer{{Player: player, Name: "WorldTester"}},
		Arguments:     "open example:arena",
	})
	if err != nil || output.Failed {
		t.Fatalf("output=%+v error=%v", output, err)
	}
	if host.worldOpened != "example:arena" || host.worldDimension != WorldDimensionOverworld ||
		host.worldTime != 6000 || host.worldSpawn != (BlockPos{X: 0, Y: 64, Z: 0}) {
		t.Fatalf("world calls = name %q dimension %d time %d spawn %+v", host.worldOpened, host.worldDimension, host.worldTime, host.worldSpawn)
	}
	if !slices.Contains(host.texts, "Opened example:arena.") {
		t.Fatalf("texts = %q", host.texts)
	}

	host.texts = nil
	output, err = runtime.HandleCommand(world.Index, CommandInput{
		Source: "WorldTester", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
		OnlinePlayers: []CommandPlayer{{Player: player, Name: "WorldTester"}},
		Arguments:     "inspect 2 3 4",
	})
	if err != nil || output.Failed {
		t.Fatalf("inspect output=%+v error=%v", output, err)
	}
	if host.worldLookup != "minecraft:overworld" || host.worldBlockPos != (BlockPos{X: 2, Y: 3, Z: 4}) ||
		!slices.ContainsFunc(host.texts, func(message string) bool { return strings.Contains(message, "minecraft:gold_block") }) {
		t.Fatalf("inspect calls = lookup %q position %+v texts %q", host.worldLookup, host.worldBlockPos, host.texts)
	}

	host.texts = nil
	output, err = runtime.HandleCommand(world.Index, CommandInput{
		Source: "WorldTester", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
		OnlinePlayers: []CommandPlayer{{Player: player, Name: "WorldTester"}},
		Arguments:     "set-stone -2 70 9",
	})
	if err != nil || output.Failed {
		t.Fatalf("set output=%+v error=%v", output, err)
	}
	if host.worldBlockPos != (BlockPos{X: -2, Y: 70, Z: 9}) || host.worldBlockSet.Identifier != "minecraft:stone" || len(host.worldBlockSet.PropertiesNBT) == 0 {
		t.Fatalf("set call = position %+v block %+v", host.worldBlockPos, host.worldBlockSet)
	}
}

func TestEntityCommandHostCalls(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	host := &recordingHost{
		worldID: 42, worldLookupOK: true,
		entityState:    EntityState{World: 42, Type: "minecraft:player", Position: Vec3{X: 2, Y: 64, Z: 3}, CanTeleport: true},
		spawnedEntity:  EntityID{UUID: [16]byte{7}, Generation: 88},
		worldEntityIDs: []EntityID{{UUID: [16]byte{7}, Generation: 88}},
	}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	command := commandNamed(t, commands, "entity")
	player := PlayerID{UUID: [16]byte{1, 2}, Generation: 17}
	host.worldPlayerIDs = []PlayerID{player}
	for _, arguments := range []string{"text", "sword", "list"} {
		output, err := runtime.HandleCommand(command.Index, CommandInput{
			Source: "Spawner", SourceKind: CommandSourcePlayer, SourcePlayer: &player, Arguments: arguments,
			OnlinePlayers: []CommandPlayer{{Player: player, Name: "Spawner"}},
		})
		if err != nil || output.Failed {
			t.Fatalf("%s: output=%+v error=%v", arguments, output, err)
		}
	}
	if host.entityStateID != (EntityID{UUID: player.UUID, Generation: player.Generation}) {
		t.Fatalf("state entity = %#v", host.entityStateID)
	}
	if len(host.entitySpawns) != 2 || host.entitySpawns[0].Kind != EntityText || host.entitySpawns[0].Text != "Native Rust entity" || host.entitySpawns[0].Position.Y != 65.5 {
		t.Fatalf("text spawn = %#v", host.entitySpawns)
	}
	item := host.entitySpawns[1].Item
	if item == nil || item.Identifier != "minecraft:diamond_sword" || item.Count != 1 {
		t.Fatalf("item spawn = %#v", host.entitySpawns[1])
	}
	if len(host.texts) == 0 || host.texts[len(host.texts)-1] != "1 entities, 1 players." {
		t.Fatalf("list message = %q", host.texts)
	}
}

func TestParticleCommandHostCalls(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	host := &recordingHost{
		worldID: 42, worldLookupOK: true,
		entityState: EntityState{World: 42, Type: "minecraft:player", Position: Vec3{X: 2, Y: 64, Z: 3}, CanTeleport: true},
	}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	command := commandNamed(t, commands, "particle")
	player := PlayerID{UUID: [16]byte{1, 2}, Generation: 17}
	for _, arguments := range []string{"coloured-flame", "block-break", "note"} {
		output, err := runtime.HandleCommand(command.Index, CommandInput{
			Source: "Spawner", SourceKind: CommandSourcePlayer, SourcePlayer: &player, Arguments: arguments,
			OnlinePlayers: []CommandPlayer{{Player: player, Name: "Spawner"}},
		})
		if err != nil || output.Failed {
			t.Fatalf("%s: output=%+v error=%v", arguments, output, err)
		}
	}
	if host.particleWorldID != 42 || len(host.particles) != 3 {
		t.Fatalf("particle calls = world %d, values %#v", host.particleWorldID, host.particles)
	}
	if flame := host.particles[0]; flame.Kind != ParticleFlame || flame.Colour != (RGBA{R: 80, G: 180, B: 255, A: 255}) {
		t.Fatalf("flame = %#v", flame)
	}
	if block := host.particles[1].Block; block == nil || block.Identifier != "minecraft:diamond_block" || len(block.PropertiesNBT) == 0 {
		t.Fatalf("block particle = %#v", host.particles[1])
	}
	if note := host.particles[2]; note.Kind != ParticleNote || note.Data != 6 || note.Pitch != 12 {
		t.Fatalf("note = %#v", note)
	}
	for _, position := range host.particlePositions {
		if position != (Vec3{X: 2, Y: 65.5, Z: 3}) {
			t.Fatalf("particle position = %#v", position)
		}
	}
}

func TestSoundCommandHostCalls(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	host := &recordingHost{
		worldID: 42, worldLookupOK: true,
		entityState: EntityState{World: 42, Type: "minecraft:player", Position: Vec3{X: 2, Y: 64, Z: 3}, CanTeleport: true},
	}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	command := commandNamed(t, commands, "sound")
	player := PlayerID{UUID: [16]byte{1, 2}, Generation: 17}
	arguments := []string{"player", "explosion", "door", "note", "equip", "bucket", "disc", "horn", "attack", "crossbow"}
	for _, argument := range arguments {
		output, err := runtime.HandleCommand(command.Index, CommandInput{
			Source: "Spawner", SourceKind: CommandSourcePlayer, SourcePlayer: &player, Arguments: argument,
			OnlinePlayers: []CommandPlayer{{Player: player, Name: "Spawner"}},
		})
		if err != nil || output.Failed {
			t.Fatalf("%s: output=%+v error=%v", argument, output, err)
		}
	}
	if len(host.playerSounds) != 1 || host.playerSounds[0].Kind != SoundLevelUp {
		t.Fatalf("player sounds = %#v", host.playerSounds)
	}
	wantKinds := []SoundKind{
		SoundExplosion, SoundDoorOpen, SoundNote, SoundEquipItem, SoundBucketEmpty,
		SoundMusicDiscPlay, SoundGoatHorn, SoundAttack, SoundCrossbowLoad,
	}
	if host.worldSoundID != 42 || len(host.worldSounds) != len(wantKinds) {
		t.Fatalf("world sounds = world %d, values %#v", host.worldSoundID, host.worldSounds)
	}
	for index, want := range wantKinds {
		if host.worldSounds[index].Kind != want {
			t.Fatalf("sound %d = %#v, want kind %d", index, host.worldSounds[index], want)
		}
		if host.worldSoundPos[index] != (Vec3{X: 2, Y: 64, Z: 3}) {
			t.Fatalf("sound position %d = %#v", index, host.worldSoundPos[index])
		}
	}
	if value := host.worldSounds[1].Block; value == nil || value.Identifier != "minecraft:wooden_door" || len(value.PropertiesNBT) == 0 {
		t.Fatalf("door sound = %#v", host.worldSounds[1])
	}
	if value := host.worldSounds[2]; value.Data != 6 || value.Integer != 12 {
		t.Fatalf("note sound = %#v", value)
	}
	if value := host.worldSounds[3].Item; value == nil || value.Identifier != "minecraft:diamond_sword" || value.Count != 1 {
		t.Fatalf("equip sound = %#v", host.worldSounds[3])
	}
	if value := host.worldSounds[4]; value.Data != 0 {
		t.Fatalf("bucket sound = %#v", value)
	}
	if value := host.worldSounds[5]; value.Data != 13 {
		t.Fatalf("disc sound = %#v", value)
	}
	if value := host.worldSounds[6]; value.Data != 7 {
		t.Fatalf("horn sound = %#v", value)
	}
	if value := host.worldSounds[7]; value.Flags != 1 {
		t.Fatalf("attack sound = %#v", value)
	}
	if value := host.worldSounds[8]; value.Integer != 2 || value.Flags != 1 {
		t.Fatalf("crossbow sound = %#v", value)
	}
}

func TestWorldCommandRejectsFailedAndMalformedHostReads(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	host := &recordingHost{worldID: 42}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	world := commandNamed(t, commands, "world")
	player := PlayerID{Generation: 24}
	input := CommandInput{
		Source: "WorldTester", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
		OnlinePlayers: []CommandPlayer{{Player: player, Name: "WorldTester"}}, Arguments: "inspect 0 64 0",
	}
	run := func(t *testing.T, want string) {
		t.Helper()
		host.texts = nil
		output, err := runtime.HandleCommand(world.Index, input)
		if err != nil || output.Failed {
			t.Fatalf("output=%+v error=%v", output, err)
		}
		if !slices.Contains(host.texts, want) {
			t.Fatalf("texts = %q, want %q", host.texts, want)
		}
	}

	t.Run("lookup failure", func(t *testing.T) {
		host.worldLookupOK = false
		run(t, "Overworld is unavailable.")
	})
	t.Run("stale handle", func(t *testing.T) {
		host.worldLookupOK, host.worldBlockOK = true, false
		run(t, "Could not read block.")
	})
	t.Run("malformed properties", func(t *testing.T) {
		host.worldBlockOK = true
		host.worldBlock = WorldBlock{Identifier: "minecraft:stone", PropertiesNBT: []byte{0xff}}
		run(t, "Could not read block.")
	})
	t.Run("identifier buffer too small", func(t *testing.T) {
		host.worldBlock = WorldBlock{Identifier: strings.Repeat("x", maxBlockIdentifierBytes+1)}
		run(t, "Could not read block.")
	})
	t.Run("properties buffer too small", func(t *testing.T) {
		host.worldBlock = WorldBlock{Identifier: "minecraft:stone", PropertiesNBT: make([]byte, maxBlockPropertiesBytes+1)}
		run(t, "Could not read block.")
	})
}

func commandNamed(t *testing.T, commands []Command, name string) Command {
	t.Helper()
	for _, command := range commands {
		if command.Name == name {
			return command
		}
	}
	t.Fatalf("command %q not found in %#v", name, commands)
	return Command{}
}

func TestPingCommandUsesPlayerLatency(t *testing.T) {
	runtime := openTestRuntime(t)
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	ping := commandNamed(t, commands, "ping")
	id := PlayerID{Generation: 9}
	id.UUID[0] = 1
	input := CommandInput{
		Source:       "Danick",
		SourcePlayer: &id,
		OnlinePlayers: []CommandPlayer{{
			Player:              id,
			Name:                "Danick",
			LatencyMilliseconds: 37,
		}},
	}
	output, err := runtime.HandleCommand(ping.Index, input)
	if err != nil {
		t.Fatal(err)
	}
	if output.Failed || output.Message != "Danick's ping: 37ms" {
		t.Fatalf("output = %#v", output)
	}
}

func TestDynamicCommandEnum(t *testing.T) {
	runtime := openTestRuntime(t)
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	options, err := runtime.CommandEnumOptions(commandNamed(t, commands, "hello").Index, 6, 1, "Danick", []string{"Danick", "RestartFU"})
	if err != nil {
		t.Fatal(err)
	}
	if len(options) != 3 || options[0] != "Danick" || options[1] != "RestartFU" || options[2] != "everyone" {
		t.Fatalf("options = %#v", options)
	}
}

func TestChatFilter(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerChatSubscription == 0 {
		t.Fatal("chat subscription missing")
	}

	output, err := runtime.HandlePlayerChat(0, PlayerChatInput{Message: "foo fighters"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if output.Cancelled {
		t.Fatal("ordinary chat cancelled")
	}
	if output.Replacement == nil || *output.Replacement != "bar fighters" {
		t.Fatalf("replacement = %v, want bar fighters", output.Replacement)
	}

	output, err = runtime.HandlePlayerChat(0, PlayerChatInput{Message: "blocked"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !output.Cancelled {
		t.Fatal("blocked chat was not cancelled")
	}
}

func TestPlayerJoinAndQuit(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerJoinSubscription == 0 || runtime.Subscriptions()&PlayerQuitSubscription == 0 {
		t.Fatal("join or quit subscription missing")
	}
	id := PlayerID{Generation: 12}
	cancelled, err := runtime.HandlePlayerJoin(0, PlayerJoinInput{Player: id, Name: "Danick"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("lifecycle logger cancelled join")
	}
	if err := runtime.HandlePlayerQuit(0, PlayerQuitInput{Player: id, Name: "Danick"}); err != nil {
		t.Fatal(err)
	}
}

func TestPlayerHurtAndHeal(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerHurtSubscription == 0 || runtime.Subscriptions()&PlayerHealSubscription == 0 {
		t.Fatal("hurt or heal subscription missing")
	}
	hurt, err := runtime.HandlePlayerHurt(0, PlayerHurtInput{
		Damage:         4,
		AttackImmunity: 500 * time.Millisecond,
		Source:         DamageSource{Name: "testDamageSource", ReducedByArmour: true},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if hurt.Cancelled || hurt.Damage != 4 || hurt.AttackImmunity != 500*time.Millisecond {
		t.Fatalf("hurt = %+v", hurt)
	}
	heal, err := runtime.HandlePlayerHeal(0, PlayerHealInput{Health: 2, Source: HealingSource{Name: "testHealingSource"}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if heal.Cancelled || heal.Health != 2 {
		t.Fatalf("heal = %+v", heal)
	}
}

func TestPlayerBlockBreakAndPlace(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerBlockBreakSubscription == 0 || runtime.Subscriptions()&PlayerBlockPlaceSubscription == 0 {
		t.Fatal("block-break or block-place subscription missing")
	}
	broken, err := runtime.HandlePlayerBlockBreak(0, PlayerBlockBreakInput{
		Position:   BlockPos{X: 1, Y: 2, Z: 3},
		Block:      "minecraft:stone",
		Experience: 4,
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if broken.Cancelled || broken.Experience != 4 {
		t.Fatalf("block break = %+v", broken)
	}
	cancelled, err := runtime.HandlePlayerBlockPlace(0, PlayerBlockPlaceInput{
		Position: BlockPos{X: 4, Y: 5, Z: 6},
		Block:    "minecraft:dirt",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("block place cancelled")
	}
}

func TestPlayerFoodLossAndDeath(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerFoodLossSubscription == 0 || runtime.Subscriptions()&PlayerDeathSubscription == 0 {
		t.Fatal("food-loss or death subscription missing")
	}
	food, err := runtime.HandlePlayerFoodLoss(0, PlayerFoodLossInput{From: 10, To: 9}, false)
	if err != nil {
		t.Fatal(err)
	}
	if food.Cancelled || food.To != 9 {
		t.Fatalf("food loss = %+v", food)
	}
	keep, err := runtime.HandlePlayerDeath(0, PlayerDeathInput{Source: DamageSource{Name: "testDamageSource"}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if keep {
		t.Fatal("lifecycle logger changed keep inventory")
	}
}

func TestPlayerToggleSprintAndSneak(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerToggleSprintSubscription == 0 || runtime.Subscriptions()&PlayerToggleSneakSubscription == 0 {
		t.Fatal("sprint-toggle or sneak-toggle subscription missing")
	}
	for name, handle := range map[string]func(InvocationID, PlayerToggleInput, bool) (bool, error){
		"sprint": runtime.HandlePlayerToggleSprint,
		"sneak":  runtime.HandlePlayerToggleSneak,
	} {
		cancelled, err := handle(0, PlayerToggleInput{After: true}, false)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if cancelled {
			t.Fatalf("%s toggle cancelled", name)
		}
	}
}

func TestPlayerJumpAndTeleport(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerJumpSubscription == 0 || runtime.Subscriptions()&PlayerTeleportSubscription == 0 {
		t.Fatal("jump or teleport subscription missing")
	}
	if err := runtime.HandlePlayerJump(0, PlayerID{}); err != nil {
		t.Fatal(err)
	}
	cancelled, err := runtime.HandlePlayerTeleport(0, PlayerTeleportInput{Position: Vec3{X: 1, Y: 64, Z: 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("teleport cancelled")
	}
}

func TestPlayerExperienceGainAndPunchAir(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerExperienceGainSubscription == 0 || runtime.Subscriptions()&PlayerPunchAirSubscription == 0 {
		t.Fatal("experience-gain or punch-air subscription missing")
	}
	output, err := runtime.HandlePlayerExperienceGain(0, PlayerID{}, 5, false)
	if err != nil {
		t.Fatal(err)
	}
	if output.Cancelled || output.Amount != 5 {
		t.Fatalf("experience gain = %+v", output)
	}
	cancelled, err := runtime.HandlePlayerPunchAir(0, PlayerID{}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("punch air cancelled")
	}
}

func TestPlayerHeldSlotChangeAndSleep(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerHeldSlotChangeSubscription == 0 || runtime.Subscriptions()&PlayerSleepSubscription == 0 {
		t.Fatal("held-slot-change or sleep subscription missing")
	}
	cancelled, err := runtime.HandlePlayerHeldSlotChange(0, PlayerHeldSlotChangeInput{From: 1, To: 2}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("held-slot change cancelled")
	}
	output, err := runtime.HandlePlayerSleep(0, PlayerID{}, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if output.Cancelled || !output.SendReminder {
		t.Fatalf("sleep = %+v", output)
	}
}

func TestPlayerBlockPickAndLecternPageTurn(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerBlockPickSubscription == 0 || runtime.Subscriptions()&PlayerLecternPageTurnSubscription == 0 {
		t.Fatal("block-pick or lectern-page subscription missing")
	}
	cancelled, err := runtime.HandlePlayerBlockPick(0, PlayerBlockPickInput{Block: "minecraft:stone"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("block pick cancelled")
	}
	output, err := runtime.HandlePlayerLecternPageTurn(0, PlayerLecternPageTurnInput{OldPage: 1, NewPage: 2}, false)
	if err != nil {
		t.Fatal(err)
	}
	if output.Cancelled || output.NewPage != 2 {
		t.Fatalf("lectern page = %+v", output)
	}
}

func TestPlayerSignEditAndItemUse(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerSignEditSubscription == 0 || runtime.Subscriptions()&PlayerItemUseSubscription == 0 {
		t.Fatal("sign-edit or item-use subscription missing")
	}
	cancelled, err := runtime.HandlePlayerSignEdit(0, PlayerSignEditInput{OldText: "old", NewText: "new"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("sign edit cancelled")
	}
	cancelled, err = runtime.HandlePlayerItemUse(0, PlayerID{}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("item use cancelled")
	}
}

func TestPlayerItemUseOnBlock(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerItemUseOnBlockSubscription == 0 {
		t.Fatal("item-use-on-block subscription missing")
	}
	cancelled, err := runtime.HandlePlayerItemUseOnBlock(0, PlayerItemUseOnBlockInput{Face: 1, ClickPosition: Vec3{X: 0.5, Y: 1, Z: 0.5}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("item use on block cancelled")
	}
}

func TestPlayerItemConsumeAndRelease(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerItemConsumeSubscription == 0 || runtime.Subscriptions()&PlayerItemReleaseSubscription == 0 {
		t.Fatal("item-consume or item-release subscription missing")
	}
	stack := ItemStack{Identifier: "minecraft:apple", Count: 1}
	cancelled, err := runtime.HandlePlayerItemConsume(0, PlayerID{}, stack, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("item consume cancelled")
	}
	cancelled, err = runtime.HandlePlayerItemRelease(0, PlayerID{}, stack, time.Second, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("item release cancelled")
	}
}

func TestPlayerItemEventPreservesFullStack(t *testing.T) {
	runtime := openTestRuntime(t)
	nbtData := []byte{10, 0, 0, 8, 4, 0, 'k', 'i', 'n', 'd', 5, 0, 'e', 'v', 'e', 'n', 't', 0}
	valuesNBT := []byte{10, 0, 0, 8, 5, 0, 'o', 'w', 'n', 'e', 'r', 4, 0, 'r', 'u', 's', 't', 0}
	stack := ItemStack{
		Identifier: "minecraft:diamond_sword", Count: 1, Damage: 7,
		Unbreakable: true, AnvilCost: 4, CustomName: "__snapshot_test__",
		Lore: []string{"one", "two"}, NBT: nbtData, ValuesNBT: valuesNBT,
		Enchantments: []ItemEnchantment{{ID: 9, Level: 5}},
	}
	cancelled, err := runtime.HandlePlayerItemConsume(0, PlayerID{}, stack, false)
	if err != nil {
		t.Fatal(err)
	}
	if !cancelled {
		t.Fatal("full item snapshot was not preserved through event dispatch")
	}
}

func TestPlayerItemDamageAndDrop(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerItemDamageSubscription == 0 || runtime.Subscriptions()&PlayerItemDropSubscription == 0 {
		t.Fatal("item-damage or item-drop subscription missing")
	}
	stack := ItemStack{Identifier: "minecraft:diamond_sword", Count: 1}
	output, err := runtime.HandlePlayerItemDamage(0, PlayerID{}, stack, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	if output.Cancelled || output.Damage != 1 {
		t.Fatalf("item damage = %+v", output)
	}
	cancelled, err := runtime.HandlePlayerItemDrop(0, PlayerID{}, stack, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("item drop cancelled")
	}
}

func TestPlayerAttackEntityRoundTrip(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerAttackEntitySubscription == 0 {
		t.Fatal("attack-entity subscription missing")
	}
	player := PlayerID{UUID: [16]byte{1, 2, 3}, Generation: 41}
	target := EntityID{UUID: [16]byte{9, 8, 7}, Generation: 52}
	output, err := runtime.HandlePlayerAttackEntity(73, PlayerAttackEntityInput{Player: player, Target: target}, 0.45, 0.3608, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if output.Cancelled || output.KnockbackForce != 0.45 || output.KnockbackHeight != 0.3608 || !output.Critical {
		t.Fatalf("attack output = %#v", output)
	}
}

func TestPlayerItemUseOnEntityRoundTrip(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerItemUseOnEntitySubscription == 0 {
		t.Fatal("item-use-on-entity subscription missing")
	}
	player := PlayerID{UUID: [16]byte{1, 2, 3}, Generation: 41}
	target := EntityID{UUID: [16]byte{9, 8, 7}, Generation: 52}
	cancelled, err := runtime.HandlePlayerItemUseOnEntity(73, PlayerItemUseOnEntityInput{Player: player, Target: target}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("item use was unexpectedly cancelled")
	}
	cancelled, err = runtime.HandlePlayerItemUseOnEntity(73, PlayerItemUseOnEntityInput{Player: player, Target: target}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !cancelled {
		t.Fatal("plugin cleared existing cancellation")
	}
}

func TestPlayerChangeWorldRoundTrip(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerChangeWorldSubscription == 0 {
		t.Fatal("change-world subscription missing")
	}
	player := PlayerID{UUID: [16]byte{1, 2, 3}, Generation: 41}
	before, after := WorldID(52), WorldID(53)
	if err := runtime.HandlePlayerChangeWorld(73, PlayerChangeWorldInput{
		Player: player, Before: &before, After: after,
	}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.HandlePlayerChangeWorld(73, PlayerChangeWorldInput{
		Player: player, After: after,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestPlayerRespawnRoundTrip(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	host := &recordingHost{worldID: 71, worldLookupOK: true}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	if runtime.Subscriptions()&PlayerRespawnSubscription == 0 {
		t.Fatal("respawn subscription missing")
	}
	player := PlayerID{UUID: [16]byte{1, 2, 3}, Generation: 41}
	position := Vec3{X: 2, Y: -5, Z: -4}
	output, err := runtime.HandlePlayerRespawn(73, PlayerRespawnInput{Player: player}, position, 53)
	if err != nil {
		t.Fatal(err)
	}
	if output.Position != (Vec3{X: 2, Y: 64, Z: -4}) || output.World != 71 {
		t.Fatalf("respawn output = %#v", output)
	}
	if host.worldLookup != "minecraft:overworld" {
		t.Fatalf("respawn world lookup = %q", host.worldLookup)
	}
}

func TestCancellationIsMonotonic(t *testing.T) {
	runtime := openTestRuntime(t)
	cancelled, err := runtime.HandlePlayerMove(0, PlayerMoveInput{NewPosition: Vec3{Y: 64}}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !cancelled {
		t.Fatal("plugin cleared existing cancellation")
	}
}

func TestLifecycleControlsDispatch(t *testing.T) {
	runtime := openTestRuntime(t)
	runtime.Disable()
	cancelled, err := runtime.HandlePlayerMove(0, PlayerMoveInput{NewPosition: Vec3{Y: -65}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("disabled plugin handled movement")
	}
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	cancelled, err = runtime.HandlePlayerMove(0, PlayerMoveInput{NewPosition: Vec3{Y: -65}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !cancelled {
		t.Fatal("enabled plugin did not handle movement")
	}
}

//go:noinline
func rawGoMovement(input PlayerMoveInput, cancelled *bool) {
	if input.NewPosition.Y < 0 {
		*cancelled = true
	}
}

func BenchmarkRawGoMovement(b *testing.B) {
	input := PlayerMoveInput{NewPosition: Vec3{Y: 64}}
	for b.Loop() {
		cancelled := false
		rawGoMovement(input, &cancelled)
	}
}

func BenchmarkNativeRustMovement(b *testing.B) {
	runtime := openTestRuntime(b)
	input := PlayerMoveInput{NewPosition: Vec3{Y: 64}}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := runtime.HandlePlayerMove(0, input, false); err != nil {
			b.Fatal(err)
		}
	}
}
