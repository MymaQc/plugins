#include "bridge.h"

#include <dlfcn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef DfStatus (*RuntimeCreateFn)(const DfRuntimeConfig *, DfRuntime **, uint8_t *, uint64_t);
typedef void (*RuntimeDestroyFn)(DfRuntime *);
typedef DfStatus (*RuntimeEnableFn)(DfRuntime *);
typedef void (*RuntimeDisableFn)(DfRuntime *);
typedef uint64_t (*RuntimeCountFn)(const DfRuntime *);
typedef DfStatus (*RuntimeCommandAtFn)(const DfRuntime *, uint64_t, DfCommandDescriptor *);
typedef DfStatus (*RuntimeCommandFn)(DfRuntime *, uint64_t, const DfCommandInput *, DfCommandState *);
typedef DfStatus (*RuntimeCommandEnumFn)(DfRuntime *, uint64_t, uint64_t, uint64_t, const DfCommandEnumContext *, DfStringBuffer *);
typedef DfStatus (*RuntimeMoveFn)(DfRuntime *, const DfPlayerMoveInput *, DfPlayerMoveState *);
typedef DfStatus (*RuntimeChatFn)(DfRuntime *, const DfPlayerChatInput *, DfPlayerChatState *);
typedef DfStatus (*RuntimeJoinFn)(DfRuntime *, const DfPlayerJoinInput *, DfPlayerJoinState *);
typedef DfStatus (*RuntimeQuitFn)(DfRuntime *, const DfPlayerQuitInput *, DfPlayerQuitState *);
typedef DfStatus (*RuntimeHurtFn)(DfRuntime *, const DfPlayerHurtInput *, DfPlayerHurtState *);
typedef DfStatus (*RuntimeHealFn)(DfRuntime *, const DfPlayerHealInput *, DfPlayerHealState *);
typedef DfStatus (*RuntimeBlockBreakFn)(DfRuntime *, const DfPlayerBlockBreakInput *, DfPlayerBlockBreakState *);
typedef DfStatus (*RuntimeBlockPlaceFn)(DfRuntime *, const DfPlayerBlockPlaceInput *, DfPlayerBlockPlaceState *);
typedef DfStatus (*RuntimeFoodLossFn)(DfRuntime *, const DfPlayerFoodLossInput *, DfPlayerFoodLossState *);
typedef DfStatus (*RuntimeDeathFn)(DfRuntime *, const DfPlayerDeathInput *, DfPlayerDeathState *);
typedef DfStatus (*RuntimeStartBreakFn)(DfRuntime *, const DfPlayerStartBreakInput *, DfPlayerStartBreakState *);
typedef DfStatus (*RuntimeFireExtinguishFn)(DfRuntime *, const DfPlayerFireExtinguishInput *, DfPlayerFireExtinguishState *);
typedef DfStatus (*RuntimeToggleSprintFn)(DfRuntime *, const DfPlayerToggleSprintInput *, DfPlayerToggleSprintState *);
typedef DfStatus (*RuntimeToggleSneakFn)(DfRuntime *, const DfPlayerToggleSneakInput *, DfPlayerToggleSneakState *);

