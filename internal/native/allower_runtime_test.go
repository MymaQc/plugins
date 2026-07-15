package native

import (
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

func TestAllowerRuntimeRoundTrip(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("allower C fixture does not support Windows")
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	extension := ".so"
	flags := []string{"-shared", "-fPIC", "-std=c11"}
	if runtime.GOOS == "darwin" {
		extension = ".dylib"
		flags[0] = "-dynamiclib"
	}
	runtimeLibrary := filepath.Join(root, "build", "lib", "libdragonfly_plugin_runtime"+extension)
	if _, err := os.Stat(runtimeLibrary); err != nil {
		t.Skipf("C# runtime not built: run make build-native (%v)", err)
	}
	pluginDirectory := t.TempDir()
	plugin := filepath.Join(pluginDirectory, "allower-test"+extension)
	arguments := append(flags,
		"-I", filepath.Join(root, "abi", "include"),
		filepath.Join(root, "internal", "native", "testdata", "allower_plugin.c"),
		"-o", plugin,
	)
	if output, err := exec.Command("cc", arguments...).CombinedOutput(); err != nil {
		t.Fatalf("compile allower fixture: %v\n%s", err, output)
	}
	pluginRuntime, err := Open(runtimeLibrary, pluginDirectory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pluginRuntime.Close)
	if err := pluginRuntime.Enable(); err != nil {
		t.Fatal(err)
	}

	message, allowed, err := pluginRuntime.Allow(
		&net.UDPAddr{IP: net.ParseIP("fe80::1"), Port: 19132, Zone: "eth0"},
		login.IdentityData{DisplayName: "Unicode玩家", PlayFabID: "ignored-by-normal-json"},
		login.ClientData{DeviceOS: protocol.DeviceLinux},
	)
	if err != nil {
		t.Fatal(err)
	}
	if allowed || message != "denied by fixture" {
		t.Fatalf("Allow() = (%q, %v)", message, allowed)
	}
}
