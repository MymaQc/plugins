package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

var selectedPlayerPresentationMethods = []string{
	"EnableInstantRespawn",
	"DisableInstantRespawn",
	"ShowCoordinates",
	"HideCoordinates",
	"SendSleepingIndicator",
	"CloseDialogue",
	"RemoveBossBar",
	"RemoveScoreboard",
	"NameTag",
	"ScoreTag",
	"SendToast",
}

type playerPresentationMethod struct {
	Name       string
	Parameters []string
}

func inspectPlayerPresentationMethods(path string) ([]playerPresentationMethod, error) {
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
		"EnableInstantRespawn":  {},
		"DisableInstantRespawn": {},
		"ShowCoordinates":       {},
		"HideCoordinates":       {},
		"SendSleepingIndicator": {Parameters: "int, int"},
		"CloseDialogue":         {},
		"RemoveBossBar":         {},
		"RemoveScoreboard":      {},
		"NameTag":               {Results: "string"},
		"ScoreTag":              {Results: "string"},
		"SendToast":             {Parameters: "string, string"},
	}
	methods := make([]playerPresentationMethod, 0, len(selectedPlayerPresentationMethods))
	for _, name := range selectedPlayerPresentationMethods {
		function := found[name]
		if function == nil {
			return nil, fmt.Errorf("Dragonfly player.Player has no %s method", name)
		}
		if got := goFunctionSignature(function); got != want[name] {
			return nil, fmt.Errorf("Dragonfly player.Player.%s signature changed: %+v", name, got)
		}
		method := playerPresentationMethod{Name: name}
		if function.Type.Params != nil {
			for _, field := range function.Type.Params.List {
				for _, parameter := range field.Names {
					method.Parameters = append(method.Parameters, parameter.Name)
				}
			}
		}
		wantParameters := 0
		if want[name].Parameters != "" {
			wantParameters = 2
		}
		if len(method.Parameters) != wantParameters {
			return nil, fmt.Errorf("Dragonfly player.Player.%s parameter names changed", name)
		}
		methods = append(methods, method)
	}
	return methods, nil
}

func generatePlayerPresentationMethods(methods []playerPresentationMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing Dragonfly.Native;\n\nnamespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class Player\n{\n")
	for _, method := range methods {
		switch method.Name {
		case "EnableInstantRespawn", "DisableInstantRespawn", "ShowCoordinates", "HideCoordinates", "CloseDialogue", "RemoveBossBar":
			fmt.Fprintf(&output, "    public void %s() => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerAction%s, default);\n", method.Name, method.Name)
		case "SendSleepingIndicator":
			fmt.Fprintf(&output, "    public void SendSleepingIndicator(int %s, int %s) => PluginBridge.Host.RunPlayerAction(_invocation, Id, Abi.PlayerActionSendSleepingIndicator, new PlayerStateValue { Integer = %s, Number = %s });\n", method.Parameters[0], method.Parameters[1], method.Parameters[0], method.Parameters[1])
		case "RemoveScoreboard":
			output.WriteString("    public void RemoveScoreboard() => PluginBridge.Host.RemovePlayerScoreboard(_invocation, Id);\n")
		case "NameTag", "ScoreTag":
			fmt.Fprintf(&output, "    public string %s() => PluginBridge.Host.PlayerString(_invocation, Id, Abi.PlayerString%s);\n", method.Name, method.Name)
		case "SendToast":
			fmt.Fprintf(&output, "    public void SendToast(string %s, string %s) => PluginBridge.Host.SendPlayerToast(_invocation, Id, %s, %s);\n", method.Parameters[0], method.Parameters[1], method.Parameters[0], method.Parameters[1])
		default:
			panic("unsupported player presentation method: " + method.Name)
		}
	}
	output.WriteString("}\n")
	return output.Bytes()
}
