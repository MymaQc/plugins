package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlayerHandlerMethodsUsesGoAST(t *testing.T) {
	path := filepath.Join(t.TempDir(), "handler.go")
	source := `package player
type Handler interface {
	HandleMove(ctx *Context, newPos mgl64.Vec3, newRot cube.Rotation)
	HandleJump(p *Player)
	HandleTeleport(ctx *Context, pos mgl64.Vec3)
	HandleToggleSprint(ctx *Context, after bool)
	HandleToggleSneak(ctx *Context, after bool)
	HandleChat(ctx *Context, message *string)
	HandleFoodLoss(ctx *Context, from int, to *int)
	HandlePunchAir(ctx *Context)
	HandleQuit(p *Player)
}`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	methods, err := playerHandlerMethods(path)
	if err != nil {
		t.Fatal(err)
	}
	output := string(generatePlayerHandler(methods))
	for _, expected := range []string{
		"void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot);",
		"void HandleChat(Player.Context ctx, ref string message);",
		"void HandleFoodLoss(Player.Context ctx, int from, ref int to);",
		"void HandleQuit(Player p);",
		"[HandlerSubscription(1UL)]",
		"public virtual void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot) { }",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("generated output missing %q:\n%s", expected, output)
		}
	}
}

func TestPlayerTextMethodsUseGoAST(t *testing.T) {
	path := filepath.Join(t.TempDir(), "player.go")
	source := `package player
func (p *Player) Message(a ...any) {}
func (p *Player) SendPopup(a ...any) {}
func (p *Player) SendTip(a ...any) {}
func (p *Player) SendJukeboxPopup(a ...any) {}
func (p *Player) SetNameTag(name string) {}
func (p *Player) Disconnect(msg ...any) {}`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	methods, err := playerTextMethods(path)
	if err != nil {
		t.Fatal(err)
	}
	output := string(generatePlayerTextMethods(methods))
	for _, expected := range []string{
		"public void Message(params object?[] a) => SendText(Abi.PlayerTextMessage, FormatArguments(a));",
		"public void SendPopup(params object?[] a) => SendText(Abi.PlayerTextPopup, FormatArguments(a));",
		"public void SendTip(params object?[] a) => SendText(Abi.PlayerTextTip, FormatArguments(a));",
		"public void SendJukeboxPopup(params object?[] a) => SendText(Abi.PlayerTextJukeboxPopup, FormatArguments(a));",
		"public void SetNameTag(string name) => SendText(Abi.PlayerTextNameTag, name);",
		"public void Disconnect(params object?[] msg) => SendText(Abi.PlayerTextDisconnect, FormatArguments(msg));",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("generated output missing %q:\n%s", expected, output)
		}
	}
}

