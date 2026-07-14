# C# plugin architecture

## Direction

- C# NativeAOT is the only plugin language.
- The public namespace mirrors Dragonfly's packages and exported types as closely as C# permits.
- Plugins subclass `Plugin`; generated build plumbing supplies the native entry point and project-name identity.
- The Go host owns Dragonfly and exposes a private flat C ABI. Plugins never use ABI types.
- Code generation reads the pinned Dragonfly Go source with `go/ast` and emits C#; there is no second public API schema.

## Shape

```text
Dragonfly Go API -> Go AST generator -> C# Dragonfly API
                                         |
                                         v
plugin source -> NativeAOT .so -> private C ABI -> Go host -> Dragonfly
```

The ABI is transport, not the API. C# names, interfaces, constructors, and behavior should come from Dragonfly. Hand-written code is limited to marshalling semantics that cannot be inferred from Go types.

## Order

1. NativeAOT loading and `OnEnable`/`OnDisable`.
2. `player.Handler` events. Movement, chat, food loss, jump, teleport, sprint/sneak toggles, punch-air, and quit are implemented.
3. Player methods and commands. Command interfaces and the implemented `Player` method surface are generated from Dragonfly's Go AST. C# uses `Cmd.New`/`Cmd.Register`, one `Cmd.Runnable` per overload, and reflected public fields as Dragonfly uses reflected Go struct fields. Supported command fields include subcommands, native enums, dynamic `Cmd.Enum` values, players, vectors, optional values, and `Cmd.Varargs`. The generator roots runnable fields and field types for NativeAOT; runnable types use `internal` visibility and require no linker annotations. Bedrock-facing subcommands and enum/player suggestions are always lowercase. The generated game-mode slice includes the exact `World.GameMode` interface, four registered values, `GameModeByID`, `GameModeID`, `Player.SetGameMode`, and `Player.GameMode`. Custom C# game modes cross the private ABI as their eight Dragonfly capabilities and remain unregistered, matching raw Dragonfly behavior. The text slice includes `Message`, `SendPopup`, `SendTip`, `SendJukeboxPopup`, `SetNameTag`, and `Disconnect`. `Messagef` remains absent until Go `fmt.Sprintf` semantics can be preserved honestly.
4. World and block parity. The first landed slice generates `Cube.Pos`, `Cube.Range`,
   `Cube.Face`, `World.Block`, `World.SetOpts`, `World.Tx.Range`, `World.Tx.Block`,
   `World.Tx.BlockLoaded`, `World.Tx.BlocksWithin`, `World.Tx.SetBlock`,
   `World.Tx.ScheduleBlockUpdate`,
   `World.Tx.HighestLightBlocker`, `World.Tx.HighestBlock`, `World.Tx.Light`,
   `World.Tx.SkyLight`, `World.Liquid`, `World.Tx.Liquid`, `World.Tx.SetLiquid`, 79 stateless
   block types, `Block.Sand`, `Block.Water`, and `Block.Lava`. The biome slice adds
   `World.Biome`, all 88 registered vanilla biome types, `SetBiome`, `Biome`, `Temperature`,
   `RainingAt`, `SnowingAt`, `ThunderingAt`, `Raining`, `Thundering`, and `CurrentTick`.
   The particle slice adds `World.Particle`, `World.Tx.AddParticle`, all 20 concrete Dragonfly
   particle types, `Color.RGBA`, and all 16 `Sound.Instrument` constructors. Their exported shapes
   and instrument values are AST-validated; particle kinds and payload layout remain private.
   Transaction method signatures and parameter names come from Dragonfly's `world.Tx` Go AST;
   `BlockLoaded` preserves Dragonfly's non-loading query through a C# tuple. `BlocksWithin` maps
   `iter.Seq[cube.Pos]` to lazy `IEnumerable<Cube.Pos>` backed by a transaction-scoped native
   iterator; it is not materialised into a snapshot. `Player.Context : World.Context : World.Tx`,
   so event handlers use world operations directly just like embedded Dragonfly contexts.
   Registered block/liquid states and registered biomes supply canonical private codecs; public
   plugins never handle identifiers, NBT, numeric biome IDs, world handles, iterator handles, or
   host errors. `Liquid` preserves Dragonfly's `(Liquid, bool)` result, and passing `null` to
   `SetLiquid` removes the liquid. Host
   ABI 30 transports that distinction, signed `time.Duration` nanoseconds, private biome IDs,
   particles, registered/custom game-mode capabilities, and the transaction owner's current tick
   without exposing them publicly. Form response callbacks additionally receive a borrowed full
   player snapshot, and ownership transfer guarantees exactly one response or drop callback.
   `ScheduleBlockUpdate`
   maps Go `time.Duration` to C# `TimeSpan`
   and preserves the transaction-owned call. `BuildStructure` remains absent until its synchronous `At` and
   `blockAt` callbacks can be implemented without materialising or changing Dragonfly semantics.
   Remaining stateful blocks, structures, world methods, custom blocks, and block models land
   incrementally.
