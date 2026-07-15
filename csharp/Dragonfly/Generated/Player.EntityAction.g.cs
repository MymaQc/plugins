// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.
using Dragonfly.Native;

namespace Dragonfly;

public sealed partial class Player
{
    public bool UseItemOnEntity(World.Entity e) =>
        PluginBridge.Host.RunPlayerEntityAction(_invocation, Id, e, Abi.PlayerEntityActionUseItemOnEntity);
    public bool AttackEntity(World.Entity e) =>
        PluginBridge.Host.RunPlayerEntityAction(_invocation, Id, e, Abi.PlayerEntityActionAttackEntity);
}
