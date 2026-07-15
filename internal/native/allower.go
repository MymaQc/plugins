package native

/*
#include "bridge.h"
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"unicode/utf8"
	"unsafe"

	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

const (
	maxAllowerDataBytes = 16 << 20
	maxAllowerMessage   = 4096
)

// Allow calls each enabled plugin's Dragonfly server.Allower implementation.
// The first denial wins. Connection data is copied before crossing the ABI.
func (r *Runtime) Allow(addr net.Addr, identity login.IdentityData, client login.ClientData) (string, bool, error) {
	if r == nil || r.ptr == nil {
		return "", false, errors.New("native runtime is closed")
	}
	if addr == nil {
		return "", false, errors.New("connection address is nil")
	}
	identityJSON, err := marshalAllowerValue(identity)
	if err != nil {
		return "", false, fmt.Errorf("snapshot identity data: %w", err)
	}
	clientJSON, err := marshalAllowerValue(client)
	if err != nil {
		return "", false, fmt.Errorf("snapshot client data: %w", err)
	}
	if len(identityJSON) > maxAllowerDataBytes || len(clientJSON) > maxAllowerDataBytes {
		return "", false, errors.New("connection data exceeds native limit")
	}

	arena := &nativeViewArena{}
	defer arena.release()
	input := C.DfAllowInput{}
	var ok bool
	if input.network, ok = arena.stringView([]byte(addr.Network())); !ok {
		return "", false, errors.New("allocate connection network")
	}
	if input.address, ok = arena.stringView([]byte(addr.String())); !ok {
		return "", false, errors.New("allocate connection address")
	}
	if input.identity_json, ok = arena.stringView(identityJSON); !ok {
		return "", false, errors.New("allocate identity data")
	}
	if input.client_json, ok = arena.stringView(clientJSON); !ok {
		return "", false, errors.New("allocate client data")
	}
	if udp, isUDP := addr.(*net.UDPAddr); isUDP {
		input.is_udp = 1
		input.port = C.int32_t(udp.Port)
		if input.ip, ok = arena.stringView([]byte(udp.IP)); !ok {
			return "", false, errors.New("allocate connection IP")
		}
		if input.zone, ok = arena.stringView([]byte(udp.Zone)); !ok {
			return "", false, errors.New("allocate connection zone")
		}
	}

	buffer := C.malloc(maxAllowerMessage)
	if buffer == nil {
		return "", false, errors.New("allocate allower message")
	}
	defer C.free(buffer)
	message := C.DfStringBuffer{data: (*C.uint8_t)(buffer), capacity: maxAllowerMessage}
	var allowed C.uint8_t
	if status := C.bg_runtime_allow(r.ptr, &input, &message, &allowed); status != C.DF_STATUS_OK {
		return "", false, fmt.Errorf("native allower failed with status %d", int32(status))
	}
	if allowed > 1 || uint64(message.len) > uint64(message.capacity) {
		return "", false, errors.New("native allower returned invalid state")
	}
	value := C.GoBytes(unsafe.Pointer(message.data), C.int(message.len))
	if !utf8.Valid(value) {
		return "", false, errors.New("native allower returned invalid UTF-8")
	}
	if allowed != 0 {
		return "", true, nil
	}
	return string(value), false, nil
}

// marshalAllowerValue deliberately uses Go field names instead of JSON tags.
// This preserves fields such as PlayFabID that gophertunnel excludes from its
// wire JSON while keeping the private transport derived from the current type.
func marshalAllowerValue(value any) ([]byte, error) {
	snapshot, err := snapshotAllowerValue(reflect.ValueOf(value))
	if err != nil {
		return nil, err
	}
	return json.Marshal(snapshot)
}

func snapshotAllowerValue(value reflect.Value) (any, error) {
	if !value.IsValid() {
		return nil, nil
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, nil
		}
		return snapshotAllowerValue(value.Elem())
	}
	switch value.Kind() {
	case reflect.Struct:
		result := make(map[string]any, value.NumField())
		typeInfo := value.Type()
		for index := 0; index < value.NumField(); index++ {
			field := typeInfo.Field(index)
			if field.PkgPath != "" {
				continue
			}
			item, err := snapshotAllowerValue(value.Field(index))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", field.Name, err)
			}
			result[field.Name] = item
		}
		return result, nil
	case reflect.Array, reflect.Slice:
		if value.Kind() == reflect.Slice && value.IsNil() {
			return []any{}, nil
		}
		result := make([]any, value.Len())
		for index := range result {
			item, err := snapshotAllowerValue(value.Index(index))
			if err != nil {
				return nil, err
			}
			result[index] = item
		}
		return result, nil
	case reflect.String:
		return value.String(), nil
	case reflect.Bool:
		return value.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return value.Float(), nil
	default:
		return nil, fmt.Errorf("unsupported %s", value.Type())
	}
}
