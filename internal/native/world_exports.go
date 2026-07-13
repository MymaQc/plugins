package native

/*
#include "bridge.h"
*/
import "C"

import (
	"unicode/utf8"
	"unsafe"
)

const (
	maxWorldNameBytes       = 256
	maxBlockIdentifierBytes = 256
	maxBlockPropertiesBytes = 64 << 10
	maxSourceNameBytes      = 64 << 10
)

//export bg_go_world_lookup
func bg_go_world_lookup(context C.uint64_t, invocation C.DfInvocationId, name C.DfStringView, output *C.DfWorldId) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	value, valid := copyWorldString(name, maxWorldNameBytes)
	if !ok || !valid || output == nil {
		return C.DF_STATUS_ERROR
	}
	id, ok := host.WorldByName(InvocationID(invocation), value)
	if !ok || id == 0 {
		return C.DF_STATUS_ERROR
	}
	output.value = C.uint64_t(id)
	return C.DF_STATUS_OK
}

//export bg_go_world_open
func bg_go_world_open(context C.uint64_t, invocation C.DfInvocationId, name C.DfStringView, dimension C.uint32_t, output *C.DfWorldId) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	value, valid := copyWorldString(name, maxWorldNameBytes)
	if !ok || !valid || output == nil || uint32(dimension) > uint32(WorldDimensionEnd) {
		return C.DF_STATUS_ERROR
	}
	id, ok := host.OpenWorld(InvocationID(invocation), value, WorldDimension(dimension))
	if !ok || id == 0 {
		return C.DF_STATUS_ERROR
	}
	output.value = C.uint64_t(id)
	return C.DF_STATUS_OK
}

//export bg_go_world_name
func bg_go_world_name(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, output *C.DfStringBuffer) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	name, ok := host.WorldName(InvocationID(invocation), WorldID(world.value))
	if !ok || name == "" || len(name) > maxWorldNameBytes || !utf8.ValidString(name) || !canWriteSkinBuffer(output, []byte(name)) {
		return C.DF_STATUS_ERROR
	}
	writeSkinBuffer(output, []byte(name))
	return C.DF_STATUS_OK
}

//export bg_go_world_unload
func bg_go_world_unload(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.UnloadWorld(InvocationID(invocation), WorldID(world.value)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_world_save
func bg_go_world_save(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SaveWorld(InvocationID(invocation), WorldID(world.value)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_world_block_get
func bg_go_world_block_get(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, output *C.DfBlockData) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	block, ok := host.WorldBlock(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position))
	identifier, properties := []byte(block.Identifier), block.PropertiesNBT
	if !ok || len(identifier) == 0 || len(identifier) > maxBlockIdentifierBytes || !utf8.Valid(identifier) || len(properties) > maxBlockPropertiesBytes {
		return C.DF_STATUS_ERROR
	}
	output.identifier.len = C.uint64_t(len(identifier))
	output.properties_nbt.len = C.uint64_t(len(properties))
	if !canWriteSkinBuffer(&output.identifier, identifier) || !canWriteSkinBuffer(&output.properties_nbt, properties) {
		return C.DF_STATUS_ERROR
	}
	writeSkinBuffer(&output.identifier, identifier)
	writeSkinBuffer(&output.properties_nbt, properties)
	return C.DF_STATUS_OK
}

//export bg_go_world_block_set
func bg_go_world_block_set(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, view *C.DfBlockView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || view == nil {
		return C.DF_STATUS_ERROR
	}
	identifier, validIdentifier := copyWorldBytes(view.identifier, maxBlockIdentifierBytes)
	properties, validProperties := copyWorldBytes(view.properties_nbt, maxBlockPropertiesBytes)
	if !validIdentifier || !validProperties || len(identifier) == 0 || !utf8.Valid(identifier) ||
		!host.SetWorldBlock(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position), WorldBlock{
			Identifier: string(identifier), PropertiesNBT: properties,
		}) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_world_time_get
func bg_go_world_time_get(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, output *C.int64_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.WorldTime(InvocationID(invocation), WorldID(world.value))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*output = C.int64_t(value)
	return C.DF_STATUS_OK
}

//export bg_go_world_time_set
func bg_go_world_time_set(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, value C.int64_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SetWorldTime(InvocationID(invocation), WorldID(world.value), int64(value)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_world_spawn_get
func bg_go_world_spawn_get(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, output *C.DfBlockPos) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	position, ok := host.WorldSpawn(InvocationID(invocation), WorldID(world.value))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*output = C.DfBlockPos{x: C.int32_t(position.X), y: C.int32_t(position.Y), z: C.int32_t(position.Z)}
	return C.DF_STATUS_OK
}

//export bg_go_world_spawn_set
func bg_go_world_spawn_set(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SetWorldSpawn(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

func nativeBlockPosition(position C.DfBlockPos) BlockPos {
	return BlockPos{X: int32(position.x), Y: int32(position.y), Z: int32(position.z)}
}

func copyWorldString(view C.DfStringView, maximum int) (string, bool) {
	value, ok := copyWorldBytes(view, maximum)
	return string(value), ok && len(value) != 0 && utf8.Valid(value)
}

func copyWorldBytes(view C.DfStringView, maximum int) ([]byte, bool) {
	if uint64(view.len) > uint64(maximum) || (view.len != 0 && view.data == nil) {
		return nil, false
	}
	if view.len == 0 {
		return nil, true
	}
	return C.GoBytes(unsafe.Pointer(view.data), C.int(view.len)), true
}
