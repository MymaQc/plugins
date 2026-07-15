package native

/*
#include "dragonfly_plugin.h"
*/
import "C"

import (
	"unicode/utf8"
	"unsafe"
)

const maxPlayerPresentationStringBytes = 16 << 20

//export bg_go_player_string_get
func bg_go_player_string_get(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, kind C.uint32_t, output *C.DfStringBuffer) C.DfStatus {
	if output == nil {
		return C.DF_STATUS_ERROR
	}
	output.len = 0
	host, ok := resolveHost(uint64(context))
	if !ok {
		return C.DF_STATUS_ERROR
	}
	value, ok := host.PlayerString(InvocationID(invocation), playerID(player), PlayerStringKind(kind))
	if !ok || !utf8.ValidString(value) || !writePlayerStringBuffer(output, []byte(value)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

//export bg_go_player_toast
func bg_go_player_toast(context C.uint64_t, invocation C.DfInvocationId, player C.DfPlayerId, title, message C.DfStringView) C.DfStatus {
	host, ok := resolveHost(uint64(context))
	titleBytes, titleOK := copyNativeBytes(title, maxPlayerPresentationStringBytes)
	messageBytes, messageOK := copyNativeBytes(message, maxPlayerPresentationStringBytes)
	if !ok || !titleOK || !messageOK || !utf8.Valid(titleBytes) || !utf8.Valid(messageBytes) ||
		!host.SendPlayerToast(InvocationID(invocation), playerID(player), string(titleBytes), string(messageBytes)) {
		return C.DF_STATUS_ERROR
	}
	return C.DF_STATUS_OK
}

func writePlayerStringBuffer(output *C.DfStringBuffer, value []byte) bool {
	output.len = C.uint64_t(len(value))
	if len(value) > maxPlayerPresentationStringBytes || uint64(output.capacity) < uint64(len(value)) ||
		len(value) != 0 && output.data == nil {
		return false
	}
	if len(value) != 0 {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(output.data)), len(value)), value)
	}
	return true
}
