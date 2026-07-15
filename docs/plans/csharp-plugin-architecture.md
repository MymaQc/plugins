# C# plugin architecture

## Direction

- C# NativeAOT is the only plugin language.
- The public namespace mirrors Dragonfly's packages and exported types as closely as C# permits.
- Plugins subclass `Plugin`; generated build plumbing supplies the native entry point and project-name identity.
- The Go host owns Dragonfly and exposes a private flat C ABI. Plugins never use ABI types.
- Code generation reads the pinned Dragonfly Go source with `go/ast` and emits C#; there is no second public API schema.
- Same generator owns matching private Go player transport constants and dispatch; deleted schema-era `abi-gen` output cannot drift outside `make check-generated`.
- Packet generation reads the pinned gophertunnel packet structs and the intercept handler contract with
  `go/ast`; intercept hands the host packets that gophertunnel has already decoded.

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
   Connection admission separately mirrors Dragonfly's `server.Allower` exactly through
   `Plugin.Allow(Net.Addr, Login.IdentityData, Login.ClientData)`. It is not an event and receives
   no invented context or player: Dragonfly calls it before player construction. The Allower,
   identity/client snapshot classes, and `Protocol.DeviceOS` values are generated from the pinned
   Dragonfly and gophertunnel Go AST. Plugins allow by default; enabled plugins are evaluated in
   deterministic library order and the first denial wins. Any preconfigured Go Allower runs first.
   Callbacks may be concurrent. Identity data retains all six upstream fields, including the two
   PlayFab fields excluded from gophertunnel's normal JSON. Client data remains explicitly
   untrusted. Internal callback failures fail closed with a fixed public rejection while only a
   private runtime status reaches the server log. Plugin ABI 11 adds this dedicated callback.
2. Generated handler events. All 37 methods in the pinned Dragonfly `player.Handler` interface are generated and
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
   All 13 methods in `world.Handler` are also AST-generated and transported for every managed
   world: liquid flow/decay/hardening, sound, fire spread, block burn, crop trample, leaves decay,
   entity spawn/despawn, explosion, redstone updates, and close. `HandleExplosion` preserves its
   arbitrary replacement semantics for entity and block slices plus mutable item-drop chance and
   fire spawning. Entity values materialise as concrete `Player` objects where applicable.
   `RedstoneUpdate` includes all 12 upstream fields, nullable `After`, and all three causes.
   Context callbacks use sticky cancellation; notification callbacks receive their borrowed
   transaction without inventing cancellation state.
   The packet slice uses the exact `intercept.Handler` names, `HandleClientPacket` and
   `HandleServerPacket`. Both are allowed by default and cancellable. The host borrows each decoded
   `packet.Packet` for one synchronous callback and exposes all 233 registered concrete packet
   types. The generator currently maps 681 primitive, string, byte, UUID, and float-vector fields
   to mutable typed properties. The remaining 205 top-level fields are callback-scoped lazy
   `Packet.Value` objects with JSON inspection while recursive protocol structs, optionals, unions,
   maps, and slices are added to the AST generator. Incoming packet mutation changes the exact
   decoded object consumed by Dragonfly. Outgoing inspection and cancellation are safe; outgoing
   mutation is rejected because intercept v0.3 passes potentially shared packet pointers and has
   no per-connection clone/replacement contract. Packet callbacks intentionally expose the
   connection XUID rather than a fake `Player.Context`: login packets can precede player
   registration and the intercept callback does not own a Dragonfly world transaction.
   `Player.WritePacket(Packet.Packet)` is AST-validated against
   `bedrock-gophers/unsafe.WritePacket` and forwards the borrowed object without exposing its
   native handle. Stale handles fail silently like other void player methods. Outgoing writes
   re-enter outgoing interception exactly as raw Go does. Owned packet construction remains future
   work.