5. Forms. `Form` interfaces, element fields and memberships, constructors, fluent methods,
   response value types, `Custom`/`Menu`/`Modal`, and `Player.SendForm`/`CloseForm` are generated
   from Dragonfly's `server/player/form` and `player.go` AST. Hand-written `FormCodec` code is
   limited to Dragonfly's reflection, JSON, response-validation, and callback semantics; the
   private ABI transports only JSON plus the borrowed submitting-player snapshot. `Form.Value`
   remains open for plugin-defined forms and exposes byte-oriented `MarshalJSON`/`SubmitJSON`;
   a null response preserves Dragonfly's close signal. `Element` and `MenuElement` retain their
   public JSON-marshalling contract too.
6. Items. The current item slice generates `World.Item`, `Item.ToolTier`, all seven tool-tier
   values, five tiered tools, and 131 concrete item structs. Typed finite stateful families now
   include colours, potions and tipped arrows, banner patterns, smithing templates, suspicious stews,
   pottery sherds, goat horns, and music discs. Dependency factories and encoded states are
   derived from Dragonfly's Go AST and live registries rather than a handwritten schema. Their
   scalar, string, and colour methods are generated from live Dragonfly behavior too; methods
   returning effects wait for the typed effect slice.
   Generated `BookAndQuill`, `WrittenBook`, and `WrittenBookGeneration` mirror Dragonfly's fields,
   page operations, and UTF-8 byte limits. A private bounded LittleEndian NBT codec carries their
   typed state; raw NBT remains absent from the plugin API.
   Generated `Firework`, `FireworkStar`, `FireworkExplosion`, and `FireworkShape` expose typed
   duration, explosion, colour, fade, twinkle, trail, and shape behavior. Their rocket and star
   state uses the same private NBT transport.
   Generated `Armour`, `ArmourTier`, `Helmet`, `Chestplate`, `Leggings`, and `Boots` expose all
   seven Dragonfly armour tiers and all 28 registered piece states. `ArmourTrim` and its 11 typed
   materials include the indirectly discovered `RedstoneWire` item. Piece methods retain
   Dragonfly's defence, toughness, knockback-resistance, enchantment, durability, repair,
   smelting, and trim behavior. The private NBT transport preserves leather dye and trim state.
   Generated `Crossbow(Stack Item)` mirrors Dragonfly's charged-projectile field and its max-count,
   durability, fuel, and enchantment behavior. Its bounded private recursive stack transport
   preserves typed item NBT, plugin values, and enchantments without exposing disk NBT or another
   public stack model. `Fuel` and `FuelInfo` are live-derived for every fuel implementation in this
   slice, including the zero-duration states of tiered tool types.
   Dragonfly's live item registry supplies the private identifier/metadata and capability codecs.
   `Item.Stack` exposes `NewStack`, count/growth/max-count, durability/damage/unbreakable,
   attack damage, custom names, lore, anvil cost, comparison/equality, and stack merging.
   Generated player methods expose `Inventory`, `Armour`, `HeldItems`,
   `SetHeldItems`, and `SetHeldSlot`; `Inventory.Value` exposes `Size`, `Item`, `SetItem`, and
   `AddItem`. C# setters return `void` as the chosen language adaptation, and invalid slots
   throw `ArgumentOutOfRangeException`; host statuses never enter the public API. The existing
   ABI 30 adds one atomic held-item pair snapshot, so `HeldItems` observes the same player state
   with one host read. Bounded open/read/close item snapshots preserve damage, unbreakable state, anvil cost, custom
   names, lore, item NBT, plugin values, and enchantments internally. Unknown registered stateful
   NBT-backed items decode to a private opaque item and round-trip losslessly. Bucket content,
   public enchantment/value mutation, `WithItem`,
   ender chests, custom items, and item events remain next; no public identifier fallback is added.