struct BgRuntimeLibrary {
    void *handle;
    DfRuntime *runtime;
    RuntimeDestroyFn destroy;
    RuntimeEnableFn enable;
    RuntimeDisableFn disable;
    RuntimeCountFn plugin_count;
    RuntimeCountFn subscriptions;
    RuntimeCountFn command_count;
    RuntimeCommandAtFn command_at;
    RuntimeCommandFn handle_command;
    RuntimeCommandEnumFn command_enum_options;
    RuntimeMoveFn handle_move;
    RuntimeChatFn handle_chat;
    RuntimeJoinFn handle_join;
    RuntimeQuitFn handle_quit;
    RuntimeHurtFn handle_hurt;
    RuntimeHealFn handle_heal;
    RuntimeBlockBreakFn handle_block_break;
    RuntimeBlockPlaceFn handle_block_place;
    RuntimeFoodLossFn handle_food_loss;
    RuntimeDeathFn handle_death;
    RuntimeStartBreakFn handle_start_break;
    RuntimeFireExtinguishFn handle_fire_extinguish;
    RuntimeToggleSprintFn handle_toggle_sprint;
    RuntimeToggleSneakFn handle_toggle_sneak;
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
    RuntimeMoveFn handle_move = (RuntimeMoveFn) load_symbol(handle, "df_runtime_handle_player_move", error, error_capacity);
    RuntimeChatFn handle_chat = (RuntimeChatFn) load_symbol(handle, "df_runtime_handle_player_chat", error, error_capacity);
    RuntimeJoinFn handle_join = (RuntimeJoinFn) load_symbol(handle, "df_runtime_handle_player_join", error, error_capacity);
    RuntimeQuitFn handle_quit = (RuntimeQuitFn) load_symbol(handle, "df_runtime_handle_player_quit", error, error_capacity);
    RuntimeHurtFn handle_hurt = (RuntimeHurtFn) load_symbol(handle, "df_runtime_handle_player_hurt", error, error_capacity);
    RuntimeHealFn handle_heal = (RuntimeHealFn) load_symbol(handle, "df_runtime_handle_player_heal", error, error_capacity);
    RuntimeBlockBreakFn handle_block_break = (RuntimeBlockBreakFn) load_symbol(handle, "df_runtime_handle_player_block_break", error, error_capacity);
    RuntimeBlockPlaceFn handle_block_place = (RuntimeBlockPlaceFn) load_symbol(handle, "df_runtime_handle_player_block_place", error, error_capacity);
    RuntimeFoodLossFn handle_food_loss = (RuntimeFoodLossFn) load_symbol(handle, "df_runtime_handle_player_food_loss", error, error_capacity);
    RuntimeDeathFn handle_death = (RuntimeDeathFn) load_symbol(handle, "df_runtime_handle_player_death", error, error_capacity);
    RuntimeStartBreakFn handle_start_break = (RuntimeStartBreakFn) load_symbol(handle, "df_runtime_handle_player_start_break", error, error_capacity);
    RuntimeFireExtinguishFn handle_fire_extinguish = (RuntimeFireExtinguishFn) load_symbol(handle, "df_runtime_handle_player_fire_extinguish", error, error_capacity);
    RuntimeToggleSprintFn handle_toggle_sprint = (RuntimeToggleSprintFn) load_symbol(handle, "df_runtime_handle_player_toggle_sprint", error, error_capacity);
    RuntimeToggleSneakFn handle_toggle_sneak = (RuntimeToggleSneakFn) load_symbol(handle, "df_runtime_handle_player_toggle_sneak", error, error_capacity);
    if (create == NULL || destroy == NULL || enable == NULL || disable == NULL || plugin_count == NULL || subscriptions == NULL || command_count == NULL || command_at == NULL || handle_command == NULL || command_enum_options == NULL || handle_move == NULL || handle_chat == NULL || handle_join == NULL || handle_quit == NULL || handle_hurt == NULL || handle_heal == NULL || handle_block_break == NULL || handle_block_place == NULL || handle_food_loss == NULL || handle_death == NULL || handle_start_break == NULL || handle_fire_extinguish == NULL || handle_toggle_sprint == NULL || handle_toggle_sneak == NULL) {
        dlclose(handle);
        return DF_STATUS_ERROR;
    }

    BgRuntimeLibrary *library = calloc(1, sizeof(*library));
    if (library == NULL) {
        write_error(error, error_capacity, "allocate runtime bridge");
        dlclose(handle);
        return DF_STATUS_ERROR;
    }

