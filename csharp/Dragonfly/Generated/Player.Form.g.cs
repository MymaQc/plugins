// Code generated from Dragonfly server/player/player.go. DO NOT EDIT.
#nullable enable

namespace Dragonfly;

public sealed partial class Player : Form.Submitter
{
    public void SendForm(Form.Value f) =>
        PluginBridge.Host.SendPlayerForm(Invocation, Id, f);

    public void CloseForm() =>
        PluginBridge.Host.ClosePlayerForm(Invocation, Id);
}
