package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPlayerVisibilityMatchesPinnedDragonfly(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	directory, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	spec, err := inspectPlayerVisibility(filepath.Join(
		string(bytes.TrimSpace(directory)), "server", "player", "player.go"))
	if err != nil {
		t.Fatal(err)
	}
	generated := generatePlayerVisibility(spec)
	for _, expected := range [][]byte{
		[]byte("public void HideEntity(World.Entity e)"),
		[]byte("public void ShowEntity(World.Entity e)"),
	} {
		if !bytes.Contains(generated, expected) {
			t.Fatalf("generated player visibility API missing %q", expected)
		}
	}
}
