package framework

import (
	"context"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunFailedPluginEnableNeverBindsListener(t *testing.T) {
	t.Skip("legacy Rust lifecycle failure plugin removed")
	probe, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := probe.LocalAddr().String()
	if err := probe.Close(); err != nil {
		t.Fatal(err)
	}

	t.Setenv("BEDROCK_GOPHERS_LIFECYCLE_ERROR", "missing FFA arena database")
	config := nativeRunConfig(t, address)

	err = Run(context.Background(), config, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err == nil || !strings.Contains(err.Error(), "missing FFA arena database") {
		t.Fatalf("Run error = %v", err)
	}

	rebound, err := net.ListenPacket("udp", address)
	if err != nil {
		t.Fatalf("failed enable leaked listener %s: %v", address, err)
	}
	if err := rebound.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRunReportsListenerBindFailure(t *testing.T) {
	occupied, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()

	err = Run(
		context.Background(),
		nativeRunConfig(t, occupied.LocalAddr().String()),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err == nil || !strings.Contains(err.Error(), "open Dragonfly listeners") {
		t.Fatalf("Run error = %v", err)
	}
}

func nativeRunConfig(t *testing.T, address string) Config {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	extension := ".so"
	if runtime.GOOS == "darwin" {
		extension = ".dylib"
	} else if runtime.GOOS == "windows" {
		extension = ".dll"
	}
	library := filepath.Join(root, "build", "lib", "libdragonfly_plugin_runtime"+extension)
	if _, err := os.Stat(library); err != nil {
		t.Skipf("native runtime not built: run make build-native (%v)", err)
	}
	data := t.TempDir()
	config := DefaultConfig()
	config.Plugins.RuntimeLibrary = library
	config.Plugins.Directory = filepath.Join(root, "build", "plugins")
	config.Worlds.Directory = filepath.Join(data, "worlds")
	config.Dragonfly.Network.Address = address
	config.Dragonfly.World.SaveData = false
	config.Dragonfly.Players.SaveData = false
	config.Dragonfly.Resources.AutoBuildPack = false
	config.Dragonfly.Resources.Folder = filepath.Join(data, "resources")
	return config
}
