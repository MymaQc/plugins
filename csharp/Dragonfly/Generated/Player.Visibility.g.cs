// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.
namespace Dragonfly;

public sealed partial class Player
{
    public void HideEntity(World.Entity e) => PluginBridge.Host.SetPlayerEntityVisible(_invocation, Id, e, false);
    public void ShowEntity(World.Entity e) => PluginBridge.Host.SetPlayerEntityVisible(_invocation, Id, e, true);
}
