package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
)

var selectedPlayerControlMethods = []string{
	"ShowHudElement",
	"HideHudElement",
	"HudElementHidden",
	"LockInput",
	"UnlockInput",
	"InputLocked",
}

type playerControlMethod struct {
	Name      string
	Parameter string
	Returns   bool
}

type playerControlValue struct {
	Name  string
	Value uint32
}

type playerControlSpec struct {
	Methods     []playerControlMethod
	HudElements []playerControlValue
	InputLocks  []playerControlValue
}

func inspectPlayerControls(playerPath, hudPath, inputPath, packetPath string) (playerControlSpec, error) {
	methods, err := inspectPlayerControlMethods(playerPath)
	if err != nil {
		return playerControlSpec{}, err
	}
	hudElements, err := inspectOpaqueConstructors(hudPath, "Element", nil)
	if err != nil {
		return playerControlSpec{}, fmt.Errorf("inspect hud elements: %w", err)
	}
	packetLocks, err := inspectInputLockConstants(packetPath)
	if err != nil {
		return playerControlSpec{}, err
	}
	inputLocks, err := inspectOpaqueConstructors(inputPath, "Lock", packetLocks)
	if err != nil {
		return playerControlSpec{}, fmt.Errorf("inspect input locks: %w", err)
	}
	return playerControlSpec{Methods: methods, HudElements: hudElements, InputLocks: inputLocks}, nil
}

func inspectPlayerControlMethods(path string) ([]playerControlMethod, error) {
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
		"ShowHudElement":   {Parameters: "hud.Element"},
		"HideHudElement":   {Parameters: "hud.Element"},
		"HudElementHidden": {Parameters: "hud.Element", Results: "bool"},
		"LockInput":        {Parameters: "input.Lock"},
		"UnlockInput":      {Parameters: "input.Lock"},
		"InputLocked":      {Parameters: "input.Lock", Results: "bool"},
	}
	methods := make([]playerControlMethod, 0, len(selectedPlayerControlMethods))
	for _, name := range selectedPlayerControlMethods {
		function := found[name]
		if function == nil {
			return nil, fmt.Errorf("Dragonfly player.Player has no %s method", name)
		}
		if signature := goFunctionSignature(function); signature != want[name] {
			return nil, fmt.Errorf("Dragonfly player.Player.%s signature changed: %+v", name, signature)
		}
		parameter := function.Type.Params.List[0].Names
		if len(parameter) != 1 {
			return nil, fmt.Errorf("Dragonfly player.Player.%s parameter name changed", name)
		}
		methods = append(methods, playerControlMethod{
			Name: name, Parameter: parameter[0].Name, Returns: want[name].Results == "bool",
		})
	}
	return methods, nil
}

func inspectOpaqueConstructors(path, resultType string, constants map[string]uint32) ([]playerControlValue, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	var values []playerControlValue
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil || !ast.IsExported(function.Name.Name) || function.Name.Name == "All" ||
			goFunctionSignature(function) != (goSignature{Results: resultType}) || function.Body == nil || len(function.Body.List) != 1 {
			continue
		}
		statement, ok := function.Body.List[0].(*ast.ReturnStmt)
		if !ok || len(statement.Results) != 1 {
			return nil, fmt.Errorf("%s constructor body changed", function.Name.Name)
		}
		literal, ok := statement.Results[0].(*ast.CompositeLit)
		if !ok || formatGoExpression(literal.Type) != resultType || len(literal.Elts) != 1 {
			return nil, fmt.Errorf("%s constructor value changed", function.Name.Name)
		}
		value, err := opaqueConstructorValue(literal.Elts[0], constants)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", function.Name.Name, err)
		}
		values = append(values, playerControlValue{Name: function.Name.Name, Value: value})
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("no %s constructors found", resultType)
	}
	all, err := inspectOpaqueAll(file, resultType)
	if err != nil {
		return nil, err
	}
	wantAll := make([]string, 0, len(values))
	for _, value := range values {
		wantAll = append(wantAll, value.Name)
	}
	if !slices.Equal(all, wantAll) {
		return nil, fmt.Errorf("%s All changed: got %v, want %v", resultType, all, wantAll)
	}
	return values, nil
}

func inspectOpaqueAll(file *ast.File, resultType string) ([]string, error) {
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil || function.Name.Name != "All" {
			continue
		}
		if goFunctionSignature(function) != (goSignature{Results: "[]" + resultType}) ||
			function.Body == nil || len(function.Body.List) != 1 {
			return nil, fmt.Errorf("%s All signature or body changed", resultType)
		}
		statement, ok := function.Body.List[0].(*ast.ReturnStmt)
		if !ok || len(statement.Results) != 1 {
			return nil, fmt.Errorf("%s All return changed", resultType)
		}
		literal, ok := statement.Results[0].(*ast.CompositeLit)
		if !ok || formatGoExpression(literal.Type) != "[]"+resultType {
			return nil, fmt.Errorf("%s All value changed", resultType)
		}
		values := make([]string, 0, len(literal.Elts))
		for _, expression := range literal.Elts {
			call, ok := expression.(*ast.CallExpr)
			if !ok || len(call.Args) != 0 {
				return nil, fmt.Errorf("%s All entry changed", resultType)
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("%s All entry changed", resultType)
			}
			values = append(values, name.Name)
		}
		return values, nil
	}
	return nil, fmt.Errorf("%s All not found", resultType)
}

