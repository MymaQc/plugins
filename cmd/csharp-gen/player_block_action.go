package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

var selectedPlayerBlockActionMethods = []string{
	"BreakBlock",
	"ContinueBreaking",
	"PickBlock",
	"Sleep",
	"StartBreaking",
	"UseItemOnBlock",
}

type playerBlockActionMethod struct {
	Name       string
	Parameters []string
}

func inspectPlayerBlockActionMethods(path string) ([]playerBlockActionMethod, error) {
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
	want := map[string]goSignature{
		"BreakBlock":       {Parameters: "cube.Pos"},
		"ContinueBreaking": {Parameters: "cube.Face"},
		"PickBlock":        {Parameters: "cube.Pos"},
		"Sleep":            {Parameters: "cube.Pos"},
		"StartBreaking":    {Parameters: "cube.Pos, cube.Face"},
		"UseItemOnBlock":   {Parameters: "cube.Pos, cube.Face, mgl64.Vec3"},
	}
	methods := make([]playerBlockActionMethod, 0, len(selectedPlayerBlockActionMethods))
	for _, name := range selectedPlayerBlockActionMethods {
		function := found[name]
		if function == nil {
			return nil, fmt.Errorf("Dragonfly player.Player has no %s method", name)
		}
		if got := goFunctionSignature(function); got != want[name] {
			return nil, fmt.Errorf("Dragonfly player.Player.%s signature changed: %+v", name, got)
		}
		method := playerBlockActionMethod{Name: name}
		for _, field := range function.Type.Params.List {
			for _, parameter := range field.Names {
				method.Parameters = append(method.Parameters, parameter.Name)
			}
		}
		methods = append(methods, method)
	}
	return methods, nil
}

func generatePlayerBlockActions(methods []playerBlockActionMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\n")
	output.WriteString("using Dragonfly.Native;\n\nnamespace Dragonfly;\n\npublic sealed partial class Player\n{\n")
	for _, method := range methods {
		switch method.Name {
		case "BreakBlock", "PickBlock", "Sleep":
			fmt.Fprintf(&output, "    public void %[1]s(Cube.Pos %[2]s) =>\n        PluginBridge.Host.RunPlayerBlockAction(_invocation, Id, Abi.PlayerBlockAction%[1]s, %[2]s, default, default);\n", method.Name, method.Parameters[0])
		case "ContinueBreaking":
			fmt.Fprintf(&output, "    public void ContinueBreaking(Cube.Face %[1]s) =>\n        PluginBridge.Host.RunPlayerBlockAction(_invocation, Id, Abi.PlayerBlockActionContinueBreaking, default, %[1]s, default);\n", method.Parameters[0])
		case "StartBreaking":
			fmt.Fprintf(&output, "    public void StartBreaking(Cube.Pos %[1]s, Cube.Face %[2]s) =>\n        PluginBridge.Host.RunPlayerBlockAction(_invocation, Id, Abi.PlayerBlockActionStartBreaking, %[1]s, %[2]s, default);\n", method.Parameters[0], method.Parameters[1])
		case "UseItemOnBlock":
			fmt.Fprintf(&output, "    public void UseItemOnBlock(Cube.Pos %[1]s, Cube.Face %[2]s, Vector3 %[3]s) =>\n        PluginBridge.Host.RunPlayerBlockAction(_invocation, Id, Abi.PlayerBlockActionUseItemOnBlock, %[1]s, %[2]s, %[3]s);\n", method.Parameters[0], method.Parameters[1], method.Parameters[2])
		default:
			panic("unsupported player block action method: " + method.Name)
		}
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generateNativePlayerBlockActions(methods []playerBlockActionMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\npackage native\n\n")
	output.WriteString("type PlayerBlockActionKind uint32\n\nconst (\n")
	for index, method := range methods {
		fmt.Fprintf(&output, "\t%-45s PlayerBlockActionKind = %d\n", "PlayerBlockAction"+method.Name, index)
	}
	output.WriteString(")\n")
	return output.Bytes()
}

func generateCSharpPlayerBlockActions(methods []playerBlockActionMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\nnamespace Dragonfly.Native;\n\npublic static partial class Abi\n{\n")
	for index, method := range methods {
		fmt.Fprintf(&output, "    public const uint PlayerBlockAction%s = %d;\n", method.Name, index)
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generateHostPlayerBlockActions(methods []playerBlockActionMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\npackage host\n\n")
	output.WriteString("import (\n\t\"github.com/bedrock-gophers/plugins/internal/native\"\n\t\"github.com/df-mc/dragonfly/server/block/cube\"\n\t\"github.com/df-mc/dragonfly/server/player\"\n\t\"github.com/go-gl/mathgl/mgl64\"\n)\n\n")
	output.WriteString("func runPlayerBlockAction(connected *player.Player, kind native.PlayerBlockActionKind, position native.BlockPos, face int32, clickPosition native.Vec3) bool {\n")
	output.WriteString("\tpos := cube.Pos{int(position.X), int(position.Y), int(position.Z)}\n")
	output.WriteString("\tswitch kind {\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "\tcase native.PlayerBlockAction%s:\n", method.Name)
		switch method.Name {
		case "BreakBlock", "PickBlock", "Sleep":
			fmt.Fprintf(&output, "\t\tconnected.%s(pos)\n", method.Name)
		case "ContinueBreaking":
			output.WriteString("\t\tconnected.ContinueBreaking(cube.Face(face))\n")
		case "StartBreaking":
			output.WriteString("\t\tconnected.StartBreaking(pos, cube.Face(face))\n")
		case "UseItemOnBlock":
			output.WriteString("\t\tconnected.UseItemOnBlock(pos, cube.Face(face), mgl64.Vec3{clickPosition.X, clickPosition.Y, clickPosition.Z})\n")
		}
		output.WriteString("\t\treturn true\n")
	}
	output.WriteString("\tdefault:\n\t\treturn false\n\t}\n}\n")
	return output.Bytes()
}
