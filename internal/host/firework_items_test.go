package host

import (
	"reflect"
	"testing"
	"time"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/item"
)

func TestFireworkItemsFromNative(t *testing.T) {
	explosionNBT := fireworkExplosionNBT(2, 0, [1]byte{1}, true, true)
	rocketNBT := mustMarshalItemNBT(t, map[string]any{
		"Fireworks": map[string]any{
			"Explosions": []any{explosionNBT},
			"Flight":     uint8(2),
		},
	})
	stack, ok := itemStackFromNative(native.ItemStack{
		Identifier: "minecraft:firework_rocket",
		Count:      1,
		NBT:        rocketNBT,
	})
	if !ok {
		t.Fatal("decode firework rocket")
	}
	rocket, ok := stack.Item().(item.Firework)
	if !ok {
		t.Fatalf("rocket item type = %T", stack.Item())
	}
	if stack.Count() != 1 || rocket.Duration != 1500*time.Millisecond || len(rocket.Explosions) != 1 {
		t.Fatalf("rocket = %#v count=%d", rocket, stack.Count())
	}
	wantExplosion := item.FireworkExplosion{
		Shape: item.FireworkShapeStar(), Colour: item.ColourBlack(),
		Fade: item.ColourRed(), Fades: true, Twinkle: true, Trail: true,
	}
	if !reflect.DeepEqual(rocket.Explosions[0], wantExplosion) {
		t.Fatalf("rocket explosion = %#v, want %#v", rocket.Explosions[0], wantExplosion)
	}

	starNBT := mustMarshalItemNBT(t, map[string]any{
		"FireworksItem": fireworkExplosionNBT(4, 6, [0]byte{}, false, false),
		"customColor":   fireworkCustomColour(item.ColourCyan()),
	})
	stack, ok = itemStackFromNative(native.ItemStack{
		Identifier: "minecraft:firework_star",
		Metadata:   6,
		Count:      1,
		NBT:        starNBT,
	})
	if !ok {
		t.Fatal("decode firework star")
	}
	star, ok := stack.Item().(item.FireworkStar)
	if !ok {
		t.Fatalf("star item type = %T", stack.Item())
	}
	wantExplosion = item.FireworkExplosion{
		Shape: item.FireworkShapeBurst(), Colour: item.ColourCyan(),
	}
	if stack.Count() != 1 || !reflect.DeepEqual(star.FireworkExplosion, wantExplosion) {
		t.Fatalf("star = %#v count=%d, want explosion %#v", star, stack.Count(), wantExplosion)
	}
}

func TestFireworkItemsToNative(t *testing.T) {
	explosion := item.FireworkExplosion{
		Shape: item.FireworkShapeStar(), Colour: item.ColourBlack(),
		Fade: item.ColourRed(), Fades: true, Twinkle: true, Trail: true,
	}
	got, ok := itemStackToNative(item.NewStack(item.Firework{
		Duration: 1500 * time.Millisecond, Explosions: []item.FireworkExplosion{explosion},
	}, 1))
	if !ok {
		t.Fatal("encode firework rocket")
	}
	if got.Identifier != "minecraft:firework_rocket" || got.Metadata != 0 || got.Count != 1 {
		t.Fatalf("rocket stack = %#v", got)
	}
	wantNBT := map[string]any{
		"Fireworks": map[string]any{
			"Explosions": []any{fireworkExplosionNBT(2, 0, [1]byte{1}, true, true)},
			"Flight":     uint8(2),
		},
	}
	assertItemNBT(t, got.NBT, wantNBT)

	star := item.FireworkStar{FireworkExplosion: item.FireworkExplosion{
		Shape: item.FireworkShapeBurst(), Colour: item.ColourCyan(),
	}}
	got, ok = itemStackToNative(item.NewStack(star, 1))
	if !ok {
		t.Fatal("encode firework star")
	}
	if got.Identifier != "minecraft:firework_star" || got.Metadata != 6 || got.Count != 1 {
		t.Fatalf("star stack = %#v", got)
	}
	wantNBT = map[string]any{
		"FireworksItem": fireworkExplosionNBT(4, 6, [0]byte{}, false, false),
		"customColor":   fireworkCustomColour(item.ColourCyan()),
	}
	assertItemNBT(t, got.NBT, wantNBT)
}

func fireworkExplosionNBT(shape, colour uint8, fade any, twinkle, trail bool) map[string]any {
	return map[string]any{
		"FireworkType":    shape,
		"FireworkColor":   [1]byte{colour},
		"FireworkFade":    fade,
		"FireworkFlicker": boolByteForTest(twinkle),
		"FireworkTrail":   boolByteForTest(trail),
	}
}

func boolByteForTest(value bool) uint8 {
	if value {
		return 1
	}
	return 0
}

func fireworkCustomColour(colour item.Colour) int32 {
	rgba := colour.RGBA()
	return int32(uint32(rgba.A)<<24 | uint32(rgba.R)<<16 | uint32(rgba.G)<<8 | uint32(rgba.B))
}

func mustMarshalItemNBT(t *testing.T, value map[string]any) []byte {
	t.Helper()
	encoded, ok := marshalItemNBT(value)
	if !ok {
		t.Fatal("marshal item NBT")
	}
	return encoded
}

func assertItemNBT(t *testing.T, encoded []byte, want map[string]any) {
	t.Helper()
	got, ok := unmarshalItemNBT(encoded)
	if !ok {
		t.Fatal("unmarshal item NBT")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("item NBT = %#v, want %#v", got, want)
	}
}
