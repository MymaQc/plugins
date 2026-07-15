package native

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
	gtpacket "github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const csharpBuiltinGameModeDescriptor = int64(-1 << 63)

type csharpEntityHost struct {
	*recordingHost
	despawnInvocation       InvocationID
	despawnEntity           EntityID
	entityHandle            EntityHandleID
	entityHandleEntity      EntityID
	entityHandleActive      bool
	entityHandleCalls       []EntityID
	entityHandleEntityCalls int
}

type csharpPacketHost struct {
	*recordingHost
	fields map[uint32]PacketFieldValue
	sets   []PacketFieldValue
}

func (h *csharpPacketHost) PacketField(_ PacketHandle, field uint32) (PacketFieldValue, bool) {
	value, ok := h.fields[field]
	value.Data = append([]byte(nil), value.Data...)
	return value, ok
}

func (h *csharpPacketHost) SetPacketField(_ PacketHandle, field uint32, value PacketFieldValue) bool {
	value.Data = append([]byte(nil), value.Data...)
	h.fields[field] = value
	h.sets = append(h.sets, value)
	return true
}

func (h *csharpEntityHost) DespawnEntity(invocation InvocationID, entity EntityID) bool {
	h.despawnInvocation, h.despawnEntity = invocation, entity
	return true
}

func (h *csharpEntityHost) EntityHandle(_ InvocationID, entity EntityID) (EntityHandleID, bool) {
	h.entityHandleCalls = append(h.entityHandleCalls, entity)
	return h.entityHandle, h.entityHandle.Valid()
}

func (h *csharpEntityHost) EntityHandleEntity(_ InvocationID, handle EntityHandleID) (EntityID, bool, bool) {
	h.entityHandleEntityCalls++
	if handle != h.entityHandle {
		return EntityID{}, false, false
	}
	return h.entityHandleEntity, h.entityHandleActive, true
}

func openCSharpRuntime(t testing.TB) *Runtime {
	return openCSharpRuntimeWithHost(t, nil)
}

func csharpArtifacts(t testing.TB) (string, string) {
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
	return library, plugins
}

func openCSharpRuntimeWithHost(t testing.TB, host Host) *Runtime {
	t.Helper()
	library, plugins := csharpArtifacts(t)
	pluginRuntime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pluginRuntime.Close)

	if got := pluginRuntime.PluginCount(); got != 5 {
		t.Fatalf("PluginCount() = %d, want 5", got)
	}
	wantSubscriptions := PlayerMoveSubscription | PlayerChatSubscription | PlayerJoinSubscription | PlayerQuitSubscription |
		PlayerHurtSubscription | PlayerHealSubscription | PlayerDeathSubscription |
		PlayerBlockBreakSubscription | PlayerBlockPlaceSubscription |
		PlayerFoodLossSubscription | PlayerToggleSprintSubscription | PlayerToggleSneakSubscription |
		PlayerStartBreakSubscription | PlayerFireExtinguishSubscription |
		PlayerJumpSubscription | PlayerTeleportSubscription | PlayerExperienceGainSubscription |
		PlayerPunchAirSubscription | PlayerHeldSlotChangeSubscription | PlayerSleepSubscription |
		PlayerBlockPickSubscription | PlayerLecternPageTurnSubscription | PlayerSignEditSubscription |
		PlayerItemUseSubscription | PlayerItemUseOnBlockSubscription | PlayerItemConsumeSubscription |
		PlayerItemReleaseSubscription | PlayerItemDamageSubscription | PlayerItemPickupSubscription |
		PlayerItemDropSubscription | PlayerAttackEntitySubscription | PlayerItemUseOnEntitySubscription |
		PlayerChangeWorldSubscription | PlayerRespawnSubscription | PlayerSkinChangeSubscription |
		PlayerTransferSubscription | PlayerCommandExecutionSubscription | PlayerDiagnosticsSubscription |
		WorldLiquidFlowSubscription | WorldLiquidDecaySubscription | WorldLiquidHardenSubscription |
		WorldSoundSubscription | WorldFireSpreadSubscription | WorldBlockBurnSubscription |
		WorldCropTrampleSubscription | WorldLeavesDecaySubscription | WorldEntitySpawnSubscription |
		WorldEntityDespawnSubscription | WorldExplosionSubscription | WorldRedstoneUpdateSubscription |
		WorldCloseSubscription | PacketClientSubscription | PacketServerSubscription
	if got := pluginRuntime.Subscriptions(); got != wantSubscriptions {
		t.Fatalf("Subscriptions() = %d, want %d", got, wantSubscriptions)
	}
	if err := pluginRuntime.Enable(); err != nil {
		t.Fatal(err)
	}
	return pluginRuntime
}

