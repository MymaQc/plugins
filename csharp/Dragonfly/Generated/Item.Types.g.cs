// Code generated from Dragonfly server/item Go AST and live registry. DO NOT EDIT.
#nullable enable
using System;

namespace Dragonfly
{
    public static partial class Item
    {
        public readonly record struct ToolTier(int HarvestLevel, double BaseMiningEfficiency, double BaseAttackDamage, int EnchantmentValue, int Durability, string Name);

        public static readonly ToolTier ToolTierWood = new(1, 2d, 1d, 15, 59, "wooden");
        public static readonly ToolTier ToolTierGold = new(1, 12d, 1d, 22, 32, "golden");
        public static readonly ToolTier ToolTierStone = new(2, 4d, 2d, 5, 131, "stone");
        public static readonly ToolTier ToolTierCopper = new(2, 5d, 2d, 13, 190, "copper");
        public static readonly ToolTier ToolTierIron = new(3, 6d, 3d, 14, 250, "iron");
        public static readonly ToolTier ToolTierDiamond = new(4, 8d, 4d, 10, 1561, "diamond");
        public static readonly ToolTier ToolTierNetherite = new(4, 9d, 5d, 15, 2031, "netherite");

        public readonly record struct Colour
        {
            private readonly int _value;
            internal Colour(int value) => _value = value;
            internal int Id => _value;

            public Color.RGBA RGBA() => _value switch
            {
                0 => new Color.RGBA(240, 240, 240, 255),
                1 => new Color.RGBA(249, 128, 29, 255),
                2 => new Color.RGBA(199, 78, 189, 255),
                3 => new Color.RGBA(58, 179, 218, 255),
                4 => new Color.RGBA(254, 216, 61, 255),
                5 => new Color.RGBA(128, 199, 31, 255),
                6 => new Color.RGBA(243, 139, 170, 255),
                7 => new Color.RGBA(71, 79, 82, 255),
                8 => new Color.RGBA(157, 157, 151, 255),
                9 => new Color.RGBA(22, 156, 156, 255),
                10 => new Color.RGBA(137, 50, 184, 255),
                11 => new Color.RGBA(60, 68, 170, 255),
                12 => new Color.RGBA(131, 84, 50, 255),
                13 => new Color.RGBA(94, 124, 22, 255),
                14 => new Color.RGBA(176, 46, 38, 255),
                15 => new Color.RGBA(29, 29, 33, 255),
                _ => throw new InvalidOperationException("Invalid Colour value."),
            };

            public Color.RGBA SignRGBA() => _value switch
            {
                0 => new Color.RGBA(240, 240, 240, 255),
                1 => new Color.RGBA(249, 128, 29, 255),
                2 => new Color.RGBA(199, 78, 189, 255),
                3 => new Color.RGBA(58, 179, 218, 255),
                4 => new Color.RGBA(254, 216, 61, 255),
                5 => new Color.RGBA(128, 199, 31, 255),
                6 => new Color.RGBA(243, 139, 170, 255),
                7 => new Color.RGBA(71, 79, 82, 255),
                8 => new Color.RGBA(157, 157, 151, 255),
                9 => new Color.RGBA(22, 156, 156, 255),
                10 => new Color.RGBA(137, 50, 184, 255),
                11 => new Color.RGBA(60, 68, 170, 255),
                12 => new Color.RGBA(131, 84, 50, 255),
                13 => new Color.RGBA(94, 124, 22, 255),
                14 => new Color.RGBA(176, 46, 38, 255),
                15 => new Color.RGBA(0, 0, 0, 255),
                _ => throw new InvalidOperationException("Invalid Colour value."),
            };

            public string String() => _value switch
            {
                0 => "white",
                1 => "orange",
                2 => "magenta",
                3 => "light_blue",
                4 => "yellow",
                5 => "lime",
                6 => "pink",
                7 => "gray",
                8 => "light_gray",
                9 => "cyan",
                10 => "purple",
                11 => "blue",
                12 => "brown",
                13 => "green",
                14 => "red",
                15 => "black",
                _ => throw new InvalidOperationException("Invalid Colour value."),
            };

            public string SilverString() => _value switch
            {
                0 => "white",
                1 => "orange",
                2 => "magenta",
                3 => "light_blue",
                4 => "yellow",
                5 => "lime",
                6 => "pink",
                7 => "gray",
                8 => "silver",
                9 => "cyan",
                10 => "purple",
                11 => "blue",
                12 => "brown",
                13 => "green",
                14 => "red",
                15 => "black",
                _ => throw new InvalidOperationException("Invalid Colour value."),
            };

            public byte Uint8() => _value switch
            {
                0 => 0,
                1 => 1,
                2 => 2,
                3 => 3,
                4 => 4,
                5 => 5,
                6 => 6,
                7 => 7,
                8 => 8,
                9 => 9,
                10 => 10,
                11 => 11,
                12 => 12,
                13 => 13,
                14 => 14,
                15 => 15,
                _ => throw new InvalidOperationException("Invalid Colour value."),
            };
        }

        public static Colour ColourWhite() => new(0);
        public static Colour ColourOrange() => new(1);
        public static Colour ColourMagenta() => new(2);
        public static Colour ColourLightBlue() => new(3);
        public static Colour ColourYellow() => new(4);
        public static Colour ColourLime() => new(5);
        public static Colour ColourPink() => new(6);
        public static Colour ColourGrey() => new(7);
        public static Colour ColourLightGrey() => new(8);
        public static Colour ColourCyan() => new(9);
        public static Colour ColourPurple() => new(10);
        public static Colour ColourBlue() => new(11);
        public static Colour ColourBrown() => new(12);
        public static Colour ColourGreen() => new(13);
        public static Colour ColourRed() => new(14);
        public static Colour ColourBlack() => new(15);

        public readonly record struct SmithingTemplateType
        {
            private readonly int _value;
            internal SmithingTemplateType(int value) => _value = value;
            internal int Id => _value;

            public string String() => _value switch
            {
                0 => "netherite_upgrade",
                1 => "sentry",
                2 => "vex",
                3 => "wild",
                4 => "coast",
                5 => "dune",
                6 => "wayfinder",
                7 => "raiser",
                8 => "shaper",
                9 => "host",
                10 => "ward",
                11 => "silence",
                12 => "tide",
                13 => "snout",
                14 => "rib",
                15 => "eye",
                16 => "spire",
                17 => "flow",
                18 => "bolt",
                _ => throw new InvalidOperationException("Invalid SmithingTemplateType value."),
            };
        }

        public static SmithingTemplateType TemplateNetheriteUpgrade() => new(0);
        public static SmithingTemplateType TemplateSentry() => new(1);
        public static SmithingTemplateType TemplateVex() => new(2);
        public static SmithingTemplateType TemplateWild() => new(3);
        public static SmithingTemplateType TemplateCoast() => new(4);
        public static SmithingTemplateType TemplateDune() => new(5);
        public static SmithingTemplateType TemplateWayFinder() => new(6);
        public static SmithingTemplateType TemplateRaiser() => new(7);
        public static SmithingTemplateType TemplateShaper() => new(8);
        public static SmithingTemplateType TemplateHost() => new(9);
        public static SmithingTemplateType TemplateWard() => new(10);
        public static SmithingTemplateType TemplateSilence() => new(11);
        public static SmithingTemplateType TemplateTide() => new(12);
        public static SmithingTemplateType TemplateSnout() => new(13);
        public static SmithingTemplateType TemplateRib() => new(14);
        public static SmithingTemplateType TemplateEye() => new(15);
        public static SmithingTemplateType TemplateSpire() => new(16);
        public static SmithingTemplateType TemplateFlow() => new(17);
        public static SmithingTemplateType TemplateBolt() => new(18);

        public readonly record struct BannerPatternType
        {
            private readonly int _value;
            internal BannerPatternType(int value) => _value = value;
            internal int Id => _value;

            public byte Uint8() => _value switch
            {
                0 => 0,
                1 => 1,
                2 => 2,
                3 => 3,
                4 => 4,
                5 => 5,
                6 => 6,
                7 => 7,
                8 => 8,
                9 => 9,
                _ => throw new InvalidOperationException("Invalid BannerPatternType value."),
            };

            public string String() => _value switch
            {
                0 => "creeper",
                1 => "skull",
                2 => "flower",
                3 => "mojang",
                4 => "field_masoned",
                5 => "bordure_indented",
                6 => "piglin",
                7 => "globe",
                8 => "flow",
                9 => "guster",
                _ => throw new InvalidOperationException("Invalid BannerPatternType value."),
            };
        }

        public static BannerPatternType CreeperBannerPattern() => new(0);
        public static BannerPatternType SkullBannerPattern() => new(1);
        public static BannerPatternType FlowerBannerPattern() => new(2);
        public static BannerPatternType MojangBannerPattern() => new(3);
        public static BannerPatternType FieldMasonedBannerPattern() => new(4);
        public static BannerPatternType BordureIndentedBannerPattern() => new(5);
        public static BannerPatternType PiglinBannerPattern() => new(6);
        public static BannerPatternType GlobeBannerPattern() => new(7);
        public static BannerPatternType FlowBannerPattern() => new(8);
        public static BannerPatternType GusterBannerPattern() => new(9);

        public readonly record struct StewType
        {
            private readonly int _value;
            internal StewType(int value) => _value = value;
            internal int Id => _value;

            public byte Uint8() => _value switch
            {
                0 => 0,
                1 => 1,
                2 => 2,
                3 => 3,
                4 => 4,
                5 => 5,
                6 => 6,
                7 => 7,
                8 => 8,
                9 => 9,
                10 => 10,
                11 => 11,
                12 => 12,
                _ => throw new InvalidOperationException("Invalid StewType value."),
            };
        }

        public static StewType NightVisionPoppyStew() => new(0);
        public static StewType JumpBoostStew() => new(1);
        public static StewType WeaknessStew() => new(2);
        public static StewType BlindnessBluetStew() => new(3);
        public static StewType PoisonStew() => new(4);
        public static StewType SaturationDandelionStew() => new(5);
        public static StewType SaturationOrchidStew() => new(6);
        public static StewType FireResistanceStew() => new(7);
        public static StewType RegenerationStew() => new(8);
        public static StewType WitherStew() => new(9);
        public static StewType NightVisionTorchflowerStew() => new(10);
        public static StewType BlindnessEyeblossomStew() => new(11);
        public static StewType NauseaStew() => new(12);

        public readonly record struct SherdType
        {
            private readonly int _value;
            internal SherdType(int value) => _value = value;
            internal int Id => _value;

            public string String() => _value switch
            {
                0 => "angler",
                1 => "archer",
                2 => "arms_up",
                3 => "blade",
                4 => "brewer",
                5 => "burn",
                6 => "danger",
                7 => "explorer",
                8 => "friend",
                9 => "heart",
                10 => "heartbreak",
                11 => "howl",
                12 => "miner",
                13 => "mourner",
                14 => "plenty",
                15 => "prize",
                16 => "sheaf",
                17 => "shelter",
                18 => "skull",
                19 => "snort",
                20 => "flow",
                21 => "guster",
                22 => "scrape",
                _ => throw new InvalidOperationException("Invalid SherdType value."),
            };

            public byte Uint8() => _value switch
            {
                0 => 0,
                1 => 1,
                2 => 2,
                3 => 3,
                4 => 4,
                5 => 5,
                6 => 6,
                7 => 7,
                8 => 8,
                9 => 9,
                10 => 10,
                11 => 11,
                12 => 12,
                13 => 13,
                14 => 14,
                15 => 15,
                16 => 16,
                17 => 17,
                18 => 18,
                19 => 19,
                20 => 20,
                21 => 21,
                22 => 22,
                _ => throw new InvalidOperationException("Invalid SherdType value."),
            };
        }

        public static SherdType SherdTypeAngler() => new(0);
        public static SherdType SherdTypeArcher() => new(1);
        public static SherdType SherdTypeArmsUp() => new(2);
        public static SherdType SherdTypeBlade() => new(3);
        public static SherdType SherdTypeBrewer() => new(4);
        public static SherdType SherdTypeBurn() => new(5);
        public static SherdType SherdTypeDanger() => new(6);
        public static SherdType SherdTypeExplorer() => new(7);
        public static SherdType SherdTypeFriend() => new(8);
        public static SherdType SherdTypeHeart() => new(9);
        public static SherdType SherdTypeHeartbreak() => new(10);
        public static SherdType SherdTypeHowl() => new(11);
        public static SherdType SherdTypeMiner() => new(12);
        public static SherdType SherdTypeMourner() => new(13);
        public static SherdType SherdTypePlenty() => new(14);
        public static SherdType SherdTypePrize() => new(15);
        public static SherdType SherdTypeSheaf() => new(16);
        public static SherdType SherdTypeShelter() => new(17);
        public static SherdType SherdTypeSkull() => new(18);
        public static SherdType SherdTypeSnort() => new(19);
        public static SherdType SherdTypeFlow() => new(20);
        public static SherdType SherdTypeGuster() => new(21);
        public static SherdType SherdTypeScrape() => new(22);

        public readonly record struct AmethystShard : World.Item;
        public readonly record struct Apple : World.Item;
        public readonly record struct Arrow(global::Dragonfly.Potion.Value Tip) : World.Item;
        public readonly record struct Axe(ToolTier Tier) : World.Item;
        public readonly record struct BakedPotato : World.Item;
        public readonly record struct BannerPattern(BannerPatternType Type) : World.Item;
        public readonly record struct Beef(bool Cooked) : World.Item;
        public readonly record struct Beetroot : World.Item;
        public readonly record struct BeetrootSoup : World.Item;
        public readonly record struct BlazePowder : World.Item;
        public readonly record struct BlazeRod : World.Item;
        public readonly record struct Bone : World.Item;
        public readonly record struct BoneMeal : World.Item;
        public readonly record struct Book : World.Item;
        public readonly record struct BottleOfEnchanting : World.Item;
        public readonly record struct Bow : World.Item;
        public readonly record struct Bowl : World.Item;
        public readonly record struct Bread : World.Item;
        public readonly record struct Brick : World.Item;
        public readonly record struct CarrotOnAStick : World.Item;
        public readonly record struct Charcoal : World.Item;
        public readonly record struct Chicken(bool Cooked) : World.Item;
        public readonly record struct ClayBall : World.Item;
        public readonly record struct Clock : World.Item;
        public readonly record struct Coal : World.Item;
        public readonly record struct Cod(bool Cooked) : World.Item;
        public readonly record struct Compass : World.Item;
        public readonly record struct Cookie : World.Item;
        public readonly record struct CopperIngot : World.Item;
        public readonly record struct CopperNugget : World.Item;
        public readonly record struct Diamond : World.Item;
        public readonly record struct DiscFragment : World.Item;
        public readonly record struct DragonBreath : World.Item;
        public readonly record struct DriedKelp : World.Item;
        public readonly record struct Dye(Colour Colour) : World.Item;
        public readonly record struct EchoShard : World.Item;
        public readonly record struct Egg : World.Item;
        public readonly record struct Elytra : World.Item;
        public readonly record struct Emerald : World.Item;
        public readonly record struct EnchantedApple : World.Item;
        public readonly record struct EnchantedBook : World.Item;
        public readonly record struct EnderEye : World.Item;
        public readonly record struct EnderPearl : World.Item;
        public readonly record struct Feather : World.Item;
        public readonly record struct FermentedSpiderEye : World.Item;
        public readonly record struct FireCharge : World.Item;
        public readonly record struct Flint : World.Item;
        public readonly record struct FlintAndSteel : World.Item;
        public readonly record struct GhastTear : World.Item;
        public readonly record struct GlassBottle : World.Item;
        public readonly record struct GlisteringMelonSlice : World.Item;
        public readonly record struct GlowstoneDust : World.Item;
        public readonly record struct GoatHorn(Sound.Horn Type) : World.Item;
        public readonly record struct GoldIngot : World.Item;
        public readonly record struct GoldNugget : World.Item;
        public readonly record struct GoldenApple : World.Item;
        public readonly record struct GoldenCarrot : World.Item;
        public readonly record struct Gunpowder : World.Item;
        public readonly record struct HeartOfTheSea : World.Item;
        public readonly record struct Hoe(ToolTier Tier) : World.Item;
        public readonly record struct HoneyBottle : World.Item;
        public readonly record struct Honeycomb : World.Item;
        public readonly record struct InkSac(bool Glowing) : World.Item;
        public readonly record struct IronIngot : World.Item;
        public readonly record struct IronNugget : World.Item;
        public readonly record struct LapisLazuli : World.Item;
        public readonly record struct Leather : World.Item;
        public readonly record struct LingeringPotion(global::Dragonfly.Potion.Value Type) : World.Item;
        public readonly record struct MagmaCream : World.Item;
        public readonly record struct MelonSlice : World.Item;
        public readonly record struct MushroomStew : World.Item;
        public readonly record struct MusicDisc(Sound.DiscType DiscType) : World.Item;
        public readonly record struct Mutton(bool Cooked) : World.Item;
        public readonly record struct NautilusShell : World.Item;
        public readonly record struct NetherBrick : World.Item;
        public readonly record struct NetherQuartz : World.Item;
        public readonly record struct NetherStar : World.Item;
        public readonly record struct NetheriteIngot : World.Item;
        public readonly record struct NetheriteScrap : World.Item;
        public readonly record struct Paper : World.Item;
        public readonly record struct PhantomMembrane : World.Item;
        public readonly record struct Pickaxe(ToolTier Tier) : World.Item;
        public readonly record struct PoisonousPotato : World.Item;
        public readonly record struct PoppedChorusFruit : World.Item;
        public readonly record struct Porkchop(bool Cooked) : World.Item;
        public readonly record struct Potion(global::Dragonfly.Potion.Value Type) : World.Item;
        public readonly record struct PotterySherd(SherdType Type) : World.Item;
        public readonly record struct PrismarineCrystals : World.Item;
        public readonly record struct PrismarineShard : World.Item;
        public readonly record struct Pufferfish : World.Item;
        public readonly record struct PumpkinPie : World.Item;
        public readonly record struct Rabbit(bool Cooked) : World.Item;
        public readonly record struct RabbitFoot : World.Item;
        public readonly record struct RabbitHide : World.Item;
        public readonly record struct RabbitStew : World.Item;
        public readonly record struct RawCopper : World.Item;
        public readonly record struct RawGold : World.Item;
        public readonly record struct RawIron : World.Item;
        public readonly record struct RecoveryCompass : World.Item;
        public readonly record struct ResinBrick : World.Item;
        public readonly record struct RottenFlesh : World.Item;
        public readonly record struct Salmon(bool Cooked) : World.Item;
        public readonly record struct Scute : World.Item;
        public readonly record struct Shears : World.Item;
        public readonly record struct Shovel(ToolTier Tier) : World.Item;
        public readonly record struct ShulkerShell : World.Item;
        public readonly record struct Slimeball : World.Item;
        public readonly record struct SmithingTemplate(SmithingTemplateType Template) : World.Item;
        public readonly record struct Snowball : World.Item;
        public readonly record struct SpiderEye : World.Item;
        public readonly record struct SplashPotion(global::Dragonfly.Potion.Value Type) : World.Item;
        public readonly record struct Spyglass : World.Item;
        public readonly record struct Stick : World.Item;
        public readonly record struct Sugar : World.Item;
        public readonly record struct SuspiciousStew(StewType Type) : World.Item;
        public readonly record struct Sword(ToolTier Tier) : World.Item;
        public readonly record struct Totem : World.Item;
        public readonly record struct TropicalFish : World.Item;
        public readonly record struct TurtleShell : World.Item;
        public readonly record struct WarpedFungusOnAStick : World.Item;
        public readonly record struct Wheat : World.Item;
    }