func opaqueConstructorValue(expression ast.Expr, constants map[string]uint32) (uint32, error) {
	if literal, ok := expression.(*ast.BasicLit); ok && literal.Kind == token.INT && constants == nil {
		value, err := strconv.ParseUint(literal.Value, 0, 32)
		return uint32(value), err
	}
	call, ok := expression.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return 0, fmt.Errorf("opaque constructor expression changed")
	}
	selector, ok := call.Args[0].(*ast.SelectorExpr)
	if !ok || formatGoExpression(selector.X) != "packet" {
		return 0, fmt.Errorf("packet constant expression changed")
	}
	value, ok := constants[selector.Sel.Name]
	if !ok {
		return 0, fmt.Errorf("unknown packet constant %s", selector.Sel.Name)
	}
	return value, nil
}

func inspectInputLockConstants(path string) (map[string]uint32, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.CONST || len(general.Specs) == 0 {
			continue
		}
		first, ok := general.Specs[0].(*ast.ValueSpec)
		if !ok || len(first.Values) != 1 || !inputLockIotaShift(first.Values[0]) {
			continue
		}
		values := map[string]uint32{}
		for index, raw := range general.Specs {
			spec, ok := raw.(*ast.ValueSpec)
			if !ok || len(spec.Names) != 1 || index > 0 && (len(spec.Values) != 0 || spec.Type != nil) || index+1 >= 32 {
				return nil, fmt.Errorf("gophertunnel input lock constants changed")
			}
			if name := spec.Names[0].Name; name != "_" {
				values[name] = uint32(1) << uint(index+1)
			}
		}
		return values, nil
	}
	return nil, fmt.Errorf("gophertunnel input lock constants not found in %s", filepath.Base(path))
}

func inputLockIotaShift(expression ast.Expr) bool {
	shift, ok := expression.(*ast.BinaryExpr)
	if !ok || shift.Op != token.SHL {
		return false
	}
	one, ok := shift.X.(*ast.BasicLit)
	if !ok || one.Kind != token.INT || one.Value != "1" {
		return false
	}
	parenthesised, ok := shift.Y.(*ast.ParenExpr)
	if !ok {
		return false
	}
	addition, ok := parenthesised.X.(*ast.BinaryExpr)
	if !ok || addition.Op != token.ADD {
		return false
	}
	iota, ok := addition.X.(*ast.Ident)
	if !ok || iota.Name != "iota" {
		return false
	}
	one, ok = addition.Y.(*ast.BasicLit)
	return ok && one.Kind == token.INT && one.Value == "1"
}

func generatePlayerControls(spec playerControlSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly player/hud, player/input, and player.go Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing Dragonfly.Native;\n\nnamespace Dragonfly;\n\n")
	generateOpaqueControlType(&output, "Hud", "Element", "byte", spec.HudElements)
	generateOpaqueControlType(&output, "Input", "Lock", "uint", spec.InputLocks)
	output.WriteString("public sealed partial class Player\n{\n")
	for _, method := range spec.Methods {
		typeName := "Hud.Element"
		if method.Name == "LockInput" || method.Name == "UnlockInput" || method.Name == "InputLocked" {
			typeName = "Input.Lock"
		}
		if method.Returns {
			fmt.Fprintf(&output, "    public bool %s(%s %s) => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerAction%s, new PlayerStateValue { Integer = %s.Value }).Integer != 0;\n", method.Name, typeName, method.Parameter, method.Name, method.Parameter)
		} else {
			fmt.Fprintf(&output, "    public void %s(%s %s) => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerAction%s, new PlayerStateValue { Integer = %s.Value });\n", method.Name, typeName, method.Parameter, method.Name, method.Parameter)
		}
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generateOpaqueControlType(output *bytes.Buffer, namespace, typeName, valueType string, values []playerControlValue) {
	fmt.Fprintf(output, "public static class %s\n{\n", namespace)
	fmt.Fprintf(output, "    public readonly record struct %s\n    {\n", typeName)
	fmt.Fprintf(output, "        internal %s(%s value) => Value = value;\n", typeName, valueType)
	fmt.Fprintf(output, "        internal %s Value { get; }\n", valueType)
	output.WriteString("    }\n\n")
	for _, value := range values {
		fmt.Fprintf(output, "    public static %s %s() => new(%d);\n", typeName, value.Name, value.Value)
	}
	fmt.Fprintf(output, "    public static %s[] All() => [", typeName)
	for index, value := range values {
		if index != 0 {
			output.WriteString(", ")
		}
		fmt.Fprintf(output, "%s()", value.Name)
	}
	output.WriteString("];\n}\n\n")
}
