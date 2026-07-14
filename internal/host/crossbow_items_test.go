package host

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
)

func TestCrossbowItemRoundTrip(t *testing.T) {
	explosion := item.FireworkExplosion{
		Shape: item.FireworkShapeStar(), Colour: item.ColourBlack(),
		Fade: item.ColourRed(), Fades: true, Twinkle: true, Trail: true,
	}
	projectile := item.NewStack(item.Firework{
		Duration: 1500 * time.Millisecond, Explosions: []item.FireworkExplosion{explosion},
	}, 1).
		WithCustomName("charged rocket").
		WithLore("nested", "stack").
		WithValue("plugin", "preserved").
		WithForcedEnchantments(item.NewEnchantment(enchantment.QuickCharge, 2))
	source := item.NewStack(item.Crossbow{Item: projectile}, 1).
		Damage(7).
		AsUnbreakable().
		WithCustomName("loaded crossbow").
		WithLore("outer")

	encoded, ok := itemStackToNative(source)
	if !ok {
		t.Fatal("encode crossbow")
	}
	if encoded.Identifier != "minecraft:crossbow" || encoded.Count != 1 || encoded.Damage != 7 ||
		!encoded.Unbreakable || encoded.CustomName != "loaded crossbow" || !reflect.DeepEqual(encoded.Lore, []string{"outer"}) {
		t.Fatalf("crossbow transport = %#v", encoded)
	}
	root, ok := unmarshalItemNBT(encoded.NBT)
	if !ok {
		t.Fatalf("crossbow NBT invalid: %x", encoded.NBT)
	}
	charged, ok := root["chargedItem"].(map[string]any)
	if !ok || charged[nestedItemVersionName] != nestedItemVersion || charged["identifier"] != "minecraft:firework_rocket" {
		t.Fatalf("charged item transport = %#v", root["chargedItem"])
	}

	decoded, ok := itemStackFromNative(encoded)
	if !ok {
		t.Fatal("decode crossbow")
	}
	if decoded.Durability() != source.Durability() || !decoded.Unbreakable() || decoded.CustomName() != "loaded crossbow" ||
		!reflect.DeepEqual(decoded.Lore(), []string{"outer"}) {
		t.Fatalf("decoded crossbow stack = %#v", decoded)
	}
	crossbow, ok := decoded.Item().(item.Crossbow)
	if !ok {
		t.Fatalf("decoded item type = %T", decoded.Item())
	}
	rocket, ok := crossbow.Item.Item().(item.Firework)
	if !ok || rocket.Duration != 1500*time.Millisecond || !reflect.DeepEqual(rocket.Explosions, []item.FireworkExplosion{explosion}) {
		t.Fatalf("decoded charged item = %#v", crossbow.Item.Item())
	}
	if crossbow.Item.CustomName() != "charged rocket" || !reflect.DeepEqual(crossbow.Item.Lore(), []string{"nested", "stack"}) {
		t.Fatalf("decoded charged stack = %#v", crossbow.Item)
	}
	if value, found := crossbow.Item.Value("plugin"); !found || value != "preserved" {
		t.Fatalf("decoded charged values = %#v, %v", value, found)
	}
	enchantments := crossbow.Item.Enchantments()
	if len(enchantments) != 1 || enchantments[0].Type() != enchantment.QuickCharge || enchantments[0].Level() != 2 {
		t.Fatalf("decoded charged enchantments = %#v", enchantments)
	}
}

func TestCrossbowEmptyAndMalformedChargedItem(t *testing.T) {
	encoded, ok := itemStackToNative(item.NewStack(item.Crossbow{}, 1))
	if !ok || len(encoded.NBT) != 0 {
		t.Fatalf("empty crossbow transport = %#v, %v", encoded, ok)
	}
	decoded, ok := itemStackFromNative(native.ItemStack{
		Identifier: "minecraft:crossbow", Count: 1,
		NBT: mustMarshalItemNBT(t, map[string]any{"chargedItem": "invalid"}),
	})
	if !ok {
		t.Fatal("malformed charged item should decode as empty")
	}
	crossbow, ok := decoded.Item().(item.Crossbow)
	if !ok || !crossbow.Item.Empty() {
		t.Fatalf("malformed crossbow = %#v", decoded.Item())
	}
	loaded, ok := itemStackToNative(item.NewStack(item.Crossbow{Item: item.NewStack(item.Arrow{}, 1)}, 1))
	if !ok {
		t.Fatal("encode crossbow with plain arrow")
	}
	decoded, ok = itemStackFromNative(loaded)
	if !ok {
		t.Fatal("decode crossbow with plain arrow")
	}
	crossbow, ok = decoded.Item().(item.Crossbow)
	if !ok || crossbow.Item.Item() != (item.Arrow{}) {
		t.Fatalf("plain charged arrow = %#v", decoded.Item())
	}
}

func TestCrossbowNestedDepthLimit(t *testing.T) {
	stack := item.NewStack(item.Arrow{}, 1)
	for range maxNestedItemDepth + 1 {
		stack = item.NewStack(item.Crossbow{Item: stack}, 1)
	}
	if _, ok := itemStackToNative(stack); ok {
		t.Fatal("nested crossbow exceeded depth limit")
	}
}

func TestCrossbowNestedTransportLimits(t *testing.T) {
	base := native.ItemStack{Identifier: "minecraft:arrow", Count: 1}
	tests := []native.ItemStack{
		func() native.ItemStack {
			value := base
			value.Identifier = strings.Repeat("x", maxNestedItemIDBytes+1)
			return value
		}(),
		func() native.ItemStack {
			value := base
			value.CustomName = strings.Repeat("x", maxNestedItemText+1)
			return value
		}(),
		func() native.ItemStack {
			value := base
			value.Lore = make([]string, maxNestedItemEntries+1)
			return value
		}(),
		func() native.ItemStack {
			value := base
			value.Enchantments = []native.ItemEnchantment{{ID: 1, Level: 0}}
			return value
		}(),
	}
	for index, value := range tests {
		if _, ok := nestedItemCompound(value); ok {
			t.Fatalf("invalid nested stack %d was accepted", index)
		}
	}

	compound, ok := nestedItemCompound(base)
	if !ok {
		t.Fatal("encode valid nested stack")
	}
	compound["enchantments"] = make([]any, maxNestedItemEntries+1)
	if _, ok := nestedItemFromCompound(compound); ok {
		t.Fatal("oversized nested enchantment list was accepted")
	}
}
