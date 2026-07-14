package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type itemStackSpec struct {
	NewStack itemStackFunction
	Methods  []itemStackFunction
}

type itemStackFunction struct {
	Name       string
	Parameters []itemStackParameter
}

type itemStackParameter struct {
	Name     string
	Type     string
	Variadic bool
}

type itemStackMethodSpec struct {
	Signature goSignature
	Result    string
}

var itemStackMethods = map[string]itemStackMethodSpec{
	"Count":                  {goSignature{Results: "int"}, "int"},
	"MaxCount":               {goSignature{Results: "int"}, "int"},
	"Grow":                   {goSignature{Parameters: "int", Results: "Stack"}, "Stack"},
	"Durability":             {goSignature{Results: "int"}, "int"},
	"MaxDurability":          {goSignature{Results: "int"}, "int"},
	"Damage":                 {goSignature{Parameters: "int", Results: "Stack"}, "Stack"},
	"WithDurability":         {goSignature{Parameters: "int", Results: "Stack"}, "Stack"},
	"Unbreakable":            {goSignature{Results: "bool"}, "bool"},
	"AsUnbreakable":          {goSignature{Results: "Stack"}, "Stack"},
	"AsBreakable":            {goSignature{Results: "Stack"}, "Stack"},
	"Empty":                  {goSignature{Results: "bool"}, "bool"},
	"Item":                   {goSignature{Results: "world.Item"}, "World.Item?"},
	"AttackDamage":           {goSignature{Results: "float64"}, "double"},
	"WithCustomName":         {goSignature{Parameters: "...any", Results: "Stack"}, "Stack"},
	"CustomName":             {goSignature{Results: "string"}, "string"},
	"WithLore":               {goSignature{Parameters: "...string", Results: "Stack"}, "Stack"},
	"Lore":                   {goSignature{Results: "[]string"}, "string[]"},
	"WithValue":              {goSignature{Parameters: "string, any", Results: "Stack"}, "Stack"},
	"Value":                  {goSignature{Parameters: "string", Results: "any, bool"}, "(object? Value, bool Ok)"},
	"WithEnchantments":       {goSignature{Parameters: "...Enchantment", Results: "Stack"}, "Stack"},
	"WithForcedEnchantments": {goSignature{Parameters: "...Enchantment", Results: "Stack"}, "Stack"},
	"WithoutEnchantments":    {goSignature{Parameters: "...EnchantmentType", Results: "Stack"}, "Stack"},
	"Enchantment":            {goSignature{Parameters: "EnchantmentType", Results: "Enchantment, bool"}, "(Enchantment Enchantment, bool Ok)"},
	"Enchantments":           {goSignature{Results: "[]Enchantment"}, "Enchantment[]"},
	"AnvilCost":              {goSignature{Results: "int"}, "int"},
	"WithAnvilCost":          {goSignature{Parameters: "int", Results: "Stack"}, "Stack"},
	"WithItem":               {goSignature{Parameters: "world.Item", Results: "Stack"}, "Stack"},
	"AddStack":               {goSignature{Parameters: "Stack", Results: "Stack, Stack"}, "(Stack A, Stack B)"},
	"Equal":                  {goSignature{Parameters: "Stack", Results: "bool"}, "bool"},
	"Comparable":             {goSignature{Parameters: "Stack", Results: "bool"}, "bool"},
	"String":                 {goSignature{Results: "string"}, "string"},
	"Values":                 {goSignature{Results: "map[string]any"}, "IReadOnlyDictionary<string, object>"},
}

func inspectItemStack(path string) (itemStackSpec, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return itemStackSpec{}, err
	}
	var spec itemStackSpec
	found := map[string]bool{}
	hasStack := false
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.GenDecl:
			for _, specification := range declaration.Specs {
				typeSpec, ok := specification.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != "Stack" {
					continue
				}
				_, hasStack = typeSpec.Type.(*ast.StructType)
			}
		case *ast.FuncDecl:
			if declaration.Recv == nil && declaration.Name.Name == "NewStack" {
				if got, want := goFunctionSignature(declaration), (goSignature{Parameters: "world.Item, int", Results: "Stack"}); got != want {
					return itemStackSpec{}, fmt.Errorf("Dragonfly item.NewStack signature changed: %+v", got)
				}
				function, err := inspectItemStackFunction(declaration)
				if err != nil {
					return itemStackSpec{}, err
				}
				spec.NewStack = function
				continue
			}
			if receiverTypeName(declaration) != "Stack" || !ast.IsExported(declaration.Name.Name) {
				continue
			}
			method, ok := itemStackMethods[declaration.Name.Name]
			if !ok {
				return itemStackSpec{}, fmt.Errorf("unsupported Dragonfly item.Stack.%s method", declaration.Name.Name)
			}
			if got := goFunctionSignature(declaration); got != method.Signature {
				return itemStackSpec{}, fmt.Errorf("Dragonfly item.Stack.%s signature changed: %+v", declaration.Name.Name, got)
			}
			function, err := inspectItemStackFunction(declaration)
			if err != nil {
				return itemStackSpec{}, err
			}
			spec.Methods = append(spec.Methods, function)
			found[function.Name] = true
		}
	}
	if !hasStack {
		return itemStackSpec{}, fmt.Errorf("Dragonfly item has no Stack struct")
	}
	if spec.NewStack.Name == "" {
		return itemStackSpec{}, fmt.Errorf("Dragonfly item has no NewStack function")
	}
	for name := range itemStackMethods {
		if !found[name] {
			return itemStackSpec{}, fmt.Errorf("Dragonfly item.Stack has no %s method", name)
		}
	}
	return spec, nil
}

