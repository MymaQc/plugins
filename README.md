# bedrock-gophers/plugins

C# NativeAOT plugins for [df-mc/dragonfly](https://github.com/df-mc/dragonfly).

The implemented public C# surface follows Dragonfly. Full Dragonfly parity is still in progress.
The native ABI is internal plumbing, and generated C# API code is derived from Dragonfly's Go AST
rather than a separate public schema.

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

    public override void OnJoin(Player.Context ctx)
    {
        Console.WriteLine($"{ctx.Player().Name()} joined");
    }

    public override void HandleQuit(Player player)
    {
        Console.WriteLine($"{player.Name()} quit");
    }
}
```

The project name is the plugin ID. A compile-time generator emits the hidden native entry point.

Connection admission mirrors Dragonfly's `server.Allower` directly:

```csharp
public override (string Message, bool Allowed) Allow(
    Net.Addr addr,
    Login.IdentityData identity,
    Login.ClientData client)
{
    if (Banned(identity.XUID)) return ("You are banned.", false);
    return (string.Empty, true);
}
```

All plugins allow connections by default. Implementations run in deterministic plugin load order;
the first denial wins. The callback may run concurrently, before a `Player` exists, and Dragonfly's
warning applies: identity data is verified, but client data is controlled by the connecting client.
An empty denial message preserves Dragonfly's hidden disconnect-screen behavior.

`OnJoin` is the one player-lifecycle extension supplied by the host rather than a generated
`player.Handler` method: accepting a player belongs to the server loop in Dragonfly, so its public
handler interface has no join callback. The host emits it after installing its handler. It receives
the same transaction-owned `Player.Context` and may cancel admission. The remaining player
callbacks continue to mirror `player.Handler`.

Current C# slice: loading, lifecycle, reflected commands, decoded packet interception, player text actions, game mode, typed
effects, forms, items and player inventories, server connection admission, all 37 methods in Dragonfly's current
`player.Handler`, and all 13 methods in `world.Handler`. Player callbacks include movement, world
changes, damage/healing/death, mutable respawns and
skins, block and item interactions, entity use/attacks, transfers, command execution, client
diagnostics, and quit. The separate host `OnJoin` lifecycle hook allows first-join initialization
before those generated handler callbacks. World callbacks include liquids, sounds, fire, crops,
leaves, entity spawn/despawn, mutable explosions, full redstone updates, and close. Player and world
handler signatures, command-interface,
player-text, player-state, game-mode, effect, form, item, and player-inventory surfaces are generated
from Dragonfly's Go AST. Generated player state parity currently includes `Food`/`SetFood`,
`Health`, `MaxHealth`/`SetMaxHealth`, level and progress experience accessors, scale accessors,
visibility toggles, mobility toggles, direct `Heal`, and Dragonfly's walking, flight, and
vertical-flight speed getters and setters. Physical-state parity covers absorption, fall distance,
death, ground contact, body heights, and breathing. Sprint, sneak, swim, crawl, glide, and fly
start/stop/state methods also map directly to Dragonfly. Fireproof/on-fire state and current/maximum
air supply use Dragonfly's exact methods, with Go `time.Duration` mapped to C# `TimeSpan`.
Hunger and experience parity includes `AddFood`, `Saturate`, `Exhaust`, total experience,
enchantment seed access/reset, add/remove experience, and collection checks/actions. These call
Dragonfly directly, preserving food-loss and experience-gain handlers, mending, and pickup delay.
`Player.Skin()` and `Player.SetSkin(Skin)` round-trip the complete Dragonfly skin value, including
persona IDs, model data, cape pixels, and animations.
`Player.HideEntity(World.Entity)` and `ShowEntity(World.Entity)` preserve Dragonfly's per-viewer
entity visibility behavior for players, host entities, and plugin-defined entities.
`Player.KnockBack(source, force, height)` calls Dragonfly directly, preserving damage immunity and
armour knockback resistance instead of approximating the result with `SetVelocity`.
`Player.FinalDamageFrom(damage, source)` uses Dragonfly's live armour, enchantments, and resistance
calculation without mutating player health.
`Player.UsingItem()` and `Player.Sleeping()` read Dragonfly's live item-use and bed state directly.
Presentation parity includes instant-respawn and coordinate toggles, sleeping indicators,
dialogue/boss-bar/scoreboard removal, live name/score tags, and toasts, using the exact Dragonfly
`Player` method names.
Block-interaction parity includes AST-generated `BreakBlock`, `ContinueBreaking`, `PickBlock`,
`Sleep`, `StartBreaking`, and `UseItemOnBlock`, with typed `Cube.Pos`, `Cube.Face`, and `Vector3`
arguments passed to Dragonfly inside the active player transaction.
`World.New()` creates a writable in-memory world.
`World.Config.New()` accepts Dragonfly's dimension and runtime settings, and an
`MCDB.Config().Open(...)` provider creates a writable, saveable world below the configured worlds
directory. `World.BlockByName` resolves dynamic names and typed state properties through
Dragonfly's default block registry, which makes external formats such as VAP usable without
exposing raw transport data. Their AST-generated `Name`, `Range`, `HighestLightBlocker`, `Time`,
`SetTime`, `Spawn`, `SetSpawn`, `PlayerSpawn`, `SetPlayerSpawn`, `Save`, and `Close` methods use the
same `World` type as handler callbacks. `Player.ChangeWorld` is the deliberately named host
extension for safe cross-world movement; Dragonfly's `Transfer` still means another server.
`World.Dimension` is the exact implementable Dragonfly interface; custom ranges, liquid behavior,
lava spread duration, weather cycle, and time cycle cross world creation and live reads.
`Player.DeathPosition()` preserves both registered and custom dimensions.
`World.GameMode` includes Dragonfly's four registered values and exact `GameModeByID`/`GameModeID`
lookups. `Player.SetGameMode` accepts custom implementations just like Dragonfly, and
`Player.GameMode` returns their capabilities without exposing the transport descriptor.
Player-backed attack and entity-use targets retain their concrete `Player` type, so normal
`entity is Player target` checks and `target.Name()` work without public handle adapters.
`Title.New(...)` and its immutable `WithSubtitle`, `WithActionText`, and duration methods mirror
Dragonfly's title value API. `Player.SendTitle` carries signed nanosecond durations without the old
unsigned-millisecond truncation.
`Player.HasCooldown(item)` and `SetCooldown(item, duration)` use the generated typed item codec and
call Dragonfly's live cooldown map directly.
`Scoreboard.New(...)` exposes Dragonfly's mutable line, padding, and descending-order API;
`Player.SendScoreboard` sends its raw state and `RemoveScoreboard` clears it.
Forms use Dragonfly's reflected public-field model through `Form.New`, `NewMenu`, and `NewModal`,
with typed elements, submitted values, `Closer`, and callback-owned `World.Tx`. `Form.Value`
remains open for custom implementations, matching Dragonfly's public `form.Form` interface.
`examples/plugins/kitchen-sink` compiles against every exposed C# API and grows with each parity
slice.

Packet plugins override the exact intercept contract names:

```csharp
using Packet = Dragonfly.Packet;

