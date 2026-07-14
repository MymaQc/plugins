// Code generated from Dragonfly server/world/world.go Go AST. DO NOT EDIT.
namespace Dragonfly;

public sealed partial class World
{
    public string Name() => PluginBridge.Host.WorldName(_invocation, Id) ?? string.Empty;
    public Cube.Pos Spawn() => PluginBridge.Host.WorldSpawn(_invocation, Id);
    public void SetSpawn(Cube.Pos pos) =>
        PluginBridge.Host.SetWorldSpawn(_invocation, Id, pos);
    public void Save() => PluginBridge.Host.SaveWorld(_invocation, Id);
    public void Close() => PluginBridge.Host.CloseWorld(_invocation, Id);
}
