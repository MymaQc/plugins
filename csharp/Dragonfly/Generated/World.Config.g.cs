// Code generated from Dragonfly world.Config, Dimension, Provider, and mcdb.Config Go AST. DO NOT EDIT.
#nullable enable
using System;

namespace Dragonfly;

public sealed partial class World
{
    public interface Dimension
    {
        Cube.Range Range();
        bool WaterEvaporates();
        TimeSpan LavaSpreadDuration();
        bool WeatherCycle();
        bool TimeCycle();
    }

    internal sealed record BuiltinDimension(uint Id, Cube.Range BuildRange, bool EvaporatesWater, TimeSpan LavaDuration, bool HasWeatherCycle, bool HasTimeCycle) : Dimension
    {
        public Cube.Range Range() => BuildRange;
        public bool WaterEvaporates() => EvaporatesWater;
        public TimeSpan LavaSpreadDuration() => LavaDuration;
        public bool WeatherCycle() => HasWeatherCycle;
        public bool TimeCycle() => HasTimeCycle;
    }

    internal sealed record TransportDimension(Cube.Range BuildRange, bool EvaporatesWater, TimeSpan LavaDuration, bool HasWeatherCycle, bool HasTimeCycle) : Dimension
    {
        public Cube.Range Range() => BuildRange;
        public bool WaterEvaporates() => EvaporatesWater;
        public TimeSpan LavaSpreadDuration() => LavaDuration;
        public bool WeatherCycle() => HasWeatherCycle;
        public bool TimeCycle() => HasTimeCycle;
    }

    public static Dimension Overworld { get; } = new BuiltinDimension(0, new Cube.Range(-64, 319), false, TimeSpan.FromMilliseconds(1500), true, true);
    public static Dimension Nether { get; } = new BuiltinDimension(1, new Cube.Range(0, 127), true, TimeSpan.FromMilliseconds(250), false, false);
    public static Dimension End { get; } = new BuiltinDimension(2, new Cube.Range(0, 255), false, TimeSpan.FromMilliseconds(1500), false, false);

    public abstract record Provider
    {
        private protected Provider() { }
    }

    public sealed record NopProvider : Provider;

    public sealed record Config
    {
        public Dimension? Dim { get; init; }
        public Provider? Provider { get; init; }
        public bool ReadOnly { get; init; }
        public TimeSpan SaveInterval { get; init; }
        public TimeSpan ChunkUnloadInterval { get; init; }
        public int RandomTickSpeed { get; init; }

        public World New() => PluginBridge.Host.NewWorld(this);
    }

    public static World New() => new Config().New();
}

public static class MCDB
{
    public sealed record Config
    {
        public DB Open(string dir) => new(dir);
    }

    public sealed record DB : World.Provider
    {
        internal DB(string dir) => Directory = dir;
        internal string Directory { get; }
    }
}
