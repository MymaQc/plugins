package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestBlockFieldAtASTPathFollowsUnderlyingAndPromotedFields(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "blocks.go", `package block
type crop struct { Growth int }
type WheatSeeds struct { crop }
type polishable struct { Polished bool }
type Andesite polishable
type shadow struct { Growth bool }
type Direct struct { crop; Growth bool }
`, 0)
	if err != nil {
		t.Fatal(err)
	}
	declarations := map[string]*ast.TypeSpec{}
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, raw := range general.Specs {
			spec := raw.(*ast.TypeSpec)
			declarations[spec.Name.Name] = spec
		}
	}
	for _, test := range []struct {
		name string
		path []int
		want blockField
	}{
		{name: "WheatSeeds", path: []int{0, 0}, want: blockField{Name: "Growth", Type: "int"}},
		{name: "Andesite", path: []int{0}, want: blockField{Name: "Polished", Type: "bool"}},
		{name: "Direct", path: []int{1}, want: blockField{Name: "Growth", Type: "bool"}},
	} {
		got, err := blockFieldAtASTPath(test.name, test.path, declarations)
		if err != nil || got != test.want {
			t.Fatalf("%s path %v = %#v, %v; want %#v", test.name, test.path, got, err, test.want)
		}
	}
}

func TestInspectBlocksIncludesPrimitiveRegistryStates(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	spec, err := inspectBlocks(filepath.Join(string(bytes.TrimSpace(output)), "server", "block"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(spec.Types), 112; got != want {
		t.Fatalf("primitive block types = %d, want %d", got, want)
	}
	states := 0
	for _, definition := range spec.Types {
		states += len(definition.States)
		for _, state := range definition.States {
			if len(state.Values) != len(definition.Fields) {
				t.Fatalf("block.%s state has %d values for %d fields", definition.Name, len(state.Values), len(definition.Fields))
			}
		}
	}
	if want := 306; states != want {
		t.Fatalf("primitive block states = %d, want %d", states, want)
	}
	generated := string(generateBlocks(spec))
	blockCases, liquidCases := blockDecodeCases(spec)
	allCases := append(blockCases, liquidCases...)
	if got, want := len(allCases), 370; got != want {
		t.Fatalf("generated codec states = %d, want %d", got, want)
	}
	for index, state := range allCases {
		if state.State != index {
			t.Fatalf("codec state index %d = %d", index, state.State)
		}
		decode := fmt.Sprintf("if (properties.SequenceEqual(State%d)) return %s;", state.State, state.Constructor)
		encode := fmt.Sprintf("identifier = %s; properties = State%d; return true;", strconv.Quote(state.Identifier), state.State)
		if !strings.Contains(generated, decode) || !strings.Contains(generated, encode) {
			t.Fatalf("generated codec is missing state %d (%s)", state.State, state.Identifier)
		}
	}
	for _, name := range []string{"BeetrootSeeds", "Carrot", "Potato", "WheatSeeds"} {
		definition := findBlockType(t, spec, name)
		if len(definition.Fields) != 1 || definition.Fields[0] != (blockField{Name: "Growth", Type: "int"}) {
			t.Fatalf("block.%s fields = %#v", name, definition.Fields)
		}
		if len(definition.States) != 8 {
			t.Fatalf("block.%s states = %d, want 8", name, len(definition.States))
		}
		var seen [8]bool
		for _, state := range definition.States {
			growth := state.Values[0].Int
			if growth < 0 || growth >= len(seen) || seen[growth] {
				t.Fatalf("block.%s has invalid or duplicate growth %d", name, growth)
			}
			seen[growth] = true
		}
	}
	for _, name := range []string{"Note", "ShortGrass"} {
		definition := findBlockType(t, spec, name)
		if len(definition.Fields) != 0 {
			t.Fatalf("block.%s exposes non-registry fields %#v", name, definition.Fields)
		}
	}
}

func findBlockType(t *testing.T, spec blockSpec, name string) blockTypeSpec {
	t.Helper()
	for _, definition := range spec.Types {
		if definition.Name == name {
			return definition
		}
	}
	t.Fatalf("block.%s was not generated", name)
	return blockTypeSpec{}
}
