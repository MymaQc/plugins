package native

import (
	"sync"
	"sync/atomic"
	"time"
)

type PlayerTextKind uint32
type PlayerTransformKind uint32

const (
	PlayerTextMessage PlayerTextKind = iota
	PlayerTextTip
	PlayerTextPopup
	PlayerTextJukeboxPopup
)

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
}

type noopHost struct{}

func (noopHost) SendPlayerText(PlayerID, PlayerTextKind, string) bool { return false }
func (noopHost) SendPlayerTitle(PlayerID, PlayerTitle) bool           { return false }
func (noopHost) TransformPlayer(PlayerID, PlayerTransformKind, Vec3, float64, float64) bool {
	return false
}
func (noopHost) PlayerRotation(PlayerID) (Rotation, bool) { return Rotation{}, false }

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
