package framework

import (
	"errors"
	"io"
	"log/slog"
	"reflect"
	"testing"
)

func TestRunCleanupStartedOrder(t *testing.T) {
	var calls []string
	customOpen := true
	cleanup := runCleanup{
		log:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		started: true,
		closeStarted: func() error {
			calls = append(calls, "server")
			return nil
		},
		beginPlugins: func() {
			if !customOpen {
				t.Fatal("custom worlds closed before plugin disable")
			}
			calls = append(calls, "begin plugins")
		},
		closeCustom: func() error {
			customOpen = false
			calls = append(calls, "custom")
			return nil
		},
		drainDetached: func() { calls = append(calls, "detached") },
		finishPlugins: func() {
			if customOpen {
				t.Fatal("plugins finalized before custom worlds closed")
			}
			calls = append(calls, "finish plugins")
		},
		closeUnstarted: func() { calls = append(calls, "unstarted") },
		closeRuntime:   func() { calls = append(calls, "runtime") },
	}
	cleanup.close()
	want := []string{"server", "begin plugins", "custom", "detached", "finish plugins", "runtime"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("cleanup order = %v, want %v", calls, want)
	}
}

func TestRunCleanupUnstartedEnabledOrder(t *testing.T) {
	var calls []string
	cleanup := runCleanup{
		log:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		closeStarted: func() error { return errors.New("must not run") },
		beginPlugins: func() { calls = append(calls, "begin plugins") },
		closeCustom: func() error {
			calls = append(calls, "custom")
			return nil
		},
		drainDetached:  func() { calls = append(calls, "detached") },
		finishPlugins:  func() { calls = append(calls, "finish plugins") },
		closeUnstarted: func() { calls = append(calls, "core") },
		closeRuntime:   func() { calls = append(calls, "runtime") },
	}
	cleanup.close()
	want := []string{"begin plugins", "custom", "detached", "core", "finish plugins", "runtime"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("cleanup order = %v, want %v", calls, want)
	}
}

func TestRunCleanupFailedEnableFinalizesPartialState(t *testing.T) {
	var calls []string
	cleanup := runCleanup{
		log:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		closeStarted: func() error { return errors.New("must not run") },
		beginPlugins: func() { calls = append(calls, "begin plugins") },
		closeCustom: func() error {
			calls = append(calls, "custom")
			return nil
		},
		drainDetached:  func() { calls = append(calls, "detached") },
		finishPlugins:  func() { calls = append(calls, "finish plugins") },
		closeUnstarted: func() { calls = append(calls, "core") },
		closeRuntime:   func() { calls = append(calls, "runtime") },
	}
	cleanup.close()
	want := []string{"begin plugins", "custom", "detached", "core", "finish plugins", "runtime"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("cleanup order = %v, want %v", calls, want)
	}
}
