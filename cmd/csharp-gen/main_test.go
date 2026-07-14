package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/df-mc/dragonfly/server/world"
)

var expectedPlayerHandlerNames = []string{
	"HandleMove", "HandleJump", "HandleTeleport", "HandleChangeWorld", "HandleToggleSprint",
	"HandleToggleSneak", "HandleChat", "HandleFoodLoss", "HandleHeal", "HandleHurt", "HandleDeath",
	"HandleRespawn", "HandleSkinChange", "HandleFireExtinguish", "HandleStartBreak", "HandleBlockBreak",
	"HandleBlockPlace", "HandleBlockPick", "HandleItemUse", "HandleItemUseOnBlock", "HandleItemUseOnEntity",
	"HandleItemRelease", "HandleItemConsume", "HandleAttackEntity", "HandleExperienceGain", "HandlePunchAir",
	"HandleSignEdit", "HandleSleep", "HandleLecternPageTurn", "HandleItemDamage", "HandleItemPickup",
	"HandleHeldSlotChange", "HandleItemDrop", "HandleTransfer", "HandleCommandExecution", "HandleQuit",
	"HandleDiagnostics",
}

func TestPlayerHandlerMethodsUsesGoAST(t *testing.T) {
	path := filepath.Join(t.TempDir(), "handler.go")
	source := `package player
type Handler interface {
	HandleMove(ctx *Context, newPos mgl64.Vec3, newRot cube.Rotation)
	HandleJump(p *Player)
	HandleTeleport(ctx *Context, pos mgl64.Vec3)
	HandleChangeWorld(p *Player, before, after *world.World)
	HandleToggleSprint(ctx *Context, after bool)
	HandleToggleSneak(ctx *Context, after bool)
	HandleChat(ctx *Context, message *string)
	HandleFoodLoss(ctx *Context, from int, to *int)
	HandleHeal(ctx *Context, health *float64, src world.HealingSource)
	HandleHurt(ctx *Context, damage *float64, immune bool, attackImmunity *time.Duration, src world.DamageSource)
	HandleDeath(p *Player, src world.DamageSource, keepInv *bool)
	HandleRespawn(p *Player, pos *mgl64.Vec3, w **world.World)
	HandleSkinChange(ctx *Context, skin *skin.Skin)
	HandleFireExtinguish(ctx *Context, pos cube.Pos)
	HandleStartBreak(ctx *Context, pos cube.Pos)
	HandleBlockBreak(ctx *Context, pos cube.Pos, drops *[]item.Stack, xp *int)
	HandleBlockPlace(ctx *Context, pos cube.Pos, b world.Block)
	HandleBlockPick(ctx *Context, pos cube.Pos, b world.Block)
	HandleItemUse(ctx *Context)
	HandleItemUseOnBlock(ctx *Context, pos cube.Pos, face cube.Face, clickPos mgl64.Vec3)
	HandleItemUseOnEntity(ctx *Context, e world.Entity)
	HandleItemRelease(ctx *Context, item item.Stack, dur time.Duration)
	HandleItemConsume(ctx *Context, item item.Stack)
	HandleAttackEntity(ctx *Context, e world.Entity, force, height *float64, critical *bool)
	HandleExperienceGain(ctx *Context, amount *int)
	HandlePunchAir(ctx *Context)
	HandleSignEdit(ctx *Context, pos cube.Pos, frontSide bool, oldText, newText string)
	HandleSleep(ctx *Context, sendReminder *bool)
	HandleLecternPageTurn(ctx *Context, pos cube.Pos, oldPage int, newPage *int)
	HandleItemDamage(ctx *Context, i item.Stack, damage *int)
	HandleItemPickup(ctx *Context, i *item.Stack)
	HandleHeldSlotChange(ctx *Context, from, to int)
	HandleItemDrop(ctx *Context, s item.Stack)
	HandleTransfer(ctx *Context, addr *net.UDPAddr)
	HandleCommandExecution(ctx *Context, command cmd.Command, args []string)
	HandleQuit(p *Player)
	HandleDiagnostics(p *Player, d session.Diagnostics)
}`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	methods, err := playerHandlerMethods(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(methods) != 37 || len(supportedPlayerHandlers) != 37 {
		t.Fatalf("generated %d methods from %d supported handlers, want exactly 37", len(methods), len(supportedPlayerHandlers))
	}
	for index, method := range methods {
		if method.Name != expectedPlayerHandlerNames[index] {
			t.Fatalf("method %d = %s, want %s", index, method.Name, expectedPlayerHandlerNames[index])
		}
		if method.Subscription != supportedPlayerHandlers[method.Name] {
			t.Fatalf("%s subscription = %d, want %d", method.Name, method.Subscription, supportedPlayerHandlers[method.Name])
		}
	}
	output := string(generatePlayerHandler(methods))
	for _, expected := range []string{
		"void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot);",
		"void HandleChangeWorld(Player p, World? before, World after);",
		"void HandleChat(Player.Context ctx, ref string message);",
		"void HandleFoodLoss(Player.Context ctx, int from, ref int to);",
		"void HandleHeal(Player.Context ctx, ref double health, World.HealingSource src);",
		"void HandleHurt(Player.Context ctx, ref double damage, bool immune, ref TimeSpan attackImmunity, World.DamageSource src);",
		"void HandleDeath(Player p, World.DamageSource src, ref bool keepInv);",
		"void HandleRespawn(Player p, ref Vector3 pos, ref World w);",
		"void HandleSkinChange(Player.Context ctx, ref Skin skin);",
		"void HandleFireExtinguish(Player.Context ctx, Cube.Pos pos);",
		"void HandleBlockBreak(Player.Context ctx, Cube.Pos pos, ref Item.Stack[] drops, ref int xp);",
		"void HandleBlockPlace(Player.Context ctx, Cube.Pos pos, World.Block b);",
		"void HandleItemUseOnBlock(Player.Context ctx, Cube.Pos pos, Cube.Face face, Vector3 clickPos);",
		"void HandleItemUseOnEntity(Player.Context ctx, World.Entity e);",
		"void HandleItemRelease(Player.Context ctx, Item.Stack item, TimeSpan dur);",
		"void HandleAttackEntity(Player.Context ctx, World.Entity e, ref double force, ref double height, ref bool critical);",
		"void HandleItemPickup(Player.Context ctx, ref Item.Stack i);",
		"void HandleTransfer(Player.Context ctx, ref Net.UDPAddr addr);",
		"void HandleCommandExecution(Player.Context ctx, Cmd.Command command, string[] args);",
		"void HandleQuit(Player p);",
		"void HandleDiagnostics(Player p, Session.Diagnostics d);",
		"[HandlerSubscription(1UL)]",
		"[HandlerSubscription(16UL)]",
		"[HandlerSubscription(536870912UL)]",
		"[HandlerSubscription(137438953472UL)]",
		"[HandlerSubscription(274877906944UL)]",
		"[HandlerSubscription(549755813888UL)]",
		"[HandlerSubscription(1099511627776UL)]",
		"public virtual void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot) { }",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("generated output missing %q:\n%s", expected, output)
		}
	}
}

