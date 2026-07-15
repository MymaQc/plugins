// Code generated from Dragonfly Go AST and live registries by csharp-gen. DO NOT EDIT.

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
	case native.PlayerTextMessage:
		connected.Message(message)
	case native.PlayerTextTip:
		connected.SendTip(message)
	case native.PlayerTextPopup:
		connected.SendPopup(message)
	case native.PlayerTextJukeboxPopup:
		connected.SendJukeboxPopup(message)
	case native.PlayerTextNameTag:
		connected.SetNameTag(message)
	case native.PlayerTextDisconnect:
		connected.Disconnect(message)
	default:
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
	case 0:
		return hud.PaperDoll(), true
	case 1:
		return hud.Armour(), true
	case 2:
		return hud.ToolTips(), true
	case 3:
		return hud.TouchControls(), true
	case 4:
		return hud.Crosshair(), true
	case 5:
		return hud.HotBar(), true
	case 6:
		return hud.Health(), true
	case 7:
		return hud.ProgressBar(), true
	case 8:
		return hud.Hunger(), true
	case 9:
		return hud.AirBubbles(), true
	case 10:
		return hud.HorseHealth(), true
	case 11:
		return hud.StatusEffects(), true
	case 12:
		return hud.ItemText(), true
	default:
		return hud.Element{}, false
	}
}

func playerInputLock(value int64) (input.Lock, bool) {
	switch value {
	case 2:
		return input.Camera(), true
	case 4:
		return input.Movement(), true
	case 16:
		return input.LateralMovement(), true
	case 32:
		return input.Sneak(), true
	case 64:
		return input.Jump(), true
	case 128:
		return input.Mount(), true
	case 256:
		return input.Dismount(), true
	case 512:
		return input.MoveForward(), true
	case 1024:
		return input.MoveBackward(), true
	case 2048:
		return input.MoveLeft(), true
	case 4096:
		return input.MoveRight(), true
	default:
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
