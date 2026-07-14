package native

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"sync"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

const csharpBuiltinGameModeDescriptor = int64(-1 << 63)

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
	for id, mode := range []string{"survival", "creative", "adventure", "spectator"} {
		output, err := pluginRuntime.HandleCommand(command.Index, CommandInput{
			Invocation: 42, Source: "Danick", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
			Arguments: []string{mode}, OnlinePlayers: []CommandPlayer{{Player: player, Name: "Danick"}},
		})
		if err != nil || output.Failed || output.Message != "Set Danick's game mode to "+mode+"." {
			t.Fatalf("mode %s: output=%#v error=%v", mode, output, err)
		}
		if got := host.values[len(host.values)-1].Integer; got != csharpBuiltinGameModeDescriptor|int64(id) {
			t.Fatalf("mode %s descriptor=%d, want %d", mode, got, csharpBuiltinGameModeDescriptor|int64(id))
		}
	}
	if !slices.Equal(host.states, []PlayerStateKind{
		PlayerStateGameMode, PlayerStateGameMode, PlayerStateGameMode, PlayerStateGameMode,
	}) || len(host.values) != 4 {
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
	blocksWithinInvocation     InvocationID
	blocksWithinWorld          WorldID
	blocksWithinPosition       BlockPos
	blocksWithinRadius         int32
	blocksWithinBlocks         []WorldBlock
	blockIterator              BlockIteratorID
	blockIteratorPositions     []BlockPos
	blockIteratorIndex         int
	blockIteratorOpenCalls     int
	blockIteratorNextCalls     int
	blockIteratorCloseCalls    int
	blockIteratorInvocation    InvocationID
	blockIteratorClosed        BlockIteratorID
	highestLightBlockerCall    csharpWorldQueryCall
	highestLightBlocker        int32
	highestBlockCall           csharpWorldQueryCall
	highestBlock               int32
	lightCall                  csharpWorldQueryCall
	light                      uint8
	skyLightCall               csharpWorldQueryCall
	skyLight                   uint8
	worldQueryOperations       []string
	worldLiquid                WorldBlock
	worldLiquidOK              bool
	worldLiquidInvocation      InvocationID
	worldLiquidWorld           WorldID
	worldLiquidPosition        BlockPos
	worldLiquidCalls           int
	worldLiquidReadCalls       []csharpWorldQueryCall
	worldLiquidSets            []csharpWorldLiquidSetCall
	worldBlockUpdates          []csharpWorldBlockUpdateCall
	worldBiome                 int32
	worldBiomeCalls            []csharpWorldQueryCall
	worldBiomeSets             []csharpWorldBiomeSetCall
	worldTemperature           float64
	worldTemperatureCall       csharpWorldQueryCall
	worldRainingAt             bool
	worldRainingAtCall         csharpWorldQueryCall
	worldSnowingAt             bool
	worldSnowingAtCall         csharpWorldQueryCall
	worldThunderingAt          bool
	worldThunderingAtCall      csharpWorldQueryCall
	worldRaining               bool
	worldRainingCall           csharpWorldQueryCall
	worldThundering            bool
	worldThunderingCall        csharpWorldQueryCall
	worldCurrentTick           int64
	worldCurrentTickInvocation InvocationID
	worldCurrentTickWorld      WorldID
	worldCurrentTickCalls      int
	worldParticles             []csharpWorldParticleCall
}

type csharpWorldQueryCall struct {
	invocation InvocationID
	world      WorldID
	position   BlockPos
	x          int32
	z          int32
}

type csharpWorldLiquidSetCall struct {
	invocation InvocationID
	world      WorldID
	position   BlockPos
	liquid     *WorldBlock
}

type csharpWorldBlockUpdateCall struct {
	invocation       InvocationID
	world            WorldID
	position         BlockPos
	block            WorldBlock
	delayNanoseconds int64
}

type csharpWorldBiomeSetCall struct {
	csharpWorldQueryCall
	biome int32
}

type csharpWorldParticleCall struct {
	invocation InvocationID
	world      WorldID
	position   Vec3
	particle   WorldParticle
}

type csharpFormHost struct {
	*recordingHost
	formCalls  []csharpFormSendCall
	closeCalls []csharpFormCloseCall
}

type csharpFormSendCall struct {
	invocation InvocationID
	player     PlayerID
	form       PlayerForm
}

type csharpFormCloseCall struct {
	invocation InvocationID
	player     PlayerID
}

func (h *csharpFormHost) SendPlayerForm(invocation InvocationID, player PlayerID, form PlayerForm) bool {
	form.RequestJSON = append([]byte(nil), form.RequestJSON...)
	h.formCalls = append(h.formCalls, csharpFormSendCall{invocation: invocation, player: player, form: form})
	return true
}

