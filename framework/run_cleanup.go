package framework

import "log/slog"

// runCleanup keeps teardown ordering explicit. A started Dragonfly server is
// closed first because Close synchronously rejects and drains player callbacks.
// Plugin disable is two-phase: on_disable runs with custom worlds and host calls
// available, then custom worlds drain entity callbacks, then runtime admission
// closes. An unstarted server closes its core worlds before final admission
// closes, because their providers may still save plugin-owned entities.
type runCleanup struct {
	log            *slog.Logger
	started        bool
	closeStarted   func() error
	beginPlugins   func()
	closeCustom    func() error
	drainDetached  func()
	finishPlugins  func()
	closeUnstarted func()
	closeRuntime   func()
}

func (cleanup *runCleanup) close() {
	if cleanup.started {
		if err := cleanup.closeStarted(); err != nil {
			cleanup.log.Error("close Dragonfly server", "error", err)
		}
	}
	cleanup.beginPlugins()
	if err := cleanup.closeCustom(); err != nil {
		cleanup.log.Error("close custom worlds", "error", err)
	}
	cleanup.drainDetached()
	if !cleanup.started {
		cleanup.closeUnstarted()
	}
	cleanup.finishPlugins()
	cleanup.closeRuntime()
}
