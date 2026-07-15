package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPlayerPresentationMethodsMatchPinnedDragonfly(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	directory, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	methods, err := inspectPlayerPresentationMethods(filepath.Join(string(bytes.TrimSpace(directory)), "server", "player", "player.go"))
	if err != nil {
		t.Fatal(err)
	}
	generated := generatePlayerPresentationMethods(methods)
	for _, expected := range [][]byte{
		[]byte("public void EnableInstantRespawn()"),
		[]byte("public void DisableInstantRespawn()"),
		[]byte("public void ShowCoordinates()"),
		[]byte("public void HideCoordinates()"),
		[]byte("public void SendSleepingIndicator(int sleeping, int max)"),
		[]byte("public void CloseDialogue()"),
		[]byte("public void RemoveBossBar()"),
		[]byte("public void RemoveScoreboard()"),
		[]byte("public string NameTag()"),
		[]byte("public string ScoreTag()"),
		[]byte("public void SendToast(string title, string message)"),
	} {
		if !bytes.Contains(generated, expected) {
			t.Fatalf("generated presentation methods missing %q", expected)
		}
	}
}
