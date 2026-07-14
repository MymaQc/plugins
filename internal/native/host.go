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

type PlayerTransformKind uint32

const (
	PlayerTransformTeleport PlayerTransformKind = iota
	PlayerTransformMove
	PlayerTransformVelocity
)

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

type WorldDimension uint32

const (
	WorldDimensionOverworld WorldDimension = iota
	WorldDimensionNether
	WorldDimensionEnd
)

type WorldOpenMode uint32

const (
	WorldOpenOrCreate WorldOpenMode = iota
	WorldOpenExisting
	WorldCreateNew
)

type WorldSavePolicy uint32

const (
	WorldSaveAutomatic WorldSavePolicy = iota
	WorldSaveManual
)

type WorldRandomTickPolicy uint32

const (
	WorldRandomTicksDisabled WorldRandomTickPolicy = iota
	WorldRandomTicksPerSubchunk
)

type WorldTimePolicy uint32

const (
	WorldTimePreserve WorldTimePolicy = iota
	WorldTimeCycle
	WorldTimeFixed
)

type WorldWeatherPolicy uint32

const (
	WorldWeatherPreserve WorldWeatherPolicy = iota
	WorldWeatherCycle
	WorldWeatherClear
)

type WorldChunkUnloadPolicy uint32

const WorldChunkUnloadAfter WorldChunkUnloadPolicy = 0

type WorldOpenSpec struct {
	ProviderPath     string
	Dimension        WorldDimension
	OpenMode         WorldOpenMode
	ReadOnly         bool
	Save             WorldSavePolicy
	SaveInterval     time.Duration
	RandomTicks      WorldRandomTickPolicy
	RandomTickRate   uint32
	Time             WorldTimePolicy
	FixedTime        int64
	Weather          WorldWeatherPolicy
	ChunkUnload      WorldChunkUnloadPolicy
	ChunkUnloadAfter time.Duration
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

type WorldSetOpts struct {
	DisableBlockUpdates       bool
	DisableLiquidDisplacement bool
	DisableRedstoneUpdates    bool
}

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
	Min, Max          Vec3
	TypeKey           uint64
	Family            EntityFamily
	CallbackFlags     uint32
	InitialHealth     float64
	MaxHealth         float64
	Speed             float64
	StateVersion      uint32
	Physics           *EntityPhysics
}

type EntityFamily uint32

const (
	EntityFamilyBase EntityFamily = iota
	EntityFamilyTicking
	EntityFamilyLiving
)

const (
	EntityCallbackState uint32 = 1 << iota
	EntityCallbackTick
	EntityCallbackHurt
	EntityCallbackHeal
	EntityCallbackDeath
)

type EntityPhysics struct {
	Gravity, Drag     float64
	DragBeforeGravity bool
}