public override void HandleClientPacket(Packet.Context ctx, Packet.Packet packet)
{
    if (packet is Packet.Text text) text.Message = text.Message.Trim();
    if (packet is Packet.CommandRequest command && command.CommandLine.Length == 0) ctx.Cancel();
}

public override void HandleServerPacket(Packet.Context ctx, Packet.Packet packet)
{
    Console.WriteLine($"send {packet.ID()} to {ctx.XUID()}");
}
```

gophertunnel has already decoded these objects; the plugin does not parse raw packet bytes. All
233 registered packet structs are generated from its Go AST. Simple top-level fields are typed and
mutable. Complex protocol fields currently expose lazy `Packet.Value.Json()` while recursive AST
generation lands. Outgoing packets may be inspected or cancelled, but mutation is rejected:
intercept v0.3 does not clone potentially shared broadcast packet objects per connection.

`Player.WritePacket(packet)` forwards a callback-borrowed packet through
`bedrock-gophers/unsafe.WritePacket`. Packet objects expire when their callback returns. Writing
from an outgoing handler re-enters that handler, matching raw Go behavior, so plugins must prevent
recursive forwarding. Owned packet construction is not exposed yet.

Item code uses Dragonfly types, never Minecraft identifiers:

```csharp
var sword = Item.NewStack(new Item.Sword(Item.ToolTierDiamond), 1)
    .WithCustomName("Arena sword")
    .WithLore("Unranked")
    .WithValue("practice:item", "lobby_ffa_selector")
    .WithEnchantments(Item.NewEnchantment(Item.Unbreaking, 1));
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
`Item.Stack` is generated from Dragonfly's `item.Stack` AST, including `WithItem`, `String`,
max-count, durability, unbreakable, attack-damage, anvil-cost, values, comparison, equality, and
stack-merging behavior. Capability tables are derived from the live item implementations; they are
not duplicated in a public adapter model.
`Inventory.Value` currently exposes `Size`, `Item`, `SetItem`, and `AddItem`; player armour
and held-item access use the same typed `Item.Stack`. Invalid C# slot indices throw
`ArgumentOutOfRangeException`; setters return `void`.

