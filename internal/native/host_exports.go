package native

/*
#include "bridge.h"
*/
import "C"

import (
	"time"
	"unsafe"
)

const (
	maxSkinDataBytes  = 64 << 20
	maxSkinAnimations = 256
)

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

//export bg_go_player_skin_open
func bg_go_player_skin_open(context C.uint64_t, player C.DfPlayerId, snapshot *C.uint64_t, info *C.DfSkinInfo) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || snapshot == nil || info == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.PlayerSkin(playerID(player))
	if !ok || !validPlayerSkinPayload(value) {
		return C.DF_STATUS_ERROR
	}
	snapshotID, ok := registerSkinSnapshot(uint64(context), value)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*snapshot = C.uint64_t(snapshotID)
	fillSkinInfo(info, value)
	return C.DF_STATUS_OK
}

//export bg_go_player_skin_animation_info
func bg_go_player_skin_animation_info(context C.uint64_t, snapshot C.uint64_t, index C.uint64_t, info *C.DfSkinAnimationInfo) C.DfStatus {
	value, ok := resolveSkinSnapshot(uint64(context), uint64(snapshot))
	if !ok || info == nil || uint64(index) >= uint64(len(value.Animations)) {
		return C.DF_STATUS_ERROR
	}
	animation := value.Animations[uint64(index)]
	*info = C.DfSkinAnimationInfo{
		width: C.uint32_t(animation.Width), height: C.uint32_t(animation.Height),
		animation_type: C.uint32_t(animation.Type), frame_count: C.int64_t(animation.FrameCount),
		expression: C.int64_t(animation.Expression), pixels_len: C.uint64_t(len(animation.Pixels)),
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_skin_read
func bg_go_player_skin_read(context C.uint64_t, snapshot C.uint64_t, data *C.DfSkinData) C.DfStatus {
	value, ok := resolveSkinSnapshot(uint64(context), uint64(snapshot))
	if !ok || data == nil || len(value.Animations) > maxSkinAnimations || uint64(data.animation_capacity) < uint64(len(value.Animations)) {
		return C.DF_STATUS_ERROR
	}
	if !canWriteSkinBuffer(&data.play_fab_id, []byte(value.PlayFabID)) ||
		!canWriteSkinBuffer(&data.full_id, []byte(value.FullID)) ||
		!canWriteSkinBuffer(&data.pixels, value.Pixels) ||
		!canWriteSkinBuffer(&data.model_default, []byte(value.ModelDefault)) ||
		!canWriteSkinBuffer(&data.model_animated_face, []byte(value.ModelAnimatedFace)) ||
		!canWriteSkinBuffer(&data.model, value.Model) ||
		!canWriteSkinBuffer(&data.cape_pixels, value.CapePixels) {
		return C.DF_STATUS_ERROR
	}
	if len(value.Animations) != 0 && data.animation_pixels == nil {
		return C.DF_STATUS_ERROR
	}
	var buffers []C.DfStringBuffer
	if len(value.Animations) != 0 {
		buffers = unsafe.Slice(data.animation_pixels, len(value.Animations))
	}
	for index, animation := range value.Animations {
		if !canWriteSkinBuffer(&buffers[index], animation.Pixels) {
			return C.DF_STATUS_ERROR
		}
	}
	writeSkinBuffer(&data.play_fab_id, []byte(value.PlayFabID))
	writeSkinBuffer(&data.full_id, []byte(value.FullID))
	writeSkinBuffer(&data.pixels, value.Pixels)
	writeSkinBuffer(&data.model_default, []byte(value.ModelDefault))
	writeSkinBuffer(&data.model_animated_face, []byte(value.ModelAnimatedFace))
	writeSkinBuffer(&data.model, value.Model)
	writeSkinBuffer(&data.cape_pixels, value.CapePixels)
	for index, animation := range value.Animations {
		writeSkinBuffer(&buffers[index], animation.Pixels)
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_skin_close
func bg_go_player_skin_close(context C.uint64_t, snapshot C.uint64_t) {
	unregisterSkinSnapshot(uint64(context), uint64(snapshot))
}

//export bg_go_player_skin_set
func bg_go_player_skin_set(context C.uint64_t, player C.DfPlayerId, view *C.DfSkinView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || view == nil || !validSkinViewPayload(view) {
		return C.DF_STATUS_ERROR
	}
	playFabID, ok := copySkinView(view.play_fab_id)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	fullID, ok := copySkinView(view.full_id)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	pixels, ok := copySkinView(view.pixels)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	modelDefault, ok := copySkinView(view.model_default)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	modelAnimatedFace, ok := copySkinView(view.model_animated_face)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	model, ok := copySkinView(view.model)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	capePixels, ok := copySkinView(view.cape_pixels)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	value := PlayerSkin{
		Width: uint32(view.width), Height: uint32(view.height), Persona: view.persona != 0,
		PlayFabID: string(playFabID), FullID: string(fullID), Pixels: pixels,
		ModelDefault: string(modelDefault), ModelAnimatedFace: string(modelAnimatedFace), Model: model,
		CapeWidth: uint32(view.cape_width), CapeHeight: uint32(view.cape_height), CapePixels: capePixels,
	}
	if view.animation_count != 0 {
		if view.animations == nil {
			return C.DF_STATUS_ERROR
		}
		animations := unsafe.Slice(view.animations, int(view.animation_count))
		value.Animations = make([]SkinAnimation, len(animations))
		for index, animation := range animations {
			animationPixels, ok := copySkinView(animation.pixels)
			if !ok {
				return C.DF_STATUS_ERROR
			}
			value.Animations[index] = SkinAnimation{
				Width: uint32(animation.width), Height: uint32(animation.height), Type: uint32(animation.animation_type),
				FrameCount: int64(animation.frame_count), Expression: int64(animation.expression), Pixels: animationPixels,
			}
		}
	}
	if !host.SetPlayerSkin(playerID(player), value) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

func fillSkinInfo(info *C.DfSkinInfo, value PlayerSkin) {
	*info = C.DfSkinInfo{
		width: C.uint32_t(value.Width), height: C.uint32_t(value.Height), persona: C.uint8_t(boolByte(value.Persona)),
		play_fab_id_len: C.uint64_t(len(value.PlayFabID)), full_id_len: C.uint64_t(len(value.FullID)),
		pixels_len: C.uint64_t(len(value.Pixels)), model_default_len: C.uint64_t(len(value.ModelDefault)),
		model_animated_face_len: C.uint64_t(len(value.ModelAnimatedFace)), model_len: C.uint64_t(len(value.Model)),
		cape_width: C.uint32_t(value.CapeWidth), cape_height: C.uint32_t(value.CapeHeight),
		cape_pixels_len: C.uint64_t(len(value.CapePixels)), animation_count: C.uint64_t(len(value.Animations)),
	}
}

func canWriteSkinBuffer(buffer *C.DfStringBuffer, source []byte) bool {
	return len(source) <= maxSkinDataBytes &&
		uint64(buffer.capacity) >= uint64(len(source)) &&
		(len(source) == 0 || buffer.data != nil)
}

func writeSkinBuffer(buffer *C.DfStringBuffer, source []byte) bool {
	if !canWriteSkinBuffer(buffer, source) {
		return false
	}
	if len(source) != 0 {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(buffer.data)), len(source)), source)
	}
	buffer.len = C.uint64_t(len(source))
	return true
}

func copySkinView(view C.DfStringView) ([]byte, bool) {
	length := uint64(view.len)
	if length > maxSkinDataBytes || length != 0 && view.data == nil {
		return nil, false
	}
	if length == 0 {
		return nil, true
	}
	return append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(view.data)), int(length))...), true
}

