// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.
using Dragonfly.Native;

namespace Dragonfly;

public sealed partial class Player
{
    public void BreakBlock(Cube.Pos pos) =>
        PluginBridge.Host.RunPlayerBlockAction(_invocation, Id, Abi.PlayerBlockActionBreakBlock, pos, default, default);
    public void ContinueBreaking(Cube.Face face) =>
        PluginBridge.Host.RunPlayerBlockAction(_invocation, Id, Abi.PlayerBlockActionContinueBreaking, default, face, default);
    public void PickBlock(Cube.Pos pos) =>
        PluginBridge.Host.RunPlayerBlockAction(_invocation, Id, Abi.PlayerBlockActionPickBlock, pos, default, default);
    public void Sleep(Cube.Pos pos) =>
        PluginBridge.Host.RunPlayerBlockAction(_invocation, Id, Abi.PlayerBlockActionSleep, pos, default, default);
    public void StartBreaking(Cube.Pos pos, Cube.Face face) =>
        PluginBridge.Host.RunPlayerBlockAction(_invocation, Id, Abi.PlayerBlockActionStartBreaking, pos, face, default);
    public void UseItemOnBlock(Cube.Pos pos, Cube.Face face, Vector3 clickPos) =>
        PluginBridge.Host.RunPlayerBlockAction(_invocation, Id, Abi.PlayerBlockActionUseItemOnBlock, pos, face, clickPos);
}
