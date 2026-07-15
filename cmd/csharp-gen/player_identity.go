package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

var selectedPlayerIdentityMethods = []string{
	"Name",
	"UUID",
	"XUID",
	"DeviceID",
	"DeviceModel",
	"SelfSignedID",
	"Locale",
	"Addr",
}

func inspectPlayerIdentityMethods(path string) ([]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	found := map[string]*ast.FuncDecl{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !playerMethod(function) {
			continue
		}
		found[function.Name.Name] = function
	}
	want := map[string]goSignature{
		"Name":         {Results: "string"},
		"UUID":         {Results: "uuid.UUID"},
		"XUID":         {Results: "string"},
		"DeviceID":     {Results: "string"},
		"DeviceModel":  {Results: "string"},
		"SelfSignedID": {Results: "string"},
		"Locale":       {Results: "language.Tag"},
		"Addr":         {Results: "net.Addr"},
	}
	for _, name := range selectedPlayerIdentityMethods {
		function := found[name]
		if function == nil {
			return nil, fmt.Errorf("Dragonfly player.Player has no %s method", name)
		}
		if got := goFunctionSignature(function); got != want[name] {
			return nil, fmt.Errorf("Dragonfly player.Player.%s signature changed: %+v", name, got)
		}
	}
	return append([]string(nil), selectedPlayerIdentityMethods...), nil
}

func generatePlayerIdentityMethods(methods []string) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing System;\nusing Dragonfly.Native;\n\nnamespace Dragonfly;\n\n")
	output.WriteString(`public static class Language
{
    public readonly record struct Tag
    {
        private readonly string? _value;
        public Tag(string value) => _value = value ?? "und";
        public string String() => _value ?? "und";
        public override string ToString() => String();
    }
}

public sealed partial class Player
{
`)
	for _, method := range methods {
		switch method {
		case "Name":
			output.WriteString("    public string Name() => PlayerName;\n")
		case "UUID":
			output.WriteString("    public Guid UUID() => PluginBridge.Host.PlayerUUID(Id);\n")
		case "XUID":
			output.WriteString("    public string XUID() => PluginBridge.Host.PlayerXUID(_invocation, Id);\n")
		case "DeviceID", "DeviceModel", "SelfSignedID":
			fmt.Fprintf(&output, "    public string %s() => PluginBridge.Host.PlayerString(_invocation, Id, Abi.PlayerString%s);\n", method, method)
		case "Locale":
			output.WriteString("    public Language.Tag Locale() => new(PluginBridge.Host.PlayerString(_invocation, Id, Abi.PlayerStringLocale));\n")
		case "Addr":
			output.WriteString(`    public Net.Addr? Addr()
    {
        if (!PluginBridge.Host.TryPlayerString(_invocation, Id, Abi.PlayerStringAddrNetwork, out var network) ||
            !PluginBridge.Host.TryPlayerString(_invocation, Id, Abi.PlayerStringAddrString, out var address)) return null;
        return new Net.AddrSnapshot(network, address);
    }
`)
		default:
			panic("unsupported player identity method: " + method)
		}
	}
	output.WriteString("}\n")
	return output.Bytes()
}
