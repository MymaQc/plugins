#include "bridge.h"

#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#ifdef _WIN32
#include <windows.h>

#define RTLD_NOW 0
#define RTLD_LOCAL 0

static char bg_dl_error[512];

static const char *dlerror(void) {
    DWORD code = GetLastError();
    if (code == 0) {
        return NULL;
    }

    DWORD written = FormatMessageA(
        FORMAT_MESSAGE_FROM_SYSTEM | FORMAT_MESSAGE_IGNORE_INSERTS,
        NULL,
        code,
        MAKELANGID(LANG_NEUTRAL, SUBLANG_DEFAULT),
        bg_dl_error,
        sizeof(bg_dl_error),
        NULL
    );
    if (written == 0) {
        snprintf(bg_dl_error, sizeof(bg_dl_error), "Windows error %lu", (unsigned long) code);
    }

    SetLastError(0);
    return bg_dl_error;
}

static void *dlopen(const char *path, int flags) {
    (void) flags;
    SetLastError(0);
    return (void *) LoadLibraryA(path);
}

static void *dlsym(void *handle, const char *name) {
    SetLastError(0);
    return (void *) GetProcAddress((HMODULE) handle, name);
}

static int dlclose(void *handle) {
    return FreeLibrary((HMODULE) handle) ? 0 : 1;
}
#else
#include <dlfcn.h>
#endif

