package native

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestStickyCancellation(t *testing.T) {
	for _, test := range []struct {
		name               string
		incoming, returned bool
		want               bool
	}{
		{name: "allowed", want: false},
		{name: "cancelled by native", returned: true, want: true},
		{name: "incoming cancellation", incoming: true, want: true},
		{name: "both cancelled", incoming: true, returned: true, want: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := stickyCancellation(test.incoming, test.returned); got != test.want {
				t.Fatalf("stickyCancellation(%t, %t) = %t, want %t", test.incoming, test.returned, got, test.want)
			}
		})
	}
}

func TestCancellablePlayerCallbacksKeepIncomingCancellationOnRuntimeError(t *testing.T) {
	runtime := &Runtime{}
	tests := []struct {
		name string
		call func() (bool, error)
	}{
		{name: "move", call: func() (bool, error) { return runtime.HandlePlayerMove(0, PlayerMoveInput{}, true) }},
		{name: "chat", call: func() (bool, error) {
			output, err := runtime.HandlePlayerChat(0, PlayerChatInput{}, true)
			return output.Cancelled, err
		}},
		{name: "join", call: func() (bool, error) { return runtime.HandlePlayerJoin(0, PlayerJoinInput{}, true) }},
		{name: "hurt", call: func() (bool, error) {
			output, err := runtime.HandlePlayerHurt(0, PlayerHurtInput{}, true)
			return output.Cancelled, err
		}},
		{name: "heal", call: func() (bool, error) {
			output, err := runtime.HandlePlayerHeal(0, PlayerHealInput{}, true)
			return output.Cancelled, err
		}},
		{name: "block break", call: func() (bool, error) {
			output, err := runtime.HandlePlayerBlockBreak(0, PlayerBlockBreakInput{}, true)
			return output.Cancelled, err
		}},
		{name: "block place", call: func() (bool, error) { return runtime.HandlePlayerBlockPlace(0, PlayerBlockPlaceInput{}, true) }},
		{name: "food loss", call: func() (bool, error) {
			output, err := runtime.HandlePlayerFoodLoss(0, PlayerFoodLossInput{}, true)
			return output.Cancelled, err
		}},
		{name: "start break", call: func() (bool, error) { return runtime.HandlePlayerStartBreak(0, PlayerPositionInput{}, true) }},
		{name: "fire extinguish", call: func() (bool, error) { return runtime.HandlePlayerFireExtinguish(0, PlayerPositionInput{}, true) }},
		{name: "toggle sprint", call: func() (bool, error) { return runtime.HandlePlayerToggleSprint(0, PlayerToggleInput{}, true) }},
		{name: "toggle sneak", call: func() (bool, error) { return runtime.HandlePlayerToggleSneak(0, PlayerToggleInput{}, true) }},
		{name: "teleport", call: func() (bool, error) { return runtime.HandlePlayerTeleport(0, PlayerTeleportInput{}, true) }},
		{name: "experience gain", call: func() (bool, error) {
			output, err := runtime.HandlePlayerExperienceGain(0, PlayerSnapshot{}, 0, true)
			return output.Cancelled, err
		}},
		{name: "punch air", call: func() (bool, error) { return runtime.HandlePlayerPunchAir(0, PlayerSnapshot{}, true) }},
		{name: "held slot change", call: func() (bool, error) { return runtime.HandlePlayerHeldSlotChange(0, PlayerHeldSlotChangeInput{}, true) }},
		{name: "sleep", call: func() (bool, error) {
			output, err := runtime.HandlePlayerSleep(0, PlayerSnapshot{}, false, true)
			return output.Cancelled, err
		}},
		{name: "block pick", call: func() (bool, error) { return runtime.HandlePlayerBlockPick(0, PlayerBlockPickInput{}, true) }},
		{name: "lectern page turn", call: func() (bool, error) {
			output, err := runtime.HandlePlayerLecternPageTurn(0, PlayerLecternPageTurnInput{}, true)
			return output.Cancelled, err
		}},
		{name: "sign edit", call: func() (bool, error) { return runtime.HandlePlayerSignEdit(0, PlayerSignEditInput{}, true) }},
		{name: "item use", call: func() (bool, error) { return runtime.HandlePlayerItemUse(0, PlayerSnapshot{}, true) }},
		{name: "item use on block", call: func() (bool, error) { return runtime.HandlePlayerItemUseOnBlock(0, PlayerItemUseOnBlockInput{}, true) }},
		{name: "item consume", call: func() (bool, error) { return runtime.HandlePlayerItemConsume(0, PlayerSnapshot{}, ItemStack{}, true) }},
		{name: "item release", call: func() (bool, error) {
			return runtime.HandlePlayerItemRelease(0, PlayerSnapshot{}, ItemStack{}, 0, true)
		}},
		{name: "item damage", call: func() (bool, error) {
			output, err := runtime.HandlePlayerItemDamage(0, PlayerSnapshot{}, ItemStack{}, 0, true)
			return output.Cancelled, err
		}},
		{name: "item drop", call: func() (bool, error) { return runtime.HandlePlayerItemDrop(0, PlayerSnapshot{}, ItemStack{}, true) }},
		{name: "item pickup", call: func() (bool, error) {
			output, err := runtime.HandlePlayerItemPickup(0, PlayerItemPickupInput{}, true)
			return output.Cancelled, err
		}},
		{name: "attack entity", call: func() (bool, error) {
			output, err := runtime.HandlePlayerAttackEntity(0, PlayerAttackEntityInput{}, 0, 0, false, true)
			return output.Cancelled, err
		}},
		{name: "item use on entity", call: func() (bool, error) {
			return runtime.HandlePlayerItemUseOnEntity(0, PlayerItemUseOnEntityInput{}, true)
		}},
		{name: "skin change", call: func() (bool, error) {
			output, err := runtime.HandlePlayerSkinChange(0, PlayerSkinChangeInput{}, PlayerSkin{}, true)
			return output.Cancelled, err
		}},
		{name: "transfer", call: func() (bool, error) {
			output, err := runtime.HandlePlayerTransfer(0, PlayerTransferInput{}, true)
			return output.Cancelled, err
		}},
		{name: "command execution", call: func() (bool, error) {
			output, err := runtime.HandlePlayerCommandExecution(0, PlayerCommandExecutionInput{}, true)
			return output.Cancelled, err
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cancelled, err := test.call()
			if err == nil {
				t.Fatal("closed runtime call returned no error")
			}
			if !cancelled {
				t.Fatal("incoming cancellation was cleared on error")
			}
		})
	}
}

