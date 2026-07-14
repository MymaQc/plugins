using System.Runtime.CompilerServices;
using System.Runtime.InteropServices;
using System.Text;
using Dragonfly.Native;

namespace Dragonfly;

internal static unsafe class PluginBridge
{
    private static Func<Plugin>? Factory;
    private static PluginApi* Descriptor;

    internal static class Host
    {
        internal static HostApi* Api;

        internal static void SendPlayerText(ulong invocation, PlayerId player, uint kind, string message)
        {
            var api = Api;
            if (api is null || api->PlayerText == null) return;
            var bytes = Encoding.UTF8.GetBytes(message);
            fixed (byte* data = bytes)
            {
                _ = api->PlayerText(
                    api->Context,
                    invocation,
                    player,
                    kind,
                    new StringView { Data = data, Length = (ulong)bytes.Length });
            }
        }

        internal static void SetPlayerState(ulong invocation, PlayerId player, uint kind, PlayerStateValue value)
        {
            var api = Api;
            if (api is null || api->PlayerStateSet == null) return;
            _ = api->PlayerStateSet(api->Context, invocation, player, kind, value);
        }

        internal static PlayerStateValue GetPlayerState(ulong invocation, PlayerId player, uint kind)
        {
            var api = Api;
            if (api is null || api->PlayerStateGet == null)
                throw new InvalidOperationException("player is unavailable");
            PlayerStateValue value;
            if (api->PlayerStateGet(api->Context, invocation, player, kind, &value) != Abi.Ok)
                throw new InvalidOperationException("player is no longer available");
            return value;
        }

        internal static double HealPlayer(
            ulong invocation,
            PlayerId player,
            double health,
            World.HealingSource source)
        {
            ArgumentNullException.ThrowIfNull(source);
            var api = Api;
            if (api is null || api->PlayerHeal == null) return 0;

            var (kind, data, name) = source switch
            {
                World.FoodHealingSource food =>
                    (Abi.HealingSourceFood, food.QuickRegeneration() ? (byte)1 : (byte)0, ""),
                World.InstantHealingSource => (Abi.HealingSourceInstant, (byte)0, ""),
                World.RegenerationHealingSource => (Abi.HealingSourceRegeneration, (byte)0, ""),
                _ => (Abi.HealingSourceCustom, (byte)0, source.GetType().FullName ?? source.GetType().Name),
            };
            var nameBytes = Encoding.UTF8.GetBytes(name);
            fixed (byte* nameData = nameBytes)
            {
                var view = new HealingSourceView
                {
                    Name = new StringView { Data = nameData, Length = (ulong)nameBytes.Length },
                    Kind = kind,
                    Data = data,
                };
                PlayerHealResult result;
                return api->PlayerHeal(api->Context, invocation, player, health, &view, &result) == Abi.Ok
                    ? result.Healed
                    : 0;
            }
        }

        internal static World.GameMode PlayerGameMode(ulong invocation, PlayerId player) =>
            World.GameModeFromDescriptor(GetPlayerState(invocation, player, 0).Integer);

        internal readonly record struct EntitySnapshot(
            Vector3 Position,
            Vector3 Velocity,
            Rotation Rotation);

        internal static EntitySnapshot ReadEntityState(ulong invocation, EntityId entity)
        {
            if (!TryReadEntityState(invocation, entity, out var state))
                throw new InvalidOperationException("entity is no longer available");
            return state;
        }

        internal static EntitySnapshot ReadPlayerKinematics(ulong invocation, PlayerId player)
        {
            var api = Api;
            if (api is null || api->PlayerKinematics == null)
                throw new InvalidOperationException("player is unavailable");
            NativePlayerKinematics state;
            if (api->PlayerKinematics(api->Context, invocation, player, &state) != Abi.Ok)
                throw new InvalidOperationException("player is no longer available");
            return new EntitySnapshot(
                new Vector3(state.Position.X, state.Position.Y, state.Position.Z),
                new Vector3(state.Velocity.X, state.Velocity.Y, state.Velocity.Z),
                new Rotation(state.Rotation.Yaw, state.Rotation.Pitch));
        }

        internal static bool TryReadPlayerKinematics(
            ulong invocation,
            PlayerId player,
            out EntitySnapshot snapshot)
        {
            snapshot = default;
            var api = Api;
            if (api is null || api->PlayerKinematics == null) return false;
            NativePlayerKinematics state;
            if (api->PlayerKinematics(api->Context, invocation, player, &state) != Abi.Ok) return false;
            snapshot = new EntitySnapshot(
                new Vector3(state.Position.X, state.Position.Y, state.Position.Z),
                new Vector3(state.Velocity.X, state.Velocity.Y, state.Velocity.Z),
                new Rotation(state.Rotation.Yaw, state.Rotation.Pitch));
            return true;
        }

        internal static bool TryReadEntityState(
            ulong invocation,
            EntityId entity,
            out EntitySnapshot snapshot)
        {
            snapshot = default;
            var api = Api;
            if (api is null || api->EntityState == null) return false;

            const int maxEntityTypeBytes = 256;
            const int maxNameTagBytes = 4096;
            byte* entityType = stackalloc byte[maxEntityTypeBytes];
            byte* nameTag = stackalloc byte[maxNameTagBytes];
            Dragonfly.Native.EntityState state = new()
            {
                EntityType = new StringBuffer { Data = entityType, Capacity = maxEntityTypeBytes },
                NameTag = new StringBuffer { Data = nameTag, Capacity = maxNameTagBytes },
            };
            if (api->EntityState(api->Context, invocation, entity, &state) != Abi.Ok ||
                state.EntityType.Length == 0 || state.EntityType.Length > maxEntityTypeBytes ||
                state.NameTag.Length > maxNameTagBytes)
                return false;
            snapshot = new EntitySnapshot(
                new Vector3(state.Position.X, state.Position.Y, state.Position.Z),
                new Vector3(state.Velocity.X, state.Velocity.Y, state.Velocity.Z),
                new Rotation(state.Rotation.Yaw, state.Rotation.Pitch));
            return true;
        }

        internal static Player? ResolveEntityPlayer(ulong invocation, EntityId entity)
        {
            var api = Api;
            if (api is null || api->EntityPlayer == null) return null;

            const int maxPlayerNameBytes = 256;
            byte* name = stackalloc byte[maxPlayerNameBytes];
            PlayerSnapshotBuffer snapshot = new()
            {
                Name = new StringBuffer { Data = name, Capacity = maxPlayerNameBytes },
            };
            if (api->EntityPlayer(api->Context, invocation, entity, &snapshot) != Abi.Ok ||
                snapshot.Player.Generation == 0 || snapshot.Player.Generation != entity.Generation ||
                snapshot.Name.Data != name || snapshot.Name.Capacity != maxPlayerNameBytes ||
                snapshot.Name.Length == 0 || snapshot.Name.Length > maxPlayerNameBytes ||
                !double.IsFinite(snapshot.Position.X) || !double.IsFinite(snapshot.Position.Y) ||
                !double.IsFinite(snapshot.Position.Z))
                return null;
            for (var index = 0; index < 16; index++)
            {
                if (snapshot.Player.Bytes[index] != entity.Bytes[index]) return null;
            }
            return new Player(
                snapshot.Player,
                Utf8(snapshot.Name),
                TimeSpan.FromMilliseconds(Math.Min(
                    (double)snapshot.LatencyMilliseconds,
                    TimeSpan.MaxValue.TotalMilliseconds)),
                new Vector3(snapshot.Position.X, snapshot.Position.Y, snapshot.Position.Z),
                invocation: invocation);
        }

        internal static void TransformPlayer(
            ulong invocation,
            PlayerId player,
            uint kind,
            Vector3 vector,
            double yaw,
            double pitch)
        {
            var api = Api;
            if (api is null || api->PlayerTransform == null) return;
            _ = api->PlayerTransform(
                api->Context,
                invocation,
                player,
                kind,
                new Vec3 { X = vector.X, Y = vector.Y, Z = vector.Z },
                yaw,
                pitch);
        }

        internal static void CloseEntity(ulong invocation, EntityId entity)
        {
            var api = Api;
            if (api is null || api->EntityDespawn == null) return;
            _ = api->EntityDespawn(api->Context, invocation, entity);
        }

        internal static World? LookupManagedWorld(string name)
        {
            var api = Api;
            if (api is null || api->WorldLookup == null || !TryWorldText(name, 256, out var bytes))
                return null;
            fixed (byte* data = bytes)
            {
                WorldId world;
                var status = api->WorldLookup(
                    api->Context,
                    0,
                    new StringView { Data = data, Length = (ulong)bytes.Length },
                    &world);
                return status == Abi.Ok && world.Value != 0 ? new World(0, world) : null;
            }
        }

        internal static World? OpenManagedWorld(string name, World.ManagedDimension dimension)
        {
            var api = Api;
            if (api is null || api->WorldOpen == null ||
                (uint)dimension > (uint)World.ManagedDimension.End ||
                !TryWorldText(name, 256, out var bytes))
                return null;
            fixed (byte* data = bytes)
            {
                WorldId world;
                var status = api->WorldOpen(
                    api->Context,
                    0,
                    new StringView { Data = data, Length = (ulong)bytes.Length },
                    (uint)dimension,
                    &world);
                return status == Abi.Ok && world.Value != 0 ? new World(0, world) : null;
            }
        }

        internal static World? OpenManagedWorld(string name, World.ManagedOpenSpec spec)
        {
            ArgumentNullException.ThrowIfNull(spec);
            var api = Api;
            if (api is null || api->WorldOpenSpec == null ||
                !TryWorldText(name, 256, out var nameBytes) ||
                !TryWorldText(spec.ProviderPath, 4096, out var providerBytes) ||
                !TryManagedWorldSpec(spec, out var view))
                return null;
            fixed (byte* nameData = nameBytes)
            fixed (byte* providerData = providerBytes)
            {
                view.ProviderPath = new StringView
                {
                    Data = providerData,
                    Length = (ulong)providerBytes.Length,
                };
                WorldId world;
                var status = api->WorldOpenSpec(
                    api->Context,
                    0,
                    new StringView { Data = nameData, Length = (ulong)nameBytes.Length },
                    &view,
                    &world);
                return status == Abi.Ok && world.Value != 0 ? new World(0, world) : null;
            }
        }

        internal static string? ManagedWorldName(ulong invocation, WorldId world)
        {
            var api = Api;
            if (api is null || api->WorldName == null) return null;
            const int capacity = 256;
            byte* data = stackalloc byte[capacity];
            var output = new StringBuffer { Data = data, Capacity = capacity };
            if (api->WorldName(api->Context, invocation, world, &output) != Abi.Ok ||
                output.Length == 0 || output.Length > capacity)
                return null;
            return Encoding.UTF8.GetString(data, checked((int)output.Length));
        }

        internal static Cube.Pos ManagedWorldSpawn(ulong invocation, WorldId world)
        {
            var api = Api;
            if (api is null || api->WorldSpawnGet == null) return default;
            BlockPos position;
            return api->WorldSpawnGet(api->Context, invocation, world, &position) == Abi.Ok
                ? new Cube.Pos(position.X, position.Y, position.Z)
                : default;
        }

        internal static void SetManagedWorldSpawn(ulong invocation, WorldId world, Cube.Pos position)
        {
            var api = Api;
            if (api is null || api->WorldSpawnSet == null) return;
            _ = api->WorldSpawnSet(
                api->Context,
                invocation,
                world,
                new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() });
        }

        internal static void SaveManagedWorld(ulong invocation, WorldId world)
        {
            var api = Api;
            if (api is null || api->WorldSave == null) return;
            _ = api->WorldSave(api->Context, invocation, world);
        }

        internal static void CloseManagedWorld(ulong invocation, WorldId world)
        {
            var api = Api;
            if (api is null || api->WorldUnload == null) return;
            _ = api->WorldUnload(api->Context, invocation, world);
        }

        internal static void ChangePlayerWorld(
            ulong invocation,
            PlayerId player,
            WorldId world,
            Vector3 position)
        {
            var api = Api;
            if (api is null || api->PlayerTransfer == null) return;
            _ = api->PlayerTransfer(
                api->Context,
                invocation,
                player,
                world,
                new Vec3 { X = position.X, Y = position.Y, Z = position.Z });
        }

