// Command abi-gen generates language-neutral native ABI types from schema.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const generatorVersion = "1"

type schemaFile struct {
	Domain string  `yaml:"domain"`
	Events []event `yaml:"events"`
}

type event struct {
	Domain string
	Name   string  `yaml:"name"`
	ID     uint32  `yaml:"id"`
	Input  []field `yaml:"input"`
	State  []field `yaml:"state"`
}

type field struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type playerSchema struct {
	States  []playerState `yaml:"states"`
	Effects []effectType  `yaml:"effects"`
	Texts   []playerText  `yaml:"texts"`
	Sounds  []soundType   `yaml:"sounds"`
}

type soundType struct {
	Name string `yaml:"name"`
	ID   uint32 `yaml:"id"`
	Go   string `yaml:"go"`
}

type playerText struct {
	Name string `yaml:"name"`
	ID   uint32 `yaml:"id"`
	Set  string `yaml:"set"`
	Rust string `yaml:"rust"`
}

type effectType struct {
	Name string `yaml:"name"`
	ID   uint32 `yaml:"id"`
	Kind string `yaml:"kind"`
}

type playerState struct {
	Name     string `yaml:"name"`
	ID       uint32 `yaml:"id"`
	Type     string `yaml:"type"`
	Set      string `yaml:"set"`
	Unset    string `yaml:"unset"`
	Get      string `yaml:"get"`
	RustSet  string `yaml:"rust_set"`
	RustGet  string `yaml:"rust_get"`
	Adapter  string `yaml:"adapter"`
	Validate string `yaml:"validate"`
}

func main() {
	root := flag.String("root", ".", "repository root")
	check := flag.Bool("check", false, "fail if generated files are stale")
	flag.Parse()

	events, err := readEvents(filepath.Join(*root, "schema", "events"))
	if err != nil {
		fatal(err)
	}
	player, err := readPlayer(filepath.Join(*root, "schema", "player.yaml"))
	if err != nil {
		fatal(err)
	}
	outputs := map[string][]byte{
		filepath.Join(*root, "abi", "include", "dragonfly_plugin.h"):                         generateC(events, player),
		filepath.Join(*root, "rust", "dragonfly-plugin-sys", "src", "generated.rs"):          generateRust(events, player),
		filepath.Join(*root, "internal", "host", "player_state_generated.go"):                generateGoPlayerStates(player),
		filepath.Join(*root, "internal", "native", "player_state_generated.go"):              generateGoNativePlayerStates(player),
		filepath.Join(*root, "rust", "dragonfly-plugin", "src", "player_state_generated.rs"): generateRustPlayerStates(player),
	}
	for path, data := range outputs {
		if strings.HasSuffix(path, ".rs") {
			// Rustfmt runs later through Cargo. Keep generated source deterministic here.
			data = bytes.TrimSpace(data)
			data = append(data, '\n')
		}
		if err := writeGenerated(path, data, *check); err != nil {
			fatal(err)
		}
	}
}

func readPlayer(path string) (playerSchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return playerSchema{}, err
	}
	var schema playerSchema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return playerSchema{}, fmt.Errorf("decode %s: %w", path, err)
	}
	ids, names := map[uint32]bool{}, map[string]bool{}
	validTypes := map[string]bool{"f64": true, "i32": true, "bool": true, "game_mode": true, "sound": true}
	validAdapters := map[string]bool{"": true, "game_mode": true, "healing_source": true, "damage_source": true, "toggle": true, "sound": true}
	validValidation := map[string]bool{"": true, "non_negative": true, "positive": true, "unit_interval": true}
	for _, state := range schema.States {
		if state.Name == "" || names[state.Name] || ids[state.ID] || !validTypes[state.Type] || !validAdapters[state.Adapter] || !validValidation[state.Validate] {
			return playerSchema{}, fmt.Errorf("invalid or duplicate player state %+v", state)
		}
		if state.Set == "" && state.Get == "" || state.Set != "" && state.RustSet == "" || state.Get != "" && state.RustGet == "" {
			return playerSchema{}, fmt.Errorf("incomplete player state %+v", state)
		}
		if state.Adapter == "toggle" && (state.Type != "bool" || state.Unset == "") || state.Unset != "" && state.Adapter != "toggle" {
			return playerSchema{}, fmt.Errorf("invalid toggle player state %+v", state)
		}
		ids[state.ID], names[state.Name] = true, true
	}
	sort.Slice(schema.States, func(i, j int) bool { return schema.States[i].ID < schema.States[j].ID })
	effectIDs, effectNames := map[uint32]bool{}, map[string]bool{}
	for _, effect := range schema.Effects {
		if effect.Name == "" || effect.ID == 0 || effectIDs[effect.ID] || effectNames[effect.Name] || effect.Kind != "lasting" && effect.Kind != "instant" {
			return playerSchema{}, fmt.Errorf("invalid or duplicate effect %+v", effect)
		}
		effectIDs[effect.ID], effectNames[effect.Name] = true, true
	}
	sort.Slice(schema.Effects, func(i, j int) bool { return schema.Effects[i].ID < schema.Effects[j].ID })
	textIDs, textNames := map[uint32]bool{}, map[string]bool{}
	for _, text := range schema.Texts {
		if text.Name == "" || text.Set == "" || text.Rust == "" || textIDs[text.ID] || textNames[text.Name] {
			return playerSchema{}, fmt.Errorf("invalid or duplicate player text %+v", text)
		}
		textIDs[text.ID], textNames[text.Name] = true, true
	}
	sort.Slice(schema.Texts, func(i, j int) bool { return schema.Texts[i].ID < schema.Texts[j].ID })
	soundIDs, soundNames := map[uint32]bool{}, map[string]bool{}
	for _, sound := range schema.Sounds {
		if sound.Name == "" || sound.Go == "" || soundIDs[sound.ID] || soundNames[sound.Name] {
			return playerSchema{}, fmt.Errorf("invalid or duplicate sound %+v", sound)
		}
		soundIDs[sound.ID], soundNames[sound.Name] = true, true
	}
	sort.Slice(schema.Sounds, func(i, j int) bool { return schema.Sounds[i].ID < schema.Sounds[j].ID })
	return schema, nil
}

