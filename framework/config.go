// Package framework owns Dragonfly server construction, plugins, worlds, players, and shutdown.
package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/df-mc/dragonfly/server"
	"github.com/pelletier/go-toml"
)

type PluginConfig struct {
	// RuntimeLibrary is derived from the config directory by RunFile and is not serialized.
	RuntimeLibrary string `toml:"-"`
	Directory      string `toml:"directory"`
}

type WorldConfig struct {
	Directory string `toml:"directory"`
}

type Config struct {
	Plugins   PluginConfig      `toml:"plugins"`
	Worlds    WorldConfig       `toml:"worlds"`
	Dragonfly server.UserConfig `toml:"dragonfly"`
}

func DefaultConfig() Config {
	return Config{
		Plugins: PluginConfig{
			Directory: "plugins",
		},
		Worlds:    WorldConfig{Directory: ".data/worlds"},
		Dragonfly: server.DefaultConfig(),
	}
}

// LoadConfig reads path. If it does not exist, LoadConfig writes and returns defaults.
func LoadConfig(path string) (Config, error) {
	config := DefaultConfig()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		data, err = toml.Marshal(config)
		if err != nil {
			return Config{}, fmt.Errorf("encode default config: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return Config{}, fmt.Errorf("create config directory: %w", err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return Config{}, fmt.Errorf("write default config: %w", err)
		}
		return config, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	if err := toml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if config.Plugins.Directory == "" {
		return Config{}, fmt.Errorf("plugins.directory is required")
	}
	if config.Worlds.Directory == "" {
		return Config{}, fmt.Errorf("worlds.directory is required")
	}
	return config, nil
}

func runtimeLibraryFilename() string {
	if runtime.GOOS == "windows" {
		return "dragonfly_plugin_runtime.dll"
	}
	if runtime.GOOS == "darwin" {
		return "libdragonfly_plugin_runtime.dylib"
	}
	return "libdragonfly_plugin_runtime.so"
}
