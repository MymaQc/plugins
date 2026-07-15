package host

import (
	"encoding/gob"
	"math"
	"reflect"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

const (
	maxItemDataBytes      = 16 << 20
	maxNestedItemDepth    = 16
	maxNestedItemEntries  = 256
	maxNestedItemIDBytes  = 256
	maxNestedItemText     = 4096
	nestedItemVersion     = int32(1)
	nestedItemVersionName = "bedrock_gophers_version"
)

func init() {
	gob.Register(map[string]any{})
	gob.Register([]any{})
	gob.Register([]byte{})
	gob.Register([]int32{})
	gob.Register([]int64{})
}

func (p *Players) InventorySize(invocation native.InvocationID, id native.InventoryID) (uint32, bool) {
	value, ok := readPlayer(p, invocation, id.Player, func(connected *player.Player) uint32 {
		if id.Kind == native.InventoryOffhand {
			return 1
		}
		inv, valid := playerInventory(connected, id.Kind)
		if !valid {
			return 0
		}
		return uint32(inv.Size())
	})
	return value, ok && value != 0
}

func (p *Players) InventoryItem(invocation native.InvocationID, id native.InventoryID, slot uint32) (native.ItemStack, bool) {
	value, ok := readPlayer(p, invocation, id.Player, func(connected *player.Player) struct {
		value native.ItemStack
		ok    bool
	} {
		var stack item.Stack
		if id.Kind == native.InventoryOffhand {
			if slot != 0 {
				return struct {
					value native.ItemStack
					ok    bool
				}{}
			}
			_, stack = connected.HeldItems()
		} else {
			inv, valid := playerInventory(connected, id.Kind)
			if !valid {
				return struct {
					value native.ItemStack
					ok    bool
				}{}
			}
			var err error
			stack, err = inv.Item(int(slot))
			if err != nil {
				return struct {
					value native.ItemStack
					ok    bool
				}{}
			}
		}
		converted, valid := itemStackToNative(stack)
		return struct {
			value native.ItemStack
			ok    bool
		}{converted, valid}
	})
	return value.value, ok && value.ok
}

func (p *Players) SetInventoryItem(invocation native.InvocationID, id native.InventoryID, slot uint32, value native.ItemStack) bool {
	stack, ok := itemStackFromNative(value)
	if !ok {
		return false
	}
	changed := false
	ok = p.mutatePlayer(invocation, id.Player, func(connected *player.Player) {
		if id.Kind == native.InventoryOffhand {
			if slot != 0 {
				return
			}
			main, _ := connected.HeldItems()
			connected.SetHeldItems(main, stack)
			changed = true
			return
		}
		inv, valid := playerInventory(connected, id.Kind)
		changed = valid && inv.SetItem(int(slot), stack) == nil
	})
	return ok && changed
}

func (p *Players) AddInventoryItem(invocation native.InvocationID, id native.InventoryID, value native.ItemStack) (uint32, bool) {
	stack, ok := itemStackFromNative(value)
	if !ok {
		return 0, false
	}
	type addResult struct {
		added uint32
		valid bool
	}
	result := addResult{}
	ok = p.mutatePlayer(invocation, id.Player, func(connected *player.Player) {
		if id.Kind == native.InventoryOffhand {
			main, offhand := connected.HeldItems()
			temporary := inventory.New(1, nil)
			_ = temporary.SetItem(0, offhand)
			count, _ := temporary.AddItem(stack)
			offhand, _ = temporary.Item(0)
			connected.SetHeldItems(main, offhand)
			result = addResult{uint32(count), true}
			return
		}
		inv, valid := playerInventory(connected, id.Kind)
		if !valid {
			return
		}
		count, _ := inv.AddItem(stack)
		result = addResult{uint32(count), true}
	})
	return result.added, ok && result.valid
}

func (p *Players) ClearInventory(invocation native.InvocationID, id native.InventoryID) bool {
	cleared := false
	ok := p.mutatePlayer(invocation, id.Player, func(connected *player.Player) {
		if id.Kind == native.InventoryOffhand {
			main, _ := connected.HeldItems()
			connected.SetHeldItems(main, item.Stack{})
			cleared = true
			return
		}
		if inv, valid := playerInventory(connected, id.Kind); valid {
			inv.Clear()
			cleared = true
		}
	})
	return ok && cleared
}

func (p *Players) HeldItem(invocation native.InvocationID, id native.PlayerID, hand uint32) (native.ItemStack, bool) {
	if hand > 1 {
		return native.ItemStack{}, false
	}
	value, ok := readPlayer(p, invocation, id, func(connected *player.Player) struct {
		value native.ItemStack
		ok    bool
	} {
		main, offhand := connected.HeldItems()
		if hand == 1 {
			main = offhand
		}
		converted, valid := itemStackToNative(main)
		return struct {
			value native.ItemStack
			ok    bool
		}{converted, valid}
	})
	return value.value, ok && value.ok
}

func (p *Players) HeldItems(invocation native.InvocationID, id native.PlayerID) (native.ItemStack, native.ItemStack, bool) {
	value, ok := readPlayer(p, invocation, id, func(connected *player.Player) struct {
		main, offhand native.ItemStack
		ok            bool
	} {
		main, offhand := connected.HeldItems()
		mainValue, mainOK := itemStackToNative(main)
		offhandValue, offhandOK := itemStackToNative(offhand)
		return struct {
			main, offhand native.ItemStack
			ok            bool
		}{main: mainValue, offhand: offhandValue, ok: mainOK && offhandOK}
	})
	return value.main, value.offhand, ok && value.ok
}

func (p *Players) SetHeldItems(invocation native.InvocationID, id native.PlayerID, mainValue, offhandValue native.ItemStack) bool {
	main, ok := itemStackFromNative(mainValue)
	if !ok {
		return false
	}
	offhand, ok := itemStackFromNative(offhandValue)
	if !ok {
		return false
	}
	return p.mutatePlayer(invocation, id, func(connected *player.Player) { connected.SetHeldItems(main, offhand) })
}

func (p *Players) SetHeldSlot(invocation native.InvocationID, id native.PlayerID, slot uint32) bool {
	if slot > 8 {
		return false
	}
	return p.mutatePlayer(invocation, id, func(connected *player.Player) { _ = connected.SetHeldSlot(int(slot)) })
}

func playerInventory(connected *player.Player, kind native.InventoryKind) (*inventory.Inventory, bool) {
	switch kind {
	case native.InventoryMain:
		return connected.Inventory(), true
	case native.InventoryArmour:
		return connected.Armour().Inventory(), true
	case native.InventoryEnderChest:
		return connected.EnderChestInventory(), true
	default:
		return nil, false
	}
}

func itemStackToNative(stack item.Stack) (value native.ItemStack, ok bool) {
	return itemStackToNativeDepth(stack, 0)
}

func itemStackToNativeDepth(stack item.Stack, depth int) (value native.ItemStack, ok bool) {
	defer func() {
		if recover() != nil {
			value, ok = native.ItemStack{}, false
		}
	}()
	if depth > maxNestedItemDepth {
		return native.ItemStack{}, false
	}
	if stack.Empty() {
		return native.ItemStack{}, true
	}
	identifier, metadata := stack.Item().EncodeItem()
	anvilCost := stack.AnvilCost()
	if anvilCost < math.MinInt32 || anvilCost > math.MaxInt32 {
		return native.ItemStack{}, false
	}
	value = native.ItemStack{
		Identifier: identifier, Metadata: int32(metadata), Count: uint32(stack.Count()),
		Unbreakable: stack.Unbreakable(), AnvilCost: int32(anvilCost),
		CustomName: stack.CustomName(), Lore: append([]string(nil), stack.Lore()...),
	}
	if maximum := stack.MaxDurability(); maximum >= 0 {
		value.Damage = uint32(maximum - stack.Durability())
	}
	if crossbow, crossbowOK := stack.Item().(item.Crossbow); crossbowOK {
		value.NBT, ok = marshalCrossbowNBT(crossbow, depth)
		if !ok {
			return native.ItemStack{}, false
		}
	} else if itemNBT, itemNBTOK := stack.Item().(world.NBTer); itemNBTOK {
		value.NBT, ok = marshalItemNBT(itemNBT.EncodeNBT())
		if !ok {
			return native.ItemStack{}, false
		}
	}
	if values := stack.Values(); len(values) != 0 {
		value.ValuesNBT, ok = marshalItemNBT(values)
		if !ok {
			return native.ItemStack{}, false
		}
	}
	for _, enchantment := range stack.Enchantments() {
		id, found := item.EnchantmentID(enchantment.Type())
		if !found || id < 0 || enchantment.Level() < 1 {
			return native.ItemStack{}, false
		}
		value.Enchantments = append(value.Enchantments, native.ItemEnchantment{ID: uint32(id), Level: uint32(enchantment.Level())})
	}
	return value, itemStackDataSize(value) <= maxItemDataBytes
}

func itemStackFromNative(value native.ItemStack) (stack item.Stack, ok bool) {
	return itemStackFromNativeDepth(value, 0)
}

func itemStackFromNativeDepth(value native.ItemStack, depth int) (stack item.Stack, ok bool) {
	defer func() {
		if recover() != nil {
			stack, ok = item.Stack{}, false
		}
	}()
	if depth > maxNestedItemDepth {
		return item.Stack{}, false
	}
	if value.Count == 0 {
		return item.Stack{}, true
	}
	if value.Identifier == "" || value.Metadata < math.MinInt16 || value.Metadata > math.MaxInt16 || itemStackDataSize(value) > maxItemDataBytes {
		return item.Stack{}, false
	}
	base, ok := world.ItemByName(value.Identifier, int16(value.Metadata))
	if !ok {
		return item.Stack{}, false
	}
	if crossbow, crossbowOK := base.(item.Crossbow); crossbowOK {
		if len(value.NBT) != 0 {
			charged, found, valid := unmarshalCrossbowNBT(value.NBT)
			if !valid {
				return item.Stack{}, false
			}
			if found {
				crossbow.Item, ok = itemStackFromNativeDepth(charged, depth+1)
				if !ok {
					return item.Stack{}, false
				}
			}
		}
		base = crossbow
	} else if len(value.NBT) != 0 {
		decoded, ok := unmarshalItemNBT(value.NBT)
		if !ok {
			return item.Stack{}, false
		}
		itemNBT, ok := base.(world.NBTer)
		if !ok {
			return item.Stack{}, false
		}
		base, ok = itemNBT.DecodeNBT(decoded).(world.Item)
		if !ok {
			return item.Stack{}, false
		}
	}
	stack = item.NewStack(base, int(value.Count)).Damage(int(value.Damage))
	if value.Unbreakable {
		stack = stack.AsUnbreakable()
	}
	stack = stack.WithAnvilCost(int(value.AnvilCost))
	if value.CustomName != "" {
		stack = stack.WithCustomName(value.CustomName)
	}
	stack = stack.WithLore(value.Lore...)
	if len(value.ValuesNBT) != 0 {
		values, ok := unmarshalItemNBT(value.ValuesNBT)
		if !ok {
			return item.Stack{}, false
		}
		values, ok = normalizeItemNBTMap(values)
		if !ok {
			return item.Stack{}, false
		}
		for key, entry := range values {
			stack = stack.WithValue(key, entry)
		}
	}
	for _, enchantment := range value.Enchantments {
		if enchantment.ID > math.MaxInt32 || enchantment.Level == 0 || enchantment.Level > math.MaxInt16 {
			return item.Stack{}, false
		}
		typeValue, found := item.EnchantmentByID(int(enchantment.ID))
		if !found {
			return item.Stack{}, false
		}
		stack = stack.WithForcedEnchantments(item.NewEnchantment(typeValue, int(enchantment.Level)))
	}
	return stack, true
}

func marshalCrossbowNBT(crossbow item.Crossbow, depth int) ([]byte, bool) {
	if crossbow.Item.Empty() {
		return nil, true
	}
	charged, ok := itemStackToNativeDepth(crossbow.Item, depth+1)
	if !ok {
		return nil, false
	}
	compound, ok := nestedItemCompound(charged)
	if !ok {
		return nil, false
	}
	return marshalItemNBT(map[string]any{"chargedItem": compound})
}

func unmarshalCrossbowNBT(data []byte) (native.ItemStack, bool, bool) {
	root, ok := unmarshalItemNBT(data)
	if !ok {
		return native.ItemStack{}, false, false
	}
	charged, found := root["chargedItem"]
	if !found {
		return native.ItemStack{}, false, true
	}
	compound, ok := charged.(map[string]any)
	if !ok {
		return native.ItemStack{}, false, true
	}
	value, ok := nestedItemFromCompound(compound)
	return value, ok, ok
}

func nestedItemCompound(value native.ItemStack) (map[string]any, bool) {
	if !validNestedItemStack(value) || value.Count > math.MaxInt32 || value.Damage > math.MaxInt32 {
		return nil, false
	}
	enchantments := make([]any, len(value.Enchantments))
	for index, enchantment := range value.Enchantments {
		if enchantment.ID > math.MaxInt32 || enchantment.Level > math.MaxInt32 {
			return nil, false
		}
		enchantments[index] = map[string]any{"id": int32(enchantment.ID), "level": int32(enchantment.Level)}
	}
	return map[string]any{
		nestedItemVersionName: nestedItemVersion,
		"identifier":          value.Identifier,
		"metadata":            value.Metadata,
		"count":               int32(value.Count),
		"damage":              int32(value.Damage),
		"unbreakable":         boolByte(value.Unbreakable),
		"anvilCost":           value.AnvilCost,
		"customName":          value.CustomName,
		"lore":                append([]string(nil), value.Lore...),
		"itemNbt":             fixedByteArray(value.NBT),
		"valuesNbt":           fixedByteArray(value.ValuesNBT),
		"enchantments":        enchantments,
	}, true
}

func nestedItemFromCompound(value map[string]any) (native.ItemStack, bool) {
	version, versionOK := value[nestedItemVersionName].(int32)
	identifier, identifierOK := value["identifier"].(string)
	metadata, metadataOK := value["metadata"].(int32)
	count, countOK := value["count"].(int32)
	damage, damageOK := value["damage"].(int32)
	unbreakable, unbreakableOK := value["unbreakable"].(uint8)
	anvilCost, anvilCostOK := value["anvilCost"].(int32)
	customName, customNameOK := value["customName"].(string)
	if !versionOK || version != nestedItemVersion || !identifierOK || identifier == "" || !metadataOK ||
		!countOK || count <= 0 || !damageOK || damage < 0 || !unbreakableOK || unbreakable > 1 ||
		!anvilCostOK || !customNameOK {
		return native.ItemStack{}, false
	}
	lore, ok := stringList(value["lore"])
	if !ok {
		return native.ItemStack{}, false
	}
	itemNBT, ok := byteArray(value["itemNbt"])
	if !ok {
		return native.ItemStack{}, false
	}
	valuesNBT, ok := byteArray(value["valuesNbt"])
	if !ok {
		return native.ItemStack{}, false
	}
	encodedEnchantments, ok := value["enchantments"].([]any)
	if !ok || len(encodedEnchantments) > maxNestedItemEntries {
		return native.ItemStack{}, false
	}
	enchantments := make([]native.ItemEnchantment, len(encodedEnchantments))
	for index, encoded := range encodedEnchantments {
		compound, ok := encoded.(map[string]any)
		if !ok {
			return native.ItemStack{}, false
		}
		id, idOK := compound["id"].(int32)
		level, levelOK := compound["level"].(int32)
		if !idOK || id < 0 || !levelOK || level <= 0 {
			return native.ItemStack{}, false
		}
		enchantments[index] = native.ItemEnchantment{ID: uint32(id), Level: uint32(level)}
	}
	result := native.ItemStack{
		Identifier: identifier, Metadata: metadata, Count: uint32(count), Damage: uint32(damage),
		Unbreakable: unbreakable == 1, AnvilCost: anvilCost, CustomName: customName,
		Lore: lore, NBT: itemNBT, ValuesNBT: valuesNBT, Enchantments: enchantments,
	}
	return result, validNestedItemStack(result)
}

func validNestedItemStack(value native.ItemStack) bool {
	if value.Count == 0 || len(value.Identifier) == 0 || len(value.Identifier) > maxNestedItemIDBytes ||
		len(value.CustomName) > maxNestedItemText || len(value.Lore) > maxNestedItemEntries ||
		len(value.Enchantments) > maxNestedItemEntries || itemStackDataSize(value) > maxItemDataBytes {
		return false
	}
	for _, line := range value.Lore {
		if len(line) > maxNestedItemText {
			return false
		}
	}
	for _, enchantment := range value.Enchantments {
		if enchantment.Level == 0 {
			return false
		}
	}
	return true
}

func fixedByteArray(value []byte) any {
	array := reflect.New(reflect.ArrayOf(len(value), reflect.TypeFor[byte]())).Elem()
	reflect.Copy(array, reflect.ValueOf(value))
	return array.Interface()
}

func boolByte(value bool) uint8 {
	if value {
		return 1
	}
	return 0
}

func stringList(value any) ([]string, bool) {
	switch list := value.(type) {
	case []string:
		if len(list) > maxNestedItemEntries {
			return nil, false
		}
		for _, entry := range list {
			if len(entry) > maxNestedItemText {
				return nil, false
			}
		}
		return append([]string(nil), list...), true
	case []any:
		if len(list) > maxNestedItemEntries {
			return nil, false
		}
		result := make([]string, len(list))
		for index, entry := range list {
			var ok bool
			result[index], ok = entry.(string)
			if !ok {
				return nil, false
			}
			if len(result[index]) > maxNestedItemText {
				return nil, false
			}
		}
		return result, true
	default:
		return nil, false
	}
}

func byteArray(value any) ([]byte, bool) {
	if list, ok := value.([]any); ok && len(list) == 0 {
		return []byte{}, true
	}
	reflected := reflect.ValueOf(value)
	if !reflected.IsValid() || (reflected.Kind() != reflect.Array && reflected.Kind() != reflect.Slice) || reflected.Type().Elem().Kind() != reflect.Uint8 {
		return nil, false
	}
	result := make([]byte, reflected.Len())
	for index := range result {
		result[index] = byte(reflected.Index(index).Uint())
	}
	return result, true
}

func normalizeItemNBTMap(value map[string]any) (map[string]any, bool) {
	result := make(map[string]any, len(value))
	for key, entry := range value {
		normalized, ok := normalizeItemNBTValue(entry)
		if !ok {
			return nil, false
		}
		result[key] = normalized
	}
	return result, true
}

func normalizeItemNBTValue(value any) (any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return normalizeItemNBTMap(typed)
	case []any:
		result := make([]any, len(typed))
		for index, entry := range typed {
			normalized, ok := normalizeItemNBTValue(entry)
			if !ok {
				return nil, false
			}
			result[index] = normalized
		}
		return result, true
	}
	reflected := reflect.ValueOf(value)
	if !reflected.IsValid() {
		return nil, false
	}
	if (reflected.Kind() == reflect.Array || reflected.Kind() == reflect.Slice) && reflected.Type().Elem().Kind() == reflect.Uint8 {
		result := make([]byte, reflected.Len())
		for index := range result {
			result[index] = byte(reflected.Index(index).Uint())
		}
		return result, true
	}
	return value, true
}

