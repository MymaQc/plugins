#include "dragonfly_plugin.h"

#include <stdint.h>
#include <string.h>

static const uint8_t plugin_id[] = "test:allower";
static const uint8_t denial[] = "denied by fixture";

static void *create(void) { return (void *)(uintptr_t)1; }
static DfStatus enable(void *instance, DfStringBuffer *error) { (void)instance; (void)error; return DF_STATUS_OK; }
static DfStatus disable(void *instance) { (void)instance; return DF_STATUS_OK; }
static DfStatus set_host(void *instance, const DfHostApiV27 *host) { (void)instance; (void)host; return DF_STATUS_OK; }
static void destroy(void *instance) { (void)instance; }

static DfStatus allow(void *instance, const DfAllowInput *input, DfStringBuffer *message, uint8_t *allowed) {
    (void)instance;
    if (input == NULL || message == NULL || allowed == NULL ||
        input->identity_json.len == 0 || input->client_json.len == 0 ||
        message->capacity < sizeof(denial) - 1) return DF_STATUS_ERROR;
    memcpy(message->data, denial, sizeof(denial) - 1);
    message->len = sizeof(denial) - 1;
    *allowed = 0;
    return DF_STATUS_OK;
}

static const DfPluginApiV11 api = {
    .header = {
        .abi_version = DF_ABI_VERSION,
        .struct_size = sizeof(DfPluginApiV11),
    },
    .plugin_id = { .data = plugin_id, .len = sizeof(plugin_id) - 1 },
    .create = create,
    .enable = enable,
    .disable = disable,
    .set_host = set_host,
    .destroy = destroy,
    .allow = allow,
};

const DfPluginApiV11 *df_plugin_entry_v11(void) { return &api; }
