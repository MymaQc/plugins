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

1. NativeAOT loading and `OnEnable`/`OnDisable`. The host also exposes `OnJoin(Player.Context)` as
   an explicit lifecycle extension. It is not presented as AST-generated `player.Handler`: raw
   Dragonfly leaves player acceptance to the server loop and that interface has no join method. The
   host invokes the callback after installing its handler. It is transaction-owned, cancellable,
   and subscribed only when overridden.
2. `player.Handler` events. All 37 methods in the pinned Dragonfly interface are generated and
   transported: movement, chat, world changes, damage/healing/death, respawn, skin changes, every
   block/item interaction, entity use/attacks, transfer, command execution, diagnostics, and quit.
   Signatures, order, and subscription bits come from Dragonfly's Go AST; generator tests fail on
   any unknown upstream method instead of silently omitting it. Private plugin ABI 7 gives every
   callback a borrowed full player snapshot, transports stateful blocks, items, skins, source
   interfaces, worlds, entities, UDP addresses, command metadata/arguments, and all nine diagnostics
   fields. Mutable block-break drops, item-pickup replacements, transfer addresses, and same-count
   command arguments use callback-owned views with an exact-once drop callback. C# item-release
   durations expose `TimeSpan`'s 100 ns precision. Unchanged hurt-immunity values preserve their
   exact signed Go nanoseconds; values mutated by C# are rounded to `TimeSpan` precision. Signed
   durations, including negative values, remain signed. Cancellation remains
   allowed by default and can only transition to cancelled.
3. Player methods and commands. Command interfaces and the implemented `Player` method surface are generated from Dragonfly's Go AST. C# uses `Cmd.New`/`Cmd.Register`, one `Cmd.Runnable` per overload, and reflected public fields as Dragonfly uses reflected Go struct fields. Supported command fields include subcommands, native enums, dynamic `Cmd.Enum` values, players, vectors, optional values, and `Cmd.Varargs`. The generator roots runnable fields and field types for NativeAOT; runnable types use `internal` visibility and require no linker annotations. Bedrock-facing subcommands and enum/player suggestions are always lowercase. The generated game-mode slice includes the exact `World.GameMode` interface, four registered values, `GameModeByID`, `GameModeID`, `Player.SetGameMode`, and `Player.GameMode`. Custom C# game modes cross the private ABI as their eight Dragonfly capabilities and remain unregistered, matching raw Dragonfly behavior. The text slice includes `Message`, `SendPopup`, `SendTip`, `SendJukeboxPopup`, `SetNameTag`, and `Disconnect`. The state slice adds the exact 17 `Food`, health, experience-level/progress, scale, visibility, and mobility methods. The kinematics slice adds AST-generated `Teleport`, `Move`, `Displace`, `Position`, `Velocity`, `SetVelocity`, and `Rotation`; the host invokes those exact Dragonfly methods and does not reject non-finite values that raw Dragonfly accepts. Setters return `void`; private host status never enters the public API. `SetMaxHealth` preserves Dragonfly's clamp-to-one behavior for non-positive values. `Messagef` remains absent until Go `fmt.Sprintf` semantics can be preserved honestly.
4. World and block parity. The first landed slice generates `Cube.Pos`, `Cube.Range`,
   `Cube.Face`, `World.Block`, `World.SetOpts`, `World.Tx.Range`, `World.Tx.Block`,
   `World.Tx.BlockLoaded`, `World.Tx.BlocksWithin`, `World.Tx.SetBlock`,
   `World.Tx.ScheduleBlockUpdate`,
   `World.Tx.HighestLightBlocker`, `World.Tx.HighestBlock`, `World.Tx.Light`,
   `World.Tx.SkyLight`, `World.Liquid`, `World.Tx.Liquid`, `World.Tx.SetLiquid`, 112 non-liquid
   block types covering all 306 registered states whose varying registry-state fields are primitive,
   plus `Block.Water` and `Block.Lava`. Promoted fields are expanded through Dragonfly's AST, so
   all eight growth stages of `BeetrootSeeds`, `Carrot`, `Potato`, and `WheatSeeds` remain typed.
   The biome slice adds
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
   ABI 32 transports that distinction, signed `time.Duration` nanoseconds, private biome IDs,
   particles, registered/custom game-mode capabilities, and the transaction owner's current tick
   without exposing them publicly. Form response callbacks additionally receive a borrowed full
   player snapshot, and ownership transfer guarantees exactly one response or drop callback.
   `ScheduleBlockUpdate`
   maps Go `time.Duration` to C# `TimeSpan`
   and preserves the transaction-owned call. `BuildStructure` remains absent until its synchronous `At` and
   `blockAt` callbacks can be implemented without materialising or changing Dragonfly semantics.
   Descriptor-backed, nested, and NBT-backed block types, structures, remaining world methods,
   custom blocks, and block models land incrementally. Unsupported block shapes remain absent;
   they do not gain a public identifier/NBT escape hatch.
