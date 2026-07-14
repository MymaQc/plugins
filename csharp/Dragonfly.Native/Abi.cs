using System.Runtime.InteropServices;

namespace Dragonfly.Native;

public static class Abi
{
    public const uint PluginVersion = 4;
    public const uint HostVersion = 20;
    public const int Ok = 0;
    public const int Error = 1;
    public const uint PlayerMoveEvent = 1;
    public const ulong PlayerMoveSubscription = 1UL;
    public const uint PlayerChatEvent = 2;
    public const ulong PlayerChatSubscription = 1UL << 1;
    public const uint PlayerQuitEvent = 4;
    public const ulong PlayerQuitSubscription = 1UL << 3;
    public const uint PlayerFoodLossEvent = 9;
    public const ulong PlayerFoodLossSubscription = 1UL << 8;
    public const uint PlayerToggleSprintEvent = 13;
    public const ulong PlayerToggleSprintSubscription = 1UL << 12;
    public const uint PlayerToggleSneakEvent = 14;
    public const ulong PlayerToggleSneakSubscription = 1UL << 13;
    public const uint PlayerJumpEvent = 15;
    public const ulong PlayerJumpSubscription = 1UL << 14;
    public const uint PlayerTeleportEvent = 16;
    public const ulong PlayerTeleportSubscription = 1UL << 15;
    public const uint PlayerPunchAirEvent = 18;
    public const ulong PlayerPunchAirSubscription = 1UL << 17;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct StringView
{
    public byte* Data;
    public ulong Length;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct StringBuffer
{
    public byte* Data;
    public ulong Length;
    public ulong Capacity;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct RuntimeConfig
{
    public StringView PluginDirectory;
    public void* Host;
}

[StructLayout(LayoutKind.Sequential)]
public struct AbiHeader
{
    public uint Version;
    public uint Size;
    public ulong Subscriptions;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct HostHeader
{
    public uint Version;
    public uint Size;
    public ulong Context;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PluginApi
{
    public AbiHeader Header;
    public StringView Id;
    public delegate* unmanaged[Cdecl]<void*> Create;
    public delegate* unmanaged[Cdecl]<void*, StringBuffer*, int> Enable;
    public delegate* unmanaged[Cdecl]<void*, int> Disable;
    public void* Commands;
    public void* EntityTypeCount;
    public void* EntityTypeAt;
    public void* HandleEntity;
    public void* HandleCommand;
    public void* CommandEnumOptions;
    public delegate* unmanaged[Cdecl]<void*, void*, int> SetHost;
    public delegate* unmanaged[Cdecl]<void*, void> Destroy;
    public delegate* unmanaged[Cdecl]<void*, uint, void*, void*, int> HandleEvent;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerId
{
    public fixed byte Bytes[16];
    public ulong Generation;
}

[StructLayout(LayoutKind.Sequential)]
public struct Vec3
{
    public double X;
    public double Y;
    public double Z;
}

[StructLayout(LayoutKind.Sequential)]
public struct NativeRotation
{
    public double Yaw;
    public double Pitch;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerMoveInput
{
    public ulong Invocation;
    public PlayerId Player;
    public Vec3 OldPosition;
    public Vec3 NewPosition;
    public NativeRotation Rotation;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerMoveState
{
    public byte Cancelled;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerChatInput
{
    public ulong Invocation;
    public PlayerId Player;
    public StringView Message;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerChatState
{
    public byte Cancelled;
    public byte HasReplacement;
    public StringBuffer Replacement;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerQuitInput
{
    public ulong Invocation;
    public PlayerId Player;
    public StringView Name;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerQuitState
{
    public byte Reserved;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerFoodLossInput
{
    public ulong Invocation;
    public PlayerId Player;
    public int From;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerFoodLossState
{
    public byte Cancelled;
    public int To;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerToggleInput
{
    public ulong Invocation;
    public PlayerId Player;
    public byte After;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerEventInput
{
    public ulong Invocation;
    public PlayerId Player;
}

[StructLayout(LayoutKind.Sequential)]
public struct CancellableState
{
    public byte Cancelled;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerTeleportInput
{
    public ulong Invocation;
    public PlayerId Player;
    public Vec3 Position;
}
