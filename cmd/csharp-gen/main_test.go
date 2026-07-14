package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/df-mc/dragonfly/server/world"
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
func (tx *Tx) BlocksWithin(pos cube.Pos, radius int, blocks ...Block) iter.Seq[cube.Pos] { return nil }
func (tx *Tx) Liquid(pos cube.Pos) (Liquid, bool) { return nil, false }
func (tx *Tx) SetLiquid(pos cube.Pos, b Liquid) {}
func (tx *Tx) ScheduleBlockUpdate(pos cube.Pos, b Block, delay time.Duration) {}
func (tx *Tx) HighestLightBlocker(x, z int) int { return 0 }
func (tx *Tx) HighestBlock(x, z int) int { return 0 }
func (tx *Tx) Light(pos cube.Pos) uint8 { return 0 }
func (tx *Tx) SkyLight(pos cube.Pos) uint8 { return 0 }
func (tx *Tx) SetBiome(pos cube.Pos, b Biome) {}
func (tx *Tx) Biome(pos cube.Pos) Biome { return nil }
func (tx *Tx) Temperature(pos cube.Pos) float64 { return 0 }
func (tx *Tx) RainingAt(pos cube.Pos) bool { return false }
func (tx *Tx) SnowingAt(pos cube.Pos) bool { return false }
func (tx *Tx) ThunderingAt(pos cube.Pos) bool { return false }
func (tx *Tx) Raining() bool { return false }
func (tx *Tx) Thundering() bool { return false }`
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
		"public IEnumerable<Cube.Pos> BlocksWithin(Cube.Pos pos, int radius, params Block[] blocks)",
		"public interface Liquid : Block { }",
		"public (Liquid? Liquid, bool Ok) Liquid(Cube.Pos pos)",
		"public void SetLiquid(Cube.Pos pos, Liquid? b)",
		"using System;",
		"public void ScheduleBlockUpdate(Cube.Pos pos, Block b, TimeSpan delay)",
		"public int HighestLightBlocker(int x, int z)",
		"public int HighestBlock(int x, int z)",
		"public byte Light(Cube.Pos pos)",
		"public byte SkyLight(Cube.Pos pos)",
		"public interface Biome { }",
		"public void SetBiome(Cube.Pos pos, Biome b)",
		"public Biome Biome(Cube.Pos pos)",
		"public double Temperature(Cube.Pos pos)",
		"public bool RainingAt(Cube.Pos pos)",
		"public bool SnowingAt(Cube.Pos pos)",
		"public bool ThunderingAt(Cube.Pos pos)",
		"public bool Raining()",
		"public bool Thundering()",
		"public bool DisableRedstoneUpdates;",
	} {
		if !strings.Contains(worldOutput, expected) {
			t.Fatalf("generated world output missing %q:\n%s", expected, worldOutput)
		}
	}
	if strings.Contains(worldOutput, "Identifier") || strings.Contains(worldOutput, "PropertiesNBT") {
		t.Fatalf("public world surface exposes transport:\n%s", worldOutput)
	}
	withoutDuration := string(generateWorldBlock(nil, []commandMethod{{Name: "Range", ReturnType: "Cube.Range"}}))
	if strings.Contains(withoutDuration, "using System;") {
		t.Fatalf("world surface imports System without a System type:\n%s", withoutDuration)
	}

	blockOutput := string(generateBlocks(blockSpec{
		Stateless: []encodedBlock{{Name: "Air", Identifier: "minecraft:air", PropertiesNBT: []byte{10, 0, 0, 0}}},
		Sand: [2]encodedBlock{
			{Name: "Sand", Identifier: "minecraft:sand", PropertiesNBT: []byte{10, 0, 0, 0}},
			{Name: "Sand", Identifier: "minecraft:red_sand", PropertiesNBT: []byte{10, 0, 0, 0}},
		},
		Liquids: []liquidSpec{
			{Name: "Water", States: []encodedLiquid{{
				encodedBlock: encodedBlock{Name: "Water", Identifier: "minecraft:water", PropertiesNBT: []byte{10, 1}},
				Still:        true, Depth: 8,
			}}},
			{Name: "Lava", States: []encodedLiquid{{
				encodedBlock: encodedBlock{Name: "Lava", Identifier: "minecraft:flowing_lava", PropertiesNBT: []byte{10, 2}},
				Depth:        4, Falling: true,
			}}},
		},
	}))
	for _, expected := range []string{
		"public readonly record struct Air : World.Block;",
		"public readonly record struct Sand(bool Red = false) : World.Block;",
		"public readonly record struct Water(bool Still, int Depth, bool Falling) : World.Liquid;",
		"public readonly record struct Lava(bool Still, int Depth, bool Falling) : World.Liquid;",
		"internal static class BlockCodec",
		"case Block.Sand { Red: true }:",
		"case Block.Water { Still: true, Depth: 8, Falling: false }:",
		"internal static World.Liquid DecodeLiquid",
		"private sealed record EncodedBlock",
		"private sealed record EncodedLiquid",
	} {
		if !strings.Contains(blockOutput, expected) {
			t.Fatalf("generated block output missing %q:\n%s", expected, blockOutput)
		}
	}
	if strings.Contains(blockOutput, "public (string") || strings.Contains(blockOutput, "EncodeBlock()") {
		t.Fatalf("typed blocks expose encoded state:\n%s", blockOutput)
	}

	biomeOutput := string(generateBiomes([]encodedBiome{{Name: "Desert", ID: 2}, {Name: "Plains", ID: 1}}))
	for _, expected := range []string{
		"public readonly record struct Desert : World.Biome;",
		"public readonly record struct Plains : World.Biome;",
		"internal static class BiomeCodec",
		"case Biome.Desert _:",
		"id = 2; return true;",
		"if (id == 1) return new Biome.Plains();",
		"return new EncodedBiome(id);",
		"private sealed record EncodedBiome(int Id) : World.Biome;",
	} {
		if !strings.Contains(biomeOutput, expected) {
			t.Fatalf("generated biome output missing %q:\n%s", expected, biomeOutput)
		}
	}
}

func TestInspectWorldTxRejectsScheduleBlockUpdateDrift(t *testing.T) {
	for name, declaration := range map[string]string{
		"position name": "func (tx *Tx) ScheduleBlockUpdate(position cube.Pos, b Block, delay time.Duration) {}",
		"block type":    "func (tx *Tx) ScheduleBlockUpdate(pos cube.Pos, b Liquid, delay time.Duration) {}",
		"delay type":    "func (tx *Tx) ScheduleBlockUpdate(pos cube.Pos, b Block, delay int) {}",
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tx.go")
			if err := os.WriteFile(path, []byte("package world\n"+declaration), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := inspectWorldTx(path)
			if err == nil || !strings.Contains(err.Error(), "world.Tx.ScheduleBlockUpdate: signature changed") {
				t.Fatalf("expected ScheduleBlockUpdate signature drift error, got %v", err)
			}
		})
	}
}

func TestValidateLiquidFieldsRejectsASTDrift(t *testing.T) {
	parse := func(source string) *ast.TypeSpec {
		t.Helper()
		file, err := parser.ParseFile(token.NewFileSet(), "liquid.go", source, 0)
		if err != nil {
			t.Fatal(err)
		}
		var result *ast.TypeSpec
		ast.Inspect(file, func(node ast.Node) bool {
			spec, ok := node.(*ast.TypeSpec)
			if ok && spec.Name.Name == "Water" {
				result = spec
			}
			return true
		})
		return result
	}
	valid := parse(`package block
