package native

import (
	"math"
	"testing"
	"time"
)

func TestPlayerEffectSnapshotViewValidation(t *testing.T) {
	valid := PlayerEffect{Type: EffectSpeed, Level: 2, Duration: time.Second, Potency: 1, Mode: PlayerEffectTimed}
	ambient := valid
	ambient.Mode, ambient.ParticlesHidden = PlayerEffectAmbient, true
	infinite := valid
	infinite.Mode, infinite.Duration = PlayerEffectInfinite, 0
	for name, value := range map[string]PlayerEffect{"timed": valid, "ambient": ambient, "infinite": infinite} {
		t.Run(name, func(t *testing.T) {
			view, ok := playerEffectSnapshotView(value)
			if !ok || int32(view.effect_type) != int32(value.Type) || int32(view.level) != value.Level || uint32(view.mode) != uint32(value.Mode) || (view.particles_hidden != 0) != value.ParticlesHidden {
				t.Fatalf("view = %#v ok=%v", view, ok)
			}
		})
	}

	invalid := map[string]PlayerEffect{
		"zero level":        func() PlayerEffect { value := valid; value.Level = 0; return value }(),
		"negative duration": func() PlayerEffect { value := valid; value.Duration = -1; return value }(),
		"zero potency":      func() PlayerEffect { value := valid; value.Potency = 0; return value }(),
		"nan potency":       func() PlayerEffect { value := valid; value.Potency = math.NaN(); return value }(),
		"instant":           func() PlayerEffect { value := valid; value.Mode = PlayerEffectInstant; return value }(),
		"unknown mode":      func() PlayerEffect { value := valid; value.Mode = 99; return value }(),
		"infinite duration": func() PlayerEffect { value := infinite; value.Duration = time.Second; return value }(),
	}
	for name, value := range invalid {
		t.Run(name, func(t *testing.T) {
			if _, ok := playerEffectSnapshotView(value); ok {
				t.Fatalf("invalid effect accepted: %#v", value)
			}
		})
	}
}
