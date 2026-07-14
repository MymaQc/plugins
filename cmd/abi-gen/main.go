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
	Operations []playerOperation `yaml:"operations"`
	States     []playerState     `yaml:"states"`
	Effects    []effectType      `yaml:"effects"`
	Texts      []playerText      `yaml:"texts"`
	Sounds     []soundType       `yaml:"sounds"`
}

type playerOperation struct {
	Name   string               `yaml:"name"`
	ID     uint32               `yaml:"id"`
	Go     string               `yaml:"go"`
	Rust   string               `yaml:"rust"`
	Source string               `yaml:"source"`
	Result string               `yaml:"result"`
	Args   []playerOperationArg `yaml:"args"`
}

type playerOperationArg struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Validate string `yaml:"validate"`
}

type itemSchema struct {
	SimpleItems  []simpleItemType  `yaml:"simple_items"`
	BooleanItems []booleanItemType `yaml:"boolean_items"`
	ToolTiers    []toolTierType    `yaml:"tool_tiers"`
	ToolFamilies []toolFamilyType  `yaml:"tool_families"`
	ArmourTiers  []toolTierType    `yaml:"armour_tiers"`
	ArmourPieces []toolFamilyType  `yaml:"armour_pieces"`
	Enchantments []enchantmentType `yaml:"enchantments"`
	Potions      []potionType      `yaml:"potions"`
}

type simpleItemType struct {
	Name       string `yaml:"name"`
	Identifier string `yaml:"identifier"`
}

type booleanItemType struct {
	Name            string `yaml:"name"`
	Property        string `yaml:"property"`
	FalseName       string `yaml:"false_name"`
	FalseIdentifier string `yaml:"false_identifier"`
	TrueName        string `yaml:"true_name"`
	TrueIdentifier  string `yaml:"true_identifier"`
}

type toolTierType struct {
	Name       string `yaml:"name"`
	Identifier string `yaml:"identifier"`
}

type toolFamilyType struct {
	Name       string `yaml:"name"`
	Identifier string `yaml:"identifier"`
}

type enchantmentType struct {
	Name     string `yaml:"name"`
	ID       uint32 `yaml:"id"`
	MaxLevel uint32 `yaml:"max_level"`
}

type potionType struct {
	Name string `yaml:"name"`
	ID   uint32 `yaml:"id"`
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
	ID   int32  `yaml:"id"`
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
	items, err := readItems(filepath.Join(*root, "schema", "items.yaml"))
	if err != nil {
		fatal(err)
	}
	outputs := map[string][]byte{
		filepath.Join(*root, "abi", "include", "dragonfly_plugin.h"):                         generateC(events, player),
		filepath.Join(*root, "rust", "dragonfly-plugin-sys", "src", "generated.rs"):          generateRust(events, player),
		filepath.Join(*root, "internal", "host", "player_state_generated.go"):                generateGoPlayerStates(player),
		filepath.Join(*root, "internal", "native", "player_state_generated.go"):              generateGoNativePlayerStates(player),
		filepath.Join(*root, "internal", "native", "player_operations_generated.go"):         generateGoNativePlayerOperations(player),
		filepath.Join(*root, "internal", "native", "player_operations_generated.h"):          generateCNativePlayerOperations(player),
		filepath.Join(*root, "rust", "dragonfly-plugin", "src", "player_state_generated.rs"): generateRustPlayerStates(player),
		filepath.Join(*root, "rust", "dragonfly-plugin", "src", "effects_generated.rs"):      generateRustEffects(player),
		filepath.Join(*root, "rust", "dragonfly-plugin", "src", "items_generated.rs"):        generateRustItems(items),
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

func readItems(path string) (itemSchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return itemSchema{}, err
	}
	var schema itemSchema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return itemSchema{}, fmt.Errorf("decode %s: %w", path, err)
	}
	if len(schema.SimpleItems) == 0 || len(schema.BooleanItems) == 0 || len(schema.ToolTiers) == 0 || len(schema.ToolFamilies) == 0 || len(schema.ArmourTiers) == 0 || len(schema.ArmourPieces) == 0 || len(schema.Enchantments) == 0 || len(schema.Potions) == 0 {
		return itemSchema{}, fmt.Errorf("%s: item, tool, armour, enchantment, and potion lists must not be empty", path)
	}
	itemNames, itemIdentifiers := map[string]bool{}, map[string]bool{}
	typeNames := map[string]bool{"ArmourTier": true, "Custom": true, "ToolTier": true}
	for _, item := range schema.SimpleItems {
		if !validSchemaName(item.Name) || item.Identifier == "" || itemNames[item.Name] || itemIdentifiers[item.Identifier] || typeNames[title(item.Name)] {
			return itemSchema{}, fmt.Errorf("invalid or duplicate simple item %+v", item)
		}
		itemNames[item.Name], itemIdentifiers[item.Identifier] = true, true
		typeNames[title(item.Name)] = true
	}
	for _, item := range schema.BooleanItems {
		if !validSchemaName(item.Name) || !validRustValueName(item.Property) || !validRustValueName(item.FalseName) || !validRustValueName(item.TrueName) || item.FalseIdentifier == "" || item.TrueIdentifier == "" || item.FalseName == item.TrueName || item.FalseName == "is_"+item.Property || item.TrueName == "is_"+item.Property || item.FalseIdentifier == item.TrueIdentifier || itemNames[item.Name] || itemIdentifiers[item.FalseIdentifier] || itemIdentifiers[item.TrueIdentifier] || typeNames[title(item.Name)] {
			return itemSchema{}, fmt.Errorf("invalid or duplicate boolean item %+v", item)
		}
		itemNames[item.Name], itemIdentifiers[item.FalseIdentifier], itemIdentifiers[item.TrueIdentifier] = true, true, true
		typeNames[title(item.Name)] = true
	}
	tierNames, tierIdentifiers := map[string]bool{}, map[string]bool{}
	for _, tier := range schema.ToolTiers {
		if !validSchemaName(tier.Name) || tier.Identifier == "" || tierNames[tier.Name] || tierIdentifiers[tier.Identifier] {
			return itemSchema{}, fmt.Errorf("invalid or duplicate tool tier %+v", tier)
		}
		tierNames[tier.Name], tierIdentifiers[tier.Identifier] = true, true
	}
	familyNames, familyIdentifiers := map[string]bool{}, map[string]bool{}
	for _, family := range schema.ToolFamilies {
		if !validSchemaName(family.Name) || family.Identifier == "" || familyNames[family.Name] || familyIdentifiers[family.Identifier] || typeNames[title(family.Name)] {
			return itemSchema{}, fmt.Errorf("invalid or duplicate tool family %+v", family)
		}
		familyNames[family.Name], familyIdentifiers[family.Identifier] = true, true
		typeNames[title(family.Name)] = true
	}
	armourTierNames, armourTierIdentifiers := map[string]bool{}, map[string]bool{}
	for _, tier := range schema.ArmourTiers {
		if !validSchemaName(tier.Name) || tier.Identifier == "" || armourTierNames[tier.Name] || armourTierIdentifiers[tier.Identifier] {
			return itemSchema{}, fmt.Errorf("invalid or duplicate armour tier %+v", tier)
		}
		armourTierNames[tier.Name], armourTierIdentifiers[tier.Identifier] = true, true
	}
	armourPieceNames, armourPieceIdentifiers := map[string]bool{}, map[string]bool{}
	for _, piece := range schema.ArmourPieces {
		if !validSchemaName(piece.Name) || piece.Identifier == "" || armourPieceNames[piece.Name] || armourPieceIdentifiers[piece.Identifier] || typeNames[title(piece.Name)] {
			return itemSchema{}, fmt.Errorf("invalid or duplicate armour piece %+v", piece)
		}
		armourPieceNames[piece.Name], armourPieceIdentifiers[piece.Identifier] = true, true
		typeNames[title(piece.Name)] = true
	}
	enchantmentIDs, enchantmentNames := map[uint32]bool{}, map[string]bool{}
	for _, enchantment := range schema.Enchantments {
		if enchantment.Name == "" || enchantment.MaxLevel == 0 || enchantmentIDs[enchantment.ID] || enchantmentNames[enchantment.Name] {
			return itemSchema{}, fmt.Errorf("invalid or duplicate enchantment %+v", enchantment)
		}
		enchantmentIDs[enchantment.ID], enchantmentNames[enchantment.Name] = true, true
	}
	potionIDs, potionNames := map[uint32]bool{}, map[string]bool{}
	for _, potion := range schema.Potions {
		if potion.Name == "" || potion.ID > 255 || potionIDs[potion.ID] || potionNames[potion.Name] {
			return itemSchema{}, fmt.Errorf("invalid or duplicate potion %+v", potion)
		}
		potionIDs[potion.ID], potionNames[potion.Name] = true, true
	}
	sort.Slice(schema.Enchantments, func(i, j int) bool { return schema.Enchantments[i].ID < schema.Enchantments[j].ID })
	sort.Slice(schema.Potions, func(i, j int) bool { return schema.Potions[i].ID < schema.Potions[j].ID })
	sort.Slice(schema.SimpleItems, func(i, j int) bool { return schema.SimpleItems[i].Name < schema.SimpleItems[j].Name })
	sort.Slice(schema.BooleanItems, func(i, j int) bool { return schema.BooleanItems[i].Name < schema.BooleanItems[j].Name })
	return schema, nil
}

func validSchemaName(value string) bool {
	if value == "" || value[0] < 'a' || value[0] > 'z' || value[len(value)-1] == '_' {
		return false
	}
	previousUnderscore := false
	for _, character := range value {
		underscore := character == '_'
		if !(character >= 'a' && character <= 'z') && !(character >= '0' && character <= '9') && !underscore || underscore && previousUnderscore {
			return false
		}
		previousUnderscore = underscore
	}
	return true
}

