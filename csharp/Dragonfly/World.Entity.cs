using Dragonfly.Native;

namespace Dragonfly;

public sealed partial class World
{
    public sealed partial class EntityHandle : IEquatable<EntityHandle>
    {
        internal EntityHandleId Id { get; }

        internal EntityHandle(EntityHandleId id) => Id = id;

        public bool Equals(EntityHandle? other) => other is not null && Id.Value == other.Id.Value && Id.Generation == other.Id.Generation;
        public override bool Equals(object? obj) => obj is EntityHandle other && Equals(other);
        public override int GetHashCode() => HashCode.Combine(Id.Value, Id.Generation);
    }

    internal sealed class HostEntity : Entity, IEquatable<HostEntity>
    {
        private readonly ulong _invocation;
        internal EntityId Id { get; }

        internal HostEntity(ulong invocation, EntityId id) => (_invocation, Id) = (invocation, id);

        public EntityHandle H() => PluginBridge.Host.EntityHandle(_invocation, Id);
        public Vector3 Position() => PluginBridge.Host.ReadEntityState(_invocation, Id).Position;
        public Rotation Rotation() => PluginBridge.Host.ReadEntityState(_invocation, Id).Rotation;
        public void Close() => PluginBridge.Host.CloseEntity(_invocation, Id);

        public bool Equals(HostEntity? other) => other is not null && SameId(Id, other.Id);
        public override bool Equals(object? obj) => obj is HostEntity other && Equals(other);
        public override int GetHashCode() => EntityHashCode(Id);
    }

    internal static Entity HostEntityFrom(ulong invocation, EntityId id) => new HostEntity(invocation, id);

    internal static EntityId EntityIdOf(Entity entity)
    {
        if (TryEntityIdOf(entity, out var id)) return id;
        throw new ArgumentException("entity is not owned by Dragonfly", nameof(entity));
    }

    internal static bool TryEntityIdOf(Entity entity, out EntityId id)
    {
        switch (entity)
        {
            case HostEntity host:
                id = host.Id;
                return true;
            case Player player:
                id = player.NativeEntityId();
                return true;
            default:
                id = default;
                return false;
        }
    }

    internal static unsafe bool SameId(EntityId left, EntityId right)
    {
        if (left.Generation != right.Generation) return false;
        for (var index = 0; index < 16; index++)
            if (left.Bytes[index] != right.Bytes[index]) return false;
        return true;
    }

    private static unsafe int EntityHashCode(EntityId value)
    {
        var hash = new HashCode();
        hash.Add(value.Generation);
        for (var index = 0; index < 16; index++) hash.Add(value.Bytes[index]);
        return hash.ToHashCode();
    }
}