        private static bool TryManagedWorldSpec(
            World.ManagedOpenSpec spec,
            out WorldOpenSpecV1 view)
        {
            view = default;
            if ((uint)spec.Dimension > (uint)World.ManagedDimension.End ||
                (uint)spec.OpenMode > (uint)World.ManagedOpenMode.CreateNew ||
                (uint)spec.Save > (uint)World.ManagedSavePolicy.Manual ||
                (uint)spec.RandomTicks > (uint)World.ManagedRandomTickPolicy.PerSubchunk ||
                (uint)spec.Time > (uint)World.ManagedTimePolicy.Fixed ||
                (uint)spec.Weather > (uint)World.ManagedWeatherPolicy.Clear ||
                !TryPositiveMilliseconds(spec.ChunkUnloadAfter, out var chunkUnloadMilliseconds))
                return false;

            ulong saveMilliseconds = 0;
            var save = spec.ReadOnly ? World.ManagedSavePolicy.Manual : spec.Save;
            if (save == World.ManagedSavePolicy.Automatic &&
                !TryPositiveMilliseconds(spec.SaveInterval, out saveMilliseconds))
                return false;

            var randomTickRate = spec.RandomTicks == World.ManagedRandomTickPolicy.Disabled
                ? 0
                : spec.RandomTickRate;
            if (spec.RandomTicks == World.ManagedRandomTickPolicy.PerSubchunk &&
                (randomTickRate == 0 || randomTickRate > int.MaxValue))
                return false;

            view = new WorldOpenSpecV1
            {
                StructSize = (uint)sizeof(WorldOpenSpecV1),
                Dimension = (uint)spec.Dimension,
                SaveIntervalMilliseconds = saveMilliseconds,
                ChunkUnloadIntervalMilliseconds = chunkUnloadMilliseconds,
                FixedTime = spec.Time == World.ManagedTimePolicy.Fixed ? spec.FixedTime : 0,
                OpenMode = (uint)spec.OpenMode,
                SavePolicy = (uint)save,
                RandomTickPolicy = (uint)spec.RandomTicks,
                RandomTickRate = randomTickRate,
                TimePolicy = (uint)spec.Time,
                WeatherPolicy = (uint)spec.Weather,
                ChunkUnloadPolicy = 0,
                ReadOnly = spec.ReadOnly ? (byte)1 : (byte)0,
            };
            return true;
        }

        private static bool TryPositiveMilliseconds(TimeSpan duration, out ulong milliseconds)
        {
            milliseconds = 0;
            if (duration.Ticks < TimeSpan.TicksPerMillisecond) return false;
            milliseconds = checked((ulong)(duration.Ticks / TimeSpan.TicksPerMillisecond));
            return true;
        }

        private static bool TryWorldText(string? value, int maxBytes, out byte[] bytes)
        {
            bytes = Array.Empty<byte>();
            if (string.IsNullOrEmpty(value) || value.IndexOf('\0') >= 0) return false;
            bytes = Encoding.UTF8.GetBytes(value);
            return bytes.Length <= maxBytes;
        }

        internal static void AddPlayerEffect(ulong invocation, PlayerId player, Effect.Value effect)
        {
            var type = effect.Type();
            if (type is null || !Effect.TryID(type, out var typeID))
                throw new ArgumentException("effect type is not registered", nameof(effect));
            if (effect.Level() <= 0)
                throw new ArgumentOutOfRangeException(nameof(effect), "effect level must be positive");
            if (effect.Duration() < TimeSpan.Zero)
                throw new ArgumentOutOfRangeException(nameof(effect), "effect duration must not be negative");

            var api = Api;
            if (api is null || api->PlayerEffect == null)
                throw new InvalidOperationException("player is unavailable");
            var view = EncodeEffect(effect, typeID);
            if (api->PlayerEffect(api->Context, invocation, player, 0, view) != Abi.Ok)
                throw new InvalidOperationException("player is no longer available");
        }

        internal static void RemovePlayerEffect(ulong invocation, PlayerId player, Effect.Type type)
        {
            ArgumentNullException.ThrowIfNull(type);
            if (!Effect.TryID(type, out var typeID))
                throw new ArgumentException("effect type is not registered", nameof(type));
            var api = Api;
            if (api is null || api->PlayerEffect == null)
                throw new InvalidOperationException("player is unavailable");
            var view = new EffectView { Type = typeID };
            if (api->PlayerEffect(api->Context, invocation, player, 1, view) != Abi.Ok)
                throw new InvalidOperationException("player is no longer available");
        }

        internal static (Effect.Value Effect, bool Ok) PlayerEffect(ulong invocation, PlayerId player, Effect.Type type)
        {
            ArgumentNullException.ThrowIfNull(type);
            if (!Effect.TryID(type, out var typeID))
                return (default, false);
            foreach (var view in PlayerEffectViews(invocation, player))
            {
                if (view.Type == typeID)
                    return (DecodeEffect(view), true);
            }
            return (default, false);
        }

        internal static IReadOnlyList<Effect.Value> PlayerEffects(ulong invocation, PlayerId player)
        {
            var views = PlayerEffectViews(invocation, player);
            var effects = new Effect.Value[views.Length];
            for (var index = 0; index < effects.Length; index++)
                effects[index] = DecodeEffect(views[index]);
            return effects;
        }

        private static EffectView[] PlayerEffectViews(ulong invocation, PlayerId player)
        {
            var api = Api;
            if (api is null || api->PlayerEffects == null)
                throw new InvalidOperationException("player is unavailable");

            EffectBuffer buffer = default;
            var status = api->PlayerEffects(api->Context, invocation, player, &buffer);
            if (buffer.Length == 0)
            {
                if (status != Abi.Ok)
                    throw new InvalidOperationException("player is no longer available");
                return Array.Empty<EffectView>();
            }
            if (buffer.Length > int.MaxValue)
                throw new InvalidOperationException("invalid effects returned by server");

            var views = new EffectView[checked((int)buffer.Length)];
            fixed (EffectView* data = views)
            {
                buffer = new EffectBuffer { Data = data, Capacity = (ulong)views.Length };
                if (api->PlayerEffects(api->Context, invocation, player, &buffer) != Abi.Ok ||
                    buffer.Length > (ulong)views.Length)
                    throw new InvalidOperationException("player is no longer available");
            }

            if (buffer.Length != (ulong)views.Length)
                Array.Resize(ref views, checked((int)buffer.Length));
            return views;
        }

        private static EffectView EncodeEffect(Effect.Value effect, int typeID)
        {
            return new EffectView
            {
                Type = typeID,
                Level = effect.Level(),
                DurationNanoseconds = checked(effect.Duration().Ticks * 100),
                Potency = effect.Potency,
                Ambient = effect.Ambient() ? (byte)1 : (byte)0,
                ParticlesHidden = effect.ParticlesHidden() ? (byte)1 : (byte)0,
                Infinite = effect.Infinite() ? (byte)1 : (byte)0,
                Tick = effect.Tick(),
            };
        }

        private static Effect.Value DecodeEffect(EffectView view)
        {
            var type = Effect.TypeByID(view.Type);
            if (type is null || view.Level <= 0 || view.Tick < 0 || view.Tick > int.MaxValue ||
                view.DurationNanoseconds % 100 != 0 ||
                view.Ambient > 1 || view.ParticlesHidden > 1 || view.Infinite > 1)
                throw new InvalidOperationException("invalid effect returned by server");
            return new Effect.Value(
                type,
                TimeSpan.FromTicks(view.DurationNanoseconds / 100),
                view.Level,
                view.Potency,
                view.Ambient != 0,
                view.ParticlesHidden != 0,
                view.Infinite != 0,
                (int)view.Tick);
        }

        internal static int InventorySize(ulong invocation, InventoryId inventory)
        {
            var api = Api;
            if (api is null || api->InventorySize == null)
                throw new InvalidOperationException("inventory is unavailable");
            uint size;
            if (api->InventorySize(api->Context, invocation, inventory, &size) != Abi.Ok || size > int.MaxValue)
                throw new InvalidOperationException("inventory is no longer available");
            return (int)size;
        }

        internal static Item.Stack InventoryItem(ulong invocation, InventoryId inventory, int slot)
        {
            var api = Api;
            if (api is null || api->InventoryItemOpen == null || api->ItemStackRead == null || api->ItemStackClose == null)
                throw new InvalidOperationException("inventory is unavailable");
            ulong snapshot;
            ItemStackInfo info;
            if (api->InventoryItemOpen(api->Context, invocation, inventory, checked((uint)slot), &snapshot, &info) != Abi.Ok)
                throw new InvalidOperationException("inventory is no longer available");
            return ReadItemStack(api, invocation, snapshot, info);
        }

        internal static Item.Stack HeldItem(ulong invocation, PlayerId player, uint hand)
        {
            var api = Api;
            if (api is null || api->PlayerHeldItemOpen == null || api->ItemStackRead == null || api->ItemStackClose == null)
                throw new InvalidOperationException("player is unavailable");
            ulong snapshot;
            ItemStackInfo info;
            if (api->PlayerHeldItemOpen(api->Context, invocation, player, hand, &snapshot, &info) != Abi.Ok)
                throw new InvalidOperationException("player is no longer available");
            return ReadItemStack(api, invocation, snapshot, info);
        }

        internal static (Item.Stack MainHand, Item.Stack OffHand) HeldItems(ulong invocation, PlayerId player)
        {
            var api = Api;
            if (api is null || api->PlayerHeldItemsOpen == null || api->ItemStackRead == null || api->ItemStackClose == null)
                throw new InvalidOperationException("player is unavailable");
            ItemStackSnapshot mainHand;
            ItemStackSnapshot offHand;
            if (api->PlayerHeldItemsOpen(api->Context, invocation, player, &mainHand, &offHand) != Abi.Ok)
                throw new InvalidOperationException("player is no longer available");
            var mainOpen = true;
            var offOpen = true;
            try
            {
                mainOpen = false;
                var main = ReadItemStack(api, invocation, mainHand.Snapshot, mainHand.Info);
                offOpen = false;
                var off = ReadItemStack(api, invocation, offHand.Snapshot, offHand.Info);
                return (main, off);
            }
            finally
            {
                if (mainOpen) api->ItemStackClose(api->Context, invocation, mainHand.Snapshot);
                if (offOpen) api->ItemStackClose(api->Context, invocation, offHand.Snapshot);
            }
        }

        internal static Item.Stack EventItem(ItemStackViewV3 item)
        {
            const ulong maxData = 16UL << 20;
            if (item.Identifier.Length > 256 || item.CustomName.Length > 4096 ||
                item.LoreCount > 256 || item.EnchantmentCount > 256 || item.Count > int.MaxValue ||
                item.Nbt.Length > maxData || item.ValuesNbt.Length > maxData ||
                item.Identifier.Length + item.CustomName.Length + item.Nbt.Length + item.ValuesNbt.Length > maxData ||
                item.LoreCount != 0 && item.Lore is null ||
                item.EnchantmentCount != 0 && item.Enchantments is null)
                throw new InvalidOperationException("invalid item stack returned by server");
            if (item.Count == 0) return default;

            var identifier = Copy(item.Identifier);
            var customName = Copy(item.CustomName);
            var itemNbt = Copy(item.Nbt);
            var valuesNbt = Copy(item.ValuesNbt);
            var lore = new string[checked((int)item.LoreCount)];
            ulong total = item.Identifier.Length + item.CustomName.Length + item.Nbt.Length + item.ValuesNbt.Length;
            for (var index = 0; index < lore.Length; index++)
            {
                var line = item.Lore[index];
                if (line.Length > 4096 || total > maxData - line.Length)
                    throw new InvalidOperationException("invalid item stack returned by server");
                total += line.Length;
                lore[index] = Encoding.UTF8.GetString(Copy(line));
            }
            var enchantments = item.EnchantmentCount == 0
                ? Array.Empty<ItemEnchantment>()
                : new ReadOnlySpan<ItemEnchantment>(
                    item.Enchantments,
                    checked((int)item.EnchantmentCount)).ToArray();
            var decoded = ItemCodec.Decode(Encoding.UTF8.GetString(identifier), item.Metadata);
            decoded = ItemNbtCodec.Decode(decoded, itemNbt, out var itemNbtConsumed);
            return new Item.Stack(
                decoded,
                checked((int)item.Count),
                item.Damage,
                item.Unbreakable != 0,
                item.AnvilCost,
                Encoding.UTF8.GetString(customName),
                lore,
                itemNbtConsumed ? null : itemNbt,
                valuesNbt,
                enchantments);
        }

        private static byte[] Copy(StringView value)
        {
            if (value.Length == 0) return Array.Empty<byte>();
            if (value.Data is null || value.Length > int.MaxValue)
                throw new InvalidOperationException("invalid data returned by server");
            return new ReadOnlySpan<byte>(value.Data, checked((int)value.Length)).ToArray();
        }

        internal static void SetInventoryItem(ulong invocation, InventoryId inventory, int slot, Item.Stack item)
        {
            var api = Api;
            if (api is null || api->InventoryItemSet == null)
                throw new InvalidOperationException("inventory is unavailable");
            using var lease = new ItemViewLease(item);
            var view = lease.View;
            if (api->InventoryItemSet(api->Context, invocation, inventory, checked((uint)slot), &view) != Abi.Ok)
                throw new InvalidOperationException("inventory is no longer available");
        }

