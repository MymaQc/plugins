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

        internal static World.GameMode PlayerGameMode(ulong invocation, PlayerId player) =>
            World.GameModeFromDescriptor(GetPlayerState(invocation, player, 0).Integer);

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
                return new Item.Stack(
                    item,
                    checked((int)info.Count),
                    info.Damage,
                    info.Unbreakable != 0,
                    info.AnvilCost,
                    Encoding.UTF8.GetString(customName),
                    lines,
                    itemNbt,
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

        private sealed class ItemViewLease : IDisposable
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
                    ValidateItemView(stack, identifier, lore);
                    var loreViews = AllocateViews(lore.Length);
                    for (var index = 0; index < lore.Length; index++) loreViews[index] = AllocateUtf8(lore[index]);
                    var enchantments = stack.Enchantments;
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
                        CustomName = AllocateUtf8(stack.CustomName()),
                        Lore = loreViews,
                        LoreCount = (ulong)lore.Length,
                        Nbt = Allocate(stack.ItemNbt),
                        ValuesNbt = Allocate(stack.ValuesNbt),
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

            private static void ValidateItemView(Item.Stack stack, string identifier, string[] lore)
            {
                const int maxData = 16 << 20;
                if (Encoding.UTF8.GetByteCount(identifier) > 256 ||
                    Encoding.UTF8.GetByteCount(stack.CustomName()) > 4096 ||
                    lore.Length > 256 || stack.Enchantments.Length > 256)
                    throw new ArgumentException("item stack data exceeds server limits", nameof(stack));
                long total = Encoding.UTF8.GetByteCount(identifier) + Encoding.UTF8.GetByteCount(stack.CustomName()) +
                    stack.ItemNbt.Length + stack.ValuesNbt.Length;
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
        if (header->Version != Abi.HostVersion || header->Size < 664) return Abi.Error;
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
            var plugin = Get(instance);
            switch (eventId)
            {
                case Abi.PlayerChatEvent:
                {
                    var value = (PlayerChatInput*)input;
                    var result = (PlayerChatState*)state;
                    var original = Utf8(value->Message);
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
                case Abi.PlayerJumpEvent:
                {
                    var value = (PlayerEventInput*)input;
                    plugin.HandleJump(new Player(value->Player, invocation: value->Invocation));
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
                    plugin.HandleQuit(new Player(value->Player, Utf8(value->Name), invocation: value->Invocation));
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

    private static Player.Context Event(PlayerId player, byte cancelled, ulong invocation = 0) =>
        new(new Player(player, invocation: invocation), cancelled != 0);

    private static void ApplyCancellation(Player.Context context, byte* cancelled)
    {
        if (context.Cancelled()) *cancelled = 1;
    }

    private static string Utf8(StringView value) => value.Length == 0
        ? string.Empty
        : Encoding.UTF8.GetString(new ReadOnlySpan<byte>(value.Data, checked((int)value.Length)));

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
