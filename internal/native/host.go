package native

import (
	"sync"
	"sync/atomic"
	"time"
)

type PlayerForm struct {
	ID          uint64
	RequestJSON []byte
}

// PlayerSnapshot is a callback-scoped view of a connected player. The ID may
// be retained, but values that require an invocation are only usable while the
// callback that supplied the snapshot is running.
type PlayerSnapshot struct {
	Player              PlayerID
	Name                string
	LatencyMilliseconds uint64
	Position            Vec3
}

type PlayerTransformKind uint32

const (
	PlayerTransformTeleport PlayerTransformKind = iota
	PlayerTransformMove
	PlayerTransformVelocity
	PlayerTransformDisplace
)

type PlayerKinematics struct {
	Position, Velocity Vec3
	Rotation           Rotation
}

type PlayerTitle struct {
	Text       string
	Subtitle   string
	ActionText string
	FadeIn     time.Duration
	Duration   time.Duration
	FadeOut    time.Duration
}

type PlayerScoreboard struct {
	Name       string
	Lines      []string
	Padding    bool
	Descending bool
}

type SkinAnimation struct {
	Width, Height uint32
	Type          uint32
	FrameCount    int64
	Expression    int64
	Pixels        []byte
}

type PlayerSkin struct {
	Width, Height                   uint32
	Persona                         bool
	PlayFabID, FullID               string
	Pixels                          []byte
	ModelDefault, ModelAnimatedFace string
	Model                           []byte
	CapeWidth, CapeHeight           uint32
	CapePixels                      []byte
	Animations                      []SkinAnimation
}

type InventoryKind uint32

const (
	InventoryMain InventoryKind = iota
	InventoryArmour
	InventoryOffhand
	InventoryEnderChest
)

type InventoryID struct {
	Player PlayerID
	Kind   InventoryKind
}

type ItemEnchantment struct {
	ID, Level uint32
}

type ItemStack struct {
	Identifier     string
	Metadata       int32
	Count, Damage  uint32
	Unbreakable    bool
	AnvilCost      int32
	CustomName     string
	Lore           []string
	NBT, ValuesNBT []byte
	Enchantments   []ItemEnchantment
}

// WorldID is an opaque, process-local handle. Handles are never reused.
type WorldID uint64

// PacketHandle identifies a decoded gophertunnel packet borrowed for one
// synchronous intercept callback.
type PacketHandle uint64

type PacketFieldKind uint32

const (
	PacketFieldInvalid PacketFieldKind = iota
	PacketFieldBool
	PacketFieldSigned
	PacketFieldUnsigned
	PacketFieldFloat
	PacketFieldString
	PacketFieldBytes
	PacketFieldVec2
	PacketFieldVec3
	PacketFieldUUID
	PacketFieldJSON
)

type PacketFieldValue struct {
	Kind     PacketFieldKind
	Signed   int64
	Unsigned uint64
	Number   float64
	X, Y, Z  float64
	UUID     [16]byte
	Data     []byte
}

type WorldDimension uint32

const (
	WorldDimensionOverworld WorldDimension = iota
	WorldDimensionNether
	WorldDimensionEnd
)

type DifficultyView struct {
	ID                    uint32
	Builtin               bool
	FoodRegenerates       bool
	StarvationHealthLimit float64
	FireSpreadIncrease    int32
}

type WorldProviderKind uint32

const (
	WorldProviderNop WorldProviderKind = iota
	WorldProviderMCDB
)

type WorldConfig struct {
	Dimension           WorldDimension
	Provider            WorldProviderKind
	ProviderPath        string
	ReadOnly            bool
	SaveInterval        time.Duration
	ChunkUnloadInterval time.Duration
	RandomTickSpeed     int
}

type WorldBlock struct {
	Identifier    string
	PropertiesNBT []byte
}

type BlockRange struct {
	Min int32
	Max int32
}

type BlockIteratorID uint64

// EntityIteratorID identifies one invocation-scoped pull iterator over the
// entities in the current Dragonfly transaction.
type EntityIteratorID uint64

// PlayerIteratorID identifies one pull iterator over the players connected to
// the Dragonfly server.
type PlayerIteratorID uint64

type WorldSetOpts struct {
	DisableBlockUpdates       bool
	DisableLiquidDisplacement bool
	DisableRedstoneUpdates    bool
}

type WorldTaskPhase uint32

const (
	WorldTaskExecute WorldTaskPhase = iota
	WorldTaskComplete
)

type WorldTaskResult uint32

const (
	WorldTaskSuccess WorldTaskResult = iota
	WorldTaskCancelled
	WorldTaskWorldClosed
	WorldTaskPanicked
	WorldTaskFailed
)

type EntityKind uint32

const (
	EntityText EntityKind = iota
	EntityLightning
	EntityTNT
	EntityExperienceOrb
	EntityItem
	EntityFallingBlock
	EntityArrow
	EntityEgg
	EntitySnowball
	EntityEnderPearl
	EntityBottleOfEnchanting
	EntitySplashPotion
	EntityLingeringPotion
	EntityCustom
)

type EntityTypeDefinition struct {
	SaveID, NetworkID string
	TypeKey           uint64
}

// EntityInstanceID identifies plugin-owned state for one custom entity.
// It is opaque outside the native runtime and is never reused by that runtime.
type EntityInstanceID uint64

type EntityOpenID uint64

type EntityCommonData struct {
	Position, Velocity Vec3
	Rotation           Rotation
	Name               string
	FireDuration, Age  time.Duration
}

