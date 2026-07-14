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

	if got := pluginRuntime.PluginCount(); got != 4 {
		t.Fatalf("PluginCount() = %d, want 4", got)
	}
	wantSubscriptions := PlayerMoveSubscription | PlayerChatSubscription | PlayerQuitSubscription |
		PlayerFoodLossSubscription | PlayerToggleSprintSubscription | PlayerToggleSneakSubscription |
		PlayerJumpSubscription | PlayerTeleportSubscription | PlayerPunchAirSubscription
	if got := pluginRuntime.Subscriptions(); got != wantSubscriptions {
		t.Fatalf("Subscriptions() = %d, want %d", got, wantSubscriptions)
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
	chat, err := pluginRuntime.HandlePlayerChat(4, PlayerChatInput{Message: "BADWORD"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if chat.Replacement == nil || *chat.Replacement != "***" {
		t.Fatalf("chat replacement = %v, want ***", chat.Replacement)
	}
	food, err := pluginRuntime.HandlePlayerFoodLoss(5, PlayerFoodLossInput{From: 1, To: -1}, false)
	if err != nil {
		t.Fatal(err)
	}
	if food.To != 0 {
		t.Fatalf("food = %d, want 0", food.To)
	}
	if err := pluginRuntime.HandlePlayerJump(6, PlayerID{}); err != nil {
		t.Fatal(err)
	}
	for name, call := range map[string]func() (bool, error){
		"teleport": func() (bool, error) {
			return pluginRuntime.HandlePlayerTeleport(7, PlayerTeleportInput{Position: Vec3{Y: 64}}, false)
		},
		"toggle sprint": func() (bool, error) {
			return pluginRuntime.HandlePlayerToggleSprint(8, PlayerToggleInput{After: true}, false)
		},
		"toggle sneak": func() (bool, error) {
			return pluginRuntime.HandlePlayerToggleSneak(9, PlayerToggleInput{After: true}, false)
		},
		"punch air": func() (bool, error) {
			return pluginRuntime.HandlePlayerPunchAir(10, PlayerID{}, false)
		},
	} {
		cancelled, err := call()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if cancelled {
			t.Fatalf("%s unexpectedly cancelled", name)
		}
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
