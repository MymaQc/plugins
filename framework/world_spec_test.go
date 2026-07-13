package framework

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/world"
)

func validWorldSpec(path string) WorldSpec {
	return WorldSpec{
		ProviderPath:     path,
		Dimension:        WorldDimensionOverworld,
		OpenMode:         WorldOpenOrCreate,
		Save:             WorldSaveAutomatic,
		SaveInterval:     10 * time.Minute,
		RandomTicks:      WorldRandomTicksPerSubchunk,
		RandomTickRate:   3,
		Time:             WorldTimePreserve,
		Weather:          WorldWeatherPreserve,
		ChunkUnload:      WorldChunkUnloadAfter,
		ChunkUnloadAfter: 2 * time.Minute,
	}
}

func TestNormalizeWorldSpecUsesExplicitDragonflyDefaults(t *testing.T) {
	spec, err := normalizeWorldSpec(t.TempDir(), validWorldSpec("arenas/one"))
	if err != nil {
		t.Fatal(err)
	}
	config := spec.config(nil, nil, nil, world.EntityRegistry{})
	if config.SaveInterval != 10*time.Minute ||
		config.ChunkUnloadInterval != 2*time.Minute ||
		config.RandomTickSpeed != 3 || config.ReadOnly {
		t.Fatalf("config = %#v", config)
	}
}

func TestNormalizeWorldSpecRejectsInvalidPolicies(t *testing.T) {
	root := t.TempDir()
	tests := map[string]func(*WorldSpec){
		"unknown dimension":       func(spec *WorldSpec) { spec.Dimension = 99 },
		"unknown open mode":       func(spec *WorldSpec) { spec.OpenMode = 99 },
		"automatic zero interval": func(spec *WorldSpec) { spec.SaveInterval = 0 },
		"manual non-zero interval": func(spec *WorldSpec) {
			spec.Save = WorldSaveManual
		},
		"disabled ticks with rate": func(spec *WorldSpec) {
			spec.RandomTicks = WorldRandomTicksDisabled
		},
		"ticks with zero rate":      func(spec *WorldSpec) { spec.RandomTickRate = 0 },
		"ticks above max int32":     func(spec *WorldSpec) { spec.RandomTickRate = math.MaxInt32 + 1 },
		"preserved time with value": func(spec *WorldSpec) { spec.FixedTime = 1 },
		"unknown time":              func(spec *WorldSpec) { spec.Time = 99 },
		"unknown weather":           func(spec *WorldSpec) { spec.Weather = 99 },
		"unknown unload":            func(spec *WorldSpec) { spec.ChunkUnload = 99 },
		"zero unload interval":      func(spec *WorldSpec) { spec.ChunkUnloadAfter = 0 },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			spec := validWorldSpec("arenas/one")
			mutate(&spec)
			if _, err := normalizeWorldSpec(root, spec); err == nil {
				t.Fatal("invalid specification accepted")
			}
		})
	}
}

func TestNormalizeWorldSpecCanonicalizesReadOnlyToManual(t *testing.T) {
	spec := validWorldSpec("arenas/one")
	spec.ReadOnly = true
	normalized, err := normalizeWorldSpec(t.TempDir(), spec)
	if err != nil {
		t.Fatal(err)
	}
	if normalized.Save != WorldSaveManual || normalized.SaveInterval != 0 {
		t.Fatalf("save = %v, interval = %v", normalized.Save, normalized.SaveInterval)
	}
	config := normalized.config(nil, nil, nil, world.EntityRegistry{})
	if !config.ReadOnly || config.SaveInterval != -1 {
		t.Fatalf("config = %#v", config)
	}

	manual := validWorldSpec("arenas/one")
	manual.ReadOnly = true
	manual.Save = WorldSaveManual
	manual.SaveInterval = 0
	want, err := normalizeWorldSpec(t.TempDir(), manual)
	if err != nil {
		t.Fatal(err)
	}
	// Absolute roots differ, but canonical policy values must not.
	normalized.absoluteProviderPath, want.absoluteProviderPath = "", ""
	if normalized != want {
		t.Fatalf("read-only specs differ:\n got %#v\nwant %#v", normalized, want)
	}
}

func TestWorldSpecSettings(t *testing.T) {
	tests := []struct {
		name    string
		time    WorldTimePolicy
		fixed   int64
		weather WorldWeatherPolicy
		check   func(*testing.T, *world.Settings)
	}{
		{name: "preserve", time: WorldTimePreserve, weather: WorldWeatherPreserve, check: func(t *testing.T, settings *world.Settings) {
			if settings.Time != 42 || !settings.TimeCycle || !settings.Raining || !settings.Thundering || !settings.WeatherCycle {
				t.Fatalf("settings changed: %#v", settings)
			}
		}},
		{name: "cycle", time: WorldTimeCycle, weather: WorldWeatherCycle, check: func(t *testing.T, settings *world.Settings) {
			if !settings.TimeCycle || !settings.WeatherCycle {
				t.Fatalf("cycles disabled: %#v", settings)
			}
		}},
		{name: "fixed and clear", time: WorldTimeFixed, fixed: -6000, weather: WorldWeatherClear, check: func(t *testing.T, settings *world.Settings) {
			if settings.Time != -6000 || settings.TimeCycle || settings.Raining || settings.Thundering || settings.RainTime != 0 || settings.ThunderTime != 0 || settings.WeatherCycle {
				t.Fatalf("settings = %#v", settings)
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := validWorldSpec("arenas/one")
			spec.Time, spec.FixedTime, spec.Weather = test.time, test.fixed, test.weather
			normalized, err := normalizeWorldSpec(t.TempDir(), spec)
			if err != nil {
				t.Fatal(err)
			}
			settings := &world.Settings{Time: 42, TimeCycle: true, RainTime: 20, Raining: true, ThunderTime: 30, Thundering: true, WeatherCycle: true}
			normalized.applySettings(settings)
			test.check(t, settings)
		})
	}
}

func TestWorldSpecPathValidation(t *testing.T) {
	root := t.TempDir()
	invalid := []string{"", "/absolute", `back\\slash`, ".", "..", "a//b", "a/./b", "a/../b", "C:/world", "nul\x00world", strings.Repeat("x", 4097)}
	for _, path := range invalid {
		t.Run(strings.ReplaceAll(path, "/", "_"), func(t *testing.T) {
			spec := validWorldSpec(path)
			if _, err := normalizeWorldSpec(root, spec); err == nil {
				t.Fatalf("path %q accepted", path)
			}
		})
	}
}

func TestWorldSpecPathRejectsExistingSymlinkComponents(t *testing.T) {
	root := t.TempDir()
	target := t.TempDir()
	if err := os.Symlink(target, filepath.Join(root, "alias")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := normalizeWorldSpec(root, validWorldSpec("alias/world")); err == nil {
		t.Fatal("symlink component accepted")
	}
}
