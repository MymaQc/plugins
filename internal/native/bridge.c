#include "bridge.h"

#include <dlfcn.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#if UINTPTR_MAX == UINT64_MAX
_Static_assert(sizeof(DfSkinAnimationInfo) == 40, "DfSkinAnimationInfo ABI layout changed");
_Static_assert(offsetof(DfSkinAnimationInfo, frame_count) == 16, "DfSkinAnimationInfo.frame_count ABI offset changed");
_Static_assert(offsetof(DfSkinAnimationInfo, pixels_len) == 32, "DfSkinAnimationInfo.pixels_len ABI offset changed");
_Static_assert(sizeof(DfSkinInfo) == 88, "DfSkinInfo ABI layout changed");
_Static_assert(offsetof(DfSkinInfo, play_fab_id_len) == 16, "DfSkinInfo.play_fab_id_len ABI offset changed");
_Static_assert(offsetof(DfSkinInfo, cape_pixels_len) == 72, "DfSkinInfo.cape_pixels_len ABI offset changed");
_Static_assert(sizeof(DfSkinData) == 184, "DfSkinData ABI layout changed");
_Static_assert(offsetof(DfSkinData, animation_pixels) == 168, "DfSkinData.animation_pixels ABI offset changed");
_Static_assert(sizeof(DfSkinAnimationView) == 48, "DfSkinAnimationView ABI layout changed");
_Static_assert(offsetof(DfSkinAnimationView, pixels) == 32, "DfSkinAnimationView.pixels ABI offset changed");
_Static_assert(sizeof(DfSkinView) == 152, "DfSkinView ABI layout changed");
_Static_assert(offsetof(DfSkinView, animations) == 136, "DfSkinView.animations ABI offset changed");
_Static_assert(sizeof(DfPlayerSkinChangeInput) == 40, "DfPlayerSkinChangeInput ABI layout changed");
_Static_assert(offsetof(DfPlayerSkinChangeInput, snapshot) == 32, "DfPlayerSkinChangeInput.snapshot ABI offset changed");
_Static_assert(sizeof(DfPlayerSkinChangeState) == 1, "DfPlayerSkinChangeState ABI layout changed");
_Static_assert(sizeof(DfInventoryId) == 32, "DfInventoryId ABI layout changed");
_Static_assert(sizeof(DfItemStackInfo) == 80, "DfItemStackInfo ABI layout changed");
_Static_assert(sizeof(DfItemStackSnapshot) == 88, "DfItemStackSnapshot ABI layout changed");
_Static_assert(sizeof(DfItemStackData) == 152, "DfItemStackData ABI layout changed");
_Static_assert(sizeof(DfItemStackViewV3) == 120, "DfItemStackViewV3 ABI layout changed");
_Static_assert(sizeof(DfItemStackSnapshot) == 88, "DfItemStackSnapshot ABI layout changed");
_Static_assert(sizeof(DfScoreboardView) == 40, "DfScoreboardView ABI layout changed");
_Static_assert(offsetof(DfScoreboardView, lines) == 16, "DfScoreboardView.lines ABI offset changed");
_Static_assert(sizeof(DfFormView) == 40, "DfFormView ABI layout changed");
_Static_assert(sizeof(DfWorldId) == 8, "DfWorldId ABI layout changed");
_Static_assert(sizeof(DfBlockData) == 48, "DfBlockData ABI layout changed");
_Static_assert(sizeof(DfBlockView) == 32, "DfBlockView ABI layout changed");
_Static_assert(sizeof(DfDamageSourceView) == 88, "DfDamageSourceView ABI layout changed");
_Static_assert(offsetof(DfDamageSourceView, entity) == 24, "DfDamageSourceView.entity ABI offset changed");
_Static_assert(offsetof(DfDamageSourceView, block) == 72, "DfDamageSourceView.block ABI offset changed");
_Static_assert(sizeof(DfHealingSourceView) == 24, "DfHealingSourceView ABI layout changed");
_Static_assert(sizeof(DfPlayerHealResult) == 8, "DfPlayerHealResult ABI layout changed");
_Static_assert(sizeof(DfPlayerHurtResult) == 16, "DfPlayerHurtResult ABI layout changed");
_Static_assert(sizeof(DfEffectView) == 32, "DfEffectView ABI layout changed");
_Static_assert(offsetof(DfEffectView, potency) == 16, "DfEffectView.potency ABI offset changed");
_Static_assert(offsetof(DfEffectView, mode) == 24, "DfEffectView.mode ABI offset changed");
_Static_assert(sizeof(DfEffectBuffer) == 24, "DfEffectBuffer ABI layout changed");
_Static_assert(offsetof(DfEffectBuffer, data) == 0, "DfEffectBuffer.data ABI offset changed");
_Static_assert(offsetof(DfEffectBuffer, len) == 8, "DfEffectBuffer.len ABI offset changed");
_Static_assert(offsetof(DfEffectBuffer, capacity) == 16, "DfEffectBuffer.capacity ABI offset changed");
_Static_assert(sizeof(DfEntitySpawnOptions) == 80, "DfEntitySpawnOptions ABI layout changed");
_Static_assert(sizeof(DfEntityTypeDescriptorV2) == 144, "DfEntityTypeDescriptorV2 ABI layout changed");
_Static_assert(offsetof(DfEntityTypeDescriptorV2, type_key) == 80, "DfEntityTypeDescriptorV2.type_key ABI offset changed");
_Static_assert(sizeof(DfEntitySpawnViewV3) == 200, "DfEntitySpawnViewV3 ABI layout changed");
_Static_assert(offsetof(DfEntitySpawnViewV3, custom_instance) == 176, "DfEntitySpawnViewV3.custom_instance ABI offset changed");
_Static_assert(sizeof(DfPluginApiV3) == 128, "DfPluginApiV3 ABI layout changed");
_Static_assert(offsetof(DfPluginApiV3, entity_type_count) == 64, "DfPluginApiV3.entity_type_count ABI offset changed");
_Static_assert(offsetof(DfPluginApiV3, handle_entity) == 80, "DfPluginApiV3.handle_entity ABI offset changed");
_Static_assert(sizeof(DfEntityState) == 128, "DfEntityState ABI layout changed");
_Static_assert(offsetof(DfEntityState, world) == 72, "DfEntityState.world ABI offset changed");
_Static_assert(sizeof(DfParticleViewV1) == 40, "DfParticleViewV1 ABI layout changed");
_Static_assert(offsetof(DfParticleViewV1, block) == 32, "DfParticleViewV1.block ABI offset changed");
_Static_assert(sizeof(DfSoundViewV1) == 40, "DfSoundViewV1 ABI layout changed");
_Static_assert(offsetof(DfSoundViewV1, scalar) == 16, "DfSoundViewV1.scalar ABI offset changed");
_Static_assert(offsetof(DfSoundViewV1, item) == 32, "DfSoundViewV1.item ABI offset changed");
_Static_assert(sizeof(DfWorldOpenSpecV1) == 80, "DfWorldOpenSpecV1 ABI layout changed");
_Static_assert(offsetof(DfWorldOpenSpecV1, provider_path) == 8, "DfWorldOpenSpecV1.provider_path ABI offset changed");
_Static_assert(offsetof(DfWorldOpenSpecV1, fixed_time) == 40, "DfWorldOpenSpecV1.fixed_time ABI offset changed");
_Static_assert(offsetof(DfWorldOpenSpecV1, open_mode) == 48, "DfWorldOpenSpecV1.open_mode ABI offset changed");
_Static_assert(offsetof(DfWorldOpenSpecV1, read_only) == 76, "DfWorldOpenSpecV1.read_only ABI offset changed");
_Static_assert(sizeof(DfHostApiV18) == 480, "DfHostApiV18 ABI layout changed");
_Static_assert(offsetof(DfHostApiV18, player_skin_open) == 80, "DfHostApiV18.player_skin_open ABI offset changed");
_Static_assert(offsetof(DfHostApiV18, player_skin_set) == 112, "DfHostApiV18.player_skin_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV18, inventory_size) == 120, "DfHostApiV18.inventory_size ABI offset changed");
_Static_assert(offsetof(DfHostApiV18, player_held_slot_set) == 200, "DfHostApiV18.player_held_slot_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV18, player_scoreboard) == 208, "DfHostApiV18.player_scoreboard ABI offset changed");
_Static_assert(offsetof(DfHostApiV18, player_heal) == 416, "DfHostApiV18.player_heal ABI offset changed");
_Static_assert(offsetof(DfHostApiV18, player_hurt) == 424, "DfHostApiV18.player_hurt ABI offset changed");
_Static_assert(offsetof(DfHostApiV18, skin_snapshot_info) == 432, "DfHostApiV18.skin_snapshot_info ABI offset changed");
_Static_assert(offsetof(DfHostApiV18, skin_snapshot_set) == 440, "DfHostApiV18.skin_snapshot_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV18, world_open_spec) == 448, "DfHostApiV18.world_open_spec ABI offset changed");
_Static_assert(offsetof(DfHostApiV18, player_transfer) == 456, "DfHostApiV18.player_transfer ABI offset changed");
_Static_assert(offsetof(DfHostApiV18, player_effects) == 464, "DfHostApiV18.player_effects ABI offset changed");
_Static_assert(offsetof(DfHostApiV18, player_effects_clear) == 472, "DfHostApiV18.player_effects_clear ABI offset changed");
#endif

extern DfStatus bg_go_player_text(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfStringView message);
extern DfStatus bg_go_player_title(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfTitleView title);
extern DfStatus bg_go_player_scoreboard(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfScoreboardView scoreboard);
extern DfStatus bg_go_player_scoreboard_remove(uint64_t context, DfInvocationId invocation, DfPlayerId player);
extern DfStatus bg_go_player_form_send(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfFormView *form);
extern DfStatus bg_go_player_form_close(uint64_t context, DfInvocationId invocation, DfPlayerId player);

DfStatus bg_call_form_response(DfFormResponseFn callback, void *callback_context, DfInvocationId invocation, DfPlayerId submitter, uint32_t outcome, DfStringView response_json) { return callback(callback_context, invocation, submitter, outcome, response_json); }
void bg_call_form_drop(DfFormDropFn callback, void *callback_context) { callback(callback_context); }
extern DfStatus bg_go_player_transform(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfVec3 vector, double yaw, double pitch);
extern DfStatus bg_go_player_transfer(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfWorldId world, DfVec3 position);
extern DfStatus bg_go_player_rotation(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfRotation *rotation);
extern DfStatus bg_go_player_state_set(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue value);
extern DfStatus bg_go_player_state_get(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue *value);
extern DfStatus bg_go_player_heal(uint64_t context, DfInvocationId invocation, DfPlayerId player, double health, const DfHealingSourceView *source, DfPlayerHealResult *result);
extern DfStatus bg_go_player_hurt(uint64_t context, DfInvocationId invocation, DfPlayerId player, double damage, const DfDamageSourceView *source, DfPlayerHurtResult *result);
extern DfStatus bg_go_player_effect(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t operation, DfEffectView effect);
extern DfStatus bg_go_player_effects(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfEffectBuffer *output);
extern DfStatus bg_go_player_effects_clear(uint64_t context, DfInvocationId invocation, DfPlayerId player);
extern DfStatus bg_go_player_entity_visibility(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfEntityId entity, uint8_t visible);
extern DfStatus bg_go_player_skin_open(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint64_t *snapshot, DfSkinInfo *info);
extern DfStatus bg_go_player_skin_animation_info(uint64_t context, DfInvocationId invocation, uint64_t snapshot, uint64_t index, DfSkinAnimationInfo *info);
extern DfStatus bg_go_player_skin_read(uint64_t context, DfInvocationId invocation, uint64_t snapshot, DfSkinData *data);
extern void bg_go_player_skin_close(uint64_t context, DfInvocationId invocation, uint64_t snapshot);
extern DfStatus bg_go_player_skin_set(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfSkinView *skin);
extern DfStatus bg_go_skin_snapshot_info(uint64_t context, DfInvocationId invocation, uint64_t snapshot, DfSkinInfo *info);
extern DfStatus bg_go_skin_snapshot_set(uint64_t context, DfInvocationId invocation, uint64_t snapshot, const DfSkinView *skin);
extern DfStatus bg_go_inventory_size(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, uint32_t *size);
extern DfStatus bg_go_inventory_item_open(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, uint32_t slot, uint64_t *snapshot, DfItemStackInfo *info);
extern DfStatus bg_go_player_held_item_open(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t hand, uint64_t *snapshot, DfItemStackInfo *info);
extern DfStatus bg_go_item_stack_read(uint64_t context, DfInvocationId invocation, uint64_t snapshot, DfItemStackData *data);
extern void bg_go_item_stack_close(uint64_t context, DfInvocationId invocation, uint64_t snapshot);
extern DfStatus bg_go_inventory_item_set(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, uint32_t slot, const DfItemStackViewV3 *item);
extern DfStatus bg_go_inventory_item_add(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, const DfItemStackViewV3 *item, uint32_t *added);
extern DfStatus bg_go_inventory_clear_slot(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, uint32_t slot);
extern DfStatus bg_go_inventory_clear(uint64_t context, DfInvocationId invocation, DfInventoryId inventory);
extern DfStatus bg_go_player_held_items_set(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfItemStackViewV3 *main_hand, const DfItemStackViewV3 *off_hand);
extern DfStatus bg_go_player_held_slot_set(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t slot);
extern DfStatus bg_go_world_lookup(uint64_t context, DfInvocationId invocation, DfStringView name, DfWorldId *world);
extern DfStatus bg_go_world_open(uint64_t context, DfInvocationId invocation, DfStringView name, uint32_t dimension, DfWorldId *world);
extern DfStatus bg_go_world_open_spec(uint64_t context, DfInvocationId invocation, DfStringView name, const DfWorldOpenSpecV1 *spec, DfWorldId *world);
extern DfStatus bg_go_world_name(uint64_t context, DfInvocationId invocation, DfWorldId world, DfStringBuffer *name);
extern DfStatus bg_go_world_unload(uint64_t context, DfInvocationId invocation, DfWorldId world);
extern DfStatus bg_go_world_save(uint64_t context, DfInvocationId invocation, DfWorldId world);
extern DfStatus bg_go_world_block_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, DfBlockData *block);
extern DfStatus bg_go_world_block_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, const DfBlockView *block);
extern DfStatus bg_go_world_time_get(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t *time);
extern DfStatus bg_go_world_time_set(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t time);
extern DfStatus bg_go_world_spawn_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos *position);
extern DfStatus bg_go_world_spawn_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position);
extern DfStatus bg_go_world_entity_spawn(uint64_t context, DfInvocationId invocation, DfWorldId world, const DfEntitySpawnViewV3 *entity, DfEntityId *output);
extern DfStatus bg_go_world_entities(uint64_t context, DfInvocationId invocation, DfWorldId world, DfEntityIdBuffer *output);
extern DfStatus bg_go_world_players(uint64_t context, DfInvocationId invocation, DfWorldId world, DfPlayerIdBuffer *output);
extern DfStatus bg_go_entity_state(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfEntityState *state);
extern DfStatus bg_go_entity_teleport(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfVec3 position);
extern DfStatus bg_go_entity_velocity_set(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfVec3 velocity);
extern DfStatus bg_go_entity_name_tag_set(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfStringView name_tag);
extern DfStatus bg_go_entity_despawn(uint64_t context, DfInvocationId invocation, DfEntityId entity);
extern DfStatus bg_go_world_particle_add(uint64_t context, DfInvocationId invocation, DfWorldId world, DfVec3 position, const DfParticleViewV1 *particle);
extern DfStatus bg_go_world_sound_play(uint64_t context, DfInvocationId invocation, DfWorldId world, DfVec3 position, const DfSoundViewV1 *sound);
extern DfStatus bg_go_player_sound_play(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfSoundViewV1 *sound);

