package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlayerViewLayerMatchesPinnedDragonfly(t *testing.T) {
	directory, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly").Output()
	if err != nil {
		t.Fatal(err)
	}
	root := string(bytes.TrimSpace(directory))
	spec, err := inspectPlayerViewLayer(
		filepath.Join(root, "server", "player", "player.go"),
		filepath.Join(root, "server", "world", "visibility_level.go"),
	)
	if err != nil {
		t.Fatal(err)
	}
	generated := string(generatePlayerViewLayer(spec))
	for _, expected := range []string{
		"public static VisibilityLevel PublicVisibility() => new(0)",
		"public static VisibilityLevel EnforceInvisible() => new(1)",
		"public static VisibilityLevel EnforceVisible() => new(2)",
		"public void ViewNameTag(World.Entity entity, string nameTag)",
		"public void ViewPublicNameTag(World.Entity entity)",
		"public void ViewScoreTag(World.Entity entity, string scoreTag)",
		"public void ViewPublicScoreTag(World.Entity entity)",
		"public void ViewVisibility(World.Entity entity, World.VisibilityLevel level)",
		"public void RemoveViewLayer(World.Entity entity)",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated player view layer missing %q:\n%s", expected, generated)
		}
	}
	nativeGenerated := string(generateNativePlayerViewLayer(spec))
	csharpNative := string(generateCSharpPlayerViewLayer(spec))
	hostGenerated := string(generateHostPlayerViewLayer(spec))
	for _, name := range selectedPlayerViewLayerMethods {
		if !strings.Contains(nativeGenerated, "PlayerViewLayer"+name) ||
			!strings.Contains(csharpNative, "PlayerViewLayer"+name+" = ") ||
			!strings.Contains(hostGenerated, "viewer."+name+"(") {
			t.Fatalf("generated transport missing %s", name)
		}
	}
}
