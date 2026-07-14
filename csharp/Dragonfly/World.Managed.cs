namespace Dragonfly;

public sealed partial class World
{
    // Managed worlds are persistent MCDB worlds owned by the plugin host. These
    // entry points are host extensions: regular World instance methods below
    // retain Dragonfly's API shape.
    public enum ManagedDimension
    {
        Overworld,
        Nether,
        End,
    }

    public enum ManagedOpenMode
    {
        OpenOrCreate,
        OpenExisting,
        CreateNew,
    }

    public enum ManagedSavePolicy
    {
        Automatic,
        Manual,
    }

    public enum ManagedRandomTickPolicy
    {
        Disabled,
        PerSubchunk,
    }

    public enum ManagedTimePolicy
    {
        Preserve,
        Cycle,
        Fixed,
    }

    public enum ManagedWeatherPolicy
    {
        Preserve,
        Cycle,
        Clear,
    }

    public sealed record ManagedOpenSpec(string ProviderPath)
    {
        public ManagedDimension Dimension { get; init; } = ManagedDimension.Overworld;
        public ManagedOpenMode OpenMode { get; init; } = ManagedOpenMode.OpenOrCreate;
        public bool ReadOnly { get; init; }
        public ManagedSavePolicy Save { get; init; } = ManagedSavePolicy.Automatic;
        public TimeSpan SaveInterval { get; init; } = TimeSpan.FromMinutes(10);
        public ManagedRandomTickPolicy RandomTicks { get; init; } = ManagedRandomTickPolicy.PerSubchunk;
        public uint RandomTickRate { get; init; } = 3;
        public ManagedTimePolicy Time { get; init; } = ManagedTimePolicy.Preserve;
        public long FixedTime { get; init; }
        public ManagedWeatherPolicy Weather { get; init; } = ManagedWeatherPolicy.Preserve;
        public TimeSpan ChunkUnloadAfter { get; init; } = TimeSpan.FromMinutes(2);
    }

    public static World? LookupManaged(string name) =>
        PluginBridge.Host.LookupManagedWorld(name);

    public static World? OpenManaged(
        string name,
        ManagedDimension dimension = ManagedDimension.Overworld) =>
        PluginBridge.Host.OpenManagedWorld(name, dimension);

    public static World? OpenManaged(string name, ManagedOpenSpec spec) =>
        PluginBridge.Host.OpenManagedWorld(name, spec);
}