func TestCSharpPacketHandlersMutateIncomingAndCancel(t *testing.T) {
	host := &csharpPacketHost{recordingHost: &recordingHost{}, fields: map[uint32]PacketFieldValue{
		3: {Kind: PacketFieldString, Data: []byte("  hello  ")},
	}}
	pluginRuntime := openCSharpRuntimeWithHost(t, host)
	cancelled, err := pluginRuntime.HandlePacket(PacketClientEvent, 1, (&gtpacket.Text{}).ID(), "xuid", false)
	if err != nil || cancelled {
		t.Fatalf("HandlePacket(Text) cancelled=%v error=%v", cancelled, err)
	}
	if len(host.sets) != 1 || string(host.sets[0].Data) != "hello" {
		t.Fatalf("packet sets = %#v", host.sets)
	}

	wantUUID := [16]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	host.fields = map[uint32]PacketFieldValue{0: {Kind: PacketFieldUUID, UUID: wantUUID}}
	host.sets = nil
	cancelled, err = pluginRuntime.HandlePacket(PacketClientEvent, 4, (&gtpacket.PlayerSkin{}).ID(), "xuid", false)
	if err != nil || cancelled || len(host.sets) != 1 || host.sets[0].UUID != wantUUID {
		t.Fatalf("HandlePacket(PlayerSkin) cancelled=%v sets=%#v error=%v", cancelled, host.sets, err)
	}

	host.fields = map[uint32]PacketFieldValue{0: {Kind: PacketFieldString}}
	cancelled, err = pluginRuntime.HandlePacket(PacketClientEvent, 2, (&gtpacket.CommandRequest{}).ID(), "xuid", false)
	if err != nil || !cancelled {
		t.Fatalf("HandlePacket(CommandRequest) cancelled=%v error=%v", cancelled, err)
	}
	if cancelled, err = pluginRuntime.HandlePacket(PacketServerEvent, 3, (&gtpacket.Text{}).ID(), "xuid", false); err != nil || cancelled {
		t.Fatalf("HandlePacket(server Text) cancelled=%v error=%v", cancelled, err)
	}
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

func TestCSharpPlayerStateMethods(t *testing.T) {
	host := &recordingHost{rejectStateWrites: true, stateValues: map[PlayerStateKind]PlayerStateValue{
		PlayerStateFood:                {Integer: 10},
		PlayerStateHealth:              {Number: 16},
		PlayerStateMaxHealth:           {Number: 20},
		PlayerStateExperienceLevel:     {Integer: 3},
		PlayerStateExperienceProgress:  {Number: 0.25},
		PlayerStateScale:               {Number: 1},
		PlayerStateInvisible:           {},
		PlayerStateImmobile:            {},
		PlayerStateSpeed:               {Number: 0.1},
		PlayerStateFlightSpeed:         {Number: 0.05},
		PlayerStateVerticalFlightSpeed: {Number: 1},
	}}
	pluginRuntime := openCSharpRuntimeWithHost(t, host)
	commands, err := pluginRuntime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	kitchen := commandNamed(t, commands, "kitchen")
	var overload uint64
	found := false
	for index, candidate := range kitchen.Overloads {
		if len(candidate.Parameters) == 1 && candidate.Parameters[0].Name == "state" {
			overload, found = uint64(index), true
			break
		}
	}
	if !found {
		t.Fatalf("state overload missing: %#v", kitchen.Overloads)
	}
	player := PlayerID{UUID: [16]byte{5}, Generation: 4}
	output, err := pluginRuntime.HandleCommand(kitchen.Index, CommandInput{
		Invocation: 42, Source: "Danick", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
		Overload: overload, Arguments: []string{"state"},
		OnlinePlayers: []CommandPlayer{{Player: player, Name: "Danick"}},
	})
	if err != nil || output.Failed || output.Message != "food=10, health=16/20, experience=3:0.25, scale=1, invisible=false, immobile=false, speed=0.1/0.05/1" {
		t.Fatalf("state output=%#v error=%v", output, err)
	}
	wantReads := []PlayerStateKind{
		PlayerStateFood,
		PlayerStateHealth,
		PlayerStateMaxHealth,
		PlayerStateExperienceLevel,
		PlayerStateExperienceProgress,
		PlayerStateScale,
		PlayerStateInvisible,
		PlayerStateImmobile,
		PlayerStateSpeed,
		PlayerStateFlightSpeed,
		PlayerStateVerticalFlightSpeed,
	}
	if !slices.Equal(host.reads, wantReads) {
		t.Fatalf("state reads=%v, want %v", host.reads, wantReads)
	}
	wantWrites := []PlayerStateKind{
		PlayerStateFood,
		PlayerStateMaxHealth,
		PlayerStateExperienceLevel,
		PlayerStateExperienceProgress,
		PlayerStateScale,
		PlayerStateInvisible,
		PlayerStateImmobile,
		PlayerStateSpeed,
		PlayerStateFlightSpeed,
		PlayerStateVerticalFlightSpeed,
	}
	if !slices.Equal(host.states, wantWrites) {
		t.Fatalf("state writes=%v, want %v", host.states, wantWrites)
	}
}

func TestCSharpTypedEffects(t *testing.T) {
	host := &recordingHost{}
	pluginRuntime := openCSharpRuntimeWithHost(t, host)
	commands, err := pluginRuntime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	kitchen := commandNamed(t, commands, "kitchen")
	var overload uint64
	found := false
	for index, candidate := range kitchen.Overloads {
		if len(candidate.Parameters) == 1 && candidate.Parameters[0].Name == "effect" {
			overload, found = uint64(index), true
			break
		}
	}
	if !found {
		t.Fatalf("effect overload missing: %#v", kitchen.Overloads)
	}

	player := PlayerID{UUID: [16]byte{4}, Generation: 3}
	output, err := pluginRuntime.HandleCommand(kitchen.Index, CommandInput{
		Invocation: 42, Source: "Danick", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
		Overload: overload, Arguments: []string{"effect"},
		OnlinePlayers: []CommandPlayer{{Player: player, Name: "Danick"}},
	})
	if err != nil || output.Failed || output.Message != "effects=28, potions=43, stews=13, active=true" {
		t.Fatalf("effect output=%#v error=%v", output, err)
	}
	if !slices.Equal(host.effectOps, []PlayerEffectOperation{PlayerEffectAdd, PlayerEffectRemove}) ||
		len(host.effects) != 2 || len(host.activeEffects) != 0 {
		t.Fatalf("effect calls: operations=%v effects=%#v active=%#v", host.effectOps, host.effects, host.activeEffects)
	}
	added := host.effects[0]
	if added.Type != EffectRegeneration || added.Level != 1 || added.Duration != 2*time.Second ||
		added.Potency != 0 || !added.Ambient || !added.ParticlesHidden || added.Infinite || added.Tick != 0 {
		t.Fatalf("added effect=%+v", added)
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
	currentWorldCalls          int
	currentWorldInvocation     InvocationID
	entityIterator             EntityIteratorID
	entityIteratorEntities     []EntityID
	playerIteratorEntities     []EntityID
	entityIteratorValues       []EntityID
	entityIteratorIndex        int
	entityIteratorOpenCalls    int
	entityIteratorNextCalls    int
	entityIteratorCloseCalls   int
	entityIteratorInvocation   InvocationID
	entityIteratorWorld        WorldID
	entityIteratorPlayersOnly  []bool
	entityIteratorBoxes        []BBox
	entityIteratorClosed       EntityIteratorID
	entityHandle               EntityHandleID
	entityHandleEntity         EntityID
	entityHandleUUID           [16]byte
	entityHandleActive         bool
	entityHandleClosed         bool
	entityHandleCalls          []EntityID
	entityHandleEntityCalls    int
	entityHandleRemoved        EntityID
	entityHandleAddedPosition  *Vec3
	entityHandleAdded          EntityID
	scheduledWorlds            []WorldID
	scheduledPlugins           []uint64
	scheduledCallbacks         []uint64
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

func (h *csharpWorldHost) CurrentWorld(invocation InvocationID) (WorldID, bool) {
	h.currentWorldCalls++
	h.currentWorldInvocation = invocation
	return h.worldID, h.worldID != 0
}

func (h *csharpWorldHost) ScheduleWorld(world WorldID, plugin, callback uint64, _ int64) bool {
	h.scheduledWorlds = append(h.scheduledWorlds, world)
	h.scheduledPlugins = append(h.scheduledPlugins, plugin)
	h.scheduledCallbacks = append(h.scheduledCallbacks, callback)
	return world != 0 && plugin != 0 && callback != 0
}

func (*csharpWorldHost) CancelWorldTask(uint64, uint64) (bool, bool) {
	return false, false
}

func (h *csharpWorldHost) OpenWorldEntityIterator(invocation InvocationID, world WorldID, playersOnly bool) (EntityIteratorID, bool) {
	h.entityIteratorInvocation, h.entityIteratorWorld = invocation, world
	h.entityIteratorPlayersOnly = append(h.entityIteratorPlayersOnly, playersOnly)
	h.entityIteratorOpenCalls++
	h.entityIteratorIndex = 0
	if playersOnly {
		h.entityIteratorValues = append(h.entityIteratorValues[:0], h.playerIteratorEntities...)
	} else {
		h.entityIteratorValues = append(h.entityIteratorValues[:0], h.entityIteratorEntities...)
	}
	return h.entityIterator, h.entityIterator != 0 && world == h.worldID
}

func (h *csharpWorldHost) OpenWorldEntitiesWithin(invocation InvocationID, world WorldID, box BBox) (EntityIteratorID, bool) {
	h.entityIteratorInvocation, h.entityIteratorWorld = invocation, world
	h.entityIteratorBoxes = append(h.entityIteratorBoxes, box)
	h.entityIteratorOpenCalls++
	h.entityIteratorIndex = 0
	h.entityIteratorValues = append(h.entityIteratorValues[:0], h.entityIteratorEntities...)
	return h.entityIterator, h.entityIterator != 0 && world == h.worldID
}

func (h *csharpWorldHost) NextWorldEntity(invocation InvocationID, iterator EntityIteratorID) (EntityID, bool, bool) {
	h.entityIteratorInvocation = invocation
	h.entityIteratorNextCalls++
	if iterator != h.entityIterator {
		return EntityID{}, false, false
	}
	if h.entityIteratorIndex >= len(h.entityIteratorValues) {
		return EntityID{}, false, true
	}
	entity := h.entityIteratorValues[h.entityIteratorIndex]
	h.entityIteratorIndex++
	return entity, true, true
}

func (h *csharpWorldHost) CloseWorldEntities(invocation InvocationID, iterator EntityIteratorID) {
	h.entityIteratorInvocation, h.entityIteratorClosed = invocation, iterator
	h.entityIteratorCloseCalls++
}

func (h *csharpWorldHost) EntityHandle(_ InvocationID, entity EntityID) (EntityHandleID, bool) {
	h.entityHandleCalls = append(h.entityHandleCalls, entity)
	return h.entityHandle, h.entityHandle.Valid()
}

func (h *csharpWorldHost) EntityHandleEntity(_ InvocationID, handle EntityHandleID) (EntityID, bool, bool) {
	h.entityHandleEntityCalls++
	if handle != h.entityHandle {
		return EntityID{}, false, false
	}
	return h.entityHandleEntity, h.entityHandleActive, true
}

func (h *csharpWorldHost) EntityHandleUUID(handle EntityHandleID) ([16]byte, bool) {
	return h.entityHandleUUID, handle == h.entityHandle
}

func (h *csharpWorldHost) EntityHandleClosed(handle EntityHandleID) (bool, bool) {
	return h.entityHandleClosed, handle == h.entityHandle
}

func (h *csharpWorldHost) CloseEntityHandle(handle EntityHandleID) bool {
	if handle != h.entityHandle {
		return false
	}
	h.entityHandleClosed = true
	return true
}

func (h *csharpWorldHost) RemoveEntity(_ InvocationID, entity EntityID) (EntityHandleID, bool) {
	h.entityHandleRemoved = entity
	if !h.entityHandleActive || entity != h.entityHandleEntity {
		return EntityHandleID{}, false
	}
	h.entityHandleActive = false
	return h.entityHandle, true
}

func (h *csharpWorldHost) AddEntity(_ InvocationID, handle EntityHandleID, position *Vec3) (EntityID, bool) {
	if handle != h.entityHandle || h.entityHandleActive {
		return EntityID{}, false
	}
	if position != nil {
		value := *position
		h.entityHandleAddedPosition = &value
	}
	h.entityHandleEntity = h.entityHandleAdded
	h.entityHandleActive = true
	return h.entityHandleAdded, h.entityHandleAdded.Generation != 0
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

func (h *csharpWorldHost) PlayWorldSound(invocation InvocationID, world WorldID, position Vec3, sound WorldSound) bool {
	_ = h.recordingHost.PlayWorldSound(invocation, world, position, sound)
	return true
}

func TestCSharpReflectedCommands(t *testing.T) {
	host := &csharpWorldHost{
		recordingHost: &recordingHost{entityState: EntityState{
			Type: "minecraft:player", Position: Vec3{X: 1, Y: 64, Z: 2},
			Velocity: Vec3{X: 0.25, Y: 0.5, Z: -0.25},
			Rotation: Rotation{Yaw: 90, Pitch: -15},
		}, worldID: 91, worldName: "kitchen:arena", worldSpawn: BlockPos{X: 8, Y: 70, Z: -4},
			worldTime:        6000,
			worldTimeCycle:   true,
			worldDefaultMode: csharpBuiltinGameModeDescriptor,
			worldDifficulty:  DifficultyView{ID: 2, Builtin: true, StarvationHealthLimit: 2, FireSpreadIncrease: 14},
			healed:           4,
			hurtResult:       PlayerHurtResult{Damage: 1, Vulnerable: true},
			stateValues: map[PlayerStateKind]PlayerStateValue{
				PlayerStateHealth:    {Number: 16},
				PlayerStateMaxHealth: {Number: 20},
			}},
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
	if !slices.Contains(kitchen.Aliases, "ks") || len(kitchen.Overloads) != 27 {
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
		{0, nil, "jumps=0, punches=0, sprints=0, sneaks=0, quits=0, scheduled=0, packets=0/0"},
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
	kinematics := base
	kinematics.Overload = 18
	kinematics.Arguments = []string{"kinematics"}
	output, err = pluginRuntime.HandleCommand(kitchen.Index, kinematics)
	if err != nil || output.Failed || output.Message !=
		"position=1,64,2, velocity=0.25,0.5,-0.25, rotation=90,-15" {
		t.Fatalf("kinematics output=%#v error=%v", output, err)
	}
	if !slices.Equal(host.transforms, []PlayerTransformKind{
		PlayerTransformTeleport,
		PlayerTransformMove,
		PlayerTransformDisplace,
		PlayerTransformVelocity,
	}) || host.vectors[0] != (Vec3{X: 1, Y: 64, Z: 2}) ||
		host.vectors[3] != (Vec3{X: 0.25, Y: 0.5, Z: -0.25}) {
		t.Fatalf("kinematics transforms=%v vectors=%+v", host.transforms, host.vectors)
	}
	heal := base
	heal.Overload = 19
	heal.Arguments = []string{"heal"}
	output, err = pluginRuntime.HandleCommand(kitchen.Index, heal)
	if err != nil || output.Failed || output.Message != "healed=4, damage=1, vulnerable=True, health=16" || len(host.heals) != 1 ||
		host.heals[0].Health != 20 || host.heals[0].Source.Kind != HealingSourceInstant {
		t.Fatalf("heal output=%#v calls=%+v error=%v", output, host.heals, err)
	}
	if len(host.hurts) != 1 || host.hurts[0].Damage != 1 || host.hurts[0].Source.Kind != DamageSourceFall ||
		!host.hurts[0].Source.FeatherFalling {
		t.Fatalf("hurt calls=%+v", host.hurts)
	}
	sources := base
	sources.Overload = 20
	sources.Arguments = []string{"sources"}
	playerEntityID := EntityID{UUID: player.UUID, Generation: player.Generation}
	host.entityHandle = EntityHandleID{Value: 70, Generation: 4}
	host.entityHandleEntity = playerEntityID
	host.entityHandleActive = true
	output, err = pluginRuntime.HandleCommand(kitchen.Index, sources)
	if err != nil || output.Failed || output.Message != "damage=21, healing=4" {
		t.Fatalf("source output=%#v error=%v", output, err)
	}
	if len(host.hurts) != 22 || host.hurts[1].Source.Kind != DamageSourceAttack ||
		host.hurts[1].Source.Entity != playerEntityID ||
		host.hurts[2].Source.Kind != DamageSourceAttack || host.hurts[2].Source.Entity != playerEntityID ||
		host.hurts[9].Source.Kind != DamageSourceProjectile ||
		host.hurts[9].Source.Entity != playerEntityID || host.hurts[9].Source.SecondaryEntity != playerEntityID ||
		host.hurts[13].Source.Kind != DamageSourcePoison || !host.hurts[13].Source.Data ||
		host.hurts[15].Source.Kind != DamageSourceBlock || host.hurts[15].Source.Block == nil ||
		host.hurts[15].Source.Block.Identifier != "minecraft:sand" ||
		host.hurts[16].Source.Kind != DamageSourceCustom || host.hurts[16].Source.Name != "block.DamageSource" ||
		host.hurts[20].Source.Kind != DamageSourceThorns || host.hurts[20].Source.Entity != playerEntityID {
		t.Fatalf("typed source calls=%+v", host.hurts)
	}
	custom := host.hurts[21].Source
	if custom.Kind != DamageSourceCustom || !strings.Contains(custom.Name, "KitchenDamageSource") ||
		!custom.ReducedByArmour || custom.ReducedByResistance || !custom.Fire || !custom.IgnoresTotem ||
		!custom.FireProtection || !custom.BlastProtection || custom.FeatherFalling || custom.ProjectileProtection {
		t.Fatalf("custom damage source=%+v", custom)
	}
	if len(host.heals) != 5 || host.heals[1].Source.Kind != HealingSourceFood || !host.heals[1].Source.Data ||
		host.heals[2].Source.Kind != HealingSourceInstant || host.heals[3].Source.Kind != HealingSourceRegeneration ||
		host.heals[4].Source.Kind != HealingSourceCustom ||
		!strings.Contains(host.heals[4].Source.Name, "KitchenHealingSource") {
		t.Fatalf("healing source calls=%+v", host.heals)
	}
	configuredWorld := base
	configuredWorld.Overload = 21
	configuredWorld.Arguments = []string{"world"}
	output, err = pluginRuntime.HandleCommand(kitchen.Index, configuredWorld)
	if err != nil || output.Failed || output.Message != "memory=World, persistent=kitchen:arena, spawn=8,70,-4, range=-64..319, highest_light_blocker=70, time=6000, overworld=true, cycle=true, difficulty=true" {
		t.Fatalf("configured world output=%#v error=%v", output, err)
	}
	if len(host.worldConfigs) != 2 || host.worldConfigs[0] != (WorldConfig{
		Dimension: WorldDimensionOverworld,
		Provider:  WorldProviderNop,
	}) || host.worldConfigs[1] != (WorldConfig{
		Dimension:       WorldDimensionOverworld,
		Provider:        WorldProviderMCDB,
		ProviderPath:    "kitchen/arena",
		SaveInterval:    10 * time.Minute,
		RandomTickSpeed: -1,
	}) || !host.worldSaved ||
		host.worldSpawn != (BlockPos{X: 8, Y: 70, Z: -4}) || host.transferInvocation != 42 ||
		host.transferPlayer != player || host.transferWorld != 91 ||
		host.transferPosition != (Vec3{X: 8.5, Y: 70, Z: -3.5}) ||
		host.worldRangeCalls != 1 || host.worldRangeInvocation != 0 || host.worldRangeWorld != 91 ||
		host.highestLightBlockerCall != (csharpWorldQueryCall{world: 91, x: 8, z: -4}) ||
		host.worldTime != 6000 || !host.worldTimeCycle || host.worldSleepDuration != time.Second ||
		host.worldDefaultMode != csharpBuiltinGameModeDescriptor || host.worldTickRange != 4 ||
		host.worldDifficulty != (DifficultyView{ID: 2, Builtin: true, StarvationHealthLimit: 2, FireSpreadIncrease: 14}) ||
		!slices.Equal(host.scheduledWorlds, []WorldID{90}) || len(host.scheduledPlugins) != 1 ||
		len(host.scheduledCallbacks) != 1 {
		t.Fatalf("configured world host state: configs=%+v saved=%v spawn=%+v transfer=%d/%+v/%d/%+v range=%d/%d/%d highest=%+v time=%d scheduled=%v/%v/%v",
			host.worldConfigs, host.worldSaved, host.worldSpawn, host.transferInvocation, host.transferPlayer,
			host.transferWorld, host.transferPosition, host.worldRangeCalls, host.worldRangeInvocation,
			host.worldRangeWorld, host.highestLightBlockerCall, host.worldTime, host.scheduledWorlds,
			host.scheduledPlugins, host.scheduledCallbacks)
	}
	longWorldName := strings.Repeat("arena", 80)
	host.worldName = longWorldName
	output, err = pluginRuntime.HandleCommand(kitchen.Index, configuredWorld)
	if err != nil || output.Failed || output.Message !=
		"memory=World, persistent="+longWorldName+", spawn=8,70,-4, range=-64..319, highest_light_blocker=70, time=6000, overworld=true, cycle=true, difficulty=true" {
		t.Fatalf("long world name output=%#v error=%v", output, err)
	}
	host.worldName = "kitchen:arena"
	playerEntity := EntityID{UUID: player.UUID, Generation: player.Generation}
	nonPlayerEntity := EntityID{UUID: [16]byte{9}, Generation: 4}
	host.entityIterator = 12
	host.entityIteratorEntities = []EntityID{playerEntity, nonPlayerEntity}
	host.playerIteratorEntities = []EntityID{playerEntity}
	host.entityPlayer = PlayerSnapshot{
		Player: player, Name: "Danick", LatencyMilliseconds: 37,
		Position: Vec3{X: 1, Y: 64, Z: 2},
	}
	entities := base
	entities.Overload = 22
	entities.Arguments = []string{"entities"}
	output, err = pluginRuntime.HandleCommand(kitchen.Index, entities)
	if err != nil || output.Failed || output.Message != "world=kitchen:arena, entities=2, nearby=2, players=1" {
		t.Fatalf("entity iteration output=%#v error=%v", output, err)
	}
	if host.currentWorldCalls != 4 || host.currentWorldInvocation != 42 ||
		host.entityIteratorOpenCalls != 3 || host.entityIteratorNextCalls != 8 ||
		host.entityIteratorCloseCalls != 3 || host.entityIteratorInvocation != 42 ||
		host.entityIteratorWorld != 91 || host.entityIteratorClosed != 12 ||
		!slices.Equal(host.entityIteratorPlayersOnly, []bool{false, true}) ||
		!slices.Equal(host.entityIteratorBoxes, []BBox{{
			Min: Vec3{X: -15, Y: 48, Z: -14},
			Max: Vec3{X: 17, Y: 80, Z: 18},
		}}) ||
		host.entityPlayerCalls != 5 || len(host.texts) == 0 ||
		host.texts[len(host.texts)-1] != "Kitchen entity iteration is live." {
		t.Fatalf("entity iterator host state: %+v", host)
	}
	host.entityHandle = EntityHandleID{Value: 71, Generation: 5}
	host.entityHandleEntity = nonPlayerEntity
	host.entityHandleUUID = nonPlayerEntity.UUID
	host.entityHandleActive = true
	host.entityHandleAdded = EntityID{UUID: nonPlayerEntity.UUID, Generation: 6}
	host.entityIteratorEntities = []EntityID{playerEntity, nonPlayerEntity}
	host.entityHandleCalls = nil
	host.entityHandleEntityCalls = 0
	handle := base
	handle.Overload = 24
	handle.Arguments = []string{"handle"}
	output, err = pluginRuntime.HandleCommand(kitchen.Index, handle)
	if err != nil || output.Failed || output.Message !=
		"same=true, uuid=09000000-0000-0000-0000-000000000000, before=true, detached=false, after=true, closed=false" {
		t.Fatalf("entity handle output=%#v error=%v", output, err)
	}
	if host.entityHandleRemoved != nonPlayerEntity || host.entityHandleEntity != host.entityHandleAdded ||
		host.entityHandleAddedPosition == nil || *host.entityHandleAddedPosition != base.SourcePosition ||
		!host.entityHandleActive || host.entityHandleEntityCalls != 3 ||
		!slices.Equal(host.entityHandleCalls, []EntityID{nonPlayerEntity, host.entityHandleAdded}) {
		t.Fatalf("entity handle host state: %+v", host)
	}
	host.reads = nil
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
	host.blockByName = WorldBlock{Identifier: "minecraft:wheat", PropertiesNBT: properties}
	host.blockByNameOK = true
	input := base
	input.Overload = 7
	input.Arguments = []string{"block"}
	output, err = pluginRuntime.HandleCommand(kitchen.Index, input)
	if err != nil || output.Failed || output.Message != "block=(1,63,2), lookup=true, range=-64..319, loaded=true, was_sand=true, nearby_sand=(0,63,0), highest_light_blocker=70, highest_block=72, light=9, sky_light=15, liquid_before=false, liquid=true:Water(still=true,depth=8,falling=false), scheduled_update=water:250ms" {
		t.Fatalf("block output=%#v error=%v", output, err)
	}
	wantUniqueLookups := []struct {
		name       string
		properties map[string]any
	}{
		{"minecraft:wheat", map[string]any{"growth": map[string]any{"kind": int32(2), "value": int32(7)}}},
		{"minecraft:candle", map[string]any{
			"candles": map[string]any{"kind": int32(2), "value": int32(0)},
			"lit":     map[string]any{"kind": int32(0), "value": uint8(0)},
		}},
		{"minecraft:barrel", map[string]any{
			"open_bit":         map[string]any{"kind": int32(1), "value": uint8(0)},
			"facing_direction": map[string]any{"kind": int32(2), "value": int32(2)},
		}},
		{"minecraft:quartz_block", map[string]any{
			"pillar_axis": map[string]any{"kind": int32(3), "value": "y"},
		}},
	}
	wantLookups := make([]struct {
		name       string
		properties map[string]any
	}, 0, len(wantUniqueLookups)*2)
	for _, want := range wantUniqueLookups {
		wantLookups = append(wantLookups, want, want)
	}
	if !slices.Equal(host.blockByNameNames, []string{
		"minecraft:wheat", "minecraft:wheat",
		"minecraft:candle", "minecraft:candle",
		"minecraft:barrel", "minecraft:barrel",
		"minecraft:quartz_block", "minecraft:quartz_block",
	}) || len(host.blockByNameProps) != len(wantLookups) {
		t.Fatalf("BlockByName calls = %v", host.blockByNameNames)
	}
	for index, want := range wantLookups {
		var lookupProperties map[string]any
		if err := nbt.UnmarshalEncoding(host.blockByNameProps[index], &lookupProperties, nbt.LittleEndian); err != nil ||
			!reflect.DeepEqual(lookupProperties, want.properties) {
			t.Fatalf("BlockByName(%q, %#v): %v", want.name, lookupProperties, err)
		}
	}
	if host.worldRangeCalls != 3 || host.worldRangeInvocation != 42 || host.worldRangeWorld != 0 {
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
		"highest-light-blocker", "highest-light-blocker",
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
	if err != nil || output.Failed || output.Message != "block=(1,63,2), lookup=true, range=-64..319, loaded=false, was_sand=true, nearby_sand=(0,63,0), highest_light_blocker=70, highest_block=72, light=9, sky_light=15, liquid_before=true, liquid=true:Water(still=true,depth=8,falling=false), scheduled_update=water:250ms" {
		t.Fatalf("unloaded block output=%#v error=%v", output, err)
	}
	if host.worldRangeCalls != 4 || host.worldBlockLoadedCalls != 3 || host.worldBlockCalls != 2 {
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

	var soundOverload uint64
	soundFound := false
	for index, candidate := range kitchen.Overloads {
		if len(candidate.Parameters) == 1 && candidate.Parameters[0].Name == "sound" {
			soundOverload, soundFound = uint64(index), true
			break
		}
	}
	if !soundFound {
		t.Fatalf("sound overload missing: %#v", kitchen.Overloads)
	}
	input = base
	input.Overload = soundOverload
	input.Arguments = []string{"sound"}
	output, err = pluginRuntime.HandleCommand(kitchen.Index, input)
	if err != nil || output.Failed || output.Message != "world_sounds=11, player_sounds=1" {
		t.Fatalf("sound output=%#v error=%v", output, err)
	}
	if len(host.worldSounds) != 11 || len(host.playerSounds) != 1 ||
		host.worldSoundID != 0 || host.playerSounds[0].Kind != SoundLevelUp || host.player != *base.SourcePlayer {
		t.Fatalf("sound host calls: world=%+v player=%+v target=%+v", host.worldSounds, host.playerSounds, host.player)
	}
	wantSoundKinds := []SoundKind{
		SoundExplosion, SoundAttack, SoundFall, SoundBlockPlace, SoundNote, SoundMusicDiscPlay,
		SoundDecoratedPotInserted, SoundEquipItem, SoundBucketFill, SoundCrossbowLoad, SoundGoatHorn,
	}
	for index, sound := range host.worldSounds {
		if sound.Kind != wantSoundKinds[index] || host.worldSoundPos[index] != base.SourcePosition {
			t.Fatalf("world sound %d=%+v position=%+v", index, sound, host.worldSoundPos[index])
		}
	}
	if host.worldSounds[1].Flags != 1 || host.worldSounds[2].Scalar != 2.5 ||
		host.worldSounds[3].Block == nil || host.worldSounds[3].Block.Identifier != "minecraft:sand" ||
		host.worldSounds[4].Data != 0 || host.worldSounds[4].Integer != 12 ||
		host.worldSounds[5].Data != 0 || host.worldSounds[6].Scalar != 0.5 ||
		host.worldSounds[7].Item == nil || host.worldSounds[7].Item.Identifier != "minecraft:diamond_sword" ||
		host.worldSounds[8].Data != 0 || host.worldSounds[8].Block == nil ||
		host.worldSounds[8].Block.Identifier != "minecraft:water" ||
		host.worldSounds[9].Integer != 1 || host.worldSounds[9].Flags != 1 ||
		host.worldSounds[10].Data != 0 {
		t.Fatalf("sound payloads=%+v", host.worldSounds)
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
	input = base
	input.Overload = 17
	input.Arguments = []string{"crop"}
	output, err = pluginRuntime.HandleCommand(kitchen.Index, input)
	if err != nil || output.Failed || output.Message != "crop=-1, planted=7" {
		t.Fatalf("crop output=%#v error=%v", output, err)
	}
	if host.worldBlockSet.Identifier != "minecraft:wheat" || len(host.worldBlockSet.PropertiesNBT) == 0 {
		t.Fatalf("crop host block=%+v", host.worldBlockSet)
	}
	host.worldBlock = host.worldBlockSet
	host.worldBlockOK = true
	output, err = pluginRuntime.HandleCommand(kitchen.Index, input)
	if err != nil || output.Failed || output.Message != "crop=7, planted=7" {
		t.Fatalf("crop round trip output=%#v error=%v", output, err)
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
	cancelled, err := pluginRuntime.HandlePlayerMove(2, PlayerMoveInput{NewPosition: Vec3{Y: math.NaN()}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !cancelled {
		t.Fatal("non-finite movement was not cancelled")
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
	if err := pluginRuntime.HandlePlayerJump(6, PlayerSnapshot{}); err != nil {
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
			return pluginRuntime.HandlePlayerPunchAir(10, PlayerSnapshot{}, false)
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

func TestCSharpRuntimeReflectedWorldHandlers(t *testing.T) {
	host := &recordingHost{}
	pluginRuntime := openCSharpRuntimeWithHost(t, host)
	properties, err := nbt.MarshalEncoding(map[string]any{}, nbt.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	water := WorldBlock{Identifier: "minecraft:water", PropertiesNBT: properties}
	lava := WorldBlock{Identifier: "minecraft:lava", PropertiesNBT: properties}
	stone := WorldBlock{Identifier: "minecraft:stone", PropertiesNBT: properties}
	from, into := BlockPos{X: 1, Y: 64, Z: 2}, BlockPos{X: 3, Y: 65, Z: 4}
	allowed := func(name string, call func() (bool, error)) {
		t.Helper()
		cancelled, err := call()
		if err != nil || cancelled {
			t.Fatalf("%s cancelled=%v error=%v", name, cancelled, err)
		}
	}
	allowed("liquid flow", func() (bool, error) {
		return pluginRuntime.HandleWorldLiquidFlow(300, WorldLiquidFlowInput{
			From: from, Into: into, Liquid: water, Replaced: stone,
		}, false)
	})
	allowed("liquid decay", func() (bool, error) {
		return pluginRuntime.HandleWorldLiquidDecay(301, WorldLiquidDecayInput{
			Position: from, Before: water, After: &lava,
		}, false)
	})
	allowed("liquid harden", func() (bool, error) {
		return pluginRuntime.HandleWorldLiquidHarden(302, WorldLiquidHardenInput{
			Position: into, LiquidHardened: water, OtherLiquid: lava, NewBlock: stone,
		}, false)
	})

	item := ItemStack{Identifier: "minecraft:diamond_sword", Count: 1}
	for kind := SoundKind(0); kind <= SoundGoatHorn; kind++ {
		sound := WorldSound{Kind: kind}
		switch kind {
		case SoundAttack:
			sound.Flags = 1
		case SoundFall, SoundDecoratedPotInserted:
			sound.Scalar = 0.75
		case SoundBlockPlace, SoundBlockBreaking, SoundDoorOpen, SoundDoorClose,
			SoundTrapdoorOpen, SoundTrapdoorClose, SoundFenceGateOpen, SoundFenceGateClose,
			SoundItemUseOn:
			sound.Block = &stone
		case SoundNote:
			sound.Data, sound.Integer = 3, 7
		case SoundMusicDiscPlay:
			sound.Data = 5
		case SoundEquipItem:
			sound.Item = &item
		case SoundBucketFill, SoundBucketEmpty:
			sound.Data, sound.Block = 0, &water
		case SoundCrossbowLoad:
			sound.Integer, sound.Flags = 1, 1
		case SoundGoatHorn:
			sound.Data = 2
		}
		allowed("sound", func() (bool, error) {
			return pluginRuntime.HandleWorldSound(303, WorldSoundInput{
				Sound: sound, Position: Vec3{X: 1, Y: 65, Z: 2},
			}, false)
		})
	}
	cancelled, err := pluginRuntime.HandleWorldSound(304, WorldSoundInput{
		Sound: WorldSound{Kind: SoundExplosion}, Position: Vec3{X: math.NaN()},
	}, false)
	if err != nil || !cancelled {
		t.Fatalf("non-finite sound cancelled=%v error=%v", cancelled, err)
	}

	allowed("fire spread", func() (bool, error) {
		return pluginRuntime.HandleWorldFireSpread(305, WorldFireSpreadInput{From: from, To: into}, false)
	})
	for name, call := range map[string]func() (bool, error){
		"block burn": func() (bool, error) {
			return pluginRuntime.HandleWorldBlockBurn(306, WorldPositionInput{Position: from}, false)
		},
		"crop trample": func() (bool, error) {
			return pluginRuntime.HandleWorldCropTrample(307, WorldPositionInput{Position: from}, false)
		},
		"leaves decay": func() (bool, error) {
			return pluginRuntime.HandleWorldLeavesDecay(308, WorldPositionInput{Position: from}, false)
		},
	} {
		allowed(name, call)
	}

	entity := EntityID{UUID: [16]byte{0x55}, Generation: 12}
	host.entityState = EntityState{
		Type: "minecraft:item", Position: Vec3{X: 1, Y: 65, Z: 2}, Rotation: Rotation{Yaw: 30, Pitch: 10},
	}
	if err := pluginRuntime.HandleWorldEntitySpawn(309, WorldEntityInput{Entity: entity}); err != nil {
		t.Fatal(err)
	}
	if err := pluginRuntime.HandleWorldEntityDespawn(310, WorldEntityInput{Entity: entity}); err != nil {
		t.Fatal(err)
	}
	explosion, err := pluginRuntime.HandleWorldExplosion(311, WorldExplosionInput{
		Position: Vec3{Y: 65}, Entities: []EntityID{entity}, Blocks: []BlockPos{from, from, into},
	}, -0.25, true, false)
	if err != nil || explosion.Cancelled || explosion.ItemDropChance != 0 || !explosion.SpawnFire ||
		!reflect.DeepEqual(explosion.Entities, []EntityID{entity}) ||
		!reflect.DeepEqual(explosion.Blocks, []BlockPos{from, into}) {
		t.Fatalf("explosion=%+v error=%v", explosion, err)
	}
	redstoneCancelled, err := pluginRuntime.HandleWorldRedstoneUpdate(312, WorldRedstoneUpdateInput{
		Position: from, ChangedNeighbour: into, HasChangedNeighbour: true,
		ChangedRedstoneRelevant: true, Source: from, HasSource: true,
		Before: stone, After: &stone, OldPower: 2, NewPower: 13, CurrentTick: 99,
		Cause: RedstoneUpdateCauseCompilerRebuild,
	}, false)
	if err != nil || redstoneCancelled {
		t.Fatalf("redstone cancelled=%v error=%v", redstoneCancelled, err)
	}
	if err := pluginRuntime.HandleWorldClose(313); err != nil {
		t.Fatal(err)
	}
}

func TestCSharpRuntimeReflectedPlayerHandlers(t *testing.T) {
	host := &csharpEntityHost{recordingHost: &recordingHost{}}
	pluginRuntime := openCSharpRuntimeWithHost(t, host)
	properties, err := nbt.MarshalEncoding(map[string]any{}, nbt.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	player := PlayerSnapshot{
		Player:              PlayerID{UUID: [16]byte{0x31, 0x32}, Generation: 47},
		Name:                "HandlerPlayer",
		LatencyMilliseconds: 83,
		Position:            Vec3{X: 12.25, Y: 70, Z: -8.5},
	}
	position := BlockPos{X: 12, Y: 69, Z: -9}
	sand := WorldBlock{Identifier: "minecraft:sand", PropertiesNBT: properties}
	stack := ItemStack{
		Identifier:  "minecraft:diamond_sword",
		Count:       1,
		Damage:      17,
		Unbreakable: true,
		AnvilCost:   4,
		CustomName:  "  Handler blade  ",
		Lore:        []string{"full item view"},
		Enchantments: []ItemEnchantment{{
			ID: 9, Level: 3,
		}},
	}
	invocation := InvocationID(200)
	next := func() InvocationID {
		current := invocation
		invocation++
		return current
	}
	allowed := func(name string, call func() (bool, error)) {
		t.Helper()
		cancelled, err := call()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if cancelled {
			t.Fatalf("%s cancelled by default", name)
		}
	}
	allowed("join", func() (bool, error) {
		return pluginRuntime.HandlePlayerJoin(next(), PlayerJoinInput{Player: player, Name: player.Name}, false)
	})
	blankName := player
	blankName.Name = ""
	rejected, err := pluginRuntime.HandlePlayerJoin(next(), PlayerJoinInput{Player: blankName}, false)
	if err != nil || !rejected {
		t.Fatalf("blank join rejected=%v error=%v", rejected, err)
	}

	allowed("fire extinguish", func() (bool, error) {
		return pluginRuntime.HandlePlayerFireExtinguish(next(), PlayerPositionInput{Player: player, Position: position}, false)
	})
	allowed("start break", func() (bool, error) {
		return pluginRuntime.HandlePlayerStartBreak(next(), PlayerPositionInput{Player: player, Position: position}, false)
	})

	broken, err := pluginRuntime.HandlePlayerBlockBreak(next(), PlayerBlockBreakInput{
		Player: player, Position: position, Block: sand,
		Drops: []ItemStack{{}, stack}, Experience: -7,
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if broken.Cancelled || broken.Experience != 0 || len(broken.Drops) != 1 {
		t.Fatalf("block break output=%+v", broken)
	}
	assertCSharpHandlerStack(t, broken.Drops[0], stack, stack.CustomName)

	allowed("block place", func() (bool, error) {
		return pluginRuntime.HandlePlayerBlockPlace(next(), PlayerBlockPlaceInput{
			Player: player, Position: position, Block: sand,
		}, false)
	})
	allowed("block pick", func() (bool, error) {
		return pluginRuntime.HandlePlayerBlockPick(next(), PlayerBlockPickInput{
			Player: player, Position: position, Block: sand,
		}, false)
	})

	host.entityState = EntityState{Type: "minecraft:player", Position: player.Position}
	allowed("item use", func() (bool, error) {
		return pluginRuntime.HandlePlayerItemUse(next(), player, false)
	})
	host.entityState.Position.Y = math.NaN()
	cancelled, err := pluginRuntime.HandlePlayerItemUse(next(), player, false)
	if err != nil {
		t.Fatal(err)
	}
	if !cancelled {
		t.Fatal("item use did not reject the non-finite live player position")
	}
	host.entityState.Position = player.Position
	allowed("item use on block", func() (bool, error) {
		return pluginRuntime.HandlePlayerItemUseOnBlock(next(), PlayerItemUseOnBlockInput{
			Player: player, Position: position, Face: 3,
			ClickPosition: Vec3{X: 0.125, Y: 0.5, Z: 0.875},
		}, false)
	})
	allowed("item release", func() (bool, error) {
		return pluginRuntime.HandlePlayerItemRelease(next(), player, stack, -123_456_789*time.Nanosecond, false)
	})
	allowed("item consume", func() (bool, error) {
		return pluginRuntime.HandlePlayerItemConsume(next(), player, stack, false)
	})

	experience, err := pluginRuntime.HandlePlayerExperienceGain(next(), player, -5, false)
	if err != nil || experience.Cancelled || experience.Amount != 0 {
		t.Fatalf("experience output=%+v error=%v", experience, err)
	}
	allowed("sign edit", func() (bool, error) {
		return pluginRuntime.HandlePlayerSignEdit(next(), PlayerSignEditInput{
			Player: player, Position: position, FrontSide: true,
			OldText: "before", NewText: "after",
		}, false)
	})
	sleep, err := pluginRuntime.HandlePlayerSleep(next(), player, true, false)
	if err != nil || sleep.Cancelled || !sleep.SendReminder {
		t.Fatalf("sleep output=%+v error=%v", sleep, err)
	}
	page, err := pluginRuntime.HandlePlayerLecternPageTurn(next(), PlayerLecternPageTurnInput{
		Player: player, Position: position, OldPage: 3, NewPage: -2,
	}, false)
	if err != nil || page.Cancelled || page.NewPage != 0 {
		t.Fatalf("lectern output=%+v error=%v", page, err)
	}
	damage, err := pluginRuntime.HandlePlayerItemDamage(next(), player, stack, -4, false)
	if err != nil || damage.Cancelled || damage.Damage != 0 {
		t.Fatalf("item damage output=%+v error=%v", damage, err)
	}
	pickup, err := pluginRuntime.HandlePlayerItemPickup(next(), PlayerItemPickupInput{Player: player, Item: stack}, false)
	if err != nil || pickup.Cancelled {
		t.Fatalf("item pickup output=%+v error=%v", pickup, err)
	}
	assertCSharpHandlerStack(t, pickup.Item, stack, "Handler blade")
	allowed("held slot change", func() (bool, error) {
		return pluginRuntime.HandlePlayerHeldSlotChange(next(), PlayerHeldSlotChangeInput{Player: player, From: 2, To: 7}, false)
	})
	allowed("item drop", func() (bool, error) {
		return pluginRuntime.HandlePlayerItemDrop(next(), player, stack, false)
	})

	heal, err := pluginRuntime.HandlePlayerHeal(next(), PlayerHealInput{
		Player: player, Health: -2, Source: HealingSource{Name: "custom-heal", Kind: HealingSourceCustom},
	}, false)
	if err != nil || heal.Cancelled || heal.Health != 0 {
		t.Fatalf("heal output=%+v error=%v", heal, err)
	}
	hurt, err := pluginRuntime.HandlePlayerHurt(next(), PlayerHurtInput{
		Player: player, Damage: -3, AttackImmunity: -123_456_789 * time.Nanosecond,
		Source: DamageSource{
			Name: "custom-hurt", Kind: DamageSourceCustom,
			ReducedByArmour: true, Fire: true, IgnoresTotem: true,
			FireProtection: true, BlastProtection: true,
		},
	}, false)
	if err != nil || hurt.Cancelled || hurt.Damage != 0 || hurt.AttackImmunity != -123_456_789*time.Nanosecond {
		t.Fatalf("hurt output=%+v error=%v", hurt, err)
	}
	commands, err := pluginRuntime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	kitchen := commandNamed(t, commands, "kitchen")
	var sourceOverload uint64
	sourceFound := false
	for index, overload := range kitchen.Overloads {
		if len(overload.Parameters) == 1 && overload.Parameters[0].Name == "sources" {
			sourceOverload = uint64(index)
			sourceFound = true
			break
		}
	}
	if !sourceFound {
		t.Fatalf("sources overload missing: %#v", kitchen.Overloads)
	}
	host.entityHandle = EntityHandleID{Value: 72, Generation: 5}
	host.entityHandleEntity = EntityID{UUID: player.Player.UUID, Generation: player.Player.Generation}
	host.entityHandleActive = true
	sourceOutput, err := pluginRuntime.HandleCommand(kitchen.Index, CommandInput{
		Invocation: next(), Source: player.Name, SourceKind: CommandSourcePlayer, SourcePlayer: &player.Player,
		Overload: sourceOverload, Arguments: []string{"sources"},
		OnlinePlayers: []CommandPlayer{{Player: player.Player, Name: player.Name, Position: player.Position}},
	})
	if err != nil || sourceOutput.Failed || sourceOutput.Message != "damage=22, healing=5" {
		t.Fatalf("opaque source output=%+v error=%v", sourceOutput, err)
	}
	roundTripDamage := host.hurts[len(host.hurts)-1].Source
	if roundTripDamage.Name != "custom-hurt" || roundTripDamage.Kind != DamageSourceCustom ||
		!roundTripDamage.ReducedByArmour || roundTripDamage.ReducedByResistance || !roundTripDamage.Fire ||
		!roundTripDamage.IgnoresTotem || !roundTripDamage.FireProtection || !roundTripDamage.BlastProtection ||
		roundTripDamage.FeatherFalling || roundTripDamage.ProjectileProtection {
		t.Fatalf("opaque damage round trip=%+v", roundTripDamage)
	}
	roundTripHealing := host.heals[len(host.heals)-1].Source
	if roundTripHealing.Name != "custom-heal" || roundTripHealing.Kind != HealingSourceCustom {
		t.Fatalf("opaque healing round trip=%+v", roundTripHealing)
	}
	keepInventory, err := pluginRuntime.HandlePlayerDeath(next(), PlayerDeathInput{
		Player: player, Source: DamageSource{Kind: DamageSourceVoid},
	}, true)
	if err != nil || !keepInventory {
		t.Fatalf("death keep-inventory=%v error=%v", keepInventory, err)
	}

	before, after := WorldID(17), WorldID(18)
	if err := pluginRuntime.HandlePlayerChangeWorld(next(), PlayerChangeWorldInput{
		Player: player, Before: &before, After: after,
	}); err != nil {
		t.Fatal(err)
	}
	respawn, err := pluginRuntime.HandlePlayerRespawn(
		next(), PlayerRespawnInput{Player: player}, Vec3{X: 1, Y: 64, Z: 2}, after)
	if err != nil || respawn.Position != (Vec3{X: 1, Y: 64, Z: 2}) || respawn.World != after {
		t.Fatalf("respawn output=%+v error=%v", respawn, err)
	}
	skin := PlayerSkin{Width: 2, Height: 2, Pixels: make([]byte, 16), FullID: "handler-skin"}
	skinChange, err := pluginRuntime.HandlePlayerSkinChange(
		next(), PlayerSkinChangeInput{Player: player}, skin, false)
	if err != nil || skinChange.Cancelled || !reflect.DeepEqual(skinChange.Skin, skin) {
		t.Fatalf("skin-change output=%+v error=%v", skinChange, err)
	}

	target := EntityID{UUID: [16]byte{0x44}, Generation: 9}
	host.entityPlayer = PlayerSnapshot{
		Player: PlayerID{UUID: target.UUID, Generation: target.Generation},
		Name:   "TargetPlayer", LatencyMilliseconds: 23,
		Position: Vec3{X: 4, Y: 65, Z: 8},
	}
	host.entityState = EntityState{
		Type: "minecraft:player", Position: Vec3{X: 4, Y: 65, Z: 8},
		Rotation: Rotation{Yaw: 90, Pitch: -15},
	}
	allowed("item use on entity", func() (bool, error) {
		return pluginRuntime.HandlePlayerItemUseOnEntity(
			next(), PlayerItemUseOnEntityInput{Player: player, Target: target}, false)
	})
	attack, err := pluginRuntime.HandlePlayerAttackEntity(
		next(), PlayerAttackEntityInput{Player: player, Target: target}, -1, -2, true, false)
	if err != nil || attack.Cancelled || attack.KnockbackForce != 0 ||
		attack.KnockbackHeight != 0 || !attack.Critical {
		t.Fatalf("attack output=%+v error=%v", attack, err)
	}
	if host.entityPlayerID != target || host.entityPlayerCalls != 2 {
		t.Fatalf("event player resolution used %#v %d times, want %#v twice",
			host.entityPlayerID, host.entityPlayerCalls, target)
	}
	host.entityState.Position.X = math.NaN()
	closeInvocation := next()
	cancelled, err = pluginRuntime.HandlePlayerItemUseOnEntity(
		closeInvocation, PlayerItemUseOnEntityInput{Player: player, Target: target}, false)
	if err != nil || !cancelled || host.despawnInvocation != closeInvocation || host.despawnEntity != target {
		t.Fatalf("invalid entity close: cancelled=%v invocation=%d entity=%#v error=%v",
			cancelled, host.despawnInvocation, host.despawnEntity, err)
	}
	host.entityState.Position.X = 4

	transfer, err := pluginRuntime.HandlePlayerTransfer(next(), PlayerTransferInput{
		Player:  player,
		Address: UDPAddress{IP: []byte{127, 0, 0, 1}, Port: 70_000, Zone: "example"},
	}, false)
	if err != nil || transfer.Cancelled || transfer.Address.Port != 65_535 ||
		!slices.Equal(transfer.Address.IP, []byte{127, 0, 0, 1}) || transfer.Address.Zone != "example" {
		t.Fatalf("transfer output=%+v error=%v", transfer, err)
	}
	commandExecution, err := pluginRuntime.HandlePlayerCommandExecution(next(), PlayerCommandExecutionInput{
		Player:    player,
		Command:   CommandInfo{Name: "kitchen", Description: "desc", Usage: "/kitchen", Aliases: []string{"ks"}},
		Arguments: []string{" first ", "SECOND"},
	}, false)
	if err != nil || commandExecution.Cancelled ||
		!slices.Equal(commandExecution.Arguments, []string{"first", "SECOND"}) {
		t.Fatalf("command-execution output=%+v error=%v", commandExecution, err)
	}
	if err := pluginRuntime.HandlePlayerDiagnostics(next(), PlayerDiagnosticsInput{
		Player: player, AverageFramesPerSecond: 60, AverageServerSimTickTime: 1,
		AverageClientSimTickTime: 2, AverageBeginFrameTime: 3, AverageInputTime: 4,
		AverageRenderTime: 5, AverageEndFrameTime: 6, AverageRemainderTimePercent: 7,
		AverageUnaccountedTimePercent: 8,
	}); err != nil {
		t.Fatal(err)
	}
}

func assertCSharpHandlerStack(t *testing.T, got, want ItemStack, customName string) {
	t.Helper()
	if got.Identifier != want.Identifier || got.Metadata != want.Metadata || got.Count != want.Count ||
		got.Damage != want.Damage || got.Unbreakable != want.Unbreakable || got.AnvilCost != want.AnvilCost ||
		got.CustomName != customName || !slices.Equal(got.Lore, want.Lore) ||
		!slices.Equal(got.Enchantments, want.Enchantments) {
		t.Fatalf("item stack=%+v, want %+v with custom name %q", got, want, customName)
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
	input := PlayerMoveInput{
		Player:      PlayerSnapshot{Name: "BenchmarkPlayer"},
		NewPosition: Vec3{Y: 64},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := pluginRuntime.HandlePlayerMove(1, input, false); err != nil {
			b.Fatal(err)
		}
	}
}