    public static partial class Potion
    {
        public readonly record struct Value
        {
            private readonly int _value;
            internal Value(int value) => _value = value;
            internal int Id => _value;

            public byte Uint8() => _value switch
            {
                0 => 0,
                1 => 1,
                2 => 2,
                3 => 3,
                4 => 4,
                5 => 5,
                6 => 6,
                7 => 7,
                8 => 8,
                9 => 9,
                10 => 10,
                11 => 11,
                12 => 12,
                13 => 13,
                14 => 14,
                15 => 15,
                16 => 16,
                17 => 17,
                18 => 18,
                19 => 19,
                20 => 20,
                21 => 21,
                22 => 22,
                23 => 23,
                24 => 24,
                25 => 25,
                26 => 26,
                27 => 27,
                28 => 28,
                29 => 29,
                30 => 30,
                31 => 31,
                32 => 32,
                33 => 33,
                34 => 34,
                35 => 35,
                36 => 36,
                37 => 37,
                38 => 38,
                39 => 39,
                40 => 40,
                41 => 41,
                42 => 42,
                _ => throw new InvalidOperationException("Invalid Value value."),
            };
        }

        public static Value Water() => new(0);
        public static Value Mundane() => new(1);
        public static Value LongMundane() => new(2);
        public static Value Thick() => new(3);
        public static Value Awkward() => new(4);
        public static Value NightVision() => new(5);
        public static Value LongNightVision() => new(6);
        public static Value Invisibility() => new(7);
        public static Value LongInvisibility() => new(8);
        public static Value Leaping() => new(9);
        public static Value LongLeaping() => new(10);
        public static Value StrongLeaping() => new(11);
        public static Value FireResistance() => new(12);
        public static Value LongFireResistance() => new(13);
        public static Value Swiftness() => new(14);
        public static Value LongSwiftness() => new(15);
        public static Value StrongSwiftness() => new(16);
        public static Value Slowness() => new(17);
        public static Value LongSlowness() => new(18);
        public static Value WaterBreathing() => new(19);
        public static Value LongWaterBreathing() => new(20);
        public static Value Healing() => new(21);
        public static Value StrongHealing() => new(22);
        public static Value Harming() => new(23);
        public static Value StrongHarming() => new(24);
        public static Value Poison() => new(25);
        public static Value LongPoison() => new(26);
        public static Value StrongPoison() => new(27);
        public static Value Regeneration() => new(28);
        public static Value LongRegeneration() => new(29);
        public static Value StrongRegeneration() => new(30);
        public static Value Strength() => new(31);
        public static Value LongStrength() => new(32);
        public static Value StrongStrength() => new(33);
        public static Value Weakness() => new(34);
        public static Value LongWeakness() => new(35);
        public static Value Wither() => new(36);
        public static Value TurtleMaster() => new(37);
        public static Value LongTurtleMaster() => new(38);
        public static Value StrongTurtleMaster() => new(39);
        public static Value SlowFalling() => new(40);
        public static Value LongSlowFalling() => new(41);
        public static Value StrongSlowness() => new(42);

    }

    public static partial class Sound
    {
        public readonly record struct Horn
        {
            private readonly int _value;
            internal Horn(int value) => _value = value;
            internal int Id => _value;

            public byte Uint8() => _value switch
            {
                0 => 0,
                1 => 1,
                2 => 2,
                3 => 3,
                4 => 4,
                5 => 5,
                6 => 6,
                7 => 7,
                _ => throw new InvalidOperationException("Invalid Horn value."),
            };

            public string Name() => _value switch
            {
                0 => "Ponder",
                1 => "Sing",
                2 => "Seek",
                3 => "Feel",
                4 => "Admire",
                5 => "Call",
                6 => "Yearn",
                7 => "Dream",
                _ => throw new InvalidOperationException("Invalid Horn value."),
            };
        }

        public static Horn Ponder() => new(0);
        public static Horn Sing() => new(1);
        public static Horn Seek() => new(2);
        public static Horn Feel() => new(3);
        public static Horn Admire() => new(4);
        public static Horn Call() => new(5);
        public static Horn Yearn() => new(6);
        public static Horn Dream() => new(7);

        public readonly record struct DiscType
        {
            private readonly int _value;
            internal DiscType(int value) => _value = value;
            internal int Id => _value;

            public byte Uint8() => _value switch
            {
                0 => 0,
                1 => 1,
                2 => 2,
                3 => 3,
                4 => 4,
                5 => 5,
                6 => 6,
                7 => 7,
                8 => 8,
                9 => 9,
                10 => 10,
                11 => 11,
                12 => 12,
                13 => 13,
                14 => 14,
                15 => 15,
                16 => 16,
                17 => 17,
                18 => 18,
                19 => 19,
                20 => 20,
                _ => throw new InvalidOperationException("Invalid DiscType value."),
            };

            public string String() => _value switch
            {
                0 => "13",
                1 => "cat",
                2 => "blocks",
                3 => "chirp",
                4 => "far",
                5 => "mall",
                6 => "mellohi",
                7 => "stal",
                8 => "strad",
                9 => "ward",
                10 => "11",
                11 => "wait",
                12 => "otherside",
                13 => "pigstep",
                14 => "5",
                15 => "relic",
                16 => "creator",
                17 => "creator_music_box",
                18 => "precipice",
                19 => "tears",
                20 => "lava_chicken",
                _ => throw new InvalidOperationException("Invalid DiscType value."),
            };

            public string DisplayName() => _value switch
            {
                0 => "13",
                1 => "cat",
                2 => "blocks",
                3 => "chirp",
                4 => "far",
                5 => "mall",
                6 => "mellohi",
                7 => "stal",
                8 => "strad",
                9 => "ward",
                10 => "11",
                11 => "wait",
                12 => "otherside",
                13 => "Pigstep",
                14 => "5",
                15 => "Relic",
                16 => "Creator",
                17 => "Creator (Music Box)",
                18 => "Precipice",
                19 => "Tears",
                20 => "Lava Chicken",
                _ => throw new InvalidOperationException("Invalid DiscType value."),
            };

            public string Author() => _value switch
            {
                0 => "C418",
                1 => "C418",
                2 => "C418",
                3 => "C418",
                4 => "C418",
                5 => "C418",
                6 => "C418",
                7 => "C418",
                8 => "C418",
                9 => "C418",
                10 => "C418",
                11 => "C418",
                12 => "Lena Raine",
                13 => "Lena Raine",
                14 => "Samuel Åberg",
                15 => "Aaron Cherof",
                16 => "Lena Raine",
                17 => "Lena Raine",
                18 => "Aaron Cherof",
                19 => "Amos Roddy",
                20 => "Hyper Potions",
                _ => throw new InvalidOperationException("Invalid DiscType value."),
            };
        }

        public static DiscType Disc13() => new(0);
        public static DiscType DiscCat() => new(1);
        public static DiscType DiscBlocks() => new(2);
        public static DiscType DiscChirp() => new(3);
        public static DiscType DiscFar() => new(4);
        public static DiscType DiscMall() => new(5);
        public static DiscType DiscMellohi() => new(6);
        public static DiscType DiscStal() => new(7);
        public static DiscType DiscStrad() => new(8);
        public static DiscType DiscWard() => new(9);
        public static DiscType Disc11() => new(10);
        public static DiscType DiscWait() => new(11);
        public static DiscType DiscOtherside() => new(12);
        public static DiscType DiscPigstep() => new(13);
        public static DiscType Disc5() => new(14);
        public static DiscType DiscRelic() => new(15);
        public static DiscType DiscCreator() => new(16);
        public static DiscType DiscCreatorMusicBox() => new(17);
        public static DiscType DiscPrecipice() => new(18);
        public static DiscType DiscTears() => new(19);
        public static DiscType DiscLavaChicken() => new(20);

    }

