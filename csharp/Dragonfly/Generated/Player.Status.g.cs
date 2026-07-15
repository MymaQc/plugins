// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.
namespace Dragonfly;

public sealed partial class Player
{
    public bool UsingItem() => PluginBridge.Host.PlayerUsingItem(_invocation, Id);
    public (Cube.Pos Position, bool Sleeping) Sleeping() => PluginBridge.Host.PlayerSleeping(_invocation, Id);
}
