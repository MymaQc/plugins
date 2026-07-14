package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	if spec.AirIdentifier != "minecraft:air" || len(spec.ToolTiers) != 7 || len(spec.ValueTypes) != 10 || len(spec.Types) != 131 {
		t.Fatalf("item spec has air=%q tiers=%d types=%d", spec.AirIdentifier, len(spec.ToolTiers), len(spec.Types))
	}
	if crossbow := spec.Crossbow; !crossbow.Present || crossbow.MaxCount != 1 || crossbow.MaxDurability != 464 ||
		crossbow.EnchantmentValue != 1 || crossbow.FuelDuration != 15*time.Second {
		t.Fatalf("crossbow spec = %#v", crossbow)
	}
	crossbow := itemTypeByName(spec.Types, "Crossbow")
	if crossbow == nil || !crossbow.NBT || len(crossbow.States) != 1 || crossbow.States[0].Identifier != "minecraft:crossbow" ||
		!crossbow.States[0].Capability.Fuel || crossbow.States[0].Capability.FuelDuration != 15*time.Second {
		t.Fatalf("crossbow item type = %#v", crossbow)
	}
	fuelTypes, fuelStates := 0, 0
	for _, definition := range spec.Types {
		if itemTypeFuel(definition) {
			fuelTypes++
		}
		for _, state := range definition.States {
			if state.Capability.Fuel {
				fuelStates++
				if state.Capability.FuelIdentifier != "" || state.Capability.FuelMetadata != 0 || state.Capability.FuelCount != 0 {
					t.Fatalf("generated fuel %s unexpectedly has residue: %#v", definition.Name, state.Capability)
				}
			}
		}
	}
	if fuelTypes != 12 || fuelStates != 42 {
		t.Fatalf("fuel spec has types=%d states=%d", fuelTypes, fuelStates)
	}
	for _, name := range []string{"Axe", "Hoe", "Pickaxe", "Shovel", "Sword"} {
		definition := itemTypeByName(spec.Types, name)
		if definition == nil || len(definition.States) != 7 {
			t.Fatalf("fuel tool %s missing or incomplete", name)
		}
		for tier, state := range definition.States {
			want := time.Duration(0)
			if tier == 0 {
				want = 10 * time.Second
			}
			if !state.Capability.Fuel || state.Capability.FuelDuration != want {
				t.Fatalf("fuel tool %s tier %d = %#v, want duration %s", name, tier, state.Capability, want)
			}
		}
	}
	if len(spec.Armour.Tiers) != 7 || len(spec.Armour.Pieces) != 4 || len(spec.Armour.TrimMaterials) != 11 {
		t.Fatalf("armour spec has tiers=%d pieces=%d trim_materials=%d", len(spec.Armour.Tiers), len(spec.Armour.Pieces), len(spec.Armour.TrimMaterials))
	}
	if leather := spec.Armour.Tiers[0]; leather.Name != "ArmourTierLeather" || leather.BaseDurability != 55 ||
		leather.EnchantmentValue != 15 || leather.IdentifierName != "leather" || !leather.Colour {
		t.Fatalf("leather armour tier = %#v", leather)
	}
	if netherite := spec.Armour.Tiers[6]; netherite.Name != "ArmourTierNetherite" || netherite.BaseDurability != 408 ||
		netherite.Toughness != 3 || netherite.KnockBackResistance != 0.1 || netherite.IdentifierName != "netherite" {
		t.Fatalf("netherite armour tier = %#v", netherite)
	}
	wantDurability := map[string][]int{
		"Helmet":     {55, 121, 77, 166, 165, 363, 408},
		"Chestplate": {80, 176, 112, 241, 240, 528, 593},
		"Leggings":   {77, 169, 107, 232, 231, 508, 571},
		"Boots":      {65, 143, 91, 196, 195, 429, 482},
	}
	wantDefence := map[string][]float64{
		"Helmet":     {1, 2, 2, 2, 2, 3, 3},
		"Chestplate": {3, 4, 5, 5, 6, 8, 8},
		"Leggings":   {2, 3, 3, 4, 5, 6, 6},
		"Boots":      {1, 1, 1, 1, 2, 3, 3},
	}
	for pieceName, durability := range wantDurability {
		piece := armourPieceByName(spec.Armour.Pieces, pieceName)
		definition := itemTypeByName(spec.Types, pieceName)
		if piece == nil || definition == nil || len(definition.States) != 7 {
			t.Fatalf("armour piece %s missing or incomplete", pieceName)
		}
		for tierIndex, tier := range spec.Armour.Tiers {
			state := definition.States[tierIndex]
			wantIdentifier := "minecraft:" + tier.IdentifierName + "_" + strings.ToLower(pieceName)
			if state.ArmourTier != tierIndex || state.Identifier != wantIdentifier || state.Metadata != 0 ||
				state.Capability.MaxCount != 1 || state.Capability.MaxDurability != durability[tierIndex] ||
				piece.DefencePoints[tierIndex] != wantDefence[pieceName][tierIndex] || piece.RepairItems[tierIndex] == "" {
				t.Fatalf("armour %s tier %s state=%#v piece=%#v", pieceName, tier.Name, state, piece)
			}
		}
	}
	wantMaterials := []string{
		"AmethystShard:amethyst:§u", "CopperIngot:copper:§n", "Diamond:diamond:§s",
		"Emerald:emerald:§q", "GoldIngot:gold:§p", "IronIngot:iron:§i", "LapisLazuli:lapis:§t",
		"NetheriteIngot:netherite:§j", "NetherQuartz:quartz:§h", "ResinBrick:resin:§v", "RedstoneWire:redstone:§m",
	}
	for index, material := range spec.Armour.TrimMaterials {
		if got := material.ItemName + ":" + material.Material + ":" + material.MaterialColour; got != wantMaterials[index] {
			t.Fatalf("armour trim material %d=%q, want %q", index, got, wantMaterials[index])
		}
	}
	generated := string(generateItems(spec))
	for _, material := range spec.Armour.TrimMaterials {
		mapping := `case "` + material.Material + `": material = new Item.` + material.ItemName + `(); return true`
		if !strings.Contains(generated, mapping) {
			t.Fatalf("generated items missing armour material mapping %q", mapping)
		}
	}
	templates := findItemValueType(spec.ValueTypes, "SmithingTemplateType")
	for index, value := range templates.Values {
		name := value.(interface{ String() string }).String()
		mapping := `case "` + name + `": template = ` + itemValueFactory(*templates, index) + `; return true`
		if !strings.Contains(generated, mapping) {
			t.Fatalf("generated items missing armour template mapping %q", mapping)
		}
	}
	for _, value := range []string{
		"public readonly record struct Sword(ToolTier Tier) : World.Item, Fuel",
		"public readonly record struct Coal : World.Item, Fuel",
		"public FuelInfo FuelInfo() => ItemCapabilities.FuelInfo(this)",
		"public interface Fuel { FuelInfo FuelInfo(); }",
		"public readonly record struct FuelInfo(TimeSpan Duration = default, Stack Residue = default)",
		"public FuelInfo WithResidue(Stack residue) => this with { Residue = residue }",
		"public readonly record struct Crossbow(Stack Item = default) : World.Item, MaxCounter, Durable, Fuel, Enchantable",
		"public DurabilityInfo DurabilityInfo() => new(464, static () => default)",
		"public FuelInfo FuelInfo() => new(TimeSpan.FromTicks(150000000))",
		"public int EnchantmentValue() => 1",
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
		"public interface ArmourTier",
		"public interface ArmourTrimMaterial",
		"public readonly record struct ArmourTierLeather(global::Dragonfly.Color.RGBA Colour = default) : ArmourTier",
		"public readonly record struct ArmourTrim(SmithingTemplateType Template, ArmourTrimMaterial? Material)",
		"public readonly record struct Helmet(ArmourTier Tier, ArmourTrim Trim = default) : World.Item, HelmetType, Trimmable, MaxCounter, Enchantable, Repairable, Smeltable",
		"bool HelmetType.Helmet() => true",
		"public readonly record struct RedstoneWire : World.Item, ArmourTrimMaterial",
		`public string TrimMaterial() => "redstone"`,
		`identifier = "minecraft:leather_helmet"; metadata = 0`,
		`identifier = "minecraft:netherite_boots"; metadata = 0`,
		"case Item.Helmet value when value.Tier is Item.ArmourTierLeather:",
		"return new Item.Helmet(new Item.ArmourTierLeather())",
		`case "diamond": material = new Item.Diamond(); return true`,
		`case "bolt": template = Item.TemplateBolt(); return true`,
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
		"Item.Sword value when value.Tier == Item.ToolTierWood => new(TimeSpan.FromTicks(100000000), default)",
		"Item.Sword value when value.Tier == Item.ToolTierDiamond => new(TimeSpan.FromTicks(0), default)",
		"Item.BlazeRod _ => new(TimeSpan.FromTicks(1200000000), default)",
		"Item.Crossbow _ => new(TimeSpan.FromTicks(150000000), default)",
		"Item.EnchantedBook _ => true",
		"private sealed record EncodedItem",
	} {
		if !strings.Contains(generated, value) {
			t.Fatalf("generated items missing %q", value)
		}
	}
}