func readEvents(dir string) ([]event, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var events []event
	ids := map[uint32]string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var file schemaFile
		if err := yaml.Unmarshal(data, &file); err != nil {
			return nil, fmt.Errorf("decode %s: %w", path, err)
		}
		if file.Domain == "" {
			return nil, fmt.Errorf("%s: empty domain", path)
		}
		for _, evt := range file.Events {
			evt.Domain = file.Domain
			fullName := evt.Domain + "_" + evt.Name
			if previous, ok := ids[evt.ID]; ok {
				return nil, fmt.Errorf("event ID %d used by %s and %s", evt.ID, previous, fullName)
			}
			ids[evt.ID] = fullName
			if err := validateFields(fullName, append(append([]field{}, evt.Input...), evt.State...)); err != nil {
				return nil, err
			}
			events = append(events, evt)
		}
	}
	sort.Slice(events, func(i, j int) bool { return events[i].ID < events[j].ID })
	return events, nil
}

func validateFields(eventName string, fields []field) error {
	valid := map[string]bool{
		"bool": true, "player_id": true, "rotation": true, "string_buffer": true,
		"string_view": true, "vec3": true, "f64": true, "u64": true, "i32": true,
		"block_pos": true, "item_stack": true,
	}
	seen := map[string]bool{}
	for _, f := range fields {
		if f.Name == "" || !valid[f.Type] {
			return fmt.Errorf("event %s: invalid field %q type %q", eventName, f.Name, f.Type)
		}
		if seen[f.Name] {
			return fmt.Errorf("event %s: duplicate field %q", eventName, f.Name)
		}
		seen[f.Name] = true
	}
	return nil
}