World callbacks now carry Dragonfly-shaped transactions. `Player.Context` inherits
`World.Context`, which inherits `World.Tx`; commands receive the same `World.Tx`. `Cube.Pos`,
`World.Block`, `World.Liquid`, `World.Biome`, `World.SetOpts`, and the current `World.Tx` block and
biome surface are generated from Dragonfly source. This includes `Range`, `Block`, `BlockLoaded`,
`Liquid`, `SetLiquid`, `ScheduleBlockUpdate`, `HighestLightBlocker`, `HighestBlock`, `Light`, and
`SkyLight`, `CurrentTick`, `AddParticle`, `PlaySound`, plus 118 concrete non-liquid block types covering 314
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
var currentWorld = tx.World();
var dimension = currentWorld.Dimension();
var cycling = currentWorld.TimeCycle();
currentWorld.StopTime();
if (cycling) currentWorld.StartTime();
currentWorld.SetRequiredSleepDuration(TimeSpan.FromSeconds(1));
var defaultGameMode = currentWorld.DefaultGameMode();
currentWorld.SetDefaultGameMode(defaultGameMode);
currentWorld.SetTickRange(4);
var difficulty = currentWorld.Difficulty();
currentWorld.SetDifficulty(difficulty);
foreach (var entity in tx.Entities()) { /* live World.Entity */ }
foreach (var player in tx.Players().OfType<Player>()) player.Message("Hello, world!");
foreach (var entity in tx.EntitiesWithin(Cube.Box(-16, -16, -16, 16, 16, 16))) { /* nearby */ }
var entity = tx.Entities().First(value => value is not Player);
var handle = tx.RemoveEntity(entity);
var moved = tx.AddEntityAt(handle, source.Position());
tx.AddParticle(source.Position(), new Particle.Flame(new Color.RGBA(255, 96, 32, 255)));
tx.AddParticle(source.Position(), new Particle.Note(Sound.Piano(), 12));
tx.PlaySound(source.Position(), new Sound.Explosion());
player.PlaySound(new Sound.LevelUp());
```

These world state methods keep Dragonfly's exact names and signatures. `World.Dimension()` and
`World.Difficulty()` are emitted as extension methods because C# cannot place a method beside the
same-named nested type; plugin call syntax remains unchanged. The four registered difficulties and
custom structural `World.Difficulty` implementations round-trip through private ABI 47.

`BlocksWithin`, `Entities`, `EntitiesWithin`, and `Players` stay lazy across the private ABI: each C# enumerator owns
a transaction-scoped Dragonfly iterator and closes it on exhaustion, early exit, or callback
completion. `Players` keeps Dragonfly's `IEnumerable<World.Entity>` shape, with player values
materialised as concrete `Player` objects. `EntityHandle.Entity`, `UUID`, `Closed`, and `Close`, plus
`Tx.AddEntity`, `AddEntityAt`, and `RemoveEntity`, are AST-generated. Handles retain stable identity
while the world-bound entity ID changes on remove/re-add. Generic player removal remains blocked;
use `Player.ChangeWorld` so Dragonfly's session transfer completes safely.

Custom entities use Dragonfly's own high-level contracts. Plugins implement `World.EntityType`,
optionally return `World.TickerEntity` from `Open`, and construct worldless handles through
`World.EntitySpawnOpts.New`. `EntityConfig.Apply`, `EntityData`, `BBox`, `DecodeNBT`, and
`EncodeNBT` keep their Dragonfly roles; the native ABI only carries their calls and state. See
`/kitchen custom-entity` for a complete create/add/remove/re-add/close example.

Server-wide player access follows Dragonfly's `Server` directly:

```csharp
var server = Server();
var overworld = server.World();
var nether = server.Nether();
var end = server.End();
var online = server.PlayerCount();
var capacity = server.MaxPlayerCount();
foreach (var player in server.Players(tx))
{
    player.Message("Hello, server!");
    var identity = (player.Name(), player.UUID(), player.XUID());
    var stable = player.H(); // Safe to retain after this iteration.
}
var (byName, found) = server.PlayerByName("Steve");
if (found && byName is not null)
{
    var (byUuid, foundUuid) = server.Player(byName.UUID());
}
var (byXuid, foundXuid) = server.PlayerByXUID("2533274790000000");

