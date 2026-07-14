namespace Dragonfly;

public sealed partial class Player
{
    // ChangeWorld is a host extension. Dragonfly's Transfer method transfers a
    // player to another server, so that name is intentionally not reused here.
    public void ChangeWorld(World world, Vector3 position)
    {
        ArgumentNullException.ThrowIfNull(world);
        PluginBridge.Host.ChangePlayerWorld(_invocation, Id, world.Id, position);
    }
}