const EntityCapabilityTicker uint32 = 1

type EntityLoadInput struct {
	Data    []byte
	Version uint32
}

type EntitySaveOutput struct {
	Data    []byte
	Version uint32
}

type EntityTickInput struct {
	Invocation InvocationID
	Entity     EntityID
	Current    int64
	Age        time.Duration
}

type EntityTickOutput struct {
	Despawn bool
}

type EntityHurtInput struct {
	Invocation InvocationID
	Entity     EntityID
	Source     DamageSource
	Health     float64
	MaxHealth  float64
	Damage     float64
}

type EntityHurtOutput struct {
	Damage    float64
	Cancelled bool
}

type EntityHealInput struct {
	Invocation InvocationID
	Entity     EntityID
	Source     HealingSource
	Health     float64
	MaxHealth  float64
	Amount     float64
}

type EntityHealOutput struct {
	Amount    float64
	Cancelled bool
}

type EntityDeathInput struct {
	Invocation InvocationID
	Entity     EntityID
	Source     DamageSource
	Health     float64
	Damage     float64
}

type EntityDeathOutput struct {
	Cancelled bool
}

const (
	EntityArrowCritical uint32 = 1 << iota
	EntityArrowDisablePickup
	EntityArrowObtainOnPickup
)

const (
	EntityLightningBlockFire uint32 = 1
	EntityItemHasPickupDelay uint32 = 1
)

type EntitySpawn struct {
	Kind                        EntityKind
	Flags                       uint32
	Position, Velocity          Vec3
	Rotation                    Rotation
	NameTag, Text, Type         string
	Owner                       EntityID
	Damage                      float64
	FuseMilliseconds            uint64
	Experience, Punch, Piercing int32
	Potion                      uint32
	Item                        *ItemStack
	Block                       *WorldBlock
	CustomInstance              uint64
}

// EntitySpawnOptions mirrors world.EntitySpawnOpts for worldless custom entity
// construction. Opaque is plugin-owned state prepared by EntityConfig.Apply.
type EntitySpawnOptions struct {
	Position, Velocity Vec3
	Rotation           Rotation
	ID                 [16]byte
	NameTag, Type      string
	Plugin, LocalType  uint64
	Opaque             uint64
	FireDuration, Age  time.Duration
}

type EntityState struct {
	Position, Velocity Vec3
	Rotation           Rotation
	World              WorldID
	Type, NameTag      string
	HasVelocity        bool
	HasNameTag         bool
	CanTeleport        bool
}

type ParticleKind uint32

const (
	ParticleFlame ParticleKind = iota
	ParticleDust
	ParticleBlockBreak
	ParticlePunchBlock
	ParticleBlockForceField
	ParticleBoneMeal
	ParticleNote
	ParticleDragonEggTeleport
	ParticleEvaporate
	ParticleWaterDrip
	ParticleLavaDrip
	ParticleLava
	ParticleDustPlume
	ParticleHugeExplosion
	ParticleEndermanTeleport
	ParticleSnowballPoof
	ParticleEggSmash
	ParticleSplash
	ParticleEffect
	ParticleEntityFlame
)

type RGBA struct{ R, G, B, A uint8 }

type WorldParticle struct {
	Kind   ParticleKind
	Data   uint32
	Pitch  int32
	Colour RGBA
	Diff   BlockPos
	Block  *WorldBlock
}

type WorldSound struct {
	Kind    SoundKind
	Data    uint32
	Integer int32
	Flags   uint32
	Scalar  float64
	Block   *WorldBlock
	Item    *ItemStack
}

