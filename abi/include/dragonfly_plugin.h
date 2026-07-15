/* Private native transport ABI. Public C# API is generated from Dragonfly's Go AST. */
#ifndef BEDROCK_GOPHERS_DRAGONFLY_PLUGIN_H
#define BEDROCK_GOPHERS_DRAGONFLY_PLUGIN_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#define DF_ABI_VERSION 11u
// Host version 58 adds exact player block-interaction methods.
#define DF_HOST_ABI_VERSION 58u
#define DF_STATUS_OK 0
#define DF_STATUS_ERROR 1

typedef int32_t DfStatus;
typedef uint32_t DfEventId;

typedef struct { uint8_t bytes[16]; uint64_t generation; } DfPlayerId;
typedef struct { uint8_t bytes[16]; uint64_t generation; } DfEntityId;
typedef struct { uint64_t value; uint64_t generation; } DfEntityHandleId;
typedef struct { uint8_t bytes[16]; } DfUuid;
typedef uint64_t DfInvocationId;
typedef uint64_t DfBlockIteratorId;
typedef uint64_t DfEntityIteratorId;
typedef uint64_t DfPlayerIteratorId;
typedef struct { double x; double y; double z; } DfVec3;
typedef struct { DfVec3 min; DfVec3 max; } DfBBox;
typedef struct { double yaw; double pitch; } DfRotation;
typedef struct { DfVec3 position; DfVec3 velocity; DfRotation rotation; } DfPlayerKinematics;
typedef struct { int32_t x; int32_t y; int32_t z; } DfBlockPos;
typedef struct { int32_t min; int32_t max; } DfBlockRange;
typedef struct { uint64_t value; } DfWorldId;
typedef struct { const uint8_t *data; uint64_t len; } DfStringView;
typedef struct { uint8_t *data; uint64_t len; uint64_t capacity; } DfStringBuffer;
typedef struct {
    DfStringView network;
    DfStringView address;
    DfStringView ip;
    DfStringView zone;
    DfStringView identity_json;
    DfStringView client_json;
    int32_t port;
    uint8_t is_udp;
} DfAllowInput;
typedef void (*DfEventDropFn)(void *context);
typedef struct { DfStringView ip; int32_t port; DfStringView zone; } DfUDPAddrView;
#define DF_DAMAGE_SOURCE_REDUCED_BY_ARMOUR 1u
#define DF_DAMAGE_SOURCE_REDUCED_BY_RESISTANCE 2u
#define DF_DAMAGE_SOURCE_FIRE 4u
#define DF_DAMAGE_SOURCE_IGNORES_TOTEM 8u
#define DF_DAMAGE_SOURCE_FIRE_PROTECTION 16u
#define DF_DAMAGE_SOURCE_FEATHER_FALLING 32u
#define DF_DAMAGE_SOURCE_BLAST_PROTECTION 64u
#define DF_DAMAGE_SOURCE_PROJECTILE_PROTECTION 128u
#define DF_INVENTORY_MAIN 0u
#define DF_INVENTORY_ARMOUR 1u
#define DF_INVENTORY_OFFHAND 2u
#define DF_INVENTORY_ENDER_CHEST 3u
typedef struct { DfPlayerId player; uint32_t kind; uint32_t reserved; } DfInventoryId;
typedef struct { uint64_t offset; uint64_t len; } DfByteSpan;
typedef struct { uint32_t id; uint32_t level; } DfItemEnchantment;
typedef struct { int32_t metadata; uint32_t count; uint32_t damage; uint8_t unbreakable; int32_t anvil_cost; uint64_t identifier_len; uint64_t custom_name_len; uint64_t lore_bytes_len; uint64_t lore_count; uint64_t nbt_len; uint64_t values_nbt_len; uint64_t enchantment_count; } DfItemStackInfo;
typedef struct { uint64_t snapshot; DfItemStackInfo info; } DfItemStackSnapshot;
typedef struct { DfStringBuffer identifier; DfStringBuffer custom_name; DfStringBuffer lore_bytes; DfStringBuffer nbt; DfStringBuffer values_nbt; DfByteSpan *lore; uint64_t lore_capacity; DfItemEnchantment *enchantments; uint64_t enchantment_capacity; } DfItemStackData;
typedef struct { DfStringView identifier; int32_t metadata; uint32_t count; uint32_t damage; uint8_t unbreakable; int32_t anvil_cost; DfStringView custom_name; const DfStringView *lore; uint64_t lore_count; DfStringView nbt; DfStringView values_nbt; const DfItemEnchantment *enchantments; uint64_t enchantment_count; } DfItemStackViewV3;
/*
 * Event item views are borrowed only for the handle_event call. A non-null
 * replacement_drop transfers a replacement view lease to the host. The host
 * copies it synchronously and invokes replacement_drop exactly once before
 * event dispatch returns to Dragonfly, including status and validation errors.
 */
