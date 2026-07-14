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
	if len(methods) != 5 {
		t.Fatalf("generated %d world lifecycle methods, want 5", len(methods))
	}
	generated := string(generateWorldLifecycleMethods(methods))
	for _, expected := range []string{
		"string Name()",
		"Cube.Pos Spawn()",
		"void SetSpawn(Cube.Pos pos)",
		"void Save()",
		"void Close()",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated world lifecycle missing %q:\n%s", expected, generated)
		}
	}
}
