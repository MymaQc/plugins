package host

import (
	"math"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

const maxItemDataBytes = 16 << 20

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
	if id.Kind == native.InventoryOffhand && slot != 0 {
		return false
	}
	return p.mutatePlayer(invocation, id.Player, func(connected *player.Player) {
		if id.Kind == native.InventoryOffhand {
			main, _ := connected.HeldItems()
			connected.SetHeldItems(main, stack)
			return
		}
		if inv, valid := playerInventory(connected, id.Kind); valid {
			_ = inv.SetItem(int(slot), stack)
		}
	})
}

func (p *Players) AddInventoryItem(invocation native.InvocationID, id native.InventoryID, value native.ItemStack) (uint32, bool) {
	stack, ok := itemStackFromNative(value)
	if !ok {
		return 0, false
	}
	return readPlayer(p, invocation, id.Player, func(connected *player.Player) uint32 {
		if id.Kind == native.InventoryOffhand {
			main, offhand := connected.HeldItems()
			temporary := inventory.New(1, nil)
			_ = temporary.SetItem(0, offhand)
			added, _ := temporary.AddItem(stack)
			offhand, _ = temporary.Item(0)
			connected.SetHeldItems(main, offhand)
			return uint32(added)
		}
		inv, valid := playerInventory(connected, id.Kind)
		if !valid {
			return 0
		}
		added, _ := inv.AddItem(stack)
		return uint32(added)
	})
}

func (p *Players) ClearInventory(invocation native.InvocationID, id native.InventoryID) bool {
	return p.mutatePlayer(invocation, id.Player, func(connected *player.Player) {
		if id.Kind == native.InventoryOffhand {
			main, _ := connected.HeldItems()
			connected.SetHeldItems(main, item.Stack{})
			return
		}
		if inv, valid := playerInventory(connected, id.Kind); valid {
			inv.Clear()
		}
	})
}

func (p *Players) HeldItem(invocation native.InvocationID, id native.PlayerID, hand uint32) (native.ItemStack, bool) {
	if hand > 1 {
		return native.ItemStack{}, false
	}
	value, ok := readPlayer(p, invocation, id, func(connected *player.Player) struct {
		value native.ItemStack
		ok    bool
	} { main, offhand := connected.HeldItems(); if hand == 1 {
		main = offhand
	}; converted, valid := itemStackToNative(main); return struct {
		value native.ItemStack
		ok    bool
	}{converted, valid} })
	return value.value, ok && value.ok
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
	default:
		return nil, false
	}
}

func itemStackToNative(stack item.Stack) (value native.ItemStack, ok bool) {
	defer func() {
		if recover() != nil {
			value, ok = native.ItemStack{}, false
		}
	}()
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
	if itemNBT, ok := stack.Item().(world.NBTer); ok {
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
	defer func() {
		if recover() != nil {
			stack, ok = item.Stack{}, false
		}
	}()
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
	if len(value.NBT) != 0 {
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
