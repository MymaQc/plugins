package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type entityMethod struct {
	Name       string
	ReturnType string
}

var selectedEntityHandleMethods = []string{"Entity", "UUID", "Closed", "Close"}

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

func generateWorldEntity(methods []entityMethod, handleMethods []commandMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/world/entity.go Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing System;\n\nnamespace Dragonfly;\n\npublic sealed partial class World\n{\n    public interface Entity\n    {\n")
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
