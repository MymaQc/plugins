// Code generated from Dragonfly server/world/biome Go AST and registry. DO NOT EDIT.
#nullable enable

namespace Dragonfly
{
    public static partial class Biome
    {
        public readonly record struct Badlands : World.Biome;
        public readonly record struct BadlandsPlateau : World.Biome;
        public readonly record struct BambooJungle : World.Biome;
        public readonly record struct BambooJungleHills : World.Biome;
        public readonly record struct BasaltDeltas : World.Biome;
        public readonly record struct Beach : World.Biome;
        public readonly record struct BirchForest : World.Biome;
        public readonly record struct BirchForestHills : World.Biome;
        public readonly record struct CherryGrove : World.Biome;
        public readonly record struct ColdOcean : World.Biome;
        public readonly record struct CrimsonForest : World.Biome;
        public readonly record struct DarkForest : World.Biome;
        public readonly record struct DarkForestHills : World.Biome;
        public readonly record struct DeepColdOcean : World.Biome;
        public readonly record struct DeepDark : World.Biome;
        public readonly record struct DeepFrozenOcean : World.Biome;
        public readonly record struct DeepLukewarmOcean : World.Biome;
        public readonly record struct DeepOcean : World.Biome;
        public readonly record struct DeepWarmOcean : World.Biome;
        public readonly record struct Desert : World.Biome;
        public readonly record struct DesertHills : World.Biome;
        public readonly record struct DesertLakes : World.Biome;
        public readonly record struct DripstoneCaves : World.Biome;
        public readonly record struct End : World.Biome;
        public readonly record struct ErodedBadlands : World.Biome;
        public readonly record struct FlowerForest : World.Biome;
        public readonly record struct Forest : World.Biome;
        public readonly record struct FrozenOcean : World.Biome;
        public readonly record struct FrozenPeaks : World.Biome;
        public readonly record struct FrozenRiver : World.Biome;
        public readonly record struct GiantSpruceTaigaHills : World.Biome;
        public readonly record struct GiantTreeTaigaHills : World.Biome;
        public readonly record struct GravellyMountainsPlus : World.Biome;
        public readonly record struct Grove : World.Biome;
        public readonly record struct IceSpikes : World.Biome;
        public readonly record struct JaggedPeaks : World.Biome;
        public readonly record struct Jungle : World.Biome;
        public readonly record struct JungleEdge : World.Biome;
        public readonly record struct JungleHills : World.Biome;
        public readonly record struct LegacyFrozenOcean : World.Biome;
        public readonly record struct LukewarmOcean : World.Biome;
        public readonly record struct LushCaves : World.Biome;
        public readonly record struct MangroveSwamp : World.Biome;
        public readonly record struct Meadow : World.Biome;
        public readonly record struct ModifiedBadlandsPlateau : World.Biome;
        public readonly record struct ModifiedJungle : World.Biome;
        public readonly record struct ModifiedJungleEdge : World.Biome;
        public readonly record struct ModifiedWoodedBadlandsPlateau : World.Biome;
        public readonly record struct MountainEdge : World.Biome;
        public readonly record struct MushroomFieldShore : World.Biome;
        public readonly record struct MushroomFields : World.Biome;
        public readonly record struct NetherWastes : World.Biome;
        public readonly record struct Ocean : World.Biome;
        public readonly record struct OldGrowthBirchForest : World.Biome;
        public readonly record struct OldGrowthPineTaiga : World.Biome;
        public readonly record struct OldGrowthSpruceTaiga : World.Biome;
        public readonly record struct PaleGarden : World.Biome;
        public readonly record struct Plains : World.Biome;
        public readonly record struct River : World.Biome;
        public readonly record struct Savanna : World.Biome;
        public readonly record struct SavannaPlateau : World.Biome;
        public readonly record struct ShatteredSavannaPlateau : World.Biome;
        public readonly record struct SnowyBeach : World.Biome;
        public readonly record struct SnowyMountains : World.Biome;
        public readonly record struct SnowyPlains : World.Biome;
        public readonly record struct SnowySlopes : World.Biome;
        public readonly record struct SnowyTaiga : World.Biome;
        public readonly record struct SnowyTaigaHills : World.Biome;
        public readonly record struct SnowyTaigaMountains : World.Biome;
        public readonly record struct SoulSandValley : World.Biome;
        public readonly record struct StonyPeaks : World.Biome;
        public readonly record struct StonyShore : World.Biome;
        public readonly record struct SulfurCaves : World.Biome;
        public readonly record struct SunflowerPlains : World.Biome;
        public readonly record struct Swamp : World.Biome;
        public readonly record struct SwampHills : World.Biome;
        public readonly record struct Taiga : World.Biome;
        public readonly record struct TaigaHills : World.Biome;
        public readonly record struct TaigaMountains : World.Biome;
        public readonly record struct TallBirchHills : World.Biome;
        public readonly record struct WarmOcean : World.Biome;
        public readonly record struct WarpedForest : World.Biome;
        public readonly record struct WindsweptForest : World.Biome;
        public readonly record struct WindsweptGravellyHills : World.Biome;
        public readonly record struct WindsweptHills : World.Biome;
        public readonly record struct WindsweptSavanna : World.Biome;
        public readonly record struct WoodedBadlandsPlateau : World.Biome;
        public readonly record struct WoodedHills : World.Biome;
    }

