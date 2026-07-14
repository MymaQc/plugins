#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "dragonfly_plugin.h"

typedef struct {
    uint32_t explosion_drops;
    DfEntityId replacement_entity;
    DfBlockPos replacement_block;
} TestState;

static const uint8_t plugin_id[] = "world-event-test";

static int view_equal(DfStringView value, const char *expected) {
    size_t length = strlen(expected);
    return value.len == length && (length == 0 || memcmp(value.data, expected, length) == 0);
}

static int position_equal(DfBlockPos value, int32_t x, int32_t y, int32_t z) {
    return value.x == x && value.y == y && value.z == z;
}

static void *create(void) {
    TestState *state = calloc(1, sizeof(TestState));
    if (state != NULL) {
        state->replacement_entity.bytes[0] = 9;
        state->replacement_entity.generation = 99;
        state->replacement_block = (DfBlockPos){-9, 90, 9};
    }
    return state;
}

static void destroy(void *instance) {
    free(instance);
}

static DfStatus set_host(void *instance, const DfHostApiV27 *host) {
    return instance != NULL && host != NULL ? DF_STATUS_OK : DF_STATUS_ERROR;
}

static void drop_explosion(void *context) {
    ((TestState *)context)->explosion_drops++;
}

static DfStatus handle_event(void *opaque, DfEventId event, const void *raw_input, void *raw_state) {
    TestState *instance = opaque;
    if (instance == NULL || raw_input == NULL || raw_state == NULL) return DF_STATUS_ERROR;
    switch (event) {
        case DF_EVENT_WORLD_LIQUID_FLOW: {
            const DfWorldLiquidFlowInput *input = raw_input;
            if (input->invocation != 73 || !position_equal(input->from, 1, 2, 3) ||
                !position_equal(input->into, 4, 5, 6) || !view_equal(input->liquid.identifier, "minecraft:water") ||
                !view_equal(input->replaced.identifier, "minecraft:air")) return DF_STATUS_ERROR;
            ((DfWorldCancellableState *)raw_state)->cancelled = 1;
            return DF_STATUS_OK;
        }
        case DF_EVENT_WORLD_LIQUID_DECAY: {
            const DfWorldLiquidDecayInput *input = raw_input;
            if (input->invocation != 73 || input->after != NULL ||
                !view_equal(input->before.identifier, "minecraft:water")) return DF_STATUS_ERROR;
            ((DfWorldCancellableState *)raw_state)->cancelled = 1;
            return DF_STATUS_OK;
        }
        case DF_EVENT_WORLD_LIQUID_HARDEN: {
            const DfWorldLiquidHardenInput *input = raw_input;
            if (input->invocation != 73 || !view_equal(input->liquid_hardened.identifier, "minecraft:water") ||
                !view_equal(input->other_liquid.identifier, "minecraft:lava") ||
                !view_equal(input->new_block.identifier, "minecraft:stone")) return DF_STATUS_ERROR;
            ((DfWorldCancellableState *)raw_state)->cancelled = 1;
            return DF_STATUS_OK;
        }
        case DF_EVENT_WORLD_SOUND: {
            const DfWorldSoundInput *input = raw_input;
            if (input->invocation != 73 || input->sound.kind != 0 || input->position.y != 64) return DF_STATUS_ERROR;
            ((DfWorldCancellableState *)raw_state)->cancelled = 1;
            return DF_STATUS_OK;
        }
        case DF_EVENT_WORLD_FIRE_SPREAD: {
            const DfWorldFireSpreadInput *input = raw_input;
            if (input->invocation != 73 || !position_equal(input->from, 1, 2, 3) || !position_equal(input->to, 4, 5, 6)) return DF_STATUS_ERROR;
            ((DfWorldCancellableState *)raw_state)->cancelled = 1;
            return DF_STATUS_OK;
        }
        case DF_EVENT_WORLD_BLOCK_BURN:
        case DF_EVENT_WORLD_CROP_TRAMPLE:
        case DF_EVENT_WORLD_LEAVES_DECAY: {
            const DfWorldPositionInput *input = raw_input;
            if (input->invocation != 73 || !position_equal(input->position, 7, 8, 9)) return DF_STATUS_ERROR;
            ((DfWorldCancellableState *)raw_state)->cancelled = 1;
            return DF_STATUS_OK;
        }
        case DF_EVENT_WORLD_ENTITY_SPAWN:
        case DF_EVENT_WORLD_ENTITY_DESPAWN: {
            const DfWorldEntityInput *input = raw_input;
            return input->invocation == 73 && input->entity.bytes[0] == 3 && input->entity.generation == 4
                ? DF_STATUS_OK : DF_STATUS_ERROR;
        }
        case DF_EVENT_WORLD_EXPLOSION: {
            const DfWorldExplosionInput *input = raw_input;
            DfWorldExplosionState *state = raw_state;
            if (input->invocation != 73 || input->position.x != 1 || input->entity_count != 1 || input->block_count != 2 ||
                input->entities == NULL || input->entities[0].generation != 4 || input->blocks == NULL ||
                !position_equal(input->blocks[1], 4, 5, 6) || state->item_drop_chance != 0.25 || state->spawn_fire != 0)
                return DF_STATUS_ERROR;
            state->cancelled = 1;
            state->spawn_fire = 1;
            state->item_drop_chance = 0.75;
            state->replacement_entities = &instance->replacement_entity;
            state->replacement_entity_count = 1;
            state->replacement_blocks = &instance->replacement_block;
            state->replacement_block_count = 1;
            state->replacement_context = instance;
            state->replacement_drop = drop_explosion;
            return DF_STATUS_OK;
        }
        case DF_EVENT_WORLD_REDSTONE_UPDATE: {
            const DfWorldRedstoneUpdateInput *input = raw_input;
            if (input->invocation != 73 || !position_equal(input->position, 1, 2, 3) ||
                !position_equal(input->changed_neighbour, 4, 5, 6) || input->has_changed_neighbour != 1 ||
                input->changed_redstone_relevant != 1 || !position_equal(input->source, 7, 8, 9) || input->has_source != 1 ||
                !view_equal(input->before.identifier, "minecraft:redstone_wire") || input->after == NULL ||
                !view_equal(input->after->identifier, "minecraft:redstone_wire") || input->old_power != 2 ||
                input->new_power != 13 || input->current_tick != 1234 || input->cause != 2) return DF_STATUS_ERROR;
            ((DfWorldCancellableState *)raw_state)->cancelled = 1;
            return DF_STATUS_OK;
        }
        case DF_EVENT_WORLD_CLOSE: {
            const DfWorldCloseInput *input = raw_input;
            return input->invocation == 73 && instance->explosion_drops == 1 ? DF_STATUS_OK : DF_STATUS_ERROR;
        }
        default:
            return DF_STATUS_ERROR;
    }
}

static const DfPluginApiV9 api = {
    .header = {
        .abi_version = DF_ABI_VERSION,
        .struct_size = sizeof(DfPluginApiV9),
        .subscriptions = ((UINT64_C(1) << 13) - 1) << (DF_EVENT_WORLD_LIQUID_FLOW - 1),
    },
    .plugin_id = {plugin_id, sizeof(plugin_id) - 1},
    .create = create,
    .set_host = set_host,
    .destroy = destroy,
    .handle_event = handle_event,
};

const DfPluginApiV9 *df_plugin_entry_v9(void) {
    return &api;
}
