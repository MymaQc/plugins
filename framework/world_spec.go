package framework

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/df-mc/dragonfly/server/world"
)

const maxWorldProviderPathBytes = 4096

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

// WorldSpec describes immutable creation policies for a persistent world.
type WorldSpec struct {
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

type normalizedWorldSpec struct {
	providerPath         string
	absoluteProviderPath string
	Dimension            WorldDimension
	OpenMode             WorldOpenMode
	ReadOnly             bool
	Save                 WorldSavePolicy
	SaveInterval         time.Duration
	RandomTicks          WorldRandomTickPolicy
	RandomTickRate       uint32
	Time                 WorldTimePolicy
	FixedTime            int64
	Weather              WorldWeatherPolicy
	ChunkUnload          WorldChunkUnloadPolicy
	ChunkUnloadAfter     time.Duration
}

func normalizeWorldSpec(root string, spec WorldSpec) (normalizedWorldSpec, error) {
	providerPath, absoluteProviderPath, err := normalizeProviderPath(root, spec.ProviderPath)
	if err != nil {
		return normalizedWorldSpec{}, err
	}
	if _, ok := worldSpecDimension(spec.Dimension); !ok {
		return normalizedWorldSpec{}, fmt.Errorf("invalid world dimension %d", spec.Dimension)
	}
	if spec.OpenMode > WorldCreateNew {
		return normalizedWorldSpec{}, fmt.Errorf("invalid world open mode %d", spec.OpenMode)
	}
	if spec.Save > WorldSaveManual {
		return normalizedWorldSpec{}, fmt.Errorf("invalid world save policy %d", spec.Save)
	}
	if spec.ReadOnly {
		spec.Save = WorldSaveManual
		spec.SaveInterval = 0
	} else {
		switch spec.Save {
		case WorldSaveAutomatic:
			if spec.SaveInterval <= 0 {
				return normalizedWorldSpec{}, errors.New("automatic save interval must be positive")
			}
		case WorldSaveManual:
			if spec.SaveInterval != 0 {
				return normalizedWorldSpec{}, errors.New("manual save interval must be zero")
			}
		}
	}
	if spec.RandomTicks > WorldRandomTicksPerSubchunk {
		return normalizedWorldSpec{}, fmt.Errorf("invalid random tick policy %d", spec.RandomTicks)
	}
	switch spec.RandomTicks {
	case WorldRandomTicksDisabled:
		if spec.RandomTickRate != 0 {
			return normalizedWorldSpec{}, errors.New("disabled random ticks require a zero rate")
		}
	case WorldRandomTicksPerSubchunk:
		if spec.RandomTickRate == 0 || spec.RandomTickRate > math.MaxInt32 {
			return normalizedWorldSpec{}, fmt.Errorf("random tick rate must be between 1 and %d", math.MaxInt32)
		}
	}
	if spec.Time > WorldTimeFixed {
		return normalizedWorldSpec{}, fmt.Errorf("invalid time policy %d", spec.Time)
	}
	if spec.Time != WorldTimeFixed && spec.FixedTime != 0 {
		return normalizedWorldSpec{}, errors.New("non-fixed time policy requires a zero fixed time")
	}
	if spec.Weather > WorldWeatherClear {
		return normalizedWorldSpec{}, fmt.Errorf("invalid weather policy %d", spec.Weather)
	}
	if spec.ChunkUnload != WorldChunkUnloadAfter {
		return normalizedWorldSpec{}, fmt.Errorf("invalid chunk unload policy %d", spec.ChunkUnload)
	}
	if spec.ChunkUnloadAfter <= 0 {
		return normalizedWorldSpec{}, errors.New("chunk unload interval must be positive")
	}
	return normalizedWorldSpec{
		providerPath: providerPath, absoluteProviderPath: absoluteProviderPath,
		Dimension: spec.Dimension, OpenMode: spec.OpenMode, ReadOnly: spec.ReadOnly,
		Save: spec.Save, SaveInterval: spec.SaveInterval,
		RandomTicks: spec.RandomTicks, RandomTickRate: spec.RandomTickRate,
		Time: spec.Time, FixedTime: spec.FixedTime, Weather: spec.Weather,
		ChunkUnload: spec.ChunkUnload, ChunkUnloadAfter: spec.ChunkUnloadAfter,
	}, nil
}

func normalizeProviderPath(root, providerPath string) (string, string, error) {
	if root == "" {
		return "", "", errors.New("persistent world root is not configured")
	}
	if providerPath == "" {
		return "", "", errors.New("world provider path is required")
	}
	if len(providerPath) > maxWorldProviderPathBytes {
		return "", "", fmt.Errorf("world provider path exceeds %d bytes", maxWorldProviderPathBytes)
	}
	if strings.IndexByte(providerPath, 0) >= 0 || strings.Contains(providerPath, `\`) {
		return "", "", errors.New("world provider path contains an invalid character")
	}
	if filepath.IsAbs(providerPath) || filepath.VolumeName(providerPath) != "" || windowsVolumePath(providerPath) {
		return "", "", errors.New("world provider path must be relative")
	}
	components := strings.Split(providerPath, "/")
	for _, component := range components {
		if component == "" || component == "." || component == ".." {
			return "", "", errors.New("world provider path contains an invalid component")
		}
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", "", fmt.Errorf("resolve world root: %w", err)
	}
	absolutePath := filepath.Join(append([]string{absoluteRoot}, components...)...)
	relative, err := filepath.Rel(absoluteRoot, absolutePath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", "", errors.New("world provider path escapes configured root")
	}
	current := absoluteRoot
	for _, component := range components {
		current = filepath.Join(current, component)
		info, statErr := os.Lstat(current)
		if errors.Is(statErr, os.ErrNotExist) {
			break
		}
		if statErr != nil {
			return "", "", fmt.Errorf("inspect world provider path: %w", statErr)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", "", fmt.Errorf("world provider path component %q is a symlink", component)
		}
	}
	return strings.Join(components, "/"), absolutePath, nil
}

func windowsVolumePath(path string) bool {
	return len(path) >= 2 && ((path[0] >= 'a' && path[0] <= 'z') || (path[0] >= 'A' && path[0] <= 'Z')) && path[1] == ':'
}

func (spec normalizedWorldSpec) config(log *slog.Logger, provider world.Provider, blocks world.BlockRegistry, entities world.EntityRegistry) world.Config {
	dimension, _ := worldSpecDimension(spec.Dimension)
	saveInterval := spec.SaveInterval
	if spec.Save == WorldSaveManual {
		saveInterval = -1
	}
	randomTickSpeed := -1
	if spec.RandomTicks == WorldRandomTicksPerSubchunk {
		randomTickSpeed = int(spec.RandomTickRate)
	}
	return world.Config{
		Log: log, Dim: dimension, Provider: provider, ReadOnly: spec.ReadOnly,
		SaveInterval: saveInterval, ChunkUnloadInterval: spec.ChunkUnloadAfter,
		RandomTickSpeed: randomTickSpeed, Blocks: blocks, Entities: entities,
	}
}

func (spec normalizedWorldSpec) applySettings(settings *world.Settings) {
	settings.Lock()
	defer settings.Unlock()
	switch spec.Time {
	case WorldTimeCycle:
		settings.TimeCycle = true
	case WorldTimeFixed:
		settings.Time = spec.FixedTime
		settings.TimeCycle = false
	}
	switch spec.Weather {
	case WorldWeatherCycle:
		settings.WeatherCycle = true
	case WorldWeatherClear:
		settings.RainTime = 0
		settings.Raining = false
		settings.ThunderTime = 0
		settings.Thundering = false
		settings.WeatherCycle = false
	}
}

func worldSpecDimension(value WorldDimension) (world.Dimension, bool) {
	switch value {
	case WorldDimensionOverworld:
		return world.Overworld, true
	case WorldDimensionNether:
		return world.Nether, true
	case WorldDimensionEnd:
		return world.End, true
	default:
		return nil, false
	}
}
