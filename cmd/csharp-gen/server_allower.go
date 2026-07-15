package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strconv"
)

var loginDataTypeNames = []string{
	"IdentityData",
	"ClientData",
	"PersonaPiece",
	"PersonaPieceTintColour",
	"SkinAnimation",
}

type serverAllowerSpec struct {
	Method commandMethod
}

type loginDataSpec struct {
	Types []loginDataTypeSpec
}

type loginDataTypeSpec struct {
	Name   string
	Fields []parameter
}

type deviceOSSpec struct {
	Values []deviceOSValue
}

type deviceOSValue struct {
	Name  string
	Value int
}

func inspectServerAllower(path string) (serverAllowerSpec, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return serverAllowerSpec{}, err
	}
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, raw := range general.Specs {
			typeSpec, ok := raw.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "Allower" {
				continue
			}
			allower, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok || allower.Methods == nil || len(allower.Methods.List) != 1 {
				return serverAllowerSpec{}, fmt.Errorf("Dragonfly server.Allower changed")
			}
			field := allower.Methods.List[0]
			if len(field.Names) != 1 || field.Names[0].Name != "Allow" {
				return serverAllowerSpec{}, fmt.Errorf("Dragonfly server.Allower.Allow changed")
			}
			function, ok := field.Type.(*ast.FuncType)
			if !ok || !serverAllowParameters(function.Params) || !serverAllowResults(function.Results) {
				return serverAllowerSpec{}, fmt.Errorf("Dragonfly server.Allower.Allow signature changed")
			}
			return serverAllowerSpec{Method: commandMethod{
				Name: "Allow",
				Parameters: []parameter{
					{Name: "addr", Type: "Net.Addr"},
					{Name: "d", Type: "Login.IdentityData"},
					{Name: "c", Type: "Login.ClientData"},
				},
				ReturnType: "(string Message, bool Allowed)",
			}}, nil
		}
	}
	return serverAllowerSpec{}, fmt.Errorf("Dragonfly server.Allower is missing")
}

func serverAllowParameters(fields *ast.FieldList) bool {
	if fields == nil || len(fields.List) != 3 {
		return false
	}
	want := []struct{ name, pkg, typ string }{
		{name: "addr", pkg: "net", typ: "Addr"},
		{name: "d", pkg: "login", typ: "IdentityData"},
		{name: "c", pkg: "login", typ: "ClientData"},
	}
	for index, field := range fields.List {
		if len(field.Names) != 1 || field.Names[0].Name != want[index].name {
			return false
		}
		selector, ok := field.Type.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != want[index].typ {
			return false
		}
		pkg, ok := selector.X.(*ast.Ident)
		if !ok || pkg.Name != want[index].pkg {
			return false
		}
	}
	return true
}

func serverAllowResults(fields *ast.FieldList) bool {
	return fields != nil && len(fields.List) == 2 &&
		len(fields.List[0].Names) == 0 && loginDataExpression(fields.List[0].Type) == "string" &&
		len(fields.List[1].Names) == 0 && loginDataExpression(fields.List[1].Type) == "bool"
}

func inspectLoginData(path string) (loginDataSpec, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return loginDataSpec{}, err
	}
	found := map[string]*ast.StructType{}
	foundDeviceID := false
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, raw := range general.Specs {
			typeSpec, ok := raw.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if structure, ok := typeSpec.Type.(*ast.StructType); ok {
				found[typeSpec.Name.Name] = structure
			} else if identifier, ok := typeSpec.Type.(*ast.Ident); ok {
				if typeSpec.Name.Name == "DeviceID" {
					foundDeviceID = identifier.Name == "string"
				}
			}
		}
	}
	if !foundDeviceID {
		return loginDataSpec{}, fmt.Errorf("gophertunnel login.DeviceID surface changed")
	}
	spec := loginDataSpec{Types: make([]loginDataTypeSpec, 0, len(loginDataTypeNames))}
	for _, name := range loginDataTypeNames {
		structure, ok := found[name]
		if !ok {
			return loginDataSpec{}, fmt.Errorf("gophertunnel login.%s is missing or no longer a struct", name)
		}
		fields, err := inspectLoginDataFields(name, structure)
		if err != nil {
			return loginDataSpec{}, err
		}
		spec.Types = append(spec.Types, loginDataTypeSpec{Name: name, Fields: fields})
	}
	return spec, nil
}

