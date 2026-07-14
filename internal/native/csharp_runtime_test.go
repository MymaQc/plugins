package native

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func openCSharpRuntime(t testing.TB) *Runtime {
	t.Helper()
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

	if got := pluginRuntime.PluginCount(); got != 2 {
		t.Fatalf("PluginCount() = %d, want 2", got)
	}
	if got := pluginRuntime.Subscriptions(); got != PlayerMoveSubscription|PlayerQuitSubscription {
		t.Fatalf("Subscriptions() = %d, want %d", got, PlayerMoveSubscription|PlayerQuitSubscription)
	}
	if err := pluginRuntime.Enable(); err != nil {
		t.Fatal(err)
	}
	return pluginRuntime
}

func TestCSharpRuntimeLifecycleAndQuit(t *testing.T) {
	pluginRuntime := openCSharpRuntime(t)
	if err := pluginRuntime.HandlePlayerQuit(1, PlayerQuitInput{Name: "Gopher"}); err != nil {
		t.Fatal(err)
	}
	cancelled, err := pluginRuntime.HandlePlayerMove(2, PlayerMoveInput{NewPosition: Vec3{Y: -65}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !cancelled {
		t.Fatal("movement below the world was not cancelled")
	}
	cancelled, err = pluginRuntime.HandlePlayerMove(3, PlayerMoveInput{NewPosition: Vec3{Y: 64}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("ordinary movement was cancelled")
	}
}

func TestCSharpRuntimeHandlesMovementConcurrently(t *testing.T) {
	pluginRuntime := openCSharpRuntime(t)
	var wait sync.WaitGroup
	errors := make(chan error, 8)
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 1_000 {
				if _, err := pluginRuntime.HandlePlayerMove(1, PlayerMoveInput{NewPosition: Vec3{Y: 64}}, false); err != nil {
					errors <- err
					return
				}
			}
		}()
	}
	wait.Wait()
	close(errors)
	for err := range errors {
		t.Fatal(err)
	}
}

func BenchmarkCSharpMovement(b *testing.B) {
	pluginRuntime := openCSharpRuntime(b)
	input := PlayerMoveInput{NewPosition: Vec3{Y: 64}}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := pluginRuntime.HandlePlayerMove(1, input, false); err != nil {
			b.Fatal(err)
		}
	}
}
