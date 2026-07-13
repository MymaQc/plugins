package native

/*
#include "bridge.h"
*/
import "C"

import (
	"unicode/utf8"
	"unsafe"
)

const (
	maxItemDataBytes    = 16 << 20
	maxItemLore         = 256
	maxItemEnchantments = 256
	maxItemIdentifier   = 256
	maxItemText         = 4096
)

//export bg_go_inventory_size
func bg_go_inventory_size(context C.uint64_t, invocation C.DfInvocationId, inventory C.DfInventoryId, size *C.uint32_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || size == nil {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.InventorySize(InvocationID(invocation), inventoryID(inventory))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*size = C.uint32_t(value)
	return C.DF_STATUS_OK
}

//export bg_go_inventory_item_open
func bg_go_inventory_item_open(context C.uint64_t, invocation C.DfInvocationId, inventory C.DfInventoryId, slot C.uint32_t, snapshot *C.uint64_t, info *C.DfItemStackInfo) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.InventoryItem(InvocationID(invocation), inventoryID(inventory), uint32(slot))
	return openItemSnapshot(uint64(context), value, ok, snapshot, info)
}

//export bg_go_player_held_item_open
func bg_go_player_held_item_open(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, hand C.uint32_t, snapshot *C.uint64_t, info *C.DfItemStackInfo) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.HeldItem(InvocationID(invocation), playerID(player), uint32(hand))
	return openItemSnapshot(uint64(context), value, ok, snapshot, info)
}

//export bg_go_item_stack_read
func bg_go_item_stack_read(context C.uint64_t, invocation C.DfInvocationId, snapshot C.uint64_t, data *C.DfItemStackData) C.DfStatus {
	value, ok := resolveItemSnapshot(uint64(context), uint64(snapshot))
	if !ok || data == nil || !validNativeItem(value) ||
		uint64(data.lore_capacity) < uint64(len(value.Lore)) ||
		uint64(data.enchantment_capacity) < uint64(len(value.Enchantments)) {
		return C.DF_STATUS_ERROR
	}
	loreBytes := make([]byte, 0)
	for _, line := range value.Lore {
		loreBytes = append(loreBytes, line...)
	}
	if !canWriteSkinBuffer(&data.identifier, []byte(value.Identifier)) ||
		!canWriteSkinBuffer(&data.custom_name, []byte(value.CustomName)) ||
		!canWriteSkinBuffer(&data.lore_bytes, loreBytes) ||
		!canWriteSkinBuffer(&data.nbt, value.NBT) ||
		!canWriteSkinBuffer(&data.values_nbt, value.ValuesNBT) ||
		len(value.Lore) != 0 && data.lore == nil ||
		len(value.Enchantments) != 0 && data.enchantments == nil {
		return C.DF_STATUS_ERROR
	}
	writeSkinBuffer(&data.identifier, []byte(value.Identifier))
	writeSkinBuffer(&data.custom_name, []byte(value.CustomName))
	writeSkinBuffer(&data.lore_bytes, loreBytes)
	writeSkinBuffer(&data.nbt, value.NBT)
	writeSkinBuffer(&data.values_nbt, value.ValuesNBT)
	offset := uint64(0)
	if len(value.Lore) != 0 {
		for index, line := range value.Lore {
			unsafe.Slice(data.lore, len(value.Lore))[index] = C.DfByteSpan{offset: C.uint64_t(offset), len: C.uint64_t(len(line))}
			offset += uint64(len(line))
		}
	}
	if len(value.Enchantments) != 0 {
		output := unsafe.Slice(data.enchantments, len(value.Enchantments))
		for index, enchantment := range value.Enchantments {
			output[index] = C.DfItemEnchantment{id: C.uint32_t(enchantment.ID), level: C.uint32_t(enchantment.Level)}
		}
	}
	return C.DF_STATUS_OK
}

//export bg_go_item_stack_close
func bg_go_item_stack_close(context C.uint64_t, invocation C.DfInvocationId, snapshot C.uint64_t) {
	unregisterItemSnapshot(uint64(context), uint64(snapshot))
}

//export bg_go_inventory_item_set
func bg_go_inventory_item_set(context C.uint64_t, invocation C.DfInvocationId, inventory C.DfInventoryId, slot C.uint32_t, view *C.DfItemStackViewV3) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	value, valid := copyItemStackView(view)
	if !ok || !valid || !host.SetInventoryItem(InvocationID(invocation), inventoryID(inventory), uint32(slot), value) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_inventory_item_add
func bg_go_inventory_item_add(context C.uint64_t, invocation C.DfInvocationId, inventory C.DfInventoryId, view *C.DfItemStackViewV3, added *C.uint32_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	value, valid := copyItemStackView(view)
	if !ok || added == nil || !valid {
		return C.DF_STATUS_ERROR
	}
	count, ok := host.AddInventoryItem(InvocationID(invocation), inventoryID(inventory), value)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*added = C.uint32_t(count)
	return C.DF_STATUS_OK
}

