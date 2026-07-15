package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

var selectedPlayerActionMethods = []string{
	"AbortBreaking",
	"ClearInputLocks",
	"FinishBreaking",
	"Jump",
	"MoveItemsToInventory",
	"PunchAir",
	"ReleaseItem",
	"RemoveAllDebugShapes",
	"SwingArm",
	"UseItem",
	"Wake",
}

type playerActionMethod struct {
	Name string
}

func inspectPlayerActionMethods(path string) ([]playerActionMethod, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	found := map[string]*ast.FuncDecl{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && playerMethod(function) {
			found[function.Name.Name] = function
		}
	}
	methods := make([]playerActionMethod, 0, len(selectedPlayerActionMethods))
	for _, name := range selectedPlayerActionMethods {
		function := found[name]
		if function == nil {
			return nil, fmt.Errorf("Dragonfly player.Player has no %s method", name)
		}
		if signature := goFunctionSignature(function); signature != (goSignature{}) {
			return nil, fmt.Errorf("Dragonfly player.Player.%s signature changed: %+v", name, signature)
		}
		methods = append(methods, playerActionMethod{Name: name})
	}
	return methods, nil
}

func generatePlayerActionMethods(methods []playerActionMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing Dragonfly.Native;\n\nnamespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class Player\n{\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "    public void %s() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerAction%s, default);\n", method.Name, method.Name)
	}
	output.WriteString("}\n")
	return output.Bytes()
}