func generateC(events []event, player playerSchema) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "/* Code generated by abi-gen v%s. DO NOT EDIT. */\n", generatorVersion)
	b.WriteString(`#ifndef BEDROCK_GOPHERS_DRAGONFLY_PLUGIN_H
#define BEDROCK_GOPHERS_DRAGONFLY_PLUGIN_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#define DF_ABI_VERSION 1u
#define DF_STATUS_OK 0
#define DF_STATUS_ERROR 1

typedef int32_t DfStatus;
typedef uint32_t DfEventId;

typedef struct { uint8_t bytes[16]; uint64_t generation; } DfPlayerId;
typedef struct { uint8_t bytes[16]; uint64_t generation; } DfEntityId;
typedef struct { double x; double y; double z; } DfVec3;
typedef struct { double yaw; double pitch; } DfRotation;
typedef struct { int32_t x; int32_t y; int32_t z; } DfBlockPos;
typedef struct { const uint8_t *data; uint64_t len; } DfStringView;
typedef struct { uint8_t *data; uint64_t len; uint64_t capacity; } DfStringBuffer;
typedef struct { DfStringView identifier; int32_t metadata; int32_t count; int32_t damage; } DfItemStackView;
#define DF_PLAYER_TRANSFORM_TELEPORT 0u
#define DF_PLAYER_TRANSFORM_MOVE 1u
#define DF_PLAYER_TRANSFORM_VELOCITY 2u
`)
	for _, text := range player.Texts {
		fmt.Fprintf(&b, "#define DF_PLAYER_TEXT_%s %du\n", strings.ToUpper(text.Name), text.ID)
	}
	for _, state := range player.States {
		fmt.Fprintf(&b, "#define DF_PLAYER_STATE_%s %du\n", strings.ToUpper(state.Name), state.ID)
	}
	for _, effect := range player.Effects {
		fmt.Fprintf(&b, "#define DF_EFFECT_%s %du\n", strings.ToUpper(effect.Name), effect.ID)
	}
	for _, sound := range player.Sounds {
		fmt.Fprintf(&b, "#define DF_SOUND_%s %du\n", strings.ToUpper(sound.Name), sound.ID)
	}
	b.WriteString(`
typedef struct { DfStringView text; DfStringView subtitle; DfStringView action_text; uint64_t fade_in_milliseconds; uint64_t duration_milliseconds; uint64_t fade_out_milliseconds; } DfTitleView;
typedef struct { double number; int64_t integer; } DfPlayerStateValue;
#define DF_PLAYER_EFFECT_ADD 0u
#define DF_PLAYER_EFFECT_REMOVE 1u
typedef struct { uint32_t effect_type; int32_t level; uint64_t duration_milliseconds; uint8_t ambient; uint8_t infinite; uint8_t particles_hidden; } DfEffectView;
typedef DfStatus (*DfHostPlayerTextFn)(uint64_t context, DfPlayerId player, uint32_t kind, DfStringView message);
typedef DfStatus (*DfHostPlayerTitleFn)(uint64_t context, DfPlayerId player, DfTitleView title);
typedef DfStatus (*DfHostPlayerTransformFn)(uint64_t context, DfPlayerId player, uint32_t kind, DfVec3 vector, double yaw, double pitch);
typedef DfStatus (*DfHostPlayerRotationFn)(uint64_t context, DfPlayerId player, DfRotation *rotation);
typedef DfStatus (*DfHostPlayerStateSetFn)(uint64_t context, DfPlayerId player, uint32_t kind, DfPlayerStateValue value);
typedef DfStatus (*DfHostPlayerStateGetFn)(uint64_t context, DfPlayerId player, uint32_t kind, DfPlayerStateValue *value);
typedef DfStatus (*DfHostPlayerEffectFn)(uint64_t context, DfPlayerId player, uint32_t operation, DfEffectView effect);
typedef DfStatus (*DfHostPlayerEntityVisibilityFn)(uint64_t context, DfPlayerId player, DfEntityId entity, uint8_t visible);
typedef struct {
    uint32_t abi_version;
    uint32_t struct_size;
    uint64_t context;
    DfHostPlayerTextFn player_text;
    DfHostPlayerTitleFn player_title;
    DfHostPlayerTransformFn player_transform;
    DfHostPlayerRotationFn player_rotation;
    DfHostPlayerStateSetFn player_state_set;
    DfHostPlayerStateGetFn player_state_get;
    DfHostPlayerEffectFn player_effect;
    DfHostPlayerEntityVisibilityFn player_entity_visibility;
} DfHostApiV1;
#define DF_COMMAND_PARAMETER_SUBCOMMAND 1u
#define DF_COMMAND_PARAMETER_ENUM 2u
#define DF_COMMAND_PARAMETER_STRING 3u
#define DF_COMMAND_PARAMETER_INTEGER 4u
#define DF_COMMAND_PARAMETER_FLOAT 5u
#define DF_COMMAND_PARAMETER_BOOL 6u
#define DF_COMMAND_PARAMETER_DYNAMIC_ENUM 7u
#define DF_COMMAND_PARAMETER_PLAYER 8u
#define DF_COMMAND_PARAMETER_RAW_TEXT 9u
typedef struct { uint32_t kind; uint8_t optional; DfStringView name; const DfStringView *values; uint64_t value_count; } DfCommandParameter;
typedef struct { const DfCommandParameter *parameters; uint64_t parameter_count; } DfCommandOverload;
typedef struct { DfStringView name; DfStringView description; const DfCommandOverload *overloads; uint64_t overload_count; } DfCommandDescriptor;
typedef struct { DfStringView source; const DfStringView *online_players; uint64_t online_player_count; } DfCommandEnumContext;
typedef struct { DfPlayerId player; DfStringView name; uint64_t latency_milliseconds; } DfCommandPlayer;
#define DF_COMMAND_SOURCE_UNKNOWN 0u
#define DF_COMMAND_SOURCE_PLAYER 1u
#define DF_COMMAND_SOURCE_CONSOLE 2u
typedef struct { DfStringView source; DfStringView arguments; uint32_t source_kind; DfPlayerId source_player; const DfCommandPlayer *online_players; uint64_t online_player_count; } DfCommandInput;
typedef struct { uint8_t failed; DfStringBuffer output; } DfCommandState;

typedef struct {
    uint32_t abi_version;
    uint32_t struct_size;
    uint64_t subscriptions;
} DfAbiHeader;

`)
	for _, evt := range events {
		upper := strings.ToUpper(evt.Domain + "_" + evt.Name)
		name := cName(evt)
		fmt.Fprintf(&b, "#define DF_EVENT_%s %du\n\n", upper, evt.ID)
		fmt.Fprintf(&b, "typedef struct {\n")
		if len(evt.Input) == 0 {
			fmt.Fprintf(&b, "    uint8_t _reserved;\n")
		}
		for _, f := range evt.Input {
			fmt.Fprintf(&b, "    %s %s;\n", cType(f.Type), f.Name)
		}
		fmt.Fprintf(&b, "} %sInput;\n\n", name)
		fmt.Fprintf(&b, "typedef struct {\n")
		if len(evt.State) == 0 {
			fmt.Fprintf(&b, "    uint8_t _reserved;\n")
		}
		for _, f := range evt.State {
			fmt.Fprintf(&b, "    %s %s;\n", cType(f.Type), f.Name)
		}
		fmt.Fprintf(&b, "} %sState;\n\n", name)
	}
	b.WriteString(`typedef DfStatus (*DfHandleEventFn)(void *instance, DfEventId event_id, const void *input, void *state);
typedef void *(*DfPluginCreateFn)(void);
typedef DfStatus (*DfPluginLifecycleFn)(void *instance);
typedef const DfCommandDescriptor *(*DfPluginCommandsFn)(void *instance, uint64_t *count);
typedef DfStatus (*DfHandleCommandFn)(void *instance, uint64_t command, const DfCommandInput *input, DfCommandState *state);
typedef DfStatus (*DfCommandEnumOptionsFn)(void *instance, uint64_t command, uint64_t overload, uint64_t parameter, const DfCommandEnumContext *context, DfStringBuffer *output);
typedef DfStatus (*DfPluginSetHostFn)(void *instance, const DfHostApiV1 *host);
typedef void (*DfPluginDestroyFn)(void *instance);

typedef struct {
    DfAbiHeader header;
    DfStringView plugin_id;
    DfPluginCreateFn create;
    DfPluginLifecycleFn enable;
    DfPluginLifecycleFn disable;
    DfPluginCommandsFn commands;
    DfHandleCommandFn handle_command;
    DfCommandEnumOptionsFn command_enum_options;
    DfPluginSetHostFn set_host;
    DfPluginDestroyFn destroy;
    DfHandleEventFn handle_event;
} DfPluginApiV1;

typedef const DfPluginApiV1 *(*DfPluginEntryV1Fn)(void);

typedef struct DfRuntime DfRuntime;
typedef struct { DfStringView plugin_directory; const DfHostApiV1 *host; } DfRuntimeConfig;

DfStatus df_runtime_create(const DfRuntimeConfig *config, DfRuntime **out, uint8_t *error, uint64_t error_capacity);
DfStatus df_runtime_enable(DfRuntime *runtime);
void df_runtime_disable(DfRuntime *runtime);
void df_runtime_destroy(DfRuntime *runtime);
uint64_t df_runtime_plugin_count(const DfRuntime *runtime);
uint64_t df_runtime_subscriptions(const DfRuntime *runtime);
uint64_t df_runtime_command_count(const DfRuntime *runtime);
DfStatus df_runtime_command_at(const DfRuntime *runtime, uint64_t index, DfCommandDescriptor *out);
DfStatus df_runtime_handle_command(DfRuntime *runtime, uint64_t index, const DfCommandInput *input, DfCommandState *state);
DfStatus df_runtime_command_enum_options(DfRuntime *runtime, uint64_t index, uint64_t overload, uint64_t parameter, const DfCommandEnumContext *context, DfStringBuffer *output);
DfStatus df_runtime_handle_event(DfRuntime *runtime, DfEventId event_id, const void *input, void *state);
`)
	b.WriteString(`
#ifdef __cplusplus
}
#endif

#endif
`)
	return b.Bytes()
}