// Host executes synchronous actions requested by native plugins.
type Host interface {
	SendPlayerText(InvocationID, PlayerID, PlayerTextKind, string) bool
	SendPlayerTitle(InvocationID, PlayerID, PlayerTitle) bool
	SendPlayerScoreboard(InvocationID, PlayerID, PlayerScoreboard) bool
	RemovePlayerScoreboard(InvocationID, PlayerID) bool
	SendPlayerForm(InvocationID, PlayerID, PlayerForm) bool
	ClosePlayerForm(InvocationID, PlayerID) bool
	TransformPlayer(InvocationID, PlayerID, PlayerTransformKind, Vec3, float64, float64) bool
	TransferPlayer(InvocationID, PlayerID, WorldID, Vec3) bool
	PlayerKinematics(InvocationID, PlayerID) (PlayerKinematics, bool)
	SetPlayerState(InvocationID, PlayerID, PlayerStateKind, PlayerStateValue) bool
	PlayerState(InvocationID, PlayerID, PlayerStateKind) (PlayerStateValue, bool)
	PlayerAction(InvocationID, PlayerID, PlayerActionKind, PlayerStateValue) (PlayerStateValue, bool)
	PlayerString(InvocationID, PlayerID, PlayerStringKind) (string, bool)
	SendPlayerToast(InvocationID, PlayerID, string, string) bool
	HealPlayer(InvocationID, PlayerID, float64, HealingSource) (float64, bool)
	HurtPlayer(InvocationID, PlayerID, float64, DamageSource) (PlayerHurtResult, bool)
	ChangePlayerEffect(InvocationID, PlayerID, PlayerEffectOperation, PlayerEffect) bool
	PlayerEffects(InvocationID, PlayerID) ([]PlayerEffect, bool)
	ClearPlayerEffects(InvocationID, PlayerID) bool
	SetPlayerEntityVisible(InvocationID, PlayerID, EntityID, bool) bool
	PlayerSkin(InvocationID, PlayerID) (PlayerSkin, bool)
	SetPlayerSkin(InvocationID, PlayerID, PlayerSkin) bool
	InventorySize(InvocationID, InventoryID) (uint32, bool)
	InventoryItem(InvocationID, InventoryID, uint32) (ItemStack, bool)
	SetInventoryItem(InvocationID, InventoryID, uint32, ItemStack) bool
	AddInventoryItem(InvocationID, InventoryID, ItemStack) (uint32, bool)
	ClearInventory(InvocationID, InventoryID) bool
	HeldItem(InvocationID, PlayerID, uint32) (ItemStack, bool)
	HeldItems(InvocationID, PlayerID) (ItemStack, ItemStack, bool)
	SetHeldItems(InvocationID, PlayerID, ItemStack, ItemStack) bool
	SetHeldSlot(InvocationID, PlayerID, uint32) bool
	CurrentWorld(InvocationID) (WorldID, bool)
	WorldName(InvocationID, WorldID) (string, bool)
	CreateWorld(WorldConfig) (WorldID, bool)
	UnloadWorld(InvocationID, WorldID) bool
	BlockByName(string, []byte) (WorldBlock, bool)
	WorldBlock(InvocationID, WorldID, BlockPos) (WorldBlock, bool)
	WorldBlockLoaded(InvocationID, WorldID, BlockPos) (WorldBlock, bool, bool)
	OpenWorldBlocksWithin(InvocationID, WorldID, BlockPos, int32, []WorldBlock) (BlockIteratorID, bool)
	NextWorldBlock(InvocationID, BlockIteratorID) (BlockPos, bool, bool)
	CloseWorldBlocks(InvocationID, BlockIteratorID)
	WorldLiquid(InvocationID, WorldID, BlockPos) (WorldBlock, bool, bool)
	SetWorldLiquid(InvocationID, WorldID, BlockPos, *WorldBlock) bool
	SetWorldBlock(InvocationID, WorldID, BlockPos, WorldBlock, WorldSetOpts) bool
	ScheduleWorldBlockUpdate(InvocationID, WorldID, BlockPos, WorldBlock, int64) bool
	WorldBiome(InvocationID, WorldID, BlockPos) (int32, bool)
	SetWorldBiome(InvocationID, WorldID, BlockPos, int32) bool
	WorldTemperature(InvocationID, WorldID, BlockPos) (float64, bool)
	WorldRainingAt(InvocationID, WorldID, BlockPos) (bool, bool)
	WorldSnowingAt(InvocationID, WorldID, BlockPos) (bool, bool)
	WorldThunderingAt(InvocationID, WorldID, BlockPos) (bool, bool)
	WorldRaining(InvocationID, WorldID) (bool, bool)
	WorldThundering(InvocationID, WorldID) (bool, bool)
	WorldCurrentTick(InvocationID, WorldID) (int64, bool)
	WorldRange(InvocationID, WorldID) (BlockRange, bool)
	WorldHighestLightBlocker(InvocationID, WorldID, int32, int32) (int32, bool)
	WorldHighestBlock(InvocationID, WorldID, int32, int32) (int32, bool)
	WorldLight(InvocationID, WorldID, BlockPos) (uint8, bool)
	WorldSkyLight(InvocationID, WorldID, BlockPos) (uint8, bool)
	WorldTime(InvocationID, WorldID) (int64, bool)
	SetWorldTime(InvocationID, WorldID, int64) bool
	WorldSpawn(InvocationID, WorldID) (BlockPos, bool)
	SetWorldSpawn(InvocationID, WorldID, BlockPos) bool
	WorldPlayerSpawn(InvocationID, WorldID, [16]byte) (BlockPos, bool)
	SetWorldPlayerSpawn(InvocationID, WorldID, [16]byte, BlockPos) bool
	WorldDimension(InvocationID, WorldID) (WorldDimension, bool)
	WorldTimeCycle(InvocationID, WorldID) (bool, bool)
	SetWorldTimeCycle(InvocationID, WorldID, bool) bool
	SetWorldRequiredSleepDuration(InvocationID, WorldID, time.Duration) bool
	WorldDefaultGameMode(InvocationID, WorldID) (int64, bool)
	SetWorldDefaultGameMode(InvocationID, WorldID, int64) bool
	SetWorldTickRange(InvocationID, WorldID, int32) bool
	WorldDifficulty(InvocationID, WorldID) (DifficultyView, bool)
	SetWorldDifficulty(InvocationID, WorldID, DifficultyView) bool
	SaveWorld(InvocationID, WorldID) bool
	SpawnWorldEntity(InvocationID, WorldID, EntitySpawn) (EntityID, bool)
	OpenWorldEntityIterator(InvocationID, WorldID, bool) (EntityIteratorID, bool)
	OpenWorldEntitiesWithin(InvocationID, WorldID, BBox) (EntityIteratorID, bool)
	NextWorldEntity(InvocationID, EntityIteratorID) (EntityID, bool, bool)
	CloseWorldEntities(InvocationID, EntityIteratorID)
	OpenServerPlayerIterator(InvocationID) (PlayerIteratorID, bool)
	NextServerPlayer(InvocationID, PlayerIteratorID) (InvocationID, PlayerSnapshot, bool, bool)
	CloseServerPlayers(InvocationID, PlayerIteratorID)
	ServerMaxPlayerCount() (int64, bool)
	ServerPlayerCount() (int64, bool)
	ServerPlayer([16]byte) (EntityHandleID, bool, bool)
	ServerPlayerByName(string) (EntityHandleID, bool, bool)
	ServerPlayerByXUID(string) (EntityHandleID, bool, bool)
	PlayerXUID(InvocationID, PlayerID) (string, bool)
	ServerWorld(WorldDimension) (WorldID, bool)
	ScheduleWorld(WorldID, uint64, uint64, int64) bool
	CancelWorldTask(uint64, uint64) (bool, bool)
	EntityHandle(InvocationID, EntityID) (EntityHandleID, bool)
	EntityHandleEntity(InvocationID, EntityHandleID) (EntityID, bool, bool)
	EntityHandleUUID(EntityHandleID) ([16]byte, bool)
	EntityHandleClosed(EntityHandleID) (bool, bool)
	CloseEntityHandle(EntityHandleID) bool
	RemoveEntity(InvocationID, EntityID) (EntityHandleID, bool)
	AddEntity(InvocationID, EntityHandleID, *Vec3) (EntityID, bool)
	EntityState(InvocationID, EntityID) (EntityState, bool)
	EntityPlayer(InvocationID, EntityID) (PlayerSnapshot, bool)
	TeleportEntity(InvocationID, EntityID, Vec3) bool
	SetEntityVelocity(InvocationID, EntityID, Vec3) bool
	SetEntityNameTag(InvocationID, EntityID, string) bool
	DespawnEntity(InvocationID, EntityID) bool
	AddWorldParticle(InvocationID, WorldID, Vec3, WorldParticle) bool
	PlayWorldSound(InvocationID, WorldID, Vec3, WorldSound) bool
	PlayPlayerSound(InvocationID, PlayerID, WorldSound) bool
}