// ItemStackFromNative decodes a plugin item stack using Dragonfly registries.
func ItemStackFromNative(value native.ItemStack) (item.Stack, bool) {
	return itemStackFromNative(value)
}

func marshalItemNBT(value map[string]any) ([]byte, bool) {
	encoded, err := nbt.MarshalEncoding(value, nbt.LittleEndian)
	return encoded, err == nil && len(encoded) <= maxItemDataBytes
}

// MarshalNBT encodes a compound using Dragonfly's little-endian NBT transport.
func MarshalNBT(value map[string]any) ([]byte, bool) { return marshalItemNBT(value) }

func unmarshalItemNBT(data []byte) (map[string]any, bool) {
	if len(data) > maxItemDataBytes {
		return nil, false
	}
	var value map[string]any
	err := nbt.UnmarshalEncoding(data, &value, nbt.LittleEndian)
	return value, err == nil
}

// UnmarshalNBT decodes a Dragonfly little-endian NBT compound.
func UnmarshalNBT(data []byte) (map[string]any, bool) { return unmarshalItemNBT(data) }

func itemStackDataSize(value native.ItemStack) int {
	total := len(value.Identifier) + len(value.CustomName) + len(value.NBT) + len(value.ValuesNBT)
	for _, line := range value.Lore {
		total += len(line)
		if total > maxItemDataBytes {
			return total
		}
	}
	return total
}