var task = overworld.Do(tx =>
{
    // Runs on the overworld owner with a fresh Dragonfly transaction.
    tx.SetBlock(new Cube.Pos(0, 64, 0), new Block.Stone());
});

var temporary = World.New();
var arena = new World.Config
{
    Dim = World.Overworld,
    Provider = new MCDB.Config().Open("arenas/nodebuff"),
    SaveInterval = TimeSpan.FromMinutes(10),
    RandomTickSpeed = -1,
}.New();

var (barrel, validState) = World.BlockByName("minecraft:barrel", new()
{
    ["open_bit"] = (byte)0,
    ["facing_direction"] = 2,
});
if (validState && barrel is not null)
{
    arena.Do(tx => tx.SetBlock(new Cube.Pos(0, 64, 0), barrel));
}
```

`World.Do` and `DoAfter` are AST-generated from Dragonfly. Returned `World.Task` exposes `Done`,
`Err`, `Wait`, `OnDone`, and `Cancel`. The callback runs once on that world's owner. Its `World.Tx`
is borrowed and expires when the callback returns; do not retain it or values borrowed through it.
Calling `Wait` from the same world owner deadlocks, matching Dragonfly. Shutdown rejects new tasks,
cancels pending delays, and drains running callbacks before plugin disable.
`World.New()` uses Dragonfly's `NopProvider`, so it is genuinely in-memory. An MCDB provider is
opened atomically with the world and persists normal `Save()` calls and automatic saves. Provider
paths are relative to the server's worlds directory; traversal, symlink escapes, and opening the
same live database twice are rejected. Set `ReadOnly = true` only when writes must be discarded.
`World.BlockByName` mirrors Dragonfly's `(Block, bool)` lookup. Property values preserve the exact
registry types: C# `bool`, `byte`, `int`, or `string`. Unknown names, missing properties, and
type-mismatched states return `Ok = false` rather than creating an opaque block.

Pass the current `World.Tx` when iterating from a callback or command, and pass `null` outside a
transaction. Iteration is deliberately lazy: a yielded `Player` is valid only inside its current
`foreach` body and expires on the next `MoveNext`, disposal, or callback completion. Do not collect
the players or retain them; retain `Player.H()` values when stable identity is needed. Each body
runs on that player's world owner, matching Dragonfly, so blocking or re-entering the same owner can
deadlock and mirrored scans from different world handlers can deadlock each other. Keep the body
short and avoid nested world-owner calls. `Player`, `PlayerByName`, and `PlayerByXUID` return stable
`World.EntityHandle` values and do not borrow a player transaction.

`World.Handler` is generated from Dragonfly's Go AST and installed on every framework-managed
world. Cancellable callbacks remain allowed by default. `HandleExplosion` may resize, reorder, or
replace both entity and block lists and mutate item-drop chance and fire spawning. Entity callback
values remain concrete `Player` objects where applicable, and `RedstoneUpdate` carries every
Dragonfly field and cause.

Public block, liquid, biome, particle, colour, instrument, sound, and item types come from Dragonfly's Go AST.
Live registries feed internal generated codecs, so Minecraft identifiers, state NBT, numeric biome
IDs, particle kinds, and instrument IDs never enter plugin code. Private host ABI 44 preserves the
separate “no liquid” result, nullable liquid removal, signed nanosecond scheduling delays,
biome/weather queries, particle payloads, registered/custom game-mode capabilities, and full
callback-scoped player snapshots for form responses. Structurally valid form contexts receive
exactly one response or drop callback, including synchronous send failures. World handles,
capability descriptors, and ABI errors also remain private transport details.

All 87 concrete `sound` package types are AST-generated under `Sound.*`. `World.Tx.PlaySound` broadcasts
at a position and `Player.PlaySound` targets one player for those concrete sounds. `HandleSound` receives
their exact exported fields, including stateful blocks and bucket liquids, typed items,
instruments, music discs, goat horns, crossbow stages, and scalar values.
Custom implementations of Dragonfly's `Sound.Play` callback remain unsupported; `World.Sound` is
currently limited to generated concrete values.

The generated effect slice exposes all 28 registered Dragonfly effects, `Effect.Value`, the five
constructors, value methods, colour mixing, registry lookup, and `Player.AddEffect`, `RemoveEffect`,
`Effect`, and `Effects`. ABI 39 carries effect duration, potency, ambient/particle/infinite flags,
and the current tick; C# exposes duration at `TimeSpan`'s 100 ns precision. Custom effect callbacks,
registration, and the remaining concrete effect-specific multiplier methods are not exposed yet;
no identifier-based fallback is exposed.