type noopHost struct{}

func (noopHost) SendPlayerText(InvocationID, PlayerID, PlayerTextKind, string) bool { return false }
func (noopHost) SendPlayerTitle(InvocationID, PlayerID, PlayerTitle) bool           { return false }
func (noopHost) SendPlayerScoreboard(InvocationID, PlayerID, PlayerScoreboard) bool { return false }
func (noopHost) RemovePlayerScoreboard(InvocationID, PlayerID) bool                 { return false }
func (noopHost) SendPlayerForm(InvocationID, PlayerID, PlayerForm) bool             { return false }
func (noopHost) ClosePlayerForm(InvocationID, PlayerID) bool                        { return false }
func (noopHost) TransformPlayer(InvocationID, PlayerID, PlayerTransformKind, Vec3, float64, float64) bool {
	return false
}
func (noopHost) TransferPlayer(InvocationID, PlayerID, WorldID, Vec3) bool { return false }
func (noopHost) PlayerKinematics(InvocationID, PlayerID) (PlayerKinematics, bool) {
	return PlayerKinematics{}, false
}
func (noopHost) SetPlayerState(InvocationID, PlayerID, PlayerStateKind, PlayerStateValue) bool {
	return false
}
func (noopHost) PlayerState(InvocationID, PlayerID, PlayerStateKind) (PlayerStateValue, bool) {
	return PlayerStateValue{}, false
}
func (noopHost) PlayerAction(InvocationID, PlayerID, PlayerActionKind, PlayerStateValue) (PlayerStateValue, bool) {
	return PlayerStateValue{}, false
}
func (noopHost) PlayerString(InvocationID, PlayerID, PlayerStringKind) (string, bool) {
	return "", false
}
func (noopHost) SendPlayerToast(InvocationID, PlayerID, string, string) bool { return false }
func (noopHost) HealPlayer(InvocationID, PlayerID, float64, HealingSource) (float64, bool) {
	return 0, false
}
func (noopHost) HurtPlayer(InvocationID, PlayerID, float64, DamageSource) (PlayerHurtResult, bool) {
	return PlayerHurtResult{}, false
}
func (noopHost) ChangePlayerEffect(InvocationID, PlayerID, PlayerEffectOperation, PlayerEffect) bool {
	return false
}
func (noopHost) PlayerEffects(InvocationID, PlayerID) ([]PlayerEffect, bool)        { return nil, false }
func (noopHost) ClearPlayerEffects(InvocationID, PlayerID) bool                     { return false }
func (noopHost) SetPlayerEntityVisible(InvocationID, PlayerID, EntityID, bool) bool { return false }
func (noopHost) PlayerSkin(InvocationID, PlayerID) (PlayerSkin, bool)               { return PlayerSkin{}, false }
func (noopHost) SetPlayerSkin(InvocationID, PlayerID, PlayerSkin) bool              { return false }
func (noopHost) InventorySize(InvocationID, InventoryID) (uint32, bool)             { return 0, false }
func (noopHost) InventoryItem(InvocationID, InventoryID, uint32) (ItemStack, bool) {
	return ItemStack{}, false
}
func (noopHost) SetInventoryItem(InvocationID, InventoryID, uint32, ItemStack) bool { return false }
func (noopHost) AddInventoryItem(InvocationID, InventoryID, ItemStack) (uint32, bool) {
	return 0, false
}
func (noopHost) ClearInventory(InvocationID, InventoryID) bool             { return false }
func (noopHost) HeldItem(InvocationID, PlayerID, uint32) (ItemStack, bool) { return ItemStack{}, false }
func (noopHost) HeldItems(InvocationID, PlayerID) (ItemStack, ItemStack, bool) {
	return ItemStack{}, ItemStack{}, false
}
func (noopHost) SetHeldItems(InvocationID, PlayerID, ItemStack, ItemStack) bool { return false }
func (noopHost) SetHeldSlot(InvocationID, PlayerID, uint32) bool                { return false }
func (noopHost) CurrentWorld(InvocationID) (WorldID, bool)                      { return 0, false }
func (noopHost) WorldName(InvocationID, WorldID) (string, bool)                 { return "", false }
func (noopHost) CreateWorld(WorldConfig) (WorldID, bool)                        { return 0, false }
func (noopHost) UnloadWorld(InvocationID, WorldID) bool                         { return false }
func (noopHost) BlockByName(string, []byte) (WorldBlock, bool)                  { return WorldBlock{}, false }
func (noopHost) WorldBlock(InvocationID, WorldID, BlockPos) (WorldBlock, bool) {
	return WorldBlock{}, false
}
func (noopHost) WorldBlockLoaded(InvocationID, WorldID, BlockPos) (WorldBlock, bool, bool) {
	return WorldBlock{}, false, false
}
func (noopHost) OpenWorldBlocksWithin(InvocationID, WorldID, BlockPos, int32, []WorldBlock) (BlockIteratorID, bool) {
	return 0, false
}
func (noopHost) NextWorldBlock(InvocationID, BlockIteratorID) (BlockPos, bool, bool) {
	return BlockPos{}, false, false
}
func (noopHost) CloseWorldBlocks(InvocationID, BlockIteratorID) {}
func (noopHost) WorldLiquid(InvocationID, WorldID, BlockPos) (WorldBlock, bool, bool) {
	return WorldBlock{}, false, false
}
func (noopHost) SetWorldLiquid(InvocationID, WorldID, BlockPos, *WorldBlock) bool { return false }
func (noopHost) SetWorldBlock(InvocationID, WorldID, BlockPos, WorldBlock, WorldSetOpts) bool {
	return false
}
func (noopHost) ScheduleWorldBlockUpdate(InvocationID, WorldID, BlockPos, WorldBlock, int64) bool {
	return false
}
func (noopHost) WorldBiome(InvocationID, WorldID, BlockPos) (int32, bool) {
	return 0, false
}
func (noopHost) SetWorldBiome(InvocationID, WorldID, BlockPos, int32) bool { return false }
func (noopHost) WorldTemperature(InvocationID, WorldID, BlockPos) (float64, bool) {
	return 0, false
}
func (noopHost) WorldRainingAt(InvocationID, WorldID, BlockPos) (bool, bool) {
	return false, false
}
func (noopHost) WorldSnowingAt(InvocationID, WorldID, BlockPos) (bool, bool) {
	return false, false
}
func (noopHost) WorldThunderingAt(InvocationID, WorldID, BlockPos) (bool, bool) {
	return false, false
}
func (noopHost) WorldRaining(InvocationID, WorldID) (bool, bool) {
	return false, false
}
func (noopHost) WorldThundering(InvocationID, WorldID) (bool, bool) {
	return false, false
}
func (noopHost) WorldCurrentTick(InvocationID, WorldID) (int64, bool) { return 0, false }
func (noopHost) WorldRange(InvocationID, WorldID) (BlockRange, bool)  { return BlockRange{}, false }
func (noopHost) WorldHighestLightBlocker(InvocationID, WorldID, int32, int32) (int32, bool) {
	return 0, false
}
func (noopHost) WorldHighestBlock(InvocationID, WorldID, int32, int32) (int32, bool) {
	return 0, false
}
func (noopHost) WorldLight(InvocationID, WorldID, BlockPos) (uint8, bool)    { return 0, false }
func (noopHost) WorldSkyLight(InvocationID, WorldID, BlockPos) (uint8, bool) { return 0, false }
func (noopHost) WorldTime(InvocationID, WorldID) (int64, bool)               { return 0, false }
func (noopHost) SetWorldTime(InvocationID, WorldID, int64) bool              { return false }
func (noopHost) WorldSpawn(InvocationID, WorldID) (BlockPos, bool)           { return BlockPos{}, false }
func (noopHost) SetWorldSpawn(InvocationID, WorldID, BlockPos) bool          { return false }
func (noopHost) WorldPlayerSpawn(InvocationID, WorldID, [16]byte) (BlockPos, bool) {
	return BlockPos{}, false
}
func (noopHost) SetWorldPlayerSpawn(InvocationID, WorldID, [16]byte, BlockPos) bool {
	return false
}
func (noopHost) WorldDimension(InvocationID, WorldID) (WorldDimension, bool) {
	return 0, false
}
func (noopHost) WorldTimeCycle(InvocationID, WorldID) (bool, bool)  { return false, false }
func (noopHost) SetWorldTimeCycle(InvocationID, WorldID, bool) bool { return false }
func (noopHost) SetWorldRequiredSleepDuration(InvocationID, WorldID, time.Duration) bool {
	return false
}
func (noopHost) WorldDefaultGameMode(InvocationID, WorldID) (int64, bool)  { return 0, false }
func (noopHost) SetWorldDefaultGameMode(InvocationID, WorldID, int64) bool { return false }
func (noopHost) SetWorldTickRange(InvocationID, WorldID, int32) bool       { return false }
func (noopHost) WorldDifficulty(InvocationID, WorldID) (DifficultyView, bool) {
	return DifficultyView{}, false
}
func (noopHost) SetWorldDifficulty(InvocationID, WorldID, DifficultyView) bool { return false }
func (noopHost) SaveWorld(InvocationID, WorldID) bool                          { return false }
func (noopHost) SpawnWorldEntity(InvocationID, WorldID, EntitySpawn) (EntityID, bool) {
	return EntityID{}, false
}
func (noopHost) OpenWorldEntityIterator(InvocationID, WorldID, bool) (EntityIteratorID, bool) {
	return 0, false
}
func (noopHost) OpenWorldEntitiesWithin(InvocationID, WorldID, BBox) (EntityIteratorID, bool) {
	return 0, false
}
func (noopHost) NextWorldEntity(InvocationID, EntityIteratorID) (EntityID, bool, bool) {
	return EntityID{}, false, false
}
func (noopHost) CloseWorldEntities(InvocationID, EntityIteratorID) {}
func (noopHost) OpenServerPlayerIterator(InvocationID) (PlayerIteratorID, bool) {
	return 0, false
}
func (noopHost) NextServerPlayer(InvocationID, PlayerIteratorID) (InvocationID, PlayerSnapshot, bool, bool) {
	return 0, PlayerSnapshot{}, false, false
}
func (noopHost) CloseServerPlayers(InvocationID, PlayerIteratorID) {}
func (noopHost) ServerMaxPlayerCount() (int64, bool)               { return 0, false }
func (noopHost) ServerPlayerCount() (int64, bool)                  { return 0, false }
func (noopHost) ServerPlayer([16]byte) (EntityHandleID, bool, bool) {
	return EntityHandleID{}, false, false
}
func (noopHost) ServerPlayerByName(string) (EntityHandleID, bool, bool) {
	return EntityHandleID{}, false, false
}
func (noopHost) ServerPlayerByXUID(string) (EntityHandleID, bool, bool) {
	return EntityHandleID{}, false, false
}
func (noopHost) PlayerXUID(InvocationID, PlayerID) (string, bool)  { return "", false }
func (noopHost) ServerWorld(WorldDimension) (WorldID, bool)        { return 0, false }
func (noopHost) ScheduleWorld(WorldID, uint64, uint64, int64) bool { return false }
func (noopHost) CancelWorldTask(uint64, uint64) (bool, bool)       { return false, false }
func (noopHost) EntityHandle(InvocationID, EntityID) (EntityHandleID, bool) {
	return EntityHandleID{}, false
}
func (noopHost) EntityHandleEntity(InvocationID, EntityHandleID) (EntityID, bool, bool) {
	return EntityID{}, false, false
}
func (noopHost) EntityHandleUUID(EntityHandleID) ([16]byte, bool) { return [16]byte{}, false }
func (noopHost) EntityHandleClosed(EntityHandleID) (bool, bool)   { return false, false }
func (noopHost) CloseEntityHandle(EntityHandleID) bool            { return false }
func (noopHost) RemoveEntity(InvocationID, EntityID) (EntityHandleID, bool) {
	return EntityHandleID{}, false
}
func (noopHost) AddEntity(InvocationID, EntityHandleID, *Vec3) (EntityID, bool) {
	return EntityID{}, false
}
func (noopHost) EntityState(InvocationID, EntityID) (EntityState, bool) {
	return EntityState{}, false
}
func (noopHost) EntityPlayer(InvocationID, EntityID) (PlayerSnapshot, bool) {
	return PlayerSnapshot{}, false
}
func (noopHost) TeleportEntity(InvocationID, EntityID, Vec3) bool     { return false }
func (noopHost) SetEntityVelocity(InvocationID, EntityID, Vec3) bool  { return false }
func (noopHost) SetEntityNameTag(InvocationID, EntityID, string) bool { return false }
func (noopHost) DespawnEntity(InvocationID, EntityID) bool            { return false }
func (noopHost) AddWorldParticle(InvocationID, WorldID, Vec3, WorldParticle) bool {
	return false
}
func (noopHost) PlayWorldSound(InvocationID, WorldID, Vec3, WorldSound) bool { return false }
func (noopHost) PlayPlayerSound(InvocationID, PlayerID, WorldSound) bool     { return false }

