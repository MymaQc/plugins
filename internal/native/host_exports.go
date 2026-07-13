package native

/*
#include "bridge.h"
*/
import "C"

import "time"

//export bg_go_player_text
func bg_go_player_text(context C.uint64_t, player C.DfPlayerId, kind C.uint32_t, message C.DfStringView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	var id PlayerID
	for index := range id.UUID {
		id.UUID[index] = byte(player.bytes[index])
	}
	id.Generation = uint64(player.generation)
	if !host.SendPlayerText(id, PlayerTextKind(kind), stringView(message)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_title
func bg_go_player_title(context C.uint64_t, player C.DfPlayerId, value C.DfTitleView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	var id PlayerID
	for index := range id.UUID {
		id.UUID[index] = byte(player.bytes[index])
	}
	id.Generation = uint64(player.generation)
	title := PlayerTitle{
		Text: stringView(value.text), Subtitle: stringView(value.subtitle),
		ActionText: stringView(value.action_text),
		FadeIn:     milliseconds(value.fade_in_milliseconds),
		Duration:   milliseconds(value.duration_milliseconds),
		FadeOut:    milliseconds(value.fade_out_milliseconds),
	}
	if !host.SendPlayerTitle(id, title) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_transform
func bg_go_player_transform(context C.uint64_t, player C.DfPlayerId, kind C.uint32_t, vector C.DfVec3, yaw C.double, pitch C.double) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	id := playerID(player)
	position := Vec3{X: float64(vector.x), Y: float64(vector.y), Z: float64(vector.z)}
	if !host.TransformPlayer(id, PlayerTransformKind(kind), position, float64(yaw), float64(pitch)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_rotation
func bg_go_player_rotation(context C.uint64_t, player C.DfPlayerId, rotation *C.DfRotation) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || rotation == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.PlayerRotation(playerID(player))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	rotation.yaw = C.double(value.Yaw)
	rotation.pitch = C.double(value.Pitch)
	return C.DF_STATUS_OK
}

//export bg_go_player_state_set
func bg_go_player_state_set(context C.uint64_t, player C.DfPlayerId, kind C.uint32_t, value C.DfPlayerStateValue) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SetPlayerState(playerID(player), PlayerStateKind(kind), PlayerStateValue{
		Number: float64(value.number), Integer: int64(value.integer),
	}) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_state_get
func bg_go_player_state_get(context C.uint64_t, player C.DfPlayerId, kind C.uint32_t, value *C.DfPlayerStateValue) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || value == nil {
		return C.DF_STATUS_ERROR
	}
	state, ok := host.PlayerState(playerID(player), PlayerStateKind(kind))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	value.number = C.double(state.Number)
	value.integer = C.int64_t(state.Integer)
	return C.DF_STATUS_OK
}

//export bg_go_player_effect
func bg_go_player_effect(context C.uint64_t, player C.DfPlayerId, operation C.uint32_t, value C.DfEffectView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.ChangePlayerEffect(playerID(player), PlayerEffectOperation(operation), PlayerEffect{
		Type: EffectType(value.effect_type), Level: int32(value.level),
		Duration: milliseconds(value.duration_milliseconds), Ambient: value.ambient != 0,
		Infinite: value.infinite != 0, ParticlesHidden: value.particles_hidden != 0,
	}) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_entity_visibility
func bg_go_player_entity_visibility(context C.uint64_t, player C.DfPlayerId, entity C.DfEntityId, visible C.uint8_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SetPlayerEntityVisible(playerID(player), entityID(entity), visible != 0) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

func playerID(player C.DfPlayerId) PlayerID {
	var id PlayerID
	for index := range id.UUID {
		id.UUID[index] = byte(player.bytes[index])
	}
	id.Generation = uint64(player.generation)
	return id
}

func entityID(entity C.DfEntityId) EntityID {
	var id EntityID
	for index := range id.UUID {
		id.UUID[index] = byte(entity.bytes[index])
	}
	id.Generation = uint64(entity.generation)
	return id
}

func milliseconds(value C.uint64_t) time.Duration {
	const maximum = uint64((1<<63 - 1) / int64(time.Millisecond))
	millis := min(uint64(value), maximum)
	return time.Duration(millis) * time.Millisecond
}
