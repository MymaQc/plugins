package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

var selectedWorldLifecycleMethods = []string{
	"Name",
	"Dimension",
	"Range",
	"HighestLightBlocker",
	"Time",
	"SetTime",
	"StopTime",
	"StartTime",
	"TimeCycle",
	"Spawn",
	"SetSpawn",
	"PlayerSpawn",
	"SetPlayerSpawn",
	"SetRequiredSleepDuration",
	"DefaultGameMode",
	"SetTickRange",
	"SetDefaultGameMode",
	"Difficulty",
	"SetDifficulty",
	"Save",
	"Close",
}

type worldLifecycleMethod struct {
	Name       string
	Parameters []string
}

func inspectWorldLifecycleMethods(path string) ([]worldLifecycleMethod, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	found := map[string]*ast.FuncDecl{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && receiverTypeName(function) == "World" {
			found[function.Name.Name] = function
		}
	}
	want := map[string]goSignature{
		"Name":                     {Results: "string"},
		"Dimension":                {Results: "Dimension"},
		"Range":                    {Results: "cube.Range"},
		"HighestLightBlocker":      {Parameters: "int, int", Results: "int"},
		"Time":                     {Results: "int"},
		"SetTime":                  {Parameters: "int"},
		"StopTime":                 {},
		"StartTime":                {},
		"TimeCycle":                {Results: "bool"},
		"Spawn":                    {Results: "cube.Pos"},
		"SetSpawn":                 {Parameters: "cube.Pos"},
		"PlayerSpawn":              {Parameters: "uuid.UUID", Results: "cube.Pos"},
		"SetPlayerSpawn":           {Parameters: "uuid.UUID, cube.Pos"},
		"SetRequiredSleepDuration": {Parameters: "time.Duration"},
		"DefaultGameMode":          {Results: "GameMode"},
		"SetTickRange":             {Parameters: "int"},
		"SetDefaultGameMode":       {Parameters: "GameMode"},
		"Difficulty":               {Results: "Difficulty"},
		"SetDifficulty":            {Parameters: "Difficulty"},
		"Save":                     {},
		"Close":                    {Results: "error"},
	}
	wantParameterCount := map[string]int{
		"Name": 0, "Dimension": 0, "Range": 0, "HighestLightBlocker": 2, "Time": 0,
		"SetTime": 1, "StopTime": 0, "StartTime": 0, "TimeCycle": 0, "Spawn": 0,
		"SetSpawn": 1, "SetRequiredSleepDuration": 1, "DefaultGameMode": 0, "SetTickRange": 1,
		"PlayerSpawn": 1, "SetPlayerSpawn": 2,
		"SetDefaultGameMode": 1, "Difficulty": 0, "SetDifficulty": 1, "Save": 0, "Close": 0,
	}
	methods := make([]worldLifecycleMethod, 0, len(selectedWorldLifecycleMethods))
	for _, name := range selectedWorldLifecycleMethods {
		function := found[name]
		if function == nil {
			return nil, fmt.Errorf("Dragonfly world.World has no %s method", name)
		}
		if got := goFunctionSignature(function); got != want[name] {
			return nil, fmt.Errorf("Dragonfly world.World.%s signature changed: %+v", name, got)
		}
		method := worldLifecycleMethod{Name: name}
		if function.Type.Params != nil {
			for _, field := range function.Type.Params.List {
				for _, parameter := range field.Names {
					method.Parameters = append(method.Parameters, parameter.Name)
				}
			}
		}
		if len(method.Parameters) != wantParameterCount[name] {
			return nil, fmt.Errorf("Dragonfly world.World.%s parameter names changed", name)
		}
		methods = append(methods, method)
	}
	return methods, nil
}

