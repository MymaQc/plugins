package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

var selectedPlayerStateMethods = []string{
	"Food",
	"SetFood",
	"Health",
	"MaxHealth",
	"SetMaxHealth",
	"Heal",
	"ExperienceLevel",
	"SetExperienceLevel",
	"ExperienceProgress",
	"SetExperienceProgress",
	"Scale",
	"SetScale",
	"Invisible",
	"SetInvisible",
	"SetVisible",
	"Immobile",
	"SetImmobile",
	"SetMobile",
}

type playerStateMethod struct {
	Name       string
	Parameters []string
}

func inspectPlayerStateMethods(path string) ([]playerStateMethod, error) {
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
		"Food":                  {Results: "int"},
		"SetFood":               {Parameters: "int"},
		"Health":                {Results: "float64"},
		"MaxHealth":             {Results: "float64"},
		"SetMaxHealth":          {Parameters: "float64"},
		"Heal":                  {Parameters: "float64, world.HealingSource", Results: "float64"},
		"ExperienceLevel":       {Results: "int"},
		"SetExperienceLevel":    {Parameters: "int"},
		"ExperienceProgress":    {Results: "float64"},
		"SetExperienceProgress": {Parameters: "float64"},
		"Scale":                 {Results: "float64"},
		"SetScale":              {Parameters: "float64"},
		"Invisible":             {Results: "bool"},
		"SetInvisible":          {},
		"SetVisible":            {},
		"Immobile":              {Results: "bool"},
		"SetImmobile":           {},
		"SetMobile":             {},
	}
	for _, name := range selectedPlayerStateMethods {
		function := found[name]
		if function == nil {
			return nil, fmt.Errorf("Dragonfly player.Player has no %s method", name)
		}
		if got := goFunctionSignature(function); got != want[name] {
			return nil, fmt.Errorf("Dragonfly player.Player.%s signature changed: %+v", name, got)
		}
	}
	methods := make([]playerStateMethod, 0, len(selectedPlayerStateMethods))
	for _, name := range selectedPlayerStateMethods {
		function := found[name]
		method := playerStateMethod{Name: name}
		if function.Type.Params != nil {
			for _, field := range function.Type.Params.List {
				for _, parameter := range field.Names {
					method.Parameters = append(method.Parameters, parameter.Name)
				}
			}
		}
		wantParameters := 0
		if want[name].Parameters != "" {
			wantParameters = len(function.Type.Params.List)
		}
		if len(method.Parameters) != wantParameters {
			return nil, fmt.Errorf("Dragonfly player.Player.%s parameter names changed", name)
		}
		methods = append(methods, method)
	}
	return methods, nil
}

func generatePlayerStateMethods(methods []playerStateMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing Dragonfly.Native;\n\nnamespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class Player\n{\n")
	for _, method := range methods {
		parameter := ""
		if len(method.Parameters) != 0 {
			parameter = method.Parameters[0]
		}
		switch method.Name {
		case "Food":
			output.WriteString("    public int Food() => checked((int)PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateFood).Integer);\n")
		case "SetFood":
			fmt.Fprintf(&output, "    public void SetFood(int %s) => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateFood, new PlayerStateValue { Integer = %s });\n", parameter, parameter)
		case "Health":
			output.WriteString("    public double Health() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateHealth).Number;\n")
		case "MaxHealth":
			output.WriteString("    public double MaxHealth() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateMaxHealth).Number;\n")
		case "SetMaxHealth":
			fmt.Fprintf(&output, "    public void SetMaxHealth(double %s) => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateMaxHealth, new PlayerStateValue { Number = %s });\n", parameter, parameter)
		case "Heal":
			fmt.Fprintf(&output, "    public double Heal(double %s, World.HealingSource %s) => PluginBridge.Host.HealPlayer(_invocation, Id, %s, %s);\n", methodsParameter(method, 0), methodsParameter(method, 1), methodsParameter(method, 0), methodsParameter(method, 1))
		case "ExperienceLevel":
			output.WriteString("    public int ExperienceLevel() => checked((int)PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateExperienceLevel).Integer);\n")
		case "SetExperienceLevel":
			fmt.Fprintf(&output, `    public void SetExperienceLevel(int %[1]s)
    {
        if (%[1]s < 0) throw new ArgumentOutOfRangeException(nameof(%[1]s));
        PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateExperienceLevel, new PlayerStateValue { Integer = %[1]s });
    }
`, parameter)
		case "ExperienceProgress":
			output.WriteString("    public double ExperienceProgress() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateExperienceProgress).Number;\n")
		case "SetExperienceProgress":
			fmt.Fprintf(&output, `    public void SetExperienceProgress(double %[1]s)
    {
        if (%[1]s is < 0 or > 1)
            throw new ArgumentOutOfRangeException(nameof(%[1]s));
        PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateExperienceProgress, new PlayerStateValue { Number = %[1]s });
    }
`, parameter)
		case "Scale":
			output.WriteString("    public double Scale() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateScale).Number;\n")
		case "SetScale":
			fmt.Fprintf(&output, "    public void SetScale(double %s) => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateScale, new PlayerStateValue { Number = %s });\n", parameter, parameter)
		case "Invisible":
			output.WriteString("    public bool Invisible() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateInvisible).Integer != 0;\n")
		case "SetInvisible":
			output.WriteString("    public void SetInvisible() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateInvisible, new PlayerStateValue { Integer = 1 });\n")
		case "SetVisible":
			output.WriteString("    public void SetVisible() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateInvisible, default);\n")
		case "Immobile":
			output.WriteString("    public bool Immobile() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateImmobile).Integer != 0;\n")
		case "SetImmobile":
			output.WriteString("    public void SetImmobile() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateImmobile, new PlayerStateValue { Integer = 1 });\n")
		case "SetMobile":
			output.WriteString("    public void SetMobile() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateImmobile, default);\n")
		default:
			panic("unsupported player state method: " + method.Name)
		}
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func methodsParameter(method playerStateMethod, index int) string {
	return method.Parameters[index]
}
