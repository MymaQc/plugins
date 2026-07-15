// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.
#nullable enable
namespace Dragonfly;

public sealed partial class Player
{
    public bool UsingItem() => PluginBridge.Host.PlayerUsingItem(_invocation, Id);
    public (Cube.Pos Position, bool Sleeping) Sleeping() => PluginBridge.Host.PlayerSleeping(_invocation, Id);
    public (Vector3 Position, World.Dimension? Dimension, bool Found) DeathPosition() => PluginBridge.Host.PlayerDeathPosition(_invocation, Id);
}