7. Entities, remaining sounds, and remaining world/block methods.
8. Convert practice-core and expand parity tests against Dragonfly.

Each slice removes the replaced legacy implementation. Unsupported API remains absent rather than gaining a parallel abstraction.

Block access belongs on `World.Tx`. Do not revive `WorldSync`, expose the framework's
`WorldManager` or namespaced world IDs, or restore schema/block-gen as a second public model. Go
AST owns exported shape; Dragonfly's live registry supplies encoded state bytes; the C ABI only
transports calls.

`examples/plugins/kitchen-sink` must use every exposed API. Its `/kitchen block` overload reads the
block below the source through `World.Tx`, writes typed `Block.Sand`, and exercises all three
`World.SetOpts` flags, reads the world range, performs a non-loading block lookup, lazily searches
nearby blocks, reads height/light data, inspects typed water, and writes/removes typed liquid.
It also leaves typed water present and schedules its update with an exact 250 ms delay.
Its separate `/kitchen biome` overload changes and restores a typed biome while exercising every
temperature and weather query in this slice.
`/kitchen tick` reads Dragonfly's transaction-owned current tick; it does not alias world day-time.
`/kitchen particle` emits all 20 particle types and exercises every one of Dragonfly's 16 note
instruments through the transaction-owned `AddParticle` call.
`/kitchen game-mode` exercises registered lookup, player reads, and a custom capability-backed
game mode.
`/kitchen item` exercises stack count, durability, unbreakable, attack damage, anvil cost,
comparison and merging, then round-trips all eleven finite stateful item families plus both typed
book and firework item families through player inventory before restoring all changed player
state. Its firework coverage also exercises typed explosion shapes, colours, fades, twinkle,
trail, off-hand support, and randomised duration. Its armour coverage checks tier, defence,
durability, repair, smelting, trim-material, and private dyed/trim NBT behavior, then round-trips
all 28 tier-and-piece combinations. It also constructs a charged typed crossbow, checks its pure
capabilities, and round-trips the full nested firework stack.
`/kitchen form` exercises reflected menu, custom, and modal forms, every built-in element,
submitted values, closers, and nested sends. `/kitchen raw-form` exercises the open `Form.Value`
contract plus public element/menu-element JSON marshalling.
NativeAOT and host-call tests verify the public shape, lazy iterator cleanup, and transaction-safe
transport. Runtime close destroys plugin instances but deliberately leaves NativeAOT libraries
mapped until process exit: NativeAOT installs process-wide signal handlers, so unloading a library
would leave those handlers pointing into unmapped code. Failed server startup follows the same
lifetime rule and returns its Dragonfly configuration error without crashing during cleanup.

`examples/plugins/vanilla-commands` keeps its plugin entry tiny and one command per file. It currently exercises `/gamemode`, `/help`, `/ping`, and `/position`, and expands with each gameplay parity slice.

Practice remains out of the parity loop until the framework API needed by practice exists. Feature work lands and is tested in this repository first.
