package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

var selectedWorldConfigFields = map[string]string{
	"Dim":                 "Dimension",
	"Provider":            "Provider",
	"ReadOnly":            "bool",
	"SaveInterval":        "time.Duration",
	"ChunkUnloadInterval": "time.Duration",
	"RandomTickSpeed":     "int",
}

func inspectWorldConfig(configPath, worldPath, dimensionPath, providerPath, mcdbPath string) error {
	config, err := parser.ParseFile(token.NewFileSet(), configPath, nil, 0)
	if err != nil {
		return err
	}
	var configType *ast.StructType
	var configNew bool
	for _, declaration := range config.Decls {
		switch value := declaration.(type) {
		case *ast.GenDecl:
			for _, raw := range value.Specs {
				if spec, ok := raw.(*ast.TypeSpec); ok && spec.Name.Name == "Config" {
					configType, _ = spec.Type.(*ast.StructType)
				}
			}
		case *ast.FuncDecl:
			configNew = configNew || value.Name.Name == "New" && valueReceiver(value, "Config") &&
				goFunctionSignature(value) == (goSignature{Results: "*World"})
		}
	}
	if configType == nil || !configNew {
		return fmt.Errorf("Dragonfly world.Config or Config.New changed")
	}
	found := map[string]string{}
	for _, field := range configType.Fields.List {
		for _, name := range field.Names {
			found[name.Name] = formatGoExpression(field.Type)
		}
	}
	for name, want := range selectedWorldConfigFields {
		if found[name] != want {
			return fmt.Errorf("Dragonfly world.Config.%s changed from %s to %s", name, want, found[name])
		}
	}
	if err := inspectPackageFunction(worldPath, "New", goSignature{Results: "*World"}); err != nil {
		return err
	}
	if err := inspectWorldDimension(dimensionPath); err != nil {
		return err
	}
	if err := inspectWorldNamedTypes(providerPath, "Provider", nil); err != nil {
		return err
	}
	return inspectMethod(mcdbPath, "Config", "Open", goSignature{Parameters: "string", Results: "*DB, error"})
}

func inspectWorldDimension(path string) error {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return err
	}
	want := map[string]string{
		"Range":              "func() cube.Range",
		"WaterEvaporates":    "func() bool",
		"LavaSpreadDuration": "func() time.Duration",
		"WeatherCycle":       "func() bool",
		"TimeCycle":          "func() bool",
	}
	var found map[string]string
	variables := map[string]bool{}
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, raw := range general.Specs {
			switch spec := raw.(type) {
			case *ast.TypeSpec:
				if spec.Name.Name != "Dimension" {
					continue
				}
				iface, ok := spec.Type.(*ast.InterfaceType)
				if !ok {
					return fmt.Errorf("Dragonfly world.Dimension is no longer an interface")
				}
				found = map[string]string{}
				for _, field := range iface.Methods.List {
					if len(field.Names) == 1 {
						found[field.Names[0].Name] = formatGoExpression(field.Type)
					}
				}
			case *ast.ValueSpec:
				for _, name := range spec.Names {
					variables[name.Name] = true
				}
			}
		}
	}
	if found == nil {
		return fmt.Errorf("Dragonfly world.Dimension is missing")
	}
	for name, signature := range want {
		if found[name] != signature {
			return fmt.Errorf("Dragonfly world.Dimension.%s changed from %s to %s", name, signature, found[name])
		}
	}
	for _, name := range []string{"Overworld", "Nether", "End"} {
		if !variables[name] {
			return fmt.Errorf("Dragonfly world.%s is missing", name)
		}
	}
	return nil
}

func inspectPackageFunction(path, name string, signature goSignature) error {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return err
	}
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv == nil && function.Name.Name == name {
			if goFunctionSignature(function) != signature {
				return fmt.Errorf("Dragonfly %s signature changed", name)
			}
			return nil
		}
	}
	return fmt.Errorf("Dragonfly has no %s function", name)
}

func inspectWorldNamedTypes(path, typeName string, variables []string) error {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return err
	}
	foundType := false
	foundVariables := map[string]bool{}
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, raw := range general.Specs {
			switch spec := raw.(type) {
			case *ast.TypeSpec:
				foundType = foundType || spec.Name.Name == typeName
			case *ast.ValueSpec:
				for _, name := range spec.Names {
					foundVariables[name.Name] = true
				}
			}
		}
	}
	if !foundType {
		return fmt.Errorf("Dragonfly world.%s is missing", typeName)
	}
	for _, name := range variables {
		if !foundVariables[name] {
			return fmt.Errorf("Dragonfly world.%s is missing", name)
		}
	}
	return nil
}

