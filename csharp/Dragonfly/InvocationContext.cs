namespace Dragonfly;

internal static class InvocationContext
{
    [ThreadStatic]
    private static ulong _current;

    internal static ulong Resolve(ulong invocation) => _current == 0 ? invocation : _current;

    internal static Scope Enter(ulong invocation)
    {
        var previous = _current;
        _current = invocation;
        return new Scope(previous);
    }

    internal readonly struct Scope(ulong previous) : IDisposable
    {
        public void Dispose() => _current = previous;
    }
}
