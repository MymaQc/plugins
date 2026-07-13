package framework

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigCreatesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "server.toml")
	created, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Plugins != created.Plugins {
		t.Fatalf("plugins = %+v, want %+v", loaded.Plugins, created.Plugins)
	}
	if loaded.Dragonfly.Network.Address != ":19132" {
		t.Fatalf("address = %q", loaded.Dragonfly.Network.Address)
	}
	if loaded.Worlds.Directory != ".data/worlds" {
		t.Fatalf("world directory = %q", loaded.Worlds.Directory)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "runtime-library") || strings.Contains(string(data), ".so") {
		t.Fatalf("config exposes native runtime path:\n%s", data)
	}
}