func TestPinnedDragonflyPlayerHandlerHasExactSupportedSurface(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	methods, err := playerHandlerMethods(filepath.Join(
		string(bytes.TrimSpace(output)), "server", "player", "handler.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(methods) != len(expectedPlayerHandlerNames) {
		t.Fatalf("generated %d pinned Dragonfly handlers, want %d", len(methods), len(expectedPlayerHandlerNames))
	}
	for index, method := range methods {
		if method.Name != expectedPlayerHandlerNames[index] {
			t.Fatalf("pinned method %d = %s, want %s", index, method.Name, expectedPlayerHandlerNames[index])
		}
	}
}

func TestPlayerHandlerMethodsRejectUnknownMethod(t *testing.T) {
	path := filepath.Join(t.TempDir(), "handler.go")
	if err := os.WriteFile(path, []byte(`package player
type Handler interface { HandleFuture(ctx *Context) }
`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := playerHandlerMethods(path)
	if err == nil || !strings.Contains(err.Error(), "unsupported player.Handler.HandleFuture method") {
		t.Fatalf("playerHandlerMethods() error = %v, want unsupported-method error", err)
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

func TestFormsUseGoAST(t *testing.T) {
	directory := t.TempDir()
	source := `package form
type Element interface { json.Marshaler; elem() }
type MenuElement interface { json.Marshaler; menuElem() }
type Form interface { json.Marshaler; SubmitJSON([]byte, Submitter, *world.Tx) error }
type Divider struct{}
type Header struct { Text string }
type Label struct { Text string }
type Input struct { Text, Default, Placeholder, Tooltip string; value string }
type Toggle struct { Text string; Default bool; Tooltip string; value bool }
type Slider struct { Text string; Min, Max float64; StepSize float64; Default float64; Tooltip string; value float64 }
type Dropdown struct { Text string; Options []string; DefaultIndex int; Tooltip string; value int }
type StepSlider Dropdown
type Button struct { Text, Image string }
type Custom struct{}
type Menu struct{}
type Modal struct{}
type Submittable interface { Submit(Submitter, *world.Tx) }
type MenuSubmittable interface { Submit(Submitter, Button, *world.Tx) }
type ModalSubmittable MenuSubmittable
type Closer interface { Close(Submitter, *world.Tx) }
type Submitter interface { SendForm(Form); CloseForm() }
func NewHeader(text string) Header { return Header{} }
func NewLabel(text string) Label { return Label{} }
func NewInput(text, defaultValue, placeholder string) Input { return Input{} }
func NewToggle(text string, defaultValue bool) Toggle { return Toggle{} }
func NewSlider(text string, min, max, stepSize, defaultValue float64) Slider { return Slider{} }
func NewDropdown(text string, options []string, defaultIndex int) Dropdown { return Dropdown{} }
func NewStepSlider(text string, options []string, defaultIndex int) StepSlider { return StepSlider{} }
func NewButton(text, image string) Button { return Button{} }
func New(Submittable, ...any) Custom { return Custom{} }
func NewMenu(MenuSubmittable, ...any) Menu { return Menu{} }
func NewModal(ModalSubmittable, ...any) Modal { return Modal{} }
func YesButton() Button { return Button{} }
func NoButton() Button { return Button{} }
func (Divider) elem() {}
func (Header) elem() {}
func (Label) elem() {}
func (Input) elem() {}
func (Toggle) elem() {}
func (Slider) elem() {}
func (Dropdown) elem() {}
func (StepSlider) elem() {}
func (Divider) menuElem() {}
func (Header) menuElem() {}
func (Label) menuElem() {}
func (Button) menuElem() {}
func (Divider) MarshalJSON() ([]byte, error) { return nil, nil }
func (Header) MarshalJSON() ([]byte, error) { return nil, nil }
func (Label) MarshalJSON() ([]byte, error) { return nil, nil }
func (Input) MarshalJSON() ([]byte, error) { return nil, nil }
func (Toggle) MarshalJSON() ([]byte, error) { return nil, nil }
func (Slider) MarshalJSON() ([]byte, error) { return nil, nil }
func (Dropdown) MarshalJSON() ([]byte, error) { return nil, nil }
func (StepSlider) MarshalJSON() ([]byte, error) { return nil, nil }
func (Button) MarshalJSON() ([]byte, error) { return nil, nil }
func (Input) WithTooltip(string) Input { return Input{} }
func (Input) Value() string { return "" }
func (Toggle) WithTooltip(string) Toggle { return Toggle{} }
func (Toggle) Value() bool { return false }
func (Slider) WithTooltip(string) Slider { return Slider{} }
func (Slider) Value() float64 { return 0 }
func (Dropdown) WithTooltip(string) Dropdown { return Dropdown{} }
func (Dropdown) Value() int { return 0 }
func (StepSlider) WithTooltip(string) StepSlider { return StepSlider{} }
func (StepSlider) Value() int { return 0 }
func (Custom) Title() string { return "" }
func (Custom) Elements() []Element { return nil }
func (Custom) SubmitJSON([]byte, Submitter, *world.Tx) error { return nil }
func (Custom) MarshalJSON() ([]byte, error) { return nil, nil }
func (Menu) WithBody(...any) Menu { return Menu{} }
func (Menu) AddButton(Button) Menu { return Menu{} }
func (Menu) AddDivider(Divider) Menu { return Menu{} }
func (Menu) AddHeader(Header) Menu { return Menu{} }
func (Menu) AddLabel(Label) Menu { return Menu{} }
func (Menu) WithButtons(...Button) Menu { return Menu{} }
func (Menu) WithElements(...MenuElement) Menu { return Menu{} }
func (Menu) Title() string { return "" }
func (Menu) Body() string { return "" }
func (Menu) Buttons() []Button { return nil }
func (Menu) Elements() []MenuElement { return nil }
func (Menu) SubmitJSON([]byte, Submitter, *world.Tx) error { return nil }
func (Menu) MarshalJSON() ([]byte, error) { return nil, nil }
func (Modal) WithBody(...any) Modal { return Modal{} }
func (Modal) Title() string { return "" }
func (Modal) Body() string { return "" }
func (Modal) Buttons() []Button { return nil }
func (Modal) SubmitJSON([]byte, Submitter, *world.Tx) error { return nil }
func (Modal) MarshalJSON() ([]byte, error) { return nil, nil }`
	if err := os.WriteFile(filepath.Join(directory, "form.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	spec, err := inspectForms(directory)
	if err != nil {
		t.Fatal(err)
	}
	output := string(generateForms(spec))
	for _, expected := range []string{
		"public interface Value\n    {\n        byte[] MarshalJSON();\n        void SubmitJSON(byte[]? response, Submitter submitter, World.Tx tx);\n    }",
		"public interface Element\n    {\n        byte[] MarshalJSON();\n    }",
		"public interface MenuElement\n    {\n        byte[] MarshalJSON();\n    }",
		"public interface ModalSubmittable : MenuSubmittable { }",
		"public struct Divider : Element, MenuElement",
		"public readonly byte[] MarshalJSON() => FormCodec.EncodeElement(this);",
		"public struct Input : Element",
		"public readonly Input WithTooltip(string tooltip)",
		"public readonly string Value() => _value ?? string.Empty;",
		"public static Slider NewSlider(string text, double min, double max, double stepSize, double defaultValue)",
		"public struct Button : MenuElement",
		"public readonly byte[] MarshalJSON() => FormCodec.EncodeMenuElement(this);",
		"public static Custom New(Submittable submittable, params object?[] title)",
		"public static Menu NewMenu(MenuSubmittable submittable, params object?[] title)",
		"public static Modal NewModal(ModalSubmittable submittable, params object?[] title)",
		"public Menu WithElements(params MenuElement[] elements)",
		"public static Button YesButton()",
		"public void SubmitJSON(byte[]? response, Submitter submitter, World.Tx tx)",
		"FormCodec.Respond(this, submitter, tx, response is null, response ?? Array.Empty<byte>());",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("generated forms missing %q:\n%s", expected, output)
		}
	}
}

func TestPlayerFormMethodsUseGoAST(t *testing.T) {
	path := filepath.Join(t.TempDir(), "player.go")
	source := `package player
func (p *Player) SendForm(f form.Form) {}
func (p *Player) CloseForm() {}`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	methods, err := inspectPlayerFormMethods(path)
	if err != nil {
		t.Fatal(err)
	}
	output := string(generatePlayerFormMethods(methods))
	for _, expected := range []string{
		"public sealed partial class Player : Form.Submitter",
		"public void SendForm(Form.Value f)",
		"PluginBridge.Host.SendPlayerForm(Invocation, Id, f);",
		"public void CloseForm()",
		"PluginBridge.Host.ClosePlayerForm(Invocation, Id);",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("generated player form surface missing %q:\n%s", expected, output)
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
func (tx *Tx) Thundering() bool { return false }
func (tx *Tx) CurrentTick() int64 { return 0 }
func (tx *Tx) AddParticle(pos mgl64.Vec3, p Particle) {}`
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
		"public interface Liquid : Block { string LiquidType(); }",
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
		"public long CurrentTick()",
		"PluginBridge.Host.WorldCurrentTick(Invocation)",
		"public interface Particle { }",
		"public void AddParticle(Vector3 pos, Particle p)",
		"PluginBridge.Host.AddWorldParticle(Invocation, pos, p)",
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
		Types: []blockTypeSpec{
			{Name: "Air", States: []encodedBlockState{{encodedBlock: encodedBlock{Name: "Air", Identifier: "minecraft:air", PropertiesNBT: []byte{10, 0, 0, 0}}}}},
			{Name: "Sand", Fields: []blockField{{Name: "Red", Type: "bool"}}, States: []encodedBlockState{
				{encodedBlock: encodedBlock{Name: "Sand", Identifier: "minecraft:sand", PropertiesNBT: []byte{10, 0, 0, 0}}, Values: []blockFieldValue{{}}},
				{encodedBlock: encodedBlock{Name: "Sand", Identifier: "minecraft:red_sand", PropertiesNBT: []byte{10, 0, 0, 0}}, Values: []blockFieldValue{{Bool: true}}},
			}},
		},
		Liquids: []liquidSpec{
			{Name: "Water", LiquidType: "water", States: []encodedLiquid{{
				encodedBlock: encodedBlock{Name: "Water", Identifier: "minecraft:water", PropertiesNBT: []byte{10, 1}},
				Still:        true, Depth: 8,
			}}},
			{Name: "Lava", LiquidType: "lava", States: []encodedLiquid{{
				encodedBlock: encodedBlock{Name: "Lava", Identifier: "minecraft:flowing_lava", PropertiesNBT: []byte{10, 2}},
				Depth:        4, Falling: true,
			}}},
		},
	}))
	for _, expected := range []string{
		"public readonly record struct Air : World.Block;",
		"public readonly record struct Sand(bool Red = false) : World.Block;",
		"public readonly record struct Water(bool Still, int Depth, bool Falling) : World.Liquid",
		`public string LiquidType() => "water";`,
		"public readonly record struct Lava(bool Still, int Depth, bool Falling) : World.Liquid",
		`public string LiquidType() => "lava";`,
		"internal static class BlockCodec",
		"case Block.Sand { Red: true }:",
		"case Block.Water { Still: true, Depth: 8, Falling: false }:",
		"internal static World.Liquid DecodeLiquid",
		"private sealed record EncodedBlock",
		"private sealed record EncodedLiquid",
		`public string LiquidType() => throw new InvalidOperationException("Opaque liquid type was not transported by the host.");`,
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

	particleOutput := string(generateParticles(particleSpec{
		Types: []particleType{
			{Name: "Flame", Kind: 0, Fields: []parameter{{Name: "Colour", Type: "Color.RGBA"}}},
			{Name: "PunchBlock", Kind: 3, Fields: []parameter{{Name: "Block", Type: "World.Block"}, {Name: "Face", Type: "Cube.Face"}}},
			{Name: "Note", Kind: 6, Fields: []parameter{{Name: "Instrument", Type: "Sound.Instrument"}, {Name: "Pitch", Type: "int"}}},
			{Name: "DragonEggTeleport", Kind: 7, Fields: []parameter{{Name: "Diff", Type: "Cube.Pos"}}},
			{Name: "EntityFlame", Kind: 19},
		},
		Instruments: []instrumentSpec{{Name: "Piano", ID: 0}, {Name: "Pling", ID: 15}},
		RGBAFields:  []parameter{{Name: "R", Type: "byte"}, {Name: "G", Type: "byte"}, {Name: "B", Type: "byte"}, {Name: "A", Type: "byte"}},
	}))
	for _, expected := range []string{
		"public readonly record struct RGBA(byte R, byte G, byte B, byte A);",
		"public readonly struct Instrument",
		"private readonly uint _id;",
		"public static Instrument Piano() => new(0u);",
		"public static Instrument Pling() => new(15u);",
		"public readonly record struct Flame(Color.RGBA Colour) : World.Particle;",
		"public readonly record struct PunchBlock(World.Block Block, Cube.Face Face) : World.Particle;",
		"public readonly record struct Note(Sound.Instrument Instrument, int Pitch) : World.Particle;",
		"public readonly record struct DragonEggTeleport(Cube.Pos Diff) : World.Particle;",
		"public readonly record struct EntityFlame : World.Particle;",
		"internal readonly record struct EncodedParticle(",
		"case Particle.PunchBlock value:",
		"encoded = new(3u, (uint)value.Face, 0, default, default, value.Block); return true;",
		"encoded = new(6u, value.Instrument.Id, value.Pitch, default, default, null); return true;",
		"internal static class ParticleCodec",
	} {
		if !strings.Contains(particleOutput, expected) {
			t.Fatalf("generated particle output missing %q:\n%s", expected, particleOutput)
		}
	}
	if strings.Contains(particleOutput, "public uint Kind") || strings.Contains(particleOutput, "public uint Id") || strings.Contains(particleOutput, "minecraft:") {
		t.Fatalf("particle transport leaks into public API:\n%s", particleOutput)
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

func TestInspectParticlesUsesCompleteASTSurface(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	directory := string(bytes.TrimSpace(output))
	spec, err := inspectParticles(
		filepath.Join(directory, "server", "world", "particle"),
		filepath.Join(directory, "server", "world", "sound", "instrument.go"),
		filepath.Join(runtime.GOROOT(), "src", "image", "color", "color.go"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Types) != 20 || len(spec.Types) != len(particleKindNames) {
		t.Fatalf("generated %d particle types", len(spec.Types))
	}
	for index, particle := range spec.Types {
		if particle.Name != particleKindNames[index] || particle.Kind != uint32(index) {
			t.Fatalf("particle %d = %+v", index, particle)
		}
	}
	if len(spec.Instruments) != 16 || len(spec.Instruments) != len(instrumentNames) {
		t.Fatalf("generated %d instruments", len(spec.Instruments))
	}
	for index, instrument := range spec.Instruments {
		if instrument.Name != instrumentNames[index] || instrument.ID != uint32(index) {
			t.Fatalf("instrument %d = %+v", index, instrument)
		}
	}
	if !reflect.DeepEqual(spec.RGBAFields, []parameter{{Name: "R", Type: "byte"}, {Name: "G", Type: "byte"}, {Name: "B", Type: "byte"}, {Name: "A", Type: "byte"}}) {
		t.Fatalf("RGBA fields = %+v", spec.RGBAFields)
	}
	generated := string(generateParticles(spec))
	if got := strings.Count(generated, ": World.Particle;"); got != 20 {
		t.Fatalf("generated output contains %d particle records", got)
	}
	for _, name := range particleKindNames {
		if !strings.Contains(generated, "record struct "+name) {
			t.Fatalf("generated output missing particle.%s", name)
		}
	}
}

func TestInspectParticlesRejectsFieldDrift(t *testing.T) {
	directory := t.TempDir()
	var source strings.Builder
	source.WriteString("package particle\ntype particle struct{}\n")
	for _, name := range particleKindNames {
		if name == "Flame" {
			source.WriteString("type Flame struct { particle; colour color.RGBA }\n")
		} else {
			fmt.Fprintf(&source, "type %s struct { particle }\n", name)
		}
	}
	if err := os.WriteFile(filepath.Join(directory, "particle.go"), []byte(source.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := inspectParticles(directory, "unused", "unused")
	if err == nil || !strings.Contains(err.Error(), "field colour is not exported") {
		t.Fatalf("expected particle field drift error, got %v", err)
	}
}

func TestInspectInstrumentsRejectsASTDrift(t *testing.T) {
	var source strings.Builder
	source.WriteString("package sound\ntype Instrument struct { instrument }\ntype instrument int32\n")
	for id, name := range instrumentNames {
		if name == "Pling" {
			id++
		}
		fmt.Fprintf(&source, "func %s() Instrument { return Instrument{%d} }\n", name, id)
	}
	path := filepath.Join(t.TempDir(), "instrument.go")
	if err := os.WriteFile(path, []byte(source.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := inspectInstruments(path)
	if err == nil || !strings.Contains(err.Error(), "sound.Pling is not Instrument{15}") {
		t.Fatalf("expected instrument ID drift error, got %v", err)
	}
}

func TestInspectGameModesUsesASTAndLiveRegistry(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	directory := string(bytes.TrimSpace(output))
	spec, err := inspectGameModes(filepath.Join(directory, "server", "world", "game_mode.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(spec.Methods, gameModeMethodNames) || len(spec.Modes) != 4 {
		t.Fatalf("game mode spec = %+v", spec)
	}
	wantMasks := []uint64{0x6b, 0xfd, 0x6a, 0x10}
	for index, mode := range spec.Modes {
		if mode.Name != gameModeVariableNames[index] || mode.ID != index || gameModeCapabilityMask(mode.Capabilities) != wantMasks[index] {
			t.Fatalf("game mode %d = %+v", index, mode)
		}
	}
	generated := string(generateGameModes(spec))
	for _, expected := range []string{
		"public interface GameMode",
		"public static readonly GameMode GameModeSurvival = new BuiltinGameMode(0, 0x6bUL);",
		"public static readonly GameMode GameModeSpectator = new BuiltinGameMode(3, 0x10UL);",
		"public static (GameMode GameMode, bool Ok) GameModeByID(int id)",
		"_ => (GameModeSurvival, false)",
		"public static (int ID, bool Ok) GameModeID(GameMode mode)",
		"internal static long GameModeDescriptor(GameMode mode)",
		"BuiltinGameModeFlag | (uint)builtin.ID",
		"if (mode.InstantPortalTravel()) capabilities |= 1UL << 7;",
		"internal static GameMode GameModeFromDescriptor(long descriptor)",
		"if ((value & ~CustomGameModeMask) != 0)",
		"if (!ok) throw new InvalidOperationException(\"invalid game mode descriptor\");",
		"private sealed class BuiltinGameMode",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated game mode output missing %q:\n%s", expected, generated)
		}
	}
	playerMethods, err := inspectPlayerGameModeMethods(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		t.Fatal(err)
	}
	playerOutput := string(generatePlayerGameModes(playerMethods))
	for _, expected := range []string{
		"public void SetGameMode(World.GameMode mode)",
		"Integer = World.GameModeDescriptor(mode)",
		"public World.GameMode GameMode() => PluginBridge.Host.PlayerGameMode(_invocation, Id);",
	} {
		if !strings.Contains(playerOutput, expected) {
			t.Fatalf("generated player game mode output missing %q:\n%s", expected, playerOutput)
		}
	}
}

func TestInspectGameModesRejectsASTDrift(t *testing.T) {
	for name, test := range map[string]struct {
		mutate   func(string) string
		expected string
	}{
		"interface": {
			mutate: func(source string) string {
				return strings.Replace(source, "AllowsEditing() bool", "CanEdit() bool", 1)
			},
			expected: "world.GameMode methods changed",
		},
		"registry": {
			mutate: func(source string) string {
				return strings.Replace(source, "0: GameModeSurvival", "0: GameModeCreative", 1)
			},
			expected: "invalid ID/name",
		},
		"lookup": {
			mutate: func(source string) string {
				return strings.Replace(source, "GameModeByID(id int)", "GameModeByID(id int64)", 1)
			},
			expected: "GameModeByID signature changed",
		},
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "game_mode.go")
			if err := os.WriteFile(path, []byte(test.mutate(gameModeFixtureSource())), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := inspectGameModes(path)
			if err == nil || !strings.Contains(err.Error(), test.expected) {
				t.Fatalf("expected %q drift error, got %v", test.expected, err)
			}
		})
	}
}

func gameModeFixtureSource() string {
	return `package world
type GameMode interface {
	AllowsEditing() bool
	AllowsTakingDamage() bool
	CreativeInventory() bool
	HasCollision() bool
	AllowsFlying() bool
	AllowsInteraction() bool
	Visible() bool
	InstantPortalTravel() bool
}
var (
	GameModeSurvival survival
	GameModeCreative creative
	GameModeAdventure adventure
	GameModeSpectator spectator
)
var gameModeReg = newGameModeRegistry(map[int]GameMode{
	0: GameModeSurvival,
	1: GameModeCreative,
	2: GameModeAdventure,
	3: GameModeSpectator,
})
func GameModeByID(id int) (GameMode, bool) { return nil, false }
func GameModeID(mode GameMode) (int, bool) { return 0, false }
type survival struct{}
type creative struct{}
type adventure struct{}
type spectator struct{}
`
}

func TestInspectPlayerGameModeMethodsRejectsDrift(t *testing.T) {
	for name, source := range map[string]string{
		"setter parameter": `package player
func (p *Player) SetGameMode(value world.GameMode) {}
func (p *Player) GameMode() world.GameMode { return nil }`,
		"getter result": `package player
func (p *Player) SetGameMode(mode world.GameMode) {}
func (p *Player) GameMode() int { return 0 }`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "player.go")
			if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := inspectPlayerGameModeMethods(path)
			if err == nil || !strings.Contains(err.Error(), "signature changed") {
				t.Fatalf("expected player game mode signature drift error, got %v", err)
			}
		})
	}
}

func TestInspectWorldTxRejectsAddParticleDrift(t *testing.T) {
	for name, declaration := range map[string]string{
		"position name": "func (tx *Tx) AddParticle(position mgl64.Vec3, p Particle) {}",
		"position type": "func (tx *Tx) AddParticle(pos cube.Pos, p Particle) {}",
		"particle type": "func (tx *Tx) AddParticle(pos mgl64.Vec3, p Block) {}",
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tx.go")
			if err := os.WriteFile(path, []byte("package world\n"+declaration), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := inspectWorldTx(path)
			if err == nil || !strings.Contains(err.Error(), "world.Tx.AddParticle: signature changed") {
				t.Fatalf("expected AddParticle signature drift error, got %v", err)
			}
		})
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
