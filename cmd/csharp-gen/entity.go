package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
)

type entityMethod struct {
	Name       string
	ReturnType string
}

type entityConstructionSpec struct {
	EntityTypeMethods   []commandMethod
	EntityConfigMethods []commandMethod
	TickerMethods       []commandMethod
	EntityDataFields    []parameter
	SpawnFields         []parameter
	SpawnNew            commandMethod
	NewEntity           commandMethod
}

var selectedEntityHandleMethods = []string{"Entity", "UUID", "Closed", "Close"}

func inspectWorldEntityConstruction(path string) (entityConstructionSpec, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return entityConstructionSpec{}, err
	}
	types := map[string]*ast.TypeSpec{}
	var spawnNew, newEntity *ast.FuncDecl
	for _, declaration := range file.Decls {
		switch value := declaration.(type) {
		case *ast.GenDecl:
			for _, raw := range value.Specs {
				if spec, ok := raw.(*ast.TypeSpec); ok {
					types[spec.Name.Name] = spec
				}
			}
		case *ast.FuncDecl:
			if value.Name.Name == "New" && valueReceiver(value, "EntitySpawnOpts") {
				spawnNew = value
			}
			if value.Name.Name == "NewEntity" && value.Recv == nil {
				newEntity = value
			}
		}
	}

	entityTypeMethods, err := inspectEntityInterface(types, "EntityType")
	if err != nil {
		return entityConstructionSpec{}, err
	}
	entityConfigMethods, err := inspectEntityInterface(types, "EntityConfig")
	if err != nil {
		return entityConstructionSpec{}, err
	}
	tickerMethods, err := inspectTickerEntityInterface(types)
	if err != nil {
		return entityConstructionSpec{}, err
	}
	entityDataFields, err := inspectEntityStruct(types, "EntityData")
	if err != nil {
		return entityConstructionSpec{}, err
	}
	spawnFields, err := inspectEntityStruct(types, "EntitySpawnOpts")
	if err != nil {
		return entityConstructionSpec{}, err
	}
	spawnNewMethod, err := inspectEntityFunction(spawnNew, "EntitySpawnOpts.New")
	if err != nil {
		return entityConstructionSpec{}, err
	}
	newEntityMethod, err := inspectEntityFunction(newEntity, "NewEntity")
	if err != nil {
		return entityConstructionSpec{}, err
	}

	wantType := []commandMethod{
		{Name: "Open", Parameters: []parameter{{Name: "tx", Type: "Tx"}, {Name: "handle", Type: "EntityHandle"}, {Name: "data", Type: "EntityData"}}, ReturnType: "Entity"},
		{Name: "EncodeEntity", ReturnType: "string"},
		{Name: "BBox", Parameters: []parameter{{Name: "e", Type: "Entity"}}, ReturnType: "Cube.BBox"},
		{Name: "DecodeNBT", Parameters: []parameter{{Name: "m", Type: "Dictionary<string, object?>"}, {Name: "data", Type: "EntityData"}}, ReturnType: "void"},
		{Name: "EncodeNBT", Parameters: []parameter{{Name: "data", Type: "EntityData"}}, ReturnType: "Dictionary<string, object?>"},
	}
	wantConfig := []commandMethod{{Name: "Apply", Parameters: []parameter{{Name: "data", Type: "EntityData"}}, ReturnType: "void"}}
	wantTicker := []commandMethod{{Name: "Tick", Parameters: []parameter{{Name: "tx", Type: "Tx"}, {Name: "current", Type: "long"}}, ReturnType: "void"}}
	wantData := []parameter{
		{Name: "Pos", Type: "Vector3"}, {Name: "Vel", Type: "Vector3"}, {Name: "Rot", Type: "Rotation"},
		{Name: "Name", Type: "string"}, {Name: "FireDuration", Type: "TimeSpan"}, {Name: "Age", Type: "TimeSpan"},
		{Name: "Data", Type: "object?"},
	}
	wantSpawn := []parameter{
		{Name: "Position", Type: "Vector3"}, {Name: "Rotation", Type: "Rotation"}, {Name: "Velocity", Type: "Vector3"},
		{Name: "ID", Type: "Guid"}, {Name: "NameTag", Type: "string"},
	}
	wantNew := commandMethod{Name: "New", Parameters: []parameter{{Name: "t", Type: "EntityType"}, {Name: "conf", Type: "EntityConfig"}}, ReturnType: "EntityHandle"}
	wantPackageNew := wantNew
	wantPackageNew.Name = "NewEntity"
	for name, gotWant := range map[string][2]any{
		"EntityType":          {entityTypeMethods, wantType},
		"EntityConfig":        {entityConfigMethods, wantConfig},
		"TickerEntity":        {tickerMethods, wantTicker},
		"EntityData":          {entityDataFields, wantData},
		"EntitySpawnOpts":     {spawnFields, wantSpawn},
		"EntitySpawnOpts.New": {spawnNewMethod, wantNew},
		"NewEntity":           {newEntityMethod, wantPackageNew},
	} {
		if !reflect.DeepEqual(gotWant[0], gotWant[1]) {
			return entityConstructionSpec{}, fmt.Errorf("Dragonfly world.%s signature changed: got %+v", name, gotWant[0])
		}
	}
	return entityConstructionSpec{
		EntityTypeMethods: entityTypeMethods, EntityConfigMethods: entityConfigMethods,
		TickerMethods: tickerMethods,
		EntityDataFields: entityDataFields, SpawnFields: spawnFields,
		SpawnNew: spawnNewMethod, NewEntity: newEntityMethod,
	}, nil
}

