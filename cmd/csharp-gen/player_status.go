package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type playerStatusSpec struct {
	UsingItem     bool
	Sleeping      bool
	DeathPosition bool
}

func inspectPlayerStatus(path string) (playerStatusSpec, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return playerStatusSpec{}, err
	}
	want := map[string]goSignature{
		"UsingItem":     {Results: "bool"},
		"Sleeping":      {Results: "cube.Pos, bool"},
		"DeathPosition": {Results: "mgl64.Vec3, world.Dimension, bool"},
	}
	found := map[string]bool{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !playerMethod(function) {
			continue
		}
		expected, selected := want[function.Name.Name]
		if !selected {
			continue
		}
		if got := goFunctionSignature(function); got != expected {
			return playerStatusSpec{}, fmt.Errorf("Dragonfly player.Player.%s signature changed: %+v", function.Name.Name, got)
		}
		found[function.Name.Name] = true
	}
	for name := range want {
		if !found[name] {
			return playerStatusSpec{}, fmt.Errorf("Dragonfly player.Player has no %s method", name)
		}
	}
	return playerStatusSpec{UsingItem: true, Sleeping: true, DeathPosition: true}, nil
}

func generatePlayerStatus(spec playerStatusSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\n")
	output.WriteString("namespace Dragonfly;\n\npublic sealed partial class Player\n{\n")
	if spec.UsingItem {
		output.WriteString("    public bool UsingItem() => PluginBridge.Host.PlayerUsingItem(_invocation, Id);\n")
	}
	if spec.Sleeping {
		output.WriteString("    public (Cube.Pos Position, bool Sleeping) Sleeping() => PluginBridge.Host.PlayerSleeping(_invocation, Id);\n")
	}
	if spec.DeathPosition {
		output.WriteString("    public (Vector3 Position, World.Dimension? Dimension, bool Found) DeathPosition() => PluginBridge.Host.PlayerDeathPosition(_invocation, Id);\n")
	}
	output.WriteString("}\n")
	return output.Bytes()
}
