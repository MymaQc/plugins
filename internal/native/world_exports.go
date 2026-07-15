package native

/*
#include "bridge.h"
*/
import "C"

import (
	"time"
	"unicode/utf8"
	"unsafe"
)

const (
	maxBlockIdentifierBytes = 256
	maxBlockPropertiesBytes = 64 << 10
	maxBlocksWithinTargets  = 1 << 16
	maxSourceNameBytes      = 64 << 10
	setBlockDisableUpdates  = 1
	setBlockDisableLiquid   = 2
	setBlockDisableRedstone = 4
)

//export bg_go_world_name
func bg_go_world_name(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, output *C.DfStringBuffer) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	name, ok := host.WorldName(InvocationID(invocation), WorldID(world.value))
	if !ok || name == "" || !utf8.ValidString(name) {
		return C.DF_STATUS_ERROR
	}
	if uint64(output.capacity) < uint64(len(name)) || len(name) != 0 && output.data == nil {
		output.len = C.uint64_t(len(name))
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
	return writeWorldBlock(output, block, ok)
}

//export bg_go_block_by_name
func bg_go_block_by_name(context C.uint64_t, name C.DfStringView, properties C.DfStringView, found *C.uint8_t, output *C.DfBlockData) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || found == nil || output == nil {
		return C.DF_STATUS_ERROR
	}
	*found = 0
	output.identifier.len = 0
	output.properties_nbt.len = 0
	identifier, validIdentifier := copyWorldBytes(name, maxBlockIdentifierBytes)
	encodedProperties, validProperties := copyWorldBytes(properties, maxBlockPropertiesBytes)
	if !validIdentifier || !validProperties || len(identifier) == 0 || !utf8.Valid(identifier) {
		return C.DF_STATUS_ERROR
	}
	block, ok := host.BlockByName(string(identifier), encodedProperties)
	if !ok {
		return C.DF_STATUS_OK
	}
	*found = 1
	return writeWorldBlock(output, block, true)
}

//export bg_go_world_block_loaded
func bg_go_world_block_loaded(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, loaded *C.uint8_t, output *C.DfBlockData) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || loaded == nil || output == nil {
		return C.DF_STATUS_ERROR
	}
	*loaded = 0
	output.identifier.len = 0
	output.properties_nbt.len = 0
	block, found, valid := host.WorldBlockLoaded(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position))
	if !valid {
		return C.DF_STATUS_ERROR
	}
	if !found {
		return C.DF_STATUS_OK
	}
	*loaded = 1
	return writeWorldBlock(output, block, true)
}

//export bg_go_world_range
func bg_go_world_range(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, output *C.DfBlockRange) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.WorldRange(InvocationID(invocation), WorldID(world.value))
	if !ok || value.Min > value.Max {
		return C.DF_STATUS_ERROR
	}
	output.min = C.int32_t(value.Min)
	output.max = C.int32_t(value.Max)
	return C.DF_STATUS_OK
}

//export bg_go_world_blocks_within_open
func bg_go_world_blocks_within_open(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, radius C.int32_t, blocks *C.DfBlockView, blockCount C.uint64_t, output *C.DfBlockIteratorId) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if output == nil {
		return C.DF_STATUS_ERROR
	}
	*output = 0
	count := uint64(blockCount)
	if !ok || count > maxBlocksWithinTargets || count != 0 && blocks == nil {
		return C.DF_STATUS_ERROR
	}
	values := make([]WorldBlock, int(count))
	for index, view := range unsafe.Slice(blocks, int(count)) {
		identifier, validIdentifier := copyWorldBytes(view.identifier, maxBlockIdentifierBytes)
		properties, validProperties := copyWorldBytes(view.properties_nbt, maxBlockPropertiesBytes)
		if !validIdentifier || !validProperties || len(identifier) == 0 || !utf8.Valid(identifier) {
			return C.DF_STATUS_ERROR
		}
		values[index] = WorldBlock{Identifier: string(identifier), PropertiesNBT: properties}
	}
	id, ok := host.OpenWorldBlocksWithin(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position), int32(radius), values)
	if !ok || id == 0 {
		return C.DF_STATUS_ERROR
	}
	*output = C.DfBlockIteratorId(id)
	return C.DF_STATUS_OK
}

