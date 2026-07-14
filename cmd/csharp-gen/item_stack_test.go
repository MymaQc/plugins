package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const itemStackFixture = `package item

import "example/world"

type Enchantment struct{}
type EnchantmentType interface{}
type Stack struct{}

func NewStack(t world.Item, count int) Stack { return Stack{} }
func (Stack) Count() int { return 0 }
func (Stack) MaxCount() int { return 0 }
func (Stack) Grow(n int) Stack { return Stack{} }
func (Stack) Durability() int { return 0 }
func (Stack) MaxDurability() int { return 0 }
func (Stack) Damage(d int) Stack { return Stack{} }
func (Stack) WithDurability(d int) Stack { return Stack{} }
func (Stack) Unbreakable() bool { return false }
func (Stack) AsUnbreakable() Stack { return Stack{} }
func (Stack) AsBreakable() Stack { return Stack{} }
func (Stack) Empty() bool { return false }
func (Stack) Item() world.Item { return nil }
func (Stack) AttackDamage() float64 { return 0 }
func (Stack) WithCustomName(a ...any) Stack { return Stack{} }
func (Stack) CustomName() string { return "" }
func (Stack) WithLore(lines ...string) Stack { return Stack{} }
func (Stack) Lore() []string { return nil }
func (Stack) WithValue(key string, val any) Stack { return Stack{} }
func (Stack) Value(key string) (any, bool) { return nil, false }
func (Stack) WithEnchantments(enchants ...Enchantment) Stack { return Stack{} }
func (Stack) WithForcedEnchantments(enchants ...Enchantment) Stack { return Stack{} }
func (Stack) WithoutEnchantments(enchants ...EnchantmentType) Stack { return Stack{} }
func (Stack) Enchantment(enchant EnchantmentType) (Enchantment, bool) { return Enchantment{}, false }
func (Stack) Enchantments() []Enchantment { return nil }
func (Stack) AnvilCost() int { return 0 }
func (Stack) WithAnvilCost(anvilCost int) Stack { return Stack{} }
func (Stack) WithItem(t world.Item) Stack { return Stack{} }
func (Stack) AddStack(s2 Stack) (Stack, Stack) { return Stack{}, Stack{} }
func (Stack) Equal(s2 Stack) bool { return false }
func (Stack) Comparable(s2 Stack) bool { return false }
func (Stack) String() string { return "" }
func (Stack) Values() map[string]any { return nil }
`

func TestItemStackUsesGoAST(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stack.go")
	if err := os.WriteFile(path, []byte(itemStackFixture), 0o600); err != nil {
		t.Fatal(err)
	}
	spec, err := inspectItemStack(path)
	if err != nil {
		t.Fatal(err)
	}
	generated := string(generateItemStack(spec))
	for _, expected := range []string{
		"Code generated from Dragonfly server/item/stack.go Go AST",
		"public static Stack NewStack(World.Item t, int count)",
		"NewStackImpl(t, count)",
		"public readonly partial struct Stack",
		"public Stack WithItem(World.Item t)",
		"WithItemImpl(t)",
		"public Stack WithValue(string key, object? val)",
		"WithValueImpl(key, val)",
		"public Stack WithCustomName(params object?[] a)",
		"public string String()",
		"StringImpl()",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated item stack missing %q:\n%s", expected, generated)
		}
	}
	for _, privateDetail := range []string{"_damage", "private Stack Copy(", "ItemCapabilities"} {
		if strings.Contains(generated, privateDetail) {
			t.Fatalf("generated public surface leaked implementation %q:\n%s", privateDetail, generated)
		}
	}
}

func TestItemStackASTControlsParameterNamesAndMethodOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stack.go")
	source := strings.Replace(itemStackFixture,
		"func (Stack) WithValue(key string, val any)",
		"func (Stack) WithValue(namespace string, payload any)", 1)
	source = strings.Replace(source,
		"func (Stack) Count() int { return 0 }\nfunc (Stack) MaxCount() int { return 0 }",
		"func (Stack) MaxCount() int { return 0 }\nfunc (Stack) Count() int { return 0 }", 1)
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	spec, err := inspectItemStack(path)
	if err != nil {
		t.Fatal(err)
	}
	generated := string(generateItemStack(spec))
	if !strings.Contains(generated, "WithValue(string @namespace, object? payload)") ||
		!strings.Contains(generated, "WithValueImpl(@namespace, payload)") {
		t.Fatalf("AST parameter names not reflected in output:\n%s", generated)
	}
	if strings.Index(generated, " MaxCount()") > strings.Index(generated, " Count()") {
		t.Fatalf("AST method order not reflected in output:\n%s", generated)
	}
}

func TestItemStackRejectsSignatureDrift(t *testing.T) {
	tests := map[string][2]string{
		"new stack": {"func NewStack(t world.Item, count int) Stack", "func NewStack(t world.Item, count uint) Stack"},
		"with item": {"func (Stack) WithItem(t world.Item) Stack", "func (Stack) WithItem(t string) Stack"},
		"string":    {"func (Stack) String() string", "func (Stack) String() []byte"},
		"values":    {"func (Stack) Values() map[string]any", "func (Stack) Values() map[int]any"},
	}
	for name, replacement := range tests {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "stack.go")
			source := strings.Replace(itemStackFixture, replacement[0], replacement[1], 1)
			if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := inspectItemStack(path); err == nil || !strings.Contains(err.Error(), "signature changed") {
				t.Fatalf("expected signature drift error, got %v", err)
			}
		})
	}
}

func TestItemStackRejectsMissingStackStruct(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stack.go")
	source := strings.Replace(itemStackFixture, "type Stack struct{}", "type Stack interface{}", 1)
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := inspectItemStack(path); err == nil || !strings.Contains(err.Error(), "no Stack struct") {
		t.Fatalf("expected missing Stack struct error, got %v", err)
	}
}

func TestItemStackRejectsUnknownExportedMethod(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stack.go")
	source := itemStackFixture + "\nfunc (Stack) NewUpstreamMethod() bool { return false }\n"
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := inspectItemStack(path); err == nil || !strings.Contains(err.Error(), "unsupported Dragonfly item.Stack.NewUpstreamMethod method") {
		t.Fatalf("expected unsupported method error, got %v", err)
	}
}

func TestPinnedDragonflyItemHasExactStackSurface(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := inspectItemStack(filepath.Join(
		string(bytes.TrimSpace(output)), "server", "item", "stack.go")); err != nil {
		t.Fatal(err)
	}
}
