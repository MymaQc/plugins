package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlayerStatusMatchesPinnedDragonfly(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	directory, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	spec, err := inspectPlayerStatus(filepath.Join(
		string(bytes.TrimSpace(directory)), "server", "player", "player.go"))
	if err != nil {
		t.Fatal(err)
	}
	generated := string(generatePlayerStatus(spec))
	for _, expected := range []string{
		"bool UsingItem()",
		"(Cube.Pos Position, bool Sleeping) Sleeping()",
		"(Vector3 Position, World.Dimension? Dimension, bool Found) DeathPosition()",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated player status missing %q:\n%s", expected, generated)
		}
	}
}
