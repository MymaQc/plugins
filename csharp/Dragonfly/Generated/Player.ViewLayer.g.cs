// Code generated from Dragonfly player.go and visibility_level.go Go AST. DO NOT EDIT.
using Dragonfly.Native;

namespace Dragonfly;

public sealed partial class World
{
    public readonly record struct VisibilityLevel
    {
        internal VisibilityLevel(byte value) => Value = value;
        internal byte Value { get; }
    }

    public static VisibilityLevel PublicVisibility() => new(0);
    public static VisibilityLevel EnforceInvisible() => new(1);
    public static VisibilityLevel EnforceVisible() => new(2);
}

public sealed partial class Player
{
    public void ViewNameTag(World.Entity entity, string nameTag) =>
        PluginBridge.Host.RunPlayerViewLayer(_invocation, Id, entity, Abi.PlayerViewLayerViewNameTag, nameTag, default);
    public void ViewPublicNameTag(World.Entity entity) =>
        PluginBridge.Host.RunPlayerViewLayer(_invocation, Id, entity, Abi.PlayerViewLayerViewPublicNameTag, string.Empty, default);
    public void ViewScoreTag(World.Entity entity, string scoreTag) =>
        PluginBridge.Host.RunPlayerViewLayer(_invocation, Id, entity, Abi.PlayerViewLayerViewScoreTag, scoreTag, default);
    public void ViewPublicScoreTag(World.Entity entity) =>
        PluginBridge.Host.RunPlayerViewLayer(_invocation, Id, entity, Abi.PlayerViewLayerViewPublicScoreTag, string.Empty, default);
    public void ViewVisibility(World.Entity entity, World.VisibilityLevel level) =>
        PluginBridge.Host.RunPlayerViewLayer(_invocation, Id, entity, Abi.PlayerViewLayerViewVisibility, string.Empty, level);
    public void RemoveViewLayer(World.Entity entity) =>
        PluginBridge.Host.RunPlayerViewLayer(_invocation, Id, entity, Abi.PlayerViewLayerRemoveViewLayer, string.Empty, default);
}
