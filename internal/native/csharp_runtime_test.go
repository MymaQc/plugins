package native

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

func openCSharpRuntime(t testing.TB) *Runtime {
	return openCSharpRuntimeWithHost(t, nil)
}

func openCSharpRuntimeWithHost(t testing.TB, host Host) *Runtime {
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
		t.Skipf("C# runtime not built: run make build-native (%v)", err)
	}
	pluginRuntime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pluginRuntime.Close)

	if got := pluginRuntime.PluginCount(); got != 5 {
		t.Fatalf("PluginCount() = %d, want 5", got)
	}
	wantSubscriptions := PlayerMoveSubscription | PlayerChatSubscription | PlayerQuitSubscription |
		PlayerFoodLossSubscription | PlayerToggleSprintSubscription | PlayerToggleSneakSubscription |
		PlayerJumpSubscription | PlayerTeleportSubscription | PlayerPunchAirSubscription
	if got := pluginRuntime.Subscriptions(); got != wantSubscriptions {
		t.Fatalf("Subscriptions() = %d, want %d", got, wantSubscriptions)
	}
	if err := pluginRuntime.Enable(); err != nil {
		t.Fatal(err)
	}
	return pluginRuntime
}

func TestCSharpVanillaGameModeCommand(t *testing.T) {
	host := &recordingHost{}
	pluginRuntime := openCSharpRuntimeWithHost(t, host)
	commands, err := pluginRuntime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	command := commandNamed(t, commands, "gamemode")
	if len(command.Overloads) != 1 || len(command.Overloads[0].Parameters) != 2 ||
		!slices.Equal(command.Overloads[0].Parameters[0].Values, []string{"survival", "creative", "adventure", "spectator"}) {
		t.Fatalf("gamemode descriptor = %#v", command)
	}
	player := PlayerID{UUID: [16]byte{3}, Generation: 9}
	output, err := pluginRuntime.HandleCommand(command.Index, CommandInput{
		Invocation: 42, Source: "Danick", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
		Arguments: []string{"creative"}, OnlinePlayers: []CommandPlayer{{Player: player, Name: "Danick"}},
	})
	if err != nil || output.Failed || output.Message != "Set Danick's game mode to creative." {
		t.Fatalf("output=%#v error=%v", output, err)
	}
	if !slices.Equal(host.states, []PlayerStateKind{PlayerStateGameMode}) || len(host.values) != 1 || host.values[0].Integer != 1 {
		t.Fatalf("game mode host calls: states=%v values=%#v", host.states, host.values)
	}
}

type csharpWorldHost struct {
	*recordingHost
	worldRange                 BlockRange
	worldRangeInvocation       InvocationID
	worldRangeWorld            WorldID
	worldRangeCalls            int
	worldBlockLoaded           WorldBlock
	worldBlockLoadedInvocation InvocationID
	worldBlockLoadedWorld      WorldID
	worldBlockLoadedPos        BlockPos
	worldBlockLoadedCalls      int
	worldBlockLoadedOK         bool
	worldBlockCalls            int
}

func (h *csharpWorldHost) WorldRange(invocation InvocationID, world WorldID) (BlockRange, bool) {
	h.worldRangeInvocation, h.worldRangeWorld = invocation, world
	h.worldRangeCalls++
	return h.worldRange, true
}

func (h *csharpWorldHost) WorldBlockLoaded(invocation InvocationID, world WorldID, position BlockPos) (WorldBlock, bool, bool) {
	h.worldBlockLoadedInvocation, h.worldBlockLoadedWorld, h.worldBlockLoadedPos = invocation, world, position
	h.worldBlockLoadedCalls++
	return h.worldBlockLoaded, h.worldBlockLoadedOK, true
}

func (h *csharpWorldHost) WorldBlock(invocation InvocationID, world WorldID, position BlockPos) (WorldBlock, bool) {
	h.worldBlockCalls++
	return h.recordingHost.WorldBlock(invocation, world, position)
}