#if UINTPTR_MAX == UINT64_MAX
#define ASSERT_PLAYER_EVENT_LAYOUT(TYPE) _Static_assert(offsetof(TYPE, player) == 8, #TYPE ".player ABI offset changed")
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerMoveInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerChatInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerJoinInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerQuitInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerHurtInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerHealInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerBlockBreakInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerBlockPlaceInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerFoodLossInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerDeathInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerStartBreakInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerFireExtinguishInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerToggleSprintInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerToggleSneakInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerJumpInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerTeleportInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerExperienceGainInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerPunchAirInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerHeldSlotChangeInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerSleepInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerBlockPickInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerLecternPageTurnInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerSignEditInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerItemUseInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerItemUseOnBlockInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerItemConsumeInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerItemReleaseInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerItemDamageInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerItemDropInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerAttackEntityInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerItemUseOnEntityInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerChangeWorldInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerRespawnInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerSkinChangeInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerItemPickupInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerTransferInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerCommandExecutionInput);
ASSERT_PLAYER_EVENT_LAYOUT(DfPlayerDiagnosticsInput);
#undef ASSERT_PLAYER_EVENT_LAYOUT
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
_Static_assert(sizeof(DfPlayerMoveInput) == 144, "DfPlayerMoveInput ABI layout changed");
_Static_assert(offsetof(DfPlayerMoveInput, player) == 8, "DfPlayerMoveInput.player ABI offset changed");
_Static_assert(sizeof(DfPlayerBlockBreakInput) == 144, "DfPlayerBlockBreakInput ABI layout changed");
_Static_assert(offsetof(DfPlayerBlockBreakInput, block) == 96, "DfPlayerBlockBreakInput.block ABI offset changed");
_Static_assert(offsetof(DfPlayerBlockBreakInput, drops) == 128, "DfPlayerBlockBreakInput.drops ABI offset changed");
_Static_assert(sizeof(DfPlayerBlockBreakState) == 40, "DfPlayerBlockBreakState ABI layout changed");
_Static_assert(offsetof(DfPlayerBlockBreakState, replacement_drops) == 8, "DfPlayerBlockBreakState.replacement_drops ABI offset changed");
_Static_assert(sizeof(DfPlayerBlockPlaceInput) == 128, "DfPlayerBlockPlaceInput ABI layout changed");
_Static_assert(offsetof(DfPlayerBlockPlaceInput, block) == 96, "DfPlayerBlockPlaceInput.block ABI offset changed");
_Static_assert(sizeof(DfPlayerItemConsumeInput) == 200, "DfPlayerItemConsumeInput ABI layout changed");
_Static_assert(offsetof(DfPlayerItemConsumeInput, item) == 80, "DfPlayerItemConsumeInput.item ABI offset changed");
_Static_assert(sizeof(DfPlayerItemReleaseInput) == 208, "DfPlayerItemReleaseInput ABI layout changed");
_Static_assert(offsetof(DfPlayerItemReleaseInput, duration_nanoseconds) == 200, "DfPlayerItemReleaseInput.duration_nanoseconds ABI offset changed");
_Static_assert(sizeof(DfPlayerItemDamageInput) == 200, "DfPlayerItemDamageInput ABI layout changed");
_Static_assert(offsetof(DfPlayerItemDamageInput, item) == 80, "DfPlayerItemDamageInput.item ABI offset changed");
_Static_assert(sizeof(DfPlayerItemDropInput) == 200, "DfPlayerItemDropInput ABI layout changed");
_Static_assert(offsetof(DfPlayerItemDropInput, item) == 80, "DfPlayerItemDropInput.item ABI offset changed");
_Static_assert(sizeof(DfPlayerSkinChangeInput) == 88, "DfPlayerSkinChangeInput ABI layout changed");
_Static_assert(offsetof(DfPlayerSkinChangeInput, snapshot) == 80, "DfPlayerSkinChangeInput.snapshot ABI offset changed");
_Static_assert(sizeof(DfPlayerSkinChangeState) == 1, "DfPlayerSkinChangeState ABI layout changed");
_Static_assert(sizeof(DfPlayerItemPickupInput) == 200, "DfPlayerItemPickupInput ABI layout changed");
_Static_assert(offsetof(DfPlayerItemPickupInput, item) == 80, "DfPlayerItemPickupInput.item ABI offset changed");
_Static_assert(sizeof(DfPlayerItemPickupState) == 32, "DfPlayerItemPickupState ABI layout changed");
_Static_assert(offsetof(DfPlayerItemPickupState, replacement) == 8, "DfPlayerItemPickupState.replacement ABI offset changed");
_Static_assert(sizeof(DfPlayerHurtState) == 24, "DfPlayerHurtState ABI layout changed");
_Static_assert(offsetof(DfPlayerHurtState, attack_immunity_nanoseconds) == 16, "DfPlayerHurtState.attack_immunity_nanoseconds ABI offset changed");
_Static_assert(_Generic(((DfPlayerHurtState *)0)->attack_immunity_nanoseconds, int64_t: 1, default: 0), "DfPlayerHurtState.attack_immunity_nanoseconds must be signed");
_Static_assert(sizeof(DfUDPAddrView) == 40, "DfUDPAddrView ABI layout changed");
_Static_assert(offsetof(DfUDPAddrView, port) == 16, "DfUDPAddrView.port ABI offset changed");
_Static_assert(offsetof(DfUDPAddrView, zone) == 24, "DfUDPAddrView.zone ABI offset changed");
_Static_assert(sizeof(DfPlayerTransferInput) == 80, "DfPlayerTransferInput ABI layout changed");
_Static_assert(sizeof(DfPlayerTransferState) == 64, "DfPlayerTransferState ABI layout changed");
_Static_assert(offsetof(DfPlayerTransferState, address) == 8, "DfPlayerTransferState.address ABI offset changed");
_Static_assert(offsetof(DfPlayerTransferState, replacement_drop) == 56, "DfPlayerTransferState.replacement_drop ABI offset changed");
_Static_assert(sizeof(DfPlayerCommandExecutionInput) == 160, "DfPlayerCommandExecutionInput ABI layout changed");
_Static_assert(offsetof(DfPlayerCommandExecutionInput, command_name) == 80, "DfPlayerCommandExecutionInput.command_name ABI offset changed");
_Static_assert(offsetof(DfPlayerCommandExecutionInput, command_aliases) == 128, "DfPlayerCommandExecutionInput.command_aliases ABI offset changed");
_Static_assert(offsetof(DfPlayerCommandExecutionInput, arguments) == 144, "DfPlayerCommandExecutionInput.arguments ABI offset changed");
_Static_assert(sizeof(DfPlayerCommandExecutionState) == 40, "DfPlayerCommandExecutionState ABI layout changed");
_Static_assert(offsetof(DfPlayerCommandExecutionState, replacement_arguments) == 8, "DfPlayerCommandExecutionState.replacement_arguments ABI offset changed");
_Static_assert(offsetof(DfPlayerCommandExecutionState, replacement_drop) == 32, "DfPlayerCommandExecutionState.replacement_drop ABI offset changed");
_Static_assert(sizeof(DfPlayerDiagnosticsInput) == 152, "DfPlayerDiagnosticsInput ABI layout changed");
_Static_assert(offsetof(DfPlayerDiagnosticsInput, average_frames_per_second) == 80, "DfPlayerDiagnosticsInput.average_frames_per_second ABI offset changed");
_Static_assert(offsetof(DfPlayerDiagnosticsInput, average_unaccounted_time_percent) == 144, "DfPlayerDiagnosticsInput.average_unaccounted_time_percent ABI offset changed");
_Static_assert(sizeof(DfPlayerDiagnosticsState) == 1, "DfPlayerDiagnosticsState ABI layout changed");
_Static_assert(sizeof(DfCommandParameter) == 56, "DfCommandParameter ABI layout changed");
_Static_assert(offsetof(DfCommandParameter, suffix) == 24, "DfCommandParameter.suffix ABI offset changed");
_Static_assert(sizeof(DfCommandDescriptor) == 64, "DfCommandDescriptor ABI layout changed");
_Static_assert(sizeof(DfCommandPlayer) == 72, "DfCommandPlayer ABI layout changed");
_Static_assert(sizeof(DfCommandEnumContext) == 88, "DfCommandEnumContext ABI layout changed");
_Static_assert(sizeof(DfCommandInput) == 120, "DfCommandInput ABI layout changed");
_Static_assert(sizeof(DfCommandState) == 32, "DfCommandState ABI layout changed");
_Static_assert(sizeof(DfAllowInput) == 104, "DfAllowInput ABI layout changed");
_Static_assert(offsetof(DfAllowInput, identity_json) == 64, "DfAllowInput.identity_json ABI offset changed");
_Static_assert(offsetof(DfAllowInput, port) == 96, "DfAllowInput.port ABI offset changed");
_Static_assert(sizeof(DfPluginApiV11) == 144, "DfPluginApiV11 ABI layout changed");
_Static_assert(sizeof(DfInventoryId) == 32, "DfInventoryId ABI layout changed");
_Static_assert(sizeof(DfItemStackInfo) == 80, "DfItemStackInfo ABI layout changed");
_Static_assert(sizeof(DfItemStackSnapshot) == 88, "DfItemStackSnapshot ABI layout changed");
_Static_assert(sizeof(DfItemStackData) == 152, "DfItemStackData ABI layout changed");
_Static_assert(sizeof(DfItemStackViewV3) == 120, "DfItemStackViewV3 ABI layout changed");
_Static_assert(sizeof(DfItemStackSnapshot) == 88, "DfItemStackSnapshot ABI layout changed");
_Static_assert(sizeof(DfScoreboardView) == 40, "DfScoreboardView ABI layout changed");
_Static_assert(offsetof(DfScoreboardView, lines) == 16, "DfScoreboardView.lines ABI offset changed");
_Static_assert(sizeof(DfPlayerSnapshot) == 72, "DfPlayerSnapshot ABI layout changed");
_Static_assert(offsetof(DfPlayerSnapshot, name) == 24, "DfPlayerSnapshot.name ABI offset changed");
_Static_assert(offsetof(DfPlayerSnapshot, latency_milliseconds) == 40, "DfPlayerSnapshot.latency_milliseconds ABI offset changed");
_Static_assert(offsetof(DfPlayerSnapshot, position) == 48, "DfPlayerSnapshot.position ABI offset changed");
_Static_assert(sizeof(DfPlayerSnapshotBuffer) == 80, "DfPlayerSnapshotBuffer ABI layout changed");
_Static_assert(offsetof(DfPlayerSnapshotBuffer, name) == 24, "DfPlayerSnapshotBuffer.name ABI offset changed");
_Static_assert(offsetof(DfPlayerSnapshotBuffer, latency_milliseconds) == 48, "DfPlayerSnapshotBuffer.latency_milliseconds ABI offset changed");
_Static_assert(offsetof(DfPlayerSnapshotBuffer, position) == 56, "DfPlayerSnapshotBuffer.position ABI offset changed");
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
_Static_assert(sizeof(DfEffectView) == 40, "DfEffectView ABI layout changed");
_Static_assert(offsetof(DfEffectView, duration_nanoseconds) == 8, "DfEffectView.duration_nanoseconds ABI offset changed");
_Static_assert(offsetof(DfEffectView, potency) == 16, "DfEffectView.potency ABI offset changed");
_Static_assert(offsetof(DfEffectView, ambient) == 24, "DfEffectView.ambient ABI offset changed");
_Static_assert(offsetof(DfEffectView, particles_hidden) == 25, "DfEffectView.particles_hidden ABI offset changed");
_Static_assert(offsetof(DfEffectView, infinite) == 26, "DfEffectView.infinite ABI offset changed");
_Static_assert(offsetof(DfEffectView, tick) == 32, "DfEffectView.tick ABI offset changed");
_Static_assert(sizeof(DfEffectBuffer) == 24, "DfEffectBuffer ABI layout changed");
_Static_assert(offsetof(DfEffectBuffer, data) == 0, "DfEffectBuffer.data ABI offset changed");
_Static_assert(offsetof(DfEffectBuffer, len) == 8, "DfEffectBuffer.len ABI offset changed");
_Static_assert(offsetof(DfEffectBuffer, capacity) == 16, "DfEffectBuffer.capacity ABI offset changed");
_Static_assert(sizeof(DfEntitySpawnOptions) == 80, "DfEntitySpawnOptions ABI layout changed");
_Static_assert(sizeof(DfEntityTypeDescriptorV2) == 40, "DfEntityTypeDescriptorV2 ABI layout changed");
_Static_assert(offsetof(DfEntityTypeDescriptorV2, type_key) == 32, "DfEntityTypeDescriptorV2.type_key ABI offset changed");
_Static_assert(sizeof(DfEntitySpawnViewV3) == 200, "DfEntitySpawnViewV3 ABI layout changed");
_Static_assert(offsetof(DfEntitySpawnViewV3, custom_instance) == 176, "DfEntitySpawnViewV3.custom_instance ABI offset changed");
_Static_assert(sizeof(DfPluginApiV11) == 144, "DfPluginApiV11 ABI layout changed");
_Static_assert(offsetof(DfPluginApiV11, entity_type_count) == 64, "DfPluginApiV11.entity_type_count ABI offset changed");
_Static_assert(offsetof(DfPluginApiV11, handle_entity) == 80, "DfPluginApiV11.handle_entity ABI offset changed");
_Static_assert(offsetof(DfPluginApiV11, handle_event) == 120, "DfPluginApiV11.handle_event ABI offset changed");
_Static_assert(offsetof(DfPluginApiV11, handle_scheduled) == 128, "DfPluginApiV11.handle_scheduled ABI offset changed");
_Static_assert(offsetof(DfPluginApiV11, allow) == 136, "DfPluginApiV11.allow ABI offset changed");
_Static_assert(DF_ABI_VERSION == 11u, "plugin ABI version changed without bridge review");
_Static_assert(sizeof(DfEntityState) == 128, "DfEntityState ABI layout changed");
_Static_assert(offsetof(DfEntityState, world) == 72, "DfEntityState.world ABI offset changed");
_Static_assert(sizeof(DfPlayerKinematics) == 64, "DfPlayerKinematics ABI layout changed");
_Static_assert(offsetof(DfPlayerKinematics, rotation) == 48, "DfPlayerKinematics.rotation ABI offset changed");
_Static_assert(DF_PLAYER_TRANSFORM_DISPLACE == 3u, "player transform ABI changed");
_Static_assert(sizeof(DfParticleViewV1) == 40, "DfParticleViewV1 ABI layout changed");
_Static_assert(offsetof(DfParticleViewV1, block) == 32, "DfParticleViewV1.block ABI offset changed");
_Static_assert(sizeof(DfSoundViewV1) == 40, "DfSoundViewV1 ABI layout changed");
_Static_assert(offsetof(DfSoundViewV1, scalar) == 16, "DfSoundViewV1.scalar ABI offset changed");
_Static_assert(offsetof(DfSoundViewV1, item) == 32, "DfSoundViewV1.item ABI offset changed");
_Static_assert(sizeof(DfWorldCancellableState) == 1, "DfWorldCancellableState ABI layout changed");
_Static_assert(sizeof(DfWorldNotificationState) == 1, "DfWorldNotificationState ABI layout changed");
_Static_assert(sizeof(DfWorldLiquidFlowInput) == 96, "DfWorldLiquidFlowInput ABI layout changed");
_Static_assert(offsetof(DfWorldLiquidFlowInput, liquid) == 32, "DfWorldLiquidFlowInput.liquid ABI offset changed");
_Static_assert(sizeof(DfWorldLiquidDecayInput) == 64, "DfWorldLiquidDecayInput ABI layout changed");
_Static_assert(offsetof(DfWorldLiquidDecayInput, before) == 24, "DfWorldLiquidDecayInput.before ABI offset changed");
_Static_assert(offsetof(DfWorldLiquidDecayInput, after) == 56, "DfWorldLiquidDecayInput.after ABI offset changed");
_Static_assert(sizeof(DfWorldLiquidHardenInput) == 120, "DfWorldLiquidHardenInput ABI layout changed");
_Static_assert(offsetof(DfWorldLiquidHardenInput, liquid_hardened) == 24, "DfWorldLiquidHardenInput.liquid_hardened ABI offset changed");
_Static_assert(sizeof(DfWorldSoundInput) == 72, "DfWorldSoundInput ABI layout changed");
_Static_assert(offsetof(DfWorldSoundInput, sound) == 8, "DfWorldSoundInput.sound ABI offset changed");
_Static_assert(offsetof(DfWorldSoundInput, position) == 48, "DfWorldSoundInput.position ABI offset changed");
_Static_assert(sizeof(DfWorldFireSpreadInput) == 32, "DfWorldFireSpreadInput ABI layout changed");
_Static_assert(sizeof(DfWorldPositionInput) == 24, "DfWorldPositionInput ABI layout changed");
_Static_assert(sizeof(DfWorldEntityInput) == 32, "DfWorldEntityInput ABI layout changed");
_Static_assert(sizeof(DfWorldExplosionInput) == 64, "DfWorldExplosionInput ABI layout changed");
_Static_assert(offsetof(DfWorldExplosionInput, entities) == 32, "DfWorldExplosionInput.entities ABI offset changed");
_Static_assert(offsetof(DfWorldExplosionInput, blocks) == 48, "DfWorldExplosionInput.blocks ABI offset changed");
_Static_assert(sizeof(DfWorldExplosionState) == 64, "DfWorldExplosionState ABI layout changed");
_Static_assert(offsetof(DfWorldExplosionState, item_drop_chance) == 8, "DfWorldExplosionState.item_drop_chance ABI offset changed");
_Static_assert(offsetof(DfWorldExplosionState, replacement_entities) == 16, "DfWorldExplosionState.replacement_entities ABI offset changed");
_Static_assert(offsetof(DfWorldExplosionState, replacement_blocks) == 32, "DfWorldExplosionState.replacement_blocks ABI offset changed");
_Static_assert(offsetof(DfWorldExplosionState, replacement_drop) == 56, "DfWorldExplosionState.replacement_drop ABI offset changed");
_Static_assert(sizeof(DfWorldRedstoneUpdateInput) == 120, "DfWorldRedstoneUpdateInput ABI layout changed");
_Static_assert(offsetof(DfWorldRedstoneUpdateInput, has_changed_neighbour) == 32, "DfWorldRedstoneUpdateInput.has_changed_neighbour ABI offset changed");
_Static_assert(offsetof(DfWorldRedstoneUpdateInput, source) == 36, "DfWorldRedstoneUpdateInput.source ABI offset changed");
_Static_assert(offsetof(DfWorldRedstoneUpdateInput, before) == 56, "DfWorldRedstoneUpdateInput.before ABI offset changed");
_Static_assert(offsetof(DfWorldRedstoneUpdateInput, after) == 88, "DfWorldRedstoneUpdateInput.after ABI offset changed");
_Static_assert(offsetof(DfWorldRedstoneUpdateInput, current_tick) == 104, "DfWorldRedstoneUpdateInput.current_tick ABI offset changed");
_Static_assert(offsetof(DfWorldRedstoneUpdateInput, cause) == 112, "DfWorldRedstoneUpdateInput.cause ABI offset changed");
_Static_assert(sizeof(DfWorldCloseInput) == 8, "DfWorldCloseInput ABI layout changed");
_Static_assert(sizeof(DfBlockRange) == 8, "DfBlockRange ABI layout changed");
_Static_assert(sizeof(DfEntityHandleId) == 16, "DfEntityHandleId ABI layout changed");
_Static_assert(sizeof(DfUuid) == 16, "DfUuid ABI layout changed");
_Static_assert(sizeof(DfPacketFieldValue) == 96, "DfPacketFieldValue ABI layout changed");
_Static_assert(offsetof(DfPacketFieldValue, data) == 72, "DfPacketFieldValue.data ABI offset changed");
_Static_assert(sizeof(DfPacketInput) == 32, "DfPacketInput ABI layout changed");
_Static_assert(sizeof(DfPacketState) == 1, "DfPacketState ABI layout changed");
_Static_assert(sizeof(DfTitleView) == 72, "DfTitleView ABI layout changed");
_Static_assert(offsetof(DfTitleView, fade_in_duration_nanoseconds) == 48, "DfTitleView.fade_in_duration_nanoseconds ABI offset changed");
_Static_assert(offsetof(DfTitleView, fade_out_duration_nanoseconds) == 64, "DfTitleView.fade_out_duration_nanoseconds ABI offset changed");
_Static_assert(DF_PLAYER_COOLDOWN_HAS == 0u, "player cooldown has operation changed");
_Static_assert(DF_PLAYER_COOLDOWN_SET == 1u, "player cooldown set operation changed");
_Static_assert(sizeof(DfDifficultyView) == 24, "DfDifficultyView ABI layout changed");
_Static_assert(offsetof(DfDifficultyView, starvation_health_limit) == 8, "DfDifficultyView.starvation_health_limit ABI offset changed");
_Static_assert(sizeof(DfHostApiV27) == 1040, "DfHostApiV27 ABI layout changed");
_Static_assert(DF_HOST_ABI_VERSION == 58u, "host ABI version changed without bridge review");
_Static_assert(offsetof(DfHostApiV27, player_skin_open) == 80, "DfHostApiV27.player_skin_open ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_skin_set) == 112, "DfHostApiV27.player_skin_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, inventory_size) == 120, "DfHostApiV27.inventory_size ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_held_slot_set) == 200, "DfHostApiV27.player_held_slot_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_scoreboard) == 208, "DfHostApiV27.player_scoreboard ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_heal) == 384, "DfHostApiV27.player_heal ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, server_players_open) == 720, "DfHostApiV27.server_players_open ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, server_player_by_name) == 752, "DfHostApiV27.server_player_by_name ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, server_max_player_count) == 760, "DfHostApiV27.server_max_player_count ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_xuid) == 784, "DfHostApiV27.player_xuid ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, server_world) == 792, "DfHostApiV27.server_world ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_schedule) == 800, "DfHostApiV27.world_schedule ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_new) == 808, "DfHostApiV27.world_new ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_entities_within_open) == 816, "DfHostApiV27.world_entities_within_open ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, block_by_name) == 824, "DfHostApiV27.block_by_name ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, entity_new) == 832, "DfHostApiV27.entity_new ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, entity_handle_type) == 840, "DfHostApiV27.entity_handle_type ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_task_cancel) == 848, "DfHostApiV27.world_task_cancel ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, packet_field_get) == 856, "DfHostApiV27.packet_field_get ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, packet_field_set) == 864, "DfHostApiV27.packet_field_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_dimension_get) == 872, "DfHostApiV27.world_dimension_get ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_time_cycle_get) == 880, "DfHostApiV27.world_time_cycle_get ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_time_cycle_set) == 888, "DfHostApiV27.world_time_cycle_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_required_sleep_duration_set) == 896, "DfHostApiV27.world_required_sleep_duration_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_default_game_mode_get) == 904, "DfHostApiV27.world_default_game_mode_get ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_default_game_mode_set) == 912, "DfHostApiV27.world_default_game_mode_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_tick_range_set) == 920, "DfHostApiV27.world_tick_range_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_difficulty_get) == 928, "DfHostApiV27.world_difficulty_get ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_difficulty_set) == 936, "DfHostApiV27.world_difficulty_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_packet_write) == 944, "DfHostApiV27.player_packet_write ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_player_spawn_get) == 952, "DfHostApiV27.world_player_spawn_get ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_player_spawn_set) == 960, "DfHostApiV27.world_player_spawn_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_string_get) == 968, "DfHostApiV27.player_string_get ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_toast) == 976, "DfHostApiV27.player_toast ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_cooldown) == 984, "DfHostApiV27.player_cooldown ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_knock_back) == 992, "DfHostApiV27.player_knock_back ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_final_damage) == 1000, "DfHostApiV27.player_final_damage ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_using_item) == 1008, "DfHostApiV27.player_using_item ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_sleeping) == 1016, "DfHostApiV27.player_sleeping ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_death_position) == 1024, "DfHostApiV27.player_death_position ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_block_action) == 1032, "DfHostApiV27.player_block_action ABI offset changed");
_Static_assert(sizeof(DfEntityNewView) == 152, "DfEntityNewView ABI layout changed");
_Static_assert(sizeof(DfBBox) == 48, "DfBBox ABI layout changed");
_Static_assert(offsetof(DfBBox, min) == 0, "DfBBox.min ABI offset changed");
_Static_assert(offsetof(DfBBox, max) == 24, "DfBBox.max ABI offset changed");
_Static_assert(sizeof(DfDimensionView) == 24, "DfDimensionView ABI layout changed");
_Static_assert(offsetof(DfDimensionView, lava_spread_nanoseconds) == 16, "DfDimensionView.lava_spread_nanoseconds ABI offset changed");
_Static_assert(sizeof(DfWorldConfigV1) == 80, "DfWorldConfigV1 ABI layout changed");
_Static_assert(offsetof(DfWorldConfigV1, dimension_view) == 56, "DfWorldConfigV1.dimension_view ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_hurt) == 392, "DfHostApiV27.player_hurt ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, skin_snapshot_info) == 400, "DfHostApiV27.skin_snapshot_info ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, skin_snapshot_set) == 408, "DfHostApiV27.skin_snapshot_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_transfer) == 416, "DfHostApiV27.player_transfer ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_effects) == 424, "DfHostApiV27.player_effects ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_effects_clear) == 432, "DfHostApiV27.player_effects_clear ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_liquid_get) == 440, "DfHostApiV27.world_liquid_get ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_action) == 448, "DfHostApiV27.player_action ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_range) == 456, "DfHostApiV27.world_range ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_block_loaded) == 464, "DfHostApiV27.world_block_loaded ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_blocks_within_open) == 472, "DfHostApiV27.world_blocks_within_open ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_blocks_within_next) == 480, "DfHostApiV27.world_blocks_within_next ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_blocks_within_close) == 488, "DfHostApiV27.world_blocks_within_close ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_highest_light_blocker) == 496, "DfHostApiV27.world_highest_light_blocker ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_highest_block) == 504, "DfHostApiV27.world_highest_block ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_light) == 512, "DfHostApiV27.world_light ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_sky_light) == 520, "DfHostApiV27.world_sky_light ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_liquid_set) == 528, "DfHostApiV27.world_liquid_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_block_update_schedule) == 536, "DfHostApiV27.world_block_update_schedule ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_biome_get) == 544, "DfHostApiV27.world_biome_get ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_biome_set) == 552, "DfHostApiV27.world_biome_set ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_temperature) == 560, "DfHostApiV27.world_temperature ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_raining_at) == 568, "DfHostApiV27.world_raining_at ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_snowing_at) == 576, "DfHostApiV27.world_snowing_at ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_thundering_at) == 584, "DfHostApiV27.world_thundering_at ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_raining) == 592, "DfHostApiV27.world_raining ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_thundering) == 600, "DfHostApiV27.world_thundering ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_current_tick) == 608, "DfHostApiV27.world_current_tick ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, player_held_items_open) == 616, "DfHostApiV27.player_held_items_open ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, entity_player) == 624, "DfHostApiV27.entity_player ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_current) == 632, "DfHostApiV27.world_current ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_entity_iterator_open) == 640, "DfHostApiV27.world_entity_iterator_open ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_entity_iterator_next) == 648, "DfHostApiV27.world_entity_iterator_next ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_entity_iterator_close) == 656, "DfHostApiV27.world_entity_iterator_close ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, entity_handle) == 664, "DfHostApiV27.entity_handle ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, entity_handle_entity) == 672, "DfHostApiV27.entity_handle_entity ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, entity_handle_uuid) == 680, "DfHostApiV27.entity_handle_uuid ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, entity_handle_closed) == 688, "DfHostApiV27.entity_handle_closed ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, entity_handle_close) == 696, "DfHostApiV27.entity_handle_close ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_entity_remove) == 704, "DfHostApiV27.world_entity_remove ABI offset changed");
_Static_assert(offsetof(DfHostApiV27, world_entity_add) == 712, "DfHostApiV27.world_entity_add ABI offset changed");
#endif