3. Player methods and commands. Command interfaces and the implemented `Player` method surface are generated from Dragonfly's Go AST. C# uses `Cmd.New`/`Cmd.Register`, one `Cmd.Runnable` per overload, and reflected public fields as Dragonfly uses reflected Go struct fields. Supported command fields include subcommands, native enums, dynamic `Cmd.Enum` values, players, vectors, optional values, and `Cmd.Varargs`. The generator roots runnable fields and field types for NativeAOT; runnable types use `internal` visibility and require no linker annotations. Bedrock-facing subcommands and enum/player suggestions are always lowercase. The generated game-mode slice includes the exact `World.GameMode` interface, four registered values, `GameModeByID`, `GameModeID`, `Player.SetGameMode`, and `Player.GameMode`. Custom C# game modes cross the private ABI as their eight Dragonfly capabilities and remain unregistered, matching raw Dragonfly behavior. The text slice includes `Message`, `SendPopup`, `SendTip`, `SendJukeboxPopup`, `SetNameTag`, and `Disconnect`. State methods include food, health, experience, scale, visibility, mobility, physical/activity state, fire state, and air supply. Every zero-argument, void Dragonfly player method is exposed, including breaking, input-lock, debug-shape, inventory, item-use, movement-action, and wake controls. The exact `Hud.Element` and `Input.Lock` opaque value sets and their six player methods are generated from Dragonfly and gophertunnel AST; plugins use constructors such as `Hud.Crosshair()` and `Input.Jump()` without raw IDs. Presentation methods include instant-respawn and coordinate toggles, sleeping indicators, `CloseDialogue`, `RemoveBossBar`, `RemoveScoreboard`, live `NameTag`/`ScoreTag` reads, and `SendToast`. Go `time.Duration` maps to C# `TimeSpan`; host calls remain exact Dragonfly methods. State, action, and string transport IDs are generated for Go and C# from one AST-validated spec. The kinematics slice adds AST-generated `Teleport`, `Move`, `Displace`, `Position`, `Velocity`, `SetVelocity`, and `Rotation`; the host invokes those exact Dragonfly methods and does not reject non-finite values that raw Dragonfly accepts. Setters return `void`; private host status never enters the public API. `SetMaxHealth` preserves Dragonfly's clamp-to-one behavior for non-positive values. `Messagef` and `SetScoreTag` remain absent until Go formatting semantics can be preserved honestly.
4. World and block parity. The first landed slice generates `Cube.Pos`, `Cube.Range`,
   `Cube.Face`, `World.Block`, `World.SetOpts`, `World.Tx.Range`, `World.Tx.Block`,
   `World.Tx.BlockLoaded`, `World.Tx.BlocksWithin`, `World.Tx.SetBlock`,
   `World.Tx.ScheduleBlockUpdate`,
   `World.Tx.HighestLightBlocker`, `World.Tx.HighestBlock`, `World.Tx.Light`,
   `World.Tx.SkyLight`, `World.Liquid`, `World.Tx.Liquid`, `World.Tx.SetLiquid`, 118 non-liquid
   block types covering all 314 registered states whose varying registry-state fields are primitive,
   plus `Block.Water` and `Block.Lava`. Promoted fields are expanded through Dragonfly's AST, so
   all eight growth stages of `BeetrootSeeds`, `Carrot`, `Potato`, and `WheatSeeds` remain typed.
   World values also expose AST-generated `Range`, `HighestLightBlocker`, `Time`, and `SetTime`
   through the same typed world transport used by transactions. The biome slice adds
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
   ABI 39 transports that distinction, signed `time.Duration` nanoseconds, private biome IDs,
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
   `Item.Stack` is generated from Dragonfly's `item.Stack` AST and exposes the complete public
   method set: `NewStack`, count/growth/max-count, durability/damage/unbreakable, attack damage,
   custom names, lore, values, enchantments, anvil cost, `WithItem`, `String`, comparison/equality,
   and stack merging.
   Generated player methods expose `Inventory`, `EnderChestInventory`, `Armour`, `HeldItems`,
   `SetHeldItems`, and `SetHeldSlot`; `Inventory.Value` exposes `Size`, `Item`, `SetItem`, and
   `AddItem`. C# setters return `void` as the chosen language adaptation, and invalid slots
   throw `ArgumentOutOfRangeException`; host statuses never enter the public API. The existing
   ABI 39 includes one atomic held-item pair snapshot, so `HeldItems` observes the same player state
   with one host read. Main and ender-chest inventory sizes are read from the live Dragonfly
   inventory, preserving custom `player.Config` sizes. Bounded open/read/close item snapshots preserve damage, unbreakable state, anvil cost, custom
   names, lore, item NBT, plugin values, and enchantments internally. `Stack.WithValue`, `Value`,
   and `Values` expose the cross-language NBT-compatible value set. The generator discovers all 27
   registered enchantments from Dragonfly's AST and live registry; normal and forced addition,
   removal, lookup, deterministic listing, item compatibility, and enchantment compatibility mirror
   `item.Stack`. Unknown registered stateful
   NBT-backed items decode to a private opaque item and round-trip losslessly. `WithItem` rebuilds
   the stack in Dragonfly's exact order, retaining only damage, enchantments, and anvil cost valid
   for the replacement typed item. Private opaque items retain their NBT through `WithItem`, and
   `Comparable` compares decoded NBT values rather than encoded key order. Custom items remain
   next; no public identifier fallback is added.
