package native

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"testing"
	"time"
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
	player        PlayerID
	texts         []string
	kinds         []PlayerTextKind
	title         PlayerTitle
	rotation      Rotation
	transforms    []PlayerTransformKind
	vectors       []Vec3
	yaws          []float64
	pitches       []float64
	states        []PlayerStateKind
	values        []PlayerStateValue
	state         PlayerStateValue
	reads         []PlayerStateKind
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
	inventoryAdds []ItemStack
}

func (h *recordingHost) SendPlayerText(player PlayerID, kind PlayerTextKind, message string) bool {
	h.player = player
	h.kinds = append(h.kinds, kind)
	h.texts = append(h.texts, message)
	return true
}

func (h *recordingHost) SendPlayerTitle(player PlayerID, title PlayerTitle) bool {
	h.player, h.title = player, title
	return true
}

func (h *recordingHost) TransformPlayer(_ PlayerID, kind PlayerTransformKind, vector Vec3, yaw, pitch float64) bool {
	h.transforms = append(h.transforms, kind)
	h.vectors = append(h.vectors, vector)
	h.yaws = append(h.yaws, yaw)
	h.pitches = append(h.pitches, pitch)
	return true
}

func (h *recordingHost) PlayerRotation(PlayerID) (Rotation, bool) {
	return h.rotation, true
}

func (h *recordingHost) SetPlayerState(_ PlayerID, kind PlayerStateKind, value PlayerStateValue) bool {
	h.states = append(h.states, kind)
	h.values = append(h.values, value)
	return true
}

func (h *recordingHost) PlayerState(_ PlayerID, kind PlayerStateKind) (PlayerStateValue, bool) {
	h.reads = append(h.reads, kind)
	return h.state, true
}

func (h *recordingHost) ChangePlayerEffect(_ PlayerID, operation PlayerEffectOperation, effect PlayerEffect) bool {
	h.effectOps = append(h.effectOps, operation)
	h.effects = append(h.effects, effect)
	return true
}

func (h *recordingHost) SetPlayerEntityVisible(_ PlayerID, entity EntityID, visible bool) bool {
	h.entities = append(h.entities, entity)
	h.visible = append(h.visible, visible)
	return true
}

func (h *recordingHost) PlayerSkin(PlayerID) (PlayerSkin, bool) {
	return h.skin, true
}

func (h *recordingHost) SetPlayerSkin(_ PlayerID, skin PlayerSkin) bool {
	h.setSkins = append(h.setSkins, skin)
	return true
}

func (h *recordingHost) InventorySize(InventoryID) (uint32, bool) { return 36, true }
func (h *recordingHost) InventoryItem(InventoryID, uint32) (ItemStack, bool) {
	return h.inventoryItem, true
}
func (h *recordingHost) SetInventoryItem(inventory InventoryID, slot uint32, item ItemStack) bool {
	h.inventorySets = append(h.inventorySets, struct {
		Inventory InventoryID
		Slot      uint32
		Item      ItemStack
	}{inventory, slot, item})
	return true
}
func (h *recordingHost) AddInventoryItem(_ InventoryID, item ItemStack) (uint32, bool) {
	h.inventoryAdds = append(h.inventoryAdds, item)
	return item.Count, true
}
func (h *recordingHost) ClearInventory(InventoryID) bool                  { return true }
func (h *recordingHost) HeldItem(PlayerID, uint32) (ItemStack, bool)      { return h.inventoryItem, true }
func (h *recordingHost) SetHeldItems(PlayerID, ItemStack, ItemStack) bool { return true }
func (h *recordingHost) SetHeldSlot(PlayerID, uint32) bool                { return true }

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
	if _, err := runtime.HandlePlayerJoin(PlayerJoinInput{Player: id, Name: "TestPlayer"}, false); err != nil {
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
}