extern DfStatus bg_go_player_text(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfStringView message);
extern DfStatus bg_go_player_title(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfTitleView title);
extern DfStatus bg_go_player_scoreboard(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfScoreboardView scoreboard);
extern DfStatus bg_go_player_scoreboard_remove(uint64_t context, DfInvocationId invocation, DfPlayerId player);
extern DfStatus bg_go_player_form_send(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfFormView *form);
extern DfStatus bg_go_player_form_close(uint64_t context, DfInvocationId invocation, DfPlayerId player);

DfStatus bg_call_form_response(DfFormResponseFn callback, void *callback_context, DfInvocationId invocation, const DfPlayerSnapshot *submitter, uint32_t outcome, DfStringView response_json) { return callback(callback_context, invocation, submitter, outcome, response_json); }
void bg_call_form_drop(DfFormDropFn callback, void *callback_context) { callback(callback_context); }
void bg_call_item_stack_views_drop(DfItemStackViewsDropFn callback, void *context) { callback(context); }
extern DfStatus bg_go_player_transform(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfVec3 vector, double yaw, double pitch);
extern DfStatus bg_go_player_transfer(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfWorldId world, DfVec3 position);
extern DfStatus bg_go_player_kinematics(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfPlayerKinematics *kinematics);
extern DfStatus bg_go_player_state_set(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue value);
extern DfStatus bg_go_player_state_get(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue *value);
extern DfStatus bg_go_player_action(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue value, DfPlayerStateValue *result);
extern DfStatus bg_go_player_string_get(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfStringBuffer *value);
extern DfStatus bg_go_player_toast(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfStringView title, DfStringView message);
extern DfStatus bg_go_player_cooldown(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t operation, DfStringView identifier, int32_t metadata, int64_t duration_nanoseconds, uint8_t *active);
extern DfStatus bg_go_player_knock_back(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfVec3 source, double force, double height);
extern DfStatus bg_go_player_final_damage(uint64_t context, DfInvocationId invocation, DfPlayerId player, double damage, const DfDamageSourceView *source, double *result);
extern DfStatus bg_go_player_using_item(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint8_t *using_item);
extern DfStatus bg_go_player_sleeping(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfBlockPos *position, uint8_t *sleeping);
extern DfStatus bg_go_player_death_position(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfVec3 *position, DfDimensionView *dimension, uint8_t *found);
extern DfStatus bg_go_player_block_action(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfBlockPos position, int32_t face, DfVec3 click_position);
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
extern DfStatus bg_go_player_held_items_open(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfItemStackSnapshot *main_hand, DfItemStackSnapshot *off_hand);
extern DfStatus bg_go_world_name(uint64_t context, DfInvocationId invocation, DfWorldId world, DfStringBuffer *name);
extern DfStatus bg_go_world_unload(uint64_t context, DfInvocationId invocation, DfWorldId world);
extern DfStatus bg_go_world_save(uint64_t context, DfInvocationId invocation, DfWorldId world);
extern DfStatus bg_go_world_block_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, DfBlockData *block);
extern DfStatus bg_go_block_by_name(uint64_t context, DfStringView name, DfStringView properties_nbt, uint8_t *found, DfBlockData *block);
extern DfStatus bg_go_world_block_loaded(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *loaded, DfBlockData *block);
extern DfStatus bg_go_world_liquid_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *found, DfBlockData *block);
extern DfStatus bg_go_world_liquid_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, const DfBlockView *liquid);
extern DfStatus bg_go_world_block_update_schedule(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, const DfBlockView *block, int64_t delay_nanoseconds);
extern DfStatus bg_go_world_biome_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, int32_t *biome);
extern DfStatus bg_go_world_biome_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, int32_t biome);
extern DfStatus bg_go_world_temperature(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, double *temperature);
extern DfStatus bg_go_world_raining_at(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *value);
extern DfStatus bg_go_world_snowing_at(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *value);
extern DfStatus bg_go_world_thundering_at(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *value);
extern DfStatus bg_go_world_raining(uint64_t context, DfInvocationId invocation, DfWorldId world, uint8_t *value);
extern DfStatus bg_go_world_thundering(uint64_t context, DfInvocationId invocation, DfWorldId world, uint8_t *value);
extern DfStatus bg_go_world_current_tick(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t *tick);
extern DfStatus bg_go_world_block_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, const DfBlockView *block, uint32_t flags);
extern DfStatus bg_go_world_range(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockRange *range);
extern DfStatus bg_go_world_blocks_within_open(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, int32_t radius, const DfBlockView *blocks, uint64_t block_count, DfBlockIteratorId *iterator);
extern DfStatus bg_go_world_blocks_within_next(uint64_t context, DfInvocationId invocation, DfBlockIteratorId iterator, DfBlockPos *position, uint8_t *found);
extern void bg_go_world_blocks_within_close(uint64_t context, DfInvocationId invocation, DfBlockIteratorId iterator);
extern DfStatus bg_go_world_highest_light_blocker(uint64_t context, DfInvocationId invocation, DfWorldId world, int32_t x, int32_t z, int32_t *height);
extern DfStatus bg_go_world_highest_block(uint64_t context, DfInvocationId invocation, DfWorldId world, int32_t x, int32_t z, int32_t *height);
extern DfStatus bg_go_world_light(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *level);
extern DfStatus bg_go_world_sky_light(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *level);
extern DfStatus bg_go_world_time_get(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t *time);
extern DfStatus bg_go_world_time_set(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t time);
extern DfStatus bg_go_world_spawn_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos *position);
extern DfStatus bg_go_world_spawn_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position);
extern DfStatus bg_go_world_player_spawn_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfUuid player, DfBlockPos *position);
extern DfStatus bg_go_world_player_spawn_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfUuid player, DfBlockPos position);
extern DfStatus bg_go_world_dimension_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfDimensionView *dimension);
extern DfStatus bg_go_world_time_cycle_get(uint64_t context, DfInvocationId invocation, DfWorldId world, uint8_t *value);
extern DfStatus bg_go_world_time_cycle_set(uint64_t context, DfInvocationId invocation, DfWorldId world, uint8_t value);
extern DfStatus bg_go_world_required_sleep_duration_set(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t duration_nanoseconds);
extern DfStatus bg_go_world_default_game_mode_get(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t *game_mode);
extern DfStatus bg_go_world_default_game_mode_set(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t game_mode);
extern DfStatus bg_go_world_tick_range_set(uint64_t context, DfInvocationId invocation, DfWorldId world, int32_t value);
extern DfStatus bg_go_world_difficulty_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfDifficultyView *difficulty);
extern DfStatus bg_go_world_difficulty_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfDifficultyView difficulty);
extern DfStatus bg_go_player_packet_write(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint64_t packet);
extern DfStatus bg_go_world_entity_spawn(uint64_t context, DfInvocationId invocation, DfWorldId world, const DfEntitySpawnViewV3 *entity, DfEntityId *output);
extern DfStatus bg_go_entity_state(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfEntityState *state);
extern DfStatus bg_go_entity_player(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfPlayerSnapshotBuffer *output);
extern DfStatus bg_go_entity_teleport(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfVec3 position);
extern DfStatus bg_go_entity_velocity_set(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfVec3 velocity);
extern DfStatus bg_go_entity_name_tag_set(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfStringView name_tag);
extern DfStatus bg_go_entity_despawn(uint64_t context, DfInvocationId invocation, DfEntityId entity);
extern DfStatus bg_go_world_particle_add(uint64_t context, DfInvocationId invocation, DfWorldId world, DfVec3 position, const DfParticleViewV1 *particle);
extern DfStatus bg_go_world_sound_play(uint64_t context, DfInvocationId invocation, DfWorldId world, DfVec3 position, const DfSoundViewV1 *sound);
extern DfStatus bg_go_player_sound_play(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfSoundViewV1 *sound);
extern DfStatus bg_go_world_current(uint64_t context, DfInvocationId invocation, DfWorldId *world);
extern DfStatus bg_go_world_entity_iterator_open(uint64_t context, DfInvocationId invocation, DfWorldId world, uint8_t players_only, DfEntityIteratorId *iterator);
extern DfStatus bg_go_world_entities_within_open(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBBox box, DfEntityIteratorId *iterator);
extern DfStatus bg_go_world_entity_iterator_next(uint64_t context, DfInvocationId invocation, DfEntityIteratorId iterator, DfEntityId *entity, uint8_t *found);
extern void bg_go_world_entity_iterator_close(uint64_t context, DfInvocationId invocation, DfEntityIteratorId iterator);
extern DfStatus bg_go_entity_handle(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfEntityHandleId *handle);
extern DfStatus bg_go_entity_handle_entity(uint64_t context, DfInvocationId invocation, DfEntityHandleId handle, DfEntityId *entity, uint8_t *found);
extern DfStatus bg_go_entity_handle_uuid(uint64_t context, DfEntityHandleId handle, DfUuid *uuid);
extern DfStatus bg_go_entity_handle_closed(uint64_t context, DfEntityHandleId handle, uint8_t *closed);
extern DfStatus bg_go_entity_handle_close(uint64_t context, DfEntityHandleId handle);
extern DfStatus bg_go_entity_new(uint64_t context, const DfEntityNewView *entity, DfEntityHandleId *handle);
extern DfStatus bg_go_entity_handle_type(uint64_t context, DfEntityHandleId handle, DfStringBuffer *entity_type);
extern DfStatus bg_go_world_entity_remove(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfEntityHandleId *handle);
extern DfStatus bg_go_world_entity_add(uint64_t context, DfInvocationId invocation, DfEntityHandleId handle, const DfVec3 *position, DfEntityId *entity);
extern DfStatus bg_go_server_players_open(uint64_t context, DfInvocationId invocation, DfPlayerIteratorId *iterator);
extern DfStatus bg_go_server_players_next(uint64_t context, DfInvocationId invocation, DfPlayerIteratorId iterator, DfInvocationId *player_invocation, DfPlayerSnapshotBuffer *player, uint8_t *found);
extern void bg_go_server_players_close(uint64_t context, DfInvocationId invocation, DfPlayerIteratorId iterator);
extern DfStatus bg_go_server_player(uint64_t context, DfUuid uuid, DfEntityHandleId *player, uint8_t *found);
extern DfStatus bg_go_server_player_by_name(uint64_t context, DfStringView name, DfEntityHandleId *player, uint8_t *found);
extern DfStatus bg_go_server_max_player_count(uint64_t context, int64_t *count);
extern DfStatus bg_go_server_player_count(uint64_t context, int64_t *count);
extern DfStatus bg_go_server_player_by_xuid(uint64_t context, DfStringView xuid, DfEntityHandleId *player, uint8_t *found);
extern DfStatus bg_go_player_xuid(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfStringBuffer *xuid);
extern DfStatus bg_go_server_world(uint64_t context, uint32_t dimension, DfWorldId *world);
extern DfStatus bg_go_world_schedule(uint64_t context, DfWorldId world, uint64_t plugin, uint64_t callback, int64_t delay_nanoseconds);
extern DfStatus bg_go_world_task_cancel(uint64_t context, uint64_t plugin, uint64_t callback, uint8_t *cancelled);
extern DfStatus bg_go_world_new(uint64_t context, const DfWorldConfigV1 *config, DfWorldId *world);
extern DfStatus bg_go_packet_field_get(uint64_t context, uint64_t packet, uint32_t field, DfPacketFieldValue *value);
extern DfStatus bg_go_packet_field_set(uint64_t context, uint64_t packet, uint32_t field, const DfPacketFieldValue *value);

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

