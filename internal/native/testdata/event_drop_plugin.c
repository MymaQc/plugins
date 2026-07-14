#include <stdlib.h>
#include <string.h>

#include "dragonfly_plugin.h"

typedef struct {
    uint32_t transfer_drops;
    uint32_t command_drops;
} TestState;

static const uint8_t plugin_id[] = "event-drop-test";
static const uint8_t replacement_ip[] = {127, 0, 0, 2};
static const uint8_t invalid_ip[] = {127, 0, 0};
static const uint8_t replacement_argument[] = "changed";
static const DfStringView replacement_arguments[] = {
    {replacement_argument, sizeof(replacement_argument) - 1},
};

static void *create(void) {
    return calloc(1, sizeof(TestState));
}

static void destroy(void *instance) {
    free(instance);
}

static DfStatus set_host(void *instance, const DfHostApiV27 *host) {
    return instance != NULL && host != NULL ? DF_STATUS_OK : DF_STATUS_ERROR;
}

static void drop_transfer(void *context) {
    ((TestState *)context)->transfer_drops++;
}

static void drop_command(void *context) {
    ((TestState *)context)->command_drops++;
}

static int view_equal(DfStringView view, const char *value) {
    size_t length = strlen(value);
    return view.len == length && (length == 0 || memcmp(view.data, value, length) == 0);
}

static DfStatus handle_transfer(TestState *instance, DfPlayerTransferState *state) {
    int32_t mode = state->address.port;
    if (mode == 1003) {
        return instance->transfer_drops == 3 ? DF_STATUS_OK : DF_STATUS_ERROR;
    }
    if ((mode == 1001 && instance->transfer_drops != 1) ||
        (mode == 1002 && instance->transfer_drops != 2)) {
        return DF_STATUS_ERROR;
    }
    state->replacement_context = instance;
    state->replacement_drop = drop_transfer;
    state->address.zone = (DfStringView){NULL, 0};
    if (mode == 1001) {
        state->address.ip = (DfStringView){invalid_ip, sizeof(invalid_ip)};
        return DF_STATUS_OK;
    }
    state->address.ip = (DfStringView){replacement_ip, sizeof(replacement_ip)};
    state->address.port = 19133;
    return mode == 1002 ? DF_STATUS_ERROR : DF_STATUS_OK;
}

static DfStatus handle_command(
    TestState *instance,
    const DfPlayerCommandExecutionInput *input,
    DfPlayerCommandExecutionState *state
) {
    if (view_equal(input->command_name, "verify")) {
        return instance->command_drops == 3 ? DF_STATUS_OK : DF_STATUS_ERROR;
    }
    if ((view_equal(input->command_name, "invalid") && instance->command_drops != 1) ||
        (view_equal(input->command_name, "status") && instance->command_drops != 2)) {
        return DF_STATUS_ERROR;
    }
    state->replacement_arguments = replacement_arguments;
    state->replacement_argument_count = view_equal(input->command_name, "invalid") ? 2 : 1;
    state->replacement_context = instance;
    state->replacement_drop = drop_command;
    return view_equal(input->command_name, "status") ? DF_STATUS_ERROR : DF_STATUS_OK;
}

static DfStatus handle_event(void *opaque, DfEventId event, const void *input, void *state) {
    TestState *instance = opaque;
    if (instance == NULL || input == NULL || state == NULL) {
        return DF_STATUS_ERROR;
    }
    if (event == DF_EVENT_PLAYER_TRANSFER) {
        return handle_transfer(instance, state);
    }
    if (event == DF_EVENT_PLAYER_COMMAND_EXECUTION) {
        return handle_command(instance, input, state);
    }
    return DF_STATUS_ERROR;
}

static const DfPluginApiV7 api = {
    .header = {
        .abi_version = DF_ABI_VERSION,
        .struct_size = sizeof(DfPluginApiV7),
        .subscriptions = (UINT64_C(1) << 38) | (UINT64_C(1) << 39),
    },
    .plugin_id = {plugin_id, sizeof(plugin_id) - 1},
    .create = create,
    .set_host = set_host,
    .destroy = destroy,
    .handle_event = handle_event,
};

const DfPluginApiV7 *df_plugin_entry_v7(void) {
    return &api;
}
