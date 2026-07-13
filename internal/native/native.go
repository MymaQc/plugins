// Package native loads and calls the native plugin runtime.
package native

/*
#cgo CFLAGS: -I../../abi/include
#cgo linux LDFLAGS: -ldl
#include <stdlib.h>
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"
	"unsafe"
)

const (
	PlayerMoveSubscription            uint64 = 1
	PlayerChatSubscription            uint64 = 2
	PlayerJoinSubscription            uint64 = 4
	PlayerQuitSubscription            uint64 = 8
	PlayerHurtSubscription            uint64 = 16
	PlayerHealSubscription            uint64 = 32
	PlayerBlockBreakSubscription      uint64 = 64
	PlayerBlockPlaceSubscription      uint64 = 128
	PlayerFoodLossSubscription        uint64 = 256
	PlayerDeathSubscription           uint64 = 512
	PlayerStartBreakSubscription      uint64 = 1024
	PlayerFireExtinguishSubscription  uint64 = 2048
	PlayerToggleSprintSubscription    uint64 = 4096
	PlayerToggleSneakSubscription     uint64 = 8192
	PlayerJumpSubscription            uint64 = 16384
	PlayerTeleportSubscription        uint64 = 32768
	PlayerExperienceGainSubscription  uint64 = 65536
	PlayerPunchAirSubscription        uint64 = 131072
	PlayerHeldSlotChangeSubscription  uint64 = 262144
	PlayerSleepSubscription           uint64 = 524288
	PlayerBlockPickSubscription       uint64 = 1048576
	PlayerLecternPageTurnSubscription uint64 = 2097152
	PlayerSignEditSubscription        uint64 = 4194304
	PlayerItemUseSubscription         uint64 = 8388608
	PlayerItemUseOnBlockSubscription  uint64 = 16777216
	PlayerItemConsumeSubscription     uint64 = 33554432
	PlayerItemReleaseSubscription     uint64 = 67108864
	PlayerItemDamageSubscription      uint64 = 134217728
	PlayerItemDropSubscription        uint64 = 268435456
	PlayerAttackEntitySubscription    uint64 = 536870912
	PlayerItemUseOnEntitySubscription uint64 = 1073741824
	PlayerChangeWorldSubscription     uint64 = 2147483648
	PlayerRespawnSubscription         uint64 = 4294967296
	MaxChatReplacementBytes                  = 4096
	MaxCommandOutputBytes                    = 4096
	MaxCommandEnumBytes                      = 4096
)

// InvocationID identifies one synchronous command, event, or form callback.
// Zero means that no Dragonfly owner transaction is attached.
type InvocationID uint64

type PlayerID struct {
	UUID       [16]byte
	Generation uint64
}

type EntityID struct {
	UUID       [16]byte
	Generation uint64
}

type Vec3 struct {
	X, Y, Z float64
}

type Rotation struct {
	Yaw, Pitch float64
}

type BlockPos struct {
	X, Y, Z int32
}

type PlayerMoveInput struct {
	Player      PlayerID
	OldPosition Vec3
	NewPosition Vec3
	Rotation    Rotation
}

type PlayerChatInput struct {
	Player  PlayerID
	Message string
}

type PlayerChatOutput struct {
	Cancelled   bool
	Replacement *string
}

type PlayerJoinInput struct {
	Player PlayerID
	Name   string
}

type PlayerQuitInput struct {
	Player PlayerID
	Name   string
}

type DamageSource struct {
	Name                                            string
	Kind                                            DamageSourceKind
	ReducedByArmour, ReducedByResistance, Fire      bool
	IgnoresTotem                                    bool
	FireProtection, FeatherFalling, BlastProtection bool
	ProjectileProtection                            bool
	Entity, SecondaryEntity                         EntityID
	Block                                           *WorldBlock
	Data                                            bool
}

type HealingSource struct {
	Name string
	Kind HealingSourceKind
	Data bool
}

type DamageSourceKind uint32

const (
	DamageSourceCustom DamageSourceKind = iota
	DamageSourceAttack
	DamageSourceBlock
	DamageSourceDrowning
	DamageSourceExplosion
	DamageSourceFall
	DamageSourceFire
	DamageSourceGlide
	DamageSourceInstant
	DamageSourceLava
	DamageSourceLightning
	DamageSourceMagma
	DamageSourcePoison
	DamageSourceProjectile
	DamageSourceStarvation
	DamageSourceSuffocation
	DamageSourceThorns
	DamageSourceVoid
	DamageSourceWither
)

type HealingSourceKind uint32

const (
	HealingSourceCustom HealingSourceKind = iota
	HealingSourceFood
	HealingSourceInstant
	HealingSourceRegeneration
)

type PlayerHurtResult struct {
	Damage     float64
	Vulnerable bool
}

type PlayerHurtInput struct {
	Player         PlayerID
	Damage         float64
	Immune         bool
	AttackImmunity time.Duration
	Source         DamageSource
}

type PlayerHurtOutput struct {
	Cancelled      bool
	Damage         float64
	AttackImmunity time.Duration
}

type PlayerHealInput struct {
	Player PlayerID
	Health float64
	Source HealingSource
}

type PlayerHealOutput struct {
	Cancelled bool
	Health    float64
}

type PlayerBlockBreakInput struct {
	Player     PlayerID
	Position   BlockPos
	Block      string
	Experience int32
}

type PlayerBlockBreakOutput struct {
	Cancelled  bool
	Experience int32
}

type PlayerBlockPlaceInput struct {
	Player   PlayerID
	Position BlockPos
	Block    string
}

type PlayerFoodLossInput struct {
	Player PlayerID
	From   int32
	To     int32
}

type PlayerFoodLossOutput struct {
	Cancelled bool
	To        int32
}

type PlayerDeathInput struct {
	Player PlayerID
	Source DamageSource
}

type PlayerPositionInput struct {
	Player   PlayerID
	Position BlockPos
}
type PlayerToggleInput struct {
	Player PlayerID
	After  bool
}
type PlayerTeleportInput struct {
	Player   PlayerID
	Position Vec3
}
type PlayerExperienceGainOutput struct {
	Cancelled bool
	Amount    int
}
type PlayerHeldSlotChangeInput struct {
	Player PlayerID
	From   int
	To     int
}
type PlayerSleepOutput struct {
	Cancelled    bool
	SendReminder bool
}
type PlayerBlockPickInput struct {
	Player   PlayerID
	Position BlockPos
	Block    string
}
type PlayerLecternPageTurnInput struct {
	Player   PlayerID
	Position BlockPos
	OldPage  int
	NewPage  int
}
type PlayerLecternPageTurnOutput struct {
	Cancelled bool
	NewPage   int
}
type PlayerSignEditInput struct {
	Player    PlayerID
	Position  BlockPos
	FrontSide bool
	OldText   string
	NewText   string
}
type PlayerItemUseOnBlockInput struct {
	Player        PlayerID
	Position      BlockPos
	Face          int
	ClickPosition Vec3
}
type PlayerItemDamageOutput struct {
	Cancelled bool
	Damage    int
}
type PlayerAttackEntityInput struct {
	Player PlayerID
	Target EntityID
}
type PlayerAttackEntityOutput struct {
	Cancelled       bool
	KnockbackForce  float64
	KnockbackHeight float64
	Critical        bool
}
type PlayerItemUseOnEntityInput struct {
	Player PlayerID
	Target EntityID
}
type PlayerChangeWorldInput struct {
	Player PlayerID
	Before *WorldID
	After  WorldID
}
type PlayerRespawnInput struct {
	Player PlayerID
}
type PlayerRespawnOutput struct {
	Position Vec3
	World    WorldID
}

type Command struct {
	Index       uint64
	Name        string
	Description string
	Overloads   []CommandOverload
}

type CommandOverload struct {
	Parameters []CommandParameter
}

type CommandParameter struct {
	Kind     CommandParameterKind
	Name     string
	Values   []string
	Optional bool
}

type CommandParameterKind uint32

const (
	CommandParameterSubcommand  CommandParameterKind = 1
	CommandParameterEnum        CommandParameterKind = 2
	CommandParameterString      CommandParameterKind = 3
	CommandParameterInteger     CommandParameterKind = 4
	CommandParameterFloat       CommandParameterKind = 5
	CommandParameterBool        CommandParameterKind = 6
	CommandParameterDynamicEnum CommandParameterKind = 7
	CommandParameterPlayer      CommandParameterKind = 8
	CommandParameterRawText     CommandParameterKind = 9
)

type CommandInput struct {
	Invocation    InvocationID
	Source        string
	Arguments     string
	SourceKind    CommandSourceKind
	SourcePlayer  *PlayerID
	OnlinePlayers []CommandPlayer
}

type CommandSourceKind uint32

const (
	CommandSourceUnknown CommandSourceKind = iota
	CommandSourcePlayer
	CommandSourceConsole
)

type CommandPlayer struct {
	Player              PlayerID
	Name                string
	LatencyMilliseconds uint64
}

type CommandOutput struct {
	Failed  bool
	Message string
}

// Runtime owns a loaded Rust runtime and its plugin libraries.
// Close must not run concurrently with any other method.
type Runtime struct {
	ptr         *C.BgRuntimeLibrary
	hostContext uint64
}

func Open(runtimeLibrary, pluginDirectory string) (*Runtime, error) {
	return OpenWithHost(runtimeLibrary, pluginDirectory, nil)
}

func OpenWithHost(runtimeLibrary, pluginDirectory string, host Host) (*Runtime, error) {
	if runtimeLibrary == "" || pluginDirectory == "" {
		return nil, errors.New("runtime library and plugin directory are required")
	}
	libraryPath := C.CString(runtimeLibrary)
	defer C.free(unsafe.Pointer(libraryPath))
	pluginPath := C.CString(pluginDirectory)
	defer C.free(unsafe.Pointer(pluginPath))

	hostContext := registerHost(host)
	var ptr *C.BgRuntimeLibrary
	var errorBuffer [1024]C.uint8_t
	status := C.bg_runtime_open(
		libraryPath,
		pluginPath,
		C.uint64_t(hostContext),
		&ptr,
		&errorBuffer[0],
		C.uint64_t(len(errorBuffer)),
	)
	if status != C.DF_STATUS_OK {
		unregisterHost(hostContext)
		message := C.GoString((*C.char)(unsafe.Pointer(&errorBuffer[0])))
		if message == "" {
			message = "unknown native runtime error"
		}
		return nil, fmt.Errorf("open native runtime: %s", message)
	}
	r := &Runtime{ptr: ptr, hostContext: hostContext}
	runtime.SetFinalizer(r, func(runtime *Runtime) { runtime.Close() })
	return r, nil
}

func (r *Runtime) Close() {
	if r == nil || r.ptr == nil {
		return
	}
	drainHostForms(r.hostContext, true)
	C.bg_runtime_close(r.ptr)
	unregisterHost(r.hostContext)
	r.ptr = nil
	r.hostContext = 0
	runtime.SetFinalizer(r, nil)
}

func (r *Runtime) Enable() error {
	if r == nil || r.ptr == nil {
		return errors.New("native runtime is closed")
	}
	if status := C.bg_runtime_enable(r.ptr); status != C.DF_STATUS_OK {
		return fmt.Errorf("enable native plugins: status %d", int32(status))
	}
	return nil
}

func (r *Runtime) Disable() {
	if r != nil && r.ptr != nil {
		drainHostForms(r.hostContext, false)
		C.bg_runtime_disable(r.ptr)
	}
}

func (r *Runtime) PluginCount() uint64 {
	if r == nil || r.ptr == nil {
		return 0
	}
	return uint64(C.bg_runtime_plugin_count(r.ptr))
}

func (r *Runtime) Subscriptions() uint64 {
	if r == nil || r.ptr == nil {
		return 0
	}
	return uint64(C.bg_runtime_subscriptions(r.ptr))
}

func (r *Runtime) Commands() ([]Command, error) {
	if r == nil || r.ptr == nil {
		return nil, errors.New("native runtime is closed")
	}
	count := uint64(C.bg_runtime_command_count(r.ptr))
	commands := make([]Command, 0, count)
	for index := uint64(0); index < count; index++ {
		var descriptor C.DfCommandDescriptor
		if status := C.bg_runtime_command_at(r.ptr, C.uint64_t(index), &descriptor); status != C.DF_STATUS_OK {
			return nil, fmt.Errorf("read native command %d: status %d", index, int32(status))
		}
		command := Command{
			Index:       index,
			Name:        stringView(descriptor.name),
			Description: stringView(descriptor.description),
		}
		if descriptor.overload_count > 0 && descriptor.overloads == nil {
			return nil, fmt.Errorf("read native command %d: null overloads", index)
		}
		overloads := unsafe.Slice(descriptor.overloads, int(descriptor.overload_count))
		for _, nativeOverload := range overloads {
			overload := CommandOverload{}
			if nativeOverload.parameter_count > 0 && nativeOverload.parameters == nil {
				return nil, fmt.Errorf("read native command %d: null parameters", index)
			}
			parameters := unsafe.Slice(nativeOverload.parameters, int(nativeOverload.parameter_count))
			for _, nativeParameter := range parameters {
				parameter := CommandParameter{
					Kind:     CommandParameterKind(nativeParameter.kind),
					Name:     stringView(nativeParameter.name),
					Optional: nativeParameter.optional != 0,
				}
				if nativeParameter.value_count > 0 && nativeParameter.values == nil {
					return nil, fmt.Errorf("read native command %d: null enum values", index)
				}
				values := unsafe.Slice(nativeParameter.values, int(nativeParameter.value_count))
				for _, value := range values {
					parameter.Values = append(parameter.Values, stringView(value))
				}
				overload.Parameters = append(overload.Parameters, parameter)
			}
			command.Overloads = append(command.Overloads, overload)
		}
		commands = append(commands, command)
	}
	return commands, nil
}

func (r *Runtime) HandleCommand(index uint64, input CommandInput) (CommandOutput, error) {
	var output CommandOutput
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	source := C.CBytes([]byte(input.Source))
	defer C.free(source)
	arguments := C.CBytes([]byte(input.Arguments))
	defer C.free(arguments)
	message := C.malloc(MaxCommandOutputBytes)
	if message == nil {
		return output, errors.New("allocate command output buffer")
	}
	defer C.free(message)

	nativeInput := C.DfCommandInput{
		invocation:  C.DfInvocationId(input.Invocation),
		source:      C.DfStringView{data: (*C.uint8_t)(source), len: C.uint64_t(len(input.Source))},
		arguments:   C.DfStringView{data: (*C.uint8_t)(arguments), len: C.uint64_t(len(input.Arguments))},
		source_kind: C.uint32_t(input.SourceKind),
	}
	if input.SourcePlayer != nil {
		fillPlayerID(&nativeInput.source_player, *input.SourcePlayer)
		nativeInput.source_kind = C.DF_COMMAND_SOURCE_PLAYER
	}
	var playerNames []unsafe.Pointer
	if len(input.OnlinePlayers) != 0 {
		memory := C.malloc(C.size_t(len(input.OnlinePlayers)) * C.size_t(unsafe.Sizeof(C.DfCommandPlayer{})))
		if memory == nil {
			return output, errors.New("allocate command player snapshots")
		}
		defer C.free(memory)
		nativeInput.online_players = (*C.DfCommandPlayer)(memory)
		nativeInput.online_player_count = C.uint64_t(len(input.OnlinePlayers))
		players := unsafe.Slice(nativeInput.online_players, len(input.OnlinePlayers))
		for index, snapshot := range input.OnlinePlayers {
			name := C.CBytes([]byte(snapshot.Name))
			playerNames = append(playerNames, name)
			fillPlayerID(&players[index].player, snapshot.Player)
			players[index].name = C.DfStringView{data: (*C.uint8_t)(name), len: C.uint64_t(len(snapshot.Name))}
			players[index].latency_milliseconds = C.uint64_t(snapshot.LatencyMilliseconds)
		}
		defer func() {
			for _, name := range playerNames {
				C.free(name)
			}
		}()
	}
	state := C.DfCommandState{
		output: C.DfStringBuffer{data: (*C.uint8_t)(message), capacity: MaxCommandOutputBytes},
	}
	if status := C.bg_runtime_handle_command(r.ptr, C.uint64_t(index), &nativeInput, &state); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native command handler failed with status %d", int32(status))
	}
	output.Failed = state.failed != 0
	output.Message = string(C.GoBytes(message, C.int(state.output.len)))
	return output, nil
}

func (r *Runtime) CommandEnumOptions(index, overload, parameter uint64, sourceName string, onlinePlayers []string) ([]string, error) {
	if r == nil || r.ptr == nil {
		return nil, errors.New("native runtime is closed")
	}
	source := C.CBytes([]byte(sourceName))
	defer C.free(source)
	var viewsPointer *C.DfStringView
	var playerStrings []unsafe.Pointer
	if len(onlinePlayers) != 0 {
		viewsMemory := C.malloc(C.size_t(len(onlinePlayers)) * C.size_t(unsafe.Sizeof(C.DfStringView{})))
		if viewsMemory == nil {
			return nil, errors.New("allocate online player views")
		}
		defer C.free(viewsMemory)
		viewsPointer = (*C.DfStringView)(viewsMemory)
		views := unsafe.Slice(viewsPointer, len(onlinePlayers))
		for index, name := range onlinePlayers {
			value := C.CBytes([]byte(name))
			playerStrings = append(playerStrings, value)
			views[index] = C.DfStringView{data: (*C.uint8_t)(value), len: C.uint64_t(len(name))}
		}
		defer func() {
			for _, value := range playerStrings {
				C.free(value)
			}
		}()
	}
	buffer := C.malloc(MaxCommandEnumBytes)
	if buffer == nil {
		return nil, errors.New("allocate command enum buffer")
	}
	defer C.free(buffer)
	output := C.DfStringBuffer{data: (*C.uint8_t)(buffer), capacity: MaxCommandEnumBytes}
	context := C.DfCommandEnumContext{
		source:              C.DfStringView{data: (*C.uint8_t)(source), len: C.uint64_t(len(sourceName))},
		online_players:      viewsPointer,
		online_player_count: C.uint64_t(len(onlinePlayers)),
	}
	status := C.bg_runtime_command_enum_options(
		r.ptr,
		C.uint64_t(index),
		C.uint64_t(overload),
		C.uint64_t(parameter),
		&context,
		&output,
	)
	if status != C.DF_STATUS_OK {
		return nil, fmt.Errorf("native command enum handler failed with status %d", int32(status))
	}
	if output.len == 0 {
		return nil, nil
	}
	return strings.Split(string(C.GoBytes(buffer, C.int(output.len))), "\n"), nil
}

func (r *Runtime) HandlePlayerMove(invocation InvocationID, input PlayerMoveInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerMoveInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.old_position = C.DfVec3{x: C.double(input.OldPosition.X), y: C.double(input.OldPosition.Y), z: C.double(input.OldPosition.Z)}
	nativeInput.new_position = C.DfVec3{x: C.double(input.NewPosition.X), y: C.double(input.NewPosition.Y), z: C.double(input.NewPosition.Z)}
	nativeInput.rotation = C.DfRotation{yaw: C.double(input.Rotation.Yaw), pitch: C.double(input.Rotation.Pitch)}
	var state C.DfPlayerMoveState
	if cancelled {
		state.cancelled = 1
	}
	packed := uint64(C.bg_runtime_handle_player_move_value(r.ptr, nativeInput, state.cancelled))
	status := int32(uint32(packed >> 32))
	finalCancelled := uint8(packed) != 0
	if status != C.DF_STATUS_OK {
		return finalCancelled, fmt.Errorf("native movement handler failed with status %d", status)
	}
	return finalCancelled, nil
}

func (r *Runtime) HandlePlayerChat(invocation InvocationID, input PlayerChatInput, cancelled bool) (PlayerChatOutput, error) {
	output := PlayerChatOutput{Cancelled: cancelled}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	message := C.CBytes([]byte(input.Message))
	defer C.free(message)
	replacement := C.malloc(MaxChatReplacementBytes)
	if replacement == nil {
		return output, errors.New("allocate chat replacement buffer")
	}
	defer C.free(replacement)

	var nativeInput C.DfPlayerChatInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.message = C.DfStringView{
		data: (*C.uint8_t)(message),
		len:  C.uint64_t(len(input.Message)),
	}
	state := C.DfPlayerChatState{
		replacement: C.DfStringBuffer{
			data:     (*C.uint8_t)(replacement),
			capacity: MaxChatReplacementBytes,
		},
	}
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_CHAT, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		output.Cancelled = state.cancelled != 0
		return output, fmt.Errorf("native chat handler failed with status %d", int32(status))
	}
	output.Cancelled = state.cancelled != 0
	if state.has_replacement != 0 {
		value := string(C.GoBytes(replacement, C.int(state.replacement.len)))
		output.Replacement = &value
	}
	return output, nil
}

func (r *Runtime) HandlePlayerJoin(invocation InvocationID, input PlayerJoinInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	name := C.CBytes([]byte(input.Name))
	defer C.free(name)
	var nativeInput C.DfPlayerJoinInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.name = C.DfStringView{data: (*C.uint8_t)(name), len: C.uint64_t(len(input.Name))}
	var state C.DfPlayerJoinState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_JOIN, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native join handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerQuit(invocation InvocationID, input PlayerQuitInput) error {
	if r == nil || r.ptr == nil {
		return errors.New("native runtime is closed")
	}
	name := C.CBytes([]byte(input.Name))
	defer C.free(name)
	var nativeInput C.DfPlayerQuitInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.name = C.DfStringView{data: (*C.uint8_t)(name), len: C.uint64_t(len(input.Name))}
	var state C.DfPlayerQuitState
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_QUIT, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return fmt.Errorf("native quit handler failed with status %d", int32(status))
	}
	return nil
}

func (r *Runtime) HandlePlayerHurt(invocation InvocationID, input PlayerHurtInput, cancelled bool) (PlayerHurtOutput, error) {
	output := PlayerHurtOutput{Cancelled: cancelled, Damage: input.Damage, AttackImmunity: input.AttackImmunity}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	source, releaseSource, ok := nativeDamageSource(input.Source)
	if !ok {
		return output, errors.New("encode hurt damage source")
	}
	defer releaseSource()
	var nativeInput C.DfPlayerHurtInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.immune = C.uint8_t(boolByte(input.Immune))
	nativeInput.source = source
	state := C.DfPlayerHurtState{
		damage:                       C.double(input.Damage),
		attack_immunity_milliseconds: C.uint64_t(max(input.AttackImmunity.Milliseconds(), 0)),
	}
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_HURT, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native hurt handler failed with status %d", int32(status))
	}
	output.Cancelled = state.cancelled != 0
	output.Damage = float64(state.damage)
	output.AttackImmunity = time.Duration(state.attack_immunity_milliseconds) * time.Millisecond
	return output, nil
}

func (r *Runtime) HandlePlayerHeal(invocation InvocationID, input PlayerHealInput, cancelled bool) (PlayerHealOutput, error) {
	output := PlayerHealOutput{Cancelled: cancelled, Health: input.Health}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	source := C.CBytes([]byte(input.Source.Name))
	defer C.free(source)
	var nativeInput C.DfPlayerHealInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.source = C.DfHealingSourceView{
		name: C.DfStringView{data: (*C.uint8_t)(source), len: C.uint64_t(len(input.Source.Name))},
		kind: C.uint32_t(input.Source.Kind), data: C.uint8_t(boolByte(input.Source.Data)),
	}
	state := C.DfPlayerHealState{health: C.double(input.Health)}
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_HEAL, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native heal handler failed with status %d", int32(status))
	}
	output.Cancelled = state.cancelled != 0
	output.Health = float64(state.health)
	return output, nil
}

func (r *Runtime) HandlePlayerBlockBreak(invocation InvocationID, input PlayerBlockBreakInput, cancelled bool) (PlayerBlockBreakOutput, error) {
	output := PlayerBlockBreakOutput{Cancelled: cancelled, Experience: input.Experience}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	block := C.CBytes([]byte(input.Block))
	defer C.free(block)
	var nativeInput C.DfPlayerBlockBreakInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.position = nativeBlockPos(input.Position)
	nativeInput.block = C.DfStringView{data: (*C.uint8_t)(block), len: C.uint64_t(len(input.Block))}
	state := C.DfPlayerBlockBreakState{experience: C.int32_t(input.Experience)}
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_BLOCK_BREAK, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native block-break handler failed with status %d", int32(status))
	}
	output.Cancelled = state.cancelled != 0
	output.Experience = int32(state.experience)
	return output, nil
}

func (r *Runtime) HandlePlayerBlockPlace(invocation InvocationID, input PlayerBlockPlaceInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	block := C.CBytes([]byte(input.Block))
	defer C.free(block)
	var nativeInput C.DfPlayerBlockPlaceInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.position = nativeBlockPos(input.Position)
	nativeInput.block = C.DfStringView{data: (*C.uint8_t)(block), len: C.uint64_t(len(input.Block))}
	var state C.DfPlayerBlockPlaceState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_BLOCK_PLACE, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native block-place handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerFoodLoss(invocation InvocationID, input PlayerFoodLossInput, cancelled bool) (PlayerFoodLossOutput, error) {
	output := PlayerFoodLossOutput{Cancelled: cancelled, To: input.To}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerFoodLossInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.from = C.int32_t(input.From)
	state := C.DfPlayerFoodLossState{to: C.int32_t(input.To)}
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_FOOD_LOSS, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native food-loss handler failed with status %d", int32(status))
	}
	output.Cancelled = state.cancelled != 0
	output.To = int32(state.to)
	return output, nil
}

func (r *Runtime) HandlePlayerDeath(invocation InvocationID, input PlayerDeathInput, keepInventory bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return keepInventory, errors.New("native runtime is closed")
	}
	source, releaseSource, ok := nativeDamageSource(input.Source)
	if !ok {
		return keepInventory, errors.New("encode death damage source")
	}
	defer releaseSource()
	var nativeInput C.DfPlayerDeathInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.source = source
	var state C.DfPlayerDeathState
	if keepInventory {
		state.keep_inventory = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_DEATH, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return keepInventory, fmt.Errorf("native death handler failed with status %d", int32(status))
	}
	return state.keep_inventory != 0, nil
}

func (r *Runtime) HandlePlayerStartBreak(invocation InvocationID, input PlayerPositionInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerStartBreakInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.position = nativeBlockPos(input.Position)
	var state C.DfPlayerStartBreakState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_START_BREAK, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native start-break handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerFireExtinguish(invocation InvocationID, input PlayerPositionInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerFireExtinguishInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.position = nativeBlockPos(input.Position)
	var state C.DfPlayerFireExtinguishState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_FIRE_EXTINGUISH, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native fire-extinguish handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerToggleSprint(invocation InvocationID, input PlayerToggleInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerToggleSprintInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.after = C.uint8_t(boolByte(input.After))
	var state C.DfPlayerToggleSprintState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_TOGGLE_SPRINT, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native toggle-sprint handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}
func (r *Runtime) HandlePlayerToggleSneak(invocation InvocationID, input PlayerToggleInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerToggleSneakInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.after = C.uint8_t(boolByte(input.After))
	var state C.DfPlayerToggleSneakState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_TOGGLE_SNEAK, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native toggle-sneak handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerJump(invocation InvocationID, player PlayerID) error {
	if r == nil || r.ptr == nil {
		return errors.New("native runtime is closed")
	}
	var input C.DfPlayerJumpInput
	input.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&input.player, player)
	var state C.DfPlayerJumpState
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_JUMP, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return fmt.Errorf("native jump handler failed with status %d", int32(status))
	}
	return nil
}

func (r *Runtime) HandlePlayerTeleport(invocation InvocationID, input PlayerTeleportInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerTeleportInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.position = C.DfVec3{x: C.double(input.Position.X), y: C.double(input.Position.Y), z: C.double(input.Position.Z)}
	var state C.DfPlayerTeleportState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_TELEPORT, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native teleport handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerExperienceGain(invocation InvocationID, player PlayerID, amount int, cancelled bool) (PlayerExperienceGainOutput, error) {
	output := PlayerExperienceGainOutput{Cancelled: cancelled, Amount: amount}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	var input C.DfPlayerExperienceGainInput
	input.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&input.player, player)
	state := C.DfPlayerExperienceGainState{amount: C.int32_t(amount)}
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_EXPERIENCE_GAIN, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native experience-gain handler failed with status %d", int32(status))
	}
	return PlayerExperienceGainOutput{Cancelled: state.cancelled != 0, Amount: int(state.amount)}, nil
}

func (r *Runtime) HandlePlayerPunchAir(invocation InvocationID, player PlayerID, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var input C.DfPlayerPunchAirInput
	input.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&input.player, player)
	var state C.DfPlayerPunchAirState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_PUNCH_AIR, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native punch-air handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerHeldSlotChange(invocation InvocationID, input PlayerHeldSlotChangeInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerHeldSlotChangeInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.from = C.int32_t(input.From)
	nativeInput.to = C.int32_t(input.To)
	var state C.DfPlayerHeldSlotChangeState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_HELD_SLOT_CHANGE, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native held-slot-change handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerSleep(invocation InvocationID, player PlayerID, sendReminder, cancelled bool) (PlayerSleepOutput, error) {
	output := PlayerSleepOutput{Cancelled: cancelled, SendReminder: sendReminder}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	var input C.DfPlayerSleepInput
	input.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&input.player, player)
	state := C.DfPlayerSleepState{send_reminder: C.uint8_t(boolByte(sendReminder))}
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_SLEEP, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native sleep handler failed with status %d", int32(status))
	}
	return PlayerSleepOutput{Cancelled: state.cancelled != 0, SendReminder: state.send_reminder != 0}, nil
}

func (r *Runtime) HandlePlayerBlockPick(invocation InvocationID, input PlayerBlockPickInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	block := unsafe.StringData(input.Block)
	nativeInput := C.DfPlayerBlockPickInput{invocation: C.DfInvocationId(invocation), position: nativeBlockPos(input.Position), block: C.DfStringView{data: (*C.uint8_t)(unsafe.Pointer(block)), len: C.uint64_t(len(input.Block))}}
	fillPlayerID(&nativeInput.player, input.Player)
	var state C.DfPlayerBlockPickState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_BLOCK_PICK, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native block-pick handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerLecternPageTurn(invocation InvocationID, input PlayerLecternPageTurnInput, cancelled bool) (PlayerLecternPageTurnOutput, error) {
	output := PlayerLecternPageTurnOutput{Cancelled: cancelled, NewPage: input.NewPage}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerLecternPageTurnInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.position = nativeBlockPos(input.Position)
	nativeInput.old_page = C.int32_t(input.OldPage)
	state := C.DfPlayerLecternPageTurnState{new_page: C.int32_t(input.NewPage)}
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_LECTERN_PAGE_TURN, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native lectern-page-turn handler failed with status %d", int32(status))
	}
	return PlayerLecternPageTurnOutput{Cancelled: state.cancelled != 0, NewPage: int(state.new_page)}, nil
}

func (r *Runtime) HandlePlayerSignEdit(invocation InvocationID, input PlayerSignEditInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	oldText, newText := unsafe.StringData(input.OldText), unsafe.StringData(input.NewText)
	nativeInput := C.DfPlayerSignEditInput{
		invocation: C.DfInvocationId(invocation),
		position:   nativeBlockPos(input.Position),
		front_side: C.uint8_t(boolByte(input.FrontSide)),
		old_text:   C.DfStringView{data: (*C.uint8_t)(unsafe.Pointer(oldText)), len: C.uint64_t(len(input.OldText))},
		new_text:   C.DfStringView{data: (*C.uint8_t)(unsafe.Pointer(newText)), len: C.uint64_t(len(input.NewText))},
	}
	fillPlayerID(&nativeInput.player, input.Player)
	var state C.DfPlayerSignEditState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_SIGN_EDIT, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native sign-edit handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerItemUse(invocation InvocationID, player PlayerID, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var input C.DfPlayerItemUseInput
	input.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&input.player, player)
	var state C.DfPlayerItemUseState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_ITEM_USE, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native item-use handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerItemUseOnBlock(invocation InvocationID, input PlayerItemUseOnBlockInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerItemUseOnBlockInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.position = nativeBlockPos(input.Position)
	nativeInput.face = C.int32_t(input.Face)
	nativeInput.click_position = C.DfVec3{x: C.double(input.ClickPosition.X), y: C.double(input.ClickPosition.Y), z: C.double(input.ClickPosition.Z)}
	var state C.DfPlayerItemUseOnBlockState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_ITEM_USE_ON_BLOCK, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native item-use-on-block handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerItemConsume(invocation InvocationID, player PlayerID, item ItemStack, cancelled bool) (bool, error) {
	return r.handlePlayerItemStackEvent(invocation, C.DF_EVENT_PLAYER_ITEM_CONSUME, player, item, 0, cancelled)
}

func (r *Runtime) HandlePlayerItemRelease(invocation InvocationID, player PlayerID, item ItemStack, duration time.Duration, cancelled bool) (bool, error) {
	milliseconds := duration.Milliseconds()
	if milliseconds < 0 {
		milliseconds = 0
	}
	return r.handlePlayerItemStackEvent(invocation, C.DF_EVENT_PLAYER_ITEM_RELEASE, player, item, uint64(milliseconds), cancelled)
}

func (r *Runtime) handlePlayerItemStackEvent(invocation InvocationID, event C.DfEventId, player PlayerID, item ItemStack, duration uint64, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	stack, ok := r.openEventItemSnapshot(item)
	if !ok {
		return cancelled, errors.New("open item event snapshot")
	}
	defer unregisterItemSnapshot(r.hostContext, uint64(stack.snapshot))
	var state C.DfPlayerItemConsumeState
	if cancelled {
		state.cancelled = 1
	}
	if event == C.DF_EVENT_PLAYER_ITEM_CONSUME {
		var input C.DfPlayerItemConsumeInput
		input.invocation = C.DfInvocationId(invocation)
		fillPlayerID(&input.player, player)
		input.item = stack
		if status := C.bg_runtime_handle_event(r.ptr, event, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
			return state.cancelled != 0, fmt.Errorf("native item-consume handler failed with status %d", int32(status))
		}
		return state.cancelled != 0, nil
	}
	var input C.DfPlayerItemReleaseInput
	input.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&input.player, player)
	input.item = stack
	input.duration_milliseconds = C.uint64_t(duration)
	if status := C.bg_runtime_handle_event(r.ptr, event, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native item-release handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerItemDamage(invocation InvocationID, player PlayerID, item ItemStack, damage int, cancelled bool) (PlayerItemDamageOutput, error) {
	output := PlayerItemDamageOutput{Cancelled: cancelled, Damage: damage}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	var input C.DfPlayerItemDamageInput
	input.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&input.player, player)
	stack, ok := r.openEventItemSnapshot(item)
	if !ok {
		return output, errors.New("open item event snapshot")
	}
	defer unregisterItemSnapshot(r.hostContext, uint64(stack.snapshot))
	input.item = stack
	state := C.DfPlayerItemDamageState{damage: C.int32_t(damage)}
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_ITEM_DAMAGE, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native item-damage handler failed with status %d", int32(status))
	}
	return PlayerItemDamageOutput{Cancelled: state.cancelled != 0, Damage: int(state.damage)}, nil
}

func (r *Runtime) HandlePlayerItemDrop(invocation InvocationID, player PlayerID, item ItemStack, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var input C.DfPlayerItemDropInput
	input.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&input.player, player)
	stack, ok := r.openEventItemSnapshot(item)
	if !ok {
		return cancelled, errors.New("open item event snapshot")
	}
	defer unregisterItemSnapshot(r.hostContext, uint64(stack.snapshot))
	input.item = stack
	var state C.DfPlayerItemDropState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_ITEM_DROP, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return state.cancelled != 0, fmt.Errorf("native item-drop handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerAttackEntity(invocation InvocationID, input PlayerAttackEntityInput, force, height float64, critical, cancelled bool) (PlayerAttackEntityOutput, error) {
	output := PlayerAttackEntityOutput{Cancelled: cancelled, KnockbackForce: force, KnockbackHeight: height, Critical: critical}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerAttackEntityInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	fillEntityID(&nativeInput.target, input.Target)
	state := C.DfPlayerAttackEntityState{
		knockback_force: C.double(force), knockback_height: C.double(height),
		critical: C.uint8_t(boolByte(critical)), cancelled: C.uint8_t(boolByte(cancelled)),
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_ATTACK_ENTITY, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native attack-entity handler failed with status %d", int32(status))
	}
	return PlayerAttackEntityOutput{
		Cancelled: state.cancelled != 0, KnockbackForce: float64(state.knockback_force),
		KnockbackHeight: float64(state.knockback_height), Critical: state.critical != 0,
	}, nil
}

func (r *Runtime) HandlePlayerItemUseOnEntity(invocation InvocationID, input PlayerItemUseOnEntityInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerItemUseOnEntityInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	fillEntityID(&nativeInput.target, input.Target)
	var state C.DfPlayerItemUseOnEntityState
	state.cancelled = C.uint8_t(boolByte(cancelled))
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_ITEM_USE_ON_ENTITY, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return cancelled, fmt.Errorf("native item-use-on-entity handler failed with status %d", int32(status))
	}
	return state.cancelled != 0, nil
}

func (r *Runtime) HandlePlayerChangeWorld(invocation InvocationID, input PlayerChangeWorldInput) error {
	if r == nil || r.ptr == nil {
		return errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerChangeWorldInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	if input.Before != nil {
		nativeInput.before.value = C.uint64_t(*input.Before)
	}
	nativeInput.after.value = C.uint64_t(input.After)
	var state C.DfPlayerChangeWorldState
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_CHANGE_WORLD, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return fmt.Errorf("native change-world handler failed with status %d", int32(status))
	}
	return nil
}

func (r *Runtime) HandlePlayerRespawn(invocation InvocationID, input PlayerRespawnInput, position Vec3, world WorldID) (PlayerRespawnOutput, error) {
	output := PlayerRespawnOutput{Position: position, World: world}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerRespawnInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillPlayerID(&nativeInput.player, input.Player)
	state := C.DfPlayerRespawnState{
		position: C.DfVec3{x: C.double(position.X), y: C.double(position.Y), z: C.double(position.Z)},
		world:    C.DfWorldId{value: C.uint64_t(world)},
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_RESPAWN, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native respawn handler failed with status %d", int32(status))
	}
	return PlayerRespawnOutput{
		Position: Vec3{X: float64(state.position.x), Y: float64(state.position.y), Z: float64(state.position.z)},
		World:    WorldID(state.world.value),
	}, nil
}

func (r *Runtime) openEventItemSnapshot(item ItemStack) (C.DfItemStackSnapshot, bool) {
	if r == nil || r.ptr == nil || !validNativeItem(item) {
		return C.DfItemStackSnapshot{}, false
	}
	id, ok := registerItemSnapshot(r.hostContext, item)
	if !ok {
		return C.DfItemStackSnapshot{}, false
	}
	var info C.DfItemStackInfo
	fillItemStackInfo(&info, item)
	return C.DfItemStackSnapshot{snapshot: C.uint64_t(id), info: info}, true
}

func nativeBlockPos(position BlockPos) C.DfBlockPos {
	return C.DfBlockPos{x: C.int32_t(position.X), y: C.int32_t(position.Y), z: C.int32_t(position.Z)}
}

func boolByte(value bool) uint8 {
	if value {
		return 1
	}
	return 0
}

func nativeDamageSource(source DamageSource) (C.DfDamageSourceView, func(), bool) {
	allocations := make([]unsafe.Pointer, 0, 4)
	release := func() {
		for _, allocation := range allocations {
			C.free(allocation)
		}
	}
	name := C.CBytes([]byte(source.Name))
	if len(source.Name) != 0 && name == nil {
		return C.DfDamageSourceView{}, func() {}, false
	}
	allocations = append(allocations, name)
	var flags uint32
	if source.ReducedByArmour {
		flags |= C.DF_DAMAGE_SOURCE_REDUCED_BY_ARMOUR
	}
	if source.ReducedByResistance {
		flags |= C.DF_DAMAGE_SOURCE_REDUCED_BY_RESISTANCE
	}
	if source.Fire {
		flags |= C.DF_DAMAGE_SOURCE_FIRE
	}
	if source.IgnoresTotem {
		flags |= C.DF_DAMAGE_SOURCE_IGNORES_TOTEM
	}
	if source.FireProtection {
		flags |= C.DF_DAMAGE_SOURCE_FIRE_PROTECTION
	}
	if source.FeatherFalling {
		flags |= C.DF_DAMAGE_SOURCE_FEATHER_FALLING
	}
	if source.BlastProtection {
		flags |= C.DF_DAMAGE_SOURCE_BLAST_PROTECTION
	}
	if source.ProjectileProtection {
		flags |= C.DF_DAMAGE_SOURCE_PROJECTILE_PROTECTION
	}
	view := C.DfDamageSourceView{
		name: C.DfStringView{data: (*C.uint8_t)(name), len: C.uint64_t(len(source.Name))},
		kind: C.uint32_t(source.Kind), flags: C.uint32_t(flags), data: C.uint8_t(boolByte(source.Data)),
	}
	fillEntityID(&view.entity, source.Entity)
	fillEntityID(&view.secondary_entity, source.SecondaryEntity)
	if source.Block != nil {
		block := (*C.DfBlockView)(C.malloc(C.size_t(C.sizeof_DfBlockView)))
		identifier := C.CBytes([]byte(source.Block.Identifier))
		properties := C.CBytes(source.Block.PropertiesNBT)
		if block == nil || len(source.Block.Identifier) != 0 && identifier == nil || len(source.Block.PropertiesNBT) != 0 && properties == nil {
			C.free(unsafe.Pointer(block))
			C.free(identifier)
			C.free(properties)
			release()
			return C.DfDamageSourceView{}, func() {}, false
		}
		allocations = append(allocations, unsafe.Pointer(block), identifier, properties)
		*block = C.DfBlockView{
			identifier:     C.DfStringView{data: (*C.uint8_t)(identifier), len: C.uint64_t(len(source.Block.Identifier))},
			properties_nbt: C.DfStringView{data: (*C.uint8_t)(properties), len: C.uint64_t(len(source.Block.PropertiesNBT))},
		}
		view.block = block
	}
	return view, release, true
}

func fillPlayerID(destination *C.DfPlayerId, source PlayerID) {
	for i, value := range source.UUID {
		destination.bytes[i] = C.uint8_t(value)
	}
	destination.generation = C.uint64_t(source.Generation)
}

func fillEntityID(destination *C.DfEntityId, source EntityID) {
	for i, value := range source.UUID {
		destination.bytes[i] = C.uint8_t(value)
	}
	destination.generation = C.uint64_t(source.Generation)
}

func stringView(view C.DfStringView) string {
	if view.len == 0 {
		return ""
	}
	return string(C.GoBytes(unsafe.Pointer(view.data), C.int(view.len)))
}
