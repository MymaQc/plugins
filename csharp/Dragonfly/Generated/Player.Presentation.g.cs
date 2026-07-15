// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.
#nullable enable
using Dragonfly.Native;

namespace Dragonfly;

public sealed partial class Player
{
    public void EnableInstantRespawn() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionEnableInstantRespawn, default);
    public void DisableInstantRespawn() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionDisableInstantRespawn, default);
    public void ShowCoordinates() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionShowCoordinates, default);
    public void HideCoordinates() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionHideCoordinates, default);
    public void SendSleepingIndicator(int sleeping, int max) => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionSendSleepingIndicator, new PlayerStateValue { Integer = sleeping, Number = max });
    public void CloseDialogue() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionCloseDialogue, default);
    public void RemoveBossBar() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionRemoveBossBar, default);
    public void RemoveScoreboard() => PluginBridge.Host.RemovePlayerScoreboard(_invocation, Id);
    public string NameTag() => PluginBridge.Host.PlayerString(_invocation, Id, Abi.PlayerStringNameTag);
    public string ScoreTag() => PluginBridge.Host.PlayerString(_invocation, Id, Abi.PlayerStringScoreTag);
    public void SendToast(string title, string message) => PluginBridge.Host.SendPlayerToast(_invocation, Id, title, message);
}
