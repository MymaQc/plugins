// Command csharp-gen emits the supported C# surface directly from Dragonfly's Go AST.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type method struct {
	Name         string
	Parameters   []parameter
	Subscription uint64
}

type parameter struct {
	Name string
	Type string
}

var supportedPlayerHandlers = map[string]uint64{
	"HandleChat":         1 << 1,
	"HandleFoodLoss":     1 << 8,
	"HandleJump":         1 << 14,
	"HandleMove":         1,
	"HandlePunchAir":     1 << 17,
	"HandleQuit":         1 << 3,
	"HandleTeleport":     1 << 15,
	"HandleToggleSneak":  1 << 13,
	"HandleToggleSprint": 1 << 12,
}

func main() {
	root := flag.String("root", ".", "repository root")
	dragonfly := flag.String("dragonfly", "", "Dragonfly module directory")
	check := flag.Bool("check", false, "fail if generated output differs")
	flag.Parse()

	directory := *dragonfly
	if directory == "" {
		command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
		command.Dir = *root
		output, err := command.Output()
		if err != nil {
			fatal(err)
		}
		directory = string(bytes.TrimSpace(output))
	}
	methods, err := playerHandlerMethods(filepath.Join(directory, "server", "player", "handler.go"))
	if err != nil {
		fatal(err)
	}
	generated := generatePlayerHandler(methods)
	path := filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Handler.g.cs")
	if *check {
		current, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(current, generated) {
			fatal(fmt.Errorf("%s is stale; run make generate", path))
		}
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(path, generated, 0o644); err != nil {
		fatal(err)
	}
}

func playerHandlerMethods(path string) ([]method, error) {
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
			if !ok || typeSpec.Name.Name != "Handler" {
				continue
			}
			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				return nil, fmt.Errorf("player.Handler is not an interface")
			}
			var methods []method
			for _, field := range interfaceType.Methods.List {
				if len(field.Names) != 1 {
					continue
				}
				subscription, supported := supportedPlayerHandlers[field.Names[0].Name]
				if !supported {
					continue
				}
				function, ok := field.Type.(*ast.FuncType)
				if !ok {
					return nil, fmt.Errorf("player.Handler.%s is not a method", field.Names[0].Name)
				}
				translated, err := translateParameters(function.Params)
				if err != nil {
					return nil, fmt.Errorf("player.Handler.%s: %w", field.Names[0].Name, err)
				}
				methods = append(methods, method{Name: field.Names[0].Name, Parameters: translated, Subscription: subscription})
			}
			for name := range supportedPlayerHandlers {
				found := false
				for _, method := range methods {
					found = found || method.Name == name
				}
				if !found {
					return nil, fmt.Errorf("Dragonfly player.Handler has no supported %s method", name)
				}
			}
			return methods, nil
		}
	}
	return nil, fmt.Errorf("Dragonfly player.Handler interface not found")
}

func translateParameters(fields *ast.FieldList) ([]parameter, error) {
	var parameters []parameter
	for _, field := range fields.List {
		typeName, ok := csharpType(field.Type)
		if !ok {
			return nil, fmt.Errorf("unsupported parameter type %T", field.Type)
		}
		for _, name := range field.Names {
			parameters = append(parameters, parameter{Name: name.Name, Type: typeName})
		}
	}
	return parameters, nil
}

func csharpType(expression ast.Expr) (string, bool) {
	switch value := expression.(type) {
	case *ast.StarExpr:
		typeName, ok := csharpType(value.X)
		if !ok {
			return "", false
		}
		if typeName == "string" || typeName == "int" || typeName == "double" || typeName == "bool" {
			return "ref " + typeName, true
		}
		return typeName, true
	case *ast.Ident:
		typeName, ok := map[string]string{
			"bool":    "bool",
			"Context": "Player.Context",
			"int":     "int",
			"Player":  "Player",
			"string":  "string",
		}[value.Name]
		return typeName, ok
	case *ast.SelectorExpr:
		packageName, ok := value.X.(*ast.Ident)
		if !ok {
			return "", false
		}
		typeName, ok := map[string]string{
			"mgl64.Vec3":    "Vector3",
			"cube.Rotation": "Rotation",
		}[packageName.Name+"."+value.Sel.Name]
		return typeName, ok
	default:
		return "", false
	}
}

func generatePlayerHandler(methods []method) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/handler.go. DO NOT EDIT.\n")
	output.WriteString("namespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class Player\n{\n    public interface Handler\n    {\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "        void %s(%s);\n", method.Name, formatParameters(method.Parameters))
	}
	output.WriteString("    }\n}\n\n")
	output.WriteString("public abstract partial class Plugin\n{\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "    [HandlerSubscription(%dUL)]\n", method.Subscription)
		fmt.Fprintf(&output, "    public virtual void %s(%s) { }\n", method.Name, formatParameters(method.Parameters))
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func formatParameters(parameters []parameter) string {
	values := make([]string, len(parameters))
	for index, parameter := range parameters {
		values[index] = parameter.Type + " " + parameter.Name
	}
	return strings.Join(values, ", ")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
