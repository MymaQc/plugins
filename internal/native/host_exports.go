package native

/*
#include <stdlib.h>
#include "bridge.h"
*/
import "C"

import (
	"encoding/json"
	"math"
	"sync"
	"time"
	"unicode/utf8"
	"unsafe"
)

const (
	maxSkinDimension   = 4096
	maxSkinDataBytes   = 64 << 20
	maxSkinAnimations  = 64
	maxSkinIDBytes     = 4096
	maxScoreboardLines = 15
	maxScoreboardBytes = 16 << 20
	maxFormJSONBytes   = 1 << 20
)

//export bg_go_player_form_send
func bg_go_player_form_send(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, view *C.DfFormView) C.DfStatus {
	if view == nil || view.response == nil || view.drop == nil {
		return C.DF_STATUS_ERROR
	}
	// A structurally valid view transfers callback-context ownership. From this
	// point every failure must terminate it through drop exactly once.
	callbackContext, responseCallback, dropCallback := view.callback_context, view.response, view.drop
	drop := func() { C.bg_call_form_drop(dropCallback, callbackContext) }
	if view.request_json.len == 0 || view.request_json.len > maxFormJSONBytes || view.request_json.data == nil {
		drop()
		return C.DF_STATUS_ERROR
	}
	request := append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(view.request_json.data)), int(view.request_json.len))...)
	if !json.Valid(request) {
		drop()
		return C.DF_STATUS_ERROR
	}
	ok := sendPlayerForm(uint64(context), InvocationID(invocation), playerID(player), request, func(responseInvocation InvocationID, submitter PlayerSnapshot, closed bool, response []byte) bool {
		outcome := C.uint32_t(C.DF_FORM_RESPONSE_SUBMITTED)
		if closed {
			outcome = C.uint32_t(C.DF_FORM_RESPONSE_CLOSED)
		}
		name := C.CBytes([]byte(submitter.Name))
		if name != nil {
			defer C.free(name)
		}
		snapshot := C.DfPlayerSnapshot{
			player:               cPlayerID(submitter.Player),
			name:                 C.DfStringView{data: (*C.uint8_t)(name), len: C.uint64_t(len(submitter.Name))},
			latency_milliseconds: C.uint64_t(submitter.LatencyMilliseconds),
			position: C.DfVec3{
				x: C.double(submitter.Position.X), y: C.double(submitter.Position.Y), z: C.double(submitter.Position.Z),
			},
		}
		var responseView C.DfStringView
		if len(response) != 0 {
			responseView.data = (*C.uint8_t)(unsafe.Pointer(&response[0]))
			responseView.len = C.uint64_t(len(response))
		}
		return C.bg_call_form_response(responseCallback, callbackContext, C.DfInvocationId(responseInvocation), &snapshot, outcome, responseView) == C.DF_STATUS_OK
	}, drop)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

// sendPlayerForm consumes drop/respond ownership. It invokes exactly one of
// them, including when the host is unavailable or rejects the send.
func sendPlayerForm(context uint64, invocation InvocationID, player PlayerID, request []byte, respond func(InvocationID, PlayerSnapshot, bool, []byte) bool, drop func()) bool {
	terminal := &formTerminal{respondCallback: respond, dropCallback: drop}
	host, ok := resolveHost(context)
	if !ok {
		terminal.drop()
		return false
	}
	id, ok := registerForm(context, player, terminal.respond, terminal.drop)
	if !ok {
		terminal.drop()
		return false
	}
	if !host.SendPlayerForm(invocation, player, PlayerForm{ID: id, RequestJSON: request}) {
		CancelPlayerForm(id)
		// Cancel may lose a race with a drain or response that already removed the
		// registration. The terminal guard waits for that winner or drops here,
		// so callback_context is always released before this error is returned.
		terminal.drop()
		return false
	}
	return true
}

type formTerminal struct {
	once            sync.Once
	respondCallback func(InvocationID, PlayerSnapshot, bool, []byte) bool
	dropCallback    func()
}

func (t *formTerminal) respond(invocation InvocationID, submitter PlayerSnapshot, closed bool, response []byte) bool {
	accepted := false
	t.once.Do(func() { accepted = t.respondCallback(invocation, submitter, closed, response) })
	return accepted
}

