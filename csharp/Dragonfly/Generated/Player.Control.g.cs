// Code generated from Dragonfly player/hud, player/input, and player.go Go AST. DO NOT EDIT.
#nullable enable
using Dragonfly.Native;

namespace Dragonfly;

public static class Hud
{
    public readonly record struct Element
    {
        internal Element(byte value) => Value = value;
        internal byte Value { get; }
    }

    public static Element PaperDoll() => new(0);
    public static Element Armour() => new(1);
    public static Element ToolTips() => new(2);
    public static Element TouchControls() => new(3);
    public static Element Crosshair() => new(4);
    public static Element HotBar() => new(5);
    public static Element Health() => new(6);
    public static Element ProgressBar() => new(7);
    public static Element Hunger() => new(8);
    public static Element AirBubbles() => new(9);
    public static Element HorseHealth() => new(10);
    public static Element StatusEffects() => new(11);
    public static Element ItemText() => new(12);
    public static Element[] All() => [PaperDoll(), Armour(), ToolTips(), TouchControls(), Crosshair(), HotBar(), Health(), ProgressBar(), Hunger(), AirBubbles(), HorseHealth(), StatusEffects(), ItemText()];
}

public static class Input
{
    public readonly record struct Lock
    {
        internal Lock(uint value) => Value = value;
        internal uint Value { get; }
    }

    public static Lock Camera() => new(2);
    public static Lock Movement() => new(4);
    public static Lock LateralMovement() => new(16);
    public static Lock Sneak() => new(32);
    public static Lock Jump() => new(64);
    public static Lock Mount() => new(128);
    public static Lock Dismount() => new(256);
    public static Lock MoveForward() => new(512);
    public static Lock MoveBackward() => new(1024);
    public static Lock MoveLeft() => new(2048);
    public static Lock MoveRight() => new(4096);
    public static Lock[] All() => [Camera(), Movement(), LateralMovement(), Sneak(), Jump(), Mount(), Dismount(), MoveForward(), MoveBackward(), MoveLeft(), MoveRight()];
}

public sealed partial class Player
{
    public void ShowHudElement(Hud.Element e) => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionShowHudElement, new PlayerStateValue { Integer = e.Value });
    public void HideHudElement(Hud.Element e) => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionHideHudElement, new PlayerStateValue { Integer = e.Value });
    public bool HudElementHidden(Hud.Element e) => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionHudElementHidden, new PlayerStateValue { Integer = e.Value }).Integer != 0;
    public void LockInput(Input.Lock l) => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionLockInput, new PlayerStateValue { Integer = l.Value });
    public void UnlockInput(Input.Lock l) => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionUnlockInput, new PlayerStateValue { Integer = l.Value });
    public bool InputLocked(Input.Lock l) => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionInputLocked, new PlayerStateValue { Integer = l.Value }).Integer != 0;
}
