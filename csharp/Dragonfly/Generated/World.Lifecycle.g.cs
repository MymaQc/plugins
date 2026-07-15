// Code generated from Dragonfly server/world/world.go Go AST. DO NOT EDIT.
namespace Dragonfly;

public sealed partial class World
{
    public string Name() => PluginBridge.Host.WorldName(Invocation, Id) ?? string.Empty;
    public Cube.Range Range() => PluginBridge.Host.WorldRange(Invocation, Id);
    public int HighestLightBlocker(int x, int z) =>
        PluginBridge.Host.WorldHighestLightBlocker(Invocation, Id, x, z);
    public int Time() => PluginBridge.Host.WorldTime(Invocation, Id);
    public void SetTime(int @new) => PluginBridge.Host.SetWorldTime(Invocation, Id, @new);
    public void StopTime() => PluginBridge.Host.SetWorldTimeCycle(Invocation, Id, false);
    public void StartTime() => PluginBridge.Host.SetWorldTimeCycle(Invocation, Id, true);
    public bool TimeCycle() => PluginBridge.Host.WorldTimeCycle(Invocation, Id);
    public Cube.Pos Spawn() => PluginBridge.Host.WorldSpawn(Invocation, Id);
    public void SetSpawn(Cube.Pos pos) =>
        PluginBridge.Host.SetWorldSpawn(Invocation, Id, pos);
    public Cube.Pos PlayerSpawn(Guid id) =>
        PluginBridge.Host.WorldPlayerSpawn(Invocation, Id, id);
    public void SetPlayerSpawn(Guid id, Cube.Pos pos) =>
        PluginBridge.Host.SetWorldPlayerSpawn(Invocation, Id, id, pos);
    public void SetRequiredSleepDuration(TimeSpan duration) =>
        PluginBridge.Host.SetWorldRequiredSleepDuration(Invocation, Id, duration);
    public GameMode DefaultGameMode() => PluginBridge.Host.WorldDefaultGameMode(Invocation, Id);
    public void SetTickRange(int v) => PluginBridge.Host.SetWorldTickRange(Invocation, Id, v);
    public void SetDefaultGameMode(GameMode mode) =>
        PluginBridge.Host.SetWorldDefaultGameMode(Invocation, Id, mode);
    public void SetDifficulty(Difficulty d) =>
        PluginBridge.Host.SetWorldDifficulty(Invocation, Id, d);
    public void Save() => PluginBridge.Host.SaveWorld(Invocation, Id);
    public void Close() => PluginBridge.Host.CloseWorld(Invocation, Id);
}

public static class WorldStateExtensions
{
    public static World.Dimension Dimension(this World world)
    {
        ArgumentNullException.ThrowIfNull(world);
        return PluginBridge.Host.WorldDimension(world.Invocation, world.Id);
    }

    public static World.Difficulty Difficulty(this World world)
    {
        ArgumentNullException.ThrowIfNull(world);
        return PluginBridge.Host.WorldDifficulty(world.Invocation, world.Id);
    }
}
