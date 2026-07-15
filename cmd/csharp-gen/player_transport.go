package main

import (
	"bytes"
	"fmt"
	"sort"
)

// Transport IDs are private ABI, not Dragonfly registry order. Keep every ID
// explicit so adding or reordering an AST-discovered API cannot renumber it.
type playerTransportSpec struct {
	StateMethods        []playerStateMethod
	ActionMethods       []playerActionMethod
	Controls            playerControlSpec
	IdentityMethods     []string
	PresentationMethods []playerPresentationMethod
	TextMethods         []method
	Effects             effectSpec
	Sounds              []soundTypeSpec
	GameModeMethods     []commandMethod
}

type transportNameID struct {
	Name   string
	GoName string
	ID     uint32
}

var playerStateTransportIDs = []transportNameID{
	{Name: "GameMode", ID: 0},
	{Name: "Food", ID: 3},
	{Name: "MaxHealth", ID: 4},
	{Name: "Health", ID: 5},
	{Name: "ExperienceLevel", ID: 6},
	{Name: "ExperienceProgress", ID: 7},
	{Name: "Scale", ID: 8},
	{Name: "Invisible", ID: 9},
	{Name: "Immobile", ID: 10},
	{Name: "Speed", ID: 11},
	{Name: "FlightSpeed", ID: 12},
	{Name: "VerticalFlightSpeed", ID: 13},
	{Name: "FallDistance", ID: 14},
	{Name: "Absorption", ID: 15},
	{Name: "Dead", ID: 16},
	{Name: "OnGround", ID: 17},
	{Name: "EyeHeight", ID: 18},
	{Name: "TorsoHeight", ID: 19},
	{Name: "Breathing", ID: 20},
	{Name: "Sprinting", ID: 21},
	{Name: "Sneaking", ID: 22},
	{Name: "Swimming", ID: 23},
	{Name: "Crawling", ID: 24},
	{Name: "Gliding", ID: 25},
	{Name: "Flying", ID: 26},
	{Name: "OnFireDuration", ID: 27},
	{Name: "FireProof", ID: 28},
	{Name: "AirSupply", ID: 29},
	{Name: "MaxAirSupply", ID: 30},
	{Name: "Experience", ID: 31},
	{Name: "EnchantmentSeed", ID: 32},
	{Name: "CanCollectExperience", ID: 33},
}

var playerActionTransportIDs = []transportNameID{
	{Name: "AddFood", ID: 0},
	{Name: "Saturate", ID: 1},
	{Name: "Exhaust", ID: 2},
	{Name: "ResetEnchantmentSeed", ID: 3},
	{Name: "AddExperience", ID: 4},
	{Name: "RemoveExperience", ID: 5},
	{Name: "CollectExperience", ID: 6},
	{Name: "EnableInstantRespawn", ID: 7},
	{Name: "DisableInstantRespawn", ID: 8},
	{Name: "ShowCoordinates", ID: 9},
	{Name: "HideCoordinates", ID: 10},
	{Name: "SendSleepingIndicator", ID: 11},
	{Name: "CloseDialogue", ID: 12},
	{Name: "RemoveBossBar", ID: 13},
	{Name: "AbortBreaking", ID: 14},
	{Name: "ClearInputLocks", ID: 15},
	{Name: "FinishBreaking", ID: 16},
	{Name: "Jump", ID: 17},
	{Name: "MoveItemsToInventory", ID: 18},
	{Name: "PunchAir", ID: 19},
	{Name: "ReleaseItem", ID: 20},
	{Name: "RemoveAllDebugShapes", ID: 21},
	{Name: "SwingArm", ID: 22},
	{Name: "UseItem", ID: 23},
	{Name: "Wake", ID: 24},
	{Name: "ShowHudElement", ID: 25},
	{Name: "HideHudElement", ID: 26},
	{Name: "HudElementHidden", ID: 27},
	{Name: "LockInput", ID: 28},
	{Name: "UnlockInput", ID: 29},
	{Name: "InputLocked", ID: 30},
}

var playerStringTransportIDs = []transportNameID{
	{Name: "NameTag", ID: 0},
	{Name: "ScoreTag", ID: 1},
	{Name: "DeviceID", ID: 2},
	{Name: "DeviceModel", ID: 3},
	{Name: "SelfSignedID", ID: 4},
	{Name: "Locale", ID: 5},
	{Name: "AddrNetwork", ID: 6},
	{Name: "AddrString", ID: 7},
}