//export bg_go_world_blocks_within_next
func bg_go_world_blocks_within_next(context C.uint64_t, invocation C.DfInvocationId, iterator C.DfBlockIteratorId, position *C.DfBlockPos, found *C.uint8_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || iterator == 0 || position == nil || found == nil {
		return C.DF_STATUS_ERROR
	}
	*position = C.DfBlockPos{}
	*found = 0
	value, hasNext, valid := host.NextWorldBlock(InvocationID(invocation), BlockIteratorID(iterator))
	if !valid {
		return C.DF_STATUS_ERROR
	}
	if !hasNext {
		return C.DF_STATUS_OK
	}
	*position = C.DfBlockPos{x: C.int32_t(value.X), y: C.int32_t(value.Y), z: C.int32_t(value.Z)}
	*found = 1
	return C.DF_STATUS_OK
}

//export bg_go_world_blocks_within_close
func bg_go_world_blocks_within_close(context C.uint64_t, invocation C.DfInvocationId, iterator C.DfBlockIteratorId) {
	host, ok := resolveHost(uint64(context))
	if ok && iterator != 0 {
		host.CloseWorldBlocks(InvocationID(invocation), BlockIteratorID(iterator))
	}
}

//export bg_go_world_highest_light_blocker
func bg_go_world_highest_light_blocker(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, x, z C.int32_t, output *C.int32_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.WorldHighestLightBlocker(InvocationID(invocation), WorldID(world.value), int32(x), int32(z))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*output = C.int32_t(value)
	return C.DF_STATUS_OK
}

//export bg_go_world_highest_block
func bg_go_world_highest_block(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, x, z C.int32_t, output *C.int32_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.WorldHighestBlock(InvocationID(invocation), WorldID(world.value), int32(x), int32(z))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*output = C.int32_t(value)
	return C.DF_STATUS_OK
}

//export bg_go_world_light
func bg_go_world_light(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, output *C.uint8_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.WorldLight(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position))
	if !ok || value > 15 {
		return C.DF_STATUS_ERROR
	}
	*output = C.uint8_t(value)
	return C.DF_STATUS_OK
}

//export bg_go_world_sky_light
func bg_go_world_sky_light(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, output *C.uint8_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.WorldSkyLight(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position))
	if !ok || value > 15 {
		return C.DF_STATUS_ERROR
	}
	*output = C.uint8_t(value)
	return C.DF_STATUS_OK
}

//export bg_go_world_liquid_get
func bg_go_world_liquid_get(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, found *C.uint8_t, output *C.DfBlockData) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || found == nil || output == nil {
		return C.DF_STATUS_ERROR
	}
	*found = 0
	output.identifier.len = 0
	output.properties_nbt.len = 0
	block, present, valid := host.WorldLiquid(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position))
	if !valid {
		return C.DF_STATUS_ERROR
	}
	if !present {
		return C.DF_STATUS_OK
	}
	*found = 1
	return writeWorldBlock(output, block, true)
}

//export bg_go_world_liquid_set
func bg_go_world_liquid_set(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, view *C.DfBlockView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	var value *WorldBlock
	if view != nil {
		identifier, validIdentifier := copyWorldBytes(view.identifier, maxBlockIdentifierBytes)
		properties, validProperties := copyWorldBytes(view.properties_nbt, maxBlockPropertiesBytes)
		if !validIdentifier || !validProperties || len(identifier) == 0 || !utf8.Valid(identifier) {
			return C.DF_STATUS_ERROR
		}
		value = &WorldBlock{Identifier: string(identifier), PropertiesNBT: properties}
	}
	if !host.SetWorldLiquid(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position), value) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_world_block_update_schedule
func bg_go_world_block_update_schedule(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, view *C.DfBlockView, delayNanoseconds C.int64_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || view == nil {
		return C.DF_STATUS_ERROR
	}
	identifier, validIdentifier := copyWorldBytes(view.identifier, maxBlockIdentifierBytes)
	properties, validProperties := copyWorldBytes(view.properties_nbt, maxBlockPropertiesBytes)
	if !validIdentifier || !validProperties || len(identifier) == 0 || !utf8.Valid(identifier) ||
		!host.ScheduleWorldBlockUpdate(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position), WorldBlock{
			Identifier: string(identifier), PropertiesNBT: properties,
		}, int64(delayNanoseconds)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_world_biome_get
func bg_go_world_biome_get(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, output *C.int32_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.WorldBiome(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*output = C.int32_t(value)
	return C.DF_STATUS_OK
}

//export bg_go_world_biome_set
func bg_go_world_biome_set(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, biome C.int32_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SetWorldBiome(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position), int32(biome)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_world_temperature
func bg_go_world_temperature(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, output *C.double) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.WorldTemperature(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*output = C.double(value)
	return C.DF_STATUS_OK
}

//export bg_go_world_raining_at
func bg_go_world_raining_at(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, output *C.uint8_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, valid := host.WorldRainingAt(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position))
	return writeWorldBool(output, value, valid)
}

//export bg_go_world_snowing_at
func bg_go_world_snowing_at(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, output *C.uint8_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, valid := host.WorldSnowingAt(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position))
	return writeWorldBool(output, value, valid)
}

//export bg_go_world_thundering_at
func bg_go_world_thundering_at(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, output *C.uint8_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, valid := host.WorldThunderingAt(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position))
	return writeWorldBool(output, value, valid)
}

