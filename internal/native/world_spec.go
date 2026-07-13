package native

/*
#include "bridge.h"
*/
import "C"

import (
	"math"
	"time"
	"unicode/utf8"
)

const maxWorldProviderPathBytes = 4096

type worldOpenSpecWire struct {
	structSize                   uint32
	providerPath                 []byte
	saveMilliseconds             uint64
	chunkUnloadMilliseconds      uint64
	fixedTime                    int64
	dimension, openMode          uint32
	savePolicy, randomTickPolicy uint32
	randomTickRate, timePolicy   uint32
	weatherPolicy, unloadPolicy  uint32
	readOnly                     uint8
	reserved                     [3]uint8
}

func copyWorldOpenSpec(view *C.DfWorldOpenSpecV1) (WorldOpenSpec, bool) {
	if view == nil {
		return WorldOpenSpec{}, false
	}
	providerPath, ok := copyWorldBytes(view.provider_path, maxWorldProviderPathBytes)
	if !ok {
		return WorldOpenSpec{}, false
	}
	return validateWorldOpenSpecWire(worldOpenSpecWire{
		structSize: uint32(view.struct_size), providerPath: providerPath,
		saveMilliseconds:        uint64(view.save_interval_milliseconds),
		chunkUnloadMilliseconds: uint64(view.chunk_unload_interval_milliseconds),
		fixedTime:               int64(view.fixed_time), dimension: uint32(view.dimension),
		openMode: uint32(view.open_mode), savePolicy: uint32(view.save_policy),
		randomTickPolicy: uint32(view.random_tick_policy), randomTickRate: uint32(view.random_tick_rate),
		timePolicy: uint32(view.time_policy), weatherPolicy: uint32(view.weather_policy),
		unloadPolicy: uint32(view.chunk_unload_policy), readOnly: uint8(view.read_only),
		reserved: [3]uint8{uint8(view.reserved[0]), uint8(view.reserved[1]), uint8(view.reserved[2])},
	})
}

func validateWorldOpenSpecWire(wire worldOpenSpecWire) (WorldOpenSpec, bool) {
	const knownStructSize = 80
	if wire.structSize < knownStructSize || len(wire.providerPath) == 0 || len(wire.providerPath) > maxWorldProviderPathBytes || !utf8.Valid(wire.providerPath) ||
		wire.dimension > uint32(WorldDimensionEnd) || wire.openMode > uint32(WorldCreateNew) ||
		wire.savePolicy > uint32(WorldSaveManual) || wire.randomTickPolicy > uint32(WorldRandomTicksPerSubchunk) ||
		wire.timePolicy > uint32(WorldTimeFixed) || wire.weatherPolicy > uint32(WorldWeatherClear) ||
		wire.unloadPolicy != uint32(WorldChunkUnloadAfter) || wire.readOnly > 1 || wire.reserved != [3]uint8{} {
		return WorldOpenSpec{}, false
	}
	if wire.readOnly == 1 && (wire.savePolicy != uint32(WorldSaveManual) || wire.saveMilliseconds != 0) {
		return WorldOpenSpec{}, false
	}
	if wire.readOnly == 0 {
		switch WorldSavePolicy(wire.savePolicy) {
		case WorldSaveAutomatic:
			if !validWorldDuration(wire.saveMilliseconds) {
				return WorldOpenSpec{}, false
			}
		case WorldSaveManual:
			if wire.saveMilliseconds != 0 {
				return WorldOpenSpec{}, false
			}
		}
	}
	switch WorldRandomTickPolicy(wire.randomTickPolicy) {
	case WorldRandomTicksDisabled:
		if wire.randomTickRate != 0 {
			return WorldOpenSpec{}, false
		}
	case WorldRandomTicksPerSubchunk:
		if wire.randomTickRate == 0 || wire.randomTickRate > math.MaxInt32 {
			return WorldOpenSpec{}, false
		}
	}
	if WorldTimePolicy(wire.timePolicy) != WorldTimeFixed && wire.fixedTime != 0 {
		return WorldOpenSpec{}, false
	}
	if !validWorldDuration(wire.chunkUnloadMilliseconds) {
		return WorldOpenSpec{}, false
	}
	return WorldOpenSpec{
		ProviderPath: string(wire.providerPath), Dimension: WorldDimension(wire.dimension),
		OpenMode: WorldOpenMode(wire.openMode), ReadOnly: wire.readOnly == 1,
		Save: WorldSavePolicy(wire.savePolicy), SaveInterval: time.Duration(wire.saveMilliseconds) * time.Millisecond,
		RandomTicks: WorldRandomTickPolicy(wire.randomTickPolicy), RandomTickRate: wire.randomTickRate,
		Time: WorldTimePolicy(wire.timePolicy), FixedTime: wire.fixedTime,
		Weather: WorldWeatherPolicy(wire.weatherPolicy), ChunkUnload: WorldChunkUnloadPolicy(wire.unloadPolicy),
		ChunkUnloadAfter: time.Duration(wire.chunkUnloadMilliseconds) * time.Millisecond,
	}, true
}

func validWorldDuration(milliseconds uint64) bool {
	return milliseconds > 0 && milliseconds <= uint64(math.MaxInt64/int64(time.Millisecond))
}
