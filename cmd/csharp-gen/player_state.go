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
	"Hurt",
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
	"Speed",
	"SetSpeed",
	"FlightSpeed",
	"SetFlightSpeed",
	"VerticalFlightSpeed",
	"SetVerticalFlightSpeed",
	"ResetFallDistance",
	"FallDistance",
	"SetAbsorption",
	"Absorption",
	"Dead",
	"OnGround",
	"EyeHeight",
	"TorsoHeight",
	"Breathing",
	"StartSprinting",
	"StopSprinting",
	"Sprinting",
	"StartSneaking",
	"StopSneaking",
	"Sneaking",
	"StartSwimming",
	"StopSwimming",
	"Swimming",
	"StartCrawling",
	"StopCrawling",
	"Crawling",
	"StartGliding",
	"StopGliding",
	"Gliding",
	"StartFlying",
	"StopFlying",
	"Flying",
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
		"Food":                   {Results: "int"},
		"SetFood":                {Parameters: "int"},
		"Health":                 {Results: "float64"},
		"MaxHealth":              {Results: "float64"},
		"SetMaxHealth":           {Parameters: "float64"},
		"Heal":                   {Parameters: "float64, world.HealingSource", Results: "float64"},
		"Hurt":                   {Parameters: "float64, world.DamageSource", Results: "float64, bool"},
		"ExperienceLevel":        {Results: "int"},
		"SetExperienceLevel":     {Parameters: "int"},
		"ExperienceProgress":     {Results: "float64"},
		"SetExperienceProgress":  {Parameters: "float64"},
		"Scale":                  {Results: "float64"},
		"SetScale":               {Parameters: "float64"},
		"Invisible":              {Results: "bool"},
		"SetInvisible":           {},
		"SetVisible":             {},
		"Immobile":               {Results: "bool"},
		"SetImmobile":            {},
		"SetMobile":              {},
		"Speed":                  {Results: "float64"},
		"SetSpeed":               {Parameters: "float64"},
		"FlightSpeed":            {Results: "float64"},
		"SetFlightSpeed":         {Parameters: "float64"},
		"VerticalFlightSpeed":    {Results: "float64"},
		"SetVerticalFlightSpeed": {Parameters: "float64"},
		"ResetFallDistance":      {},
		"FallDistance":           {Results: "float64"},
		"SetAbsorption":          {Parameters: "float64"},
		"Absorption":             {Results: "float64"},
		"Dead":                   {Results: "bool"},
		"OnGround":               {Results: "bool"},
		"EyeHeight":              {Results: "float64"},
		"TorsoHeight":            {Results: "float64"},
		"Breathing":              {Results: "bool"},
		"StartSprinting":         {},
		"StopSprinting":          {},
		"Sprinting":              {Results: "bool"},
		"StartSneaking":          {},
		"StopSneaking":           {},
		"Sneaking":               {Results: "bool"},
		"StartSwimming":          {},
		"StopSwimming":           {},
		"Swimming":               {Results: "bool"},
		"StartCrawling":          {},
		"StopCrawling":           {},
		"Crawling":               {Results: "bool"},
		"StartGliding":           {},
		"StopGliding":            {},
		"Gliding":                {Results: "bool"},
		"StartFlying":            {},
		"StopFlying":             {},
		"Flying":                 {Results: "bool"},
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
		case "Hurt":
			fmt.Fprintf(&output, "    public (double Damage, bool Vulnerable) Hurt(double %s, World.DamageSource %s) => PluginBridge.Host.HurtPlayer(_invocation, Id, %s, %s);\n", methodsParameter(method, 0), methodsParameter(method, 1), methodsParameter(method, 0), methodsParameter(method, 1))
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
		case "Speed":
			output.WriteString("    public double Speed() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateSpeed).Number;\n")
		case "SetSpeed":
			fmt.Fprintf(&output, "    public void SetSpeed(double %s) => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateSpeed, new PlayerStateValue { Number = %s });\n", parameter, parameter)
		case "FlightSpeed":
			output.WriteString("    public double FlightSpeed() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateFlightSpeed).Number;\n")
		case "SetFlightSpeed":
			fmt.Fprintf(&output, "    public void SetFlightSpeed(double %s) => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateFlightSpeed, new PlayerStateValue { Number = %s });\n", parameter, parameter)
		case "VerticalFlightSpeed":
			output.WriteString("    public double VerticalFlightSpeed() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateVerticalFlightSpeed).Number;\n")
		case "SetVerticalFlightSpeed":
			fmt.Fprintf(&output, "    public void SetVerticalFlightSpeed(double %s) => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateVerticalFlightSpeed, new PlayerStateValue { Number = %s });\n", parameter, parameter)
		case "ResetFallDistance":
			output.WriteString("    public void ResetFallDistance() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateFallDistance, default);\n")
		case "FallDistance":
			output.WriteString("    public double FallDistance() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateFallDistance).Number;\n")
		case "SetAbsorption":
			fmt.Fprintf(&output, "    public void SetAbsorption(double %s) => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateAbsorption, new PlayerStateValue { Number = %s });\n", parameter, parameter)
		case "Absorption":
			output.WriteString("    public double Absorption() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateAbsorption).Number;\n")
		case "Dead":
			output.WriteString("    public bool Dead() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateDead).Integer != 0;\n")
		case "OnGround":
			output.WriteString("    public bool OnGround() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateOnGround).Integer != 0;\n")
		case "EyeHeight":
			output.WriteString("    public double EyeHeight() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateEyeHeight).Number;\n")
		case "TorsoHeight":
			output.WriteString("    public double TorsoHeight() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateTorsoHeight).Number;\n")
		case "Breathing":
			output.WriteString("    public bool Breathing() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateBreathing).Integer != 0;\n")
		case "StartSprinting":
			output.WriteString("    public void StartSprinting() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateSprinting, new PlayerStateValue { Integer = 1 });\n")
		case "StopSprinting":
			output.WriteString("    public void StopSprinting() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateSprinting, default);\n")
		case "Sprinting":
			output.WriteString("    public bool Sprinting() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateSprinting).Integer != 0;\n")
		case "StartSneaking":
			output.WriteString("    public void StartSneaking() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateSneaking, new PlayerStateValue { Integer = 1 });\n")
		case "StopSneaking":
			output.WriteString("    public void StopSneaking() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateSneaking, default);\n")
		case "Sneaking":
			output.WriteString("    public bool Sneaking() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateSneaking).Integer != 0;\n")
		case "StartSwimming":
			output.WriteString("    public void StartSwimming() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateSwimming, new PlayerStateValue { Integer = 1 });\n")
		case "StopSwimming":
			output.WriteString("    public void StopSwimming() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateSwimming, default);\n")
		case "Swimming":
			output.WriteString("    public bool Swimming() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateSwimming).Integer != 0;\n")
		case "StartCrawling":
			output.WriteString("    public void StartCrawling() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateCrawling, new PlayerStateValue { Integer = 1 });\n")
		case "StopCrawling":
			output.WriteString("    public void StopCrawling() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateCrawling, default);\n")
		case "Crawling":
			output.WriteString("    public bool Crawling() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateCrawling).Integer != 0;\n")
		case "StartGliding":
			output.WriteString("    public void StartGliding() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateGliding, new PlayerStateValue { Integer = 1 });\n")
		case "StopGliding":
			output.WriteString("    public void StopGliding() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateGliding, default);\n")
		case "Gliding":
			output.WriteString("    public bool Gliding() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateGliding).Integer != 0;\n")
		case "StartFlying":
			output.WriteString("    public void StartFlying() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateFlying, new PlayerStateValue { Integer = 1 });\n")
		case "StopFlying":
			output.WriteString("    public void StopFlying() => PluginBridge.Host.SetPlayerState(_invocation, Id, Abi.PlayerStateFlying, default);\n")
		case "Flying":
			output.WriteString("    public bool Flying() => PluginBridge.Host.GetPlayerState(_invocation, Id, Abi.PlayerStateFlying).Integer != 0;\n")
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