func inspectLoginDataFields(owner string, structure *ast.StructType) ([]parameter, error) {
	var fields []parameter
	for _, field := range structure.Fields.List {
		if len(field.Names) == 0 {
			return nil, fmt.Errorf("gophertunnel login.%s has an unsupported embedded field", owner)
		}
		fieldType, err := translateLoginDataType(field.Type)
		if err != nil {
			return nil, fmt.Errorf("gophertunnel login.%s.%s: %w", owner, field.Names[0].Name, err)
		}
		for _, name := range field.Names {
			fields = append(fields, parameter{Name: name.Name, Type: fieldType})
		}
	}
	return fields, nil
}

func translateLoginDataType(expression ast.Expr) (string, error) {
	switch value := expression.(type) {
	case *ast.Ident:
		switch value.Name {
		case "string":
			return "string", nil
		case "bool":
			return "bool", nil
		case "int":
			return "int", nil
		case "int64":
			return "long", nil
		case "float64":
			return "double", nil
		case "DeviceID":
			return "DeviceID", nil
		case "SkinAnimation", "PersonaPiece", "PersonaPieceTintColour":
			return value.Name, nil
		}
	case *ast.SelectorExpr:
		pkg, ok := value.X.(*ast.Ident)
		if ok && pkg.Name == "protocol" && value.Sel.Name == "DeviceOS" {
			return "Protocol.DeviceOS", nil
		}
	case *ast.StarExpr:
		if identifier, ok := value.X.(*ast.Ident); ok && identifier.Name == "bool" {
			return "bool?", nil
		}
	case *ast.ArrayType:
		element, err := translateLoginDataType(value.Elt)
		if err != nil {
			return "", err
		}
		if value.Len != nil {
			length, ok := value.Len.(*ast.BasicLit)
			if !ok || length.Kind != token.INT || length.Value != "4" {
				return "", fmt.Errorf("unsupported fixed array length")
			}
		}
		return element + "[]", nil
	}
	return "", fmt.Errorf("unsupported type %s", loginDataExpression(expression))
}

func loginDataExpression(expression ast.Expr) string {
	var output bytes.Buffer
	if err := formatNode(&output, expression); err != nil {
		return fmt.Sprintf("%T", expression)
	}
	return output.String()
}

func formatNode(output *bytes.Buffer, node ast.Node) error {
	return printerFprint(output, token.NewFileSet(), node)
}

// printerFprint is a variable so the small AST formatter stays testable without
// making parser errors look like unsupported source types.
var printerFprint = func(output *bytes.Buffer, files *token.FileSet, node any) error {
	return printer.Fprint(output, files, node)
}

func inspectDeviceOS(path string) (deviceOSSpec, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return deviceOSSpec{}, err
	}
	foundType := false
	var values []deviceOSValue
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		if general.Tok == token.TYPE {
			for _, raw := range general.Specs {
				typeSpec, ok := raw.(*ast.TypeSpec)
				if ok && typeSpec.Name.Name == "DeviceOS" {
					identifier, ok := typeSpec.Type.(*ast.Ident)
					foundType = ok && identifier.Name == "int"
				}
			}
			continue
		}
		if general.Tok != token.CONST {
			continue
		}
		block, ok := inspectDeviceOSConstants(general)
		if ok {
			values = block
		}
	}
	if !foundType || len(values) == 0 {
		return deviceOSSpec{}, fmt.Errorf("gophertunnel protocol.DeviceOS changed")
	}
	return deviceOSSpec{Values: values}, nil
}

