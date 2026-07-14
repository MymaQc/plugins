// Code generated from Dragonfly server/world Go AST. DO NOT EDIT.
#nullable enable

namespace Dragonfly;

public sealed partial class World
{
    public interface Block { }

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
    }
}