static DfStatus host_player_text(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfStringView message) {
    return bg_go_player_text(context, invocation, player, kind, message);
}

static DfStatus host_player_title(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfTitleView title) {
    return bg_go_player_title(context, invocation, player, title);
}

static DfStatus host_player_scoreboard(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfScoreboardView scoreboard) {
    return bg_go_player_scoreboard(context, invocation, player, scoreboard);
}

static DfStatus host_player_scoreboard_remove(uint64_t context, DfInvocationId invocation, DfPlayerId player) {
    return bg_go_player_scoreboard_remove(context, invocation, player);
}
static DfStatus host_player_form_send(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfFormView *form) { return bg_go_player_form_send(context, invocation, player, form); }
static DfStatus host_player_form_close(uint64_t context, DfInvocationId invocation, DfPlayerId player) { return bg_go_player_form_close(context, invocation, player); }

static DfStatus host_player_transform(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfVec3 vector, double yaw, double pitch) {
    return bg_go_player_transform(context, invocation, player, kind, vector, yaw, pitch);
}

static DfStatus host_player_rotation(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfRotation *rotation) {
    return bg_go_player_rotation(context, invocation, player, rotation);
}

static DfStatus host_player_state_set(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue value) {
    return bg_go_player_state_set(context, invocation, player, kind, value);
}

