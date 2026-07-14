package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlayerKinematicsMethodsUseGoAST(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	module, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	methods, err := inspectPlayerKinematicsMethods(filepath.Join(
		string(bytes.TrimSpace(module)), "server", "player", "player.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(methods) != 7 {
		t.Fatalf("generated %d player kinematics methods, want 7", len(methods))
	}
	generated := string(generatePlayerKinematicsMethods(methods))
	for _, expected := range []string{
		"void Teleport(Vector3 pos)",
		"void Move(Vector3 deltaPos, double deltaYaw, double deltaPitch)",
		"void Displace(Vector3 deltaPos)",
		"Vector3 Position()",
		"Vector3 Velocity()",
		"void SetVelocity(Vector3 velocity)",
		"Rotation Rotation()",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated player kinematics missing %q:\n%s", expected, generated)
		}
	}
}
