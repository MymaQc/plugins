package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlayerStateMethodsFollowDragonflyAST(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	module, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	methods, err := inspectPlayerStateMethods(filepath.Join(string(bytes.TrimSpace(module)), "server", "player", "player.go"))
	if err != nil {
		t.Fatal(err)
	}
	generated := string(generatePlayerStateMethods(methods))
	for _, expected := range []string{
		"int Food()",
		"void SetFood(int level)",
		"double Health()",
		"double MaxHealth()",
		"void SetMaxHealth(double health)",
		"double Heal(double health, World.HealingSource source)",
		"(double Damage, bool Vulnerable) Hurt(double dmg, World.DamageSource src)",
		"int ExperienceLevel()",
		"void SetExperienceLevel(int level)",
		"double ExperienceProgress()",
		"void SetExperienceProgress(double progress)",
		"double Scale()",
		"void SetScale(double s)",
		"bool Invisible()",
		"void SetInvisible()",
		"void SetVisible()",
		"bool Immobile()",
		"void SetImmobile()",
		"void SetMobile()",
		"double Speed()",
		"void SetSpeed(double speed)",
		"double FlightSpeed()",
		"void SetFlightSpeed(double flightSpeed)",
		"double VerticalFlightSpeed()",
		"void SetVerticalFlightSpeed(double flightSpeed)",
		"void ResetFallDistance()",
		"double FallDistance()",
		"void SetAbsorption(double health)",
		"double Absorption()",
		"bool Dead()",
		"bool OnGround()",
		"double EyeHeight()",
		"double TorsoHeight()",
		"bool Breathing()",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated player state methods missing %q", expected)
		}
	}
	if strings.Contains(generated, "DfPlayerState") || strings.Contains(generated, "bool Set") {
		t.Fatal("public Player state API exposes transport status")
	}
	if strings.Contains(generated, "double.IsFinite") {
		t.Fatal("SetExperienceProgress rejects NaN unlike Dragonfly")
	}
}