//export bg_go_inventory_clear_slot
func bg_go_inventory_clear_slot(context C.uint64_t, invocation C.DfInvocationId, inventory C.DfInventoryId, slot C.uint32_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SetInventoryItem(InvocationID(invocation), inventoryID(inventory), uint32(slot), ItemStack{}) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_inventory_clear
func bg_go_inventory_clear(context C.uint64_t, invocation C.DfInvocationId, inventory C.DfInventoryId) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.ClearInventory(InvocationID(invocation), inventoryID(inventory)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_held_items_set
func bg_go_player_held_items_set(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, mainView, offhandView *C.DfItemStackViewV3) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	main, mainOK := copyItemStackView(mainView)
	offhand, offhandOK := copyItemStackView(offhandView)
	if !ok || !mainOK || !offhandOK || !host.SetHeldItems(InvocationID(invocation), playerID(player), main, offhand) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_held_slot_set
func bg_go_player_held_slot_set(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, slot C.uint32_t) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok || !host.SetHeldSlot(InvocationID(invocation), playerID(player), uint32(slot)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

func openItemSnapshot(context uint64, value ItemStack, ok bool, snapshot *C.uint64_t, info *C.DfItemStackInfo) C.DfStatus {
	if !ok || snapshot == nil || info == nil || !validNativeItem(value) {
		return C.DF_STATUS_ERROR
	}
	id, ok := registerItemSnapshot(context, value)
	if !ok {
		return C.DF_STATUS_ERROR
	}
	*snapshot = C.uint64_t(id)
	fillItemStackInfo(info, value)
	return C.DF_STATUS_OK
}

func fillItemStackInfo(info *C.DfItemStackInfo, value ItemStack) {
	loreBytes := 0
	for _, line := range value.Lore {
		loreBytes += len(line)
	}
	*info = C.DfItemStackInfo{
		metadata: C.int32_t(value.Metadata), count: C.uint32_t(value.Count), damage: C.uint32_t(value.Damage),
		unbreakable: C.uint8_t(boolByte(value.Unbreakable)), anvil_cost: C.int32_t(value.AnvilCost),
		identifier_len: C.uint64_t(len(value.Identifier)), custom_name_len: C.uint64_t(len(value.CustomName)),
		lore_bytes_len: C.uint64_t(loreBytes), lore_count: C.uint64_t(len(value.Lore)),
		nbt_len: C.uint64_t(len(value.NBT)), values_nbt_len: C.uint64_t(len(value.ValuesNBT)),
		enchantment_count: C.uint64_t(len(value.Enchantments)),
	}
}

func copyItemStackView(view *C.DfItemStackViewV3) (ItemStack, bool) {
	if view == nil || uint64(view.lore_count) > maxItemLore || uint64(view.enchantment_count) > maxItemEnchantments ||
		view.lore_count != 0 && view.lore == nil || view.enchantment_count != 0 && view.enchantments == nil {
		return ItemStack{}, false
	}
	identifier, ok := copySkinView(view.identifier)
	if !ok {
		return ItemStack{}, false
	}
	customName, ok := copySkinView(view.custom_name)
	if !ok {
		return ItemStack{}, false
	}
	nbtData, ok := copySkinView(view.nbt)
	if !ok {
		return ItemStack{}, false
	}
	valuesNBT, ok := copySkinView(view.values_nbt)
	if !ok {
		return ItemStack{}, false
	}
	value := ItemStack{
		Identifier: string(identifier), Metadata: int32(view.metadata), Count: uint32(view.count), Damage: uint32(view.damage),
		Unbreakable: view.unbreakable != 0, AnvilCost: int32(view.anvil_cost),
		CustomName: string(customName), NBT: nbtData, ValuesNBT: valuesNBT,
	}
	if view.lore_count != 0 {
		for _, lineView := range unsafe.Slice(view.lore, int(view.lore_count)) {
			line, ok := copySkinView(lineView)
			if !ok {
				return ItemStack{}, false
			}
			value.Lore = append(value.Lore, string(line))
		}
	}
	if view.enchantment_count != 0 {
		for _, enchantment := range unsafe.Slice(view.enchantments, int(view.enchantment_count)) {
			value.Enchantments = append(value.Enchantments, ItemEnchantment{ID: uint32(enchantment.id), Level: uint32(enchantment.level)})
		}
	}
	return value, validNativeItem(value)
}

func validNativeItem(value ItemStack) bool {
	if len(value.Identifier) > maxItemIdentifier || len(value.CustomName) > maxItemText ||
		len(value.Lore) > maxItemLore || len(value.Enchantments) > maxItemEnchantments ||
		!utf8.ValidString(value.Identifier) || !utf8.ValidString(value.CustomName) {
		return false
	}
	total := len(value.Identifier) + len(value.CustomName) + len(value.NBT) + len(value.ValuesNBT)
	for _, line := range value.Lore {
		if len(line) > maxItemText || !utf8.ValidString(line) {
			return false
		}
		total += len(line)
		if total > maxItemDataBytes {
			return false
		}
	}
	if total > maxItemDataBytes || value.Count != 0 && value.Identifier == "" {
		return false
	}
	for _, enchantment := range value.Enchantments {
		if enchantment.Level == 0 {
			return false
		}
	}
	return true
}

func inventoryID(value C.DfInventoryId) InventoryID {
	return InventoryID{Player: playerID(value.player), Kind: InventoryKind(value.kind)}
}
