// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.
namespace Dragonfly;

public sealed partial class Player
{
    public double FinalDamageFrom(double dmg, World.DamageSource src) =>
        PluginBridge.Host.FinalPlayerDamage(_invocation, Id, dmg, src);
}
