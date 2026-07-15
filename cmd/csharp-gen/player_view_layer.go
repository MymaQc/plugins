package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

var selectedPlayerViewLayerMethods = []string{
	"ViewNameTag",
	"ViewPublicNameTag",
	"ViewScoreTag",
	"ViewPublicScoreTag",
	"ViewVisibility",
	"RemoveViewLayer",
}

var selectedVisibilityLevels = []string{
	"PublicVisibility",
	"EnforceInvisible",
	"EnforceVisible",
}

type playerViewLayerMethod struct {
	Name       string
	Parameters []string
}

type visibilityLevel struct {
	Name  string
	Value uint8
}

type playerViewLayerSpec struct {
	Methods []playerViewLayerMethod
	Levels  []visibilityLevel
}

func inspectPlayerViewLayer(playerPath, visibilityPath string) (playerViewLayerSpec, error) {
	playerFile, err := parser.ParseFile(token.NewFileSet(), playerPath, nil, 0)
	if err != nil {
		return playerViewLayerSpec{}, err
	}
	found := map[string]*ast.FuncDecl{}
	for _, declaration := range playerFile.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && playerMethod(function) {
			found[function.Name.Name] = function
		}
	}
	want := map[string]goSignature{
		"ViewNameTag":        {Parameters: "world.Entity, string"},
		"ViewPublicNameTag":  {Parameters: "world.Entity"},
		"ViewScoreTag":       {Parameters: "world.Entity, string"},
		"ViewPublicScoreTag": {Parameters: "world.Entity"},
		"ViewVisibility":     {Parameters: "world.Entity, world.VisibilityLevel"},
		"RemoveViewLayer":    {Parameters: "world.Entity"},
	}
	methods := make([]playerViewLayerMethod, 0, len(selectedPlayerViewLayerMethods))
	for _, name := range selectedPlayerViewLayerMethods {
		function := found[name]
		if function == nil {
			return playerViewLayerSpec{}, fmt.Errorf("Dragonfly player.Player has no %s method", name)
		}
		if got := goFunctionSignature(function); got != want[name] {
			return playerViewLayerSpec{}, fmt.Errorf("Dragonfly player.Player.%s signature changed: %+v", name, got)
		}
		method := playerViewLayerMethod{Name: name}
		for _, field := range function.Type.Params.List {
			for _, parameter := range field.Names {
				method.Parameters = append(method.Parameters, parameter.Name)
			}
		}
		methods = append(methods, method)
	}

	visibilityFile, err := parser.ParseFile(token.NewFileSet(), visibilityPath, nil, 0)
	if err != nil {
		return playerViewLayerSpec{}, err
	}
	constructors := map[string]*ast.FuncDecl{}
	visibilityType := false
	for _, declaration := range visibilityFile.Decls {
		switch declaration := declaration.(type) {
		case *ast.GenDecl:
			for _, raw := range declaration.Specs {
				typeSpec, ok := raw.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != "VisibilityLevel" {
					continue
				}
				structure, ok := typeSpec.Type.(*ast.StructType)
				if !ok || len(structure.Fields.List) != 1 {
					continue
				}
				field := structure.Fields.List[0]
				identifier, ok := field.Type.(*ast.Ident)
				visibilityType = ok && len(field.Names) == 0 && identifier.Name == "visibility"
			}
		case *ast.FuncDecl:
			constructors[declaration.Name.Name] = declaration
		}
	}
	if !visibilityType {
		return playerViewLayerSpec{}, fmt.Errorf("Dragonfly world.VisibilityLevel shape changed")
	}
	levels := make([]visibilityLevel, 0, len(selectedVisibilityLevels))
	for _, name := range selectedVisibilityLevels {
		function := constructors[name]
		if function == nil {
			return playerViewLayerSpec{}, fmt.Errorf("Dragonfly world has no %s constructor", name)
		}
		if got := goFunctionSignature(function); got != (goSignature{Results: "VisibilityLevel"}) {
			return playerViewLayerSpec{}, fmt.Errorf("Dragonfly world.%s signature changed: %+v", name, got)
		}
		if function.Body == nil || len(function.Body.List) != 1 {
			return playerViewLayerSpec{}, fmt.Errorf("Dragonfly world.%s body changed", name)
		}
		statement, ok := function.Body.List[0].(*ast.ReturnStmt)
		if !ok || len(statement.Results) != 1 {
			return playerViewLayerSpec{}, fmt.Errorf("Dragonfly world.%s return changed", name)
		}
		literal, ok := statement.Results[0].(*ast.CompositeLit)
		if !ok || len(literal.Elts) != 1 {
			return playerViewLayerSpec{}, fmt.Errorf("Dragonfly world.%s value changed", name)
		}
		valueLiteral, ok := literal.Elts[0].(*ast.BasicLit)
		if !ok || valueLiteral.Kind != token.INT {
			return playerViewLayerSpec{}, fmt.Errorf("Dragonfly world.%s value changed", name)
		}
		value, err := strconv.ParseUint(valueLiteral.Value, 0, 8)
		if err != nil {
			return playerViewLayerSpec{}, fmt.Errorf("Dragonfly world.%s value changed: %w", name, err)
		}
		levels = append(levels, visibilityLevel{Name: name, Value: uint8(value)})
	}
	return playerViewLayerSpec{Methods: methods, Levels: levels}, nil
}

