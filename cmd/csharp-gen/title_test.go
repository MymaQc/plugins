package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestTitleMatchesPinnedDragonfly(t *testing.T) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
	directory, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	root := string(bytes.TrimSpace(directory))
	if err := inspectTitle(
		filepath.Join(root, "server", "player", "title", "title.go"),
		filepath.Join(root, "server", "player", "player.go"),
	); err != nil {
		t.Fatal(err)
	}
	generated := generateTitle()
	for _, expected := range [][]byte{
		[]byte("public readonly record struct Title"),
		[]byte("public static Title New(params object?[] text)"),
		[]byte("TimeSpan.FromTicks(500000), TimeSpan.FromSeconds(2), TimeSpan.FromTicks(500000)"),
		[]byte("public Title WithSubtitle(params object?[] text)"),
		[]byte("public Title WithActionText(params object?[] text)"),
		[]byte("public Title WithDuration(TimeSpan d)"),
		[]byte("public Title WithFadeInDuration(TimeSpan d)"),
		[]byte("public Title WithFadeOutDuration(TimeSpan d)"),
		[]byte("public void SendTitle(Title t)"),
	} {
		if !bytes.Contains(generated, expected) {
			t.Fatalf("generated title API missing %q", expected)
		}
	}
}

func TestTitleInspectorRejectsDrift(t *testing.T) {
	title := `package title
import "time"
type Title struct{}
func New(text ...any) Title { return Title{} }
func (Title) Text() string { return "" }
func (Title) WithSubtitle(text ...any) Title { return Title{} }
func (Title) Subtitle() string { return "" }
func (Title) WithActionText(text ...any) Title { return Title{} }
func (Title) ActionText() string { return "" }
func (Title) Duration() time.Duration { return 0 }
func (Title) WithDuration(d time.Duration) Title { return Title{} }
func (Title) WithFadeInDuration(d time.Duration) Title { return Title{} }
func (Title) FadeInDuration() time.Duration { return 0 }
func (Title) WithFadeOutDuration(d time.Duration) Title { return Title{} }
func (Title) FadeOutDuration() time.Duration { return 0 }
`
	player := `package player
import "example/title"
type Player struct{}
func (*Player) SendTitle(t title.Title) {}
`
	directory := t.TempDir()
	titlePath, playerPath := filepath.Join(directory, "title.go"), filepath.Join(directory, "player.go")
	if err := os.WriteFile(titlePath, []byte(title), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(playerPath, []byte(player), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := inspectTitle(titlePath, playerPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(playerPath, bytes.ReplaceAll([]byte(player), []byte("title.Title"), []byte("string")), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := inspectTitle(titlePath, playerPath); err == nil {
		t.Fatal("changed SendTitle signature accepted")
	}
}