var playerStateTransportASTMethods = []string{
	"Food", "SetFood", "AddFood", "Saturate", "Exhaust", "Health", "MaxHealth", "SetMaxHealth", "Heal", "Hurt",
	"ExperienceLevel", "SetExperienceLevel", "ExperienceProgress", "SetExperienceProgress",
	"Experience", "EnchantmentSeed", "ResetEnchantmentSeed", "AddExperience", "RemoveExperience", "CanCollectExperience", "CollectExperience",
	"Scale", "SetScale", "Invisible", "SetInvisible", "SetVisible", "Immobile", "SetImmobile", "SetMobile",
	"Speed", "SetSpeed", "FlightSpeed", "SetFlightSpeed", "VerticalFlightSpeed", "SetVerticalFlightSpeed",
	"ResetFallDistance", "FallDistance", "SetAbsorption", "Absorption", "Dead", "OnGround", "EyeHeight", "TorsoHeight", "Breathing",
	"StartSprinting", "StopSprinting", "Sprinting", "StartSneaking", "StopSneaking", "Sneaking",
	"StartSwimming", "StopSwimming", "Swimming", "StartCrawling", "StopCrawling", "Crawling",
	"StartGliding", "StopGliding", "Gliding", "StartFlying", "StopFlying", "Flying",
	"FireProof", "OnFireDuration", "SetOnFire", "Extinguish",
	"AirSupply", "SetAirSupply", "MaxAirSupply", "SetMaxAirSupply",
}

var playerTextTransportIDs = []transportNameID{
	{Name: "Message", ID: 0},
	{Name: "Tip", ID: 1},
	{Name: "Popup", ID: 2},
	{Name: "JukeboxPopup", ID: 3},
	{Name: "NameTag", ID: 4},
	{Name: "Disconnect", ID: 5},
}

var playerTextTransportASTMethods = []string{
	"Message", "SendPopup", "SendTip", "SendJukeboxPopup", "SetNameTag", "Disconnect",
}

var effectTransportIDs = []transportNameID{
	{Name: "Speed", ID: 1},
	{Name: "Slowness", ID: 2},
	{Name: "Haste", ID: 3},
	{Name: "MiningFatigue", ID: 4},
	{Name: "Strength", ID: 5},
	{Name: "InstantHealth", ID: 6},
	{Name: "InstantDamage", ID: 7},
	{Name: "JumpBoost", ID: 8},
	{Name: "Nausea", ID: 9},
	{Name: "Regeneration", ID: 10},
	{Name: "Resistance", ID: 11},
	{Name: "FireResistance", ID: 12},
	{Name: "WaterBreathing", ID: 13},
	{Name: "Invisibility", ID: 14},
	{Name: "Blindness", ID: 15},
	{Name: "NightVision", ID: 16},
	{Name: "Hunger", ID: 17},
	{Name: "Weakness", ID: 18},
	{Name: "Poison", ID: 19},
	{Name: "Wither", ID: 20},
	{Name: "HealthBoost", ID: 21},
	{Name: "Absorption", ID: 22},
	{Name: "Saturation", ID: 23},
	{Name: "Levitation", ID: 24},
	{Name: "FatalPoison", ID: 25},
	{Name: "ConduitPower", ID: 26},
	{Name: "SlowFalling", ID: 27},
	{Name: "Darkness", ID: 30},
}

