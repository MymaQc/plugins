// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.
#nullable enable
using Dragonfly.Native;

namespace Dragonfly;

public sealed partial class Player
{
    public int Food() => checked((int)PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateFood).Integer);
    public void SetFood(int level) => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateFood, new PlayerStateValue { Integer = level });
    public double Health() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateHealth).Number;
    public double MaxHealth() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateMaxHealth).Number;
    public void SetMaxHealth(double health) => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateMaxHealth, new PlayerStateValue { Number = health });
    public double Heal(double health, World.HealingSource source) => PluginBridge.Host.HealPlayer(_invocation, Id, health, source);
    public (double Damage, bool Vulnerable) Hurt(double dmg, World.DamageSource src) => PluginBridge.Host.HurtPlayer(_invocation, Id, dmg, src);
    public int ExperienceLevel() => checked((int)PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateExperienceLevel).Integer);
    public void SetExperienceLevel(int level)
    {
        if (level < 0) throw new ArgumentOutOfRangeException(nameof(level));
        PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateExperienceLevel, new PlayerStateValue { Integer = level });
    }
    public double ExperienceProgress() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateExperienceProgress).Number;
    public void SetExperienceProgress(double progress)
    {
        if (progress is < 0 or > 1)
            throw new ArgumentOutOfRangeException(nameof(progress));
        PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateExperienceProgress, new PlayerStateValue { Number = progress });
    }
    public double Scale() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateScale).Number;
    public void SetScale(double s) => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateScale, new PlayerStateValue { Number = s });
    public bool Invisible() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateInvisible).Integer != 0;
    public void SetInvisible() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateInvisible, new PlayerStateValue { Integer = 1 });
    public void SetVisible() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateInvisible, default);
    public bool Immobile() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateImmobile).Integer != 0;
    public void SetImmobile() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateImmobile, new PlayerStateValue { Integer = 1 });
    public void SetMobile() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateImmobile, default);
    public double Speed() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateSpeed).Number;
    public void SetSpeed(double speed) => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateSpeed, new PlayerStateValue { Number = speed });
    public double FlightSpeed() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateFlightSpeed).Number;
    public void SetFlightSpeed(double flightSpeed) => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateFlightSpeed, new PlayerStateValue { Number = flightSpeed });
    public double VerticalFlightSpeed() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateVerticalFlightSpeed).Number;
    public void SetVerticalFlightSpeed(double flightSpeed) => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateVerticalFlightSpeed, new PlayerStateValue { Number = flightSpeed });
    public void ResetFallDistance() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateFallDistance, default);
    public double FallDistance() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateFallDistance).Number;
    public void SetAbsorption(double health) => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateAbsorption, new PlayerStateValue { Number = health });
    public double Absorption() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateAbsorption).Number;
    public bool Dead() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateDead).Integer != 0;
    public bool OnGround() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateOnGround).Integer != 0;
    public double EyeHeight() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateEyeHeight).Number;
    public double TorsoHeight() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateTorsoHeight).Number;
    public bool Breathing() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateBreathing).Integer != 0;
}
