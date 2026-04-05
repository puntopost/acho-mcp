package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/puntopost/acho-mcp/internal/cli/config"
)

func loadConfig() (config.Config, error) {
	if ConfigPath != "" {
		return loadConfigFrom(ConfigPath)
	}
	return loadConfigFrom(config.DefaultPath())
}

func configPath() string {
	if ConfigPath != "" {
		return ConfigPath
	}
	return config.DefaultPath()
}

func loadConfigFrom(path string) (config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := config.Default()
			if path == config.DefaultPath() {
				if err := saveConfig(cfg); err != nil {
					return cfg, fmt.Errorf("failed to create default config: %w", err)
				}
			}
			return cfg, nil
		}
		return config.Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := config.Default()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return config.Config{}, fmt.Errorf("failed to parse config: %w", err)
	}
	cfg.EnabledProjects = config.NormalizeProjects(cfg.EnabledProjects)

	return cfg, nil
}

func saveConfig(cfg config.Config) error {
	cfg.EnabledProjects = config.NormalizeProjects(cfg.EnabledProjects)
	path := configPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".config.json.*")
	if err != nil {
		return fmt.Errorf("failed to create temp config: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to write temp config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to close temp config: %w", err)
	}
	if err := os.Chmod(tmpName, 0644); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to chmod temp config: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to replace config: %w", err)
	}
	return nil
}
