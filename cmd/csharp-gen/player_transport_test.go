package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratePlayerTransportMatchesCurrentFiles(t *testing.T) {
	spec := inspectPinnedPlayerTransport(t)
	tests := []struct {
		name     string
		path     string
		generate func(playerTransportSpec) ([]byte, error)
	}{
		{name: "native", path: "internal/native/player_state_generated.go", generate: generateNativePlayerTransport},
		{name: "host", path: "internal/host/player_state_generated.go", generate: generateHostPlayerTransport},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			generated, err := test.generate(spec)
			if err != nil {
				t.Fatal(err)
			}
			current, err := os.ReadFile(filepath.Join("..", "..", test.path))
			if err != nil {
				t.Fatal(err)
			}
			// PlayerTextKick was stale private transport with no C# API. New
			// generation deliberately removes it while preserving every live ID.
			current = removeLegacyPlayerTextKick(current)
			if !bytes.Equal(generated, current) {
				generatedLines, currentLines := strings.Split(string(generated), "\n"), strings.Split(string(current), "\n")
				for index := 0; index < len(generatedLines) && index < len(currentLines); index++ {
					if generatedLines[index] != currentLines[index] {
						t.Fatalf("generated %s differs at line %d:\ngot  %q\nwant %q", test.path, index+1, generatedLines[index], currentLines[index])
					}
				}
				t.Fatalf("generated %s line count differs: got %d, want %d", test.path, len(generatedLines), len(currentLines))
			}
		})
	}
}

func TestGeneratorDoesNotDependOnGeneratedHostPackages(t *testing.T) {
	command := exec.Command("go", "list", "-deps", "./cmd/csharp-gen")
	command.Dir = filepath.Join("..", "..")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		"github.com/bedrock-gophers/plugins/internal/host",
		"github.com/bedrock-gophers/plugins/internal/native",
	} {
		for _, dependency := range strings.Fields(string(output)) {
			if dependency == forbidden {
				t.Fatalf("csharp-gen depends on %s, so generated host files cannot bootstrap", forbidden)
			}
		}
	}
}

func removeLegacyPlayerTextKick(source []byte) []byte {
	lines := strings.SplitAfter(string(source), "\n")
	filtered := make([]string, 0, len(lines))
	skipDisconnect := false
	for _, line := range lines {
		if strings.Contains(line, "PlayerTextKick") {
			skipDisconnect = strings.Contains(line, "case native.")
			continue
		}
		if skipDisconnect && strings.TrimSpace(line) == "connected.Disconnect(message)" {
			skipDisconnect = false
			continue
		}
		skipDisconnect = false
		filtered = append(filtered, line)
	}
	return []byte(strings.Join(filtered, ""))
}

func TestPlayerTransportPreservesExplicitIDs(t *testing.T) {
	spec := inspectPinnedPlayerTransport(t)
	generated, err := generateNativePlayerTransport(spec)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(generated), "abi-gen") {
		t.Fatal("generated player transport still names removed abi-gen")
	}
	for _, expected := range []string{
		"PlayerStateGameMode            PlayerStateKind = 0",
		"PlayerStateFood                PlayerStateKind = 3",
		"PlayerStateSpeed               PlayerStateKind = 11",
		"PlayerStateVerticalFlightSpeed PlayerStateKind = 13",
		"EffectSlowFalling    EffectType = 27",
		"EffectDarkness       EffectType = 30",
		"PlayerTextTip          PlayerTextKind = 1",
		"PlayerTextPopup        PlayerTextKind = 2",
		"SoundTnt                        SoundKind = 57",
		"SoundGoatHorn                   SoundKind = 86",
	} {
		if !strings.Contains(string(generated), expected) {
			t.Fatalf("generated native transport missing %q", expected)
		}
	}
}

func TestPlayerTransportRejectsSpecDrift(t *testing.T) {
	base := inspectPinnedPlayerTransport(t)
	tests := map[string]struct {
		mutate func(*playerTransportSpec)
		want   string
	}{
		"state name": {
			mutate: func(spec *playerTransportSpec) { spec.StateMethods[0].Name = "Changed" },
			want:   "player state methods changed",
		},
		"text name": {
			mutate: func(spec *playerTransportSpec) { spec.TextMethods[0].Name = "Changed" },
			want:   "player text methods changed",
		},
		"game mode name": {
			mutate: func(spec *playerTransportSpec) { spec.GameModeMethods[0].Name = "Changed" },
			want:   "player game mode methods changed",
		},
		"effect id": {
			mutate: func(spec *playerTransportSpec) { spec.Effects.Types[0].ID = 99 },
			want:   "effects transport ID",
		},
		"sound order": {
			mutate: func(spec *playerTransportSpec) {
				spec.Sounds[0], spec.Sounds[1] = spec.Sounds[1], spec.Sounds[0]
			},
			want: "sounds transport ID",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			spec := clonePlayerTransportSpec(base)
			test.mutate(&spec)
			_, err := generateNativePlayerTransport(spec)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func inspectPinnedPlayerTransport(t *testing.T) playerTransportSpec {
	t.Helper()
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	module, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	directory := string(bytes.TrimSpace(module))
	playerPath := filepath.Join(directory, "server", "player", "player.go")
	state, err := inspectPlayerStateMethods(playerPath)
	if err != nil {
		t.Fatal(err)
	}
	text, err := playerTextMethods(playerPath)
	if err != nil {
		t.Fatal(err)
	}
	effects, err := inspectEffects(filepath.Join(directory, "server", "entity", "effect"))
	if err != nil {
		t.Fatal(err)
	}
	sounds, err := inspectSounds(filepath.Join(directory, "server", "world", "sound"))
	if err != nil {
		t.Fatal(err)
	}
	gameModes, err := inspectPlayerGameModeMethods(playerPath)
	if err != nil {
		t.Fatal(err)
	}
	return playerTransportSpec{
		StateMethods: state, TextMethods: text, Effects: effects, Sounds: sounds, GameModeMethods: gameModes,
	}
}

func clonePlayerTransportSpec(spec playerTransportSpec) playerTransportSpec {
	spec.StateMethods = append([]playerStateMethod(nil), spec.StateMethods...)
	spec.TextMethods = append([]method(nil), spec.TextMethods...)
	spec.Effects.Types = append([]effectTypeSpec(nil), spec.Effects.Types...)
	spec.Sounds = append([]soundTypeSpec(nil), spec.Sounds...)
	spec.GameModeMethods = append([]commandMethod(nil), spec.GameModeMethods...)
	return spec
}
