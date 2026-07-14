using System;
using Dragonfly;

public sealed class KitchenSink : Plugin
{
    private bool _enabled;

    public override void OnEnable() => _enabled = true;
    public override void OnDisable() => _enabled = false;

    public override void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot)
    {
        if (!_enabled || !Finite(newPos) || !double.IsFinite(newRot.Yaw) || !double.IsFinite(newRot.Pitch))
            ctx.Cancel();
    }

    public override void HandleJump(Player p) => _ = p;

    public override void HandleTeleport(Player.Context ctx, Vector3 pos)
    {
        if (!_enabled || !Finite(pos)) ctx.Cancel();
    }

    public override void HandleToggleSprint(Player.Context ctx, bool after) => _ = (ctx, after);
    public override void HandleToggleSneak(Player.Context ctx, bool after) => _ = (ctx, after);

    public override void HandleChat(Player.Context ctx, ref string message)
    {
        if (!_enabled) ctx.Cancel();
        message = message.Trim();
    }

    public override void HandleFoodLoss(Player.Context ctx, int from, ref int to)
    {
        _ = (ctx, from);
        to = Math.Max(0, to);
    }

    public override void HandlePunchAir(Player.Context ctx) => _ = ctx;
    public override void HandleQuit(Player p) => _ = p.Name;

    private static bool Finite(Vector3 value) =>
        double.IsFinite(value.X) && double.IsFinite(value.Y) && double.IsFinite(value.Z);
}