func generateRust(events []event, player playerSchema) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated by abi-gen v%s. DO NOT EDIT.\n\n", generatorVersion)
	b.WriteString(`use core::ffi::c_void;

pub const DF_ABI_VERSION: u32 = 1;
pub const DF_STATUS_OK: DfStatus = 0;
pub const DF_STATUS_ERROR: DfStatus = 1;
pub type DfStatus = i32;
pub type DfEventId = u32;

#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfPlayerId { pub bytes: [u8; 16], pub generation: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfEntityId { pub bytes: [u8; 16], pub generation: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfVec3 { pub x: f64, pub y: f64, pub z: f64 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfRotation { pub yaw: f64, pub pitch: f64 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfBlockPos { pub x: i32, pub y: i32, pub z: i32 }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfStringView { pub data: *const u8, pub len: u64 }
impl Default for DfStringView {
    fn default() -> Self { Self { data: core::ptr::null(), len: 0 } }
}
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfStringBuffer { pub data: *mut u8, pub len: u64, pub capacity: u64 }
impl Default for DfStringBuffer {
    fn default() -> Self { Self { data: core::ptr::null_mut(), len: 0, capacity: 0 } }
}
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfItemStackView { pub identifier: DfStringView, pub metadata: i32, pub count: i32, pub damage: i32 }
pub const DF_PLAYER_TRANSFORM_TELEPORT: u32 = 0;
pub const DF_PLAYER_TRANSFORM_MOVE: u32 = 1;
pub const DF_PLAYER_TRANSFORM_VELOCITY: u32 = 2;
`)
	for _, text := range player.Texts {
		fmt.Fprintf(&b, "pub const DF_PLAYER_TEXT_%s: u32 = %d;\n", strings.ToUpper(text.Name), text.ID)
	}
	for _, state := range player.States {
		fmt.Fprintf(&b, "pub const DF_PLAYER_STATE_%s: u32 = %d;\n", strings.ToUpper(state.Name), state.ID)
	}
	for _, effect := range player.Effects {
		fmt.Fprintf(&b, "pub const DF_EFFECT_%s: u32 = %d;\n", strings.ToUpper(effect.Name), effect.ID)
	}
	for _, sound := range player.Sounds {
		fmt.Fprintf(&b, "pub const DF_SOUND_%s: u32 = %d;\n", strings.ToUpper(sound.Name), sound.ID)
	}
	b.WriteString(`
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfTitleView { pub text: DfStringView, pub subtitle: DfStringView, pub action_text: DfStringView, pub fade_in_milliseconds: u64, pub duration_milliseconds: u64, pub fade_out_milliseconds: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfPlayerStateValue { pub number: f64, pub integer: i64 }
pub const DF_PLAYER_EFFECT_ADD: u32 = 0;
pub const DF_PLAYER_EFFECT_REMOVE: u32 = 1;
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfEffectView { pub effect_type: u32, pub level: i32, pub duration_milliseconds: u64, pub ambient: u8, pub infinite: u8, pub particles_hidden: u8 }
pub type DfHostPlayerTextFn = unsafe extern "C" fn(context: u64, player: DfPlayerId, kind: u32, message: DfStringView) -> DfStatus;
pub type DfHostPlayerTitleFn = unsafe extern "C" fn(context: u64, player: DfPlayerId, title: DfTitleView) -> DfStatus;
pub type DfHostPlayerTransformFn = unsafe extern "C" fn(context: u64, player: DfPlayerId, kind: u32, vector: DfVec3, yaw: f64, pitch: f64) -> DfStatus;
pub type DfHostPlayerRotationFn = unsafe extern "C" fn(context: u64, player: DfPlayerId, rotation: *mut DfRotation) -> DfStatus;
pub type DfHostPlayerStateSetFn = unsafe extern "C" fn(context: u64, player: DfPlayerId, kind: u32, value: DfPlayerStateValue) -> DfStatus;
pub type DfHostPlayerStateGetFn = unsafe extern "C" fn(context: u64, player: DfPlayerId, kind: u32, value: *mut DfPlayerStateValue) -> DfStatus;
pub type DfHostPlayerEffectFn = unsafe extern "C" fn(context: u64, player: DfPlayerId, operation: u32, effect: DfEffectView) -> DfStatus;
pub type DfHostPlayerEntityVisibilityFn = unsafe extern "C" fn(context: u64, player: DfPlayerId, entity: DfEntityId, visible: u8) -> DfStatus;
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfHostApiV1 { pub abi_version: u32, pub struct_size: u32, pub context: u64, pub player_text: Option<DfHostPlayerTextFn>, pub player_title: Option<DfHostPlayerTitleFn>, pub player_transform: Option<DfHostPlayerTransformFn>, pub player_rotation: Option<DfHostPlayerRotationFn>, pub player_state_set: Option<DfHostPlayerStateSetFn>, pub player_state_get: Option<DfHostPlayerStateGetFn>, pub player_effect: Option<DfHostPlayerEffectFn>, pub player_entity_visibility: Option<DfHostPlayerEntityVisibilityFn> }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfCommandParameter { pub kind: u32, pub optional: u8, pub name: DfStringView, pub values: *const DfStringView, pub value_count: u64 }
pub const DF_COMMAND_PARAMETER_SUBCOMMAND: u32 = 1;
pub const DF_COMMAND_PARAMETER_ENUM: u32 = 2;
pub const DF_COMMAND_PARAMETER_STRING: u32 = 3;
pub const DF_COMMAND_PARAMETER_INTEGER: u32 = 4;
pub const DF_COMMAND_PARAMETER_FLOAT: u32 = 5;
pub const DF_COMMAND_PARAMETER_BOOL: u32 = 6;
pub const DF_COMMAND_PARAMETER_DYNAMIC_ENUM: u32 = 7;
pub const DF_COMMAND_PARAMETER_PLAYER: u32 = 8;
pub const DF_COMMAND_PARAMETER_RAW_TEXT: u32 = 9;
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfCommandOverload { pub parameters: *const DfCommandParameter, pub parameter_count: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfCommandDescriptor { pub name: DfStringView, pub description: DfStringView, pub overloads: *const DfCommandOverload, pub overload_count: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfCommandEnumContext { pub source: DfStringView, pub online_players: *const DfStringView, pub online_player_count: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfCommandPlayer { pub player: DfPlayerId, pub name: DfStringView, pub latency_milliseconds: u64 }
pub const DF_COMMAND_SOURCE_UNKNOWN: u32 = 0;
pub const DF_COMMAND_SOURCE_PLAYER: u32 = 1;
pub const DF_COMMAND_SOURCE_CONSOLE: u32 = 2;
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfCommandInput { pub source: DfStringView, pub arguments: DfStringView, pub source_kind: u32, pub source_player: DfPlayerId, pub online_players: *const DfCommandPlayer, pub online_player_count: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfCommandState { pub failed: u8, pub output: DfStringBuffer }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfAbiHeader { pub abi_version: u32, pub struct_size: u32, pub subscriptions: u64 }

`)
	for _, evt := range events {
		upper := strings.ToUpper(evt.Domain + "_" + evt.Name)
		name := rustName(evt)
		fmt.Fprintf(&b, "pub const DF_EVENT_%s: DfEventId = %d;\n", upper, evt.ID)
		fmt.Fprintf(&b, "pub const DF_SUBSCRIPTION_%s: u64 = 1u64 << %d;\n", upper, evt.ID-1)
		b.WriteString("#[repr(C)]\n#[derive(Clone, Copy, Debug, Default)]\n")
		fmt.Fprintf(&b, "pub struct %sInput {\n", name)
		if len(evt.Input) == 0 {
			b.WriteString("    pub _reserved: u8,\n")
		}
		for _, f := range evt.Input {
			fmt.Fprintf(&b, "    pub %s: %s,\n", f.Name, rustType(f.Type))
		}
		b.WriteString("}\n")
		b.WriteString("#[repr(C)]\n#[derive(Clone, Copy, Debug, Default)]\n")
		fmt.Fprintf(&b, "pub struct %sState {\n", name)
		if len(evt.State) == 0 {
			b.WriteString("    pub _reserved: u8,\n")
		}
		for _, f := range evt.State {
			fmt.Fprintf(&b, "    pub %s: %s,\n", f.Name, rustType(f.Type))
		}
		b.WriteString("}\n\n")
	}
	b.WriteString(`pub type DfPluginCreateFn = unsafe extern "C" fn() -> *mut c_void;
pub type DfPluginLifecycleFn = unsafe extern "C" fn(instance: *mut c_void) -> DfStatus;
pub type DfPluginCommandsFn = unsafe extern "C" fn(instance: *mut c_void, count: *mut u64) -> *const DfCommandDescriptor;
pub type DfHandleCommandFn = unsafe extern "C" fn(instance: *mut c_void, command: u64, input: *const DfCommandInput, state: *mut DfCommandState) -> DfStatus;
pub type DfCommandEnumOptionsFn = unsafe extern "C" fn(instance: *mut c_void, command: u64, overload: u64, parameter: u64, context: *const DfCommandEnumContext, output: *mut DfStringBuffer) -> DfStatus;
pub type DfPluginSetHostFn = unsafe extern "C" fn(instance: *mut c_void, host: *const DfHostApiV1) -> DfStatus;
pub type DfPluginDestroyFn = unsafe extern "C" fn(instance: *mut c_void);
pub type DfHandleEventFn = unsafe extern "C" fn(instance: *mut c_void, event_id: DfEventId, input: *const c_void, state: *mut c_void) -> DfStatus;

#[repr(C)]
pub struct DfPluginApiV1 {
    pub header: DfAbiHeader,
    pub plugin_id: DfStringView,
    pub create: Option<DfPluginCreateFn>,
    pub enable: Option<DfPluginLifecycleFn>,
    pub disable: Option<DfPluginLifecycleFn>,
    pub commands: Option<DfPluginCommandsFn>,
    pub handle_command: Option<DfHandleCommandFn>,
    pub command_enum_options: Option<DfCommandEnumOptionsFn>,
    pub set_host: Option<DfPluginSetHostFn>,
    pub destroy: Option<DfPluginDestroyFn>,
    pub handle_event: Option<DfHandleEventFn>,
}

pub type DfPluginEntryV1Fn = unsafe extern "C" fn() -> *const DfPluginApiV1;

unsafe impl Sync for DfPluginApiV1 {}
`)
	return b.Bytes()
}