func (t *formTerminal) drop() {
	t.once.Do(t.dropCallback)
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
	name, valid := copyNativeBytes(view.name, maxScoreboardBytes)
	if !valid || !utf8.Valid(name) {
		return C.DF_STATUS_ERROR
	}
	lineViews := unsafe.Slice(view.lines, int(view.line_count))
	lines := make([]string, len(lineViews))
	total := len(name)
	for index, line := range lineViews {
		value, valid := copyNativeBytes(line, maxScoreboardBytes-total)
		if !valid || !utf8.Valid(value) {
			return C.DF_STATUS_ERROR
		}
		total += len(value)
		lines[index] = string(value)
	}
	if !host.SendPlayerScoreboard(InvocationID(invocation), playerID(player), PlayerScoreboard{
		Name: string(name), Lines: lines, Padding: view.padding != 0, Descending: view.descending != 0,
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
	title := playerTitle(
		stringView(value.text), stringView(value.subtitle), stringView(value.action_text),
		int64(value.fade_in_duration_nanoseconds), int64(value.duration_nanoseconds), int64(value.fade_out_duration_nanoseconds),
	)
	if !host.SendPlayerTitle(InvocationID(invocation), id, title) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

func playerTitle(text, subtitle, actionText string, fadeIn, duration, fadeOut int64) PlayerTitle {
	return PlayerTitle{
		Text: text, Subtitle: subtitle, ActionText: actionText,
		FadeIn: time.Duration(fadeIn), Duration: time.Duration(duration), FadeOut: time.Duration(fadeOut),
	}
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

//export bg_go_player_knock_back
func bg_go_player_knock_back(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, source C.DfVec3, force C.double, height C.double) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.KnockBackPlayer(
		InvocationID(invocation),
		playerID(player),
		Vec3{X: float64(source.x), Y: float64(source.y), Z: float64(source.z)},
		float64(force),
		float64(height),
	) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_kinematics
func bg_go_player_kinematics(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, output *C.DfPlayerKinematics) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || output == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.PlayerKinematics(InvocationID(invocation), playerID(player))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	output.position = cEntityVec3(value.Position)
	output.velocity = cEntityVec3(value.Velocity)
	output.rotation = C.DfRotation{yaw: C.double(value.Rotation.Yaw), pitch: C.double(value.Rotation.Pitch)}
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

//export bg_go_player_action
func bg_go_player_action(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, kind C.uint32_t, value C.DfPlayerStateValue, result *C.DfPlayerStateValue) C.DfStatus {
	if result == nil {
		return C.DF_STATUS_ERROR
	}
	*result = C.DfPlayerStateValue{}
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	output, ok := host.PlayerAction(InvocationID(invocation), playerID(player), PlayerActionKind(kind), PlayerStateValue{
		Number: float64(value.number), Integer: int64(value.integer),
	})
	if !ok {
		return C.DF_STATUS_ERROR
	}
	result.number = C.double(output.Number)
	result.integer = C.int64_t(output.Integer)
	return C.DF_STATUS_OK
}

//export bg_go_player_block_action
func bg_go_player_block_action(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, kind C.uint32_t, position C.DfBlockPos, face C.int32_t, clickPosition C.DfVec3) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.PlayerBlockAction(
		InvocationID(invocation),
		playerID(player),
		PlayerBlockActionKind(kind),
		nativeBlockPosition(position),
		int32(face),
		nativeEntityVec3(clickPosition),
	) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_view_layer
func bg_go_player_view_layer(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, entity C.DfEntityId, kind C.uint32_t, text C.DfStringView, visibility C.uint8_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.PlayerViewLayer(
		InvocationID(invocation),
		playerID(player),
		entityID(entity),
		PlayerViewLayerKind(kind),
		stringView(text),
		uint8(visibility),
	) {
		return C.DF_STATUS_ERROR
	}
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

//export bg_go_player_final_damage
func bg_go_player_final_damage(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, damage C.double, view *C.DfDamageSourceView, result *C.double) C.DfStatus {
	if result == nil {
		return C.DF_STATUS_ERROR
	}
	*result = 0
	host, ok := resolveHost(uint64(context))
	source, valid := copyDamageSourceView(view)
	if !ok || !valid {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.FinalPlayerDamage(InvocationID(invocation), playerID(player), float64(damage), source)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*result = C.double(value)
	return C.DF_STATUS_OK
}

//export bg_go_player_using_item
func bg_go_player_using_item(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, usingItem *C.uint8_t) C.DfStatus {
	if usingItem == nil {
		return C.DF_STATUS_ERROR
	}
	*usingItem = 0
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.PlayerUsingItem(InvocationID(invocation), playerID(player))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	if value {
		*usingItem = 1
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_sleeping
func bg_go_player_sleeping(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, position *C.DfBlockPos, sleeping *C.uint8_t) C.DfStatus {
	if position == nil || sleeping == nil {
		return C.DF_STATUS_ERROR
	}
	*position = C.DfBlockPos{}
	*sleeping = 0
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	value, active, ok := host.PlayerSleeping(InvocationID(invocation), playerID(player))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	position.x, position.y, position.z = C.int32_t(value.X), C.int32_t(value.Y), C.int32_t(value.Z)
	if active {
		*sleeping = 1
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_death_position
func bg_go_player_death_position(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, position *C.DfVec3, dimension *C.DfDimensionView, found *C.uint8_t) C.DfStatus {
	if position == nil || dimension == nil || found == nil {
		return C.DF_STATUS_ERROR
	}
	*position, *dimension, *found = C.DfVec3{}, C.DfDimensionView{}, 0
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	value, dimensionValue, exists, ok := host.PlayerDeathPosition(InvocationID(invocation), playerID(player))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	if !exists {
		return C.DF_STATUS_OK
	}
	position.x, position.y, position.z = C.double(value.X), C.double(value.Y), C.double(value.Z)
	dimension.id = C.uint32_t(dimensionValue.ID)
	if dimensionValue.Custom != nil {
		dimension.custom = 1
		dimension.range_min = C.int32_t(dimensionValue.Custom.Range.Min)
		dimension.range_max = C.int32_t(dimensionValue.Custom.Range.Max)
		dimension.lava_spread_nanoseconds = C.int64_t(dimensionValue.Custom.LavaSpreadDuration)
		if dimensionValue.Custom.WaterEvaporates {
			dimension.water_evaporates = 1
		}
		if dimensionValue.Custom.WeatherCycle {
			dimension.weather_cycle = 1
		}
		if dimensionValue.Custom.TimeCycle {
			dimension.time_cycle = 1
		}
	} else if dimensionValue.ID > WorldDimensionEnd {
		return C.DF_STATUS_ERROR
	}
	*found = 1
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
	effect, valid := copyPlayerEffect(PlayerEffectOperation(operation), value)
	if !ok || !valid || !host.ChangePlayerEffect(InvocationID(invocation), playerID(player), PlayerEffectOperation(operation), effect) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

func copyPlayerEffect(operation PlayerEffectOperation, value C.DfEffectView) (PlayerEffect, bool) {
	if operation > PlayerEffectRemove || value.ambient > 1 || value.particles_hidden > 1 || value.infinite > 1 {
		return PlayerEffect{}, false
	}
	effect := PlayerEffect{
		Type: EffectType(value.effect_type), Level: int32(value.level),
		Duration: time.Duration(value.duration_nanoseconds), Potency: float64(value.potency),
		Ambient: value.ambient != 0, ParticlesHidden: value.particles_hidden != 0,
		Infinite: value.infinite != 0, Tick: int64(value.tick),
	}
	return effect, true
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
	snapshotID, ok := registerSkinSnapshot(uint64(context), InvocationID(invocation), value, false)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*snapshot = C.uint64_t(snapshotID)
	fillSkinInfo(info, value)
	return C.DF_STATUS_OK
}

//export bg_go_player_skin_animation_info
func bg_go_player_skin_animation_info(context C.uint64_t, invocation C.DfInvocationId, snapshot C.uint64_t, index C.uint64_t, info *C.DfSkinAnimationInfo) C.DfStatus {
	value, ok := resolveSkinSnapshot(uint64(context), InvocationID(invocation), uint64(snapshot))
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
	value, ok := resolveSkinSnapshot(uint64(context), InvocationID(invocation), uint64(snapshot))
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
	unregisterSkinSnapshot(uint64(context), InvocationID(invocation), uint64(snapshot))
}

//export bg_go_player_skin_set
func bg_go_player_skin_set(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, view *C.DfSkinView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || view == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := copyPlayerSkinView(view)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	if !host.SetPlayerSkin(InvocationID(invocation), playerID(player), value) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_skin_snapshot_info
func bg_go_skin_snapshot_info(context C.uint64_t, invocation C.DfInvocationId, snapshot C.uint64_t, info *C.DfSkinInfo) C.DfStatus {
	value, ok := resolveSkinSnapshot(uint64(context), InvocationID(invocation), uint64(snapshot))
	if !ok || info == nil {
		return C.DF_STATUS_ERROR
	}
	fillSkinInfo(info, value)
	return C.DF_STATUS_OK
}

//export bg_go_skin_snapshot_set
func bg_go_skin_snapshot_set(context C.uint64_t, invocation C.DfInvocationId, snapshot C.uint64_t, view *C.DfSkinView) C.DfStatus {
	if view == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := copyPlayerSkinView(view)
	if !ok || !replaceEventSkinSnapshot(uint64(context), InvocationID(invocation), uint64(snapshot), value) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

func copyPlayerSkinView(view *C.DfSkinView) (PlayerSkin, bool) {
	if view == nil || !validSkinViewPayload(view) {
		return PlayerSkin{}, false
	}
	playFabID, ok := copySkinView(view.play_fab_id)
	if !ok {
		return PlayerSkin{}, false
	}
	fullID, ok := copySkinView(view.full_id)
	if !ok {
		return PlayerSkin{}, false
	}
	pixels, ok := copySkinView(view.pixels)
	if !ok {
		return PlayerSkin{}, false
	}
	modelDefault, ok := copySkinView(view.model_default)
	if !ok {
		return PlayerSkin{}, false
	}
	modelAnimatedFace, ok := copySkinView(view.model_animated_face)
	if !ok {
		return PlayerSkin{}, false
	}
	model, ok := copySkinView(view.model)
	if !ok {
		return PlayerSkin{}, false
	}
	capePixels, ok := copySkinView(view.cape_pixels)
	if !ok {
		return PlayerSkin{}, false
	}
	value := PlayerSkin{
		Width: uint32(view.width), Height: uint32(view.height), Persona: view.persona != 0,
		PlayFabID: string(playFabID), FullID: string(fullID), Pixels: pixels,
		ModelDefault: string(modelDefault), ModelAnimatedFace: string(modelAnimatedFace), Model: model,
		CapeWidth: uint32(view.cape_width), CapeHeight: uint32(view.cape_height), CapePixels: capePixels,
	}
	if view.animation_count != 0 {
		animations := unsafe.Slice(view.animations, int(view.animation_count))
		value.Animations = make([]SkinAnimation, len(animations))
		for index, animation := range animations {
			animationPixels, ok := copySkinView(animation.pixels)
			if !ok {
				return PlayerSkin{}, false
			}
			value.Animations[index] = SkinAnimation{
				Width: uint32(animation.width), Height: uint32(animation.height), Type: uint32(animation.animation_type),
				FrameCount: int64(animation.frame_count), Expression: int64(animation.expression), Pixels: animationPixels,
			}
		}
	}
	if !validPlayerSkinPayload(value) {
		return PlayerSkin{}, false
	}
	return value, true
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
	if !validNativeSkinPixels(value.Width, value.Height, value.Pixels, true) ||
		!validNativeSkinPixels(value.CapeWidth, value.CapeHeight, value.CapePixels, true) ||
		len(value.Animations) > maxSkinAnimations ||
		len(value.PlayFabID) > maxSkinIDBytes || len(value.FullID) > maxSkinIDBytes ||
		len(value.ModelDefault) > maxSkinIDBytes || len(value.ModelAnimatedFace) > maxSkinIDBytes ||
		!utf8.ValidString(value.PlayFabID) || !utf8.ValidString(value.FullID) ||
		!utf8.ValidString(value.ModelDefault) || !utf8.ValidString(value.ModelAnimatedFace) {
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
		if !validNativeSkinPixels(animation.Width, animation.Height, animation.Pixels, false) ||
			animation.Type > 2 || animation.FrameCount <= 0 || animation.FrameCount > int64(math.MaxInt) ||
			animation.Expression < int64(math.MinInt) || animation.Expression > int64(math.MaxInt) {
			return false
		}
		if uint64(len(animation.Pixels)) > uint64(maxSkinDataBytes)-total {
			return false
		}
		total += uint64(len(animation.Pixels))
	}
	return true
}

func validNativeSkinPixels(width, height uint32, pixels []byte, empty bool) bool {
	if width > maxSkinDimension || height > maxSkinDimension {
		return false
	}
	if width == 0 || height == 0 {
		return empty && width == 0 && height == 0 && len(pixels) == 0
	}
	return uint64(width)*uint64(height)*4 == uint64(len(pixels))
}

func validSkinViewPayload(view *C.DfSkinView) bool {
	if view.persona > 1 || uint64(view.animation_count) > maxSkinAnimations || view.animation_count != 0 && view.animations == nil {
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