static DfStatus host_player_kinematics(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfPlayerKinematics *kinematics) {
    return bg_go_player_kinematics(context, invocation, player, kinematics);
}

static DfStatus host_player_state_set(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue value) {
    return bg_go_player_state_set(context, invocation, player, kind, value);
}

static DfStatus host_player_state_get(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue *value) {
    return bg_go_player_state_get(context, invocation, player, kind, value);
}

static DfStatus host_player_action(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue value, DfPlayerStateValue *result) {
    return bg_go_player_action(context, invocation, player, kind, value, result);
}

static DfStatus host_player_string_get(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfStringBuffer *value) {
    return bg_go_player_string_get(context, invocation, player, kind, value);
}

static DfStatus host_player_toast(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfStringView title, DfStringView message) {
    return bg_go_player_toast(context, invocation, player, title, message);
}

static DfStatus host_player_cooldown(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t operation, DfStringView identifier, int32_t metadata, int64_t duration_nanoseconds, uint8_t *active) {
    return bg_go_player_cooldown(context, invocation, player, operation, identifier, metadata, duration_nanoseconds, active);
}

static DfStatus host_player_knock_back(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfVec3 source, double force, double height) {
    return bg_go_player_knock_back(context, invocation, player, source, force, height);
}

static DfStatus host_player_final_damage(uint64_t context, DfInvocationId invocation, DfPlayerId player, double damage, const DfDamageSourceView *source, double *result) {
    return bg_go_player_final_damage(context, invocation, player, damage, source, result);
}

