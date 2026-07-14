package native

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCSharpRuntimeLifecycleAndQuit(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
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
	plugins := filepath.Join(root, "build", "plugins")
	if _, err := os.Stat(library); err != nil {
		t.Skipf("C# runtime not built: run make build-native (%v)", err)
	}
	pluginRuntime, err := Open(library, plugins)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pluginRuntime.Close)

	if got := pluginRuntime.PluginCount(); got != 1 {
		t.Fatalf("PluginCount() = %d, want 1", got)
	}
	if got := pluginRuntime.Subscriptions(); got != 1<<3 {
		t.Fatalf("Subscriptions() = %d, want %d", got, 1<<3)
	}
	if err := pluginRuntime.Enable(); err != nil {
		t.Fatal(err)
	}
	if err := pluginRuntime.HandlePlayerQuit(1, PlayerQuitInput{Name: "Gopher"}); err != nil {
		t.Fatal(err)
	}
}