func validPlayerSkinPayload(value PlayerSkin) bool {
	if len(value.Animations) > maxSkinAnimations {
		return false
	}
	total := uint64(0)
	for _, length := range []int{
		len(value.PlayFabID), len(value.FullID), len(value.Pixels), len(value.ModelDefault),
		len(value.ModelAnimatedFace), len(value.Model), len(value.CapePixels),
	} {
		if uint64(length) > uint64(maxSkinDataBytes)-total {
			return false
		}
		total += uint64(length)
	}
	for _, animation := range value.Animations {
		if uint64(len(animation.Pixels)) > uint64(maxSkinDataBytes)-total {
			return false
		}
		total += uint64(len(animation.Pixels))
	}
	return true
}

func validSkinViewPayload(view *C.DfSkinView) bool {
	if uint64(view.animation_count) > maxSkinAnimations || view.animation_count != 0 && view.animations == nil {
		return false
	}
	total := uint64(0)
	views := [...]C.DfStringView{
		view.play_fab_id, view.full_id, view.pixels, view.model_default,
		view.model_animated_face, view.model, view.cape_pixels,
	}
	for _, field := range views {
		if !addSkinViewLength(&total, field) {
			return false
		}
	}
	if view.animation_count == 0 {
		return true
	}
	for _, animation := range unsafe.Slice(view.animations, int(view.animation_count)) {
		if !addSkinViewLength(&total, animation.pixels) {
			return false
		}
	}
	return true
}

func addSkinViewLength(total *uint64, view C.DfStringView) bool {
	length := uint64(view.len)
	if length != 0 && view.data == nil || length > uint64(maxSkinDataBytes)-*total {
		return false
	}
	*total += length
	return true
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