func (h *csharpFormHost) ClosePlayerForm(invocation InvocationID, player PlayerID) bool {
	h.closeCalls = append(h.closeCalls, csharpFormCloseCall{invocation: invocation, player: player})
	return true
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

func (h *csharpWorldHost) OpenWorldBlocksWithin(invocation InvocationID, world WorldID, position BlockPos, radius int32, blocks []WorldBlock) (BlockIteratorID, bool) {
	h.blocksWithinInvocation, h.blocksWithinWorld = invocation, world
	h.blocksWithinPosition, h.blocksWithinRadius = position, radius
	h.blocksWithinBlocks = append([]WorldBlock(nil), blocks...)
	h.blockIteratorIndex = 0
	h.blockIteratorOpenCalls++
	h.worldQueryOperations = append(h.worldQueryOperations, "open")
	return h.blockIterator, true
}

func (h *csharpWorldHost) NextWorldBlock(invocation InvocationID, iterator BlockIteratorID) (BlockPos, bool, bool) {
	h.blockIteratorInvocation = invocation
	h.blockIteratorNextCalls++
	h.worldQueryOperations = append(h.worldQueryOperations, "next")
	if iterator != h.blockIterator || h.blockIteratorIndex >= len(h.blockIteratorPositions) {
		return BlockPos{}, false, iterator == h.blockIterator
	}
	position := h.blockIteratorPositions[h.blockIteratorIndex]
	h.blockIteratorIndex++
	return position, true, true
}

func (h *csharpWorldHost) CloseWorldBlocks(invocation InvocationID, iterator BlockIteratorID) {
	h.blockIteratorInvocation, h.blockIteratorClosed = invocation, iterator
	h.blockIteratorCloseCalls++
	h.worldQueryOperations = append(h.worldQueryOperations, "close")
}

func (h *csharpWorldHost) WorldHighestLightBlocker(invocation InvocationID, world WorldID, x, z int32) (int32, bool) {
	h.highestLightBlockerCall = csharpWorldQueryCall{invocation: invocation, world: world, x: x, z: z}
	h.worldQueryOperations = append(h.worldQueryOperations, "highest-light-blocker")
	return h.highestLightBlocker, true
}

func (h *csharpWorldHost) WorldHighestBlock(invocation InvocationID, world WorldID, x, z int32) (int32, bool) {
	h.highestBlockCall = csharpWorldQueryCall{invocation: invocation, world: world, x: x, z: z}
	h.worldQueryOperations = append(h.worldQueryOperations, "highest-block")
	return h.highestBlock, true
}

func (h *csharpWorldHost) WorldLight(invocation InvocationID, world WorldID, position BlockPos) (uint8, bool) {
	h.lightCall = csharpWorldQueryCall{invocation: invocation, world: world, position: position}
	h.worldQueryOperations = append(h.worldQueryOperations, "light")
	return h.light, true
}

func (h *csharpWorldHost) WorldSkyLight(invocation InvocationID, world WorldID, position BlockPos) (uint8, bool) {
	h.skyLightCall = csharpWorldQueryCall{invocation: invocation, world: world, position: position}
	h.worldQueryOperations = append(h.worldQueryOperations, "sky-light")
	return h.skyLight, true
}

func (h *csharpWorldHost) WorldLiquid(invocation InvocationID, world WorldID, position BlockPos) (WorldBlock, bool, bool) {
	h.worldLiquidInvocation, h.worldLiquidWorld, h.worldLiquidPosition = invocation, world, position
	h.worldLiquidCalls++
	h.worldLiquidReadCalls = append(h.worldLiquidReadCalls, csharpWorldQueryCall{
		invocation: invocation,
		world:      world,
		position:   position,
	})
	return h.worldLiquid, h.worldLiquidOK, true
}

func (h *csharpWorldHost) SetWorldLiquid(invocation InvocationID, world WorldID, position BlockPos, liquid *WorldBlock) bool {
	var copied *WorldBlock
	if liquid != nil {
		value := *liquid
		value.PropertiesNBT = append([]byte(nil), liquid.PropertiesNBT...)
		copied = &value
		h.worldLiquid, h.worldLiquidOK = value, true
	} else {
		h.worldLiquid, h.worldLiquidOK = WorldBlock{}, false
	}
	h.worldLiquidSets = append(h.worldLiquidSets, csharpWorldLiquidSetCall{
		invocation: invocation,
		world:      world,
		position:   position,
		liquid:     copied,
	})
	return true
}

func (h *csharpWorldHost) ScheduleWorldBlockUpdate(invocation InvocationID, world WorldID, position BlockPos, block WorldBlock, delayNanoseconds int64) bool {
	block.PropertiesNBT = append([]byte(nil), block.PropertiesNBT...)
	h.worldBlockUpdates = append(h.worldBlockUpdates, csharpWorldBlockUpdateCall{
		invocation:       invocation,
		world:            world,
		position:         position,
		block:            block,
		delayNanoseconds: delayNanoseconds,
	})
	return true
}

func (h *csharpWorldHost) WorldBiome(invocation InvocationID, world WorldID, position BlockPos) (int32, bool) {
	h.worldBiomeCalls = append(h.worldBiomeCalls, csharpWorldQueryCall{invocation: invocation, world: world, position: position})
	return h.worldBiome, true
}

func (h *csharpWorldHost) SetWorldBiome(invocation InvocationID, world WorldID, position BlockPos, biome int32) bool {
	h.worldBiomeSets = append(h.worldBiomeSets, csharpWorldBiomeSetCall{
		csharpWorldQueryCall: csharpWorldQueryCall{invocation: invocation, world: world, position: position},
		biome:                biome,
	})
	h.worldBiome = biome
	return true
}

func (h *csharpWorldHost) WorldTemperature(invocation InvocationID, world WorldID, position BlockPos) (float64, bool) {
	h.worldTemperatureCall = csharpWorldQueryCall{invocation: invocation, world: world, position: position}
	return h.worldTemperature, true
}

func (h *csharpWorldHost) WorldRainingAt(invocation InvocationID, world WorldID, position BlockPos) (bool, bool) {
	h.worldRainingAtCall = csharpWorldQueryCall{invocation: invocation, world: world, position: position}
	return h.worldRainingAt, true
}

func (h *csharpWorldHost) WorldSnowingAt(invocation InvocationID, world WorldID, position BlockPos) (bool, bool) {
	h.worldSnowingAtCall = csharpWorldQueryCall{invocation: invocation, world: world, position: position}
	return h.worldSnowingAt, true
}

func (h *csharpWorldHost) WorldThunderingAt(invocation InvocationID, world WorldID, position BlockPos) (bool, bool) {
	h.worldThunderingAtCall = csharpWorldQueryCall{invocation: invocation, world: world, position: position}
	return h.worldThunderingAt, true
}

func (h *csharpWorldHost) WorldRaining(invocation InvocationID, world WorldID) (bool, bool) {
	h.worldRainingCall = csharpWorldQueryCall{invocation: invocation, world: world}
	return h.worldRaining, true
}

func (h *csharpWorldHost) WorldThundering(invocation InvocationID, world WorldID) (bool, bool) {
	h.worldThunderingCall = csharpWorldQueryCall{invocation: invocation, world: world}
	return h.worldThundering, true
}

func (h *csharpWorldHost) WorldCurrentTick(invocation InvocationID, world WorldID) (int64, bool) {
	h.worldCurrentTickInvocation, h.worldCurrentTickWorld = invocation, world
	h.worldCurrentTickCalls++
	return h.worldCurrentTick, true
}

func (h *csharpWorldHost) AddWorldParticle(invocation InvocationID, world WorldID, position Vec3, particle WorldParticle) bool {
	if particle.Block != nil {
		block := *particle.Block
		block.PropertiesNBT = append([]byte(nil), block.PropertiesNBT...)
		particle.Block = &block
	}
	h.worldParticles = append(h.worldParticles, csharpWorldParticleCall{
		invocation: invocation,
		world:      world,
		position:   position,
		particle:   particle,
	})
	return true
}

func TestCSharpReflectedCommands(t *testing.T) {
	host := &csharpWorldHost{
		recordingHost:      &recordingHost{},
		worldRange:         BlockRange{Min: -64, Max: 319},
		worldBlockLoaded:   WorldBlock{Identifier: "minecraft:sand"},
		worldBlockLoadedOK: true,
		blockIterator:      7,
		blockIteratorPositions: []BlockPos{
			{X: 0, Y: 63, Z: 0},
			{X: 2, Y: 63, Z: 2},
		},
		highestLightBlocker: 70,
		highestBlock:        72,
		light:               9,
		skyLight:            15,
		worldBiome:          1,
		worldTemperature:    0.75,
		worldRainingAt:      true,
		worldSnowingAt:      false,
		worldThunderingAt:   true,
		worldRaining:        true,
		worldThundering:     true,
		worldCurrentTick:    123_456,
	}
	pluginRuntime := openCSharpRuntimeWithHost(t, host)
	commands, err := pluginRuntime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	kitchen := commandNamed(t, commands, "kitchen")
	if !slices.Contains(kitchen.Aliases, "ks") || len(kitchen.Overloads) != 14 {
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
	if err != nil || output.Failed || output.Message != "block=(1,63,2), range=-64..319, loaded=true, was_sand=true, nearby_sand=(0,63,0), highest_light_blocker=70, highest_block=72, light=9, sky_light=15, liquid_before=false, liquid=true:Water(still=true,depth=8,falling=false), scheduled_update=water:250ms" {
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
	queryPosition := BlockPos{X: 1, Y: 63, Z: 2}
	if len(host.worldBlockUpdates) != 1 || host.worldBlockUpdates[0].invocation != 42 ||
		host.worldBlockUpdates[0].world != 0 || host.worldBlockUpdates[0].position != queryPosition ||
		host.worldBlockUpdates[0].block.Identifier != "minecraft:water" || len(host.worldLiquidSets) < 1 ||
		host.worldLiquidSets[0].liquid == nil ||
		!slices.Equal(host.worldBlockUpdates[0].block.PropertiesNBT, host.worldLiquidSets[0].liquid.PropertiesNBT) ||
		host.worldBlockUpdates[0].delayNanoseconds != 250_000_000 {
		t.Fatalf("scheduled block update=%+v", host.worldBlockUpdates)
	}
	columnCall := csharpWorldQueryCall{invocation: 42, x: 1, z: 2}
	positionCall := csharpWorldQueryCall{invocation: 42, position: queryPosition}
	if host.blocksWithinInvocation != 42 || host.blocksWithinWorld != 0 || host.blocksWithinPosition != queryPosition ||
		host.blocksWithinRadius != 2 || len(host.blocksWithinBlocks) != 1 ||
		host.blocksWithinBlocks[0].Identifier != "minecraft:sand" || !slices.Equal(host.blocksWithinBlocks[0].PropertiesNBT, properties) {
		t.Fatalf("blocks within open: invocation=%d world=%d position=%+v radius=%d blocks=%+v",
			host.blocksWithinInvocation, host.blocksWithinWorld, host.blocksWithinPosition, host.blocksWithinRadius, host.blocksWithinBlocks)
	}
	if host.highestLightBlockerCall != columnCall || host.highestBlockCall != columnCall ||
		host.lightCall != positionCall || host.skyLightCall != positionCall {
		t.Fatalf("scalar query calls: highest_light=%+v highest=%+v light=%+v sky_light=%+v",
			host.highestLightBlockerCall, host.highestBlockCall, host.lightCall, host.skyLightCall)
	}
	if host.blockIteratorOpenCalls != 1 || host.blockIteratorNextCalls != 1 || host.blockIteratorCloseCalls != 1 ||
		host.blockIteratorInvocation != 42 || host.blockIteratorClosed != 7 || host.blockIteratorIndex != 1 {
		t.Fatalf("block iterator: open=%d next=%d close=%d invocation=%d closed=%d index=%d",
			host.blockIteratorOpenCalls, host.blockIteratorNextCalls, host.blockIteratorCloseCalls,
			host.blockIteratorInvocation, host.blockIteratorClosed, host.blockIteratorIndex)
	}
	if !slices.Equal(host.worldQueryOperations, []string{
		"highest-light-blocker", "highest-block", "light", "sky-light", "open", "next", "close",
	}) {
		t.Fatalf("world query operations=%v", host.worldQueryOperations)
	}
	if host.worldLiquidCalls != 3 || !slices.Equal(host.worldLiquidReadCalls, []csharpWorldQueryCall{
		positionCall, positionCall, positionCall,
	}) {
		t.Fatalf("liquid read calls: calls=%d invocation=%d world=%d position=%+v",
			host.worldLiquidCalls, host.worldLiquidInvocation, host.worldLiquidWorld, host.worldLiquidPosition)
	}
	if len(host.worldLiquidSets) != 3 || host.worldLiquidSets[0].invocation != 42 ||
		host.worldLiquidSets[0].world != 0 || host.worldLiquidSets[0].position != queryPosition ||
		host.worldLiquidSets[0].liquid == nil || host.worldLiquidSets[0].liquid.Identifier != "minecraft:water" ||
		len(host.worldLiquidSets[0].liquid.PropertiesNBT) == 0 ||
		host.worldLiquidSets[1].invocation != 42 || host.worldLiquidSets[1].world != 0 ||
		host.worldLiquidSets[1].position != queryPosition || host.worldLiquidSets[1].liquid != nil ||
		host.worldLiquidSets[2].invocation != 42 || host.worldLiquidSets[2].world != 0 ||
		host.worldLiquidSets[2].position != queryPosition || host.worldLiquidSets[2].liquid == nil ||
		!slices.Equal(host.worldLiquidSets[2].liquid.PropertiesNBT, host.worldLiquidSets[0].liquid.PropertiesNBT) {
		t.Fatalf("liquid set calls=%+v", host.worldLiquidSets)
	}

	host.worldBlockLoadedOK = false
	host.worldBlock, host.worldBlockOK = host.worldBlockLoaded, true
	output, err = pluginRuntime.HandleCommand(kitchen.Index, input)
	if err != nil || output.Failed || output.Message != "block=(1,63,2), range=-64..319, loaded=false, was_sand=true, nearby_sand=(0,63,0), highest_light_blocker=70, highest_block=72, light=9, sky_light=15, liquid_before=true, liquid=true:Water(still=true,depth=8,falling=false), scheduled_update=water:250ms" {
		t.Fatalf("unloaded block output=%#v error=%v", output, err)
	}
	if host.worldRangeCalls != 2 || host.worldBlockLoadedCalls != 3 || host.worldBlockCalls != 2 {
		t.Fatalf("unloaded fallback host calls: range=%d loaded=%d block=%d",
			host.worldRangeCalls, host.worldBlockLoadedCalls, host.worldBlockCalls)
	}
	if host.blockIteratorOpenCalls != 2 || host.blockIteratorNextCalls != 2 || host.blockIteratorCloseCalls != 2 {
		t.Fatalf("second block iterator: open=%d next=%d close=%d",
			host.blockIteratorOpenCalls, host.blockIteratorNextCalls, host.blockIteratorCloseCalls)
	}
	if host.worldLiquidCalls != 7 || len(host.worldLiquidSets) != 6 || host.worldLiquidSets[3].liquid == nil ||
		host.worldLiquidSets[3].liquid.Identifier != "minecraft:water" ||
		!slices.Equal(host.worldLiquidSets[3].liquid.PropertiesNBT, host.worldLiquidSets[0].liquid.PropertiesNBT) ||
		host.worldLiquidSets[4].liquid != nil || host.worldLiquidSets[5].liquid == nil ||
		!slices.Equal(host.worldLiquidSets[5].liquid.PropertiesNBT, host.worldLiquidSets[0].liquid.PropertiesNBT) {
		t.Fatalf("second liquid pass: reads=%d sets=%+v", host.worldLiquidCalls, host.worldLiquidSets)
	}
	if len(host.worldBlockUpdates) != 2 || host.worldBlockUpdates[1].invocation != 42 ||
		host.worldBlockUpdates[1].world != 0 || host.worldBlockUpdates[1].position != queryPosition ||
		host.worldBlockUpdates[1].block.Identifier != "minecraft:water" ||
		!slices.Equal(host.worldBlockUpdates[1].block.PropertiesNBT, host.worldBlockUpdates[0].block.PropertiesNBT) ||
		host.worldBlockUpdates[1].delayNanoseconds != 250_000_000 {
		t.Fatalf("second scheduled block update=%+v", host.worldBlockUpdates)
	}

	input = base
	input.Overload = 8
	input.Arguments = []string{"biome"}
	output, err = pluginRuntime.HandleCommand(kitchen.Index, input)
	if err != nil || output.Failed || output.Message != "biome=Desert, applied=true, temperature=0.75, raining_at=true, snowing_at=false, thundering_at=true, raining=true, thundering=true, restored=true" {
		t.Fatalf("biome output=%#v error=%v", output, err)
	}
	biomePosition := BlockPos{X: 1, Y: 64, Z: 2}
	biomeCall := csharpWorldQueryCall{invocation: 42, position: biomePosition}
	if !slices.Equal(host.worldBiomeCalls, []csharpWorldQueryCall{biomeCall, biomeCall, biomeCall}) ||
		len(host.worldBiomeSets) != 2 || host.worldBiomeSets[0].csharpWorldQueryCall != biomeCall ||
		host.worldBiomeSets[0].biome != 2 || host.worldBiomeSets[1].csharpWorldQueryCall != biomeCall ||
		host.worldBiomeSets[1].biome != 1 || host.worldBiome != 1 {
		t.Fatalf("biome host calls: reads=%+v sets=%+v current=%d", host.worldBiomeCalls, host.worldBiomeSets, host.worldBiome)
	}
	if host.worldTemperatureCall != biomeCall || host.worldRainingAtCall != biomeCall ||
		host.worldSnowingAtCall != biomeCall || host.worldThunderingAtCall != biomeCall ||
		host.worldRainingCall != (csharpWorldQueryCall{invocation: 42}) ||
		host.worldThunderingCall != (csharpWorldQueryCall{invocation: 42}) {
		t.Fatalf("biome weather calls: temperature=%+v raining_at=%+v snowing_at=%+v thundering_at=%+v raining=%+v thundering=%+v",
			host.worldTemperatureCall, host.worldRainingAtCall, host.worldSnowingAtCall,
			host.worldThunderingAtCall, host.worldRainingCall, host.worldThunderingCall)
	}

	input = base
	input.Overload = 9
	input.Arguments = []string{"tick"}
	output, err = pluginRuntime.HandleCommand(kitchen.Index, input)
	if err != nil || output.Failed || output.Message != "tick=123456" {
		t.Fatalf("tick output=%#v error=%v", output, err)
	}
	if host.worldCurrentTickCalls != 1 || host.worldCurrentTickInvocation != 42 || host.worldCurrentTickWorld != 0 {
		t.Fatalf("current tick host calls: calls=%d invocation=%d world=%d",
			host.worldCurrentTickCalls, host.worldCurrentTickInvocation, host.worldCurrentTickWorld)
	}

	input = base
	input.Overload = 10
	input.Arguments = []string{"particle"}
	output, err = pluginRuntime.HandleCommand(kitchen.Index, input)
	if err != nil || output.Failed || output.Message != "particles=35" {
		t.Fatalf("particle output=%#v error=%v", output, err)
	}
	if len(host.worldParticles) != 35 {
		t.Fatalf("particle host calls=%d, want 35", len(host.worldParticles))
	}
	sand := WorldBlock{Identifier: "minecraft:sand", PropertiesNBT: properties}
	wantParticles := []WorldParticle{
		{Kind: ParticleFlame, Colour: RGBA{R: 1, G: 2, B: 3, A: 4}},
		{Kind: ParticleDust, Colour: RGBA{R: 5, G: 6, B: 7, A: 8}},
		{Kind: ParticleBlockBreak, Block: &sand},
		{Kind: ParticlePunchBlock, Data: 5, Block: &sand},
		{Kind: ParticleBlockForceField},
		{Kind: ParticleBoneMeal, Data: 1},
	}
	for instrument := range 16 {
		wantParticles = append(wantParticles, WorldParticle{Kind: ParticleNote, Data: uint32(instrument), Pitch: 24})
	}
	wantParticles = append(wantParticles,
		WorldParticle{Kind: ParticleDragonEggTeleport, Diff: BlockPos{X: -3, Y: 4, Z: 5}},
		WorldParticle{Kind: ParticleEvaporate},
		WorldParticle{Kind: ParticleWaterDrip},
		WorldParticle{Kind: ParticleLavaDrip},
		WorldParticle{Kind: ParticleLava},
		WorldParticle{Kind: ParticleDustPlume},
		WorldParticle{Kind: ParticleHugeExplosion},
		WorldParticle{Kind: ParticleEndermanTeleport},
		WorldParticle{Kind: ParticleSnowballPoof},
		WorldParticle{Kind: ParticleEggSmash},
		WorldParticle{Kind: ParticleSplash, Colour: RGBA{R: 9, G: 10, B: 11, A: 12}},
		WorldParticle{Kind: ParticleEffect, Colour: RGBA{R: 13, G: 14, B: 15, A: 16}},
		WorldParticle{Kind: ParticleEntityFlame},
	)
	for index, call := range host.worldParticles {
		if call.invocation != 42 || call.world != 0 || call.position != base.SourcePosition ||
			!reflect.DeepEqual(call.particle, wantParticles[index]) {
			t.Fatalf("particle call %d=%+v, want invocation=42 world=0 position=%+v particle=%+v",
				index, call, base.SourcePosition, wantParticles[index])
		}
	}

	host.state = PlayerStateValue{Integer: csharpBuiltinGameModeDescriptor | 2}
	input = base
	input.Overload = 11
	input.Arguments = []string{"game-mode"}
	output, err = pluginRuntime.HandleCommand(kitchen.Index, input)
	if err != nil || output.Failed || output.Message != "game_mode_id=2, registered=true, round_trip=true, custom_registered=false" {
		t.Fatalf("game mode output=%#v error=%v", output, err)
	}
	if !slices.Equal(host.reads, []PlayerStateKind{PlayerStateGameMode}) ||
		host.state.Integer != csharpBuiltinGameModeDescriptor|2 {
		t.Fatalf("game mode getter: reads=%v descriptor=%d", host.reads, host.state.Integer)
	}
	if len(host.states) < 2 || !slices.Equal(host.states[len(host.states)-2:], []PlayerStateKind{PlayerStateGameMode, PlayerStateGameMode}) ||
		len(host.values) < 2 || host.values[len(host.values)-2].Integer != 0x6b ||
		host.values[len(host.values)-1].Integer != csharpBuiltinGameModeDescriptor|2 {
		t.Fatalf("custom game mode setter: states=%v values=%#v", host.states, host.values)
	}

	host.state = PlayerStateValue{Integer: 0xa5}
	output, err = pluginRuntime.HandleCommand(kitchen.Index, input)
	if err != nil || output.Failed || output.Message != "game_mode_id=0, registered=false, round_trip=false, custom_registered=false" {
		t.Fatalf("custom game mode output=%#v error=%v", output, err)
	}
	if len(host.reads) < 2 || !slices.Equal(host.reads[len(host.reads)-2:], []PlayerStateKind{PlayerStateGameMode, PlayerStateGameMode}) {
		t.Fatalf("custom game mode getter reads=%v", host.reads)
	}
	if len(host.values) < 2 || host.values[len(host.values)-2].Integer != 0x6b || host.values[len(host.values)-1].Integer != 0xa5 {
		t.Fatalf("custom game mode round trip values=%#v", host.values)
	}

	options, err := pluginRuntime.CommandEnumOptions(kitchen.Index, 5, 1, CommandEnumContext{
		Source: "Danick", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
		SourcePosition: base.SourcePosition, OnlinePlayers: base.OnlinePlayers,
	})
	if err != nil || !slices.Equal(options, []string{"spawn", "source"}) {
		t.Fatalf("dynamic enum options=%q error=%v", options, err)
	}
}

func TestCSharpTypedFormFlow(t *testing.T) {
	host := &csharpFormHost{recordingHost: &recordingHost{}}
	pluginRuntime := openCSharpRuntimeWithHost(t, host)
	commands, err := pluginRuntime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	kitchen := commandNamed(t, commands, "kitchen")
	formOverload, rawFormOverload := -1, -1
	for index, overload := range kitchen.Overloads {
		if len(overload.Parameters) != 1 || overload.Parameters[0].Kind != CommandParameterSubcommand {
			continue
		}
		switch overload.Parameters[0].Name {
		case "form":
			formOverload = index
		case "raw-form":
			rawFormOverload = index
		}
	}
	if formOverload < 0 || rawFormOverload < 0 {
		t.Fatalf("kitchen form overloads missing: %#v", kitchen.Overloads)
	}

	player := PlayerID{UUID: [16]byte{0x42}, Generation: 17}
	snapshot := PlayerSnapshot{
		Player: player, Name: "FormPlayer", LatencyMilliseconds: 41,
		Position: Vec3{X: 12.5, Y: 64, Z: -3.25},
	}
	invocation := InvocationID(100)
	sendMenu := func() PlayerForm {
		t.Helper()
		before := len(host.formCalls)
		output, err := pluginRuntime.HandleCommand(kitchen.Index, CommandInput{
			Invocation: invocation, Overload: uint64(formOverload),
			Source: "FormPlayer", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
			SourcePosition: snapshot.Position,
			Arguments:      []string{"form"},
			OnlinePlayers: []CommandPlayer{{
				Player: player, Name: snapshot.Name, LatencyMilliseconds: snapshot.LatencyMilliseconds,
				Position: snapshot.Position,
			}},
		})
		if err != nil || output.Failed || output.Message != "" {
			t.Fatalf("form command: output=%#v error=%v", output, err)
		}
		if len(host.formCalls) != before+1 {
			t.Fatalf("form sends = %d, want %d", len(host.formCalls), before+1)
		}
		call := host.formCalls[before]
		if call.invocation != invocation || call.player != player {
			t.Fatalf("menu send = invocation %d player %+v, want %d %+v", call.invocation, call.player, invocation, player)
		}
		invocation++
		return call.form
	}
	complete := func(form PlayerForm, closed bool, response string) bool {
		t.Helper()
		var body []byte
		if !closed {
			body = []byte(response)
		}
		accepted := CompletePlayerForm(form.ID, invocation, snapshot, closed, body)
		invocation++
		return accepted
	}

	menu := sendMenu()
	requireJSONEqual(t, menu.RequestJSON, `{
		"type":"form",
		"title":"Kitchen sink forms",
		"content":"Dragonfly's reflected menu API from C#.",
		"elements":[
			{"type":"button","text":"Open every custom element","image":{"type":"path","data":"textures/ui/icon_recipe_nature"}},
			{"type":"button","text":"Skip to the modal","image":{"type":"url","data":"https://raw.githubusercontent.com/df-mc/dragonfly/master/.github/assets/logo.png"}},
			{"type":"header","text":"Generated from Dragonfly"},
			{"type":"divider","text":""},
			{"type":"label","text":"The first two buttons are reflected fields."},
			{"type":"button","text":"Extra button"},
			{"type":"button","text":"Close"},
			{"type":"label","text":"Menu elements may be appended together."},
			{"type":"divider","text":""},
			{"type":"label","text":"Kitchen sink forms: Dragonfly's reflected menu API from C#. (4 buttons, 9 elements)"}
		]
	}`)
	if !complete(menu, false, "0") {
		t.Fatal("menu response rejected")
	}
	_ = CompletePlayerForm(menu.ID, invocation, snapshot, false, []byte("0"))
	if len(host.formCalls) != 2 {
		t.Fatalf("duplicate response changed form sends: got %d, want 2", len(host.formCalls))
	}
	customCall := host.formCalls[1]
	if customCall.invocation != invocation-1 || customCall.player != player {
		t.Fatalf("custom send = invocation %d player %+v, want %d %+v", customCall.invocation, customCall.player, invocation-1, player)
	}
	requireJSONEqual(t, customCall.form.RequestJSON, `{
		"type":"custom_form",
		"title":"Kitchen custom form",
		"content":[
			{"type":"header","text":"Every custom element"},
			{"type":"divider","text":""},
			{"type":"label","text":"Kitchen custom form contains 8 reflected elements."},
			{"type":"input","text":"Name","default":"Dragonfly","placeholder":"Type a name","tooltip":"A UTF-8 string value."},
			{"type":"toggle","text":"Enabled","default":true,"tooltip":"A boolean value."},
			{"type":"slider","text":"Power","min":0,"max":10,"step":0.5,"default":5,"tooltip":"A bounded numeric value."},
			{"type":"dropdown","text":"Colour","default":1,"options":["Red","Green","Blue"],"tooltip":"An option index."},
			{"type":"step_slider","text":"Size","default":1,"steps":["Small","Medium","Large"],"tooltip":"A stepped option index."}
		]
	}`)
	if !complete(customCall.form, false, `[null,null,null,"Alex",false,7.5,2,0]`) {
		t.Fatal("custom response rejected")
	}
	if len(host.formCalls) != 3 {
		t.Fatalf("form sends after custom = %d, want 3", len(host.formCalls))
	}
	modalCall := host.formCalls[2]
	if modalCall.invocation != invocation-1 || modalCall.player != player {
		t.Fatalf("modal send = invocation %d player %+v, want %d %+v", modalCall.invocation, modalCall.player, invocation-1, player)
	}
	requireJSONEqual(t, modalCall.form.RequestJSON, `{
		"type":"modal",
		"title":"Confirm kitchen values",
		"content":"Confirm kitchen values: name=Alex, enabled=False, power=7.5, colour=2, size=0 (2 choices)",
		"button1":"gui.yes",
		"button2":"gui.no"
	}`)
	if !complete(modalCall.form, false, "true") {
		t.Fatal("modal response rejected")
	}
	if len(host.texts) == 0 || host.texts[len(host.texts)-1] != "Accepted: name=Alex, enabled=False, power=7.5, colour=2, size=0" ||
		host.textPlayers[len(host.textPlayers)-1] != player {
		t.Fatalf("modal response messages: players=%+v texts=%q", host.textPlayers, host.texts)
	}

	t.Run("dismissals and close", func(t *testing.T) {
		menu := sendMenu()
		if !complete(menu, true, "") || host.texts[len(host.texts)-1] != "Kitchen menu dismissed." {
			t.Fatalf("menu dismissal texts=%q", host.texts)
		}

		menu = sendMenu()
		if !complete(menu, false, "0") {
			t.Fatal("menu response rejected")
		}
		custom := host.formCalls[len(host.formCalls)-1].form
		if !complete(custom, true, "") || host.texts[len(host.texts)-1] != "Kitchen custom form dismissed." {
			t.Fatalf("custom dismissal texts=%q", host.texts)
		}

		menu = sendMenu()
		if !complete(menu, false, "1") {
			t.Fatal("direct modal response rejected")
		}
		modal := host.formCalls[len(host.formCalls)-1]
		requireJSONEqual(t, modal.form.RequestJSON, `{
			"type":"modal",
			"title":"Confirm kitchen values",
			"content":"Confirm kitchen values: Opened directly from the menu. (2 choices)",
			"button1":"gui.yes",
			"button2":"gui.no"
		}`)
		if !complete(modal.form, true, "") || host.texts[len(host.texts)-1] != "Kitchen modal dismissed." {
			t.Fatalf("modal dismissal texts=%q", host.texts)
		}

		menu = sendMenu()
		beforeClose := len(host.closeCalls)
		if !complete(menu, false, "3") {
			t.Fatal("close button response rejected")
		}
		if len(host.closeCalls) != beforeClose+1 ||
			host.closeCalls[beforeClose] != (csharpFormCloseCall{invocation: invocation - 1, player: player}) ||
			host.texts[len(host.texts)-1] != "Kitchen form closed." {
			t.Fatalf("close calls=%+v texts=%q", host.closeCalls, host.texts)
		}
	})

	t.Run("invalid terminal responses", func(t *testing.T) {
		menu := sendMenu()
		wrong := snapshot
		wrong.Player.Generation++
		if CompletePlayerForm(menu.ID, invocation, wrong, false, []byte("0")) {
			t.Fatal("wrong-player response accepted")
		}
		invocation++
		beforeForms := len(host.formCalls)
		_ = CompletePlayerForm(menu.ID, invocation, snapshot, false, []byte("0"))
		if len(host.formCalls) != beforeForms {
			t.Fatal("response after wrong-player terminal response had an effect")
		}
		invocation++

		menu = sendMenu()
		beforeForms = len(host.formCalls)
		if complete(menu, false, `[]`) {
			t.Fatal("malformed menu response accepted")
		}
		if len(host.formCalls) != beforeForms {
			t.Fatalf("malformed menu response sent another form: %d -> %d", beforeForms, len(host.formCalls))
		}
		_ = CompletePlayerForm(menu.ID, invocation, snapshot, false, []byte("0"))
		if len(host.formCalls) != beforeForms {
			t.Fatal("second response after malformed JSON had an effect")
		}
		invocation++

		menu = sendMenu()
		if !complete(menu, false, "0") {
			t.Fatal("menu response rejected")
		}
		custom := host.formCalls[len(host.formCalls)-1].form
		beforeForms = len(host.formCalls)
		if complete(custom, false, `[null,null,null,"Alex",true,99,1,1]`) {
			t.Fatal("out-of-range custom response accepted")
		}
		if len(host.formCalls) != beforeForms {
			t.Fatalf("invalid custom response sent modal: %d -> %d", beforeForms, len(host.formCalls))
		}
	})

	t.Run("open custom form value", func(t *testing.T) {
		send := func() PlayerForm {
			t.Helper()
			before := len(host.formCalls)
			output, err := pluginRuntime.HandleCommand(kitchen.Index, CommandInput{
				Invocation: invocation, Overload: uint64(rawFormOverload),
				Source: "FormPlayer", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
				SourcePosition: snapshot.Position, Arguments: []string{"raw-form"},
				OnlinePlayers: []CommandPlayer{{
					Player: player, Name: snapshot.Name, LatencyMilliseconds: snapshot.LatencyMilliseconds,
					Position: snapshot.Position,
				}},
			})
			if err != nil || output.Failed || output.Message != "" || len(host.formCalls) != before+1 {
				t.Fatalf("raw form command: output=%#v calls=%d error=%v", output, len(host.formCalls), err)
			}
			invocation++
			return host.formCalls[before].form
		}

		raw := send()
		requireJSONEqual(t, raw.RequestJSON, `{
			"type":"form",
			"title":"Custom Form.Value",
			"content":"Open form interface",
			"elements":[
				{"type":"header","text":"Custom Form.Value"},
				{"type":"button","text":"Submit"}
			]
		}`)
		if !complete(raw, false, "0") || host.texts[len(host.texts)-1] !=
			"raw=0, player=FormPlayer, latency=41ms, position=12.5,64,-3.25" {
			t.Fatalf("raw form response texts=%q", host.texts)
		}

		raw = send()
		if !complete(raw, true, "") || host.texts[len(host.texts)-1] != "Custom Form.Value dismissed." {
			t.Fatalf("raw form dismissal texts=%q", host.texts)
		}
	})
}

func requireJSONEqual(t *testing.T, got []byte, want string) {
	t.Helper()
	var gotValue, wantValue any
	if err := json.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("invalid form JSON %q: %v", got, err)
	}
	if err := json.Unmarshal([]byte(want), &wantValue); err != nil {
		t.Fatalf("invalid expected JSON %q: %v", want, err)
	}
	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Fatalf("form JSON:\n got: %s\nwant: %s", got, want)
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
