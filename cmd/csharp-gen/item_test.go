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
	module := string(bytes.TrimSpace(output))
	effects, err := inspectEffects(filepath.Join(module, "server", "entity", "effect"))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := inspectItems(filepath.Join(module, "server", "item"), effects)
	if err != nil {
		t.Fatal(err)
	}
	if spec.AirIdentifier != "minecraft:air" || len(spec.ToolTiers) != 7 || len(spec.ValueTypes) != 10 || len(spec.Types) != 132 || len(spec.Enchantments) != 27 {
		t.Fatalf("item spec has air=%q tiers=%d types=%d enchantments=%d", spec.AirIdentifier, len(spec.ToolTiers), len(spec.Types), len(spec.Enchantments))
	}
	protection := spec.Enchantments[0]
	if protection.Name != "Protection" || protection.ID != 0 || protection.DisplayName != "Protection" || protection.MaxLevel != 4 ||
		protection.CompatibleEnchantments&(1<<1) != 0 || !hasEncodedItem(protection.CompatibleItems, "minecraft:diamond_chestplate", 0) ||
		hasEncodedItem(protection.CompatibleItems, "minecraft:diamond_sword", 0) {
		t.Fatalf("protection enchantment spec = %#v", protection)
	}
	unbreaking := enchantmentByName(spec.Enchantments, "Unbreaking")
	if unbreaking == nil || unbreaking.ID != 17 || unbreaking.MaxLevel != 3 ||
		!hasEncodedItem(unbreaking.CompatibleItems, "minecraft:diamond_sword", 0) {
		t.Fatalf("unbreaking enchantment spec = %#v", unbreaking)
	}
	if bucket := spec.Bucket; !bucket.Present || bucket.ConsumeDuration != 1610*time.Millisecond || bucket.EmptyMaxCount != 16 || bucket.FullMaxCount != 1 {
		t.Fatalf("bucket spec = %#v", bucket)
	}
	bucket := itemTypeByName(spec.Types, "Bucket")
	if bucket == nil || !bucket.Bucket || len(bucket.States) != 4 {
		t.Fatalf("bucket item type = %#v", bucket)
	}
	wantBuckets := []struct {
		identifier string
		content    bucketContentKind
		maxCount   int
		duration   time.Duration
		residue    string
		count      int
	}{
		{"minecraft:bucket", bucketEmpty, 16, 0, "", 0},
		{"minecraft:water_bucket", bucketWater, 1, 0, "", 0},
		{"minecraft:lava_bucket", bucketLava, 1, 1000 * time.Second, "minecraft:bucket", 1},
		{"minecraft:milk_bucket", bucketMilk, 1, 0, "", 0},
	}
	for index, want := range wantBuckets {
		state := bucket.States[index]
		if state.Identifier != want.identifier || state.Metadata != 0 || state.Bucket != want.content ||
			state.Capability.MaxCount != want.maxCount || !state.Capability.Fuel || state.Capability.FuelDuration != want.duration ||
			state.Capability.FuelIdentifier != want.residue || state.Capability.FuelMetadata != 0 || state.Capability.FuelCount != want.count {
			t.Fatalf("bucket state %d = %#v", index, state)
		}
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
				lavaBucket := definition.Bucket && state.Bucket == bucketLava
				if !lavaBucket && (state.Capability.FuelIdentifier != "" || state.Capability.FuelMetadata != 0 || state.Capability.FuelCount != 0) {
					t.Fatalf("generated fuel %s unexpectedly has residue: %#v", definition.Name, state.Capability)
				}
			}
		}
	}
	if fuelTypes != 13 || fuelStates != 46 {
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
		"public readonly struct BucketContent",
		"public (World.Liquid? Liquid, bool Ok) Liquid() => (_liquid, _liquid is not null)",
		"public string String() => _milk ? \"milk\" : _liquid?.LiquidType() ?? string.Empty",
		"public string LiquidType() => _liquid is null ? \"milk\" : String()",
		"public override string ToString() => String()",
		"public static BucketContent LiquidBucketContent(World.Liquid liquid)",
		"public static BucketContent MilkBucketContent() => new(null, true)",
		"public readonly record struct Bucket(BucketContent Content = default) : World.Item, MaxCounter, Fuel",
		"public int MaxCount() => Empty() ? 16 : 1",
		"public bool AlwaysConsumable() => Content.Milk",
		"public bool CanConsume() => Content.Milk",
		"public TimeSpan ConsumeDuration() => TimeSpan.FromTicks(16100000)",
		"public bool Empty() => Content.RawLiquid is null && !Content.Milk",
		"case Item.Bucket value when value.Empty():",
		"case Item.Bucket value when value.Content.RawLiquid is Block.Water:",
		"case Item.Bucket value when value.Content.RawLiquid is Block.Lava:",
		"case Item.Bucket value when value.Content.Milk:",
		"case Item.Bucket value when value.Content.RawLiquid is not null:",
		`identifier = "minecraft:bucket"; metadata = 0`,
		`identifier = "minecraft:water_bucket"; metadata = 0`,
		`identifier = "minecraft:lava_bucket"; metadata = 0`,
		`identifier = "minecraft:milk_bucket"; metadata = 0`,
		`return new Item.Bucket(Item.LiquidBucketContent(new Block.Water(false, 0, false)))`,
		`return new Item.Bucket(Item.LiquidBucketContent(new Block.Lava(false, 0, false)))`,
		`return new Item.Bucket(Item.MilkBucketContent())`,
		`Item.Bucket value => value.FuelInfo()`,
		"Item.Bucket value => value.MaxCount()",
		"public readonly record struct Sword(ToolTier Tier) : World.Item, Fuel",
		"public readonly record struct Coal : World.Item, Fuel",
		`public FuelInfo FuelInfo() => Content.RawLiquid?.LiquidType() == "lava"`,
		`? new FuelInfo(TimeSpan.FromTicks(10000000000), NewStack(new Bucket(), 1))`,
		"public interface Fuel { FuelInfo FuelInfo(); }",
		"public readonly record struct FuelInfo(TimeSpan Duration = default, Stack Residue = default)",
		"public interface EnchantmentType",
		"public readonly record struct Enchantment",
		"public static Enchantment NewEnchantment(EnchantmentType type, int level)",
		"public static readonly EnchantmentType Protection = new BuiltinEnchantmentType(0, \"Protection\", 4",
		"public static readonly EnchantmentType Unbreaking = new BuiltinEnchantmentType(17, \"Unbreaking\", 3",
		"public static (EnchantmentType? Type, bool Ok) EnchantmentByID(int id)",
		"internal static bool EnchantmentCompatibleWithItem(int enchantment, World.Item? item)",
		"(0, \"minecraft:diamond_chestplate\", 0) => true",
		"(17, \"minecraft:diamond_sword\", 0) => true",
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
		"public static IReadOnlyList<StewType> StewTypes() => new StewType[]",
		"public static Value From(int id) => new(unchecked((byte)id))",
		"public static IReadOnlyList<Value> All() => new Value[]",
		"public IReadOnlyList<Effect.Value> Effects() => _value switch",
		"_ => unchecked((byte)_value)",
		"_ => Array.Empty<Effect.Value>()",
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
	bucketSurfaceStart := strings.Index(generated, "public readonly struct BucketContent")
	codecStart := strings.Index(generated, "internal static class ItemCodec")
	if bucketSurfaceStart < 0 || codecStart < bucketSurfaceStart {
		t.Fatal("generated bucket public surface boundaries missing")
	}
	bucketSurface := generated[bucketSurfaceStart:codecStart]
	for _, forbidden := range []string{"minecraft:", "public Stack Consume(", "public bool UseOnBlock("} {
		if strings.Contains(bucketSurface, forbidden) {
			t.Fatalf("generated bucket public surface exposes %q", forbidden)
		}
	}
}

func enchantmentByName(values []enchantmentSpec, name string) *enchantmentSpec {
	for index := range values {
		if values[index].Name == name {
			return &values[index]
		}
	}
	return nil
}

func hasEncodedItem(values []encodedItemKey, identifier string, metadata int) bool {
	for _, value := range values {
		if value.Identifier == identifier && value.Metadata == metadata {
			return true
		}
	}
	return false
}
