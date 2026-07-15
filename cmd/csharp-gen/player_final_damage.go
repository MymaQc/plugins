package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type playerFinalDamageSpec struct {
	Damage string
	Source string
}

func inspectPlayerFinalDamage(path string) (playerFinalDamageSpec, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return playerFinalDamageSpec{}, err
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !playerMethod(function) || function.Name.Name != "FinalDamageFrom" {
			continue
		}
		want := goSignature{Parameters: "float64, world.DamageSource", Results: "float64"}
		if got := goFunctionSignature(function); got != want {
			return playerFinalDamageSpec{}, fmt.Errorf("Dragonfly player.Player.FinalDamageFrom signature changed: %+v", got)
		}
		if function.Type.Params == nil || len(function.Type.Params.List) != 2 ||
			len(function.Type.Params.List[0].Names) != 1 || len(function.Type.Params.List[1].Names) != 1 {
			return playerFinalDamageSpec{}, fmt.Errorf("Dragonfly player.Player.FinalDamageFrom parameter names changed")
		}
		return playerFinalDamageSpec{
			Damage: function.Type.Params.List[0].Names[0].Name,
			Source: function.Type.Params.List[1].Names[0].Name,
		}, nil
	}
	return playerFinalDamageSpec{}, fmt.Errorf("Dragonfly player.Player has no FinalDamageFrom method")
}

func generatePlayerFinalDamage(spec playerFinalDamageSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\n")
	output.WriteString("namespace Dragonfly;\n\npublic sealed partial class Player\n{\n")
	fmt.Fprintf(&output, "    public double FinalDamageFrom(double %[1]s, World.DamageSource %[2]s) =>\n        PluginBridge.Host.FinalPlayerDamage(_invocation, Id, %[1]s, %[2]s);\n", spec.Damage, spec.Source)
	output.WriteString("}\n")
	return output.Bytes()
}