type Water struct {
	empty
	Still bool
	Depth int
	Falling bool
}`)
	if err := validateLiquidFields(valid, "Water"); err != nil {
		t.Fatalf("valid liquid fields rejected: %v", err)
	}
	changed := parse(`package block
type Water struct {
	empty
	Still bool
	Depth int32
	Falling bool
}`)
	if err := validateLiquidFields(changed, "Water"); err == nil || !strings.Contains(err.Error(), "fields changed") {
		t.Fatalf("expected liquid field drift error, got %v", err)
	}
}

func TestInspectBlocksUsesCanonicalLiquidStates(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	spec, err := inspectBlocks(filepath.Join(string(bytes.TrimSpace(output)), "server", "block"))
	if err != nil {
		t.Fatal(err)
	}
	state := func(name string, still bool, depth int, falling bool) encodedLiquid {
		t.Helper()
		for _, liquid := range spec.Liquids {
			if liquid.Name != name {
				continue
			}
			for _, candidate := range liquid.States {
				if candidate.Still == still && candidate.Depth == depth && candidate.Falling == falling {
					return candidate
				}
			}
		}
		t.Fatalf("%s state (%t, %d, %t) not generated", name, still, depth, falling)
		return encodedLiquid{}
	}
	for _, test := range []struct {
		name       string
		state      encodedLiquid
		identifier string
		depth      byte
	}{
		{name: "still water", state: state("Water", true, 8, false), identifier: "minecraft:water", depth: 0},
		{name: "falling lava", state: state("Lava", false, 4, true), identifier: "minecraft:flowing_lava", depth: 12},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.state.Identifier != test.identifier || !bytes.Equal(test.state.PropertiesNBT, canonicalLiquidNBT(test.depth)) {
				t.Fatalf("state = %q, %v", test.state.Identifier, test.state.PropertiesNBT)
			}
		})
	}
}

func TestInspectBiomesUsesASTAndRegistry(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	biomes, err := inspectBiomes(filepath.Join(string(bytes.TrimSpace(output)), "server", "world", "biome"))
	if err != nil {
		t.Fatal(err)
	}
	if len(biomes) != len(world.Biomes()) || len(biomes) == 0 {
		t.Fatalf("generated %d biomes for %d registry entries", len(biomes), len(world.Biomes()))
	}
	foundPlains := false
	for index, biome := range biomes {
		if index != 0 && biomes[index-1].Name >= biome.Name {
			t.Fatalf("biomes are not strictly sorted at %q, %q", biomes[index-1].Name, biome.Name)
		}
		if biome.Name == "Plains" {
			foundPlains = true
			if biome.ID != 1 {
				t.Fatalf("Plains ID = %d, want 1", biome.ID)
			}
		}
	}
	if !foundPlains {
		t.Fatal("Plains biome was not generated")
	}
}

func canonicalLiquidNBT(depth byte) []byte {
	return []byte{
		10, 0, 0,
		10, 12, 0, 'l', 'i', 'q', 'u', 'i', 'd', '_', 'd', 'e', 'p', 't', 'h',
		3, 4, 0, 'k', 'i', 'n', 'd', 2, 0, 0, 0,
		3, 5, 0, 'v', 'a', 'l', 'u', 'e', depth, 0, 0, 0,
		0,
		0,
	}
}

func TestInspectWorldTxRejectsVariadicDrift(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tx.go")
	source := `package world
func (tx *Tx) Range() cube.Range { return cube.Range{} }
func (tx *Tx) SetBlock(pos cube.Pos, b Block, opts *SetOpts) {}
func (tx *Tx) Block(pos cube.Pos) Block { return nil }
func (tx *Tx) BlockLoaded(pos cube.Pos) (Block, bool) { return nil, false }
func (tx *Tx) BlocksWithin(pos cube.Pos, radius int, blocks []Block) iter.Seq[cube.Pos] { return nil }
func (tx *Tx) HighestLightBlocker(x, z int) int { return 0 }
func (tx *Tx) HighestBlock(x, z int) int { return 0 }
func (tx *Tx) Light(pos cube.Pos) uint8 { return 0 }
func (tx *Tx) SkyLight(pos cube.Pos) uint8 { return 0 }`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := inspectWorldTx(path)
	if err == nil || !strings.Contains(err.Error(), "world.Tx.BlocksWithin: unsupported parameter type []Block") {
		t.Fatalf("expected changed BlocksWithin variadic error, got %v", err)
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