typedef void (*DfItemStackViewsDropFn)(void *context);
#define DF_WORLD_DIMENSION_OVERWORLD 0u
#define DF_WORLD_DIMENSION_NETHER 1u
#define DF_WORLD_DIMENSION_END 2u
#define DF_WORLD_PROVIDER_NOP 0u
#define DF_WORLD_PROVIDER_MCDB 1u
typedef struct { uint32_t id; uint8_t custom; uint8_t water_evaporates; uint8_t weather_cycle; uint8_t time_cycle; int32_t range_min; int32_t range_max; int64_t lava_spread_nanoseconds; } DfDimensionView;
typedef struct { uint32_t struct_size; uint32_t dimension; uint32_t provider_kind; uint32_t read_only; DfStringView provider_path; int64_t save_interval_nanoseconds; int64_t chunk_unload_interval_nanoseconds; int32_t random_tick_speed; uint32_t reserved; DfDimensionView dimension_view; } DfWorldConfigV1;
typedef struct { uint32_t id; uint8_t builtin; uint8_t food_regenerates; uint16_t reserved; double starvation_health_limit; int32_t fire_spread_increase; uint32_t reserved2; } DfDifficultyView;
typedef struct { DfStringBuffer identifier; DfStringBuffer properties_nbt; } DfBlockData;
typedef struct { DfStringView identifier; DfStringView properties_nbt; } DfBlockView;
#define DF_SET_BLOCK_DISABLE_BLOCK_UPDATES 1u
#define DF_SET_BLOCK_DISABLE_LIQUID_DISPLACEMENT 2u
#define DF_SET_BLOCK_DISABLE_REDSTONE_UPDATES 4u
#define DF_DAMAGE_SOURCE_CUSTOM 0u
#define DF_DAMAGE_SOURCE_ATTACK 1u
#define DF_DAMAGE_SOURCE_BLOCK 2u
#define DF_DAMAGE_SOURCE_DROWNING 3u
#define DF_DAMAGE_SOURCE_EXPLOSION 4u
#define DF_DAMAGE_SOURCE_FALL 5u
#define DF_DAMAGE_SOURCE_FIRE_KIND 6u
#define DF_DAMAGE_SOURCE_GLIDE 7u
#define DF_DAMAGE_SOURCE_INSTANT 8u
#define DF_DAMAGE_SOURCE_LAVA 9u
#define DF_DAMAGE_SOURCE_LIGHTNING 10u
#define DF_DAMAGE_SOURCE_MAGMA 11u
#define DF_DAMAGE_SOURCE_POISON 12u
#define DF_DAMAGE_SOURCE_PROJECTILE 13u
#define DF_DAMAGE_SOURCE_STARVATION 14u
#define DF_DAMAGE_SOURCE_SUFFOCATION 15u
#define DF_DAMAGE_SOURCE_THORNS 16u
#define DF_DAMAGE_SOURCE_VOID 17u
#define DF_DAMAGE_SOURCE_WITHER 18u
#define DF_HEALING_SOURCE_CUSTOM 0u
#define DF_HEALING_SOURCE_FOOD 1u
#define DF_HEALING_SOURCE_INSTANT 2u
#define DF_HEALING_SOURCE_REGENERATION 3u
typedef struct { DfStringView name; uint32_t kind; uint32_t flags; DfEntityId entity; DfEntityId secondary_entity; const DfBlockView *block; uint8_t data; } DfDamageSourceView;
typedef struct { DfStringView name; uint32_t kind; uint8_t data; } DfHealingSourceView;
#define DF_ENTITY_TEXT 0u
#define DF_ENTITY_LIGHTNING 1u
#define DF_ENTITY_TNT 2u
#define DF_ENTITY_EXPERIENCE_ORB 3u
#define DF_ENTITY_ITEM 4u
#define DF_ENTITY_FALLING_BLOCK 5u
#define DF_ENTITY_ARROW 6u
#define DF_ENTITY_EGG 7u
#define DF_ENTITY_SNOWBALL 8u
#define DF_ENTITY_ENDER_PEARL 9u
#define DF_ENTITY_BOTTLE_OF_ENCHANTING 10u
#define DF_ENTITY_SPLASH_POTION 11u
#define DF_ENTITY_LINGERING_POTION 12u
#define DF_ENTITY_CUSTOM 13u
#define DF_ENTITY_ARROW_CRITICAL 1u
#define DF_ENTITY_ARROW_DISABLE_PICKUP 2u
#define DF_ENTITY_ARROW_OBTAIN_ON_PICKUP 4u
#define DF_ENTITY_LIGHTNING_BLOCK_FIRE 1u
#define DF_ENTITY_ITEM_HAS_PICKUP_DELAY 1u
#define DF_ENTITY_HAS_VELOCITY 1u
#define DF_ENTITY_HAS_NAME_TAG 2u
#define DF_ENTITY_CAN_TELEPORT 4u
typedef struct { DfVec3 position; DfRotation rotation; DfVec3 velocity; DfStringView name_tag; } DfEntitySpawnOptions;
typedef struct { DfUuid id; DfEntitySpawnOptions options; DfStringView entity_type; uint64_t plugin; uint64_t local_type; uint64_t opaque; int64_t fire_duration_nanoseconds; int64_t age_nanoseconds; } DfEntityNewView;
#define DF_ENTITY_FAMILY_BASE 0u
#define DF_ENTITY_FAMILY_TICKING 1u
#define DF_ENTITY_FAMILY_LIVING 2u
#define DF_ENTITY_OPERATION_ADOPT 0u
#define DF_ENTITY_OPERATION_LOAD 1u
#define DF_ENTITY_OPERATION_SAVE 2u
#define DF_ENTITY_OPERATION_TICK 3u
#define DF_ENTITY_OPERATION_HURT 4u
#define DF_ENTITY_OPERATION_HEAL 5u
#define DF_ENTITY_OPERATION_DEATH 6u
#define DF_ENTITY_OPERATION_DESTROY 7u
#define DF_ENTITY_OPERATION_DECODE_NBT 8u
#define DF_ENTITY_OPERATION_ENCODE_NBT 9u
#define DF_ENTITY_OPERATION_OPEN 10u
#define DF_ENTITY_OPERATION_BBOX 11u
#define DF_ENTITY_OPERATION_CLOSE 12u
#define DF_ENTITY_OPERATION_HANDLE 13u
#define DF_ENTITY_OPERATION_POSITION 14u
#define DF_ENTITY_OPERATION_ROTATION 15u
#define DF_ENTITY_OPERATION_TICK_EXACT 16u
#define DF_ENTITY_OPERATION_RELEASE_OPEN 17u
#define DF_ENTITY_CAPABILITY_TICKER 1u
typedef uint64_t DfEntityInstanceId;
typedef struct { const uint8_t *data; uint64_t len; } DfBytesView;
typedef struct { uint8_t *data; uint64_t len; uint64_t capacity; } DfBytesBuffer;
typedef struct { DfStringView save_id; DfStringView network_id; uint64_t type_key; } DfEntityTypeDescriptorV2;
typedef struct { uint32_t kind; uint32_t flags; DfEntitySpawnOptions options; DfEntityId owner; double damage; uint64_t fuse_milliseconds; int32_t experience; uint32_t potion; int32_t punch_level; int32_t piercing_level; DfStringView text; DfStringView custom_type; DfEntityInstanceId custom_instance; const DfItemStackViewV3 *item; const DfBlockView *block; } DfEntitySpawnViewV3;
typedef struct { DfVec3 position; DfRotation rotation; DfVec3 velocity; uint32_t capabilities; DfWorldId world; DfStringBuffer entity_type; DfStringBuffer name_tag; } DfEntityState;
typedef struct { DfBytesView data; uint32_t version; uint32_t reserved; } DfEntityLoadInput;
typedef struct { uint64_t instance; } DfEntityLoadState;
typedef struct { DfBytesBuffer data; uint32_t version; uint32_t reserved; } DfEntitySaveState;
typedef struct { DfInvocationId invocation; DfEntityId entity; int64_t current; uint64_t age_milliseconds; } DfEntityTickInput;
typedef struct { uint8_t despawn; } DfEntityTickState;
typedef struct { DfVec3 position; DfVec3 velocity; DfRotation rotation; DfStringBuffer name; int64_t fire_duration_nanoseconds; int64_t age_nanoseconds; } DfEntityDataState;
typedef struct { DfInvocationId invocation; DfEntityHandleId handle; DfEntityDataState *data; DfBytesView nbt; int64_t current; } DfEntityExactInput;
typedef struct { uint64_t instance; uint32_t capabilities; uint32_t reserved; DfBBox bbox; DfEntityHandleId handle; DfVec3 position; DfRotation rotation; DfBytesBuffer nbt; } DfEntityExactState;
#define DF_PARTICLE_FLAME 0u
#define DF_PARTICLE_DUST 1u
#define DF_PARTICLE_BLOCK_BREAK 2u
#define DF_PARTICLE_PUNCH_BLOCK 3u
#define DF_PARTICLE_BLOCK_FORCE_FIELD 4u
#define DF_PARTICLE_BONE_MEAL 5u
#define DF_PARTICLE_NOTE 6u
#define DF_PARTICLE_DRAGON_EGG_TELEPORT 7u
#define DF_PARTICLE_EVAPORATE 8u
#define DF_PARTICLE_WATER_DRIP 9u
#define DF_PARTICLE_LAVA_DRIP 10u
#define DF_PARTICLE_LAVA 11u
#define DF_PARTICLE_DUST_PLUME 12u
#define DF_PARTICLE_HUGE_EXPLOSION 13u
#define DF_PARTICLE_ENDERMAN_TELEPORT 14u
#define DF_PARTICLE_SNOWBALL_POOF 15u
#define DF_PARTICLE_EGG_SMASH 16u
#define DF_PARTICLE_SPLASH 17u
#define DF_PARTICLE_EFFECT 18u
#define DF_PARTICLE_ENTITY_FLAME 19u
typedef struct { uint8_t r; uint8_t g; uint8_t b; uint8_t a; } DfRgba;
typedef struct { uint32_t kind; uint32_t data; int32_t pitch; DfRgba colour; DfBlockPos diff; const DfBlockView *block; } DfParticleViewV1;
typedef struct { uint32_t kind; uint32_t data; int32_t integer; uint32_t flags; double scalar; const DfBlockView *block; const DfItemStackViewV3 *item; } DfSoundViewV1;
#define DF_PLAYER_TRANSFORM_TELEPORT 0u
#define DF_PLAYER_TRANSFORM_MOVE 1u
#define DF_PLAYER_TRANSFORM_VELOCITY 2u
#define DF_PLAYER_TRANSFORM_DISPLACE 3u
#define DF_PLAYER_TEXT_MESSAGE 0u
#define DF_PLAYER_TEXT_TIP 1u
#define DF_PLAYER_TEXT_POPUP 2u
#define DF_PLAYER_TEXT_JUKEBOX_POPUP 3u
#define DF_PLAYER_TEXT_NAME_TAG 4u
#define DF_PLAYER_TEXT_DISCONNECT 5u
#define DF_PLAYER_TEXT_KICK 6u
#define DF_PLAYER_STATE_GAME_MODE 0u
#define DF_PLAYER_STATE_FOOD 3u
#define DF_PLAYER_STATE_MAX_HEALTH 4u
#define DF_PLAYER_STATE_HEALTH 5u
#define DF_PLAYER_STATE_EXPERIENCE_LEVEL 6u
#define DF_PLAYER_STATE_EXPERIENCE_PROGRESS 7u
#define DF_PLAYER_STATE_SCALE 8u
#define DF_PLAYER_STATE_INVISIBLE 9u
#define DF_PLAYER_STATE_IMMOBILE 10u
#define DF_PLAYER_OPERATION_HEAL 0u
#define DF_PLAYER_OPERATION_HURT 1u
#define DF_PLAYER_OPERATION_EXPERIENCE 2u
#define DF_EFFECT_SPEED 1
#define DF_EFFECT_SLOWNESS 2
#define DF_EFFECT_HASTE 3
#define DF_EFFECT_MINING_FATIGUE 4
#define DF_EFFECT_STRENGTH 5
#define DF_EFFECT_INSTANT_HEALTH 6
#define DF_EFFECT_INSTANT_DAMAGE 7
#define DF_EFFECT_JUMP_BOOST 8
#define DF_EFFECT_NAUSEA 9
#define DF_EFFECT_REGENERATION 10
#define DF_EFFECT_RESISTANCE 11
#define DF_EFFECT_FIRE_RESISTANCE 12
#define DF_EFFECT_WATER_BREATHING 13
#define DF_EFFECT_INVISIBILITY 14
#define DF_EFFECT_BLINDNESS 15
#define DF_EFFECT_NIGHT_VISION 16
#define DF_EFFECT_HUNGER 17
#define DF_EFFECT_WEAKNESS 18
#define DF_EFFECT_POISON 19
#define DF_EFFECT_WITHER 20
#define DF_EFFECT_HEALTH_BOOST 21
#define DF_EFFECT_ABSORPTION 22
#define DF_EFFECT_SATURATION 23
#define DF_EFFECT_LEVITATION 24
#define DF_EFFECT_FATAL_POISON 25
#define DF_EFFECT_CONDUIT_POWER 26
#define DF_EFFECT_SLOW_FALLING 27
#define DF_EFFECT_DARKNESS 30
#define DF_SOUND_KIND_ANVIL_BREAK 0u
#define DF_SOUND_KIND_ANVIL_LAND 1u
#define DF_SOUND_KIND_ANVIL_USE 2u
#define DF_SOUND_KIND_ARROW_HIT 3u
#define DF_SOUND_KIND_BARREL_CLOSE 4u
#define DF_SOUND_KIND_BARREL_OPEN 5u
#define DF_SOUND_KIND_BLAST_FURNACE_CRACKLE 6u
#define DF_SOUND_KIND_BOW_SHOOT 7u
#define DF_SOUND_KIND_BURNING 8u
#define DF_SOUND_KIND_BURP 9u
#define DF_SOUND_KIND_CAMPFIRE_CRACKLE 10u
#define DF_SOUND_KIND_CHEST_CLOSE 11u
#define DF_SOUND_KIND_CHEST_OPEN 12u
#define DF_SOUND_KIND_CLICK 13u
#define DF_SOUND_KIND_COMPOSTER_EMPTY 14u
#define DF_SOUND_KIND_COMPOSTER_FILL 15u
#define DF_SOUND_KIND_COMPOSTER_FILL_LAYER 16u
#define DF_SOUND_KIND_COMPOSTER_READY 17u
#define DF_SOUND_KIND_COPPER_SCRAPED 18u
#define DF_SOUND_KIND_CROSSBOW_SHOOT 19u
#define DF_SOUND_KIND_DECORATED_POT_INSERT_FAILED 20u
#define DF_SOUND_KIND_DENY 21u
#define DF_SOUND_KIND_DOOR_CRASH 22u
#define DF_SOUND_KIND_DROWNING 23u
#define DF_SOUND_KIND_ENDER_CHEST_CLOSE 24u
#define DF_SOUND_KIND_ENDER_CHEST_OPEN 25u
#define DF_SOUND_KIND_EXPERIENCE 26u
#define DF_SOUND_KIND_EXPLOSION 27u
#define DF_SOUND_KIND_FIRE_CHARGE 28u
#define DF_SOUND_KIND_FIRE_EXTINGUISH 29u
#define DF_SOUND_KIND_FIREWORK_BLAST 30u
#define DF_SOUND_KIND_FIREWORK_HUGE_BLAST 31u
#define DF_SOUND_KIND_FIREWORK_LAUNCH 32u
#define DF_SOUND_KIND_FIREWORK_TWINKLE 33u
#define DF_SOUND_KIND_FIZZ 34u
#define DF_SOUND_KIND_FURNACE_CRACKLE 35u
#define DF_SOUND_KIND_GHAST_SHOOT 36u
#define DF_SOUND_KIND_GHAST_WARNING 37u
#define DF_SOUND_KIND_GLASS_BREAK 38u
#define DF_SOUND_KIND_IGNITE 39u
#define DF_SOUND_KIND_ITEM_ADD 40u
#define DF_SOUND_KIND_ITEM_BREAK 41u
#define DF_SOUND_KIND_ITEM_FRAME_REMOVE 42u
#define DF_SOUND_KIND_ITEM_FRAME_ROTATE 43u
#define DF_SOUND_KIND_ITEM_THROW 44u
#define DF_SOUND_KIND_LECTERN_BOOK_PLACE 45u
#define DF_SOUND_KIND_LEVEL_UP 46u
#define DF_SOUND_KIND_LIGHTNING_EXPLODE 47u
#define DF_SOUND_KIND_LIGHTNING_THUNDER 48u
#define DF_SOUND_KIND_MUSIC_DISC_END 49u
#define DF_SOUND_KIND_POP 50u
#define DF_SOUND_KIND_POTION_BREWED 51u
#define DF_SOUND_KIND_POWER_OFF 52u
#define DF_SOUND_KIND_POWER_ON 53u
#define DF_SOUND_KIND_SIGN_WAXED 54u
#define DF_SOUND_KIND_SMOKER_CRACKLE 55u
#define DF_SOUND_KIND_STOP_USING_SPYGLASS 56u
#define DF_SOUND_KIND_TNT 57u
#define DF_SOUND_KIND_TELEPORT 58u
#define DF_SOUND_KIND_THUNDER 59u
#define DF_SOUND_KIND_TOTEM 60u
#define DF_SOUND_KIND_USE_SPYGLASS 61u
#define DF_SOUND_KIND_WAX_REMOVED 62u
#define DF_SOUND_KIND_WAXED_SIGN_FAILED_INTERACTION 63u
#define DF_SOUND_KIND_SHULKER_BOX_OPEN 64u
#define DF_SOUND_KIND_SHULKER_BOX_CLOSE 65u
#define DF_SOUND_KIND_ENDER_EYE_PLACED 66u
#define DF_SOUND_KIND_END_PORTAL_CREATED 67u
#define DF_SOUND_KIND_ATTACK 68u
#define DF_SOUND_KIND_FALL 69u
#define DF_SOUND_KIND_BLOCK_PLACE 70u
#define DF_SOUND_KIND_BLOCK_BREAKING 71u
#define DF_SOUND_KIND_DOOR_OPEN 72u
#define DF_SOUND_KIND_DOOR_CLOSE 73u
#define DF_SOUND_KIND_TRAPDOOR_OPEN 74u
#define DF_SOUND_KIND_TRAPDOOR_CLOSE 75u
#define DF_SOUND_KIND_FENCE_GATE_OPEN 76u
#define DF_SOUND_KIND_FENCE_GATE_CLOSE 77u
#define DF_SOUND_KIND_NOTE 78u
#define DF_SOUND_KIND_MUSIC_DISC_PLAY 79u
#define DF_SOUND_KIND_DECORATED_POT_INSERTED 80u
#define DF_SOUND_KIND_ITEM_USE_ON 81u
#define DF_SOUND_KIND_EQUIP_ITEM 82u
#define DF_SOUND_KIND_BUCKET_FILL 83u
#define DF_SOUND_KIND_BUCKET_EMPTY 84u
#define DF_SOUND_KIND_CROSSBOW_LOAD 85u
#define DF_SOUND_KIND_GOAT_HORN 86u

