package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPinnedDragonflyWorldLifecycleUsesGoAST(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	module, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	methods, err := inspectWorldLifecycleMethods(filepath.Join(
		string(bytes.TrimSpace(module)), "server", "world", "world.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(methods) != 21 {
		t.Fatalf("generated %d world lifecycle methods, want 21", len(methods))
	}
	generated := string(generateWorldLifecycleMethods(methods))
	for _, expected := range []string{
		"string Name()",
		"World.Dimension Dimension(this World world)",
		"Cube.Range Range()",
		"int HighestLightBlocker(int x, int z)",
		"int Time()",
		"void SetTime(int @new)",
		"void StopTime()",
		"void StartTime()",
		"bool TimeCycle()",
		"Cube.Pos Spawn()",
		"void SetSpawn(Cube.Pos pos)",
		"Cube.Pos PlayerSpawn(Guid id)",
		"void SetPlayerSpawn(Guid id, Cube.Pos pos)",
		"void SetRequiredSleepDuration(TimeSpan duration)",
		"GameMode DefaultGameMode()",
		"void SetTickRange(int v)",
		"void SetDefaultGameMode(GameMode mode)",
		"World.Difficulty Difficulty(this World world)",
		"void SetDifficulty(Difficulty d)",
		"void Save()",
		"void Close()",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated world lifecycle missing %q:\n%s", expected, generated)
		}
	}
	if strings.Contains(generated, "_invocation") {
		t.Fatalf("generated world lifecycle bypasses callback invocation:\n%s", generated)
	}
}