func validRustValueName(value string) bool {
	if !validSchemaName(value) {
		return false
	}
	_, keyword := map[string]struct{}{
		"as": {}, "break": {}, "const": {}, "continue": {}, "crate": {}, "else": {}, "enum": {}, "extern": {}, "false": {}, "fn": {}, "for": {}, "if": {}, "impl": {}, "in": {}, "let": {}, "loop": {}, "match": {}, "mod": {}, "move": {}, "mut": {}, "pub": {}, "ref": {}, "return": {}, "self": {}, "Self": {}, "static": {}, "struct": {}, "super": {}, "trait": {}, "true": {}, "type": {}, "unsafe": {}, "use": {}, "where": {}, "while": {}, "async": {}, "await": {}, "dyn": {},
	}[value]
	return !keyword
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
	operationIDs, operationNames := map[uint32]bool{}, map[string]bool{}
	validSources := map[string]bool{"damage": true, "healing": true}
	validResults := map[string]bool{"healed": true, "damage_vulnerable": true}
	for _, operation := range schema.Operations {
		if !validSchemaName(operation.Name) || operation.Go != title(operation.Name) || !validRustValueName(operation.Rust) || operationIDs[operation.ID] || operationNames[operation.Name] {
			return playerSchema{}, fmt.Errorf("invalid or duplicate player operation %+v", operation)
		}
		if len(operation.Args) == 0 {
			if !validSources[operation.Source] || !validResults[operation.Result] || operation.Source == "healing" && operation.Result != "healed" || operation.Source == "damage" && operation.Result != "damage_vulnerable" {
				return playerSchema{}, fmt.Errorf("incompatible player operation source and result %+v", operation)
			}
		} else {
			if operation.Source != "" || operation.Result != "" {
				return playerSchema{}, fmt.Errorf("composite player operation cannot declare source or result %+v", operation)
			}
			argumentNames := map[string]bool{}
			for _, argument := range operation.Args {
				validValidation := argument.Validate == "" || argument.Type == "i32" && argument.Validate == "non_negative" || argument.Type == "f64" && (argument.Validate == "non_negative" || argument.Validate == "positive" || argument.Validate == "unit_interval")
				if !validRustValueName(argument.Name) || argumentNames[argument.Name] || argument.Type != "i32" && argument.Type != "f64" || !validValidation {
					return playerSchema{}, fmt.Errorf("invalid composite player operation argument %+v in %+v", argument, operation)
				}
				argumentNames[argument.Name] = true
			}
		}
		operationIDs[operation.ID], operationNames[operation.Name] = true, true
	}
	if len(schema.Operations) == 0 {
		return playerSchema{}, fmt.Errorf("%s: player operations must not be empty", path)
	}
	sort.Slice(schema.Operations, func(i, j int) bool { return schema.Operations[i].ID < schema.Operations[j].ID })
	ids, names := map[uint32]bool{}, map[string]bool{}
	validTypes := map[string]bool{"f64": true, "i32": true, "bool": true, "game_mode": true}
	validAdapters := map[string]bool{"": true, "game_mode": true, "toggle": true}
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
	effectIDs, effectNames := map[int32]bool{}, map[string]bool{}
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
			if err := validateFields(fullName+" input", evt.Input); err != nil {
				return nil, err
			}
			if err := validateFields(fullName+" state", evt.State); err != nil {
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
		"bool": true, "player_id": true, "entity_id": true, "world_id": true, "rotation": true, "string_buffer": true,
		"string_view": true, "vec3": true, "f64": true, "u64": true, "i32": true,
		"block_pos": true, "item_stack": true, "damage_source": true, "healing_source": true,
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

#define DF_ABI_VERSION 4u
#define DF_HOST_ABI_VERSION 20u
#define DF_STATUS_OK 0
#define DF_STATUS_ERROR 1

typedef int32_t DfStatus;
typedef uint32_t DfEventId;

typedef struct { uint8_t bytes[16]; uint64_t generation; } DfPlayerId;
typedef struct { uint8_t bytes[16]; uint64_t generation; } DfEntityId;
typedef uint64_t DfInvocationId;
typedef struct { double x; double y; double z; } DfVec3;
typedef struct { double yaw; double pitch; } DfRotation;
typedef struct { int32_t x; int32_t y; int32_t z; } DfBlockPos;
typedef struct { uint64_t value; } DfWorldId;
typedef struct { const uint8_t *data; uint64_t len; } DfStringView;
typedef struct { uint8_t *data; uint64_t len; uint64_t capacity; } DfStringBuffer;
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
typedef struct { DfPlayerId player; uint32_t kind; uint32_t reserved; } DfInventoryId;
typedef struct { uint64_t offset; uint64_t len; } DfByteSpan;
typedef struct { uint32_t id; uint32_t level; } DfItemEnchantment;
typedef struct { int32_t metadata; uint32_t count; uint32_t damage; uint8_t unbreakable; int32_t anvil_cost; uint64_t identifier_len; uint64_t custom_name_len; uint64_t lore_bytes_len; uint64_t lore_count; uint64_t nbt_len; uint64_t values_nbt_len; uint64_t enchantment_count; } DfItemStackInfo;
typedef struct { uint64_t snapshot; DfItemStackInfo info; } DfItemStackSnapshot;
typedef struct { DfStringBuffer identifier; DfStringBuffer custom_name; DfStringBuffer lore_bytes; DfStringBuffer nbt; DfStringBuffer values_nbt; DfByteSpan *lore; uint64_t lore_capacity; DfItemEnchantment *enchantments; uint64_t enchantment_capacity; } DfItemStackData;
typedef struct { DfStringView identifier; int32_t metadata; uint32_t count; uint32_t damage; uint8_t unbreakable; int32_t anvil_cost; DfStringView custom_name; const DfStringView *lore; uint64_t lore_count; DfStringView nbt; DfStringView values_nbt; const DfItemEnchantment *enchantments; uint64_t enchantment_count; } DfItemStackViewV3;
#define DF_WORLD_DIMENSION_OVERWORLD 0u
#define DF_WORLD_DIMENSION_NETHER 1u
#define DF_WORLD_DIMENSION_END 2u
#define DF_WORLD_OPEN_OR_CREATE 0u
#define DF_WORLD_OPEN_EXISTING 1u
#define DF_WORLD_CREATE_NEW 2u
#define DF_WORLD_SAVE_AUTOMATIC 0u
#define DF_WORLD_SAVE_MANUAL 1u
#define DF_WORLD_RANDOM_TICKS_DISABLED 0u
#define DF_WORLD_RANDOM_TICKS_PER_SUBCHUNK 1u
#define DF_WORLD_TIME_PRESERVE 0u
#define DF_WORLD_TIME_CYCLE 1u
#define DF_WORLD_TIME_FIXED 2u
#define DF_WORLD_WEATHER_PRESERVE 0u
#define DF_WORLD_WEATHER_CYCLE 1u
#define DF_WORLD_WEATHER_CLEAR 2u
#define DF_WORLD_CHUNK_UNLOAD_AFTER 0u
typedef struct { uint32_t struct_size; uint32_t dimension; DfStringView provider_path; uint64_t save_interval_milliseconds; uint64_t chunk_unload_interval_milliseconds; int64_t fixed_time; uint32_t open_mode; uint32_t save_policy; uint32_t random_tick_policy; uint32_t random_tick_rate; uint32_t time_policy; uint32_t weather_policy; uint32_t chunk_unload_policy; uint8_t read_only; uint8_t reserved[3]; } DfWorldOpenSpecV1;
typedef struct { DfStringBuffer identifier; DfStringBuffer properties_nbt; } DfBlockData;
typedef struct { DfStringView identifier; DfStringView properties_nbt; } DfBlockView;
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
typedef struct { DfEntityId *data; uint64_t len; uint64_t capacity; } DfEntityIdBuffer;
typedef struct { DfPlayerId *data; uint64_t len; uint64_t capacity; } DfPlayerIdBuffer;
typedef struct { DfVec3 position; DfRotation rotation; DfVec3 velocity; DfStringView name_tag; } DfEntitySpawnOptions;
#define DF_ENTITY_FAMILY_BASE 0u
#define DF_ENTITY_FAMILY_TICKING 1u
#define DF_ENTITY_FAMILY_LIVING 2u
#define DF_ENTITY_CALLBACK_STATE 1u
#define DF_ENTITY_CALLBACK_TICK 2u
#define DF_ENTITY_CALLBACK_HURT 4u
#define DF_ENTITY_CALLBACK_HEAL 8u
#define DF_ENTITY_CALLBACK_DEATH 16u
#define DF_ENTITY_PHYSICS_ENABLED 1u
#define DF_ENTITY_PHYSICS_DRAG_BEFORE_GRAVITY 2u
#define DF_ENTITY_OPERATION_ADOPT 0u
#define DF_ENTITY_OPERATION_LOAD 1u
#define DF_ENTITY_OPERATION_SAVE 2u
#define DF_ENTITY_OPERATION_TICK 3u
#define DF_ENTITY_OPERATION_HURT 4u
#define DF_ENTITY_OPERATION_HEAL 5u
#define DF_ENTITY_OPERATION_DEATH 6u
#define DF_ENTITY_OPERATION_DESTROY 7u
typedef uint64_t DfEntityInstanceId;
typedef struct { const uint8_t *data; uint64_t len; } DfBytesView;
typedef struct { uint8_t *data; uint64_t len; uint64_t capacity; } DfBytesBuffer;
typedef struct { DfStringView save_id; DfStringView network_id; DfVec3 min; DfVec3 max; uint64_t type_key; uint32_t family; uint32_t callback_flags; double initial_health; double max_health; double speed; uint32_t state_version; uint32_t physics_flags; double gravity; double drag; } DfEntityTypeDescriptorV2;
typedef struct { uint32_t kind; uint32_t flags; DfEntitySpawnOptions options; DfEntityId owner; double damage; uint64_t fuse_milliseconds; int32_t experience; uint32_t potion; int32_t punch_level; int32_t piercing_level; DfStringView text; DfStringView custom_type; DfEntityInstanceId custom_instance; const DfItemStackViewV3 *item; const DfBlockView *block; } DfEntitySpawnViewV3;
typedef struct { DfVec3 position; DfRotation rotation; DfVec3 velocity; uint32_t capabilities; DfWorldId world; DfStringBuffer entity_type; DfStringBuffer name_tag; } DfEntityState;
typedef struct { DfBytesView data; uint32_t version; uint32_t reserved; } DfEntityLoadInput;
typedef struct { uint64_t instance; } DfEntityLoadState;
typedef struct { DfBytesBuffer data; uint32_t version; uint32_t reserved; } DfEntitySaveState;
typedef struct { DfInvocationId invocation; DfEntityId entity; int64_t current; uint64_t age_milliseconds; } DfEntityTickInput;
typedef struct { uint8_t despawn; } DfEntityTickState;
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
`)
	for _, text := range player.Texts {
		fmt.Fprintf(&b, "#define DF_PLAYER_TEXT_%s %du\n", strings.ToUpper(text.Name), text.ID)
	}
	for _, state := range player.States {
		fmt.Fprintf(&b, "#define DF_PLAYER_STATE_%s %du\n", strings.ToUpper(state.Name), state.ID)
	}
	for _, operation := range player.Operations {
		fmt.Fprintf(&b, "#define DF_PLAYER_OPERATION_%s %du\n", strings.ToUpper(operation.Name), operation.ID)
	}
	for _, effect := range player.Effects {
		fmt.Fprintf(&b, "#define DF_EFFECT_%s %d\n", strings.ToUpper(effect.Name), effect.ID)
	}
	for _, sound := range player.Sounds {
		fmt.Fprintf(&b, "#define DF_SOUND_KIND_%s %du\n", strings.ToUpper(sound.Name), sound.ID)
	}
	b.WriteString(`
typedef struct { DfStringView text; DfStringView subtitle; DfStringView action_text; uint64_t fade_in_milliseconds; uint64_t duration_milliseconds; uint64_t fade_out_milliseconds; } DfTitleView;
typedef struct { DfStringView name; const DfStringView *lines; uint64_t line_count; uint8_t padding; uint8_t descending; } DfScoreboardView;
typedef DfStatus (*DfFormResponseFn)(void *callback_context, DfInvocationId invocation, DfPlayerId submitter, uint32_t outcome, DfStringView response_json);
typedef void (*DfFormDropFn)(void *callback_context);
typedef struct { DfStringView request_json; void *callback_context; DfFormResponseFn response; DfFormDropFn drop; } DfFormView;
#define DF_FORM_RESPONSE_SUBMITTED 0u
#define DF_FORM_RESPONSE_CLOSED 1u
typedef struct { double number; int64_t integer; } DfPlayerStateValue;
typedef struct { double healed; } DfPlayerHealResult;
typedef struct { double damage; uint8_t vulnerable; } DfPlayerHurtResult;
#define DF_PLAYER_EFFECT_ADD 0u
#define DF_PLAYER_EFFECT_REMOVE 1u
#define DF_EFFECT_MODE_TIMED 0u
#define DF_EFFECT_MODE_AMBIENT 1u
#define DF_EFFECT_MODE_INFINITE 2u
#define DF_EFFECT_MODE_INSTANT 3u
typedef struct { int32_t effect_type; int32_t level; uint64_t duration_milliseconds; double potency; uint32_t mode; uint8_t particles_hidden; } DfEffectView;
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
typedef DfStatus (*DfHostPlayerRotationFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, DfRotation *rotation);
typedef DfStatus (*DfHostPlayerStateSetFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue value);
typedef DfStatus (*DfHostPlayerStateGetFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, uint32_t kind, DfPlayerStateValue *value);
`)
	for _, operation := range compositePlayerOperations(player) {
		fmt.Fprintf(&b, "typedef DfStatus (*DfHostPlayer%sSetFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player%s);\n", title(operation.Name), cPlayerOperationArguments(operation))
	}
	b.WriteString(`
typedef DfStatus (*DfHostPlayerHealFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, double health, const DfHealingSourceView *source, DfPlayerHealResult *result);
typedef DfStatus (*DfHostPlayerHurtFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, double damage, const DfDamageSourceView *source, DfPlayerHurtResult *result);
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
typedef DfStatus (*DfHostWorldLookupFn)(uint64_t context, DfInvocationId invocation, DfStringView name, DfWorldId *world);
typedef DfStatus (*DfHostWorldOpenFn)(uint64_t context, DfInvocationId invocation, DfStringView name, uint32_t dimension, DfWorldId *world);
typedef DfStatus (*DfHostWorldOpenSpecFn)(uint64_t context, DfInvocationId invocation, DfStringView name, const DfWorldOpenSpecV1 *spec, DfWorldId *world);
typedef DfStatus (*DfHostWorldNameFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfStringBuffer *name);
typedef DfStatus (*DfHostWorldUnloadFn)(uint64_t context, DfInvocationId invocation, DfWorldId world);
typedef DfStatus (*DfHostWorldSaveFn)(uint64_t context, DfInvocationId invocation, DfWorldId world);
typedef DfStatus (*DfHostWorldBlockGetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, DfBlockData *block);
typedef DfStatus (*DfHostWorldLiquidGetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, DfBlockData *block);
typedef DfStatus (*DfHostWorldBlockSetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position, const DfBlockView *block);
typedef DfStatus (*DfHostWorldTimeGetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t *time);
typedef DfStatus (*DfHostWorldTimeSetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, int64_t time);
typedef DfStatus (*DfHostWorldSpawnGetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos *position);
typedef DfStatus (*DfHostWorldSpawnSetFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfBlockPos position);
typedef DfStatus (*DfHostWorldEntitySpawnFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, const DfEntitySpawnViewV3 *entity, DfEntityId *output);
typedef DfStatus (*DfHostWorldEntitiesFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfEntityIdBuffer *output);
typedef DfStatus (*DfHostWorldPlayersFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfPlayerIdBuffer *output);
typedef DfStatus (*DfHostEntityStateFn)(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfEntityState *state);
typedef DfStatus (*DfHostEntityTeleportFn)(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfVec3 position);
typedef DfStatus (*DfHostEntityVelocitySetFn)(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfVec3 velocity);
typedef DfStatus (*DfHostEntityNameTagSetFn)(uint64_t context, DfInvocationId invocation, DfEntityId entity, DfStringView name_tag);
typedef DfStatus (*DfHostEntityDespawnFn)(uint64_t context, DfInvocationId invocation, DfEntityId entity);
typedef DfStatus (*DfHostWorldParticleAddFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfVec3 position, const DfParticleViewV1 *particle);
typedef DfStatus (*DfHostWorldSoundPlayFn)(uint64_t context, DfInvocationId invocation, DfWorldId world, DfVec3 position, const DfSoundViewV1 *sound);
typedef DfStatus (*DfHostPlayerSoundPlayFn)(uint64_t context, DfInvocationId invocation, DfPlayerId player, const DfSoundViewV1 *sound);
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
    DfHostWorldLookupFn world_lookup;
    DfHostWorldOpenFn world_open;
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
    DfHostWorldEntitiesFn world_entities;
    DfHostWorldPlayersFn world_players;
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
    DfHostWorldOpenSpecFn world_open_spec;
    DfHostPlayerTransferFn player_transfer;
    DfHostPlayerEffectsFn player_effects;
    DfHostPlayerEffectsClearFn player_effects_clear;
    DfHostWorldLiquidGetFn world_liquid_get;
`)
	for _, operation := range compositePlayerOperations(player) {
		fmt.Fprintf(&b, "    DfHostPlayer%sSetFn player_%s_set;\n", title(operation.Name), operation.Name)
	}
	b.WriteString(`
} DfHostApiV20;
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
typedef struct { DfInvocationId invocation; DfStringView source; DfStringView arguments; uint32_t source_kind; DfPlayerId source_player; const DfCommandPlayer *online_players; uint64_t online_player_count; } DfCommandInput;
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
		fmt.Fprintf(&b, "    DfInvocationId invocation;\n")
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
typedef DfStatus (*DfPluginEnableFn)(void *instance, DfStringBuffer *error);
typedef DfStatus (*DfPluginDisableFn)(void *instance);
typedef const DfCommandDescriptor *(*DfPluginCommandsFn)(void *instance, uint64_t *count);
typedef uint64_t (*DfPluginEntityTypeCountFn)(void *instance);
typedef DfStatus (*DfPluginEntityTypeAtFn)(void *instance, uint64_t index, DfEntityTypeDescriptorV2 *out);
typedef DfStatus (*DfPluginHandleEntityFn)(void *instance, uint64_t local_type, uint32_t operation, uint64_t entity_instance, const void *input, void *state);
typedef DfStatus (*DfHandleCommandFn)(void *instance, uint64_t command, const DfCommandInput *input, DfCommandState *state);
typedef DfStatus (*DfCommandEnumOptionsFn)(void *instance, uint64_t command, uint64_t overload, uint64_t parameter, const DfCommandEnumContext *context, DfStringBuffer *output);
typedef DfStatus (*DfPluginSetHostFn)(void *instance, const DfHostApiV20 *host);
typedef void (*DfPluginDestroyFn)(void *instance);

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
} DfPluginApiV4;

