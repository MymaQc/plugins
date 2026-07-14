namespace Dragonfly;

public sealed partial class World
{
    public interface Item { }

    public partial class Tx
    {
        private readonly ulong _invocation;
        private readonly InvocationLease? _lease;

        internal Tx(ulong invocation) => _invocation = invocation;
        internal Tx(InvocationLease lease) => _lease = lease;
        internal ulong Invocation => _lease?.Invocation ?? _invocation;
    }

    public class Context : Tx
    {
        private bool _cancelled;

        internal Context(ulong invocation, bool cancelled) : base(invocation) =>
            _cancelled = cancelled;

        public bool Cancelled() => _cancelled;
        public void Cancel() => _cancelled = true;
    }
}

internal sealed class InvocationLease
{
    private ulong _invocation;

    internal ulong Invocation => _invocation;

    internal Scope Enter(ulong invocation)
    {
        if (invocation == 0) throw new InvalidOperationException("world transaction is unavailable");
        var previous = _invocation;
        _invocation = invocation;
        return new Scope(this, previous);
    }

    internal readonly struct Scope : IDisposable
    {
        private readonly InvocationLease _lease;
        private readonly ulong _previous;

        internal Scope(InvocationLease lease, ulong previous)
        {
            _lease = lease;
            _previous = previous;
        }

        public void Dispose() => _lease._invocation = _previous;
    }
}
