// Code generated from Dragonfly server/player/handler.go. DO NOT EDIT.
namespace Dragonfly;

public sealed partial class Player
{
    public interface Handler
    {
        void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot);
        void HandleQuit(Player p);
    }
}

public abstract partial class Plugin
{
    public virtual void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot) { }
    public virtual void HandleQuit(Player p) { }
}