    internal static class BiomeCodec
    {
        internal static bool TryEncode(World.Biome biome, out int id)
        {
            switch (biome)
            {
                case Biome.Badlands _:
                    id = 37; return true;
                case Biome.BadlandsPlateau _:
                    id = 39; return true;
                case Biome.BambooJungle _:
                    id = 48; return true;
                case Biome.BambooJungleHills _:
                    id = 49; return true;
                case Biome.BasaltDeltas _:
                    id = 181; return true;
                case Biome.Beach _:
                    id = 16; return true;
                case Biome.BirchForest _:
                    id = 27; return true;
                case Biome.BirchForestHills _:
                    id = 28; return true;
                case Biome.CherryGrove _:
                    id = 192; return true;
                case Biome.ColdOcean _:
                    id = 44; return true;
                case Biome.CrimsonForest _:
                    id = 179; return true;
                case Biome.DarkForest _:
                    id = 29; return true;
                case Biome.DarkForestHills _:
                    id = 157; return true;
                case Biome.DeepColdOcean _:
                    id = 45; return true;
                case Biome.DeepDark _:
                    id = 190; return true;
                case Biome.DeepFrozenOcean _:
                    id = 47; return true;
                case Biome.DeepLukewarmOcean _:
                    id = 43; return true;
                case Biome.DeepOcean _:
                    id = 24; return true;
                case Biome.DeepWarmOcean _:
                    id = 41; return true;
                case Biome.Desert _:
                    id = 2; return true;
                case Biome.DesertHills _:
                    id = 17; return true;
                case Biome.DesertLakes _:
                    id = 130; return true;
                case Biome.DripstoneCaves _:
                    id = 188; return true;
                case Biome.End _:
                    id = 9; return true;
                case Biome.ErodedBadlands _:
                    id = 165; return true;
                case Biome.FlowerForest _:
                    id = 132; return true;
                case Biome.Forest _:
                    id = 4; return true;
                case Biome.FrozenOcean _:
                    id = 46; return true;
                case Biome.FrozenPeaks _:
                    id = 183; return true;
                case Biome.FrozenRiver _:
                    id = 11; return true;
                case Biome.GiantSpruceTaigaHills _:
                    id = 161; return true;
                case Biome.GiantTreeTaigaHills _:
                    id = 33; return true;
                case Biome.GravellyMountainsPlus _:
                    id = 162; return true;
                case Biome.Grove _:
                    id = 185; return true;
                case Biome.IceSpikes _:
                    id = 140; return true;
                case Biome.JaggedPeaks _:
                    id = 182; return true;
                case Biome.Jungle _:
                    id = 21; return true;
                case Biome.JungleEdge _:
                    id = 23; return true;
                case Biome.JungleHills _:
                    id = 22; return true;
                case Biome.LegacyFrozenOcean _:
                    id = 10; return true;
                case Biome.LukewarmOcean _:
                    id = 42; return true;
                case Biome.LushCaves _:
                    id = 187; return true;
                case Biome.MangroveSwamp _:
                    id = 191; return true;
                case Biome.Meadow _:
                    id = 186; return true;
                case Biome.ModifiedBadlandsPlateau _:
                    id = 167; return true;
                case Biome.ModifiedJungle _:
                    id = 149; return true;
                case Biome.ModifiedJungleEdge _:
                    id = 151; return true;
                case Biome.ModifiedWoodedBadlandsPlateau _:
                    id = 166; return true;
                case Biome.MountainEdge _:
                    id = 20; return true;
                case Biome.MushroomFieldShore _:
                    id = 15; return true;
                case Biome.MushroomFields _:
                    id = 14; return true;
                case Biome.NetherWastes _:
                    id = 8; return true;
                case Biome.Ocean _:
                    id = 0; return true;
                case Biome.OldGrowthBirchForest _:
                    id = 155; return true;
                case Biome.OldGrowthPineTaiga _:
                    id = 32; return true;
                case Biome.OldGrowthSpruceTaiga _:
                    id = 160; return true;
                case Biome.PaleGarden _:
                    id = 193; return true;
                case Biome.Plains _:
                    id = 1; return true;
                case Biome.River _:
                    id = 7; return true;
                case Biome.Savanna _:
                    id = 35; return true;
                case Biome.SavannaPlateau _:
                    id = 36; return true;
                case Biome.ShatteredSavannaPlateau _:
                    id = 164; return true;
                case Biome.SnowyBeach _:
                    id = 26; return true;
                case Biome.SnowyMountains _:
                    id = 13; return true;
                case Biome.SnowyPlains _:
                    id = 12; return true;
                case Biome.SnowySlopes _:
                    id = 184; return true;
                case Biome.SnowyTaiga _:
                    id = 30; return true;
                case Biome.SnowyTaigaHills _:
                    id = 31; return true;
                case Biome.SnowyTaigaMountains _:
                    id = 158; return true;
                case Biome.SoulSandValley _:
                    id = 178; return true;
                case Biome.StonyPeaks _:
                    id = 189; return true;
                case Biome.StonyShore _:
                    id = 25; return true;
                case Biome.SulfurCaves _:
                    id = 194; return true;
                case Biome.SunflowerPlains _:
                    id = 129; return true;
                case Biome.Swamp _:
                    id = 6; return true;
                case Biome.SwampHills _:
                    id = 134; return true;
                case Biome.Taiga _:
                    id = 5; return true;
                case Biome.TaigaHills _:
                    id = 19; return true;
                case Biome.TaigaMountains _:
                    id = 133; return true;
                case Biome.TallBirchHills _:
                    id = 156; return true;
                case Biome.WarmOcean _:
                    id = 40; return true;
                case Biome.WarpedForest _:
                    id = 180; return true;
                case Biome.WindsweptForest _:
                    id = 34; return true;
                case Biome.WindsweptGravellyHills _:
                    id = 131; return true;
                case Biome.WindsweptHills _:
                    id = 3; return true;
                case Biome.WindsweptSavanna _:
                    id = 163; return true;
                case Biome.WoodedBadlandsPlateau _:
                    id = 38; return true;
                case Biome.WoodedHills _:
                    id = 18; return true;
                case EncodedBiome encoded:
                    id = encoded.Id; return true;
                default:
                    id = 0; return false;
            }
        }