func inspectTickerEntityInterface(types map[string]*ast.TypeSpec) ([]commandMethod, error) {
	spec, ok := types["TickerEntity"]
	if !ok {
		return nil, fmt.Errorf("Dragonfly world.TickerEntity not found")
	}
	definition, ok := spec.Type.(*ast.InterfaceType)
	if !ok || len(definition.Methods.List) != 2 {
		return nil, fmt.Errorf("Dragonfly world.TickerEntity signature changed")
	}
	embedded := definition.Methods.List[0]
	identifier, ok := embedded.Type.(*ast.Ident)
	if len(embedded.Names) != 0 || !ok || identifier.Name != "Entity" {
		return nil, fmt.Errorf("Dragonfly world.TickerEntity must embed Entity")
	}
	tick := definition.Methods.List[1]
	if len(tick.Names) != 1 || tick.Names[0].Name != "Tick" {
		return nil, fmt.Errorf("Dragonfly world.TickerEntity.Tick not found")
	}
	function, ok := tick.Type.(*ast.FuncType)
	if !ok {
		return nil, fmt.Errorf("Dragonfly world.TickerEntity.Tick is not a method")
	}
	method, err := translateEntityFunction("Tick", function)
	if err != nil {
		return nil, fmt.Errorf("Dragonfly world.TickerEntity.Tick: %w", err)
	}
	return []commandMethod{method}, nil
}

func inspectEntityInterface(types map[string]*ast.TypeSpec, name string) ([]commandMethod, error) {
	spec, ok := types[name]
	if !ok {
		return nil, fmt.Errorf("Dragonfly world.%s not found", name)
	}
	definition, ok := spec.Type.(*ast.InterfaceType)
	if !ok {
		return nil, fmt.Errorf("Dragonfly world.%s is not an interface", name)
	}
	methods := make([]commandMethod, 0, len(definition.Methods.List))
	for _, field := range definition.Methods.List {
		if len(field.Names) != 1 {
			return nil, fmt.Errorf("Dragonfly world.%s contains an embedded or multiply named method", name)
		}
		function, ok := field.Type.(*ast.FuncType)
		if !ok {
			return nil, fmt.Errorf("Dragonfly world.%s.%s is not a method", name, field.Names[0].Name)
		}
		method, err := translateEntityFunction(field.Names[0].Name, function)
		if err != nil {
			return nil, fmt.Errorf("Dragonfly world.%s.%s: %w", name, field.Names[0].Name, err)
		}
		methods = append(methods, method)
	}
	return methods, nil
}

