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

func playerID(player C.DfPlayerId) PlayerID {
	var id PlayerID
	for index := range id.UUID {
		id.UUID[index] = byte(player.bytes[index])
	}
	id.Generation = uint64(player.generation)
	return id
}

func milliseconds(value C.uint64_t) time.Duration {
	const maximum = uint64((1<<63 - 1) / int64(time.Millisecond))
	millis := min(uint64(value), maximum)
	return time.Duration(millis) * time.Millisecond
}