func generateGoPlayerStates(playerSchema playerSchema) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated by abi-gen v%s. DO NOT EDIT.\n\n", generatorVersion)
	b.WriteString("package host\n\nimport (\n\t\"math\"\n\n\t\"github.com/bedrock-gophers/plugins/internal/native\"\n\t\"github.com/df-mc/dragonfly/server/player\"\n\t\"github.com/df-mc/dragonfly/server/world\"\n\tdfsound \"github.com/df-mc/dragonfly/server/world/sound\"\n)\n\n")
	b.WriteString("func sendPlayerText(connected *player.Player, kind native.PlayerTextKind, message string) bool {\n\tswitch kind {\n")
	for _, text := range playerSchema.Texts {
		fmt.Fprintf(&b, "\tcase native.PlayerText%s:\n\t\tconnected.%s(message)\n", title(text.Name), text.Set)
	}
	b.WriteString("\tdefault:\n\t\treturn false\n\t}\n\treturn true\n}\n\n")
	b.WriteString("func playerSound(kind native.SoundType) (world.Sound, bool) {\n\tswitch kind {\n")
	for _, sound := range playerSchema.Sounds {
		fmt.Fprintf(&b, "\tcase native.Sound%s:\n\t\treturn dfsound.%s{}, true\n", title(sound.Name), sound.Go)
	}
	b.WriteString("\tdefault:\n\t\treturn nil, false\n\t}\n}\n\n")
	b.WriteString("func setPlayerState(connected *player.Player, kind native.PlayerStateKind, value native.PlayerStateValue) bool {\n\tswitch kind {\n")
	for _, state := range playerSchema.States {
		if state.Set == "" {
			continue
		}
		fmt.Fprintf(&b, "\tcase native.PlayerState%s:\n", title(state.Name))
		generateGoValidation(&b, state)
		switch state.Adapter {
		case "game_mode":
			b.WriteString("\t\tmode, ok := world.GameModeByID(int(value.Integer))\n\t\tif !ok { return false }\n")
			fmt.Fprintf(&b, "\t\tconnected.%s(mode)\n", state.Set)
		case "healing_source":
			fmt.Fprintf(&b, "\t\tconnected.%s(value.Number, pluginHealingSource{})\n", state.Set)
		case "damage_source":
			fmt.Fprintf(&b, "\t\tconnected.%s(value.Number, pluginDamageSource{})\n", state.Set)
		case "toggle":
			fmt.Fprintf(&b, "\t\tif value.Integer != 0 { connected.%s() } else { connected.%s() }\n", state.Set, state.Unset)
		case "sound":
			b.WriteString("\t\tsound, ok := playerSound(native.SoundType(value.Integer))\n\t\tif !ok { return false }\n")
			fmt.Fprintf(&b, "\t\tconnected.%s(sound)\n", state.Set)
		default:
			if state.Type == "f64" {
				fmt.Fprintf(&b, "\t\tconnected.%s(value.Number)\n", state.Set)
			} else {
				fmt.Fprintf(&b, "\t\tconnected.%s(int(value.Integer))\n", state.Set)
			}
		}
	}
	b.WriteString("\tdefault:\n\t\treturn false\n\t}\n\treturn true\n}\n\n")
	b.WriteString("func readPlayerState(connected *player.Player, kind native.PlayerStateKind) (native.PlayerStateValue, bool) {\n\tswitch kind {\n")
	for _, state := range playerSchema.States {
		if state.Get == "" {
			continue
		}
		fmt.Fprintf(&b, "\tcase native.PlayerState%s:\n", title(state.Name))
		if state.Adapter == "game_mode" {
			fmt.Fprintf(&b, "\t\tvalue, ok := world.GameModeID(connected.%s())\n\t\treturn native.PlayerStateValue{Integer: int64(value)}, ok\n", state.Get)
		} else if state.Type == "bool" {
			fmt.Fprintf(&b, "\t\tif connected.%s() { return native.PlayerStateValue{Integer: 1}, true }\n\t\treturn native.PlayerStateValue{}, true\n", state.Get)
		} else if state.Type == "f64" {
			fmt.Fprintf(&b, "\t\treturn native.PlayerStateValue{Number: connected.%s()}, true\n", state.Get)
		} else {
			fmt.Fprintf(&b, "\t\treturn native.PlayerStateValue{Integer: int64(connected.%s())}, true\n", state.Get)
		}
	}
	b.WriteString("\tdefault:\n\t\treturn native.PlayerStateValue{}, false\n\t}\n}\n")
	return b.Bytes()
}

