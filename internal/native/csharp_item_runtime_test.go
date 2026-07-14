package native

import (
	"reflect"
	"slices"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

type csharpItemHost struct {
	*recordingHost
	inventory  ItemStack
	armour     ItemStack
	held       [2]ItemStack
	sets       []ItemStack
	armourSets []ItemStack
	adds       []ItemStack
	heldSets   [][2]ItemStack
	failWrites bool
}

func (h *csharpItemHost) InventorySize(_ InvocationID, inventory InventoryID) (uint32, bool) {
	if inventory.Kind == InventoryArmour {
		return 4, true
	}
	return 36, inventory.Kind == InventoryMain
}

func (h *csharpItemHost) InventoryItem(_ InvocationID, inventory InventoryID, slot uint32) (ItemStack, bool) {
	if inventory.Kind != InventoryMain || slot != 0 {
		if inventory.Kind == InventoryArmour && slot == 0 {
			return h.armour, true
		}
		return ItemStack{}, false
	}
	return h.inventory, true
}

func (h *csharpItemHost) SetInventoryItem(_ InvocationID, inventory InventoryID, slot uint32, stack ItemStack) bool {
	if h.failWrites {
		return false
	}
	if inventory.Kind != InventoryMain || slot != 0 {
		if inventory.Kind == InventoryArmour && slot == 0 {
			h.armourSets = append(h.armourSets, cloneNativeItem(stack))
			h.armour = cloneNativeItem(stack)
			return true
		}
		return false
	}
	h.sets = append(h.sets, cloneNativeItem(stack))
	h.inventory = cloneNativeItem(stack)
	return true
}

func (h *csharpItemHost) AddInventoryItem(_ InvocationID, inventory InventoryID, stack ItemStack) (uint32, bool) {
	if h.failWrites {
		return 0, false
	}
	if inventory.Kind != InventoryMain {
		return 0, false
	}
	h.adds = append(h.adds, cloneNativeItem(stack))
	return stack.Count, true
}

func (h *csharpItemHost) HeldItem(_ InvocationID, _ PlayerID, hand uint32) (ItemStack, bool) {
	if hand > 1 {
		return ItemStack{}, false
	}
	return h.held[hand], true
}

func (h *csharpItemHost) HeldItems(InvocationID, PlayerID) (ItemStack, ItemStack, bool) {
	return h.held[0], h.held[1], true
}

func (h *csharpItemHost) SetHeldItems(_ InvocationID, _ PlayerID, mainHand, offHand ItemStack) bool {
	if h.failWrites {
		return false
	}
	h.heldSets = append(h.heldSets, [2]ItemStack{cloneNativeItem(mainHand), cloneNativeItem(offHand)})
	h.held = [2]ItemStack{cloneNativeItem(mainHand), cloneNativeItem(offHand)}
	return true
}

func cloneNativeItem(value ItemStack) ItemStack {
	value.Lore = append([]string(nil), value.Lore...)
	value.NBT = append([]byte(nil), value.NBT...)
	value.ValuesNBT = append([]byte(nil), value.ValuesNBT...)
	value.Enchantments = append([]ItemEnchantment(nil), value.Enchantments...)
	return value
}

func TestCSharpTypedItemInventoryFlow(t *testing.T) {
	previous := ItemStack{Identifier: "minecraft:apple", Count: 3}
	mainHand := ItemStack{
		Identifier: "minecraft:shield", Count: 1, Damage: 4, Unbreakable: true,
		CustomName: "Opaque main", Lore: []string{"preserve"}, NBT: []byte{10, 0, 0},
		ValuesNBT: []byte{10, 0, 0}, Enchantments: []ItemEnchantment{{ID: 17, Level: 3}},
	}
	offHand := ItemStack{Identifier: "minecraft:filled_map", Metadata: 2, Count: 1, CustomName: "Opaque off"}
	host := &csharpItemHost{
		recordingHost: &recordingHost{},
		inventory:     previous,
		held:          [2]ItemStack{mainHand, offHand},
	}
	runtime := openCSharpRuntimeWithHost(t, host)
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	kitchen := commandNamed(t, commands, "kitchen")
	overload := -1
	for index, candidate := range kitchen.Overloads {
		if len(candidate.Parameters) == 1 && candidate.Parameters[0].Name == "item" {
			overload = index
			break
		}
	}
	if overload < 0 {
		t.Fatalf("kitchen item overload missing: %#v", kitchen.Overloads)
	}
	player := PlayerID{UUID: [16]byte{0x91}, Generation: 8}
	for iteration := 0; iteration < 70; iteration++ {
		output, err := runtime.HandleCommand(kitchen.Index, CommandInput{
			Invocation: 55, Source: "ItemTester", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
			Overload: uint64(overload), Arguments: []string{"item"},
			OnlinePlayers: []CommandPlayer{{Player: player, Name: "ItemTester"}},
		})
		if err != nil || output.Failed || output.Message != "item=Sword, tier=diamond, count=1, held=true, armour_slots=4, added_empty=0, variants=15" {
			t.Fatalf("iteration %d: output=%#v error=%v", iteration, output, err)
		}
	}
	if !reflect.DeepEqual(host.inventory, previous) || !reflect.DeepEqual(host.held, [2]ItemStack{mainHand, offHand}) {
		t.Fatalf("command did not restore items: inventory=%#v held=%#v", host.inventory, host.held)
	}
	if len(host.sets) != 1190 || len(host.heldSets) != 140 {
		t.Fatalf("item writes=%d held writes=%d", len(host.sets), len(host.heldSets))
	}
	if len(host.armourSets) != 70 || len(host.adds) != 70 || host.adds[0].Identifier != "" || host.adds[0].Count != 0 {
		t.Fatalf("armour writes=%d adds=%d", len(host.armourSets), len(host.adds))
	}
	sword := host.sets[0]
	if sword.Identifier != "minecraft:diamond_sword" || sword.Count != 1 ||
		sword.CustomName != "Kitchen sword" ||
		!slices.Equal(sword.Lore, []string{"Generated from Dragonfly", "Restored after this command"}) {
		t.Fatalf("typed sword transport=%#v", sword)
	}
	wantVariants := []ItemStack{
		{Identifier: "minecraft:arrow", Metadata: 6, Count: 1},
		{Identifier: "minecraft:creeper_banner_pattern", Count: 1},
		{Identifier: "minecraft:black_dye", Count: 1},
		{Identifier: "minecraft:goat_horn", Metadata: 7, Count: 1},
		{Identifier: "minecraft:potion", Metadata: 42, Count: 1},
		{Identifier: "minecraft:lingering_potion", Metadata: 21, Count: 1},
		{Identifier: "minecraft:splash_potion", Metadata: 23, Count: 1},
		{Identifier: "minecraft:music_disc_lava_chicken", Count: 1},
		{Identifier: "minecraft:scrape_pottery_sherd", Count: 1},
		{Identifier: "minecraft:bolt_armor_trim_smithing_template", Count: 1},
		{Identifier: "minecraft:suspicious_stew", Metadata: 12, Count: 1},
	}
	for index, want := range wantVariants {
		if got := host.sets[index+1]; !reflect.DeepEqual(got, want) {
			t.Fatalf("typed variant %d=%#v, want %#v", index, got, want)
		}
	}
	writable := host.sets[12]
	if writable.Identifier != "minecraft:writable_book" || writable.Count != 1 {
		t.Fatalf("typed writable book transport=%#v", writable)
	}
	writableNBT, ok := decodeTestItemNBT(writable.NBT)
	if !ok {
		t.Fatalf("typed writable book NBT invalid: %x", writable.NBT)
	}
	writablePages, ok := writableNBT["pages"].([]any)
	firstPage, firstOK := itemPageText(writablePages, 0)
	secondPage, secondOK := itemPageText(writablePages, 1)
	if !ok || len(writablePages) != 2 || !firstOK || !secondOK || firstPage != "beta" || secondPage != "first" {
		t.Fatalf("typed writable book pages=%#v", writableNBT["pages"])
	}
	written := host.sets[13]
	writtenNBT, ok := decodeTestItemNBT(written.NBT)
	if !ok || written.Identifier != "minecraft:written_book" || written.Count != 1 ||
		writtenNBT["title"] != "Kitchen" || writtenNBT["author"] != "bedrock-gophers" ||
		writtenNBT["generation"] != uint8(1) {
		t.Fatalf("typed written book transport=%#v nbt=%#v", written, writtenNBT)
	}
	firework := host.sets[14]
	fireworkNBT, ok := decodeTestItemNBT(firework.NBT)
	fireworks, fireworksOK := fireworkNBT["Fireworks"].(map[string]any)
	explosions, explosionsOK := fireworks["Explosions"].([]any)
	var explosion map[string]any
	if explosionsOK && len(explosions) == 1 {
		explosion, explosionsOK = explosions[0].(map[string]any)
	}
	if !ok || firework.Identifier != "minecraft:firework_rocket" || firework.Metadata != 0 || firework.Count != 1 ||
		!fireworksOK || fireworks["Flight"] != uint8(2) || !explosionsOK ||
		explosion["FireworkType"] != uint8(2) || explosion["FireworkColor"] != [1]byte{0} ||
		explosion["FireworkFade"] != [1]byte{1} || explosion["FireworkFlicker"] != uint8(1) ||
		explosion["FireworkTrail"] != uint8(1) {
		t.Fatalf("typed firework transport=%#v nbt=%#v", firework, fireworkNBT)
	}
	star := host.sets[15]
	starNBT, ok := decodeTestItemNBT(star.NBT)
	starExplosion, starExplosionOK := starNBT["FireworksItem"].(map[string]any)
	if !ok || star.Identifier != "minecraft:firework_star" || star.Metadata != 6 || star.Count != 1 ||
		!starExplosionOK || starExplosion["FireworkType"] != uint8(4) ||
		starExplosion["FireworkColor"] != [1]byte{6} || starExplosion["FireworkFade"] != [0]byte{} ||
		starExplosion["FireworkFlicker"] != uint8(0) || starExplosion["FireworkTrail"] != uint8(0) ||
		starNBT["customColor"] != int32(-15295332) {
		t.Fatalf("typed firework star transport=%#v nbt=%#v", star, starNBT)
	}
	host.failWrites = true
	failed, err := runtime.HandleCommand(kitchen.Index, CommandInput{
		Invocation: 55, Source: "ItemTester", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
		Overload: uint64(overload), Arguments: []string{"item"},
		OnlinePlayers: []CommandPlayer{{Player: player, Name: "ItemTester"}},
	})
	if err == nil && !failed.Failed {
		t.Fatalf("failed inventory mutation was silently accepted: %#v", failed)
	}
}

func itemPageText(pages []any, index int) (string, bool) {
	if index < 0 || index >= len(pages) {
		return "", false
	}
	page, ok := pages[index].(map[string]any)
	if !ok {
		return "", false
	}
	text, ok := page["text"].(string)
	return text, ok
}

func decodeTestItemNBT(data []byte) (map[string]any, bool) {
	var value map[string]any
	err := nbt.UnmarshalEncoding(data, &value, nbt.LittleEndian)
	return value, err == nil
}