static DfStatus host_player_state_get(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue *value) {
    return bg_go_player_state_get(context, invocation, player, kind, value);
}

static DfStatus host_player_heal(uint64_t context, DfInvocationId invocation, DfPlayerId player, double health, const DfHealingSourceView *source, DfPlayerHealResult *result) {
    return bg_go_player_heal(context, invocation, player, health, source, result);
}

static DfStatus host_player_hurt(uint64_t context, DfInvocationId invocation, DfPlayerId player, double damage, const DfDamageSourceView *source, DfPlayerHurtResult *result) {
    return bg_go_player_hurt(context, invocation, player, damage, source, result);
}

static DfStatus host_player_effect(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t operation, DfEffectView effect) {
    return bg_go_player_effect(context, invocation, player, operation, effect);
}

static DfStatus host_player_effects(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfEffectBuffer *output) {
    return bg_go_player_effects(context, invocation, player, output);
}

static DfStatus host_player_effects_clear(uint64_t context, DfInvocationId invocation, DfPlayerId player) {
    return bg_go_player_effects_clear(context, invocation, player);
}

static DfStatus host_player_entity_visibility(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfEntityId entity, uint8_t visible) {
    return bg_go_player_entity_visibility(context, invocation, player, entity, visible);
}

static DfStatus host_player_skin_open(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint64_t *snapshot, DfSkinInfo *info) {
    return bg_go_player_skin_open(context, invocation, player, snapshot, info);
}