5. Forms. `Form` interfaces, element fields and memberships, constructors, fluent methods,
   response value types, `Custom`/`Menu`/`Modal`, and `Player.SendForm`/`CloseForm` are generated
   from Dragonfly's `server/player/form` and `player.go` AST. Hand-written `FormCodec` code is
   limited to Dragonfly's reflection, JSON, response-validation, and callback semantics; the
   private ABI transports only JSON plus the borrowed submitting-player snapshot. `Form.Value`
   remains open for plugin-defined forms and exposes byte-oriented `MarshalJSON`/`SubmitJSON`;
   a null response preserves Dragonfly's close signal. `Element` and `MenuElement` retain their
   public JSON-marshalling contract too.
6. Items. The current item slice generates `World.Item`, `Item.ToolTier`, all seven tool-tier
   values, five tiered tools, and 132 concrete item structs. Typed finite stateful families now
   include colours, potions and tipped arrows, banner patterns, smithing templates, suspicious stews,
   pottery sherds, goat horns, and music discs. Dependency factories and encoded states are
   derived from Dragonfly's Go AST and live registries rather than a handwritten schema. Their
   scalar, string, colour, and typed effect methods are generated from live Dragonfly behavior too.
   `Potion.Effects`, `Potion.All`/`From`, `StewType.Effects`, and `StewTypes` preserve Dragonfly's
   exact lists and ordering.
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
   Generated `Bucket(BucketContent Content)` and its typed liquid/milk factories preserve all four
   registered empty, water, lava, and milk states. Pure count, consumption, duration, empty, and
   fuel behavior matches Dragonfly, including lava's typed empty-bucket residue. Runtime consume
   and block-use methods wait for the `Consumer`, `User`, and `UseContext` slices. Liquid types
   registered only by the host at runtime still need their semantic `LiquidType` transported across
   the ABI; the private opaque fallback fails explicitly instead of treating a block identifier as
   a liquid type.
   Dragonfly's live item registry supplies the private identifier/metadata and capability codecs.
   `Item.Stack` exposes `NewStack`, count/growth/max-count, durability/damage/unbreakable,
   attack damage, custom names, lore, anvil cost, comparison/equality, and stack merging.
   Generated player methods expose `Inventory`, `EnderChestInventory`, `Armour`, `HeldItems`,
   `SetHeldItems`, and `SetHeldSlot`; `Inventory.Value` exposes `Size`, `Item`, `SetItem`, and
   `AddItem`. C# setters return `void` as the chosen language adaptation, and invalid slots
   throw `ArgumentOutOfRangeException`; host statuses never enter the public API. The existing
   ABI 32 includes one atomic held-item pair snapshot, so `HeldItems` observes the same player state
   with one host read. Main and ender-chest inventory sizes are read from the live Dragonfly
   inventory, preserving custom `player.Config` sizes. Bounded open/read/close item snapshots preserve damage, unbreakable state, anvil cost, custom
   names, lore, item NBT, plugin values, and enchantments internally. `Stack.WithValue`, `Value`,
   and `Values` expose the cross-language NBT-compatible value set. The generator discovers all 27
   registered enchantments from Dragonfly's AST and live registry; normal and forced addition,
   removal, lookup, deterministic listing, item compatibility, and enchantment compatibility mirror
   `item.Stack`. Unknown registered stateful
   NBT-backed items decode to a private opaque item and round-trip losslessly. Public
   `WithItem` and custom items remain next; no public identifier
   fallback is added.
