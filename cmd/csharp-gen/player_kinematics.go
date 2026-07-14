package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

var selectedPlayerKinematicsMethods = []string{
	"Teleport",
	"Move",
	"Displace",
	"Position",
	"Velocity",
	"SetVelocity",
	"Rotation",
}

type playerKinematicsMethod struct {
	Name       string
	Parameters []string
}

func inspectPlayerKinematicsMethods(path string) ([]playerKinematicsMethod, error) {
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
		"Teleport":    {Parameters: "mgl64.Vec3"},
		"Move":        {Parameters: "mgl64.Vec3, float64, float64"},
		"Displace":    {Parameters: "mgl64.Vec3"},
		"Position":    {Results: "mgl64.Vec3"},
		"Velocity":    {Results: "mgl64.Vec3"},
		"SetVelocity": {Parameters: "mgl64.Vec3"},
		"Rotation":    {Results: "cube.Rotation"},
	}
	wantParameterCount := map[string]int{
		"Teleport": 1, "Move": 3, "Displace": 1, "Position": 0,
		"Velocity": 0, "SetVelocity": 1, "Rotation": 0,
	}
	methods := make([]playerKinematicsMethod, 0, len(selectedPlayerKinematicsMethods))
	for _, name := range selectedPlayerKinematicsMethods {
		function := found[name]
		if function == nil {
			return nil, fmt.Errorf("Dragonfly player.Player has no %s method", name)
		}
		if got := goFunctionSignature(function); got != want[name] {
			return nil, fmt.Errorf("Dragonfly player.Player.%s signature changed: %+v", name, got)
		}
		method := playerKinematicsMethod{Name: name}
		if function.Type.Params != nil {
			for _, field := range function.Type.Params.List {
				for _, parameter := range field.Names {
					method.Parameters = append(method.Parameters, parameter.Name)
				}
			}
		}
		if len(method.Parameters) != wantParameterCount[name] {
			return nil, fmt.Errorf("Dragonfly player.Player.%s parameter names changed", name)
		}
		methods = append(methods, method)
	}
	return methods, nil
}

func generatePlayerKinematicsMethods(methods []playerKinematicsMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\n")
	output.WriteString("using Dragonfly.Native;\n\nnamespace Dragonfly;\n\npublic sealed partial class Player\n{\n")
	for _, method := range methods {
		switch method.Name {
		case "Teleport":
			fmt.Fprintf(&output, "    public void Teleport(Vector3 %s) =>\n        PluginBridge.Host.TransformPlayer(_invocation, Id, Abi.PlayerTransformTeleport, %s, 0, 0);\n", method.Parameters[0], method.Parameters[0])
		case "Move":
			fmt.Fprintf(&output, "    public void Move(Vector3 %[1]s, double %[2]s, double %[3]s) =>\n        PluginBridge.Host.TransformPlayer(_invocation, Id, Abi.PlayerTransformMove, %[1]s, %[2]s, %[3]s);\n", method.Parameters[0], method.Parameters[1], method.Parameters[2])
		case "Displace":
			fmt.Fprintf(&output, "    public void Displace(Vector3 %s) =>\n        PluginBridge.Host.TransformPlayer(_invocation, Id, Abi.PlayerTransformDisplace, %s, 0, 0);\n", method.Parameters[0], method.Parameters[0])
		case "Position":
			output.WriteString("    public Vector3 Position() => PluginBridge.Host.TryReadPlayerKinematics(_invocation, Id, out var state) ? state.Position : _position;\n")
		case "Velocity":
			output.WriteString("    public Vector3 Velocity() => PluginBridge.Host.ReadPlayerKinematics(_invocation, Id).Velocity;\n")
		case "SetVelocity":
			fmt.Fprintf(&output, "    public void SetVelocity(Vector3 %s) =>\n        PluginBridge.Host.TransformPlayer(_invocation, Id, Abi.PlayerTransformVelocity, %s, 0, 0);\n", method.Parameters[0], method.Parameters[0])
		case "Rotation":
			output.WriteString("    public Rotation Rotation() => PluginBridge.Host.ReadPlayerKinematics(_invocation, Id).Rotation;\n")
		default:
			panic("unsupported player kinematics method: " + method.Name)
		}
	}
	output.WriteString("}\n")
	return output.Bytes()
}
