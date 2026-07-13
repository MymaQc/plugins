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
	if !filepath.IsAbs(config.Worlds.Directory) {
		config.Worlds.Directory = filepath.Join(base, config.Worlds.Directory)
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
	players := host.NewPlayers()
	dragonflyConfig, err := config.Dragonfly.Config(log)
	if err != nil {
		return fmt.Errorf("configure Dragonfly: %w", err)
	}
	srv := dragonflyConfig.New()
	worlds, err := NewPersistentWorldManager(config.Worlds.Directory, log, players)
	if err != nil {
		_ = srv.Close()
		return err
	}
	var pluginRuntime *native.Runtime
	enabled := false
	defer func() {
		if enabled {
			pluginRuntime.Disable()
		}
		if err := worlds.CloseCustom(); err != nil {
			log.Error("close custom worlds", "error", err)
		}
		if err := srv.Close(); err != nil {
			log.Error("close Dragonfly server", "error", err)
		}
		if pluginRuntime != nil {
			pluginRuntime.Close()
		}
	}()
	if err := worlds.RegisterCore(OverworldID, srv.World()); err != nil {
		return err
	}
	if err := worlds.RegisterCore(NetherID, srv.Nether()); err != nil {
		return err
	}
	if err := worlds.RegisterCore(EndID, srv.End()); err != nil {
		return err
	}
	pluginRuntime, err = native.OpenWithHost(config.Plugins.RuntimeLibrary, config.Plugins.Directory, struct {
		*host.Players
		*WorldManager
	}{players, worlds})
	if err != nil {
		return err
	}
	if err := pluginRuntime.Enable(); err != nil {
		return err
	}
	enabled = true
	if err := host.RegisterCommands(pluginRuntime, players); err != nil {
		return err
	}
	srv.Listen()

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
		players.Register(p, generation.Add(1))
		handler := host.NewPlayerHandler(pluginRuntime, log, players)
		p.Handle(handler)
		if handler.Join(p) {
			p.Disconnect("Connection rejected by a plugin.")
		}
	}
	return nil
}