        internal static World.Biome Decode(int id)
        {
            if (id == 37) return new Biome.Badlands();
            if (id == 39) return new Biome.BadlandsPlateau();
            if (id == 48) return new Biome.BambooJungle();
            if (id == 49) return new Biome.BambooJungleHills();
            if (id == 181) return new Biome.BasaltDeltas();
            if (id == 16) return new Biome.Beach();
            if (id == 27) return new Biome.BirchForest();
            if (id == 28) return new Biome.BirchForestHills();
            if (id == 192) return new Biome.CherryGrove();
            if (id == 44) return new Biome.ColdOcean();
            if (id == 179) return new Biome.CrimsonForest();
            if (id == 29) return new Biome.DarkForest();
            if (id == 157) return new Biome.DarkForestHills();
            if (id == 45) return new Biome.DeepColdOcean();
            if (id == 190) return new Biome.DeepDark();
            if (id == 47) return new Biome.DeepFrozenOcean();
            if (id == 43) return new Biome.DeepLukewarmOcean();
            if (id == 24) return new Biome.DeepOcean();
            if (id == 41) return new Biome.DeepWarmOcean();
            if (id == 2) return new Biome.Desert();
            if (id == 17) return new Biome.DesertHills();
            if (id == 130) return new Biome.DesertLakes();
            if (id == 188) return new Biome.DripstoneCaves();
            if (id == 9) return new Biome.End();
            if (id == 165) return new Biome.ErodedBadlands();
            if (id == 132) return new Biome.FlowerForest();
            if (id == 4) return new Biome.Forest();
            if (id == 46) return new Biome.FrozenOcean();
            if (id == 183) return new Biome.FrozenPeaks();
            if (id == 11) return new Biome.FrozenRiver();
            if (id == 161) return new Biome.GiantSpruceTaigaHills();
            if (id == 33) return new Biome.GiantTreeTaigaHills();
            if (id == 162) return new Biome.GravellyMountainsPlus();
            if (id == 185) return new Biome.Grove();
            if (id == 140) return new Biome.IceSpikes();
            if (id == 182) return new Biome.JaggedPeaks();
            if (id == 21) return new Biome.Jungle();
            if (id == 23) return new Biome.JungleEdge();
            if (id == 22) return new Biome.JungleHills();
            if (id == 10) return new Biome.LegacyFrozenOcean();
            if (id == 42) return new Biome.LukewarmOcean();
            if (id == 187) return new Biome.LushCaves();
            if (id == 191) return new Biome.MangroveSwamp();
            if (id == 186) return new Biome.Meadow();
            if (id == 167) return new Biome.ModifiedBadlandsPlateau();
            if (id == 149) return new Biome.ModifiedJungle();
            if (id == 151) return new Biome.ModifiedJungleEdge();
            if (id == 166) return new Biome.ModifiedWoodedBadlandsPlateau();
            if (id == 20) return new Biome.MountainEdge();
            if (id == 15) return new Biome.MushroomFieldShore();
            if (id == 14) return new Biome.MushroomFields();
            if (id == 8) return new Biome.NetherWastes();
            if (id == 0) return new Biome.Ocean();
            if (id == 155) return new Biome.OldGrowthBirchForest();
            if (id == 32) return new Biome.OldGrowthPineTaiga();
            if (id == 160) return new Biome.OldGrowthSpruceTaiga();
            if (id == 193) return new Biome.PaleGarden();
            if (id == 1) return new Biome.Plains();
            if (id == 7) return new Biome.River();
            if (id == 35) return new Biome.Savanna();
            if (id == 36) return new Biome.SavannaPlateau();
            if (id == 164) return new Biome.ShatteredSavannaPlateau();
            if (id == 26) return new Biome.SnowyBeach();
            if (id == 13) return new Biome.SnowyMountains();
            if (id == 12) return new Biome.SnowyPlains();
            if (id == 184) return new Biome.SnowySlopes();
            if (id == 30) return new Biome.SnowyTaiga();
            if (id == 31) return new Biome.SnowyTaigaHills();
            if (id == 158) return new Biome.SnowyTaigaMountains();
            if (id == 178) return new Biome.SoulSandValley();
            if (id == 189) return new Biome.StonyPeaks();
            if (id == 25) return new Biome.StonyShore();
            if (id == 194) return new Biome.SulfurCaves();
            if (id == 129) return new Biome.SunflowerPlains();
            if (id == 6) return new Biome.Swamp();
            if (id == 134) return new Biome.SwampHills();
            if (id == 5) return new Biome.Taiga();
            if (id == 19) return new Biome.TaigaHills();
            if (id == 133) return new Biome.TaigaMountains();
            if (id == 156) return new Biome.TallBirchHills();
            if (id == 40) return new Biome.WarmOcean();
            if (id == 180) return new Biome.WarpedForest();
            if (id == 34) return new Biome.WindsweptForest();
            if (id == 131) return new Biome.WindsweptGravellyHills();
            if (id == 3) return new Biome.WindsweptHills();
            if (id == 163) return new Biome.WindsweptSavanna();
            if (id == 38) return new Biome.WoodedBadlandsPlateau();
            if (id == 18) return new Biome.WoodedHills();
            return new EncodedBiome(id);
        }

        private sealed record EncodedBiome(int Id) : World.Biome;
    }
}
