package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestPlayerControlsMatchPinnedDragonfly(t *testing.T) {
	dragonfly := moduleDirectoryForTest(t, "github.com/df-mc/dragonfly")
	gophertunnel := moduleDirectoryForTest(t, "github.com/sandertv/gophertunnel")
	spec, err := inspectPlayerControls(
		filepath.Join(dragonfly, "server", "player", "player.go"),
		filepath.Join(dragonfly, "server", "player", "hud", "element.go"),
		filepath.Join(dragonfly, "server", "player", "input", "lock.go"),
		filepath.Join(gophertunnel, "minecraft", "protocol", "packet", "update_client_input_locks.go"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Methods) != 6 || len(spec.HudElements) != 13 || len(spec.InputLocks) != 11 {
		t.Fatalf("control counts = %d/%d/%d", len(spec.Methods), len(spec.HudElements), len(spec.InputLocks))
	}
	if spec.HudElements[0] != (playerControlValue{Name: "PaperDoll", Value: 0}) ||
		spec.HudElements[12] != (playerControlValue{Name: "ItemText", Value: 12}) ||
		spec.InputLocks[0] != (playerControlValue{Name: "Camera", Value: 2}) ||
		spec.InputLocks[10] != (playerControlValue{Name: "MoveRight", Value: 4096}) {
		t.Fatalf("control values = hud %#v, input %#v", spec.HudElements, spec.InputLocks)
	}
	generated := generatePlayerControls(spec)
	for _, expected := range [][]byte{
		[]byte("public static Element PaperDoll()"),
		[]byte("public static Lock Camera()"),
		[]byte("public void ShowHudElement(Hud.Element e)"),
		[]byte("public bool HudElementHidden(Hud.Element e)"),
		[]byte("public void LockInput(Input.Lock l)"),
		[]byte("public bool InputLocked(Input.Lock l)"),
	} {
		if !bytes.Contains(generated, expected) {
			t.Fatalf("generated controls missing %q", expected)
		}
	}
	if bytes.Contains(generated, []byte("public Element(")) || bytes.Contains(generated, []byte("public Lock(")) {
		t.Fatal("opaque Dragonfly control values gained public constructors")
	}
}