static DfStatus host_player_skin_animation_info(uint64_t context, DfInvocationId invocation, uint64_t snapshot, uint64_t index, DfSkinAnimationInfo *info) {
    return bg_go_player_skin_animation_info(context, invocation, snapshot, index, info);
}

static DfStatus host_player_skin_read(uint64_t context, DfInvocationId invocation, uint64_t snapshot, DfSkinData *data) {
    return bg_go_player_skin_read(context, invocation, snapshot, data);
}

static void host_player_skin_close(uint64_t context, DfInvocationId invocation, uint64_t snapshot) {
    bg_go_player_skin_close(context, invocation, snapshot);
}

static DfStatus host_player_skin_set(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfSkinView *skin) {
    return bg_go_player_skin_set(context, invocation, player, skin);
}

static DfStatus host_player_transfer(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfWorldId world, DfVec3 position) {
    return bg_go_player_transfer(context, invocation, player, world, position);
}

static DfStatus host_skin_snapshot_info(uint64_t context, DfInvocationId invocation, uint64_t snapshot, DfSkinInfo *info) {
    return bg_go_skin_snapshot_info(context, invocation, snapshot, info);
}

static DfStatus host_skin_snapshot_set(uint64_t context, DfInvocationId invocation, uint64_t snapshot, const DfSkinView *skin) {
    return bg_go_skin_snapshot_set(context, invocation, snapshot, skin);
}

static DfStatus host_inventory_size(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, uint32_t *size) { return bg_go_inventory_size(context, invocation, inventory, size); }
static DfStatus host_inventory_item_open(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, uint32_t slot, uint64_t *snapshot, DfItemStackInfo *info) { return bg_go_inventory_item_open(context, invocation, inventory, slot, snapshot, info); }
static DfStatus host_player_held_item_open(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t hand, uint64_t *snapshot, DfItemStackInfo *info) { return bg_go_player_held_item_open(context, invocation, player, hand, snapshot, info); }
static DfStatus host_item_stack_read(uint64_t context, DfInvocationId invocation, uint64_t snapshot, DfItemStackData *data) { return bg_go_item_stack_read(context, invocation, snapshot, data); }
static void host_item_stack_close(uint64_t context, DfInvocationId invocation, uint64_t snapshot) { bg_go_item_stack_close(context, invocation, snapshot); }
static DfStatus host_inventory_item_set(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, uint32_t slot, const DfItemStackViewV3 *item) { return bg_go_inventory_item_set(context, invocation, inventory, slot, item); }
static DfStatus host_inventory_item_add(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, const DfItemStackViewV3 *item, uint32_t *added) { return bg_go_inventory_item_add(context, invocation, inventory, item, added); }
static DfStatus host_inventory_clear_slot(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, uint32_t slot) { return bg_go_inventory_clear_slot(context, invocation, inventory, slot); }
static DfStatus host_inventory_clear(uint64_t context, DfInvocationId invocation, DfInventoryId inventory) { return bg_go_inventory_clear(context, invocation, inventory); }
static DfStatus host_player_held_items_set(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfItemStackViewV3 *main_hand, const DfItemStackViewV3 *off_hand) { return bg_go_player_held_items_set(context, invocation, player, main_hand, off_hand); }
static DfStatus host_player_held_slot_set(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t slot) { return bg_go_player_held_slot_set(context, invocation, player, slot); }
static DfStatus host_world_lookup(uint64_t context, DfInvocationId invocation, DfStringView name, DfWorldId *world) { return bg_go_world_lookup(context, invocation, name, world); }
static DfStatus host_world_open(uint64_t context, DfInvocationId invocation, DfStringView name, uint32_t dimension, DfWorldId *world) { return bg_go_world_open(context, invocation, name, dimension, world); }
static DfStatus host_world_open_spec(uint64_t context, DfInvocationId invocation, DfStringView name, const DfWorldOpenSpecV1 *spec, DfWorldId *world) { return bg_go_world_open_spec(context, invocation, name, spec, world); }
static DfStatus host_world_name(uint64_t context, DfInvocationId invocation, DfWorldId world, DfStringBuffer *name) { return bg_go_world_name(context, invocation, world, name); }
static DfStatus host_world_unload(uint64_t context, DfInvocationId invocation, DfWorldId world) { return bg_go_world_unload(context, invocation, world); }
static DfStatus host_world_save(uint64_t context, DfInvocationId invocation, DfWorldId world) { return bg_go_world_save(context, invocation, world); }
static DfStatus host_world_block_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, DfBlockData *block) { return bg_go_world_block_get(context, invocation, world, position, block); }
static DfStatus host_world_block_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, const DfBlockView *block) { return bg_go_world_block_set(context, invocation, world, position, block); }
static DfStatus host_world_time_get(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t *time) { return bg_go_world_time_get(context, invocation, world, time); }
static DfStatus host_world_time_set(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t time) { return bg_go_world_time_set(context, invocation, world, time); }
static DfStatus host_world_spawn_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos *position) { return bg_go_world_spawn_get(context, invocation, world, position); }
static DfStatus host_world_spawn_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position) { return bg_go_world_spawn_set(context, invocation, world, position); }
static DfStatus host_world_entity_spawn(uint64_t context, DfInvocationId invocation, DfWorldId world, const DfEntitySpawnViewV3 *entity, DfEntityId *output) { return bg_go_world_entity_spawn(context, invocation, world, entity, output); }
static DfStatus host_world_entities(uint64_t context, DfInvocationId invocation, DfWorldId world, DfEntityIdBuffer *output) { return bg_go_world_entities(context, invocation, world, output); }
static DfStatus host_world_players(uint64_t context, DfInvocationId invocation, DfWorldId world, DfPlayerIdBuffer *output) { return bg_go_world_players(context, invocation, world, output); }
static DfStatus host_entity_state(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfEntityState *state) { return bg_go_entity_state(context, invocation, entity, state); }
static DfStatus host_entity_teleport(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfVec3 position) { return bg_go_entity_teleport(context, invocation, entity, position); }
static DfStatus host_entity_velocity_set(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfVec3 velocity) { return bg_go_entity_velocity_set(context, invocation, entity, velocity); }
static DfStatus host_entity_name_tag_set(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfStringView name_tag) { return bg_go_entity_name_tag_set(context, invocation, entity, name_tag); }
static DfStatus host_entity_despawn(uint64_t context, DfInvocationId invocation, DfEntityId entity) { return bg_go_entity_despawn(context, invocation, entity); }
static DfStatus host_world_particle_add(uint64_t context, DfInvocationId invocation, DfWorldId world, DfVec3 position, const DfParticleViewV1 *particle) { return bg_go_world_particle_add(context, invocation, world, position, particle); }
static DfStatus host_world_sound_play(uint64_t context, DfInvocationId invocation, DfWorldId world, DfVec3 position, const DfSoundViewV1 *sound) { return bg_go_world_sound_play(context, invocation, world, position, sound); }
static DfStatus host_player_sound_play(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfSoundViewV1 *sound) { return bg_go_player_sound_play(context, invocation, player, sound); }