typedef struct { DfStringView text; DfStringView subtitle; DfStringView action_text; int64_t fade_in_duration_nanoseconds; int64_t duration_nanoseconds; int64_t fade_out_duration_nanoseconds; } DfTitleView;
typedef struct { DfStringView name; const DfStringView *lines; uint64_t line_count; uint8_t padding; uint8_t descending; } DfScoreboardView;
typedef struct { DfPlayerId player; DfStringView name; uint64_t latency_milliseconds; DfVec3 position; } DfPlayerSnapshot;
typedef struct { DfPlayerId player; DfStringBuffer name; uint64_t latency_milliseconds; DfVec3 position; } DfPlayerSnapshotBuffer;
/* The snapshot and its name are borrowed for the duration of the callback. */
typedef DfStatus (*DfFormResponseFn)(void *callback_context, DfInvocationId invocation, const DfPlayerSnapshot *submitter, uint32_t outcome, DfStringView response_json);
typedef void (*DfFormDropFn)(void *callback_context);
/*
 * A structurally valid form transfers callback_context ownership to the host,
 * even when player_form_send returns an error. The host invokes exactly one of
 * response or drop. Synchronous send failures invoke drop before returning.
 */
typedef struct { DfStringView request_json; void *callback_context; DfFormResponseFn response; DfFormDropFn drop; } DfFormView;
#define DF_FORM_RESPONSE_SUBMITTED 0u
#define DF_FORM_RESPONSE_CLOSED 1u
typedef struct { double number; int64_t integer; } DfPlayerStateValue;
typedef struct { double healed; } DfPlayerHealResult;
typedef struct { double damage; uint8_t vulnerable; } DfPlayerHurtResult;
#define DF_PLAYER_EFFECT_ADD 0u
#define DF_PLAYER_EFFECT_REMOVE 1u
typedef struct { int32_t effect_type; int32_t level; int64_t duration_nanoseconds; double potency; uint8_t ambient; uint8_t particles_hidden; uint8_t infinite; int64_t tick; } DfEffectView;
typedef struct { DfEffectView *data; uint64_t len; uint64_t capacity; } DfEffectBuffer;
typedef struct { uint32_t width; uint32_t height; uint32_t animation_type; int64_t frame_count; int64_t expression; uint64_t pixels_len; } DfSkinAnimationInfo;
typedef struct { uint32_t width; uint32_t height; uint8_t persona; uint64_t play_fab_id_len; uint64_t full_id_len; uint64_t pixels_len; uint64_t model_default_len; uint64_t model_animated_face_len; uint64_t model_len; uint32_t cape_width; uint32_t cape_height; uint64_t cape_pixels_len; uint64_t animation_count; } DfSkinInfo;
typedef struct { DfStringBuffer play_fab_id; DfStringBuffer full_id; DfStringBuffer pixels; DfStringBuffer model_default; DfStringBuffer model_animated_face; DfStringBuffer model; DfStringBuffer cape_pixels; DfStringBuffer *animation_pixels; uint64_t animation_capacity; } DfSkinData;
typedef struct { uint32_t width; uint32_t height; uint32_t animation_type; int64_t frame_count; int64_t expression; DfStringView pixels; } DfSkinAnimationView;
typedef struct { uint32_t width; uint32_t height; uint8_t persona; DfStringView play_fab_id; DfStringView full_id; DfStringView pixels; DfStringView model_default; DfStringView model_animated_face; DfStringView model; uint32_t cape_width; uint32_t cape_height; DfStringView cape_pixels; const DfSkinAnimationView *animations; uint64_t animation_count; } DfSkinView;
typedef DfStatus (*DfHostPlayerTextFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfStringView message);
typedef DfStatus (*DfHostPlayerTitleFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfTitleView title);
typedef DfStatus (*DfHostPlayerScoreboardFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfScoreboardView scoreboard);
typedef DfStatus (*DfHostPlayerScoreboardRemoveFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player);
typedef DfStatus (*DfHostPlayerFormSendFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfFormView *form);
typedef DfStatus (*DfHostPlayerFormCloseFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player);
typedef DfStatus (*DfHostPlayerTransformFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfVec3 vector, double yaw, double pitch);
typedef DfStatus (*DfHostPlayerKinematicsFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfPlayerKinematics *kinematics);
typedef DfStatus (*DfHostPlayerKnockBackFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfVec3 source, double force, double height);
typedef DfStatus (*DfHostPlayerStateSetFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue value);
typedef DfStatus (*DfHostPlayerStateGetFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue *value);
typedef DfStatus (*DfHostPlayerActionFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue value, DfPlayerStateValue *result);
typedef DfStatus (*DfHostPlayerHealFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, double health, const DfHealingSourceView *source, DfPlayerHealResult *result);
typedef DfStatus (*DfHostPlayerHurtFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, double damage, const DfDamageSourceView *source, DfPlayerHurtResult *result);
typedef DfStatus (*DfHostPlayerFinalDamageFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, double damage, const DfDamageSourceView *source, double *result);
typedef DfStatus (*DfHostPlayerUsingItemFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint8_t *using_item);
typedef DfStatus (*DfHostPlayerSleepingFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfBlockPos *position, uint8_t *sleeping);
typedef DfStatus (*DfHostPlayerDeathPositionFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfVec3 *position, DfDimensionView *dimension, uint8_t *found);
typedef DfStatus (*DfHostPlayerBlockActionFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfBlockPos position, int32_t face, DfVec3 click_position);
typedef DfStatus (*DfHostPlayerEffectFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t operation, DfEffectView effect);
typedef DfStatus (*DfHostPlayerEffectsFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfEffectBuffer *output);
typedef DfStatus (*DfHostPlayerEffectsClearFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player);
typedef DfStatus (*DfHostPlayerEntityVisibilityFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfEntityId entity, uint8_t visible);
/* Skin snapshots freeze one skin across metadata and data reads. Open owns a snapshot until close. */
/* A zero-length buffer may have a null data pointer. Read performs no partial writes on insufficient capacity. */
typedef DfStatus (*DfHostPlayerSkinOpenFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint64_t *snapshot, DfSkinInfo *info);
typedef DfStatus (*DfHostPlayerSkinAnimationInfoFn)(uint64_t context, DfInvocationId invocation, uint64_t snapshot, uint64_t index, DfSkinAnimationInfo *info);
typedef DfStatus (*DfHostPlayerSkinReadFn)(uint64_t context, DfInvocationId invocation, uint64_t snapshot, DfSkinData *data);
typedef void (*DfHostPlayerSkinCloseFn)(uint64_t context, DfInvocationId invocation, uint64_t snapshot);
typedef DfStatus (*DfHostPlayerSkinSetFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfSkinView *skin);
typedef DfStatus (*DfHostPlayerTransferFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfWorldId world, DfVec3 position);
typedef DfStatus (*DfHostSkinSnapshotInfoFn)(uint64_t context, DfInvocationId invocation, uint64_t snapshot, DfSkinInfo *info);
typedef DfStatus (*DfHostSkinSnapshotSetFn)(uint64_t context, DfInvocationId invocation, uint64_t snapshot, const DfSkinView *skin);
typedef DfStatus (*DfHostInventorySizeFn)(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, uint32_t *size);
typedef DfStatus (*DfHostInventoryItemOpenFn)(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, uint32_t slot, uint64_t *snapshot, DfItemStackInfo *info);
typedef DfStatus (*DfHostPlayerHeldItemOpenFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t hand, uint64_t *snapshot, DfItemStackInfo *info);
typedef DfStatus (*DfHostItemStackReadFn)(uint64_t context, DfInvocationId invocation, uint64_t snapshot, DfItemStackData *data);
typedef void (*DfHostItemStackCloseFn)(uint64_t context, DfInvocationId invocation, uint64_t snapshot);
typedef DfStatus (*DfHostInventoryItemSetFn)(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, uint32_t slot, const DfItemStackViewV3 *item);
typedef DfStatus (*DfHostInventoryItemAddFn)(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, const DfItemStackViewV3 *item, uint32_t *added);
typedef DfStatus (*DfHostInventoryClearSlotFn)(uint64_t context, DfInvocationId invocation, DfInventoryId inventory, uint32_t slot);
typedef DfStatus (*DfHostInventoryClearFn)(uint64_t context, DfInvocationId invocation, DfInventoryId inventory);
typedef DfStatus (*DfHostPlayerHeldItemsSetFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfItemStackViewV3 *main_hand, const DfItemStackViewV3 *off_hand);
typedef DfStatus (*DfHostPlayerHeldSlotSetFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t slot);
/* Opens one atomic held-items read. Each returned snapshot is independently owned until item_stack_close. */
typedef DfStatus (*DfHostPlayerHeldItemsOpenFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfItemStackSnapshot *main_hand, DfItemStackSnapshot *off_hand);
typedef DfStatus (*DfHostWorldNameFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfStringBuffer *name);
typedef DfStatus (*DfHostWorldUnloadFn)(uint64_t context, DfInvocationId invocation, DfWorldId world);
typedef DfStatus (*DfHostWorldSaveFn)(uint64_t context, DfInvocationId invocation, DfWorldId world);
typedef DfStatus (*DfHostWorldBlockGetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, DfBlockData *block);
typedef DfStatus (*DfHostBlockByNameFn)(uint64_t context, DfStringView name, DfStringView properties_nbt, uint8_t *found, DfBlockData *block);
typedef DfStatus (*DfHostWorldBlockLoadedFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *loaded, DfBlockData *block);
typedef DfStatus (*DfHostWorldLiquidGetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *found, DfBlockData *block);
typedef DfStatus (*DfHostWorldLiquidSetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, const DfBlockView *liquid);
typedef DfStatus (*DfHostWorldBlockSetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, const DfBlockView *block, uint32_t flags);
typedef DfStatus (*DfHostWorldBlockUpdateScheduleFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, const DfBlockView *block, int64_t delay_nanoseconds);
typedef DfStatus (*DfHostWorldBiomeGetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, int32_t *biome);
typedef DfStatus (*DfHostWorldBiomeSetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, int32_t biome);
typedef DfStatus (*DfHostWorldTemperatureFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, double *temperature);
typedef DfStatus (*DfHostWorldWeatherAtFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *value);
typedef DfStatus (*DfHostWorldWeatherFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, uint8_t *value);
typedef DfStatus (*DfHostWorldCurrentTickFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t *tick);
typedef DfStatus (*DfHostWorldRangeFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockRange *range);
typedef DfStatus (*DfHostWorldBlocksWithinOpenFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, int32_t radius, const DfBlockView *blocks, uint64_t block_count, DfBlockIteratorId *iterator);
typedef DfStatus (*DfHostWorldBlocksWithinNextFn)(uint64_t context, DfInvocationId invocation, DfBlockIteratorId iterator, DfBlockPos *position, uint8_t *found);
typedef void (*DfHostWorldBlocksWithinCloseFn)(uint64_t context, DfInvocationId invocation, DfBlockIteratorId iterator);
typedef DfStatus (*DfHostWorldHeightFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, int32_t x, int32_t z, int32_t *height);
typedef DfStatus (*DfHostWorldLightFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, uint8_t *level);
typedef DfStatus (*DfHostWorldTimeGetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t *time);
typedef DfStatus (*DfHostWorldTimeSetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t time);
typedef DfStatus (*DfHostWorldSpawnGetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos *position);
typedef DfStatus (*DfHostWorldSpawnSetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position);
typedef DfStatus (*DfHostWorldPlayerSpawnGetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfUuid player, DfBlockPos *position);
typedef DfStatus (*DfHostWorldPlayerSpawnSetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfUuid player, DfBlockPos position);
typedef DfStatus (*DfHostWorldDimensionGetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfDimensionView *dimension);
typedef DfStatus (*DfHostWorldBoolGetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, uint8_t *value);
typedef DfStatus (*DfHostWorldBoolSetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, uint8_t value);
typedef DfStatus (*DfHostWorldDurationSetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t duration_nanoseconds);
typedef DfStatus (*DfHostWorldGameModeGetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t *game_mode);
typedef DfStatus (*DfHostWorldGameModeSetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t game_mode);
typedef DfStatus (*DfHostWorldIntSetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, int32_t value);
typedef DfStatus (*DfHostWorldDifficultyGetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfDifficultyView *difficulty);
typedef DfStatus (*DfHostWorldDifficultySetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfDifficultyView difficulty);
typedef DfStatus (*DfHostPlayerPacketWriteFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint64_t packet);
typedef DfStatus (*DfHostPlayerStringGetFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfStringBuffer *value);
typedef DfStatus (*DfHostPlayerToastFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfStringView title, DfStringView message);
typedef DfStatus (*DfHostPlayerCooldownFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t operation, DfStringView identifier, int32_t metadata, int64_t duration_nanoseconds, uint8_t *active);
#define DF_PLAYER_COOLDOWN_HAS 0u
#define DF_PLAYER_COOLDOWN_SET 1u
typedef DfStatus (*DfHostWorldEntitySpawnFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, const DfEntitySpawnViewV3 *entity, DfEntityId *output);
typedef DfStatus (*DfHostEntityStateFn)(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfEntityState *state);
typedef DfStatus (*DfHostEntityPlayerFn)(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfPlayerSnapshotBuffer *output);
typedef DfStatus (*DfHostEntityTeleportFn)(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfVec3 position);
typedef DfStatus (*DfHostEntityVelocitySetFn)(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfVec3 velocity);
typedef DfStatus (*DfHostEntityNameTagSetFn)(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfStringView name_tag);
typedef DfStatus (*DfHostEntityDespawnFn)(uint64_t context, DfInvocationId invocation, DfEntityId entity);
typedef DfStatus (*DfHostWorldParticleAddFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfVec3 position, const DfParticleViewV1 *particle);
typedef DfStatus (*DfHostWorldSoundPlayFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfVec3 position, const DfSoundViewV1 *sound);
typedef DfStatus (*DfHostPlayerSoundPlayFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfSoundViewV1 *sound);
typedef DfStatus (*DfHostWorldCurrentFn)(uint64_t context, DfInvocationId invocation, DfWorldId *world);
typedef DfStatus (*DfHostWorldEntityIteratorOpenFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, uint8_t players_only, DfEntityIteratorId *iterator);
typedef DfStatus (*DfHostWorldEntitiesWithinOpenFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBBox box, DfEntityIteratorId *iterator);
typedef DfStatus (*DfHostWorldEntityIteratorNextFn)(uint64_t context, DfInvocationId invocation, DfEntityIteratorId iterator, DfEntityId *entity, uint8_t *found);
typedef void (*DfHostWorldEntityIteratorCloseFn)(uint64_t context, DfInvocationId invocation, DfEntityIteratorId iterator);
typedef DfStatus (*DfHostServerPlayersOpenFn)(uint64_t context, DfInvocationId invocation, DfPlayerIteratorId *iterator);
typedef DfStatus (*DfHostServerPlayersNextFn)(uint64_t context, DfInvocationId invocation, DfPlayerIteratorId iterator, DfInvocationId *player_invocation, DfPlayerSnapshotBuffer *player, uint8_t *found);
typedef void (*DfHostServerPlayersCloseFn)(uint64_t context, DfInvocationId invocation, DfPlayerIteratorId iterator);
typedef DfStatus (*DfHostServerPlayerFn)(uint64_t context, DfUuid uuid, DfEntityHandleId *player, uint8_t *found);
typedef DfStatus (*DfHostServerPlayerByNameFn)(uint64_t context, DfStringView name, DfEntityHandleId *player, uint8_t *found);
typedef DfStatus (*DfHostServerCountFn)(uint64_t context, int64_t *count);
typedef DfStatus (*DfHostServerPlayerByXuidFn)(uint64_t context, DfStringView xuid, DfEntityHandleId *player, uint8_t *found);
typedef DfStatus (*DfHostPlayerXuidFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfStringBuffer *xuid);
typedef DfStatus (*DfHostServerWorldFn)(uint64_t context, uint32_t dimension, DfWorldId *world);
#define DF_WORLD_TASK_PHASE_EXECUTE 0u
#define DF_WORLD_TASK_PHASE_COMPLETE 1u
#define DF_WORLD_TASK_SUCCESS 0u
#define DF_WORLD_TASK_CANCELLED 1u
#define DF_WORLD_TASK_WORLD_CLOSED 2u
#define DF_WORLD_TASK_PANICKED 3u
#define DF_WORLD_TASK_FAILED 4u
typedef DfStatus (*DfHostWorldScheduleFn)(uint64_t context, DfWorldId world, uint64_t plugin, uint64_t callback, int64_t delay_nanoseconds);
typedef DfStatus (*DfHostWorldTaskCancelFn)(uint64_t context, uint64_t plugin, uint64_t callback, uint8_t *cancelled);
typedef DfStatus (*DfHostWorldNewFn)(uint64_t context, const DfWorldConfigV1 *config, DfWorldId *world);
typedef DfStatus (*DfHostEntityHandleFn)(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfEntityHandleId *handle);
typedef DfStatus (*DfHostEntityHandleEntityFn)(uint64_t context, DfInvocationId invocation, DfEntityHandleId handle, DfEntityId *entity, uint8_t *found);
typedef DfStatus (*DfHostEntityHandleUuidFn)(uint64_t context, DfEntityHandleId handle, DfUuid *uuid);
typedef DfStatus (*DfHostEntityHandleClosedFn)(uint64_t context, DfEntityHandleId handle, uint8_t *closed);
typedef DfStatus (*DfHostEntityHandleCloseFn)(uint64_t context, DfEntityHandleId handle);
typedef DfStatus (*DfHostWorldEntityRemoveFn)(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfEntityHandleId *handle);
typedef DfStatus (*DfHostWorldEntityAddFn)(uint64_t context, DfInvocationId invocation, DfEntityHandleId handle, const DfVec3 *position, DfEntityId *entity);
typedef DfStatus (*DfHostEntityNewFn)(uint64_t context, const DfEntityNewView *entity, DfEntityHandleId *handle);
typedef DfStatus (*DfHostEntityHandleTypeFn)(uint64_t context, DfEntityHandleId handle, DfStringBuffer *entity_type);
typedef struct {
    uint32_t kind;
    uint32_t _reserved;
    int64_t signed_value;
    uint64_t unsigned_value;
    double number;
    double x;
    double y;
    double z;
    DfUuid uuid;
    DfStringBuffer data;
} DfPacketFieldValue;
#define DF_PACKET_FIELD_INVALID 0u
#define DF_PACKET_FIELD_BOOL 1u
#define DF_PACKET_FIELD_SIGNED 2u
#define DF_PACKET_FIELD_UNSIGNED 3u
#define DF_PACKET_FIELD_FLOAT 4u
#define DF_PACKET_FIELD_STRING 5u
#define DF_PACKET_FIELD_BYTES 6u
#define DF_PACKET_FIELD_VEC2 7u
#define DF_PACKET_FIELD_VEC3 8u
#define DF_PACKET_FIELD_UUID 9u
#define DF_PACKET_FIELD_JSON 10u
typedef DfStatus (*DfHostPacketFieldGetFn)(uint64_t context, uint64_t packet, uint32_t field, DfPacketFieldValue *value);
typedef DfStatus (*DfHostPacketFieldSetFn)(uint64_t context, uint64_t packet, uint32_t field, const DfPacketFieldValue *value);
typedef struct {
    uint32_t abi_version;
    uint32_t struct_size;
    uint64_t context;
    DfHostPlayerTextFn player_text;
    DfHostPlayerTitleFn player_title;
    DfHostPlayerTransformFn player_transform;
    DfHostPlayerKinematicsFn player_kinematics;
    DfHostPlayerStateSetFn player_state_set;
    DfHostPlayerStateGetFn player_state_get;
    DfHostPlayerEffectFn player_effect;
    DfHostPlayerEntityVisibilityFn player_entity_visibility;
    DfHostPlayerSkinOpenFn player_skin_open;
    DfHostPlayerSkinAnimationInfoFn player_skin_animation_info;
    DfHostPlayerSkinReadFn player_skin_read;
    DfHostPlayerSkinCloseFn player_skin_close;
    DfHostPlayerSkinSetFn player_skin_set;
    DfHostInventorySizeFn inventory_size;
    DfHostInventoryItemOpenFn inventory_item_open;
    DfHostPlayerHeldItemOpenFn player_held_item_open;
    DfHostItemStackReadFn item_stack_read;
    DfHostItemStackCloseFn item_stack_close;
    DfHostInventoryItemSetFn inventory_item_set;
    DfHostInventoryItemAddFn inventory_item_add;
    DfHostInventoryClearSlotFn inventory_clear_slot;
    DfHostInventoryClearFn inventory_clear;
    DfHostPlayerHeldItemsSetFn player_held_items_set;
    DfHostPlayerHeldSlotSetFn player_held_slot_set;
    DfHostPlayerScoreboardFn player_scoreboard;
    DfHostPlayerScoreboardRemoveFn player_scoreboard_remove;
    DfHostPlayerFormSendFn player_form_send;
    DfHostPlayerFormCloseFn player_form_close;
    DfHostWorldNameFn world_name;
    DfHostWorldUnloadFn world_unload;
    DfHostWorldSaveFn world_save;
    DfHostWorldBlockGetFn world_block_get;
    DfHostWorldBlockSetFn world_block_set;
    DfHostWorldTimeGetFn world_time_get;
    DfHostWorldTimeSetFn world_time_set;
    DfHostWorldSpawnGetFn world_spawn_get;
    DfHostWorldSpawnSetFn world_spawn_set;
    DfHostWorldEntitySpawnFn world_entity_spawn;
    DfHostEntityStateFn entity_state;
    DfHostEntityTeleportFn entity_teleport;
    DfHostEntityVelocitySetFn entity_velocity_set;
    DfHostEntityNameTagSetFn entity_name_tag_set;
    DfHostEntityDespawnFn entity_despawn;
    DfHostWorldParticleAddFn world_particle_add;
    DfHostWorldSoundPlayFn world_sound_play;
    DfHostPlayerSoundPlayFn player_sound_play;
    DfHostPlayerHealFn player_heal;
    DfHostPlayerHurtFn player_hurt;
    DfHostSkinSnapshotInfoFn skin_snapshot_info;
    DfHostSkinSnapshotSetFn skin_snapshot_set;
    DfHostPlayerTransferFn player_transfer;
    DfHostPlayerEffectsFn player_effects;
    DfHostPlayerEffectsClearFn player_effects_clear;
    DfHostWorldLiquidGetFn world_liquid_get;
    DfHostPlayerActionFn player_action;
    DfHostWorldRangeFn world_range;
    DfHostWorldBlockLoadedFn world_block_loaded;
    DfHostWorldBlocksWithinOpenFn world_blocks_within_open;
    DfHostWorldBlocksWithinNextFn world_blocks_within_next;
    DfHostWorldBlocksWithinCloseFn world_blocks_within_close;
    DfHostWorldHeightFn world_highest_light_blocker;
    DfHostWorldHeightFn world_highest_block;
    DfHostWorldLightFn world_light;
    DfHostWorldLightFn world_sky_light;
    DfHostWorldLiquidSetFn world_liquid_set;
    DfHostWorldBlockUpdateScheduleFn world_block_update_schedule;
    DfHostWorldBiomeGetFn world_biome_get;
    DfHostWorldBiomeSetFn world_biome_set;
    DfHostWorldTemperatureFn world_temperature;
    DfHostWorldWeatherAtFn world_raining_at;
    DfHostWorldWeatherAtFn world_snowing_at;
    DfHostWorldWeatherAtFn world_thundering_at;
    DfHostWorldWeatherFn world_raining;
    DfHostWorldWeatherFn world_thundering;
    DfHostWorldCurrentTickFn world_current_tick;
    DfHostPlayerHeldItemsOpenFn player_held_items_open;
    DfHostEntityPlayerFn entity_player;
    DfHostWorldCurrentFn world_current;
    DfHostWorldEntityIteratorOpenFn world_entity_iterator_open;
    DfHostWorldEntityIteratorNextFn world_entity_iterator_next;
    DfHostWorldEntityIteratorCloseFn world_entity_iterator_close;
    DfHostEntityHandleFn entity_handle;
    DfHostEntityHandleEntityFn entity_handle_entity;
    DfHostEntityHandleUuidFn entity_handle_uuid;
    DfHostEntityHandleClosedFn entity_handle_closed;
    DfHostEntityHandleCloseFn entity_handle_close;
    DfHostWorldEntityRemoveFn world_entity_remove;
    DfHostWorldEntityAddFn world_entity_add;
    DfHostServerPlayersOpenFn server_players_open;
    DfHostServerPlayersNextFn server_players_next;
    DfHostServerPlayersCloseFn server_players_close;
    DfHostServerPlayerFn server_player;
    DfHostServerPlayerByNameFn server_player_by_name;
    DfHostServerCountFn server_max_player_count;
    DfHostServerCountFn server_player_count;
    DfHostServerPlayerByXuidFn server_player_by_xuid;
    DfHostPlayerXuidFn player_xuid;
    DfHostServerWorldFn server_world;
    DfHostWorldScheduleFn world_schedule;
    DfHostWorldNewFn world_new;
    DfHostWorldEntitiesWithinOpenFn world_entities_within_open;
    DfHostBlockByNameFn block_by_name;
    DfHostEntityNewFn entity_new;
    DfHostEntityHandleTypeFn entity_handle_type;
    DfHostWorldTaskCancelFn world_task_cancel;
    DfHostPacketFieldGetFn packet_field_get;
    DfHostPacketFieldSetFn packet_field_set;
    DfHostWorldDimensionGetFn world_dimension_get;
    DfHostWorldBoolGetFn world_time_cycle_get;
    DfHostWorldBoolSetFn world_time_cycle_set;
    DfHostWorldDurationSetFn world_required_sleep_duration_set;
    DfHostWorldGameModeGetFn world_default_game_mode_get;
    DfHostWorldGameModeSetFn world_default_game_mode_set;
    DfHostWorldIntSetFn world_tick_range_set;
    DfHostWorldDifficultyGetFn world_difficulty_get;
    DfHostWorldDifficultySetFn world_difficulty_set;
    DfHostPlayerPacketWriteFn player_packet_write;
    DfHostWorldPlayerSpawnGetFn world_player_spawn_get;
    DfHostWorldPlayerSpawnSetFn world_player_spawn_set;
    DfHostPlayerStringGetFn player_string_get;
    DfHostPlayerToastFn player_toast;
    DfHostPlayerCooldownFn player_cooldown;
    DfHostPlayerKnockBackFn player_knock_back;
    DfHostPlayerFinalDamageFn player_final_damage;
    DfHostPlayerUsingItemFn player_using_item;
    DfHostPlayerSleepingFn player_sleeping;
    DfHostPlayerDeathPositionFn player_death_position;
    DfHostPlayerBlockActionFn player_block_action;
} DfHostApiV27;
#define DF_COMMAND_PARAMETER_SUBCOMMAND 1u
#define DF_COMMAND_PARAMETER_ENUM 2u
#define DF_COMMAND_PARAMETER_STRING 3u
#define DF_COMMAND_PARAMETER_INTEGER 4u
#define DF_COMMAND_PARAMETER_FLOAT 5u
#define DF_COMMAND_PARAMETER_BOOL 6u
#define DF_COMMAND_PARAMETER_DYNAMIC_ENUM 7u
#define DF_COMMAND_PARAMETER_PLAYER 8u
#define DF_COMMAND_PARAMETER_RAW_TEXT 9u
#define DF_COMMAND_PARAMETER_VECTOR 10u
typedef struct { uint32_t kind; uint8_t optional; DfStringView name; DfStringView suffix; const DfStringView *values; uint64_t value_count; } DfCommandParameter;
typedef struct { const DfCommandParameter *parameters; uint64_t parameter_count; } DfCommandOverload;
typedef struct { DfStringView name; DfStringView description; const DfStringView *aliases; uint64_t alias_count; const DfCommandOverload *overloads; uint64_t overload_count; } DfCommandDescriptor;
#define DF_COMMAND_SOURCE_UNKNOWN 0u
#define DF_COMMAND_SOURCE_PLAYER 1u
#define DF_COMMAND_SOURCE_CONSOLE 2u
typedef struct { DfPlayerId player; DfStringView name; uint64_t latency_milliseconds; DfVec3 position; } DfCommandPlayer;
typedef struct { DfStringView source; uint32_t source_kind; DfPlayerId source_player; DfVec3 source_position; const DfCommandPlayer *online_players; uint64_t online_player_count; } DfCommandEnumContext;
typedef struct { DfInvocationId invocation; uint64_t overload; DfStringView source; const DfStringView *arguments; uint64_t argument_count; uint32_t source_kind; DfPlayerId source_player; DfVec3 source_position; const DfCommandPlayer *online_players; uint64_t online_player_count; } DfCommandInput;
typedef struct { uint8_t failed; DfStringBuffer output; } DfCommandState;