func generateGoNativePlayerStates(player playerSchema) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated by abi-gen v%s. DO NOT EDIT.\n\n", generatorVersion)
	b.WriteString("package native\n\nimport \"time\"\n\ntype PlayerStateKind uint32\n\nconst (\n")
	for _, state := range player.States {
		fmt.Fprintf(&b, "\tPlayerState%s PlayerStateKind = %d\n", title(state.Name), state.ID)
	}
	b.WriteString(")\n\ntype PlayerStateValue struct {\n\tNumber float64\n\tInteger int64\n}\n\ntype EffectType uint32\n\nconst (\n")
	for _, effect := range player.Effects {
		fmt.Fprintf(&b, "\tEffect%s EffectType = %d\n", title(effect.Name), effect.ID)
	}
	b.WriteString(")\n\ntype PlayerEffectOperation uint32\n\nconst (\n\tPlayerEffectAdd PlayerEffectOperation = iota\n\tPlayerEffectRemove\n)\n\ntype PlayerEffect struct {\n\tType EffectType\n\tLevel int32\n\tDuration time.Duration\n\tAmbient bool\n\tInfinite bool\n\tParticlesHidden bool\n}\n")
	b.WriteString("\ntype PlayerTextKind uint32\n\nconst (\n")
	for _, text := range player.Texts {
		fmt.Fprintf(&b, "\tPlayerText%s PlayerTextKind = %d\n", title(text.Name), text.ID)
	}
	b.WriteString(")\n\ntype SoundType uint32\n\nconst (\n")
	for _, sound := range player.Sounds {
		fmt.Fprintf(&b, "\tSound%s SoundType = %d\n", title(sound.Name), sound.ID)
	}
	b.WriteString(")\n")
	return b.Bytes()
}