    internal static class ItemCodec
    {
        internal static bool TryEncode(World.Item item, out string identifier, out int metadata)
        {
            switch (item)
            {
                case Item.AmethystShard _:
                    identifier = "minecraft:amethyst_shard"; metadata = 0; return true;
                case Item.Apple _:
                    identifier = "minecraft:apple"; metadata = 0; return true;
                case Item.Arrow value when value.Tip == Potion.Water():
                    identifier = "minecraft:arrow"; metadata = 0; return true;
                case Item.Arrow value when value.Tip == Potion.Mundane():
                    identifier = "minecraft:arrow"; metadata = 0; return true;
                case Item.Arrow value when value.Tip == Potion.LongMundane():
                    identifier = "minecraft:arrow"; metadata = 0; return true;
                case Item.Arrow value when value.Tip == Potion.Thick():
                    identifier = "minecraft:arrow"; metadata = 0; return true;
                case Item.Arrow value when value.Tip == Potion.Awkward():
                    identifier = "minecraft:arrow"; metadata = 0; return true;
                case Item.Arrow value when value.Tip == Potion.NightVision():
                    identifier = "minecraft:arrow"; metadata = 6; return true;
                case Item.Arrow value when value.Tip == Potion.LongNightVision():
                    identifier = "minecraft:arrow"; metadata = 7; return true;
                case Item.Arrow value when value.Tip == Potion.Invisibility():
                    identifier = "minecraft:arrow"; metadata = 8; return true;
                case Item.Arrow value when value.Tip == Potion.LongInvisibility():
                    identifier = "minecraft:arrow"; metadata = 9; return true;
                case Item.Arrow value when value.Tip == Potion.Leaping():
                    identifier = "minecraft:arrow"; metadata = 10; return true;
                case Item.Arrow value when value.Tip == Potion.LongLeaping():
                    identifier = "minecraft:arrow"; metadata = 11; return true;
                case Item.Arrow value when value.Tip == Potion.StrongLeaping():
                    identifier = "minecraft:arrow"; metadata = 12; return true;
                case Item.Arrow value when value.Tip == Potion.FireResistance():
                    identifier = "minecraft:arrow"; metadata = 13; return true;
                case Item.Arrow value when value.Tip == Potion.LongFireResistance():
                    identifier = "minecraft:arrow"; metadata = 14; return true;
                case Item.Arrow value when value.Tip == Potion.Swiftness():
                    identifier = "minecraft:arrow"; metadata = 15; return true;
                case Item.Arrow value when value.Tip == Potion.LongSwiftness():
                    identifier = "minecraft:arrow"; metadata = 16; return true;
                case Item.Arrow value when value.Tip == Potion.StrongSwiftness():
                    identifier = "minecraft:arrow"; metadata = 17; return true;
                case Item.Arrow value when value.Tip == Potion.Slowness():
                    identifier = "minecraft:arrow"; metadata = 18; return true;
                case Item.Arrow value when value.Tip == Potion.LongSlowness():
                    identifier = "minecraft:arrow"; metadata = 19; return true;
                case Item.Arrow value when value.Tip == Potion.WaterBreathing():
                    identifier = "minecraft:arrow"; metadata = 20; return true;
                case Item.Arrow value when value.Tip == Potion.LongWaterBreathing():
                    identifier = "minecraft:arrow"; metadata = 21; return true;
                case Item.Arrow value when value.Tip == Potion.Healing():
                    identifier = "minecraft:arrow"; metadata = 22; return true;
                case Item.Arrow value when value.Tip == Potion.StrongHealing():
                    identifier = "minecraft:arrow"; metadata = 23; return true;
                case Item.Arrow value when value.Tip == Potion.Harming():
                    identifier = "minecraft:arrow"; metadata = 24; return true;
                case Item.Arrow value when value.Tip == Potion.StrongHarming():
                    identifier = "minecraft:arrow"; metadata = 25; return true;
                case Item.Arrow value when value.Tip == Potion.Poison():
                    identifier = "minecraft:arrow"; metadata = 26; return true;
                case Item.Arrow value when value.Tip == Potion.LongPoison():
                    identifier = "minecraft:arrow"; metadata = 27; return true;
                case Item.Arrow value when value.Tip == Potion.StrongPoison():
                    identifier = "minecraft:arrow"; metadata = 28; return true;
                case Item.Arrow value when value.Tip == Potion.Regeneration():
                    identifier = "minecraft:arrow"; metadata = 29; return true;
                case Item.Arrow value when value.Tip == Potion.LongRegeneration():
                    identifier = "minecraft:arrow"; metadata = 30; return true;
                case Item.Arrow value when value.Tip == Potion.StrongRegeneration():
                    identifier = "minecraft:arrow"; metadata = 31; return true;
                case Item.Arrow value when value.Tip == Potion.Strength():
                    identifier = "minecraft:arrow"; metadata = 32; return true;
                case Item.Arrow value when value.Tip == Potion.LongStrength():
                    identifier = "minecraft:arrow"; metadata = 33; return true;
                case Item.Arrow value when value.Tip == Potion.StrongStrength():
                    identifier = "minecraft:arrow"; metadata = 34; return true;
                case Item.Arrow value when value.Tip == Potion.Weakness():
                    identifier = "minecraft:arrow"; metadata = 35; return true;
                case Item.Arrow value when value.Tip == Potion.LongWeakness():
                    identifier = "minecraft:arrow"; metadata = 36; return true;
                case Item.Arrow value when value.Tip == Potion.Wither():
                    identifier = "minecraft:arrow"; metadata = 37; return true;
                case Item.Arrow value when value.Tip == Potion.TurtleMaster():
                    identifier = "minecraft:arrow"; metadata = 38; return true;
                case Item.Arrow value when value.Tip == Potion.LongTurtleMaster():
                    identifier = "minecraft:arrow"; metadata = 39; return true;
                case Item.Arrow value when value.Tip == Potion.StrongTurtleMaster():
                    identifier = "minecraft:arrow"; metadata = 40; return true;
                case Item.Arrow value when value.Tip == Potion.SlowFalling():
                    identifier = "minecraft:arrow"; metadata = 41; return true;
                case Item.Arrow value when value.Tip == Potion.LongSlowFalling():
                    identifier = "minecraft:arrow"; metadata = 42; return true;
                case Item.Arrow value when value.Tip == Potion.StrongSlowness():
                    identifier = "minecraft:arrow"; metadata = 43; return true;
                case Item.Axe value when value.Tier == Item.ToolTierWood:
                    identifier = "minecraft:wooden_axe"; metadata = 0; return true;
                case Item.Axe value when value.Tier == Item.ToolTierGold:
                    identifier = "minecraft:golden_axe"; metadata = 0; return true;
                case Item.Axe value when value.Tier == Item.ToolTierStone:
                    identifier = "minecraft:stone_axe"; metadata = 0; return true;
                case Item.Axe value when value.Tier == Item.ToolTierCopper:
                    identifier = "minecraft:copper_axe"; metadata = 0; return true;
                case Item.Axe value when value.Tier == Item.ToolTierIron:
                    identifier = "minecraft:iron_axe"; metadata = 0; return true;
                case Item.Axe value when value.Tier == Item.ToolTierDiamond:
                    identifier = "minecraft:diamond_axe"; metadata = 0; return true;
                case Item.Axe value when value.Tier == Item.ToolTierNetherite:
                    identifier = "minecraft:netherite_axe"; metadata = 0; return true;
                case Item.BakedPotato _:
                    identifier = "minecraft:baked_potato"; metadata = 0; return true;
                case Item.BannerPattern value when value.Type == Item.CreeperBannerPattern():
                    identifier = "minecraft:creeper_banner_pattern"; metadata = 0; return true;
                case Item.BannerPattern value when value.Type == Item.SkullBannerPattern():
                    identifier = "minecraft:skull_banner_pattern"; metadata = 0; return true;
                case Item.BannerPattern value when value.Type == Item.FlowerBannerPattern():
                    identifier = "minecraft:flower_banner_pattern"; metadata = 0; return true;
                case Item.BannerPattern value when value.Type == Item.MojangBannerPattern():
                    identifier = "minecraft:mojang_banner_pattern"; metadata = 0; return true;
                case Item.BannerPattern value when value.Type == Item.FieldMasonedBannerPattern():
                    identifier = "minecraft:field_masoned_banner_pattern"; metadata = 0; return true;
                case Item.BannerPattern value when value.Type == Item.BordureIndentedBannerPattern():
                    identifier = "minecraft:bordure_indented_banner_pattern"; metadata = 0; return true;
                case Item.BannerPattern value when value.Type == Item.PiglinBannerPattern():
                    identifier = "minecraft:piglin_banner_pattern"; metadata = 0; return true;
                case Item.BannerPattern value when value.Type == Item.GlobeBannerPattern():
                    identifier = "minecraft:globe_banner_pattern"; metadata = 0; return true;
                case Item.BannerPattern value when value.Type == Item.FlowBannerPattern():
                    identifier = "minecraft:flow_banner_pattern"; metadata = 0; return true;
                case Item.BannerPattern value when value.Type == Item.GusterBannerPattern():
                    identifier = "minecraft:guster_banner_pattern"; metadata = 0; return true;
                case Item.Beef { Cooked: false }:
                    identifier = "minecraft:beef"; metadata = 0; return true;
                case Item.Beef { Cooked: true }:
                    identifier = "minecraft:cooked_beef"; metadata = 0; return true;
                case Item.Beetroot _:
                    identifier = "minecraft:beetroot"; metadata = 0; return true;
                case Item.BeetrootSoup _:
                    identifier = "minecraft:beetroot_soup"; metadata = 0; return true;
                case Item.BlazePowder _:
                    identifier = "minecraft:blaze_powder"; metadata = 0; return true;
                case Item.BlazeRod _:
                    identifier = "minecraft:blaze_rod"; metadata = 0; return true;
                case Item.Bone _:
                    identifier = "minecraft:bone"; metadata = 0; return true;
                case Item.BoneMeal _:
                    identifier = "minecraft:bone_meal"; metadata = 0; return true;
                case Item.Book _:
                    identifier = "minecraft:book"; metadata = 0; return true;
                case Item.BottleOfEnchanting _:
                    identifier = "minecraft:experience_bottle"; metadata = 0; return true;
                case Item.Bow _:
                    identifier = "minecraft:bow"; metadata = 0; return true;
                case Item.Bowl _:
                    identifier = "minecraft:bowl"; metadata = 0; return true;
                case Item.Bread _:
                    identifier = "minecraft:bread"; metadata = 0; return true;
                case Item.Brick _:
                    identifier = "minecraft:brick"; metadata = 0; return true;
                case Item.CarrotOnAStick _:
                    identifier = "minecraft:carrot_on_a_stick"; metadata = 0; return true;
                case Item.Charcoal _:
                    identifier = "minecraft:charcoal"; metadata = 0; return true;
                case Item.Chicken { Cooked: false }:
                    identifier = "minecraft:chicken"; metadata = 0; return true;
                case Item.Chicken { Cooked: true }:
                    identifier = "minecraft:cooked_chicken"; metadata = 0; return true;
                case Item.ClayBall _:
                    identifier = "minecraft:clay_ball"; metadata = 0; return true;
                case Item.Clock _:
                    identifier = "minecraft:clock"; metadata = 0; return true;
                case Item.Coal _:
                    identifier = "minecraft:coal"; metadata = 0; return true;
                case Item.Cod { Cooked: false }:
                    identifier = "minecraft:cod"; metadata = 0; return true;
                case Item.Cod { Cooked: true }:
                    identifier = "minecraft:cooked_cod"; metadata = 0; return true;
                case Item.Compass _:
                    identifier = "minecraft:compass"; metadata = 0; return true;
                case Item.Cookie _:
                    identifier = "minecraft:cookie"; metadata = 0; return true;
                case Item.CopperIngot _:
                    identifier = "minecraft:copper_ingot"; metadata = 0; return true;
                case Item.CopperNugget _:
                    identifier = "minecraft:copper_nugget"; metadata = 0; return true;
                case Item.Diamond _:
                    identifier = "minecraft:diamond"; metadata = 0; return true;
                case Item.DiscFragment _:
                    identifier = "minecraft:disc_fragment_5"; metadata = 0; return true;
                case Item.DragonBreath _:
                    identifier = "minecraft:dragon_breath"; metadata = 0; return true;
                case Item.DriedKelp _:
                    identifier = "minecraft:dried_kelp"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourWhite():
                    identifier = "minecraft:white_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourOrange():
                    identifier = "minecraft:orange_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourMagenta():
                    identifier = "minecraft:magenta_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourLightBlue():
                    identifier = "minecraft:light_blue_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourYellow():
                    identifier = "minecraft:yellow_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourLime():
                    identifier = "minecraft:lime_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourPink():
                    identifier = "minecraft:pink_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourGrey():
                    identifier = "minecraft:gray_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourLightGrey():
                    identifier = "minecraft:light_gray_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourCyan():
                    identifier = "minecraft:cyan_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourPurple():
                    identifier = "minecraft:purple_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourBlue():
                    identifier = "minecraft:blue_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourBrown():
                    identifier = "minecraft:brown_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourGreen():
                    identifier = "minecraft:green_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourRed():
                    identifier = "minecraft:red_dye"; metadata = 0; return true;
                case Item.Dye value when value.Colour == Item.ColourBlack():
                    identifier = "minecraft:black_dye"; metadata = 0; return true;
                case Item.EchoShard _:
                    identifier = "minecraft:echo_shard"; metadata = 0; return true;
                case Item.Egg _:
                    identifier = "minecraft:egg"; metadata = 0; return true;
                case Item.Elytra _:
                    identifier = "minecraft:elytra"; metadata = 0; return true;
                case Item.Emerald _:
                    identifier = "minecraft:emerald"; metadata = 0; return true;
                case Item.EnchantedApple _:
                    identifier = "minecraft:enchanted_golden_apple"; metadata = 0; return true;
                case Item.EnchantedBook _:
                    identifier = "minecraft:enchanted_book"; metadata = 0; return true;
                case Item.EnderEye _:
                    identifier = "minecraft:ender_eye"; metadata = 0; return true;
                case Item.EnderPearl _:
                    identifier = "minecraft:ender_pearl"; metadata = 0; return true;
                case Item.Feather _:
                    identifier = "minecraft:feather"; metadata = 0; return true;
                case Item.FermentedSpiderEye _:
                    identifier = "minecraft:fermented_spider_eye"; metadata = 0; return true;
                case Item.FireCharge _:
                    identifier = "minecraft:fire_charge"; metadata = 0; return true;
                case Item.Flint _:
                    identifier = "minecraft:flint"; metadata = 0; return true;
                case Item.FlintAndSteel _:
                    identifier = "minecraft:flint_and_steel"; metadata = 0; return true;
                case Item.GhastTear _:
                    identifier = "minecraft:ghast_tear"; metadata = 0; return true;
                case Item.GlassBottle _:
                    identifier = "minecraft:glass_bottle"; metadata = 0; return true;
                case Item.GlisteringMelonSlice _:
                    identifier = "minecraft:glistering_melon_slice"; metadata = 0; return true;
                case Item.GlowstoneDust _:
                    identifier = "minecraft:glowstone_dust"; metadata = 0; return true;
                case Item.GoatHorn value when value.Type == Sound.Ponder():
                    identifier = "minecraft:goat_horn"; metadata = 0; return true;
                case Item.GoatHorn value when value.Type == Sound.Sing():
                    identifier = "minecraft:goat_horn"; metadata = 1; return true;
                case Item.GoatHorn value when value.Type == Sound.Seek():
                    identifier = "minecraft:goat_horn"; metadata = 2; return true;
                case Item.GoatHorn value when value.Type == Sound.Feel():
                    identifier = "minecraft:goat_horn"; metadata = 3; return true;
                case Item.GoatHorn value when value.Type == Sound.Admire():
                    identifier = "minecraft:goat_horn"; metadata = 4; return true;
                case Item.GoatHorn value when value.Type == Sound.Call():
                    identifier = "minecraft:goat_horn"; metadata = 5; return true;
                case Item.GoatHorn value when value.Type == Sound.Yearn():
                    identifier = "minecraft:goat_horn"; metadata = 6; return true;
                case Item.GoatHorn value when value.Type == Sound.Dream():
                    identifier = "minecraft:goat_horn"; metadata = 7; return true;
                case Item.GoldIngot _:
                    identifier = "minecraft:gold_ingot"; metadata = 0; return true;
                case Item.GoldNugget _:
                    identifier = "minecraft:gold_nugget"; metadata = 0; return true;
                case Item.GoldenApple _:
                    identifier = "minecraft:golden_apple"; metadata = 0; return true;
                case Item.GoldenCarrot _:
                    identifier = "minecraft:golden_carrot"; metadata = 0; return true;
                case Item.Gunpowder _:
                    identifier = "minecraft:gunpowder"; metadata = 0; return true;
                case Item.HeartOfTheSea _:
                    identifier = "minecraft:heart_of_the_sea"; metadata = 0; return true;
                case Item.Hoe value when value.Tier == Item.ToolTierWood:
                    identifier = "minecraft:wooden_hoe"; metadata = 0; return true;
                case Item.Hoe value when value.Tier == Item.ToolTierGold:
                    identifier = "minecraft:golden_hoe"; metadata = 0; return true;
                case Item.Hoe value when value.Tier == Item.ToolTierStone:
                    identifier = "minecraft:stone_hoe"; metadata = 0; return true;
                case Item.Hoe value when value.Tier == Item.ToolTierCopper:
                    identifier = "minecraft:copper_hoe"; metadata = 0; return true;
                case Item.Hoe value when value.Tier == Item.ToolTierIron:
                    identifier = "minecraft:iron_hoe"; metadata = 0; return true;
                case Item.Hoe value when value.Tier == Item.ToolTierDiamond:
                    identifier = "minecraft:diamond_hoe"; metadata = 0; return true;
                case Item.Hoe value when value.Tier == Item.ToolTierNetherite:
                    identifier = "minecraft:netherite_hoe"; metadata = 0; return true;
                case Item.HoneyBottle _:
                    identifier = "minecraft:honey_bottle"; metadata = 0; return true;
                case Item.Honeycomb _:
                    identifier = "minecraft:honeycomb"; metadata = 0; return true;
                case Item.InkSac { Glowing: false }:
                    identifier = "minecraft:ink_sac"; metadata = 0; return true;
                case Item.InkSac { Glowing: true }:
                    identifier = "minecraft:glow_ink_sac"; metadata = 0; return true;
                case Item.IronIngot _:
                    identifier = "minecraft:iron_ingot"; metadata = 0; return true;
                case Item.IronNugget _:
                    identifier = "minecraft:iron_nugget"; metadata = 0; return true;
                case Item.LapisLazuli _:
                    identifier = "minecraft:lapis_lazuli"; metadata = 0; return true;
                case Item.Leather _:
                    identifier = "minecraft:leather"; metadata = 0; return true;
                case Item.LingeringPotion value when value.Type == Potion.Water():
                    identifier = "minecraft:lingering_potion"; metadata = 0; return true;
                case Item.LingeringPotion value when value.Type == Potion.Mundane():
                    identifier = "minecraft:lingering_potion"; metadata = 1; return true;
                case Item.LingeringPotion value when value.Type == Potion.LongMundane():
                    identifier = "minecraft:lingering_potion"; metadata = 2; return true;
                case Item.LingeringPotion value when value.Type == Potion.Thick():
                    identifier = "minecraft:lingering_potion"; metadata = 3; return true;
                case Item.LingeringPotion value when value.Type == Potion.Awkward():
                    identifier = "minecraft:lingering_potion"; metadata = 4; return true;
                case Item.LingeringPotion value when value.Type == Potion.NightVision():
                    identifier = "minecraft:lingering_potion"; metadata = 5; return true;
                case Item.LingeringPotion value when value.Type == Potion.LongNightVision():
                    identifier = "minecraft:lingering_potion"; metadata = 6; return true;
                case Item.LingeringPotion value when value.Type == Potion.Invisibility():
                    identifier = "minecraft:lingering_potion"; metadata = 7; return true;
                case Item.LingeringPotion value when value.Type == Potion.LongInvisibility():
                    identifier = "minecraft:lingering_potion"; metadata = 8; return true;
                case Item.LingeringPotion value when value.Type == Potion.Leaping():
                    identifier = "minecraft:lingering_potion"; metadata = 9; return true;
                case Item.LingeringPotion value when value.Type == Potion.LongLeaping():
                    identifier = "minecraft:lingering_potion"; metadata = 10; return true;
                case Item.LingeringPotion value when value.Type == Potion.StrongLeaping():
                    identifier = "minecraft:lingering_potion"; metadata = 11; return true;
                case Item.LingeringPotion value when value.Type == Potion.FireResistance():
                    identifier = "minecraft:lingering_potion"; metadata = 12; return true;
                case Item.LingeringPotion value when value.Type == Potion.LongFireResistance():
                    identifier = "minecraft:lingering_potion"; metadata = 13; return true;
                case Item.LingeringPotion value when value.Type == Potion.Swiftness():
                    identifier = "minecraft:lingering_potion"; metadata = 14; return true;
                case Item.LingeringPotion value when value.Type == Potion.LongSwiftness():
                    identifier = "minecraft:lingering_potion"; metadata = 15; return true;
                case Item.LingeringPotion value when value.Type == Potion.StrongSwiftness():
                    identifier = "minecraft:lingering_potion"; metadata = 16; return true;
                case Item.LingeringPotion value when value.Type == Potion.Slowness():
                    identifier = "minecraft:lingering_potion"; metadata = 17; return true;
                case Item.LingeringPotion value when value.Type == Potion.LongSlowness():
                    identifier = "minecraft:lingering_potion"; metadata = 18; return true;
                case Item.LingeringPotion value when value.Type == Potion.WaterBreathing():
                    identifier = "minecraft:lingering_potion"; metadata = 19; return true;
                case Item.LingeringPotion value when value.Type == Potion.LongWaterBreathing():
                    identifier = "minecraft:lingering_potion"; metadata = 20; return true;
                case Item.LingeringPotion value when value.Type == Potion.Healing():
                    identifier = "minecraft:lingering_potion"; metadata = 21; return true;
                case Item.LingeringPotion value when value.Type == Potion.StrongHealing():
                    identifier = "minecraft:lingering_potion"; metadata = 22; return true;
                case Item.LingeringPotion value when value.Type == Potion.Harming():
                    identifier = "minecraft:lingering_potion"; metadata = 23; return true;
                case Item.LingeringPotion value when value.Type == Potion.StrongHarming():
                    identifier = "minecraft:lingering_potion"; metadata = 24; return true;
                case Item.LingeringPotion value when value.Type == Potion.Poison():
                    identifier = "minecraft:lingering_potion"; metadata = 25; return true;
                case Item.LingeringPotion value when value.Type == Potion.LongPoison():
                    identifier = "minecraft:lingering_potion"; metadata = 26; return true;
                case Item.LingeringPotion value when value.Type == Potion.StrongPoison():
                    identifier = "minecraft:lingering_potion"; metadata = 27; return true;
                case Item.LingeringPotion value when value.Type == Potion.Regeneration():
                    identifier = "minecraft:lingering_potion"; metadata = 28; return true;
                case Item.LingeringPotion value when value.Type == Potion.LongRegeneration():
                    identifier = "minecraft:lingering_potion"; metadata = 29; return true;
                case Item.LingeringPotion value when value.Type == Potion.StrongRegeneration():
                    identifier = "minecraft:lingering_potion"; metadata = 30; return true;
                case Item.LingeringPotion value when value.Type == Potion.Strength():
                    identifier = "minecraft:lingering_potion"; metadata = 31; return true;
                case Item.LingeringPotion value when value.Type == Potion.LongStrength():
                    identifier = "minecraft:lingering_potion"; metadata = 32; return true;
                case Item.LingeringPotion value when value.Type == Potion.StrongStrength():
                    identifier = "minecraft:lingering_potion"; metadata = 33; return true;
                case Item.LingeringPotion value when value.Type == Potion.Weakness():
                    identifier = "minecraft:lingering_potion"; metadata = 34; return true;
                case Item.LingeringPotion value when value.Type == Potion.LongWeakness():
                    identifier = "minecraft:lingering_potion"; metadata = 35; return true;
                case Item.LingeringPotion value when value.Type == Potion.Wither():
                    identifier = "minecraft:lingering_potion"; metadata = 36; return true;
                case Item.LingeringPotion value when value.Type == Potion.TurtleMaster():
                    identifier = "minecraft:lingering_potion"; metadata = 37; return true;
                case Item.LingeringPotion value when value.Type == Potion.LongTurtleMaster():
                    identifier = "minecraft:lingering_potion"; metadata = 38; return true;
                case Item.LingeringPotion value when value.Type == Potion.StrongTurtleMaster():
                    identifier = "minecraft:lingering_potion"; metadata = 39; return true;
                case Item.LingeringPotion value when value.Type == Potion.SlowFalling():
                    identifier = "minecraft:lingering_potion"; metadata = 40; return true;
                case Item.LingeringPotion value when value.Type == Potion.LongSlowFalling():
                    identifier = "minecraft:lingering_potion"; metadata = 41; return true;
                case Item.LingeringPotion value when value.Type == Potion.StrongSlowness():
                    identifier = "minecraft:lingering_potion"; metadata = 42; return true;
                case Item.MagmaCream _:
                    identifier = "minecraft:magma_cream"; metadata = 0; return true;
                case Item.MelonSlice _:
                    identifier = "minecraft:melon_slice"; metadata = 0; return true;
                case Item.MushroomStew _:
                    identifier = "minecraft:mushroom_stew"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.Disc13():
                    identifier = "minecraft:music_disc_13"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscCat():
                    identifier = "minecraft:music_disc_cat"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscBlocks():
                    identifier = "minecraft:music_disc_blocks"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscChirp():
                    identifier = "minecraft:music_disc_chirp"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscFar():
                    identifier = "minecraft:music_disc_far"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscMall():
                    identifier = "minecraft:music_disc_mall"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscMellohi():
                    identifier = "minecraft:music_disc_mellohi"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscStal():
                    identifier = "minecraft:music_disc_stal"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscStrad():
                    identifier = "minecraft:music_disc_strad"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscWard():
                    identifier = "minecraft:music_disc_ward"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.Disc11():
                    identifier = "minecraft:music_disc_11"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscWait():
                    identifier = "minecraft:music_disc_wait"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscOtherside():
                    identifier = "minecraft:music_disc_otherside"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscPigstep():
                    identifier = "minecraft:music_disc_pigstep"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.Disc5():
                    identifier = "minecraft:music_disc_5"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscRelic():
                    identifier = "minecraft:music_disc_relic"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscCreator():
                    identifier = "minecraft:music_disc_creator"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscCreatorMusicBox():
                    identifier = "minecraft:music_disc_creator_music_box"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscPrecipice():
                    identifier = "minecraft:music_disc_precipice"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscTears():
                    identifier = "minecraft:music_disc_tears"; metadata = 0; return true;
                case Item.MusicDisc value when value.DiscType == Sound.DiscLavaChicken():
                    identifier = "minecraft:music_disc_lava_chicken"; metadata = 0; return true;
                case Item.Mutton { Cooked: false }:
                    identifier = "minecraft:mutton"; metadata = 0; return true;
                case Item.Mutton { Cooked: true }:
                    identifier = "minecraft:cooked_mutton"; metadata = 0; return true;
                case Item.NautilusShell _:
                    identifier = "minecraft:nautilus_shell"; metadata = 0; return true;
                case Item.NetherBrick _:
                    identifier = "minecraft:netherbrick"; metadata = 0; return true;
                case Item.NetherQuartz _:
                    identifier = "minecraft:quartz"; metadata = 0; return true;
                case Item.NetherStar _:
                    identifier = "minecraft:nether_star"; metadata = 0; return true;
                case Item.NetheriteIngot _:
                    identifier = "minecraft:netherite_ingot"; metadata = 0; return true;
                case Item.NetheriteScrap _:
                    identifier = "minecraft:netherite_scrap"; metadata = 0; return true;
                case Item.Paper _:
                    identifier = "minecraft:paper"; metadata = 0; return true;
                case Item.PhantomMembrane _:
                    identifier = "minecraft:phantom_membrane"; metadata = 0; return true;
                case Item.Pickaxe value when value.Tier == Item.ToolTierWood:
                    identifier = "minecraft:wooden_pickaxe"; metadata = 0; return true;
                case Item.Pickaxe value when value.Tier == Item.ToolTierGold:
                    identifier = "minecraft:golden_pickaxe"; metadata = 0; return true;
                case Item.Pickaxe value when value.Tier == Item.ToolTierStone:
                    identifier = "minecraft:stone_pickaxe"; metadata = 0; return true;
                case Item.Pickaxe value when value.Tier == Item.ToolTierCopper:
                    identifier = "minecraft:copper_pickaxe"; metadata = 0; return true;
                case Item.Pickaxe value when value.Tier == Item.ToolTierIron:
                    identifier = "minecraft:iron_pickaxe"; metadata = 0; return true;
                case Item.Pickaxe value when value.Tier == Item.ToolTierDiamond:
                    identifier = "minecraft:diamond_pickaxe"; metadata = 0; return true;
                case Item.Pickaxe value when value.Tier == Item.ToolTierNetherite:
                    identifier = "minecraft:netherite_pickaxe"; metadata = 0; return true;
                case Item.PoisonousPotato _:
                    identifier = "minecraft:poisonous_potato"; metadata = 0; return true;
                case Item.PoppedChorusFruit _:
                    identifier = "minecraft:popped_chorus_fruit"; metadata = 0; return true;
                case Item.Porkchop { Cooked: false }:
                    identifier = "minecraft:porkchop"; metadata = 0; return true;
                case Item.Porkchop { Cooked: true }:
                    identifier = "minecraft:cooked_porkchop"; metadata = 0; return true;
                case Item.Potion value when value.Type == Potion.Water():
                    identifier = "minecraft:potion"; metadata = 0; return true;
                case Item.Potion value when value.Type == Potion.Mundane():
                    identifier = "minecraft:potion"; metadata = 1; return true;
                case Item.Potion value when value.Type == Potion.LongMundane():
                    identifier = "minecraft:potion"; metadata = 2; return true;
                case Item.Potion value when value.Type == Potion.Thick():
                    identifier = "minecraft:potion"; metadata = 3; return true;
                case Item.Potion value when value.Type == Potion.Awkward():
                    identifier = "minecraft:potion"; metadata = 4; return true;
                case Item.Potion value when value.Type == Potion.NightVision():
                    identifier = "minecraft:potion"; metadata = 5; return true;
                case Item.Potion value when value.Type == Potion.LongNightVision():
                    identifier = "minecraft:potion"; metadata = 6; return true;
                case Item.Potion value when value.Type == Potion.Invisibility():
                    identifier = "minecraft:potion"; metadata = 7; return true;
                case Item.Potion value when value.Type == Potion.LongInvisibility():
                    identifier = "minecraft:potion"; metadata = 8; return true;
                case Item.Potion value when value.Type == Potion.Leaping():
                    identifier = "minecraft:potion"; metadata = 9; return true;
                case Item.Potion value when value.Type == Potion.LongLeaping():
                    identifier = "minecraft:potion"; metadata = 10; return true;
                case Item.Potion value when value.Type == Potion.StrongLeaping():
                    identifier = "minecraft:potion"; metadata = 11; return true;
                case Item.Potion value when value.Type == Potion.FireResistance():
                    identifier = "minecraft:potion"; metadata = 12; return true;
                case Item.Potion value when value.Type == Potion.LongFireResistance():
                    identifier = "minecraft:potion"; metadata = 13; return true;
                case Item.Potion value when value.Type == Potion.Swiftness():
                    identifier = "minecraft:potion"; metadata = 14; return true;
                case Item.Potion value when value.Type == Potion.LongSwiftness():
                    identifier = "minecraft:potion"; metadata = 15; return true;
                case Item.Potion value when value.Type == Potion.StrongSwiftness():
                    identifier = "minecraft:potion"; metadata = 16; return true;
                case Item.Potion value when value.Type == Potion.Slowness():
                    identifier = "minecraft:potion"; metadata = 17; return true;
                case Item.Potion value when value.Type == Potion.LongSlowness():
                    identifier = "minecraft:potion"; metadata = 18; return true;
                case Item.Potion value when value.Type == Potion.WaterBreathing():
                    identifier = "minecraft:potion"; metadata = 19; return true;
                case Item.Potion value when value.Type == Potion.LongWaterBreathing():
                    identifier = "minecraft:potion"; metadata = 20; return true;
                case Item.Potion value when value.Type == Potion.Healing():
                    identifier = "minecraft:potion"; metadata = 21; return true;
                case Item.Potion value when value.Type == Potion.StrongHealing():
                    identifier = "minecraft:potion"; metadata = 22; return true;
                case Item.Potion value when value.Type == Potion.Harming():
                    identifier = "minecraft:potion"; metadata = 23; return true;
                case Item.Potion value when value.Type == Potion.StrongHarming():
                    identifier = "minecraft:potion"; metadata = 24; return true;
                case Item.Potion value when value.Type == Potion.Poison():
                    identifier = "minecraft:potion"; metadata = 25; return true;
                case Item.Potion value when value.Type == Potion.LongPoison():
                    identifier = "minecraft:potion"; metadata = 26; return true;
                case Item.Potion value when value.Type == Potion.StrongPoison():
                    identifier = "minecraft:potion"; metadata = 27; return true;
                case Item.Potion value when value.Type == Potion.Regeneration():
                    identifier = "minecraft:potion"; metadata = 28; return true;
                case Item.Potion value when value.Type == Potion.LongRegeneration():
                    identifier = "minecraft:potion"; metadata = 29; return true;
                case Item.Potion value when value.Type == Potion.StrongRegeneration():
                    identifier = "minecraft:potion"; metadata = 30; return true;
                case Item.Potion value when value.Type == Potion.Strength():
                    identifier = "minecraft:potion"; metadata = 31; return true;
                case Item.Potion value when value.Type == Potion.LongStrength():
                    identifier = "minecraft:potion"; metadata = 32; return true;
                case Item.Potion value when value.Type == Potion.StrongStrength():
                    identifier = "minecraft:potion"; metadata = 33; return true;
                case Item.Potion value when value.Type == Potion.Weakness():
                    identifier = "minecraft:potion"; metadata = 34; return true;
                case Item.Potion value when value.Type == Potion.LongWeakness():
                    identifier = "minecraft:potion"; metadata = 35; return true;
                case Item.Potion value when value.Type == Potion.Wither():
                    identifier = "minecraft:potion"; metadata = 36; return true;
                case Item.Potion value when value.Type == Potion.TurtleMaster():
                    identifier = "minecraft:potion"; metadata = 37; return true;
                case Item.Potion value when value.Type == Potion.LongTurtleMaster():
                    identifier = "minecraft:potion"; metadata = 38; return true;
                case Item.Potion value when value.Type == Potion.StrongTurtleMaster():
                    identifier = "minecraft:potion"; metadata = 39; return true;
                case Item.Potion value when value.Type == Potion.SlowFalling():
                    identifier = "minecraft:potion"; metadata = 40; return true;
                case Item.Potion value when value.Type == Potion.LongSlowFalling():
                    identifier = "minecraft:potion"; metadata = 41; return true;
                case Item.Potion value when value.Type == Potion.StrongSlowness():
                    identifier = "minecraft:potion"; metadata = 42; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeAngler():
                    identifier = "minecraft:angler_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeArcher():
                    identifier = "minecraft:archer_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeArmsUp():
                    identifier = "minecraft:arms_up_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeBlade():
                    identifier = "minecraft:blade_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeBrewer():
                    identifier = "minecraft:brewer_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeBurn():
                    identifier = "minecraft:burn_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeDanger():
                    identifier = "minecraft:danger_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeExplorer():
                    identifier = "minecraft:explorer_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeFriend():
                    identifier = "minecraft:friend_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeHeart():
                    identifier = "minecraft:heart_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeHeartbreak():
                    identifier = "minecraft:heartbreak_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeHowl():
                    identifier = "minecraft:howl_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeMiner():
                    identifier = "minecraft:miner_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeMourner():
                    identifier = "minecraft:mourner_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypePlenty():
                    identifier = "minecraft:plenty_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypePrize():
                    identifier = "minecraft:prize_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeSheaf():
                    identifier = "minecraft:sheaf_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeShelter():
                    identifier = "minecraft:shelter_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeSkull():
                    identifier = "minecraft:skull_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeSnort():
                    identifier = "minecraft:snort_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeFlow():
                    identifier = "minecraft:flow_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeGuster():
                    identifier = "minecraft:guster_pottery_sherd"; metadata = 0; return true;
                case Item.PotterySherd value when value.Type == Item.SherdTypeScrape():
                    identifier = "minecraft:scrape_pottery_sherd"; metadata = 0; return true;
                case Item.PrismarineCrystals _:
                    identifier = "minecraft:prismarine_crystals"; metadata = 0; return true;
                case Item.PrismarineShard _:
                    identifier = "minecraft:prismarine_shard"; metadata = 0; return true;
                case Item.Pufferfish _:
                    identifier = "minecraft:pufferfish"; metadata = 0; return true;
                case Item.PumpkinPie _:
                    identifier = "minecraft:pumpkin_pie"; metadata = 0; return true;
                case Item.Rabbit { Cooked: false }:
                    identifier = "minecraft:rabbit"; metadata = 0; return true;
                case Item.Rabbit { Cooked: true }:
                    identifier = "minecraft:cooked_rabbit"; metadata = 0; return true;
                case Item.RabbitFoot _:
                    identifier = "minecraft:rabbit_foot"; metadata = 0; return true;
                case Item.RabbitHide _:
                    identifier = "minecraft:rabbit_hide"; metadata = 0; return true;
                case Item.RabbitStew _:
                    identifier = "minecraft:rabbit_stew"; metadata = 0; return true;
                case Item.RawCopper _:
                    identifier = "minecraft:raw_copper"; metadata = 0; return true;
                case Item.RawGold _:
                    identifier = "minecraft:raw_gold"; metadata = 0; return true;
                case Item.RawIron _:
                    identifier = "minecraft:raw_iron"; metadata = 0; return true;
                case Item.RecoveryCompass _:
                    identifier = "minecraft:recovery_compass"; metadata = 0; return true;
                case Item.ResinBrick _:
                    identifier = "minecraft:resin_brick"; metadata = 0; return true;
                case Item.RottenFlesh _:
                    identifier = "minecraft:rotten_flesh"; metadata = 0; return true;
                case Item.Salmon { Cooked: false }:
                    identifier = "minecraft:salmon"; metadata = 0; return true;
                case Item.Salmon { Cooked: true }:
                    identifier = "minecraft:cooked_salmon"; metadata = 0; return true;
                case Item.Scute _:
                    identifier = "minecraft:turtle_scute"; metadata = 0; return true;
                case Item.Shears _:
                    identifier = "minecraft:shears"; metadata = 0; return true;
                case Item.Shovel value when value.Tier == Item.ToolTierWood:
                    identifier = "minecraft:wooden_shovel"; metadata = 0; return true;
                case Item.Shovel value when value.Tier == Item.ToolTierGold:
                    identifier = "minecraft:golden_shovel"; metadata = 0; return true;
                case Item.Shovel value when value.Tier == Item.ToolTierStone:
                    identifier = "minecraft:stone_shovel"; metadata = 0; return true;
                case Item.Shovel value when value.Tier == Item.ToolTierCopper:
                    identifier = "minecraft:copper_shovel"; metadata = 0; return true;
                case Item.Shovel value when value.Tier == Item.ToolTierIron:
                    identifier = "minecraft:iron_shovel"; metadata = 0; return true;
                case Item.Shovel value when value.Tier == Item.ToolTierDiamond:
                    identifier = "minecraft:diamond_shovel"; metadata = 0; return true;
                case Item.Shovel value when value.Tier == Item.ToolTierNetherite:
                    identifier = "minecraft:netherite_shovel"; metadata = 0; return true;
                case Item.ShulkerShell _:
                    identifier = "minecraft:shulker_shell"; metadata = 0; return true;
                case Item.Slimeball _:
                    identifier = "minecraft:slime_ball"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateNetheriteUpgrade():
                    identifier = "minecraft:netherite_upgrade_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateSentry():
                    identifier = "minecraft:sentry_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateVex():
                    identifier = "minecraft:vex_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateWild():
                    identifier = "minecraft:wild_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateCoast():
                    identifier = "minecraft:coast_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateDune():
                    identifier = "minecraft:dune_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateWayFinder():
                    identifier = "minecraft:wayfinder_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateRaiser():
                    identifier = "minecraft:raiser_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateShaper():
                    identifier = "minecraft:shaper_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateHost():
                    identifier = "minecraft:host_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateWard():
                    identifier = "minecraft:ward_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateSilence():
                    identifier = "minecraft:silence_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateTide():
                    identifier = "minecraft:tide_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateSnout():
                    identifier = "minecraft:snout_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateRib():
                    identifier = "minecraft:rib_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateEye():
                    identifier = "minecraft:eye_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateSpire():
                    identifier = "minecraft:spire_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateFlow():
                    identifier = "minecraft:flow_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.SmithingTemplate value when value.Template == Item.TemplateBolt():
                    identifier = "minecraft:bolt_armor_trim_smithing_template"; metadata = 0; return true;
                case Item.Snowball _:
                    identifier = "minecraft:snowball"; metadata = 0; return true;
                case Item.SpiderEye _:
                    identifier = "minecraft:spider_eye"; metadata = 0; return true;
                case Item.SplashPotion value when value.Type == Potion.Water():
                    identifier = "minecraft:splash_potion"; metadata = 0; return true;
                case Item.SplashPotion value when value.Type == Potion.Mundane():
                    identifier = "minecraft:splash_potion"; metadata = 1; return true;
                case Item.SplashPotion value when value.Type == Potion.LongMundane():
                    identifier = "minecraft:splash_potion"; metadata = 2; return true;
                case Item.SplashPotion value when value.Type == Potion.Thick():
                    identifier = "minecraft:splash_potion"; metadata = 3; return true;
                case Item.SplashPotion value when value.Type == Potion.Awkward():
                    identifier = "minecraft:splash_potion"; metadata = 4; return true;
                case Item.SplashPotion value when value.Type == Potion.NightVision():
                    identifier = "minecraft:splash_potion"; metadata = 5; return true;
                case Item.SplashPotion value when value.Type == Potion.LongNightVision():
                    identifier = "minecraft:splash_potion"; metadata = 6; return true;
                case Item.SplashPotion value when value.Type == Potion.Invisibility():
                    identifier = "minecraft:splash_potion"; metadata = 7; return true;
                case Item.SplashPotion value when value.Type == Potion.LongInvisibility():
                    identifier = "minecraft:splash_potion"; metadata = 8; return true;
                case Item.SplashPotion value when value.Type == Potion.Leaping():
                    identifier = "minecraft:splash_potion"; metadata = 9; return true;
                case Item.SplashPotion value when value.Type == Potion.LongLeaping():
                    identifier = "minecraft:splash_potion"; metadata = 10; return true;
                case Item.SplashPotion value when value.Type == Potion.StrongLeaping():
                    identifier = "minecraft:splash_potion"; metadata = 11; return true;
                case Item.SplashPotion value when value.Type == Potion.FireResistance():
                    identifier = "minecraft:splash_potion"; metadata = 12; return true;
                case Item.SplashPotion value when value.Type == Potion.LongFireResistance():
                    identifier = "minecraft:splash_potion"; metadata = 13; return true;
                case Item.SplashPotion value when value.Type == Potion.Swiftness():
                    identifier = "minecraft:splash_potion"; metadata = 14; return true;
                case Item.SplashPotion value when value.Type == Potion.LongSwiftness():
                    identifier = "minecraft:splash_potion"; metadata = 15; return true;
                case Item.SplashPotion value when value.Type == Potion.StrongSwiftness():
                    identifier = "minecraft:splash_potion"; metadata = 16; return true;
                case Item.SplashPotion value when value.Type == Potion.Slowness():
                    identifier = "minecraft:splash_potion"; metadata = 17; return true;
                case Item.SplashPotion value when value.Type == Potion.LongSlowness():
                    identifier = "minecraft:splash_potion"; metadata = 18; return true;
                case Item.SplashPotion value when value.Type == Potion.WaterBreathing():
                    identifier = "minecraft:splash_potion"; metadata = 19; return true;
                case Item.SplashPotion value when value.Type == Potion.LongWaterBreathing():
                    identifier = "minecraft:splash_potion"; metadata = 20; return true;
                case Item.SplashPotion value when value.Type == Potion.Healing():
                    identifier = "minecraft:splash_potion"; metadata = 21; return true;
                case Item.SplashPotion value when value.Type == Potion.StrongHealing():
                    identifier = "minecraft:splash_potion"; metadata = 22; return true;
                case Item.SplashPotion value when value.Type == Potion.Harming():
                    identifier = "minecraft:splash_potion"; metadata = 23; return true;
                case Item.SplashPotion value when value.Type == Potion.StrongHarming():
                    identifier = "minecraft:splash_potion"; metadata = 24; return true;
                case Item.SplashPotion value when value.Type == Potion.Poison():
                    identifier = "minecraft:splash_potion"; metadata = 25; return true;
                case Item.SplashPotion value when value.Type == Potion.LongPoison():
                    identifier = "minecraft:splash_potion"; metadata = 26; return true;
                case Item.SplashPotion value when value.Type == Potion.StrongPoison():
                    identifier = "minecraft:splash_potion"; metadata = 27; return true;
                case Item.SplashPotion value when value.Type == Potion.Regeneration():
                    identifier = "minecraft:splash_potion"; metadata = 28; return true;
                case Item.SplashPotion value when value.Type == Potion.LongRegeneration():
                    identifier = "minecraft:splash_potion"; metadata = 29; return true;
                case Item.SplashPotion value when value.Type == Potion.StrongRegeneration():
                    identifier = "minecraft:splash_potion"; metadata = 30; return true;
                case Item.SplashPotion value when value.Type == Potion.Strength():
                    identifier = "minecraft:splash_potion"; metadata = 31; return true;
                case Item.SplashPotion value when value.Type == Potion.LongStrength():
                    identifier = "minecraft:splash_potion"; metadata = 32; return true;
                case Item.SplashPotion value when value.Type == Potion.StrongStrength():
                    identifier = "minecraft:splash_potion"; metadata = 33; return true;
                case Item.SplashPotion value when value.Type == Potion.Weakness():
                    identifier = "minecraft:splash_potion"; metadata = 34; return true;
                case Item.SplashPotion value when value.Type == Potion.LongWeakness():
                    identifier = "minecraft:splash_potion"; metadata = 35; return true;
                case Item.SplashPotion value when value.Type == Potion.Wither():
                    identifier = "minecraft:splash_potion"; metadata = 36; return true;
                case Item.SplashPotion value when value.Type == Potion.TurtleMaster():
                    identifier = "minecraft:splash_potion"; metadata = 37; return true;
                case Item.SplashPotion value when value.Type == Potion.LongTurtleMaster():
                    identifier = "minecraft:splash_potion"; metadata = 38; return true;
                case Item.SplashPotion value when value.Type == Potion.StrongTurtleMaster():
                    identifier = "minecraft:splash_potion"; metadata = 39; return true;
                case Item.SplashPotion value when value.Type == Potion.SlowFalling():
                    identifier = "minecraft:splash_potion"; metadata = 40; return true;
                case Item.SplashPotion value when value.Type == Potion.LongSlowFalling():
                    identifier = "minecraft:splash_potion"; metadata = 41; return true;
                case Item.SplashPotion value when value.Type == Potion.StrongSlowness():
                    identifier = "minecraft:splash_potion"; metadata = 42; return true;
                case Item.Spyglass _:
                    identifier = "minecraft:spyglass"; metadata = 0; return true;
                case Item.Stick _:
                    identifier = "minecraft:stick"; metadata = 0; return true;
                case Item.Sugar _:
                    identifier = "minecraft:sugar"; metadata = 0; return true;
                case Item.SuspiciousStew value when value.Type == Item.NightVisionPoppyStew():
                    identifier = "minecraft:suspicious_stew"; metadata = 0; return true;
                case Item.SuspiciousStew value when value.Type == Item.JumpBoostStew():
                    identifier = "minecraft:suspicious_stew"; metadata = 1; return true;
                case Item.SuspiciousStew value when value.Type == Item.WeaknessStew():
                    identifier = "minecraft:suspicious_stew"; metadata = 2; return true;
                case Item.SuspiciousStew value when value.Type == Item.BlindnessBluetStew():
                    identifier = "minecraft:suspicious_stew"; metadata = 3; return true;
                case Item.SuspiciousStew value when value.Type == Item.PoisonStew():
                    identifier = "minecraft:suspicious_stew"; metadata = 4; return true;
                case Item.SuspiciousStew value when value.Type == Item.SaturationDandelionStew():
                    identifier = "minecraft:suspicious_stew"; metadata = 5; return true;
                case Item.SuspiciousStew value when value.Type == Item.SaturationOrchidStew():
                    identifier = "minecraft:suspicious_stew"; metadata = 6; return true;
                case Item.SuspiciousStew value when value.Type == Item.FireResistanceStew():
                    identifier = "minecraft:suspicious_stew"; metadata = 7; return true;
                case Item.SuspiciousStew value when value.Type == Item.RegenerationStew():
                    identifier = "minecraft:suspicious_stew"; metadata = 8; return true;
                case Item.SuspiciousStew value when value.Type == Item.WitherStew():
                    identifier = "minecraft:suspicious_stew"; metadata = 9; return true;
                case Item.SuspiciousStew value when value.Type == Item.NightVisionTorchflowerStew():
                    identifier = "minecraft:suspicious_stew"; metadata = 10; return true;
                case Item.SuspiciousStew value when value.Type == Item.BlindnessEyeblossomStew():
                    identifier = "minecraft:suspicious_stew"; metadata = 11; return true;
                case Item.SuspiciousStew value when value.Type == Item.NauseaStew():
                    identifier = "minecraft:suspicious_stew"; metadata = 12; return true;
                case Item.Sword value when value.Tier == Item.ToolTierWood:
                    identifier = "minecraft:wooden_sword"; metadata = 0; return true;
                case Item.Sword value when value.Tier == Item.ToolTierGold:
                    identifier = "minecraft:golden_sword"; metadata = 0; return true;
                case Item.Sword value when value.Tier == Item.ToolTierStone:
                    identifier = "minecraft:stone_sword"; metadata = 0; return true;
                case Item.Sword value when value.Tier == Item.ToolTierCopper:
                    identifier = "minecraft:copper_sword"; metadata = 0; return true;
                case Item.Sword value when value.Tier == Item.ToolTierIron:
                    identifier = "minecraft:iron_sword"; metadata = 0; return true;
                case Item.Sword value when value.Tier == Item.ToolTierDiamond:
                    identifier = "minecraft:diamond_sword"; metadata = 0; return true;
                case Item.Sword value when value.Tier == Item.ToolTierNetherite:
                    identifier = "minecraft:netherite_sword"; metadata = 0; return true;
                case Item.Totem _:
                    identifier = "minecraft:totem_of_undying"; metadata = 0; return true;
                case Item.TropicalFish _:
                    identifier = "minecraft:tropical_fish"; metadata = 0; return true;
                case Item.TurtleShell _:
                    identifier = "minecraft:turtle_helmet"; metadata = 0; return true;
                case Item.WarpedFungusOnAStick _:
                    identifier = "minecraft:warped_fungus_on_a_stick"; metadata = 0; return true;
                case Item.Wheat _:
                    identifier = "minecraft:wheat"; metadata = 0; return true;
                case EncodedItem encoded:
                    identifier = encoded.Identifier; metadata = encoded.Metadata; return true;
                default:
                    identifier = string.Empty; metadata = 0; return false;
            }
        }