// EntityInstanceID identifies plugin-owned state for one custom entity.
// It is opaque outside the native runtime and is never reused by that runtime.
type EntityInstanceID uint64

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
	generatedPlayerOperationsHost
	SendPlayerText(InvocationID, PlayerID, PlayerTextKind, string) bool
	SendPlayerTitle(InvocationID, PlayerID, PlayerTitle) bool
	SendPlayerScoreboard(InvocationID, PlayerID, PlayerScoreboard) bool
	RemovePlayerScoreboard(InvocationID, PlayerID) bool
	SendPlayerForm(InvocationID, PlayerID, PlayerForm) bool
	ClosePlayerForm(InvocationID, PlayerID) bool
	TransformPlayer(InvocationID, PlayerID, PlayerTransformKind, Vec3, float64, float64) bool
	TransferPlayer(InvocationID, PlayerID, WorldID, Vec3) bool
	PlayerRotation(InvocationID, PlayerID) (Rotation, bool)
	SetPlayerState(InvocationID, PlayerID, PlayerStateKind, PlayerStateValue) bool
	PlayerState(InvocationID, PlayerID, PlayerStateKind) (PlayerStateValue, bool)
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
	SetHeldItems(InvocationID, PlayerID, ItemStack, ItemStack) bool
	SetHeldSlot(InvocationID, PlayerID, uint32) bool
	WorldByName(InvocationID, string) (WorldID, bool)
	WorldName(InvocationID, WorldID) (string, bool)
	OpenWorld(InvocationID, string, WorldDimension) (WorldID, bool)
	OpenWorldSpec(InvocationID, string, WorldOpenSpec) (WorldID, bool)
	UnloadWorld(InvocationID, WorldID) bool
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
	WorldRange(InvocationID, WorldID) (BlockRange, bool)
	WorldHighestLightBlocker(InvocationID, WorldID, int32, int32) (int32, bool)
	WorldHighestBlock(InvocationID, WorldID, int32, int32) (int32, bool)
	WorldLight(InvocationID, WorldID, BlockPos) (uint8, bool)
	WorldSkyLight(InvocationID, WorldID, BlockPos) (uint8, bool)
	WorldTime(InvocationID, WorldID) (int64, bool)
	SetWorldTime(InvocationID, WorldID, int64) bool
	WorldSpawn(InvocationID, WorldID) (BlockPos, bool)
	SetWorldSpawn(InvocationID, WorldID, BlockPos) bool
	SaveWorld(InvocationID, WorldID) bool
	SpawnWorldEntity(InvocationID, WorldID, EntitySpawn) (EntityID, bool)
	WorldEntities(InvocationID, WorldID) ([]EntityID, bool)
	WorldPlayers(InvocationID, WorldID) ([]PlayerID, bool)
	EntityState(InvocationID, EntityID) (EntityState, bool)
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
func (noopHost) PlayerRotation(InvocationID, PlayerID) (Rotation, bool)    { return Rotation{}, false }
func (noopHost) SetPlayerState(InvocationID, PlayerID, PlayerStateKind, PlayerStateValue) bool {
	return false
}
func (noopHost) PlayerState(InvocationID, PlayerID, PlayerStateKind) (PlayerStateValue, bool) {
	return PlayerStateValue{}, false
}
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
func (noopHost) ClearInventory(InvocationID, InventoryID) bool                  { return false }
func (noopHost) HeldItem(InvocationID, PlayerID, uint32) (ItemStack, bool)      { return ItemStack{}, false }
func (noopHost) SetHeldItems(InvocationID, PlayerID, ItemStack, ItemStack) bool { return false }
func (noopHost) SetHeldSlot(InvocationID, PlayerID, uint32) bool                { return false }
func (noopHost) WorldByName(InvocationID, string) (WorldID, bool)               { return 0, false }
func (noopHost) WorldName(InvocationID, WorldID) (string, bool)                 { return "", false }
func (noopHost) OpenWorld(InvocationID, string, WorldDimension) (WorldID, bool) { return 0, false }
func (noopHost) OpenWorldSpec(InvocationID, string, WorldOpenSpec) (WorldID, bool) {
	return 0, false
}
func (noopHost) UnloadWorld(InvocationID, WorldID) bool { return false }
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
func (noopHost) WorldRange(InvocationID, WorldID) (BlockRange, bool) { return BlockRange{}, false }
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
func (noopHost) SaveWorld(InvocationID, WorldID) bool                        { return false }
func (noopHost) SpawnWorldEntity(InvocationID, WorldID, EntitySpawn) (EntityID, bool) {
	return EntityID{}, false
}
func (noopHost) WorldEntities(InvocationID, WorldID) ([]EntityID, bool) { return nil, false }
func (noopHost) WorldPlayers(InvocationID, WorldID) ([]PlayerID, bool)  { return nil, false }
func (noopHost) EntityState(InvocationID, EntityID) (EntityState, bool) {
	return EntityState{}, false
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
	respond func(InvocationID, PlayerID, bool, []byte) bool
	drop    func()
}
type formState struct {
	closing, draining bool
	inflight, count   int
	players           map[PlayerID]int
}

func registerForm(host uint64, player PlayerID, respond func(InvocationID, PlayerID, bool, []byte) bool, drop func()) (uint64, bool) {
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

func CompletePlayerForm(id uint64, invocation InvocationID, submitter PlayerID, closed bool, response []byte) bool {
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
	if submitter != registration.player || len(response) > maxFormJSONBytes {
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

func abandonForm(id uint64) {
	formMu.Lock()
	defer formMu.Unlock()
	r, ok := forms[id]
	if !ok {
		return
	}
	delete(forms, id)
	s := formHostState[r.host]
	s.count--
	s.players[r.player]--
	if s.players[r.player] == 0 {
		delete(s.players, r.player)
	}
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