func generateGoValidation(b *bytes.Buffer, state playerState) {
	if state.Type == "f64" {
		b.WriteString("\t\tif math.IsNaN(value.Number) || math.IsInf(value.Number, 0)")
		switch state.Validate {
		case "non_negative":
			b.WriteString(" || value.Number < 0")
		case "positive":
			b.WriteString(" || value.Number <= 0")
		case "unit_interval":
			b.WriteString(" || value.Number < 0 || value.Number > 1")
		}
		b.WriteString(" { return false }\n")
		return
	}
	if state.Type == "i32" {
		b.WriteString("\t\tif value.Integer < math.MinInt32 || value.Integer > math.MaxInt32")
		if state.Validate == "non_negative" {
			b.WriteString(" || value.Integer < 0")
		}
		b.WriteString(" { return false }\n")
	}
	if state.Type == "bool" {
		b.WriteString("\t\tif value.Integer != 0 && value.Integer != 1 { return false }\n")
	}
}

func generateRustPlayerStates(player playerSchema) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated by abi-gen v%s. DO NOT EDIT.\n\n", generatorVersion)
	b.WriteString("#[repr(u32)]\n#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]\npub enum EffectType {\n")
	for _, effect := range player.Effects {
		fmt.Fprintf(&b, "    %s = dragonfly_plugin_sys::DF_EFFECT_%s,\n", title(effect.Name), strings.ToUpper(effect.Name))
	}
	b.WriteString("}\n\n#[repr(u32)]\n#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]\npub enum Sound {\n")
	for _, sound := range player.Sounds {
		fmt.Fprintf(&b, "    %s = dragonfly_plugin_sys::DF_SOUND_%s,\n", title(sound.Name), strings.ToUpper(sound.Name))
	}
	b.WriteString("}\n\n")
	b.WriteString("impl Player {\n")
	for _, text := range player.Texts {
		fmt.Fprintf(&b, "    pub fn %s(&self, value: &str) { self.send_text(dragonfly_plugin_sys::DF_PLAYER_TEXT_%s, value); }\n", text.Rust, strings.ToUpper(text.Name))
	}
	for _, state := range player.States {
		constant := "dragonfly_plugin_sys::DF_PLAYER_STATE_" + strings.ToUpper(state.Name)
		if state.Set != "" {
			switch state.Type {
			case "game_mode":
				fmt.Fprintf(&b, "    pub fn %s(&self, value: GameMode) { self.set_state(%s, 0.0, value as i64); }\n", state.RustSet, constant)
			case "f64":
				fmt.Fprintf(&b, "    pub fn %s(&self, value: f64) { self.set_state(%s, value, 0); }\n", state.RustSet, constant)
			case "i32":
				fmt.Fprintf(&b, "    pub fn %s(&self, value: i32) { self.set_state(%s, 0.0, i64::from(value)); }\n", state.RustSet, constant)
			case "bool":
				fmt.Fprintf(&b, "    pub fn %s(&self, value: bool) { self.set_state(%s, 0.0, value as i64); }\n", state.RustSet, constant)
			case "sound":
				fmt.Fprintf(&b, "    pub fn %s(&self, value: Sound) { self.set_state(%s, 0.0, value as i64); }\n", state.RustSet, constant)
			}
		}
		if state.Get != "" {
			switch state.Type {
			case "game_mode":
				fmt.Fprintf(&b, "    pub fn %s(&self) -> GameMode { match self.state(%s).integer { 1 => GameMode::Creative, 2 => GameMode::Adventure, 3 => GameMode::Spectator, _ => GameMode::Survival } }\n", state.RustGet, constant)
			case "f64":
				fmt.Fprintf(&b, "    pub fn %s(&self) -> f64 { self.state(%s).number }\n", state.RustGet, constant)
			case "i32":
				fmt.Fprintf(&b, "    pub fn %s(&self) -> i32 { self.state(%s).integer as i32 }\n", state.RustGet, constant)
			case "bool":
				fmt.Fprintf(&b, "    pub fn %s(&self) -> bool { self.state(%s).integer != 0 }\n", state.RustGet, constant)
			}
		}
	}
	b.WriteString("}\n")
	return b.Bytes()
}

