// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.
#nullable enable
using System;
using Dragonfly.Native;

namespace Dragonfly;

public sealed partial class Player
{
    public void SetGameMode(World.GameMode mode)
    {
        ArgumentNullException.ThrowIfNull(mode);
        PluginBridge.Host.SetPlayerState(
            _invocation,
            Id,
            Abi.PlayerStateGameMode,
            new PlayerStateValue { Integer = World.GameModeDescriptor(mode) });
    }

    public World.GameMode GameMode() => PluginBridge.Host.PlayerGameMode(_invocation, Id);
}
