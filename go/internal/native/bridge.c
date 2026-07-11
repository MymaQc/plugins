#include "bridge.h"

#include <dlfcn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef DfStatus (*RuntimeCreateFn)(const DfRuntimeConfig *, DfRuntime **, uint8_t *, uint64_t);
typedef void (*RuntimeDestroyFn)(DfRuntime *);
typedef uint64_t (*RuntimeCountFn)(const DfRuntime *);
typedef DfStatus (*RuntimeMoveFn)(DfRuntime *, const DfPlayerMoveInput *, DfPlayerMoveState *);

struct BgRuntimeLibrary {
    void *handle;
    DfRuntime *runtime;
    RuntimeDestroyFn destroy;
    RuntimeCountFn plugin_count;
    RuntimeCountFn subscriptions;
    RuntimeMoveFn handle_move;
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
    RuntimeCountFn plugin_count = (RuntimeCountFn) load_symbol(handle, "df_runtime_plugin_count", error, error_capacity);
    RuntimeCountFn subscriptions = (RuntimeCountFn) load_symbol(handle, "df_runtime_subscriptions", error, error_capacity);
    RuntimeMoveFn handle_move = (RuntimeMoveFn) load_symbol(handle, "df_runtime_handle_player_move", error, error_capacity);
    if (create == NULL || destroy == NULL || plugin_count == NULL || subscriptions == NULL || handle_move == NULL) {
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
    library->plugin_count = plugin_count;
    library->subscriptions = subscriptions;
    library->handle_move = handle_move;
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

uint64_t bg_runtime_plugin_count(const BgRuntimeLibrary *library) {
    return library == NULL ? 0 : library->plugin_count(library->runtime);
}

uint64_t bg_runtime_subscriptions(const BgRuntimeLibrary *library) {
    return library == NULL ? 0 : library->subscriptions(library->runtime);
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

uint64_t bg_runtime_handle_player_move_value(
    BgRuntimeLibrary *library,
    DfPlayerMoveInput input,
    uint8_t cancelled
) {
    DfPlayerMoveState state = {.cancelled = cancelled};
    DfStatus status = bg_runtime_handle_player_move(library, &input, &state);
    return ((uint64_t) (uint32_t) status << 32u) | (uint64_t) state.cancelled;
}
