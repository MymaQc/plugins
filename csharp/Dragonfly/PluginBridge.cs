using System.Runtime.CompilerServices;
using System.Runtime.InteropServices;
using System.Text;
using System.Text.Json;
using System.Collections.Concurrent;
using Dragonfly.Native;

namespace Dragonfly;

internal static unsafe class PluginBridge
{
    private static Func<Plugin>? Factory;
    private static Func<World.EntityType[]>? EntityTypes;
    private static PluginState? CurrentState;
    private static PluginApi* Descriptor;
    private static readonly Dictionary<string, nint> EntityTypeStrings = [];
    private sealed record ScheduledWorldTask(Action<World.Tx> Callback, World.Task Task);
    private static readonly ConcurrentDictionary<ulong, ScheduledWorldTask> Scheduled = new();
    private static long NextScheduled;

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

        internal static void WritePlayerPacket(ulong invocation, PlayerId player, ulong packet)
        {
            var api = Api;
            if (api is null || api->PlayerPacketWrite == null || packet == 0) return;
            _ = api->PlayerPacketWrite(api->Context, invocation, player, packet);
        }

        internal static void SendPlayerTitle(ulong invocation, PlayerId player, Title title)
        {
            var api = Api;
            if (api is null || api->PlayerTitle == null) return;
            var text = Encoding.UTF8.GetBytes(title.Text());
            var subtitle = Encoding.UTF8.GetBytes(title.Subtitle());
            var actionText = Encoding.UTF8.GetBytes(title.ActionText());
            fixed (byte* textData = text)
            fixed (byte* subtitleData = subtitle)
            fixed (byte* actionTextData = actionText)
            {
                var view = new TitleView
                {
                    Text = new StringView { Data = textData, Length = (ulong)text.Length },
                    Subtitle = new StringView { Data = subtitleData, Length = (ulong)subtitle.Length },
                    ActionText = new StringView { Data = actionTextData, Length = (ulong)actionText.Length },
                    FadeInDurationNanoseconds = DurationNanoseconds(title.FadeInDuration(), nameof(title)),
                    DurationNanoseconds = DurationNanoseconds(title.Duration(), nameof(title)),
                    FadeOutDurationNanoseconds = DurationNanoseconds(title.FadeOutDuration(), nameof(title)),
                };
                _ = api->PlayerTitle(api->Context, invocation, player, view);
            }
        }

        internal static void SendPlayerScoreboard(ulong invocation, PlayerId player, Scoreboard scoreboard)
        {
            var api = Api;
            if (api is null || api->PlayerScoreboard == null) return;
            using var lease = new ScoreboardViewLease(scoreboard);
            _ = api->PlayerScoreboard(api->Context, invocation, player, lease.View);
        }

        private sealed class ScoreboardViewLease : IDisposable
        {
            private readonly List<nint> _allocations = [];

            internal ScoreboardViewLease(Scoreboard scoreboard)
            {
                try
                {
                    var lines = scoreboard.RawLines();
                    var views = AllocateArray<StringView>(lines.Length);
                    for (var index = 0; index < lines.Length; index++) views[index] = AllocateUtf8(lines[index]);
                    View = new ScoreboardView
                    {
                        Name = AllocateUtf8(scoreboard.Name()),
                        Lines = views,
                        LineCount = (ulong)lines.Length,
                        Padding = scoreboard.Padding() ? (byte)1 : (byte)0,
                        Descending = scoreboard.Descending() ? (byte)1 : (byte)0,
                    };
                }
                catch
                {
                    Dispose();
                    throw;
                }
            }

            internal ScoreboardView View { get; }

            public void Dispose()
            {
                foreach (var allocation in _allocations) NativeMemory.Free((void*)allocation);
                _allocations.Clear();
            }

            private StringView AllocateUtf8(string value)
            {
                var bytes = Encoding.UTF8.GetBytes(value);
                var data = AllocateArray<byte>(bytes.Length);
                if (bytes.Length != 0) bytes.CopyTo(new Span<byte>(data, bytes.Length));
                return new StringView { Data = data, Length = (ulong)bytes.Length };
            }

            private T* AllocateArray<T>(int length) where T : unmanaged
            {
                if (length == 0) return null;
                var pointer = (T*)NativeMemory.Alloc((nuint)length, (nuint)sizeof(T));
                if (pointer is null) throw new OutOfMemoryException();
                _allocations.Add((nint)pointer);
                return pointer;
            }
        }

        internal static void RemovePlayerScoreboard(ulong invocation, PlayerId player)
        {
            var api = Api;
            if (api is null || api->PlayerScoreboardRemove == null) return;
            _ = api->PlayerScoreboardRemove(api->Context, invocation, player);
        }

        internal static string PlayerString(ulong invocation, PlayerId player, uint kind) =>
            TryPlayerString(invocation, player, kind, out var value) ? value : string.Empty;

        internal static bool TryPlayerString(ulong invocation, PlayerId player, uint kind, out string value)
        {
            value = string.Empty;
            var api = Api;
            if (api is null || api->PlayerStringGet == null) return false;
            const ulong maxBytes = 16UL << 20;
            StringBuffer output = default;
            var status = api->PlayerStringGet(api->Context, invocation, player, kind, &output);
            if (status == Abi.Ok)
            {
                if (output.Length != 0 || output.Data is not null || output.Capacity != 0)
                    return false;
                return true;
            }
            for (var attempt = 0; attempt < 3; attempt++)
            {
                if (output.Length == 0 || output.Length > maxBytes)
                    return false;
                var data = GC.AllocateUninitializedArray<byte>(checked((int)output.Length));
                fixed (byte* pointer = data)
                {
                    output = new StringBuffer { Data = pointer, Capacity = (ulong)data.Length };
                    status = api->PlayerStringGet(api->Context, invocation, player, kind, &output);
                    if (status == Abi.Ok)
                    {
                        if (output.Data != pointer || output.Capacity != (ulong)data.Length || output.Length > output.Capacity)
                            return false;
                        value = Utf8(output);
                        return true;
                    }
                }
            }
            return false;
        }