var (
	hostSequence         atomic.Uint64
	hosts                sync.Map
	skinSnapshotSequence atomic.Uint64
	skinSnapshotMu       sync.Mutex
	skinSnapshots        = map[uint64]skinSnapshot{}
	skinSnapshotCounts   = map[uint64]int{}
	itemSnapshotSequence atomic.Uint64
	itemSnapshotMu       sync.Mutex
	itemSnapshots        = map[uint64]itemSnapshot{}
	itemSnapshotCounts   = map[uint64]int{}
	formSequence         atomic.Uint64
	formMu               sync.Mutex
	formCond             = sync.NewCond(&formMu)
	forms                = map[uint64]formRegistration{}
	formHostState        = map[uint64]*formState{}
)

type registeredHost struct {
	host   Host
	active atomic.Bool
}

const maxSkinSnapshotsPerHost = 32
const maxItemSnapshotsPerHost = 64
const maxFormsPerHost = 128
const maxFormsPerPlayer = 16

type formRegistration struct {
	host    uint64
	player  PlayerID
	respond func(InvocationID, PlayerSnapshot, bool, []byte) bool
	drop    func()
}
type formState struct {
	closing, draining bool
	inflight, count   int
	players           map[PlayerID]int
}

func registerForm(host uint64, player PlayerID, respond func(InvocationID, PlayerSnapshot, bool, []byte) bool, drop func()) (uint64, bool) {
	formMu.Lock()
	defer formMu.Unlock()
	state := formHostState[host]
	if state == nil {
		state = &formState{players: map[PlayerID]int{}}
		formHostState[host] = state
	}
	if state.closing || state.draining || state.count >= maxFormsPerHost || state.players[player] >= maxFormsPerPlayer {
		return 0, false
	}
	id := formSequence.Add(1)
	forms[id] = formRegistration{host: host, player: player, respond: respond, drop: drop}
	state.count++
	state.players[player]++
	return id, true
}