func TestValidEntityTypeDefinition(t *testing.T) {
	base := EntityTypeDefinition{
		SaveID: "example:marker", NetworkID: "minecraft:armor_stand", TypeKey: 1,
		Min: Vec3{X: -0.25, Z: -0.25}, Max: Vec3{X: 0.25, Y: 1.0, Z: 0.25},
	}
	ticking := base
	ticking.Family = EntityFamilyTicking
	ticking.CallbackFlags = EntityCallbackState | EntityCallbackTick
	ticking.StateVersion = 1
	ticking.Physics = &EntityPhysics{Gravity: 0.08, Drag: 0.02, DragBeforeGravity: true}
	living := ticking
	living.Family = EntityFamilyLiving
	living.CallbackFlags |= EntityCallbackHurt | EntityCallbackHeal | EntityCallbackDeath
	living.InitialHealth, living.MaxHealth, living.Speed = 20, 40, 0.1

	for name, definition := range map[string]EntityTypeDefinition{
		"base": base, "ticking": ticking, "living": living,
	} {
		t.Run(name, func(t *testing.T) {
			if !validEntityTypeDefinition(definition) {
				t.Fatalf("valid definition rejected: %#v", definition)
			}
		})
	}

	invalid := map[string]EntityTypeDefinition{
		"zero type key":    func() EntityTypeDefinition { value := base; value.TypeKey = 0; return value }(),
		"unknown family":   func() EntityTypeDefinition { value := base; value.Family = 99; return value }(),
		"unknown callback": func() EntityTypeDefinition { value := ticking; value.CallbackFlags |= 1 << 31; return value }(),
		"base callback":    func() EntityTypeDefinition { value := base; value.CallbackFlags = EntityCallbackTick; return value }(),
		"base physics":     func() EntityTypeDefinition { value := base; value.Physics = &EntityPhysics{}; return value }(),
		"ticking hurt":     func() EntityTypeDefinition { value := ticking; value.CallbackFlags |= EntityCallbackHurt; return value }(),
		"state version unused": func() EntityTypeDefinition {
			value := ticking
			value.CallbackFlags &^= EntityCallbackState
			return value
		}(),
		"living no health": func() EntityTypeDefinition { value := living; value.MaxHealth = 0; return value }(),
		"health over max":  func() EntityTypeDefinition { value := living; value.InitialHealth = 41; return value }(),
		"negative speed":   func() EntityTypeDefinition { value := living; value.Speed = -0.1; return value }(),
		"infinite physics": func() EntityTypeDefinition { value := ticking; value.Physics.Gravity = math.Inf(1); return value }(),
		"invalid drag": func() EntityTypeDefinition {
			value := ticking
			value.Physics = &EntityPhysics{Gravity: 0.08, Drag: 1.1}
			return value
		}(),
	}
	for name, definition := range invalid {
		t.Run(name, func(t *testing.T) {
			if validEntityTypeDefinition(definition) {
				t.Fatalf("invalid definition accepted: %#v", definition)
			}
		})
	}
}

