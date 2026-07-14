// Package native loads and calls the native plugin runtime.
package native

/*
#cgo CFLAGS: -I../../abi/include
#cgo linux LDFLAGS: -ldl
#include <stdlib.h>
#include "bridge.h"

static inline void bg_call_event_drop(DfEventDropFn callback, void *context) {
    if (callback != NULL) callback(context);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"
)

const (
	maxEntityTypes                 = 1024
	maxEntityStateBytes            = 16 << 20
	maxEventStringBytes            = 64 << 10
	maxCommandExecutionStringCount = 1024
	maxCommandExecutionAliasCount  = 1024
	maxTransferIPBytes             = 16
	maxTransferZoneBytes           = 4096
)

const (
	PlayerMoveSubscription             uint64 = 1
	PlayerChatSubscription             uint64 = 2
	PlayerJoinSubscription             uint64 = 4
	PlayerQuitSubscription             uint64 = 8
	PlayerHurtSubscription             uint64 = 16
	PlayerHealSubscription             uint64 = 32
	PlayerBlockBreakSubscription       uint64 = 64
	PlayerBlockPlaceSubscription       uint64 = 128
	PlayerFoodLossSubscription         uint64 = 256
	PlayerDeathSubscription            uint64 = 512
	PlayerStartBreakSubscription       uint64 = 1024
	PlayerFireExtinguishSubscription   uint64 = 2048
	PlayerToggleSprintSubscription     uint64 = 4096
	PlayerToggleSneakSubscription      uint64 = 8192
	PlayerJumpSubscription             uint64 = 16384
	PlayerTeleportSubscription         uint64 = 32768
	PlayerExperienceGainSubscription   uint64 = 65536
	PlayerPunchAirSubscription         uint64 = 131072
	PlayerHeldSlotChangeSubscription   uint64 = 262144
	PlayerSleepSubscription            uint64 = 524288
	PlayerBlockPickSubscription        uint64 = 1048576
	PlayerLecternPageTurnSubscription  uint64 = 2097152
	PlayerSignEditSubscription         uint64 = 4194304
	PlayerItemUseSubscription          uint64 = 8388608
	PlayerItemUseOnBlockSubscription   uint64 = 16777216
	PlayerItemConsumeSubscription      uint64 = 33554432
	PlayerItemReleaseSubscription      uint64 = 67108864
	PlayerItemDamageSubscription       uint64 = 134217728
	PlayerItemDropSubscription         uint64 = 268435456
	PlayerAttackEntitySubscription     uint64 = 536870912
	PlayerItemUseOnEntitySubscription  uint64 = 1073741824
	PlayerChangeWorldSubscription      uint64 = 2147483648
	PlayerRespawnSubscription          uint64 = 4294967296
	PlayerSkinChangeSubscription       uint64 = 8589934592
	PlayerItemPickupSubscription       uint64 = 137438953472
	PlayerTransferSubscription         uint64 = 274877906944
	PlayerCommandExecutionSubscription uint64 = 549755813888
	PlayerDiagnosticsSubscription      uint64 = 1099511627776
	MaxChatReplacementBytes                   = 4096
	MaxCommandOutputBytes                     = 4096
	MaxCommandEnumBytes                       = 4096
)

// InvocationID identifies one synchronous command, event, or form callback.
// Zero means that no Dragonfly owner transaction is attached.
type InvocationID uint64

type PlayerID struct {
	UUID       [16]byte
	Generation uint64
}

type EntityID struct {
	UUID       [16]byte
	Generation uint64
}

type Vec3 struct {
	X, Y, Z float64
}

type BBox struct {
	Min, Max Vec3
}

type Rotation struct {
	Yaw, Pitch float64
}

type BlockPos struct {
	X, Y, Z int32
}

type PlayerMoveInput struct {
	Player      PlayerSnapshot
	OldPosition Vec3
	NewPosition Vec3
	Rotation    Rotation
}

type PlayerChatInput struct {
	Player  PlayerSnapshot
	Message string
}

type PlayerChatOutput struct {
	Cancelled   bool
	Replacement *string
}

type PlayerJoinInput struct {
	Player PlayerSnapshot
	Name   string
}

type PlayerQuitInput struct {
	Player PlayerSnapshot
	Name   string
}

type DamageSource struct {
	Name                                            string
	Kind                                            DamageSourceKind
	ReducedByArmour, ReducedByResistance, Fire      bool
	IgnoresTotem                                    bool
	FireProtection, FeatherFalling, BlastProtection bool
	ProjectileProtection                            bool
	Entity, SecondaryEntity                         EntityID
	Block                                           *WorldBlock
	Data                                            bool
}

type HealingSource struct {
	Name string
	Kind HealingSourceKind
	Data bool
}

type DamageSourceKind uint32

const (
	DamageSourceCustom DamageSourceKind = iota
	DamageSourceAttack
	DamageSourceBlock
	DamageSourceDrowning
	DamageSourceExplosion
	DamageSourceFall
	DamageSourceFire
	DamageSourceGlide
	DamageSourceInstant
	DamageSourceLava
	DamageSourceLightning
	DamageSourceMagma
	DamageSourcePoison
	DamageSourceProjectile
	DamageSourceStarvation
	DamageSourceSuffocation
	DamageSourceThorns
	DamageSourceVoid
	DamageSourceWither
)

type HealingSourceKind uint32

const (
	HealingSourceCustom HealingSourceKind = iota
	HealingSourceFood
	HealingSourceInstant
	HealingSourceRegeneration
)

type PlayerHurtResult struct {
	Damage     float64
	Vulnerable bool
}

type PlayerHurtInput struct {
	Player         PlayerSnapshot
	Damage         float64
	Immune         bool
	AttackImmunity time.Duration
	Source         DamageSource
}

type PlayerHurtOutput struct {
	Cancelled      bool
	Damage         float64
	AttackImmunity time.Duration
}

type PlayerHealInput struct {
	Player PlayerSnapshot
	Health float64
	Source HealingSource
}

type PlayerHealOutput struct {
	Cancelled bool
	Health    float64
}

type PlayerBlockBreakInput struct {
	Player     PlayerSnapshot
	Position   BlockPos
	Block      WorldBlock
	Drops      []ItemStack
	Experience int32
}

type PlayerBlockBreakOutput struct {
	Cancelled  bool
	Drops      []ItemStack
	Experience int32
}

type PlayerBlockPlaceInput struct {
	Player   PlayerSnapshot
	Position BlockPos
	Block    WorldBlock
}

type PlayerFoodLossInput struct {
	Player PlayerSnapshot
	From   int32
	To     int32
}

type PlayerFoodLossOutput struct {
	Cancelled bool
	To        int32
}

type PlayerDeathInput struct {
	Player PlayerSnapshot
	Source DamageSource
}

type PlayerPositionInput struct {
	Player   PlayerSnapshot
	Position BlockPos
}
type PlayerToggleInput struct {
	Player PlayerSnapshot
	After  bool
}
type PlayerTeleportInput struct {
	Player   PlayerSnapshot
	Position Vec3
}
type PlayerExperienceGainOutput struct {
	Cancelled bool
	Amount    int
}
type PlayerHeldSlotChangeInput struct {
	Player PlayerSnapshot
	From   int
	To     int
}
type PlayerSleepOutput struct {
	Cancelled    bool
	SendReminder bool
}
type PlayerBlockPickInput struct {
	Player   PlayerSnapshot
	Position BlockPos
	Block    WorldBlock
}
type PlayerLecternPageTurnInput struct {
	Player   PlayerSnapshot
	Position BlockPos
	OldPage  int
	NewPage  int
}
type PlayerLecternPageTurnOutput struct {
	Cancelled bool
	NewPage   int
}
type PlayerSignEditInput struct {
	Player    PlayerSnapshot
	Position  BlockPos
	FrontSide bool
	OldText   string
	NewText   string
}
type PlayerItemUseOnBlockInput struct {
	Player        PlayerSnapshot
	Position      BlockPos
	Face          int
	ClickPosition Vec3
}
type PlayerItemDamageOutput struct {
	Cancelled bool
	Damage    int
}
type PlayerAttackEntityInput struct {
	Player PlayerSnapshot
	Target EntityID
}
type PlayerAttackEntityOutput struct {
	Cancelled       bool
	KnockbackForce  float64
	KnockbackHeight float64
	Critical        bool
}
type PlayerItemUseOnEntityInput struct {
	Player PlayerSnapshot
	Target EntityID
}
type PlayerChangeWorldInput struct {
	Player PlayerSnapshot
	Before *WorldID
	After  WorldID
}
type PlayerRespawnInput struct {
	Player PlayerSnapshot
}
type PlayerRespawnOutput struct {
	Position Vec3
	World    WorldID
}
type PlayerSkinChangeInput struct {
	Player PlayerSnapshot
}
type PlayerSkinChangeOutput struct {
	Cancelled bool
	Skin      PlayerSkin
}

type UDPAddress struct {
	IP   []byte
	Port int
	Zone string
}

type PlayerTransferInput struct {
	Player  PlayerSnapshot
	Address UDPAddress
}

type PlayerTransferOutput struct {
	Cancelled bool
	Address   UDPAddress
}

type CommandInfo struct {
	Name        string
	Description string
	Usage       string
	Aliases     []string
}

type PlayerCommandExecutionInput struct {
	Player    PlayerSnapshot
	Command   CommandInfo
	Arguments []string
}

type PlayerCommandExecutionOutput struct {
	Cancelled bool
	Arguments []string
}

type PlayerDiagnosticsInput struct {
	Player                        PlayerSnapshot
	AverageFramesPerSecond        float64
	AverageServerSimTickTime      float64
	AverageClientSimTickTime      float64
	AverageBeginFrameTime         float64
	AverageInputTime              float64
	AverageRenderTime             float64
	AverageEndFrameTime           float64
	AverageRemainderTimePercent   float64
	AverageUnaccountedTimePercent float64
}

type PlayerItemPickupInput struct {
	Player PlayerSnapshot
	Item   ItemStack
}

type PlayerItemPickupOutput struct {
	Cancelled bool
	Item      ItemStack
}

type Command struct {
	Index       uint64
	Name        string
	Description string
	Aliases     []string
	Overloads   []CommandOverload
}

type CommandOverload struct {
	Parameters []CommandParameter
}

type CommandParameter struct {
	Kind     CommandParameterKind
	Name     string
	Suffix   string
	Values   []string
	Optional bool
}

type CommandParameterKind uint32

const (
	CommandParameterSubcommand  CommandParameterKind = 1
	CommandParameterEnum        CommandParameterKind = 2
	CommandParameterString      CommandParameterKind = 3
	CommandParameterInteger     CommandParameterKind = 4
	CommandParameterFloat       CommandParameterKind = 5
	CommandParameterBool        CommandParameterKind = 6
	CommandParameterDynamicEnum CommandParameterKind = 7
	CommandParameterPlayer      CommandParameterKind = 8
	CommandParameterRawText     CommandParameterKind = 9
	CommandParameterVector      CommandParameterKind = 10
)

type CommandInput struct {
	Invocation     InvocationID
	Overload       uint64
	Source         string
	Arguments      []string
	SourceKind     CommandSourceKind
	SourcePlayer   *PlayerID
	SourcePosition Vec3
	OnlinePlayers  []CommandPlayer
}

type CommandEnumContext struct {
	Source         string
	SourceKind     CommandSourceKind
	SourcePlayer   *PlayerID
	SourcePosition Vec3
	OnlinePlayers  []CommandPlayer
}

type CommandSourceKind uint32

const (
	CommandSourceUnknown CommandSourceKind = iota
	CommandSourcePlayer
	CommandSourceConsole
)

type CommandPlayer struct {
	Player              PlayerID
	Name                string
	LatencyMilliseconds uint64
	Position            Vec3
}

type CommandOutput struct {
	Failed  bool
	Message string
}

// Runtime owns a loaded C# NativeAOT runtime and its plugin libraries.
// Close must not run concurrently with any other method.
type Runtime struct {
	ptr         *C.BgRuntimeLibrary
	hostContext uint64
}

func Open(runtimeLibrary, pluginDirectory string) (*Runtime, error) {
	return OpenWithHost(runtimeLibrary, pluginDirectory, nil)
}

func OpenWithHost(runtimeLibrary, pluginDirectory string, host Host) (*Runtime, error) {
	if runtimeLibrary == "" || pluginDirectory == "" {
		return nil, errors.New("runtime library and plugin directory are required")
	}
	libraryPath := C.CString(runtimeLibrary)
	defer C.free(unsafe.Pointer(libraryPath))
	pluginPath := C.CString(pluginDirectory)
	defer C.free(unsafe.Pointer(pluginPath))

	hostContext := registerInactiveHost(host)
	var ptr *C.BgRuntimeLibrary
	var errorBuffer [1024]C.uint8_t
	status := C.bg_runtime_open(
		libraryPath,
		pluginPath,
		C.uint64_t(hostContext),
		&ptr,
		&errorBuffer[0],
		C.uint64_t(len(errorBuffer)),
	)
	if status != C.DF_STATUS_OK {
		unregisterHost(hostContext)
		message := C.GoString((*C.char)(unsafe.Pointer(&errorBuffer[0])))
		if message == "" {
			message = "unknown native runtime error"
		}
		return nil, fmt.Errorf("open native runtime: %s", message)
	}
	r := &Runtime{ptr: ptr, hostContext: hostContext}
	runtime.SetFinalizer(r, func(runtime *Runtime) { runtime.Close() })
	return r, nil
}

func (r *Runtime) Close() {
	if r == nil || r.ptr == nil {
		return
	}
	r.BeginDisable()
	r.FinishDisable()
	drainHostForms(r.hostContext, true)
	C.bg_runtime_close(r.ptr)
	unregisterHost(r.hostContext)
	r.ptr = nil
	r.hostContext = 0
	runtime.SetFinalizer(r, nil)
}

func (r *Runtime) Enable() error {
	if r == nil || r.ptr == nil {
		return errors.New("native runtime is closed")
	}
	if !setHostActive(r.hostContext, true) {
		return errors.New("native runtime host is closed")
	}
	activateHostForms(r.hostContext)
	var errorBuffer [4096]C.uint8_t
	if status := C.bg_runtime_enable(
		r.ptr,
		&errorBuffer[0],
		C.uint64_t(len(errorBuffer)),
	); status != C.DF_STATUS_OK {
		message := C.GoString((*C.char)(unsafe.Pointer(&errorBuffer[0])))
		if message == "" {
			message = "native plugins failed to enable"
		}
		return fmt.Errorf("enable native plugins: %s", message)
	}
	return nil
}

// BeginDisable rejects and drains ordinary plugin calls before running on_disable.
// The host remains active so custom worlds may close and dispatch entity callbacks.
func (r *Runtime) BeginDisable() {
	if r != nil && r.ptr != nil {
		drainHostForms(r.hostContext, true)
		C.bg_runtime_begin_disable(r.ptr)
	}
}

// FinishDisable rejects and drains entity callbacks, then deactivates the host.
func (r *Runtime) FinishDisable() {
	if r != nil && r.ptr != nil {
		C.bg_runtime_finish_disable(r.ptr)
		setHostActive(r.hostContext, false)
	}
}

func (r *Runtime) Disable() {
	r.BeginDisable()
	r.FinishDisable()
}

func (r *Runtime) PluginCount() uint64 {
	if r == nil || r.ptr == nil {
		return 0
	}
	return uint64(C.bg_runtime_plugin_count(r.ptr))
}

func (r *Runtime) Subscriptions() uint64 {
	if r == nil || r.ptr == nil {
		return 0
	}
	return uint64(C.bg_runtime_subscriptions(r.ptr))
}

func (r *Runtime) EntityTypes() ([]EntityTypeDefinition, error) {
	if r == nil || r.ptr == nil {
		return nil, errors.New("native runtime is closed")
	}
	count := uint64(C.bg_runtime_entity_type_count(r.ptr))
	if count > maxEntityTypes {
		return nil, fmt.Errorf("native runtime returned too many entity types: %d", count)
	}
	types := make([]EntityTypeDefinition, 0, count)
	for index := uint64(0); index < count; index++ {
		var descriptor C.DfEntityTypeDescriptorV2
		if status := C.bg_runtime_entity_type_at(r.ptr, C.uint64_t(index), &descriptor); status != C.DF_STATUS_OK {
			return nil, fmt.Errorf("read native entity type %d: status %d", index, int32(status))
		}
		definition := EntityTypeDefinition{
			SaveID: stringView(descriptor.save_id), NetworkID: stringView(descriptor.network_id),
			TypeKey: uint64(descriptor.type_key),
		}
		if !validEntityTypeDefinition(definition) {
			return nil, fmt.Errorf("invalid native entity type %d", index)
		}
		types = append(types, definition)
	}
	return types, nil
}

func validEntityTypeDefinition(definition EntityTypeDefinition) bool {
	if definition.TypeKey == 0 ||
		len(definition.SaveID) == 0 || len(definition.SaveID) > maxEntityTypeBytes || !utf8.ValidString(definition.SaveID) ||
		len(definition.NetworkID) == 0 || len(definition.NetworkID) > maxEntityTypeBytes || !utf8.ValidString(definition.NetworkID) {
		return false
	}
	return true
}

func validEntityBounds(minimum, maximum Vec3) bool {
	values := [...]float64{minimum.X, minimum.Y, minimum.Z, maximum.X, maximum.Y, maximum.Z}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return minimum.X <= maximum.X && minimum.Y <= maximum.Y && minimum.Z <= maximum.Z
}

func (r *Runtime) EntityAdopt(typeKey, opaque uint64) (EntityInstanceID, error) {
	if r == nil || r.ptr == nil {
		return 0, errors.New("native runtime is closed")
	}
	var instance C.DfEntityInstanceId
	if status := C.bg_runtime_entity_adopt(r.ptr, C.uint64_t(typeKey), C.uint64_t(opaque), &instance); status != C.DF_STATUS_OK {
		return 0, fmt.Errorf("adopt native entity: status %d", int32(status))
	}
	return EntityInstanceID(instance), nil
}

// EntityAdoptLocal routes plugin-local type identity from a direct host call
// through the runtime's global instance table.
func (r *Runtime) EntityAdoptLocal(plugin, typeKey, opaque uint64) (EntityInstanceID, error) {
	if r == nil || r.ptr == nil {
		return 0, errors.New("native runtime is closed")
	}
	var instance C.DfEntityInstanceId
	if status := C.bg_runtime_entity_adopt_local(r.ptr, C.uint64_t(plugin), C.uint64_t(typeKey), C.uint64_t(opaque), &instance); status != C.DF_STATUS_OK {
		return 0, fmt.Errorf("adopt local native entity: status %d", int32(status))
	}
	return EntityInstanceID(instance), nil
}

func (r *Runtime) EntityLoad(typeKey uint64, input EntityLoadInput) (EntityInstanceID, error) {
	if r == nil || r.ptr == nil {
		return 0, errors.New("native runtime is closed")
	}
	if len(input.Data) > maxEntityStateBytes {
		return 0, errors.New("native entity state is too large")
	}
	data := C.CBytes(input.Data)
	defer C.free(data)
	nativeInput := C.DfEntityLoadInput{
		data:    C.DfBytesView{data: (*C.uint8_t)(data), len: C.uint64_t(len(input.Data))},
		version: C.uint32_t(input.Version),
	}
	var instance C.DfEntityInstanceId
	if status := C.bg_runtime_entity_load(r.ptr, C.uint64_t(typeKey), &nativeInput, &instance); status != C.DF_STATUS_OK {
		return 0, fmt.Errorf("load native entity: status %d", int32(status))
	}
	return EntityInstanceID(instance), nil
}

func (r *Runtime) EntitySave(instance EntityInstanceID) (EntitySaveOutput, error) {
	if r == nil || r.ptr == nil {
		return EntitySaveOutput{}, errors.New("native runtime is closed")
	}
	capacity := 4 << 10
	for attempts := 0; attempts < 2; attempts++ {
		data := C.malloc(C.size_t(capacity))
		if data == nil {
			return EntitySaveOutput{}, errors.New("allocate native entity state")
		}
		state := C.DfEntitySaveState{data: C.DfBytesBuffer{data: (*C.uint8_t)(data), capacity: C.uint64_t(capacity)}}
		status := C.bg_runtime_entity_save(r.ptr, C.DfEntityInstanceId(instance), &state)
		length := uint64(state.data.len)
		if status == C.DF_STATUS_OK && length <= uint64(capacity) {
			output := EntitySaveOutput{Version: uint32(state.version), Data: C.GoBytes(data, C.int(length))}
			C.free(data)
			return output, nil
		}
		C.free(data)
		if length <= uint64(capacity) || length > maxEntityStateBytes {
			return EntitySaveOutput{}, fmt.Errorf("save native entity: status %d", int32(status))
		}
		capacity = int(length)
	}
	return EntitySaveOutput{}, errors.New("save native entity state did not fit")
}

func (r *Runtime) EntityTick(instance EntityInstanceID, input EntityTickInput) (EntityTickOutput, error) {
	output := EntityTickOutput{}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	nativeInput := C.DfEntityTickInput{
		invocation: C.DfInvocationId(input.Invocation), current: C.int64_t(input.Current),
		age_milliseconds: C.uint64_t(max(input.Age.Milliseconds(), 0)),
	}
	fillEntityID(&nativeInput.entity, input.Entity)
	var state C.DfEntityTickState
	if status := C.bg_runtime_entity_tick(r.ptr, C.DfEntityInstanceId(instance), &nativeInput, &state); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("tick native entity: status %d", int32(status))
	}
	output.Despawn = state.despawn != 0
	return output, nil
}

func (r *Runtime) EntityHurt(instance EntityInstanceID, input EntityHurtInput) (EntityHurtOutput, error) {
	output := EntityHurtOutput{Damage: input.Damage}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	source, release, ok := nativeDamageSource(input.Source)
	if !ok {
		return output, errors.New("encode entity damage source")
	}
	defer release()
	nativeInput := C.DfEntityHurtInput{
		invocation: C.DfInvocationId(input.Invocation), source: source,
		health: C.double(input.Health), max_health: C.double(input.MaxHealth),
	}
	fillEntityID(&nativeInput.entity, input.Entity)
	state := C.DfEntityHurtState{damage: C.double(input.Damage)}
	if status := C.bg_runtime_entity_hurt(r.ptr, C.DfEntityInstanceId(instance), &nativeInput, &state); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("hurt native entity: status %d", int32(status))
	}
	output.Damage, output.Cancelled = float64(state.damage), state.cancelled != 0
	return output, nil
}

func (r *Runtime) EntityHeal(instance EntityInstanceID, input EntityHealInput) (EntityHealOutput, error) {
	output := EntityHealOutput{Amount: input.Amount}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	source, release, ok := nativeHealingSource(input.Source)
	if !ok {
		return output, errors.New("encode entity healing source")
	}
	defer release()
	nativeInput := C.DfEntityHealInput{
		invocation: C.DfInvocationId(input.Invocation), source: source,
		health: C.double(input.Health), max_health: C.double(input.MaxHealth),
	}
	fillEntityID(&nativeInput.entity, input.Entity)
	state := C.DfEntityHealState{health: C.double(input.Amount)}
	if status := C.bg_runtime_entity_heal(r.ptr, C.DfEntityInstanceId(instance), &nativeInput, &state); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("heal native entity: status %d", int32(status))
	}
	output.Amount, output.Cancelled = float64(state.health), state.cancelled != 0
	return output, nil
}

func (r *Runtime) EntityDeath(instance EntityInstanceID, input EntityDeathInput) (EntityDeathOutput, error) {
	output := EntityDeathOutput{}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	source, release, ok := nativeDamageSource(input.Source)
	if !ok {
		return output, errors.New("encode entity death source")
	}
	defer release()
	nativeInput := C.DfEntityDeathInput{
		invocation: C.DfInvocationId(input.Invocation), source: source,
		health: C.double(input.Health), damage: C.double(input.Damage),
	}
	fillEntityID(&nativeInput.entity, input.Entity)
	state := C.DfEntityDeathState{}
	if status := C.bg_runtime_entity_death(r.ptr, C.DfEntityInstanceId(instance), &nativeInput, &state); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("death native entity: status %d", int32(status))
	}
	output.Cancelled = state.cancelled != 0
	return output, nil
}

func (r *Runtime) EntityDestroy(instance EntityInstanceID) {
	if r != nil && r.ptr != nil && instance != 0 {
		C.bg_runtime_entity_destroy(r.ptr, C.DfEntityInstanceId(instance))
	}
}

func (r *Runtime) EntityDecodeNBT(typeKey uint64, data EntityCommonData, encoded []byte) (EntityInstanceID, EntityCommonData, error) {
	if r == nil || r.ptr == nil {
		return 0, data, errors.New("native runtime is closed")
	}
	if typeKey == 0 || len(encoded) > maxEntityStateBytes {
		return 0, data, errors.New("invalid native entity decode input")
	}
	encodedData := C.CBytes(encoded)
	defer C.free(encodedData)
	var instance EntityInstanceID
	updated, err := withNativeEntityCommon(data, func(common *C.DfEntityDataState) C.DfStatus {
		input := (*C.DfEntityExactInput)(C.malloc(C.size_t(C.sizeof_DfEntityExactInput)))
		state := (*C.DfEntityExactState)(C.malloc(C.size_t(C.sizeof_DfEntityExactState)))
		if input == nil || state == nil {
			C.free(unsafe.Pointer(input))
			C.free(unsafe.Pointer(state))
			return C.DF_STATUS_ERROR
		}
		defer C.free(unsafe.Pointer(input))
		defer C.free(unsafe.Pointer(state))
		*input = C.DfEntityExactInput{
			data: common,
			nbt:  C.DfBytesView{data: (*C.uint8_t)(encodedData), len: C.uint64_t(len(encoded))},
		}
		*state = C.DfEntityExactState{}
		status := C.bg_runtime_entity_decode_nbt(r.ptr, C.uint64_t(typeKey), input, state)
		instance = EntityInstanceID(state.instance)
		return status
	})
	if err != nil || instance == 0 {
		return 0, data, fmt.Errorf("decode native entity NBT: %w", err)
	}
	return instance, updated, nil
}

func (r *Runtime) EntityEncodeNBT(instance EntityInstanceID, data EntityCommonData) ([]byte, EntityCommonData, error) {
	if r == nil || r.ptr == nil || instance == 0 {
		return nil, data, errors.New("native runtime is closed or entity is invalid")
	}
	capacity := 4 << 10
	for attempts := 0; attempts < 2; attempts++ {
		buffer := C.malloc(C.size_t(capacity))
		if buffer == nil {
			return nil, data, errors.New("allocate native entity NBT")
		}
		var required uint64
		updated, err := withNativeEntityCommon(data, func(common *C.DfEntityDataState) C.DfStatus {
			input := (*C.DfEntityExactInput)(C.malloc(C.size_t(C.sizeof_DfEntityExactInput)))
			state := (*C.DfEntityExactState)(C.malloc(C.size_t(C.sizeof_DfEntityExactState)))
			if input == nil || state == nil {
				C.free(unsafe.Pointer(input))
				C.free(unsafe.Pointer(state))
				return C.DF_STATUS_ERROR
			}
			defer C.free(unsafe.Pointer(input))
			defer C.free(unsafe.Pointer(state))
			*input = C.DfEntityExactInput{data: common}
			*state = C.DfEntityExactState{nbt: C.DfBytesBuffer{data: (*C.uint8_t)(buffer), capacity: C.uint64_t(capacity)}}
			status := C.bg_runtime_entity_call(r.ptr, C.DfEntityInstanceId(instance), C.uint32_t(C.DF_ENTITY_OPERATION_ENCODE_NBT), input, state)
			required = uint64(state.nbt.len)
			return status
		})
		if err == nil && required <= uint64(capacity) {
			encoded := C.GoBytes(buffer, C.int(required))
			C.free(buffer)
			return encoded, updated, nil
		}
		C.free(buffer)
		if required <= uint64(capacity) || required > maxEntityStateBytes {
			return nil, data, fmt.Errorf("encode native entity NBT: %w", err)
		}
		capacity = int(required)
	}
	return nil, data, errors.New("native entity NBT did not fit")
}

func (r *Runtime) EntityOpen(instance EntityInstanceID, invocation InvocationID, handle EntityHandleID, data EntityCommonData) (EntityOpenID, uint32, EntityCommonData, error) {
	var open EntityOpenID
	var capabilities uint32
	updated, err := r.entityExactCall(EntityOpenID(instance), uint32(C.DF_ENTITY_OPERATION_OPEN), invocation, handle, data, func(state *C.DfEntityExactState) {
		open = EntityOpenID(state.instance)
		capabilities = uint32(state.capabilities)
	})
	if err != nil || open == 0 || capabilities&^EntityCapabilityTicker != 0 {
		return 0, 0, data, fmt.Errorf("open native entity: %w", err)
	}
	return open, capabilities, updated, nil
}

func (r *Runtime) EntityBBox(open EntityOpenID, invocation InvocationID, data EntityCommonData) (BBox, EntityCommonData, error) {
	var box BBox
	updated, err := r.entityExactCall(open, uint32(C.DF_ENTITY_OPERATION_BBOX), invocation, EntityHandleID{}, data, func(state *C.DfEntityExactState) {
		box = BBox{Min: nativeEntityVec3(state.bbox.min), Max: nativeEntityVec3(state.bbox.max)}
	})
	return box, updated, err
}

func (r *Runtime) EntityClose(open EntityOpenID, invocation InvocationID, data EntityCommonData) (EntityCommonData, error) {
	return r.entityExactCall(open, uint32(C.DF_ENTITY_OPERATION_CLOSE), invocation, EntityHandleID{}, data, nil)
}

func (r *Runtime) EntityH(open EntityOpenID, invocation InvocationID, data EntityCommonData) (EntityHandleID, EntityCommonData, error) {
	var handle EntityHandleID
	updated, err := r.entityExactCall(open, uint32(C.DF_ENTITY_OPERATION_HANDLE), invocation, EntityHandleID{}, data, func(state *C.DfEntityExactState) {
		handle = entityHandleID(state.handle)
	})
	return handle, updated, err
}

func (r *Runtime) EntityPosition(open EntityOpenID, invocation InvocationID, data EntityCommonData) (Vec3, EntityCommonData, error) {
	var position Vec3
	updated, err := r.entityExactCall(open, uint32(C.DF_ENTITY_OPERATION_POSITION), invocation, EntityHandleID{}, data, func(state *C.DfEntityExactState) {
		position = nativeEntityVec3(state.position)
	})
	return position, updated, err
}

func (r *Runtime) EntityRotation(open EntityOpenID, invocation InvocationID, data EntityCommonData) (Rotation, EntityCommonData, error) {
	var rotation Rotation
	updated, err := r.entityExactCall(open, uint32(C.DF_ENTITY_OPERATION_ROTATION), invocation, EntityHandleID{}, data, func(state *C.DfEntityExactState) {
		rotation = Rotation{Yaw: float64(state.rotation.yaw), Pitch: float64(state.rotation.pitch)}
	})
	return rotation, updated, err
}

func (r *Runtime) EntityTickExact(open EntityOpenID, invocation InvocationID, current int64, data EntityCommonData) (EntityCommonData, error) {
	return r.entityExactCallCurrent(open, uint32(C.DF_ENTITY_OPERATION_TICK_EXACT), invocation, current, data)
}

func (r *Runtime) EntityReleaseOpen(open EntityOpenID) {
	if r != nil && r.ptr != nil && open != 0 {
		_ = C.bg_runtime_entity_call(r.ptr, C.DfEntityInstanceId(open), C.uint32_t(C.DF_ENTITY_OPERATION_RELEASE_OPEN), nil, nil)
	}
}

func (r *Runtime) entityExactCall(
	identity EntityOpenID,
	operation uint32,
	invocation InvocationID,
	handle EntityHandleID,
	data EntityCommonData,
	read func(*C.DfEntityExactState),
) (EntityCommonData, error) {
	return r.entityExactCallWith(identity, operation, invocation, handle, 0, data, read)
}

func (r *Runtime) entityExactCallCurrent(identity EntityOpenID, operation uint32, invocation InvocationID, current int64, data EntityCommonData) (EntityCommonData, error) {
	return r.entityExactCallWith(identity, operation, invocation, EntityHandleID{}, current, data, nil)
}

func (r *Runtime) entityExactCallWith(
	identity EntityOpenID,
	operation uint32,
	invocation InvocationID,
	handle EntityHandleID,
	current int64,
	data EntityCommonData,
	read func(*C.DfEntityExactState),
) (EntityCommonData, error) {
	if r == nil || r.ptr == nil || identity == 0 || invocation == 0 {
		return data, errors.New("native runtime, entity, or invocation is invalid")
	}
	return withNativeEntityCommon(data, func(common *C.DfEntityDataState) C.DfStatus {
		input := (*C.DfEntityExactInput)(C.malloc(C.size_t(C.sizeof_DfEntityExactInput)))
		state := (*C.DfEntityExactState)(C.malloc(C.size_t(C.sizeof_DfEntityExactState)))
		if input == nil || state == nil {
			C.free(unsafe.Pointer(input))
			C.free(unsafe.Pointer(state))
			return C.DF_STATUS_ERROR
		}
		defer C.free(unsafe.Pointer(input))
		defer C.free(unsafe.Pointer(state))
		*input = C.DfEntityExactInput{invocation: C.DfInvocationId(invocation), data: common, current: C.int64_t(current)}
		input.handle = cEntityHandleID(handle)
		*state = C.DfEntityExactState{}
		status := C.bg_runtime_entity_call(r.ptr, C.DfEntityInstanceId(identity), C.uint32_t(operation), input, state)
		if status == C.DF_STATUS_OK && read != nil {
			read(state)
		}
		return status
	})
}

func withNativeEntityCommon(data EntityCommonData, call func(*C.DfEntityDataState) C.DfStatus) (EntityCommonData, error) {
	if len(data.Name) > maxEntityTagBytes || !utf8.ValidString(data.Name) {
		return data, errors.New("invalid native entity name")
	}
	name := C.malloc(C.size_t(maxEntityTagBytes))
	if name == nil {
		return data, errors.New("allocate native entity name")
	}
	defer C.free(name)
	common := C.malloc(C.size_t(C.sizeof_DfEntityDataState))
	if common == nil {
		return data, errors.New("allocate native entity data")
	}
	defer C.free(common)
	copy(unsafe.Slice((*byte)(name), maxEntityTagBytes), data.Name)
	state := (*C.DfEntityDataState)(common)
	*state = C.DfEntityDataState{
		position:                  C.DfVec3{x: C.double(data.Position.X), y: C.double(data.Position.Y), z: C.double(data.Position.Z)},
		velocity:                  C.DfVec3{x: C.double(data.Velocity.X), y: C.double(data.Velocity.Y), z: C.double(data.Velocity.Z)},
		rotation:                  C.DfRotation{yaw: C.double(data.Rotation.Yaw), pitch: C.double(data.Rotation.Pitch)},
		name:                      C.DfStringBuffer{data: (*C.uint8_t)(name), len: C.uint64_t(len(data.Name)), capacity: C.uint64_t(maxEntityTagBytes)},
		fire_duration_nanoseconds: C.int64_t(data.FireDuration), age_nanoseconds: C.int64_t(data.Age),
	}
	if status := call(state); status != C.DF_STATUS_OK {
		return data, fmt.Errorf("native entity callback status %d", int32(status))
	}
	if uint64(state.name.len) > maxEntityTagBytes {
		return data, errors.New("native entity name is too large")
	}
	updated := EntityCommonData{
		Position: nativeEntityVec3(state.position), Velocity: nativeEntityVec3(state.velocity),
		Rotation:     Rotation{Yaw: float64(state.rotation.yaw), Pitch: float64(state.rotation.pitch)},
		Name:         string(unsafe.Slice((*byte)(name), int(state.name.len))),
		FireDuration: time.Duration(state.fire_duration_nanoseconds), Age: time.Duration(state.age_nanoseconds),
	}
	if !utf8.ValidString(updated.Name) {
		return data, errors.New("native entity returned invalid name")
	}
	return updated, nil
}

func (r *Runtime) Commands() ([]Command, error) {
	if r == nil || r.ptr == nil {
		return nil, errors.New("native runtime is closed")
	}
	count := uint64(C.bg_runtime_command_count(r.ptr))
	commands := make([]Command, 0, count)
	for index := uint64(0); index < count; index++ {
		var descriptor C.DfCommandDescriptor
		if status := C.bg_runtime_command_at(r.ptr, C.uint64_t(index), &descriptor); status != C.DF_STATUS_OK {
			return nil, fmt.Errorf("read native command %d: status %d", index, int32(status))
		}
		command := Command{
			Index:       index,
			Name:        stringView(descriptor.name),
			Description: stringView(descriptor.description),
		}
		if descriptor.alias_count > 0 && descriptor.aliases == nil {
			return nil, fmt.Errorf("read native command %d: null aliases", index)
		}
		for _, alias := range unsafe.Slice(descriptor.aliases, int(descriptor.alias_count)) {
			command.Aliases = append(command.Aliases, stringView(alias))
		}
		if descriptor.overload_count > 0 && descriptor.overloads == nil {
			return nil, fmt.Errorf("read native command %d: null overloads", index)
		}
		overloads := unsafe.Slice(descriptor.overloads, int(descriptor.overload_count))
		for _, nativeOverload := range overloads {
			overload := CommandOverload{}
			if nativeOverload.parameter_count > 0 && nativeOverload.parameters == nil {
				return nil, fmt.Errorf("read native command %d: null parameters", index)
			}
			parameters := unsafe.Slice(nativeOverload.parameters, int(nativeOverload.parameter_count))
			for _, nativeParameter := range parameters {
				parameter := CommandParameter{
					Kind:     CommandParameterKind(nativeParameter.kind),
					Name:     stringView(nativeParameter.name),
					Suffix:   stringView(nativeParameter.suffix),
					Optional: nativeParameter.optional != 0,
				}
				if nativeParameter.value_count > 0 && nativeParameter.values == nil {
					return nil, fmt.Errorf("read native command %d: null enum values", index)
				}
				values := unsafe.Slice(nativeParameter.values, int(nativeParameter.value_count))
				for _, value := range values {
					parameter.Values = append(parameter.Values, stringView(value))
				}
				overload.Parameters = append(overload.Parameters, parameter)
			}
			command.Overloads = append(command.Overloads, overload)
		}
		commands = append(commands, command)
	}
	return commands, nil
}

func (r *Runtime) HandleCommand(index uint64, input CommandInput) (CommandOutput, error) {
	var output CommandOutput
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	source := C.CBytes([]byte(input.Source))
	defer C.free(source)
	var argumentViews *C.DfStringView
	var argumentValues []unsafe.Pointer
	if len(input.Arguments) != 0 {
		memory := C.malloc(C.size_t(len(input.Arguments)) * C.size_t(unsafe.Sizeof(C.DfStringView{})))
		if memory == nil {
			return output, errors.New("allocate command argument views")
		}
		defer C.free(memory)
		argumentViews = (*C.DfStringView)(memory)
		views := unsafe.Slice(argumentViews, len(input.Arguments))
		for index, argument := range input.Arguments {
			value := C.CBytes([]byte(argument))
			argumentValues = append(argumentValues, value)
			views[index] = C.DfStringView{data: (*C.uint8_t)(value), len: C.uint64_t(len(argument))}
		}
		defer func() {
			for _, value := range argumentValues {
				C.free(value)
			}
		}()
	}
	message := C.malloc(MaxCommandOutputBytes)
	if message == nil {
		return output, errors.New("allocate command output buffer")
	}
	defer C.free(message)

	nativeInput := C.DfCommandInput{
		invocation:      C.DfInvocationId(input.Invocation),
		overload:        C.uint64_t(input.Overload),
		source:          C.DfStringView{data: (*C.uint8_t)(source), len: C.uint64_t(len(input.Source))},
		arguments:       argumentViews,
		argument_count:  C.uint64_t(len(input.Arguments)),
		source_kind:     C.uint32_t(input.SourceKind),
		source_position: C.DfVec3{x: C.double(input.SourcePosition.X), y: C.double(input.SourcePosition.Y), z: C.double(input.SourcePosition.Z)},
	}
	if input.SourcePlayer != nil {
		fillPlayerID(&nativeInput.source_player, *input.SourcePlayer)
		nativeInput.source_kind = C.DF_COMMAND_SOURCE_PLAYER
	}
	var playerNames []unsafe.Pointer
	if len(input.OnlinePlayers) != 0 {
		memory := C.malloc(C.size_t(len(input.OnlinePlayers)) * C.size_t(unsafe.Sizeof(C.DfCommandPlayer{})))
		if memory == nil {
			return output, errors.New("allocate command player snapshots")
		}
		defer C.free(memory)
		nativeInput.online_players = (*C.DfCommandPlayer)(memory)
		nativeInput.online_player_count = C.uint64_t(len(input.OnlinePlayers))
		players := unsafe.Slice(nativeInput.online_players, len(input.OnlinePlayers))
		for index, snapshot := range input.OnlinePlayers {
			name := C.CBytes([]byte(snapshot.Name))
			playerNames = append(playerNames, name)
			fillPlayerID(&players[index].player, snapshot.Player)
			players[index].name = C.DfStringView{data: (*C.uint8_t)(name), len: C.uint64_t(len(snapshot.Name))}
			players[index].latency_milliseconds = C.uint64_t(snapshot.LatencyMilliseconds)
			players[index].position = C.DfVec3{x: C.double(snapshot.Position.X), y: C.double(snapshot.Position.Y), z: C.double(snapshot.Position.Z)}
		}
		defer func() {
			for _, name := range playerNames {
				C.free(name)
			}
		}()
	}
	state := C.DfCommandState{
		output: C.DfStringBuffer{data: (*C.uint8_t)(message), capacity: MaxCommandOutputBytes},
	}
	if status := C.bg_runtime_handle_command(r.ptr, C.uint64_t(index), &nativeInput, &state); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native command handler failed with status %d", int32(status))
	}
	output.Failed = state.failed != 0
	output.Message = string(C.GoBytes(message, C.int(state.output.len)))
	return output, nil
}

func (r *Runtime) CommandEnumOptions(index, overload, parameter uint64, input CommandEnumContext) ([]string, error) {
	if r == nil || r.ptr == nil {
		return nil, errors.New("native runtime is closed")
	}
	source := C.CBytes([]byte(input.Source))
	defer C.free(source)
	var playersPointer *C.DfCommandPlayer
	var playerStrings []unsafe.Pointer
	if len(input.OnlinePlayers) != 0 {
		playersMemory := C.malloc(C.size_t(len(input.OnlinePlayers)) * C.size_t(unsafe.Sizeof(C.DfCommandPlayer{})))
		if playersMemory == nil {
			return nil, errors.New("allocate online player snapshots")
		}
		defer C.free(playersMemory)
		playersPointer = (*C.DfCommandPlayer)(playersMemory)
		players := unsafe.Slice(playersPointer, len(input.OnlinePlayers))
		for index, snapshot := range input.OnlinePlayers {
			value := C.CBytes([]byte(snapshot.Name))
			playerStrings = append(playerStrings, value)
			fillPlayerID(&players[index].player, snapshot.Player)
			players[index].name = C.DfStringView{data: (*C.uint8_t)(value), len: C.uint64_t(len(snapshot.Name))}
			players[index].latency_milliseconds = C.uint64_t(snapshot.LatencyMilliseconds)
			players[index].position = C.DfVec3{x: C.double(snapshot.Position.X), y: C.double(snapshot.Position.Y), z: C.double(snapshot.Position.Z)}
		}
		defer func() {
			for _, value := range playerStrings {
				C.free(value)
			}
		}()
	}
	buffer := C.malloc(MaxCommandEnumBytes)
	if buffer == nil {
		return nil, errors.New("allocate command enum buffer")
	}
	defer C.free(buffer)
	output := C.DfStringBuffer{data: (*C.uint8_t)(buffer), capacity: MaxCommandEnumBytes}
	context := C.DfCommandEnumContext{
		source:              C.DfStringView{data: (*C.uint8_t)(source), len: C.uint64_t(len(input.Source))},
		source_kind:         C.uint32_t(input.SourceKind),
		source_position:     C.DfVec3{x: C.double(input.SourcePosition.X), y: C.double(input.SourcePosition.Y), z: C.double(input.SourcePosition.Z)},
		online_players:      playersPointer,
		online_player_count: C.uint64_t(len(input.OnlinePlayers)),
	}
	if input.SourcePlayer != nil {
		fillPlayerID(&context.source_player, *input.SourcePlayer)
		context.source_kind = C.DF_COMMAND_SOURCE_PLAYER
	}
	status := C.bg_runtime_command_enum_options(
		r.ptr,
		C.uint64_t(index),
		C.uint64_t(overload),
		C.uint64_t(parameter),
		&context,
		&output,
	)
	if status != C.DF_STATUS_OK {
		return nil, fmt.Errorf("native command enum handler failed with status %d", int32(status))
	}
	if output.len == 0 {
		return nil, nil
	}
	return strings.Split(string(C.GoBytes(buffer, C.int(output.len))), "\n"), nil
}

func (r *Runtime) HandlePlayerMove(invocation InvocationID, input PlayerMoveInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerMoveInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	fillBorrowedPlayerSnapshot(&nativeInput.player, input.Player)
	nativeInput.old_position = C.DfVec3{x: C.double(input.OldPosition.X), y: C.double(input.OldPosition.Y), z: C.double(input.OldPosition.Z)}
	nativeInput.new_position = C.DfVec3{x: C.double(input.NewPosition.X), y: C.double(input.NewPosition.Y), z: C.double(input.NewPosition.Z)}
	nativeInput.rotation = C.DfRotation{yaw: C.double(input.Rotation.Yaw), pitch: C.double(input.Rotation.Pitch)}
	var state C.DfPlayerMoveState
	if cancelled {
		state.cancelled = 1
	}
	packed := uint64(C.bg_runtime_handle_player_move_value(r.ptr, nativeInput, state.cancelled))
	runtime.KeepAlive(input.Player.Name)
	status := int32(uint32(packed >> 32))
	finalCancelled := stickyCancellation(cancelled, uint8(packed) != 0)
	if status != C.DF_STATUS_OK {
		return finalCancelled, fmt.Errorf("native movement handler failed with status %d", status)
	}
	return finalCancelled, nil
}

func (r *Runtime) HandlePlayerChat(invocation InvocationID, input PlayerChatInput, cancelled bool) (PlayerChatOutput, error) {
	output := PlayerChatOutput{Cancelled: cancelled}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	message := C.CBytes([]byte(input.Message))
	defer C.free(message)
	replacement := C.malloc(MaxChatReplacementBytes)
	if replacement == nil {
		return output, errors.New("allocate chat replacement buffer")
	}
	defer C.free(replacement)

	var nativeInput C.DfPlayerChatInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.message = C.DfStringView{
		data: (*C.uint8_t)(message),
		len:  C.uint64_t(len(input.Message)),
	}
	state := C.DfPlayerChatState{
		replacement: C.DfStringBuffer{
			data:     (*C.uint8_t)(replacement),
			capacity: MaxChatReplacementBytes,
		},
	}
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_CHAT, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
		return output, fmt.Errorf("native chat handler failed with status %d", int32(status))
	}
	output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
	if state.has_replacement != 0 {
		value := string(C.GoBytes(replacement, C.int(state.replacement.len)))
		output.Replacement = &value
	}
	return output, nil
}

func (r *Runtime) HandlePlayerJoin(invocation InvocationID, input PlayerJoinInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	name := C.CBytes([]byte(input.Name))
	defer C.free(name)
	var nativeInput C.DfPlayerJoinInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.name = C.DfStringView{data: (*C.uint8_t)(name), len: C.uint64_t(len(input.Name))}
	var state C.DfPlayerJoinState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_JOIN, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native join handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerQuit(invocation InvocationID, input PlayerQuitInput) error {
	if r == nil || r.ptr == nil {
		return errors.New("native runtime is closed")
	}
	name := C.CBytes([]byte(input.Name))
	defer C.free(name)
	var nativeInput C.DfPlayerQuitInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.name = C.DfStringView{data: (*C.uint8_t)(name), len: C.uint64_t(len(input.Name))}
	var state C.DfPlayerQuitState
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_QUIT, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return fmt.Errorf("native quit handler failed with status %d", int32(status))
	}
	return nil
}

func (r *Runtime) HandlePlayerHurt(invocation InvocationID, input PlayerHurtInput, cancelled bool) (PlayerHurtOutput, error) {
	output := PlayerHurtOutput{Cancelled: cancelled, Damage: input.Damage, AttackImmunity: input.AttackImmunity}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	source, releaseSource, ok := nativeDamageSource(input.Source)
	if !ok {
		return output, errors.New("encode hurt damage source")
	}
	defer releaseSource()
	var nativeInput C.DfPlayerHurtInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.immune = C.uint8_t(boolByte(input.Immune))
	nativeInput.source = source
	state := C.DfPlayerHurtState{
		damage:                      C.double(input.Damage),
		attack_immunity_nanoseconds: C.int64_t(input.AttackImmunity),
	}
	if cancelled {
		state.cancelled = 1
	}
	status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_HURT, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state))
	output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
	if status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native hurt handler failed with status %d", int32(status))
	}
	output.Damage = float64(state.damage)
	output.AttackImmunity = time.Duration(state.attack_immunity_nanoseconds)
	return output, nil
}

func (r *Runtime) HandlePlayerHeal(invocation InvocationID, input PlayerHealInput, cancelled bool) (PlayerHealOutput, error) {
	output := PlayerHealOutput{Cancelled: cancelled, Health: input.Health}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	source, releaseSource, ok := nativeHealingSource(input.Source)
	if !ok {
		return output, errors.New("encode healing source")
	}
	defer releaseSource()
	var nativeInput C.DfPlayerHealInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.source = source
	state := C.DfPlayerHealState{health: C.double(input.Health)}
	if cancelled {
		state.cancelled = 1
	}
	status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_HEAL, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state))
	output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
	if status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native heal handler failed with status %d", int32(status))
	}
	output.Health = float64(state.health)
	return output, nil
}

func (r *Runtime) HandlePlayerBlockBreak(invocation InvocationID, input PlayerBlockBreakInput, cancelled bool) (PlayerBlockBreakOutput, error) {
	output := PlayerBlockBreakOutput{Cancelled: cancelled, Drops: input.Drops, Experience: input.Experience}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	blockArena := &nativeViewArena{}
	defer blockArena.release()
	block, ok := nativeBlockView(input.Block, blockArena)
	if !ok {
		return output, errors.New("encode block-break block")
	}
	drops, dropCount, releaseDrops, ok := nativeItemStackViews(input.Drops)
	defer releaseDrops()
	if !ok {
		return output, errors.New("encode block-break drops")
	}
	var nativeInput C.DfPlayerBlockBreakInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.position = nativeBlockPos(input.Position)
	nativeInput.block = block
	nativeInput.drops = drops
	nativeInput.drop_count = C.uint64_t(dropCount)
	state := C.DfPlayerBlockBreakState{experience: C.int32_t(input.Experience)}
	if cancelled {
		state.cancelled = 1
	}
	status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_BLOCK_BREAK, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state))
	output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
	hasReplacement := state.replacement_drop != nil
	replacementFields := state.replacement_drops != nil || state.replacement_drop_count != 0 || state.replacement_context != nil
	var replacements []ItemStack
	validReplacement := true
	if hasReplacement {
		replacements, validReplacement = copyOwnedItemStackViews(
			state.replacement_drops, uint64(state.replacement_drop_count), state.replacement_context, state.replacement_drop,
		)
	} else if replacementFields {
		validReplacement = false
	}
	if status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native block-break handler failed with status %d", int32(status))
	}
	if !validReplacement {
		return output, errors.New("native block-break handler returned invalid replacement drops")
	}
	output.Experience = int32(state.experience)
	if hasReplacement {
		output.Drops = replacements
	}
	return output, nil
}

func (r *Runtime) HandlePlayerBlockPlace(invocation InvocationID, input PlayerBlockPlaceInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	blockArena := &nativeViewArena{}
	defer blockArena.release()
	block, ok := nativeBlockView(input.Block, blockArena)
	if !ok {
		return cancelled, errors.New("encode block-place block")
	}
	var nativeInput C.DfPlayerBlockPlaceInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.position = nativeBlockPos(input.Position)
	nativeInput.block = block
	var state C.DfPlayerBlockPlaceState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_BLOCK_PLACE, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native block-place handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerFoodLoss(invocation InvocationID, input PlayerFoodLossInput, cancelled bool) (PlayerFoodLossOutput, error) {
	output := PlayerFoodLossOutput{Cancelled: cancelled, To: input.To}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerFoodLossInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.from = C.int32_t(input.From)
	state := C.DfPlayerFoodLossState{to: C.int32_t(input.To)}
	if cancelled {
		state.cancelled = 1
	}
	status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_FOOD_LOSS, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state))
	output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
	if status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native food-loss handler failed with status %d", int32(status))
	}
	output.To = int32(state.to)
	return output, nil
}

func (r *Runtime) HandlePlayerDeath(invocation InvocationID, input PlayerDeathInput, keepInventory bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return keepInventory, errors.New("native runtime is closed")
	}
	source, releaseSource, ok := nativeDamageSource(input.Source)
	if !ok {
		return keepInventory, errors.New("encode death damage source")
	}
	defer releaseSource()
	var nativeInput C.DfPlayerDeathInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.source = source
	var state C.DfPlayerDeathState
	if keepInventory {
		state.keep_inventory = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_DEATH, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return keepInventory, fmt.Errorf("native death handler failed with status %d", int32(status))
	}
	return state.keep_inventory != 0, nil
}

func (r *Runtime) HandlePlayerStartBreak(invocation InvocationID, input PlayerPositionInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerStartBreakInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.position = nativeBlockPos(input.Position)
	var state C.DfPlayerStartBreakState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_START_BREAK, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native start-break handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerFireExtinguish(invocation InvocationID, input PlayerPositionInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerFireExtinguishInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.position = nativeBlockPos(input.Position)
	var state C.DfPlayerFireExtinguishState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_FIRE_EXTINGUISH, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native fire-extinguish handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerToggleSprint(invocation InvocationID, input PlayerToggleInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerToggleSprintInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.after = C.uint8_t(boolByte(input.After))
	var state C.DfPlayerToggleSprintState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_TOGGLE_SPRINT, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native toggle-sprint handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}
func (r *Runtime) HandlePlayerToggleSneak(invocation InvocationID, input PlayerToggleInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerToggleSneakInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.after = C.uint8_t(boolByte(input.After))
	var state C.DfPlayerToggleSneakState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_TOGGLE_SNEAK, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native toggle-sneak handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerJump(invocation InvocationID, player PlayerSnapshot) error {
	if r == nil || r.ptr == nil {
		return errors.New("native runtime is closed")
	}
	var input C.DfPlayerJumpInput
	input.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&input.player, player)
	defer C.free(playerName)
	var state C.DfPlayerJumpState
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_JUMP, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return fmt.Errorf("native jump handler failed with status %d", int32(status))
	}
	return nil
}

func (r *Runtime) HandlePlayerTeleport(invocation InvocationID, input PlayerTeleportInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerTeleportInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.position = C.DfVec3{x: C.double(input.Position.X), y: C.double(input.Position.Y), z: C.double(input.Position.Z)}
	var state C.DfPlayerTeleportState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_TELEPORT, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native teleport handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerExperienceGain(invocation InvocationID, player PlayerSnapshot, amount int, cancelled bool) (PlayerExperienceGainOutput, error) {
	output := PlayerExperienceGainOutput{Cancelled: cancelled, Amount: amount}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	var input C.DfPlayerExperienceGainInput
	input.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&input.player, player)
	defer C.free(playerName)
	state := C.DfPlayerExperienceGainState{amount: C.int32_t(amount)}
	if cancelled {
		state.cancelled = 1
	}
	status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_EXPERIENCE_GAIN, unsafe.Pointer(&input), unsafe.Pointer(&state))
	output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
	if status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native experience-gain handler failed with status %d", int32(status))
	}
	return PlayerExperienceGainOutput{Cancelled: output.Cancelled, Amount: int(state.amount)}, nil
}

func (r *Runtime) HandlePlayerPunchAir(invocation InvocationID, player PlayerSnapshot, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var input C.DfPlayerPunchAirInput
	input.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&input.player, player)
	defer C.free(playerName)
	var state C.DfPlayerPunchAirState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_PUNCH_AIR, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native punch-air handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerHeldSlotChange(invocation InvocationID, input PlayerHeldSlotChangeInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerHeldSlotChangeInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.from = C.int32_t(input.From)
	nativeInput.to = C.int32_t(input.To)
	var state C.DfPlayerHeldSlotChangeState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_HELD_SLOT_CHANGE, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native held-slot-change handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerSleep(invocation InvocationID, player PlayerSnapshot, sendReminder, cancelled bool) (PlayerSleepOutput, error) {
	output := PlayerSleepOutput{Cancelled: cancelled, SendReminder: sendReminder}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	var input C.DfPlayerSleepInput
	input.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&input.player, player)
	defer C.free(playerName)
	state := C.DfPlayerSleepState{send_reminder: C.uint8_t(boolByte(sendReminder))}
	if cancelled {
		state.cancelled = 1
	}
	status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_SLEEP, unsafe.Pointer(&input), unsafe.Pointer(&state))
	output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
	if status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native sleep handler failed with status %d", int32(status))
	}
	return PlayerSleepOutput{Cancelled: output.Cancelled, SendReminder: state.send_reminder != 0}, nil
}

func (r *Runtime) HandlePlayerBlockPick(invocation InvocationID, input PlayerBlockPickInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	blockArena := &nativeViewArena{}
	defer blockArena.release()
	block, ok := nativeBlockView(input.Block, blockArena)
	if !ok {
		return cancelled, errors.New("encode block-pick block")
	}
	nativeInput := C.DfPlayerBlockPickInput{invocation: C.DfInvocationId(invocation), position: nativeBlockPos(input.Position), block: block}
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	var state C.DfPlayerBlockPickState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_BLOCK_PICK, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native block-pick handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerLecternPageTurn(invocation InvocationID, input PlayerLecternPageTurnInput, cancelled bool) (PlayerLecternPageTurnOutput, error) {
	output := PlayerLecternPageTurnOutput{Cancelled: cancelled, NewPage: input.NewPage}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerLecternPageTurnInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.position = nativeBlockPos(input.Position)
	nativeInput.old_page = C.int32_t(input.OldPage)
	state := C.DfPlayerLecternPageTurnState{new_page: C.int32_t(input.NewPage)}
	if cancelled {
		state.cancelled = 1
	}
	status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_LECTERN_PAGE_TURN, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state))
	output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
	if status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native lectern-page-turn handler failed with status %d", int32(status))
	}
	return PlayerLecternPageTurnOutput{Cancelled: output.Cancelled, NewPage: int(state.new_page)}, nil
}

func (r *Runtime) HandlePlayerSignEdit(invocation InvocationID, input PlayerSignEditInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	oldText, newText := unsafe.StringData(input.OldText), unsafe.StringData(input.NewText)
	nativeInput := C.DfPlayerSignEditInput{
		invocation: C.DfInvocationId(invocation),
		position:   nativeBlockPos(input.Position),
		front_side: C.uint8_t(boolByte(input.FrontSide)),
		old_text:   C.DfStringView{data: (*C.uint8_t)(unsafe.Pointer(oldText)), len: C.uint64_t(len(input.OldText))},
		new_text:   C.DfStringView{data: (*C.uint8_t)(unsafe.Pointer(newText)), len: C.uint64_t(len(input.NewText))},
	}
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	var state C.DfPlayerSignEditState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_SIGN_EDIT, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native sign-edit handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerItemUse(invocation InvocationID, player PlayerSnapshot, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var input C.DfPlayerItemUseInput
	input.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&input.player, player)
	defer C.free(playerName)
	var state C.DfPlayerItemUseState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_ITEM_USE, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native item-use handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerItemUseOnBlock(invocation InvocationID, input PlayerItemUseOnBlockInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerItemUseOnBlockInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.position = nativeBlockPos(input.Position)
	nativeInput.face = C.int32_t(input.Face)
	nativeInput.click_position = C.DfVec3{x: C.double(input.ClickPosition.X), y: C.double(input.ClickPosition.Y), z: C.double(input.ClickPosition.Z)}
	var state C.DfPlayerItemUseOnBlockState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_ITEM_USE_ON_BLOCK, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native item-use-on-block handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerItemConsume(invocation InvocationID, player PlayerSnapshot, item ItemStack, cancelled bool) (bool, error) {
	return r.handlePlayerItemStackEvent(invocation, C.DF_EVENT_PLAYER_ITEM_CONSUME, player, item, 0, cancelled)
}

func (r *Runtime) HandlePlayerItemRelease(invocation InvocationID, player PlayerSnapshot, item ItemStack, duration time.Duration, cancelled bool) (bool, error) {
	return r.handlePlayerItemStackEvent(invocation, C.DF_EVENT_PLAYER_ITEM_RELEASE, player, item, int64(duration), cancelled)
}

func (r *Runtime) handlePlayerItemStackEvent(invocation InvocationID, event C.DfEventId, player PlayerSnapshot, item ItemStack, duration int64, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	arena := &nativeViewArena{}
	defer arena.release()
	itemView, ok := nativeItemStackView(item, arena)
	if !ok {
		return cancelled, errors.New("encode item event stack")
	}
	var state C.DfPlayerItemConsumeState
	if cancelled {
		state.cancelled = 1
	}
	if event == C.DF_EVENT_PLAYER_ITEM_CONSUME {
		var input C.DfPlayerItemConsumeInput
		input.invocation = C.DfInvocationId(invocation)
		playerName := fillPlayerSnapshot(&input.player, player)
		defer C.free(playerName)
		input.item = itemView
		if status := C.bg_runtime_handle_event(r.ptr, event, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
			return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native item-consume handler failed with status %d", int32(status))
		}
		return stickyCancellation(cancelled, state.cancelled != 0), nil
	}
	var input C.DfPlayerItemReleaseInput
	input.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&input.player, player)
	defer C.free(playerName)
	input.item = itemView
	input.duration_nanoseconds = C.int64_t(duration)
	if status := C.bg_runtime_handle_event(r.ptr, event, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native item-release handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerItemDamage(invocation InvocationID, player PlayerSnapshot, item ItemStack, damage int, cancelled bool) (PlayerItemDamageOutput, error) {
	output := PlayerItemDamageOutput{Cancelled: cancelled, Damage: damage}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	var input C.DfPlayerItemDamageInput
	input.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&input.player, player)
	defer C.free(playerName)
	arena := &nativeViewArena{}
	defer arena.release()
	itemView, ok := nativeItemStackView(item, arena)
	if !ok {
		return output, errors.New("encode item-damage stack")
	}
	input.item = itemView
	state := C.DfPlayerItemDamageState{damage: C.int32_t(damage)}
	if cancelled {
		state.cancelled = 1
	}
	status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_ITEM_DAMAGE, unsafe.Pointer(&input), unsafe.Pointer(&state))
	output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
	if status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native item-damage handler failed with status %d", int32(status))
	}
	return PlayerItemDamageOutput{Cancelled: output.Cancelled, Damage: int(state.damage)}, nil
}

func (r *Runtime) HandlePlayerItemDrop(invocation InvocationID, player PlayerSnapshot, item ItemStack, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var input C.DfPlayerItemDropInput
	input.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&input.player, player)
	defer C.free(playerName)
	arena := &nativeViewArena{}
	defer arena.release()
	itemView, ok := nativeItemStackView(item, arena)
	if !ok {
		return cancelled, errors.New("encode item-drop stack")
	}
	input.item = itemView
	var state C.DfPlayerItemDropState
	if cancelled {
		state.cancelled = 1
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_ITEM_DROP, unsafe.Pointer(&input), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native item-drop handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerItemPickup(invocation InvocationID, input PlayerItemPickupInput, cancelled bool) (PlayerItemPickupOutput, error) {
	output := PlayerItemPickupOutput{Cancelled: cancelled, Item: input.Item}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	arena := &nativeViewArena{}
	defer arena.release()
	itemView, ok := nativeItemStackView(input.Item, arena)
	if !ok {
		return output, errors.New("encode item-pickup item")
	}
	nativeInput := C.DfPlayerItemPickupInput{invocation: C.DfInvocationId(invocation), item: itemView}
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	state := C.DfPlayerItemPickupState{cancelled: C.uint8_t(boolByte(cancelled))}
	status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_ITEM_PICKUP, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state))
	output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
	hasReplacement := state.replacement_drop != nil
	replacementFields := state.replacement != nil || state.replacement_context != nil
	var replacement []ItemStack
	validReplacement := true
	if hasReplacement {
		replacement, validReplacement = copyOwnedItemStackViews(state.replacement, 1, state.replacement_context, state.replacement_drop)
	} else if replacementFields {
		validReplacement = false
	}
	if status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native item-pickup handler failed with status %d", int32(status))
	}
	if !validReplacement {
		return output, errors.New("native item-pickup handler returned invalid replacement")
	}
	if hasReplacement {
		output.Item = replacement[0]
	}
	return output, nil
}

func (r *Runtime) HandlePlayerAttackEntity(invocation InvocationID, input PlayerAttackEntityInput, force, height float64, critical, cancelled bool) (PlayerAttackEntityOutput, error) {
	output := PlayerAttackEntityOutput{Cancelled: cancelled, KnockbackForce: force, KnockbackHeight: height, Critical: critical}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerAttackEntityInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	fillEntityID(&nativeInput.target, input.Target)
	state := C.DfPlayerAttackEntityState{
		knockback_force: C.double(force), knockback_height: C.double(height),
		critical: C.uint8_t(boolByte(critical)), cancelled: C.uint8_t(boolByte(cancelled)),
	}
	status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_ATTACK_ENTITY, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state))
	output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
	if status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native attack-entity handler failed with status %d", int32(status))
	}
	return PlayerAttackEntityOutput{
		Cancelled: output.Cancelled, KnockbackForce: float64(state.knockback_force),
		KnockbackHeight: float64(state.knockback_height), Critical: state.critical != 0,
	}, nil
}

func (r *Runtime) HandlePlayerItemUseOnEntity(invocation InvocationID, input PlayerItemUseOnEntityInput, cancelled bool) (bool, error) {
	if r == nil || r.ptr == nil {
		return cancelled, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerItemUseOnEntityInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	fillEntityID(&nativeInput.target, input.Target)
	var state C.DfPlayerItemUseOnEntityState
	state.cancelled = C.uint8_t(boolByte(cancelled))
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_ITEM_USE_ON_ENTITY, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return stickyCancellation(cancelled, state.cancelled != 0), fmt.Errorf("native item-use-on-entity handler failed with status %d", int32(status))
	}
	return stickyCancellation(cancelled, state.cancelled != 0), nil
}

func (r *Runtime) HandlePlayerChangeWorld(invocation InvocationID, input PlayerChangeWorldInput) error {
	if r == nil || r.ptr == nil {
		return errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerChangeWorldInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	if input.Before != nil {
		nativeInput.before.value = C.uint64_t(*input.Before)
	}
	nativeInput.after.value = C.uint64_t(input.After)
	var state C.DfPlayerChangeWorldState
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_CHANGE_WORLD, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return fmt.Errorf("native change-world handler failed with status %d", int32(status))
	}
	return nil
}

func (r *Runtime) HandlePlayerRespawn(invocation InvocationID, input PlayerRespawnInput, position Vec3, world WorldID) (PlayerRespawnOutput, error) {
	output := PlayerRespawnOutput{Position: position, World: world}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerRespawnInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	state := C.DfPlayerRespawnState{
		position: C.DfVec3{x: C.double(position.X), y: C.double(position.Y), z: C.double(position.Z)},
		world:    C.DfWorldId{value: C.uint64_t(world)},
	}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_RESPAWN, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native respawn handler failed with status %d", int32(status))
	}
	return PlayerRespawnOutput{
		Position: Vec3{X: float64(state.position.x), Y: float64(state.position.y), Z: float64(state.position.z)},
		World:    WorldID(state.world.value),
	}, nil
}

func (r *Runtime) HandlePlayerSkinChange(invocation InvocationID, input PlayerSkinChangeInput, skin PlayerSkin, cancelled bool) (PlayerSkinChangeOutput, error) {
	output := PlayerSkinChangeOutput{Cancelled: cancelled, Skin: clonePlayerSkin(skin)}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	if input.Player.Player.Generation == 0 || !validPlayerSkinPayload(skin) {
		return output, errors.New("invalid skin-change input")
	}
	snapshot, ok := registerSkinSnapshot(r.hostContext, invocation, skin, true)
	if !ok {
		return output, errors.New("skin snapshot limit reached")
	}
	defer forceUnregisterSkinSnapshot(r.hostContext, invocation, snapshot)

	var nativeInput C.DfPlayerSkinChangeInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.snapshot = C.uint64_t(snapshot)
	state := C.DfPlayerSkinChangeState{cancelled: C.uint8_t(boolByte(cancelled))}
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_SKIN_CHANGE, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
		return output, fmt.Errorf("native skin-change handler failed with status %d", int32(status))
	}
	output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
	finalSkin, ok := resolveSkinSnapshot(r.hostContext, invocation, snapshot)
	if !ok || !validPlayerSkinPayload(finalSkin) {
		return output, errors.New("native skin-change handler returned an invalid skin")
	}
	return PlayerSkinChangeOutput{Cancelled: output.Cancelled, Skin: finalSkin}, nil
}

func (r *Runtime) HandlePlayerTransfer(invocation InvocationID, input PlayerTransferInput, cancelled bool) (PlayerTransferOutput, error) {
	output := PlayerTransferOutput{Cancelled: cancelled, Address: cloneUDPAddress(input.Address)}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	if input.Address.Port < math.MinInt32 || input.Address.Port > math.MaxInt32 ||
		!validTransferIP(input.Address.IP) || len(input.Address.Zone) > maxTransferZoneBytes || !utf8.ValidString(input.Address.Zone) {
		return output, errors.New("invalid transfer address")
	}
	arena := &nativeViewArena{}
	defer arena.release()
	ip, ok := arena.stringView(input.Address.IP)
	if !ok {
		return output, errors.New("allocate transfer IP")
	}
	zone, ok := arena.stringView([]byte(input.Address.Zone))
	if !ok {
		return output, errors.New("allocate transfer zone")
	}
	var nativeInput C.DfPlayerTransferInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	state := C.DfPlayerTransferState{
		cancelled: C.uint8_t(boolByte(cancelled)),
		address: C.DfUDPAddrView{
			ip: ip, port: C.int32_t(input.Address.Port), zone: zone,
		},
	}
	status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_TRANSFER, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state))
	if state.replacement_drop != nil {
		defer C.bg_call_event_drop(state.replacement_drop, state.replacement_context)
	}
	output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
	if status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native transfer handler failed with status %d", int32(status))
	}
	address, ok := copyNativeUDPAddress(state.address)
	if !ok {
		return output, errors.New("native transfer handler returned an invalid address")
	}
	output.Address = address
	return output, nil
}

func (r *Runtime) HandlePlayerCommandExecution(invocation InvocationID, input PlayerCommandExecutionInput, cancelled bool) (PlayerCommandExecutionOutput, error) {
	output := PlayerCommandExecutionOutput{Cancelled: cancelled, Arguments: append([]string(nil), input.Arguments...)}
	if r == nil || r.ptr == nil {
		return output, errors.New("native runtime is closed")
	}
	if len(input.Command.Aliases) > maxCommandExecutionAliasCount || len(input.Arguments) > maxCommandExecutionStringCount ||
		!validEventString(input.Command.Name) || !validEventString(input.Command.Description) || !validEventString(input.Command.Usage) {
		return output, errors.New("command execution input exceeds limits")
	}
	arena := &nativeViewArena{}
	defer arena.release()
	name, ok := arena.stringView([]byte(input.Command.Name))
	if !ok {
		return output, errors.New("allocate command name")
	}
	description, ok := arena.stringView([]byte(input.Command.Description))
	if !ok {
		return output, errors.New("allocate command description")
	}
	usage, ok := arena.stringView([]byte(input.Command.Usage))
	if !ok {
		return output, errors.New("allocate command usage")
	}
	aliases, ok := arena.stringViews(input.Command.Aliases)
	if !ok {
		return output, errors.New("allocate command aliases")
	}
	arguments, ok := arena.stringViews(input.Arguments)
	if !ok {
		return output, errors.New("allocate command arguments")
	}
	var nativeInput C.DfPlayerCommandExecutionInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.command_name = name
	nativeInput.command_description = description
	nativeInput.command_usage = usage
	nativeInput.command_aliases = aliases
	nativeInput.command_alias_count = C.uint64_t(len(input.Command.Aliases))
	nativeInput.arguments = arguments
	nativeInput.argument_count = C.uint64_t(len(input.Arguments))
	state := C.DfPlayerCommandExecutionState{cancelled: C.uint8_t(boolByte(cancelled))}
	status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_COMMAND_EXECUTION, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state))
	if state.replacement_drop != nil {
		defer C.bg_call_event_drop(state.replacement_drop, state.replacement_context)
	}
	output.Cancelled = stickyCancellation(cancelled, state.cancelled != 0)
	if status != C.DF_STATUS_OK {
		return output, fmt.Errorf("native command-execution handler failed with status %d", int32(status))
	}
	if state.replacement_drop == nil {
		return output, nil
	}
	if uint64(len(input.Arguments)) != uint64(state.replacement_argument_count) {
		return output, errors.New("native command-execution handler changed argument count")
	}
	replacement, ok := copyNativeStrings(state.replacement_arguments, uint64(state.replacement_argument_count), maxCommandExecutionStringCount)
	if !ok {
		return output, errors.New("native command-execution handler returned invalid arguments")
	}
	output.Arguments = replacement
	return output, nil
}

func (r *Runtime) HandlePlayerDiagnostics(invocation InvocationID, input PlayerDiagnosticsInput) error {
	if r == nil || r.ptr == nil {
		return errors.New("native runtime is closed")
	}
	var nativeInput C.DfPlayerDiagnosticsInput
	nativeInput.invocation = C.DfInvocationId(invocation)
	playerName := fillPlayerSnapshot(&nativeInput.player, input.Player)
	defer C.free(playerName)
	nativeInput.average_frames_per_second = C.double(input.AverageFramesPerSecond)
	nativeInput.average_server_sim_tick_time = C.double(input.AverageServerSimTickTime)
	nativeInput.average_client_sim_tick_time = C.double(input.AverageClientSimTickTime)
	nativeInput.average_begin_frame_time = C.double(input.AverageBeginFrameTime)
	nativeInput.average_input_time = C.double(input.AverageInputTime)
	nativeInput.average_render_time = C.double(input.AverageRenderTime)
	nativeInput.average_end_frame_time = C.double(input.AverageEndFrameTime)
	nativeInput.average_remainder_time_percent = C.double(input.AverageRemainderTimePercent)
	nativeInput.average_unaccounted_time_percent = C.double(input.AverageUnaccountedTimePercent)
	var state C.DfPlayerDiagnosticsState
	if status := C.bg_runtime_handle_event(r.ptr, C.DF_EVENT_PLAYER_DIAGNOSTICS, unsafe.Pointer(&nativeInput), unsafe.Pointer(&state)); status != C.DF_STATUS_OK {
		return fmt.Errorf("native diagnostics handler failed with status %d", int32(status))
	}
	return nil
}

type nativeViewArena struct {
	allocations []unsafe.Pointer
}

func (a *nativeViewArena) allocate(size uintptr) (unsafe.Pointer, bool) {
	if size == 0 {
		return nil, true
	}
	pointer := C.malloc(C.size_t(size))
	if pointer == nil {
		return nil, false
	}
	a.allocations = append(a.allocations, pointer)
	return pointer, true
}

func (a *nativeViewArena) stringView(value []byte) (C.DfStringView, bool) {
	pointer, ok := a.allocate(uintptr(len(value)))
	if !ok {
		return C.DfStringView{}, false
	}
	if len(value) != 0 {
		copy(unsafe.Slice((*byte)(pointer), len(value)), value)
	}
	return C.DfStringView{data: (*C.uint8_t)(pointer), len: C.uint64_t(len(value))}, true
}

func (a *nativeViewArena) stringViews(values []string) (*C.DfStringView, bool) {
	if len(values) == 0 {
		return nil, true
	}
	pointer, ok := a.allocate(uintptr(len(values)) * C.sizeof_DfStringView)
	if !ok {
		return nil, false
	}
	views := unsafe.Slice((*C.DfStringView)(pointer), len(values))
	for index, value := range values {
		if !validEventString(value) {
			return nil, false
		}
		view, ok := a.stringView([]byte(value))
		if !ok {
			return nil, false
		}
		views[index] = view
	}
	return (*C.DfStringView)(pointer), true
}

func (a *nativeViewArena) release() {
	for index := len(a.allocations) - 1; index >= 0; index-- {
		C.free(a.allocations[index])
	}
}

func validEventString(value string) bool {
	return len(value) <= maxEventStringBytes && utf8.ValidString(value)
}

func validTransferIP(value []byte) bool {
	return len(value) == 0 || len(value) == netIPv4Bytes || len(value) == maxTransferIPBytes
}

const netIPv4Bytes = 4

func cloneUDPAddress(value UDPAddress) UDPAddress {
	value.IP = append([]byte(nil), value.IP...)
	return value
}

func copyNativeUDPAddress(value C.DfUDPAddrView) (UDPAddress, bool) {
	ip, ok := copyNativeBytes(value.ip, maxTransferIPBytes)
	if !ok || !validTransferIP(ip) {
		return UDPAddress{}, false
	}
	zoneBytes, ok := copyNativeBytes(value.zone, maxTransferZoneBytes)
	if !ok || !utf8.Valid(zoneBytes) {
		return UDPAddress{}, false
	}
	return UDPAddress{IP: ip, Port: int(value.port), Zone: string(zoneBytes)}, true
}

func copyNativeStrings(values *C.DfStringView, count uint64, maxCount int) ([]string, bool) {
	if count > uint64(maxCount) {
		return nil, false
	}
	if count == 0 {
		return []string{}, true
	}
	if values == nil {
		return nil, false
	}
	views := unsafe.Slice(values, int(count))
	result := make([]string, len(views))
	for index, view := range views {
		value, ok := copyNativeBytes(view, maxEventStringBytes)
		if !ok || !utf8.Valid(value) {
			return nil, false
		}
		result[index] = string(value)
	}
	return result, true
}

func copyNativeBytes(value C.DfStringView, maxBytes int) ([]byte, bool) {
	if uint64(value.len) > uint64(maxBytes) {
		return nil, false
	}
	if value.len == 0 {
		return []byte{}, true
	}
	if value.data == nil {
		return nil, false
	}
	return C.GoBytes(unsafe.Pointer(value.data), C.int(value.len)), true
}

func nativeItemStackView(value ItemStack, arena *nativeViewArena) (C.DfItemStackViewV3, bool) {
	identifier, ok := arena.stringView([]byte(value.Identifier))
	if !ok {
		return C.DfItemStackViewV3{}, false
	}
	customName, ok := arena.stringView([]byte(value.CustomName))
	if !ok {
		return C.DfItemStackViewV3{}, false
	}
	nbtData, ok := arena.stringView(value.NBT)
	if !ok {
		return C.DfItemStackViewV3{}, false
	}
	valuesNBT, ok := arena.stringView(value.ValuesNBT)
	if !ok {
		return C.DfItemStackViewV3{}, false
	}
	view := C.DfItemStackViewV3{
		identifier: identifier, metadata: C.int32_t(value.Metadata), count: C.uint32_t(value.Count),
		damage: C.uint32_t(value.Damage), unbreakable: C.uint8_t(boolByte(value.Unbreakable)),
		anvil_cost: C.int32_t(value.AnvilCost), custom_name: customName, nbt: nbtData, values_nbt: valuesNBT,
		lore_count: C.uint64_t(len(value.Lore)), enchantment_count: C.uint64_t(len(value.Enchantments)),
	}
	if len(value.Lore) != 0 {
		pointer, allocated := arena.allocate(uintptr(len(value.Lore)) * C.sizeof_DfStringView)
		if !allocated {
			return C.DfItemStackViewV3{}, false
		}
		view.lore = (*C.DfStringView)(pointer)
		for index, line := range value.Lore {
			lineView, valid := arena.stringView([]byte(line))
			if !valid {
				return C.DfItemStackViewV3{}, false
			}
			unsafe.Slice(view.lore, len(value.Lore))[index] = lineView
		}
	}
	if len(value.Enchantments) != 0 {
		pointer, allocated := arena.allocate(uintptr(len(value.Enchantments)) * C.sizeof_DfItemEnchantment)
		if !allocated {
			return C.DfItemStackViewV3{}, false
		}
		view.enchantments = (*C.DfItemEnchantment)(pointer)
		for index, enchantment := range value.Enchantments {
			unsafe.Slice(view.enchantments, len(value.Enchantments))[index] = C.DfItemEnchantment{
				id: C.uint32_t(enchantment.ID), level: C.uint32_t(enchantment.Level),
			}
		}
	}
	return view, true
}

func nativeItemStackViews(values []ItemStack) (*C.DfItemStackViewV3, uint64, func(), bool) {
	arena := &nativeViewArena{}
	release := arena.release
	if len(values) == 0 {
		return nil, 0, release, true
	}
	pointer, ok := arena.allocate(uintptr(len(values)) * C.sizeof_DfItemStackViewV3)
	if !ok {
		return nil, 0, release, false
	}
	views := unsafe.Slice((*C.DfItemStackViewV3)(pointer), len(values))
	for index, value := range values {
		view, valid := nativeItemStackView(value, arena)
		if !valid {
			return nil, 0, release, false
		}
		views[index] = view
	}
	return (*C.DfItemStackViewV3)(pointer), uint64(len(values)), release, true
}

func nativeBlockView(value WorldBlock, arena *nativeViewArena) (C.DfBlockView, bool) {
	identifier, ok := arena.stringView([]byte(value.Identifier))
	if !ok {
		return C.DfBlockView{}, false
	}
	properties, ok := arena.stringView(value.PropertiesNBT)
	if !ok {
		return C.DfBlockView{}, false
	}
	return C.DfBlockView{identifier: identifier, properties_nbt: properties}, true
}

func copyOwnedItemStackViews(data *C.DfItemStackViewV3, count uint64, context unsafe.Pointer, drop C.DfItemStackViewsDropFn) ([]ItemStack, bool) {
	if drop == nil {
		return nil, false
	}
	defer C.bg_call_item_stack_views_drop(drop, context)
	if count == 0 {
		return []ItemStack{}, true
	}
	if data == nil || count > uint64(^uint(0)>>1) {
		return nil, false
	}
	views := unsafe.Slice(data, int(count))
	values := make([]ItemStack, len(views))
	for index := range views {
		value, ok := copyItemStackView(&views[index])
		if !ok {
			return nil, false
		}
		values[index] = value
	}
	return values, true
}

func nativeBlockPos(position BlockPos) C.DfBlockPos {
	return C.DfBlockPos{x: C.int32_t(position.X), y: C.int32_t(position.Y), z: C.int32_t(position.Z)}
}

func boolByte(value bool) uint8 {
	if value {
		return 1
	}
	return 0
}

func stickyCancellation(incoming, returned bool) bool {
	return incoming || returned
}

func nativeDamageSource(source DamageSource) (C.DfDamageSourceView, func(), bool) {
	allocations := make([]unsafe.Pointer, 0, 4)
	release := func() {
		for _, allocation := range allocations {
			C.free(allocation)
		}
	}
	name := C.CBytes([]byte(source.Name))
	if len(source.Name) != 0 && name == nil {
		return C.DfDamageSourceView{}, func() {}, false
	}
	allocations = append(allocations, name)
	var flags uint32
	if source.ReducedByArmour {
		flags |= C.DF_DAMAGE_SOURCE_REDUCED_BY_ARMOUR
	}
	if source.ReducedByResistance {
		flags |= C.DF_DAMAGE_SOURCE_REDUCED_BY_RESISTANCE
	}
	if source.Fire {
		flags |= C.DF_DAMAGE_SOURCE_FIRE
	}
	if source.IgnoresTotem {
		flags |= C.DF_DAMAGE_SOURCE_IGNORES_TOTEM
	}
	if source.FireProtection {
		flags |= C.DF_DAMAGE_SOURCE_FIRE_PROTECTION
	}
	if source.FeatherFalling {
		flags |= C.DF_DAMAGE_SOURCE_FEATHER_FALLING
	}
	if source.BlastProtection {
		flags |= C.DF_DAMAGE_SOURCE_BLAST_PROTECTION
	}
	if source.ProjectileProtection {
		flags |= C.DF_DAMAGE_SOURCE_PROJECTILE_PROTECTION
	}
	view := C.DfDamageSourceView{
		name: C.DfStringView{data: (*C.uint8_t)(name), len: C.uint64_t(len(source.Name))},
		kind: C.uint32_t(source.Kind), flags: C.uint32_t(flags), data: C.uint8_t(boolByte(source.Data)),
	}
	fillEntityID(&view.entity, source.Entity)
	fillEntityID(&view.secondary_entity, source.SecondaryEntity)
	if source.Block != nil {
		block := (*C.DfBlockView)(C.malloc(C.size_t(C.sizeof_DfBlockView)))
		identifier := C.CBytes([]byte(source.Block.Identifier))
		properties := C.CBytes(source.Block.PropertiesNBT)
		if block == nil || len(source.Block.Identifier) != 0 && identifier == nil || len(source.Block.PropertiesNBT) != 0 && properties == nil {
			C.free(unsafe.Pointer(block))
			C.free(identifier)
			C.free(properties)
			release()
			return C.DfDamageSourceView{}, func() {}, false
		}
		allocations = append(allocations, unsafe.Pointer(block), identifier, properties)
		*block = C.DfBlockView{
			identifier:     C.DfStringView{data: (*C.uint8_t)(identifier), len: C.uint64_t(len(source.Block.Identifier))},
			properties_nbt: C.DfStringView{data: (*C.uint8_t)(properties), len: C.uint64_t(len(source.Block.PropertiesNBT))},
		}
		view.block = block
	}
	return view, release, true
}

func nativeHealingSource(source HealingSource) (C.DfHealingSourceView, func(), bool) {
	name := C.CBytes([]byte(source.Name))
	if len(source.Name) != 0 && name == nil {
		return C.DfHealingSourceView{}, func() {}, false
	}
	return C.DfHealingSourceView{
		name: C.DfStringView{data: (*C.uint8_t)(name), len: C.uint64_t(len(source.Name))},
		kind: C.uint32_t(source.Kind), data: C.uint8_t(boolByte(source.Data)),
	}, func() { C.free(name) }, true
}

func fillPlayerID(destination *C.DfPlayerId, source PlayerID) {
	for i, value := range source.UUID {
		destination.bytes[i] = C.uint8_t(value)
	}
	destination.generation = C.uint64_t(source.Generation)
}

func fillPlayerSnapshot(destination *C.DfPlayerSnapshot, source PlayerSnapshot) unsafe.Pointer {
	name := C.CBytes([]byte(source.Name))
	fillPlayerID(&destination.player, source.Player)
	destination.name = C.DfStringView{data: (*C.uint8_t)(name), len: C.uint64_t(len(source.Name))}
	destination.latency_milliseconds = C.uint64_t(source.LatencyMilliseconds)
	destination.position = C.DfVec3{x: C.double(source.Position.X), y: C.double(source.Position.Y), z: C.double(source.Position.Z)}
	return name
}

// fillBorrowedPlayerSnapshot avoids allocating the stable player name for the hot movement path.
// The containing input must be passed to C by value and source.Name kept alive until the call returns.
func fillBorrowedPlayerSnapshot(destination *C.DfPlayerSnapshot, source PlayerSnapshot) {
	name := unsafe.StringData(source.Name)
	fillPlayerID(&destination.player, source.Player)
	destination.name = C.DfStringView{data: (*C.uint8_t)(unsafe.Pointer(name)), len: C.uint64_t(len(source.Name))}
	destination.latency_milliseconds = C.uint64_t(source.LatencyMilliseconds)
	destination.position = C.DfVec3{x: C.double(source.Position.X), y: C.double(source.Position.Y), z: C.double(source.Position.Z)}
}

func fillEntityID(destination *C.DfEntityId, source EntityID) {
	for i, value := range source.UUID {
		destination.bytes[i] = C.uint8_t(value)
	}
	destination.generation = C.uint64_t(source.Generation)
}

func stringView(view C.DfStringView) string {
	if view.len == 0 {
		return ""
	}
	return string(C.GoBytes(unsafe.Pointer(view.data), C.int(view.len)))
}
