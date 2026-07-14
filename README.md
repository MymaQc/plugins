# bedrock-gophers/plugins

C# NativeAOT plugins for [df-mc/dragonfly](https://github.com/df-mc/dragonfly).

The public C# API follows Dragonfly. The native ABI is internal plumbing, and generated C# API code is derived from Dragonfly's Go AST rather than a separate public schema.

## Run

Requirements: Go 1.26+, .NET 10 SDK, and a C compiler/linker.

```shell
make run
```

`make run` always regenerates the AST-derived C# API, republishes the runtime and every
`examples/plugins/*/*.csproj` as NativeAOT, replaces staged example `.so` files, then starts the
server. It does not reuse an older plugin build.

Consumer repositories keep plugins in `plugins/`; see the `minimal` branch for the local and
Docker setup. The loader also accepts compatible precompiled `.so` files. Source-owned build
commands replace their staged output, so precompiled files are staged after that build or run
directly with the server.

## Plugin

Plugin code does not declare an ID, shared-library path, or native entry point:

```csharp
using Dragonfly;

public sealed class Example : Plugin
{
    public override void OnEnable() => Console.WriteLine("enabled");
    public override void OnDisable() => Console.WriteLine("disabled");

    public override void HandleQuit(Player player)
    {
        Console.WriteLine($"{player.Name()} quit");
    }
}
```

The project name is the plugin ID. A compile-time generator emits the hidden native entry point.

Current C# slice: loading, lifecycle, reflected commands, player text actions, game mode, typed
effects, forms, items and player inventories, and all 37 methods in Dragonfly's current
`player.Handler`. This includes movement, world changes, damage/healing/death, mutable respawns and
skins, block and item interactions, entity use/attacks, transfers, command execution, client
diagnostics, and quit. Player handler, command-interface,
player-text, player-state, game-mode, effect, form, item, and player-inventory surfaces are generated
from Dragonfly's Go AST. Generated player state parity currently includes `Food`/`SetFood`,
`Health`, `MaxHealth`/`SetMaxHealth`, level and progress experience accessors, scale accessors,
visibility toggles, and mobility toggles.
`World.GameMode` includes Dragonfly's four registered values and exact `GameModeByID`/`GameModeID`
lookups. `Player.SetGameMode` accepts custom implementations just like Dragonfly, and
`Player.GameMode` returns their capabilities without exposing the transport descriptor.
Forms use Dragonfly's reflected public-field model through `Form.New`, `NewMenu`, and `NewModal`,
with typed elements, submitted values, `Closer`, and callback-owned `World.Tx`. `Form.Value`
remains open for custom implementations, matching Dragonfly's public `form.Form` interface.
`examples/plugins/kitchen-sink` compiles against every exposed C# API and grows with each parity
slice.

Item code uses Dragonfly types, never Minecraft identifiers:

```csharp
var sword = Item.NewStack(new Item.Sword(Item.ToolTierDiamond), 1)
    .WithCustomName("Arena sword")
    .WithLore("Unranked");
var inventory = player.Inventory();
var enderChest = player.EnderChestInventory();
var previous = inventory.Item(0);
inventory.SetItem(0, sword);
enderChest.SetItem(0, sword);
var (mainHand, offHand) = player.HeldItems();
player.SetHeldItems(sword, offHand);
```

The current generated item slice contains 132 concrete Dragonfly item structs: stateless items,
boolean variants, five tool types, the four tiered armour pieces, and the finite stateful families
for colours, potions, banner patterns, smithing templates, suspicious stews, pottery sherds, goat
horns, and music discs. `RedstoneWire` is also generated as a typed trim material even though it is
not directly registered as an item implementation.
Their exact factory values come from Dragonfly's Go AST and live registries. NBT-backed item
families that have not landed remain opaque during transport; raw identifiers,
metadata, NBT, enchantment IDs, snapshots, and host statuses stay private.
Generated value methods include colour conversions, numeric IDs, horn names, and music-disc
identifiers, display names, and authors. Potions and suspicious stews expose their exact typed
Dragonfly effect lists through `Effects`, `Potion.All`/`From`, and `Item.StewTypes`.
`BookAndQuill`, `WrittenBook`, and `WrittenBookGeneration` mirror Dragonfly fields and page
operations. Private bounded LittleEndian NBT transport preserves typed pages, title, author, and
generation without exposing NBT to plugins.
`Firework`, `FireworkStar`, `FireworkExplosion`, and `FireworkShape` likewise expose Dragonfly's
typed duration, explosions, colours, fades, effects, and shape behavior. The same private NBT
transport preserves rocket and star state without exposing identifiers or encoded tags.
`Armour`, `ArmourTier`, `Helmet`, `Chestplate`, `Leggings`, and `Boots` mirror Dragonfly's concrete
`ArmourTierLeather`, `ArmourTierCopper`, `ArmourTierGold`, `ArmourTierChain`, `ArmourTierIron`,
`ArmourTierDiamond`, and `ArmourTierNetherite` tiers and all 28 registered armour states.
`ArmourTrim` and `ArmourTrimMaterial` cover typed amethyst, copper, diamond, emerald, gold, iron,
lapis, netherite, quartz, resin, and redstone materials. Generated pieces expose
`DefencePoints`, `Toughness`, `KnockBackResistance`, `EnchantmentValue`, `DurabilityInfo`,
`RepairableBy`, `SmeltInfo`, and `WithTrim`. Private NBT preserves leather dye and trim state
without exposing encoded tags.
`Crossbow` carries its charged projectile as a full typed `Item.Stack`, matching Dragonfly's field.
Its max-count, durability, fuel, and enchantment values are generated from Dragonfly. A bounded
private recursive transport preserves charged item state, including typed item NBT, stack values,
and enchantments, without exposing disk NBT or an adapter type. Generated `Fuel`/`FuelInfo` also
cover every fuel implementation in the current item slice, including zero-duration non-wood tool
states whose concrete Dragonfly types still implement `Fuel`.
`Bucket` and its private-state `BucketContent` mirror Dragonfly's empty, water, lava, and milk
states through typed content factories. Pure count, consumption, duration, empty, and fuel behavior
is preserved; lava burns for 1000 seconds and leaves a typed empty bucket residue.
`Item.Stack` also exposes Dragonfly's generated max-count, durability, unbreakable, attack-damage,
anvil-cost, comparison, equality, and stack-merging behavior. Capability tables are derived from
the live item implementations; they are not duplicated in a public adapter model.
`Inventory.Value` currently exposes `Size`, `Item`, `SetItem`, and `AddItem`; player armour
and held-item access use the same typed `Item.Stack`. Invalid C# slot indices throw
`ArgumentOutOfRangeException`; setters return `void`.