func CompletePlayerForm(id uint64, invocation InvocationID, submitter PlayerSnapshot, closed bool, response []byte) bool {
	formMu.Lock()
	registration, ok := forms[id]
	if !ok {
		formMu.Unlock()
		return true
	}
	delete(forms, id)
	state := formHostState[registration.host]
	state.count--
	state.players[registration.player]--
	if state.players[registration.player] == 0 {
		delete(state.players, registration.player)
	}
	state.inflight++
	formMu.Unlock()
	if submitter.Player != registration.player || len(response) > maxFormJSONBytes {
		registration.drop()
		formMu.Lock()
		state.inflight--
		formCond.Broadcast()
		formMu.Unlock()
		return false
	}
	ok = registration.respond(invocation, submitter, closed, response)
	formMu.Lock()
	state.inflight--
	formCond.Broadcast()
	formMu.Unlock()
	return ok
}

func CancelPlayerForms(player PlayerID) {
	cancelMatchingForms(func(r formRegistration) bool { return r.player == player })
}
func CancelPlayerForm(id uint64) {
	var callback func()
	var state *formState
	formMu.Lock()
	if r, ok := forms[id]; ok {
		delete(forms, id)
		state = formHostState[r.host]
		state.count--
		state.players[r.player]--
		if state.players[r.player] == 0 {
			delete(state.players, r.player)
		}
		state.inflight++
		callback = r.drop
	}
	formMu.Unlock()
	if callback != nil {
		callback()
		formMu.Lock()
		state.inflight--
		formCond.Broadcast()
		formMu.Unlock()
	}
}
func cancelMatchingForms(match func(formRegistration) bool) {
	type pendingDrop struct {
		callback func()
		state    *formState
	}
	var callbacks []pendingDrop
	formMu.Lock()
	for id, r := range forms {
		if match(r) {
			delete(forms, id)
			s := formHostState[r.host]
			s.count--
			s.players[r.player]--
			if s.players[r.player] == 0 {
				delete(s.players, r.player)
			}
			s.inflight++
			callbacks = append(callbacks, pendingDrop{callback: r.drop, state: s})
		}
	}
	formMu.Unlock()
	for _, callback := range callbacks {
		callback.callback()
		formMu.Lock()
		callback.state.inflight--
		formCond.Broadcast()
		formMu.Unlock()
	}
}