typedef DfStatus (*RuntimeCreateFn)(const DfRuntimeConfig *, DfRuntime **, uint8_t *, uint64_t);
typedef void (*RuntimeDestroyFn)(DfRuntime *);
typedef DfStatus (*RuntimeEnableFn)(DfRuntime *);
typedef void (*RuntimeDisableFn)(DfRuntime *);
typedef uint64_t (*RuntimeCountFn)(const DfRuntime *);
typedef DfStatus (*RuntimeEntityTypeAtFn)(const DfRuntime *, uint64_t, DfEntityTypeDescriptorV2 *);
typedef DfStatus (*RuntimeEntityAdoptFn)(DfRuntime *, uint64_t, uint64_t, DfEntityInstanceId *);
typedef DfStatus (*RuntimeEntityLoadFn)(DfRuntime *, uint64_t, const DfEntityLoadInput *, DfEntityInstanceId *);
typedef DfStatus (*RuntimeEntitySaveFn)(DfRuntime *, DfEntityInstanceId, DfEntitySaveState *);
typedef DfStatus (*RuntimeEntityTickFn)(DfRuntime *, DfEntityInstanceId, const DfEntityTickInput *, DfEntityTickState *);
typedef DfStatus (*RuntimeEntityHurtFn)(DfRuntime *, DfEntityInstanceId, const DfEntityHurtInput *, DfEntityHurtState *);
typedef DfStatus (*RuntimeEntityHealFn)(DfRuntime *, DfEntityInstanceId, const DfEntityHealInput *, DfEntityHealState *);
typedef DfStatus (*RuntimeEntityDeathFn)(DfRuntime *, DfEntityInstanceId, const DfEntityDeathInput *, DfEntityDeathState *);
typedef DfStatus (*RuntimeEntityDestroyFn)(DfRuntime *, DfEntityInstanceId);
typedef DfStatus (*RuntimeCommandAtFn)(const DfRuntime *, uint64_t, DfCommandDescriptor *);
typedef DfStatus (*RuntimeCommandFn)(DfRuntime *, uint64_t, const DfCommandInput *, DfCommandState *);
typedef DfStatus (*RuntimeCommandEnumFn)(DfRuntime *, uint64_t, uint64_t, uint64_t, const DfCommandEnumContext *, DfStringBuffer *);
typedef DfStatus (*RuntimeEventFn)(DfRuntime *, DfEventId, const void *, void *);

struct BgRuntimeLibrary {
    void *handle;
    DfRuntime *runtime;
    DfHostApiV18 host_api;
    RuntimeDestroyFn destroy;
    RuntimeEnableFn enable;
    RuntimeDisableFn disable;
    RuntimeCountFn plugin_count;
    RuntimeCountFn subscriptions;
    RuntimeCountFn entity_type_count;
    RuntimeEntityTypeAtFn entity_type_at;
    RuntimeEntityAdoptFn entity_adopt;
    RuntimeEntityLoadFn entity_load;
    RuntimeEntitySaveFn entity_save;
    RuntimeEntityTickFn entity_tick;
    RuntimeEntityHurtFn entity_hurt;
    RuntimeEntityHealFn entity_heal;
    RuntimeEntityDeathFn entity_death;
    RuntimeEntityDestroyFn entity_destroy;
    RuntimeCountFn command_count;
    RuntimeCommandAtFn command_at;
    RuntimeCommandFn handle_command;
    RuntimeCommandEnumFn command_enum_options;
    RuntimeEventFn handle_event;
};

static void write_error(uint8_t *error, uint64_t capacity, const char *message) {
    if (error == NULL || capacity == 0) {
        return;
    }
    snprintf((char *) error, (size_t) capacity, "%s", message == NULL ? "unknown error" : message);
}

static void *load_symbol(void *handle, const char *name, uint8_t *error, uint64_t capacity) {
    dlerror();
    void *symbol = dlsym(handle, name);
    const char *failure = dlerror();
    if (failure != NULL) {
        write_error(error, capacity, failure);
        return NULL;
    }
    return symbol;
}