type recordingHost struct {
	noopHost
	player             PlayerID
	textPlayers        []PlayerID
	texts              []string
	kinds              []PlayerTextKind
	title              PlayerTitle
	scoreboard         PlayerScoreboard
	scoreboardRemoved  bool
	transforms         []PlayerTransformKind
	vectors            []Vec3
	yaws               []float64
	pitches            []float64
	states             []PlayerStateKind
	values             []PlayerStateValue
	state              PlayerStateValue
	stateValues        map[PlayerStateKind]PlayerStateValue
	rejectStateWrites  bool
	reads              []PlayerStateKind
	experienceLevel    int32
	experienceProgress float64
	experienceCalls    int
	heals              []struct {
		Health float64
		Source HealingSource
	}
	healed float64
	hurts  []struct {
		Damage float64
		Source DamageSource
	}
	hurtResult    PlayerHurtResult
	effectOps     []PlayerEffectOperation
	effects       []PlayerEffect
	activeEffects []PlayerEffect
	entities      []EntityID
	visible       []bool
	skin          PlayerSkin
	setSkins      []PlayerSkin
	inventoryItem ItemStack
	inventorySets []struct {
		Inventory InventoryID
		Slot      uint32
		Item      ItemStack
	}
	inventoryAdds      []ItemStack
	forms              []PlayerForm
	formClosed         bool
	worldOpened        string
	worldDimension     WorldDimension
	worldOpenSpec      WorldOpenSpec
	worldID            WorldID
	worldLookup        string
	worldLookupOK      bool
	worldName          string
	worldBlock         WorldBlock
	worldBlockOK       bool
	worldLiquid        WorldBlock
	worldLiquidOK      bool
	worldBlockPos      BlockPos
	worldBlockSet      WorldBlock
	worldSetOpts       WorldSetOpts
	worldUpdateDelay   int64
	worldBiome         int32
	worldTemperature   float64
	worldRainingAt     bool
	worldSnowingAt     bool
	worldThunderingAt  bool
	worldRaining       bool
	worldThundering    bool
	worldCurrentTick   int64
	worldSaved         bool
	worldUnloaded      bool
	worldTime          int64
	worldSpawn         BlockPos
	entityStateID      EntityID
	entityState        EntityState
	entityPlayerID     EntityID
	entityPlayer       PlayerSnapshot
	entityPlayerCalls  int
	entitySpawns       []EntitySpawn
	spawnedEntity      EntityID
	particleWorldID    WorldID
	particlePositions  []Vec3
	particles          []WorldParticle
	worldSoundID       WorldID
	worldSoundPos      []Vec3
	worldSounds        []WorldSound
	playerSounds       []WorldSound
	transferInvocation InvocationID
	transferPlayer     PlayerID
	transferWorld      WorldID
	transferPosition   Vec3
}

func (h *recordingHost) SendPlayerText(_ InvocationID, player PlayerID, kind PlayerTextKind, message string) bool {
	h.player = player
	h.textPlayers = append(h.textPlayers, player)
	h.kinds = append(h.kinds, kind)
	h.texts = append(h.texts, message)
	return true
}

func (h *recordingHost) SendPlayerTitle(_ InvocationID, player PlayerID, title PlayerTitle) bool {
	h.player, h.title = player, title
	return true
}

