using System.ComponentModel;

namespace Dragonfly;

[AttributeUsage(AttributeTargets.Method)]
[EditorBrowsable(EditorBrowsableState.Never)]
public sealed class HandlerSubscriptionAttribute(ulong value) : Attribute
{
    public ulong Value { get; } = value;
}