7. Effects. The generator reads `server/entity/effect` registrations, interfaces, constructors,
   value methods, and player method signatures from Dragonfly's Go AST, then validates all 28
   built-ins against the live registry. C# exposes `Effect.Value`, registered `Type`/`LastingType`
   values, all five constructors, `ResultingColour`, `ByID`/`ID`, and the four player effect methods.
   ABI 32 transports signed nanosecond duration, level, potency, ambient/particle/infinite flags, and
   tick. C# `TimeSpan` has 100 ns precision and rejects snapshots outside that precision instead of
   truncating them. Re-adding a snapshot is bounded to one million elapsed ticks because Dragonfly
   exposes `Tick` but no constructor or setter for it. Pending initial instant-effect potency is
   normalised because Dragonfly exposes no potency getter. Custom `Type.Apply`, `LastingType.Start`/`End`,
   `Register`, and concrete multiplier methods wait for entity and damage-source callbacks instead
   of receiving a second abstraction.
8. Entities, remaining sounds, and remaining world/block methods. The entity foundation now
   generates Dragonfly's exact `World.Entity` interface (`Close`, `H`, `Position`, `Rotation`)
   from `server/world/entity.go`. Handler entities are live host-backed values, not public ID
   tokens; `EntityHandle` retains stable identity without conflating handle closure with entity
   despawning. `Player` implements `World.Entity`, as it does in
   Dragonfly. Entity IDs, generations, state buffers, and host status codes remain private.
   Next entity slices add the remaining `EntityHandle` surface, `entity.Ent` capabilities,
   transaction add/remove/list methods, then plugin-defined entity types and callbacks.
9. Convert practice-core and expand parity tests against Dragonfly.

## FFA parity

The combat callback foundation is present: hurt, attack, death, respawn, typed damage sources,
mutable damage/knockback/hit delay, live entity transforms, player kinematics, and direct player
teleportation and healing. `OnJoin` supplies the missing lobby-entry lifecycle without claiming to
be an upstream handler method. The host's explicit managed-MCDB extension now covers lookup/open,
world spawn, save/close, and safe `Player.ChangeWorld`; exact `World` instance methods remain
AST-generated. Stack values and typed enchantments cover selector metadata and Protection kits.

Remaining raw-Dragonfly parity work is concentrated in the world/entity transaction slice:

- `Tx.World`, lazy entity/player iteration, and reusable-handle `AddEntity`, `AddEntityAt`, and
  `RemoveEntity` semantics;
- concrete player identity/name resolution for attacked entity handles;
- arbitrary Dragonfly `World.Config`, provider, generator, and entity-type construction beyond the
  bounded managed-MCDB bootstrap extension.

Dragonfly's `HandleTransfer` remains a transfer to another server address. It must not be reused or
documented as cross-world movement.

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
Its separate `/kitchen crop` overload decodes and writes `Block.WheatSeeds(Growth: 7)`, proving
that promoted private Go embeddings round-trip through the typed C# API.
Its separate `/kitchen biome` overload changes and restores a typed biome while exercising every
temperature and weather query in this slice.
`/kitchen tick` reads Dragonfly's transaction-owned current tick; it does not alias world day-time.
`/kitchen particle` emits all 20 particle types and exercises every one of Dragonfly's 16 note
instruments through the transaction-owned `AddParticle` call.
`/kitchen game-mode` exercises registered lookup, player reads, and a custom capability-backed
game mode.
`/kitchen state` exercises all 17 generated food, health, experience-level/progress, scale,
visibility, and mobility methods without changing the final player state.
`/kitchen item` exercises stack count, durability, unbreakable, attack damage, anvil cost,
comparison and merging, then round-trips all eleven finite stateful item families plus both typed
book and firework item families through player inventory before restoring all changed player
state. Its firework coverage also exercises typed explosion shapes, colours, fades, twinkle,
trail, off-hand support, and randomised duration. Its armour coverage checks tier, defence,
durability, repair, smelting, trim-material, and private dyed/trim NBT behavior, then round-trips
all 28 tier-and-piece combinations. It also constructs a charged typed crossbow, checks its pure
capabilities, and round-trips the full nested firework stack. Empty, water, lava, and milk buckets
exercise typed content queries, consumption flags, duration, max counts, and lava fuel residue.
`/kitchen effect` exercises every effect constructor and value method, registry and colour lookup,
all potion/stew effect families, and a real player add/read/list/remove round trip.
The same one-file plugin overrides the host `OnJoin` lifecycle extension and all 37
`Player.Handler` methods; mutable damage, healing,
immunity, keep-inventory, respawn, skin, knockback, transfer, command arguments, drops, pickup,
experience, page, and reminder values traverse the real private ABI.
Its entity-use handler reads the target's live position and rotation and checks its stable handle,
covering the generated `World.Entity` foundation without exposing the private entity ID.
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