typedef struct {
    uint32_t abi_version;
    uint32_t struct_size;
    uint64_t subscriptions;
} DfAbiHeader;

#define DF_EVENT_PLAYER_MOVE 1u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfVec3 old_position;
    DfVec3 new_position;
    DfRotation rotation;
} DfPlayerMoveInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerMoveState;

#define DF_EVENT_PLAYER_CHAT 2u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfStringView message;
} DfPlayerChatInput;

typedef struct {
    uint8_t cancelled;
    uint8_t has_replacement;
    DfStringBuffer replacement;
} DfPlayerChatState;

#define DF_EVENT_PLAYER_JOIN 3u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfStringView name;
} DfPlayerJoinInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerJoinState;

#define DF_EVENT_PLAYER_QUIT 4u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfStringView name;
} DfPlayerQuitInput;

typedef struct {
    uint8_t _reserved;
} DfPlayerQuitState;

#define DF_EVENT_PLAYER_HURT 5u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    uint8_t immune;
    DfDamageSourceView source;
} DfPlayerHurtInput;

typedef struct {
    uint8_t cancelled;
    double damage;
    int64_t attack_immunity_nanoseconds;
} DfPlayerHurtState;

#define DF_EVENT_PLAYER_HEAL 6u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfHealingSourceView source;
} DfPlayerHealInput;

