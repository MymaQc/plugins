package native

/*
#include "bridge.h"
*/
import "C"

import (
	"encoding/json"
	"time"
	"unicode/utf8"
	"unsafe"
)

const (
	maxSkinDataBytes   = 64 << 20
	maxSkinAnimations  = 256
	maxScoreboardLines = 15
	maxFormJSONBytes   = 1 << 20
)

//export bg_go_player_form_send
func bg_go_player_form_send(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, view *C.DfFormView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || view == nil || view.response == nil || view.drop == nil || view.request_json.len == 0 || view.request_json.len > maxFormJSONBytes || view.request_json.data == nil {
		return C.DF_STATUS_ERROR
	}
	request := append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(view.request_json.data)), int(view.request_json.len))...)
	if !json.Valid(request) {
		return C.DF_STATUS_ERROR
	}
	callbackContext, responseCallback, dropCallback := view.callback_context, view.response, view.drop
	id, ok := registerForm(uint64(context), playerID(player), func(responseInvocation InvocationID, submitter PlayerID, closed bool, response []byte) bool {
		outcome := C.uint32_t(C.DF_FORM_RESPONSE_SUBMITTED)
		if closed {
			outcome = C.uint32_t(C.DF_FORM_RESPONSE_CLOSED)
		}
		var responseView C.DfStringView
		if len(response) != 0 {
			responseView.data = (*C.uint8_t)(unsafe.Pointer(&response[0]))
			responseView.len = C.uint64_t(len(response))
		}
		return C.bg_call_form_response(responseCallback, callbackContext, C.DfInvocationId(responseInvocation), cPlayerID(submitter), outcome, responseView) == C.DF_STATUS_OK
	}, func() { C.bg_call_form_drop(dropCallback, callbackContext) })
	if !ok {
		return C.DF_STATUS_ERROR
	}
	if !host.SendPlayerForm(InvocationID(invocation), playerID(player), PlayerForm{ID: id, RequestJSON: request}) {
		abandonForm(id)
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_form_close
func bg_go_player_form_close(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.ClosePlayerForm(InvocationID(invocation), playerID(player)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

func cPlayerID(id PlayerID) C.DfPlayerId {
	var value C.DfPlayerId
	for index := range id.UUID {
		value.bytes[index] = C.uint8_t(id.UUID[index])
	}
	value.generation = C.uint64_t(id.Generation)
	return value
}

//export bg_go_player_text
func bg_go_player_text(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, kind C.uint32_t, message C.DfStringView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	var id PlayerID
	for index := range id.UUID {
		id.UUID[index] = byte(player.bytes[index])
	}
	id.Generation = uint64(player.generation)
	if !host.SendPlayerText(InvocationID(invocation), id, PlayerTextKind(kind), stringView(message)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_scoreboard
func bg_go_player_scoreboard(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, view C.DfScoreboardView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || view.line_count > maxScoreboardLines || (view.line_count != 0 && view.lines == nil) {
		return C.DF_STATUS_ERROR
	}
	lineViews := unsafe.Slice(view.lines, int(view.line_count))
	lines := make([]string, len(lineViews))
	for index, line := range lineViews {
		lines[index] = stringView(line)
	}
	if !host.SendPlayerScoreboard(InvocationID(invocation), playerID(player), PlayerScoreboard{
		Name: stringView(view.name), Lines: lines, Padding: view.padding != 0, Descending: view.descending != 0,
	}) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_scoreboard_remove
func bg_go_player_scoreboard_remove(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.RemovePlayerScoreboard(InvocationID(invocation), playerID(player)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_title
func bg_go_player_title(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, value C.DfTitleView) C.DfStatus {
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
	if !host.SendPlayerTitle(InvocationID(invocation), id, title) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_transform
func bg_go_player_transform(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, kind C.uint32_t, vector C.DfVec3, yaw C.double, pitch C.double) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	id := playerID(player)
	position := Vec3{X: float64(vector.x), Y: float64(vector.y), Z: float64(vector.z)}
	if !host.TransformPlayer(InvocationID(invocation), id, PlayerTransformKind(kind), position, float64(yaw), float64(pitch)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_rotation
func bg_go_player_rotation(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, rotation *C.DfRotation) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || rotation == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.PlayerRotation(InvocationID(invocation), playerID(player))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	rotation.yaw = C.double(value.Yaw)
	rotation.pitch = C.double(value.Pitch)
	return C.DF_STATUS_OK
}

//export bg_go_player_state_set
func bg_go_player_state_set(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, kind C.uint32_t, value C.DfPlayerStateValue) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SetPlayerState(InvocationID(invocation), playerID(player), PlayerStateKind(kind), PlayerStateValue{
		Number: float64(value.number), Integer: int64(value.integer),
	}) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_state_get
func bg_go_player_state_get(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, kind C.uint32_t, value *C.DfPlayerStateValue) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || value == nil {
		return C.DF_STATUS_ERROR
	}
	state, ok := host.PlayerState(InvocationID(invocation), playerID(player), PlayerStateKind(kind))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	value.number = C.double(state.Number)
	value.integer = C.int64_t(state.Integer)
	return C.DF_STATUS_OK
}

//export bg_go_player_heal
func bg_go_player_heal(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, health C.double, view *C.DfHealingSourceView, result *C.DfPlayerHealResult) C.DfStatus {
	if result == nil {
		return C.DF_STATUS_ERROR
	}
	*result = C.DfPlayerHealResult{}
	host, ok := resolveHost(uint64(context))
	source, valid := copyHealingSourceView(view)
	if !ok || !valid {
		return C.DF_STATUS_ERROR
	}
	healed, ok := host.HealPlayer(InvocationID(invocation), playerID(player), float64(health), source)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	result.healed = C.double(healed)
	return C.DF_STATUS_OK
}

//export bg_go_player_hurt
func bg_go_player_hurt(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, damage C.double, view *C.DfDamageSourceView, result *C.DfPlayerHurtResult) C.DfStatus {
	if result == nil {
		return C.DF_STATUS_ERROR
	}
	*result = C.DfPlayerHurtResult{}
	host, ok := resolveHost(uint64(context))
	source, valid := copyDamageSourceView(view)
	if !ok || !valid {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.HurtPlayer(InvocationID(invocation), playerID(player), float64(damage), source)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	result.damage = C.double(value.Damage)
	if value.Vulnerable {
		result.vulnerable = 1
	}
	return C.DF_STATUS_OK
}

func copyHealingSourceView(view *C.DfHealingSourceView) (HealingSource, bool) {
	if view == nil || uint32(view.kind) > uint32(HealingSourceRegeneration) || view.data > 1 {
		return HealingSource{}, false
	}
	kind := HealingSourceKind(view.kind)
	if view.data != 0 && kind != HealingSourceFood {
		return HealingSource{}, false
	}
	name, ok := copyWorldBytes(view.name, maxSourceNameBytes)
	if !ok || !utf8.Valid(name) {
		return HealingSource{}, false
	}
	return HealingSource{Name: string(name), Kind: kind, Data: view.data != 0}, true
}

func copyDamageSourceView(view *C.DfDamageSourceView) (DamageSource, bool) {
	if view == nil || uint32(view.kind) > uint32(DamageSourceWither) || view.data > 1 {
		return DamageSource{}, false
	}
	kind := DamageSourceKind(view.kind)
	if view.data != 0 && kind != DamageSourcePoison {
		return DamageSource{}, false
	}
	if (view.block != nil) != (kind == DamageSourceBlock) {
		return DamageSource{}, false
	}
	name, ok := copyWorldBytes(view.name, maxSourceNameBytes)
	if !ok || !utf8.Valid(name) {
		return DamageSource{}, false
	}
	flags := uint32(view.flags)
	knownFlags := uint32(C.DF_DAMAGE_SOURCE_REDUCED_BY_ARMOUR | C.DF_DAMAGE_SOURCE_REDUCED_BY_RESISTANCE |
		C.DF_DAMAGE_SOURCE_FIRE | C.DF_DAMAGE_SOURCE_IGNORES_TOTEM | C.DF_DAMAGE_SOURCE_FIRE_PROTECTION |
		C.DF_DAMAGE_SOURCE_FEATHER_FALLING | C.DF_DAMAGE_SOURCE_BLAST_PROTECTION | C.DF_DAMAGE_SOURCE_PROJECTILE_PROTECTION)
	if flags & ^knownFlags != 0 {
		return DamageSource{}, false
	}
	source := DamageSource{
		Name: string(name), Kind: kind, Entity: entityID(view.entity),
		SecondaryEntity: entityID(view.secondary_entity), Data: view.data != 0,
		ReducedByArmour:     flags&C.DF_DAMAGE_SOURCE_REDUCED_BY_ARMOUR != 0,
		ReducedByResistance: flags&C.DF_DAMAGE_SOURCE_REDUCED_BY_RESISTANCE != 0,
		Fire:                flags&C.DF_DAMAGE_SOURCE_FIRE != 0, IgnoresTotem: flags&C.DF_DAMAGE_SOURCE_IGNORES_TOTEM != 0,
		FireProtection:       flags&C.DF_DAMAGE_SOURCE_FIRE_PROTECTION != 0,
		FeatherFalling:       flags&C.DF_DAMAGE_SOURCE_FEATHER_FALLING != 0,
		BlastProtection:      flags&C.DF_DAMAGE_SOURCE_BLAST_PROTECTION != 0,
		ProjectileProtection: flags&C.DF_DAMAGE_SOURCE_PROJECTILE_PROTECTION != 0,
	}
	if view.block != nil {
		identifier, validIdentifier := copyWorldBytes(view.block.identifier, maxBlockIdentifierBytes)
		properties, validProperties := copyWorldBytes(view.block.properties_nbt, maxBlockPropertiesBytes)
		if !validIdentifier || !validProperties || len(identifier) == 0 || !utf8.Valid(identifier) {
			return DamageSource{}, false
		}
		source.Block = &WorldBlock{Identifier: string(identifier), PropertiesNBT: properties}
	}
	return source, true
}

//export bg_go_player_effect
func bg_go_player_effect(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, operation C.uint32_t, value C.DfEffectView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.ChangePlayerEffect(InvocationID(invocation), playerID(player), PlayerEffectOperation(operation), PlayerEffect{
		Type: EffectType(value.effect_type), Level: int32(value.level),
		Duration: milliseconds(value.duration_milliseconds), Ambient: value.ambient != 0,
		Infinite: value.infinite != 0, ParticlesHidden: value.particles_hidden != 0,
	}) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_entity_visibility
func bg_go_player_entity_visibility(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, entity C.DfEntityId, visible C.uint8_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SetPlayerEntityVisible(InvocationID(invocation), playerID(player), entityID(entity), visible != 0) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_skin_open
func bg_go_player_skin_open(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, snapshot *C.uint64_t, info *C.DfSkinInfo) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || snapshot == nil || info == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.PlayerSkin(InvocationID(invocation), playerID(player))
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
func bg_go_player_skin_animation_info(context C.uint64_t, invocation C.DfInvocationId, snapshot C.uint64_t, index C.uint64_t, info *C.DfSkinAnimationInfo) C.DfStatus {
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
func bg_go_player_skin_read(context C.uint64_t, invocation C.DfInvocationId, snapshot C.uint64_t, data *C.DfSkinData) C.DfStatus {
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
func bg_go_player_skin_close(context C.uint64_t, invocation C.DfInvocationId, snapshot C.uint64_t) {
	unregisterSkinSnapshot(uint64(context), uint64(snapshot))
}

//export bg_go_player_skin_set
func bg_go_player_skin_set(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, view *C.DfSkinView) C.DfStatus {
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
	if !host.SetPlayerSkin(InvocationID(invocation), playerID(player), value) {
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