        internal static void SendPlayerToast(ulong invocation, PlayerId player, string title, string message)
        {
            ArgumentNullException.ThrowIfNull(title);
            ArgumentNullException.ThrowIfNull(message);
            var api = Api;
            if (api is null || api->PlayerToast == null) return;
            var titleBytes = Encoding.UTF8.GetBytes(title);
            var messageBytes = Encoding.UTF8.GetBytes(message);
            fixed (byte* titleData = titleBytes)
            fixed (byte* messageData = messageBytes)
            {
                _ = api->PlayerToast(
                    api->Context,
                    invocation,
                    player,
                    new StringView { Data = titleData, Length = (ulong)titleBytes.Length },
                    new StringView { Data = messageData, Length = (ulong)messageBytes.Length });
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

        internal static PlayerStateValue RunPlayerAction(
            ulong invocation,
            PlayerId player,
            uint kind,
            PlayerStateValue value)
        {
            var api = Api;
            if (api is null || api->PlayerAction == null) return default;
            PlayerStateValue result = default;
            return api->PlayerAction(api->Context, invocation, player, kind, value, &result) == Abi.Ok
                ? result
                : default;
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
                Entity.FoodHealingSource food =>
                    (Abi.HealingSourceFood, food.QuickRegeneration ? (byte)1 : (byte)0, ""),
                Effect.InstantHealingSource => (Abi.HealingSourceInstant, (byte)0, ""),
                Effect.RegenerationHealingSource => (Abi.HealingSourceRegeneration, (byte)0, ""),
                OpaqueHealingSource opaque => (Abi.HealingSourceCustom, (byte)0, opaque.Name),
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

        internal static (double Damage, bool Vulnerable) HurtPlayer(
            ulong invocation,
            PlayerId player,
            double damage,
            World.DamageSource source) =>
            PlayerDamage(invocation, player, damage, source, false);

        internal static double FinalPlayerDamage(
            ulong invocation,
            PlayerId player,
            double damage,
            World.DamageSource source) =>
            PlayerDamage(invocation, player, damage, source, true).Damage;

        private static (double Damage, bool Vulnerable) PlayerDamage(
            ulong invocation,
            PlayerId player,
            double damage,
            World.DamageSource source,
            bool final)
        {
            ArgumentNullException.ThrowIfNull(source);
            var api = Api;
            if (api is null || (final ? api->PlayerFinalDamage == null : api->PlayerHurt == null))
                return default;

            uint kind;
            byte data = 0;
            var name = "";
            World.Entity? entity = null;
            World.Entity? secondaryEntity = null;
            World.Block? block = null;
            switch (source)
            {
                case Entity.AttackDamageSource value:
                    kind = Abi.DamageSourceAttack;
                    entity = value.Attacker;
                    break;
                case Block.DamageSource { Block: not null } value:
                    kind = Abi.DamageSourceBlock;
                    block = value.Block;
                    break;
                case Block.DamageSource:
                    kind = Abi.DamageSourceCustom;
                    name = "block.DamageSource";
                    break;
                case Entity.DrowningDamageSource:
                    kind = Abi.DamageSourceDrowning;
                    break;
                case Entity.ExplosionDamageSource:
                    kind = Abi.DamageSourceExplosion;
                    break;
                case Entity.FallDamageSource:
                    kind = Abi.DamageSourceFall;
                    break;
                case Block.FireDamageSource:
                    kind = Abi.DamageSourceFireKind;
                    break;
                case Entity.GlideDamageSource:
                    kind = Abi.DamageSourceGlide;
                    break;
                case Effect.InstantDamageSource:
                    kind = Abi.DamageSourceInstant;
                    break;
                case Block.LavaDamageSource:
                    kind = Abi.DamageSourceLava;
                    break;
                case Entity.LightningDamageSource:
                    kind = Abi.DamageSourceLightning;
                    break;
                case Block.MagmaDamageSource:
                    kind = Abi.DamageSourceMagma;
                    break;
                case Effect.PoisonDamageSource value:
                    kind = Abi.DamageSourcePoison;
                    data = value.Fatal ? (byte)1 : (byte)0;
                    break;
                case Entity.ProjectileDamageSource value:
                    kind = Abi.DamageSourceProjectile;
                    entity = value.Projectile;
                    secondaryEntity = value.Owner;
                    break;
                case Player.StarvationDamageSource:
                    kind = Abi.DamageSourceStarvation;
                    break;
                case Entity.SuffocationDamageSource:
                    kind = Abi.DamageSourceSuffocation;
                    break;
                case Enchantment.ThornsDamageSource value:
                    kind = Abi.DamageSourceThorns;
                    entity = value.Owner;
                    break;
                case Entity.VoidDamageSource:
                    kind = Abi.DamageSourceVoid;
                    break;
                case Effect.WitherDamageSource:
                    kind = Abi.DamageSourceWither;
                    break;
                case OpaqueDamageSource value:
                    kind = Abi.DamageSourceCustom;
                    name = value.Name;
                    break;
                default:
                    kind = Abi.DamageSourceCustom;
                    name = source.GetType().FullName ?? source.GetType().Name;
                    break;
            }

            var flags = DamageSourceFlags(source);
            var entityId = default(EntityId);
            var secondaryEntityId = default(EntityId);
            if (entity is not null && !TryEntityId(invocation, entity, out entityId) ||
                secondaryEntity is not null && !TryEntityId(invocation, secondaryEntity, out secondaryEntityId))
                return default;

            var blockIdentifier = Array.Empty<byte>();
            var blockProperties = Array.Empty<byte>();
            if (block is not null)
            {
                if (!BlockCodec.TryEncode(block, out var identifier, out blockProperties)) return default;
                blockIdentifier = Encoding.UTF8.GetBytes(identifier);
            }
            else if (kind == Abi.DamageSourceBlock)
            {
                return default;
            }

            var nameBytes = Encoding.UTF8.GetBytes(name);
            fixed (byte* nameData = nameBytes)
            fixed (byte* identifierData = blockIdentifier)
            fixed (byte* propertiesData = blockProperties)
            {
                var blockView = new BlockView
                {
                    Identifier = new StringView { Data = identifierData, Length = (ulong)blockIdentifier.Length },
                    PropertiesNbt = new StringView { Data = propertiesData, Length = (ulong)blockProperties.Length },
                };
                var view = new DamageSourceView
                {
                    Name = new StringView { Data = nameData, Length = (ulong)nameBytes.Length },
                    Kind = kind,
                    Flags = flags,
                    Entity = entityId,
                    SecondaryEntity = secondaryEntityId,
                    Block = block is null ? null : &blockView,
                    Data = data,
                };
                if (final)
                {
                    double result;
                    return api->PlayerFinalDamage(api->Context, invocation, player, damage, &view, &result) == Abi.Ok
                        ? (result, false)
                        : default;
                }
                else
                {
                    PlayerHurtResult result;
                    if (api->PlayerHurt(api->Context, invocation, player, damage, &view, &result) != Abi.Ok ||
                        result.Vulnerable > 1)
                        return default;
                    return (result.Damage, result.Vulnerable != 0);
                }
            }
        }

        private static bool TryEntityId(
            ulong invocation,
            World.Entity entity,
            out EntityId id)
        {
            if (World.TryEntityIdOf(entity, out id)) return true;
            var api = Api;
            if (api is null || api->EntityHandleEntity == null) return false;
            try
            {
                var handle = entity.H();
                byte found;
                EntityId resolved;
                if (api->EntityHandleEntity(api->Context, invocation, handle.Id, &resolved, &found) != Abi.Ok ||
                    found != 1 || resolved.Generation == 0)
                    return false;
                id = resolved;
                return true;
            }
            catch (InvalidOperationException)
            {
                id = default;
                return false;
            }
            catch (ArgumentException)
            {
                id = default;
                return false;
            }
        }

        private static uint DamageSourceFlags(World.DamageSource source)
        {
            var flags = source.ReducedByArmour() ? Abi.DamageSourceReducedByArmour : 0u;
            if (source.ReducedByResistance()) flags |= Abi.DamageSourceReducedByResistance;
            if (source.Fire()) flags |= Abi.DamageSourceFire;
            if (source.IgnoreTotem()) flags |= Abi.DamageSourceIgnoresTotem;
            if (source is Enchantment.AffectedDamageSource affected)
            {
                if (affected.AffectedByEnchantment(Item.FireProtection))
                    flags |= Abi.DamageSourceFireProtection;
                if (affected.AffectedByEnchantment(Item.FeatherFalling))
                    flags |= Abi.DamageSourceFeatherFalling;
                if (affected.AffectedByEnchantment(Item.BlastProtection))
                    flags |= Abi.DamageSourceBlastProtection;
                if (affected.AffectedByEnchantment(Item.ProjectileProtection))
                    flags |= Abi.DamageSourceProjectileProtection;
            }
            return flags;
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

        internal static void KnockBackPlayer(
            ulong invocation,
            PlayerId player,
            Vector3 source,
            double force,
            double height)
        {
            var api = Api;
            if (api is null || api->PlayerKnockBack == null) return;
            _ = api->PlayerKnockBack(
                api->Context,
                invocation,
                player,
                new Vec3 { X = source.X, Y = source.Y, Z = source.Z },
                force,
                height);
        }

        internal static bool PlayerUsingItem(ulong invocation, PlayerId player)
        {
            var api = Api;
            if (api is null || api->PlayerUsingItem == null) return false;
            byte value;
            return api->PlayerUsingItem(api->Context, invocation, player, &value) == Abi.Ok && value == 1;
        }

        internal static (Cube.Pos Position, bool Sleeping) PlayerSleeping(ulong invocation, PlayerId player)
        {
            var api = Api;
            if (api is null || api->PlayerSleeping == null) return default;
            BlockPos position;
            byte sleeping;
            if (api->PlayerSleeping(api->Context, invocation, player, &position, &sleeping) != Abi.Ok ||
                sleeping > 1)
                return default;
            return (new Cube.Pos(position.X, position.Y, position.Z), sleeping != 0);
        }

        internal static (Vector3 Position, World.Dimension? Dimension, bool Found) PlayerDeathPosition(
            ulong invocation,
            PlayerId player)
        {
            var api = Api;
            if (api is null || api->PlayerDeathPosition == null) return default;
            Vec3 position;
            DimensionView dimension;
            byte found;
            if (api->PlayerDeathPosition(api->Context, invocation, player, &position, &dimension, &found) != Abi.Ok ||
                found > 1)
                return default;
            if (found == 0) return default;
            var decoded = DecodeWorldDimension(dimension);
            return decoded is null ? default : (new Vector3(position.X, position.Y, position.Z), decoded, true);
        }

        internal static void CloseEntity(ulong invocation, EntityId entity)
        {
            var api = Api;
            if (api is null || api->EntityDespawn == null) return;
            _ = api->EntityDespawn(api->Context, invocation, entity);
        }

        internal static World.EntityHandle EntityHandle(ulong invocation, EntityId entity)
        {
            var api = Api;
            if (api is null || api->EntityHandle == null)
                throw new InvalidOperationException("entity handle is unavailable");
            EntityHandleId handle;
            if (api->EntityHandle(api->Context, invocation, entity, &handle) != Abi.Ok ||
                handle.Value == 0 || handle.Generation == 0)
                throw new InvalidOperationException("entity is no longer available");
            return new World.EntityHandle(handle);
        }

        // Entity construction is generated from Dragonfly now, but its managed
        // type/config lifetime bridge is implemented as a separate milestone.
        internal static World.EntityHandle NewEntity(
            World.EntitySpawnOpts opts,
            World.EntityType type,
            World.EntityConfig config)
        {
            ArgumentNullException.ThrowIfNull(opts);
            ArgumentNullException.ThrowIfNull(type);
            ArgumentNullException.ThrowIfNull(config);
            var api = Api;
            var state = CurrentState;
            if (api is null || api->EntityNew == null || Descriptor is null || state is null)
                throw new InvalidOperationException("entity construction is unavailable");
            var localType = state.Entities.TypeKey(type);
            var opaque = state.Entities.Prepare(opts, type, config);
            try
            {
                var data = state.Entities.PreparedData(opaque);
                var encodedType = Encoding.UTF8.GetBytes(type.EncodeEntity());
                var nameTag = Encoding.UTF8.GetBytes(data.Name ?? string.Empty);
                fixed (byte* encodedTypePointer = encodedType)
                fixed (byte* nameTagPointer = nameTag)
                {
                    var view = new EntityNewView
                    {
                        Options = new EntitySpawnOptions
                        {
                            Position = new Vec3 { X = data.Pos.X, Y = data.Pos.Y, Z = data.Pos.Z },
                            Rotation = new NativeRotation { Yaw = data.Rot.Yaw, Pitch = data.Rot.Pitch },
                            Velocity = new Vec3 { X = data.Vel.X, Y = data.Vel.Y, Z = data.Vel.Z },
                            NameTag = new StringView { Data = nameTagPointer, Length = checked((ulong)nameTag.Length) },
                        },
                        EntityType = new StringView { Data = encodedTypePointer, Length = checked((ulong)encodedType.Length) },
                        Plugin = (ulong)(nuint)Descriptor,
                        LocalType = localType,
                        Opaque = opaque,
                        FireDurationNanoseconds = checked(data.FireDuration.Ticks * 100),
                        AgeNanoseconds = checked(data.Age.Ticks * 100),
                    };
                    if (!opts.ID.TryWriteBytes(new Span<byte>(view.Id.Bytes, 16), bigEndian: true, out _))
                        throw new InvalidOperationException("entity UUID cannot be encoded");
                    EntityHandleId handle;
                    if (api->EntityNew(api->Context, &view, &handle) != Abi.Ok ||
                        handle.Value == 0 || handle.Generation == 0)
                        throw new InvalidOperationException("entity could not be created");
                    return state.Entities.BindHandle(opaque, handle.Value, handle.Generation);
                }
            }
            catch
            {
                state.Entities.ReleasePending(opaque);
                throw;
            }
        }

        internal static World.EntityType EntityHandleType(World.EntityHandle handle)
        {
            ArgumentNullException.ThrowIfNull(handle);
            var state = CurrentState ?? throw new InvalidOperationException("entity types are unavailable");
            try { return state.Entities.HandleType(handle.Id.Value, handle.Id.Generation); }
            catch (InvalidOperationException) { }
            var api = Api;
            if (api is null || api->EntityHandleType == null)
                throw new InvalidOperationException("entity type is unavailable");
            const int capacity = 256;
            byte* bytes = stackalloc byte[capacity];
            var output = new StringBuffer { Data = bytes, Capacity = capacity };
            if (api->EntityHandleType(api->Context, handle.Id, &output) != Abi.Ok ||
                output.Length == 0 || output.Length > capacity)
                throw new InvalidOperationException("entity handle is no longer available");
            return state.Entities.TypeByIdentifier(
                Encoding.UTF8.GetString(new ReadOnlySpan<byte>(bytes, checked((int)output.Length))));
        }

        internal static (World.Entity? Entity, bool Ok) EntityHandleEntity(
            ulong invocation,
            World.EntityHandle handle)
        {
            ArgumentNullException.ThrowIfNull(handle);
            var api = Api;
            if (api is null || api->EntityHandleEntity == null)
                throw new InvalidOperationException("entity handle is unavailable");
            EntityId entity;
            byte found;
            if (api->EntityHandleEntity(api->Context, invocation, handle.Id, &entity, &found) != Abi.Ok || found > 1)
                throw new InvalidOperationException("world transaction is no longer valid");
            if (found == 0) return (null, false);
            if (entity.Generation == 0)
                throw new InvalidOperationException("entity handle returned an invalid entity");
            var custom = CurrentState?.Entities.OpenedEntity(handle);
            return (custom ?? ResolveEntityPlayer(invocation, entity) ?? World.HostEntityFrom(invocation, entity), true);
        }

        internal static Guid EntityHandleUuid(World.EntityHandle handle)
        {
            ArgumentNullException.ThrowIfNull(handle);
            var api = Api;
            if (api is null || api->EntityHandleUuid == null)
                throw new InvalidOperationException("entity handle is unavailable");
            NativeUuid uuid;
            if (api->EntityHandleUuid(api->Context, handle.Id, &uuid) != Abi.Ok)
                throw new InvalidOperationException("entity handle is no longer available");
            return new Guid(new ReadOnlySpan<byte>(uuid.Bytes, 16), bigEndian: true);
        }

        internal static bool EntityHandleClosed(World.EntityHandle handle)
        {
            ArgumentNullException.ThrowIfNull(handle);
            var api = Api;
            if (api is null || api->EntityHandleClosed == null)
                throw new InvalidOperationException("entity handle is unavailable");
            byte closed;
            if (api->EntityHandleClosed(api->Context, handle.Id, &closed) != Abi.Ok || closed > 1)
                throw new InvalidOperationException("entity handle is no longer available");
            return closed != 0;
        }

        internal static void CloseEntityHandle(World.EntityHandle handle)
        {
            ArgumentNullException.ThrowIfNull(handle);
            var api = Api;
            if (api is null || api->EntityHandleClose == null) return;
            _ = api->EntityHandleClose(api->Context, handle.Id);
        }

        internal static World.Entity TransactionAddEntity(
            ulong invocation,
            World.EntityHandle handle) =>
            TransactionAddEntity(invocation, handle, null);

        internal static World.Entity TransactionAddEntity(
            ulong invocation,
            World.EntityHandle handle,
            Vector3 position) =>
            TransactionAddEntity(invocation, handle, (Vector3?)position);

        private static World.Entity TransactionAddEntity(
            ulong invocation,
            World.EntityHandle handle,
            Vector3? position)
        {
            ArgumentNullException.ThrowIfNull(handle);
            var api = Api;
            if (api is null || api->WorldEntityAdd == null)
                throw new InvalidOperationException("world transaction is unavailable");
            Vec3 nativePosition;
            Vec3* positionPointer = null;
            if (position is { } value)
            {
                nativePosition = new Vec3 { X = value.X, Y = value.Y, Z = value.Z };
                positionPointer = &nativePosition;
            }
            EntityId entity;
            if (api->WorldEntityAdd(api->Context, invocation, handle.Id, positionPointer, &entity) != Abi.Ok ||
                entity.Generation == 0)
                throw new InvalidOperationException("entity handle cannot be added to this transaction");
            return CurrentState?.Entities.OpenedEntity(handle) ??
                   ResolveEntityPlayer(invocation, entity) ??
                   World.HostEntityFrom(invocation, entity);
        }

        internal static World.EntityHandle TransactionRemoveEntity(ulong invocation, World.Entity entity)
        {
            ArgumentNullException.ThrowIfNull(entity);
            var api = Api;
            if (api is null || api->WorldEntityRemove == null)
                throw new InvalidOperationException("world transaction is unavailable");
            EntityId entityId;
            if (!World.TryEntityIdOf(entity, out entityId))
            {
                if (api->EntityHandleEntity == null)
                    throw new InvalidOperationException("entity handle is unavailable");
                var stable = entity.H();
                byte found;
                if (api->EntityHandleEntity(api->Context, invocation, stable.Id, &entityId, &found) != Abi.Ok ||
                    found != 1 || entityId.Generation == 0)
                    throw new InvalidOperationException("entity is no longer in this transaction");
            }
            EntityHandleId handle;
            if (api->WorldEntityRemove(api->Context, invocation, entityId, &handle) != Abi.Ok ||
                handle.Value == 0 || handle.Generation == 0)
                throw new InvalidOperationException("entity cannot be removed from this transaction");
            var detached = new World.EntityHandle(handle);
            return CurrentState?.Entities.CanonicalHandle(detached) ?? detached;
        }

        internal static World NewWorld(World.Config config)
        {
            ArgumentNullException.ThrowIfNull(config);
            var api = Api;
            if (api is null || api->WorldNew == null)
                throw new InvalidOperationException("world creation is unavailable");
            var dimension = config.Dim ?? World.Overworld;
            uint dimensionId = 0;
            DimensionView dimensionView = default;
            if (dimension is World.BuiltinDimension builtin && builtin.Id <= Abi.WorldDimensionEnd)
            {
                dimensionId = builtin.Id;
            }
            else
            {
                var range = dimension.Range();
                dimensionView = new DimensionView
                {
                    Custom = 1,
                    WaterEvaporates = dimension.WaterEvaporates() ? (byte)1 : (byte)0,
                    WeatherCycle = dimension.WeatherCycle() ? (byte)1 : (byte)0,
                    TimeCycle = dimension.TimeCycle() ? (byte)1 : (byte)0,
                    RangeMin = range.Min(),
                    RangeMax = range.Max(),
                    LavaSpreadNanoseconds = DurationNanoseconds(
                        dimension.LavaSpreadDuration(),
                        nameof(config)),
                };
            }
            dimensionView.Id = dimensionId;
            uint providerKind;
            byte[] providerPath;
            switch (config.Provider)
            {
                case null:
                case World.NopProvider:
                    providerKind = 0;
                    providerPath = [];
                    break;
                case MCDB.DB provider when TryWorldText(provider.Directory, 4096, out providerPath):
                    providerKind = 1;
                    break;
                case MCDB.DB:
                    throw new ArgumentException("MCDB provider path is invalid", nameof(config));
                default:
                    throw new ArgumentException("provider is not supported", nameof(config));
            }
            var view = new WorldConfigV1
            {
                StructSize = (uint)sizeof(WorldConfigV1),
                Dimension = dimensionId,
                DimensionView = dimensionView,
                ProviderKind = providerKind,
                ReadOnly = config.ReadOnly ? 1u : 0u,
                SaveIntervalNanoseconds = DurationNanoseconds(config.SaveInterval, nameof(config.SaveInterval)),
                ChunkUnloadIntervalNanoseconds = DurationNanoseconds(config.ChunkUnloadInterval, nameof(config.ChunkUnloadInterval)),
                RandomTickSpeed = config.RandomTickSpeed,
            };
            fixed (byte* path = providerPath)
            {
                view.ProviderPath = new StringView { Data = path, Length = (ulong)providerPath.Length };
                WorldId world;
                if (api->WorldNew(api->Context, &view, &world) != Abi.Ok || world.Value == 0)
                    throw new InvalidOperationException("world could not be created");
                return new World(0, world);
            }
        }

        internal static long DurationNanoseconds(TimeSpan value, string parameter)
        {
            try { return checked(value.Ticks * 100L); }
            catch (OverflowException) { throw new ArgumentOutOfRangeException(parameter); }
        }

        internal static TimeSpan PlayerDuration(long nanoseconds) => TimeSpan.FromTicks(nanoseconds / 100);

        internal static string? WorldName(ulong invocation, WorldId world)
        {
            var api = Api;
            if (api is null || api->WorldName == null) return null;
            StringBuffer probe = default;
            _ = api->WorldName(api->Context, invocation, world, &probe);
            if (probe.Length == 0 || probe.Length > int.MaxValue)
                return null;
            var data = new byte[checked((int)probe.Length)];
            fixed (byte* pointer = data)
            {
                var output = new StringBuffer { Data = pointer, Capacity = (ulong)data.Length };
                if (api->WorldName(api->Context, invocation, world, &output) != Abi.Ok ||
                    output.Length != (ulong)data.Length)
                    return null;
            }
            return Encoding.UTF8.GetString(data);
        }

        internal static Cube.Pos WorldSpawn(ulong invocation, WorldId world)
        {
            var api = Api;
            if (api is null || api->WorldSpawnGet == null) return default;
            BlockPos position;
            return api->WorldSpawnGet(api->Context, invocation, world, &position) == Abi.Ok
                ? new Cube.Pos(position.X, position.Y, position.Z)
                : default;
        }

        internal static void SetWorldSpawn(ulong invocation, WorldId world, Cube.Pos position)
        {
            var api = Api;
            if (api is null || api->WorldSpawnSet == null) return;
            _ = api->WorldSpawnSet(
                api->Context,
                invocation,
                world,
                new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() });
        }

        internal static Cube.Pos WorldPlayerSpawn(ulong invocation, WorldId world, Guid player)
        {
            var api = Api;
            if (api is null || api->WorldPlayerSpawnGet == null) return default;
            NativeUuid native = default;
            if (!player.TryWriteBytes(new Span<byte>(native.Bytes, 16), bigEndian: true, out var written) || written != 16)
                return default;
            BlockPos position;
            return api->WorldPlayerSpawnGet(api->Context, invocation, world, native, &position) == Abi.Ok
                ? new Cube.Pos(position.X, position.Y, position.Z)
                : default;
        }

        internal static void SetWorldPlayerSpawn(ulong invocation, WorldId world, Guid player, Cube.Pos position)
        {
            var api = Api;
            if (api is null || api->WorldPlayerSpawnSet == null) return;
            NativeUuid native = default;
            if (!player.TryWriteBytes(new Span<byte>(native.Bytes, 16), bigEndian: true, out var written) || written != 16)
                return;
            _ = api->WorldPlayerSpawnSet(
                api->Context,
                invocation,
                world,
                native,
                new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() });
        }

        internal static void SaveWorld(ulong invocation, WorldId world)
        {
            var api = Api;
            if (api is null || api->WorldSave == null) return;
            _ = api->WorldSave(api->Context, invocation, world);
        }

        internal static void CloseWorld(ulong invocation, WorldId world)
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

        internal static bool HasPlayerCooldown(ulong invocation, PlayerId player, World.Item? item) =>
            PlayerCooldown(invocation, player, Abi.PlayerCooldownHas, item, TimeSpan.Zero);

        internal static void SetPlayerCooldown(ulong invocation, PlayerId player, World.Item? item, TimeSpan cooldown) =>
            _ = PlayerCooldown(invocation, player, Abi.PlayerCooldownSet, item, cooldown);

        private static bool PlayerCooldown(
            ulong invocation,
            PlayerId player,
            uint operation,
            World.Item? item,
            TimeSpan duration)
        {
            var api = Api;
            if (item is null || api is null || api->PlayerCooldown == null ||
                !ItemCodec.TryEncode(item, out var identifier, out var metadata))
                return false;
            var identifierBytes = Encoding.UTF8.GetBytes(identifier);
            fixed (byte* identifierData = identifierBytes)
            {
                byte active;
                var status = api->PlayerCooldown(
                    api->Context,
                    invocation,
                    player,
                    operation,
                    new StringView { Data = identifierData, Length = (ulong)identifierBytes.Length },
                    metadata,
                    DurationNanoseconds(duration, nameof(duration)),
                    &active);
                return status == Abi.Ok && active != 0;
            }
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

        internal static void SetPlayerEntityVisible(
            ulong invocation,
            PlayerId player,
            World.Entity entity,
            bool visible)
        {
            ArgumentNullException.ThrowIfNull(entity);
            var api = Api;
            if (api is null || api->PlayerEntityVisibility == null ||
                !TryEntityId(invocation, entity, out var entityId)) return;
            _ = api->PlayerEntityVisibility(
                api->Context,
                invocation,
                player,
                entityId,
                visible ? (byte)1 : (byte)0);
        }

        internal static Skin PlayerSkin(ulong invocation, PlayerId player)
        {
            var api = Api;
            if (api is null || api->PlayerSkinOpen == null || api->PlayerSkinClose == null)
                return new Skin(0, 0);
            ulong snapshot;
            SkinInfo info;
            if (api->PlayerSkinOpen(api->Context, invocation, player, &snapshot, &info) != Abi.Ok || snapshot == 0)
                return new Skin(0, 0);
            try
            {
                return EventSkin(invocation, snapshot);
            }
            catch (InvalidOperationException)
            {
                return new Skin(0, 0);
            }
            finally
            {
                api->PlayerSkinClose(api->Context, invocation, snapshot);
            }
        }

        internal static void SetPlayerSkin(ulong invocation, PlayerId player, Skin skin)
        {
            ArgumentNullException.ThrowIfNull(skin);
            var api = Api;
            if (api is null || api->PlayerSkinSet == null) return;
            using var lease = new SkinViewLease(skin);
            var view = lease.View;
            _ = api->PlayerSkinSet(api->Context, invocation, player, &view);
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
                using var invocationScope = InvocationContext.Enter(invocation);
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

        internal static (World.Block? Block, bool Ok) BlockByName(
            string name,
            IReadOnlyDictionary<string, object?>? properties)
        {
            ArgumentNullException.ThrowIfNull(name);
            var api = Api;
            if (api is null || api->BlockByName == null) return (null, false);
            var nameBytes = Encoding.UTF8.GetBytes(name);
            if (nameBytes.Length == 0 || nameBytes.Length > 256) return (null, false);
            var encodedProperties = BlockPropertyCodec.Encode(properties);
            if (encodedProperties.Length > 64 * 1024) return (null, false);

            fixed (byte* nameData = nameBytes)
            fixed (byte* propertiesData = encodedProperties)
            {
                var nameView = new StringView { Data = nameData, Length = (ulong)nameBytes.Length };
                var propertiesView = new StringView
                {
                    Data = propertiesData,
                    Length = (ulong)encodedProperties.Length,
                };
                byte found;
                var data = new BlockData();
                var status = api->BlockByName(
                    api->Context, nameView, propertiesView, &found, &data);
                if (found == 0) return (null, false);
                if (found != 1 || data.Identifier.Length == 0 || data.Identifier.Length > 256 ||
                    data.PropertiesNbt.Length > 64 * 1024) return (null, false);
                var identifierBytes = new byte[checked((int)data.Identifier.Length)];
                var propertyBytes = new byte[checked((int)data.PropertiesNbt.Length)];
                fixed (byte* identifierData = identifierBytes)
                fixed (byte* propertyData = propertyBytes)
                {
                    data.Identifier = new StringBuffer
                    {
                        Data = identifierData,
                        Capacity = (ulong)identifierBytes.Length,
                    };
                    data.PropertiesNbt = new StringBuffer
                    {
                        Data = propertyData,
                        Capacity = (ulong)propertyBytes.Length,
                    };
                    if (api->BlockByName(
                            api->Context, nameView, propertiesView, &found, &data) != Abi.Ok || found != 1)
                        return (null, false);
                }
                return (BlockCodec.Decode(Encoding.UTF8.GetString(identifierBytes), propertyBytes), true);
            }
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

        internal static Cube.Range WorldRange(ulong invocation, WorldId world)
        {
            var api = Api;
            if (api is null || api->WorldRange == null)
                throw new InvalidOperationException("world is unavailable");
            BlockRange range;
            if (api->WorldRange(api->Context, invocation, world, &range) != Abi.Ok || range.Min > range.Max)
                throw new InvalidOperationException("world is no longer valid");
            return new Cube.Range(range.Min, range.Max);
        }

        internal static int WorldTime(ulong invocation, WorldId world)
        {
            var api = Api;
            if (api is null || api->WorldTimeGet == null)
                throw new InvalidOperationException("world is unavailable");
            long value;
            if (api->WorldTimeGet(api->Context, invocation, world, &value) != Abi.Ok)
                throw new InvalidOperationException("world is no longer valid");
            return checked((int)value);
        }

        internal static void SetWorldTime(ulong invocation, WorldId world, int value)
        {
            var api = Api;
            if (api is null || api->WorldTimeSet == null ||
                api->WorldTimeSet(api->Context, invocation, world, value) != Abi.Ok)
                return;
        }

        internal static World.Dimension WorldDimension(ulong invocation, WorldId world)
        {
            var api = Api;
            if (api is null || api->WorldDimensionGet == null) return World.Overworld;
            DimensionView value;
            if (api->WorldDimensionGet(api->Context, invocation, world, &value) != Abi.Ok)
                return World.Overworld;
            return DecodeWorldDimension(value) ?? World.Overworld;
        }

        private static World.Dimension? DecodeWorldDimension(DimensionView value)
        {
            if (value.Custom == 0)
            {
                return value.Id switch
                {
                    Abi.WorldDimensionOverworld => World.Overworld,
                    Abi.WorldDimensionNether => World.Nether,
                    Abi.WorldDimensionEnd => World.End,
                    _ => null,
                };
            }
            if (value.Custom != 1 || value.Id != 0 || value.WaterEvaporates > 1 ||
                value.WeatherCycle > 1 || value.TimeCycle > 1)
                return null;
            return new World.TransportDimension(
                new Cube.Range(value.RangeMin, value.RangeMax),
                value.WaterEvaporates != 0,
                PlayerDuration(value.LavaSpreadNanoseconds),
                value.WeatherCycle != 0,
                value.TimeCycle != 0);
        }

        internal static bool WorldTimeCycle(ulong invocation, WorldId world)
        {
            var api = Api;
            if (api is null || api->WorldTimeCycleGet == null) return false;
            byte value;
            return api->WorldTimeCycleGet(api->Context, invocation, world, &value) == Abi.Ok && value == 1;
        }

        internal static void SetWorldTimeCycle(ulong invocation, WorldId world, bool value)
        {
            var api = Api;
            if (api is null || api->WorldTimeCycleSet == null) return;
            _ = api->WorldTimeCycleSet(api->Context, invocation, world, value ? (byte)1 : (byte)0);
        }

        internal static void SetWorldRequiredSleepDuration(ulong invocation, WorldId world, TimeSpan duration)
        {
            var api = Api;
            if (api is null || api->WorldRequiredSleepDurationSet == null) return;
            _ = api->WorldRequiredSleepDurationSet(
                api->Context, invocation, world, DurationNanoseconds(duration, nameof(duration)));
        }

        internal static World.GameMode WorldDefaultGameMode(ulong invocation, WorldId world)
        {
            var api = Api;
            if (api is null || api->WorldDefaultGameModeGet == null) return World.GameModeSurvival;
            long value;
            if (api->WorldDefaultGameModeGet(api->Context, invocation, world, &value) != Abi.Ok)
                return World.GameModeSurvival;
            return World.GameModeFromDescriptor(value);
        }

        internal static void SetWorldDefaultGameMode(ulong invocation, WorldId world, World.GameMode mode)
        {
            var descriptor = World.GameModeDescriptor(mode);
            var api = Api;
            if (api is null || api->WorldDefaultGameModeSet == null) return;
            _ = api->WorldDefaultGameModeSet(api->Context, invocation, world, descriptor);
        }

        internal static void SetWorldTickRange(ulong invocation, WorldId world, int value)
        {
            var api = Api;
            if (api is null || api->WorldTickRangeSet == null) return;
            _ = api->WorldTickRangeSet(api->Context, invocation, world, value);
        }

        internal static World.Difficulty WorldDifficulty(ulong invocation, WorldId world)
        {
            var api = Api;
            if (api is null || api->WorldDifficultyGet == null) return World.DifficultyNormal;
            DifficultyView value;
            if (api->WorldDifficultyGet(api->Context, invocation, world, &value) != Abi.Ok)
                return World.DifficultyNormal;
            return World.DifficultyFromView(value);
        }

        internal static void SetWorldDifficulty(ulong invocation, WorldId world, World.Difficulty difficulty)
        {
            var value = World.DifficultyView(difficulty);
            var api = Api;
            if (api is null || api->WorldDifficultySet == null) return;
            _ = api->WorldDifficultySet(api->Context, invocation, world, value);
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

        internal static void PlayWorldSound(ulong invocation, Vector3 position, World.Sound sound) =>
            PlaySound(invocation, default, default, position, sound, world: true);

        internal static void PlayPlayerSound(ulong invocation, PlayerId player, World.Sound sound) =>
            PlaySound(invocation, default, player, default, sound, world: false);

        private static void PlaySound(
            ulong invocation,
            WorldId worldId,
            PlayerId player,
            Vector3 position,
            World.Sound sound,
            bool world)
        {
            ArgumentNullException.ThrowIfNull(sound);
            var api = Api;
            if (api is null || world && api->WorldSoundPlay == null || !world && api->PlayerSoundPlay == null)
                throw new InvalidOperationException(world ? "world transaction is unavailable" : "player is unavailable");
            if (!SoundCodec.TryEncode(sound, out var encoded))
                throw new ArgumentException("sound type is not registered", nameof(sound));

            var identifierBytes = Array.Empty<byte>();
            var propertyBytes = Array.Empty<byte>();
            if (encoded.Block is not null)
            {
                if (!BlockCodec.TryEncode(encoded.Block, out var identifier, out propertyBytes))
                    throw new ArgumentException("sound block type is not registered", nameof(sound));
                identifierBytes = Encoding.UTF8.GetBytes(identifier);
            }

            using var itemLease = encoded.Item is null
                ? null
                : new ItemViewLease(Item.NewStack(encoded.Item, 1));
            var item = itemLease?.View ?? default;
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
                var view = new SoundViewV1
                {
                    Kind = encoded.Kind,
                    Data = encoded.Data,
                    Integer = encoded.Integer,
                    Flags = encoded.Flags,
                    Scalar = encoded.Scalar,
                    Block = encoded.Block is null ? null : &block,
                    Item = encoded.Item is null ? null : &item,
                };
                var status = world
                    ? api->WorldSoundPlay(
                        api->Context,
                        invocation,
                        worldId,
                        new Vec3 { X = position.X, Y = position.Y, Z = position.Z },
                        &view)
                    : api->PlayerSoundPlay(api->Context, invocation, player, &view);
                if (status != Abi.Ok)
                    throw new InvalidOperationException(world
                        ? "world transaction is no longer valid"
                        : "player is no longer available");
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

        internal static World TransactionWorld(ulong invocation)
        {
            var api = Api;
            if (api is null || api->WorldCurrent == null)
                throw new InvalidOperationException("world transaction is unavailable");
            WorldId world;
            if (api->WorldCurrent(api->Context, invocation, &world) != Abi.Ok || world.Value == 0)
                throw new InvalidOperationException("world transaction is no longer valid");
            return new World(invocation, world);
        }

        internal static IEnumerable<World.Entity> TransactionEntities(
            ulong invocation,
            bool playersOnly)
        {
            var world = TransactionWorld(invocation);
            return new TransactionEntityEnumerable(invocation, world.Id, playersOnly, null);
        }

        internal static IEnumerable<World.Entity> TransactionEntitiesWithin(
            ulong invocation,
            Cube.BBox box)
        {
            var world = TransactionWorld(invocation);
            return new TransactionEntityEnumerable(invocation, world.Id, playersOnly: false, box);
        }

        private sealed class TransactionEntityEnumerable(
            ulong invocation,
            WorldId world,
            bool playersOnly,
            Cube.BBox? box) : IEnumerable<World.Entity>
        {
            public IEnumerator<World.Entity> GetEnumerator() =>
                new TransactionEntityEnumerator(invocation, world, playersOnly, box);

            System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator() => GetEnumerator();
        }

        private sealed class TransactionEntityEnumerator : IEnumerator<World.Entity>
        {
            private readonly HostApi* _api;
            private readonly ulong _invocation;
            private readonly bool _playersOnly;
            private ulong _iterator;
            private bool _disposed;

            internal TransactionEntityEnumerator(
                ulong invocation,
                WorldId world,
                bool playersOnly,
                Cube.BBox? box)
            {
                _api = Api;
                if (_api is null ||
                    _api->WorldEntityIteratorNext == null || _api->WorldEntityIteratorClose == null)
                    throw new InvalidOperationException("world transaction is unavailable");

                _invocation = invocation;
                _playersOnly = playersOnly;
                ulong iterator = 0;
                int status;
                if (box is { } bounds)
                {
                    if (_api->WorldEntitiesWithinOpen == null)
                        throw new InvalidOperationException("world transaction is unavailable");
                    var min = bounds.Min();
                    var max = bounds.Max();
                    status = _api->WorldEntitiesWithinOpen(
                        _api->Context,
                        invocation,
                        world,
                        new Dragonfly.Native.BBox
                        {
                            Min = new Vec3 { X = min.X, Y = min.Y, Z = min.Z },
                            Max = new Vec3 { X = max.X, Y = max.Y, Z = max.Z },
                        },
                        &iterator);
                }
                else
                {
                    if (_api->WorldEntityIteratorOpen == null)
                        throw new InvalidOperationException("world transaction is unavailable");
                    status = _api->WorldEntityIteratorOpen(
                        _api->Context,
                        invocation,
                        world,
                        playersOnly ? (byte)1 : (byte)0,
                        &iterator);
                }
                if (status != Abi.Ok || iterator == 0)
                {
                    if (iterator != 0)
                        _api->WorldEntityIteratorClose(_api->Context, invocation, iterator);
                    throw new InvalidOperationException("world transaction is no longer valid");
                }
                _iterator = iterator;
            }

            public World.Entity Current { get; private set; } = null!;
            object System.Collections.IEnumerator.Current => Current;

            public bool MoveNext()
            {
                if (_disposed) return false;
                EntityId entity;
                byte found;
                if (_api->WorldEntityIteratorNext(
                        _api->Context,
                        _invocation,
                        _iterator,
                        &entity,
                        &found) != Abi.Ok || found > 1)
                {
                    Dispose();
                    throw new InvalidOperationException("world transaction is no longer valid");
                }
                if (found == 0)
                {
                    Dispose();
                    return false;
                }
                if (entity.Generation == 0)
                {
                    Dispose();
                    throw new InvalidOperationException("world transaction returned an invalid entity");
                }
                var player = ResolveEntityPlayer(_invocation, entity);
                if (_playersOnly && player is null)
                {
                    Dispose();
                    throw new InvalidOperationException("world transaction returned a non-player entity");
                }
                Current = player ?? World.HostEntityFrom(_invocation, entity);
                return true;
            }

            public void Reset() => throw new NotSupportedException();

            public void Dispose()
            {
                if (_disposed) return;
                _disposed = true;
                _api->WorldEntityIteratorClose(_api->Context, _invocation, _iterator);
                _iterator = 0;
            }
        }

        internal static IEnumerable<Player> ServerPlayers(ulong invocation) =>
            new ServerPlayerEnumerable(invocation);

        internal static int ServerMaxPlayerCount()
        {
            var api = Api;
            if (api is null || api->ServerMaxPlayerCount == null)
                throw new InvalidOperationException("server is unavailable");
            long count;
            if (api->ServerMaxPlayerCount(api->Context, &count) != Abi.Ok || count < 0)
                throw new InvalidOperationException("server maximum player count failed");
            return checked((int)count);
        }

        internal static int ServerPlayerCount()
        {
            var api = Api;
            if (api is null || api->ServerPlayerCount == null)
                throw new InvalidOperationException("server is unavailable");
            long count;
            if (api->ServerPlayerCount(api->Context, &count) != Abi.Ok || count < 0)
                throw new InvalidOperationException("server player count failed");
            return checked((int)count);
        }

        internal static World ServerWorld(uint dimension)
        {
            var api = Api;
            if (api is null || api->ServerWorld == null)
                throw new InvalidOperationException("server is unavailable");
            WorldId world;
            if (api->ServerWorld(api->Context, dimension, &world) != Abi.Ok || world.Value == 0)
                throw new InvalidOperationException("server world is unavailable");
            return new World(0, world);
        }

        internal static World.Task ScheduleWorld(World world, TimeSpan delay, Action<World.Tx> callback)
        {
            ArgumentNullException.ThrowIfNull(world);
            ArgumentNullException.ThrowIfNull(callback);
            var delayNanoseconds = checked(delay.Ticks * 100);
            var next = Interlocked.Increment(ref NextScheduled);
            if (next <= 0) throw new InvalidOperationException("world callback IDs exhausted");
            var id = (ulong)next;
            var task = new World.Task(id);
            if (!Scheduled.TryAdd(id, new ScheduledWorldTask(callback, task)))
                throw new InvalidOperationException("world callback ID collision");
            var api = Api;
            if (api is not null && api->WorldSchedule != null && Descriptor is not null &&
                api->WorldSchedule(api->Context, world.Id, (ulong)(nuint)Descriptor, id, delayNanoseconds) == Abi.Ok)
                return task;
            Scheduled.TryRemove(id, out _);
            task.Complete(Abi.WorldTaskFailed);
            return task;
        }

        internal static bool CancelWorldTask(ulong callback)
        {
            var api = Api;
            if (api is null || api->WorldTaskCancel == null || Descriptor is null) return false;
            byte cancelled;
            return api->WorldTaskCancel(api->Context, (ulong)(nuint)Descriptor, callback, &cancelled) == Abi.Ok &&
                   cancelled == 1;
        }

        internal static (World.EntityHandle? Player, bool Ok) ServerPlayer(Guid uuid)
        {
            var api = Api;
            if (api is null || api->ServerPlayer == null)
                throw new InvalidOperationException("server is unavailable");
            NativeUuid native = default;
            if (!uuid.TryWriteBytes(new Span<byte>(native.Bytes, 16), bigEndian: true, out var written) || written != 16)
                throw new InvalidOperationException("could not encode player UUID");
            EntityHandleId player;
            byte found;
            if (api->ServerPlayer(api->Context, native, &player, &found) != Abi.Ok || found > 1)
                throw new InvalidOperationException("server player lookup failed");
            if (found == 0) return (null, false);
            if (player.Value == 0 || player.Generation == 0)
                throw new InvalidOperationException("server returned an invalid player handle");
            return (new World.EntityHandle(player), true);
        }

        internal static (World.EntityHandle? Player, bool Ok) ServerPlayerByName(string name)
        {
            ArgumentNullException.ThrowIfNull(name);
            var api = Api;
            if (api is null || api->ServerPlayerByName == null)
                throw new InvalidOperationException("server is unavailable");
            var bytes = Encoding.UTF8.GetBytes(name);
            if (bytes.Length > 256)
                return (null, false);
            fixed (byte* data = bytes)
            {
                EntityHandleId player;
                byte found;
                if (api->ServerPlayerByName(
                        api->Context,
                        new StringView { Data = data, Length = (ulong)bytes.Length },
                        &player,
                        &found) != Abi.Ok || found > 1)
                    throw new InvalidOperationException("server player lookup failed");
                if (found == 0) return (null, false);
                if (player.Value == 0 || player.Generation == 0)
                    throw new InvalidOperationException("server returned an invalid player handle");
                return (new World.EntityHandle(player), true);
            }
        }

        internal static (World.EntityHandle? Player, bool Ok) ServerPlayerByXUID(string xuid)
        {
            ArgumentNullException.ThrowIfNull(xuid);
            var api = Api;
            if (api is null || api->ServerPlayerByXuid == null)
                throw new InvalidOperationException("server is unavailable");
            var bytes = Encoding.UTF8.GetBytes(xuid);
            if (bytes.Length > 64)
                return (null, false);
            fixed (byte* data = bytes)
            {
                EntityHandleId player;
                byte found;
                if (api->ServerPlayerByXuid(
                        api->Context,
                        new StringView { Data = data, Length = (ulong)bytes.Length },
                        &player,
                        &found) != Abi.Ok || found > 1)
                    throw new InvalidOperationException("server player lookup failed");
                if (found == 0) return (null, false);
                if (player.Value == 0 || player.Generation == 0)
                    throw new InvalidOperationException("server returned an invalid player handle");
                return (new World.EntityHandle(player), true);
            }
        }

        internal static Guid PlayerUUID(PlayerId player)
        {
            Span<byte> bytes = stackalloc byte[16];
            for (var index = 0; index < bytes.Length; index++) bytes[index] = player.Bytes[index];
            return new Guid(bytes, bigEndian: true);
        }

        internal static string PlayerXUID(ulong invocation, PlayerId player)
        {
            var api = Api;
            if (api is null || api->PlayerXuid == null)
                throw new InvalidOperationException("player is unavailable");
            const int capacity = 64;
            byte* data = stackalloc byte[capacity];
            var output = new StringBuffer { Data = data, Capacity = capacity };
            if (api->PlayerXuid(api->Context, invocation, player, &output) != Abi.Ok ||
                output.Data != data || output.Capacity != capacity || output.Length > capacity)
                throw new InvalidOperationException("player is no longer valid");
            return Utf8(output);
        }

        private sealed class ServerPlayerEnumerable(ulong invocation) : IEnumerable<Player>
        {
            public IEnumerator<Player> GetEnumerator() => new ServerPlayerEnumerator(invocation);

            System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator() => GetEnumerator();
        }

        private sealed class ServerPlayerEnumerator : IEnumerator<Player>
        {
            private readonly HostApi* _api;
            private readonly ulong _invocation;
            private ulong _iterator;
            private bool _disposed;

            internal ServerPlayerEnumerator(ulong invocation)
            {
                _api = Api;
                if (_api is null || _api->ServerPlayersOpen == null ||
                    _api->ServerPlayersNext == null || _api->ServerPlayersClose == null)
                    throw new InvalidOperationException("server is unavailable");
                _invocation = invocation;
                ulong iterator = 0;
                if (_api->ServerPlayersOpen(_api->Context, invocation, &iterator) != Abi.Ok || iterator == 0)
                {
                    if (iterator != 0)
                        _api->ServerPlayersClose(_api->Context, invocation, iterator);
                    throw new InvalidOperationException("server player iteration failed");
                }
                _iterator = iterator;
            }

            public Player Current { get; private set; } = null!;
            object System.Collections.IEnumerator.Current => Current;

            public bool MoveNext()
            {
                if (_disposed) return false;
                const int maxPlayerNameBytes = 256;
                byte* name = stackalloc byte[maxPlayerNameBytes];
                PlayerSnapshotBuffer snapshot = new()
                {
                    Name = new StringBuffer { Data = name, Capacity = maxPlayerNameBytes },
                };
                ulong playerInvocation;
                byte found;
                if (_api->ServerPlayersNext(
                        _api->Context,
                        _invocation,
                        _iterator,
                        &playerInvocation,
                        &snapshot,
                        &found) != Abi.Ok || found > 1)
                {
                    Dispose();
                    throw new InvalidOperationException("server player iteration failed");
                }
                if (found == 0)
                {
                    if (playerInvocation != 0)
                    {
                        Dispose();
                        throw new InvalidOperationException("server returned an invalid player invocation");
                    }
                    Dispose();
                    return false;
                }
                if (playerInvocation == 0 || snapshot.Player.Generation == 0 ||
                    snapshot.Name.Data != name || snapshot.Name.Capacity != maxPlayerNameBytes ||
                    snapshot.Name.Length == 0 || snapshot.Name.Length > maxPlayerNameBytes ||
                    !double.IsFinite(snapshot.Position.X) || !double.IsFinite(snapshot.Position.Y) ||
                    !double.IsFinite(snapshot.Position.Z))
                {
                    Dispose();
                    throw new InvalidOperationException("server returned an invalid player");
                }
                Current = new Player(
                    snapshot.Player,
                    Utf8(snapshot.Name),
                    TimeSpan.FromMilliseconds(Math.Min(
                        (double)snapshot.LatencyMilliseconds,
                        TimeSpan.MaxValue.TotalMilliseconds)),
                    new Vector3(snapshot.Position.X, snapshot.Position.Y, snapshot.Position.Z),
                    invocation: playerInvocation);
                return true;
            }

            public void Reset() => throw new NotSupportedException();

            public void Dispose()
            {
                if (_disposed) return;
                _disposed = true;
                _api->ServerPlayersClose(_api->Context, _invocation, _iterator);
                _iterator = 0;
            }
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

        internal static int WorldHighestLightBlocker(ulong invocation, WorldId world, int x, int z)
        {
            var api = Api;
            if (api is null || api->WorldHighestLightBlocker == null)
                throw new InvalidOperationException("world is unavailable");
            int y;
            if (api->WorldHighestLightBlocker(api->Context, invocation, world, x, z, &y) != Abi.Ok)
                throw new InvalidOperationException("world is no longer valid");
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

    internal static PluginApi* Initialize(
        Func<Plugin> factory,
        string id,
        ulong subscriptions,
        Func<World.EntityType[]> entityTypes)
    {
        if (Descriptor is not null) return Descriptor;
        Factory = factory;
        EntityTypes = entityTypes;
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
            EntityTypeCount = (void*)(delegate* unmanaged[Cdecl]<void*, ulong>)&EntityTypeCount,
            EntityTypeAt = (void*)(delegate* unmanaged[Cdecl]<void*, ulong, EntityTypeDescriptorV2*, int>)&EntityTypeAt,
            HandleEntity = (void*)(delegate* unmanaged[Cdecl]<void*, ulong, uint, ulong, void*, void*, int>)&HandleEntity,
            HandleCommand = &HandleCommand,
            CommandEnumOptions = &CommandEnumOptions,
            SetHost = &SetHost,
            Destroy = &Destroy,
            HandleEvent = &HandleEvent,
            HandleScheduled = &HandleScheduled,
            Allow = &Allow,
        };
        return Descriptor;
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static void* Create()
    {
        try
        {
            CommandRegistry.Clear();
            var state = new PluginState(Factory!, EntityTypes!);
            CurrentState = state;
            return (void*)GCHandle.ToIntPtr(GCHandle.Alloc(state));
        }
        catch { return null; }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int SetHost(void* instance, void* host)
    {
        if (host is null) return Abi.Error;
        var header = (HostHeader*)host;
        if (header->Version != Abi.HostVersion || header->Size < (uint)sizeof(HostApi)) return Abi.Error;
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
        if (instance is not null)
        {
            var handle = GCHandle.FromIntPtr((nint)instance);
            (handle.Target as PluginState)?.Dispose();
            handle.Free();
        }
        CurrentState = null;
        foreach (var pointer in EntityTypeStrings.Values)
            if (pointer != 0) NativeMemory.Free((void*)pointer);
        EntityTypeStrings.Clear();
        Host.Api = null;
        CommandRegistry.Clear();
        foreach (var task in Scheduled.Values) task.Task.Complete(Abi.WorldTaskFailed);
        Scheduled.Clear();
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
    private static ulong EntityTypeCount(void* instance)
    {
        try { return instance is null ? 0 : checked((ulong)State(instance).Entities.Count); }
        catch { return 0; }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int EntityTypeAt(void* instance, ulong index, EntityTypeDescriptorV2* output)
    {
        try
        {
            if (instance is null || output is null || index > int.MaxValue) return Abi.Error;
            var (key, type) = State(instance).Entities.TypeAt((int)index);
            var identifier = PersistentEntityTypeString(type.EncodeEntity());
            *output = new EntityTypeDescriptorV2
            {
                SaveId = identifier,
                NetworkId = identifier,
                TypeKey = key,
            };
            return Abi.Ok;
        }
        catch { return Abi.Error; }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int HandleEntity(
        void* instance,
        ulong localType,
        uint operation,
        ulong entityInstance,
        void* input,
        void* state)
    {
        try
        {
            if (instance is null) return Abi.Error;
            var entities = State(instance).Entities;
            if (operation == Abi.EntityOperationAdopt)
            {
                if (input is null || state is null || entityInstance != 0) return Abi.Error;
                *(ulong*)state = entities.Adopt(localType, *(ulong*)input);
                return Abi.Ok;
            }
            if (operation == Abi.EntityOperationDestroy && entityInstance != 0)
            {
                entities.Destroy(entityInstance);
                return Abi.Ok;
            }
            if (operation == Abi.EntityOperationDecodeNbt)
            {
                if (localType == 0 || entityInstance != 0 || input is null || state is null) return Abi.Error;
                var value = (EntityExactInput*)input;
                var result = (EntityExactState*)state;
                if (value->Data is null || value->Nbt.Length > Nbt.MaxBytes ||
                    value->Nbt.Length != 0 && value->Nbt.Data is null) return Abi.Error;
                var data = new World.EntityData();
                ReadEntityData(data, value->Data);
                result->Instance = entities.DecodeNBT(
                    localType,
                    EntityNbtCodec.Decode(new ReadOnlySpan<byte>(value->Nbt.Data, checked((int)value->Nbt.Length))),
                    data);
                return WriteEntityData(data, value->Data) ? Abi.Ok : Abi.Error;
            }
            if (entityInstance == 0 || input is null && operation != Abi.EntityOperationReleaseOpen)
                return Abi.Error;
            var exactInput = (EntityExactInput*)input;
            var exactState = (EntityExactState*)state;
            if (operation == Abi.EntityOperationReleaseOpen)
            {
                entities.ReleaseOpen(entityInstance);
                return Abi.Ok;
            }
            if (exactInput->Data is null || exactState is null) return Abi.Error;
            var entityData = operation is Abi.EntityOperationEncodeNbt or Abi.EntityOperationOpen
                ? entities.Data(entityInstance)
                : entities.OpenData(entityInstance);
            ReadEntityData(entityData, exactInput->Data);
            switch (operation)
            {
                case Abi.EntityOperationEncodeNbt:
                {
                    var encoded = EntityNbtCodec.Encode(entities.EncodeNBT(entityInstance));
                    exactState->Nbt.Length = checked((ulong)encoded.Length);
                    if ((ulong)encoded.Length > exactState->Nbt.Capacity ||
                        encoded.Length != 0 && exactState->Nbt.Data is null) return Abi.Error;
                    encoded.CopyTo(new Span<byte>(exactState->Nbt.Data, encoded.Length));
                    break;
                }
                case Abi.EntityOperationOpen:
                {
                    if (exactInput->Invocation == 0 || exactInput->Handle.Value == 0 || exactInput->Handle.Generation == 0)
                        return Abi.Error;
                    var handle = entities.BindHandle(entityInstance, exactInput->Handle.Value, exactInput->Handle.Generation);
                    var opened = entities.Open(
                        entityInstance,
                        exactInput->Invocation,
                        handle);
                    exactState->Instance = opened.Open;
                    exactState->Capabilities = opened.Capabilities;
                    break;
                }
                case Abi.EntityOperationBBox:
                {
                    var box = entities.BBox(entityInstance, exactInput->Invocation);
                    var minimum = box.Min();
                    var maximum = box.Max();
                    exactState->BBox = new Native.BBox
                    {
                        Min = new Vec3 { X = minimum.X, Y = minimum.Y, Z = minimum.Z },
                        Max = new Vec3 { X = maximum.X, Y = maximum.Y, Z = maximum.Z },
                    };
                    break;
                }
                case Abi.EntityOperationClose:
                    entities.DispatchClose(entityInstance, exactInput->Invocation);
                    break;
                case Abi.EntityOperationHandle:
                    exactState->Handle = entities.DispatchHandle(entityInstance, exactInput->Invocation).Id;
                    break;
                case Abi.EntityOperationPosition:
                {
                    var position = entities.DispatchPosition(entityInstance, exactInput->Invocation);
                    exactState->Position = new Vec3 { X = position.X, Y = position.Y, Z = position.Z };
                    break;
                }
                case Abi.EntityOperationRotation:
                {
                    var rotation = entities.DispatchRotation(entityInstance, exactInput->Invocation);
                    exactState->Rotation = new NativeRotation { Yaw = rotation.Yaw, Pitch = rotation.Pitch };
                    break;
                }
                case Abi.EntityOperationTickExact:
                    entities.DispatchTick(entityInstance, exactInput->Invocation, exactInput->Current);
                    break;
                default:
                    return Abi.Error;
            }
            return WriteEntityData(entityData, exactInput->Data) ? Abi.Ok : Abi.Error;
        }
        catch { return Abi.Error; }
    }

    private static void ReadEntityData(World.EntityData output, EntityDataState* input)
    {
        output.Pos = new Vector3(input->Position.X, input->Position.Y, input->Position.Z);
        output.Vel = new Vector3(input->Velocity.X, input->Velocity.Y, input->Velocity.Z);
        output.Rot = new Rotation(input->Rotation.Yaw, input->Rotation.Pitch);
        output.Name = Utf8(input->Name);
        output.FireDuration = TimeSpan.FromTicks(input->FireDurationNanoseconds / 100);
        output.Age = TimeSpan.FromTicks(input->AgeNanoseconds / 100);
    }

    private static bool WriteEntityData(World.EntityData input, EntityDataState* output)
    {
        output->Position = new Vec3 { X = input.Pos.X, Y = input.Pos.Y, Z = input.Pos.Z };
        output->Velocity = new Vec3 { X = input.Vel.X, Y = input.Vel.Y, Z = input.Vel.Z };
        output->Rotation = new NativeRotation { Yaw = input.Rot.Yaw, Pitch = input.Rot.Pitch };
        output->FireDurationNanoseconds = checked(input.FireDuration.Ticks * 100);
        output->AgeNanoseconds = checked(input.Age.Ticks * 100);
        return WriteExact(&output->Name, input.Name ?? string.Empty);
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
    private static int HandleScheduled(void* instance, ulong callback, ulong invocation, uint phase, uint result)
    {
        try
        {
            if (instance is null || phase > Abi.WorldTaskComplete || result > Abi.WorldTaskFailed)
                return Abi.Error;
            if (phase == Abi.WorldTaskComplete)
            {
                if (!Scheduled.TryRemove(callback, out var completed)) return Abi.Ok;
                completed.Task.Complete(result);
                return Abi.Ok;
            }
            if (invocation == 0 || !Scheduled.TryGetValue(callback, out var scheduled)) return Abi.Error;
            try
            {
                using var invocationScope = InvocationContext.Enter(invocation);
                scheduled.Callback(new World.Tx(invocation));
                return Abi.Ok;
            }
            catch (Exception error)
            {
                scheduled.Task.CallbackFailed(error);
                return Abi.Error;
            }
        }
        catch { return Abi.Error; }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int Allow(void* instance, AllowInput* input, StringBuffer* message, byte* allowed)
    {
        try
        {
            if (instance is null || input is null || message is null || allowed is null) return Abi.Error;
            Net.Addr address;
            if (input->IsUdp != 0)
            {
                if (input->IP.Length is not (4 or 16) || input->IP.Data is null) return Abi.Error;
                address = new Net.UDPAddr(
                    new ReadOnlySpan<byte>(input->IP.Data, checked((int)input->IP.Length)).ToArray(),
                    input->Port,
                    Utf8(input->Zone));
            }
            else
            {
                address = new Net.AddrSnapshot(Utf8(input->Network), Utf8(input->Address));
            }
            var identity = JsonSerializer.Deserialize(
                new ReadOnlySpan<byte>(input->IdentityJson.Data, checked((int)input->IdentityJson.Length)),
                LoginJsonContext.Default.IdentityData);
            var client = JsonSerializer.Deserialize(
                new ReadOnlySpan<byte>(input->ClientJson.Data, checked((int)input->ClientJson.Length)),
                LoginJsonContext.Default.ClientData);
            if (identity is null || client is null) return Abi.Error;
            var result = Get(instance).Allow(address, identity, client);
            if (!WriteExact(message, result.Message ?? string.Empty)) return Abi.Error;
            *allowed = result.Allowed ? (byte)1 : (byte)0;
            return Abi.Ok;
        }
        catch { return Abi.Error; }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int HandleEvent(void* instance, uint eventId, void* input, void* state)
    {
        try
        {
            if (instance is null || input is null || state is null) return Abi.Error;
            using var invocationScope = InvocationContext.Enter(
                eventId >= Abi.PlayerMoveEvent && eventId <= Abi.WorldCloseEvent ? *(ulong*)input : 0);
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
                case Abi.WorldLiquidFlowEvent:
                {
                    var value = (WorldLiquidFlowInput*)input;
                    var result = (WorldCancellableState*)state;
                    var context = new World.Context(value->Invocation, result->Cancelled != 0);
                    var liquid = EventLiquid(value->Liquid);
                    plugin.HandleLiquidFlow(
                        context,
                        EventPos(value->From),
                        EventPos(value->Into),
                        liquid,
                        EventBlock(value->Replaced));
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.WorldLiquidDecayEvent:
                {
                    var value = (WorldLiquidDecayInput*)input;
                    var result = (WorldCancellableState*)state;
                    var context = new World.Context(value->Invocation, result->Cancelled != 0);
                    plugin.HandleLiquidDecay(
                        context,
                        EventPos(value->Position),
                        EventLiquid(value->Before),
                        value->After is null ? null : EventLiquid(*value->After));
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.WorldLiquidHardenEvent:
                {
                    var value = (WorldLiquidHardenInput*)input;
                    var result = (WorldCancellableState*)state;
                    var context = new World.Context(value->Invocation, result->Cancelled != 0);
                    plugin.HandleLiquidHarden(
                        context,
                        EventPos(value->Position),
                        EventBlock(value->LiquidHardened),
                        EventBlock(value->OtherLiquid),
                        EventBlock(value->NewBlock));
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.WorldSoundEvent:
                {
                    var value = (WorldSoundInput*)input;
                    var result = (WorldCancellableState*)state;
                    var context = new World.Context(value->Invocation, result->Cancelled != 0);
                    plugin.HandleSound(
                        context,
                        EventSound(value->Sound),
                        new Vector3(value->Position.X, value->Position.Y, value->Position.Z));
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.WorldFireSpreadEvent:
                {
                    var value = (WorldFireSpreadInput*)input;
                    var result = (WorldCancellableState*)state;
                    var context = new World.Context(value->Invocation, result->Cancelled != 0);
                    plugin.HandleFireSpread(context, EventPos(value->From), EventPos(value->To));
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.WorldBlockBurnEvent:
                case Abi.WorldCropTrampleEvent:
                case Abi.WorldLeavesDecayEvent:
                {
                    var value = (WorldPositionInput*)input;
                    var result = (WorldCancellableState*)state;
                    var context = new World.Context(value->Invocation, result->Cancelled != 0);
                    var position = EventPos(value->Position);
                    if (eventId == Abi.WorldBlockBurnEvent)
                        plugin.HandleBlockBurn(context, position);
                    else if (eventId == Abi.WorldCropTrampleEvent)
                        plugin.HandleCropTrample(context, position);
                    else
                        plugin.HandleLeavesDecay(context, position);
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.WorldEntitySpawnEvent:
                case Abi.WorldEntityDespawnEvent:
                {
                    var value = (WorldEntityInput*)input;
                    var entity = EventEntity(value->Entity, value->Invocation);
                    if (entity is null) return Abi.Error;
                    var tx = new World.Tx(value->Invocation);
                    if (eventId == Abi.WorldEntitySpawnEvent)
                        plugin.HandleEntitySpawn(tx, entity);
                    else
                        plugin.HandleEntityDespawn(tx, entity);
                    return Abi.Ok;
                }
                case Abi.WorldExplosionEvent:
                {
                    const ulong maxValues = 1UL << 20;
                    var value = (WorldExplosionInput*)input;
                    var result = (WorldExplosionState*)state;
                    if (result->SpawnFire > 1)
                        return Abi.Error;
                    var previousContext = result->ReplacementContext;
                    var previousDrop = result->ReplacementDrop;
                    var replacementFields = result->ReplacementEntities is not null ||
                        result->ReplacementEntityCount != 0 || result->ReplacementBlocks is not null ||
                        result->ReplacementBlockCount != 0 || result->ReplacementContext is not null;
                    if (previousDrop == null && replacementFields) return Abi.Error;
                    if (value->EntityCount > maxValues || value->BlockCount > maxValues ||
                        result->ReplacementEntityCount > maxValues || result->ReplacementBlockCount > maxValues)
                        return Abi.Error;
                    var entities = previousDrop == null
                        ? EventEntities(value->Entities, value->EntityCount, value->Invocation)
                        : EventEntities(result->ReplacementEntities, result->ReplacementEntityCount, value->Invocation);
                    var blocks = previousDrop == null
                        ? EventPositions(value->Blocks, value->BlockCount)
                        : EventPositions(result->ReplacementBlocks, result->ReplacementBlockCount);
                    var context = new World.Context(value->Invocation, result->Cancelled != 0);
                    var itemDropChance = result->ItemDropChance;
                    var spawnFire = result->SpawnFire != 0;
                    plugin.HandleExplosion(
                        context,
                        new Vector3(value->Position.X, value->Position.Y, value->Position.Z),
                        ref entities,
                        ref blocks,
                        ref itemDropChance,
                        ref spawnFire);
                    ArgumentNullException.ThrowIfNull(entities);
                    ArgumentNullException.ThrowIfNull(blocks);
                    var replacement = new PendingWorldExplosion(entities, blocks);
                    result->ReplacementEntities = replacement.Entities;
                    result->ReplacementEntityCount = replacement.EntityCount;
                    result->ReplacementBlocks = replacement.Blocks;
                    result->ReplacementBlockCount = replacement.BlockCount;
                    result->ReplacementContext = replacement.Context;
                    result->ReplacementDrop = replacement.Drop;
                    result->ItemDropChance = itemDropChance;
                    result->SpawnFire = spawnFire ? (byte)1 : (byte)0;
                    ApplyCancellation(context, &result->Cancelled);
                    if (previousDrop != null) previousDrop(previousContext);
                    return Abi.Ok;
                }
                case Abi.WorldRedstoneUpdateEvent:
                {
                    var value = (WorldRedstoneUpdateInput*)input;
                    var result = (WorldCancellableState*)state;
                    if (value->HasChangedNeighbour > 1 || value->ChangedRedstoneRelevant > 1 ||
                        value->HasSource > 1 || value->Cause > 2)
                        return Abi.Error;
                    var context = new World.Context(value->Invocation, result->Cancelled != 0);
                    plugin.HandleRedstoneUpdate(
                        context,
                        new World.RedstoneUpdate(
                            EventPos(value->Position),
                            EventPos(value->ChangedNeighbour),
                            value->HasChangedNeighbour != 0,
                            value->ChangedRedstoneRelevant != 0,
                            EventPos(value->Source),
                            value->HasSource != 0,
                            EventBlock(value->Before),
                            value->After is null ? null : EventBlock(*value->After),
                            value->OldPower,
                            value->NewPower,
                            value->CurrentTick,
                            (World.RedstoneUpdateCause)value->Cause));
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                case Abi.WorldCloseEvent:
                {
                    var value = (WorldCloseInput*)input;
                    plugin.HandleClose(new World.Tx(value->Invocation));
                    return Abi.Ok;
                }
                case Abi.PacketClientEvent:
                case Abi.PacketServerEvent:
                {
                    var value = (PacketInput*)input;
                    var result = (PacketState*)state;
                    var context = new Packet.Context(Utf8(value->Xuid), result->Cancelled != 0);
                    var packet = Packet.PacketCodec.Decode(value->Packet, value->PacketId);
                    if (eventId == Abi.PacketClientEvent)
                        plugin.HandleClientPacket(context, packet);
                    else
                        plugin.HandleServerPacket(context, packet);
                    ApplyCancellation(context, &result->Cancelled);
                    return Abi.Ok;
                }
                default:
                    return Abi.Ok;
            }
        }
        catch { return Abi.Error; }
    }

    private static Plugin Get(void* instance) => State(instance).Plugin;

    private static PluginState State(void* instance) =>
        (PluginState?)GCHandle.FromIntPtr((nint)instance).Target ??
        throw new InvalidOperationException("plugin instance is unavailable");

    private static StringView PersistentEntityTypeString(string value)
    {
        lock (EntityTypeStrings)
        {
            if (!EntityTypeStrings.TryGetValue(value, out var pointer))
            {
                var bytes = Encoding.UTF8.GetBytes(value);
                pointer = (nint)NativeMemory.Alloc((nuint)bytes.Length);
                bytes.CopyTo(new Span<byte>((void*)pointer, bytes.Length));
                EntityTypeStrings.Add(value, pointer);
            }
            return new StringView
            {
                Data = (byte*)pointer,
                Length = checked((ulong)Encoding.UTF8.GetByteCount(value)),
            };
        }
    }

    private sealed class PluginState : IDisposable
    {
        internal PluginState(Func<Plugin> plugin, Func<World.EntityType[]> entityTypes)
        {
            Plugin = plugin();
            Entities = new CustomEntityTable(entityTypes());
        }

        internal Plugin Plugin { get; }
        internal CustomEntityTable Entities { get; }

        public void Dispose() => Entities.Dispose();
    }

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

    private static World.Liquid EventLiquid(BlockView block)
    {
        if (block.Identifier.Length > 256 || block.PropertiesNbt.Length > 64 * 1024 ||
            block.Identifier.Length != 0 && block.Identifier.Data is null ||
            block.PropertiesNbt.Length != 0 && block.PropertiesNbt.Data is null)
            throw new InvalidOperationException("invalid liquid returned by server");
        var properties = block.PropertiesNbt.Length == 0
            ? Array.Empty<byte>()
            : new ReadOnlySpan<byte>(
                block.PropertiesNbt.Data,
                checked((int)block.PropertiesNbt.Length)).ToArray();
        return BlockCodec.DecodeLiquid(Utf8(block.Identifier), properties);
    }

    private static World.Sound EventSound(SoundViewV1 sound)
    {
        const uint goatHorn = 86;
        const uint attack = 68;
        const uint note = 78;
        const uint musicDiscPlay = 79;
        const uint bucketFill = 83;
        const uint bucketEmpty = 84;
        const uint crossbowLoad = 85;
        if (sound.Kind > goatHorn || !double.IsFinite(sound.Scalar) ||
            sound.Kind == attack && sound.Flags > 1 ||
            sound.Kind == note && sound.Data >= 16 ||
            sound.Kind == musicDiscPlay && sound.Data >= 21 ||
            (sound.Kind == bucketFill || sound.Kind == bucketEmpty) && sound.Data >= 2 ||
            sound.Kind == crossbowLoad && (sound.Integer is < 0 or > 2 || sound.Flags > 1) ||
            sound.Kind == goatHorn && sound.Data >= 8)
            throw new InvalidOperationException("invalid sound returned by server");
        var block = sound.Block is null
            ? null
            : sound.Kind is bucketFill or bucketEmpty
                ? EventLiquid(*sound.Block)
                : EventBlock(*sound.Block);
        return Sound.DecodeEvent(
            sound.Kind,
            sound.Data,
            sound.Integer,
            sound.Flags,
            sound.Scalar,
            block,
            sound.Item is null ? null : Host.EventItem(*sound.Item).Item());
    }

    private static World.Entity[] EventEntities(EntityId* values, ulong count, ulong invocation)
    {
        const ulong maxValues = 1UL << 20;
        if (count > maxValues || count != 0 && values is null)
            throw new InvalidOperationException("invalid entity array returned by server");
        var result = new World.Entity[checked((int)count)];
        for (var index = 0; index < result.Length; index++)
            result[index] = EventEntity(values[index], invocation) ??
                throw new InvalidOperationException("invalid entity returned by server");
        return result;
    }

    private static Cube.Pos[] EventPositions(BlockPos* values, ulong count)
    {
        const ulong maxValues = 1UL << 20;
        if (count > maxValues || count != 0 && values is null)
            throw new InvalidOperationException("invalid block-position array returned by server");
        var result = new Cube.Pos[checked((int)count)];
        for (var index = 0; index < result.Length; index++) result[index] = EventPos(values[index]);
        return result;
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
        var name = Utf8(source.Name);
        return source.Kind switch
        {
            0 => new OpaqueDamageSource(
                name,
                source.Flags,
                (source.Flags & Abi.DamageSourceReducedByArmour) != 0,
                (source.Flags & Abi.DamageSourceReducedByResistance) != 0,
                (source.Flags & Abi.DamageSourceFire) != 0,
                (source.Flags & Abi.DamageSourceIgnoresTotem) != 0),
            1 => new Entity.AttackDamageSource(EventEntity(source.Entity, invocation)),
            2 => new Block.DamageSource(EventBlock(*source.Block)),
            3 => new Entity.DrowningDamageSource(),
            4 => new Entity.ExplosionDamageSource(),
            5 => new Entity.FallDamageSource(),
            6 => new Block.FireDamageSource(),
            7 => new Entity.GlideDamageSource(),
            8 => new Effect.InstantDamageSource(),
            9 => new Block.LavaDamageSource(),
            10 => new Entity.LightningDamageSource(),
            11 => new Block.MagmaDamageSource(),
            12 => new Effect.PoisonDamageSource(source.Data != 0),
            13 => new Entity.ProjectileDamageSource(
                EventEntity(source.Entity, invocation),
                EventEntity(source.SecondaryEntity, invocation)),
            14 => new Player.StarvationDamageSource(),
            15 => new Entity.SuffocationDamageSource(),
            16 => new Enchantment.ThornsDamageSource(EventEntity(source.Entity, invocation)),
            17 => new Entity.VoidDamageSource(),
            18 => new Effect.WitherDamageSource(),
            _ => throw new InvalidOperationException("invalid damage source returned by server"),
        };
    }

    private static World.HealingSource EventHealingSource(HealingSourceView source)
    {
        if (source.Kind > Abi.HealingSourceRegeneration || source.Data > 1 ||
            source.Data != 0 && source.Kind != 1)
            throw new InvalidOperationException("invalid healing source returned by server");
        var name = Utf8(source.Name);
        return source.Kind switch
        {
            0 => new OpaqueHealingSource(name),
            1 => new Entity.FoodHealingSource(source.Data != 0),
            2 => new Effect.InstantHealingSource(),
            3 => new Effect.RegenerationHealingSource(),
            _ => throw new InvalidOperationException("invalid healing source returned by server"),
        };
    }

    private sealed record OpaqueDamageSource(
        string Name,
        uint Flags,
        bool Armour,
        bool Resistance,
        bool IsFire,
        bool IgnoresTotem) : Enchantment.AffectedDamageSource
    {
        public bool ReducedByArmour() => Armour;
        public bool ReducedByResistance() => Resistance;
        public bool Fire() => IsFire;
        public bool IgnoreTotem() => IgnoresTotem;
        public bool AffectedByEnchantment(Item.EnchantmentType e) =>
            object.Equals(e, Item.FireProtection)
                ? (Flags & Abi.DamageSourceFireProtection) != 0
                : object.Equals(e, Item.FeatherFalling)
                    ? (Flags & Abi.DamageSourceFeatherFalling) != 0
                    : object.Equals(e, Item.BlastProtection)
                        ? (Flags & Abi.DamageSourceBlastProtection) != 0
                        : object.Equals(e, Item.ProjectileProtection) &&
                          (Flags & Abi.DamageSourceProjectileProtection) != 0;
    }

    private sealed record OpaqueHealingSource(string Name) : World.HealingSource;

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
    private static void DropWorldExplosion(void* context)
    {
        if (context is null) return;
        var handle = GCHandle.FromIntPtr((nint)context);
        try { (handle.Target as PendingWorldExplosion)?.Dispose(); }
        finally { handle.Free(); }
    }

    private sealed class PendingWorldExplosion : IDisposable
    {
        private GCHandle _handle;

        internal PendingWorldExplosion(IReadOnlyList<World.Entity> entities, IReadOnlyList<Cube.Pos> blocks)
        {
            const int maxValues = 1 << 20;
            if (entities.Count > maxValues || blocks.Count > maxValues)
                throw new ArgumentException("an explosion replacement exceeds the maximum array length");
            try
            {
                Entities = entities.Count == 0
                    ? null
                    : (EntityId*)NativeMemory.Alloc((nuint)entities.Count, (nuint)sizeof(EntityId));
                Blocks = blocks.Count == 0
                    ? null
                    : (BlockPos*)NativeMemory.Alloc((nuint)blocks.Count, (nuint)sizeof(BlockPos));
                for (var index = 0; index < entities.Count; index++)
                {
                    ArgumentNullException.ThrowIfNull(entities[index]);
                    Entities[index] = World.EntityIdOf(entities[index]);
                }
                for (var index = 0; index < blocks.Count; index++)
                {
                    var position = blocks[index];
                    Blocks[index] = new BlockPos { X = position.X(), Y = position.Y(), Z = position.Z() };
                }
                EntityCount = (ulong)entities.Count;
                BlockCount = (ulong)blocks.Count;
                _handle = GCHandle.Alloc(this);
            }
            catch
            {
                Dispose();
                throw;
            }
        }

        internal EntityId* Entities { get; private set; }
        internal ulong EntityCount { get; }
        internal BlockPos* Blocks { get; private set; }
        internal ulong BlockCount { get; }
        internal void* Context => (void*)GCHandle.ToIntPtr(_handle);
        internal delegate* unmanaged[Cdecl]<void*, void> Drop => &DropWorldExplosion;

        public void Dispose()
        {
            if (Entities is not null) NativeMemory.Free(Entities);
            if (Blocks is not null) NativeMemory.Free(Blocks);
            Entities = null;
            Blocks = null;
        }
    }

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

    private static void ApplyCancellation(World.Context context, byte* cancelled)
    {
        if (context.Cancelled()) *cancelled = 1;
    }

    private static void ApplyCancellation(Packet.Context context, byte* cancelled)
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
