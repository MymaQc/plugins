using System.Runtime.InteropServices;

namespace Dragonfly.Native;

public static class Abi
{
    public const uint PluginVersion = 5;
    public const uint HostVersion = 29;
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
    public const uint CommandParameterSubcommand = 1;
    public const uint CommandParameterEnum = 2;
    public const uint CommandParameterString = 3;
    public const uint CommandParameterInteger = 4;
    public const uint CommandParameterFloat = 5;
    public const uint CommandParameterBool = 6;
    public const uint CommandParameterDynamicEnum = 7;
    public const uint CommandParameterPlayer = 8;
    public const uint CommandParameterRawText = 9;
    public const uint CommandParameterVector = 10;
    public const uint CommandSourceUnknown = 0;
    public const uint CommandSourcePlayer = 1;
    public const uint CommandSourceConsole = 2;
    public const uint PlayerTextMessage = 0;
    public const uint PlayerTextTip = 1;
    public const uint PlayerTextPopup = 2;
    public const uint PlayerTextJukeboxPopup = 3;
    public const uint PlayerTextNameTag = 4;
    public const uint PlayerTextDisconnect = 5;
    public const uint SetBlockDisableBlockUpdates = 1;
    public const uint SetBlockDisableLiquidDisplacement = 1 << 1;
    public const uint SetBlockDisableRedstoneUpdates = 1 << 2;
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
public struct PlayerStateValue
{
    public double Number;
    public long Integer;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerSnapshot
{
    public PlayerId Player;
    public StringView Name;
    public ulong LatencyMilliseconds;
    public Vec3 Position;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct FormView
{
    public StringView RequestJson;
    public void* CallbackContext;
    public delegate* unmanaged[Cdecl]<void*, ulong, PlayerSnapshot*, uint, StringView, int> Response;
    public delegate* unmanaged[Cdecl]<void*, void> Drop;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct HostApi
{
    public uint Version;
    public uint Size;
    public ulong Context;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, uint, StringView, int> PlayerText;
    public void* PlayerTitle;
    public void* PlayerTransform;
    public void* PlayerRotation;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, uint, PlayerStateValue, int> PlayerStateSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, uint, PlayerStateValue*, int> PlayerStateGet;
    public void* PlayerEffect;
    public void* PlayerEntityVisibility;
    public void* PlayerSkinOpen;
    public void* PlayerSkinAnimationInfo;
    public void* PlayerSkinRead;
    public void* PlayerSkinClose;
    public void* PlayerSkinSet;
    public void* InventorySize;
    public void* InventoryItemOpen;
    public void* PlayerHeldItemOpen;
    public void* ItemStackRead;
    public void* ItemStackClose;
    public void* InventoryItemSet;
    public void* InventoryItemAdd;
    public void* InventoryClearSlot;
    public void* InventoryClear;
    public void* PlayerHeldItemsSet;
    public void* PlayerHeldSlotSet;
    public void* PlayerScoreboard;
    public void* PlayerScoreboardRemove;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, FormView*, int> PlayerFormSend;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, int> PlayerFormClose;
    public void* WorldLookup;
    public void* WorldOpen;
    public void* WorldName;
    public void* WorldUnload;
    public void* WorldSave;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, BlockData*, int> WorldBlockGet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, BlockView*, uint, int> WorldBlockSet;
    public void* WorldTimeGet;
    public void* WorldTimeSet;
    public void* WorldSpawnGet;
    public void* WorldSpawnSet;
    public void* WorldEntitySpawn;
    public void* WorldEntities;
    public void* WorldPlayers;
    public void* EntityState;
    public void* EntityTeleport;
    public void* EntityVelocitySet;
    public void* EntityNameTagSet;
    public void* EntityDespawn;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, Vec3, ParticleView*, int> WorldParticleAdd;
    public void* WorldSoundPlay;
    public void* PlayerSoundPlay;
    public void* PlayerHeal;
    public void* PlayerHurt;
    public void* SkinSnapshotInfo;
    public void* SkinSnapshotSet;
    public void* WorldOpenSpec;
    public void* PlayerTransfer;
    public void* PlayerEffects;
    public void* PlayerEffectsClear;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, byte*, BlockData*, int> WorldLiquidGet;
    public void* PlayerExperienceSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockRange*, int> WorldRange;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, byte*, BlockData*, int> WorldBlockLoaded;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, int, BlockView*, ulong, ulong*, int> WorldBlocksWithinOpen;
    public delegate* unmanaged[Cdecl]<ulong, ulong, ulong, BlockPos*, byte*, int> WorldBlocksWithinNext;
    public delegate* unmanaged[Cdecl]<ulong, ulong, ulong, void> WorldBlocksWithinClose;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, int, int, int*, int> WorldHighestLightBlocker;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, int, int, int*, int> WorldHighestBlock;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, byte*, int> WorldLight;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, byte*, int> WorldSkyLight;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, BlockView*, int> WorldLiquidSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, BlockView*, long, int> WorldBlockUpdateSchedule;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, int*, int> WorldBiomeGet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, int, int> WorldBiomeSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, double*, int> WorldTemperature;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, byte*, int> WorldRainingAt;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, byte*, int> WorldSnowingAt;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, byte*, int> WorldThunderingAt;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, byte*, int> WorldRaining;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, byte*, int> WorldThundering;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, long*, int> WorldCurrentTick;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PluginApi
{
    public AbiHeader Header;
    public StringView Id;
    public delegate* unmanaged[Cdecl]<void*> Create;
    public delegate* unmanaged[Cdecl]<void*, StringBuffer*, int> Enable;
    public delegate* unmanaged[Cdecl]<void*, int> Disable;
    public delegate* unmanaged[Cdecl]<void*, ulong*, CommandDescriptor*> Commands;
    public void* EntityTypeCount;
    public void* EntityTypeAt;
    public void* HandleEntity;
    public delegate* unmanaged[Cdecl]<void*, ulong, CommandInput*, CommandState*, int> HandleCommand;
    public delegate* unmanaged[Cdecl]<void*, ulong, ulong, ulong, CommandEnumContext*, StringBuffer*, int> CommandEnumOptions;
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
public struct BlockPos
{
    public int X;
    public int Y;
    public int Z;
}

[StructLayout(LayoutKind.Sequential)]
public struct BlockRange
{
    public int Min;
    public int Max;
}

[StructLayout(LayoutKind.Sequential)]
public struct WorldId
{
    public ulong Value;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct BlockData
{
    public StringBuffer Identifier;
    public StringBuffer PropertiesNbt;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct BlockView
{
    public StringView Identifier;
    public StringView PropertiesNbt;
}

[StructLayout(LayoutKind.Sequential)]
public struct Rgba
{
    public byte R;
    public byte G;
    public byte B;
    public byte A;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct ParticleView
{
    public uint Kind;
    public uint Data;
    public int Pitch;
    public Rgba Colour;
    public BlockPos Diff;
    public BlockView* Block;
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

[StructLayout(LayoutKind.Sequential)]
public unsafe struct CommandParameter
{
    public uint Kind;
    public byte Optional;
    public StringView Name;
    public StringView Suffix;
    public StringView* Values;
    public ulong ValueCount;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct CommandOverload
{
    public CommandParameter* Parameters;
    public ulong ParameterCount;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct CommandDescriptor
{
    public StringView Name;
    public StringView Description;
    public StringView* Aliases;
    public ulong AliasCount;
    public CommandOverload* Overloads;
    public ulong OverloadCount;
}

[StructLayout(LayoutKind.Sequential)]
public struct CommandPlayer
{
    public PlayerId Player;
    public StringView Name;
    public ulong LatencyMilliseconds;
    public Vec3 Position;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct CommandEnumContext
{
    public StringView Source;
    public uint SourceKind;
    public PlayerId SourcePlayer;
    public Vec3 SourcePosition;
    public CommandPlayer* OnlinePlayers;
    public ulong OnlinePlayerCount;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct CommandInput
{
    public ulong Invocation;
    public ulong Overload;
    public StringView Source;
    public StringView* Arguments;
    public ulong ArgumentCount;
    public uint SourceKind;
    public PlayerId SourcePlayer;
    public Vec3 SourcePosition;
    public CommandPlayer* OnlinePlayers;
    public ulong OnlinePlayerCount;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct CommandState
{
    public byte Failed;
    public StringBuffer Output;
}