        internal static int AddInventoryItem(ulong invocation, InventoryId inventory, Item.Stack item)
        {
            var api = Api;
            if (api is null || api->InventoryItemAdd == null)
                throw new InvalidOperationException("inventory is unavailable");
            using var lease = new ItemViewLease(item);
            var view = lease.View;
            uint added;
            if (api->InventoryItemAdd(api->Context, invocation, inventory, &view, &added) != Abi.Ok)
                throw new InvalidOperationException("inventory is no longer available");
            return checked((int)added);
        }

        internal static void SetHeldItems(ulong invocation, PlayerId player, Item.Stack mainHand, Item.Stack offHand)
        {
            var api = Api;
            if (api is null || api->PlayerHeldItemsSet == null)
                throw new InvalidOperationException("player is unavailable");
            using var mainLease = new ItemViewLease(mainHand);
            using var offLease = new ItemViewLease(offHand);
            var main = mainLease.View;
            var off = offLease.View;
            if (api->PlayerHeldItemsSet(api->Context, invocation, player, &main, &off) != Abi.Ok)
                throw new InvalidOperationException("player is no longer available");
        }

        internal static void SetHeldSlot(ulong invocation, PlayerId player, int slot)
        {
            if (slot is < 0 or > 8) throw new ArgumentOutOfRangeException(nameof(slot));
            var api = Api;
            if (api is null || api->PlayerHeldSlotSet == null)
                throw new InvalidOperationException("player is unavailable");
            if (api->PlayerHeldSlotSet(api->Context, invocation, player, (uint)slot) != Abi.Ok)
                throw new InvalidOperationException("player is no longer available");
        }

        private static Item.Stack ReadItemStack(HostApi* api, ulong invocation, ulong snapshot, ItemStackInfo info)
        {
            try
            {
                const ulong maxData = 16UL << 20;
                if (info.IdentifierLength > 256 || info.CustomNameLength > 4096 || info.LoreCount > 256 ||
                    info.EnchantmentCount > 256 || info.IdentifierLength + info.CustomNameLength +
                    info.LoreBytesLength + info.NbtLength + info.ValuesNbtLength > maxData ||
                    info.Count > int.MaxValue)
                    throw new InvalidOperationException("invalid item stack returned by server");

                var identifier = new byte[checked((int)info.IdentifierLength)];
                var customName = new byte[checked((int)info.CustomNameLength)];
                var loreBytes = new byte[checked((int)info.LoreBytesLength)];
                var itemNbt = new byte[checked((int)info.NbtLength)];
                var valuesNbt = new byte[checked((int)info.ValuesNbtLength)];
                var lore = new ByteSpan[checked((int)info.LoreCount)];
                var enchantments = new ItemEnchantment[checked((int)info.EnchantmentCount)];
                fixed (byte* identifierData = identifier)
                fixed (byte* customNameData = customName)
                fixed (byte* loreData = loreBytes)
                fixed (byte* itemNbtData = itemNbt)
                fixed (byte* valuesNbtData = valuesNbt)
                fixed (ByteSpan* loreSpans = lore)
                fixed (ItemEnchantment* enchantmentData = enchantments)
                {
                    var data = new ItemStackData
                    {
                        Identifier = Buffer(identifierData, identifier.Length),
                        CustomName = Buffer(customNameData, customName.Length),
                        LoreBytes = Buffer(loreData, loreBytes.Length),
                        Nbt = Buffer(itemNbtData, itemNbt.Length),
                        ValuesNbt = Buffer(valuesNbtData, valuesNbt.Length),
                        Lore = loreSpans,
                        LoreCapacity = (ulong)lore.Length,
                        Enchantments = enchantmentData,
                        EnchantmentCapacity = (ulong)enchantments.Length,
                    };
                    if (api->ItemStackRead(api->Context, invocation, snapshot, &data) != Abi.Ok)
                        throw new InvalidOperationException("item stack is no longer available");
                }

                if (info.Count == 0) return default;
                var lines = new string[lore.Length];
                for (var index = 0; index < lore.Length; index++)
                {
                    var span = lore[index];
                    if (span.Offset > (ulong)loreBytes.Length || span.Length > (ulong)loreBytes.Length - span.Offset)
                        throw new InvalidOperationException("invalid item lore returned by server");
                    lines[index] = Encoding.UTF8.GetString(loreBytes, checked((int)span.Offset), checked((int)span.Length));
                }
                var item = ItemCodec.Decode(Encoding.UTF8.GetString(identifier), info.Metadata);
                item = ItemNbtCodec.Decode(item, itemNbt, out var itemNbtConsumed);
                return new Item.Stack(
                    item,
                    checked((int)info.Count),
                    info.Damage,
                    info.Unbreakable != 0,
                    info.AnvilCost,
                    Encoding.UTF8.GetString(customName),
                    lines,
                    itemNbtConsumed ? null : itemNbt,
                    valuesNbt,
                    enchantments);
            }
            finally
            {
                api->ItemStackClose(api->Context, invocation, snapshot);
            }
        }

        private static StringBuffer Buffer(byte* data, int length) => new()
        {
            Data = data,
            Capacity = (ulong)length,
        };

        internal sealed class ItemViewLease : IDisposable
        {
            private readonly List<nint> _allocations = [];
            internal ItemStackViewV3 View;

            internal ItemViewLease(Item.Stack stack)
            {
                try
                {
                    var empty = stack.Empty();
                    if (empty)
                    {
                        View = default;
                        return;
                    }
                    var identifier = string.Empty;
                    var metadata = 0;
                    if (!stack.TryEncode(out identifier, out metadata))
                        throw new ArgumentException("item type is not registered", nameof(stack));
                    var lore = stack.Lore();
                    var customName = stack.CustomName();
                    var itemNbt = stack.ItemNbt;
                    var valuesNbt = stack.ValuesNbt;
                    var enchantments = stack.EncodedEnchantments;
                    ValidateItemView(stack, identifier, customName, lore, itemNbt, valuesNbt, enchantments);
                    var loreViews = AllocateViews(lore.Length);
                    for (var index = 0; index < lore.Length; index++) loreViews[index] = AllocateUtf8(lore[index]);
                    var enchantmentData = AllocateArray<ItemEnchantment>(enchantments.Length);
                    if (enchantments.Length != 0) enchantments.CopyTo(new Span<ItemEnchantment>(enchantmentData, enchantments.Length));
                    View = new ItemStackViewV3
                    {
                        Identifier = AllocateUtf8(identifier),
                        Metadata = metadata,
                        Count = checked((uint)stack.Count()),
                        Damage = stack.DamageValue,
                        Unbreakable = stack.IsUnbreakable ? (byte)1 : (byte)0,
                        AnvilCost = stack.AnvilCostValue,
                        CustomName = AllocateUtf8(customName),
                        Lore = loreViews,
                        LoreCount = (ulong)lore.Length,
                        Nbt = Allocate(itemNbt),
                        ValuesNbt = Allocate(valuesNbt),
                        Enchantments = enchantmentData,
                        EnchantmentCount = (ulong)enchantments.Length,
                    };
                }
                catch
                {
                    Dispose();
                    throw;
                }
            }

            private static void ValidateItemView(
                Item.Stack stack,
                string identifier,
                string customName,
                string[] lore,
                byte[] itemNbt,
                byte[] valuesNbt,
                ItemEnchantment[] enchantments)
            {
                const int maxData = 16 << 20;
                if (Encoding.UTF8.GetByteCount(identifier) > 256 ||
                    Encoding.UTF8.GetByteCount(customName) > 4096 ||
                    lore.Length > 256 || enchantments.Length > 256)
                    throw new ArgumentException("item stack data exceeds server limits", nameof(stack));
                long total = Encoding.UTF8.GetByteCount(identifier) + Encoding.UTF8.GetByteCount(customName) +
                    itemNbt.Length + valuesNbt.Length;
                foreach (var line in lore)
                {
                    var length = Encoding.UTF8.GetByteCount(line);
                    if (length > 4096) throw new ArgumentException("item lore exceeds server limits", nameof(stack));
                    total += length;
                }
                if (total > maxData) throw new ArgumentException("item stack data exceeds server limits", nameof(stack));
            }

            private StringView AllocateUtf8(string value) => Allocate(Encoding.UTF8.GetBytes(value));

            private StringView Allocate(byte[] value)
            {
                if (value.Length == 0) return default;
                var data = (byte*)NativeMemory.Alloc((nuint)value.Length);
                _allocations.Add((nint)data);
                value.CopyTo(new Span<byte>(data, value.Length));
                return new StringView { Data = data, Length = (ulong)value.Length };
            }

            private StringView* AllocateViews(int count) => AllocateArray<StringView>(count);

            private T* AllocateArray<T>(int count) where T : unmanaged
            {
                if (count == 0) return null;
                var data = (T*)NativeMemory.Alloc((nuint)count, (nuint)sizeof(T));
                _allocations.Add((nint)data);
                return data;
            }

            public void Dispose()
            {
                foreach (var allocation in _allocations) NativeMemory.Free((void*)allocation);
                _allocations.Clear();
            }
        }

        internal static Skin EventSkin(ulong invocation, ulong snapshot)
        {
            const ulong maxSkinData = 64UL << 20;
            const ulong maxSkinAnimations = 64;
            var api = Api;
            if (api is null || api->SkinSnapshotInfo == null ||
                api->PlayerSkinAnimationInfo == null || api->PlayerSkinRead == null)
                throw new InvalidOperationException("skin is unavailable");

            SkinInfo info;
            if (api->SkinSnapshotInfo(api->Context, invocation, snapshot, &info) != Abi.Ok ||
                info.Persona > 1 || info.AnimationCount > maxSkinAnimations)
                throw new InvalidOperationException("skin is no longer available");

            var lengths = new[]
            {
                info.PlayFabIdLength,
                info.FullIdLength,
                info.PixelsLength,
                info.ModelDefaultLength,
                info.ModelAnimatedFaceLength,
                info.ModelLength,
                info.CapePixelsLength,
            };
            ulong total = 0;
            foreach (var length in lengths)
            {
                if (length > int.MaxValue || length > maxSkinData - total)
                    throw new InvalidOperationException("invalid skin returned by server");
                total += length;
            }
            if (!ValidSkinPixels(info.Width, info.Height, info.PixelsLength, true) ||
                !ValidSkinPixels(info.CapeWidth, info.CapeHeight, info.CapePixelsLength, true))
                throw new InvalidOperationException("invalid skin returned by server");

            var animationInfos = new SkinAnimationInfo[checked((int)info.AnimationCount)];
            for (var index = 0; index < animationInfos.Length; index++)
            {
                fixed (SkinAnimationInfo* animation = &animationInfos[index])
                {
                    if (api->PlayerSkinAnimationInfo(
                            api->Context,
                            invocation,
                            snapshot,
                            (ulong)index,
                            animation) != Abi.Ok)
                        throw new InvalidOperationException("skin is no longer available");
                }
                var value = animationInfos[index];
                if (value.AnimationType > 2 || value.FrameCount <= 0 ||
                    !ValidSkinPixels(value.Width, value.Height, value.PixelsLength, false) ||
                    value.PixelsLength > int.MaxValue || value.PixelsLength > maxSkinData - total)
                    throw new InvalidOperationException("invalid skin returned by server");
                total += value.PixelsLength;
            }

            var allocations = new List<nint>();
            try
            {
                StringBuffer Allocate(ulong length)
                {
                    if (length == 0) return default;
                    var data = (byte*)NativeMemory.Alloc((nuint)length);
                    allocations.Add((nint)data);
                    return new StringBuffer { Data = data, Capacity = length };
                }

                var animationBuffers = animationInfos.Length == 0
                    ? null
                    : (StringBuffer*)NativeMemory.Alloc(
                        (nuint)animationInfos.Length,
                        (nuint)sizeof(StringBuffer));
                if (animationBuffers is not null) allocations.Add((nint)animationBuffers);
                for (var index = 0; index < animationInfos.Length; index++)
                    animationBuffers[index] = Allocate(animationInfos[index].PixelsLength);

                var data = new SkinData
                {
                    PlayFabId = Allocate(info.PlayFabIdLength),
                    FullId = Allocate(info.FullIdLength),
                    Pixels = Allocate(info.PixelsLength),
                    ModelDefault = Allocate(info.ModelDefaultLength),
                    ModelAnimatedFace = Allocate(info.ModelAnimatedFaceLength),
                    Model = Allocate(info.ModelLength),
                    CapePixels = Allocate(info.CapePixelsLength),
                    AnimationPixels = animationBuffers,
                    AnimationCapacity = info.AnimationCount,
                };
                if (api->PlayerSkinRead(api->Context, invocation, snapshot, &data) != Abi.Ok ||
                    data.PlayFabId.Length != info.PlayFabIdLength ||
                    data.FullId.Length != info.FullIdLength ||
                    data.Pixels.Length != info.PixelsLength ||
                    data.ModelDefault.Length != info.ModelDefaultLength ||
                    data.ModelAnimatedFace.Length != info.ModelAnimatedFaceLength ||
                    data.Model.Length != info.ModelLength ||
                    data.CapePixels.Length != info.CapePixelsLength)
                    throw new InvalidOperationException("skin is no longer available");

                var animations = new SkinAnimation[animationInfos.Length];
                for (var index = 0; index < animations.Length; index++)
                {
                    if (animationBuffers[index].Length != animationInfos[index].PixelsLength)
                        throw new InvalidOperationException("skin is no longer available");
                    var animation = animationInfos[index];
                    if (animation.FrameCount > int.MaxValue || animation.Expression is < int.MinValue or > int.MaxValue)
                        throw new InvalidOperationException("invalid skin returned by server");
                    animations[index] = new SkinAnimation(
                        checked((int)animation.Width),
                        checked((int)animation.Height),
                        (SkinAnimationType)animation.AnimationType,
                        CopySkinBuffer(animationBuffers[index]),
                        checked((int)animation.FrameCount),
                        checked((int)animation.Expression));
                }
                return new Skin(
                    checked((int)info.Width),
                    checked((int)info.Height),
                    CopySkinBuffer(data.Pixels))
                {
                    Persona = info.Persona != 0,
                    PlayFabID = SkinText(data.PlayFabId),
                    FullID = SkinText(data.FullId),
                    ModelConfig = new SkinModelConfig
                    {
                        Default = SkinText(data.ModelDefault),
                        AnimatedFace = SkinText(data.ModelAnimatedFace),
                    },
                    Model = CopySkinBuffer(data.Model),
                    Cape = new SkinCape(
                        checked((int)info.CapeWidth),
                        checked((int)info.CapeHeight),
                        CopySkinBuffer(data.CapePixels)),
                    Animations = animations,
                };
            }
            finally
            {
                foreach (var allocation in allocations) NativeMemory.Free((void*)allocation);
            }
        }

