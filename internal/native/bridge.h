#ifndef BEDROCK_GOPHERS_BRIDGE_H
#define BEDROCK_GOPHERS_BRIDGE_H

#include "dragonfly_plugin.h"

typedef struct BgRuntimeLibrary BgRuntimeLibrary;
DfStatus bg_call_form_response(DfFormResponseFn callback, void *callback_context, DfInvocationId invocation, const DfPlayerSnapshot *submitter, uint32_t outcome, DfStringView response_json);
void bg_call_form_drop(DfFormDropFn callback, void *callback_context);

DfStatus bg_runtime_open(
    const char *library_path,
    const char *plugin_directory,
    uint64_t host_context,
    BgRuntimeLibrary **out,
    uint8_t *error,
    uint64_t error_capacity
);

void bg_runtime_close(BgRuntimeLibrary *library);
DfStatus bg_runtime_enable(BgRuntimeLibrary *library, uint8_t *error, uint64_t error_capacity);
void bg_runtime_begin_disable(BgRuntimeLibrary *library);
void bg_runtime_finish_disable(BgRuntimeLibrary *library);
void bg_runtime_disable(BgRuntimeLibrary *library);
uint64_t bg_runtime_plugin_count(const BgRuntimeLibrary *library);
uint64_t bg_runtime_subscriptions(const BgRuntimeLibrary *library);
uint64_t bg_runtime_entity_type_count(const BgRuntimeLibrary *library);
DfStatus bg_runtime_entity_type_at(
    const BgRuntimeLibrary *library,
    uint64_t index,
    DfEntityTypeDescriptorV2 *out
);
DfStatus bg_runtime_entity_adopt(BgRuntimeLibrary *library, uint64_t type_key, uint64_t opaque, DfEntityInstanceId *out);
DfStatus bg_runtime_entity_load(BgRuntimeLibrary *library, uint64_t type_key, const DfEntityLoadInput *input, DfEntityInstanceId *out);
DfStatus bg_runtime_entity_save(BgRuntimeLibrary *library, DfEntityInstanceId instance, DfEntitySaveState *state);
DfStatus bg_runtime_entity_tick(BgRuntimeLibrary *library, DfEntityInstanceId instance, const DfEntityTickInput *input, DfEntityTickState *state);
DfStatus bg_runtime_entity_hurt(BgRuntimeLibrary *library, DfEntityInstanceId instance, const DfEntityHurtInput *input, DfEntityHurtState *state);
DfStatus bg_runtime_entity_heal(BgRuntimeLibrary *library, DfEntityInstanceId instance, const DfEntityHealInput *input, DfEntityHealState *state);
DfStatus bg_runtime_entity_death(BgRuntimeLibrary *library, DfEntityInstanceId instance, const DfEntityDeathInput *input, DfEntityDeathState *state);
DfStatus bg_runtime_entity_destroy(BgRuntimeLibrary *library, DfEntityInstanceId instance);
uint64_t bg_runtime_command_count(const BgRuntimeLibrary *library);
DfStatus bg_runtime_command_at(
    const BgRuntimeLibrary *library,
    uint64_t index,
    DfCommandDescriptor *out
);
DfStatus bg_runtime_handle_command(
    BgRuntimeLibrary *library,
    uint64_t index,
    const DfCommandInput *input,
    DfCommandState *state
);
DfStatus bg_runtime_command_enum_options(
    BgRuntimeLibrary *library,
    uint64_t index,
    uint64_t overload,
    uint64_t parameter,
    const DfCommandEnumContext *context,
    DfStringBuffer *output
);
DfStatus bg_runtime_handle_event(
    BgRuntimeLibrary *library,
    DfEventId event_id,
    const void *input,
    void *state
);
DfStatus bg_runtime_handle_player_move(
    BgRuntimeLibrary *library,
    const DfPlayerMoveInput *input,
    DfPlayerMoveState *state
);
DfStatus bg_runtime_handle_player_chat(
    BgRuntimeLibrary *library,
    const DfPlayerChatInput *input,
    DfPlayerChatState *state
);
DfStatus bg_runtime_handle_player_join(
    BgRuntimeLibrary *library,
    const DfPlayerJoinInput *input,
    DfPlayerJoinState *state
);
DfStatus bg_runtime_handle_player_quit(
    BgRuntimeLibrary *library,
    const DfPlayerQuitInput *input,
    DfPlayerQuitState *state
);
DfStatus bg_runtime_handle_player_hurt(
    BgRuntimeLibrary *library,
    const DfPlayerHurtInput *input,
    DfPlayerHurtState *state
);
DfStatus bg_runtime_handle_player_heal(
    BgRuntimeLibrary *library,
    const DfPlayerHealInput *input,
    DfPlayerHealState *state
);
DfStatus bg_runtime_handle_player_block_break(
    BgRuntimeLibrary *library,
    const DfPlayerBlockBreakInput *input,
    DfPlayerBlockBreakState *state
);
DfStatus bg_runtime_handle_player_block_place(
    BgRuntimeLibrary *library,
    const DfPlayerBlockPlaceInput *input,
    DfPlayerBlockPlaceState *state
);
DfStatus bg_runtime_handle_player_food_loss(
    BgRuntimeLibrary *library,
    const DfPlayerFoodLossInput *input,
    DfPlayerFoodLossState *state
);
DfStatus bg_runtime_handle_player_death(
    BgRuntimeLibrary *library,
    const DfPlayerDeathInput *input,
    DfPlayerDeathState *state
);
DfStatus bg_runtime_handle_player_start_break(BgRuntimeLibrary *library, const DfPlayerStartBreakInput *input, DfPlayerStartBreakState *state);
DfStatus bg_runtime_handle_player_fire_extinguish(BgRuntimeLibrary *library, const DfPlayerFireExtinguishInput *input, DfPlayerFireExtinguishState *state);
DfStatus bg_runtime_handle_player_toggle_sprint(BgRuntimeLibrary *library, const DfPlayerToggleSprintInput *input, DfPlayerToggleSprintState *state);
DfStatus bg_runtime_handle_player_toggle_sneak(BgRuntimeLibrary *library, const DfPlayerToggleSneakInput *input, DfPlayerToggleSneakState *state);
DfStatus bg_runtime_handle_player_jump(BgRuntimeLibrary *library, const DfPlayerJumpInput *input, DfPlayerJumpState *state);
DfStatus bg_runtime_handle_player_teleport(BgRuntimeLibrary *library, const DfPlayerTeleportInput *input, DfPlayerTeleportState *state);
DfStatus bg_runtime_handle_player_experience_gain(BgRuntimeLibrary *library, const DfPlayerExperienceGainInput *input, DfPlayerExperienceGainState *state);
DfStatus bg_runtime_handle_player_punch_air(BgRuntimeLibrary *library, const DfPlayerPunchAirInput *input, DfPlayerPunchAirState *state);
uint64_t bg_runtime_handle_player_move_value(
    BgRuntimeLibrary *library,
    DfPlayerMoveInput input,
    uint8_t cancelled
);

#endif
