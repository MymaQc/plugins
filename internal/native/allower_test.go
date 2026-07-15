package native

import (
	"encoding/json"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

func TestMarshalAllowerValuePreservesEveryIdentityField(t *testing.T) {
	data, err := marshalAllowerValue(login.IdentityData{
		XUID: "1", Identity: "identity", DisplayName: "Player", TitleID: "title",
		PlayFabTitleID: "playfab-title", PlayFabID: "playfab",
	})
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]any
	if err := json.Unmarshal(data, &values); err != nil {
		t.Fatal(err)
	}
	for field, want := range map[string]string{
		"XUID": "1", "Identity": "identity", "DisplayName": "Player", "TitleID": "title",
		"PlayFabTitleID": "playfab-title", "PlayFabID": "playfab",
	} {
		if got := values[field]; got != want {
			t.Fatalf("%s = %#v, want %q", field, got, want)
		}
	}
}

func TestMarshalAllowerValueCopiesNestedClientData(t *testing.T) {
	data, err := marshalAllowerValue(login.ClientData{
		DeviceOS:          protocol.DeviceLinux,
		AnimatedImageData: []login.SkinAnimation{{Frames: 2, Image: "image", ImageWidth: 4}},
		PersonaPieces:     []login.PersonaPiece{{PieceID: "piece", Default: true}},
		PieceTintColours: []login.PersonaPieceTintColour{{
			Colours: [4]string{"one", "two", "three", "four"}, PieceType: "eyes",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]any
	if err := json.Unmarshal(data, &values); err != nil {
		t.Fatal(err)
	}
	if got := values["DeviceOS"]; got != float64(protocol.DeviceLinux) {
		t.Fatalf("DeviceOS = %#v", got)
	}
	animations := values["AnimatedImageData"].([]any)
	if got := animations[0].(map[string]any)["Image"]; got != "image" {
		t.Fatalf("animation image = %#v", got)
	}
	colours := values["PieceTintColours"].([]any)[0].(map[string]any)["Colours"].([]any)
	if len(colours) != 4 || colours[3] != "four" {
		t.Fatalf("colours = %#v", colours)
	}
}