7. Effects. The generator reads `server/entity/effect` registrations, interfaces, constructors,
   value methods, and player method signatures from Dragonfly's Go AST, then validates all 28
   built-ins against the live registry. C# exposes `Effect.Value`, registered `Type`/`LastingType`
   values, all five constructors, `ResultingColour`, `ByID`/`ID`, and the four player effect methods.
   ABI 39 transports signed nanosecond duration, level, potency, ambient/particle/infinite flags, and
   tick. C# `TimeSpan` has 100 ns precision and rejects snapshots outside that precision instead of
   truncating them. Re-adding a snapshot is bounded to one million elapsed ticks because Dragonfly
   exposes `Tick` but no constructor or setter for it. Pending initial instant-effect potency is
   normalised because Dragonfly exposes no potency getter. Custom `Type.Apply`, `LastingType.Start`/`End`,
   `Register`, and concrete multiplier methods wait for entity and damage-source callbacks instead
   of receiving a second abstraction.
8. Entities, sounds, and remaining world/block methods. The entity foundation now
   generates Dragonfly's exact `World.Entity` interface (`Close`, `H`, `Position`, `Rotation`)
   from `server/world/entity.go`. Handler entities are live host-backed values, not public ID
   tokens; `EntityHandle` retains stable identity without conflating handle closure with entity
   despawning. Player-backed entity handles resolve to the concrete `Player` type, preserving
   `entity is Player` checks and `Player.Name()` in attack and entity-use handlers. `Player`
   implements `World.Entity`, as it does in Dragonfly. Entity IDs, generations, state buffers, and
   host status codes remain private. AST-generated `Tx.World`, `Tx.Entities`, and `Tx.Players`
   preserve Dragonfly's exact signatures; entity/player iteration is lazy, reads the live world,
   and is closed on exhaustion, early disposal, or invocation end. The replaced eager private
   entity/player snapshots have been removed. ABI 39 adds stable private handle identities and the
   AST-generated `EntityHandle.Entity`, `UUID`, `Closed`, and `Close` methods plus exact public
   `Tx.AddEntity`, `AddEntityAt`, and `RemoveEntity` signatures. `Cube.BBox` and
   `Tx.EntitiesWithin` are AST-generated too; the latter lazily selects entities whose positions
   are strictly inside the box, matching Dragonfly rather than intersecting entity hitboxes.
   Removing an entity expires its
   world-bound identity; adding the same handle creates a fresh identity while preserving handle
   equality. Abandoned detached custom entities are closed before plugin runtime shutdown.
   Generic player removal is intentionally rejected for now because Dragonfly's connected session
   must complete its coordinated world transfer; `Player.ChangeWorld` is the safe current path.
   The AST-generated public `Server` surface adds direct `Plugin.Server()`, lazy
   `Server.World()`, `Server.Nether()`, `Server.End()`, `Server.MaxPlayerCount()`,
   `Server.PlayerCount()`, `Server.Players(World.Tx?)`, and stable-handle
   `Server.Player(Guid)`/`PlayerByName(string)`/`PlayerByXUID(string)` lookups. AST-generated
   `Player.Name()`, `Player.UUID()`, and `Player.XUID()` preserve Dragonfly's identity surface while
   UUIDs, XUID buffers, and lookup handles remain private transport details. A non-null current
   transaction must be passed when iteration begins in a callback or
   command; `null` remains valid outside a transaction and is never replaced with an inferred one.
   Every `foreach` body runs synchronously on the yielded player's world owner. Advancing or
   disposing the enumerator expires the prior borrowed `Player`, so players must not be collected
   or retained; only `Player.H()` is stable beyond that body. Re-entering the same world owner from
   the body can deadlock, and mirrored server scans in handlers on different worlds can deadlock
   each other, exactly as in Dragonfly. The private iterator closes on exhaustion, early disposal,
   callback completion, or runtime shutdown.
   `World.Do`, `World.DoAfter`, and `World.Task` are generated from Dragonfly's exact AST shapes.
   `Task.Done`, `Err`, `Wait`, `OnDone`, and `Cancel` preserve completion and cancellation semantics;
   C# maps Dragonfly task errors to domain exceptions without exposing ABI status. The private
   callback trampoline keeps delegates plugin-owned and borrows a fresh transaction only during
   execution. Framework teardown rejects new tasks, cancels pending delays, and drains running
   callbacks before `OnDisable`. Host ABI 47 and plugin ABI 11 carry execution, terminal outcome,
   signed delay, and cancellation. `World.New()`
   constructs Dragonfly's in-memory NOP-provider world. `World.Config.New()` transports the
   selected upstream config fields atomically; `MCDB.Config.Open(path)` selects a writable,
   persistent provider rooted below the configured worlds directory. Created worlds are owned and
   closed by the framework, but internal registry keys and provider handles never enter plugin
   code. `World.Name()` remains Dragonfly's display name. ABI 43 adds the AST-pinned package-level
   `World.BlockByName(string, Dictionary<string, object?>?)` surface. Its private property codec
   preserves Dragonfly's exact `bool`, `uint8`, `int32`, and `string` state types, and the host
   resolves them through `world.BlockByName`; failed names or state hashes return `(null, false)`.
   This is the dynamic block path needed by VAP-style arena data, while generated concrete block
   types remain the normal compile-time path.
   Host ABI 47 adds AST-generated `World.Dimension`, `StopTime`, `StartTime`, `TimeCycle`,
   `SetRequiredSleepDuration`, `DefaultGameMode`, `SetDefaultGameMode`, `SetTickRange`,
   `Difficulty`, and `SetDifficulty`. Registered game modes/difficulties preserve identity; custom
   structural implementations preserve their exact interface capabilities. C# emits `Dimension()`
   and `Difficulty()` as extension methods solely to avoid the language collision with the nested
   `World.Dimension` and `World.Difficulty` types while retaining identical call syntax.
   Host ABI 49 adds exact AST-generated `World.PlayerSpawn(Guid)` and
   `World.SetPlayerSpawn(Guid, Cube.Pos)` calls backed directly by Dragonfly's provider state.
   Host ABI 50 activates one generated player-action transport in the old reserved slot. Exact
   hunger and experience methods call Dragonfly directly; return-producing `AddExperience` and
   `CollectExperience` return their real results instead of being faked as state setters.
   Host ABI 51 appends typed player-string reads and two-string toast sends. Name and score tags
   are read live with a bounded retrying buffer; toast title and message remain distinct strings.
   ABI 44 and plugin ABI 9 add AST-generated `EntitySpawnOpts`, `EntityType`, `EntityConfig`,
   `EntityData`, and `TickerEntity`. `EntitySpawnOpts.New` creates a worldless Dragonfly handle;
   `Tx.AddEntity`, `RemoveEntity`, and `AddEntityAt` preserve that handle across a fresh world-bound
   identity. Each Dragonfly `EntityType.Open` call creates the plugin's actual C# entity object for
   that transaction. `BBox`, `Close`, `H`, `Position`, `Rotation`, and `TickerEntity.Tick` dispatch
   to it directly, while `DecodeNBT` and `EncodeNBT` persist the plugin-owned `EntityData.Data`.
   The transaction defers release of each open C# view until Dragonfly finishes the transaction.
   There is one exact proxy path; the abandoned family/callback/physics entity DSL is removed.
   The sound slice AST-generates all 87 concrete `server/world/sound` structs as `Sound.*` values
   implementing `World.Sound`. Exact AST-generated `World.Tx.PlaySound` and `Player.PlaySound` use the
   same generated reverse codec and existing private sound descriptor as `HandleSound`; world playback
   broadcasts at a position while player playback targets only that player. `HandleSound` materialises
   their exported bool, scalar, block, item,
   liquid, instrument, disc, horn, pitch, and stage fields. Bucket sounds preserve the exact typed
   liquid block state rather than reducing it to a Minecraft identifier or liquid-kind enum.
   Playback currently accepts the 87 generated concrete sounds. Exact custom `World.Sound.Play`
   callback dispatch remains part of the next entity/world callback slice; the public interface must
   gain that AST-generated method before custom sounds can claim raw-Dragonfly parity.
   Later entity slices add the remaining concrete `entity.Ent` capabilities and transaction methods.