//export bg_go_world_raining
func bg_go_world_raining(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, output *C.uint8_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, valid := host.WorldRaining(InvocationID(invocation), WorldID(world.value))
	return writeWorldBool(output, value, valid)
}

//export bg_go_world_thundering
func bg_go_world_thundering(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, output *C.uint8_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, valid := host.WorldThundering(InvocationID(invocation), WorldID(world.value))
	return writeWorldBool(output, value, valid)
}

//export bg_go_world_current_tick
func bg_go_world_current_tick(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, output *C.int64_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.WorldCurrentTick(InvocationID(invocation), WorldID(world.value))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*output = C.int64_t(value)
	return C.DF_STATUS_OK
}

func writeWorldBool(output *C.uint8_t, value, valid bool) C.DfStatus {
	if !valid {
		return C.DF_STATUS_ERROR
	}
	if value {
		*output = 1
	} else {
		*output = 0
	}
	return C.DF_STATUS_OK
}

func writeWorldBlock(output *C.DfBlockData, block WorldBlock, ok bool) C.DfStatus {
	identifier, properties, valid := worldBlockPayload(block, ok)
	if !valid {
		return C.DF_STATUS_ERROR
	}
	output.identifier.len = C.uint64_t(len(identifier))
	output.properties_nbt.len = C.uint64_t(len(properties))
	if !worldBlockFits(identifier, properties, uint64(output.identifier.capacity), uint64(output.properties_nbt.capacity)) ||
		!canWriteSkinBuffer(&output.identifier, identifier) || !canWriteSkinBuffer(&output.properties_nbt, properties) {
		return C.DF_STATUS_ERROR
	}
	writeSkinBuffer(&output.identifier, identifier)
	writeSkinBuffer(&output.properties_nbt, properties)
	return C.DF_STATUS_OK
}

func worldBlockPayload(block WorldBlock, ok bool) ([]byte, []byte, bool) {
	identifier, properties := []byte(block.Identifier), block.PropertiesNBT
	if !ok || len(identifier) == 0 || len(identifier) > maxBlockIdentifierBytes || !utf8.Valid(identifier) || len(properties) > maxBlockPropertiesBytes {
		return nil, nil, false
	}
	return identifier, properties, true
}

func worldBlockFits(identifier, properties []byte, identifierCapacity, propertiesCapacity uint64) bool {
	return uint64(len(identifier)) <= identifierCapacity && uint64(len(properties)) <= propertiesCapacity
}