typedef const DfPluginApiV4 *(*DfPluginEntryV4Fn)(void);

typedef struct DfRuntime DfRuntime;
typedef struct { DfStringView plugin_directory; const DfHostApiV20 *host; } DfRuntimeConfig;

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

pub const DF_ABI_VERSION: u32 = 4;
pub const DF_HOST_ABI_VERSION: u32 = 20;
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
pub type DfInvocationId = u64;
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
#[derive(Clone, Copy, Debug, Default)]
pub struct DfWorldId { pub value: u64 }
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
pub const DF_DAMAGE_SOURCE_REDUCED_BY_ARMOUR: u32 = 1;
pub const DF_DAMAGE_SOURCE_REDUCED_BY_RESISTANCE: u32 = 2;
pub const DF_DAMAGE_SOURCE_FIRE: u32 = 4;
pub const DF_DAMAGE_SOURCE_IGNORES_TOTEM: u32 = 8;
pub const DF_DAMAGE_SOURCE_FIRE_PROTECTION: u32 = 16;
pub const DF_DAMAGE_SOURCE_FEATHER_FALLING: u32 = 32;
pub const DF_DAMAGE_SOURCE_BLAST_PROTECTION: u32 = 64;
pub const DF_DAMAGE_SOURCE_PROJECTILE_PROTECTION: u32 = 128;
pub const DF_INVENTORY_MAIN: u32 = 0;
pub const DF_INVENTORY_ARMOUR: u32 = 1;
pub const DF_INVENTORY_OFFHAND: u32 = 2;
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfInventoryId { pub player: DfPlayerId, pub kind: u32, pub reserved: u32 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfByteSpan { pub offset: u64, pub len: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfItemEnchantment { pub id: u32, pub level: u32 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfItemStackInfo { pub metadata: i32, pub count: u32, pub damage: u32, pub unbreakable: u8, pub anvil_cost: i32, pub identifier_len: u64, pub custom_name_len: u64, pub lore_bytes_len: u64, pub lore_count: u64, pub nbt_len: u64, pub values_nbt_len: u64, pub enchantment_count: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfItemStackSnapshot { pub snapshot: u64, pub info: DfItemStackInfo }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfItemStackData { pub identifier: DfStringBuffer, pub custom_name: DfStringBuffer, pub lore_bytes: DfStringBuffer, pub nbt: DfStringBuffer, pub values_nbt: DfStringBuffer, pub lore: *mut DfByteSpan, pub lore_capacity: u64, pub enchantments: *mut DfItemEnchantment, pub enchantment_capacity: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfItemStackViewV3 { pub identifier: DfStringView, pub metadata: i32, pub count: u32, pub damage: u32, pub unbreakable: u8, pub anvil_cost: i32, pub custom_name: DfStringView, pub lore: *const DfStringView, pub lore_count: u64, pub nbt: DfStringView, pub values_nbt: DfStringView, pub enchantments: *const DfItemEnchantment, pub enchantment_count: u64 }
pub const DF_WORLD_DIMENSION_OVERWORLD: u32 = 0;
pub const DF_WORLD_DIMENSION_NETHER: u32 = 1;
pub const DF_WORLD_DIMENSION_END: u32 = 2;
pub const DF_WORLD_OPEN_OR_CREATE: u32 = 0;
pub const DF_WORLD_OPEN_EXISTING: u32 = 1;
pub const DF_WORLD_CREATE_NEW: u32 = 2;
pub const DF_WORLD_SAVE_AUTOMATIC: u32 = 0;
pub const DF_WORLD_SAVE_MANUAL: u32 = 1;
pub const DF_WORLD_RANDOM_TICKS_DISABLED: u32 = 0;
pub const DF_WORLD_RANDOM_TICKS_PER_SUBCHUNK: u32 = 1;
pub const DF_WORLD_TIME_PRESERVE: u32 = 0;
pub const DF_WORLD_TIME_CYCLE: u32 = 1;
pub const DF_WORLD_TIME_FIXED: u32 = 2;
pub const DF_WORLD_WEATHER_PRESERVE: u32 = 0;
pub const DF_WORLD_WEATHER_CYCLE: u32 = 1;
pub const DF_WORLD_WEATHER_CLEAR: u32 = 2;
pub const DF_WORLD_CHUNK_UNLOAD_AFTER: u32 = 0;
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfWorldOpenSpecV1 { pub struct_size: u32, pub dimension: u32, pub provider_path: DfStringView, pub save_interval_milliseconds: u64, pub chunk_unload_interval_milliseconds: u64, pub fixed_time: i64, pub open_mode: u32, pub save_policy: u32, pub random_tick_policy: u32, pub random_tick_rate: u32, pub time_policy: u32, pub weather_policy: u32, pub chunk_unload_policy: u32, pub read_only: u8, pub reserved: [u8; 3] }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfBlockData { pub identifier: DfStringBuffer, pub properties_nbt: DfStringBuffer }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfBlockView { pub identifier: DfStringView, pub properties_nbt: DfStringView }
pub const DF_DAMAGE_SOURCE_CUSTOM: u32 = 0;
pub const DF_DAMAGE_SOURCE_ATTACK: u32 = 1;
pub const DF_DAMAGE_SOURCE_BLOCK: u32 = 2;
pub const DF_DAMAGE_SOURCE_DROWNING: u32 = 3;
pub const DF_DAMAGE_SOURCE_EXPLOSION: u32 = 4;
pub const DF_DAMAGE_SOURCE_FALL: u32 = 5;
pub const DF_DAMAGE_SOURCE_FIRE_KIND: u32 = 6;
pub const DF_DAMAGE_SOURCE_GLIDE: u32 = 7;
pub const DF_DAMAGE_SOURCE_INSTANT: u32 = 8;
pub const DF_DAMAGE_SOURCE_LAVA: u32 = 9;
pub const DF_DAMAGE_SOURCE_LIGHTNING: u32 = 10;
pub const DF_DAMAGE_SOURCE_MAGMA: u32 = 11;
pub const DF_DAMAGE_SOURCE_POISON: u32 = 12;
pub const DF_DAMAGE_SOURCE_PROJECTILE: u32 = 13;
pub const DF_DAMAGE_SOURCE_STARVATION: u32 = 14;
pub const DF_DAMAGE_SOURCE_SUFFOCATION: u32 = 15;
pub const DF_DAMAGE_SOURCE_THORNS: u32 = 16;
pub const DF_DAMAGE_SOURCE_VOID: u32 = 17;
pub const DF_DAMAGE_SOURCE_WITHER: u32 = 18;
pub const DF_HEALING_SOURCE_CUSTOM: u32 = 0;
pub const DF_HEALING_SOURCE_FOOD: u32 = 1;
pub const DF_HEALING_SOURCE_INSTANT: u32 = 2;
pub const DF_HEALING_SOURCE_REGENERATION: u32 = 3;
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfDamageSourceView { pub name: DfStringView, pub kind: u32, pub flags: u32, pub entity: DfEntityId, pub secondary_entity: DfEntityId, pub block: *const DfBlockView, pub data: u8 }
impl Default for DfDamageSourceView {
    fn default() -> Self { Self { name: DfStringView::default(), kind: 0, flags: 0, entity: DfEntityId::default(), secondary_entity: DfEntityId::default(), block: core::ptr::null(), data: 0 } }
}
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfHealingSourceView { pub name: DfStringView, pub kind: u32, pub data: u8 }
pub const DF_ENTITY_TEXT: u32 = 0;
pub const DF_ENTITY_LIGHTNING: u32 = 1;
pub const DF_ENTITY_TNT: u32 = 2;
pub const DF_ENTITY_EXPERIENCE_ORB: u32 = 3;
pub const DF_ENTITY_ITEM: u32 = 4;
pub const DF_ENTITY_FALLING_BLOCK: u32 = 5;
pub const DF_ENTITY_ARROW: u32 = 6;
pub const DF_ENTITY_EGG: u32 = 7;
pub const DF_ENTITY_SNOWBALL: u32 = 8;
pub const DF_ENTITY_ENDER_PEARL: u32 = 9;
pub const DF_ENTITY_BOTTLE_OF_ENCHANTING: u32 = 10;
pub const DF_ENTITY_SPLASH_POTION: u32 = 11;
pub const DF_ENTITY_LINGERING_POTION: u32 = 12;
pub const DF_ENTITY_CUSTOM: u32 = 13;
pub const DF_ENTITY_ARROW_CRITICAL: u32 = 1;
pub const DF_ENTITY_ARROW_DISABLE_PICKUP: u32 = 2;
pub const DF_ENTITY_ARROW_OBTAIN_ON_PICKUP: u32 = 4;
pub const DF_ENTITY_LIGHTNING_BLOCK_FIRE: u32 = 1;
pub const DF_ENTITY_ITEM_HAS_PICKUP_DELAY: u32 = 1;
pub const DF_ENTITY_HAS_VELOCITY: u32 = 1;
pub const DF_ENTITY_HAS_NAME_TAG: u32 = 2;
pub const DF_ENTITY_CAN_TELEPORT: u32 = 4;
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfEntityIdBuffer { pub data: *mut DfEntityId, pub len: u64, pub capacity: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfPlayerIdBuffer { pub data: *mut DfPlayerId, pub len: u64, pub capacity: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfEntitySpawnOptions { pub position: DfVec3, pub rotation: DfRotation, pub velocity: DfVec3, pub name_tag: DfStringView }
pub const DF_ENTITY_FAMILY_BASE: u32 = 0;
pub const DF_ENTITY_FAMILY_TICKING: u32 = 1;
pub const DF_ENTITY_FAMILY_LIVING: u32 = 2;
pub const DF_ENTITY_CALLBACK_STATE: u32 = 1;
pub const DF_ENTITY_CALLBACK_TICK: u32 = 2;
pub const DF_ENTITY_CALLBACK_HURT: u32 = 4;
pub const DF_ENTITY_CALLBACK_HEAL: u32 = 8;
pub const DF_ENTITY_CALLBACK_DEATH: u32 = 16;
pub const DF_ENTITY_PHYSICS_ENABLED: u32 = 1;
pub const DF_ENTITY_PHYSICS_DRAG_BEFORE_GRAVITY: u32 = 2;
pub const DF_ENTITY_OPERATION_ADOPT: u32 = 0;
pub const DF_ENTITY_OPERATION_LOAD: u32 = 1;
pub const DF_ENTITY_OPERATION_SAVE: u32 = 2;
pub const DF_ENTITY_OPERATION_TICK: u32 = 3;
pub const DF_ENTITY_OPERATION_HURT: u32 = 4;
pub const DF_ENTITY_OPERATION_HEAL: u32 = 5;
pub const DF_ENTITY_OPERATION_DEATH: u32 = 6;
pub const DF_ENTITY_OPERATION_DESTROY: u32 = 7;
pub type DfEntityInstanceId = u64;
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfBytesView { pub data: *const u8, pub len: u64 }
impl Default for DfBytesView { fn default() -> Self { Self { data: core::ptr::null(), len: 0 } } }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfBytesBuffer { pub data: *mut u8, pub len: u64, pub capacity: u64 }
impl Default for DfBytesBuffer { fn default() -> Self { Self { data: core::ptr::null_mut(), len: 0, capacity: 0 } } }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfEntityTypeDescriptorV2 { pub save_id: DfStringView, pub network_id: DfStringView, pub min: DfVec3, pub max: DfVec3, pub type_key: u64, pub family: u32, pub callback_flags: u32, pub initial_health: f64, pub max_health: f64, pub speed: f64, pub state_version: u32, pub physics_flags: u32, pub gravity: f64, pub drag: f64 }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfEntitySpawnViewV3 { pub kind: u32, pub flags: u32, pub options: DfEntitySpawnOptions, pub owner: DfEntityId, pub damage: f64, pub fuse_milliseconds: u64, pub experience: i32, pub potion: u32, pub punch_level: i32, pub piercing_level: i32, pub text: DfStringView, pub custom_type: DfStringView, pub custom_instance: DfEntityInstanceId, pub item: *const DfItemStackViewV3, pub block: *const DfBlockView }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfEntityState { pub position: DfVec3, pub rotation: DfRotation, pub velocity: DfVec3, pub capabilities: u32, pub world: DfWorldId, pub entity_type: DfStringBuffer, pub name_tag: DfStringBuffer }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfEntityLoadInput { pub data: DfBytesView, pub version: u32, pub reserved: u32 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfEntityLoadState { pub instance: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfEntitySaveState { pub data: DfBytesBuffer, pub version: u32, pub reserved: u32 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfEntityTickInput { pub invocation: DfInvocationId, pub entity: DfEntityId, pub current: i64, pub age_milliseconds: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfEntityTickState { pub despawn: u8 }
pub const DF_PARTICLE_FLAME: u32 = 0;
pub const DF_PARTICLE_DUST: u32 = 1;
pub const DF_PARTICLE_BLOCK_BREAK: u32 = 2;
pub const DF_PARTICLE_PUNCH_BLOCK: u32 = 3;
pub const DF_PARTICLE_BLOCK_FORCE_FIELD: u32 = 4;
pub const DF_PARTICLE_BONE_MEAL: u32 = 5;
pub const DF_PARTICLE_NOTE: u32 = 6;
pub const DF_PARTICLE_DRAGON_EGG_TELEPORT: u32 = 7;
pub const DF_PARTICLE_EVAPORATE: u32 = 8;
pub const DF_PARTICLE_WATER_DRIP: u32 = 9;
pub const DF_PARTICLE_LAVA_DRIP: u32 = 10;
pub const DF_PARTICLE_LAVA: u32 = 11;
pub const DF_PARTICLE_DUST_PLUME: u32 = 12;
pub const DF_PARTICLE_HUGE_EXPLOSION: u32 = 13;
pub const DF_PARTICLE_ENDERMAN_TELEPORT: u32 = 14;
pub const DF_PARTICLE_SNOWBALL_POOF: u32 = 15;
pub const DF_PARTICLE_EGG_SMASH: u32 = 16;
pub const DF_PARTICLE_SPLASH: u32 = 17;
pub const DF_PARTICLE_EFFECT: u32 = 18;
pub const DF_PARTICLE_ENTITY_FLAME: u32 = 19;
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfRgba { pub r: u8, pub g: u8, pub b: u8, pub a: u8 }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfParticleViewV1 { pub kind: u32, pub data: u32, pub pitch: i32, pub colour: DfRgba, pub diff: DfBlockPos, pub block: *const DfBlockView }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfSoundViewV1 { pub kind: u32, pub data: u32, pub integer: i32, pub flags: u32, pub scalar: f64, pub block: *const DfBlockView, pub item: *const DfItemStackViewV3 }
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
	for _, operation := range player.Operations {
		fmt.Fprintf(&b, "pub const DF_PLAYER_OPERATION_%s: u32 = %d;\n", strings.ToUpper(operation.Name), operation.ID)
	}
	for _, effect := range player.Effects {
		fmt.Fprintf(&b, "pub const DF_EFFECT_%s: i32 = %d;\n", strings.ToUpper(effect.Name), effect.ID)
	}
	for _, sound := range player.Sounds {
		fmt.Fprintf(&b, "pub const DF_SOUND_KIND_%s: u32 = %d;\n", strings.ToUpper(sound.Name), sound.ID)
	}
	b.WriteString(`
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfTitleView { pub text: DfStringView, pub subtitle: DfStringView, pub action_text: DfStringView, pub fade_in_milliseconds: u64, pub duration_milliseconds: u64, pub fade_out_milliseconds: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfScoreboardView { pub name: DfStringView, pub lines: *const DfStringView, pub line_count: u64, pub padding: u8, pub descending: u8 }
pub const DF_FORM_RESPONSE_SUBMITTED: u32 = 0;
pub const DF_FORM_RESPONSE_CLOSED: u32 = 1;
pub type DfFormResponseFn = unsafe extern "C" fn(callback_context: *mut c_void, invocation: DfInvocationId, submitter: DfPlayerId, outcome: u32, response_json: DfStringView) -> DfStatus;
pub type DfFormDropFn = unsafe extern "C" fn(callback_context: *mut c_void);
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfFormView { pub request_json: DfStringView, pub callback_context: *mut c_void, pub response: Option<DfFormResponseFn>, pub drop: Option<DfFormDropFn> }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfPlayerStateValue { pub number: f64, pub integer: i64 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfPlayerHealResult { pub healed: f64 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfPlayerHurtResult { pub damage: f64, pub vulnerable: u8 }
pub const DF_PLAYER_EFFECT_ADD: u32 = 0;
pub const DF_PLAYER_EFFECT_REMOVE: u32 = 1;
pub const DF_EFFECT_MODE_TIMED: u32 = 0;
pub const DF_EFFECT_MODE_AMBIENT: u32 = 1;
pub const DF_EFFECT_MODE_INFINITE: u32 = 2;
pub const DF_EFFECT_MODE_INSTANT: u32 = 3;
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfEffectView { pub effect_type: i32, pub level: i32, pub duration_milliseconds: u64, pub potency: f64, pub mode: u32, pub particles_hidden: u8 }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfEffectBuffer { pub data: *mut DfEffectView, pub len: u64, pub capacity: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfSkinAnimationInfo { pub width: u32, pub height: u32, pub animation_type: u32, pub frame_count: i64, pub expression: i64, pub pixels_len: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct DfSkinInfo { pub width: u32, pub height: u32, pub persona: u8, pub play_fab_id_len: u64, pub full_id_len: u64, pub pixels_len: u64, pub model_default_len: u64, pub model_animated_face_len: u64, pub model_len: u64, pub cape_width: u32, pub cape_height: u32, pub cape_pixels_len: u64, pub animation_count: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfSkinData { pub play_fab_id: DfStringBuffer, pub full_id: DfStringBuffer, pub pixels: DfStringBuffer, pub model_default: DfStringBuffer, pub model_animated_face: DfStringBuffer, pub model: DfStringBuffer, pub cape_pixels: DfStringBuffer, pub animation_pixels: *mut DfStringBuffer, pub animation_capacity: u64 }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfSkinAnimationView { pub width: u32, pub height: u32, pub animation_type: u32, pub frame_count: i64, pub expression: i64, pub pixels: DfStringView }
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfSkinView { pub width: u32, pub height: u32, pub persona: u8, pub play_fab_id: DfStringView, pub full_id: DfStringView, pub pixels: DfStringView, pub model_default: DfStringView, pub model_animated_face: DfStringView, pub model: DfStringView, pub cape_width: u32, pub cape_height: u32, pub cape_pixels: DfStringView, pub animations: *const DfSkinAnimationView, pub animation_count: u64 }
pub type DfHostPlayerTextFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, kind: u32, message: DfStringView) -> DfStatus;
pub type DfHostPlayerTitleFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, title: DfTitleView) -> DfStatus;
pub type DfHostPlayerScoreboardFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, scoreboard: DfScoreboardView) -> DfStatus;
pub type DfHostPlayerScoreboardRemoveFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId) -> DfStatus;
pub type DfHostPlayerFormSendFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, form: *const DfFormView) -> DfStatus;
pub type DfHostPlayerFormCloseFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId) -> DfStatus;
pub type DfHostPlayerTransformFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, kind: u32, vector: DfVec3, yaw: f64, pitch: f64) -> DfStatus;
pub type DfHostPlayerRotationFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, rotation: *mut DfRotation) -> DfStatus;
pub type DfHostPlayerStateSetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, kind: u32, value: DfPlayerStateValue) -> DfStatus;
pub type DfHostPlayerStateGetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, kind: u32, value: *mut DfPlayerStateValue) -> DfStatus;
`)
	for _, operation := range compositePlayerOperations(player) {
		fmt.Fprintf(&b, "pub type DfHostPlayer%sSetFn = unsafe extern \"C\" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId%s) -> DfStatus;\n", title(operation.Name), rustPlayerOperationArguments(operation))
	}
	b.WriteString(`
pub type DfHostPlayerHealFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, health: f64, source: *const DfHealingSourceView, result: *mut DfPlayerHealResult) -> DfStatus;
pub type DfHostPlayerHurtFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, damage: f64, source: *const DfDamageSourceView, result: *mut DfPlayerHurtResult) -> DfStatus;
pub type DfHostPlayerEffectFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, operation: u32, effect: DfEffectView) -> DfStatus;
pub type DfHostPlayerEffectsFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, output: *mut DfEffectBuffer) -> DfStatus;
pub type DfHostPlayerEffectsClearFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId) -> DfStatus;
pub type DfHostPlayerEntityVisibilityFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, entity: DfEntityId, visible: u8) -> DfStatus;
pub type DfHostPlayerSkinOpenFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, snapshot: *mut u64, info: *mut DfSkinInfo) -> DfStatus;
pub type DfHostPlayerSkinAnimationInfoFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, snapshot: u64, index: u64, info: *mut DfSkinAnimationInfo) -> DfStatus;
pub type DfHostPlayerSkinReadFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, snapshot: u64, data: *mut DfSkinData) -> DfStatus;
pub type DfHostPlayerSkinCloseFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, snapshot: u64);
pub type DfHostPlayerSkinSetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, skin: *const DfSkinView) -> DfStatus;
pub type DfHostPlayerTransferFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, world: DfWorldId, position: DfVec3) -> DfStatus;
pub type DfHostSkinSnapshotInfoFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, snapshot: u64, info: *mut DfSkinInfo) -> DfStatus;
pub type DfHostSkinSnapshotSetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, snapshot: u64, skin: *const DfSkinView) -> DfStatus;
pub type DfHostInventorySizeFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, inventory: DfInventoryId, size: *mut u32) -> DfStatus;
pub type DfHostInventoryItemOpenFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, inventory: DfInventoryId, slot: u32, snapshot: *mut u64, info: *mut DfItemStackInfo) -> DfStatus;
pub type DfHostPlayerHeldItemOpenFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, hand: u32, snapshot: *mut u64, info: *mut DfItemStackInfo) -> DfStatus;
pub type DfHostItemStackReadFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, snapshot: u64, data: *mut DfItemStackData) -> DfStatus;
pub type DfHostItemStackCloseFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, snapshot: u64);
pub type DfHostInventoryItemSetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, inventory: DfInventoryId, slot: u32, item: *const DfItemStackViewV3) -> DfStatus;
pub type DfHostInventoryItemAddFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, inventory: DfInventoryId, item: *const DfItemStackViewV3, added: *mut u32) -> DfStatus;
pub type DfHostInventoryClearSlotFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, inventory: DfInventoryId, slot: u32) -> DfStatus;
pub type DfHostInventoryClearFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, inventory: DfInventoryId) -> DfStatus;
pub type DfHostPlayerHeldItemsSetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, main_hand: *const DfItemStackViewV3, off_hand: *const DfItemStackViewV3) -> DfStatus;
pub type DfHostPlayerHeldSlotSetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, slot: u32) -> DfStatus;
pub type DfHostWorldLookupFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, name: DfStringView, world: *mut DfWorldId) -> DfStatus;
pub type DfHostWorldOpenFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, name: DfStringView, dimension: u32, world: *mut DfWorldId) -> DfStatus;
pub type DfHostWorldOpenSpecFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, name: DfStringView, spec: *const DfWorldOpenSpecV1, world: *mut DfWorldId) -> DfStatus;
pub type DfHostWorldNameFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId, name: *mut DfStringBuffer) -> DfStatus;
pub type DfHostWorldUnloadFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId) -> DfStatus;
pub type DfHostWorldSaveFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId) -> DfStatus;
pub type DfHostWorldBlockGetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId, position: DfBlockPos, block: *mut DfBlockData) -> DfStatus;
pub type DfHostWorldLiquidGetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId, position: DfBlockPos, block: *mut DfBlockData) -> DfStatus;
pub type DfHostWorldBlockSetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId, position: DfBlockPos, block: *const DfBlockView) -> DfStatus;
pub type DfHostWorldTimeGetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId, time: *mut i64) -> DfStatus;
pub type DfHostWorldTimeSetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId, time: i64) -> DfStatus;
pub type DfHostWorldSpawnGetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId, position: *mut DfBlockPos) -> DfStatus;
pub type DfHostWorldSpawnSetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId, position: DfBlockPos) -> DfStatus;
pub type DfHostWorldEntitySpawnFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId, entity: *const DfEntitySpawnViewV3, output: *mut DfEntityId) -> DfStatus;
pub type DfHostWorldEntitiesFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId, output: *mut DfEntityIdBuffer) -> DfStatus;
pub type DfHostWorldPlayersFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId, output: *mut DfPlayerIdBuffer) -> DfStatus;
pub type DfHostEntityStateFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, entity: DfEntityId, state: *mut DfEntityState) -> DfStatus;
pub type DfHostEntityTeleportFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, entity: DfEntityId, position: DfVec3) -> DfStatus;
pub type DfHostEntityVelocitySetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, entity: DfEntityId, velocity: DfVec3) -> DfStatus;
pub type DfHostEntityNameTagSetFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, entity: DfEntityId, name_tag: DfStringView) -> DfStatus;
pub type DfHostEntityDespawnFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, entity: DfEntityId) -> DfStatus;
pub type DfHostWorldParticleAddFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId, position: DfVec3, particle: *const DfParticleViewV1) -> DfStatus;
pub type DfHostWorldSoundPlayFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, world: DfWorldId, position: DfVec3, sound: *const DfSoundViewV1) -> DfStatus;
pub type DfHostPlayerSoundPlayFn = unsafe extern "C" fn(context: u64, invocation: DfInvocationId, player: DfPlayerId, sound: *const DfSoundViewV1) -> DfStatus;
#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct DfHostApiV20 { pub abi_version: u32, pub struct_size: u32, pub context: u64, pub player_text: Option<DfHostPlayerTextFn>, pub player_title: Option<DfHostPlayerTitleFn>, pub player_transform: Option<DfHostPlayerTransformFn>, pub player_rotation: Option<DfHostPlayerRotationFn>, pub player_state_set: Option<DfHostPlayerStateSetFn>, pub player_state_get: Option<DfHostPlayerStateGetFn>, pub player_effect: Option<DfHostPlayerEffectFn>, pub player_entity_visibility: Option<DfHostPlayerEntityVisibilityFn>, pub player_skin_open: Option<DfHostPlayerSkinOpenFn>, pub player_skin_animation_info: Option<DfHostPlayerSkinAnimationInfoFn>, pub player_skin_read: Option<DfHostPlayerSkinReadFn>, pub player_skin_close: Option<DfHostPlayerSkinCloseFn>, pub player_skin_set: Option<DfHostPlayerSkinSetFn>, pub inventory_size: Option<DfHostInventorySizeFn>, pub inventory_item_open: Option<DfHostInventoryItemOpenFn>, pub player_held_item_open: Option<DfHostPlayerHeldItemOpenFn>, pub item_stack_read: Option<DfHostItemStackReadFn>, pub item_stack_close: Option<DfHostItemStackCloseFn>, pub inventory_item_set: Option<DfHostInventoryItemSetFn>, pub inventory_item_add: Option<DfHostInventoryItemAddFn>, pub inventory_clear_slot: Option<DfHostInventoryClearSlotFn>, pub inventory_clear: Option<DfHostInventoryClearFn>, pub player_held_items_set: Option<DfHostPlayerHeldItemsSetFn>, pub player_held_slot_set: Option<DfHostPlayerHeldSlotSetFn>, pub player_scoreboard: Option<DfHostPlayerScoreboardFn>, pub player_scoreboard_remove: Option<DfHostPlayerScoreboardRemoveFn>, pub player_form_send: Option<DfHostPlayerFormSendFn>, pub player_form_close: Option<DfHostPlayerFormCloseFn>, pub world_lookup: Option<DfHostWorldLookupFn>, pub world_open: Option<DfHostWorldOpenFn>, pub world_name: Option<DfHostWorldNameFn>, pub world_unload: Option<DfHostWorldUnloadFn>, pub world_save: Option<DfHostWorldSaveFn>, pub world_block_get: Option<DfHostWorldBlockGetFn>, pub world_block_set: Option<DfHostWorldBlockSetFn>, pub world_time_get: Option<DfHostWorldTimeGetFn>, pub world_time_set: Option<DfHostWorldTimeSetFn>, pub world_spawn_get: Option<DfHostWorldSpawnGetFn>, pub world_spawn_set: Option<DfHostWorldSpawnSetFn>, pub world_entity_spawn: Option<DfHostWorldEntitySpawnFn>, pub world_entities: Option<DfHostWorldEntitiesFn>, pub world_players: Option<DfHostWorldPlayersFn>, pub entity_state: Option<DfHostEntityStateFn>, pub entity_teleport: Option<DfHostEntityTeleportFn>, pub entity_velocity_set: Option<DfHostEntityVelocitySetFn>, pub entity_name_tag_set: Option<DfHostEntityNameTagSetFn>, pub entity_despawn: Option<DfHostEntityDespawnFn>, pub world_particle_add: Option<DfHostWorldParticleAddFn>, pub world_sound_play: Option<DfHostWorldSoundPlayFn>, pub player_sound_play: Option<DfHostPlayerSoundPlayFn>, pub player_heal: Option<DfHostPlayerHealFn>, pub player_hurt: Option<DfHostPlayerHurtFn>, pub skin_snapshot_info: Option<DfHostSkinSnapshotInfoFn>, pub skin_snapshot_set: Option<DfHostSkinSnapshotSetFn>, pub world_open_spec: Option<DfHostWorldOpenSpecFn>, pub player_transfer: Option<DfHostPlayerTransferFn>, pub player_effects: Option<DfHostPlayerEffectsFn>, pub player_effects_clear: Option<DfHostPlayerEffectsClearFn>, pub world_liquid_get: Option<DfHostWorldLiquidGetFn>`)
	for _, operation := range compositePlayerOperations(player) {
		fmt.Fprintf(&b, ", pub player_%s_set: Option<DfHostPlayer%sSetFn>", operation.Name, title(operation.Name))
	}
	b.WriteString(" }\n")
	b.WriteString(`
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
pub struct DfCommandInput { pub invocation: DfInvocationId, pub source: DfStringView, pub arguments: DfStringView, pub source_kind: u32, pub source_player: DfPlayerId, pub online_players: *const DfCommandPlayer, pub online_player_count: u64 }
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
		b.WriteString("    pub invocation: DfInvocationId,\n")
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
pub type DfPluginEnableFn = unsafe extern "C" fn(instance: *mut c_void, error: *mut DfStringBuffer) -> DfStatus;
pub type DfPluginDisableFn = unsafe extern "C" fn(instance: *mut c_void) -> DfStatus;
pub type DfPluginCommandsFn = unsafe extern "C" fn(instance: *mut c_void, count: *mut u64) -> *const DfCommandDescriptor;
pub type DfPluginEntityTypeCountFn = unsafe extern "C" fn(instance: *mut c_void) -> u64;
pub type DfPluginEntityTypeAtFn = unsafe extern "C" fn(instance: *mut c_void, index: u64, out: *mut DfEntityTypeDescriptorV2) -> DfStatus;
pub type DfPluginHandleEntityFn = unsafe extern "C" fn(instance: *mut c_void, local_type: u64, operation: u32, entity_instance: u64, input: *const c_void, state: *mut c_void) -> DfStatus;
pub type DfHandleCommandFn = unsafe extern "C" fn(instance: *mut c_void, command: u64, input: *const DfCommandInput, state: *mut DfCommandState) -> DfStatus;
pub type DfCommandEnumOptionsFn = unsafe extern "C" fn(instance: *mut c_void, command: u64, overload: u64, parameter: u64, context: *const DfCommandEnumContext, output: *mut DfStringBuffer) -> DfStatus;
pub type DfPluginSetHostFn = unsafe extern "C" fn(instance: *mut c_void, host: *const DfHostApiV20) -> DfStatus;
pub type DfPluginDestroyFn = unsafe extern "C" fn(instance: *mut c_void);
pub type DfHandleEventFn = unsafe extern "C" fn(instance: *mut c_void, event_id: DfEventId, input: *const c_void, state: *mut c_void) -> DfStatus;

#[repr(C)]
pub struct DfPluginApiV4 {
    pub header: DfAbiHeader,
    pub plugin_id: DfStringView,
    pub create: Option<DfPluginCreateFn>,
    pub enable: Option<DfPluginEnableFn>,
    pub disable: Option<DfPluginDisableFn>,
    pub commands: Option<DfPluginCommandsFn>,
    pub entity_type_count: Option<DfPluginEntityTypeCountFn>,
    pub entity_type_at: Option<DfPluginEntityTypeAtFn>,
    pub handle_entity: Option<DfPluginHandleEntityFn>,
    pub handle_command: Option<DfHandleCommandFn>,
    pub command_enum_options: Option<DfCommandEnumOptionsFn>,
    pub set_host: Option<DfPluginSetHostFn>,
    pub destroy: Option<DfPluginDestroyFn>,
    pub handle_event: Option<DfHandleEventFn>,
}

pub type DfPluginEntryV4Fn = unsafe extern "C" fn() -> *const DfPluginApiV4;

unsafe impl Sync for DfPluginApiV4 {}
`)
	return b.Bytes()
}

func generateGoPlayerStates(playerSchema playerSchema) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated by abi-gen v%s. DO NOT EDIT.\n\n", generatorVersion)
	b.WriteString("package host\n\nimport (\n\t\"math\"\n\n\t\"github.com/bedrock-gophers/plugins/internal/native\"\n\t\"github.com/df-mc/dragonfly/server/player\"\n\t\"github.com/df-mc/dragonfly/server/world\"\n)\n\n")
	b.WriteString("func sendPlayerText(connected *player.Player, kind native.PlayerTextKind, message string) bool {\n\tswitch kind {\n")
	for _, text := range playerSchema.Texts {
		fmt.Fprintf(&b, "\tcase native.PlayerText%s:\n\t\tconnected.%s(message)\n", title(text.Name), text.Set)
	}
	b.WriteString("\tdefault:\n\t\treturn false\n\t}\n\treturn true\n}\n\n")
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
		case "toggle":
			fmt.Fprintf(&b, "\t\tif value.Integer != 0 { connected.%s() } else { connected.%s() }\n", state.Set, state.Unset)
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
	b.WriteString(")\n\ntype PlayerStateValue struct {\n\tNumber float64\n\tInteger int64\n}\n\ntype EffectType int32\n\nconst (\n")
	for _, effect := range player.Effects {
		fmt.Fprintf(&b, "\tEffect%s EffectType = %d\n", title(effect.Name), effect.ID)
	}
	b.WriteString(")\n\ntype PlayerEffectOperation uint32\n\nconst (\n\tPlayerEffectAdd PlayerEffectOperation = iota\n\tPlayerEffectRemove\n)\n\ntype PlayerEffectMode uint32\n\nconst (\n\tPlayerEffectTimed PlayerEffectMode = iota\n\tPlayerEffectAmbient\n\tPlayerEffectInfinite\n\tPlayerEffectInstant\n)\n\ntype PlayerEffect struct {\n\tType EffectType\n\tLevel int32\n\tDuration time.Duration\n\tPotency float64\n\tMode PlayerEffectMode\n\tParticlesHidden bool\n}\n")
	b.WriteString("\ntype PlayerTextKind uint32\n\nconst (\n")
	for _, text := range player.Texts {
		fmt.Fprintf(&b, "\tPlayerText%s PlayerTextKind = %d\n", title(text.Name), text.ID)
	}
	b.WriteString(")\n\ntype SoundKind uint32\n\nconst (\n")
	for _, sound := range player.Sounds {
		fmt.Fprintf(&b, "\tSound%s SoundKind = %d\n", title(sound.Name), sound.ID)
	}
	b.WriteString(")\n")
	return b.Bytes()
}

func generateGoNativePlayerOperations(player playerSchema) []byte {
	operations := compositePlayerOperations(player)
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated by abi-gen v%s. DO NOT EDIT.\n\n", generatorVersion)
	b.WriteString("package native\n\n")
	if len(operations) != 0 {
		b.WriteString("/*\n#include \"bridge.h\"\n*/\nimport \"C\"\n")
		hasFloat := false
		for _, operation := range operations {
			for _, argument := range operation.Args {
				hasFloat = hasFloat || argument.Type == "f64"
			}
		}
		if hasFloat {
			b.WriteString("\nimport \"math\"\n")
		}
		b.WriteString("\n")
	}
	b.WriteString("type generatedPlayerOperationsHost interface {\n")
	for _, operation := range operations {
		fmt.Fprintf(&b, "\tSetPlayer%s(InvocationID, PlayerID%s) bool\n", operation.Go, goPlayerOperationArguments(operation))
	}
	b.WriteString("}\n\n")
	for _, operation := range operations {
		fmt.Fprintf(&b, "func (noopHost) SetPlayer%s(InvocationID, PlayerID%s) bool { return false }\n\n", operation.Go, goPlayerOperationArguments(operation))
		fmt.Fprintf(&b, "//export bg_go_player_%s_set\n", operation.Name)
		fmt.Fprintf(&b, "func bg_go_player_%s_set(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId%s) C.DfStatus {\n", operation.Name, goCPlayerOperationArguments(operation))
		for _, argument := range operation.Args {
			fmt.Fprintf(&b, "\t%sValue := %s(%s)\n", argument.Name, goType(argument.Type), argument.Name)
		}
		b.WriteString("\thost, ok := resolveHost(uint64(context))\n")
		validations := goPlayerOperationValidations(operation)
		if validations != "" {
			fmt.Fprintf(&b, "\tif !ok || %s || !host.SetPlayer%s(InvocationID(invocation), playerID(player)%s) {\n", validations, operation.Go, goPlayerOperationValues(operation))
		} else {
			fmt.Fprintf(&b, "\tif !ok || !host.SetPlayer%s(InvocationID(invocation), playerID(player)%s) {\n", operation.Go, goPlayerOperationValues(operation))
		}
		b.WriteString("\t\treturn C.DF_STATUS_ERROR\n\t}\n\treturn C.DF_STATUS_OK\n}\n\n")
	}
	return b.Bytes()
}

func generateCNativePlayerOperations(player playerSchema) []byte {
	operations := compositePlayerOperations(player)
	var b bytes.Buffer
	fmt.Fprintf(&b, "/* Code generated by abi-gen v%s. DO NOT EDIT. */\n", generatorVersion)
	b.WriteString("#ifndef BEDROCK_GOPHERS_PLAYER_OPERATIONS_GENERATED_H\n#define BEDROCK_GOPHERS_PLAYER_OPERATIONS_GENERATED_H\n\n")
	b.WriteString("#if UINTPTR_MAX == UINT64_MAX\n")
	for index, operation := range operations {
		offset := 488 + index*8
		fmt.Fprintf(&b, "_Static_assert(offsetof(DfHostApiV20, player_%s_set) == %d, \"DfHostApiV20.player_%s_set ABI offset changed\");\n", operation.Name, offset, operation.Name)
	}
	b.WriteString("#endif\n\n")
	for _, operation := range operations {
		fmt.Fprintf(&b, "extern DfStatus bg_go_player_%s_set(uint64_t context, DfInvocationId invocation, DfPlayerId player%s);\n", operation.Name, cPlayerOperationArguments(operation))
		fmt.Fprintf(&b, "static DfStatus host_player_%s_set(uint64_t context, DfInvocationId invocation, DfPlayerId player%s) {\n", operation.Name, cPlayerOperationArguments(operation))
		fmt.Fprintf(&b, "    return bg_go_player_%s_set(context, invocation, player%s);\n}\n\n", operation.Name, playerOperationValues(operation))
	}
	b.WriteString("#define BG_PLAYER_OPERATION_HOST_FIELDS")
	for _, operation := range operations {
		fmt.Fprintf(&b, " \\\n    .player_%s_set = host_player_%s_set,", operation.Name, operation.Name)
	}
	b.WriteString("\n\n#endif\n")
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
	for _, operation := range compositePlayerOperations(player) {
		fmt.Fprintf(&b, "    /// Performs the `%s` player operation. Invalid values and unavailable hosts are ignored.\n", operation.Name)
		fmt.Fprintf(&b, "    pub fn %s(&self%s) {\n", operation.Rust, rustPlayerOperationArguments(operation))
		if validation := rustPlayerOperationValidations(operation); validation != "" {
			fmt.Fprintf(&b, "        if %s { return; }\n", validation)
		}
		fmt.Fprintf(&b, "        let Some(host) = host_api() else { return };\n        let Some(call) = host.player_%s_set else { return };\n", operation.Name)
		fmt.Fprintf(&b, "        let _ = unsafe { call(host.context, current_invocation(), self.raw_id()%s) };\n", playerOperationValues(operation))
		b.WriteString("    }\n")
	}
	b.WriteString("}\n")
	return b.Bytes()
}

func generateRustEffects(player playerSchema) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated by abi-gen v%s. DO NOT EDIT.\n\n", generatorVersion)
	for _, effect := range player.Effects {
		name := title(effect.Name)
		fmt.Fprintf(&b, "#[derive(Clone, Copy, Debug, Default, Eq, Hash, PartialEq)]\npub struct %s;\n", name)
		fmt.Fprintf(&b, "impl private::Sealed for %s {}\n", name)
		fmt.Fprintf(&b, "impl Type for %s { fn id(self) -> i32 { dragonfly_plugin_sys::DF_EFFECT_%s } }\n", name, strings.ToUpper(effect.Name))
		switch effect.Kind {
		case "lasting":
			fmt.Fprintf(&b, "impl LastingType for %s {}\n\n", name)
		case "instant":
			fmt.Fprintf(&b, "impl InstantType for %s {}\n\n", name)
		default:
			panic(fmt.Sprintf("effect %q has unknown kind %q", effect.Name, effect.Kind))
		}
	}
	return b.Bytes()
}

func generateRustItems(items itemSchema) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated by abi-gen v%s. DO NOT EDIT.\n\n", generatorVersion)
	b.WriteString("pub trait Item {\n    fn identifier(&self) -> &str;\n\n    fn metadata(&self) -> i16 { 0 }\n}\n\nimpl<T: Item + ?Sized> Item for &T {\n    fn identifier(&self) -> &str { (*self).identifier() }\n\n    fn metadata(&self) -> i16 { (*self).metadata() }\n}\n\n")
	b.WriteString("pub mod item {\n    use super::Item;\n\n")
	for _, item := range items.SimpleItems {
		name := title(item.Name)
		fmt.Fprintf(&b, "    #[derive(Clone, Copy, Debug, Default, Eq, Hash, PartialEq)]\n    pub struct %s;\n\n    impl Item for %s {\n        fn identifier(&self) -> &str { %q }\n    }\n\n", name, name, item.Identifier)
	}
	for _, item := range items.BooleanItems {
		name := title(item.Name)
		fmt.Fprintf(&b, "    #[derive(Clone, Copy, Debug, Default, Eq, Hash, PartialEq)]\n    pub struct %s { %s: bool }\n\n    impl %s {\n        pub const fn %s() -> Self { Self { %s: false } }\n\n        pub const fn %s() -> Self { Self { %s: true } }\n\n        pub const fn is_%s(self) -> bool { self.%s }\n    }\n\n    impl Item for %s {\n        fn identifier(&self) -> &str {\n            if self.%s { %q } else { %q }\n        }\n    }\n\n", name, item.Property, name, item.FalseName, item.Property, item.TrueName, item.Property, item.Property, item.Property, name, item.Property, item.TrueIdentifier, item.FalseIdentifier)
	}
	b.WriteString("    #[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]\n    pub enum ToolTier {\n")
	for _, tier := range items.ToolTiers {
		fmt.Fprintf(&b, "        %s,\n", title(tier.Name))
	}
	b.WriteString("    }\n\n")
	for _, family := range items.ToolFamilies {
		name := title(family.Name)
		fmt.Fprintf(&b, "    #[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]\n    pub struct %s { tier: ToolTier }\n\n    impl %s {\n        pub const fn new(tier: ToolTier) -> Self { Self { tier } }\n\n        pub const fn tier(self) -> ToolTier { self.tier }\n    }\n\n    impl Item for %s {\n        fn identifier(&self) -> &str {\n            match self.tier {\n", name, name, name)
		for _, tier := range items.ToolTiers {
			identifier := "minecraft:" + tier.Identifier + "_" + family.Identifier
			fmt.Fprintf(&b, "                ToolTier::%s => %q,\n", title(tier.Name), identifier)
		}
		b.WriteString("            }\n        }\n    }\n\n")
	}
	b.WriteString("    #[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]\n    pub enum ArmourTier {\n")
	for _, tier := range items.ArmourTiers {
		fmt.Fprintf(&b, "        %s,\n", title(tier.Name))
	}
	b.WriteString("    }\n\n")
	for _, piece := range items.ArmourPieces {
		name := title(piece.Name)
		fmt.Fprintf(&b, "    #[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]\n    pub struct %s { tier: ArmourTier }\n\n    impl %s {\n        pub const fn new(tier: ArmourTier) -> Self { Self { tier } }\n\n        pub const fn tier(self) -> ArmourTier { self.tier }\n    }\n\n    impl Item for %s {\n        fn identifier(&self) -> &str {\n            match self.tier {\n", name, name, name)
		for _, tier := range items.ArmourTiers {
			identifier := "minecraft:" + tier.Identifier + "_" + piece.Identifier
			fmt.Fprintf(&b, "                ArmourTier::%s => %q,\n", title(tier.Name), identifier)
		}
		b.WriteString("            }\n        }\n    }\n\n")
	}
	b.WriteString("    #[derive(Clone, Debug, Eq, Hash, PartialEq)]\n    pub struct Custom { identifier: std::string::String, metadata: i16 }\n\n    impl Custom {\n        pub fn new(identifier: impl Into<std::string::String>) -> Self {\n            Self { identifier: identifier.into(), metadata: 0 }\n        }\n\n        pub fn with_metadata(mut self, metadata: i16) -> Self {\n            self.metadata = metadata;\n            self\n        }\n    }\n\n    impl Item for Custom {\n        fn identifier(&self) -> &str { &self.identifier }\n\n        fn metadata(&self) -> i16 { self.metadata }\n    }\n\n    pub fn new(item: impl Item, count: u32) -> super::ItemStack {\n        super::ItemStack::new(item, count)\n    }\n}\n\n")
	b.WriteString("#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]\npub enum Enchantment {\n")
	for _, enchantment := range items.Enchantments {
		fmt.Fprintf(&b, "    %s,\n", title(enchantment.Name))
	}
	b.WriteString("    Custom(u32),\n}\n\nimpl Enchantment {\n    pub const fn id(self) -> u32 {\n        match self {\n")
	for _, enchantment := range items.Enchantments {
		fmt.Fprintf(&b, "            Self::%s => %d,\n", title(enchantment.Name), enchantment.ID)
	}
	b.WriteString("            Self::Custom(id) => id,\n        }\n    }\n\n    pub const fn max_level(self) -> Option<u32> {\n        match self {\n")
	for _, enchantment := range items.Enchantments {
		fmt.Fprintf(&b, "            Self::%s => Some(%d),\n", title(enchantment.Name), enchantment.MaxLevel)
	}
	b.WriteString("            Self::Custom(_) => None,\n        }\n    }\n\n    pub const fn from_id(id: u32) -> Self {\n        match id {\n")
	for _, enchantment := range items.Enchantments {
		fmt.Fprintf(&b, "            %d => Self::%s,\n", enchantment.ID, title(enchantment.Name))
	}
	b.WriteString("            _ => Self::Custom(id),\n        }\n    }\n}\n\n")
	b.WriteString("#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]\npub enum Potion {\n")
	for _, potion := range items.Potions {
		fmt.Fprintf(&b, "    %s,\n", title(potion.Name))
	}
	b.WriteString("    Custom(u8),\n}\n\nimpl Potion {\n    pub const fn id(self) -> u8 {\n        match self {\n")
	for _, potion := range items.Potions {
		fmt.Fprintf(&b, "            Self::%s => %d,\n", title(potion.Name), potion.ID)
	}
	b.WriteString("            Self::Custom(id) => id,\n        }\n    }\n\n    pub const fn from_id(id: u8) -> Self {\n        match id {\n")
	for _, potion := range items.Potions {
		fmt.Fprintf(&b, "            %d => Self::%s,\n", potion.ID, title(potion.Name))
	}
	b.WriteString("            _ => Self::Custom(id),\n        }\n    }\n}\n")
	return b.Bytes()
}

func compositePlayerOperations(player playerSchema) []playerOperation {
	operations := make([]playerOperation, 0, len(player.Operations))
	for _, operation := range player.Operations {
		if len(operation.Args) != 0 {
			operations = append(operations, operation)
		}
	}
	return operations
}

func cPlayerOperationArguments(operation playerOperation) string {
	var b strings.Builder
	for _, argument := range operation.Args {
		fmt.Fprintf(&b, ", %s %s", cType(argument.Type), argument.Name)
	}
	return b.String()
}

func rustPlayerOperationArguments(operation playerOperation) string {
	var b strings.Builder
	for _, argument := range operation.Args {
		fmt.Fprintf(&b, ", %s: %s", argument.Name, argument.Type)
	}
	return b.String()
}

func goPlayerOperationArguments(operation playerOperation) string {
	var b strings.Builder
	for _, argument := range operation.Args {
		fmt.Fprintf(&b, ", %s", goType(argument.Type))
	}
	return b.String()
}

func goCPlayerOperationArguments(operation playerOperation) string {
	var b strings.Builder
	for _, argument := range operation.Args {
		fmt.Fprintf(&b, ", %s C.%s", argument.Name, goCType(argument.Type))
	}
	return b.String()
}

func playerOperationValues(operation playerOperation) string {
	var b strings.Builder
	for _, argument := range operation.Args {
		fmt.Fprintf(&b, ", %s", argument.Name)
	}
	return b.String()
}

func goPlayerOperationValues(operation playerOperation) string {
	var b strings.Builder
	for _, argument := range operation.Args {
		fmt.Fprintf(&b, ", %sValue", argument.Name)
	}
	return b.String()
}

func rustPlayerOperationValidations(operation playerOperation) string {
	validations := make([]string, 0, len(operation.Args)*2)
	for _, argument := range operation.Args {
		if argument.Type == "f64" {
			validations = append(validations, "!"+argument.Name+".is_finite()")
		}
		switch argument.Validate {
		case "non_negative":
			validations = append(validations, argument.Name+" < 0")
			if argument.Type == "f64" {
				validations[len(validations)-1] += ".0"
			}
		case "positive":
			validations = append(validations, argument.Name+" <= 0.0")
		case "unit_interval":
			validations = append(validations, "!(0.0..=1.0).contains(&"+argument.Name+")")
		}
	}
	return strings.Join(validations, " || ")
}

func goPlayerOperationValidations(operation playerOperation) string {
	validations := make([]string, 0, len(operation.Args)*2)
	for _, argument := range operation.Args {
		value := argument.Name + "Value"
		if argument.Type == "f64" {
			validations = append(validations, "math.IsNaN("+value+")", "math.IsInf("+value+", 0)")
		}
		switch argument.Validate {
		case "non_negative":
			validations = append(validations, value+" < 0")
		case "positive":
			validations = append(validations, value+" <= 0")
		case "unit_interval":
			validations = append(validations, value+" < 0", value+" > 1")
		}
	}
	return strings.Join(validations, " || ")
}

func goType(t string) string {
	return map[string]string{"f64": "float64", "i32": "int32"}[t]
}

func goCType(t string) string {
	return map[string]string{"f64": "double", "i32": "int32_t"}[t]
}

func cType(t string) string {
	return map[string]string{
		"bool": "uint8_t", "player_id": "DfPlayerId", "entity_id": "DfEntityId", "world_id": "DfWorldId", "rotation": "DfRotation",
		"string_buffer": "DfStringBuffer", "string_view": "DfStringView", "vec3": "DfVec3",
		"f64": "double", "u64": "uint64_t",
		"i32": "int32_t", "block_pos": "DfBlockPos", "item_stack": "DfItemStackSnapshot",
		"damage_source": "DfDamageSourceView", "healing_source": "DfHealingSourceView",
	}[t]
}

func rustType(t string) string {
	return map[string]string{
		"bool": "u8", "player_id": "DfPlayerId", "entity_id": "DfEntityId", "world_id": "DfWorldId", "rotation": "DfRotation",
		"string_buffer": "DfStringBuffer", "string_view": "DfStringView", "vec3": "DfVec3",
		"f64": "f64", "u64": "u64",
		"i32": "i32", "block_pos": "DfBlockPos", "item_stack": "DfItemStackSnapshot",
		"damage_source": "DfDamageSourceView", "healing_source": "DfHealingSourceView",
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