func (h *recordingHost) SendPlayerScoreboard(_ InvocationID, player PlayerID, scoreboard PlayerScoreboard) bool {
	h.player, h.scoreboard = player, scoreboard
	return true
}

func (h *recordingHost) RemovePlayerScoreboard(_ InvocationID, player PlayerID) bool {
	h.player, h.scoreboardRemoved = player, true
	return true
}
func (h *recordingHost) SendPlayerForm(_ InvocationID, _ PlayerID, form PlayerForm) bool {
	h.forms = append(h.forms, form)
	return true
}
func (h *recordingHost) ClosePlayerForm(InvocationID, PlayerID) bool {
	h.formClosed = true
	return true
}

func (h *recordingHost) TransformPlayer(_ InvocationID, _ PlayerID, kind PlayerTransformKind, vector Vec3, yaw, pitch float64) bool {
	h.transforms = append(h.transforms, kind)
	h.vectors = append(h.vectors, vector)
	h.yaws = append(h.yaws, yaw)
	h.pitches = append(h.pitches, pitch)
	return true
}

func (h *recordingHost) TransferPlayer(invocation InvocationID, player PlayerID, world WorldID, position Vec3) bool {
	h.transferInvocation = invocation
	h.transferPlayer = player
	h.transferWorld = world
	h.transferPosition = position
	return true
}

func (h *recordingHost) PlayerKinematics(InvocationID, PlayerID) (PlayerKinematics, bool) {
	if h.entityState.Type == "" {
		return PlayerKinematics{}, false
	}
	return PlayerKinematics{
		Position: h.entityState.Position,
		Velocity: h.entityState.Velocity,
		Rotation: h.entityState.Rotation,
	}, true
}

func (h *recordingHost) SetPlayerState(_ InvocationID, _ PlayerID, kind PlayerStateKind, value PlayerStateValue) bool {
	h.states = append(h.states, kind)
	h.values = append(h.values, value)
	return !h.rejectStateWrites
}

func (h *recordingHost) PlayerState(_ InvocationID, _ PlayerID, kind PlayerStateKind) (PlayerStateValue, bool) {
	h.reads = append(h.reads, kind)
	if value, ok := h.stateValues[kind]; ok {
		return value, true
	}
	return h.state, true
}

func (h *recordingHost) SetPlayerExperience(_ InvocationID, _ PlayerID, level int32, progress float64) bool {
	h.experienceLevel = level
	h.experienceProgress = progress
	h.experienceCalls++
	return true
}

func (h *recordingHost) HealPlayer(_ InvocationID, _ PlayerID, health float64, source HealingSource) (float64, bool) {
	h.heals = append(h.heals, struct {
		Health float64
		Source HealingSource
	}{health, source})
	return h.healed, true
}

func (h *recordingHost) HurtPlayer(_ InvocationID, _ PlayerID, damage float64, source DamageSource) (PlayerHurtResult, bool) {
	h.hurts = append(h.hurts, struct {
		Damage float64
		Source DamageSource
	}{damage, source})
	return h.hurtResult, true
}

func (h *recordingHost) ChangePlayerEffect(_ InvocationID, _ PlayerID, operation PlayerEffectOperation, effect PlayerEffect) bool {
	h.effectOps = append(h.effectOps, operation)
	h.effects = append(h.effects, effect)
	for index := range h.activeEffects {
		if h.activeEffects[index].Type != effect.Type {
			continue
		}
		if operation == PlayerEffectRemove {
			h.activeEffects = append(h.activeEffects[:index], h.activeEffects[index+1:]...)
		} else {
			h.activeEffects[index] = effect
		}
		return true
	}
	if operation == PlayerEffectAdd {
		h.activeEffects = append(h.activeEffects, effect)
	}
	return true
}

func (h *recordingHost) PlayerEffects(InvocationID, PlayerID) ([]PlayerEffect, bool) {
	return append([]PlayerEffect(nil), h.activeEffects...), true
}

func (h *recordingHost) ClearPlayerEffects(InvocationID, PlayerID) bool {
	h.activeEffects = nil
	return true
}

func (h *recordingHost) SetPlayerEntityVisible(_ InvocationID, _ PlayerID, entity EntityID, visible bool) bool {
	h.entities = append(h.entities, entity)
	h.visible = append(h.visible, visible)
	return true
}

