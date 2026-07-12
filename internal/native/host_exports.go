package native

/*
#include "bridge.h"
*/
import "C"

import "time"

//export bg_go_player_text
func bg_go_player_text(context C.uint64_t, player C.DfPlayerId, kind C.uint32_t, message C.DfStringView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	var id PlayerID
	for index := range id.UUID {
		id.UUID[index] = byte(player.bytes[index])
	}
	id.Generation = uint64(player.generation)
	if !host.SendPlayerText(id, PlayerTextKind(kind), stringView(message)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_title
func bg_go_player_title(context C.uint64_t, player C.DfPlayerId, value C.DfTitleView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	var id PlayerID
	for index := range id.UUID {
		id.UUID[index] = byte(player.bytes[index])
	}
	id.Generation = uint64(player.generation)
	title := PlayerTitle{
		Text: stringView(value.text), Subtitle: stringView(value.subtitle),
		ActionText: stringView(value.action_text),
		FadeIn:     milliseconds(value.fade_in_milliseconds),
		Duration:   milliseconds(value.duration_milliseconds),
		FadeOut:    milliseconds(value.fade_out_milliseconds),
	}
	if !host.SendPlayerTitle(id, title) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

func milliseconds(value C.uint64_t) time.Duration {
	const maximum = uint64((1<<63 - 1) / int64(time.Millisecond))
	millis := min(uint64(value), maximum)
	return time.Duration(millis) * time.Millisecond
}
