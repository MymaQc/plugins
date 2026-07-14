// Code generated from Dragonfly server/world Go AST. DO NOT EDIT.
#nullable enable
using System;

namespace Dragonfly;

public sealed partial class World
{
    public interface Block { }

    public interface Biome { }

    public interface Particle { }

    public interface Liquid : Block { string LiquidType(); }

    public sealed class SetOpts
    {
        public bool DisableBlockUpdates;
        public bool DisableLiquidDisplacement;
        public bool DisableRedstoneUpdates;
    }

    public partial class Tx
    {
        public World World() =>
            PluginBridge.Host.TransactionWorld(Invocation);

        public Cube.Range Range() =>
            PluginBridge.Host.WorldRange(Invocation);

        public void SetBlock(Cube.Pos pos, Block? b, SetOpts? opts = null) =>
            PluginBridge.Host.SetWorldBlock(Invocation, pos, b, opts);

        public Block Block(Cube.Pos pos) =>
            PluginBridge.Host.WorldBlock(Invocation, pos);

        public (Block? Block, bool Ok) BlockLoaded(Cube.Pos pos) =>
            PluginBridge.Host.WorldBlockLoaded(Invocation, pos);

        public IEnumerable<Cube.Pos> BlocksWithin(Cube.Pos pos, int radius, params Block[] blocks) =>
            PluginBridge.Host.WorldBlocksWithin(Invocation, pos, radius, blocks);

        public (Liquid? Liquid, bool Ok) Liquid(Cube.Pos pos) =>
            PluginBridge.Host.WorldLiquid(Invocation, pos);

        public void SetLiquid(Cube.Pos pos, Liquid? b) =>
            PluginBridge.Host.SetWorldLiquid(Invocation, pos, b);

        public void ScheduleBlockUpdate(Cube.Pos pos, Block b, TimeSpan delay) =>
            PluginBridge.Host.ScheduleWorldBlockUpdate(Invocation, pos, b, delay);

        public int HighestLightBlocker(int x, int z) =>
            PluginBridge.Host.WorldHighestLightBlocker(Invocation, x, z);

        public int HighestBlock(int x, int z) =>
            PluginBridge.Host.WorldHighestBlock(Invocation, x, z);

        public byte Light(Cube.Pos pos) =>
            PluginBridge.Host.WorldLight(Invocation, pos);

        public byte SkyLight(Cube.Pos pos) =>
            PluginBridge.Host.WorldSkyLight(Invocation, pos);

        public void SetBiome(Cube.Pos pos, Biome b) =>
            PluginBridge.Host.SetWorldBiome(Invocation, pos, b);

        public Biome Biome(Cube.Pos pos) =>
            PluginBridge.Host.WorldBiome(Invocation, pos);

        public double Temperature(Cube.Pos pos) =>
            PluginBridge.Host.WorldTemperature(Invocation, pos);

        public bool RainingAt(Cube.Pos pos) =>
            PluginBridge.Host.WorldRainingAt(Invocation, pos);

        public bool SnowingAt(Cube.Pos pos) =>
            PluginBridge.Host.WorldSnowingAt(Invocation, pos);

        public bool ThunderingAt(Cube.Pos pos) =>
            PluginBridge.Host.WorldThunderingAt(Invocation, pos);

        public bool Raining() =>
            PluginBridge.Host.WorldRaining(Invocation);

        public bool Thundering() =>
            PluginBridge.Host.WorldThundering(Invocation);

        public long CurrentTick() =>
            PluginBridge.Host.WorldCurrentTick(Invocation);

        public void AddParticle(Vector3 pos, Particle p) =>
            PluginBridge.Host.AddWorldParticle(Invocation, pos, p);

        public Entity AddEntity(EntityHandle e) =>
            PluginBridge.Host.TransactionAddEntity(Invocation, e);

        public Entity AddEntityAt(EntityHandle e, Vector3 pos) =>
            PluginBridge.Host.TransactionAddEntity(Invocation, e, pos);

        public EntityHandle RemoveEntity(Entity e) =>
            PluginBridge.Host.TransactionRemoveEntity(Invocation, e);

        public IEnumerable<Entity> Entities() =>
            PluginBridge.Host.TransactionEntities(Invocation, playersOnly: false);

        public IEnumerable<Entity> Players() =>
            PluginBridge.Host.TransactionEntities(Invocation, playersOnly: true);
    }
}
