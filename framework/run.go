package framework

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync/atomic"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
)

// RunFile loads configuration and runs the owned Dragonfly server until ctx is cancelled.
func RunFile(ctx context.Context, configPath string, log *slog.Logger) error {
	config, err := LoadConfig(configPath)
	if err != nil {
		return err
	}
	base := filepath.Dir(configPath)
	config.Plugins.RuntimeLibrary = filepath.Join(base, "lib", runtimeLibraryFilename())
	if !filepath.IsAbs(config.Plugins.Directory) {
		config.Plugins.Directory = filepath.Join(base, config.Plugins.Directory)
	}
	resolveDataPath(base, &config.Dragonfly.World.Folder)
	resolveDataPath(base, &config.Dragonfly.Players.Folder)
	resolveDataPath(base, &config.Dragonfly.Resources.Folder)
	return Run(ctx, config, log)
}

func resolveDataPath(base string, path *string) {
	if *path != "" && !filepath.IsAbs(*path) {
		*path = filepath.Join(base, *path)
	}
}

// Run constructs and owns the plugin runtime and Dragonfly server lifecycle.
func Run(ctx context.Context, config Config, log *slog.Logger) error {
	if log == nil {
		log = slog.Default()
	}
	pluginRuntime, err := native.Open(config.Plugins.RuntimeLibrary, config.Plugins.Directory)
	if err != nil {
		return err
	}
	defer pluginRuntime.Close()

	dragonflyConfig, err := config.Dragonfly.Config(log)
	if err != nil {
		return fmt.Errorf("configure Dragonfly: %w", err)
	}
	srv := dragonflyConfig.New()
	srv.World().Handle(host.NewWorldHandler())
	srv.Nether().Handle(host.NewWorldHandler())
	srv.End().Handle(host.NewWorldHandler())
	srv.Listen()
	defer func() {
		if err := srv.Close(); err != nil {
			log.Error("close Dragonfly server", "error", err)
		}
	}()

	stopped := make(chan struct{})
	defer close(stopped)
	go func() {
		select {
		case <-ctx.Done():
			if err := srv.Close(); err != nil {
				log.Error("close Dragonfly server", "error", err)
			}
		case <-stopped:
		}
	}()

	var generation atomic.Uint64
	for p := range srv.Accept() {
		p.Handle(host.NewPlayerHandler(pluginRuntime, log, generation.Add(1)))
	}
	return nil
}
