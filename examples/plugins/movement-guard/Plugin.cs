using Dragonfly;

public sealed class MovementGuard : Plugin
{
    public override void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot)
    {
        if (newPos.Y < -64) ctx.Cancel();
    }
}