typedef struct {
    uint8_t cancelled;
    double health;
} DfPlayerHealState;

#define DF_EVENT_PLAYER_BLOCK_BREAK 7u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfBlockPos position;
    DfBlockView block;
    const DfItemStackViewV3 *drops;
    uint64_t drop_count;
} DfPlayerBlockBreakInput;

typedef struct {
    uint8_t cancelled;
    int32_t experience;
    const DfItemStackViewV3 *replacement_drops;
    uint64_t replacement_drop_count;
    void *replacement_context;
    DfItemStackViewsDropFn replacement_drop;
} DfPlayerBlockBreakState;

#define DF_EVENT_PLAYER_BLOCK_PLACE 8u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfBlockPos position;
    DfBlockView block;
} DfPlayerBlockPlaceInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerBlockPlaceState;

#define DF_EVENT_PLAYER_FOOD_LOSS 9u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    int32_t from;
} DfPlayerFoodLossInput;

typedef struct {
    uint8_t cancelled;
    int32_t to;
} DfPlayerFoodLossState;

#define DF_EVENT_PLAYER_DEATH 10u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfDamageSourceView source;
} DfPlayerDeathInput;

typedef struct {
    uint8_t keep_inventory;
} DfPlayerDeathState;

#define DF_EVENT_PLAYER_START_BREAK 11u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfBlockPos position;
} DfPlayerStartBreakInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerStartBreakState;

