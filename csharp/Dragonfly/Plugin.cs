namespace Dragonfly;

public abstract partial class Plugin : Player.Handler
{
    public virtual void OnEnable() { }
    public virtual void OnDisable() { }

    [HandlerSubscription(4UL)]
    public virtual void OnJoin(Player.Context ctx) { }
}