var soundTransportIDs = []transportNameID{
	{Name: "AnvilBreak", ID: 0},
	{Name: "AnvilLand", ID: 1},
	{Name: "AnvilUse", ID: 2},
	{Name: "ArrowHit", ID: 3},
	{Name: "BarrelClose", ID: 4},
	{Name: "BarrelOpen", ID: 5},
	{Name: "BlastFurnaceCrackle", ID: 6},
	{Name: "BowShoot", ID: 7},
	{Name: "Burning", ID: 8},
	{Name: "Burp", ID: 9},
	{Name: "CampfireCrackle", ID: 10},
	{Name: "ChestClose", ID: 11},
	{Name: "ChestOpen", ID: 12},
	{Name: "Click", ID: 13},
	{Name: "ComposterEmpty", ID: 14},
	{Name: "ComposterFill", ID: 15},
	{Name: "ComposterFillLayer", ID: 16},
	{Name: "ComposterReady", ID: 17},
	{Name: "CopperScraped", ID: 18},
	{Name: "CrossbowShoot", ID: 19},
	{Name: "DecoratedPotInsertFailed", ID: 20},
	{Name: "Deny", ID: 21},
	{Name: "DoorCrash", ID: 22},
	{Name: "Drowning", ID: 23},
	{Name: "EnderChestClose", ID: 24},
	{Name: "EnderChestOpen", ID: 25},
	{Name: "Experience", ID: 26},
	{Name: "Explosion", ID: 27},
	{Name: "FireCharge", ID: 28},
	{Name: "FireExtinguish", ID: 29},
	{Name: "FireworkBlast", ID: 30},
	{Name: "FireworkHugeBlast", ID: 31},
	{Name: "FireworkLaunch", ID: 32},
	{Name: "FireworkTwinkle", ID: 33},
	{Name: "Fizz", ID: 34},
	{Name: "FurnaceCrackle", ID: 35},
	{Name: "GhastShoot", ID: 36},
	{Name: "GhastWarning", ID: 37},
	{Name: "GlassBreak", ID: 38},
	{Name: "Ignite", ID: 39},
	{Name: "ItemAdd", ID: 40},
	{Name: "ItemBreak", ID: 41},
	{Name: "ItemFrameRemove", ID: 42},
	{Name: "ItemFrameRotate", ID: 43},
	{Name: "ItemThrow", ID: 44},
	{Name: "LecternBookPlace", ID: 45},
	{Name: "LevelUp", ID: 46},
	{Name: "LightningExplode", ID: 47},
	{Name: "LightningThunder", ID: 48},
	{Name: "MusicDiscEnd", ID: 49},
	{Name: "Pop", ID: 50},
	{Name: "PotionBrewed", ID: 51},
	{Name: "PowerOff", ID: 52},
	{Name: "PowerOn", ID: 53},
	{Name: "SignWaxed", ID: 54},
	{Name: "SmokerCrackle", ID: 55},
	{Name: "StopUsingSpyglass", ID: 56},
	{Name: "TNT", GoName: "Tnt", ID: 57},
	{Name: "Teleport", ID: 58},
	{Name: "Thunder", ID: 59},
	{Name: "Totem", ID: 60},
	{Name: "UseSpyglass", ID: 61},
	{Name: "WaxRemoved", ID: 62},
	{Name: "WaxedSignFailedInteraction", ID: 63},
	{Name: "ShulkerBoxOpen", ID: 64},
	{Name: "ShulkerBoxClose", ID: 65},
	{Name: "EnderEyePlaced", ID: 66},
	{Name: "EndPortalCreated", ID: 67},
	{Name: "Attack", ID: 68},
	{Name: "Fall", ID: 69},
	{Name: "BlockPlace", ID: 70},
	{Name: "BlockBreaking", ID: 71},
	{Name: "DoorOpen", ID: 72},
	{Name: "DoorClose", ID: 73},
	{Name: "TrapdoorOpen", ID: 74},
	{Name: "TrapdoorClose", ID: 75},
	{Name: "FenceGateOpen", ID: 76},
	{Name: "FenceGateClose", ID: 77},
	{Name: "Note", ID: 78},
	{Name: "MusicDiscPlay", ID: 79},
	{Name: "DecoratedPotInserted", ID: 80},
	{Name: "ItemUseOn", ID: 81},
	{Name: "EquipItem", ID: 82},
	{Name: "BucketFill", ID: 83},
	{Name: "BucketEmpty", ID: 84},
	{Name: "CrossbowLoad", ID: 85},
	{Name: "GoatHorn", ID: 86},
}