func drainHostForms(host uint64, closing bool) {
	formMu.Lock()
	state := formHostState[host]
	if state == nil {
		if closing {
			formHostState[host] = &formState{closing: true, draining: true, players: map[PlayerID]int{}}
		}
		formMu.Unlock()
		return
	}
	state.draining = true
	if closing {
		state.closing = true
	}
	var callbacks []func()
	for id, r := range forms {
		if r.host == host {
			delete(forms, id)
			callbacks = append(callbacks, r.drop)
		}
	}
	state.count = 0
	state.players = map[PlayerID]int{}
	for state.inflight != 0 {
		formCond.Wait()
	}
	if !closing {
		state.draining = false
	}
	formMu.Unlock()
	for _, callback := range callbacks {
		callback()
	}
}

func activateHostForms(host uint64) {
	formMu.Lock()
	state := formHostState[host]
	if state == nil {
		formHostState[host] = &formState{players: map[PlayerID]int{}}
	} else {
		state.closing = false
		state.draining = false
	}
	formMu.Unlock()
}

type skinSnapshot struct {
	host       uint64
	invocation InvocationID
	skin       PlayerSkin
	eventOwned bool
}

type itemSnapshot struct {
	host uint64
	item ItemStack
}

func registerHost(host Host) uint64 {
	return registerHostState(host, true)
}

func registerInactiveHost(host Host) uint64 {
	return registerHostState(host, false)
}

func registerHostState(host Host, active bool) uint64 {
	if host == nil {
		host = noopHost{}
	}
	id := hostSequence.Add(1)
	registered := &registeredHost{host: host}
	registered.active.Store(active)
	hosts.Store(id, registered)
	return id
}

func setHostActive(id uint64, active bool) bool {
	value, ok := hosts.Load(id)
	if !ok {
		return false
	}
	value.(*registeredHost).active.Store(active)
	return true
}

func unregisterHost(id uint64) {
	if id != 0 {
		drainHostForms(id, true)
		formMu.Lock()
		delete(formHostState, id)
		formMu.Unlock()
		hosts.Delete(id)
		skinSnapshotMu.Lock()
		for snapshotID, snapshot := range skinSnapshots {
			if snapshot.host == id {
				delete(skinSnapshots, snapshotID)
			}
		}
		delete(skinSnapshotCounts, id)
		skinSnapshotMu.Unlock()
		itemSnapshotMu.Lock()
		for snapshotID, snapshot := range itemSnapshots {
			if snapshot.host == id {
				delete(itemSnapshots, snapshotID)
			}
		}
		delete(itemSnapshotCounts, id)
		itemSnapshotMu.Unlock()
	}
}

