// Code generated from Dragonfly world.Config, Dimension, Provider, and mcdb.Config Go AST. DO NOT EDIT.
#nullable enable
using System;

namespace Dragonfly;

public sealed partial class World
{
    public abstract record Dimension
    {
        private protected Dimension() { }
    }

    internal sealed record BuiltinDimension(uint Id) : Dimension;
    public static Dimension Overworld { get; } = new BuiltinDimension(0);
    public static Dimension Nether { get; } = new BuiltinDimension(1);
    public static Dimension End { get; } = new BuiltinDimension(2);

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