static DfStatus host_player_using_item(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint8_t *using_item) {
    return bg_go_player_using_item(context, invocation, player, using_item);
}

static DfStatus host_player_sleeping(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfBlockPos *position, uint8_t *sleeping) {
    return bg_go_player_sleeping(context, invocation, player, position, sleeping);
}

static DfStatus host_player_death_position(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfVec3 *position, DfDimensionView *dimension, uint8_t *found) {
    return bg_go_player_death_position(context, invocation, player, position, dimension, found);
}

static DfStatus host_player_block_action(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfBlockPos position, int32_t face, DfVec3 click_position) {
    return bg_go_player_block_action(context, invocation, player, kind, position, face, click_position);
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
static DfStatus host_player_held_items_open(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfItemStackSnapshot *main_hand, DfItemStackSnapshot *off_hand) { return bg_go_player_held_items_open(context, invocation, player, main_hand, off_hand); }
static DfStatus host_world_name(uint64_t context, DfInvocationId invocation, DfWorldId world, DfStringBuffer *name) { return bg_go_world_name(context, invocation, world, name); }
static DfStatus host_world_unload(uint64_t context, DfInvocationId invocation, DfWorldId world) { return bg_go_world_unload(context, invocation, world); }
static DfStatus host_world_save(uint64_t context, DfInvocationId invocation, DfWorldId world) { return bg_go_world_save(context, invocation, world); }
static DfStatus host_world_block_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, DfBlockData *block) { return bg_go_world_block_get(context, invocation, world, position, block); }
static DfStatus host_block_by_name(uint64_t context, DfStringView name, DfStringView properties_nbt, uint8_t *found, DfBlockData *block) { return bg_go_block_by_name(context, name, properties_nbt, found, block); }
static DfStatus host_world_block_loaded(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *loaded, DfBlockData *block) { return bg_go_world_block_loaded(context, invocation, world, position, loaded, block); }
static DfStatus host_world_liquid_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *found, DfBlockData *block) { return bg_go_world_liquid_get(context, invocation, world, position, found, block); }
static DfStatus host_world_liquid_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, const DfBlockView *liquid) { return bg_go_world_liquid_set(context, invocation, world, position, liquid); }
static DfStatus host_world_block_update_schedule(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, const DfBlockView *block, int64_t delay_nanoseconds) { return bg_go_world_block_update_schedule(context, invocation, world, position, block, delay_nanoseconds); }
static DfStatus host_world_biome_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, int32_t *biome) { return bg_go_world_biome_get(context, invocation, world, position, biome); }
static DfStatus host_world_biome_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, int32_t biome) { return bg_go_world_biome_set(context, invocation, world, position, biome); }
static DfStatus host_world_temperature(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, double *temperature) { return bg_go_world_temperature(context, invocation, world, position, temperature); }
static DfStatus host_world_raining_at(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *value) { return bg_go_world_raining_at(context, invocation, world, position, value); }
static DfStatus host_world_snowing_at(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *value) { return bg_go_world_snowing_at(context, invocation, world, position, value); }
static DfStatus host_world_thundering_at(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *value) { return bg_go_world_thundering_at(context, invocation, world, position, value); }
static DfStatus host_world_raining(uint64_t context, DfInvocationId invocation, DfWorldId world, uint8_t *value) { return bg_go_world_raining(context, invocation, world, value); }
static DfStatus host_world_thundering(uint64_t context, DfInvocationId invocation, DfWorldId world, uint8_t *value) { return bg_go_world_thundering(context, invocation, world, value); }
static DfStatus host_world_current_tick(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t *tick) { return bg_go_world_current_tick(context, invocation, world, tick); }
static DfStatus host_world_block_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, const DfBlockView *block, uint32_t flags) { return bg_go_world_block_set(context, invocation, world, position, block, flags); }
static DfStatus host_world_range(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockRange *range) { return bg_go_world_range(context, invocation, world, range); }
static DfStatus host_world_blocks_within_open(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, int32_t radius, const DfBlockView *blocks, uint64_t block_count, DfBlockIteratorId *iterator) { return bg_go_world_blocks_within_open(context, invocation, world, position, radius, blocks, block_count, iterator); }
static DfStatus host_world_blocks_within_next(uint64_t context, DfInvocationId invocation, DfBlockIteratorId iterator, DfBlockPos *position, uint8_t *found) { return bg_go_world_blocks_within_next(context, invocation, iterator, position, found); }
static void host_world_blocks_within_close(uint64_t context, DfInvocationId invocation, DfBlockIteratorId iterator) { bg_go_world_blocks_within_close(context, invocation, iterator); }
static DfStatus host_world_highest_light_blocker(uint64_t context, DfInvocationId invocation, DfWorldId world, int32_t x, int32_t z, int32_t *height) { return bg_go_world_highest_light_blocker(context, invocation, world, x, z, height); }
static DfStatus host_world_highest_block(uint64_t context, DfInvocationId invocation, DfWorldId world, int32_t x, int32_t z, int32_t *height) { return bg_go_world_highest_block(context, invocation, world, x, z, height); }
static DfStatus host_world_light(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *level) { return bg_go_world_light(context, invocation, world, position, level); }
static DfStatus host_world_sky_light(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *level) { return bg_go_world_sky_light(context, invocation, world, position, level); }
static DfStatus host_world_time_get(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t *time) { return bg_go_world_time_get(context, invocation, world, time); }
static DfStatus host_world_time_set(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t time) { return bg_go_world_time_set(context, invocation, world, time); }
static DfStatus host_world_spawn_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos *position) { return bg_go_world_spawn_get(context, invocation, world, position); }
static DfStatus host_world_spawn_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position) { return bg_go_world_spawn_set(context, invocation, world, position); }
static DfStatus host_world_player_spawn_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfUuid player, DfBlockPos *position) { return bg_go_world_player_spawn_get(context, invocation, world, player, position); }
static DfStatus host_world_player_spawn_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfUuid player, DfBlockPos position) { return bg_go_world_player_spawn_set(context, invocation, world, player, position); }
static DfStatus host_world_dimension_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfDimensionView *dimension) { return bg_go_world_dimension_get(context, invocation, world, dimension); }
static DfStatus host_world_time_cycle_get(uint64_t context, DfInvocationId invocation, DfWorldId world, uint8_t *value) { return bg_go_world_time_cycle_get(context, invocation, world, value); }
static DfStatus host_world_time_cycle_set(uint64_t context, DfInvocationId invocation, DfWorldId world, uint8_t value) { return bg_go_world_time_cycle_set(context, invocation, world, value); }
static DfStatus host_world_required_sleep_duration_set(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t duration_nanoseconds) { return bg_go_world_required_sleep_duration_set(context, invocation, world, duration_nanoseconds); }
static DfStatus host_world_default_game_mode_get(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t *game_mode) { return bg_go_world_default_game_mode_get(context, invocation, world, game_mode); }
static DfStatus host_world_default_game_mode_set(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t game_mode) { return bg_go_world_default_game_mode_set(context, invocation, world, game_mode); }
static DfStatus host_world_tick_range_set(uint64_t context, DfInvocationId invocation, DfWorldId world, int32_t value) { return bg_go_world_tick_range_set(context, invocation, world, value); }
static DfStatus host_world_difficulty_get(uint64_t context, DfInvocationId invocation, DfWorldId world, DfDifficultyView *difficulty) { return bg_go_world_difficulty_get(context, invocation, world, difficulty); }
static DfStatus host_world_difficulty_set(uint64_t context, DfInvocationId invocation, DfWorldId world, DfDifficultyView difficulty) { return bg_go_world_difficulty_set(context, invocation, world, difficulty); }
static DfStatus host_player_packet_write(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint64_t packet) { return bg_go_player_packet_write(context, invocation, player, packet); }
static DfStatus host_packet_field_get(uint64_t context, uint64_t packet, uint32_t field, DfPacketFieldValue *value) { return bg_go_packet_field_get(context, packet, field, value); }
static DfStatus host_packet_field_set(uint64_t context, uint64_t packet, uint32_t field, const DfPacketFieldValue *value) { return bg_go_packet_field_set(context, packet, field, value); }
static DfStatus host_world_entity_spawn(uint64_t context, DfInvocationId invocation, DfWorldId world, const DfEntitySpawnViewV3 *entity, DfEntityId *output) { return bg_go_world_entity_spawn(context, invocation, world, entity, output); }
static DfStatus host_entity_state(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfEntityState *state) { return bg_go_entity_state(context, invocation, entity, state); }
static DfStatus host_entity_player(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfPlayerSnapshotBuffer *output) { return bg_go_entity_player(context, invocation, entity, output); }
static DfStatus host_entity_teleport(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfVec3 position) { return bg_go_entity_teleport(context, invocation, entity, position); }
static DfStatus host_entity_velocity_set(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfVec3 velocity) { return bg_go_entity_velocity_set(context, invocation, entity, velocity); }
static DfStatus host_entity_name_tag_set(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfStringView name_tag) { return bg_go_entity_name_tag_set(context, invocation, entity, name_tag); }
static DfStatus host_entity_despawn(uint64_t context, DfInvocationId invocation, DfEntityId entity) { return bg_go_entity_despawn(context, invocation, entity); }
static DfStatus host_world_particle_add(uint64_t context, DfInvocationId invocation, DfWorldId world, DfVec3 position, const DfParticleViewV1 *particle) { return bg_go_world_particle_add(context, invocation, world, position, particle); }
static DfStatus host_world_sound_play(uint64_t context, DfInvocationId invocation, DfWorldId world, DfVec3 position, const DfSoundViewV1 *sound) { return bg_go_world_sound_play(context, invocation, world, position, sound); }
static DfStatus host_player_sound_play(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfSoundViewV1 *sound) { return bg_go_player_sound_play(context, invocation, player, sound); }
static DfStatus host_world_current(uint64_t context, DfInvocationId invocation, DfWorldId *world) { return bg_go_world_current(context, invocation, world); }
static DfStatus host_world_entity_iterator_open(uint64_t context, DfInvocationId invocation, DfWorldId world, uint8_t players_only, DfEntityIteratorId *iterator) { return bg_go_world_entity_iterator_open(context, invocation, world, players_only, iterator); }
static DfStatus host_world_entities_within_open(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBBox box, DfEntityIteratorId *iterator) { return bg_go_world_entities_within_open(context, invocation, world, box, iterator); }
static DfStatus host_world_entity_iterator_next(uint64_t context, DfInvocationId invocation, DfEntityIteratorId iterator, DfEntityId *entity, uint8_t *found) { return bg_go_world_entity_iterator_next(context, invocation, iterator, entity, found); }
static void host_world_entity_iterator_close(uint64_t context, DfInvocationId invocation, DfEntityIteratorId iterator) { bg_go_world_entity_iterator_close(context, invocation, iterator); }
static DfStatus host_entity_handle(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfEntityHandleId *handle) { return bg_go_entity_handle(context, invocation, entity, handle); }
static DfStatus host_entity_handle_entity(uint64_t context, DfInvocationId invocation, DfEntityHandleId handle, DfEntityId *entity, uint8_t *found) { return bg_go_entity_handle_entity(context, invocation, handle, entity, found); }
static DfStatus host_entity_handle_uuid(uint64_t context, DfEntityHandleId handle, DfUuid *uuid) { return bg_go_entity_handle_uuid(context, handle, uuid); }
static DfStatus host_entity_handle_closed(uint64_t context, DfEntityHandleId handle, uint8_t *closed) { return bg_go_entity_handle_closed(context, handle, closed); }
static DfStatus host_entity_handle_close(uint64_t context, DfEntityHandleId handle) { return bg_go_entity_handle_close(context, handle); }
static DfStatus host_entity_new(uint64_t context, const DfEntityNewView *entity, DfEntityHandleId *handle) { return bg_go_entity_new(context, entity, handle); }
static DfStatus host_entity_handle_type(uint64_t context, DfEntityHandleId handle, DfStringBuffer *entity_type) { return bg_go_entity_handle_type(context, handle, entity_type); }
static DfStatus host_world_entity_remove(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfEntityHandleId *handle) { return bg_go_world_entity_remove(context, invocation, entity, handle); }
static DfStatus host_world_entity_add(uint64_t context, DfInvocationId invocation, DfEntityHandleId handle, const DfVec3 *position, DfEntityId *entity) { return bg_go_world_entity_add(context, invocation, handle, position, entity); }
static DfStatus host_server_players_open(uint64_t context, DfInvocationId invocation, DfPlayerIteratorId *iterator) { return bg_go_server_players_open(context, invocation, iterator); }
static DfStatus host_server_players_next(uint64_t context, DfInvocationId invocation, DfPlayerIteratorId iterator, DfInvocationId *player_invocation, DfPlayerSnapshotBuffer *player, uint8_t *found) { return bg_go_server_players_next(context, invocation, iterator, player_invocation, player, found); }
static void host_server_players_close(uint64_t context, DfInvocationId invocation, DfPlayerIteratorId iterator) { bg_go_server_players_close(context, invocation, iterator); }
static DfStatus host_server_player(uint64_t context, DfUuid uuid, DfEntityHandleId *player, uint8_t *found) { return bg_go_server_player(context, uuid, player, found); }
static DfStatus host_server_player_by_name(uint64_t context, DfStringView name, DfEntityHandleId *player, uint8_t *found) { return bg_go_server_player_by_name(context, name, player, found); }
static DfStatus host_server_max_player_count(uint64_t context, int64_t *count) { return bg_go_server_max_player_count(context, count); }
static DfStatus host_server_player_count(uint64_t context, int64_t *count) { return bg_go_server_player_count(context, count); }
static DfStatus host_server_player_by_xuid(uint64_t context, DfStringView xuid, DfEntityHandleId *player, uint8_t *found) { return bg_go_server_player_by_xuid(context, xuid, player, found); }
static DfStatus host_player_xuid(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfStringBuffer *xuid) { return bg_go_player_xuid(context, invocation, player, xuid); }
static DfStatus host_server_world(uint64_t context, uint32_t dimension, DfWorldId *world) { return bg_go_server_world(context, dimension, world); }
static DfStatus host_world_schedule(uint64_t context, DfWorldId world, uint64_t plugin, uint64_t callback, int64_t delay_nanoseconds) { return bg_go_world_schedule(context, world, plugin, callback, delay_nanoseconds); }
static DfStatus host_world_task_cancel(uint64_t context, uint64_t plugin, uint64_t callback, uint8_t *cancelled) { return bg_go_world_task_cancel(context, plugin, callback, cancelled); }
static DfStatus host_world_new(uint64_t context, const DfWorldConfigV1 *config, DfWorldId *world) { return bg_go_world_new(context, config, world); }