func cType(t string) string {
	return map[string]string{
		"bool": "uint8_t", "player_id": "DfPlayerId", "rotation": "DfRotation",
		"string_buffer": "DfStringBuffer", "string_view": "DfStringView", "vec3": "DfVec3",
		"f64": "double", "u64": "uint64_t",
		"i32": "int32_t", "block_pos": "DfBlockPos", "item_stack": "DfItemStackView",
	}[t]
}

func rustType(t string) string {
	return map[string]string{
		"bool": "u8", "player_id": "DfPlayerId", "rotation": "DfRotation",
		"string_buffer": "DfStringBuffer", "string_view": "DfStringView", "vec3": "DfVec3",
		"f64": "f64", "u64": "u64",
		"i32": "i32", "block_pos": "DfBlockPos", "item_stack": "DfItemStackView",
	}[t]
}

func cName(evt event) string    { return "Df" + title(evt.Domain) + title(evt.Name) }
func rustName(evt event) string { return "Df" + title(evt.Domain) + title(evt.Name) }
func title(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}

func writeGenerated(path string, data []byte, check bool) error {
	if strings.HasSuffix(path, ".go") {
		formatted, err := format.Source(data)
		if err != nil {
			return err
		}
		data = formatted
	}
	existing, err := os.ReadFile(path)
	if err == nil && bytes.Equal(existing, data) {
		return nil
	}
	if check {
		return fmt.Errorf("generated file is stale: %s", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "abi-gen:", err)
	os.Exit(1)
}