func inspectEntityStruct(types map[string]*ast.TypeSpec, name string) ([]parameter, error) {
	spec, ok := types[name]
	if !ok {
		return nil, fmt.Errorf("Dragonfly world.%s not found", name)
	}
	definition, ok := spec.Type.(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("Dragonfly world.%s is not a struct", name)
	}
	fields, err := translateEntityFields(definition.Fields)
	if err != nil {
		return nil, fmt.Errorf("Dragonfly world.%s: %w", name, err)
	}
	return fields, nil
}

func inspectEntityFunction(function *ast.FuncDecl, name string) (commandMethod, error) {
	if function == nil {
		return commandMethod{}, fmt.Errorf("Dragonfly world.%s not found", name)
	}
	methodName := function.Name.Name
	method, err := translateEntityFunction(methodName, function.Type)
	if err != nil {
		return commandMethod{}, fmt.Errorf("Dragonfly world.%s: %w", name, err)
	}
	return method, nil
}

func translateEntityFunction(name string, function *ast.FuncType) (commandMethod, error) {
	parameters, err := translateEntityFields(function.Params)
	if err != nil {
		return commandMethod{}, err
	}
	method := commandMethod{Name: name, Parameters: parameters, ReturnType: "void"}
	if function.Results == nil {
		return method, nil
	}
	if len(function.Results.List) != 1 || len(function.Results.List[0].Names) > 1 {
		return commandMethod{}, fmt.Errorf("multiple results are unsupported")
	}
	result, ok := entityConstructionCSharpType(function.Results.List[0].Type)
	if !ok {
		return commandMethod{}, fmt.Errorf("unsupported result %s", formatGoExpression(function.Results.List[0].Type))
	}
	method.ReturnType = result
	return method, nil
}

func translateEntityFields(fields *ast.FieldList) ([]parameter, error) {
	if fields == nil {
		return nil, nil
	}
	var translated []parameter
	for _, field := range fields.List {
		typeName, ok := entityConstructionCSharpType(field.Type)
		if !ok {
			return nil, fmt.Errorf("unsupported type %s", formatGoExpression(field.Type))
		}
		if len(field.Names) == 0 {
			return nil, fmt.Errorf("unnamed fields are unsupported")
		}
		for _, name := range field.Names {
			translated = append(translated, parameter{Name: name.Name, Type: typeName})
		}
	}
	return translated, nil
}

func entityConstructionCSharpType(expression ast.Expr) (string, bool) {
	goType := formatGoExpression(expression)
	typeName, ok := map[string]string{
		"*Tx":            "Tx",
		"*EntityHandle":  "EntityHandle",
		"*EntityData":    "EntityData",
		"Entity":         "Entity",
		"EntityType":     "EntityType",
		"EntityConfig":   "EntityConfig",
		"mgl64.Vec3":     "Vector3",
		"cube.Rotation":  "Rotation",
		"cube.BBox":      "Cube.BBox",
		"uuid.UUID":      "Guid",
		"time.Duration":  "TimeSpan",
		"string":         "string",
		"int64":          "long",
		"any":            "object?",
		"map[string]any": "Dictionary<string, object?>",
	}[goType]
	return typeName, ok
}

func inspectWorldEntity(path string) ([]entityMethod, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "Entity" {
				continue
			}
			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				return nil, fmt.Errorf("world.Entity is not an interface")
			}
			methods := make([]entityMethod, 0, 4)
			for _, field := range interfaceType.Methods.List {
				if len(field.Names) == 0 {
					selector, ok := field.Type.(*ast.SelectorExpr)
					if !ok {
						return nil, fmt.Errorf("world.Entity has unsupported embedded interface")
					}
					pkg, pkgOK := selector.X.(*ast.Ident)
					if !pkgOK || pkg.Name != "io" || selector.Sel.Name != "Closer" {
						return nil, fmt.Errorf("world.Entity has unsupported embedded interface")
					}
					methods = append(methods, entityMethod{Name: "Close", ReturnType: "void"})
					continue
				}
				if len(field.Names) != 1 {
					return nil, fmt.Errorf("world.Entity has multiply named method")
				}
				function, ok := field.Type.(*ast.FuncType)
				if !ok || function.Params.NumFields() != 0 || function.Results == nil || function.Results.NumFields() != 1 {
					return nil, fmt.Errorf("world.Entity.%s has unsupported signature", field.Names[0].Name)
				}
				name := field.Names[0].Name
				returnType, ok := entityReturnType(function.Results.List[0].Type)
				if !ok {
					return nil, fmt.Errorf("world.Entity.%s has unsupported return type", name)
				}
				methods = append(methods, entityMethod{Name: name, ReturnType: returnType})
			}
			if len(methods) != 4 {
				return nil, fmt.Errorf("world.Entity has %d methods, want 4", len(methods))
			}
			return methods, nil
		}
	}
	return nil, fmt.Errorf("world.Entity interface not found")
}

