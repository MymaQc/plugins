package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

var selectedPlayerEntityActionMethods = []string{
	"UseItemOnEntity",
	"AttackEntity",
}

type playerEntityActionMethod struct {
	Name      string
	Parameter string
}

func inspectPlayerEntityActionMethods(path string) ([]playerEntityActionMethod, error) {
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
	methods := make([]playerEntityActionMethod, 0, len(selectedPlayerEntityActionMethods))
	for _, name := range selectedPlayerEntityActionMethods {
		function := found[name]
		if function == nil {
			return nil, fmt.Errorf("Dragonfly player.Player has no %s method", name)
		}
		if got := goFunctionSignature(function); got != (goSignature{Parameters: "world.Entity", Results: "bool"}) {
			return nil, fmt.Errorf("Dragonfly player.Player.%s signature changed: %+v", name, got)
		}
		parameters := function.Type.Params.List
		if len(parameters) != 1 || len(parameters[0].Names) != 1 {
			return nil, fmt.Errorf("Dragonfly player.Player.%s parameter changed", name)
		}
		methods = append(methods, playerEntityActionMethod{Name: name, Parameter: parameters[0].Names[0].Name})
	}
	return methods, nil
}

func generatePlayerEntityActions(methods []playerEntityActionMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\n")
	output.WriteString("using Dragonfly.Native;\n\nnamespace Dragonfly;\n\npublic sealed partial class Player\n{\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "    public bool %[1]s(World.Entity %[2]s) =>\n        PluginBridge.Host.RunPlayerEntityAction(_invocation, Id, %[2]s, Abi.PlayerEntityAction%[1]s);\n", method.Name, method.Parameter)
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generateNativePlayerEntityActions(methods []playerEntityActionMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\npackage native\n\n")
	output.WriteString("type PlayerEntityActionKind uint32\n\nconst (\n")
	for index, method := range methods {
		fmt.Fprintf(&output, "\t%-44s PlayerEntityActionKind = %d\n", "PlayerEntityAction"+method.Name, index)
	}
	output.WriteString(")\n")
	return output.Bytes()
}

func generateCSharpPlayerEntityActions(methods []playerEntityActionMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\nnamespace Dragonfly.Native;\n\npublic static partial class Abi\n{\n")
	for index, method := range methods {
		fmt.Fprintf(&output, "    public const uint PlayerEntityAction%s = %d;\n", method.Name, index)
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generateHostPlayerEntityActions(methods []playerEntityActionMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\npackage host\n\n")
	output.WriteString("import (\n\t\"github.com/bedrock-gophers/plugins/internal/native\"\n\t\"github.com/df-mc/dragonfly/server/player\"\n\t\"github.com/df-mc/dragonfly/server/world\"\n)\n\n")
	output.WriteString("func runExactPlayerEntityAction(connected *player.Player, entity world.Entity, kind native.PlayerEntityActionKind) (bool, bool) {\n\tswitch kind {\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "\tcase native.PlayerEntityAction%s:\n\t\treturn connected.%s(entity), true\n", method.Name, method.Name)
	}
	output.WriteString("\tdefault:\n\t\treturn false, false\n\t}\n}\n")
	return output.Bytes()
}
