using System;
using Dragonfly;

public sealed class LifecycleLogger : Plugin
{
    public override void OnEnable() => Console.WriteLine("enabled");
    public override void OnDisable() => Console.WriteLine("disabled");
    public override void HandleQuit(Player player) => Console.WriteLine($"{player.Name} quit");
}