typedef DfStatus (*RuntimeCreateFn)(const DfRuntimeConfig *, DfRuntime **, uint8_t *, uint64_t);
typedef void (*RuntimeDestroyFn)(DfRuntime *);
typedef DfStatus (*RuntimeEnableFn)(DfRuntime *, uint8_t *, uint64_t);
typedef void (*RuntimeDisableFn)(DfRuntime *);
typedef uint64_t (*RuntimeCountFn)(const DfRuntime *);
typedef DfStatus (*RuntimeEntityTypeAtFn)(const DfRuntime *, uint64_t, DfEntityTypeDescriptorV2 *);
typedef DfStatus (*RuntimeEntityAdoptFn)(DfRuntime *, uint64_t, uint64_t, DfEntityInstanceId *);
typedef DfStatus (*RuntimeEntityAdoptLocalFn)(DfRuntime *, uint64_t, uint64_t, uint64_t, DfEntityInstanceId *);
typedef DfStatus (*RuntimeEntityLoadFn)(DfRuntime *, uint64_t, const DfEntityLoadInput *, DfEntityInstanceId *);
typedef DfStatus (*RuntimeEntitySaveFn)(DfRuntime *, DfEntityInstanceId, DfEntitySaveState *);
typedef DfStatus (*RuntimeEntityTickFn)(DfRuntime *, DfEntityInstanceId, const DfEntityTickInput *, DfEntityTickState *);
typedef DfStatus (*RuntimeEntityHurtFn)(DfRuntime *, DfEntityInstanceId, const DfEntityHurtInput *, DfEntityHurtState *);
typedef DfStatus (*RuntimeEntityHealFn)(DfRuntime *, DfEntityInstanceId, const DfEntityHealInput *, DfEntityHealState *);
typedef DfStatus (*RuntimeEntityDeathFn)(DfRuntime *, DfEntityInstanceId, const DfEntityDeathInput *, DfEntityDeathState *);
typedef DfStatus (*RuntimeEntityDestroyFn)(DfRuntime *, DfEntityInstanceId);
typedef DfStatus (*RuntimeEntityDecodeNbtFn)(DfRuntime *, uint64_t, const DfEntityExactInput *, DfEntityExactState *);
typedef DfStatus (*RuntimeEntityCallFn)(DfRuntime *, DfEntityInstanceId, uint32_t, const DfEntityExactInput *, DfEntityExactState *);
typedef DfStatus (*RuntimeCommandAtFn)(const DfRuntime *, uint64_t, DfCommandDescriptor *);
typedef DfStatus (*RuntimeCommandFn)(DfRuntime *, uint64_t, const DfCommandInput *, DfCommandState *);
typedef DfStatus (*RuntimeCommandEnumFn)(DfRuntime *, uint64_t, uint64_t, uint64_t, const DfCommandEnumContext *, DfStringBuffer *);
typedef DfStatus (*RuntimeEventFn)(DfRuntime *, DfEventId, const void *, void *);
typedef DfStatus (*RuntimeScheduledFn)(DfRuntime *, uint64_t, uint64_t, DfInvocationId, uint32_t, uint32_t);
typedef DfStatus (*RuntimeAllowFn)(DfRuntime *, const DfAllowInput *, DfStringBuffer *, uint8_t *);

