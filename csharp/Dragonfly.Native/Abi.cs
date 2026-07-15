using System.Runtime.InteropServices;

namespace Dragonfly.Native;

public static partial class Abi
{
    public const uint PluginVersion = 11;
    public const uint HostVersion = 53;
    public const int Ok = 0;
    public const int Error = 1;
    public const uint PlayerCooldownHas = 0;
    public const uint PlayerCooldownSet = 1;

    public const uint WorldTaskExecute = 0;
    public const uint WorldTaskComplete = 1;
    public const uint WorldTaskSuccess = 0;
    public const uint WorldTaskCancelled = 1;
    public const uint WorldTaskWorldClosed = 2;
    public const uint WorldTaskPanicked = 3;
    public const uint WorldTaskFailed = 4;
    public const uint EntityOperationAdopt = 0;
    public const uint EntityOperationLoad = 1;
    public const uint EntityOperationSave = 2;
    public const uint EntityOperationTick = 3;
    public const uint EntityOperationHurt = 4;
    public const uint EntityOperationHeal = 5;
    public const uint EntityOperationDeath = 6;
    public const uint EntityOperationDestroy = 7;
    public const uint EntityOperationDecodeNbt = 8;
    public const uint EntityOperationEncodeNbt = 9;
    public const uint EntityOperationOpen = 10;
    public const uint EntityOperationBBox = 11;
    public const uint EntityOperationClose = 12;
    public const uint EntityOperationHandle = 13;
    public const uint EntityOperationPosition = 14;
    public const uint EntityOperationRotation = 15;
    public const uint EntityOperationTickExact = 16;
    public const uint EntityOperationReleaseOpen = 17;
    public const uint EntityCapabilityTicker = 1;
    public const uint WorldDimensionOverworld = 0;
    public const uint WorldDimensionNether = 1;
    public const uint WorldDimensionEnd = 2;
    public const uint PlayerMoveEvent = 1;
    public const ulong PlayerMoveSubscription = 1UL;
    public const uint PlayerChatEvent = 2;
    public const ulong PlayerChatSubscription = 1UL << 1;
    public const uint PlayerJoinEvent = 3;
    public const ulong PlayerJoinSubscription = 1UL << 2;
    public const uint PlayerQuitEvent = 4;
    public const ulong PlayerQuitSubscription = 1UL << 3;
    public const uint PlayerHurtEvent = 5;
    public const ulong PlayerHurtSubscription = 1UL << 4;
    public const uint PlayerHealEvent = 6;
    public const ulong PlayerHealSubscription = 1UL << 5;
    public const uint PlayerBlockBreakEvent = 7;
    public const ulong PlayerBlockBreakSubscription = 1UL << 6;
    public const uint PlayerBlockPlaceEvent = 8;
    public const ulong PlayerBlockPlaceSubscription = 1UL << 7;
    public const uint PlayerFoodLossEvent = 9;
    public const ulong PlayerFoodLossSubscription = 1UL << 8;
    public const uint PlayerDeathEvent = 10;
    public const ulong PlayerDeathSubscription = 1UL << 9;
    public const uint PlayerStartBreakEvent = 11;
    public const ulong PlayerStartBreakSubscription = 1UL << 10;
    public const uint PlayerFireExtinguishEvent = 12;
    public const ulong PlayerFireExtinguishSubscription = 1UL << 11;
    public const uint PlayerToggleSprintEvent = 13;
    public const ulong PlayerToggleSprintSubscription = 1UL << 12;
    public const uint PlayerToggleSneakEvent = 14;
    public const ulong PlayerToggleSneakSubscription = 1UL << 13;
    public const uint PlayerJumpEvent = 15;
    public const ulong PlayerJumpSubscription = 1UL << 14;
    public const uint PlayerTeleportEvent = 16;
    public const ulong PlayerTeleportSubscription = 1UL << 15;
    public const uint PlayerExperienceGainEvent = 17;
    public const ulong PlayerExperienceGainSubscription = 1UL << 16;
    public const uint PlayerPunchAirEvent = 18;
    public const ulong PlayerPunchAirSubscription = 1UL << 17;
    public const uint PlayerHeldSlotChangeEvent = 19;
    public const ulong PlayerHeldSlotChangeSubscription = 1UL << 18;
    public const uint PlayerSleepEvent = 20;
    public const ulong PlayerSleepSubscription = 1UL << 19;
    public const uint PlayerBlockPickEvent = 21;
    public const ulong PlayerBlockPickSubscription = 1UL << 20;
    public const uint PlayerLecternPageTurnEvent = 22;
    public const ulong PlayerLecternPageTurnSubscription = 1UL << 21;
    public const uint PlayerSignEditEvent = 23;
    public const ulong PlayerSignEditSubscription = 1UL << 22;
    public const uint PlayerItemUseEvent = 24;
    public const ulong PlayerItemUseSubscription = 1UL << 23;
    public const uint PlayerItemUseOnBlockEvent = 25;
    public const ulong PlayerItemUseOnBlockSubscription = 1UL << 24;
    public const uint PlayerItemConsumeEvent = 26;
    public const ulong PlayerItemConsumeSubscription = 1UL << 25;
    public const uint PlayerItemReleaseEvent = 27;
    public const ulong PlayerItemReleaseSubscription = 1UL << 26;
    public const uint PlayerItemDamageEvent = 28;
    public const ulong PlayerItemDamageSubscription = 1UL << 27;
    public const uint PlayerItemDropEvent = 29;
    public const ulong PlayerItemDropSubscription = 1UL << 28;
    public const uint PlayerAttackEntityEvent = 30;
    public const ulong PlayerAttackEntitySubscription = 1UL << 29;
    public const uint PlayerItemUseOnEntityEvent = 31;
    public const ulong PlayerItemUseOnEntitySubscription = 1UL << 30;
    public const uint PlayerChangeWorldEvent = 32;
    public const ulong PlayerChangeWorldSubscription = 1UL << 31;
    public const uint PlayerRespawnEvent = 33;
    public const ulong PlayerRespawnSubscription = 1UL << 32;
    public const uint PlayerSkinChangeEvent = 34;
    public const ulong PlayerSkinChangeSubscription = 1UL << 33;
    public const uint PlayerItemPickupEvent = 38;
    public const ulong PlayerItemPickupSubscription = 1UL << 37;
    public const uint PlayerTransferEvent = 39;
    public const ulong PlayerTransferSubscription = 1UL << 38;
    public const uint PlayerCommandExecutionEvent = 40;
    public const ulong PlayerCommandExecutionSubscription = 1UL << 39;
    public const uint PlayerDiagnosticsEvent = 41;
    public const ulong PlayerDiagnosticsSubscription = 1UL << 40;
    public const uint WorldLiquidFlowEvent = 42;
    public const ulong WorldLiquidFlowSubscription = 1UL << 41;
    public const uint WorldLiquidDecayEvent = 43;
    public const ulong WorldLiquidDecaySubscription = 1UL << 42;
    public const uint WorldLiquidHardenEvent = 44;
    public const ulong WorldLiquidHardenSubscription = 1UL << 43;
    public const uint WorldSoundEvent = 45;
    public const ulong WorldSoundSubscription = 1UL << 44;
    public const uint WorldFireSpreadEvent = 46;
    public const ulong WorldFireSpreadSubscription = 1UL << 45;
    public const uint WorldBlockBurnEvent = 47;
    public const ulong WorldBlockBurnSubscription = 1UL << 46;
    public const uint WorldCropTrampleEvent = 48;
    public const ulong WorldCropTrampleSubscription = 1UL << 47;
    public const uint WorldLeavesDecayEvent = 49;
    public const ulong WorldLeavesDecaySubscription = 1UL << 48;
    public const uint WorldEntitySpawnEvent = 50;
    public const ulong WorldEntitySpawnSubscription = 1UL << 49;
    public const uint WorldEntityDespawnEvent = 51;
    public const ulong WorldEntityDespawnSubscription = 1UL << 50;
    public const uint WorldExplosionEvent = 52;
    public const ulong WorldExplosionSubscription = 1UL << 51;
    public const uint WorldRedstoneUpdateEvent = 53;
    public const ulong WorldRedstoneUpdateSubscription = 1UL << 52;
    public const uint WorldCloseEvent = 54;
    public const ulong WorldCloseSubscription = 1UL << 53;
    public const uint PacketClientEvent = 55;
    public const ulong PacketClientSubscription = 1UL << 54;
    public const uint PacketServerEvent = 56;
    public const ulong PacketServerSubscription = 1UL << 55;
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
    public const uint InventoryMain = 0;
    public const uint InventoryArmour = 1;
    public const uint InventoryOffhand = 2;
    public const uint InventoryEnderChest = 3;
    public const uint SetBlockDisableBlockUpdates = 1;
    public const uint SetBlockDisableLiquidDisplacement = 1 << 1;
    public const uint SetBlockDisableRedstoneUpdates = 1 << 2;
    public const uint DamageSourceReducedByArmour = 1;
    public const uint DamageSourceReducedByResistance = 1 << 1;
    public const uint DamageSourceFire = 1 << 2;
    public const uint DamageSourceIgnoresTotem = 1 << 3;
    public const uint DamageSourceFireProtection = 1 << 4;
    public const uint DamageSourceFeatherFalling = 1 << 5;
    public const uint DamageSourceBlastProtection = 1 << 6;
    public const uint DamageSourceProjectileProtection = 1 << 7;
    public const uint DamageSourceCustom = 0;
    public const uint DamageSourceAttack = 1;
    public const uint DamageSourceBlock = 2;
    public const uint DamageSourceDrowning = 3;
    public const uint DamageSourceExplosion = 4;
    public const uint DamageSourceFall = 5;
    public const uint DamageSourceFireKind = 6;
    public const uint DamageSourceGlide = 7;
    public const uint DamageSourceInstant = 8;
    public const uint DamageSourceLava = 9;
    public const uint DamageSourceLightning = 10;
    public const uint DamageSourceMagma = 11;
    public const uint DamageSourcePoison = 12;
    public const uint DamageSourceProjectile = 13;
    public const uint DamageSourceStarvation = 14;
    public const uint DamageSourceSuffocation = 15;
    public const uint DamageSourceThorns = 16;
    public const uint DamageSourceVoid = 17;
    public const uint DamageSourceWither = 18;
    public const uint HealingSourceCustom = 0;
    public const uint HealingSourceFood = 1;
    public const uint HealingSourceInstant = 2;
    public const uint HealingSourceRegeneration = 3;
    public const uint EntityHasVelocity = 1;
    public const uint EntityHasNameTag = 2;
    public const uint EntityCanTeleport = 4;
    public const uint PlayerTransformTeleport = 0;
    public const uint PlayerTransformMove = 1;
    public const uint PlayerTransformVelocity = 2;
    public const uint PlayerTransformDisplace = 3;
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
public unsafe struct AllowInput
{
    public StringView Network;
    public StringView Address;
    public StringView IP;
    public StringView Zone;
    public StringView IdentityJson;
    public StringView ClientJson;
    public int Port;
    public byte IsUdp;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PacketFieldValue
{
    public uint Kind;
    public uint Reserved;
    public long Signed;
    public ulong Unsigned;
    public double Number;
    public double X;
    public double Y;
    public double Z;
    public NativeUuid Uuid;
    public StringBuffer Data;
}

[StructLayout(LayoutKind.Sequential)]
public struct PacketInput
{
    public ulong Packet;
    public uint PacketId;
    public uint Reserved;
    public StringView Xuid;
}

[StructLayout(LayoutKind.Sequential)]
public struct PacketState
{
    public byte Cancelled;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct BytesView
{
    public byte* Data;
    public ulong Length;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct BytesBuffer
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
public struct PlayerHealResult
{
    public double Healed;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerHurtResult
{
    public double Damage;
    public byte Vulnerable;
}

[StructLayout(LayoutKind.Sequential)]
public struct EffectView
{
    public int Type;
    public int Level;
    public long DurationNanoseconds;
    public double Potency;
    public byte Ambient;
    public byte ParticlesHidden;
    public byte Infinite;
    public long Tick;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct EffectBuffer
{
    public EffectView* Data;
    public ulong Length;
    public ulong Capacity;
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
public unsafe struct PlayerSnapshotBuffer
{
    public PlayerId Player;
    public StringBuffer Name;
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
public unsafe struct TitleView
{
    public StringView Text;
    public StringView Subtitle;
    public StringView ActionText;
    public long FadeInDurationNanoseconds;
    public long DurationNanoseconds;
    public long FadeOutDurationNanoseconds;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct ScoreboardView
{
    public StringView Name;
    public StringView* Lines;
    public ulong LineCount;
    public byte Padding;
    public byte Descending;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct HostApi
{
    public uint Version;
    public uint Size;
    public ulong Context;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, uint, StringView, int> PlayerText;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, TitleView, int> PlayerTitle;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, uint, Vec3, double, double, int> PlayerTransform;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, NativePlayerKinematics*, int> PlayerKinematics;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, uint, PlayerStateValue, int> PlayerStateSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, uint, PlayerStateValue*, int> PlayerStateGet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, uint, EffectView, int> PlayerEffect;
    public void* PlayerEntityVisibility;
    public void* PlayerSkinOpen;
    public delegate* unmanaged[Cdecl]<ulong, ulong, ulong, ulong, SkinAnimationInfo*, int> PlayerSkinAnimationInfo;
    public delegate* unmanaged[Cdecl]<ulong, ulong, ulong, SkinData*, int> PlayerSkinRead;
    public void* PlayerSkinClose;
    public void* PlayerSkinSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, InventoryId, uint*, int> InventorySize;
    public delegate* unmanaged[Cdecl]<ulong, ulong, InventoryId, uint, ulong*, ItemStackInfo*, int> InventoryItemOpen;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, uint, ulong*, ItemStackInfo*, int> PlayerHeldItemOpen;
    public delegate* unmanaged[Cdecl]<ulong, ulong, ulong, ItemStackData*, int> ItemStackRead;
    public delegate* unmanaged[Cdecl]<ulong, ulong, ulong, void> ItemStackClose;
    public delegate* unmanaged[Cdecl]<ulong, ulong, InventoryId, uint, ItemStackViewV3*, int> InventoryItemSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, InventoryId, ItemStackViewV3*, uint*, int> InventoryItemAdd;
    public delegate* unmanaged[Cdecl]<ulong, ulong, InventoryId, uint, int> InventoryClearSlot;
    public delegate* unmanaged[Cdecl]<ulong, ulong, InventoryId, int> InventoryClear;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, ItemStackViewV3*, ItemStackViewV3*, int> PlayerHeldItemsSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, uint, int> PlayerHeldSlotSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, ScoreboardView, int> PlayerScoreboard;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, int> PlayerScoreboardRemove;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, FormView*, int> PlayerFormSend;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, int> PlayerFormClose;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, StringBuffer*, int> WorldName;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, int> WorldUnload;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, int> WorldSave;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, BlockData*, int> WorldBlockGet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, BlockView*, uint, int> WorldBlockSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, long*, int> WorldTimeGet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, long, int> WorldTimeSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos*, int> WorldSpawnGet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, int> WorldSpawnSet;
    public void* WorldEntitySpawn;
    public delegate* unmanaged[Cdecl]<ulong, ulong, EntityId, EntityState*, int> EntityState;
    public void* EntityTeleport;
    public void* EntityVelocitySet;
    public void* EntityNameTagSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, EntityId, int> EntityDespawn;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, Vec3, ParticleView*, int> WorldParticleAdd;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, Vec3, SoundViewV1*, int> WorldSoundPlay;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, SoundViewV1*, int> PlayerSoundPlay;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, double, HealingSourceView*, PlayerHealResult*, int> PlayerHeal;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, double, DamageSourceView*, PlayerHurtResult*, int> PlayerHurt;
    public delegate* unmanaged[Cdecl]<ulong, ulong, ulong, SkinInfo*, int> SkinSnapshotInfo;
    public delegate* unmanaged[Cdecl]<ulong, ulong, ulong, SkinView*, int> SkinSnapshotSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, WorldId, Vec3, int> PlayerTransfer;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, EffectBuffer*, int> PlayerEffects;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, int> PlayerEffectsClear;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BlockPos, byte*, BlockData*, int> WorldLiquidGet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, uint, PlayerStateValue, PlayerStateValue*, int> PlayerAction;
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
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, ItemStackSnapshot*, ItemStackSnapshot*, int> PlayerHeldItemsOpen;
    public delegate* unmanaged[Cdecl]<ulong, ulong, EntityId, PlayerSnapshotBuffer*, int> EntityPlayer;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId*, int> WorldCurrent;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, byte, ulong*, int> WorldEntityIteratorOpen;
    public delegate* unmanaged[Cdecl]<ulong, ulong, ulong, EntityId*, byte*, int> WorldEntityIteratorNext;
    public delegate* unmanaged[Cdecl]<ulong, ulong, ulong, void> WorldEntityIteratorClose;
    public delegate* unmanaged[Cdecl]<ulong, ulong, EntityId, EntityHandleId*, int> EntityHandle;
    public delegate* unmanaged[Cdecl]<ulong, ulong, EntityHandleId, EntityId*, byte*, int> EntityHandleEntity;
    public delegate* unmanaged[Cdecl]<ulong, EntityHandleId, NativeUuid*, int> EntityHandleUuid;
    public delegate* unmanaged[Cdecl]<ulong, EntityHandleId, byte*, int> EntityHandleClosed;
    public delegate* unmanaged[Cdecl]<ulong, EntityHandleId, int> EntityHandleClose;
    public delegate* unmanaged[Cdecl]<ulong, ulong, EntityId, EntityHandleId*, int> WorldEntityRemove;
    public delegate* unmanaged[Cdecl]<ulong, ulong, EntityHandleId, Vec3*, EntityId*, int> WorldEntityAdd;
    public delegate* unmanaged[Cdecl]<ulong, ulong, ulong*, int> ServerPlayersOpen;
    public delegate* unmanaged[Cdecl]<ulong, ulong, ulong, ulong*, PlayerSnapshotBuffer*, byte*, int> ServerPlayersNext;
    public delegate* unmanaged[Cdecl]<ulong, ulong, ulong, void> ServerPlayersClose;
    public delegate* unmanaged[Cdecl]<ulong, NativeUuid, EntityHandleId*, byte*, int> ServerPlayer;
    public delegate* unmanaged[Cdecl]<ulong, StringView, EntityHandleId*, byte*, int> ServerPlayerByName;
    public delegate* unmanaged[Cdecl]<ulong, long*, int> ServerMaxPlayerCount;
    public delegate* unmanaged[Cdecl]<ulong, long*, int> ServerPlayerCount;
    public delegate* unmanaged[Cdecl]<ulong, StringView, EntityHandleId*, byte*, int> ServerPlayerByXuid;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, StringBuffer*, int> PlayerXuid;
    public delegate* unmanaged[Cdecl]<ulong, uint, WorldId*, int> ServerWorld;
    public delegate* unmanaged[Cdecl]<ulong, WorldId, ulong, ulong, long, int> WorldSchedule;
    public delegate* unmanaged[Cdecl]<ulong, WorldConfigV1*, WorldId*, int> WorldNew;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, BBox, ulong*, int> WorldEntitiesWithinOpen;
    public delegate* unmanaged[Cdecl]<ulong, StringView, StringView, byte*, BlockData*, int> BlockByName;
    public delegate* unmanaged[Cdecl]<ulong, EntityNewView*, EntityHandleId*, int> EntityNew;
    public delegate* unmanaged[Cdecl]<ulong, EntityHandleId, StringBuffer*, int> EntityHandleType;
    public delegate* unmanaged[Cdecl]<ulong, ulong, ulong, byte*, int> WorldTaskCancel;
    public delegate* unmanaged[Cdecl]<ulong, ulong, uint, PacketFieldValue*, int> PacketFieldGet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, uint, PacketFieldValue*, int> PacketFieldSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, uint*, int> WorldDimensionGet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, byte*, int> WorldTimeCycleGet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, byte, int> WorldTimeCycleSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, long, int> WorldRequiredSleepDurationSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, long*, int> WorldDefaultGameModeGet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, long, int> WorldDefaultGameModeSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, int, int> WorldTickRangeSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, DifficultyView*, int> WorldDifficultyGet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, DifficultyView, int> WorldDifficultySet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, ulong, int> PlayerPacketWrite;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, NativeUuid, BlockPos*, int> WorldPlayerSpawnGet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, WorldId, NativeUuid, BlockPos, int> WorldPlayerSpawnSet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, uint, StringBuffer*, int> PlayerStringGet;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, StringView, StringView, int> PlayerToast;
    public delegate* unmanaged[Cdecl]<ulong, ulong, PlayerId, uint, StringView, int, long, byte*, int> PlayerCooldown;
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
    public delegate* unmanaged[Cdecl]<void*, ulong, ulong, uint, uint, int> HandleScheduled;
    public delegate* unmanaged[Cdecl]<void*, AllowInput*, StringBuffer*, byte*, int> Allow;
}

[StructLayout(LayoutKind.Sequential)]
public struct EntityTypeDescriptorV2
{
    public StringView SaveId;
    public StringView NetworkId;
    public ulong TypeKey;
}

[StructLayout(LayoutKind.Sequential)]
public struct EntitySpawnOptions
{
    public Vec3 Position;
    public NativeRotation Rotation;
    public Vec3 Velocity;
    public StringView NameTag;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct EntityNewView
{
    public NativeUuid Id;
    public EntitySpawnOptions Options;
    public StringView EntityType;
    public ulong Plugin;
    public ulong LocalType;
    public ulong Opaque;
    public long FireDurationNanoseconds;
    public long AgeNanoseconds;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct EntityDataState
{
    public Vec3 Position;
    public Vec3 Velocity;
    public NativeRotation Rotation;
    public StringBuffer Name;
    public long FireDurationNanoseconds;
    public long AgeNanoseconds;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct EntityExactInput
{
    public ulong Invocation;
    public EntityHandleId Handle;
    public EntityDataState* Data;
    public BytesView Nbt;
    public long Current;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct EntityExactState
{
    public ulong Instance;
    public uint Capabilities;
    public uint Reserved;
    public BBox BBox;
    public EntityHandleId Handle;
    public Vec3 Position;
    public NativeRotation Rotation;
    public BytesBuffer Nbt;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerId
{
    public fixed byte Bytes[16];
    public ulong Generation;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct EntityId
{
    public fixed byte Bytes[16];
    public ulong Generation;
}

[StructLayout(LayoutKind.Sequential)]
public struct EntityHandleId
{
    public ulong Value;
    public ulong Generation;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct NativeUuid
{
    public fixed byte Bytes[16];
}

[StructLayout(LayoutKind.Sequential)]
public struct InventoryId
{
    public PlayerId Player;
    public uint Kind;
    public uint Reserved;
}

[StructLayout(LayoutKind.Sequential)]
public struct ByteSpan
{
    public ulong Offset;
    public ulong Length;
}

[StructLayout(LayoutKind.Sequential)]
public struct ItemEnchantment
{
    public uint Id;
    public uint Level;
}

[StructLayout(LayoutKind.Sequential)]
public struct ItemStackInfo
{
    public int Metadata;
    public uint Count;
    public uint Damage;
    public byte Unbreakable;
    public int AnvilCost;
    public ulong IdentifierLength;
    public ulong CustomNameLength;
    public ulong LoreBytesLength;
    public ulong LoreCount;
    public ulong NbtLength;
    public ulong ValuesNbtLength;
    public ulong EnchantmentCount;
}

[StructLayout(LayoutKind.Sequential)]
public struct ItemStackSnapshot
{
    public ulong Snapshot;
    public ItemStackInfo Info;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct ItemStackData
{
    public StringBuffer Identifier;
    public StringBuffer CustomName;
    public StringBuffer LoreBytes;
    public StringBuffer Nbt;
    public StringBuffer ValuesNbt;
    public ByteSpan* Lore;
    public ulong LoreCapacity;
    public ItemEnchantment* Enchantments;
    public ulong EnchantmentCapacity;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct ItemStackViewV3
{
    public StringView Identifier;
    public int Metadata;
    public uint Count;
    public uint Damage;
    public byte Unbreakable;
    public int AnvilCost;
    public StringView CustomName;
    public StringView* Lore;
    public ulong LoreCount;
    public StringView Nbt;
    public StringView ValuesNbt;
    public ItemEnchantment* Enchantments;
    public ulong EnchantmentCount;
}

[StructLayout(LayoutKind.Sequential)]
public struct Vec3
{
    public double X;
    public double Y;
    public double Z;
}

[StructLayout(LayoutKind.Sequential)]
public struct BBox
{
    public Vec3 Min;
    public Vec3 Max;
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
public unsafe struct WorldConfigV1
{
    public uint StructSize;
    public uint Dimension;
    public uint ProviderKind;
    public uint ReadOnly;
    public StringView ProviderPath;
    public long SaveIntervalNanoseconds;
    public long ChunkUnloadIntervalNanoseconds;
    public int RandomTickSpeed;
    public uint Reserved;
}

[StructLayout(LayoutKind.Sequential)]
public struct DifficultyView
{
    public uint ID;
    public byte Builtin;
    public byte FoodRegenerates;
    public ushort Reserved;
    public double StarvationHealthLimit;
    public int FireSpreadIncrease;
    public uint Reserved2;
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
public unsafe struct DamageSourceView
{
    public StringView Name;
    public uint Kind;
    public uint Flags;
    public EntityId Entity;
    public EntityId SecondaryEntity;
    public BlockView* Block;
    public byte Data;
}

[StructLayout(LayoutKind.Sequential)]
public struct HealingSourceView
{
    public StringView Name;
    public uint Kind;
    public byte Data;
}

[StructLayout(LayoutKind.Sequential)]
public struct SkinAnimationInfo
{
    public uint Width;
    public uint Height;
    public uint AnimationType;
    public long FrameCount;
    public long Expression;
    public ulong PixelsLength;
}

[StructLayout(LayoutKind.Sequential)]
public struct SkinInfo
{
    public uint Width;
    public uint Height;
    public byte Persona;
    public ulong PlayFabIdLength;
    public ulong FullIdLength;
    public ulong PixelsLength;
    public ulong ModelDefaultLength;
    public ulong ModelAnimatedFaceLength;
    public ulong ModelLength;
    public uint CapeWidth;
    public uint CapeHeight;
    public ulong CapePixelsLength;
    public ulong AnimationCount;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct SkinData
{
    public StringBuffer PlayFabId;
    public StringBuffer FullId;
    public StringBuffer Pixels;
    public StringBuffer ModelDefault;
    public StringBuffer ModelAnimatedFace;
    public StringBuffer Model;
    public StringBuffer CapePixels;
    public StringBuffer* AnimationPixels;
    public ulong AnimationCapacity;
}

[StructLayout(LayoutKind.Sequential)]
public struct SkinAnimationView
{
    public uint Width;
    public uint Height;
    public uint AnimationType;
    public long FrameCount;
    public long Expression;
    public StringView Pixels;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct SkinView
{
    public uint Width;
    public uint Height;
    public byte Persona;
    public StringView PlayFabId;
    public StringView FullId;
    public StringView Pixels;
    public StringView ModelDefault;
    public StringView ModelAnimatedFace;
    public StringView Model;
    public uint CapeWidth;
    public uint CapeHeight;
    public StringView CapePixels;
    public SkinAnimationView* Animations;
    public ulong AnimationCount;
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
public unsafe struct SoundViewV1
{
    public uint Kind;
    public uint Data;
    public int Integer;
    public uint Flags;
    public double Scalar;
    public BlockView* Block;
    public ItemStackViewV3* Item;
}

[StructLayout(LayoutKind.Sequential)]
public struct NativeRotation
{
    public double Yaw;
    public double Pitch;
}

[StructLayout(LayoutKind.Sequential)]
public struct NativePlayerKinematics
{
    public Vec3 Position;
    public Vec3 Velocity;
    public NativeRotation Rotation;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct EntityState
{
    public Vec3 Position;
    public NativeRotation Rotation;
    public Vec3 Velocity;
    public uint Capabilities;
    public WorldId World;
    public StringBuffer EntityType;
    public StringBuffer NameTag;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerMoveInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
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
    public PlayerSnapshot Player;
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
public unsafe struct PlayerJoinInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public StringView Name;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerJoinState
{
    public byte Cancelled;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerQuitInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public StringView Name;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerQuitState
{
    public byte Reserved;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerHurtInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public byte Immune;
    public DamageSourceView Source;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerHurtState
{
    public byte Cancelled;
    public double Damage;
    public long AttackImmunityNanoseconds;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerHealInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public HealingSourceView Source;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerHealState
{
    public byte Cancelled;
    public double Health;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerFoodLossInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public int From;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerFoodLossState
{
    public byte Cancelled;
    public int To;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerDeathInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public DamageSourceView Source;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerDeathState
{
    public byte KeepInventory;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerToggleInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public byte After;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerEventInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
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
    public PlayerSnapshot Player;
    public Vec3 Position;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerBlockBreakInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public BlockPos Position;
    public BlockView Block;
    public ItemStackViewV3* Drops;
    public ulong DropCount;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerBlockBreakState
{
    public byte Cancelled;
    public int Experience;
    public ItemStackViewV3* ReplacementDrops;
    public ulong ReplacementDropCount;
    public void* ReplacementContext;
    public delegate* unmanaged[Cdecl]<void*, void> ReplacementDrop;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerBlockInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public BlockPos Position;
    public BlockView Block;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerBlockPositionInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public BlockPos Position;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerHeldSlotChangeInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public int From;
    public int To;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerSleepState
{
    public byte Cancelled;
    public byte SendReminder;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerLecternPageTurnInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public BlockPos Position;
    public int OldPage;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerLecternPageTurnState
{
    public byte Cancelled;
    public int NewPage;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerSignEditInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public BlockPos Position;
    public byte FrontSide;
    public StringView OldText;
    public StringView NewText;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerItemUseOnBlockInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public BlockPos Position;
    public int Face;
    public Vec3 ClickPosition;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerItemInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public ItemStackViewV3 Item;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerItemReleaseInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public ItemStackViewV3 Item;
    public long DurationNanoseconds;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerIntegerState
{
    public byte Cancelled;
    public int Value;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerItemPickupInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public ItemStackViewV3 Item;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerItemPickupState
{
    public byte Cancelled;
    public ItemStackViewV3* Replacement;
    public void* ReplacementContext;
    public delegate* unmanaged[Cdecl]<void*, void> ReplacementDrop;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerAttackEntityInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public EntityId Target;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerAttackEntityState
{
    public byte Cancelled;
    public double KnockbackForce;
    public double KnockbackHeight;
    public byte Critical;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerItemUseOnEntityInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public EntityId Target;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerChangeWorldInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public WorldId Before;
    public WorldId After;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerRespawnState
{
    public Vec3 Position;
    public WorldId World;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerSkinChangeInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public ulong Snapshot;
}

[StructLayout(LayoutKind.Sequential)]
public struct UDPAddrView
{
    public StringView IP;
    public int Port;
    public StringView Zone;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerTransferInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerTransferState
{
    public byte Cancelled;
    public UDPAddrView Address;
    public void* ReplacementContext;
    public delegate* unmanaged[Cdecl]<void*, void> ReplacementDrop;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerCommandExecutionInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public StringView CommandName;
    public StringView CommandDescription;
    public StringView CommandUsage;
    public StringView* CommandAliases;
    public ulong CommandAliasCount;
    public StringView* Arguments;
    public ulong ArgumentCount;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct PlayerCommandExecutionState
{
    public byte Cancelled;
    public StringView* ReplacementArguments;
    public ulong ReplacementArgumentCount;
    public void* ReplacementContext;
    public delegate* unmanaged[Cdecl]<void*, void> ReplacementDrop;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerDiagnosticsInput
{
    public ulong Invocation;
    public PlayerSnapshot Player;
    public double AverageFramesPerSecond;
    public double AverageServerSimTickTime;
    public double AverageClientSimTickTime;
    public double AverageBeginFrameTime;
    public double AverageInputTime;
    public double AverageRenderTime;
    public double AverageEndFrameTime;
    public double AverageRemainderTimePercent;
    public double AverageUnaccountedTimePercent;
}

[StructLayout(LayoutKind.Sequential)]
public struct PlayerDiagnosticsState
{
    public byte Reserved;
}

[StructLayout(LayoutKind.Sequential)]
public struct WorldCancellableState
{
    public byte Cancelled;
}

[StructLayout(LayoutKind.Sequential)]
public struct WorldNotificationState
{
    public byte Reserved;
}

[StructLayout(LayoutKind.Sequential)]
public struct WorldLiquidFlowInput
{
    public ulong Invocation;
    public BlockPos From;
    public BlockPos Into;
    public BlockView Liquid;
    public BlockView Replaced;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct WorldLiquidDecayInput
{
    public ulong Invocation;
    public BlockPos Position;
    public BlockView Before;
    public BlockView* After;
}

[StructLayout(LayoutKind.Sequential)]
public struct WorldLiquidHardenInput
{
    public ulong Invocation;
    public BlockPos Position;
    public BlockView LiquidHardened;
    public BlockView OtherLiquid;
    public BlockView NewBlock;
}

[StructLayout(LayoutKind.Sequential)]
public struct WorldSoundInput
{
    public ulong Invocation;
    public SoundViewV1 Sound;
    public Vec3 Position;
}

[StructLayout(LayoutKind.Sequential)]
public struct WorldFireSpreadInput
{
    public ulong Invocation;
    public BlockPos From;
    public BlockPos To;
}

[StructLayout(LayoutKind.Sequential)]
public struct WorldPositionInput
{
    public ulong Invocation;
    public BlockPos Position;
}

[StructLayout(LayoutKind.Sequential)]
public struct WorldEntityInput
{
    public ulong Invocation;
    public EntityId Entity;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct WorldExplosionInput
{
    public ulong Invocation;
    public Vec3 Position;
    public EntityId* Entities;
    public ulong EntityCount;
    public BlockPos* Blocks;
    public ulong BlockCount;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct WorldExplosionState
{
    public byte Cancelled;
    public byte SpawnFire;
    public double ItemDropChance;
    public EntityId* ReplacementEntities;
    public ulong ReplacementEntityCount;
    public BlockPos* ReplacementBlocks;
    public ulong ReplacementBlockCount;
    public void* ReplacementContext;
    public delegate* unmanaged[Cdecl]<void*, void> ReplacementDrop;
}

[StructLayout(LayoutKind.Sequential)]
public unsafe struct WorldRedstoneUpdateInput
{
    public ulong Invocation;
    public BlockPos Position;
    public BlockPos ChangedNeighbour;
    public byte HasChangedNeighbour;
    public byte ChangedRedstoneRelevant;
    public BlockPos Source;
    public byte HasSource;
    public BlockView Before;
    public BlockView* After;
    public int OldPower;
    public int NewPower;
    public long CurrentTick;
    public uint Cause;
}

[StructLayout(LayoutKind.Sequential)]
public struct WorldCloseInput
{
    public ulong Invocation;
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