func (h *recordingHost) PlayerSkin(InvocationID, PlayerID) (PlayerSkin, bool) {
	return h.skin, true
}

func (h *recordingHost) SetPlayerSkin(_ InvocationID, _ PlayerID, skin PlayerSkin) bool {
	h.setSkins = append(h.setSkins, skin)
	return true
}

func (h *recordingHost) InventorySize(InvocationID, InventoryID) (uint32, bool) { return 36, true }
func (h *recordingHost) InventoryItem(InvocationID, InventoryID, uint32) (ItemStack, bool) {
	return h.inventoryItem, true
}
func (h *recordingHost) SetInventoryItem(_ InvocationID, inventory InventoryID, slot uint32, item ItemStack) bool {
	h.inventorySets = append(h.inventorySets, struct {
		Inventory InventoryID
		Slot      uint32
		Item      ItemStack
	}{inventory, slot, item})
	return true
}
func (h *recordingHost) AddInventoryItem(_ InvocationID, _ InventoryID, item ItemStack) (uint32, bool) {
	h.inventoryAdds = append(h.inventoryAdds, item)
	return item.Count, true
}
func (h *recordingHost) ClearInventory(InvocationID, InventoryID) bool { return true }
func (h *recordingHost) HeldItem(InvocationID, PlayerID, uint32) (ItemStack, bool) {
	return h.inventoryItem, true
}
func (h *recordingHost) SetHeldItems(InvocationID, PlayerID, ItemStack, ItemStack) bool {
	return true
}
func (h *recordingHost) SetHeldSlot(InvocationID, PlayerID, uint32) bool { return true }
func (h *recordingHost) OpenWorld(_ InvocationID, name string, dimension WorldDimension) (WorldID, bool) {
	h.worldOpened, h.worldDimension = name, dimension
	return h.worldID, h.worldID != 0
}
func (h *recordingHost) OpenWorldSpec(_ InvocationID, name string, spec WorldOpenSpec) (WorldID, bool) {
	h.worldOpened, h.worldOpenSpec = name, spec
	return h.worldID, h.worldID != 0
}
func (h *recordingHost) WorldByName(_ InvocationID, name string) (WorldID, bool) {
	h.worldLookup = name
	return h.worldID, h.worldLookupOK && h.worldID != 0
}
func (h *recordingHost) CurrentWorld(_ InvocationID) (WorldID, bool) {
	return h.worldID, h.worldID != 0
}
func (h *recordingHost) WorldName(_ InvocationID, id WorldID) (string, bool) {
	return h.worldName, id == h.worldID && h.worldName != ""
}
func (h *recordingHost) WorldBlock(_ InvocationID, id WorldID, position BlockPos) (WorldBlock, bool) {
	h.worldBlockPos = position
	return h.worldBlock, (id == 0 || id == h.worldID) && h.worldBlockOK
}
func (h *recordingHost) WorldLiquid(_ InvocationID, id WorldID, position BlockPos) (WorldBlock, bool, bool) {
	h.worldBlockPos = position
	if id != h.worldID {
		return WorldBlock{}, false, false
	}
	return h.worldLiquid, h.worldLiquidOK, true
}
func (h *recordingHost) SetWorldLiquid(_ InvocationID, id WorldID, position BlockPos, value *WorldBlock) bool {
	h.worldBlockPos = position
	if value == nil {
		h.worldLiquid, h.worldLiquidOK = WorldBlock{}, false
	} else {
		h.worldLiquid, h.worldLiquidOK = *value, true
	}
	return id == 0 || id == h.worldID
}
func (h *recordingHost) SetWorldBlock(_ InvocationID, id WorldID, position BlockPos, value WorldBlock, options WorldSetOpts) bool {
	h.worldBlockPos, h.worldBlockSet, h.worldSetOpts = position, value, options
	return id == 0 || id == h.worldID
}
func (h *recordingHost) ScheduleWorldBlockUpdate(_ InvocationID, id WorldID, position BlockPos, value WorldBlock, delayNanoseconds int64) bool {
	h.worldBlockPos, h.worldBlockSet, h.worldUpdateDelay = position, value, delayNanoseconds
	return id == 0 || id == h.worldID
}
func (h *recordingHost) WorldBiome(_ InvocationID, id WorldID, position BlockPos) (int32, bool) {
	h.worldBlockPos = position
	return h.worldBiome, id == 0 || id == h.worldID
}
func (h *recordingHost) SetWorldBiome(_ InvocationID, id WorldID, position BlockPos, biome int32) bool {
	h.worldBlockPos, h.worldBiome = position, biome
	return id == 0 || id == h.worldID
}
func (h *recordingHost) WorldTemperature(_ InvocationID, id WorldID, position BlockPos) (float64, bool) {
	h.worldBlockPos = position
	return h.worldTemperature, id == 0 || id == h.worldID
}
func (h *recordingHost) WorldRainingAt(_ InvocationID, id WorldID, position BlockPos) (bool, bool) {
	h.worldBlockPos = position
	return h.worldRainingAt, id == 0 || id == h.worldID
}
func (h *recordingHost) WorldSnowingAt(_ InvocationID, id WorldID, position BlockPos) (bool, bool) {
	h.worldBlockPos = position
	return h.worldSnowingAt, id == 0 || id == h.worldID
}
func (h *recordingHost) WorldThunderingAt(_ InvocationID, id WorldID, position BlockPos) (bool, bool) {
	h.worldBlockPos = position
	return h.worldThunderingAt, id == 0 || id == h.worldID
}
func (h *recordingHost) WorldRaining(_ InvocationID, id WorldID) (bool, bool) {
	return h.worldRaining, id == 0 || id == h.worldID
}
func (h *recordingHost) WorldThundering(_ InvocationID, id WorldID) (bool, bool) {
	return h.worldThundering, id == 0 || id == h.worldID
}
func (h *recordingHost) WorldCurrentTick(_ InvocationID, id WorldID) (int64, bool) {
	return h.worldCurrentTick, id == 0 || id == h.worldID
}
func (h *recordingHost) SaveWorld(_ InvocationID, id WorldID) bool {
	h.worldSaved = id == h.worldID
	return h.worldSaved
}
func (h *recordingHost) UnloadWorld(_ InvocationID, id WorldID) bool {
	h.worldUnloaded = id == h.worldID
	return h.worldUnloaded
}
func (h *recordingHost) SetWorldTime(_ InvocationID, _ WorldID, value int64) bool {
	h.worldTime = value
	return true
}
func (h *recordingHost) SetWorldSpawn(_ InvocationID, _ WorldID, position BlockPos) bool {
	h.worldSpawn = position
	return true
}
func (h *recordingHost) WorldSpawn(_ InvocationID, id WorldID) (BlockPos, bool) {
	return h.worldSpawn, id == h.worldID
}
func (h *recordingHost) EntityState(_ InvocationID, id EntityID) (EntityState, bool) {
	h.entityStateID = id
	return h.entityState, h.entityState.Type != ""
}

