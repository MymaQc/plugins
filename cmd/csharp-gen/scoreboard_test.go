package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestScoreboardMatchesPinnedDragonfly(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	directory, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	root := string(bytes.TrimSpace(directory))
	if err := inspectScoreboard(
		filepath.Join(root, "server", "player", "scoreboard", "scoreboard.go"),
		filepath.Join(root, "server", "player", "player.go"),
	); err != nil {
		t.Fatal(err)
	}
	generated := generateScoreboard()
	for _, expected := range [][]byte{
		[]byte("public static Scoreboard New(params object?[] name)"),
		[]byte("public int Write(byte[] p)"),
		[]byte("public int WriteString(string s)"),
		[]byte("public void Set(int index, string s)"),
		[]byte("public void Remove(int index)"),
		[]byte("public void RemovePadding()"),
		[]byte("public string[] Lines()"),
		[]byte("public bool Descending()"),
		[]byte("public void SetDescending()"),
		[]byte("public void SendScoreboard(Scoreboard scoreboard)"),
	} {
		if !bytes.Contains(generated, expected) {
			t.Fatalf("generated scoreboard API missing %q", expected)
		}
	}
}