DfStatus bg_runtime_open(
    const char *library_path,
    const char *plugin_directory,
    uint64_t host_context,
    BgRuntimeLibrary **out,
    uint8_t *error,
    uint64_t error_capacity
) {
    if (library_path == NULL || plugin_directory == NULL || out == NULL) {
        write_error(error, error_capacity, "null runtime path, plugin directory, or output");
        return DF_STATUS_ERROR;
    }
    *out = NULL;
    void *handle = dlopen(library_path, RTLD_NOW | RTLD_LOCAL);
    if (handle == NULL) {
        write_error(error, error_capacity, dlerror());
        return DF_STATUS_ERROR;
    }

    RuntimeCreateFn create = (RuntimeCreateFn) load_symbol(handle, "df_runtime_create", error, error_capacity);
    RuntimeDestroyFn destroy = (RuntimeDestroyFn) load_symbol(handle, "df_runtime_destroy", error, error_capacity);
    RuntimeEnableFn enable = (RuntimeEnableFn) load_symbol(handle, "df_runtime_enable", error, error_capacity);
    RuntimeDisableFn disable = (RuntimeDisableFn) load_symbol(handle, "df_runtime_disable", error, error_capacity);
    RuntimeCountFn plugin_count = (RuntimeCountFn) load_symbol(handle, "df_runtime_plugin_count", error, error_capacity);
    RuntimeCountFn subscriptions = (RuntimeCountFn) load_symbol(handle, "df_runtime_subscriptions", error, error_capacity);
    RuntimeCountFn entity_type_count = (RuntimeCountFn) load_symbol(handle, "df_runtime_entity_type_count", error, error_capacity);
    RuntimeEntityTypeAtFn entity_type_at = (RuntimeEntityTypeAtFn) load_symbol(handle, "df_runtime_entity_type_at", error, error_capacity);
    RuntimeEntityAdoptFn entity_adopt = (RuntimeEntityAdoptFn) load_symbol(handle, "df_runtime_entity_adopt", error, error_capacity);
    RuntimeEntityLoadFn entity_load = (RuntimeEntityLoadFn) load_symbol(handle, "df_runtime_entity_load", error, error_capacity);
    RuntimeEntitySaveFn entity_save = (RuntimeEntitySaveFn) load_symbol(handle, "df_runtime_entity_save", error, error_capacity);
    RuntimeEntityTickFn entity_tick = (RuntimeEntityTickFn) load_symbol(handle, "df_runtime_entity_tick", error, error_capacity);
    RuntimeEntityHurtFn entity_hurt = (RuntimeEntityHurtFn) load_symbol(handle, "df_runtime_entity_hurt", error, error_capacity);
    RuntimeEntityHealFn entity_heal = (RuntimeEntityHealFn) load_symbol(handle, "df_runtime_entity_heal", error, error_capacity);
    RuntimeEntityDeathFn entity_death = (RuntimeEntityDeathFn) load_symbol(handle, "df_runtime_entity_death", error, error_capacity);
    RuntimeEntityDestroyFn entity_destroy = (RuntimeEntityDestroyFn) load_symbol(handle, "df_runtime_entity_destroy", error, error_capacity);
    RuntimeCountFn command_count = (RuntimeCountFn) load_symbol(handle, "df_runtime_command_count", error, error_capacity);
    RuntimeCommandAtFn command_at = (RuntimeCommandAtFn) load_symbol(handle, "df_runtime_command_at", error, error_capacity);
    RuntimeCommandFn handle_command = (RuntimeCommandFn) load_symbol(handle, "df_runtime_handle_command", error, error_capacity);
    RuntimeCommandEnumFn command_enum_options = (RuntimeCommandEnumFn) load_symbol(handle, "df_runtime_command_enum_options", error, error_capacity);
    RuntimeEventFn handle_event = (RuntimeEventFn) load_symbol(handle, "df_runtime_handle_event", error, error_capacity);
    if (create == NULL || destroy == NULL || enable == NULL || disable == NULL || plugin_count == NULL || subscriptions == NULL || entity_type_count == NULL || entity_type_at == NULL || entity_adopt == NULL || entity_load == NULL || entity_save == NULL || entity_tick == NULL || entity_hurt == NULL || entity_heal == NULL || entity_death == NULL || entity_destroy == NULL || command_count == NULL || command_at == NULL || handle_command == NULL || command_enum_options == NULL || handle_event == NULL) {
        dlclose(handle);
        return DF_STATUS_ERROR;
    }

    BgRuntimeLibrary *library = calloc(1, sizeof(*library));
    if (library == NULL) {
        write_error(error, error_capacity, "allocate runtime bridge");
        dlclose(handle);
        return DF_STATUS_ERROR;
    }

    library->host_api = (DfHostApiV18) {
        .abi_version = DF_HOST_ABI_VERSION,
        .struct_size = sizeof(DfHostApiV18),
        .context = host_context,
        .player_text = host_player_text,
        .player_title = host_player_title,
        .player_transform = host_player_transform,
        .player_rotation = host_player_rotation,
        .player_state_set = host_player_state_set,
        .player_state_get = host_player_state_get,
        .player_effect = host_player_effect,
        .player_entity_visibility = host_player_entity_visibility,
        .player_skin_open = host_player_skin_open,
        .player_skin_animation_info = host_player_skin_animation_info,
        .player_skin_read = host_player_skin_read,
        .player_skin_close = host_player_skin_close,
        .player_skin_set = host_player_skin_set,
        .inventory_size = host_inventory_size,
        .inventory_item_open = host_inventory_item_open,
        .player_held_item_open = host_player_held_item_open,
        .item_stack_read = host_item_stack_read,
        .item_stack_close = host_item_stack_close,
        .inventory_item_set = host_inventory_item_set,
        .inventory_item_add = host_inventory_item_add,
        .inventory_clear_slot = host_inventory_clear_slot,
        .inventory_clear = host_inventory_clear,
        .player_held_items_set = host_player_held_items_set,
        .player_held_slot_set = host_player_held_slot_set,
        .player_scoreboard = host_player_scoreboard,
        .player_scoreboard_remove = host_player_scoreboard_remove,
        .player_form_send = host_player_form_send,
        .player_form_close = host_player_form_close,
        .world_lookup = host_world_lookup,
        .world_open = host_world_open,
        .world_name = host_world_name,
        .world_unload = host_world_unload,
        .world_save = host_world_save,
        .world_block_get = host_world_block_get,
        .world_block_set = host_world_block_set,
        .world_time_get = host_world_time_get,
        .world_time_set = host_world_time_set,
        .world_spawn_get = host_world_spawn_get,
        .world_spawn_set = host_world_spawn_set,
        .world_entity_spawn = host_world_entity_spawn,
        .world_entities = host_world_entities,
        .world_players = host_world_players,
        .entity_state = host_entity_state,
        .entity_teleport = host_entity_teleport,
        .entity_velocity_set = host_entity_velocity_set,
        .entity_name_tag_set = host_entity_name_tag_set,
        .entity_despawn = host_entity_despawn,
        .world_particle_add = host_world_particle_add,
        .world_sound_play = host_world_sound_play,
        .player_sound_play = host_player_sound_play,
        .player_heal = host_player_heal,
        .player_hurt = host_player_hurt,
        .skin_snapshot_info = host_skin_snapshot_info,
        .skin_snapshot_set = host_skin_snapshot_set,
        .world_open_spec = host_world_open_spec,
        .player_transfer = host_player_transfer,
        .player_effects = host_player_effects,
        .player_effects_clear = host_player_effects_clear,
    };
    DfRuntimeConfig config = {
        .plugin_directory = {
            .data = (const uint8_t *) plugin_directory,
            .len = (uint64_t) strlen(plugin_directory),
        },
        .host = &library->host_api,
    };
    if (create(&config, &library->runtime, error, error_capacity) != DF_STATUS_OK) {
        free(library);
        dlclose(handle);
        return DF_STATUS_ERROR;
    }

    library->handle = handle;
    library->destroy = destroy;
    library->enable = enable;
    library->disable = disable;
    library->plugin_count = plugin_count;
    library->subscriptions = subscriptions;
    library->entity_type_count = entity_type_count;
    library->entity_type_at = entity_type_at;
    library->entity_adopt = entity_adopt;
    library->entity_load = entity_load;
    library->entity_save = entity_save;
    library->entity_tick = entity_tick;
    library->entity_hurt = entity_hurt;
    library->entity_heal = entity_heal;
    library->entity_death = entity_death;
    library->entity_destroy = entity_destroy;
    library->command_count = command_count;
    library->command_at = command_at;
    library->handle_command = handle_command;
    library->command_enum_options = command_enum_options;
    library->handle_event = handle_event;
    *out = library;
    return DF_STATUS_OK;
}

