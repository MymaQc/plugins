package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlayerHandlerMethodsUsesGoAST(t *testing.T) {
	path := filepath.Join(t.TempDir(), "handler.go")
	source := `package player
type Handler interface {
	HandleMove(ctx *Context, newPos mgl64.Vec3, newRot cube.Rotation)
	HandleJump(p *Player)
	HandleQuit(p *Player)
}`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	methods, err := playerHandlerMethods(path)
	if err != nil {
		t.Fatal(err)
	}
	output := string(generatePlayerHandler(methods))
	for _, expected := range []string{
		"void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot);",
		"void HandleQuit(Player p);",
		"public virtual void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot) { }",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("generated output missing %q:\n%s", expected, output)
		}
	}
	if strings.Contains(output, "HandleJump") {
		t.Fatalf("generated unsupported method:\n%s", output)
	}
}