func generateNativePlayerTransport(spec playerTransportSpec) ([]byte, error) {
	if err := validatePlayerTransportSpec(spec); err != nil {
		return nil, err
	}
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly Go AST and live registries by csharp-gen. DO NOT EDIT.\n\npackage native\n\nimport \"time\"\n\n")
	output.WriteString("type PlayerStateKind uint32\n\nconst (\n")
	for _, entry := range playerStateTransportIDs {
		fmt.Fprintf(&output, "\t%-30s PlayerStateKind = %d\n", "PlayerState"+entry.Name, entry.ID)
	}
	output.WriteString(`)

type PlayerStateValue struct {
	Number  float64
	Integer int64
}

type PlayerActionKind uint32

const (
`)
	for _, entry := range playerActionTransportIDs {
		fmt.Fprintf(&output, "\t%-37s PlayerActionKind = %d\n", "PlayerAction"+entry.Name, entry.ID)
	}
	output.WriteString(`)

type PlayerStringKind uint32

const (
`)
	for _, entry := range playerStringTransportIDs {
		fmt.Fprintf(&output, "\t%-27s PlayerStringKind = %d\n", "PlayerString"+entry.Name, entry.ID)
	}
	output.WriteString(`)

type EffectType int32

const (
`)
	for _, entry := range effectTransportIDs {
		fmt.Fprintf(&output, "\t%-20s EffectType = %d\n", "Effect"+entry.Name, entry.ID)
	}
	output.WriteString(`)

type PlayerEffectOperation uint32

const (
	PlayerEffectAdd PlayerEffectOperation = iota
	PlayerEffectRemove
)

type PlayerEffect struct {
	Type            EffectType
	Level           int32
	Duration        time.Duration
	Potency         float64
	Ambient         bool
	ParticlesHidden bool
	Infinite        bool
	Tick            int64
}

type PlayerTextKind uint32

const (
`)
	for _, entry := range playerTextTransportIDs {
		fmt.Fprintf(&output, "\t%-22s PlayerTextKind = %d\n", "PlayerText"+entry.Name, entry.ID)
	}
	output.WriteString(`)

type SoundKind uint32

const (
`)
	for _, entry := range soundTransportIDs {
		name := entry.GoName
		if name == "" {
			name = entry.Name
		}
		fmt.Fprintf(&output, "\t%-31s SoundKind = %d\n", "Sound"+name, entry.ID)
	}
	output.WriteString(")\n")
	return output.Bytes(), nil
}

