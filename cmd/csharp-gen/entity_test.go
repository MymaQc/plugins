package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const entityHandleSource = `
func (e *EntityHandle) Entity(tx *Tx) (Entity, bool) { return nil, false }
func (e *EntityHandle) UUID() uuid.UUID { return uuid.Nil }
func (e *EntityHandle) Closed() bool { return false }
func (e *EntityHandle) Close() error { return nil }
`

func TestWorldEntityUsesGoAST(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "entity.go")
	source := `package world
import (
    "io"
    "github.com/go-gl/mathgl/mgl64"
    "github.com/df-mc/dragonfly/server/block/cube"
)
type Entity interface {
    io.Closer
    H() *EntityHandle
    Position() mgl64.Vec3
    Rotation() cube.Rotation
}` + entityHandleSource
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	methods, err := inspectWorldEntity(path)
	if err != nil {
		t.Fatal(err)
	}
	handleMethods, err := inspectWorldEntityHandle(path)
	if err != nil {
		t.Fatal(err)
	}
	generated := string(generateWorldEntity(methods, handleMethods))
	for _, expected := range []string{
		"void Close();",
		"EntityHandle H();",
		"Vector3 Position();",
		"Rotation Rotation();",
		"public (Entity? Entity, bool Ok) Entity(Tx tx)",
		"PluginBridge.Host.EntityHandleEntity(tx.Invocation, this)",
		"public Guid UUID()",
		"PluginBridge.Host.EntityHandleUuid(this)",
		"public bool Closed()",
		"PluginBridge.Host.EntityHandleClosed(this)",
		"public void Close()",
		"PluginBridge.Host.CloseEntityHandle(this)",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated entity interface missing %q:\n%s", expected, generated)
		}
	}
}

func TestPinnedDragonflyEntityHandleHasExactSurface(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	directory := string(bytes.TrimSpace(output))
	methods, err := inspectWorldEntityHandle(filepath.Join(directory, "server", "world", "entity.go"))
	if err != nil {
		t.Fatal(err)
	}
	want := []commandMethod{
		{Name: "Entity", Parameters: []parameter{{Name: "tx", Type: "Tx"}}, ReturnType: "(Entity? Entity, bool Ok)"},
		{Name: "UUID", ReturnType: "Guid"},
		{Name: "Closed", ReturnType: "bool"},
		{Name: "Close", ReturnType: "void"},
	}
	if len(methods) != len(want) {
		t.Fatalf("world.EntityHandle has %d methods, want %d", len(methods), len(want))
	}
	for index := range want {
		if methods[index].Name != want[index].Name || methods[index].ReturnType != want[index].ReturnType ||
			!equalParameters(methods[index].Parameters, want[index].Parameters) {
			t.Fatalf("world.EntityHandle method %d = %+v, want %+v", index, methods[index], want[index])
		}
	}
}

func equalParameters(left, right []parameter) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func TestPinnedDragonflyWorldEntityHasExactSurface(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	directory := string(bytes.TrimSpace(output))
	methods, err := inspectWorldEntity(filepath.Join(directory, "server", "world", "entity.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(methods) != 4 {
		t.Fatalf("world.Entity has %d methods, want 4", len(methods))
	}
	want := []entityMethod{
		{Name: "Close", ReturnType: "void"},
		{Name: "H", ReturnType: "EntityHandle"},
		{Name: "Position", ReturnType: "Vector3"},
		{Name: "Rotation", ReturnType: "Rotation"},
	}
	for index := range want {
		if methods[index] != want[index] {
			t.Fatalf("world.Entity method %d = %+v, want %+v", index, methods[index], want[index])
		}
	}
}
