// Code generated from Dragonfly Go AST and live registries by csharp-gen. DO NOT EDIT.

package host

import (
	"math"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/player"
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
	default:
		return false
	}
	return true
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