func (h *recordingHost) EntityPlayer(_ InvocationID, id EntityID) (PlayerSnapshot, bool) {
	h.entityPlayerID = id
	h.entityPlayerCalls++
	return h.entityPlayer, h.entityPlayer.Player.Generation != 0
}
func (h *recordingHost) SpawnWorldEntity(_ InvocationID, id WorldID, value EntitySpawn) (EntityID, bool) {
	h.entitySpawns = append(h.entitySpawns, value)
	return h.spawnedEntity, id == h.worldID && h.spawnedEntity.Generation != 0
}
func (h *recordingHost) OpenWorldEntityIterator(_ InvocationID, id WorldID, _ bool) (EntityIteratorID, bool) {
	return 1, id == h.worldID
}
func (h *recordingHost) NextWorldEntity(_ InvocationID, _ EntityIteratorID) (EntityID, bool, bool) {
	return EntityID{}, false, true
}
func (h *recordingHost) CloseWorldEntities(InvocationID, EntityIteratorID) {}
func (h *recordingHost) EntityHandle(_ InvocationID, entity EntityID) (EntityHandleID, bool) {
	if entity.Generation == 0 {
		return EntityHandleID{}, false
	}
	return EntityHandleID{Value: entity.Generation, Generation: entity.Generation}, true
}
func (h *recordingHost) AddWorldParticle(_ InvocationID, id WorldID, position Vec3, value WorldParticle) bool {
	h.particleWorldID = id
	h.particlePositions = append(h.particlePositions, position)
	h.particles = append(h.particles, value)
	return id == h.worldID
}
func (h *recordingHost) PlayWorldSound(_ InvocationID, id WorldID, position Vec3, value WorldSound) bool {
	h.worldSoundID = id
	h.worldSoundPos = append(h.worldSoundPos, position)
	h.worldSounds = append(h.worldSounds, value)
	return id == h.worldID
}
func (h *recordingHost) PlayPlayerSound(_ InvocationID, player PlayerID, value WorldSound) bool {
	h.player = player
	h.playerSounds = append(h.playerSounds, value)
	return true
}

