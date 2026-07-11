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
	"unsafe"
)

const (
	PlayerMoveSubscription  uint64 = 1
	PlayerChatSubscription  uint64 = 2
	MaxChatReplacementBytes        = 4096
	MaxCommandOutputBytes          = 4096
	MaxCommandEnumBytes            = 4096
)

type PlayerID struct {
	UUID       [16]byte
	Generation uint64
}

type Vec3 struct {
	X, Y, Z float64
}

type Rotation struct {
	Yaw, Pitch float32
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
	Kind   CommandParameterKind
	Name   string
	Values []string
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
	Source    string
	Arguments string
}

type CommandOutput struct {
	Failed  bool
	Message string
}

// Runtime owns a loaded Rust runtime and its plugin libraries.
// Close must not run concurrently with any other method.
type Runtime struct {
	ptr *C.BgRuntimeLibrary
}

func Open(runtimeLibrary, pluginDirectory string) (*Runtime, error) {
	if runtimeLibrary == "" || pluginDirectory == "" {
		return nil, errors.New("runtime library and plugin directory are required")
	}
	libraryPath := C.CString(runtimeLibrary)
	defer C.free(unsafe.Pointer(libraryPath))
	pluginPath := C.CString(pluginDirectory)
	defer C.free(unsafe.Pointer(pluginPath))

	var ptr *C.BgRuntimeLibrary
	var errorBuffer [1024]C.uint8_t
	status := C.bg_runtime_open(
		libraryPath,
		pluginPath,
		&ptr,
		&errorBuffer[0],
		C.uint64_t(len(errorBuffer)),
	)
	if status != C.DF_STATUS_OK {
		message := C.GoString((*C.char)(unsafe.Pointer(&errorBuffer[0])))
		if message == "" {
			message = "unknown native runtime error"
		}
		return nil, fmt.Errorf("open native runtime: %s", message)
	}
	r := &Runtime{ptr: ptr}
	runtime.SetFinalizer(r, func(runtime *Runtime) { runtime.Close() })
	return r, nil
}

func (r *Runtime) Close() {
	if r == nil || r.ptr == nil {
		return
	}
	C.bg_runtime_close(r.ptr)
	r.ptr = nil
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
					Kind: CommandParameterKind(nativeParameter.kind),
					Name: stringView(nativeParameter.name),
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
		source:    C.DfStringView{data: (*C.uint8_t)(source), len: C.uint64_t(len(input.Source))},
		arguments: C.DfStringView{data: (*C.uint8_t)(arguments), len: C.uint64_t(len(input.Arguments))},
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

func (r *Runtime) HandlePlayerMove(input PlayerMoveInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerMoveInput
	fillPlayerID(&nativeInput.player, input.Player)
	nativeInput.old_position = C.DfVec3{x: C.double(input.OldPosition.X), y: C.double(input.OldPosition.Y), z: C.double(input.OldPosition.Z)}
	nativeInput.new_position = C.DfVec3{x: C.double(input.NewPosition.X), y: C.double(input.NewPosition.Y), z: C.double(input.NewPosition.Z)}
	nativeInput.rotation = C.DfRotation{yaw: C.float(input.Rotation.Yaw), pitch: C.float(input.Rotation.Pitch)}
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

func (r *Runtime) HandlePlayerChat(input PlayerChatInput, cancelled bool) (PlayerChatOutput, error) {
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
	if status := C.bg_runtime_handle_player_chat(r.ptr, &nativeInput, &state); status != C.DF_STATUS_OK {
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

func fillPlayerID(destination *C.DfPlayerId, source PlayerID) {
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