World callbacks now carry Dragonfly-shaped transactions. `Player.Context` inherits
`World.Context`, which inherits `World.Tx`; commands receive the same `World.Tx`. `Cube.Pos`,
`World.Block`, `World.Liquid`, `World.Biome`, `World.SetOpts`, and the current `World.Tx` block and
biome surface are generated from Dragonfly source. This includes `Range`, `Block`, `BlockLoaded`,
`Liquid`, `SetLiquid`, `ScheduleBlockUpdate`, `HighestLightBlocker`, `HighestBlock`, `Light`, and
`SkyLight`, `CurrentTick`, `AddParticle`, plus 112 concrete non-liquid block types covering 306
canonical primitive-state registry entries, `Block.Water`, `Block.Lava`, all 88 registered vanilla
biome types, and all 20 Dragonfly particle types with typed colours, blocks, faces, positions, and
note instruments. Promoted Dragonfly fields remain visible, so crops use typed growth stages:

```csharp
var pos = Cube.PosFromVec3(source.Position()).Side(Cube.Face.Down);
var (block, loaded) = tx.BlockLoaded(pos);
World.Block? previous = loaded ? block : tx.Block(pos);
Cube.Range bounds = tx.Range();
var nearby = tx.BlocksWithin(pos, 8, new Block.Sand());
tx.SetBlock(pos, new Block.Sand());
tx.SetBlock(pos, new Block.WheatSeeds(Growth: 7));
var (liquid, found) = tx.Liquid(pos);
tx.SetLiquid(pos, new Block.Water(Still: true, Depth: 8, Falling: false));
tx.SetLiquid(pos, null); // Remove the liquid.
var water = new Block.Water(Still: true, Depth: 8, Falling: false);
tx.SetLiquid(pos, water);
tx.ScheduleBlockUpdate(pos, water, TimeSpan.FromMilliseconds(250));
var previousBiome = tx.Biome(pos);
tx.SetBiome(pos, new Biome.Desert());
var temperature = tx.Temperature(pos);
var rainingHere = tx.RainingAt(pos);
tx.SetBiome(pos, previousBiome);
var tick = tx.CurrentTick();
tx.AddParticle(source.Position(), new Particle.Flame(new Color.RGBA(255, 96, 32, 255)));
tx.AddParticle(source.Position(), new Particle.Note(Sound.Piano(), 12));
```

`BlocksWithin` stays lazy across the private ABI: each C# enumerator owns a transaction-scoped
Dragonfly iterator and closes it on exhaustion, early exit, or callback completion.

Public block, liquid, biome, particle, colour, instrument, and item types come from Dragonfly's Go AST.
Live registries feed internal generated codecs, so Minecraft identifiers, state NBT, numeric biome
IDs, particle kinds, and instrument IDs never enter plugin code. Private host ABI 32 preserves the
separate “no liquid” result, nullable liquid removal, signed nanosecond scheduling delays,
biome/weather queries, particle payloads, registered/custom game-mode capabilities, and full
callback-scoped player snapshots for form responses. Structurally valid form contexts receive
exactly one response or drop callback, including synchronous send failures. World handles,
capability descriptors, and ABI errors also remain private transport details.

The generated effect slice exposes all 28 registered Dragonfly effects, `Effect.Value`, the five
constructors, value methods, colour mixing, registry lookup, and `Player.AddEffect`, `RemoveEffect`,
`Effect`, and `Effects`. ABI 32 carries effect duration, potency, ambient/particle/infinite flags,
and the current tick; C# exposes duration at `TimeSpan`'s 100 ns precision. Custom effect callbacks,
registration, and concrete effect-specific multiplier methods wait for the entity and damage-source
slices; no identifier-based fallback is exposed.