func generateCSharpPlayerStateTransport(spec playerTransportSpec) ([]byte, error) {
	if err := validatePlayerTransportSpec(spec); err != nil {
		return nil, err
	}
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly Go AST by csharp-gen. DO NOT EDIT.\n\nnamespace Dragonfly.Native;\n\npublic static partial class Abi\n{\n")
	for _, entry := range playerStateTransportIDs {
		fmt.Fprintf(&output, "    public const uint PlayerState%s = %d;\n", entry.Name, entry.ID)
	}
	for _, entry := range playerActionTransportIDs {
		fmt.Fprintf(&output, "    public const uint PlayerAction%s = %d;\n", entry.Name, entry.ID)
	}
	for _, entry := range playerStringTransportIDs {
		fmt.Fprintf(&output, "    public const uint PlayerString%s = %d;\n", entry.Name, entry.ID)
	}
	output.WriteString("}\n")
	return output.Bytes(), nil
}

func generateHostPlayerTransport(spec playerTransportSpec) ([]byte, error) {
	if err := validatePlayerTransportSpec(spec); err != nil {
		return nil, err
	}
	var output bytes.Buffer
	output.WriteString(`// Code generated from Dragonfly Go AST and live registries by csharp-gen. DO NOT EDIT.

package host

import (
	"math"
	"time"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/hud"
	"github.com/df-mc/dragonfly/server/player/input"
)

func sendPlayerText(connected *player.Player, kind native.PlayerTextKind, message string) bool {
	switch kind {
`)
	for _, entry := range playerTextTransportIDs {
		method := map[string]string{
			"Message": "Message", "Tip": "SendTip", "Popup": "SendPopup",
			"JukeboxPopup": "SendJukeboxPopup", "NameTag": "SetNameTag",
			"Disconnect": "Disconnect",
		}[entry.Name]
		fmt.Fprintf(&output, "\tcase native.PlayerText%s:\n\t\tconnected.%s(message)\n", entry.Name, method)
	}
	output.WriteString(`	default:
		return false
	}
	return true
}

func setPlayerState(connected *player.Player, kind native.PlayerStateKind, value native.PlayerStateValue) bool {
	switch kind {
	case native.PlayerStateGameMode:
		mode, ok := decodeGameModeDescriptor(value.Integer)
		if !ok {
			return false
		}
		connected.SetGameMode(mode)
	case native.PlayerStateFood:
		if value.Integer < math.MinInt32 || value.Integer > math.MaxInt32 {
			return false
		}
		connected.SetFood(int(value.Integer))
	case native.PlayerStateMaxHealth:
		connected.SetMaxHealth(value.Number)
	case native.PlayerStateExperienceLevel:
		if value.Integer < math.MinInt32 || value.Integer > math.MaxInt32 || value.Integer < 0 {
			return false
		}
		connected.SetExperienceLevel(int(value.Integer))
	case native.PlayerStateExperienceProgress:
		if value.Number < 0 || value.Number > 1 {
			return false
		}
		connected.SetExperienceProgress(value.Number)
	case native.PlayerStateScale:
		connected.SetScale(value.Number)
	case native.PlayerStateInvisible:
		if value.Integer != 0 && value.Integer != 1 {
			return false
		}
		if value.Integer != 0 {
			connected.SetInvisible()
		} else {
			connected.SetVisible()
		}
	case native.PlayerStateImmobile:
		if value.Integer != 0 && value.Integer != 1 {
			return false
		}
		if value.Integer != 0 {
			connected.SetImmobile()
		} else {
			connected.SetMobile()
		}
	case native.PlayerStateSpeed:
		connected.SetSpeed(value.Number)
	case native.PlayerStateFlightSpeed:
		connected.SetFlightSpeed(value.Number)
	case native.PlayerStateVerticalFlightSpeed:
		connected.SetVerticalFlightSpeed(value.Number)
	case native.PlayerStateFallDistance:
		connected.ResetFallDistance()
	case native.PlayerStateAbsorption:
		connected.SetAbsorption(value.Number)
	case native.PlayerStateSprinting:
		if !setPlayerActivity(value.Integer, connected.StartSprinting, connected.StopSprinting) {
			return false
		}
	case native.PlayerStateSneaking:
		if !setPlayerActivity(value.Integer, connected.StartSneaking, connected.StopSneaking) {
			return false
		}
	case native.PlayerStateSwimming:
		if !setPlayerActivity(value.Integer, connected.StartSwimming, connected.StopSwimming) {
			return false
		}
	case native.PlayerStateCrawling:
		if !setPlayerActivity(value.Integer, connected.StartCrawling, connected.StopCrawling) {
			return false
		}
	case native.PlayerStateGliding:
		if !setPlayerActivity(value.Integer, connected.StartGliding, connected.StopGliding) {
			return false
		}
	case native.PlayerStateFlying:
		if !setPlayerActivity(value.Integer, connected.StartFlying, connected.StopFlying) {
			return false
		}
	case native.PlayerStateOnFireDuration:
		connected.SetOnFire(time.Duration(value.Integer))
	case native.PlayerStateAirSupply:
		connected.SetAirSupply(time.Duration(value.Integer))
	case native.PlayerStateMaxAirSupply:
		connected.SetMaxAirSupply(time.Duration(value.Integer))
	default:
		return false
	}
	return true
}

func runPlayerAction(connected *player.Player, kind native.PlayerActionKind, value native.PlayerStateValue) (native.PlayerStateValue, bool) {
	switch kind {
	case native.PlayerActionAddFood:
		connected.AddFood(int(value.Integer))
	case native.PlayerActionSaturate:
		connected.Saturate(int(value.Integer), value.Number)
	case native.PlayerActionExhaust:
		connected.Exhaust(value.Number)
	case native.PlayerActionResetEnchantmentSeed:
		connected.ResetEnchantmentSeed()
	case native.PlayerActionAddExperience:
		return native.PlayerStateValue{Integer: int64(connected.AddExperience(int(value.Integer)))}, true
	case native.PlayerActionRemoveExperience:
		connected.RemoveExperience(int(value.Integer))
	case native.PlayerActionCollectExperience:
		return native.PlayerStateValue{Integer: boolInteger(connected.CollectExperience(int(value.Integer)))}, true
	case native.PlayerActionEnableInstantRespawn:
		connected.EnableInstantRespawn()
	case native.PlayerActionDisableInstantRespawn:
		connected.DisableInstantRespawn()
	case native.PlayerActionShowCoordinates:
		connected.ShowCoordinates()
	case native.PlayerActionHideCoordinates:
		connected.HideCoordinates()
	case native.PlayerActionSendSleepingIndicator:
		if value.Integer < math.MinInt32 || value.Integer > math.MaxInt32 ||
			value.Number < math.MinInt32 || value.Number > math.MaxInt32 || math.Trunc(value.Number) != value.Number {
			return native.PlayerStateValue{}, false
		}
		connected.SendSleepingIndicator(int(value.Integer), int(value.Number))
	case native.PlayerActionCloseDialogue:
		connected.CloseDialogue()
	case native.PlayerActionRemoveBossBar:
		connected.RemoveBossBar()
	case native.PlayerActionAbortBreaking:
		connected.AbortBreaking()
	case native.PlayerActionClearInputLocks:
		connected.ClearInputLocks()
	case native.PlayerActionFinishBreaking:
		connected.FinishBreaking()
	case native.PlayerActionJump:
		connected.Jump()
	case native.PlayerActionMoveItemsToInventory:
		connected.MoveItemsToInventory()
	case native.PlayerActionPunchAir:
		connected.PunchAir()
	case native.PlayerActionReleaseItem:
		connected.ReleaseItem()
	case native.PlayerActionRemoveAllDebugShapes:
		connected.RemoveAllDebugShapes()
	case native.PlayerActionSwingArm:
		connected.SwingArm()
	case native.PlayerActionUseItem:
		connected.UseItem()
	case native.PlayerActionWake:
		connected.Wake()
	case native.PlayerActionShowHudElement:
		element, ok := playerHudElement(value.Integer)
		if !ok {
			return native.PlayerStateValue{}, false
		}
		connected.ShowHudElement(element)
	case native.PlayerActionHideHudElement:
		element, ok := playerHudElement(value.Integer)
		if !ok {
			return native.PlayerStateValue{}, false
		}
		connected.HideHudElement(element)
	case native.PlayerActionHudElementHidden:
		element, ok := playerHudElement(value.Integer)
		if !ok {
			return native.PlayerStateValue{}, false
		}
		return native.PlayerStateValue{Integer: boolInteger(connected.HudElementHidden(element))}, true
	case native.PlayerActionLockInput:
		lock, ok := playerInputLock(value.Integer)
		if !ok {
			return native.PlayerStateValue{}, false
		}
		connected.LockInput(lock)
	case native.PlayerActionUnlockInput:
		lock, ok := playerInputLock(value.Integer)
		if !ok {
			return native.PlayerStateValue{}, false
		}
		connected.UnlockInput(lock)
	case native.PlayerActionInputLocked:
		lock, ok := playerInputLock(value.Integer)
		if !ok {
			return native.PlayerStateValue{}, false
		}
		return native.PlayerStateValue{Integer: boolInteger(connected.InputLocked(lock))}, true
	default:
		return native.PlayerStateValue{}, false
	}
	return native.PlayerStateValue{}, true
}

func playerHudElement(value int64) (hud.Element, bool) {
	switch value {
`)
	for _, value := range spec.Controls.HudElements {
		fmt.Fprintf(&output, "\tcase %d:\n\t\treturn hud.%s(), true\n", value.Value, value.Name)
	}
	output.WriteString(`	default:
		return hud.Element{}, false
	}
}

func playerInputLock(value int64) (input.Lock, bool) {
	switch value {
`)
	for _, value := range spec.Controls.InputLocks {
		fmt.Fprintf(&output, "\tcase %d:\n\t\treturn input.%s(), true\n", value.Value, value.Name)
	}
	output.WriteString(`	default:
		return input.Lock{}, false
	}
}

func playerString(connected *player.Player, kind native.PlayerStringKind) (string, bool) {
	switch kind {
	case native.PlayerStringNameTag:
		return connected.NameTag(), true
	case native.PlayerStringScoreTag:
		return connected.ScoreTag(), true
	case native.PlayerStringDeviceID:
		return connected.DeviceID(), true
	case native.PlayerStringDeviceModel:
		return connected.DeviceModel(), true
	case native.PlayerStringSelfSignedID:
		return connected.SelfSignedID(), true
	case native.PlayerStringLocale:
		return connected.Locale().String(), true
	case native.PlayerStringAddrNetwork:
		if address := connected.Addr(); address != nil {
			return address.Network(), true
		}
		return "", false
	case native.PlayerStringAddrString:
		if address := connected.Addr(); address != nil {
			return address.String(), true
		}
		return "", false
	default:
		return "", false
	}
}

func readPlayerState(connected *player.Player, kind native.PlayerStateKind) (native.PlayerStateValue, bool) {
	switch kind {
	case native.PlayerStateGameMode:
		value, ok := encodeGameModeDescriptor(connected.GameMode())
		return native.PlayerStateValue{Integer: value}, ok
	case native.PlayerStateFood:
		return native.PlayerStateValue{Integer: int64(connected.Food())}, true
	case native.PlayerStateMaxHealth:
		return native.PlayerStateValue{Number: connected.MaxHealth()}, true
	case native.PlayerStateHealth:
		return native.PlayerStateValue{Number: connected.Health()}, true
	case native.PlayerStateExperienceLevel:
		return native.PlayerStateValue{Integer: int64(connected.ExperienceLevel())}, true
	case native.PlayerStateExperienceProgress:
		return native.PlayerStateValue{Number: connected.ExperienceProgress()}, true
	case native.PlayerStateScale:
		return native.PlayerStateValue{Number: connected.Scale()}, true
	case native.PlayerStateInvisible:
		if connected.Invisible() {
			return native.PlayerStateValue{Integer: 1}, true
		}
		return native.PlayerStateValue{}, true
	case native.PlayerStateImmobile:
		if connected.Immobile() {
			return native.PlayerStateValue{Integer: 1}, true
		}
		return native.PlayerStateValue{}, true
	case native.PlayerStateSpeed:
		return native.PlayerStateValue{Number: connected.Speed()}, true
	case native.PlayerStateFlightSpeed:
		return native.PlayerStateValue{Number: connected.FlightSpeed()}, true
	case native.PlayerStateVerticalFlightSpeed:
		return native.PlayerStateValue{Number: connected.VerticalFlightSpeed()}, true
	case native.PlayerStateFallDistance:
		return native.PlayerStateValue{Number: connected.FallDistance()}, true
	case native.PlayerStateAbsorption:
		return native.PlayerStateValue{Number: connected.Absorption()}, true
	case native.PlayerStateDead:
		return native.PlayerStateValue{Integer: boolInteger(connected.Dead())}, true
	case native.PlayerStateOnGround:
		return native.PlayerStateValue{Integer: boolInteger(connected.OnGround())}, true
	case native.PlayerStateEyeHeight:
		return native.PlayerStateValue{Number: connected.EyeHeight()}, true
	case native.PlayerStateTorsoHeight:
		return native.PlayerStateValue{Number: connected.TorsoHeight()}, true
	case native.PlayerStateBreathing:
		return native.PlayerStateValue{Integer: boolInteger(connected.Breathing())}, true
	case native.PlayerStateSprinting:
		return native.PlayerStateValue{Integer: boolInteger(connected.Sprinting())}, true
	case native.PlayerStateSneaking:
		return native.PlayerStateValue{Integer: boolInteger(connected.Sneaking())}, true
	case native.PlayerStateSwimming:
		return native.PlayerStateValue{Integer: boolInteger(connected.Swimming())}, true
	case native.PlayerStateCrawling:
		return native.PlayerStateValue{Integer: boolInteger(connected.Crawling())}, true
	case native.PlayerStateGliding:
		return native.PlayerStateValue{Integer: boolInteger(connected.Gliding())}, true
	case native.PlayerStateFlying:
		return native.PlayerStateValue{Integer: boolInteger(connected.Flying())}, true
	case native.PlayerStateOnFireDuration:
		return native.PlayerStateValue{Integer: int64(connected.OnFireDuration())}, true
	case native.PlayerStateFireProof:
		return native.PlayerStateValue{Integer: boolInteger(connected.FireProof())}, true
	case native.PlayerStateAirSupply:
		return native.PlayerStateValue{Integer: int64(connected.AirSupply())}, true
	case native.PlayerStateMaxAirSupply:
		return native.PlayerStateValue{Integer: int64(connected.MaxAirSupply())}, true
	case native.PlayerStateExperience:
		return native.PlayerStateValue{Integer: int64(connected.Experience())}, true
	case native.PlayerStateEnchantmentSeed:
		return native.PlayerStateValue{Integer: connected.EnchantmentSeed()}, true
	case native.PlayerStateCanCollectExperience:
		return native.PlayerStateValue{Integer: boolInteger(connected.CanCollectExperience())}, true
	default:
		return native.PlayerStateValue{}, false
	}
}

func boolInteger(value bool) int64 {
	if value {
		return 1
	}
	return 0
}

func setPlayerActivity(value int64, start, stop func()) bool {
	if value != 0 && value != 1 {
		return false
	}
	if value == 1 {
		start()
	} else {
		stop()
	}
	return true
}
`)
	return output.Bytes(), nil
}

func validatePlayerTransportSpec(spec playerTransportSpec) error {
	stateNames := make([]string, 0, len(spec.StateMethods))
	for _, value := range spec.StateMethods {
		stateNames = append(stateNames, value.Name)
	}
	if err := requireTransportNames("player state methods", stateNames, playerStateTransportASTMethods); err != nil {
		return err
	}
	actionNames := make([]string, 0, len(spec.ActionMethods))
	for _, value := range spec.ActionMethods {
		actionNames = append(actionNames, value.Name)
	}
	if err := requireTransportNames("player action methods", actionNames, selectedPlayerActionMethods); err != nil {
		return err
	}
	controlNames := make([]string, 0, len(spec.Controls.Methods))
	for _, value := range spec.Controls.Methods {
		controlNames = append(controlNames, value.Name)
	}
	if err := requireTransportNames("player control methods", controlNames, selectedPlayerControlMethods); err != nil {
		return err
	}
	if err := requireTransportNames("player identity methods", spec.IdentityMethods, selectedPlayerIdentityMethods); err != nil {
		return err
	}
	presentationNames := make([]string, 0, len(spec.PresentationMethods))
	for _, value := range spec.PresentationMethods {
		presentationNames = append(presentationNames, value.Name)
	}
	if err := requireTransportNames("player presentation methods", presentationNames, selectedPlayerPresentationMethods); err != nil {
		return err
	}
	textNames := make([]string, 0, len(spec.TextMethods))
	for _, value := range spec.TextMethods {
		textNames = append(textNames, value.Name)
	}
	if err := requireTransportNames("player text methods", textNames, playerTextTransportASTMethods); err != nil {
		return err
	}
	gameModeNames := make([]string, 0, len(spec.GameModeMethods))
	for _, value := range spec.GameModeMethods {
		gameModeNames = append(gameModeNames, value.Name)
	}
	if err := requireTransportNames("player game mode methods", gameModeNames, []string{"SetGameMode", "GameMode"}); err != nil {
		return err
	}
	effects := make([]transportNameID, 0, len(spec.Effects.Types))
	for _, value := range spec.Effects.Types {
		effects = append(effects, transportNameID{Name: value.Name, ID: uint32(value.ID)})
	}
	if err := requireTransportIDs("effects", effects, effectTransportIDs); err != nil {
		return err
	}
	sounds := make([]transportNameID, 0, len(spec.Sounds))
	for index, value := range spec.Sounds {
		sounds = append(sounds, transportNameID{Name: value.Name, ID: uint32(index)})
	}
	if err := requireTransportIDs("sounds", sounds, soundTransportIDs); err != nil {
		return err
	}
	return nil
}

func requireTransportNames(subject string, got, want []string) error {
	got = append([]string(nil), got...)
	want = append([]string(nil), want...)
	sort.Strings(got)
	sort.Strings(want)
	if fmt.Sprint(got) != fmt.Sprint(want) {
		return fmt.Errorf("%s changed: got %v, want %v", subject, got, want)
	}
	return nil
}

func requireTransportIDs(subject string, got, want []transportNameID) error {
	if len(got) != len(want) {
		return fmt.Errorf("%s transport changed: got %d entries, want %d", subject, len(got), len(want))
	}
	byName := make(map[string]uint32, len(got))
	for _, value := range got {
		if _, duplicate := byName[value.Name]; duplicate {
			return fmt.Errorf("%s transport contains duplicate %s", subject, value.Name)
		}
		byName[value.Name] = value.ID
	}
	for _, expected := range want {
		if id, ok := byName[expected.Name]; !ok || id != expected.ID {
			return fmt.Errorf("%s transport ID for %s changed: got %d, %v, want %d", subject, expected.Name, id, ok, expected.ID)
		}
	}
	return nil
}
