package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlayerFinalDamageMatchesPinnedDragonfly(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	directory, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	spec, err := inspectPlayerFinalDamage(filepath.Join(
		string(bytes.TrimSpace(directory)), "server", "player", "player.go"))
	if err != nil {
		t.Fatal(err)
	}
	generated := string(generatePlayerFinalDamage(spec))
	if !strings.Contains(generated, "double FinalDamageFrom(double dmg, World.DamageSource src)") ||
		!strings.Contains(generated, "FinalPlayerDamage(_invocation, Id, dmg, src)") {
		t.Fatalf("generated final damage method:\n%s", generated)
	}
}
