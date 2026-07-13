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
	maxEntityTextBytes = 4 << 10
	maxEntityTagBytes  = 4 << 10
	maxEntityTypeBytes = 256
	maxWorldEntities   = 1 << 20
)

//export bg_go_world_entity_spawn
func bg_go_world_entity_spawn(context C.uint64_t, invocation C.DfInvocationId, worldID C.DfWorldId, view *C.DfEntitySpawnViewV2, output *C.DfEntityId) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || view == nil || output == nil {
		return C.DF_STATUS_ERROR
	}
	nameTag, validNameTag := copyWorldBytes(view.options.name_tag, maxEntityTagBytes)
	text, validText := copyWorldBytes(view.text, maxEntityTextBytes)
	customType, validCustomType := copyWorldBytes(view.custom_type, maxEntityTypeBytes)
	if !validNameTag || !validText || !validCustomType || !utf8.Valid(nameTag) || !utf8.Valid(text) || !utf8.Valid(customType) ||
		(EntityKind(view.kind) == EntityCustom && len(customType) == 0) || EntityKind(view.kind) > EntityCustom {
		return C.DF_STATUS_ERROR
	}
	value := EntitySpawn{
		Kind: EntityKind(view.kind), Flags: uint32(view.flags), Position: nativeEntityVec3(view.options.position),
		Rotation: Rotation{Yaw: float64(view.options.rotation.yaw), Pitch: float64(view.options.rotation.pitch)},
		Velocity: nativeEntityVec3(view.options.velocity), NameTag: string(nameTag), Text: string(text), Type: string(customType),
		Owner: entityID(view.owner), Damage: float64(view.damage), FuseMilliseconds: uint64(view.fuse_milliseconds),
		Experience: int32(view.experience), Potion: uint32(view.potion), Punch: int32(view.punch_level), Piercing: int32(view.piercing_level),
	}
	if view.item != nil {
		item, ok := copyItemStackView(view.item)
		if !ok {
			return C.DF_STATUS_ERROR
		}
		value.Item = &item
	}
	if view.block != nil {
		identifier, validIdentifier := copyWorldBytes(view.block.identifier, maxBlockIdentifierBytes)
		properties, validProperties := copyWorldBytes(view.block.properties_nbt, maxBlockPropertiesBytes)
		if !validIdentifier || !validProperties || len(identifier) == 0 || !utf8.Valid(identifier) {
			return C.DF_STATUS_ERROR
		}
		block := WorldBlock{Identifier: string(identifier), PropertiesNBT: properties}
		value.Block = &block
	}
	id, ok := host.SpawnWorldEntity(InvocationID(invocation), WorldID(worldID.value), value)
	if !ok || id.Generation == 0 {
		return C.DF_STATUS_ERROR
	}
	*output = cEntityID(id)
	return C.DF_STATUS_OK
}

//export bg_go_world_entities
func bg_go_world_entities(context C.uint64_t, invocation C.DfInvocationId, worldID C.DfWorldId, output *C.DfEntityIdBuffer) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	ids, ok := host.WorldEntities(InvocationID(invocation), WorldID(worldID.value))
	if !ok || len(ids) > maxWorldEntities {
		return C.DF_STATUS_ERROR
	}
	output.len = C.uint64_t(len(ids))
	if uint64(output.capacity) < uint64(len(ids)) || len(ids) != 0 && output.data == nil {
		return C.DF_STATUS_ERROR
	}
	for index, id := range ids {
		unsafe.Slice(output.data, len(ids))[index] = cEntityID(id)
	}
	return C.DF_STATUS_OK
}

