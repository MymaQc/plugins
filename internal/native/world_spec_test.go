package native

import (
	"math"
	"strings"
	"testing"
	"time"
)

func validWorldOpenSpecWire() worldOpenSpecWire {
	return worldOpenSpecWire{
		structSize: 80, providerPath: []byte("arenas/one"),
		saveMilliseconds:        uint64((10 * time.Minute) / time.Millisecond),
		chunkUnloadMilliseconds: uint64((2 * time.Minute) / time.Millisecond),
		dimension:               uint32(WorldDimensionOverworld), openMode: uint32(WorldOpenOrCreate),
		savePolicy: uint32(WorldSaveAutomatic), randomTickPolicy: uint32(WorldRandomTicksPerSubchunk),
		randomTickRate: 3, timePolicy: uint32(WorldTimePreserve),
		weatherPolicy: uint32(WorldWeatherPreserve), unloadPolicy: uint32(WorldChunkUnloadAfter),
	}
}

func TestCopyWorldOpenSpecValidatesWireContract(t *testing.T) {
	wire := validWorldOpenSpecWire()
	got, ok := validateWorldOpenSpecWire(wire)
	if !ok {
		t.Fatal("valid specification rejected")
	}
	if got.ProviderPath != "arenas/one" || got.SaveInterval != 10*time.Minute || got.ChunkUnloadAfter != 2*time.Minute || got.RandomTickRate != 3 {
		t.Fatalf("specification = %#v", got)
	}
}

func TestCopyWorldOpenSpecRejectsMalformedWire(t *testing.T) {
	maximumDuration := uint64(math.MaxInt64 / int64(time.Millisecond))
	tests := map[string]func(*worldOpenSpecWire){
		"small struct":         func(wire *worldOpenSpecWire) { wire.structSize = 79 },
		"empty path":           func(wire *worldOpenSpecWire) { wire.providerPath = nil },
		"oversized path":       func(wire *worldOpenSpecWire) { wire.providerPath = []byte(strings.Repeat("x", 4097)) },
		"invalid UTF-8":        func(wire *worldOpenSpecWire) { wire.providerPath = []byte{0xff} },
		"unknown dimension":    func(wire *worldOpenSpecWire) { wire.dimension = 3 },
		"unknown open mode":    func(wire *worldOpenSpecWire) { wire.openMode = 3 },
		"unknown save":         func(wire *worldOpenSpecWire) { wire.savePolicy = 2 },
		"unknown ticks":        func(wire *worldOpenSpecWire) { wire.randomTickPolicy = 2 },
		"unknown time":         func(wire *worldOpenSpecWire) { wire.timePolicy = 3 },
		"unknown weather":      func(wire *worldOpenSpecWire) { wire.weatherPolicy = 3 },
		"unknown unload":       func(wire *worldOpenSpecWire) { wire.unloadPolicy = 1 },
		"invalid bool":         func(wire *worldOpenSpecWire) { wire.readOnly = 2 },
		"reserved byte":        func(wire *worldOpenSpecWire) { wire.reserved[1] = 1 },
		"automatic zero":       func(wire *worldOpenSpecWire) { wire.saveMilliseconds = 0 },
		"automatic overflow":   func(wire *worldOpenSpecWire) { wire.saveMilliseconds = maximumDuration + 1 },
		"manual duration":      func(wire *worldOpenSpecWire) { wire.savePolicy, wire.saveMilliseconds = uint32(WorldSaveManual), 1 },
		"read-only automatic":  func(wire *worldOpenSpecWire) { wire.readOnly = 1 },
		"disabled tick rate":   func(wire *worldOpenSpecWire) { wire.randomTickPolicy = uint32(WorldRandomTicksDisabled) },
		"zero tick rate":       func(wire *worldOpenSpecWire) { wire.randomTickRate = 0 },
		"overflow tick rate":   func(wire *worldOpenSpecWire) { wire.randomTickRate = math.MaxInt32 + 1 },
		"preserved fixed time": func(wire *worldOpenSpecWire) { wire.fixedTime = 1 },
		"zero unload":          func(wire *worldOpenSpecWire) { wire.chunkUnloadMilliseconds = 0 },
		"overflow unload":      func(wire *worldOpenSpecWire) { wire.chunkUnloadMilliseconds = maximumDuration + 1 },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			wire := validWorldOpenSpecWire()
			mutate(&wire)
			if _, ok := validateWorldOpenSpecWire(wire); ok {
				t.Fatal("malformed specification accepted")
			}
		})
	}
}

func TestCopyWorldOpenSpecAcceptsCanonicalReadOnly(t *testing.T) {
	wire := validWorldOpenSpecWire()
	wire.readOnly = 1
	wire.savePolicy = uint32(WorldSaveManual)
	wire.saveMilliseconds = 0
	got, ok := validateWorldOpenSpecWire(wire)
	if !ok || !got.ReadOnly || got.Save != WorldSaveManual || got.SaveInterval != 0 {
		t.Fatalf("read-only specification = %#v, %v", got, ok)
	}
}