func TestCommandInterfacesUseGoAST(t *testing.T) {
	directory := t.TempDir()
	source := `package cmd
type Runnable interface {
	Run(src Source, o *Output, tx *world.Tx)
}
type Allower interface {
	Allow(src Source) bool
}
type Target interface {
	Position() mgl64.Vec3
}
type NamedTarget interface {
	Target
	Name() string
}
type Source interface {
	Target
	SendCommandOutput(o *Output)
}
type Enum interface {
	Type() string
	Options(source Source) []string
}`
	if err := os.WriteFile(filepath.Join(directory, "cmd.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	interfaces, err := commandInterfaces(directory)
	if err != nil {
		t.Fatal(err)
	}
	output := string(generateCommandInterfaces(interfaces))
	for _, expected := range []string{
		"public static partial class Cmd\n{",
		"    public interface Runnable\n    {\n        void Run(Source src, Output o, World.Tx? tx);\n    }",
		"    public interface Allower\n    {\n        bool Allow(Source src);\n    }",
		"    public interface Target\n    {\n        Vector3 Position();\n    }",
		"    public interface NamedTarget : Target\n    {\n        string Name();\n    }",
		"    public interface Source : Target\n    {\n        void SendCommandOutput(Output o);\n    }",
		"    public interface Enum\n    {\n        string Type();\n        IReadOnlyList<string> Options(Source source);\n    }",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("generated output missing %q:\n%s", expected, output)
		}
	}
}

func TestSyncGeneratedFilesChecksEveryOutput(t *testing.T) {
	directory := t.TempDir()
	first := filepath.Join(directory, "first.g.cs")
	second := filepath.Join(directory, "second.g.cs")
	files := []generatedFile{
		{Path: first, Content: []byte("first")},
		{Path: second, Content: []byte("second")},
	}
	if err := syncGeneratedFiles(files, false); err != nil {
		t.Fatal(err)
	}
	if err := syncGeneratedFiles(files, true); err != nil {
		t.Fatalf("fresh files reported stale: %v", err)
	}
	if err := os.WriteFile(second, []byte("stale"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := syncGeneratedFiles(files, true)
	if err == nil || !strings.Contains(err.Error(), second+" is stale") {
		t.Fatalf("expected stale second output error, got %v", err)
	}
}

func TestGeneratedWorldBlockSurfaceKeepsTransportPrivate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tx.go")
	source := `package world
func (tx *Tx) Range() cube.Range { return cube.Range{} }
func (tx *Tx) SetBlock(pos cube.Pos, b Block, opts *SetOpts) {}
func (tx *Tx) Block(pos cube.Pos) Block { return nil }
func (tx *Tx) BlockLoaded(pos cube.Pos) (Block, bool) { return nil, false }
func (tx *Tx) Liquid(pos cube.Pos) (Liquid, bool) { return nil, false }`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	methods, err := inspectWorldTx(path)
	if err != nil {
		t.Fatal(err)
	}
	worldOutput := string(generateWorldBlock([]string{
		"DisableBlockUpdates", "DisableLiquidDisplacement", "DisableRedstoneUpdates",
	}, methods))
	for _, expected := range []string{
		"public interface Block { }",
		"public Cube.Range Range()",
		"public void SetBlock(Cube.Pos pos, Block? b, SetOpts? opts = null)",
		"public Block Block(Cube.Pos pos)",
		"public (Block? Block, bool Ok) BlockLoaded(Cube.Pos pos)",
		"public bool DisableRedstoneUpdates;",
	} {
		if !strings.Contains(worldOutput, expected) {
			t.Fatalf("generated world output missing %q:\n%s", expected, worldOutput)
		}
	}
	if strings.Contains(worldOutput, "Identifier") || strings.Contains(worldOutput, "PropertiesNBT") {
		t.Fatalf("public world surface exposes transport:\n%s", worldOutput)
	}

	blockOutput := string(generateBlocks(blockSpec{
		Stateless: []encodedBlock{{Name: "Air", Identifier: "minecraft:air", PropertiesNBT: []byte{10, 0, 0, 0}}},
		Sand: [2]encodedBlock{
			{Name: "Sand", Identifier: "minecraft:sand", PropertiesNBT: []byte{10, 0, 0, 0}},
			{Name: "Sand", Identifier: "minecraft:red_sand", PropertiesNBT: []byte{10, 0, 0, 0}},
		},
	}))
	for _, expected := range []string{
		"public readonly record struct Air : World.Block;",
		"public readonly record struct Sand(bool Red = false) : World.Block;",
		"internal static class BlockCodec",
		"case Block.Sand { Red: true }:",
		"private sealed record EncodedBlock",
	} {
		if !strings.Contains(blockOutput, expected) {
			t.Fatalf("generated block output missing %q:\n%s", expected, blockOutput)
		}
	}
	if strings.Contains(blockOutput, "public (string") || strings.Contains(blockOutput, "EncodeBlock()") {
		t.Fatalf("typed blocks expose encoded state:\n%s", blockOutput)
	}
}

func TestInspectWorldTxRejectsChangedSelectedSignature(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tx.go")
	source := `package world
func (tx *Tx) Range() cube.Range { return cube.Range{} }
func (tx *Tx) SetBlock(pos cube.Pos, b Block, opts *SetOpts) {}
func (tx *Tx) Block(pos cube.Pos) Block { return nil }
func (tx *Tx) BlockLoaded(pos cube.Pos) (Block, error) { return nil, nil }`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := inspectWorldTx(path)
	if err == nil || !strings.Contains(err.Error(), "world.Tx.BlockLoaded: unsupported return type error") {
		t.Fatalf("expected changed BlockLoaded signature error, got %v", err)
	}
}