func resolveHost(id uint64) (Host, bool) {
	value, ok := hosts.Load(id)
	if !ok {
		return nil, false
	}
	host := value.(*registeredHost)
	if !host.active.Load() {
		return nil, false
	}
	return host.host, true
}

func registerSkinSnapshot(host uint64, invocation InvocationID, skin PlayerSkin, eventOwned bool) (uint64, bool) {
	skinSnapshotMu.Lock()
	defer skinSnapshotMu.Unlock()
	if skinSnapshotCounts[host] >= maxSkinSnapshotsPerHost {
		return 0, false
	}
	id := skinSnapshotSequence.Add(1)
	skinSnapshots[id] = skinSnapshot{host: host, invocation: invocation, skin: clonePlayerSkin(skin), eventOwned: eventOwned}
	skinSnapshotCounts[host]++
	return id, true
}

func resolveSkinSnapshot(host uint64, invocation InvocationID, id uint64) (PlayerSkin, bool) {
	skinSnapshotMu.Lock()
	defer skinSnapshotMu.Unlock()
	value, ok := skinSnapshots[id]
	if !ok || value.host != host || value.invocation != invocation {
		return PlayerSkin{}, false
	}
	return clonePlayerSkin(value.skin), true
}

func replaceEventSkinSnapshot(host uint64, invocation InvocationID, id uint64, skin PlayerSkin) bool {
	skinSnapshotMu.Lock()
	defer skinSnapshotMu.Unlock()
	value, ok := skinSnapshots[id]
	if !ok || value.host != host || value.invocation != invocation || !value.eventOwned {
		return false
	}
	value.skin = clonePlayerSkin(skin)
	skinSnapshots[id] = value
	return true
}

func unregisterSkinSnapshot(host uint64, invocation InvocationID, id uint64) {
	skinSnapshotMu.Lock()
	defer skinSnapshotMu.Unlock()
	value, ok := skinSnapshots[id]
	if !ok || value.host != host || value.invocation != invocation || value.eventOwned {
		return
	}
	deleteSkinSnapshotLocked(host, id)
}

func forceUnregisterSkinSnapshot(host uint64, invocation InvocationID, id uint64) {
	skinSnapshotMu.Lock()
	defer skinSnapshotMu.Unlock()
	value, ok := skinSnapshots[id]
	if !ok || value.host != host || value.invocation != invocation {
		return
	}
	deleteSkinSnapshotLocked(host, id)
}

func deleteSkinSnapshotLocked(host, id uint64) {
	delete(skinSnapshots, id)
	skinSnapshotCounts[host]--
	if skinSnapshotCounts[host] == 0 {
		delete(skinSnapshotCounts, host)
	}
}

func clonePlayerSkin(value PlayerSkin) PlayerSkin {
	value.Pixels = append([]byte(nil), value.Pixels...)
	value.Model = append([]byte(nil), value.Model...)
	value.CapePixels = append([]byte(nil), value.CapePixels...)
	value.Animations = append([]SkinAnimation(nil), value.Animations...)
	for index := range value.Animations {
		value.Animations[index].Pixels = append([]byte(nil), value.Animations[index].Pixels...)
	}
	return value
}

func registerItemSnapshot(host uint64, item ItemStack) (uint64, bool) {
	itemSnapshotMu.Lock()
	defer itemSnapshotMu.Unlock()
	if itemSnapshotCounts[host] >= maxItemSnapshotsPerHost {
		return 0, false
	}
	id := itemSnapshotSequence.Add(1)
	itemSnapshots[id] = itemSnapshot{host: host, item: cloneItemStack(item)}
	itemSnapshotCounts[host]++
	return id, true
}

func registerItemSnapshotPair(host uint64, first, second ItemStack) (uint64, uint64, bool) {
	itemSnapshotMu.Lock()
	defer itemSnapshotMu.Unlock()
	if itemSnapshotCounts[host] > maxItemSnapshotsPerHost-2 {
		return 0, 0, false
	}
	firstID := itemSnapshotSequence.Add(1)
	secondID := itemSnapshotSequence.Add(1)
	itemSnapshots[firstID] = itemSnapshot{host: host, item: cloneItemStack(first)}
	itemSnapshots[secondID] = itemSnapshot{host: host, item: cloneItemStack(second)}
	itemSnapshotCounts[host] += 2
	return firstID, secondID, true
}

func resolveItemSnapshot(host, id uint64) (ItemStack, bool) {
	itemSnapshotMu.Lock()
	defer itemSnapshotMu.Unlock()
	value, ok := itemSnapshots[id]
	if !ok || value.host != host {
		return ItemStack{}, false
	}
	return value.item, true
}

func unregisterItemSnapshot(host, id uint64) {
	itemSnapshotMu.Lock()
	defer itemSnapshotMu.Unlock()
	value, ok := itemSnapshots[id]
	if ok && value.host == host {
		delete(itemSnapshots, id)
		itemSnapshotCounts[host]--
		if itemSnapshotCounts[host] == 0 {
			delete(itemSnapshotCounts, host)
		}
	}
}

func cloneItemStack(value ItemStack) ItemStack {
	value.Lore = append([]string(nil), value.Lore...)
	value.NBT = append([]byte(nil), value.NBT...)
	value.ValuesNBT = append([]byte(nil), value.ValuesNBT...)
	value.Enchantments = append([]ItemEnchantment(nil), value.Enchantments...)
	return value
}
