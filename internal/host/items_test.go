package host

import (
	"reflect"
	"testing"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/player"
)

func TestPlayersInventoryItemRoundTrip(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		playerID := players.Register(player, 7)
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()
		valuesNBT, ok := marshalItemNBT(map[string]any{
			"owner": "rust", "level": int32(12), "flags": []byte{1, 0, 1},
		})
		if !ok {
			t.Fatal("encode values")
		}
		want := native.ItemStack{
			Identifier: "minecraft:diamond_sword", Count: 1, Damage: 3,
			Unbreakable: true, AnvilCost: 7,
			CustomName: "Plugin Sword", Lore: []string{"line one", "line two"}, ValuesNBT: valuesNBT,
			Enchantments: []native.ItemEnchantment{{ID: 9, Level: 5}, {ID: 17, Level: 3}},
		}
		main := native.InventoryID{Player: playerID, Kind: native.InventoryMain}
		if size, ok := players.InventorySize(invocation, main); !ok || size != 36 {
			t.Fatalf("main inventory size=%d ok=%v", size, ok)
		}
		if !players.SetInventoryItem(invocation, main, 2, want) {
			t.Fatal("set inventory item")
		}
		got, ok := players.InventoryItem(invocation, main, 2)
		if !ok {
			t.Fatal("read inventory item")
		}
		wantValues, _ := unmarshalItemNBT(want.ValuesNBT)
		gotValues, valuesOK := unmarshalItemNBT(got.ValuesNBT)
		got.ValuesNBT, want.ValuesNBT = nil, nil
		if !valuesOK || !reflect.DeepEqual(gotValues, wantValues) || !reflect.DeepEqual(got, want) {
			t.Fatalf("item mismatch\ngot:  %#v values=%#v\nwant: %#v values=%#v", got, gotValues, want, wantValues)
		}
		if !players.SetHeldSlot(invocation, playerID, 2) {
			t.Fatal("set held slot")
		}
		held, ok := players.HeldItem(invocation, playerID, 0)
		if !ok || held.Identifier != want.Identifier || held.CustomName != want.CustomName {
			t.Fatalf("held=%#v ok=%v", held, ok)
		}
	})
}

func TestPlayersInventoryAddClearAndOffhand(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		playerID := players.Register(player, 9)
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()
		main := native.InventoryID{Player: playerID, Kind: native.InventoryMain}
		added, ok := players.AddInventoryItem(invocation, main, native.ItemStack{Identifier: "minecraft:apple", Count: 70})
		if !ok || added != 70 {
			t.Fatalf("added=%d ok=%v", added, ok)
		}
		first, _ := players.InventoryItem(invocation, main, 0)
		second, _ := players.InventoryItem(invocation, main, 1)
		if first.Count != 64 || second.Count != 6 {
			t.Fatalf("apple counts=%d,%d", first.Count, second.Count)
		}
		offhand := native.InventoryID{Player: playerID, Kind: native.InventoryOffhand}
		if !players.SetInventoryItem(invocation, offhand, 0, native.ItemStack{Identifier: "minecraft:totem_of_undying", Count: 1}) {
			t.Fatal("set offhand")
		}
		item, ok := players.HeldItem(invocation, playerID, 1)
		if !ok || item.Identifier != "minecraft:totem_of_undying" {
			t.Fatalf("offhand=%#v ok=%v", item, ok)
		}
		if !players.ClearInventory(invocation, main) || !players.ClearInventory(invocation, offhand) {
			t.Fatal("clear inventory")
		}
		first, _ = players.InventoryItem(invocation, main, 0)
		item, _ = players.InventoryItem(invocation, offhand, 0)
		if first.Count != 0 || item.Count != 0 {
			t.Fatalf("inventories not cleared: main=%#v offhand=%#v", first, item)
		}
	})
}

func TestPlayersRejectInvalidNativeItems(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		id := native.InventoryID{Player: players.Register(player, 1), Kind: native.InventoryMain}
		invocation, leave := players.BeginInvocation(player.Tx())
		defer leave()
		invalid := []native.ItemStack{
			{Identifier: "missing:item", Count: 1},
			{Identifier: "minecraft:apple", Metadata: 1 << 16, Count: 1},
			{Identifier: "minecraft:diamond_sword", Count: 1, Enchantments: []native.ItemEnchantment{{ID: 9}}},
			{Identifier: "minecraft:diamond_sword", Count: 1, ValuesNBT: []byte{1, 2, 3}},
		}
		for index, value := range invalid {
			if players.SetInventoryItem(invocation, id, 0, value) {
				t.Fatalf("invalid item %d accepted", index)
			}
		}
	})
}
