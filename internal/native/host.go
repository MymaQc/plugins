package native

import (
	"sync"
	"sync/atomic"
	"time"
)

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

// Host executes synchronous actions requested by native plugins.
type Host interface {
	SendPlayerText(PlayerID, PlayerTextKind, string) bool
	SendPlayerTitle(PlayerID, PlayerTitle) bool
	TransformPlayer(PlayerID, PlayerTransformKind, Vec3, float64, float64) bool
	PlayerRotation(PlayerID) (Rotation, bool)
	SetPlayerState(PlayerID, PlayerStateKind, PlayerStateValue) bool
	PlayerState(PlayerID, PlayerStateKind) (PlayerStateValue, bool)
	ChangePlayerEffect(PlayerID, PlayerEffectOperation, PlayerEffect) bool
	SetPlayerEntityVisible(PlayerID, EntityID, bool) bool
	PlayerSkin(PlayerID) (PlayerSkin, bool)
	SetPlayerSkin(PlayerID, PlayerSkin) bool
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
func (noopHost) ChangePlayerEffect(PlayerID, PlayerEffectOperation, PlayerEffect) bool { return false }
func (noopHost) SetPlayerEntityVisible(PlayerID, EntityID, bool) bool                  { return false }
func (noopHost) PlayerSkin(PlayerID) (PlayerSkin, bool)                                { return PlayerSkin{}, false }
func (noopHost) SetPlayerSkin(PlayerID, PlayerSkin) bool                               { return false }

var (
	hostSequence         atomic.Uint64
	hosts                sync.Map
	skinSnapshotSequence atomic.Uint64
	skinSnapshotMu       sync.Mutex
	skinSnapshots        = map[uint64]skinSnapshot{}
	skinSnapshotCounts   = map[uint64]int{}
)

const maxSkinSnapshotsPerHost = 32

type skinSnapshot struct {
	host uint64
	skin PlayerSkin
}

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
		skinSnapshotMu.Lock()
		for snapshotID, snapshot := range skinSnapshots {
			if snapshot.host == id {
				delete(skinSnapshots, snapshotID)
			}
		}
		delete(skinSnapshotCounts, id)
		skinSnapshotMu.Unlock()
	}
}

func resolveHost(id uint64) (Host, bool) {
	host, ok := hosts.Load(id)
	if !ok {
		return nil, false
	}
	return host.(Host), true
}

func registerSkinSnapshot(host uint64, skin PlayerSkin) (uint64, bool) {
	skinSnapshotMu.Lock()
	defer skinSnapshotMu.Unlock()
	if skinSnapshotCounts[host] >= maxSkinSnapshotsPerHost {
		return 0, false
	}
	id := skinSnapshotSequence.Add(1)
	skinSnapshots[id] = skinSnapshot{host: host, skin: clonePlayerSkin(skin)}
	skinSnapshotCounts[host]++
	return id, true
}

func resolveSkinSnapshot(host, id uint64) (PlayerSkin, bool) {
	skinSnapshotMu.Lock()
	defer skinSnapshotMu.Unlock()
	value, ok := skinSnapshots[id]
	if !ok || value.host != host {
		return PlayerSkin{}, false
	}
	return value.skin, true
}

func unregisterSkinSnapshot(host, id uint64) {
	skinSnapshotMu.Lock()
	defer skinSnapshotMu.Unlock()
	value, ok := skinSnapshots[id]
	if ok && value.host == host {
		delete(skinSnapshots, id)
		skinSnapshotCounts[host]--
		if skinSnapshotCounts[host] == 0 {
			delete(skinSnapshotCounts, host)
		}
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