struct BgRuntimeLibrary {
    DfRuntime *runtime;
    DfHostApiV27 host_api;
    RuntimeDestroyFn destroy;
    RuntimeEnableFn enable;
    RuntimeDisableFn begin_disable;
    RuntimeDisableFn finish_disable;
    RuntimeDisableFn disable;
    RuntimeCountFn plugin_count;
    RuntimeCountFn subscriptions;
    RuntimeCountFn entity_type_count;
    RuntimeEntityTypeAtFn entity_type_at;
    RuntimeEntityAdoptFn entity_adopt;
    RuntimeEntityAdoptLocalFn entity_adopt_local;
    RuntimeEntityLoadFn entity_load;
    RuntimeEntitySaveFn entity_save;
    RuntimeEntityTickFn entity_tick;
    RuntimeEntityHurtFn entity_hurt;
    RuntimeEntityHealFn entity_heal;
    RuntimeEntityDeathFn entity_death;
    RuntimeEntityDestroyFn entity_destroy;
    RuntimeEntityDecodeNbtFn entity_decode_nbt;
    RuntimeEntityCallFn entity_call;
    RuntimeCountFn command_count;
    RuntimeCommandAtFn command_at;
    RuntimeCommandFn handle_command;
    RuntimeCommandEnumFn command_enum_options;
    RuntimeEventFn handle_event;
    RuntimeScheduledFn handle_scheduled;
    RuntimeAllowFn allow;
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
    // NativeAOT shared libraries install process-wide signal handlers and do not support safe
    // unloading. Keep every successful dlopen reference until process exit, including failures
    // after this point, so those handlers never point into unmapped code.

    RuntimeCreateFn create = (RuntimeCreateFn) load_symbol(handle, "df_runtime_create", error, error_capacity);
    RuntimeDestroyFn destroy = (RuntimeDestroyFn) load_symbol(handle, "df_runtime_destroy", error, error_capacity);
    RuntimeEnableFn enable = (RuntimeEnableFn) load_symbol(handle, "df_runtime_enable", error, error_capacity);
    RuntimeDisableFn begin_disable = (RuntimeDisableFn) load_symbol(handle, "df_runtime_begin_disable", error, error_capacity);
    RuntimeDisableFn finish_disable = (RuntimeDisableFn) load_symbol(handle, "df_runtime_finish_disable", error, error_capacity);
    RuntimeDisableFn disable = (RuntimeDisableFn) load_symbol(handle, "df_runtime_disable", error, error_capacity);
    RuntimeCountFn plugin_count = (RuntimeCountFn) load_symbol(handle, "df_runtime_plugin_count", error, error_capacity);
    RuntimeCountFn subscriptions = (RuntimeCountFn) load_symbol(handle, "df_runtime_subscriptions", error, error_capacity);
    RuntimeCountFn entity_type_count = (RuntimeCountFn) load_symbol(handle, "df_runtime_entity_type_count", error, error_capacity);
    RuntimeEntityTypeAtFn entity_type_at = (RuntimeEntityTypeAtFn) load_symbol(handle, "df_runtime_entity_type_at", error, error_capacity);
    RuntimeEntityAdoptFn entity_adopt = (RuntimeEntityAdoptFn) load_symbol(handle, "df_runtime_entity_adopt", error, error_capacity);
    RuntimeEntityAdoptLocalFn entity_adopt_local = (RuntimeEntityAdoptLocalFn) load_symbol(handle, "df_runtime_entity_adopt_local", error, error_capacity);
    RuntimeEntityLoadFn entity_load = (RuntimeEntityLoadFn) load_symbol(handle, "df_runtime_entity_load", error, error_capacity);
    RuntimeEntitySaveFn entity_save = (RuntimeEntitySaveFn) load_symbol(handle, "df_runtime_entity_save", error, error_capacity);
    RuntimeEntityTickFn entity_tick = (RuntimeEntityTickFn) load_symbol(handle, "df_runtime_entity_tick", error, error_capacity);
    RuntimeEntityHurtFn entity_hurt = (RuntimeEntityHurtFn) load_symbol(handle, "df_runtime_entity_hurt", error, error_capacity);
    RuntimeEntityHealFn entity_heal = (RuntimeEntityHealFn) load_symbol(handle, "df_runtime_entity_heal", error, error_capacity);
    RuntimeEntityDeathFn entity_death = (RuntimeEntityDeathFn) load_symbol(handle, "df_runtime_entity_death", error, error_capacity);
    RuntimeEntityDestroyFn entity_destroy = (RuntimeEntityDestroyFn) load_symbol(handle, "df_runtime_entity_destroy", error, error_capacity);
    RuntimeEntityDecodeNbtFn entity_decode_nbt = (RuntimeEntityDecodeNbtFn) load_symbol(handle, "df_runtime_entity_decode_nbt", error, error_capacity);
    RuntimeEntityCallFn entity_call = (RuntimeEntityCallFn) load_symbol(handle, "df_runtime_entity_call", error, error_capacity);
    RuntimeCountFn command_count = (RuntimeCountFn) load_symbol(handle, "df_runtime_command_count", error, error_capacity);
    RuntimeCommandAtFn command_at = (RuntimeCommandAtFn) load_symbol(handle, "df_runtime_command_at", error, error_capacity);
    RuntimeCommandFn handle_command = (RuntimeCommandFn) load_symbol(handle, "df_runtime_handle_command", error, error_capacity);
    RuntimeCommandEnumFn command_enum_options = (RuntimeCommandEnumFn) load_symbol(handle, "df_runtime_command_enum_options", error, error_capacity);
    RuntimeEventFn handle_event = (RuntimeEventFn) load_symbol(handle, "df_runtime_handle_event", error, error_capacity);
    RuntimeScheduledFn handle_scheduled = (RuntimeScheduledFn) load_symbol(handle, "df_runtime_handle_scheduled", error, error_capacity);
    RuntimeAllowFn allow = (RuntimeAllowFn) load_symbol(handle, "df_runtime_allow", error, error_capacity);
    if (create == NULL || destroy == NULL || enable == NULL || begin_disable == NULL || finish_disable == NULL || disable == NULL || plugin_count == NULL || subscriptions == NULL || entity_type_count == NULL || entity_type_at == NULL || entity_adopt == NULL || entity_adopt_local == NULL || entity_load == NULL || entity_save == NULL || entity_tick == NULL || entity_hurt == NULL || entity_heal == NULL || entity_death == NULL || entity_destroy == NULL || entity_decode_nbt == NULL || entity_call == NULL || command_count == NULL || command_at == NULL || handle_command == NULL || command_enum_options == NULL || handle_event == NULL || handle_scheduled == NULL || allow == NULL) {
        return DF_STATUS_ERROR;
    }

