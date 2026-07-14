// Code generated from Dragonfly server/world/world.go Go AST. DO NOT EDIT.
namespace Dragonfly;

public sealed partial class World
{
    public string Name() => PluginBridge.Host.ManagedWorldName(_invocation, Id) ?? string.Empty;
    public Cube.Pos Spawn() => PluginBridge.Host.ManagedWorldSpawn(_invocation, Id);
    public void SetSpawn(Cube.Pos pos) =>
        PluginBridge.Host.SetManagedWorldSpawn(_invocation, Id, pos);
    public void Save() => PluginBridge.Host.SaveManagedWorld(_invocation, Id);
    public void Close() => PluginBridge.Host.CloseManagedWorld(_invocation, Id);
}
