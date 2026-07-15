package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type playerVisibilitySpec struct {
	Parameter string
}

func inspectPlayerVisibility(path string) (playerVisibilitySpec, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return playerVisibilitySpec{}, err
	}
	found := map[string]*ast.FuncDecl{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && playerMethod(function) && (function.Name.Name == "HideEntity" || function.Name.Name == "ShowEntity") {
			found[function.Name.Name] = function
		}
	}
	for _, name := range []string{"HideEntity", "ShowEntity"} {
		function := found[name]
		if function == nil {
			return playerVisibilitySpec{}, fmt.Errorf("Dragonfly player.Player has no %s method", name)
		}
		if got := goFunctionSignature(function); got != (goSignature{Parameters: "world.Entity"}) {
			return playerVisibilitySpec{}, fmt.Errorf("Dragonfly player.Player.%s signature changed: %+v", name, got)
		}
	}
	hide := found["HideEntity"].Type.Params.List
	show := found["ShowEntity"].Type.Params.List
	if len(hide) != 1 || len(hide[0].Names) != 1 || len(show) != 1 || len(show[0].Names) != 1 ||
		hide[0].Names[0].Name != show[0].Names[0].Name {
		return playerVisibilitySpec{}, fmt.Errorf("Dragonfly player visibility parameter names changed")
	}
	return playerVisibilitySpec{Parameter: hide[0].Names[0].Name}, nil
}

func generatePlayerVisibility(spec playerVisibilitySpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\n")
	output.WriteString("namespace Dragonfly;\n\npublic sealed partial class Player\n{\n")
	fmt.Fprintf(&output, "    public void HideEntity(World.Entity %s) => PluginBridge.Host.SetPlayerEntityVisible(_invocation, Id, %s, false);\n", spec.Parameter, spec.Parameter)
	fmt.Fprintf(&output, "    public void ShowEntity(World.Entity %s) => PluginBridge.Host.SetPlayerEntityVisible(_invocation, Id, %s, true);\n", spec.Parameter, spec.Parameter)
	output.WriteString("}\n")
	return output.Bytes()
}