func generatePlayerViewLayer(spec playerViewLayerSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly player.go and visibility_level.go Go AST. DO NOT EDIT.\n")
	output.WriteString("using Dragonfly.Native;\n\nnamespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class World\n{\n    public readonly record struct VisibilityLevel\n    {\n        internal VisibilityLevel(byte value) => Value = value;\n        internal byte Value { get; }\n    }\n\n")
	for _, level := range spec.Levels {
		fmt.Fprintf(&output, "    public static VisibilityLevel %s() => new(%d);\n", level.Name, level.Value)
	}
	output.WriteString("}\n\npublic sealed partial class Player\n{\n")
	for _, method := range spec.Methods {
		entity := method.Parameters[0]
		switch method.Name {
		case "ViewNameTag", "ViewScoreTag":
			fmt.Fprintf(&output, "    public void %[1]s(World.Entity %[2]s, string %[3]s) =>\n        PluginBridge.Host.RunPlayerViewLayer(_invocation, Id, %[2]s, Abi.PlayerViewLayer%[1]s, %[3]s, default);\n", method.Name, entity, method.Parameters[1])
		case "ViewVisibility":
			fmt.Fprintf(&output, "    public void ViewVisibility(World.Entity %[1]s, World.VisibilityLevel %[2]s) =>\n        PluginBridge.Host.RunPlayerViewLayer(_invocation, Id, %[1]s, Abi.PlayerViewLayerViewVisibility, string.Empty, %[2]s);\n", entity, method.Parameters[1])
		default:
			fmt.Fprintf(&output, "    public void %[1]s(World.Entity %[2]s) =>\n        PluginBridge.Host.RunPlayerViewLayer(_invocation, Id, %[2]s, Abi.PlayerViewLayer%[1]s, string.Empty, default);\n", method.Name, entity)
		}
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generateNativePlayerViewLayer(spec playerViewLayerSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly player.go and visibility_level.go Go AST. DO NOT EDIT.\npackage native\n\n")
	output.WriteString("type PlayerViewLayerKind uint32\n\nconst (\n")
	for index, method := range spec.Methods {
		fmt.Fprintf(&output, "\t%-42s PlayerViewLayerKind = %d\n", "PlayerViewLayer"+method.Name, index)
	}
	output.WriteString(")\n")
	return output.Bytes()
}

func generateCSharpPlayerViewLayer(spec playerViewLayerSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly player.go and visibility_level.go Go AST. DO NOT EDIT.\nnamespace Dragonfly.Native;\n\npublic static partial class Abi\n{\n")
	for index, method := range spec.Methods {
		fmt.Fprintf(&output, "    public const uint PlayerViewLayer%s = %d;\n", method.Name, index)
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generateHostPlayerViewLayer(spec playerViewLayerSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly player.go and visibility_level.go Go AST. DO NOT EDIT.\npackage host\n\n")
	output.WriteString("import (\n\t\"github.com/bedrock-gophers/plugins/internal/native\"\n\t\"github.com/df-mc/dragonfly/server/player\"\n\t\"github.com/df-mc/dragonfly/server/world\"\n)\n\n")
	output.WriteString("func runPlayerViewLayer(viewer *player.Player, entity world.Entity, kind native.PlayerViewLayerKind, text string, visibility uint8) bool {\n\tswitch kind {\n")
	for _, method := range spec.Methods {
		fmt.Fprintf(&output, "\tcase native.PlayerViewLayer%s:\n", method.Name)
		switch method.Name {
		case "ViewNameTag", "ViewScoreTag":
			fmt.Fprintf(&output, "\t\tviewer.%s(entity, text)\n", method.Name)
		case "ViewVisibility":
			output.WriteString("\t\tvar level world.VisibilityLevel\n\t\tswitch visibility {\n")
			for _, level := range spec.Levels {
				fmt.Fprintf(&output, "\t\tcase %d:\n\t\t\tlevel = world.%s()\n", level.Value, level.Name)
			}
			output.WriteString("\t\tdefault:\n\t\t\treturn false\n\t\t}\n\t\tviewer.ViewVisibility(entity, level)\n")
		default:
			fmt.Fprintf(&output, "\t\tviewer.%s(entity)\n", method.Name)
		}
		output.WriteString("\t\treturn true\n")
	}
	output.WriteString("\tdefault:\n\t\treturn false\n\t}\n}\n")
	return output.Bytes()
}