//export bg_go_world_players
func bg_go_world_players(context C.uint64_t, invocation C.DfInvocationId, worldID C.DfWorldId, output *C.DfPlayerIdBuffer) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	ids, ok := host.WorldPlayers(InvocationID(invocation), WorldID(worldID.value))
	if !ok || len(ids) > maxWorldEntities {
		return C.DF_STATUS_ERROR
	}
	output.len = C.uint64_t(len(ids))
	if uint64(output.capacity) < uint64(len(ids)) || len(ids) != 0 && output.data == nil {
		return C.DF_STATUS_ERROR
	}
	for index, id := range ids {
		unsafe.Slice(output.data, len(ids))[index] = cPlayerID(id)
	}
	return C.DF_STATUS_OK
}

//export bg_go_entity_state
func bg_go_entity_state(context C.uint64_t, invocation C.DfInvocationId, id C.DfEntityId, output *C.DfEntityState) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	state, ok := host.EntityState(InvocationID(invocation), entityID(id))
	typeBytes, nameTag := []byte(state.Type), []byte(state.NameTag)
	if !ok || len(typeBytes) == 0 || len(typeBytes) > maxEntityTypeBytes || len(nameTag) > maxEntityTagBytes || !utf8.Valid(typeBytes) || !utf8.Valid(nameTag) {
		return C.DF_STATUS_ERROR
	}
	output.entity_type.len = C.uint64_t(len(typeBytes))
	output.name_tag.len = C.uint64_t(len(nameTag))
	if !canWriteSkinBuffer(&output.entity_type, typeBytes) || !canWriteSkinBuffer(&output.name_tag, nameTag) {
		return C.DF_STATUS_ERROR
	}
	output.position = cEntityVec3(state.Position)
	output.rotation = C.DfRotation{yaw: C.double(state.Rotation.Yaw), pitch: C.double(state.Rotation.Pitch)}
	output.velocity = cEntityVec3(state.Velocity)
	output.world = C.DfWorldId{value: C.uint64_t(state.World)}
	output.capabilities = 0
	if state.HasVelocity {
		output.capabilities |= C.DF_ENTITY_HAS_VELOCITY
	}
	if state.HasNameTag {
		output.capabilities |= C.DF_ENTITY_HAS_NAME_TAG
	}
	if state.CanTeleport {
		output.capabilities |= C.DF_ENTITY_CAN_TELEPORT
	}
	writeSkinBuffer(&output.entity_type, typeBytes)
	writeSkinBuffer(&output.name_tag, nameTag)
	return C.DF_STATUS_OK
}

//export bg_go_entity_teleport
func bg_go_entity_teleport(context C.uint64_t, invocation C.DfInvocationId, id C.DfEntityId, position C.DfVec3) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.TeleportEntity(InvocationID(invocation), entityID(id), nativeEntityVec3(position)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_entity_velocity_set
func bg_go_entity_velocity_set(context C.uint64_t, invocation C.DfInvocationId, id C.DfEntityId, velocity C.DfVec3) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SetEntityVelocity(InvocationID(invocation), entityID(id), nativeEntityVec3(velocity)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_entity_name_tag_set
func bg_go_entity_name_tag_set(context C.uint64_t, invocation C.DfInvocationId, id C.DfEntityId, nameTag C.DfStringView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	value, valid := copyWorldBytes(nameTag, maxEntityTagBytes)
	if !ok || !valid || !utf8.Valid(value) || !host.SetEntityNameTag(InvocationID(invocation), entityID(id), string(value)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_entity_despawn
func bg_go_entity_despawn(context C.uint64_t, invocation C.DfInvocationId, id C.DfEntityId) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.DespawnEntity(InvocationID(invocation), entityID(id)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

func cEntityID(id EntityID) C.DfEntityId {
	var value C.DfEntityId
	for index := range id.UUID {
		value.bytes[index] = C.uint8_t(id.UUID[index])
	}
	value.generation = C.uint64_t(id.Generation)
	return value
}

func nativeEntityVec3(value C.DfVec3) Vec3 {
	return Vec3{X: float64(value.x), Y: float64(value.y), Z: float64(value.z)}
}

func cEntityVec3(value Vec3) C.DfVec3 {
	return C.DfVec3{x: C.double(value.X), y: C.double(value.Y), z: C.double(value.Z)}
}