9. Convert practice-core and expand parity tests against Dragonfly.

## FFA parity

The combat callback foundation is present: hurt, attack, death, respawn, typed damage sources,
mutable damage/knockback/hit delay, live entity transforms, player kinematics, and direct player
teleportation and healing. `OnJoin` supplies the missing lobby-entry lifecycle without claiming to
be an upstream handler method. Dragonfly-shaped `World.New()`, `World.Config.New()`, and
`MCDB.Config.Open()` cover in-memory and writable persistent world construction; world spawn,
save/close, and safe `Player.ChangeWorld` use exact AST-generated `World` methods. Stack values and
typed enchantments cover selector metadata and Protection kits.
Direct server-wide lazy player iteration and UUID/name lookup now make global broadcasts and
online-player resolution possible without a public manager abstraction. Player-backed attack and
entity-use targets are concrete `Player` values, so killer inventory inspection and refill are
available too. Functional Nodebuff and Sumo FFA are therefore implementable with the current API.
Damage and healing sources are generated directly from the Dragonfly Go AST under their real
package owners: `Entity`, `Block`, `Effect`, `Player`, and `Enchantment`, with only the two source
interfaces under `World`. Exported Go fields remain typed record properties, protection matching
is preserved through `Enchantment.AffectedDamageSource`, and `Player.Hurt` and `Player.Heal` both
accept the same source interfaces as Dragonfly. Unknown custom implementations cross the private
ABI without adding invented public wrapper types.

