package native

/*
#include "bridge.h"
*/
import "C"

import (
	"time"
	"unicode/utf8"
)

const (
	worldConfigV1Size         = 80
	maxWorldProviderPathBytes = 4096
)

//export bg_go_world_new
func bg_go_world_new(context C.uint64_t, config *C.DfWorldConfigV1, output *C.DfWorldId) C.DfStatus {
	if output == nil {
		return C.DF_STATUS_ERROR
	}
	output.value = 0
	host, ok := resolveHost(uint64(context))
	value, valid := copyWorldConfig(config)
	if !ok || !valid {
		return C.DF_STATUS_ERROR
	}
	id, ok := host.CreateWorld(value)
	if !ok || id == 0 {
		return C.DF_STATUS_ERROR
	}
	output.value = C.uint64_t(id)
	return C.DF_STATUS_OK
}

func copyWorldConfig(config *C.DfWorldConfigV1) (WorldConfig, bool) {
	if config == nil || uint32(config.struct_size) < worldConfigV1Size || config.reserved != 0 ||
		uint32(config.provider_kind) > uint32(WorldProviderMCDB) || config.read_only > 1 {
		return WorldConfig{}, false
	}
	dimension, valid := copyWorldDimension(config)
	if !valid {
		return WorldConfig{}, false
	}
	pathBytes, valid := copyWorldBytes(config.provider_path, maxWorldProviderPathBytes)
	path := string(pathBytes)
	provider := WorldProviderKind(config.provider_kind)
	if !valid || !utf8.Valid(pathBytes) || provider == WorldProviderNop && path != "" || provider == WorldProviderMCDB && path == "" {
		return WorldConfig{}, false
	}
	return WorldConfig{
		Dimension:           WorldDimension(config.dimension),
		CustomDimension:     dimension,
		Provider:            provider,
		ProviderPath:        path,
		ReadOnly:            config.read_only != 0,
		SaveInterval:        time.Duration(config.save_interval_nanoseconds),
		ChunkUnloadInterval: time.Duration(config.chunk_unload_interval_nanoseconds),
		RandomTickSpeed:     int(int32(config.random_tick_speed)),
	}, true
}

func copyWorldDimension(config *C.DfWorldConfigV1) (*CustomWorldDimension, bool) {
	view := config.dimension_view
	if view.custom > 1 || view.water_evaporates > 1 || view.weather_cycle > 1 || view.time_cycle > 1 {
		return nil, false
	}
	if view.custom == 0 {
		if uint32(config.dimension) > uint32(WorldDimensionEnd) || uint32(view.id) != uint32(config.dimension) ||
			view.water_evaporates != 0 ||
			view.weather_cycle != 0 || view.time_cycle != 0 || view.range_min != 0 ||
			view.range_max != 0 || view.lava_spread_nanoseconds != 0 {
			return nil, false
		}
		return nil, true
	}
	if view.id != 0 {
		return nil, false
	}
	return &CustomWorldDimension{
		Range:           BlockRange{Min: int32(view.range_min), Max: int32(view.range_max)},
		WaterEvaporates: view.water_evaporates != 0, LavaSpreadDuration: time.Duration(view.lava_spread_nanoseconds),
		WeatherCycle: view.weather_cycle != 0, TimeCycle: view.time_cycle != 0,
	}, true
}
