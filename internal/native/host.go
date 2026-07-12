package native

import (
	"sync"
	"sync/atomic"
	"time"
)

type PlayerTextKind uint32
type PlayerTransformKind uint32
type PlayerStateKind uint32

const (
	PlayerTextMessage PlayerTextKind = iota
	PlayerTextTip
	PlayerTextPopup
	PlayerTextJukeboxPopup
)

const (
	PlayerStateGameMode PlayerStateKind = iota
	PlayerStateHeal
	PlayerStateHurt
	PlayerStateFood
	PlayerStateMaxHealth
	PlayerStateHealth
)

type PlayerStateValue struct {
	Number  float64
	Integer int64
}

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

// Host executes synchronous actions requested by native plugins.
type Host interface {
	SendPlayerText(PlayerID, PlayerTextKind, string) bool
	SendPlayerTitle(PlayerID, PlayerTitle) bool
	TransformPlayer(PlayerID, PlayerTransformKind, Vec3, float64, float64) bool
	PlayerRotation(PlayerID) (Rotation, bool)
	SetPlayerState(PlayerID, PlayerStateKind, PlayerStateValue) bool
	PlayerState(PlayerID, PlayerStateKind) (PlayerStateValue, bool)
}

type noopHost struct{}

func (noopHost) SendPlayerText(PlayerID, PlayerTextKind, string) bool { return false }
func (noopHost) SendPlayerTitle(PlayerID, PlayerTitle) bool           { return false }
func (noopHost) TransformPlayer(PlayerID, PlayerTransformKind, Vec3, float64, float64) bool {
	return false
}
func (noopHost) PlayerRotation(PlayerID) (Rotation, bool)                        { return Rotation{}, false }
func (noopHost) SetPlayerState(PlayerID, PlayerStateKind, PlayerStateValue) bool { return false }
func (noopHost) PlayerState(PlayerID, PlayerStateKind) (PlayerStateValue, bool) {
	return PlayerStateValue{}, false
}

var (
	hostSequence atomic.Uint64
	hosts        sync.Map
)

func registerHost(host Host) uint64 {
	if host == nil {
		host = noopHost{}
	}
	id := hostSequence.Add(1)
	hosts.Store(id, host)
	return id
}

func unregisterHost(id uint64) {
	if id != 0 {
		hosts.Delete(id)
	}
}

func resolveHost(id uint64) (Host, bool) {
	host, ok := hosts.Load(id)
	if !ok {
		return nil, false
	}
	return host.(Host), true
}