void bg_runtime_close(BgRuntimeLibrary *library) {
    if (library == NULL) {
        return;
    }
    library->destroy(library->runtime);
    dlclose(library->handle);
    free(library);
}

DfStatus bg_runtime_enable(BgRuntimeLibrary *library) {
    return library == NULL ? DF_STATUS_ERROR : library->enable(library->runtime);
}

void bg_runtime_disable(BgRuntimeLibrary *library) {
    if (library != NULL) {
        library->disable(library->runtime);
    }
}

uint64_t bg_runtime_plugin_count(const BgRuntimeLibrary *library) {
    return library == NULL ? 0 : library->plugin_count(library->runtime);
}

uint64_t bg_runtime_subscriptions(const BgRuntimeLibrary *library) {
    return library == NULL ? 0 : library->subscriptions(library->runtime);
}

uint64_t bg_runtime_entity_type_count(const BgRuntimeLibrary *library) {
    return library == NULL ? 0 : library->entity_type_count(library->runtime);
}

DfStatus bg_runtime_entity_type_at(
    const BgRuntimeLibrary *library,
    uint64_t index,
    DfEntityTypeDescriptorV2 *out
) {
    if (library == NULL || out == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->entity_type_at(library->runtime, index, out);
}

DfStatus bg_runtime_entity_adopt(BgRuntimeLibrary *library, uint64_t type_key, uint64_t opaque, DfEntityInstanceId *out) {
    return library == NULL || out == NULL ? DF_STATUS_ERROR : library->entity_adopt(library->runtime, type_key, opaque, out);
}

DfStatus bg_runtime_entity_load(BgRuntimeLibrary *library, uint64_t type_key, const DfEntityLoadInput *input, DfEntityInstanceId *out) {
    return library == NULL || input == NULL || out == NULL ? DF_STATUS_ERROR : library->entity_load(library->runtime, type_key, input, out);
}

DfStatus bg_runtime_entity_save(BgRuntimeLibrary *library, DfEntityInstanceId instance, DfEntitySaveState *state) {
    return library == NULL || state == NULL ? DF_STATUS_ERROR : library->entity_save(library->runtime, instance, state);
}

DfStatus bg_runtime_entity_tick(BgRuntimeLibrary *library, DfEntityInstanceId instance, const DfEntityTickInput *input, DfEntityTickState *state) {
    return library == NULL || input == NULL || state == NULL ? DF_STATUS_ERROR : library->entity_tick(library->runtime, instance, input, state);
}

DfStatus bg_runtime_entity_hurt(BgRuntimeLibrary *library, DfEntityInstanceId instance, const DfEntityHurtInput *input, DfEntityHurtState *state) {
    return library == NULL || input == NULL || state == NULL ? DF_STATUS_ERROR : library->entity_hurt(library->runtime, instance, input, state);
}

DfStatus bg_runtime_entity_heal(BgRuntimeLibrary *library, DfEntityInstanceId instance, const DfEntityHealInput *input, DfEntityHealState *state) {
    return library == NULL || input == NULL || state == NULL ? DF_STATUS_ERROR : library->entity_heal(library->runtime, instance, input, state);
}

DfStatus bg_runtime_entity_death(BgRuntimeLibrary *library, DfEntityInstanceId instance, const DfEntityDeathInput *input, DfEntityDeathState *state) {
    return library == NULL || input == NULL || state == NULL ? DF_STATUS_ERROR : library->entity_death(library->runtime, instance, input, state);
}

DfStatus bg_runtime_entity_destroy(BgRuntimeLibrary *library, DfEntityInstanceId instance) {
    return library == NULL ? DF_STATUS_ERROR : library->entity_destroy(library->runtime, instance);
}

uint64_t bg_runtime_command_count(const BgRuntimeLibrary *library) {
    return library == NULL ? 0 : library->command_count(library->runtime);
}

DfStatus bg_runtime_command_at(
    const BgRuntimeLibrary *library,
    uint64_t index,
    DfCommandDescriptor *out
) {
    if (library == NULL || out == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->command_at(library->runtime, index, out);
}

DfStatus bg_runtime_handle_command(
    BgRuntimeLibrary *library,
    uint64_t index,
    const DfCommandInput *input,
    DfCommandState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->handle_command(library->runtime, index, input, state);
}

DfStatus bg_runtime_command_enum_options(
    BgRuntimeLibrary *library,
    uint64_t index,
    uint64_t overload,
    uint64_t parameter,
    const DfCommandEnumContext *context,
    DfStringBuffer *output
) {
    if (library == NULL || context == NULL || output == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->command_enum_options(library->runtime, index, overload, parameter, context, output);
}

DfStatus bg_runtime_handle_event(BgRuntimeLibrary *library, DfEventId event_id, const void *input, void *state) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->handle_event(library->runtime, event_id, input, state);
}

DfStatus bg_runtime_handle_player_move(
    BgRuntimeLibrary *library,
    const DfPlayerMoveInput *input,
    DfPlayerMoveState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return bg_runtime_handle_event(library, DF_EVENT_PLAYER_MOVE, input, state);
}

DfStatus bg_runtime_handle_player_chat(
    BgRuntimeLibrary *library,
    const DfPlayerChatInput *input,
    DfPlayerChatState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return bg_runtime_handle_event(library, DF_EVENT_PLAYER_CHAT, input, state);
}

DfStatus bg_runtime_handle_player_join(
    BgRuntimeLibrary *library,
    const DfPlayerJoinInput *input,
    DfPlayerJoinState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return bg_runtime_handle_event(library, DF_EVENT_PLAYER_JOIN, input, state);
}

DfStatus bg_runtime_handle_player_quit(
    BgRuntimeLibrary *library,
    const DfPlayerQuitInput *input,
    DfPlayerQuitState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return bg_runtime_handle_event(library, DF_EVENT_PLAYER_QUIT, input, state);
}

DfStatus bg_runtime_handle_player_hurt(
    BgRuntimeLibrary *library,
    const DfPlayerHurtInput *input,
    DfPlayerHurtState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return bg_runtime_handle_event(library, DF_EVENT_PLAYER_HURT, input, state);
}

DfStatus bg_runtime_handle_player_heal(
    BgRuntimeLibrary *library,
    const DfPlayerHealInput *input,
    DfPlayerHealState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return bg_runtime_handle_event(library, DF_EVENT_PLAYER_HEAL, input, state);
}

DfStatus bg_runtime_handle_player_block_break(
    BgRuntimeLibrary *library,
    const DfPlayerBlockBreakInput *input,
    DfPlayerBlockBreakState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return bg_runtime_handle_event(library, DF_EVENT_PLAYER_BLOCK_BREAK, input, state);
}

DfStatus bg_runtime_handle_player_block_place(
    BgRuntimeLibrary *library,
    const DfPlayerBlockPlaceInput *input,
    DfPlayerBlockPlaceState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return bg_runtime_handle_event(library, DF_EVENT_PLAYER_BLOCK_PLACE, input, state);
}

DfStatus bg_runtime_handle_player_food_loss(
    BgRuntimeLibrary *library,
    const DfPlayerFoodLossInput *input,
    DfPlayerFoodLossState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return bg_runtime_handle_event(library, DF_EVENT_PLAYER_FOOD_LOSS, input, state);
}

DfStatus bg_runtime_handle_player_death(
    BgRuntimeLibrary *library,
    const DfPlayerDeathInput *input,
    DfPlayerDeathState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return bg_runtime_handle_event(library, DF_EVENT_PLAYER_DEATH, input, state);
}

DfStatus bg_runtime_handle_player_start_break(BgRuntimeLibrary *library, const DfPlayerStartBreakInput *input, DfPlayerStartBreakState *state) {
    if (library == NULL || input == NULL || state == NULL) return DF_STATUS_ERROR;
    return bg_runtime_handle_event(library, DF_EVENT_PLAYER_START_BREAK, input, state);
}

DfStatus bg_runtime_handle_player_fire_extinguish(BgRuntimeLibrary *library, const DfPlayerFireExtinguishInput *input, DfPlayerFireExtinguishState *state) {
    if (library == NULL || input == NULL || state == NULL) return DF_STATUS_ERROR;
    return bg_runtime_handle_event(library, DF_EVENT_PLAYER_FIRE_EXTINGUISH, input, state);
}
DfStatus bg_runtime_handle_player_toggle_sprint(BgRuntimeLibrary *library, const DfPlayerToggleSprintInput *input, DfPlayerToggleSprintState *state) { return bg_runtime_handle_event(library, DF_EVENT_PLAYER_TOGGLE_SPRINT, input, state); }
DfStatus bg_runtime_handle_player_toggle_sneak(BgRuntimeLibrary *library, const DfPlayerToggleSneakInput *input, DfPlayerToggleSneakState *state) { return bg_runtime_handle_event(library, DF_EVENT_PLAYER_TOGGLE_SNEAK, input, state); }
DfStatus bg_runtime_handle_player_jump(BgRuntimeLibrary *library, const DfPlayerJumpInput *input, DfPlayerJumpState *state) { return bg_runtime_handle_event(library, DF_EVENT_PLAYER_JUMP, input, state); }
DfStatus bg_runtime_handle_player_teleport(BgRuntimeLibrary *library, const DfPlayerTeleportInput *input, DfPlayerTeleportState *state) { return bg_runtime_handle_event(library, DF_EVENT_PLAYER_TELEPORT, input, state); }
DfStatus bg_runtime_handle_player_experience_gain(BgRuntimeLibrary *library, const DfPlayerExperienceGainInput *input, DfPlayerExperienceGainState *state) { return bg_runtime_handle_event(library, DF_EVENT_PLAYER_EXPERIENCE_GAIN, input, state); }
DfStatus bg_runtime_handle_player_punch_air(BgRuntimeLibrary *library, const DfPlayerPunchAirInput *input, DfPlayerPunchAirState *state) { return bg_runtime_handle_event(library, DF_EVENT_PLAYER_PUNCH_AIR, input, state); }

uint64_t bg_runtime_handle_player_move_value(
    BgRuntimeLibrary *library,
    DfPlayerMoveInput input,
    uint8_t cancelled
) {
    DfPlayerMoveState state = {.cancelled = cancelled};
    DfStatus status = library == NULL
        ? DF_STATUS_ERROR
        : library->handle_event(library->runtime, DF_EVENT_PLAYER_MOVE, &input, &state);
    return ((uint64_t) (uint32_t) status << 32u) | (uint64_t) state.cancelled;
}