func generateWorldLifecycleMethods(methods []worldLifecycleMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/world/world.go Go AST. DO NOT EDIT.\n")
	output.WriteString("namespace Dragonfly;\n\npublic sealed partial class World\n{\n")
	for _, method := range methods {
		switch method.Name {
		case "Name":
			output.WriteString("    public string Name() => PluginBridge.Host.WorldName(Invocation, Id) ?? string.Empty;\n")
		case "Dimension":
			// C# does not permit a nested World.Dimension type and a World.Dimension
			// member. An extension preserves both exact plugin spellings.
		case "Range":
			output.WriteString("    public Cube.Range Range() => PluginBridge.Host.WorldRange(Invocation, Id);\n")
		case "HighestLightBlocker":
			fmt.Fprintf(&output, "    public int HighestLightBlocker(int %s, int %s) =>\n        PluginBridge.Host.WorldHighestLightBlocker(Invocation, Id, %s, %s);\n", method.Parameters[0], method.Parameters[1], method.Parameters[0], method.Parameters[1])
		case "Time":
			output.WriteString("    public int Time() => PluginBridge.Host.WorldTime(Invocation, Id);\n")
		case "SetTime":
			parameter := csharpIdentifier(method.Parameters[0])
			fmt.Fprintf(&output, "    public void SetTime(int %s) => PluginBridge.Host.SetWorldTime(Invocation, Id, %s);\n", parameter, parameter)
		case "StopTime":
			output.WriteString("    public void StopTime() => PluginBridge.Host.SetWorldTimeCycle(Invocation, Id, false);\n")
		case "StartTime":
			output.WriteString("    public void StartTime() => PluginBridge.Host.SetWorldTimeCycle(Invocation, Id, true);\n")
		case "TimeCycle":
			output.WriteString("    public bool TimeCycle() => PluginBridge.Host.WorldTimeCycle(Invocation, Id);\n")
		case "Spawn":
			output.WriteString("    public Cube.Pos Spawn() => PluginBridge.Host.WorldSpawn(Invocation, Id);\n")
		case "SetSpawn":
			fmt.Fprintf(&output, "    public void SetSpawn(Cube.Pos %s) =>\n        PluginBridge.Host.SetWorldSpawn(Invocation, Id, %s);\n", method.Parameters[0], method.Parameters[0])
		case "PlayerSpawn":
			fmt.Fprintf(&output, "    public Cube.Pos PlayerSpawn(Guid %s) =>\n        PluginBridge.Host.WorldPlayerSpawn(Invocation, Id, %s);\n", method.Parameters[0], method.Parameters[0])
		case "SetPlayerSpawn":
			fmt.Fprintf(&output, "    public void SetPlayerSpawn(Guid %s, Cube.Pos %s) =>\n        PluginBridge.Host.SetWorldPlayerSpawn(Invocation, Id, %s, %s);\n", method.Parameters[0], method.Parameters[1], method.Parameters[0], method.Parameters[1])
		case "SetRequiredSleepDuration":
			parameter := csharpIdentifier(method.Parameters[0])
			fmt.Fprintf(&output, "    public void SetRequiredSleepDuration(TimeSpan %s) =>\n        PluginBridge.Host.SetWorldRequiredSleepDuration(Invocation, Id, %s);\n", parameter, parameter)
		case "DefaultGameMode":
			output.WriteString("    public GameMode DefaultGameMode() => PluginBridge.Host.WorldDefaultGameMode(Invocation, Id);\n")
		case "SetTickRange":
			fmt.Fprintf(&output, "    public void SetTickRange(int %s) => PluginBridge.Host.SetWorldTickRange(Invocation, Id, %s);\n", method.Parameters[0], method.Parameters[0])
		case "SetDefaultGameMode":
			fmt.Fprintf(&output, "    public void SetDefaultGameMode(GameMode %s) =>\n        PluginBridge.Host.SetWorldDefaultGameMode(Invocation, Id, %s);\n", method.Parameters[0], method.Parameters[0])
		case "Difficulty":
			// See Dimension above.
		case "SetDifficulty":
			fmt.Fprintf(&output, "    public void SetDifficulty(Difficulty %s) =>\n        PluginBridge.Host.SetWorldDifficulty(Invocation, Id, %s);\n", method.Parameters[0], method.Parameters[0])
		case "Save":
			output.WriteString("    public void Save() => PluginBridge.Host.SaveWorld(Invocation, Id);\n")
		case "Close":
			output.WriteString("    public void Close() => PluginBridge.Host.CloseWorld(Invocation, Id);\n")
		default:
			panic("unsupported world lifecycle method: " + method.Name)
		}
	}
	output.WriteString("}\n\n")
	output.WriteString("public static class WorldStateExtensions\n{\n")
	output.WriteString("    public static World.Dimension Dimension(this World world)\n    {\n        ArgumentNullException.ThrowIfNull(world);\n        return PluginBridge.Host.WorldDimension(world.Invocation, world.Id);\n    }\n\n")
	output.WriteString("    public static World.Difficulty Difficulty(this World world)\n    {\n        ArgumentNullException.ThrowIfNull(world);\n        return PluginBridge.Host.WorldDifficulty(world.Invocation, world.Id);\n    }\n")
	output.WriteString("}\n")
	return output.Bytes()
}