func inspectMethod(path, receiver, name string, signature goSignature) error {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return err
	}
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == name && valueReceiver(function, receiver) {
			if goFunctionSignature(function) != signature {
				return fmt.Errorf("Dragonfly %s.%s signature changed", receiver, name)
			}
			return nil
		}
	}
	return fmt.Errorf("Dragonfly %s.%s is missing", receiver, name)
}

func generateWorldConfig() []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly world.Config, Dimension, Provider, and mcdb.Config Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing System;\n\nnamespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class World\n{\n")
	output.WriteString("    public interface Dimension\n    {\n")
	output.WriteString("        Cube.Range Range();\n")
	output.WriteString("        bool WaterEvaporates();\n")
	output.WriteString("        TimeSpan LavaSpreadDuration();\n")
	output.WriteString("        bool WeatherCycle();\n")
	output.WriteString("        bool TimeCycle();\n")
	output.WriteString("    }\n\n")
	output.WriteString("    internal sealed record BuiltinDimension(uint Id, Cube.Range BuildRange, bool EvaporatesWater, TimeSpan LavaDuration, bool HasWeatherCycle, bool HasTimeCycle) : Dimension\n    {\n")
	output.WriteString("        public Cube.Range Range() => BuildRange;\n")
	output.WriteString("        public bool WaterEvaporates() => EvaporatesWater;\n")
	output.WriteString("        public TimeSpan LavaSpreadDuration() => LavaDuration;\n")
	output.WriteString("        public bool WeatherCycle() => HasWeatherCycle;\n")
	output.WriteString("        public bool TimeCycle() => HasTimeCycle;\n")
	output.WriteString("    }\n\n")
	output.WriteString("    internal sealed record TransportDimension(Cube.Range BuildRange, bool EvaporatesWater, TimeSpan LavaDuration, bool HasWeatherCycle, bool HasTimeCycle) : Dimension\n    {\n")
	output.WriteString("        public Cube.Range Range() => BuildRange;\n")
	output.WriteString("        public bool WaterEvaporates() => EvaporatesWater;\n")
	output.WriteString("        public TimeSpan LavaSpreadDuration() => LavaDuration;\n")
	output.WriteString("        public bool WeatherCycle() => HasWeatherCycle;\n")
	output.WriteString("        public bool TimeCycle() => HasTimeCycle;\n")
	output.WriteString("    }\n\n")
	output.WriteString("    public static Dimension Overworld { get; } = new BuiltinDimension(0, new Cube.Range(-64, 319), false, TimeSpan.FromMilliseconds(1500), true, true);\n")
	output.WriteString("    public static Dimension Nether { get; } = new BuiltinDimension(1, new Cube.Range(0, 127), true, TimeSpan.FromMilliseconds(250), false, false);\n")
	output.WriteString("    public static Dimension End { get; } = new BuiltinDimension(2, new Cube.Range(0, 255), false, TimeSpan.FromMilliseconds(1500), false, false);\n\n")
	output.WriteString("    public abstract record Provider\n    {\n        private protected Provider() { }\n    }\n\n")
	output.WriteString("    public sealed record NopProvider : Provider;\n\n")
	output.WriteString("    public sealed record Config\n    {\n")
	output.WriteString("        public Dimension? Dim { get; init; }\n")
	output.WriteString("        public Provider? Provider { get; init; }\n")
	output.WriteString("        public bool ReadOnly { get; init; }\n")
	output.WriteString("        public TimeSpan SaveInterval { get; init; }\n")
	output.WriteString("        public TimeSpan ChunkUnloadInterval { get; init; }\n")
	output.WriteString("        public int RandomTickSpeed { get; init; }\n\n")
	output.WriteString("        public World New() => PluginBridge.Host.NewWorld(this);\n")
	output.WriteString("    }\n\n")
	output.WriteString("    public static World New() => new Config().New();\n")
	output.WriteString("}\n\n")
	output.WriteString("public static class MCDB\n{\n")
	output.WriteString("    public sealed record Config\n    {\n")
	output.WriteString("        public DB Open(string dir) => new(dir);\n")
	output.WriteString("    }\n\n")
	output.WriteString("    public sealed record DB : World.Provider\n    {\n")
	output.WriteString("        internal DB(string dir) => Directory = dir;\n")
	output.WriteString("        internal string Directory { get; }\n")
	output.WriteString("    }\n")
	output.WriteString("}\n")
	return output.Bytes()
}