func inspectWorldEntityHandle(path string) ([]commandMethod, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	found := map[string]commandMethod{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !selectedEntityHandleMethod(function.Name.Name) || !pointerReceiver(function, "EntityHandle") {
			continue
		}
		definition, err := translateEntityHandleMethod(function)
		if err != nil {
			return nil, fmt.Errorf("world.EntityHandle.%s: %w", function.Name.Name, err)
		}
		found[function.Name.Name] = definition
	}
	methods := make([]commandMethod, 0, len(selectedEntityHandleMethods))
	for _, name := range selectedEntityHandleMethods {
		definition, ok := found[name]
		if !ok {
			return nil, fmt.Errorf("Dragonfly world.EntityHandle has no supported %s method", name)
		}
		methods = append(methods, definition)
	}
	return methods, nil
}

func selectedEntityHandleMethod(name string) bool {
	for _, selected := range selectedEntityHandleMethods {
		if name == selected {
			return true
		}
	}
	return false
}

func translateEntityHandleMethod(function *ast.FuncDecl) (commandMethod, error) {
	method := commandMethod{Name: function.Name.Name}
	switch function.Name.Name {
	case "Entity":
		if !singlePointerParameter(function.Type.Params, "tx", "Tx") || !entityAndBoolResults(function.Type.Results) {
			return method, fmt.Errorf("signature changed")
		}
		method.Parameters = []parameter{{Name: "tx", Type: "Tx"}}
		method.ReturnType = "(Entity? Entity, bool Ok)"
	case "UUID":
		if function.Type.Params.NumFields() != 0 || !singleSelectorResult(function.Type.Results, "uuid", "UUID") {
			return method, fmt.Errorf("signature changed")
		}
		method.ReturnType = "Guid"
	case "Closed":
		if function.Type.Params.NumFields() != 0 || !singleIdentifierResult(function.Type.Results, "bool") {
			return method, fmt.Errorf("signature changed")
		}
		method.ReturnType = "bool"
	case "Close":
		if function.Type.Params.NumFields() != 0 || !singleIdentifierResult(function.Type.Results, "error") {
			return method, fmt.Errorf("signature changed")
		}
		// Dragonfly's EntityHandle.Close always returns nil. Keep host errors private.
		method.ReturnType = "void"
	default:
		return method, fmt.Errorf("unsupported method")
	}
	return method, nil
}

func singlePointerParameter(fields *ast.FieldList, name, typeName string) bool {
	if fields == nil || len(fields.List) != 1 || len(fields.List[0].Names) != 1 || fields.List[0].Names[0].Name != name {
		return false
	}
	pointer, ok := fields.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	identifier, identifierOK := pointer.X.(*ast.Ident)
	return identifierOK && identifier.Name == typeName
}

func entityAndBoolResults(fields *ast.FieldList) bool {
	return fields != nil && len(fields.List) == 2 &&
		singleResultType(fields.List[0].Type, "Entity") && singleResultType(fields.List[1].Type, "bool")
}

func singleIdentifierResult(fields *ast.FieldList, name string) bool {
	return fields != nil && len(fields.List) == 1 && singleResultType(fields.List[0].Type, name)
}

func singleSelectorResult(fields *ast.FieldList, packageName, name string) bool {
	if fields == nil || len(fields.List) != 1 {
		return false
	}
	selector, ok := fields.List[0].Type.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, packageOK := selector.X.(*ast.Ident)
	return packageOK && pkg.Name == packageName && selector.Sel.Name == name
}