        internal static void SetEventSkin(ulong invocation, ulong snapshot, Skin skin)
        {
            ArgumentNullException.ThrowIfNull(skin);
            var api = Api;
            if (api is null || api->SkinSnapshotSet == null)
                throw new InvalidOperationException("skin is unavailable");
            using var lease = new SkinViewLease(skin);
            var view = lease.View;
            if (api->SkinSnapshotSet(api->Context, invocation, snapshot, &view) != Abi.Ok)
                throw new InvalidOperationException("skin is no longer available");
        }

        private static bool ValidSkinPixels(uint width, uint height, ulong length, bool empty)
        {
            if (width > 4096 || height > 4096) return false;
            if (width == 0 || height == 0)
                return empty && width == 0 && height == 0 && length == 0;
            return (ulong)width * height * 4 == length;
        }

        private static byte[] CopySkinBuffer(StringBuffer value) => value.Length == 0
            ? []
            : new ReadOnlySpan<byte>(value.Data, checked((int)value.Length)).ToArray();

        private static string SkinText(StringBuffer value) => value.Length == 0
            ? ""
            : Encoding.UTF8.GetString(value.Data, checked((int)value.Length));

        private sealed class SkinViewLease : IDisposable
        {
            private readonly List<nint> _allocations = [];
            internal SkinView View;

            internal SkinViewLease(Skin skin)
            {
                try
                {
                    ValidateSkin(skin);
                    var animations = AllocateArray<SkinAnimationView>(skin.Animations.Length);
                    for (var index = 0; index < skin.Animations.Length; index++)
                    {
                        var animation = skin.Animations[index];
                        var animationBounds = animation.Bounds();
                        animations[index] = new SkinAnimationView
                        {
                            Width = checked((uint)animationBounds.Width),
                            Height = checked((uint)animationBounds.Height),
                            AnimationType = (uint)animation.Type(),
                            FrameCount = animation.FrameCount,
                            Expression = animation.AnimationExpression,
                            Pixels = Allocate(animation.Pix),
                        };
                    }
                    var bounds = skin.Bounds();
                    var capeBounds = skin.Cape.Bounds();
                    View = new SkinView
                    {
                        Width = checked((uint)bounds.Width),
                        Height = checked((uint)bounds.Height),
                        Persona = skin.Persona ? (byte)1 : (byte)0,
                        PlayFabId = Allocate(Encoding.UTF8.GetBytes(skin.PlayFabID)),
                        FullId = Allocate(Encoding.UTF8.GetBytes(skin.FullID)),
                        Pixels = Allocate(skin.Pix),
                        ModelDefault = Allocate(Encoding.UTF8.GetBytes(skin.ModelConfig.Default)),
                        ModelAnimatedFace = Allocate(Encoding.UTF8.GetBytes(skin.ModelConfig.AnimatedFace)),
                        Model = Allocate(skin.Model),
                        CapeWidth = checked((uint)capeBounds.Width),
                        CapeHeight = checked((uint)capeBounds.Height),
                        CapePixels = Allocate(skin.Cape.Pix),
                        Animations = animations,
                        AnimationCount = (ulong)skin.Animations.Length,
                    };
                }
                catch
                {
                    Dispose();
                    throw;
                }
            }

            private static void ValidateSkin(Skin skin)
            {
                ArgumentNullException.ThrowIfNull(skin.PlayFabID);
                ArgumentNullException.ThrowIfNull(skin.FullID);
                ArgumentNullException.ThrowIfNull(skin.Pix);
                ArgumentNullException.ThrowIfNull(skin.ModelConfig);
                ArgumentNullException.ThrowIfNull(skin.ModelConfig.Default);
                ArgumentNullException.ThrowIfNull(skin.ModelConfig.AnimatedFace);
                ArgumentNullException.ThrowIfNull(skin.Model);
                ArgumentNullException.ThrowIfNull(skin.Cape);
                ArgumentNullException.ThrowIfNull(skin.Cape.Pix);
                ArgumentNullException.ThrowIfNull(skin.Animations);
                var bounds = skin.Bounds();
                var capeBounds = skin.Cape.Bounds();
                if (skin.Animations.Length > 64 ||
                    !ValidSkinPixels(
                        checked((uint)bounds.Width),
                        checked((uint)bounds.Height),
                        (ulong)skin.Pix.Length,
                        true) ||
                    !ValidSkinPixels(
                        checked((uint)capeBounds.Width),
                        checked((uint)capeBounds.Height),
                        (ulong)skin.Cape.Pix.Length,
                        true))
                    throw new ArgumentException("invalid skin", nameof(skin));
                long total = Encoding.UTF8.GetByteCount(skin.PlayFabID) +
                    Encoding.UTF8.GetByteCount(skin.FullID) + skin.Pix.LongLength +
                    Encoding.UTF8.GetByteCount(skin.ModelConfig.Default) +
                    Encoding.UTF8.GetByteCount(skin.ModelConfig.AnimatedFace) + skin.Model.LongLength +
                    skin.Cape.Pix.LongLength;
                foreach (var animation in skin.Animations)
                {
                    ArgumentNullException.ThrowIfNull(animation);
                    ArgumentNullException.ThrowIfNull(animation.Pix);
                    var animationBounds = animation.Bounds();
                    if ((uint)animation.Type() > 2 || animation.FrameCount <= 0 ||
                        !ValidSkinPixels(
                            checked((uint)animationBounds.Width),
                            checked((uint)animationBounds.Height),
                            (ulong)animation.Pix.Length,
                            false))
                        throw new ArgumentException("invalid skin animation", nameof(skin));
                    total += animation.Pix.LongLength;
                }
                if (total > 64L << 20 || Encoding.UTF8.GetByteCount(skin.PlayFabID) > 4096 ||
                    Encoding.UTF8.GetByteCount(skin.FullID) > 4096 ||
                    Encoding.UTF8.GetByteCount(skin.ModelConfig.Default) > 4096 ||
                    Encoding.UTF8.GetByteCount(skin.ModelConfig.AnimatedFace) > 4096)
                    throw new ArgumentException("skin data exceeds server limits", nameof(skin));
            }

            private StringView Allocate(byte[] value)
            {
                if (value.Length == 0) return default;
                var data = (byte*)NativeMemory.Alloc((nuint)value.Length);
                _allocations.Add((nint)data);
                value.CopyTo(new Span<byte>(data, value.Length));
                return new StringView { Data = data, Length = (ulong)value.Length };
            }

            private T* AllocateArray<T>(int count) where T : unmanaged
            {
                if (count == 0) return null;
                var data = (T*)NativeMemory.Alloc((nuint)count, (nuint)sizeof(T));
                _allocations.Add((nint)data);
                return data;
            }

            public void Dispose()
            {
                foreach (var allocation in _allocations) NativeMemory.Free((void*)allocation);
                _allocations.Clear();
            }
        }

        internal static void SendPlayerForm(ulong invocation, PlayerId player, Form.Value form)
        {
            ArgumentNullException.ThrowIfNull(form);
            var request = form.MarshalJSON();
            ArgumentNullException.ThrowIfNull(request);
            var api = Api;
            if (api is null || api->PlayerFormSend == null) return;
            var handle = GCHandle.Alloc(new PendingForm(form));
            fixed (byte* requestData = request)
            {
                var view = new FormView
                {
                    RequestJson = new StringView { Data = requestData, Length = (ulong)request.Length },
                    CallbackContext = (void*)GCHandle.ToIntPtr(handle),
                    Response = &FormResponse,
                    Drop = &FormDrop,
                };
                // A structurally valid view transfers callback-context ownership to the host even on error.
                _ = api->PlayerFormSend(api->Context, invocation, player, &view);
            }
        }

        internal static void ClosePlayerForm(ulong invocation, PlayerId player)
        {
            var api = Api;
            if (api is null || api->PlayerFormClose == null) return;
            _ = api->PlayerFormClose(api->Context, invocation, player);
        }

        [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
        private static int FormResponse(
            void* callbackContext,
            ulong invocation,
            PlayerSnapshot* snapshot,
            uint outcome,
            StringView response)
        {
            PendingForm? pending = null;
            try
            {
                pending = TakePendingForm(callbackContext);
                if (pending is null || snapshot is null || outcome > 1 ||
                    response.Length > 1024 * 1024 || (response.Length != 0 && response.Data is null))
                    return Abi.Error;
                var latency = Math.Min((double)snapshot->LatencyMilliseconds, TimeSpan.MaxValue.TotalMilliseconds);
                var submitter = new Player(
                    snapshot->Player,
                    Utf8(snapshot->Name),
                    TimeSpan.FromMilliseconds(latency),
                    new Vector3(snapshot->Position.X, snapshot->Position.Y, snapshot->Position.Z),
                    invocation: invocation);
                byte[]? responseBytes = outcome == 1
                    ? null
                    : response.Length == 0
                        ? Array.Empty<byte>()
                        : new ReadOnlySpan<byte>(response.Data, checked((int)response.Length)).ToArray();
                pending.Form.SubmitJSON(responseBytes, submitter, new World.Tx(invocation));
                return Abi.Ok;
            }
            catch
            {
                return Abi.Error;
            }
        }

        [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
        private static void FormDrop(void* callbackContext)
        {
            try { _ = TakePendingForm(callbackContext); }
            catch { }
        }

        private static PendingForm? TakePendingForm(void* callbackContext)
        {
            if (callbackContext is null) return null;
            var handle = GCHandle.FromIntPtr((nint)callbackContext);
            var pending = handle.Target as PendingForm;
            handle.Free();
            return pending;
        }

        private sealed class PendingForm(Form.Value form)
        {
            internal Form.Value Form { get; } = form;
        }

        internal static World.Block WorldBlock(ulong invocation, Cube.Pos position)
        {
            var api = Api;
            if (api is null || api->WorldBlockGet == null)
                throw new InvalidOperationException("world transaction is unavailable");

            var nativePosition = new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() };
            var data = new BlockData();
            _ = api->WorldBlockGet(api->Context, invocation, default, nativePosition, &data);
            if (data.Identifier.Length > 256 || data.PropertiesNbt.Length > 64 * 1024)
                throw new InvalidOperationException("invalid block state returned by server");

            var identifierBytes = new byte[checked((int)data.Identifier.Length)];
            var propertyBytes = new byte[checked((int)data.PropertiesNbt.Length)];
            fixed (byte* identifier = identifierBytes)
            fixed (byte* properties = propertyBytes)
            {
                data.Identifier = new StringBuffer
                {
                    Data = identifier,
                    Capacity = (ulong)identifierBytes.Length,
                };
                data.PropertiesNbt = new StringBuffer
                {
                    Data = properties,
                    Capacity = (ulong)propertyBytes.Length,
                };
                if (api->WorldBlockGet(api->Context, invocation, default, nativePosition, &data) != Abi.Ok)
                    throw new InvalidOperationException("world transaction is no longer valid");
            }
            return BlockCodec.Decode(Encoding.UTF8.GetString(identifierBytes), propertyBytes);
        }

        internal static Cube.Range WorldRange(ulong invocation)
        {
            var api = Api;
            if (api is null || api->WorldRange == null)
                throw new InvalidOperationException("world transaction is unavailable");
            BlockRange range;
            if (api->WorldRange(api->Context, invocation, default, &range) != Abi.Ok || range.Min > range.Max)
                throw new InvalidOperationException("world transaction is no longer valid");
            return new Cube.Range(range.Min, range.Max);
        }

        internal static (World.Block? Block, bool Loaded) WorldBlockLoaded(
            ulong invocation,
            Cube.Pos position)
        {
            var api = Api;
            if (api is null || api->WorldBlockLoaded == null)
                throw new InvalidOperationException("world transaction is unavailable");

            var nativePosition = new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() };
            var data = new BlockData();
            byte loaded = 0;
            var status = api->WorldBlockLoaded(
                api->Context,
                invocation,
                default,
                nativePosition,
                &loaded,
                &data);
            if (loaded == 0)
            {
                if (status != Abi.Ok)
                    throw new InvalidOperationException("world transaction is no longer valid");
                return (null, false);
            }
            if (loaded != 1 || data.Identifier.Length > 256 || data.PropertiesNbt.Length > 64 * 1024)
                throw new InvalidOperationException("invalid block state returned by server");

            var identifierBytes = new byte[checked((int)data.Identifier.Length)];
            var propertyBytes = new byte[checked((int)data.PropertiesNbt.Length)];
            fixed (byte* identifier = identifierBytes)
            fixed (byte* properties = propertyBytes)
            {
                data.Identifier = new StringBuffer
                {
                    Data = identifier,
                    Capacity = (ulong)identifierBytes.Length,
                };
                data.PropertiesNbt = new StringBuffer
                {
                    Data = properties,
                    Capacity = (ulong)propertyBytes.Length,
                };
                loaded = 0;
                if (api->WorldBlockLoaded(
                        api->Context,
                        invocation,
                        default,
                        nativePosition,
                        &loaded,
                        &data) != Abi.Ok || loaded != 1)
                    throw new InvalidOperationException("world transaction is no longer valid");
            }
            return (BlockCodec.Decode(Encoding.UTF8.GetString(identifierBytes), propertyBytes), true);
        }

