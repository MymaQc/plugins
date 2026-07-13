#include "bridge.h"

#include <dlfcn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

extern DfStatus bg_go_player_text(uint64_t context, DfPlayerId player, uint32_t kind, DfStringView message);
extern DfStatus bg_go_player_title(uint64_t context, DfPlayerId player, DfTitleView title);
extern DfStatus bg_go_player_transform(uint64_t context, DfPlayerId player, uint32_t kind, DfVec3 vector, double yaw, double pitch);
extern DfStatus bg_go_player_rotation(uint64_t context, DfPlayerId player, DfRotation *rotation);
extern DfStatus bg_go_player_state_set(uint64_t context, DfPlayerId player, uint32_t kind, DfPlayerStateValue value);
extern DfStatus bg_go_player_state_get(uint64_t context, DfPlayerId player, uint32_t kind, DfPlayerStateValue *value);
extern DfStatus bg_go_player_effect(uint64_t context, DfPlayerId player, uint32_t operation, DfEffectView effect);
extern DfStatus bg_go_player_entity_visibility(uint64_t context, DfPlayerId player, DfEntityId entity, uint8_t visible);

static DfStatus host_player_text(uint64_t context, DfPlayerId player, uint32_t kind, DfStringView message) {
    return bg_go_player_text(context, player, kind, message);
}

static DfStatus host_player_title(uint64_t context, DfPlayerId player, DfTitleView title) {
    return bg_go_player_title(context, player, title);
}

static DfStatus host_player_transform(uint64_t context, DfPlayerId player, uint32_t kind, DfVec3 vector, double yaw, double pitch) {
    return bg_go_player_transform(context, player, kind, vector, yaw, pitch);
}

static DfStatus host_player_rotation(uint64_t context, DfPlayerId player, DfRotation *rotation) {
    return bg_go_player_rotation(context, player, rotation);
}

static DfStatus host_player_state_set(uint64_t context, DfPlayerId player, uint32_t kind, DfPlayerStateValue value) {
    return bg_go_player_state_set(context, player, kind, value);
}

static DfStatus host_player_state_get(uint64_t context, DfPlayerId player, uint32_t kind, DfPlayerStateValue *value) {
    return bg_go_player_state_get(context, player, kind, value);
}

static DfStatus host_player_effect(uint64_t context, DfPlayerId player, uint32_t operation, DfEffectView effect) {
    return bg_go_player_effect(context, player, operation, effect);
}

static DfStatus host_player_entity_visibility(uint64_t context, DfPlayerId player, DfEntityId entity, uint8_t visible) {
    return bg_go_player_entity_visibility(context, player, entity, visible);
}

typedef DfStatus (*RuntimeCreateFn)(const DfRuntimeConfig *, DfRuntime **, uint8_t *, uint64_t);
typedef void (*RuntimeDestroyFn)(DfRuntime *);
typedef DfStatus (*RuntimeEnableFn)(DfRuntime *);
typedef void (*RuntimeDisableFn)(DfRuntime *);
typedef uint64_t (*RuntimeCountFn)(const DfRuntime *);
typedef DfStatus (*RuntimeCommandAtFn)(const DfRuntime *, uint64_t, DfCommandDescriptor *);
typedef DfStatus (*RuntimeCommandFn)(DfRuntime *, uint64_t, const DfCommandInput *, DfCommandState *);
typedef DfStatus (*RuntimeCommandEnumFn)(DfRuntime *, uint64_t, uint64_t, uint64_t, const DfCommandEnumContext *, DfStringBuffer *);
typedef DfStatus (*RuntimeEventFn)(DfRuntime *, DfEventId, const void *, void *);

struct BgRuntimeLibrary {
    void *handle;
    DfRuntime *runtime;
    DfHostApiV1 host_api;
    RuntimeDestroyFn destroy;
    RuntimeEnableFn enable;
    RuntimeDisableFn disable;
    RuntimeCountFn plugin_count;
    RuntimeCountFn subscriptions;
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
    RuntimeCountFn command_count = (RuntimeCountFn) load_symbol(handle, "df_runtime_command_count", error, error_capacity);
    RuntimeCommandAtFn command_at = (RuntimeCommandAtFn) load_symbol(handle, "df_runtime_command_at", error, error_capacity);
    RuntimeCommandFn handle_command = (RuntimeCommandFn) load_symbol(handle, "df_runtime_handle_command", error, error_capacity);
    RuntimeCommandEnumFn command_enum_options = (RuntimeCommandEnumFn) load_symbol(handle, "df_runtime_command_enum_options", error, error_capacity);
    RuntimeEventFn handle_event = (RuntimeEventFn) load_symbol(handle, "df_runtime_handle_event", error, error_capacity);
    if (create == NULL || destroy == NULL || enable == NULL || disable == NULL || plugin_count == NULL || subscriptions == NULL || command_count == NULL || command_at == NULL || handle_command == NULL || command_enum_options == NULL || handle_event == NULL) {
        dlclose(handle);
        return DF_STATUS_ERROR;
    }

    BgRuntimeLibrary *library = calloc(1, sizeof(*library));
    if (library == NULL) {
        write_error(error, error_capacity, "allocate runtime bridge");
        dlclose(handle);
        return DF_STATUS_ERROR;
    }

    library->host_api = (DfHostApiV1) {
        .abi_version = DF_ABI_VERSION,
        .struct_size = sizeof(DfHostApiV1),
        .context = host_context,
        .player_text = host_player_text,
        .player_title = host_player_title,
        .player_transform = host_player_transform,
        .player_rotation = host_player_rotation,
        .player_state_set = host_player_state_set,
        .player_state_get = host_player_state_get,
        .player_effect = host_player_effect,
        .player_entity_visibility = host_player_entity_visibility,
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