        internal static World.Item Decode(string identifier, int metadata)
        {
            if (identifier == "minecraft:amethyst_shard" && metadata == 0) return new Item.AmethystShard();
            if (identifier == "minecraft:apple" && metadata == 0) return new Item.Apple();
            if (identifier == "minecraft:arrow" && metadata == 0) return new Item.Arrow(Potion.Water());
            if (identifier == "minecraft:arrow" && metadata == 0) return new Item.Arrow(Potion.Mundane());
            if (identifier == "minecraft:arrow" && metadata == 0) return new Item.Arrow(Potion.LongMundane());
            if (identifier == "minecraft:arrow" && metadata == 0) return new Item.Arrow(Potion.Thick());
            if (identifier == "minecraft:arrow" && metadata == 0) return new Item.Arrow(Potion.Awkward());
            if (identifier == "minecraft:arrow" && metadata == 6) return new Item.Arrow(Potion.NightVision());
            if (identifier == "minecraft:arrow" && metadata == 7) return new Item.Arrow(Potion.LongNightVision());
            if (identifier == "minecraft:arrow" && metadata == 8) return new Item.Arrow(Potion.Invisibility());
            if (identifier == "minecraft:arrow" && metadata == 9) return new Item.Arrow(Potion.LongInvisibility());
            if (identifier == "minecraft:arrow" && metadata == 10) return new Item.Arrow(Potion.Leaping());
            if (identifier == "minecraft:arrow" && metadata == 11) return new Item.Arrow(Potion.LongLeaping());
            if (identifier == "minecraft:arrow" && metadata == 12) return new Item.Arrow(Potion.StrongLeaping());
            if (identifier == "minecraft:arrow" && metadata == 13) return new Item.Arrow(Potion.FireResistance());
            if (identifier == "minecraft:arrow" && metadata == 14) return new Item.Arrow(Potion.LongFireResistance());
            if (identifier == "minecraft:arrow" && metadata == 15) return new Item.Arrow(Potion.Swiftness());
            if (identifier == "minecraft:arrow" && metadata == 16) return new Item.Arrow(Potion.LongSwiftness());
            if (identifier == "minecraft:arrow" && metadata == 17) return new Item.Arrow(Potion.StrongSwiftness());
            if (identifier == "minecraft:arrow" && metadata == 18) return new Item.Arrow(Potion.Slowness());
            if (identifier == "minecraft:arrow" && metadata == 19) return new Item.Arrow(Potion.LongSlowness());
            if (identifier == "minecraft:arrow" && metadata == 20) return new Item.Arrow(Potion.WaterBreathing());
            if (identifier == "minecraft:arrow" && metadata == 21) return new Item.Arrow(Potion.LongWaterBreathing());
            if (identifier == "minecraft:arrow" && metadata == 22) return new Item.Arrow(Potion.Healing());
            if (identifier == "minecraft:arrow" && metadata == 23) return new Item.Arrow(Potion.StrongHealing());
            if (identifier == "minecraft:arrow" && metadata == 24) return new Item.Arrow(Potion.Harming());
            if (identifier == "minecraft:arrow" && metadata == 25) return new Item.Arrow(Potion.StrongHarming());
            if (identifier == "minecraft:arrow" && metadata == 26) return new Item.Arrow(Potion.Poison());
            if (identifier == "minecraft:arrow" && metadata == 27) return new Item.Arrow(Potion.LongPoison());
            if (identifier == "minecraft:arrow" && metadata == 28) return new Item.Arrow(Potion.StrongPoison());
            if (identifier == "minecraft:arrow" && metadata == 29) return new Item.Arrow(Potion.Regeneration());
            if (identifier == "minecraft:arrow" && metadata == 30) return new Item.Arrow(Potion.LongRegeneration());
            if (identifier == "minecraft:arrow" && metadata == 31) return new Item.Arrow(Potion.StrongRegeneration());
            if (identifier == "minecraft:arrow" && metadata == 32) return new Item.Arrow(Potion.Strength());
            if (identifier == "minecraft:arrow" && metadata == 33) return new Item.Arrow(Potion.LongStrength());
            if (identifier == "minecraft:arrow" && metadata == 34) return new Item.Arrow(Potion.StrongStrength());
            if (identifier == "minecraft:arrow" && metadata == 35) return new Item.Arrow(Potion.Weakness());
            if (identifier == "minecraft:arrow" && metadata == 36) return new Item.Arrow(Potion.LongWeakness());
            if (identifier == "minecraft:arrow" && metadata == 37) return new Item.Arrow(Potion.Wither());
            if (identifier == "minecraft:arrow" && metadata == 38) return new Item.Arrow(Potion.TurtleMaster());
            if (identifier == "minecraft:arrow" && metadata == 39) return new Item.Arrow(Potion.LongTurtleMaster());
            if (identifier == "minecraft:arrow" && metadata == 40) return new Item.Arrow(Potion.StrongTurtleMaster());
            if (identifier == "minecraft:arrow" && metadata == 41) return new Item.Arrow(Potion.SlowFalling());
            if (identifier == "minecraft:arrow" && metadata == 42) return new Item.Arrow(Potion.LongSlowFalling());
            if (identifier == "minecraft:arrow" && metadata == 43) return new Item.Arrow(Potion.StrongSlowness());
            if (identifier == "minecraft:wooden_axe" && metadata == 0) return new Item.Axe(Item.ToolTierWood);
            if (identifier == "minecraft:golden_axe" && metadata == 0) return new Item.Axe(Item.ToolTierGold);
            if (identifier == "minecraft:stone_axe" && metadata == 0) return new Item.Axe(Item.ToolTierStone);
            if (identifier == "minecraft:copper_axe" && metadata == 0) return new Item.Axe(Item.ToolTierCopper);
            if (identifier == "minecraft:iron_axe" && metadata == 0) return new Item.Axe(Item.ToolTierIron);
            if (identifier == "minecraft:diamond_axe" && metadata == 0) return new Item.Axe(Item.ToolTierDiamond);
            if (identifier == "minecraft:netherite_axe" && metadata == 0) return new Item.Axe(Item.ToolTierNetherite);
            if (identifier == "minecraft:baked_potato" && metadata == 0) return new Item.BakedPotato();
            if (identifier == "minecraft:creeper_banner_pattern" && metadata == 0) return new Item.BannerPattern(Item.CreeperBannerPattern());
            if (identifier == "minecraft:skull_banner_pattern" && metadata == 0) return new Item.BannerPattern(Item.SkullBannerPattern());
            if (identifier == "minecraft:flower_banner_pattern" && metadata == 0) return new Item.BannerPattern(Item.FlowerBannerPattern());
            if (identifier == "minecraft:mojang_banner_pattern" && metadata == 0) return new Item.BannerPattern(Item.MojangBannerPattern());
            if (identifier == "minecraft:field_masoned_banner_pattern" && metadata == 0) return new Item.BannerPattern(Item.FieldMasonedBannerPattern());
            if (identifier == "minecraft:bordure_indented_banner_pattern" && metadata == 0) return new Item.BannerPattern(Item.BordureIndentedBannerPattern());
            if (identifier == "minecraft:piglin_banner_pattern" && metadata == 0) return new Item.BannerPattern(Item.PiglinBannerPattern());
            if (identifier == "minecraft:globe_banner_pattern" && metadata == 0) return new Item.BannerPattern(Item.GlobeBannerPattern());
            if (identifier == "minecraft:flow_banner_pattern" && metadata == 0) return new Item.BannerPattern(Item.FlowBannerPattern());
            if (identifier == "minecraft:guster_banner_pattern" && metadata == 0) return new Item.BannerPattern(Item.GusterBannerPattern());
            if (identifier == "minecraft:beef" && metadata == 0) return new Item.Beef(false);
            if (identifier == "minecraft:cooked_beef" && metadata == 0) return new Item.Beef(true);
            if (identifier == "minecraft:beetroot" && metadata == 0) return new Item.Beetroot();
            if (identifier == "minecraft:beetroot_soup" && metadata == 0) return new Item.BeetrootSoup();
            if (identifier == "minecraft:blaze_powder" && metadata == 0) return new Item.BlazePowder();
            if (identifier == "minecraft:blaze_rod" && metadata == 0) return new Item.BlazeRod();
            if (identifier == "minecraft:bone" && metadata == 0) return new Item.Bone();
            if (identifier == "minecraft:bone_meal" && metadata == 0) return new Item.BoneMeal();
            if (identifier == "minecraft:book" && metadata == 0) return new Item.Book();
            if (identifier == "minecraft:experience_bottle" && metadata == 0) return new Item.BottleOfEnchanting();
            if (identifier == "minecraft:bow" && metadata == 0) return new Item.Bow();
            if (identifier == "minecraft:bowl" && metadata == 0) return new Item.Bowl();
            if (identifier == "minecraft:bread" && metadata == 0) return new Item.Bread();
            if (identifier == "minecraft:brick" && metadata == 0) return new Item.Brick();
            if (identifier == "minecraft:carrot_on_a_stick" && metadata == 0) return new Item.CarrotOnAStick();
            if (identifier == "minecraft:charcoal" && metadata == 0) return new Item.Charcoal();
            if (identifier == "minecraft:chicken" && metadata == 0) return new Item.Chicken(false);
            if (identifier == "minecraft:cooked_chicken" && metadata == 0) return new Item.Chicken(true);
            if (identifier == "minecraft:clay_ball" && metadata == 0) return new Item.ClayBall();
            if (identifier == "minecraft:clock" && metadata == 0) return new Item.Clock();
            if (identifier == "minecraft:coal" && metadata == 0) return new Item.Coal();
            if (identifier == "minecraft:cod" && metadata == 0) return new Item.Cod(false);
            if (identifier == "minecraft:cooked_cod" && metadata == 0) return new Item.Cod(true);
            if (identifier == "minecraft:compass" && metadata == 0) return new Item.Compass();
            if (identifier == "minecraft:cookie" && metadata == 0) return new Item.Cookie();
            if (identifier == "minecraft:copper_ingot" && metadata == 0) return new Item.CopperIngot();
            if (identifier == "minecraft:copper_nugget" && metadata == 0) return new Item.CopperNugget();
            if (identifier == "minecraft:diamond" && metadata == 0) return new Item.Diamond();
            if (identifier == "minecraft:disc_fragment_5" && metadata == 0) return new Item.DiscFragment();
            if (identifier == "minecraft:dragon_breath" && metadata == 0) return new Item.DragonBreath();
            if (identifier == "minecraft:dried_kelp" && metadata == 0) return new Item.DriedKelp();
            if (identifier == "minecraft:white_dye" && metadata == 0) return new Item.Dye(Item.ColourWhite());
            if (identifier == "minecraft:orange_dye" && metadata == 0) return new Item.Dye(Item.ColourOrange());
            if (identifier == "minecraft:magenta_dye" && metadata == 0) return new Item.Dye(Item.ColourMagenta());
            if (identifier == "minecraft:light_blue_dye" && metadata == 0) return new Item.Dye(Item.ColourLightBlue());
            if (identifier == "minecraft:yellow_dye" && metadata == 0) return new Item.Dye(Item.ColourYellow());
            if (identifier == "minecraft:lime_dye" && metadata == 0) return new Item.Dye(Item.ColourLime());
            if (identifier == "minecraft:pink_dye" && metadata == 0) return new Item.Dye(Item.ColourPink());
            if (identifier == "minecraft:gray_dye" && metadata == 0) return new Item.Dye(Item.ColourGrey());
            if (identifier == "minecraft:light_gray_dye" && metadata == 0) return new Item.Dye(Item.ColourLightGrey());
            if (identifier == "minecraft:cyan_dye" && metadata == 0) return new Item.Dye(Item.ColourCyan());
            if (identifier == "minecraft:purple_dye" && metadata == 0) return new Item.Dye(Item.ColourPurple());
            if (identifier == "minecraft:blue_dye" && metadata == 0) return new Item.Dye(Item.ColourBlue());
            if (identifier == "minecraft:brown_dye" && metadata == 0) return new Item.Dye(Item.ColourBrown());
            if (identifier == "minecraft:green_dye" && metadata == 0) return new Item.Dye(Item.ColourGreen());
            if (identifier == "minecraft:red_dye" && metadata == 0) return new Item.Dye(Item.ColourRed());
            if (identifier == "minecraft:black_dye" && metadata == 0) return new Item.Dye(Item.ColourBlack());
            if (identifier == "minecraft:echo_shard" && metadata == 0) return new Item.EchoShard();
            if (identifier == "minecraft:egg" && metadata == 0) return new Item.Egg();
            if (identifier == "minecraft:elytra" && metadata == 0) return new Item.Elytra();
            if (identifier == "minecraft:emerald" && metadata == 0) return new Item.Emerald();
            if (identifier == "minecraft:enchanted_golden_apple" && metadata == 0) return new Item.EnchantedApple();
            if (identifier == "minecraft:enchanted_book" && metadata == 0) return new Item.EnchantedBook();
            if (identifier == "minecraft:ender_eye" && metadata == 0) return new Item.EnderEye();
            if (identifier == "minecraft:ender_pearl" && metadata == 0) return new Item.EnderPearl();
            if (identifier == "minecraft:feather" && metadata == 0) return new Item.Feather();
            if (identifier == "minecraft:fermented_spider_eye" && metadata == 0) return new Item.FermentedSpiderEye();
            if (identifier == "minecraft:fire_charge" && metadata == 0) return new Item.FireCharge();
            if (identifier == "minecraft:flint" && metadata == 0) return new Item.Flint();
            if (identifier == "minecraft:flint_and_steel" && metadata == 0) return new Item.FlintAndSteel();
            if (identifier == "minecraft:ghast_tear" && metadata == 0) return new Item.GhastTear();
            if (identifier == "minecraft:glass_bottle" && metadata == 0) return new Item.GlassBottle();
            if (identifier == "minecraft:glistering_melon_slice" && metadata == 0) return new Item.GlisteringMelonSlice();
            if (identifier == "minecraft:glowstone_dust" && metadata == 0) return new Item.GlowstoneDust();
            if (identifier == "minecraft:goat_horn" && metadata == 0) return new Item.GoatHorn(Sound.Ponder());
            if (identifier == "minecraft:goat_horn" && metadata == 1) return new Item.GoatHorn(Sound.Sing());
            if (identifier == "minecraft:goat_horn" && metadata == 2) return new Item.GoatHorn(Sound.Seek());
            if (identifier == "minecraft:goat_horn" && metadata == 3) return new Item.GoatHorn(Sound.Feel());
            if (identifier == "minecraft:goat_horn" && metadata == 4) return new Item.GoatHorn(Sound.Admire());
            if (identifier == "minecraft:goat_horn" && metadata == 5) return new Item.GoatHorn(Sound.Call());
            if (identifier == "minecraft:goat_horn" && metadata == 6) return new Item.GoatHorn(Sound.Yearn());
            if (identifier == "minecraft:goat_horn" && metadata == 7) return new Item.GoatHorn(Sound.Dream());
            if (identifier == "minecraft:gold_ingot" && metadata == 0) return new Item.GoldIngot();
            if (identifier == "minecraft:gold_nugget" && metadata == 0) return new Item.GoldNugget();
            if (identifier == "minecraft:golden_apple" && metadata == 0) return new Item.GoldenApple();
            if (identifier == "minecraft:golden_carrot" && metadata == 0) return new Item.GoldenCarrot();
            if (identifier == "minecraft:gunpowder" && metadata == 0) return new Item.Gunpowder();
            if (identifier == "minecraft:heart_of_the_sea" && metadata == 0) return new Item.HeartOfTheSea();
            if (identifier == "minecraft:wooden_hoe" && metadata == 0) return new Item.Hoe(Item.ToolTierWood);
            if (identifier == "minecraft:golden_hoe" && metadata == 0) return new Item.Hoe(Item.ToolTierGold);
            if (identifier == "minecraft:stone_hoe" && metadata == 0) return new Item.Hoe(Item.ToolTierStone);
            if (identifier == "minecraft:copper_hoe" && metadata == 0) return new Item.Hoe(Item.ToolTierCopper);
            if (identifier == "minecraft:iron_hoe" && metadata == 0) return new Item.Hoe(Item.ToolTierIron);
            if (identifier == "minecraft:diamond_hoe" && metadata == 0) return new Item.Hoe(Item.ToolTierDiamond);
            if (identifier == "minecraft:netherite_hoe" && metadata == 0) return new Item.Hoe(Item.ToolTierNetherite);
            if (identifier == "minecraft:honey_bottle" && metadata == 0) return new Item.HoneyBottle();
            if (identifier == "minecraft:honeycomb" && metadata == 0) return new Item.Honeycomb();
            if (identifier == "minecraft:ink_sac" && metadata == 0) return new Item.InkSac(false);
            if (identifier == "minecraft:glow_ink_sac" && metadata == 0) return new Item.InkSac(true);
            if (identifier == "minecraft:iron_ingot" && metadata == 0) return new Item.IronIngot();
            if (identifier == "minecraft:iron_nugget" && metadata == 0) return new Item.IronNugget();
            if (identifier == "minecraft:lapis_lazuli" && metadata == 0) return new Item.LapisLazuli();
            if (identifier == "minecraft:leather" && metadata == 0) return new Item.Leather();
            if (identifier == "minecraft:lingering_potion" && metadata == 0) return new Item.LingeringPotion(Potion.Water());
            if (identifier == "minecraft:lingering_potion" && metadata == 1) return new Item.LingeringPotion(Potion.Mundane());
            if (identifier == "minecraft:lingering_potion" && metadata == 2) return new Item.LingeringPotion(Potion.LongMundane());
            if (identifier == "minecraft:lingering_potion" && metadata == 3) return new Item.LingeringPotion(Potion.Thick());
            if (identifier == "minecraft:lingering_potion" && metadata == 4) return new Item.LingeringPotion(Potion.Awkward());
            if (identifier == "minecraft:lingering_potion" && metadata == 5) return new Item.LingeringPotion(Potion.NightVision());
            if (identifier == "minecraft:lingering_potion" && metadata == 6) return new Item.LingeringPotion(Potion.LongNightVision());
            if (identifier == "minecraft:lingering_potion" && metadata == 7) return new Item.LingeringPotion(Potion.Invisibility());
            if (identifier == "minecraft:lingering_potion" && metadata == 8) return new Item.LingeringPotion(Potion.LongInvisibility());
            if (identifier == "minecraft:lingering_potion" && metadata == 9) return new Item.LingeringPotion(Potion.Leaping());
            if (identifier == "minecraft:lingering_potion" && metadata == 10) return new Item.LingeringPotion(Potion.LongLeaping());
            if (identifier == "minecraft:lingering_potion" && metadata == 11) return new Item.LingeringPotion(Potion.StrongLeaping());
            if (identifier == "minecraft:lingering_potion" && metadata == 12) return new Item.LingeringPotion(Potion.FireResistance());
            if (identifier == "minecraft:lingering_potion" && metadata == 13) return new Item.LingeringPotion(Potion.LongFireResistance());
            if (identifier == "minecraft:lingering_potion" && metadata == 14) return new Item.LingeringPotion(Potion.Swiftness());
            if (identifier == "minecraft:lingering_potion" && metadata == 15) return new Item.LingeringPotion(Potion.LongSwiftness());
            if (identifier == "minecraft:lingering_potion" && metadata == 16) return new Item.LingeringPotion(Potion.StrongSwiftness());
            if (identifier == "minecraft:lingering_potion" && metadata == 17) return new Item.LingeringPotion(Potion.Slowness());
            if (identifier == "minecraft:lingering_potion" && metadata == 18) return new Item.LingeringPotion(Potion.LongSlowness());
            if (identifier == "minecraft:lingering_potion" && metadata == 19) return new Item.LingeringPotion(Potion.WaterBreathing());
            if (identifier == "minecraft:lingering_potion" && metadata == 20) return new Item.LingeringPotion(Potion.LongWaterBreathing());
            if (identifier == "minecraft:lingering_potion" && metadata == 21) return new Item.LingeringPotion(Potion.Healing());
            if (identifier == "minecraft:lingering_potion" && metadata == 22) return new Item.LingeringPotion(Potion.StrongHealing());
            if (identifier == "minecraft:lingering_potion" && metadata == 23) return new Item.LingeringPotion(Potion.Harming());
            if (identifier == "minecraft:lingering_potion" && metadata == 24) return new Item.LingeringPotion(Potion.StrongHarming());
            if (identifier == "minecraft:lingering_potion" && metadata == 25) return new Item.LingeringPotion(Potion.Poison());
            if (identifier == "minecraft:lingering_potion" && metadata == 26) return new Item.LingeringPotion(Potion.LongPoison());
            if (identifier == "minecraft:lingering_potion" && metadata == 27) return new Item.LingeringPotion(Potion.StrongPoison());
            if (identifier == "minecraft:lingering_potion" && metadata == 28) return new Item.LingeringPotion(Potion.Regeneration());
            if (identifier == "minecraft:lingering_potion" && metadata == 29) return new Item.LingeringPotion(Potion.LongRegeneration());
            if (identifier == "minecraft:lingering_potion" && metadata == 30) return new Item.LingeringPotion(Potion.StrongRegeneration());
            if (identifier == "minecraft:lingering_potion" && metadata == 31) return new Item.LingeringPotion(Potion.Strength());
            if (identifier == "minecraft:lingering_potion" && metadata == 32) return new Item.LingeringPotion(Potion.LongStrength());
            if (identifier == "minecraft:lingering_potion" && metadata == 33) return new Item.LingeringPotion(Potion.StrongStrength());
            if (identifier == "minecraft:lingering_potion" && metadata == 34) return new Item.LingeringPotion(Potion.Weakness());
            if (identifier == "minecraft:lingering_potion" && metadata == 35) return new Item.LingeringPotion(Potion.LongWeakness());
            if (identifier == "minecraft:lingering_potion" && metadata == 36) return new Item.LingeringPotion(Potion.Wither());
            if (identifier == "minecraft:lingering_potion" && metadata == 37) return new Item.LingeringPotion(Potion.TurtleMaster());
            if (identifier == "minecraft:lingering_potion" && metadata == 38) return new Item.LingeringPotion(Potion.LongTurtleMaster());
            if (identifier == "minecraft:lingering_potion" && metadata == 39) return new Item.LingeringPotion(Potion.StrongTurtleMaster());
            if (identifier == "minecraft:lingering_potion" && metadata == 40) return new Item.LingeringPotion(Potion.SlowFalling());
            if (identifier == "minecraft:lingering_potion" && metadata == 41) return new Item.LingeringPotion(Potion.LongSlowFalling());
            if (identifier == "minecraft:lingering_potion" && metadata == 42) return new Item.LingeringPotion(Potion.StrongSlowness());
            if (identifier == "minecraft:magma_cream" && metadata == 0) return new Item.MagmaCream();
            if (identifier == "minecraft:melon_slice" && metadata == 0) return new Item.MelonSlice();
            if (identifier == "minecraft:mushroom_stew" && metadata == 0) return new Item.MushroomStew();
            if (identifier == "minecraft:music_disc_13" && metadata == 0) return new Item.MusicDisc(Sound.Disc13());
            if (identifier == "minecraft:music_disc_cat" && metadata == 0) return new Item.MusicDisc(Sound.DiscCat());
            if (identifier == "minecraft:music_disc_blocks" && metadata == 0) return new Item.MusicDisc(Sound.DiscBlocks());
            if (identifier == "minecraft:music_disc_chirp" && metadata == 0) return new Item.MusicDisc(Sound.DiscChirp());
            if (identifier == "minecraft:music_disc_far" && metadata == 0) return new Item.MusicDisc(Sound.DiscFar());
            if (identifier == "minecraft:music_disc_mall" && metadata == 0) return new Item.MusicDisc(Sound.DiscMall());
            if (identifier == "minecraft:music_disc_mellohi" && metadata == 0) return new Item.MusicDisc(Sound.DiscMellohi());
            if (identifier == "minecraft:music_disc_stal" && metadata == 0) return new Item.MusicDisc(Sound.DiscStal());
            if (identifier == "minecraft:music_disc_strad" && metadata == 0) return new Item.MusicDisc(Sound.DiscStrad());
            if (identifier == "minecraft:music_disc_ward" && metadata == 0) return new Item.MusicDisc(Sound.DiscWard());
            if (identifier == "minecraft:music_disc_11" && metadata == 0) return new Item.MusicDisc(Sound.Disc11());
            if (identifier == "minecraft:music_disc_wait" && metadata == 0) return new Item.MusicDisc(Sound.DiscWait());
            if (identifier == "minecraft:music_disc_otherside" && metadata == 0) return new Item.MusicDisc(Sound.DiscOtherside());
            if (identifier == "minecraft:music_disc_pigstep" && metadata == 0) return new Item.MusicDisc(Sound.DiscPigstep());
            if (identifier == "minecraft:music_disc_5" && metadata == 0) return new Item.MusicDisc(Sound.Disc5());
            if (identifier == "minecraft:music_disc_relic" && metadata == 0) return new Item.MusicDisc(Sound.DiscRelic());
            if (identifier == "minecraft:music_disc_creator" && metadata == 0) return new Item.MusicDisc(Sound.DiscCreator());
            if (identifier == "minecraft:music_disc_creator_music_box" && metadata == 0) return new Item.MusicDisc(Sound.DiscCreatorMusicBox());
            if (identifier == "minecraft:music_disc_precipice" && metadata == 0) return new Item.MusicDisc(Sound.DiscPrecipice());
            if (identifier == "minecraft:music_disc_tears" && metadata == 0) return new Item.MusicDisc(Sound.DiscTears());
            if (identifier == "minecraft:music_disc_lava_chicken" && metadata == 0) return new Item.MusicDisc(Sound.DiscLavaChicken());
            if (identifier == "minecraft:mutton" && metadata == 0) return new Item.Mutton(false);
            if (identifier == "minecraft:cooked_mutton" && metadata == 0) return new Item.Mutton(true);
            if (identifier == "minecraft:nautilus_shell" && metadata == 0) return new Item.NautilusShell();
            if (identifier == "minecraft:netherbrick" && metadata == 0) return new Item.NetherBrick();
            if (identifier == "minecraft:quartz" && metadata == 0) return new Item.NetherQuartz();
            if (identifier == "minecraft:nether_star" && metadata == 0) return new Item.NetherStar();
            if (identifier == "minecraft:netherite_ingot" && metadata == 0) return new Item.NetheriteIngot();
            if (identifier == "minecraft:netherite_scrap" && metadata == 0) return new Item.NetheriteScrap();
            if (identifier == "minecraft:paper" && metadata == 0) return new Item.Paper();
            if (identifier == "minecraft:phantom_membrane" && metadata == 0) return new Item.PhantomMembrane();
            if (identifier == "minecraft:wooden_pickaxe" && metadata == 0) return new Item.Pickaxe(Item.ToolTierWood);
            if (identifier == "minecraft:golden_pickaxe" && metadata == 0) return new Item.Pickaxe(Item.ToolTierGold);
            if (identifier == "minecraft:stone_pickaxe" && metadata == 0) return new Item.Pickaxe(Item.ToolTierStone);
            if (identifier == "minecraft:copper_pickaxe" && metadata == 0) return new Item.Pickaxe(Item.ToolTierCopper);
            if (identifier == "minecraft:iron_pickaxe" && metadata == 0) return new Item.Pickaxe(Item.ToolTierIron);
            if (identifier == "minecraft:diamond_pickaxe" && metadata == 0) return new Item.Pickaxe(Item.ToolTierDiamond);
            if (identifier == "minecraft:netherite_pickaxe" && metadata == 0) return new Item.Pickaxe(Item.ToolTierNetherite);
            if (identifier == "minecraft:poisonous_potato" && metadata == 0) return new Item.PoisonousPotato();
            if (identifier == "minecraft:popped_chorus_fruit" && metadata == 0) return new Item.PoppedChorusFruit();
            if (identifier == "minecraft:porkchop" && metadata == 0) return new Item.Porkchop(false);
            if (identifier == "minecraft:cooked_porkchop" && metadata == 0) return new Item.Porkchop(true);
            if (identifier == "minecraft:potion" && metadata == 0) return new Item.Potion(Potion.Water());
            if (identifier == "minecraft:potion" && metadata == 1) return new Item.Potion(Potion.Mundane());
            if (identifier == "minecraft:potion" && metadata == 2) return new Item.Potion(Potion.LongMundane());
            if (identifier == "minecraft:potion" && metadata == 3) return new Item.Potion(Potion.Thick());
            if (identifier == "minecraft:potion" && metadata == 4) return new Item.Potion(Potion.Awkward());
            if (identifier == "minecraft:potion" && metadata == 5) return new Item.Potion(Potion.NightVision());
            if (identifier == "minecraft:potion" && metadata == 6) return new Item.Potion(Potion.LongNightVision());
            if (identifier == "minecraft:potion" && metadata == 7) return new Item.Potion(Potion.Invisibility());
            if (identifier == "minecraft:potion" && metadata == 8) return new Item.Potion(Potion.LongInvisibility());
            if (identifier == "minecraft:potion" && metadata == 9) return new Item.Potion(Potion.Leaping());
            if (identifier == "minecraft:potion" && metadata == 10) return new Item.Potion(Potion.LongLeaping());
            if (identifier == "minecraft:potion" && metadata == 11) return new Item.Potion(Potion.StrongLeaping());
            if (identifier == "minecraft:potion" && metadata == 12) return new Item.Potion(Potion.FireResistance());
            if (identifier == "minecraft:potion" && metadata == 13) return new Item.Potion(Potion.LongFireResistance());
            if (identifier == "minecraft:potion" && metadata == 14) return new Item.Potion(Potion.Swiftness());
            if (identifier == "minecraft:potion" && metadata == 15) return new Item.Potion(Potion.LongSwiftness());
            if (identifier == "minecraft:potion" && metadata == 16) return new Item.Potion(Potion.StrongSwiftness());
            if (identifier == "minecraft:potion" && metadata == 17) return new Item.Potion(Potion.Slowness());
            if (identifier == "minecraft:potion" && metadata == 18) return new Item.Potion(Potion.LongSlowness());
            if (identifier == "minecraft:potion" && metadata == 19) return new Item.Potion(Potion.WaterBreathing());
            if (identifier == "minecraft:potion" && metadata == 20) return new Item.Potion(Potion.LongWaterBreathing());
            if (identifier == "minecraft:potion" && metadata == 21) return new Item.Potion(Potion.Healing());
            if (identifier == "minecraft:potion" && metadata == 22) return new Item.Potion(Potion.StrongHealing());
            if (identifier == "minecraft:potion" && metadata == 23) return new Item.Potion(Potion.Harming());
            if (identifier == "minecraft:potion" && metadata == 24) return new Item.Potion(Potion.StrongHarming());
            if (identifier == "minecraft:potion" && metadata == 25) return new Item.Potion(Potion.Poison());
            if (identifier == "minecraft:potion" && metadata == 26) return new Item.Potion(Potion.LongPoison());
            if (identifier == "minecraft:potion" && metadata == 27) return new Item.Potion(Potion.StrongPoison());
            if (identifier == "minecraft:potion" && metadata == 28) return new Item.Potion(Potion.Regeneration());
            if (identifier == "minecraft:potion" && metadata == 29) return new Item.Potion(Potion.LongRegeneration());
            if (identifier == "minecraft:potion" && metadata == 30) return new Item.Potion(Potion.StrongRegeneration());
            if (identifier == "minecraft:potion" && metadata == 31) return new Item.Potion(Potion.Strength());
            if (identifier == "minecraft:potion" && metadata == 32) return new Item.Potion(Potion.LongStrength());
            if (identifier == "minecraft:potion" && metadata == 33) return new Item.Potion(Potion.StrongStrength());
            if (identifier == "minecraft:potion" && metadata == 34) return new Item.Potion(Potion.Weakness());
            if (identifier == "minecraft:potion" && metadata == 35) return new Item.Potion(Potion.LongWeakness());
            if (identifier == "minecraft:potion" && metadata == 36) return new Item.Potion(Potion.Wither());
            if (identifier == "minecraft:potion" && metadata == 37) return new Item.Potion(Potion.TurtleMaster());
            if (identifier == "minecraft:potion" && metadata == 38) return new Item.Potion(Potion.LongTurtleMaster());
            if (identifier == "minecraft:potion" && metadata == 39) return new Item.Potion(Potion.StrongTurtleMaster());
            if (identifier == "minecraft:potion" && metadata == 40) return new Item.Potion(Potion.SlowFalling());
            if (identifier == "minecraft:potion" && metadata == 41) return new Item.Potion(Potion.LongSlowFalling());
            if (identifier == "minecraft:potion" && metadata == 42) return new Item.Potion(Potion.StrongSlowness());
            if (identifier == "minecraft:angler_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeAngler());
            if (identifier == "minecraft:archer_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeArcher());
            if (identifier == "minecraft:arms_up_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeArmsUp());
            if (identifier == "minecraft:blade_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeBlade());
            if (identifier == "minecraft:brewer_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeBrewer());
            if (identifier == "minecraft:burn_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeBurn());
            if (identifier == "minecraft:danger_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeDanger());
            if (identifier == "minecraft:explorer_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeExplorer());
            if (identifier == "minecraft:friend_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeFriend());
            if (identifier == "minecraft:heart_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeHeart());
            if (identifier == "minecraft:heartbreak_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeHeartbreak());
            if (identifier == "minecraft:howl_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeHowl());
            if (identifier == "minecraft:miner_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeMiner());
            if (identifier == "minecraft:mourner_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeMourner());
            if (identifier == "minecraft:plenty_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypePlenty());
            if (identifier == "minecraft:prize_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypePrize());
            if (identifier == "minecraft:sheaf_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeSheaf());
            if (identifier == "minecraft:shelter_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeShelter());
            if (identifier == "minecraft:skull_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeSkull());
            if (identifier == "minecraft:snort_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeSnort());
            if (identifier == "minecraft:flow_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeFlow());
            if (identifier == "minecraft:guster_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeGuster());
            if (identifier == "minecraft:scrape_pottery_sherd" && metadata == 0) return new Item.PotterySherd(Item.SherdTypeScrape());
            if (identifier == "minecraft:prismarine_crystals" && metadata == 0) return new Item.PrismarineCrystals();
            if (identifier == "minecraft:prismarine_shard" && metadata == 0) return new Item.PrismarineShard();
            if (identifier == "minecraft:pufferfish" && metadata == 0) return new Item.Pufferfish();
            if (identifier == "minecraft:pumpkin_pie" && metadata == 0) return new Item.PumpkinPie();
            if (identifier == "minecraft:rabbit" && metadata == 0) return new Item.Rabbit(false);
            if (identifier == "minecraft:cooked_rabbit" && metadata == 0) return new Item.Rabbit(true);
            if (identifier == "minecraft:rabbit_foot" && metadata == 0) return new Item.RabbitFoot();
            if (identifier == "minecraft:rabbit_hide" && metadata == 0) return new Item.RabbitHide();
            if (identifier == "minecraft:rabbit_stew" && metadata == 0) return new Item.RabbitStew();
            if (identifier == "minecraft:raw_copper" && metadata == 0) return new Item.RawCopper();
            if (identifier == "minecraft:raw_gold" && metadata == 0) return new Item.RawGold();
            if (identifier == "minecraft:raw_iron" && metadata == 0) return new Item.RawIron();
            if (identifier == "minecraft:recovery_compass" && metadata == 0) return new Item.RecoveryCompass();
            if (identifier == "minecraft:resin_brick" && metadata == 0) return new Item.ResinBrick();
            if (identifier == "minecraft:rotten_flesh" && metadata == 0) return new Item.RottenFlesh();
            if (identifier == "minecraft:salmon" && metadata == 0) return new Item.Salmon(false);
            if (identifier == "minecraft:cooked_salmon" && metadata == 0) return new Item.Salmon(true);
            if (identifier == "minecraft:turtle_scute" && metadata == 0) return new Item.Scute();
            if (identifier == "minecraft:shears" && metadata == 0) return new Item.Shears();
            if (identifier == "minecraft:wooden_shovel" && metadata == 0) return new Item.Shovel(Item.ToolTierWood);
            if (identifier == "minecraft:golden_shovel" && metadata == 0) return new Item.Shovel(Item.ToolTierGold);
            if (identifier == "minecraft:stone_shovel" && metadata == 0) return new Item.Shovel(Item.ToolTierStone);
            if (identifier == "minecraft:copper_shovel" && metadata == 0) return new Item.Shovel(Item.ToolTierCopper);
            if (identifier == "minecraft:iron_shovel" && metadata == 0) return new Item.Shovel(Item.ToolTierIron);
            if (identifier == "minecraft:diamond_shovel" && metadata == 0) return new Item.Shovel(Item.ToolTierDiamond);
            if (identifier == "minecraft:netherite_shovel" && metadata == 0) return new Item.Shovel(Item.ToolTierNetherite);
            if (identifier == "minecraft:shulker_shell" && metadata == 0) return new Item.ShulkerShell();
            if (identifier == "minecraft:slime_ball" && metadata == 0) return new Item.Slimeball();
            if (identifier == "minecraft:netherite_upgrade_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateNetheriteUpgrade());
            if (identifier == "minecraft:sentry_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateSentry());
            if (identifier == "minecraft:vex_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateVex());
            if (identifier == "minecraft:wild_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateWild());
            if (identifier == "minecraft:coast_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateCoast());
            if (identifier == "minecraft:dune_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateDune());
            if (identifier == "minecraft:wayfinder_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateWayFinder());
            if (identifier == "minecraft:raiser_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateRaiser());
            if (identifier == "minecraft:shaper_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateShaper());
            if (identifier == "minecraft:host_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateHost());
            if (identifier == "minecraft:ward_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateWard());
            if (identifier == "minecraft:silence_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateSilence());
            if (identifier == "minecraft:tide_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateTide());
            if (identifier == "minecraft:snout_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateSnout());
            if (identifier == "minecraft:rib_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateRib());
            if (identifier == "minecraft:eye_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateEye());
            if (identifier == "minecraft:spire_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateSpire());
            if (identifier == "minecraft:flow_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateFlow());
            if (identifier == "minecraft:bolt_armor_trim_smithing_template" && metadata == 0) return new Item.SmithingTemplate(Item.TemplateBolt());
            if (identifier == "minecraft:snowball" && metadata == 0) return new Item.Snowball();
            if (identifier == "minecraft:spider_eye" && metadata == 0) return new Item.SpiderEye();
            if (identifier == "minecraft:splash_potion" && metadata == 0) return new Item.SplashPotion(Potion.Water());
            if (identifier == "minecraft:splash_potion" && metadata == 1) return new Item.SplashPotion(Potion.Mundane());
            if (identifier == "minecraft:splash_potion" && metadata == 2) return new Item.SplashPotion(Potion.LongMundane());
            if (identifier == "minecraft:splash_potion" && metadata == 3) return new Item.SplashPotion(Potion.Thick());
            if (identifier == "minecraft:splash_potion" && metadata == 4) return new Item.SplashPotion(Potion.Awkward());
            if (identifier == "minecraft:splash_potion" && metadata == 5) return new Item.SplashPotion(Potion.NightVision());
            if (identifier == "minecraft:splash_potion" && metadata == 6) return new Item.SplashPotion(Potion.LongNightVision());
            if (identifier == "minecraft:splash_potion" && metadata == 7) return new Item.SplashPotion(Potion.Invisibility());
            if (identifier == "minecraft:splash_potion" && metadata == 8) return new Item.SplashPotion(Potion.LongInvisibility());
            if (identifier == "minecraft:splash_potion" && metadata == 9) return new Item.SplashPotion(Potion.Leaping());
            if (identifier == "minecraft:splash_potion" && metadata == 10) return new Item.SplashPotion(Potion.LongLeaping());
            if (identifier == "minecraft:splash_potion" && metadata == 11) return new Item.SplashPotion(Potion.StrongLeaping());
            if (identifier == "minecraft:splash_potion" && metadata == 12) return new Item.SplashPotion(Potion.FireResistance());
            if (identifier == "minecraft:splash_potion" && metadata == 13) return new Item.SplashPotion(Potion.LongFireResistance());
            if (identifier == "minecraft:splash_potion" && metadata == 14) return new Item.SplashPotion(Potion.Swiftness());
            if (identifier == "minecraft:splash_potion" && metadata == 15) return new Item.SplashPotion(Potion.LongSwiftness());
            if (identifier == "minecraft:splash_potion" && metadata == 16) return new Item.SplashPotion(Potion.StrongSwiftness());
            if (identifier == "minecraft:splash_potion" && metadata == 17) return new Item.SplashPotion(Potion.Slowness());
            if (identifier == "minecraft:splash_potion" && metadata == 18) return new Item.SplashPotion(Potion.LongSlowness());
            if (identifier == "minecraft:splash_potion" && metadata == 19) return new Item.SplashPotion(Potion.WaterBreathing());
            if (identifier == "minecraft:splash_potion" && metadata == 20) return new Item.SplashPotion(Potion.LongWaterBreathing());
            if (identifier == "minecraft:splash_potion" && metadata == 21) return new Item.SplashPotion(Potion.Healing());
            if (identifier == "minecraft:splash_potion" && metadata == 22) return new Item.SplashPotion(Potion.StrongHealing());
            if (identifier == "minecraft:splash_potion" && metadata == 23) return new Item.SplashPotion(Potion.Harming());
            if (identifier == "minecraft:splash_potion" && metadata == 24) return new Item.SplashPotion(Potion.StrongHarming());
            if (identifier == "minecraft:splash_potion" && metadata == 25) return new Item.SplashPotion(Potion.Poison());
            if (identifier == "minecraft:splash_potion" && metadata == 26) return new Item.SplashPotion(Potion.LongPoison());
            if (identifier == "minecraft:splash_potion" && metadata == 27) return new Item.SplashPotion(Potion.StrongPoison());
            if (identifier == "minecraft:splash_potion" && metadata == 28) return new Item.SplashPotion(Potion.Regeneration());
            if (identifier == "minecraft:splash_potion" && metadata == 29) return new Item.SplashPotion(Potion.LongRegeneration());
            if (identifier == "minecraft:splash_potion" && metadata == 30) return new Item.SplashPotion(Potion.StrongRegeneration());
            if (identifier == "minecraft:splash_potion" && metadata == 31) return new Item.SplashPotion(Potion.Strength());
            if (identifier == "minecraft:splash_potion" && metadata == 32) return new Item.SplashPotion(Potion.LongStrength());
            if (identifier == "minecraft:splash_potion" && metadata == 33) return new Item.SplashPotion(Potion.StrongStrength());
            if (identifier == "minecraft:splash_potion" && metadata == 34) return new Item.SplashPotion(Potion.Weakness());
            if (identifier == "minecraft:splash_potion" && metadata == 35) return new Item.SplashPotion(Potion.LongWeakness());
            if (identifier == "minecraft:splash_potion" && metadata == 36) return new Item.SplashPotion(Potion.Wither());
            if (identifier == "minecraft:splash_potion" && metadata == 37) return new Item.SplashPotion(Potion.TurtleMaster());
            if (identifier == "minecraft:splash_potion" && metadata == 38) return new Item.SplashPotion(Potion.LongTurtleMaster());
            if (identifier == "minecraft:splash_potion" && metadata == 39) return new Item.SplashPotion(Potion.StrongTurtleMaster());
            if (identifier == "minecraft:splash_potion" && metadata == 40) return new Item.SplashPotion(Potion.SlowFalling());
            if (identifier == "minecraft:splash_potion" && metadata == 41) return new Item.SplashPotion(Potion.LongSlowFalling());
            if (identifier == "minecraft:splash_potion" && metadata == 42) return new Item.SplashPotion(Potion.StrongSlowness());
            if (identifier == "minecraft:spyglass" && metadata == 0) return new Item.Spyglass();
            if (identifier == "minecraft:stick" && metadata == 0) return new Item.Stick();
            if (identifier == "minecraft:sugar" && metadata == 0) return new Item.Sugar();
            if (identifier == "minecraft:suspicious_stew" && metadata == 0) return new Item.SuspiciousStew(Item.NightVisionPoppyStew());
            if (identifier == "minecraft:suspicious_stew" && metadata == 1) return new Item.SuspiciousStew(Item.JumpBoostStew());
            if (identifier == "minecraft:suspicious_stew" && metadata == 2) return new Item.SuspiciousStew(Item.WeaknessStew());
            if (identifier == "minecraft:suspicious_stew" && metadata == 3) return new Item.SuspiciousStew(Item.BlindnessBluetStew());
            if (identifier == "minecraft:suspicious_stew" && metadata == 4) return new Item.SuspiciousStew(Item.PoisonStew());
            if (identifier == "minecraft:suspicious_stew" && metadata == 5) return new Item.SuspiciousStew(Item.SaturationDandelionStew());
            if (identifier == "minecraft:suspicious_stew" && metadata == 6) return new Item.SuspiciousStew(Item.SaturationOrchidStew());
            if (identifier == "minecraft:suspicious_stew" && metadata == 7) return new Item.SuspiciousStew(Item.FireResistanceStew());
            if (identifier == "minecraft:suspicious_stew" && metadata == 8) return new Item.SuspiciousStew(Item.RegenerationStew());
            if (identifier == "minecraft:suspicious_stew" && metadata == 9) return new Item.SuspiciousStew(Item.WitherStew());
            if (identifier == "minecraft:suspicious_stew" && metadata == 10) return new Item.SuspiciousStew(Item.NightVisionTorchflowerStew());
            if (identifier == "minecraft:suspicious_stew" && metadata == 11) return new Item.SuspiciousStew(Item.BlindnessEyeblossomStew());
            if (identifier == "minecraft:suspicious_stew" && metadata == 12) return new Item.SuspiciousStew(Item.NauseaStew());
            if (identifier == "minecraft:wooden_sword" && metadata == 0) return new Item.Sword(Item.ToolTierWood);
            if (identifier == "minecraft:golden_sword" && metadata == 0) return new Item.Sword(Item.ToolTierGold);
            if (identifier == "minecraft:stone_sword" && metadata == 0) return new Item.Sword(Item.ToolTierStone);
            if (identifier == "minecraft:copper_sword" && metadata == 0) return new Item.Sword(Item.ToolTierCopper);
            if (identifier == "minecraft:iron_sword" && metadata == 0) return new Item.Sword(Item.ToolTierIron);
            if (identifier == "minecraft:diamond_sword" && metadata == 0) return new Item.Sword(Item.ToolTierDiamond);
            if (identifier == "minecraft:netherite_sword" && metadata == 0) return new Item.Sword(Item.ToolTierNetherite);
            if (identifier == "minecraft:totem_of_undying" && metadata == 0) return new Item.Totem();
            if (identifier == "minecraft:tropical_fish" && metadata == 0) return new Item.TropicalFish();
            if (identifier == "minecraft:turtle_helmet" && metadata == 0) return new Item.TurtleShell();
            if (identifier == "minecraft:warped_fungus_on_a_stick" && metadata == 0) return new Item.WarpedFungusOnAStick();
            if (identifier == "minecraft:wheat" && metadata == 0) return new Item.Wheat();
            return new EncodedItem(identifier, metadata);
        }

        internal static bool IsAir(World.Item item) =>
            TryEncode(item, out var identifier, out _) && identifier == "minecraft:air";

        private sealed record EncodedItem(string Identifier, int Metadata) : World.Item;
    }

    internal readonly record struct ItemDurability(int MaxDurability, bool Persistent, Item.Stack BrokenStack);

    internal static class ItemCapabilities
    {
        internal static int MaxCount(World.Item? item) => item switch
        {
            Item.Axe value when value.Tier == Item.ToolTierWood => 1,
            Item.Axe value when value.Tier == Item.ToolTierGold => 1,
            Item.Axe value when value.Tier == Item.ToolTierStone => 1,
            Item.Axe value when value.Tier == Item.ToolTierCopper => 1,
            Item.Axe value when value.Tier == Item.ToolTierIron => 1,
            Item.Axe value when value.Tier == Item.ToolTierDiamond => 1,
            Item.Axe value when value.Tier == Item.ToolTierNetherite => 1,
            Item.BannerPattern value when value.Type == Item.CreeperBannerPattern() => 1,
            Item.BannerPattern value when value.Type == Item.SkullBannerPattern() => 1,
            Item.BannerPattern value when value.Type == Item.FlowerBannerPattern() => 1,
            Item.BannerPattern value when value.Type == Item.MojangBannerPattern() => 1,
            Item.BannerPattern value when value.Type == Item.FieldMasonedBannerPattern() => 1,
            Item.BannerPattern value when value.Type == Item.BordureIndentedBannerPattern() => 1,
            Item.BannerPattern value when value.Type == Item.PiglinBannerPattern() => 1,
            Item.BannerPattern value when value.Type == Item.GlobeBannerPattern() => 1,
            Item.BannerPattern value when value.Type == Item.FlowBannerPattern() => 1,
            Item.BannerPattern value when value.Type == Item.GusterBannerPattern() => 1,
            Item.BeetrootSoup _ => 1,
            Item.Bow _ => 1,
            Item.CarrotOnAStick _ => 1,
            Item.Egg _ => 16,
            Item.Elytra _ => 1,
            Item.EnchantedBook _ => 1,
            Item.EnderPearl _ => 16,
            Item.FlintAndSteel _ => 1,
            Item.GoatHorn value when value.Type == Sound.Ponder() => 1,
            Item.GoatHorn value when value.Type == Sound.Sing() => 1,
            Item.GoatHorn value when value.Type == Sound.Seek() => 1,
            Item.GoatHorn value when value.Type == Sound.Feel() => 1,
            Item.GoatHorn value when value.Type == Sound.Admire() => 1,
            Item.GoatHorn value when value.Type == Sound.Call() => 1,
            Item.GoatHorn value when value.Type == Sound.Yearn() => 1,
            Item.GoatHorn value when value.Type == Sound.Dream() => 1,
            Item.Hoe value when value.Tier == Item.ToolTierWood => 1,
            Item.Hoe value when value.Tier == Item.ToolTierGold => 1,
            Item.Hoe value when value.Tier == Item.ToolTierStone => 1,
            Item.Hoe value when value.Tier == Item.ToolTierCopper => 1,
            Item.Hoe value when value.Tier == Item.ToolTierIron => 1,
            Item.Hoe value when value.Tier == Item.ToolTierDiamond => 1,
            Item.Hoe value when value.Tier == Item.ToolTierNetherite => 1,
            Item.HoneyBottle _ => 16,
            Item.LingeringPotion value when value.Type == Potion.Water() => 1,
            Item.LingeringPotion value when value.Type == Potion.Mundane() => 1,
            Item.LingeringPotion value when value.Type == Potion.LongMundane() => 1,
            Item.LingeringPotion value when value.Type == Potion.Thick() => 1,
            Item.LingeringPotion value when value.Type == Potion.Awkward() => 1,
            Item.LingeringPotion value when value.Type == Potion.NightVision() => 1,
            Item.LingeringPotion value when value.Type == Potion.LongNightVision() => 1,
            Item.LingeringPotion value when value.Type == Potion.Invisibility() => 1,
            Item.LingeringPotion value when value.Type == Potion.LongInvisibility() => 1,
            Item.LingeringPotion value when value.Type == Potion.Leaping() => 1,
            Item.LingeringPotion value when value.Type == Potion.LongLeaping() => 1,
            Item.LingeringPotion value when value.Type == Potion.StrongLeaping() => 1,
            Item.LingeringPotion value when value.Type == Potion.FireResistance() => 1,
            Item.LingeringPotion value when value.Type == Potion.LongFireResistance() => 1,
            Item.LingeringPotion value when value.Type == Potion.Swiftness() => 1,
            Item.LingeringPotion value when value.Type == Potion.LongSwiftness() => 1,
            Item.LingeringPotion value when value.Type == Potion.StrongSwiftness() => 1,
            Item.LingeringPotion value when value.Type == Potion.Slowness() => 1,
            Item.LingeringPotion value when value.Type == Potion.LongSlowness() => 1,
            Item.LingeringPotion value when value.Type == Potion.WaterBreathing() => 1,
            Item.LingeringPotion value when value.Type == Potion.LongWaterBreathing() => 1,
            Item.LingeringPotion value when value.Type == Potion.Healing() => 1,
            Item.LingeringPotion value when value.Type == Potion.StrongHealing() => 1,
            Item.LingeringPotion value when value.Type == Potion.Harming() => 1,
            Item.LingeringPotion value when value.Type == Potion.StrongHarming() => 1,
            Item.LingeringPotion value when value.Type == Potion.Poison() => 1,
            Item.LingeringPotion value when value.Type == Potion.LongPoison() => 1,
            Item.LingeringPotion value when value.Type == Potion.StrongPoison() => 1,
            Item.LingeringPotion value when value.Type == Potion.Regeneration() => 1,
            Item.LingeringPotion value when value.Type == Potion.LongRegeneration() => 1,
            Item.LingeringPotion value when value.Type == Potion.StrongRegeneration() => 1,
            Item.LingeringPotion value when value.Type == Potion.Strength() => 1,
            Item.LingeringPotion value when value.Type == Potion.LongStrength() => 1,
            Item.LingeringPotion value when value.Type == Potion.StrongStrength() => 1,
            Item.LingeringPotion value when value.Type == Potion.Weakness() => 1,
            Item.LingeringPotion value when value.Type == Potion.LongWeakness() => 1,
            Item.LingeringPotion value when value.Type == Potion.Wither() => 1,
            Item.LingeringPotion value when value.Type == Potion.TurtleMaster() => 1,
            Item.LingeringPotion value when value.Type == Potion.LongTurtleMaster() => 1,
            Item.LingeringPotion value when value.Type == Potion.StrongTurtleMaster() => 1,
            Item.LingeringPotion value when value.Type == Potion.SlowFalling() => 1,
            Item.LingeringPotion value when value.Type == Potion.LongSlowFalling() => 1,
            Item.LingeringPotion value when value.Type == Potion.StrongSlowness() => 1,
            Item.MushroomStew _ => 1,
            Item.MusicDisc value when value.DiscType == Sound.Disc13() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscCat() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscBlocks() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscChirp() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscFar() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscMall() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscMellohi() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscStal() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscStrad() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscWard() => 1,
            Item.MusicDisc value when value.DiscType == Sound.Disc11() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscWait() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscOtherside() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscPigstep() => 1,
            Item.MusicDisc value when value.DiscType == Sound.Disc5() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscRelic() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscCreator() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscCreatorMusicBox() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscPrecipice() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscTears() => 1,
            Item.MusicDisc value when value.DiscType == Sound.DiscLavaChicken() => 1,
            Item.Pickaxe value when value.Tier == Item.ToolTierWood => 1,
            Item.Pickaxe value when value.Tier == Item.ToolTierGold => 1,
            Item.Pickaxe value when value.Tier == Item.ToolTierStone => 1,
            Item.Pickaxe value when value.Tier == Item.ToolTierCopper => 1,
            Item.Pickaxe value when value.Tier == Item.ToolTierIron => 1,
            Item.Pickaxe value when value.Tier == Item.ToolTierDiamond => 1,
            Item.Pickaxe value when value.Tier == Item.ToolTierNetherite => 1,
            Item.Potion value when value.Type == Potion.Water() => 1,
            Item.Potion value when value.Type == Potion.Mundane() => 1,
            Item.Potion value when value.Type == Potion.LongMundane() => 1,
            Item.Potion value when value.Type == Potion.Thick() => 1,
            Item.Potion value when value.Type == Potion.Awkward() => 1,
            Item.Potion value when value.Type == Potion.NightVision() => 1,
            Item.Potion value when value.Type == Potion.LongNightVision() => 1,
            Item.Potion value when value.Type == Potion.Invisibility() => 1,
            Item.Potion value when value.Type == Potion.LongInvisibility() => 1,
            Item.Potion value when value.Type == Potion.Leaping() => 1,
            Item.Potion value when value.Type == Potion.LongLeaping() => 1,
            Item.Potion value when value.Type == Potion.StrongLeaping() => 1,
            Item.Potion value when value.Type == Potion.FireResistance() => 1,
            Item.Potion value when value.Type == Potion.LongFireResistance() => 1,
            Item.Potion value when value.Type == Potion.Swiftness() => 1,
            Item.Potion value when value.Type == Potion.LongSwiftness() => 1,
            Item.Potion value when value.Type == Potion.StrongSwiftness() => 1,
            Item.Potion value when value.Type == Potion.Slowness() => 1,
            Item.Potion value when value.Type == Potion.LongSlowness() => 1,
            Item.Potion value when value.Type == Potion.WaterBreathing() => 1,
            Item.Potion value when value.Type == Potion.LongWaterBreathing() => 1,
            Item.Potion value when value.Type == Potion.Healing() => 1,
            Item.Potion value when value.Type == Potion.StrongHealing() => 1,
            Item.Potion value when value.Type == Potion.Harming() => 1,
            Item.Potion value when value.Type == Potion.StrongHarming() => 1,
            Item.Potion value when value.Type == Potion.Poison() => 1,
            Item.Potion value when value.Type == Potion.LongPoison() => 1,
            Item.Potion value when value.Type == Potion.StrongPoison() => 1,
            Item.Potion value when value.Type == Potion.Regeneration() => 1,
            Item.Potion value when value.Type == Potion.LongRegeneration() => 1,
            Item.Potion value when value.Type == Potion.StrongRegeneration() => 1,
            Item.Potion value when value.Type == Potion.Strength() => 1,
            Item.Potion value when value.Type == Potion.LongStrength() => 1,
            Item.Potion value when value.Type == Potion.StrongStrength() => 1,
            Item.Potion value when value.Type == Potion.Weakness() => 1,
            Item.Potion value when value.Type == Potion.LongWeakness() => 1,
            Item.Potion value when value.Type == Potion.Wither() => 1,
            Item.Potion value when value.Type == Potion.TurtleMaster() => 1,
            Item.Potion value when value.Type == Potion.LongTurtleMaster() => 1,
            Item.Potion value when value.Type == Potion.StrongTurtleMaster() => 1,
            Item.Potion value when value.Type == Potion.SlowFalling() => 1,
            Item.Potion value when value.Type == Potion.LongSlowFalling() => 1,
            Item.Potion value when value.Type == Potion.StrongSlowness() => 1,
            Item.RabbitStew _ => 1,
            Item.Shears _ => 1,
            Item.Shovel value when value.Tier == Item.ToolTierWood => 1,
            Item.Shovel value when value.Tier == Item.ToolTierGold => 1,
            Item.Shovel value when value.Tier == Item.ToolTierStone => 1,
            Item.Shovel value when value.Tier == Item.ToolTierCopper => 1,
            Item.Shovel value when value.Tier == Item.ToolTierIron => 1,
            Item.Shovel value when value.Tier == Item.ToolTierDiamond => 1,
            Item.Shovel value when value.Tier == Item.ToolTierNetherite => 1,
            Item.Snowball _ => 16,
            Item.SplashPotion value when value.Type == Potion.Water() => 1,
            Item.SplashPotion value when value.Type == Potion.Mundane() => 1,
            Item.SplashPotion value when value.Type == Potion.LongMundane() => 1,
            Item.SplashPotion value when value.Type == Potion.Thick() => 1,
            Item.SplashPotion value when value.Type == Potion.Awkward() => 1,
            Item.SplashPotion value when value.Type == Potion.NightVision() => 1,
            Item.SplashPotion value when value.Type == Potion.LongNightVision() => 1,
            Item.SplashPotion value when value.Type == Potion.Invisibility() => 1,
            Item.SplashPotion value when value.Type == Potion.LongInvisibility() => 1,
            Item.SplashPotion value when value.Type == Potion.Leaping() => 1,
            Item.SplashPotion value when value.Type == Potion.LongLeaping() => 1,
            Item.SplashPotion value when value.Type == Potion.StrongLeaping() => 1,
            Item.SplashPotion value when value.Type == Potion.FireResistance() => 1,
            Item.SplashPotion value when value.Type == Potion.LongFireResistance() => 1,
            Item.SplashPotion value when value.Type == Potion.Swiftness() => 1,
            Item.SplashPotion value when value.Type == Potion.LongSwiftness() => 1,
            Item.SplashPotion value when value.Type == Potion.StrongSwiftness() => 1,
            Item.SplashPotion value when value.Type == Potion.Slowness() => 1,
            Item.SplashPotion value when value.Type == Potion.LongSlowness() => 1,
            Item.SplashPotion value when value.Type == Potion.WaterBreathing() => 1,
            Item.SplashPotion value when value.Type == Potion.LongWaterBreathing() => 1,
            Item.SplashPotion value when value.Type == Potion.Healing() => 1,
            Item.SplashPotion value when value.Type == Potion.StrongHealing() => 1,
            Item.SplashPotion value when value.Type == Potion.Harming() => 1,
            Item.SplashPotion value when value.Type == Potion.StrongHarming() => 1,
            Item.SplashPotion value when value.Type == Potion.Poison() => 1,
            Item.SplashPotion value when value.Type == Potion.LongPoison() => 1,
            Item.SplashPotion value when value.Type == Potion.StrongPoison() => 1,
            Item.SplashPotion value when value.Type == Potion.Regeneration() => 1,
            Item.SplashPotion value when value.Type == Potion.LongRegeneration() => 1,
            Item.SplashPotion value when value.Type == Potion.StrongRegeneration() => 1,
            Item.SplashPotion value when value.Type == Potion.Strength() => 1,
            Item.SplashPotion value when value.Type == Potion.LongStrength() => 1,
            Item.SplashPotion value when value.Type == Potion.StrongStrength() => 1,
            Item.SplashPotion value when value.Type == Potion.Weakness() => 1,
            Item.SplashPotion value when value.Type == Potion.LongWeakness() => 1,
            Item.SplashPotion value when value.Type == Potion.Wither() => 1,
            Item.SplashPotion value when value.Type == Potion.TurtleMaster() => 1,
            Item.SplashPotion value when value.Type == Potion.LongTurtleMaster() => 1,
            Item.SplashPotion value when value.Type == Potion.StrongTurtleMaster() => 1,
            Item.SplashPotion value when value.Type == Potion.SlowFalling() => 1,
            Item.SplashPotion value when value.Type == Potion.LongSlowFalling() => 1,
            Item.SplashPotion value when value.Type == Potion.StrongSlowness() => 1,
            Item.Spyglass _ => 1,
            Item.SuspiciousStew value when value.Type == Item.NightVisionPoppyStew() => 1,
            Item.SuspiciousStew value when value.Type == Item.JumpBoostStew() => 1,
            Item.SuspiciousStew value when value.Type == Item.WeaknessStew() => 1,
            Item.SuspiciousStew value when value.Type == Item.BlindnessBluetStew() => 1,
            Item.SuspiciousStew value when value.Type == Item.PoisonStew() => 1,
            Item.SuspiciousStew value when value.Type == Item.SaturationDandelionStew() => 1,
            Item.SuspiciousStew value when value.Type == Item.SaturationOrchidStew() => 1,
            Item.SuspiciousStew value when value.Type == Item.FireResistanceStew() => 1,
            Item.SuspiciousStew value when value.Type == Item.RegenerationStew() => 1,
            Item.SuspiciousStew value when value.Type == Item.WitherStew() => 1,
            Item.SuspiciousStew value when value.Type == Item.NightVisionTorchflowerStew() => 1,
            Item.SuspiciousStew value when value.Type == Item.BlindnessEyeblossomStew() => 1,
            Item.SuspiciousStew value when value.Type == Item.NauseaStew() => 1,
            Item.Sword value when value.Tier == Item.ToolTierWood => 1,
            Item.Sword value when value.Tier == Item.ToolTierGold => 1,
            Item.Sword value when value.Tier == Item.ToolTierStone => 1,
            Item.Sword value when value.Tier == Item.ToolTierCopper => 1,
            Item.Sword value when value.Tier == Item.ToolTierIron => 1,
            Item.Sword value when value.Tier == Item.ToolTierDiamond => 1,
            Item.Sword value when value.Tier == Item.ToolTierNetherite => 1,
            Item.Totem _ => 1,
            Item.TurtleShell _ => 1,
            Item.WarpedFungusOnAStick _ => 1,
            _ => 64,
        };

        internal static bool TryDurability(World.Item? item, out ItemDurability durability)
        {
            switch (item)
            {
                case Item.Axe value when value.Tier == Item.ToolTierWood:
                    durability = new(59, false, default); return true;
                case Item.Axe value when value.Tier == Item.ToolTierGold:
                    durability = new(32, false, default); return true;
                case Item.Axe value when value.Tier == Item.ToolTierStone:
                    durability = new(131, false, default); return true;
                case Item.Axe value when value.Tier == Item.ToolTierCopper:
                    durability = new(190, false, default); return true;
                case Item.Axe value when value.Tier == Item.ToolTierIron:
                    durability = new(250, false, default); return true;
                case Item.Axe value when value.Tier == Item.ToolTierDiamond:
                    durability = new(1561, false, default); return true;
                case Item.Axe value when value.Tier == Item.ToolTierNetherite:
                    durability = new(2031, false, default); return true;
                case Item.Bow _:
                    durability = new(385, false, default); return true;
                case Item.Elytra _:
                    durability = new(433, true, default); return true;
                case Item.FlintAndSteel _:
                    durability = new(65, false, default); return true;
                case Item.Hoe value when value.Tier == Item.ToolTierWood:
                    durability = new(59, false, default); return true;
                case Item.Hoe value when value.Tier == Item.ToolTierGold:
                    durability = new(32, false, default); return true;
                case Item.Hoe value when value.Tier == Item.ToolTierStone:
                    durability = new(131, false, default); return true;
                case Item.Hoe value when value.Tier == Item.ToolTierCopper:
                    durability = new(190, false, default); return true;
                case Item.Hoe value when value.Tier == Item.ToolTierIron:
                    durability = new(250, false, default); return true;
                case Item.Hoe value when value.Tier == Item.ToolTierDiamond:
                    durability = new(1561, false, default); return true;
                case Item.Hoe value when value.Tier == Item.ToolTierNetherite:
                    durability = new(2031, false, default); return true;
                case Item.Pickaxe value when value.Tier == Item.ToolTierWood:
                    durability = new(59, false, default); return true;
                case Item.Pickaxe value when value.Tier == Item.ToolTierGold:
                    durability = new(32, false, default); return true;
                case Item.Pickaxe value when value.Tier == Item.ToolTierStone:
                    durability = new(131, false, default); return true;
                case Item.Pickaxe value when value.Tier == Item.ToolTierCopper:
                    durability = new(190, false, default); return true;
                case Item.Pickaxe value when value.Tier == Item.ToolTierIron:
                    durability = new(250, false, default); return true;
                case Item.Pickaxe value when value.Tier == Item.ToolTierDiamond:
                    durability = new(1561, false, default); return true;
                case Item.Pickaxe value when value.Tier == Item.ToolTierNetherite:
                    durability = new(2031, false, default); return true;
                case Item.Shears _:
                    durability = new(238, false, default); return true;
                case Item.Shovel value when value.Tier == Item.ToolTierWood:
                    durability = new(59, false, default); return true;
                case Item.Shovel value when value.Tier == Item.ToolTierGold:
                    durability = new(32, false, default); return true;
                case Item.Shovel value when value.Tier == Item.ToolTierStone:
                    durability = new(131, false, default); return true;
                case Item.Shovel value when value.Tier == Item.ToolTierCopper:
                    durability = new(190, false, default); return true;
                case Item.Shovel value when value.Tier == Item.ToolTierIron:
                    durability = new(250, false, default); return true;
                case Item.Shovel value when value.Tier == Item.ToolTierDiamond:
                    durability = new(1561, false, default); return true;
                case Item.Shovel value when value.Tier == Item.ToolTierNetherite:
                    durability = new(2031, false, default); return true;
                case Item.Sword value when value.Tier == Item.ToolTierWood:
                    durability = new(59, false, default); return true;
                case Item.Sword value when value.Tier == Item.ToolTierGold:
                    durability = new(32, false, default); return true;
                case Item.Sword value when value.Tier == Item.ToolTierStone:
                    durability = new(131, false, default); return true;
                case Item.Sword value when value.Tier == Item.ToolTierCopper:
                    durability = new(190, false, default); return true;
                case Item.Sword value when value.Tier == Item.ToolTierIron:
                    durability = new(250, false, default); return true;
                case Item.Sword value when value.Tier == Item.ToolTierDiamond:
                    durability = new(1561, false, default); return true;
                case Item.Sword value when value.Tier == Item.ToolTierNetherite:
                    durability = new(2031, false, default); return true;
                case Item.TurtleShell _:
                    durability = new(276, false, default); return true;
                default:
                    durability = default; return false;
            }
        }

        internal static double AttackDamage(World.Item? item) => item switch
        {
            Item.Axe value when value.Tier == Item.ToolTierWood => 4d,
            Item.Axe value when value.Tier == Item.ToolTierGold => 4d,
            Item.Axe value when value.Tier == Item.ToolTierStone => 5d,
            Item.Axe value when value.Tier == Item.ToolTierCopper => 5d,
            Item.Axe value when value.Tier == Item.ToolTierIron => 6d,
            Item.Axe value when value.Tier == Item.ToolTierDiamond => 7d,
            Item.Axe value when value.Tier == Item.ToolTierNetherite => 8d,
            Item.Hoe value when value.Tier == Item.ToolTierWood => 3d,
            Item.Hoe value when value.Tier == Item.ToolTierGold => 3d,
            Item.Hoe value when value.Tier == Item.ToolTierStone => 4d,
            Item.Hoe value when value.Tier == Item.ToolTierCopper => 4d,
            Item.Hoe value when value.Tier == Item.ToolTierIron => 5d,
            Item.Hoe value when value.Tier == Item.ToolTierDiamond => 6d,
            Item.Hoe value when value.Tier == Item.ToolTierNetherite => 7d,
            Item.Pickaxe value when value.Tier == Item.ToolTierWood => 3d,
            Item.Pickaxe value when value.Tier == Item.ToolTierGold => 3d,
            Item.Pickaxe value when value.Tier == Item.ToolTierStone => 4d,
            Item.Pickaxe value when value.Tier == Item.ToolTierCopper => 4d,
            Item.Pickaxe value when value.Tier == Item.ToolTierIron => 5d,
            Item.Pickaxe value when value.Tier == Item.ToolTierDiamond => 6d,
            Item.Pickaxe value when value.Tier == Item.ToolTierNetherite => 7d,
            Item.Shovel value when value.Tier == Item.ToolTierWood => 2d,
            Item.Shovel value when value.Tier == Item.ToolTierGold => 2d,
            Item.Shovel value when value.Tier == Item.ToolTierStone => 3d,
            Item.Shovel value when value.Tier == Item.ToolTierCopper => 3d,
            Item.Shovel value when value.Tier == Item.ToolTierIron => 4d,
            Item.Shovel value when value.Tier == Item.ToolTierDiamond => 5d,
            Item.Shovel value when value.Tier == Item.ToolTierNetherite => 6d,
            Item.Sword value when value.Tier == Item.ToolTierWood => 5d,
            Item.Sword value when value.Tier == Item.ToolTierGold => 5d,
            Item.Sword value when value.Tier == Item.ToolTierStone => 6d,
            Item.Sword value when value.Tier == Item.ToolTierCopper => 6d,
            Item.Sword value when value.Tier == Item.ToolTierIron => 7d,
            Item.Sword value when value.Tier == Item.ToolTierDiamond => 8d,
            Item.Sword value when value.Tier == Item.ToolTierNetherite => 9d,
            _ => 1d,
        };

        internal static bool AllowsAnvilCost(World.Item? item) => item switch
        {
            Item.Axe value when value.Tier == Item.ToolTierWood => true,
            Item.Axe value when value.Tier == Item.ToolTierGold => true,
            Item.Axe value when value.Tier == Item.ToolTierStone => true,
            Item.Axe value when value.Tier == Item.ToolTierCopper => true,
            Item.Axe value when value.Tier == Item.ToolTierIron => true,
            Item.Axe value when value.Tier == Item.ToolTierDiamond => true,
            Item.Axe value when value.Tier == Item.ToolTierNetherite => true,
            Item.Elytra _ => true,
            Item.EnchantedBook _ => true,
            Item.Hoe value when value.Tier == Item.ToolTierWood => true,
            Item.Hoe value when value.Tier == Item.ToolTierGold => true,
            Item.Hoe value when value.Tier == Item.ToolTierStone => true,
            Item.Hoe value when value.Tier == Item.ToolTierCopper => true,
            Item.Hoe value when value.Tier == Item.ToolTierIron => true,
            Item.Hoe value when value.Tier == Item.ToolTierDiamond => true,
            Item.Hoe value when value.Tier == Item.ToolTierNetherite => true,
            Item.Pickaxe value when value.Tier == Item.ToolTierWood => true,
            Item.Pickaxe value when value.Tier == Item.ToolTierGold => true,
            Item.Pickaxe value when value.Tier == Item.ToolTierStone => true,
            Item.Pickaxe value when value.Tier == Item.ToolTierCopper => true,
            Item.Pickaxe value when value.Tier == Item.ToolTierIron => true,
            Item.Pickaxe value when value.Tier == Item.ToolTierDiamond => true,
            Item.Pickaxe value when value.Tier == Item.ToolTierNetherite => true,
            Item.Shovel value when value.Tier == Item.ToolTierWood => true,
            Item.Shovel value when value.Tier == Item.ToolTierGold => true,
            Item.Shovel value when value.Tier == Item.ToolTierStone => true,
            Item.Shovel value when value.Tier == Item.ToolTierCopper => true,
            Item.Shovel value when value.Tier == Item.ToolTierIron => true,
            Item.Shovel value when value.Tier == Item.ToolTierDiamond => true,
            Item.Shovel value when value.Tier == Item.ToolTierNetherite => true,
            Item.Sword value when value.Tier == Item.ToolTierWood => true,
            Item.Sword value when value.Tier == Item.ToolTierGold => true,
            Item.Sword value when value.Tier == Item.ToolTierStone => true,
            Item.Sword value when value.Tier == Item.ToolTierCopper => true,
            Item.Sword value when value.Tier == Item.ToolTierIron => true,
            Item.Sword value when value.Tier == Item.ToolTierDiamond => true,
            Item.Sword value when value.Tier == Item.ToolTierNetherite => true,
            Item.TurtleShell _ => true,
            _ => false,
        };
    }
}
