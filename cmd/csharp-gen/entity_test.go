package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const entityHandleSource = `
func (e *EntityHandle) Type() EntityType { return nil }
func (e *EntityHandle) Entity(tx *Tx) (Entity, bool) { return nil, false }
func (e *EntityHandle) UUID() uuid.UUID { return uuid.Nil }
func (e *EntityHandle) Closed() bool { return false }
func (e *EntityHandle) Close() error { return nil }
`

const entityConstructionSource = `
type EntityType interface {
    Open(tx *Tx, handle *EntityHandle, data *EntityData) Entity
    EncodeEntity() string
    BBox(e Entity) cube.BBox
    DecodeNBT(m map[string]any, data *EntityData)
    EncodeNBT(data *EntityData) map[string]any
}
type EntityConfig interface { Apply(data *EntityData) }
type TickerEntity interface {
    Entity
    Tick(tx *Tx, current int64)
}
type EntityData struct {
    Pos, Vel mgl64.Vec3
    Rot cube.Rotation
    Name string
    FireDuration time.Duration
    Age time.Duration
    Data any
}
type EntitySpawnOpts struct {
    Position mgl64.Vec3
    Rotation cube.Rotation
    Velocity mgl64.Vec3
    ID uuid.UUID
    NameTag string
}
func (opts EntitySpawnOpts) New(t EntityType, conf EntityConfig) *EntityHandle { return nil }
func NewEntity(t EntityType, conf EntityConfig) *EntityHandle { return nil }
`

func TestWorldEntityUsesGoAST(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "entity.go")
	source := `package world
import (
    "io"
    "github.com/go-gl/mathgl/mgl64"
    "github.com/df-mc/dragonfly/server/block/cube"
)
type Entity interface {
    io.Closer
    H() *EntityHandle
    Position() mgl64.Vec3
    Rotation() cube.Rotation
}` + entityConstructionSource + entityHandleSource
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	methods, err := inspectWorldEntity(path)
	if err != nil {
		t.Fatal(err)
	}
	handleMethods, err := inspectWorldEntityHandle(path)
	if err != nil {
		t.Fatal(err)
	}
	construction, err := inspectWorldEntityConstruction(path)
	if err != nil {
		t.Fatal(err)
	}
	generated := string(generateWorldEntity(methods, handleMethods, construction))
	for _, expected := range []string{
		"public interface EntityType",
		"Entity Open(Tx tx, EntityHandle handle, EntityData data);",
		"string EncodeEntity();",
		"Cube.BBox BBox(Entity e);",
		"void DecodeNBT(Dictionary<string, object?> m, EntityData data);",
		"Dictionary<string, object?> EncodeNBT(EntityData data);",
		"public interface EntityConfig",
		"void Apply(EntityData data);",
		"public interface TickerEntity : Entity",
		"void Tick(Tx tx, long current);",
		"public sealed class EntityData",
		"public Vector3 Pos = default;",
		"public Vector3 Vel = default;",
		"public Rotation Rot = default;",
		"public string Name = string.Empty;",
		"public TimeSpan FireDuration = default;",
		"public TimeSpan Age = default;",
		"public object? Data;",
		"public sealed class EntitySpawnOpts",
		"public Vector3 Position = default;",
		"public Rotation Rotation = default;",
		"public Vector3 Velocity = default;",
		"public Guid ID = default;",
		"public string NameTag = string.Empty;",
		"public EntityHandle New(EntityType t, EntityConfig conf)",
		"PluginBridge.Host.NewEntity(this, t, conf)",
		"public static EntityHandle NewEntity(EntityType t, EntityConfig conf)",
		"new EntitySpawnOpts().New(t, conf)",
		"void Close();",
		"EntityHandle H();",
		"Vector3 Position();",
		"Rotation Rotation();",
		"public (Entity? Entity, bool Ok) Entity(Tx tx)",
		"PluginBridge.Host.EntityHandleEntity(tx.Invocation, this)",
		"public Guid UUID()",
		"PluginBridge.Host.EntityHandleUuid(this)",
		"public bool Closed()",
		"PluginBridge.Host.EntityHandleClosed(this)",
		"public void Close()",
		"PluginBridge.Host.CloseEntityHandle(this)",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated entity interface missing %q:\n%s", expected, generated)
		}
	}
}

func TestPinnedDragonflyEntityHandleHasExactSurface(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	directory := string(bytes.TrimSpace(output))
	methods, err := inspectWorldEntityHandle(filepath.Join(directory, "server", "world", "entity.go"))
	if err != nil {
		t.Fatal(err)
	}
	want := []commandMethod{
		{Name: "Entity", Parameters: []parameter{{Name: "tx", Type: "Tx"}}, ReturnType: "(Entity? Entity, bool Ok)"},
		{Name: "UUID", ReturnType: "Guid"},
		{Name: "Closed", ReturnType: "bool"},
		{Name: "Close", ReturnType: "void"},
	}
	if len(methods) != len(want) {
		t.Fatalf("world.EntityHandle has %d methods, want %d", len(methods), len(want))
	}
	for index := range want {
		if methods[index].Name != want[index].Name || methods[index].ReturnType != want[index].ReturnType ||
			!equalParameters(methods[index].Parameters, want[index].Parameters) {
			t.Fatalf("world.EntityHandle method %d = %+v, want %+v", index, methods[index], want[index])
		}
	}
}