#define DF_EVENT_PLAYER_FIRE_EXTINGUISH 12u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfBlockPos position;
} DfPlayerFireExtinguishInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerFireExtinguishState;

#define DF_EVENT_PLAYER_TOGGLE_SPRINT 13u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    uint8_t after;
} DfPlayerToggleSprintInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerToggleSprintState;

#define DF_EVENT_PLAYER_TOGGLE_SNEAK 14u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    uint8_t after;
} DfPlayerToggleSneakInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerToggleSneakState;

#define DF_EVENT_PLAYER_JUMP 15u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
} DfPlayerJumpInput;

typedef struct {
    uint8_t _reserved;
} DfPlayerJumpState;

#define DF_EVENT_PLAYER_TELEPORT 16u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfVec3 position;
} DfPlayerTeleportInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerTeleportState;

#define DF_EVENT_PLAYER_EXPERIENCE_GAIN 17u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
} DfPlayerExperienceGainInput;

typedef struct {
    uint8_t cancelled;
    int32_t amount;
} DfPlayerExperienceGainState;

#define DF_EVENT_PLAYER_PUNCH_AIR 18u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
} DfPlayerPunchAirInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerPunchAirState;

#define DF_EVENT_PLAYER_HELD_SLOT_CHANGE 19u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    int32_t from;
    int32_t to;
} DfPlayerHeldSlotChangeInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerHeldSlotChangeState;