func inspectItemStackFunction(function *ast.FuncDecl) (itemStackFunction, error) {
	result := itemStackFunction{Name: function.Name.Name}
	if function.Type.Params == nil {
		return result, nil
	}
	for _, field := range function.Type.Params.List {
		if len(field.Names) == 0 {
			return itemStackFunction{}, fmt.Errorf("Dragonfly item.%s has unnamed parameter", function.Name.Name)
		}
		goType := formatGoExpression(field.Type)
		variadic := strings.HasPrefix(goType, "...")
		if variadic {
			goType = strings.TrimPrefix(goType, "...")
		}
		csharpType, ok := itemStackParameterType(goType)
		if !ok {
			return itemStackFunction{}, fmt.Errorf("Dragonfly item.%s has unsupported parameter type %s", function.Name.Name, goType)
		}
		for _, name := range field.Names {
			result.Parameters = append(result.Parameters, itemStackParameter{Name: name.Name, Type: csharpType, Variadic: variadic})
		}
	}
	return result, nil
}

func itemStackParameterType(goType string) (string, bool) {
	typeName, ok := map[string]string{
		"int":             "int",
		"world.Item":      "World.Item",
		"Stack":           "Stack",
		"any":             "object?",
		"string":          "string",
		"Enchantment":     "Enchantment",
		"EnchantmentType": "EnchantmentType",
	}[goType]
	return typeName, ok
}

func generateItemStack(spec itemStackSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/item/stack.go Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing System.Collections.Generic;\n\nnamespace Dragonfly;\n\npublic static partial class Item\n{\n")
	writeItemStackWrapper(&output, spec.NewStack, "Stack", true)
	output.WriteString("\n    public readonly partial struct Stack\n    {\n")
	for index, method := range spec.Methods {
		writeItemStackWrapper(&output, method, itemStackMethods[method.Name].Result, false)
		if index != len(spec.Methods)-1 {
			output.WriteByte('\n')
		}
	}
	output.WriteString("    }\n}\n")
	return output.Bytes()
}

func writeItemStackWrapper(output *bytes.Buffer, function itemStackFunction, result string, static bool) {
	output.WriteString("    ")
	if static {
		output.WriteString("public static ")
	} else {
		output.WriteString("    public ")
	}
	fmt.Fprintf(output, "%s %s(", result, function.Name)
	for index, parameter := range function.Parameters {
		if index != 0 {
			output.WriteString(", ")
		}
		if parameter.Variadic {
			output.WriteString("params ")
		}
		fmt.Fprintf(output, "%s%s %s", parameter.Type, map[bool]string{true: "[]"}[parameter.Variadic], csharpItemStackIdentifier(parameter.Name))
	}
	output.WriteString(") =>\n")
	if static {
		output.WriteString("        ")
	} else {
		output.WriteString("            ")
	}
	fmt.Fprintf(output, "%sImpl(", function.Name)
	for index, parameter := range function.Parameters {
		if index != 0 {
			output.WriteString(", ")
		}
		output.WriteString(csharpItemStackIdentifier(parameter.Name))
	}
	output.WriteString(");\n")
}

func csharpItemStackIdentifier(name string) string {
	if _, keyword := map[string]struct{}{
		"abstract": {}, "as": {}, "base": {}, "bool": {}, "break": {}, "byte": {}, "case": {}, "catch": {},
		"char": {}, "checked": {}, "class": {}, "const": {}, "continue": {}, "decimal": {}, "default": {},
		"delegate": {}, "do": {}, "double": {}, "else": {}, "enum": {}, "event": {}, "explicit": {}, "extern": {},
		"false": {}, "finally": {}, "fixed": {}, "float": {}, "for": {}, "foreach": {}, "goto": {}, "if": {},
		"implicit": {}, "in": {}, "int": {}, "interface": {}, "internal": {}, "is": {}, "lock": {}, "long": {},
		"namespace": {}, "new": {}, "null": {}, "object": {}, "operator": {}, "out": {}, "override": {}, "params": {},
		"private": {}, "protected": {}, "public": {}, "readonly": {}, "ref": {}, "return": {}, "sbyte": {}, "sealed": {},
		"short": {}, "sizeof": {}, "stackalloc": {}, "static": {}, "string": {}, "struct": {}, "switch": {}, "this": {},
		"throw": {}, "true": {}, "try": {}, "typeof": {}, "uint": {}, "ulong": {}, "unchecked": {}, "unsafe": {},
		"ushort": {}, "using": {}, "virtual": {}, "void": {}, "volatile": {}, "while": {},
	}[name]; keyword {
		return "@" + name
	}
	return name
}