func TestFormRegistryIsBoundedAndDrained(t *testing.T) {
	host := registerHost(noopHost{})
	player := PlayerID{Generation: 9}
	dropped := 0
	for index := 0; index < maxFormsPerPlayer; index++ {
		if _, ok := registerForm(host, player, func(InvocationID, PlayerSnapshot, bool, []byte) bool { return true }, func() { dropped++ }); !ok {
			t.Fatalf("registration %d rejected", index)
		}
	}
	if _, ok := registerForm(host, player, func(InvocationID, PlayerSnapshot, bool, []byte) bool { return true }, func() { dropped++ }); ok {
		t.Fatal("registration beyond per-player bound accepted")
	}
	drainHostForms(host, false)
	if dropped != maxFormsPerPlayer {
		t.Fatalf("dropped = %d, want %d", dropped, maxFormsPerPlayer)
	}
	if _, ok := registerForm(host, player, func(InvocationID, PlayerSnapshot, bool, []byte) bool { return true }, func() { dropped++ }); !ok {
		t.Fatal("registry did not reopen after non-closing drain")
	}
	unregisterHost(host)
	if dropped != maxFormsPerPlayer+1 {
		t.Fatalf("dropped after close = %d", dropped)
	}
}

func TestFormDrainWaitsForConcurrentDrop(t *testing.T) {
	host := registerHost(noopHost{})
	player := PlayerID{Generation: 10}
	started, release := make(chan struct{}), make(chan struct{})
	id, ok := registerForm(host, player, func(InvocationID, PlayerSnapshot, bool, []byte) bool { return true }, func() { close(started); <-release })
	if !ok {
		t.Fatal("form registration rejected")
	}
	go CancelPlayerForm(id)
	<-started
	drained := make(chan struct{})
	go func() { drainHostForms(host, false); close(drained) }()
	select {
	case <-drained:
		t.Fatal("drain returned while drop callback was in flight")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	select {
	case <-drained:
	case <-time.After(time.Second):
		t.Fatal("drain did not finish")
	}
	unregisterHost(host)
}

func TestClosingFormDrainKeepsRegistrationGateClosed(t *testing.T) {
	host := registerHost(noopHost{})
	drainHostForms(host, true)
	if _, ok := registerForm(host, PlayerID{Generation: 11}, func(InvocationID, PlayerSnapshot, bool, []byte) bool {
		return true
	}, func() {}); ok {
		t.Fatal("form registered after closing drain")
	}
	unregisterHost(host)
}

func TestInactiveHostRejectsNewWork(t *testing.T) {
	host := registerHost(noopHost{})
	t.Cleanup(func() { unregisterHost(host) })
	if _, ok := resolveHost(host); !ok {
		t.Fatal("new host is inactive")
	}
	if !setHostActive(host, false) {
		t.Fatal("could not deactivate host")
	}
	if _, ok := resolveHost(host); ok {
		t.Fatal("inactive host resolved")
	}
	if !setHostActive(host, true) {
		t.Fatal("could not reactivate host")
	}
	if _, ok := resolveHost(host); !ok {
		t.Fatal("reactivated host did not resolve")
	}
}

func TestFormRejectsWrongPlayerAndOversizedResponse(t *testing.T) {
	host := registerHost(noopHost{})
	t.Cleanup(func() { unregisterHost(host) })
	player := PlayerID{Generation: 11}
	dropped := 0
	register := func() uint64 {
		id, ok := registerForm(host, player, func(InvocationID, PlayerSnapshot, bool, []byte) bool {
			t.Fatal("invalid response reached plugin callback")
			return true
		}, func() { dropped++ })
		if !ok {
			t.Fatal("form registration rejected")
		}
		return id
	}
	if CompletePlayerForm(register(), 0, PlayerSnapshot{Player: PlayerID{Generation: 12}}, false, []byte("0")) {
		t.Fatal("response from wrong player accepted")
	}
	if CompletePlayerForm(register(), 0, PlayerSnapshot{Player: player}, false, make([]byte, maxFormJSONBytes+1)) {
		t.Fatal("oversized response accepted")
	}
	if dropped != 2 {
		t.Fatalf("dropped = %d, want 2", dropped)
	}
}

func TestSkinSnapshotsAreInvocationScopedAndOwned(t *testing.T) {
	host := registerHost(noopHost{})
	t.Cleanup(func() { unregisterHost(host) })
	original := PlayerSkin{
		Width: 1, Height: 1, Pixels: []byte{1, 2, 3, 4},
		Animations: []SkinAnimation{{Width: 1, Height: 1, FrameCount: 1, Pixels: []byte{5, 6, 7, 8}}},
	}
	normal, ok := registerSkinSnapshot(host, 11, original, false)
	if !ok {
		t.Fatal("normal snapshot was not registered")
	}
	original.Pixels[0] = 99
	if _, ok := resolveSkinSnapshot(host+1, 11, normal); ok {
		t.Fatal("snapshot resolved for wrong host")
	}
	if _, ok := resolveSkinSnapshot(host, 12, normal); ok {
		t.Fatal("snapshot resolved for wrong invocation")
	}
	resolved, ok := resolveSkinSnapshot(host, 11, normal)
	if !ok || resolved.Pixels[0] != 1 || resolved.Animations[0].Pixels[0] != 5 {
		t.Fatalf("snapshot did not deep-copy input: %#v", resolved)
	}
	resolved.Pixels[0], resolved.Animations[0].Pixels[0] = 88, 77
	resolvedAgain, _ := resolveSkinSnapshot(host, 11, normal)
	if resolvedAgain.Pixels[0] != 1 || resolvedAgain.Animations[0].Pixels[0] != 5 {
		t.Fatal("resolved skin aliased registry storage")
	}
	if replaceEventSkinSnapshot(host, 11, normal, original) {
		t.Fatal("ordinary snapshot was mutable")
	}
	unregisterSkinSnapshot(host, 12, normal)
	if _, ok := resolveSkinSnapshot(host, 11, normal); !ok {
		t.Fatal("wrong invocation closed snapshot")
	}
	unregisterSkinSnapshot(host, 11, normal)
	if _, ok := resolveSkinSnapshot(host, 11, normal); ok {
		t.Fatal("ordinary snapshot remained after close")
	}

	event, ok := registerSkinSnapshot(host, 21, original, true)
	if !ok {
		t.Fatal("event snapshot was not registered")
	}
	unregisterSkinSnapshot(host, 21, event)
	if _, ok := resolveSkinSnapshot(host, 21, event); !ok {
		t.Fatal("guest closed event-owned snapshot")
	}
	replacement := clonePlayerSkin(original)
	replacement.Pixels[0] = 42
	if !replaceEventSkinSnapshot(host, 21, event, replacement) {
		t.Fatal("event snapshot replacement failed")
	}
	replacement.Pixels[0] = 43
	got, _ := resolveSkinSnapshot(host, 21, event)
	if got.Pixels[0] != 42 {
		t.Fatal("event replacement aliased caller storage")
	}
	forceUnregisterSkinSnapshot(host, 21, event)
	if _, ok := resolveSkinSnapshot(host, 21, event); ok {
		t.Fatal("event snapshot remained after owner cleanup")
	}
}

func commandNamed(t *testing.T, commands []Command, name string) Command {
	t.Helper()
	for _, command := range commands {
		if command.Name == name {
			return command
		}
	}
	t.Fatalf("command %q not found in %#v", name, commands)
	return Command{}
}

func commandArguments(value string) []string { return strings.Fields(value) }

//go:noinline
func rawGoMovement(input PlayerMoveInput, cancelled *bool) {
	if input.NewPosition.Y < 0 {
		*cancelled = true
	}
}

func BenchmarkRawGoMovement(b *testing.B) {
	input := PlayerMoveInput{NewPosition: Vec3{Y: 64}}
	for b.Loop() {
		cancelled := false
		rawGoMovement(input, &cancelled)
	}
}
