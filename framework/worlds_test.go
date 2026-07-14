package framework

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
)

func TestWorldManagerRegistersCoreWorlds(t *testing.T) {
	manager := NewWorldManager()
	overworld := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = overworld.Close() })
	if err := manager.RegisterCore(OverworldID, overworld); err != nil {
		t.Fatal(err)
	}
	if _, ok := overworld.Handler().(*host.WorldHandler); !ok {
		t.Fatalf("handler type = %T", overworld.Handler())
	}
	if err := manager.Unload(OverworldID); err == nil {
		t.Fatal("core world unload succeeded")
	}
}

func TestWorldManagerCreatesAndUnloadsCustomWorld(t *testing.T) {
	manager := NewWorldManager()
	w, err := manager.Create("example:lobby", world.Config{Synchronous: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := w.Handler().(*host.WorldHandler); !ok {
		t.Fatalf("handler type = %T", w.Handler())
	}
	if got := manager.IDs(); !slices.Equal(got, []WorldID{"example:lobby"}) {
		t.Fatalf("IDs = %v", got)
	}
	if err := manager.Save("example:lobby"); err != nil {
		t.Fatal(err)
	}
	if err := manager.Unload("example:lobby"); err != nil {
		t.Fatal(err)
	}
	if _, ok := manager.World("example:lobby"); ok {
		t.Fatal("unloaded world remains registered")
	}
}

func TestWorldManagerRejectsInvalidOrDuplicateIDs(t *testing.T) {
	manager := NewWorldManager()
	if _, err := manager.Create("minecraft:custom", world.Config{Synchronous: true}); err == nil {
		t.Fatal("reserved namespace accepted")
	}
	if _, err := manager.Create("missing_namespace", world.Config{Synchronous: true}); err == nil {
		t.Fatal("invalid ID accepted")
	}
	if _, err := manager.Create("example:arena", world.Config{Synchronous: true}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	if _, err := manager.Create("example:arena", world.Config{Synchronous: true}); err == nil {
		t.Fatal("duplicate ID accepted")
	}
}

func TestWorldManagerNativeHandlesAreStableAndNeverReused(t *testing.T) {
	manager := NewWorldManager()
	first, err := manager.Create("example:first", world.Config{Synchronous: true})
	if err != nil {
		t.Fatal(err)
	}
	firstID, ok := manager.WorldByName(0, "example:first")
	if !ok {
		t.Fatal("first world missing")
	}
	if name, ok := manager.WorldName(0, firstID); !ok || name != "example:first" {
		t.Fatalf("WorldName() = %q, %v", name, ok)
	}
	if err := manager.Unload("example:first"); err != nil {
		t.Fatal(err)
	}
	_ = first
	if _, ok := manager.WorldName(0, firstID); ok {
		t.Fatal("stale handle still resolves")
	}
	if _, err := manager.Create("example:second", world.Config{Synchronous: true}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	secondID, _ := manager.WorldByName(0, "example:second")
	if secondID <= firstID {
		t.Fatalf("handle reused: first=%d second=%d", firstID, secondID)
	}
}

func TestWorldManagerBlockAndStateOperationsUseActiveTx(t *testing.T) {
	players := host.NewPlayers()
	manager := newWorldManager("", nil, players)
	w, err := manager.Create("example:blocks", world.Config{Synchronous: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	id, _ := manager.WorldByName(0, "example:blocks")
	if err := w.Do(func(tx *world.Tx) {
		invocation, leave := players.BeginInvocation(tx)
		defer leave()
		worldRange, ok := manager.WorldRange(invocation, 0)
		if !ok || worldRange.Min != int32(tx.Range().Min()) || worldRange.Max != int32(tx.Range().Max()) {
			t.Fatalf("WorldRange() = %#v, %v", worldRange, ok)
		}
		properties, ok := encodeBlockProperties(map[string]any{
			"bool": true, "byte": uint8(4), "int": int32(-9), "string": "north",
		})
		if !ok {
			t.Fatal("encode block properties failed")
		}
		decoded, ok := decodeBlockProperties(properties)
		if !ok || decoded["bool"] != true || decoded["byte"] != uint8(4) || decoded["int"] != int32(-9) || decoded["string"] != "north" {
			t.Fatalf("decoded properties = %#v", decoded)
		}
		gold, ok := w.BlockRegistry().BlockByName("minecraft:gold_block", nil)
		if !ok {
			t.Fatal("gold block is not registered")
		}
		name, state := gold.EncodeBlock()
		stateNBT, ok := encodeBlockProperties(state)
		options := native.WorldSetOpts{
			DisableBlockUpdates: true, DisableLiquidDisplacement: true, DisableRedstoneUpdates: true,
		}
		if !ok || !manager.SetWorldBlock(invocation, 0, native.BlockPos{X: 2, Y: 3, Z: 4}, native.WorldBlock{Identifier: name, PropertiesNBT: stateNBT}, options) {
			t.Fatal("SetWorldBlock() failed")
		}
		if !manager.ScheduleWorldBlockUpdate(invocation, 0, native.BlockPos{X: 2, Y: 3, Z: 4}, native.WorldBlock{
			Identifier: name, PropertiesNBT: stateNBT,
		}, int64(time.Hour)) {
			t.Fatal("ScheduleWorldBlockUpdate() failed")
		}
		got, ok := manager.WorldBlock(invocation, 0, native.BlockPos{X: 2, Y: 3, Z: 4})
		if !ok || got.Identifier != name {
			t.Fatalf("WorldBlock() = %#v, %v", got, ok)
		}
		loadedBlock, loaded, valid := manager.WorldBlockLoaded(invocation, 0, native.BlockPos{X: 2, Y: 3, Z: 4})
		if !valid || !loaded || loadedBlock.Identifier != name {
			t.Fatalf("WorldBlockLoaded() = %#v, %v, %v", loadedBlock, loaded, valid)
		}
		if _, loaded, valid := manager.WorldBlockLoaded(invocation, 0, native.BlockPos{X: 30_000_000, Y: 3, Z: 30_000_000}); !valid || loaded {
			t.Fatalf("WorldBlockLoaded(unloaded) = loaded %v, valid %v", loaded, valid)
		}
		iterator, ok := manager.OpenWorldBlocksWithin(invocation, 0, native.BlockPos{X: 2, Y: 3, Z: 4}, 1, []native.WorldBlock{{Identifier: name, PropertiesNBT: stateNBT}})
		if !ok || iterator == 0 {
			t.Fatal("OpenWorldBlocksWithin() failed")
		}
		position, found, valid := manager.NextWorldBlock(invocation, iterator)
		if !valid || !found || position != (native.BlockPos{X: 2, Y: 3, Z: 4}) {
			t.Fatalf("NextWorldBlock() = %#v, %v, %v", position, found, valid)
		}
		if _, found, valid := manager.NextWorldBlock(invocation, iterator); !valid || found {
			t.Fatalf("NextWorldBlock(end) = found %v, valid %v", found, valid)
		}
		if _, _, valid := manager.NextWorldBlock(invocation, iterator); valid {
			t.Fatal("finished block iterator remained live")
		}
		closedIterator, ok := manager.OpenWorldBlocksWithin(invocation, 0, native.BlockPos{X: 2, Y: 3, Z: 4}, 1, []native.WorldBlock{{Identifier: name, PropertiesNBT: stateNBT}})
		if !ok {
			t.Fatal("second OpenWorldBlocksWithin() failed")
		}
		manager.CloseWorldBlocks(invocation, closedIterator)
		if _, _, valid := manager.NextWorldBlock(invocation, closedIterator); valid {
			t.Fatal("closed block iterator remained live")
		}
		if got, ok := manager.WorldHighestLightBlocker(invocation, 0, 2, 4); !ok || got != int32(tx.HighestLightBlocker(2, 4)) {
			t.Fatalf("WorldHighestLightBlocker() = %d, %v", got, ok)
		}
		if got, ok := manager.WorldHighestBlock(invocation, 0, 2, 4); !ok || got != int32(tx.HighestBlock(2, 4)) {
			t.Fatalf("WorldHighestBlock() = %d, %v", got, ok)
		}
		blockPosition := native.BlockPos{X: 2, Y: 3, Z: 4}
		if got, ok := manager.WorldLight(invocation, 0, blockPosition); !ok || got != tx.Light(cube.Pos{2, 3, 4}) {
			t.Fatalf("WorldLight() = %d, %v", got, ok)
		}
		if got, ok := manager.WorldSkyLight(invocation, 0, blockPosition); !ok || got != tx.SkyLight(cube.Pos{2, 3, 4}) {
			t.Fatalf("WorldSkyLight() = %d, %v", got, ok)
		}
		plains, ok := world.BiomeByName("plains")
		if !ok {
			t.Fatal("plains biome is not registered")
		}
		biomeID := plains.EncodeBiome()
		if biomeID < math.MinInt32 || biomeID > math.MaxInt32 || !manager.SetWorldBiome(invocation, 0, blockPosition, int32(biomeID)) {
			t.Fatal("SetWorldBiome() failed")
		}
		if got, ok := manager.WorldBiome(invocation, 0, blockPosition); !ok || got != int32(biomeID) {
			t.Fatalf("WorldBiome() = %d, %v", got, ok)
		}
		if manager.SetWorldBiome(invocation, 0, blockPosition, math.MaxInt32) {
			t.Fatal("SetWorldBiome() accepted an unknown biome")
		}
		if got, ok := manager.WorldTemperature(invocation, 0, blockPosition); !ok || got != tx.Temperature(cube.Pos{2, 3, 4}) {
			t.Fatalf("WorldTemperature() = %f, %v", got, ok)
		}
		checkWeather := func(name string, got, valid, want bool) {
			if !valid || got != want {
				t.Fatalf("World%s() = %v, %v, want %v", name, got, valid, want)
			}
		}
		gotWeather, validWeather := manager.WorldRainingAt(invocation, 0, blockPosition)
		checkWeather("RainingAt", gotWeather, validWeather, tx.RainingAt(cube.Pos{2, 3, 4}))
		gotWeather, validWeather = manager.WorldSnowingAt(invocation, 0, blockPosition)
		checkWeather("SnowingAt", gotWeather, validWeather, tx.SnowingAt(cube.Pos{2, 3, 4}))
		gotWeather, validWeather = manager.WorldThunderingAt(invocation, 0, blockPosition)
		checkWeather("ThunderingAt", gotWeather, validWeather, tx.ThunderingAt(cube.Pos{2, 3, 4}))
		gotWeather, validWeather = manager.WorldRaining(invocation, 0)
		checkWeather("Raining", gotWeather, validWeather, tx.Raining())
		gotWeather, validWeather = manager.WorldThundering(invocation, 0)
		checkWeather("Thundering", gotWeather, validWeather, tx.Thundering())
		bars := block.IronBars{}
		tx.SetBlock(cube.Pos{2, 3, 4}, bars, nil)
		waterName, waterState := (block.Water{Still: true, Depth: 8}).EncodeBlock()
		waterNBT, ok := encodeBlockProperties(waterState)
		water := native.WorldBlock{Identifier: waterName, PropertiesNBT: waterNBT}
		if !ok || !manager.SetWorldLiquid(invocation, id, blockPosition, &water) {
			t.Fatal("SetWorldLiquid() failed")
		}
		liquid, found, valid := manager.WorldLiquid(invocation, id, blockPosition)
		if !valid || !found || liquid.Identifier != "minecraft:water" {
			t.Fatalf("WorldLiquid() = %#v, %v, %v", liquid, found, valid)
		}
		if !manager.SetWorldLiquid(invocation, id, blockPosition, nil) {
			t.Fatal("SetWorldLiquid(nil) failed")
		}
		if _, found, valid := manager.WorldLiquid(invocation, id, blockPosition); !valid || found {
			t.Fatalf("WorldLiquid(removed) = found %v, valid %v", found, valid)
		}
		if !manager.SetWorldLiquid(invocation, id, blockPosition, &water) {
			t.Fatal("second SetWorldLiquid() failed")
		}
		foreground, ok := manager.WorldBlock(invocation, id, native.BlockPos{X: 2, Y: 3, Z: 4})
		barsName, _ := bars.EncodeBlock()
		if !ok || foreground.Identifier != barsName {
			t.Fatalf("waterlogged foreground = %#v, %v", foreground, ok)
		}
		if manager.SaveWorld(invocation, id) {
			t.Fatal("SaveWorld succeeded from owner transaction")
		}
		leakedIterator, ok := manager.OpenWorldBlocksWithin(invocation, 0, native.BlockPos{X: 2, Y: 3, Z: 4}, 1, []native.WorldBlock{{Identifier: name, PropertiesNBT: stateNBT}})
		if !ok {
			t.Fatal("leaked OpenWorldBlocksWithin() failed")
		}
		leave()
		if _, _, valid := manager.NextWorldBlock(invocation, leakedIterator); valid {
			t.Fatal("iterator survived invocation cleanup")
		}
		manager.blockIteratorMu.Lock()
		remainingIterators := len(manager.blockIterators)
		manager.blockIteratorMu.Unlock()
		if remainingIterators != 0 {
			t.Fatalf("block iterators after invocation = %d", remainingIterators)
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := manager.WorldBlock(9999, 0, native.BlockPos{}); ok {
		t.Fatal("stale invocation resolved current world")
	}
	if _, _, valid := manager.WorldBlockLoaded(9999, 0, native.BlockPos{}); valid {
		t.Fatal("stale invocation resolved a loaded block")
	}
	if _, ok := manager.WorldRange(9999, 0); ok {
		t.Fatal("stale invocation resolved a world range")
	}
	if _, ok := manager.WorldBiome(9999, 0, native.BlockPos{}); ok {
		t.Fatal("stale invocation resolved a biome")
	}
	if manager.SetWorldBiome(9999, 0, native.BlockPos{}, 1) {
		t.Fatal("stale invocation set a biome")
	}
	if _, ok := manager.WorldTemperature(9999, 0, native.BlockPos{}); ok {
		t.Fatal("stale invocation resolved temperature")
	}
	if _, ok := manager.WorldRainingAt(9999, 0, native.BlockPos{}); ok {
		t.Fatal("stale invocation resolved position weather")
	}
	if _, ok := manager.WorldSnowingAt(9999, 0, native.BlockPos{}); ok {
		t.Fatal("stale invocation resolved snow")
	}
	if _, ok := manager.WorldThunderingAt(9999, 0, native.BlockPos{}); ok {
		t.Fatal("stale invocation resolved position thunder")
	}
	if _, ok := manager.WorldRaining(9999, 0); ok {
		t.Fatal("stale invocation resolved world weather")
	}
	if _, ok := manager.WorldThundering(9999, 0); ok {
		t.Fatal("stale invocation resolved world thunder")
	}
	if !manager.SetWorldTime(0, id, 6000) {
		t.Fatal("SetWorldTime failed")
	}
	if got, ok := manager.WorldTime(0, id); !ok || got != 6000 {
		t.Fatalf("WorldTime() = %d, %v", got, ok)
	}
	spawn := native.BlockPos{X: 8, Y: 70, Z: -3}
	if !manager.SetWorldSpawn(0, id, spawn) {
		t.Fatal("SetWorldSpawn failed")
	}
	if got, ok := manager.WorldSpawn(0, id); !ok || got != spawn {
		t.Fatalf("WorldSpawn() = %#v, %v", got, ok)
	}
}

func TestWorldManagerInvocationCannotBorrowAnotherWorldTransaction(t *testing.T) {
	players := host.NewPlayers()
	manager := newWorldManager("", nil, players)
	first, err := manager.Create("example:first", world.Config{Synchronous: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create("example:second", world.Config{Synchronous: true}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	firstID, _ := manager.WorldByName(0, "example:first")
	secondID, _ := manager.WorldByName(0, "example:second")
	position := native.BlockPos{X: 1, Y: 2, Z: 3}
	var stale native.InvocationID
	var scheduled native.WorldBlock
	if err := first.Do(func(tx *world.Tx) {
		invocation, leave := players.BeginInvocation(tx)
		stale = invocation
		gold, ok := first.BlockRegistry().BlockByName("minecraft:gold_block", nil)
		if !ok {
			t.Fatal("gold block is not registered")
		}
		name, properties := gold.EncodeBlock()
		propertiesNBT, ok := encodeBlockProperties(properties)
		if !ok {
			t.Fatal("encode gold block properties")
		}
		scheduled = native.WorldBlock{Identifier: name, PropertiesNBT: propertiesNBT}
		if _, ok := manager.WorldBlock(invocation, secondID, position); ok {
			t.Fatal("cross-world synchronous read succeeded")
		}
		if _, _, valid := manager.WorldLiquid(invocation, secondID, position); valid {
			t.Fatal("cross-world synchronous liquid read succeeded")
		}
		tx.SetLiquid(cube.Pos{1, 2, 3}, block.Water{Still: true, Depth: 8})
		if _, found, valid := manager.WorldLiquid(invocation, firstID, position); !valid || !found {
			t.Fatal("same-world liquid read did not use invocation transaction")
		}
		if _, ok := manager.WorldBlock(invocation, firstID, position); !ok {
			t.Fatal("same-world read did not use invocation transaction")
		}
		if !manager.ScheduleWorldBlockUpdate(invocation, firstID, position, scheduled, int64(time.Second)) {
			t.Fatal("same-world scheduled update did not use invocation transaction")
		}
		if !manager.ScheduleWorldBlockUpdate(invocation, secondID, position, scheduled, int64(time.Second)) {
			t.Fatal("cross-world scheduled update was not safely queued")
		}
		leave()
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := manager.WorldBlock(stale, firstID, position); ok {
		t.Fatal("stale invocation still resolves")
	}
	if _, _, valid := manager.WorldLiquid(stale, firstID, position); valid {
		t.Fatal("stale invocation still resolves liquid")
	}
	if manager.ScheduleWorldBlockUpdate(stale, firstID, position, scheduled, int64(time.Second)) {
		t.Fatal("stale invocation scheduled a block update")
	}
}

func TestWorldManagerUnloadRejectsPlayersAndCurrentOwner(t *testing.T) {
	players := host.NewPlayers()
	manager := newWorldManager("", nil, players)
	w, err := manager.Create("example:occupied", world.Config{Synchronous: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	id, _ := manager.WorldByName(0, "example:occupied")
	if err := w.Do(func(tx *world.Tx) {
		invocation, leave := players.BeginInvocation(tx)
		defer leave()
		if manager.UnloadWorld(invocation, id) {
			t.Fatal("owner unload succeeded")
		}
		playerID := uuid.MustParse("4f62ee78-9519-4f1c-b0bd-69f57b578daf")
		handle := world.EntitySpawnOpts{ID: playerID, Position: mgl64.Vec3{}}.New(
			player.Type, player.Config{UUID: playerID, Name: "Occupant"},
		)
		tx.AddEntity(handle)
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Unload("example:occupied"); err == nil || !strings.Contains(err.Error(), "contains 1 player") {
		t.Fatalf("occupied unload error = %v", err)
	}
}

func TestPersistentWorldManagerKeepsWorldsBelowRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "custom-worlds")
	manager, err := NewPersistentWorldManager(root, nil, host.NewPlayers())
	if err != nil {
		t.Fatal(err)
	}
	core := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = core.Close() })
	if err := manager.RegisterCore(OverworldID, core); err != nil {
		t.Fatal(err)
	}
	if _, ok := manager.OpenWorld(0, "example:arenas/one", native.WorldDimensionOverworld); !ok {
		t.Fatal("OpenWorld failed")
	}
	t.Cleanup(func() { _ = manager.CloseCustom() })
	if _, err := manager.Open("example:../escape", native.WorldDimensionOverworld); err == nil {
		t.Fatal("unsafe world path accepted")
	}
	if _, err := os.Stat(filepath.Join(root, "example", "arenas", "one", "db")); err != nil {
		t.Fatalf("persistent DB missing: %v", err)
	}
}