Remaining raw-Dragonfly parity work includes:

- remaining `Player`, `World`, and `World.Tx` methods, including transaction defer/event,
  world configuration, presentation, combat, and player-control surfaces;
- exact command behavior and skin types;
- player-capable raw handle transfer and remaining concrete entity capabilities;
- custom items, blocks, dimensions, providers, and generators beyond current generated registries and
  NOP/MCDB `World.Config` surface.

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
`/kitchen entities` exercises `Entities`, `Players`, and a strict-position `EntitiesWithin` query
using an AST-generated `Cube.BBox` around the command source.
`/kitchen handle` resolves a live non-player handle, checks its UUID and transaction lookup,
removes it, proves the worldless lookup fails, then re-adds it at the command source while preserving
handle identity.
`/kitchen server` iterates all online players through the direct Dragonfly `Server` surface,
messages them inside their borrowed loop bodies, retains only a stable handle, and verifies UUID
and exact-name lookup resolve to the same handle.
`/kitchen particle` emits all 20 particle types and exercises every one of Dragonfly's 16 note
instruments through the transaction-owned `AddParticle` call.
`/kitchen sound` exercises both exact playback methods and every sound payload family: flags,
scalars, blocks, instruments, discs, items, liquids, stages, and horns.
`/kitchen game-mode` exercises registered lookup, player reads, and a custom capability-backed
game mode.
`/kitchen world` exercises world dimension, time-cycle stop/start, sleep duration, default game
mode, tick range, registered/custom difficulty, and restores observable state after transport.
`/kitchen state` exercises the generated food, health, experience-level/progress, scale,
visibility, mobility, walking-speed, flight-speed, and vertical-flight-speed methods without
changing the restorable player state. It also reads absorption, fall distance, death, ground
contact, body heights, and breathing, then exercises Dragonfly's fall-distance reset.
It also exercises exact sprinting, sneaking, swimming, crawling, gliding, and flying state methods.
It exercises exact hunger mutations, total experience, enchantment seed, experience mutations,
and experience collection too. Enchantment-seed reset and the 100 ms collection delay are
write-only upstream effects, so this diagnostic command intentionally changes those two values.
`/kitchen item` exercises stack count, durability, unbreakable, attack damage, anvil cost,
values, `WithItem`, `String`, semantic NBT comparison, and merging, then round-trips all eleven
finite stateful item families plus both typed book and firework item families through player
inventory before restoring all changed player state. Its firework coverage also exercises typed
explosion shapes, colours, fades, twinkle,
trail, off-hand support, and randomised duration. Its armour coverage checks tier, defence,
durability, repair, smelting, trim-material, and private dyed/trim NBT behavior, then round-trips
all 28 tier-and-piece combinations. It also constructs a charged typed crossbow, checks its pure
capabilities, and round-trips the full nested firework stack. Empty, water, lava, and milk buckets
exercise typed content queries, consumption flags, duration, max counts, and lava fuel residue.
`/kitchen effect` exercises every effect constructor and value method, registry and colour lookup,
all potion/stew effect families, and a real player add/read/list/remove round trip.
The same one-file plugin overrides the host `OnJoin` lifecycle extension, all 37
`Player.Handler` methods, and all 13 `World.Handler` methods; mutable damage, healing,
immunity, keep-inventory, respawn, skin, knockback, transfer, command arguments, drops, pickup,
experience, page, reminder, explosion entity/block lists, item-drop chance, and fire values traverse
the real private ABI.
Its attack and entity-use handlers preserve concrete player targets, read `Player.Name()`, inspect
live position/rotation, and check the stable handle without exposing private entity or player IDs.
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
