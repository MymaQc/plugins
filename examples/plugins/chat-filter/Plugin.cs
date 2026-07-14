using System;
using Dragonfly;

public sealed class ChatFilter : Plugin
{
    public override void HandleChat(Player.Context ctx, ref string message) =>
        message = message.Replace("badword", "***", StringComparison.OrdinalIgnoreCase);
}
