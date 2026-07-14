// Code generated from Dragonfly server/player/handler.go. DO NOT EDIT.
namespace Dragonfly;

public sealed partial class Player
{
    public interface Handler
    {
        void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot);
        void HandleJump(Player p);
        void HandleTeleport(Player.Context ctx, Vector3 pos);
        void HandleToggleSprint(Player.Context ctx, bool after);
        void HandleToggleSneak(Player.Context ctx, bool after);
        void HandleChat(Player.Context ctx, ref string message);
        void HandleFoodLoss(Player.Context ctx, int from, ref int to);
        void HandlePunchAir(Player.Context ctx);
        void HandleQuit(Player p);
    }
}

public abstract partial class Plugin
{
    [HandlerSubscription(1UL)]
    public virtual void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot) { }
    [HandlerSubscription(16384UL)]
    public virtual void HandleJump(Player p) { }
    [HandlerSubscription(32768UL)]
    public virtual void HandleTeleport(Player.Context ctx, Vector3 pos) { }
    [HandlerSubscription(4096UL)]
    public virtual void HandleToggleSprint(Player.Context ctx, bool after) { }
    [HandlerSubscription(8192UL)]
    public virtual void HandleToggleSneak(Player.Context ctx, bool after) { }
    [HandlerSubscription(2UL)]
    public virtual void HandleChat(Player.Context ctx, ref string message) { }
    [HandlerSubscription(256UL)]
    public virtual void HandleFoodLoss(Player.Context ctx, int from, ref int to) { }
    [HandlerSubscription(131072UL)]
    public virtual void HandlePunchAir(Player.Context ctx) { }
    [HandlerSubscription(8UL)]
    public virtual void HandleQuit(Player p) { }
}
