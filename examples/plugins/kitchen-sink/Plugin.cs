using System;
using System.Threading;
using Dragonfly;

public sealed partial class KitchenSink : Plugin
{
    private long _jumps;
    private long _punches;
    private long _quits;
    private long _sneaks;
    private long _sprints;

    public override void OnEnable() => Console.WriteLine("kitchen-sink enabled");

    public override void OnDisable() => Console.WriteLine(
        $"kitchen-sink disabled: jumps={_jumps}, punches={_punches}, " +
        $"sprints={_sprints}, sneaks={_sneaks}, quits={_quits}");

    private static void Increment(ref long counter) => Interlocked.Increment(ref counter);
}
