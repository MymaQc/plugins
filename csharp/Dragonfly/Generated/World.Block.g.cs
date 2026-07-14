// Code generated from Dragonfly server/world Go AST. DO NOT EDIT.
#nullable enable
using System;

namespace Dragonfly;

public sealed partial class World
{
    public interface Block { }

    public interface Liquid : Block { }

    public sealed class SetOpts
    {
        public bool DisableBlockUpdates;
        public bool DisableLiquidDisplacement;
        public bool DisableRedstoneUpdates;
    }

    public partial class Tx
    {
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
    }
}