func TestCSharpReflectedCommands(t *testing.T) {
	host := &csharpWorldHost{
		recordingHost:      &recordingHost{},
		worldRange:         BlockRange{Min: -64, Max: 319},
		worldBlockLoaded:   WorldBlock{Identifier: "minecraft:sand"},
		worldBlockLoadedOK: true,
	}
	pluginRuntime := openCSharpRuntimeWithHost(t, host)
	commands, err := pluginRuntime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	kitchen := commandNamed(t, commands, "kitchen")
	if !slices.Contains(kitchen.Aliases, "ks") || len(kitchen.Overloads) != 8 {
		t.Fatalf("kitchen descriptor = %#v", kitchen)
	}
	if kitchen.Overloads[1].Parameters[0].Name != "echo" ||
		!slices.Equal(kitchen.Overloads[2].Parameters[1].Values, []string{"survival", "creative", "adventure", "spectator"}) {
		t.Fatalf("command enum values are not lowercase: %#v", kitchen.Overloads)
	}
	help := commandNamed(t, commands, "help")
	ping := commandNamed(t, commands, "ping")
	position := commandNamed(t, commands, "position")
	if len(help.Overloads) != 1 || len(help.Overloads[0].Parameters) != 1 || !help.Overloads[0].Parameters[0].Optional {
		t.Fatalf("help descriptor = %#v", help)
	}
	if len(ping.Overloads) != 1 || len(ping.Overloads[0].Parameters) != 1 || ping.Overloads[0].Parameters[0].Kind != CommandParameterPlayer {
		t.Fatalf("ping descriptor = %#v", ping)
	}
	if !slices.Contains(position.Aliases, "pos") || len(position.Overloads) != 1 || len(position.Overloads[0].Parameters) != 0 {
		t.Fatalf("position descriptor = %#v", position)
	}

	player := PlayerID{UUID: [16]byte{1}, Generation: 7}
	base := CommandInput{
		Invocation: 42, Source: "Danick", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
		SourcePosition: Vec3{X: 1, Y: 64, Z: 2},
		OnlinePlayers: []CommandPlayer{{
			Player: player, Name: "Danick", LatencyMilliseconds: 37,
			Position: Vec3{X: 1, Y: 64, Z: 2},
		}},
	}
	tests := []struct {
		overload  uint64
		arguments []string
		want      string
	}{
		{0, nil, "jumps=0, punches=0, sprints=0, sneaks=0, quits=0"},
		{1, []string{"echo", "hello world"}, "hello world"},
		{2, []string{"mode", "Creative"}, "mode=Creative"},
		{3, []string{"ping"}, "Danick's ping: 37ms"},
		{4, []string{"position", "3 70 -4"}, "position=3,70,-4"},
		{5, []string{"destination", "source"}, "destination=source"},
	}
	for _, test := range tests {
		input := base
		input.Overload, input.Arguments = test.overload, test.arguments
		output, err := pluginRuntime.HandleCommand(kitchen.Index, input)
		if err != nil || output.Failed || output.Message != test.want {
			t.Fatalf("overload %d: output=%#v error=%v, want %q", test.overload, output, err, test.want)
		}
	}
	targetArgument := "02000000000000000000000000000000:8:52:4:65:-2:RestartFU"
	targeted := base
	targeted.Overload = 3
	targeted.Arguments = []string{"ping", targetArgument}
	output, err := pluginRuntime.HandleCommand(kitchen.Index, targeted)
	if err != nil || output.Failed || output.Message != "RestartFU's ping: 52ms" {
		t.Fatalf("targeted player: output=%#v error=%v", output, err)
	}
	for _, text := range []struct {
		action string
		kind   PlayerTextKind
		want   string
	}{
		{"message", PlayerTextMessage, "hello true 12 1.5 <nil>"},
		{"popup", PlayerTextPopup, "hello"},
		{"tip", PlayerTextTip, "hello"},
		{"jukebox", PlayerTextJukeboxPopup, "hello"},
		{"nametag", PlayerTextNameTag, "hello"},
		{"disconnect", PlayerTextDisconnect, "hello"},
	} {
		input := base
		input.Overload = 6
		input.Arguments = []string{"text", text.action, "hello"}
		output, err = pluginRuntime.HandleCommand(kitchen.Index, input)
		if err != nil || output.Failed || output.Message != "" {
			t.Fatalf("player text %s: output=%#v error=%v", text.action, output, err)
		}
		if host.kinds[len(host.kinds)-1] != text.kind || host.texts[len(host.texts)-1] != text.want {
			t.Fatalf("player text %s host call: kind=%v text=%q", text.action, host.kinds[len(host.kinds)-1], host.texts[len(host.texts)-1])
		}
	}
	properties, err := nbt.MarshalEncoding(map[string]any{}, nbt.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	host.worldBlockLoaded.PropertiesNBT = properties
	input := base
	input.Overload = 7
	input.Arguments = []string{"block"}
	output, err = pluginRuntime.HandleCommand(kitchen.Index, input)
	if err != nil || output.Failed || output.Message != "block=(1,63,2), range=-64..319, loaded=true, was_sand=true" {
		t.Fatalf("block output=%#v error=%v", output, err)
	}
	if host.worldRangeCalls != 1 || host.worldRangeInvocation != 42 || host.worldRangeWorld != 0 {
		t.Fatalf("range host calls: calls=%d invocation=%d world=%d", host.worldRangeCalls, host.worldRangeInvocation, host.worldRangeWorld)
	}
	if host.worldBlockLoadedCalls != 2 || host.worldBlockLoadedInvocation != 42 || host.worldBlockLoadedWorld != 0 ||
		host.worldBlockLoadedPos != (BlockPos{X: 1, Y: 63, Z: 2}) {
		t.Fatalf("loaded block host calls: calls=%d invocation=%d world=%d position=%+v",
			host.worldBlockLoadedCalls, host.worldBlockLoadedInvocation, host.worldBlockLoadedWorld, host.worldBlockLoadedPos)
	}
	if host.worldBlockPos != (BlockPos{X: 1, Y: 63, Z: 2}) || host.worldBlockSet.Identifier != "minecraft:sand" ||
		!host.worldSetOpts.DisableBlockUpdates || !host.worldSetOpts.DisableLiquidDisplacement || !host.worldSetOpts.DisableRedstoneUpdates {
		t.Fatalf("block host call: position=%+v block=%+v options=%+v", host.worldBlockPos, host.worldBlockSet, host.worldSetOpts)
	}

	host.worldBlockLoadedOK = false
	host.worldBlock, host.worldBlockOK = host.worldBlockLoaded, true
	output, err = pluginRuntime.HandleCommand(kitchen.Index, input)
	if err != nil || output.Failed || output.Message != "block=(1,63,2), range=-64..319, loaded=false, was_sand=true" {
		t.Fatalf("unloaded block output=%#v error=%v", output, err)
	}
	if host.worldRangeCalls != 2 || host.worldBlockLoadedCalls != 3 || host.worldBlockCalls != 2 {
		t.Fatalf("unloaded fallback host calls: range=%d loaded=%d block=%d",
			host.worldRangeCalls, host.worldBlockLoadedCalls, host.worldBlockCalls)
	}

	options, err := pluginRuntime.CommandEnumOptions(kitchen.Index, 5, 1, CommandEnumContext{
		Source: "Danick", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
		SourcePosition: base.SourcePosition, OnlinePlayers: base.OnlinePlayers,
	})
	if err != nil || !slices.Equal(options, []string{"spawn", "source"}) {
		t.Fatalf("dynamic enum options=%q error=%v", options, err)
	}
}

func TestCSharpRuntimeLifecycleAndQuit(t *testing.T) {
	pluginRuntime := openCSharpRuntime(t)
	if err := pluginRuntime.HandlePlayerQuit(1, PlayerQuitInput{Name: "Gopher"}); err != nil {
		t.Fatal(err)
	}
	cancelled, err := pluginRuntime.HandlePlayerMove(2, PlayerMoveInput{NewPosition: Vec3{Y: -65}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !cancelled {
		t.Fatal("movement below the world was not cancelled")
	}
	cancelled, err = pluginRuntime.HandlePlayerMove(3, PlayerMoveInput{NewPosition: Vec3{Y: 64}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("ordinary movement was cancelled")
	}
	chat, err := pluginRuntime.HandlePlayerChat(4, PlayerChatInput{Message: "BADWORD"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if chat.Replacement == nil || *chat.Replacement != "***" {
		t.Fatalf("chat replacement = %v, want ***", chat.Replacement)
	}
	food, err := pluginRuntime.HandlePlayerFoodLoss(5, PlayerFoodLossInput{From: 1, To: -1}, false)
	if err != nil {
		t.Fatal(err)
	}
	if food.To != 0 {
		t.Fatalf("food = %d, want 0", food.To)
	}
	if err := pluginRuntime.HandlePlayerJump(6, PlayerID{}); err != nil {
		t.Fatal(err)
	}
	for name, call := range map[string]func() (bool, error){
		"teleport": func() (bool, error) {
			return pluginRuntime.HandlePlayerTeleport(7, PlayerTeleportInput{Position: Vec3{Y: 64}}, false)
		},
		"toggle sprint": func() (bool, error) {
			return pluginRuntime.HandlePlayerToggleSprint(8, PlayerToggleInput{After: true}, false)
		},
		"toggle sneak": func() (bool, error) {
			return pluginRuntime.HandlePlayerToggleSneak(9, PlayerToggleInput{After: true}, false)
		},
		"punch air": func() (bool, error) {
			return pluginRuntime.HandlePlayerPunchAir(10, PlayerID{}, false)
		},
	} {
		cancelled, err := call()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if cancelled {
			t.Fatalf("%s unexpectedly cancelled", name)
		}
	}
}

func TestCSharpRuntimeHandlesMovementConcurrently(t *testing.T) {
	pluginRuntime := openCSharpRuntime(t)
	var wait sync.WaitGroup
	errors := make(chan error, 8)
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 1_000 {
				if _, err := pluginRuntime.HandlePlayerMove(1, PlayerMoveInput{NewPosition: Vec3{Y: 64}}, false); err != nil {
					errors <- err
					return
				}
			}
		}()
	}
	wait.Wait()
	close(errors)
	for err := range errors {
		t.Fatal(err)
	}
}

func BenchmarkCSharpMovement(b *testing.B) {
	pluginRuntime := openCSharpRuntime(b)
	input := PlayerMoveInput{NewPosition: Vec3{Y: 64}}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := pluginRuntime.HandlePlayerMove(1, input, false); err != nil {
			b.Fatal(err)
		}
	}
}