func equalParameters(left, right []parameter) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func TestPinnedDragonflyWorldEntityHasExactSurface(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	directory := string(bytes.TrimSpace(output))
	methods, err := inspectWorldEntity(filepath.Join(directory, "server", "world", "entity.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(methods) != 4 {
		t.Fatalf("world.Entity has %d methods, want 4", len(methods))
	}
	want := []entityMethod{
		{Name: "Close", ReturnType: "void"},
		{Name: "H", ReturnType: "EntityHandle"},
		{Name: "Position", ReturnType: "Vector3"},
		{Name: "Rotation", ReturnType: "Rotation"},
	}
	for index := range want {
		if methods[index] != want[index] {
			t.Fatalf("world.Entity method %d = %+v, want %+v", index, methods[index], want[index])
		}
	}
}

func TestPinnedDragonflyEntityConstructionHasExactSurface(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(string(bytes.TrimSpace(output)), "server", "world", "entity.go")
	spec, err := inspectWorldEntityConstruction(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.EntityTypeMethods) != 5 || len(spec.EntityConfigMethods) != 1 || len(spec.TickerMethods) != 1 ||
		len(spec.EntityDataFields) != 7 || len(spec.SpawnFields) != 5 {
		t.Fatalf("unexpected pinned entity construction surface: %+v", spec)
	}
}

func TestEntityConstructionRejectsDrift(t *testing.T) {
	tests := map[string][2]string{
		"entity type method": {"EncodeEntity() string", "EncodeEntity() []byte"},
		"entity config":      {"Apply(data *EntityData)", "Apply(data EntityData)"},
		"entity data field":  {"Age time.Duration", "Age int64"},
		"spawn field":        {"ID uuid.UUID", "ID string"},
		"spawn receiver":     {"func (opts EntitySpawnOpts) New", "func (opts *EntitySpawnOpts) New"},
		"spawn new":          {"conf EntityConfig) *EntityHandle", "conf *EntityConfig) *EntityHandle"},
		"package new":        {"func NewEntity(t EntityType", "func NewEntity(t *EntityType"},
	}
	for name, replacement := range tests {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "entity.go")
			source := "package world\n" + strings.Replace(entityConstructionSource, replacement[0], replacement[1], 1)
			if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := inspectWorldEntityConstruction(path); err == nil {
				t.Fatal("expected entity construction drift error")
			}
		})
	}
}

func TestGeneratedEntityConstructionCompilesForPlugin(t *testing.T) {
	if _, err := exec.LookPath("dotnet"); err != nil {
		t.Skip("dotnet is not installed")
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	project := fmt.Sprintf(`<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup><TargetFramework>net10.0</TargetFramework><Nullable>enable</Nullable></PropertyGroup>
  <ItemGroup><ProjectReference Include="%s" /></ItemGroup>
</Project>
`, filepath.ToSlash(filepath.Join(root, "csharp", "Dragonfly", "Dragonfly.csproj")))
	source := `using Dragonfly;
using System;
using System.Collections.Generic;

public sealed class ExampleEntityType : World.EntityType
{
    public World.Entity Open(World.Tx tx, World.EntityHandle handle, World.EntityData data) =>
        new ExampleEntity(handle, data);
    public string EncodeEntity() => "example:entity";
    public Cube.BBox BBox(World.Entity e) => Cube.Box(-.25, 0, -.25, .25, 1, .25);
    public void DecodeNBT(Dictionary<string, object?> m, World.EntityData data) => data.Data = m;
    public Dictionary<string, object?> EncodeNBT(World.EntityData data) => new();
}

public sealed class ExampleEntityConfig : World.EntityConfig
{
    public void Apply(World.EntityData data) => data.Data = "configured";
}

public sealed class ExampleEntity(World.EntityHandle handle, World.EntityData data) : World.Entity
{
    public void Close() => handle.Close();
    public World.EntityHandle H() => handle;
    public Vector3 Position() => data.Pos;
    public Rotation Rotation() => data.Rot;
}

public static class ExampleFactory
{
    public static World.EntityHandle Spawn() => new World.EntitySpawnOpts
    {
        Position = new Vector3(1, 2, 3),
        Rotation = new Rotation(90, 0),
        Velocity = new Vector3(0, 1, 0),
        ID = Guid.NewGuid(),
        NameTag = "Example",
    }.New(new ExampleEntityType(), new ExampleEntityConfig());

    public static World.EntityHandle SpawnDefault() =>
        World.NewEntity(new ExampleEntityType(), new ExampleEntityConfig());

}
`
	for name, content := range map[string]string{"EntityPlugin.csproj": project, "Plugin.cs": source} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.Command("dotnet", "build", filepath.Join(directory, "EntityPlugin.csproj"), "--nologo")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("generated entity construction plugin failed to compile: %v\n%s", err, output)
	}
}