    DfRuntimeConfig config = {
        .plugin_directory = {
            .data = (const uint8_t *) plugin_directory,
            .len = (uint64_t) strlen(plugin_directory),
        },
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
    library->handle_move = handle_move;
    library->handle_chat = handle_chat;
    library->handle_join = handle_join;
    library->handle_quit = handle_quit;
    library->handle_hurt = handle_hurt;
    library->handle_heal = handle_heal;
    library->handle_block_break = handle_block_break;
    library->handle_block_place = handle_block_place;
    library->handle_food_loss = handle_food_loss;
    library->handle_death = handle_death;
    library->handle_start_break = handle_start_break;
    library->handle_fire_extinguish = handle_fire_extinguish;
    library->handle_toggle_sprint = handle_toggle_sprint;
    library->handle_toggle_sneak = handle_toggle_sneak;
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

DfStatus bg_runtime_handle_player_move(
    BgRuntimeLibrary *library,
    const DfPlayerMoveInput *input,
    DfPlayerMoveState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->handle_move(library->runtime, input, state);
}

DfStatus bg_runtime_handle_player_chat(
    BgRuntimeLibrary *library,
    const DfPlayerChatInput *input,
    DfPlayerChatState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->handle_chat(library->runtime, input, state);
}

DfStatus bg_runtime_handle_player_join(
    BgRuntimeLibrary *library,
    const DfPlayerJoinInput *input,
    DfPlayerJoinState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->handle_join(library->runtime, input, state);
}

DfStatus bg_runtime_handle_player_quit(
    BgRuntimeLibrary *library,
    const DfPlayerQuitInput *input,
    DfPlayerQuitState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->handle_quit(library->runtime, input, state);
}

DfStatus bg_runtime_handle_player_hurt(
    BgRuntimeLibrary *library,
    const DfPlayerHurtInput *input,
    DfPlayerHurtState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->handle_hurt(library->runtime, input, state);
}

DfStatus bg_runtime_handle_player_heal(
    BgRuntimeLibrary *library,
    const DfPlayerHealInput *input,
    DfPlayerHealState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->handle_heal(library->runtime, input, state);
}

DfStatus bg_runtime_handle_player_block_break(
    BgRuntimeLibrary *library,
    const DfPlayerBlockBreakInput *input,
    DfPlayerBlockBreakState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->handle_block_break(library->runtime, input, state);
}

DfStatus bg_runtime_handle_player_block_place(
    BgRuntimeLibrary *library,
    const DfPlayerBlockPlaceInput *input,
    DfPlayerBlockPlaceState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->handle_block_place(library->runtime, input, state);
}

DfStatus bg_runtime_handle_player_food_loss(
    BgRuntimeLibrary *library,
    const DfPlayerFoodLossInput *input,
    DfPlayerFoodLossState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->handle_food_loss(library->runtime, input, state);
}

DfStatus bg_runtime_handle_player_death(
    BgRuntimeLibrary *library,
    const DfPlayerDeathInput *input,
    DfPlayerDeathState *state
) {
    if (library == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    return library->handle_death(library->runtime, input, state);
}

DfStatus bg_runtime_handle_player_start_break(BgRuntimeLibrary *library, const DfPlayerStartBreakInput *input, DfPlayerStartBreakState *state) {
    if (library == NULL || input == NULL || state == NULL) return DF_STATUS_ERROR;
    return library->handle_start_break(library->runtime, input, state);
}

DfStatus bg_runtime_handle_player_fire_extinguish(BgRuntimeLibrary *library, const DfPlayerFireExtinguishInput *input, DfPlayerFireExtinguishState *state) {
    if (library == NULL || input == NULL || state == NULL) return DF_STATUS_ERROR;
    return library->handle_fire_extinguish(library->runtime, input, state);
}
DfStatus bg_runtime_handle_player_toggle_sprint(BgRuntimeLibrary *library, const DfPlayerToggleSprintInput *input, DfPlayerToggleSprintState *state) { if (!library || !input || !state) return DF_STATUS_ERROR; return library->handle_toggle_sprint(library->runtime, input, state); }
DfStatus bg_runtime_handle_player_toggle_sneak(BgRuntimeLibrary *library, const DfPlayerToggleSneakInput *input, DfPlayerToggleSneakState *state) { if (!library || !input || !state) return DF_STATUS_ERROR; return library->handle_toggle_sneak(library->runtime, input, state); }

uint64_t bg_runtime_handle_player_move_value(
    BgRuntimeLibrary *library,
    DfPlayerMoveInput input,
    uint8_t cancelled
) {
    DfPlayerMoveState state = {.cancelled = cancelled};
    DfStatus status = bg_runtime_handle_player_move(library, &input, &state);
    return ((uint64_t) (uint32_t) status << 32u) | (uint64_t) state.cancelled;
}
