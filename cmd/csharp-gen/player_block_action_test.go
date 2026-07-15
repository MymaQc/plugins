package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlayerBlockActionsMatchPinnedDragonfly(t *testing.T) {
	directory, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly").Output()
	if err != nil {
		t.Fatal(err)
	}
	methods, err := inspectPlayerBlockActionMethods(filepath.Join(string(bytes.TrimSpace(directory)), "server", "player", "player.go"))
	if err != nil {
		t.Fatal(err)
	}
	generated := string(generatePlayerBlockActions(methods))
	for _, expected := range []string{
		"public void BreakBlock(Cube.Pos pos)",
		"public void ContinueBreaking(Cube.Face face)",
		"public void PickBlock(Cube.Pos pos)",
		"public void Sleep(Cube.Pos pos)",
		"public void StartBreaking(Cube.Pos pos, Cube.Face face)",
		"public void UseItemOnBlock(Cube.Pos pos, Cube.Face face, Vector3 clickPos)",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated player block actions missing %q:\n%s", expected, generated)
		}
	}
	nativeGenerated := string(generateNativePlayerBlockActions(methods))
	csharpNative := string(generateCSharpPlayerBlockActions(methods))
	hostGenerated := string(generateHostPlayerBlockActions(methods))
	for index, name := range selectedPlayerBlockActionMethods {
		if !strings.Contains(nativeGenerated, "PlayerBlockAction"+name) ||
			!strings.Contains(csharpNative, "PlayerBlockAction"+name+" = ") ||
			!strings.Contains(hostGenerated, "connected."+name+"(") {
			t.Fatalf("generated transport missing method %d %s", index, name)
		}
	}
}
