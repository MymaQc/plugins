using Dragonfly.Native;

namespace Dragonfly;

public sealed partial class World : IEquatable<World>
{
    private readonly ulong _invocation;
    internal WorldId Id { get; }
    internal ulong Invocation => InvocationContext.Resolve(_invocation);

    internal World(ulong invocation, WorldId id)
    {
        _invocation = invocation;
        Id = id;
    }

    public bool Equals(World? other) => other is not null && Id.Value == other.Id.Value;
    public override bool Equals(object? obj) => obj is World other && Equals(other);
    public override int GetHashCode() => Id.Value.GetHashCode();
}
