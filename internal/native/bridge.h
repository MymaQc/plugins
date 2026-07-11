#ifndef BEDROCK_GOPHERS_BRIDGE_H
#define BEDROCK_GOPHERS_BRIDGE_H

#include "dragonfly_plugin.h"

typedef struct BgRuntimeLibrary BgRuntimeLibrary;

DfStatus bg_runtime_open(
    const char *library_path,
    const char *plugin_directory,
    BgRuntimeLibrary **out,
    uint8_t *error,
    uint64_t error_capacity
);

void bg_runtime_close(BgRuntimeLibrary *library);
DfStatus bg_runtime_enable(BgRuntimeLibrary *library);
void bg_runtime_disable(BgRuntimeLibrary *library);
uint64_t bg_runtime_plugin_count(const BgRuntimeLibrary *library);
uint64_t bg_runtime_subscriptions(const BgRuntimeLibrary *library);
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
uint64_t bg_runtime_handle_player_move_value(
    BgRuntimeLibrary *library,
    DfPlayerMoveInput input,
    uint8_t cancelled
);

#endif
