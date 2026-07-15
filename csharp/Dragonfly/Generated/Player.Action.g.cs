// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.
#nullable enable
using Dragonfly.Native;

namespace Dragonfly;

public sealed partial class Player
{
    public void AbortBreaking() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionAbortBreaking, default);
    public void ClearInputLocks() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionClearInputLocks, default);
    public void FinishBreaking() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionFinishBreaking, default);
    public void Jump() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionJump, default);
    public void MoveItemsToInventory() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionMoveItemsToInventory, default);
    public void PunchAir() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionPunchAir, default);
    public void ReleaseItem() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionReleaseItem, default);
    public void RemoveAllDebugShapes() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionRemoveAllDebugShapes, default);
    public void SwingArm() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionSwingArm, default);
    public void UseItem() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionUseItem, default);
    public void Wake() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionWake, default);
}
