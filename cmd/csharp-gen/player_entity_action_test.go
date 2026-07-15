package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlayerEntityActionsMatchPinnedDragonfly(t *testing.T) {
	directory, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly").Output()
	if err != nil {
		t.Fatal(err)
	}
	methods, err := inspectPlayerEntityActionMethods(filepath.Join(
		string(bytes.TrimSpace(directory)), "server", "player", "player.go"))
	if err != nil {
		t.Fatal(err)
	}
	generated := string(generatePlayerEntityActions(methods))
	for _, expected := range []string{
		"public bool UseItemOnEntity(World.Entity e)",
		"public bool AttackEntity(World.Entity e)",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated player entity actions missing %q:\n%s", expected, generated)
		}
	}
	nativeGenerated := string(generateNativePlayerEntityActions(methods))
	csharpNative := string(generateCSharpPlayerEntityActions(methods))
	hostGenerated := string(generateHostPlayerEntityActions(methods))
	for _, name := range selectedPlayerEntityActionMethods {
		if !strings.Contains(nativeGenerated, "PlayerEntityAction"+name) ||
			!strings.Contains(csharpNative, "PlayerEntityAction"+name+" = ") ||
			!strings.Contains(hostGenerated, "connected."+name+"(entity)") {
			t.Fatalf("generated transport missing %s", name)
		}
	}
}