        internal static (World.Liquid? Liquid, bool Ok) WorldLiquid(
            ulong invocation,
            Cube.Pos position)
        {
            var api = Api;
            if (api is null || api->WorldLiquidGet == null)
                throw new InvalidOperationException("world transaction is unavailable");

            var nativePosition = new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() };
            var data = new BlockData();
            byte found = 0;
            var status = api->WorldLiquidGet(
                api->Context,
                invocation,
                default,
                nativePosition,
                &found,
                &data);
            if (found == 0)
            {
                if (status != Abi.Ok)
                    throw new InvalidOperationException("world transaction is no longer valid");
                return (null, false);
            }
            if (found != 1 || data.Identifier.Length > 256 || data.PropertiesNbt.Length > 64 * 1024)
                throw new InvalidOperationException("invalid liquid state returned by server");

            var identifierBytes = new byte[checked((int)data.Identifier.Length)];
            var propertyBytes = new byte[checked((int)data.PropertiesNbt.Length)];
            fixed (byte* identifier = identifierBytes)
            fixed (byte* properties = propertyBytes)
            {
                data.Identifier = new StringBuffer
                {
                    Data = identifier,
                    Capacity = (ulong)identifierBytes.Length,
                };
                data.PropertiesNbt = new StringBuffer
                {
                    Data = properties,
                    Capacity = (ulong)propertyBytes.Length,
                };
                found = 0;
                if (api->WorldLiquidGet(
                        api->Context,
                        invocation,
                        default,
                        nativePosition,
                        &found,
                        &data) != Abi.Ok || found != 1)
                    throw new InvalidOperationException("world transaction is no longer valid");
            }
            return (BlockCodec.DecodeLiquid(Encoding.UTF8.GetString(identifierBytes), propertyBytes), true);
        }

        internal static void SetWorldLiquid(
            ulong invocation,
            Cube.Pos position,
            World.Liquid? liquid)
        {
            var api = Api;
            if (api is null || api->WorldLiquidSet == null) return;
            var nativePosition = new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() };
            if (liquid is null)
            {
                if (api->WorldLiquidSet(api->Context, invocation, default, nativePosition, null) != Abi.Ok)
                    throw new InvalidOperationException("world transaction is no longer valid");
                return;
            }
            if (!BlockCodec.TryEncode(liquid, out var identifier, out var properties))
                throw new ArgumentException("liquid type is not registered", nameof(liquid));

            var identifierBytes = Encoding.UTF8.GetBytes(identifier);
            fixed (byte* identifierData = identifierBytes)
            fixed (byte* propertyData = properties)
            {
                var view = new BlockView
                {
                    Identifier = new StringView { Data = identifierData, Length = (ulong)identifierBytes.Length },
                    PropertiesNbt = new StringView { Data = propertyData, Length = (ulong)properties.Length },
                };
                if (api->WorldLiquidSet(api->Context, invocation, default, nativePosition, &view) != Abi.Ok)
                    throw new InvalidOperationException("world transaction is no longer valid");
            }
        }

        internal static void ScheduleWorldBlockUpdate(
            ulong invocation,
            Cube.Pos position,
            World.Block block,
            TimeSpan delay)
        {
            var api = Api;
            if (api is null || api->WorldBlockUpdateSchedule == null)
                throw new InvalidOperationException("world transaction is unavailable");
            if (!BlockCodec.TryEncode(block, out var identifier, out var properties))
                throw new ArgumentException("block type is not registered", nameof(block));

            long delayNanoseconds;
            try
            {
                delayNanoseconds = checked(delay.Ticks * 100L);
            }
            catch (OverflowException)
            {
                throw new ArgumentOutOfRangeException(nameof(delay), "delay is outside the supported nanosecond range");
            }

            var identifierBytes = Encoding.UTF8.GetBytes(identifier);
            fixed (byte* identifierData = identifierBytes)
            fixed (byte* propertyData = properties)
            {
                var view = new BlockView
                {
                    Identifier = new StringView { Data = identifierData, Length = (ulong)identifierBytes.Length },
                    PropertiesNbt = new StringView { Data = propertyData, Length = (ulong)properties.Length },
                };
                if (api->WorldBlockUpdateSchedule(
                        api->Context,
                        invocation,
                        default,
                        new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() },
                        &view,
                        delayNanoseconds) != Abi.Ok)
                    throw new InvalidOperationException("world transaction is no longer valid");
            }
        }

        internal static void SetWorldBiome(ulong invocation, Cube.Pos position, World.Biome biome)
        {
            var api = Api;
            if (api is null || api->WorldBiomeSet == null)
                throw new InvalidOperationException("world transaction is unavailable");
            if (!BiomeCodec.TryEncode(biome, out var id))
                throw new ArgumentException("biome type is not registered", nameof(biome));
            if (api->WorldBiomeSet(
                    api->Context,
                    invocation,
                    default,
                    new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() },
                    id) != Abi.Ok)
                throw new InvalidOperationException("world transaction is no longer valid");
        }

        internal static World.Biome WorldBiome(ulong invocation, Cube.Pos position)
        {
            var api = Api;
            if (api is null || api->WorldBiomeGet == null)
                throw new InvalidOperationException("world transaction is unavailable");
            int id;
            if (api->WorldBiomeGet(
                    api->Context,
                    invocation,
                    default,
                    new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() },
                    &id) != Abi.Ok)
                throw new InvalidOperationException("world transaction is no longer valid");
            return BiomeCodec.Decode(id);
        }

        internal static double WorldTemperature(ulong invocation, Cube.Pos position)
        {
            var api = Api;
            if (api is null || api->WorldTemperature == null)
                throw new InvalidOperationException("world transaction is unavailable");
            double temperature;
            if (api->WorldTemperature(
                    api->Context,
                    invocation,
                    default,
                    new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() },
                    &temperature) != Abi.Ok)
                throw new InvalidOperationException("world transaction is no longer valid");
            return temperature;
        }

        internal static bool WorldRainingAt(ulong invocation, Cube.Pos position)
        {
            var api = Api;
            if (api is null || api->WorldRainingAt == null)
                throw new InvalidOperationException("world transaction is unavailable");
            return WorldWeatherAt(api, invocation, position, api->WorldRainingAt);
        }

        internal static bool WorldSnowingAt(ulong invocation, Cube.Pos position)
        {
            var api = Api;
            if (api is null || api->WorldSnowingAt == null)
                throw new InvalidOperationException("world transaction is unavailable");
            return WorldWeatherAt(api, invocation, position, api->WorldSnowingAt);
        }

        internal static bool WorldThunderingAt(ulong invocation, Cube.Pos position)
        {
            var api = Api;
            if (api is null || api->WorldThunderingAt == null)
                throw new InvalidOperationException("world transaction is unavailable");
            return WorldWeatherAt(api, invocation, position, api->WorldThunderingAt);
        }

        internal static bool WorldRaining(ulong invocation)
        {
            var api = Api;
            if (api is null || api->WorldRaining == null)
                throw new InvalidOperationException("world transaction is unavailable");
            return WorldWeather(api, invocation, api->WorldRaining);
        }

        internal static bool WorldThundering(ulong invocation)
        {
            var api = Api;
            if (api is null || api->WorldThundering == null)
                throw new InvalidOperationException("world transaction is unavailable");
            return WorldWeather(api, invocation, api->WorldThundering);
        }

        internal static long WorldCurrentTick(ulong invocation)
        {
            var api = Api;
            if (api is null || api->WorldCurrentTick == null)
                throw new InvalidOperationException("world transaction is unavailable");
            long tick;
            if (api->WorldCurrentTick(api->Context, invocation, default, &tick) != Abi.Ok)
                throw new InvalidOperationException("world transaction is no longer valid");
            return tick;
        }

        internal static void AddWorldParticle(ulong invocation, Vector3 position, World.Particle particle)
        {
            var api = Api;
            if (api is null || api->WorldParticleAdd == null)
                throw new InvalidOperationException("world transaction is unavailable");
            if (!ParticleCodec.TryEncode(particle, out var encoded))
                throw new ArgumentException("particle type is not registered", nameof(particle));

            var identifierBytes = Array.Empty<byte>();
            var propertyBytes = Array.Empty<byte>();
            if (encoded.Block is not null)
            {
                if (!BlockCodec.TryEncode(encoded.Block, out var identifier, out propertyBytes))
                    throw new ArgumentException("particle type is not registered", nameof(particle));
                identifierBytes = Encoding.UTF8.GetBytes(identifier);
            }

            fixed (byte* identifierData = identifierBytes)
            fixed (byte* propertyData = propertyBytes)
            {
                var block = new BlockView
                {
                    Identifier = new StringView
                    {
                        Data = identifierData,
                        Length = (ulong)identifierBytes.Length,
                    },
                    PropertiesNbt = new StringView
                    {
                        Data = propertyData,
                        Length = (ulong)propertyBytes.Length,
                    },
                };
                var view = new ParticleView
                {
                    Kind = encoded.Kind,
                    Data = encoded.Data,
                    Pitch = encoded.Pitch,
                    Colour = new Rgba
                    {
                        R = encoded.Colour.R,
                        G = encoded.Colour.G,
                        B = encoded.Colour.B,
                        A = encoded.Colour.A,
                    },
                    Diff = new BlockPos
                    {
                        X = encoded.Diff.X(),
                        Y = encoded.Diff.Y(),
                        Z = encoded.Diff.Z(),
                    },
                    Block = encoded.Block is null ? null : &block,
                };
                if (api->WorldParticleAdd(
                        api->Context,
                        invocation,
                        default,
                        new Vec3 { X = position.X, Y = position.Y, Z = position.Z },
                        &view) != Abi.Ok)
                    throw new InvalidOperationException("world transaction is no longer valid");
            }
        }

        private static bool WorldWeatherAt(
            HostApi* api,
            ulong invocation,
            Cube.Pos position,
            delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, byte*, int> callback)
        {
            byte value;
            if (callback(
                    api->Context,
                    invocation,
                    default,
                    new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() },
                    &value) != Abi.Ok || value > 1)
                throw new InvalidOperationException("world transaction is no longer valid");
            return value == 1;
        }

        private static bool WorldWeather(
            HostApi* api,
            ulong invocation,
            delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, byte*, int> callback)
        {
            byte value;
            if (callback(api->Context, invocation, default, &value) != Abi.Ok || value > 1)
                throw new InvalidOperationException("world transaction is no longer valid");
            return value == 1;
        }

        internal static IEnumerable<Cube.Pos> WorldBlocksWithin(
            ulong invocation,
            Cube.Pos position,
            int radius,
            IReadOnlyList<World.Block> blocks) =>
            new WorldBlocksWithinEnumerable(invocation, position, radius, blocks);

        internal static int WorldHighestLightBlocker(ulong invocation, int x, int z)
        {
            var api = Api;
            if (api is null || api->WorldHighestLightBlocker == null)
                throw new InvalidOperationException("world transaction is unavailable");
            int y;
            if (api->WorldHighestLightBlocker(api->Context, invocation, default, x, z, &y) != Abi.Ok)
                throw new InvalidOperationException("world transaction is no longer valid");
            return y;
        }

        internal static int WorldHighestBlock(ulong invocation, int x, int z)
        {
            var api = Api;
            if (api is null || api->WorldHighestBlock == null)
                throw new InvalidOperationException("world transaction is unavailable");
            int y;
            if (api->WorldHighestBlock(api->Context, invocation, default, x, z, &y) != Abi.Ok)
                throw new InvalidOperationException("world transaction is no longer valid");
            return y;
        }

        internal static byte WorldLight(ulong invocation, Cube.Pos position) =>
            WorldLightLevel(invocation, position, sky: false);

        internal static byte WorldSkyLight(ulong invocation, Cube.Pos position) =>
            WorldLightLevel(invocation, position, sky: true);

        private static byte WorldLightLevel(ulong invocation, Cube.Pos position, bool sky)
        {
            var api = Api;
            if (api is null || sky && api->WorldSkyLight == null || !sky && api->WorldLight == null)
                throw new InvalidOperationException("world transaction is unavailable");
            byte level;
            var nativePosition = new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() };
            var status = sky
                ? api->WorldSkyLight(api->Context, invocation, default, nativePosition, &level)
                : api->WorldLight(api->Context, invocation, default, nativePosition, &level);
            if (status != Abi.Ok)
                throw new InvalidOperationException("world transaction is no longer valid");
            return level;
        }

        private sealed class WorldBlocksWithinEnumerable(
            ulong invocation,
            Cube.Pos position,
            int radius,
            IReadOnlyList<World.Block> blocks) : IEnumerable<Cube.Pos>
        {
            public IEnumerator<Cube.Pos> GetEnumerator() =>
                new WorldBlocksWithinEnumerator(invocation, position, radius, blocks);

            System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator() => GetEnumerator();
        }

        private sealed class WorldBlocksWithinEnumerator : IEnumerator<Cube.Pos>
        {
            private readonly HostApi* _api;
            private readonly ulong _invocation;
            private ulong _iterator;
            private bool _disposed;

            internal WorldBlocksWithinEnumerator(
                ulong invocation,
                Cube.Pos position,
                int radius,
                IReadOnlyList<World.Block> blocks)
            {
                _api = Api;
                if (_api is null || _api->WorldBlocksWithinOpen == null ||
                    _api->WorldBlocksWithinNext == null || _api->WorldBlocksWithinClose == null)
                    throw new InvalidOperationException("world transaction is unavailable");

                _invocation = invocation;
                var encoded = new (byte[] Identifier, byte[] Properties)[blocks.Count];
                var storageLength = 0;
                for (var index = 0; index < blocks.Count; index++)
                {
                    if (!BlockCodec.TryEncode(blocks[index], out var identifier, out var properties))
                        throw new ArgumentException("block type is not registered", nameof(blocks));
                    var identifierBytes = Encoding.UTF8.GetBytes(identifier);
                    encoded[index] = (identifierBytes, properties);
                    storageLength = checked(storageLength + identifierBytes.Length + properties.Length);
                }

                var storage = new byte[storageLength];
                var views = new BlockView[encoded.Length];
                fixed (byte* storageData = storage)
                fixed (BlockView* viewData = views)
                {
                    var offset = 0;
                    for (var index = 0; index < encoded.Length; index++)
                    {
                        var (identifier, properties) = encoded[index];
                        identifier.CopyTo(storage, offset);
                        views[index].Identifier = new StringView
                        {
                            Data = storageData + offset,
                            Length = (ulong)identifier.Length,
                        };
                        offset += identifier.Length;
                        properties.CopyTo(storage, offset);
                        views[index].PropertiesNbt = new StringView
                        {
                            Data = storageData + offset,
                            Length = (ulong)properties.Length,
                        };
                        offset += properties.Length;
                    }

                    ulong iterator = 0;
                    var status = _api->WorldBlocksWithinOpen(
                            _api->Context,
                            invocation,
                            default,
                            new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() },
                            radius,
                            viewData,
                            (ulong)views.Length,
                            &iterator);
                    if (status != Abi.Ok || iterator == 0)
                    {
                        if (iterator != 0)
                            _api->WorldBlocksWithinClose(_api->Context, invocation, iterator);
                        throw new InvalidOperationException("world transaction is no longer valid");
                    }
                    _iterator = iterator;
                }
            }

            public Cube.Pos Current { get; private set; }
            object System.Collections.IEnumerator.Current => Current;

            public bool MoveNext()
            {
                if (_disposed) return false;
                BlockPos position;
                byte ok;
                if (_api->WorldBlocksWithinNext(
                        _api->Context,
                        _invocation,
                        _iterator,
                        &position,
                        &ok) != Abi.Ok || ok > 1)
                {
                    Dispose();
                    throw new InvalidOperationException("world transaction is no longer valid");
                }
                if (ok == 0)
                {
                    Dispose();
                    return false;
                }
                Current = new Cube.Pos(position.X, position.Y, position.Z);
                return true;
            }

            public void Reset() => throw new NotSupportedException();

            public void Dispose()
            {
                if (_disposed) return;
                _disposed = true;
                _api->WorldBlocksWithinClose(_api->Context, _invocation, _iterator);
                _iterator = 0;
            }
        }

        internal static void SetWorldBlock(
            ulong invocation,
            Cube.Pos position,
            World.Block? block,
            World.SetOpts? options)
        {
            var api = Api;
            if (api is null || api->WorldBlockSet == null) return;
            block ??= new Block.Air();
            if (!BlockCodec.TryEncode(block, out var identifier, out var properties))
                throw new ArgumentException("block type is not registered", nameof(block));

            uint flags = 0;
            if (options?.DisableBlockUpdates == true) flags |= Abi.SetBlockDisableBlockUpdates;
            if (options?.DisableLiquidDisplacement == true) flags |= Abi.SetBlockDisableLiquidDisplacement;
            if (options?.DisableRedstoneUpdates == true) flags |= Abi.SetBlockDisableRedstoneUpdates;

            var identifierBytes = Encoding.UTF8.GetBytes(identifier);
            fixed (byte* identifierData = identifierBytes)
            fixed (byte* propertyData = properties)
            {
                var view = new BlockView
                {
                    Identifier = new StringView { Data = identifierData, Length = (ulong)identifierBytes.Length },
                    PropertiesNbt = new StringView { Data = propertyData, Length = (ulong)properties.Length },
                };
                _ = api->WorldBlockSet(
                    api->Context,
                    invocation,
                    default,
                    new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() },
                    &view,
                    flags);
            }
        }
    }

    internal static PluginApi* Initialize(Func<Plugin> factory, string id, ulong subscriptions)
    {
        if (Descriptor is not null) return Descriptor;
        Factory = factory;
        var bytes = Encoding.UTF8.GetBytes(id);
        var idPointer = (byte*)NativeMemory.Alloc((nuint)bytes.Length);
        bytes.CopyTo(new Span<byte>(idPointer, bytes.Length));
        Descriptor = (PluginApi*)NativeMemory.AllocZeroed((nuint)sizeof(PluginApi));
        *Descriptor = new PluginApi
        {
            Header = new AbiHeader
            {
                Version = Abi.PluginVersion,
                Size = (uint)sizeof(PluginApi),
                Subscriptions = subscriptions,
            },
            Id = new StringView { Data = idPointer, Length = (ulong)bytes.Length },
            Create = &Create,
            Enable = &Enable,
            Disable = &Disable,
            Commands = &Commands,
            HandleCommand = &HandleCommand,
            CommandEnumOptions = &CommandEnumOptions,
            SetHost = &SetHost,
            Destroy = &Destroy,
            HandleEvent = &HandleEvent,
        };
        return Descriptor;
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static void* Create()
    {
        try
        {
            CommandRegistry.Clear();
            return (void*)GCHandle.ToIntPtr(GCHandle.Alloc(Factory!()));
        }
        catch { return null; }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int SetHost(void* instance, void* host)
    {
        if (host is null) return Abi.Error;
        var header = (HostHeader*)host;
        if (header->Version != Abi.HostVersion || header->Size < 672) return Abi.Error;
        Host.Api = (HostApi*)host;
        return Abi.Ok;
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int Enable(void* instance, StringBuffer* error)
    {
        try
        {
            Get(instance).OnEnable();
            if (error is not null) error->Length = 0;
            return Abi.Ok;
        }
        catch (Exception exception)
        {
            Write(error, exception.Message);
            return Abi.Error;
        }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int Disable(void* instance)
    {
        try { Get(instance).OnDisable(); return Abi.Ok; }
        catch { return Abi.Error; }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static void Destroy(void* instance)
    {
        if (instance is not null) GCHandle.FromIntPtr((nint)instance).Free();
        Host.Api = null;
        CommandRegistry.Clear();
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static CommandDescriptor* Commands(void* instance, ulong* count)
    {
        try
        {
            if (instance is null) return null;
            return CommandRegistry.Native(count);
        }
        catch
        {
            if (count is not null) *count = 0;
            return null;
        }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int HandleCommand(void* instance, ulong index, CommandInput* input, CommandState* state)
    {
        try { return instance is null ? Abi.Error : CommandRegistry.Execute(index, input, state); }
        catch { return Abi.Error; }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int CommandEnumOptions(
        void* instance,
        ulong index,
        ulong overload,
        ulong parameter,
        CommandEnumContext* context,
        StringBuffer* output)
    {
        try
        {
            return instance is null
                ? Abi.Error
                : CommandRegistry.EnumOptions(index, overload, parameter, context, output);
        }
        catch { return Abi.Error; }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int HandleEvent(void* instance, uint eventId, void* input, void* state)
    {
        try
        {
            if (instance is null || input is null || state is null) return Abi.Error;
            var plugin = Get(instance);
            switch (eventId)
            {
                case Abi.PlayerChatEvent:
                {
                    var value = (PlayerChatInput*)input;
                    var result = (PlayerChatState*)state;
                    var original = result->HasReplacement != 0
                        ? Utf8(result->Replacement)
                        : Utf8(value->Message);
                    var message = original;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    plugin.HandleChat(context, ref message);
                    ApplyCancellation(context, &result->Cancelled);
                    if (message != original)
                    {
                        if (!WriteExact(&result->Replacement, message)) return Abi.Error;
                        result->HasReplacement = 1;
                    }
                    return Abi.Ok;
                }
                case Abi.PlayerJoinEvent:
                {
                    var value = (PlayerJoinInput*)input;
                    var result = (PlayerJoinState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    plugin.OnJoin(context);
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerHurtEvent:
                {
                    var value = (PlayerHurtInput*)input;
                    var result = (PlayerHurtState*)state;
                    if (value->Immune > 1) return Abi.Error;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    var damage = result->Damage;
                    var originalAttackImmunityNanoseconds = result->AttackImmunityNanoseconds;
                    var originalAttackImmunity = TimeSpan.FromTicks(
                        originalAttackImmunityNanoseconds / 100);
                    var attackImmunity = originalAttackImmunity;
                    plugin.HandleHurt(
                        context,
                        ref damage,
                        value->Immune != 0,
                        ref attackImmunity,
                        EventDamageSource(value->Source, value->Invocation));
                    result->Damage = damage;
                    result->AttackImmunityNanoseconds = attackImmunity == originalAttackImmunity
                        ? originalAttackImmunityNanoseconds
                        : checked(attackImmunity.Ticks * 100);
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerHealEvent:
                {
                    var value = (PlayerHealInput*)input;
                    var result = (PlayerHealState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    var health = result->Health;
                    plugin.HandleHeal(
                        context,
                        ref health,
                        EventHealingSource(value->Source));
                    result->Health = health;
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerBlockBreakEvent:
                {
                    var value = (PlayerBlockBreakInput*)input;
                    var result = (PlayerBlockBreakState*)state;
                    var previousContext = result->ReplacementContext;
                    var previousDrop = result->ReplacementDrop;
                    var current = result->ReplacementDrop == null
                        ? value->Drops
                        : result->ReplacementDrops;
                    var count = result->ReplacementDrop == null
                        ? value->DropCount
                        : result->ReplacementDropCount;
                    if (count > int.MaxValue || count != 0 && current is null) return Abi.Error;
                    var drops = new Item.Stack[checked((int)count)];
                    for (var index = 0; index < drops.Length; index++)
                        drops[index] = Host.EventItem(current[index]);
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    var experience = result->Experience;
                    plugin.HandleBlockBreak(
                        context,
                        EventPos(value->Position),
                        ref drops,
                        ref experience);
                    ArgumentNullException.ThrowIfNull(drops);
                    var replacement = TransferEventItems(drops);
                    result->Experience = experience;
                    result->ReplacementDrops = replacement.Views;
                    result->ReplacementDropCount = replacement.Count;
                    result->ReplacementContext = replacement.Context;
                    result->ReplacementDrop = replacement.Drop;
                    ApplyCancellation(context, &result->Cancelled);
                    if (previousDrop != null) previousDrop(previousContext);
                    return Abi.Ok;
                }
                case Abi.PlayerBlockPlaceEvent:
                case Abi.PlayerBlockPickEvent:
                {
                    var value = (PlayerBlockInput*)input;
                    var result = (CancellableState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    var position = EventPos(value->Position);
                    var block = EventBlock(value->Block);
                    if (eventId == Abi.PlayerBlockPlaceEvent)
                        plugin.HandleBlockPlace(context, position, block);
                    else
                        plugin.HandleBlockPick(context, position, block);
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerStartBreakEvent:
                case Abi.PlayerFireExtinguishEvent:
                {
                    var value = (PlayerBlockPositionInput*)input;
                    var result = (CancellableState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    var position = EventPos(value->Position);
                    if (eventId == Abi.PlayerStartBreakEvent)
                        plugin.HandleStartBreak(context, position);
                    else
                        plugin.HandleFireExtinguish(context, position);
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerFoodLossEvent:
                {
                    var value = (PlayerFoodLossInput*)input;
                    var result = (PlayerFoodLossState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    var to = result->To;
                    plugin.HandleFoodLoss(context, value->From, ref to);
                    result->To = to;
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerDeathEvent:
                {
                    var value = (PlayerDeathInput*)input;
                    var result = (PlayerDeathState*)state;
                    var keepInventory = result->KeepInventory != 0;
                    plugin.HandleDeath(
                        SnapshotPlayer(value->Player, value->Invocation),
                        EventDamageSource(value->Source, value->Invocation),
                        ref keepInventory);
                    result->KeepInventory = keepInventory ? (byte)1 : (byte)0;
                    return Abi.Ok;
                }
                case Abi.PlayerExperienceGainEvent:
                {
                    var value = (PlayerEventInput*)input;
                    var result = (PlayerIntegerState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    var amount = result->Value;
                    plugin.HandleExperienceGain(context, ref amount);
                    result->Value = amount;
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerHeldSlotChangeEvent:
                {
                    var value = (PlayerHeldSlotChangeInput*)input;
                    var result = (CancellableState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    plugin.HandleHeldSlotChange(context, value->From, value->To);
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerSleepEvent:
                {
                    var value = (PlayerEventInput*)input;
                    var result = (PlayerSleepState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    var sendReminder = result->SendReminder != 0;
                    plugin.HandleSleep(context, ref sendReminder);
                    result->SendReminder = sendReminder ? (byte)1 : (byte)0;
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerLecternPageTurnEvent:
                {
                    var value = (PlayerLecternPageTurnInput*)input;
                    var result = (PlayerLecternPageTurnState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    var newPage = result->NewPage;
                    plugin.HandleLecternPageTurn(
                        context,
                        EventPos(value->Position),
                        value->OldPage,
                        ref newPage);
                    result->NewPage = newPage;
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerSignEditEvent:
                {
                    var value = (PlayerSignEditInput*)input;
                    var result = (CancellableState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    plugin.HandleSignEdit(
                        context,
                        EventPos(value->Position),
                        value->FrontSide != 0,
                        Utf8(value->OldText),
                        Utf8(value->NewText));
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerItemUseEvent:
                {
                    var value = (PlayerEventInput*)input;
                    var result = (CancellableState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    plugin.HandleItemUse(context);
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerItemUseOnBlockEvent:
                {
                    var value = (PlayerItemUseOnBlockInput*)input;
                    var result = (CancellableState*)state;
                    if (value->Face is < 0 or > 5) return Abi.Error;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    plugin.HandleItemUseOnBlock(
                        context,
                        EventPos(value->Position),
                        (Cube.Face)value->Face,
                        new Vector3(
                            value->ClickPosition.X,
                            value->ClickPosition.Y,
                            value->ClickPosition.Z));
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerItemConsumeEvent:
                case Abi.PlayerItemDropEvent:
                {
                    var value = (PlayerItemInput*)input;
                    var result = (CancellableState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    var item = Host.EventItem(value->Item);
                    if (eventId == Abi.PlayerItemConsumeEvent)
                        plugin.HandleItemConsume(context, item);
                    else
                        plugin.HandleItemDrop(context, item);
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerItemReleaseEvent:
                {
                    var value = (PlayerItemReleaseInput*)input;
                    var result = (CancellableState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    plugin.HandleItemRelease(
                        context,
                        Host.EventItem(value->Item),
                        TimeSpan.FromTicks(value->DurationNanoseconds / 100));
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerItemDamageEvent:
                {
                    var value = (PlayerItemInput*)input;
                    var result = (PlayerIntegerState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    var damage = result->Value;
                    plugin.HandleItemDamage(context, Host.EventItem(value->Item), ref damage);
                    result->Value = damage;
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerAttackEntityEvent:
                {
                    var value = (PlayerAttackEntityInput*)input;
                    var result = (PlayerAttackEntityState*)state;
                    if (result->Critical > 1) return Abi.Error;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    var target = EventEntity(value->Target, value->Invocation);
                    if (target is null) return Abi.Error;
                    var force = result->KnockbackForce;
                    var height = result->KnockbackHeight;
                    var critical = result->Critical != 0;
                    plugin.HandleAttackEntity(
                        context,
                        target,
                        ref force,
                        ref height,
                        ref critical);
                    result->KnockbackForce = force;
                    result->KnockbackHeight = height;
                    result->Critical = critical ? (byte)1 : (byte)0;
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerItemUseOnEntityEvent:
                {
                    var value = (PlayerItemUseOnEntityInput*)input;
                    var result = (CancellableState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    var target = EventEntity(value->Target, value->Invocation);
                    if (target is null) return Abi.Error;
                    plugin.HandleItemUseOnEntity(context, target);
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerChangeWorldEvent:
                {
                    var value = (PlayerChangeWorldInput*)input;
                    if (value->After.Value == 0) return Abi.Error;
                    var before = value->Before.Value == 0
                        ? null
                        : new World(value->Invocation, value->Before);
                    plugin.HandleChangeWorld(
                        SnapshotPlayer(value->Player, value->Invocation),
                        before,
                        new World(value->Invocation, value->After));
                    return Abi.Ok;
                }
                case Abi.PlayerRespawnEvent:
                {
                    var value = (PlayerEventInput*)input;
                    var result = (PlayerRespawnState*)state;
                    if (result->World.Value == 0) return Abi.Error;
                    var position = new Vector3(
                        result->Position.X,
                        result->Position.Y,
                        result->Position.Z);
                    var world = new World(value->Invocation, result->World);
                    plugin.HandleRespawn(
                        SnapshotPlayer(value->Player, value->Invocation),
                        ref position,
                        ref world);
                    ArgumentNullException.ThrowIfNull(world);
                    result->Position = new Vec3 { X = position.X, Y = position.Y, Z = position.Z };
                    result->World = world.Id;
                    return Abi.Ok;
                }
                case Abi.PlayerSkinChangeEvent:
                {
                    var value = (PlayerSkinChangeInput*)input;
                    var result = (CancellableState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    var skin = Host.EventSkin(value->Invocation, value->Snapshot);
                    plugin.HandleSkinChange(context, ref skin);
                    ArgumentNullException.ThrowIfNull(skin);
                    Host.SetEventSkin(value->Invocation, value->Snapshot, skin);
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerTransferEvent:
                {
                    var value = (PlayerTransferInput*)input;
                    var result = (PlayerTransferState*)state;
                    var previousContext = result->ReplacementContext;
                    var previousDrop = result->ReplacementDrop;
                    var address = EventUDPAddr(result->Address);
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    plugin.HandleTransfer(context, ref address);
                    ArgumentNullException.ThrowIfNull(address);
                    var replacement = new PendingTransferAddress(address);
                    result->Address = replacement.View;
                    result->ReplacementContext = replacement.Context;
                    result->ReplacementDrop = replacement.Drop;
                    ApplyCancellation(context, &result->Cancelled);
                    if (previousDrop != null) previousDrop(previousContext);
                    return Abi.Ok;
                }
                case Abi.PlayerCommandExecutionEvent:
                {
                    var value = (PlayerCommandExecutionInput*)input;
                    var result = (PlayerCommandExecutionState*)state;
                    var previousContext = result->ReplacementContext;
                    var previousDrop = result->ReplacementDrop;
                    var aliases = EventStrings(value->CommandAliases, value->CommandAliasCount);
                    var arguments = result->ReplacementDrop == null
                        ? EventStrings(value->Arguments, value->ArgumentCount)
                        : EventStrings(result->ReplacementArguments, result->ReplacementArgumentCount);
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    plugin.HandleCommandExecution(
                        context,
                        new Cmd.Command(
                            Utf8(value->CommandName),
                            Utf8(value->CommandDescription),
                            Utf8(value->CommandUsage),
                            aliases),
                        arguments);
                    var replacement = new PendingEventStrings(arguments);
                    result->ReplacementArguments = replacement.Views;
                    result->ReplacementArgumentCount = replacement.Count;
                    result->ReplacementContext = replacement.Context;
                    result->ReplacementDrop = replacement.Drop;
                    ApplyCancellation(context, &result->Cancelled);
                    if (previousDrop != null) previousDrop(previousContext);
                    return Abi.Ok;
                }
                case Abi.PlayerDiagnosticsEvent:
                {
                    var value = (PlayerDiagnosticsInput*)input;
                    plugin.HandleDiagnostics(
                        SnapshotPlayer(value->Player, value->Invocation),
                        new Session.Diagnostics(
                            value->AverageFramesPerSecond,
                            value->AverageServerSimTickTime,
                            value->AverageClientSimTickTime,
                            value->AverageBeginFrameTime,
                            value->AverageInputTime,
                            value->AverageRenderTime,
                            value->AverageEndFrameTime,
                            value->AverageRemainderTimePercent,
                            value->AverageUnaccountedTimePercent));
                    return Abi.Ok;
                }
                case Abi.PlayerItemPickupEvent:
                {
                    var value = (PlayerItemPickupInput*)input;
                    var result = (PlayerItemPickupState*)state;
                    var previousContext = result->ReplacementContext;
                    var previousDrop = result->ReplacementDrop;
                    if (previousDrop != null && result->Replacement is null) return Abi.Error;
                    var item = result->ReplacementDrop == null
                        ? Host.EventItem(value->Item)
                        : Host.EventItem(*result->Replacement);
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    plugin.HandleItemPickup(context, ref item);
                    var replacement = TransferEventItems([item]);
                    result->Replacement = replacement.Views;
                    result->ReplacementContext = replacement.Context;
                    result->ReplacementDrop = replacement.Drop;
                    ApplyCancellation(context, &result->Cancelled);
                    if (previousDrop != null) previousDrop(previousContext);
                    return Abi.Ok;
                }
                case Abi.PlayerJumpEvent:
                {
                    var value = (PlayerEventInput*)input;
                    plugin.HandleJump(SnapshotPlayer(value->Player, value->Invocation));
                    return Abi.Ok;
                }
                case Abi.PlayerMoveEvent:
                {
                    var value = (PlayerMoveInput*)input;
                    var result = (PlayerMoveState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    plugin.HandleMove(
                        context,
                        new Vector3(value->NewPosition.X, value->NewPosition.Y, value->NewPosition.Z),
                        new Rotation(value->Rotation.Yaw, value->Rotation.Pitch));
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerPunchAirEvent:
                {
                    var value = (PlayerEventInput*)input;
                    var result = (CancellableState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    plugin.HandlePunchAir(context);
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerQuitEvent:
                {
                    var value = (PlayerQuitInput*)input;
                    plugin.HandleQuit(SnapshotPlayer(value->Player, value->Invocation));
                    return Abi.Ok;
                }
                case Abi.PlayerTeleportEvent:
                {
                    var value = (PlayerTeleportInput*)input;
                    var result = (CancellableState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    plugin.HandleTeleport(context, new Vector3(value->Position.X, value->Position.Y, value->Position.Z));
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.PlayerToggleSneakEvent:
                case Abi.PlayerToggleSprintEvent:
                {
                    var value = (PlayerToggleInput*)input;
                    var result = (CancellableState*)state;
                    var context = Event(value->Player, result->Cancelled, value->Invocation);
                    if (eventId == Abi.PlayerToggleSneakEvent)
                        plugin.HandleToggleSneak(context, value->After != 0);
                    else
                        plugin.HandleToggleSprint(context, value->After != 0);
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                default:
                    return Abi.Ok;
            }
        }
        catch { return Abi.Error; }
    }

    private static Plugin Get(void* instance) => (Plugin)GCHandle.FromIntPtr((nint)instance).Target!;

    private static Player SnapshotPlayer(PlayerSnapshot snapshot, ulong invocation) => new(
        snapshot.Player,
        Utf8(snapshot.Name),
        TimeSpan.FromMilliseconds(Math.Min(
            (double)snapshot.LatencyMilliseconds,
            TimeSpan.MaxValue.TotalMilliseconds)),
        new Vector3(snapshot.Position.X, snapshot.Position.Y, snapshot.Position.Z),
        invocation: invocation);

    private static Player.Context Event(PlayerSnapshot player, byte cancelled, ulong invocation = 0) =>
        new(SnapshotPlayer(player, invocation), cancelled != 0);

    private static Cube.Pos EventPos(BlockPos position) => new(position.X, position.Y, position.Z);

    private static World.Block EventBlock(BlockView block)
    {
        if (block.Identifier.Length > 256 || block.PropertiesNbt.Length > 64 * 1024 ||
            block.Identifier.Length != 0 && block.Identifier.Data is null ||
            block.PropertiesNbt.Length != 0 && block.PropertiesNbt.Data is null)
            throw new InvalidOperationException("invalid block returned by server");
        var properties = block.PropertiesNbt.Length == 0
            ? Array.Empty<byte>()
            : new ReadOnlySpan<byte>(
                block.PropertiesNbt.Data,
                checked((int)block.PropertiesNbt.Length)).ToArray();
        return BlockCodec.Decode(Utf8(block.Identifier), properties);
    }

    private static World.Entity? EventEntity(EntityId entity, ulong invocation) =>
        entity.Generation == 0
            ? null
            : Host.ResolveEntityPlayer(invocation, entity) ?? World.HostEntityFrom(invocation, entity);

    private static World.DamageSource EventDamageSource(DamageSourceView source, ulong invocation)
    {
        const uint knownFlags = Abi.DamageSourceReducedByArmour |
            Abi.DamageSourceReducedByResistance | Abi.DamageSourceFire |
            Abi.DamageSourceIgnoresTotem | Abi.DamageSourceFireProtection |
            Abi.DamageSourceFeatherFalling | Abi.DamageSourceBlastProtection |
            Abi.DamageSourceProjectileProtection;
        if (source.Kind > Abi.DamageSourceWither || (source.Flags & ~knownFlags) != 0 ||
            source.Data > 1 || source.Data != 0 && source.Kind != 12 ||
            (source.Block is not null) != (source.Kind == 2))
            throw new InvalidOperationException("invalid damage source returned by server");
        _ = Utf8(source.Name);
        var properties = new World.DamageProperties(
            (source.Flags & Abi.DamageSourceReducedByArmour) != 0,
            (source.Flags & Abi.DamageSourceReducedByResistance) != 0,
            (source.Flags & Abi.DamageSourceFire) != 0,
            (source.Flags & Abi.DamageSourceIgnoresTotem) != 0);
        return source.Kind switch
        {
            0 => new World.CustomDamageSource(properties),
            1 => new World.AttackDamageSource(properties, EventEntity(source.Entity, invocation)),
            2 => new World.BlockDamageSource(properties, EventBlock(*source.Block)),
            3 => new World.DrowningDamageSource(properties),
            4 => new World.ExplosionDamageSource(properties),
            5 => new World.FallDamageSource(properties),
            6 => new World.FireDamageSource(properties),
            7 => new World.GlideDamageSource(properties),
            8 => new World.InstantDamageSource(properties),
            9 => new World.LavaDamageSource(properties),
            10 => new World.LightningDamageSource(properties),
            11 => new World.MagmaDamageSource(properties),
            12 => new World.PoisonDamageSource(properties, source.Data != 0),
            13 => new World.ProjectileDamageSource(
                properties,
                EventEntity(source.Entity, invocation),
                EventEntity(source.SecondaryEntity, invocation)),
            14 => new World.StarvationDamageSource(properties),
            15 => new World.SuffocationDamageSource(properties),
            16 => new World.ThornsDamageSource(properties, EventEntity(source.Entity, invocation)),
            17 => new World.VoidDamageSource(properties),
            18 => new World.WitherDamageSource(properties),
            _ => throw new InvalidOperationException("invalid damage source returned by server"),
        };
    }

    private static World.HealingSource EventHealingSource(HealingSourceView source)
    {
        if (source.Kind > Abi.HealingSourceRegeneration || source.Data > 1 ||
            source.Data != 0 && source.Kind != 1)
            throw new InvalidOperationException("invalid healing source returned by server");
        _ = Utf8(source.Name);
        return source.Kind switch
        {
            0 => new World.CustomHealingSource(),
            1 => new World.FoodHealingSource(source.Data != 0),
            2 => new World.InstantHealingSource(),
            3 => new World.RegenerationHealingSource(),
            _ => throw new InvalidOperationException("invalid healing source returned by server"),
        };
    }

    private static Net.UDPAddr EventUDPAddr(UDPAddrView address)
    {
        if (address.IP.Length > 16 || address.IP.Length != 0 && address.IP.Data is null)
            throw new InvalidOperationException("invalid transfer IP returned by server");
        var ip = address.IP.Length == 0
            ? Array.Empty<byte>()
            : new ReadOnlySpan<byte>(address.IP.Data, checked((int)address.IP.Length)).ToArray();
        return new Net.UDPAddr(ip, address.Port, Utf8(address.Zone));
    }

    private static string[] EventStrings(StringView* values, ulong count)
    {
        const ulong maxStrings = 1024;
        if (count > maxStrings || count != 0 && values is null)
            throw new InvalidOperationException("invalid string array returned by server");
        var result = new string[checked((int)count)];
        for (var index = 0; index < result.Length; index++) result[index] = Utf8(values[index]);
        return result;
    }

    private static PendingEventItems TransferEventItems(IReadOnlyList<Item.Stack> items) => new(items);

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static void DropTransferAddress(void* context)
    {
        if (context is null) return;
        var handle = GCHandle.FromIntPtr((nint)context);
        try { (handle.Target as PendingTransferAddress)?.Dispose(); }
        finally { handle.Free(); }
    }

    private sealed class PendingTransferAddress : IDisposable
    {
        private GCHandle _handle;

        internal PendingTransferAddress(Net.UDPAddr address)
        {
            ArgumentNullException.ThrowIfNull(address.IP);
            ArgumentNullException.ThrowIfNull(address.Zone);
            if (address.IP.Length is not (0 or 4 or 16))
                throw new ArgumentException("a UDP address IP must be empty, IPv4, or IPv6", nameof(address));
            var zone = Encoding.UTF8.GetBytes(address.Zone);
            if (zone.Length > 4096)
                throw new ArgumentException("a UDP address zone must not exceed 4096 UTF-8 bytes", nameof(address));
            try
            {
                View = new UDPAddrView { IP = Allocate(address.IP), Port = address.Port };
                var view = View;
                view.Zone = Allocate(zone);
                View = view;
                _handle = GCHandle.Alloc(this);
            }
            catch
            {
                Dispose();
                throw;
            }
        }

        internal UDPAddrView View { get; private set; }
        internal void* Context => (void*)GCHandle.ToIntPtr(_handle);
        internal delegate* unmanaged[Cdecl]<void*, void> Drop => &DropTransferAddress;

        public void Dispose()
        {
            if (View.IP.Data is not null) NativeMemory.Free(View.IP.Data);
            if (View.Zone.Data is not null) NativeMemory.Free(View.Zone.Data);
            View = default;
        }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static void DropEventStrings(void* context)
    {
        if (context is null) return;
        var handle = GCHandle.FromIntPtr((nint)context);
        try { (handle.Target as PendingEventStrings)?.Dispose(); }
        finally { handle.Free(); }
    }

    private sealed class PendingEventStrings : IDisposable
    {
        private GCHandle _handle;

        internal PendingEventStrings(IReadOnlyList<string> values)
        {
            try
            {
                Views = values.Count == 0
                    ? null
                    : (StringView*)NativeMemory.AllocZeroed(
                        (nuint)values.Count,
                        (nuint)sizeof(StringView));
                Count = (ulong)values.Count;
                for (var index = 0; index < values.Count; index++)
                {
                    ArgumentNullException.ThrowIfNull(values[index]);
                    var bytes = Encoding.UTF8.GetBytes(values[index]);
                    if (bytes.Length > 64 * 1024)
                        throw new ArgumentException(
                            "an event string must not exceed 64 KiB of UTF-8 data",
                            nameof(values));
                    Views[index] = Allocate(bytes);
                }
                _handle = GCHandle.Alloc(this);
            }
            catch
            {
                Dispose();
                throw;
            }
        }

        internal StringView* Views { get; private set; }
        internal ulong Count { get; }
        internal void* Context => (void*)GCHandle.ToIntPtr(_handle);
        internal delegate* unmanaged[Cdecl]<void*, void> Drop => &DropEventStrings;

        public void Dispose()
        {
            if (Views is null) return;
            for (var index = 0; index < checked((int)Count); index++)
                if (Views[index].Data is not null) NativeMemory.Free(Views[index].Data);
            NativeMemory.Free(Views);
            Views = null;
        }
    }

    private static StringView Allocate(ReadOnlySpan<byte> bytes)
    {
        if (bytes.Length == 0) return default;
        var data = (byte*)NativeMemory.Alloc((nuint)bytes.Length);
        bytes.CopyTo(new Span<byte>(data, bytes.Length));
        return new StringView { Data = data, Length = (ulong)bytes.Length };
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static void DropEventItems(void* context)
    {
        if (context is null) return;
        var handle = GCHandle.FromIntPtr((nint)context);
        try { (handle.Target as PendingEventItems)?.Dispose(); }
        finally { handle.Free(); }
    }

    private sealed class PendingEventItems : IDisposable
    {
        private readonly Host.ItemViewLease[] _leases;
        private GCHandle _handle;

        internal PendingEventItems(IReadOnlyList<Item.Stack> items)
        {
            _leases = new Host.ItemViewLease[items.Count];
            try
            {
                Views = items.Count == 0
                    ? null
                    : (ItemStackViewV3*)NativeMemory.Alloc(
                        (nuint)items.Count,
                        (nuint)sizeof(ItemStackViewV3));
                for (var index = 0; index < items.Count; index++)
                {
                    var lease = new Host.ItemViewLease(items[index]);
                    _leases[index] = lease;
                    Views[index] = lease.View;
                }
                Count = (ulong)items.Count;
                _handle = GCHandle.Alloc(this);
            }
            catch
            {
                Dispose();
                throw;
            }
        }

        internal ItemStackViewV3* Views { get; private set; }
        internal ulong Count { get; }
        internal void* Context => (void*)GCHandle.ToIntPtr(_handle);
        internal delegate* unmanaged[Cdecl]<void*, void> Drop => &DropEventItems;

        public void Dispose()
        {
            foreach (var lease in _leases) lease?.Dispose();
            if (Views is not null) NativeMemory.Free(Views);
            Views = null;
        }
    }

    private static void ApplyCancellation(Player.Context context, byte* cancelled)
    {
        if (context.Cancelled()) *cancelled = 1;
    }

    private static string Utf8(StringView value)
    {
        const ulong maxStringBytes = 16UL << 20;
        if (value.Length == 0) return string.Empty;
        if (value.Data is null || value.Length > maxStringBytes)
            throw new InvalidOperationException("invalid string returned by server");
        return Encoding.UTF8.GetString(new ReadOnlySpan<byte>(value.Data, checked((int)value.Length)));
    }

    private static string Utf8(StringBuffer value)
    {
        const ulong maxStringBytes = 16UL << 20;
        if (value.Length == 0) return string.Empty;
        if (value.Data is null || value.Length > value.Capacity || value.Length > maxStringBytes)
            throw new InvalidOperationException("invalid string returned by server");
        return Encoding.UTF8.GetString(new ReadOnlySpan<byte>(value.Data, checked((int)value.Length)));
    }

    private static void Write(StringBuffer* output, string message)
    {
        if (output is null || output->Data is null || output->Capacity == 0) return;
        var bytes = Encoding.UTF8.GetBytes(message);
        var length = Math.Min(bytes.Length, checked((int)output->Capacity));
        bytes.AsSpan(0, length).CopyTo(new Span<byte>(output->Data, length));
        output->Length = (ulong)length;
    }

    private static bool WriteExact(StringBuffer* output, string message)
    {
        if (output is null) return false;
        var bytes = Encoding.UTF8.GetBytes(message);
        if ((ulong)bytes.Length > output->Capacity || bytes.Length != 0 && output->Data is null) return false;
        bytes.CopyTo(new Span<byte>(output->Data, bytes.Length));
        output->Length = (ulong)bytes.Length;
        return true;
    }
}
