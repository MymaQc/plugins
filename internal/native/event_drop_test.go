package native

import (
	"bytes"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func openEventDropRuntime(t *testing.T) *Runtime {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("event-drop C fixture does not support Windows")
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	extension := ".so"
	compileFlags := []string{"-shared", "-fPIC", "-std=c11"}
	if runtime.GOOS == "darwin" {
		extension = ".dylib"
		compileFlags[0] = "-dynamiclib"
	}
	runtimeLibrary := filepath.Join(root, "build", "lib", "libdragonfly_plugin_runtime"+extension)
	if _, err := os.Stat(runtimeLibrary); err != nil {
		t.Skipf("C# runtime not built: run make build-native (%v)", err)
	}
	pluginDirectory := t.TempDir()
	plugin := filepath.Join(pluginDirectory, "event-drop-test"+extension)
	arguments := append(compileFlags,
		"-I", filepath.Join(root, "abi", "include"),
		filepath.Join(root, "internal", "native", "testdata", "event_drop_plugin.c"),
		"-o", plugin,
	)
	command := exec.Command("cc", arguments...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("compile event-drop fixture: %v\n%s", err, output)
	}
	pluginRuntime, err := Open(runtimeLibrary, pluginDirectory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pluginRuntime.Close)
	if err := pluginRuntime.Enable(); err != nil {
		t.Fatal(err)
	}
	return pluginRuntime
}

func TestPlayerTransferReplacementDropExactlyOnce(t *testing.T) {
	pluginRuntime := openEventDropRuntime(t)
	input := PlayerTransferInput{Address: UDPAddress{IP: net.IPv4(127, 0, 0, 1).To4()}}

	input.Address.Port = 1000
	output, err := pluginRuntime.HandlePlayerTransfer(1, input, false)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output.Address.IP, net.IPv4(127, 0, 0, 2).To4()) || output.Address.Port != 19133 {
		t.Fatalf("success replacement = %#v", output.Address)
	}

	input.Address.Port = 1001
	if _, err := pluginRuntime.HandlePlayerTransfer(2, input, false); err == nil || !strings.Contains(err.Error(), "invalid address") {
		t.Fatalf("validation error = %v", err)
	}

	input.Address.Port = 1002
	if _, err := pluginRuntime.HandlePlayerTransfer(3, input, false); err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("status error = %v", err)
	}

	input.Address.Port = 1003
	if _, err := pluginRuntime.HandlePlayerTransfer(4, input, false); err != nil {
		t.Fatalf("drop verification: %v", err)
	}
}

func TestPlayerCommandExecutionReplacementDropExactlyOnce(t *testing.T) {
	pluginRuntime := openEventDropRuntime(t)
	input := PlayerCommandExecutionInput{Arguments: []string{"original"}}

	input.Command.Name = "success"
	output, err := pluginRuntime.HandlePlayerCommandExecution(1, input, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(output.Arguments) != 1 || output.Arguments[0] != "changed" {
		t.Fatalf("success replacement = %#v", output.Arguments)
	}

	input.Command.Name = "invalid"
	if _, err := pluginRuntime.HandlePlayerCommandExecution(2, input, false); err == nil || !strings.Contains(err.Error(), "changed argument count") {
		t.Fatalf("validation error = %v", err)
	}

	input.Command.Name = "status"
	if _, err := pluginRuntime.HandlePlayerCommandExecution(3, input, false); err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("status error = %v", err)
	}

	input.Command.Name = "verify"
	if _, err := pluginRuntime.HandlePlayerCommandExecution(4, input, false); err != nil {
		t.Fatalf("drop verification: %v", err)
	}
}