#define DF_EVENT_PLAYER_SLEEP 20u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
} DfPlayerSleepInput;

typedef struct {
    uint8_t cancelled;
    uint8_t send_reminder;
} DfPlayerSleepState;

#define DF_EVENT_PLAYER_BLOCK_PICK 21u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfBlockPos position;
    DfBlockView block;
} DfPlayerBlockPickInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerBlockPickState;

#define DF_EVENT_PLAYER_LECTERN_PAGE_TURN 22u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfBlockPos position;
    int32_t old_page;
} DfPlayerLecternPageTurnInput;

typedef struct {
    uint8_t cancelled;
    int32_t new_page;
} DfPlayerLecternPageTurnState;

#define DF_EVENT_PLAYER_SIGN_EDIT 23u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfBlockPos position;
    uint8_t front_side;
    DfStringView old_text;
    DfStringView new_text;
} DfPlayerSignEditInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerSignEditState;

#define DF_EVENT_PLAYER_ITEM_USE 24u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
} DfPlayerItemUseInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerItemUseState;

#define DF_EVENT_PLAYER_ITEM_USE_ON_BLOCK 25u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfBlockPos position;
    int32_t face;
    DfVec3 click_position;
} DfPlayerItemUseOnBlockInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerItemUseOnBlockState;

#define DF_EVENT_PLAYER_ITEM_CONSUME 26u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfItemStackViewV3 item;
} DfPlayerItemConsumeInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerItemConsumeState;

#define DF_EVENT_PLAYER_ITEM_RELEASE 27u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfItemStackViewV3 item;
    int64_t duration_nanoseconds;
} DfPlayerItemReleaseInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerItemReleaseState;

#define DF_EVENT_PLAYER_ITEM_DAMAGE 28u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfItemStackViewV3 item;
} DfPlayerItemDamageInput;

typedef struct {
    uint8_t cancelled;
    int32_t damage;
} DfPlayerItemDamageState;

#define DF_EVENT_PLAYER_ITEM_DROP 29u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfItemStackViewV3 item;
} DfPlayerItemDropInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerItemDropState;

#define DF_EVENT_PLAYER_ATTACK_ENTITY 30u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfEntityId target;
} DfPlayerAttackEntityInput;

typedef struct {
    uint8_t cancelled;
    double knockback_force;
    double knockback_height;
    uint8_t critical;
} DfPlayerAttackEntityState;

#define DF_EVENT_PLAYER_ITEM_USE_ON_ENTITY 31u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfEntityId target;
} DfPlayerItemUseOnEntityInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerItemUseOnEntityState;

#define DF_EVENT_PLAYER_CHANGE_WORLD 32u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfWorldId before;
    DfWorldId after;
} DfPlayerChangeWorldInput;

typedef struct {
    uint8_t _reserved;
} DfPlayerChangeWorldState;

#define DF_EVENT_PLAYER_RESPAWN 33u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
} DfPlayerRespawnInput;

typedef struct {
    DfVec3 position;
    DfWorldId world;
} DfPlayerRespawnState;

#define DF_EVENT_PLAYER_SKIN_CHANGE 34u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    uint64_t snapshot;
} DfPlayerSkinChangeInput;

typedef struct {
    uint8_t cancelled;
} DfPlayerSkinChangeState;

#define DF_EVENT_PLAYER_ITEM_PICKUP 38u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfItemStackViewV3 item;
} DfPlayerItemPickupInput;

typedef struct {
    uint8_t cancelled;
    const DfItemStackViewV3 *replacement;
    void *replacement_context;
    DfItemStackViewsDropFn replacement_drop;
} DfPlayerItemPickupState;

#define DF_EVENT_PLAYER_TRANSFER 39u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
} DfPlayerTransferInput;

/*
 * address initially borrows the host address for the duration of handle_event.
 * A plugin replacing address.ip or address.zone with owned storage sets
 * replacement_drop and transfers that storage to the host. The host copies the
 * final address and invokes replacement_drop exactly once, including error and
 * validation paths.
 */
typedef struct {
    uint8_t cancelled;
    DfUDPAddrView address;
    void *replacement_context;
    DfEventDropFn replacement_drop;
} DfPlayerTransferState;

#define DF_EVENT_PLAYER_COMMAND_EXECUTION 40u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    DfStringView command_name;
    DfStringView command_description;
    DfStringView command_usage;
    const DfStringView *command_aliases;
    uint64_t command_alias_count;
    const DfStringView *arguments;
    uint64_t argument_count;
} DfPlayerCommandExecutionInput;

/*
 * replacement_arguments must contain exactly input.argument_count entries.
 * A non-null replacement_drop transfers the replacement array and its strings
 * to the host. The host copies them and invokes replacement_drop exactly once,
 * including error and validation paths.
 */
typedef struct {
    uint8_t cancelled;
    const DfStringView *replacement_arguments;
    uint64_t replacement_argument_count;
    void *replacement_context;
    DfEventDropFn replacement_drop;
} DfPlayerCommandExecutionState;

#define DF_EVENT_PLAYER_DIAGNOSTICS 41u

typedef struct {
    DfInvocationId invocation;
    DfPlayerSnapshot player;
    double average_frames_per_second;
    double average_server_sim_tick_time;
    double average_client_sim_tick_time;
    double average_begin_frame_time;
    double average_input_time;
    double average_render_time;
    double average_end_frame_time;
    double average_remainder_time_percent;
    double average_unaccounted_time_percent;
} DfPlayerDiagnosticsInput;

typedef struct {
    uint8_t _reserved;
} DfPlayerDiagnosticsState;

#define DF_EVENT_WORLD_LIQUID_FLOW 42u
#define DF_EVENT_WORLD_LIQUID_DECAY 43u
#define DF_EVENT_WORLD_LIQUID_HARDEN 44u
#define DF_EVENT_WORLD_SOUND 45u
#define DF_EVENT_WORLD_FIRE_SPREAD 46u
#define DF_EVENT_WORLD_BLOCK_BURN 47u
#define DF_EVENT_WORLD_CROP_TRAMPLE 48u
#define DF_EVENT_WORLD_LEAVES_DECAY 49u
#define DF_EVENT_WORLD_ENTITY_SPAWN 50u
#define DF_EVENT_WORLD_ENTITY_DESPAWN 51u
#define DF_EVENT_WORLD_EXPLOSION 52u
#define DF_EVENT_WORLD_REDSTONE_UPDATE 53u
#define DF_EVENT_WORLD_CLOSE 54u
#define DF_EVENT_PACKET_CLIENT 55u
#define DF_EVENT_PACKET_SERVER 56u

typedef struct {
    uint64_t packet;
    uint32_t packet_id;
    uint32_t _reserved;
    DfStringView xuid;
} DfPacketInput;

typedef struct {
    uint8_t cancelled;
} DfPacketState;

typedef struct {
    uint8_t cancelled;
} DfWorldCancellableState;

typedef struct {
    uint8_t _reserved;
} DfWorldNotificationState;

typedef struct {
    DfInvocationId invocation;
    DfBlockPos from;
    DfBlockPos into;
    DfBlockView liquid;
    DfBlockView replaced;
} DfWorldLiquidFlowInput;