func inspectDeviceOSConstants(general *ast.GenDecl) ([]deviceOSValue, bool) {
	var values []deviceOSValue
	for index, raw := range general.Specs {
		valueSpec, ok := raw.(*ast.ValueSpec)
		if !ok || len(valueSpec.Names) != 1 {
			return nil, false
		}
		name := valueSpec.Names[0].Name
		if index == 0 {
			if name != "DeviceAndroid" || !deviceOSIotaPlusOne(valueSpec) {
				return nil, false
			}
		} else if len(valueSpec.Values) != 0 || valueSpec.Type != nil {
			return nil, false
		}
		if len(name) < len("Device") || name[:len("Device")] != "Device" {
			return nil, false
		}
		values = append(values, deviceOSValue{Name: name[len("Device"):], Value: index + 1})
	}
	return values, len(values) != 0
}

func deviceOSIotaPlusOne(spec *ast.ValueSpec) bool {
	if len(spec.Values) != 1 {
		return false
	}
	selector, ok := spec.Type.(*ast.Ident)
	if !ok || selector.Name != "DeviceOS" {
		return false
	}
	binary, ok := spec.Values[0].(*ast.BinaryExpr)
	if !ok || binary.Op != token.ADD {
		return false
	}
	iota, ok := binary.X.(*ast.Ident)
	if !ok || iota.Name != "iota" {
		return false
	}
	one, ok := binary.Y.(*ast.BasicLit)
	return ok && one.Kind == token.INT && one.Value == strconv.Itoa(1)
}

func generateServerAllower(allower serverAllowerSpec, login loginDataSpec, deviceOS deviceOSSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/allower.go and gophertunnel login AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing System.Text.Json;\nusing System.Text.Json.Serialization;\n\nnamespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class Server\n{\n")
	output.WriteString("    public interface Allower\n    {\n")
	fmt.Fprintf(&output, "        %s %s(%s);\n", allower.Method.ReturnType, allower.Method.Name, formatParameters(allower.Method.Parameters))
	output.WriteString("    }\n}\n\n")
	output.WriteString("public abstract partial class Plugin : Server.Allower\n{\n")
	fmt.Fprintf(&output, "    public virtual %s %s(%s) => (string.Empty, true);\n", allower.Method.ReturnType, allower.Method.Name, formatParameters(allower.Method.Parameters))
	output.WriteString("}\n\n")
	output.WriteString("public static partial class Login\n{\n")
	output.WriteString(`    [JsonConverter(typeof(DeviceIDJsonConverter))]
    public readonly record struct DeviceID
    {
        private readonly string? _value;

        public DeviceID(string value) => _value = value ?? string.Empty;
        public string Value => _value ?? string.Empty;

        public override string ToString() => Value;
        public static implicit operator DeviceID(string value) => new(value);
        public static implicit operator string(DeviceID value) => value.Value;
    }

    internal sealed class DeviceIDJsonConverter : JsonConverter<DeviceID>
    {
        public override DeviceID Read(ref Utf8JsonReader reader, Type type, JsonSerializerOptions options) =>
            new(reader.GetString() ?? string.Empty);
        public override void Write(Utf8JsonWriter writer, DeviceID value, JsonSerializerOptions options) =>
            writer.WriteStringValue(value.Value);
    }

`)
	for typeIndex, dataType := range login.Types {
		fmt.Fprintf(&output, "    public sealed class %s\n    {\n", dataType.Name)
		for _, field := range dataType.Fields {
			fmt.Fprintf(&output, "        public %s %s { get; init; }%s\n", field.Type, field.Name, loginDataInitializer(field.Type))
		}
		output.WriteString("    }\n")
		if typeIndex != len(login.Types)-1 {
			output.WriteByte('\n')
		}
	}
	output.WriteString("}\n\n")
	output.WriteString("public static partial class Protocol\n{\n    public enum DeviceOS\n    {\n")
	for _, value := range deviceOS.Values {
		fmt.Fprintf(&output, "        %s = %d,\n", value.Name, value.Value)
	}
	output.WriteString("    }\n}\n")
	return output.Bytes()
}

func loginDataInitializer(csharpType string) string {
	if len(csharpType) >= 2 && csharpType[len(csharpType)-2:] == "[]" {
		return " = [];"
	}
	if csharpType == "string" {
		return " = string.Empty;"
	}
	return ""
}
