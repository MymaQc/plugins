package host

import (
	"math"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/world"
)

func WorldDimensionView(value world.Dimension) (native.WorldDimensionView, bool) {
	if id, ok := registeredDimensionID(value); ok && id >= 0 && id <= int(native.WorldDimensionEnd) {
		return native.WorldDimensionView{ID: native.WorldDimension(id)}, true
	}
	rangeValue := value.Range()
	if rangeValue.Min() < math.MinInt32 || rangeValue.Min() > math.MaxInt32 ||
		rangeValue.Max() < math.MinInt32 || rangeValue.Max() > math.MaxInt32 {
		return native.WorldDimensionView{}, false
	}
	return native.WorldDimensionView{Custom: &native.CustomWorldDimension{
		Range:           native.BlockRange{Min: int32(rangeValue.Min()), Max: int32(rangeValue.Max())},
		WaterEvaporates: value.WaterEvaporates(), LavaSpreadDuration: value.LavaSpreadDuration(),
		WeatherCycle: value.WeatherCycle(), TimeCycle: value.TimeCycle(),
	}}, true
}

func registeredDimensionID(value world.Dimension) (id int, ok bool) {
	defer func() {
		if recover() != nil {
			id, ok = 0, false
		}
	}()
	return world.DimensionID(value)
}
