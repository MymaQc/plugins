package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestPlayerHandlerMethodsUsesGoAST(t *testing.T) {
	path := filepath.Join(t.TempDir(), "handler.go")
	source := `package player
type Handler interface {
	HandleMove(ctx *Context, position int)
	HandleQuit(p *Player)
}`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	methods, err := playerHandlerMethods(path)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(methods, []string{"HandleQuit"}) {
		t.Fatalf("methods = %v", methods)
	}
	if output := string(generatePlayerHandler(methods)); !strings.Contains(output, "void HandleQuit(Player player);") {
		t.Fatalf("generated output missing HandleQuit: %s", output)
	}
}