func singleResultType(expression ast.Expr, name string) bool {
	identifier, ok := expression.(*ast.Ident)
	return ok && identifier.Name == name
}

func entityReturnType(expression ast.Expr) (string, bool) {
	switch value := expression.(type) {
	case *ast.StarExpr:
		identifier, ok := value.X.(*ast.Ident)
		if !ok || identifier.Name != "EntityHandle" {
			return "", false
		}
		return "EntityHandle", true
	case *ast.SelectorExpr:
		pkg, ok := value.X.(*ast.Ident)
		if !ok {
			return "", false
		}
		translated, ok := map[string]string{
			"mgl64.Vec3":    "Vector3",
			"cube.Rotation": "Rotation",
		}[pkg.Name+"."+value.Sel.Name]
		return translated, ok
	default:
		return "", false
	}
}

func generateWorldEntity(methods []entityMethod, handleMethods []commandMethod, construction entityConstructionSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/world/entity.go Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing System;\nusing System.Collections.Generic;\n\nnamespace Dragonfly;\n\npublic sealed partial class World\n{\n")
	output.WriteString("    public interface EntityType\n    {\n")
	generateEntityInterface(&output, construction.EntityTypeMethods)
	output.WriteString("    }\n\n    public interface EntityConfig\n    {\n")
	generateEntityInterface(&output, construction.EntityConfigMethods)
	output.WriteString("    }\n\n    public interface TickerEntity : Entity\n    {\n")
	generateEntityInterface(&output, construction.TickerMethods)
	output.WriteString("    }\n\n    public sealed class EntityData\n    {\n")
	generateEntityFields(&output, construction.EntityDataFields)
	output.WriteString("    }\n\n    public sealed class EntitySpawnOpts\n    {\n")
	generateEntityFields(&output, construction.SpawnFields)
	fmt.Fprintf(&output, "\n        public %s %s(%s) =>\n", construction.SpawnNew.ReturnType, construction.SpawnNew.Name, formatParameters(construction.SpawnNew.Parameters))
	output.WriteString("            PluginBridge.Host.NewEntity(this, t, conf);\n")
	output.WriteString("    }\n\n")
	fmt.Fprintf(&output, "    public static %s %s(%s) =>\n", construction.NewEntity.ReturnType, construction.NewEntity.Name, formatParameters(construction.NewEntity.Parameters))
	output.WriteString("        new EntitySpawnOpts().New(t, conf);\n\n")
	output.WriteString("    public interface Entity\n    {\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "        %s %s();\n", method.ReturnType, method.Name)
	}
	output.WriteString("    }\n\n    public sealed partial class EntityHandle\n    {\n")
	for index, method := range handleMethods {
		fmt.Fprintf(&output, "        public %s %s(%s) =>\n", method.ReturnType, method.Name, formatParameters(method.Parameters))
		switch method.Name {
		case "Entity":
			output.WriteString("            PluginBridge.Host.EntityHandleEntity(tx.Invocation, this);\n")
		case "UUID":
			output.WriteString("            PluginBridge.Host.EntityHandleUuid(this);\n")
		case "Closed":
			output.WriteString("            PluginBridge.Host.EntityHandleClosed(this);\n")
		case "Close":
			output.WriteString("            PluginBridge.Host.CloseEntityHandle(this);\n")
		default:
			panic("unsupported world.EntityHandle method: " + method.Name)
		}
		if index != len(handleMethods)-1 {
			output.WriteByte('\n')
		}
	}
	output.WriteString("    }\n}\n")
	return output.Bytes()
}

func generateEntityInterface(output *bytes.Buffer, methods []commandMethod) {
	for _, method := range methods {
		fmt.Fprintf(output, "        %s %s(%s);\n", method.ReturnType, method.Name, formatParameters(method.Parameters))
	}
}

func generateEntityFields(output *bytes.Buffer, fields []parameter) {
	for _, field := range fields {
		fmt.Fprintf(output, "        public %s %s", field.Type, field.Name)
		switch field.Type {
		case "string":
			output.WriteString(" = string.Empty")
		case "object?":
			// The zero value of Go's any is nil.
		default:
			output.WriteString(" = default")
		}
		output.WriteString(";\n")
	}
}
