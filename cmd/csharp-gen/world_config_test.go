package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPinnedDragonflyWorldConfigUsesGoAST(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	module, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	root := string(bytes.TrimSpace(module))
	if err := inspectWorldConfig(
		filepath.Join(root, "server", "world", "conf.go"),
		filepath.Join(root, "server", "world", "world.go"),
		filepath.Join(root, "server", "world", "dimension.go"),
		filepath.Join(root, "server", "world", "provider.go"),
		filepath.Join(root, "server", "world", "mcdb", "conf.go"),
	); err != nil {
		t.Fatal(err)
	}
	generated := string(generateWorldConfig())
	for _, expected := range []string{
		"sealed record Config",
		"Dimension? Dim",
		"interface Dimension",
		"Cube.Range Range()",
		"bool WaterEvaporates()",
		"TimeSpan LavaSpreadDuration()",
		"Provider? Provider",
		"TimeSpan SaveInterval",
		"World New()",
		"static World New()",
		"static class MCDB",
		"DB Open(string dir)",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated world config missing %q:\n%s", expected, generated)
		}
	}
}
