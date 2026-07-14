package framework

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"reflect"
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
	if err := validateConfig(config); err != nil {
		return err
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
	dragonflyConfig.Entities, err = buildEntityRegistry(dragonflyConfig.Entities, entityTypes, foreignEntityServices{
		runtime: pluginRuntime, players: players, entities: worlds.entityHandles,
	})
	if err != nil {
		pluginRuntime.Close()
		closeDragonflyProviders(dragonflyConfig, log)
		return fmt.Errorf("configure plugin entities: %w", err)
	}
	if err := applyCoreWorldPolicy(&dragonflyConfig, config.Worlds.Core); err != nil {
		pluginRuntime.Close()
		closeDragonflyProviders(dragonflyConfig, log)
		return fmt.Errorf("configure core worlds: %w", err)
	}
	listenerGate := deferListenerCreation(&dragonflyConfig)
	var srv *server.Server
	cleanup := runCleanup{
		log: log,
		closeStarted: func() error {
			return srv.Close()
		},
		beginPlugins:  pluginRuntime.BeginDisable,
		closeCustom:   worlds.CloseCustom,
		drainDetached: worlds.DrainDetachedEntities,
		finishPlugins: pluginRuntime.FinishDisable,
		closeUnstarted: func() {
			if err := listenerGate.closeAll(); err != nil {
				log.Error("close unstarted Dragonfly listeners", "error", err)
			}
			if srv == nil {
				return
			}
			for _, managed := range []*world.World{srv.End(), srv.Nether(), srv.World()} {
				if err := managed.Close(); err != nil {
					log.Error("close unstarted Dragonfly world", "error", err)
				}
			}
			if dragonflyConfig.PlayerProvider != nil {
				if err := dragonflyConfig.PlayerProvider.Close(); err != nil {
					log.Error("close unstarted player provider", "error", err)
				}
			}
		},
		closeRuntime: pluginRuntime.Close,
	}
	defer cleanup.close()
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
	if err := host.RegisterCommands(pluginRuntime, players); err != nil {
		return err
	}
	if err := listenerGate.openAll(); err != nil {
		return fmt.Errorf("open Dragonfly listeners: %w", err)
	}
	srv.Listen()
	cleanup.started = true

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
	closeDragonflyProvider("player", config.PlayerProvider, log)
	closeDragonflyProvider("world", config.WorldProvider, log)
}

func closeDragonflyProvider(name string, provider interface{ Close() error }, log *slog.Logger) {
	if provider == nil {
		return
	}
	value := reflect.ValueOf(provider)
	if value.Kind() == reflect.Pointer && value.IsNil() {
		return
	}
	if err := provider.Close(); err != nil {
		log.Error("close unused "+name+" provider", "error", err)
	}
}
