package framework

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync/atomic"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/world"
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
	worlds, err := NewPersistentWorldManager(config.Worlds.Directory, log, players)
	if err != nil {
		return err
	}
	pluginRuntime, err := native.OpenWithHost(config.Plugins.RuntimeLibrary, config.Plugins.Directory, struct {
		*host.Players
		*WorldManager
	}{players, worlds})
	if err != nil {
		return err
	}
	entityTypes, err := pluginRuntime.EntityTypes()
	if err != nil {
		pluginRuntime.Close()
		return err
	}
	dragonflyConfig, err := config.Dragonfly.Config(log)
	if err != nil {
		pluginRuntime.Close()
		closeDragonflyProviders(dragonflyConfig, log)
		return fmt.Errorf("configure Dragonfly: %w", err)
	}
	dragonflyConfig.Entities, err = buildEntityRegistry(dragonflyConfig.Entities, entityTypes)
	if err != nil {
		pluginRuntime.Close()
		closeDragonflyProviders(dragonflyConfig, log)
		return fmt.Errorf("configure plugin entities: %w", err)
	}
	var srv *server.Server
	enabled := false
	started := false
	defer func() {
		if enabled {
			pluginRuntime.Disable()
		}
		if started {
			if err := srv.Close(); err != nil {
				log.Error("close Dragonfly server", "error", err)
			}
		} else if srv != nil {
			for _, managed := range []*world.World{srv.End(), srv.Nether(), srv.World()} {
				if err := managed.Close(); err != nil {
					log.Error("close unstarted Dragonfly world", "error", err)
				}
			}
			if err := dragonflyConfig.PlayerProvider.Close(); err != nil {
				log.Error("close unstarted player provider", "error", err)
			}
		}
		if err := worlds.CloseCustom(); err != nil {
			log.Error("close custom worlds", "error", err)
		}
		pluginRuntime.Close()
	}()
	srv = dragonflyConfig.New()
	if err := worlds.RegisterCore(OverworldID, srv.World()); err != nil {
		return err
	}
	if err := worlds.RegisterCore(NetherID, srv.Nether()); err != nil {
		return err
	}
	if err := worlds.RegisterCore(EndID, srv.End()); err != nil {
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
	started = true

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
		handler := host.NewPlayerHandler(pluginRuntime, log, players, worlds)
		p.Handle(handler)
		if handler.Join(p) {
			p.Disconnect("Connection rejected by a plugin.")
		}
	}
	return nil
}

func closeDragonflyProviders(config server.Config, log *slog.Logger) {
	if config.PlayerProvider != nil {
		if err := config.PlayerProvider.Close(); err != nil {
			log.Error("close unused player provider", "error", err)
		}
	}
	if config.WorldProvider != nil {
		if err := config.WorldProvider.Close(); err != nil {
			log.Error("close unused world provider", "error", err)
		}
	}
}