    BgRuntimeLibrary *library = calloc(1, sizeof(*library));
    if (library == NULL) {
        write_error(error, error_capacity, "allocate runtime bridge");
        return DF_STATUS_ERROR;
    }

    library->host_api = (DfHostApiV27) {
        .abi_version = DF_HOST_ABI_VERSION,
        .struct_size = sizeof(DfHostApiV27),
        .context = host_context,
        .player_text = host_player_text,
        .player_title = host_player_title,
        .player_transform = host_player_transform,
        .player_kinematics = host_player_kinematics,
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
        .player_transfer = host_player_transfer,
        .player_effects = host_player_effects,
        .player_effects_clear = host_player_effects_clear,
        .world_liquid_get = host_world_liquid_get,
        .player_action = host_player_action,
        .world_range = host_world_range,
        .world_block_loaded = host_world_block_loaded,
        .world_blocks_within_open = host_world_blocks_within_open,
        .world_blocks_within_next = host_world_blocks_within_next,
        .world_blocks_within_close = host_world_blocks_within_close,
        .world_highest_light_blocker = host_world_highest_light_blocker,
        .world_highest_block = host_world_highest_block,
        .world_light = host_world_light,
        .world_sky_light = host_world_sky_light,
        .world_liquid_set = host_world_liquid_set,
        .world_block_update_schedule = host_world_block_update_schedule,
        .world_biome_get = host_world_biome_get,
        .world_biome_set = host_world_biome_set,
        .world_temperature = host_world_temperature,
        .world_raining_at = host_world_raining_at,
        .world_snowing_at = host_world_snowing_at,
        .world_thundering_at = host_world_thundering_at,
        .world_raining = host_world_raining,
        .world_thundering = host_world_thundering,
        .world_current_tick = host_world_current_tick,
        .player_held_items_open = host_player_held_items_open,
        .entity_player = host_entity_player,
        .world_current = host_world_current,
        .world_entity_iterator_open = host_world_entity_iterator_open,
        .world_entity_iterator_next = host_world_entity_iterator_next,
        .world_entity_iterator_close = host_world_entity_iterator_close,
        .entity_handle = host_entity_handle,
        .entity_handle_entity = host_entity_handle_entity,
        .entity_handle_uuid = host_entity_handle_uuid,
        .entity_handle_closed = host_entity_handle_closed,
        .entity_handle_close = host_entity_handle_close,
        .world_entity_remove = host_world_entity_remove,
        .world_entity_add = host_world_entity_add,
        .server_players_open = host_server_players_open,
        .server_players_next = host_server_players_next,
        .server_players_close = host_server_players_close,
        .server_player = host_server_player,
        .server_player_by_name = host_server_player_by_name,
        .server_max_player_count = host_server_max_player_count,
        .server_player_count = host_server_player_count,
        .server_player_by_xuid = host_server_player_by_xuid,
        .player_xuid = host_player_xuid,
        .server_world = host_server_world,
        .world_schedule = host_world_schedule,
        .world_new = host_world_new,
        .world_entities_within_open = host_world_entities_within_open,
        .block_by_name = host_block_by_name,
        .entity_new = host_entity_new,
        .entity_handle_type = host_entity_handle_type,
        .world_task_cancel = host_world_task_cancel,
        .packet_field_get = host_packet_field_get,
        .packet_field_set = host_packet_field_set,
        .world_dimension_get = host_world_dimension_get,
        .world_time_cycle_get = host_world_time_cycle_get,
        .world_time_cycle_set = host_world_time_cycle_set,
        .world_required_sleep_duration_set = host_world_required_sleep_duration_set,
        .world_default_game_mode_get = host_world_default_game_mode_get,
        .world_default_game_mode_set = host_world_default_game_mode_set,
        .world_tick_range_set = host_world_tick_range_set,
        .world_difficulty_get = host_world_difficulty_get,
        .world_difficulty_set = host_world_difficulty_set,
        .player_packet_write = host_player_packet_write,
        .world_player_spawn_get = host_world_player_spawn_get,
        .world_player_spawn_set = host_world_player_spawn_set,
        .player_string_get = host_player_string_get,
        .player_toast = host_player_toast,
        .player_cooldown = host_player_cooldown,
        .player_knock_back = host_player_knock_back,
        .player_final_damage = host_player_final_damage,
        .player_using_item = host_player_using_item,
        .player_sleeping = host_player_sleeping,
        .player_death_position = host_player_death_position,
        .player_block_action = host_player_block_action,
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
        return DF_STATUS_ERROR;
    }

    library->destroy = destroy;
    library->enable = enable;
    library->begin_disable = begin_disable;
    library->finish_disable = finish_disable;
    library->disable = disable;
    library->plugin_count = plugin_count;
    library->subscriptions = subscriptions;
    library->entity_type_count = entity_type_count;
    library->entity_type_at = entity_type_at;
    library->entity_adopt = entity_adopt;
    library->entity_adopt_local = entity_adopt_local;
    library->entity_load = entity_load;
    library->entity_save = entity_save;
    library->entity_tick = entity_tick;
    library->entity_hurt = entity_hurt;
    library->entity_heal = entity_heal;
    library->entity_death = entity_death;
    library->entity_destroy = entity_destroy;
    library->entity_decode_nbt = entity_decode_nbt;
    library->entity_call = entity_call;
    library->command_count = command_count;
    library->command_at = command_at;
    library->handle_command = handle_command;
    library->command_enum_options = command_enum_options;
    library->handle_event = handle_event;
    library->handle_scheduled = handle_scheduled;
    library->allow = allow;
    *out = library;
    return DF_STATUS_OK;
}

void bg_runtime_close(BgRuntimeLibrary *library) {
    if (library == NULL) {
        return;
    }
    library->destroy(library->runtime);
    free(library);
}

DfStatus bg_runtime_enable(BgRuntimeLibrary *library, uint8_t *error, uint64_t error_capacity) {
    if (library == NULL) {
        write_error(error, error_capacity, "native runtime is closed");
        return DF_STATUS_ERROR;
    }
    return library->enable(library->runtime, error, error_capacity);
}

void bg_runtime_disable(BgRuntimeLibrary *library) {
    if (library != NULL) {
        library->disable(library->runtime);
    }
}

void bg_runtime_begin_disable(BgRuntimeLibrary *library) {
    if (library != NULL) {
        library->begin_disable(library->runtime);
    }
}

void bg_runtime_finish_disable(BgRuntimeLibrary *library) {
    if (library != NULL) {
        library->finish_disable(library->runtime);
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

DfStatus bg_runtime_entity_adopt_local(BgRuntimeLibrary *library, uint64_t plugin, uint64_t type_key, uint64_t opaque, DfEntityInstanceId *out) {
    return library == NULL || out == NULL ? DF_STATUS_ERROR : library->entity_adopt_local(library->runtime, plugin, type_key, opaque, out);
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

DfStatus bg_runtime_entity_decode_nbt(BgRuntimeLibrary *library, uint64_t type_key, const DfEntityExactInput *input, DfEntityExactState *state) {
    return library == NULL || input == NULL || state == NULL ? DF_STATUS_ERROR : library->entity_decode_nbt(library->runtime, type_key, input, state);
}

DfStatus bg_runtime_entity_call(BgRuntimeLibrary *library, DfEntityInstanceId identity, uint32_t operation, const DfEntityExactInput *input, DfEntityExactState *state) {
    return library == NULL ? DF_STATUS_ERROR : library->entity_call(library->runtime, identity, operation, input, state);
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

DfStatus bg_runtime_handle_scheduled(
    BgRuntimeLibrary *library,
    uint64_t plugin,
    uint64_t callback,
    DfInvocationId invocation,
    uint32_t phase,
    uint32_t result
) {
    if (library == NULL || plugin == 0 || callback == 0 || phase > DF_WORLD_TASK_PHASE_COMPLETE || result > DF_WORLD_TASK_FAILED) {
        return DF_STATUS_ERROR;
    }
    return library->handle_scheduled(library->runtime, plugin, callback, invocation, phase, result);
}

DfStatus bg_runtime_allow(
    BgRuntimeLibrary *library,
    const DfAllowInput *input,
    DfStringBuffer *message,
    uint8_t *allowed
) {
    if (library == NULL || input == NULL || message == NULL || allowed == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->allow(library->runtime, input, message, allowed);
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
