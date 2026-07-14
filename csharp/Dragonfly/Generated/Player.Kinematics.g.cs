// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.
using Dragonfly.Native;

namespace Dragonfly;

public sealed partial class Player
{
    public void Teleport(Vector3 pos) =>
        PluginBridge.Host.TransformPlayer(_invocation, Id, Abi.PlayerTransformTeleport, pos, 0, 0);
    public void Move(Vector3 deltaPos, double deltaYaw, double deltaPitch) =>
        PluginBridge.Host.TransformPlayer(_invocation, Id, Abi.PlayerTransformMove, deltaPos, deltaYaw, deltaPitch);
    public void Displace(Vector3 deltaPos) =>
        PluginBridge.Host.TransformPlayer(_invocation, Id, Abi.PlayerTransformDisplace, deltaPos, 0, 0);
    public Vector3 Position() => PluginBridge.Host.TryReadPlayerKinematics(_invocation, Id, out var state) ? state.Position : _position;
    public Vector3 Velocity() => PluginBridge.Host.ReadPlayerKinematics(_invocation, Id).Velocity;
    public void SetVelocity(Vector3 velocity) =>
        PluginBridge.Host.TransformPlayer(_invocation, Id, Abi.PlayerTransformVelocity, velocity, 0, 0);
    public Rotation Rotation() => PluginBridge.Host.ReadPlayerKinematics(_invocation, Id).Rotation;
}
