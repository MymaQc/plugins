package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectItemsUsesASTAndRegistry(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	spec, err := inspectItems(filepath.Join(string(bytes.TrimSpace(output)), "server", "item"))
	if err != nil {
		t.Fatal(err)
	}
	if spec.AirIdentifier != "minecraft:air" || len(spec.ToolTiers) != 7 || len(spec.ValueTypes) != 10 || len(spec.Types) != 125 {
		t.Fatalf("item spec has air=%q tiers=%d types=%d", spec.AirIdentifier, len(spec.ToolTiers), len(spec.Types))
	}
	generated := string(generateItems(spec))
	for _, value := range []string{
		"public readonly record struct Sword(ToolTier Tier) : World.Item",
		"public readonly record struct Beef(bool Cooked) : World.Item",
		"Item.ToolTierDiamond",
		`identifier = "minecraft:diamond_sword"; metadata = 0`,
		`identifier = "minecraft:cooked_beef"; metadata = 0`,
		"public static Colour ColourBlack() => new(15)",
		"public Color.RGBA SignRGBA() => _value switch",
		`15 => "black"`,
		`8 => "silver"`,
		"public static Value StrongSlowness() => new(42)",
		"public byte Uint8() => _value switch",
		"public static Horn Dream() => new(7)",
		`7 => "Dream"`,
		"public static DiscType DiscLavaChicken() => new(20)",
		`20 => "Lava Chicken"`,
		`20 => "Hyper Potions"`,
		"public static FireworkShape FireworkShapeBurst() => new(4)",
		`4 => "Burst"`,
		`4 => "burst"`,
		"public readonly record struct FireworkExplosion",
		"public readonly struct Firework : World.Item",
		"public TimeSpan RandomisedDuration()",
		"public bool OffHand() => true",
		"public readonly record struct FireworkStar(FireworkExplosion FireworkExplosion) : World.Item",
		`identifier = "minecraft:firework_rocket"; metadata = 0`,
		`identifier = "minecraft:firework_star"; metadata = 0`,
		`identifier = "minecraft:firework_star"; metadata = 15`,
		"public static WrittenBookGeneration CopyOfCopyGeneration() => new(2)",
		`2 => "copy of copy"`,
		"public readonly struct BookAndQuill : World.Item",
		"public BookAndQuill SetPage(int page, string text)",
		"public readonly struct WrittenBook : World.Item",
		"public WrittenBook(string title, string author, WrittenBookGeneration generation, params string[] pages)",
		`identifier = "minecraft:writable_book"; metadata = 0`,
		`identifier = "minecraft:written_book"; metadata = 0`,
		"public readonly record struct Arrow(global::Dragonfly.Potion.Value Tip) : World.Item",
		"public readonly record struct Dye(Colour Colour) : World.Item",
		"public readonly record struct GoatHorn(Sound.Horn Type) : World.Item",
		"public readonly record struct MusicDisc(Sound.DiscType DiscType) : World.Item",
		"public readonly record struct SmithingTemplate(SmithingTemplateType Template) : World.Item",
		"public readonly record struct BannerPattern(BannerPatternType Type) : World.Item",
		"public readonly record struct SuspiciousStew(StewType Type) : World.Item",
		"public readonly record struct PotterySherd(SherdType Type) : World.Item",
		`identifier = "minecraft:arrow"; metadata = 43`,
		`identifier = "minecraft:black_dye"; metadata = 0`,
		`identifier = "minecraft:music_disc_lava_chicken"; metadata = 0`,
		"internal static class ItemCapabilities",
		"Item.Snowball _ => 16",
		"durability = new(1561, false, default); return true",
		"Item.Sword value when value.Tier == Item.ToolTierDiamond => 8d",
		"Item.EnchantedBook _ => true",
		"private sealed record EncodedItem",
	} {
		if !strings.Contains(generated, value) {
			t.Fatalf("generated items missing %q", value)
		}
	}
}
