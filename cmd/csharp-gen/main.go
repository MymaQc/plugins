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
	"slices"
)

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

func playerHandlerMethods(path string) ([]string, error) {
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
			var methods []string
			for _, field := range interfaceType.Methods.List {
				for _, name := range field.Names {
					if name.Name == "HandleQuit" {
						methods = append(methods, name.Name)
					}
				}
			}
			slices.Sort(methods)
			if !slices.Contains(methods, "HandleQuit") {
				return nil, fmt.Errorf("Dragonfly player.Handler has no HandleQuit method")
			}
			return methods, nil
		}
	}
	return nil, fmt.Errorf("Dragonfly player.Handler interface not found")
}

func generatePlayerHandler(methods []string) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/handler.go. DO NOT EDIT.\n")
	output.WriteString("namespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class Player\n{\n    public interface Handler\n    {\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "        void %s(Player player);\n", method)
	}
	output.WriteString("    }\n}\n")
	return output.Bytes()
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