typedef struct {
    DfInvocationId invocation;
    DfBlockPos position;
    DfBlockView before;
    const DfBlockView *after;
} DfWorldLiquidDecayInput;

typedef struct {
    DfInvocationId invocation;
    DfBlockPos position;
    DfBlockView liquid_hardened;
    DfBlockView other_liquid;
    DfBlockView new_block;
} DfWorldLiquidHardenInput;

typedef struct {
    DfInvocationId invocation;
    DfSoundViewV1 sound;
    DfVec3 position;
} DfWorldSoundInput;

typedef struct {
    DfInvocationId invocation;
    DfBlockPos from;
    DfBlockPos to;
} DfWorldFireSpreadInput;

typedef struct {
    DfInvocationId invocation;
    DfBlockPos position;
} DfWorldPositionInput;

typedef struct {
    DfInvocationId invocation;
    DfEntityId entity;
} DfWorldEntityInput;

typedef struct {
    DfInvocationId invocation;
    DfVec3 position;
    const DfEntityId *entities;
    uint64_t entity_count;
    const DfBlockPos *blocks;
    uint64_t block_count;
} DfWorldExplosionInput;

/*
 * A non-null replacement_drop transfers both replacement arrays to the host.
 * The host copies the arrays and invokes replacement_drop exactly once,
 * including validation and error paths.
 */
typedef struct {
    uint8_t cancelled;
    uint8_t spawn_fire;
    double item_drop_chance;
    const DfEntityId *replacement_entities;
    uint64_t replacement_entity_count;
    const DfBlockPos *replacement_blocks;
    uint64_t replacement_block_count;
    void *replacement_context;
    DfEventDropFn replacement_drop;
} DfWorldExplosionState;

typedef struct {
    DfInvocationId invocation;
    DfBlockPos position;
    DfBlockPos changed_neighbour;
    uint8_t has_changed_neighbour;
    uint8_t changed_redstone_relevant;
    DfBlockPos source;
    uint8_t has_source;
    DfBlockView before;
    const DfBlockView *after;
    int32_t old_power;
    int32_t new_power;
    int64_t current_tick;
    uint32_t cause;
} DfWorldRedstoneUpdateInput;

typedef struct {
    DfInvocationId invocation;
} DfWorldCloseInput;

#define DF_EVENT_ENTITY_HURT 35u

typedef struct {
    DfInvocationId invocation;
    DfEntityId entity;
    DfDamageSourceView source;
    double health;
    double max_health;
} DfEntityHurtInput;

typedef struct {
    double damage;
    uint8_t cancelled;
} DfEntityHurtState;

#define DF_EVENT_ENTITY_HEAL 36u

typedef struct {
    DfInvocationId invocation;
    DfEntityId entity;
    DfHealingSourceView source;
    double health;
    double max_health;
} DfEntityHealInput;

typedef struct {
    double health;
    uint8_t cancelled;
} DfEntityHealState;

#define DF_EVENT_ENTITY_DEATH 37u

typedef struct {
    DfInvocationId invocation;
    DfEntityId entity;
    DfDamageSourceView source;
    double health;
    double damage;
} DfEntityDeathInput;

typedef struct {
    uint8_t cancelled;
} DfEntityDeathState;

typedef DfStatus (*DfHandleEventFn)(void *instance, DfEventId event_id, const void *input, void *state);
typedef void *(*DfPluginCreateFn)(void);
typedef DfStatus (*DfPluginEnableFn)(void *instance, DfStringBuffer *error);
typedef DfStatus (*DfPluginDisableFn)(void *instance);
typedef const DfCommandDescriptor *(*DfPluginCommandsFn)(void *instance, uint64_t *count);
typedef uint64_t (*DfPluginEntityTypeCountFn)(void *instance);
typedef DfStatus (*DfPluginEntityTypeAtFn)(void *instance, uint64_t index, DfEntityTypeDescriptorV2 *out);
typedef DfStatus (*DfPluginHandleEntityFn)(void *instance, uint64_t local_type, uint32_t operation, uint64_t entity_instance, const void *input, void *state);
typedef DfStatus (*DfHandleCommandFn)(void *instance, uint64_t command, const DfCommandInput *input, DfCommandState *state);
typedef DfStatus (*DfCommandEnumOptionsFn)(void *instance, uint64_t command, uint64_t overload, uint64_t parameter, const DfCommandEnumContext *context, DfStringBuffer *output);
typedef DfStatus (*DfPluginSetHostFn)(void *instance, const DfHostApiV27 *host);
typedef void (*DfPluginDestroyFn)(void *instance);
typedef DfStatus (*DfPluginScheduledFn)(void *instance, uint64_t callback, DfInvocationId invocation, uint32_t phase, uint32_t result);
typedef DfStatus (*DfPluginAllowFn)(void *instance, const DfAllowInput *input, DfStringBuffer *message, uint8_t *allowed);

typedef struct {
    DfAbiHeader header;
    DfStringView plugin_id;
    DfPluginCreateFn create;
    DfPluginEnableFn enable;
    DfPluginDisableFn disable;
    DfPluginCommandsFn commands;
    DfPluginEntityTypeCountFn entity_type_count;
    DfPluginEntityTypeAtFn entity_type_at;
    DfPluginHandleEntityFn handle_entity;
    DfHandleCommandFn handle_command;
    DfCommandEnumOptionsFn command_enum_options;
    DfPluginSetHostFn set_host;
    DfPluginDestroyFn destroy;
    DfHandleEventFn handle_event;
    DfPluginScheduledFn handle_scheduled;
    DfPluginAllowFn allow;
} DfPluginApiV11;

typedef const DfPluginApiV11 *(*DfPluginEntryV11Fn)(void);

typedef struct DfRuntime DfRuntime;
typedef struct { DfStringView plugin_directory; const DfHostApiV27 *host; } DfRuntimeConfig;

DfStatus df_runtime_create(const DfRuntimeConfig *config, DfRuntime **out, uint8_t *error, uint64_t error_capacity);
DfStatus df_runtime_enable(DfRuntime *runtime, uint8_t *error, uint64_t error_capacity);
void df_runtime_begin_disable(DfRuntime *runtime);
void df_runtime_finish_disable(DfRuntime *runtime);
void df_runtime_disable(DfRuntime *runtime);
void df_runtime_destroy(DfRuntime *runtime);
uint64_t df_runtime_plugin_count(const DfRuntime *runtime);
uint64_t df_runtime_subscriptions(const DfRuntime *runtime);
uint64_t df_runtime_entity_type_count(const DfRuntime *runtime);
DfStatus df_runtime_entity_type_at(const DfRuntime *runtime, uint64_t index, DfEntityTypeDescriptorV2 *out);
DfStatus df_runtime_entity_adopt(DfRuntime *runtime, uint64_t type_key, uint64_t opaque, DfEntityInstanceId *out);
DfStatus df_runtime_entity_load(DfRuntime *runtime, uint64_t type_key, const DfEntityLoadInput *input, DfEntityInstanceId *out);
DfStatus df_runtime_entity_save(DfRuntime *runtime, DfEntityInstanceId instance, DfEntitySaveState *state);
DfStatus df_runtime_entity_tick(DfRuntime *runtime, DfEntityInstanceId instance, const DfEntityTickInput *input, DfEntityTickState *state);
DfStatus df_runtime_entity_hurt(DfRuntime *runtime, DfEntityInstanceId instance, const DfEntityHurtInput *input, DfEntityHurtState *state);
DfStatus df_runtime_entity_heal(DfRuntime *runtime, DfEntityInstanceId instance, const DfEntityHealInput *input, DfEntityHealState *state);
DfStatus df_runtime_entity_death(DfRuntime *runtime, DfEntityInstanceId instance, const DfEntityDeathInput *input, DfEntityDeathState *state);
DfStatus df_runtime_entity_destroy(DfRuntime *runtime, DfEntityInstanceId instance);
DfStatus df_runtime_entity_decode_nbt(DfRuntime *runtime, uint64_t type_key, const DfEntityExactInput *input, DfEntityExactState *state);
DfStatus df_runtime_entity_call(DfRuntime *runtime, DfEntityInstanceId instance, uint32_t operation, const DfEntityExactInput *input, DfEntityExactState *state);
uint64_t df_runtime_command_count(const DfRuntime *runtime);
DfStatus df_runtime_command_at(const DfRuntime *runtime, uint64_t index, DfCommandDescriptor *out);
DfStatus df_runtime_handle_command(DfRuntime *runtime, uint64_t index, const DfCommandInput *input, DfCommandState *state);
DfStatus df_runtime_command_enum_options(DfRuntime *runtime, uint64_t index, uint64_t overload, uint64_t parameter, const DfCommandEnumContext *context, DfStringBuffer *output);
DfStatus df_runtime_handle_event(DfRuntime *runtime, DfEventId event_id, const void *input, void *state);
DfStatus df_runtime_handle_scheduled(DfRuntime *runtime, uint64_t plugin, uint64_t callback, DfInvocationId invocation, uint32_t phase, uint32_t result);
DfStatus df_runtime_allow(DfRuntime *runtime, const DfAllowInput *input, DfStringBuffer *message, uint8_t *allowed);

#ifdef __cplusplus
}
#endif

#endif
