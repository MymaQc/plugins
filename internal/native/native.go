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
	"unsafe"
)

const (
	PlayerMoveSubscription  uint64 = 1
	PlayerChatSubscription  uint64 = 2
	MaxChatReplacementBytes        = 4096
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