//export bg_go_world_block_set
func bg_go_world_block_set(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, position C.DfBlockPos, view *C.DfBlockView, flags C.uint32_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	options, validOptions := worldSetOptions(uint32(flags))
	if !ok || view == nil || !validOptions {
		return C.DF_STATUS_ERROR
	}
	identifier, validIdentifier := copyWorldBytes(view.identifier, maxBlockIdentifierBytes)
	properties, validProperties := copyWorldBytes(view.properties_nbt, maxBlockPropertiesBytes)
	if !validIdentifier || !validProperties || len(identifier) == 0 || !utf8.Valid(identifier) ||
		!host.SetWorldBlock(InvocationID(invocation), WorldID(world.value), nativeBlockPosition(position), WorldBlock{
			Identifier: string(identifier), PropertiesNBT: properties,
		}, options) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

func worldSetOptions(flags uint32) (WorldSetOpts, bool) {
	if flags & ^uint32(setBlockDisableUpdates|setBlockDisableLiquid|setBlockDisableRedstone) != 0 {
		return WorldSetOpts{}, false
	}
	return WorldSetOpts{
		DisableBlockUpdates:       flags&setBlockDisableUpdates != 0,
		DisableLiquidDisplacement: flags&setBlockDisableLiquid != 0,
		DisableRedstoneUpdates:    flags&setBlockDisableRedstone != 0,
	}, true
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

//export bg_go_world_player_spawn_get
func bg_go_world_player_spawn_get(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, player C.DfUuid, output *C.DfBlockPos) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	position, ok := host.WorldPlayerSpawn(InvocationID(invocation), WorldID(world.value), uuidValue(player))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*output = C.DfBlockPos{x: C.int32_t(position.X), y: C.int32_t(position.Y), z: C.int32_t(position.Z)}
	return C.DF_STATUS_OK
}

//export bg_go_world_player_spawn_set
func bg_go_world_player_spawn_set(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, player C.DfUuid, position C.DfBlockPos) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SetWorldPlayerSpawn(InvocationID(invocation), WorldID(world.value), uuidValue(player), nativeBlockPosition(position)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

func uuidValue(value C.DfUuid) [16]byte {
	var id [16]byte
	for index := range id {
		id[index] = byte(value.bytes[index])
	}
	return id
}

//export bg_go_world_dimension_get
func bg_go_world_dimension_get(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, output *C.DfDimensionView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	dimension, ok := host.WorldDimension(InvocationID(invocation), WorldID(world.value))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*output = C.DfDimensionView{id: C.uint32_t(dimension.ID)}
	if dimension.Custom != nil {
		output.custom = 1
		output.range_min, output.range_max = C.int32_t(dimension.Custom.Range.Min), C.int32_t(dimension.Custom.Range.Max)
		output.lava_spread_nanoseconds = C.int64_t(dimension.Custom.LavaSpreadDuration)
		if dimension.Custom.WaterEvaporates {
			output.water_evaporates = 1
		}
		if dimension.Custom.WeatherCycle {
			output.weather_cycle = 1
		}
		if dimension.Custom.TimeCycle {
			output.time_cycle = 1
		}
	} else if dimension.ID > WorldDimensionEnd {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_world_time_cycle_get
func bg_go_world_time_cycle_get(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, output *C.uint8_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.WorldTimeCycle(InvocationID(invocation), WorldID(world.value))
	return writeWorldBool(output, value, ok)
}

//export bg_go_world_time_cycle_set
func bg_go_world_time_cycle_set(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, value C.uint8_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || value > 1 || !host.SetWorldTimeCycle(InvocationID(invocation), WorldID(world.value), value != 0) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_world_required_sleep_duration_set
func bg_go_world_required_sleep_duration_set(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, value C.int64_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SetWorldRequiredSleepDuration(InvocationID(invocation), WorldID(world.value), time.Duration(value)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_world_default_game_mode_get
func bg_go_world_default_game_mode_get(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, output *C.int64_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.WorldDefaultGameMode(InvocationID(invocation), WorldID(world.value))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*output = C.int64_t(value)
	return C.DF_STATUS_OK
}

//export bg_go_world_default_game_mode_set
func bg_go_world_default_game_mode_set(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, value C.int64_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SetWorldDefaultGameMode(InvocationID(invocation), WorldID(world.value), int64(value)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_world_tick_range_set
func bg_go_world_tick_range_set(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, value C.int32_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SetWorldTickRange(InvocationID(invocation), WorldID(world.value), int32(value)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_world_difficulty_get
func bg_go_world_difficulty_get(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, output *C.DfDifficultyView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.WorldDifficulty(InvocationID(invocation), WorldID(world.value))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*output = C.DfDifficultyView{
		id: C.uint32_t(value.ID), builtin: worldBool(value.Builtin), food_regenerates: worldBool(value.FoodRegenerates),
		starvation_health_limit: C.double(value.StarvationHealthLimit), fire_spread_increase: C.int32_t(value.FireSpreadIncrease),
	}
	return C.DF_STATUS_OK
}

//export bg_go_world_difficulty_set
func bg_go_world_difficulty_set(context C.uint64_t, invocation C.DfInvocationId, world C.DfWorldId, value C.DfDifficultyView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || value.builtin > 1 || value.food_regenerates > 1 || value.reserved != 0 || value.reserved2 != 0 {
		return C.DF_STATUS_ERROR
	}
	view := DifficultyView{
		ID: uint32(value.id), Builtin: value.builtin != 0, FoodRegenerates: value.food_regenerates != 0,
		StarvationHealthLimit: float64(value.starvation_health_limit), FireSpreadIncrease: int32(value.fire_spread_increase),
	}
	if !host.SetWorldDifficulty(InvocationID(invocation), WorldID(world.value), view) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

func worldBool(value bool) C.uint8_t {
	if value {
		return 1
	}
	return 0
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
