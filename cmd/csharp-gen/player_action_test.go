package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"testing"
)

func TestPlayerActionMethodsMatchPinnedDragonfly(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	directory, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	methods, err := inspectPlayerActionMethods(filepath.Join(string(bytes.TrimSpace(directory)), "server", "player", "player.go"))
	if err != nil {
		t.Fatal(err)
	}
	generated := generatePlayerActionMethods(methods)
	for _, name := range selectedPlayerActionMethods {
		expected := []byte("public void " + name + "()")
		if !bytes.Contains(generated, expected) {
			t.Fatalf("generated player actions missing %q", expected)
		}
	}
}

func TestEveryZeroArgumentVoidPlayerMethodIsSelected(t *testing.T) {
	directory := moduleDirectoryForTest(t, "github.com/df-mc/dragonfly")
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(directory, "server", "player", "player.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	selected := map[string]bool{}
	for _, names := range [][]string{
		selectedPlayerActionMethods,
		selectedPlayerStateMethods,
		selectedPlayerPresentationMethods,
		selectedPlayerFormMethods,
	} {
		for _, name := range names {
			selected[name] = true
		}
	}
	var found, covered []string
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !playerMethod(function) || !ast.IsExported(function.Name.Name) ||
			goFunctionSignature(function) != (goSignature{}) {
			continue
		}
		found = append(found, function.Name.Name)
		if selected[function.Name.Name] {
			covered = append(covered, function.Name.Name)
		}
	}
	sort.Strings(found)
	sort.Strings(covered)
	if !slices.Equal(found, covered) {
		t.Fatalf("zero-argument void Player coverage = %v, want %v", covered, found)
	}
}