func TestMovementGuard(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.PluginCount() != 6 {
		t.Fatalf("plugin count = %d, want 6", runtime.PluginCount())
	}
	if runtime.Subscriptions()&PlayerMoveSubscription == 0 {
		t.Fatal("movement subscription missing")
	}

	input := PlayerMoveInput{NewPosition: Vec3{X: 10, Y: 64, Z: 10}}
	cancelled, err := runtime.HandlePlayerMove(input, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("valid movement cancelled")
	}

	input.NewPosition.Y = -65
	cancelled, err = runtime.HandlePlayerMove(input, false)
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
		output, err := runtime.HandleCommand(commands[0].Index, CommandInput{
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
	host := &recordingHost{state: PlayerStateValue{Number: 12, Integer: 15}}
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
		output, err := runtime.HandleCommand(commands[0].Index, CommandInput{
			Source: "TestPlayer", SourceKind: CommandSourcePlayer, SourcePlayer: &id,
			OnlinePlayers: []CommandPlayer{{Player: id, Name: "TestPlayer"}}, Arguments: arguments,
		})
		if err != nil || output.Failed {
			t.Fatalf("%s: output=%+v error=%v", arguments, output, err)
		}
	}
	want := []PlayerStateKind{PlayerStateGameMode, PlayerStateHeal, PlayerStateHurt, PlayerStateFood, PlayerStateMaxHealth, PlayerStateExperienceLevel, PlayerStateExperienceProgress}
	if !slices.Equal(host.states, want) {
		t.Fatalf("states = %v, want %v", host.states, want)
	}
	if host.values[0].Integer != 1 || host.values[1].Number != 4 || host.values[3].Integer != 15 || host.values[4].Number != 40 {
		t.Fatalf("values = %+v", host.values)
	}
	if host.values[5].Integer != 12 || host.values[6].Number != 0.5 {
		t.Fatalf("experience values = %+v", host.values[5:])
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
	for _, arguments := range []string{"speed 2 30", "clear-speed"} {
		output, err := runtime.HandleCommand(commands[0].Index, CommandInput{
			Source: "TestPlayer", SourceKind: CommandSourcePlayer, SourcePlayer: &id,
			OnlinePlayers: []CommandPlayer{{Player: id, Name: "TestPlayer"}}, Arguments: arguments,
		})
		if err != nil || output.Failed {
			t.Fatalf("%s: output=%+v error=%v", arguments, output, err)
		}
	}
	if !slices.Equal(host.effectOps, []PlayerEffectOperation{PlayerEffectAdd, PlayerEffectRemove}) {
		t.Fatalf("operations = %v", host.effectOps)
	}
	if host.effects[0].Type != EffectSpeed || host.effects[0].Level != 2 || host.effects[0].Duration != 30*time.Second {
		t.Fatalf("effect = %+v", host.effects[0])
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
		output, err := runtime.HandleCommand(commands[0].Index, CommandInput{
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
		output, err := runtime.HandleCommand(commands[0].Index, CommandInput{
			Source: "TestPlayer", SourceKind: CommandSourcePlayer, SourcePlayer: &id,
			OnlinePlayers: []CommandPlayer{{Player: id, Name: "TestPlayer"}}, Arguments: arguments,
		})
		if err != nil || output.Failed {
			t.Fatalf("%s: output=%+v error=%v", arguments, output, err)
		}
	}
	if !slices.Equal(host.states, []PlayerStateKind{PlayerStateSound}) || host.values[0].Integer != int64(SoundLevelUp) {
		t.Fatalf("states=%v values=%+v", host.states, host.values)
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
		output, err := runtime.HandleCommand(commands[0].Index, CommandInput{
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
	output, err := runtime.HandleCommand(commands[0].Index, CommandInput{
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
	if len(commands) != 3 {
		t.Fatalf("commands = %#v, want hello, items, and ping", commands)
	}
	hello := commandNamed(t, commands, "hello")
	_ = commandNamed(t, commands, "items")
	_ = commandNamed(t, commands, "ping")
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
	options, err := runtime.CommandEnumOptions(0, 6, 1, "Danick", []string{"Danick", "RestartFU"})
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

	output, err := runtime.HandlePlayerChat(PlayerChatInput{Message: "foo fighters"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if output.Cancelled {
		t.Fatal("ordinary chat cancelled")
	}
	if output.Replacement == nil || *output.Replacement != "bar fighters" {
		t.Fatalf("replacement = %v, want bar fighters", output.Replacement)
	}

	output, err = runtime.HandlePlayerChat(PlayerChatInput{Message: "blocked"}, false)
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
	cancelled, err := runtime.HandlePlayerJoin(PlayerJoinInput{Player: id, Name: "Danick"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("lifecycle logger cancelled join")
	}
	if err := runtime.HandlePlayerQuit(PlayerQuitInput{Player: id, Name: "Danick"}); err != nil {
		t.Fatal(err)
	}
}

func TestPlayerHurtAndHeal(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerHurtSubscription == 0 || runtime.Subscriptions()&PlayerHealSubscription == 0 {
		t.Fatal("hurt or heal subscription missing")
	}
	hurt, err := runtime.HandlePlayerHurt(PlayerHurtInput{
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
	heal, err := runtime.HandlePlayerHeal(PlayerHealInput{Health: 2, Source: HealingSource{Name: "testHealingSource"}}, false)
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
	broken, err := runtime.HandlePlayerBlockBreak(PlayerBlockBreakInput{
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
	cancelled, err := runtime.HandlePlayerBlockPlace(PlayerBlockPlaceInput{
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
	food, err := runtime.HandlePlayerFoodLoss(PlayerFoodLossInput{From: 10, To: 9}, false)
	if err != nil {
		t.Fatal(err)
	}
	if food.Cancelled || food.To != 9 {
		t.Fatalf("food loss = %+v", food)
	}
	keep, err := runtime.HandlePlayerDeath(PlayerDeathInput{Source: DamageSource{Name: "testDamageSource"}}, false)
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
	for name, handle := range map[string]func(PlayerToggleInput, bool) (bool, error){
		"sprint": runtime.HandlePlayerToggleSprint,
		"sneak":  runtime.HandlePlayerToggleSneak,
	} {
		cancelled, err := handle(PlayerToggleInput{After: true}, false)
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
	if err := runtime.HandlePlayerJump(PlayerID{}); err != nil {
		t.Fatal(err)
	}
	cancelled, err := runtime.HandlePlayerTeleport(PlayerTeleportInput{Position: Vec3{X: 1, Y: 64, Z: 2}}, false)
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
	output, err := runtime.HandlePlayerExperienceGain(PlayerID{}, 5, false)
	if err != nil {
		t.Fatal(err)
	}
	if output.Cancelled || output.Amount != 5 {
		t.Fatalf("experience gain = %+v", output)
	}
	cancelled, err := runtime.HandlePlayerPunchAir(PlayerID{}, false)
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
	cancelled, err := runtime.HandlePlayerHeldSlotChange(PlayerHeldSlotChangeInput{From: 1, To: 2}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("held-slot change cancelled")
	}
	output, err := runtime.HandlePlayerSleep(PlayerID{}, true, false)
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
	cancelled, err := runtime.HandlePlayerBlockPick(PlayerBlockPickInput{Block: "minecraft:stone"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("block pick cancelled")
	}
	output, err := runtime.HandlePlayerLecternPageTurn(PlayerLecternPageTurnInput{OldPage: 1, NewPage: 2}, false)
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
	cancelled, err := runtime.HandlePlayerSignEdit(PlayerSignEditInput{OldText: "old", NewText: "new"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("sign edit cancelled")
	}
	cancelled, err = runtime.HandlePlayerItemUse(PlayerID{}, false)
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
	cancelled, err := runtime.HandlePlayerItemUseOnBlock(PlayerItemUseOnBlockInput{Face: 1, ClickPosition: Vec3{X: 0.5, Y: 1, Z: 0.5}}, false)
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
	cancelled, err := runtime.HandlePlayerItemConsume(PlayerID{}, stack, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("item consume cancelled")
	}
	cancelled, err = runtime.HandlePlayerItemRelease(PlayerID{}, stack, time.Second, false)
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
	cancelled, err := runtime.HandlePlayerItemConsume(PlayerID{}, stack, false)
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
	output, err := runtime.HandlePlayerItemDamage(PlayerID{}, stack, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	if output.Cancelled || output.Damage != 1 {
		t.Fatalf("item damage = %+v", output)
	}
	cancelled, err := runtime.HandlePlayerItemDrop(PlayerID{}, stack, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("item drop cancelled")
	}
}

func TestCancellationIsMonotonic(t *testing.T) {
	runtime := openTestRuntime(t)
	cancelled, err := runtime.HandlePlayerMove(PlayerMoveInput{NewPosition: Vec3{Y: 64}}, true)
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
	cancelled, err := runtime.HandlePlayerMove(PlayerMoveInput{NewPosition: Vec3{Y: -65}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("disabled plugin handled movement")
	}
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	cancelled, err = runtime.HandlePlayerMove(PlayerMoveInput{NewPosition: Vec3{Y: -65}}, false)
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
		if _, err := runtime.HandlePlayerMove(input, false); err != nil {
			b.Fatal(err)
		}
	}
}
