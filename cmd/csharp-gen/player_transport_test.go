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
		{name: "csharp", path: "csharp/Dragonfly.Native/Generated/Player.State.g.cs", generate: generateCSharpPlayerStateTransport},
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
		"PlayerStateBreathing           PlayerStateKind = 20",
		"PlayerStateFlying              PlayerStateKind = 26",
		"PlayerStateOnFireDuration      PlayerStateKind = 27",
		"PlayerStateMaxAirSupply        PlayerStateKind = 30",
		"PlayerStateCanCollectExperience PlayerStateKind = 33",
		"PlayerActionAddFood                   PlayerActionKind = 0",
		"PlayerActionCollectExperience         PlayerActionKind = 6",
		"PlayerActionRemoveBossBar             PlayerActionKind = 13",
		"PlayerActionWake                      PlayerActionKind = 24",
		"PlayerActionInputLocked               PlayerActionKind = 30",
		"PlayerStringNameTag         PlayerStringKind = 0",
		"PlayerStringScoreTag        PlayerStringKind = 1",
		"PlayerStringAddrString      PlayerStringKind = 7",
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

func TestPlayerPresentationTransportCallsExactDragonflyMethods(t *testing.T) {
	generated, err := generateHostPlayerTransport(inspectPinnedPlayerTransport(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"connected.EnableInstantRespawn()",
		"connected.DisableInstantRespawn()",
		"connected.ShowCoordinates()",
		"connected.HideCoordinates()",
		"connected.SendSleepingIndicator(int(value.Integer), int(value.Number))",
		"connected.CloseDialogue()",
		"connected.RemoveBossBar()",
		"return connected.NameTag(), true",
		"return connected.ScoreTag(), true",
		"return connected.DeviceID(), true",
		"return connected.DeviceModel(), true",
		"return connected.SelfSignedID(), true",
		"return connected.Locale().String(), true",
		"return address.Network(), true",
		"return address.String(), true",
	} {
		if !strings.Contains(string(generated), expected) {
			t.Fatalf("generated host transport missing %q", expected)
		}
	}
}

func TestPlayerActionTransportCallsExactDragonflyMethods(t *testing.T) {
	generated, err := generateHostPlayerTransport(inspectPinnedPlayerTransport(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, method := range selectedPlayerActionMethods {
		expected := "connected." + method + "()"
		if !strings.Contains(string(generated), expected) {
			t.Fatalf("generated host transport missing %q", expected)
		}
	}
}

func TestPlayerControlTransportCallsExactDragonflyMethods(t *testing.T) {
	generated, err := generateHostPlayerTransport(inspectPinnedPlayerTransport(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"connected.ShowHudElement(element)",
		"connected.HideHudElement(element)",
		"connected.HudElementHidden(element)",
		"connected.LockInput(lock)",
		"connected.UnlockInput(lock)",
		"connected.InputLocked(lock)",
		"return hud.PaperDoll(), true",
		"return input.MoveRight(), true",
	} {
		if !strings.Contains(string(generated), expected) {
			t.Fatalf("generated host transport missing %q", expected)
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
		"action name": {
			mutate: func(spec *playerTransportSpec) { spec.ActionMethods[0].Name = "Changed" },
			want:   "player action methods changed",
		},
		"control name": {
			mutate: func(spec *playerTransportSpec) { spec.Controls.Methods[0].Name = "Changed" },
			want:   "player control methods changed",
		},
		"identity name": {
			mutate: func(spec *playerTransportSpec) { spec.IdentityMethods[0] = "Changed" },
			want:   "player identity methods changed",
		},
		"text name": {
			mutate: func(spec *playerTransportSpec) { spec.TextMethods[0].Name = "Changed" },
			want:   "player text methods changed",
		},
		"presentation name": {
			mutate: func(spec *playerTransportSpec) { spec.PresentationMethods[0].Name = "Changed" },
			want:   "player presentation methods changed",
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
	actions, err := inspectPlayerActionMethods(playerPath)
	if err != nil {
		t.Fatal(err)
	}
	gophertunnel := moduleDirectoryForTest(t, "github.com/sandertv/gophertunnel")
	controls, err := inspectPlayerControls(
		playerPath,
		filepath.Join(directory, "server", "player", "hud", "element.go"),
		filepath.Join(directory, "server", "player", "input", "lock.go"),
		filepath.Join(gophertunnel, "minecraft", "protocol", "packet", "update_client_input_locks.go"),
	)
	if err != nil {
		t.Fatal(err)
	}
	identity, err := inspectPlayerIdentityMethods(playerPath)
	if err != nil {
		t.Fatal(err)
	}
	presentation, err := inspectPlayerPresentationMethods(playerPath)
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
		StateMethods: state, ActionMethods: actions, Controls: controls, IdentityMethods: identity, PresentationMethods: presentation, TextMethods: text,
		Effects: effects, Sounds: sounds, GameModeMethods: gameModes,
	}
}

func clonePlayerTransportSpec(spec playerTransportSpec) playerTransportSpec {
	spec.StateMethods = append([]playerStateMethod(nil), spec.StateMethods...)
	spec.ActionMethods = append([]playerActionMethod(nil), spec.ActionMethods...)
	spec.Controls.Methods = append([]playerControlMethod(nil), spec.Controls.Methods...)
	spec.Controls.HudElements = append([]playerControlValue(nil), spec.Controls.HudElements...)
	spec.Controls.InputLocks = append([]playerControlValue(nil), spec.Controls.InputLocks...)
	spec.IdentityMethods = append([]string(nil), spec.IdentityMethods...)
	spec.PresentationMethods = append([]playerPresentationMethod(nil), spec.PresentationMethods...)
	spec.TextMethods = append([]method(nil), spec.TextMethods...)
	spec.Effects.Types = append([]effectTypeSpec(nil), spec.Effects.Types...)
	spec.Sounds = append([]soundTypeSpec(nil), spec.Sounds...)
	spec.GameModeMethods = append([]commandMethod(nil), spec.GameModeMethods...)
	return spec
}
