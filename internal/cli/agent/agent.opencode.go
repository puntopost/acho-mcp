package agent

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var _ Setup = (*OpenCode)(nil)

func init() {
	Register(&OpenCode{})
}

// OpenCode registers acho as an MCP server in OpenCode.
type OpenCode struct{}

func (o *OpenCode) Name() string        { return "opencode" }
func (o *OpenCode) Description() string { return "Register acho as MCP server in OpenCode" }

// OpenCodePluginFS is set by main.go with the embedded OpenCode plugin files.
var OpenCodePluginFS fs.FS

// OpenCodePluginVersion is set by main.go from the CLI version.
var OpenCodePluginVersion = "0.1.0"

func (o *OpenCode) Setup() error {
	configDir, err := o.configDir()
	if err != nil {
		return err
	}
	path := filepath.Join(configDir, "opencode.json")

	cfg := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse OpenCode config: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read OpenCode config: %w", err)
	}

	if _, ok := cfg["$schema"]; !ok {
		cfg["$schema"] = "https://opencode.ai/config.json"
	}

	if err := extractOpenCodeFiles(configDir); err != nil {
		return err
	}
	if err := writeOpenCodePackageManifest(configDir); err != nil {
		return err
	}

	mcp, err := ensureObject(cfg, "mcp")
	if err != nil {
		return err
	}

	mcp["acho"] = map[string]any{
		"type":    "local",
		"command": []string{"acho", "mcp"},
		"enabled": true,
	}

	if err := ensureStringInArray(cfg, "instructions", "instructions/acho.md"); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create OpenCode config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode OpenCode config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write OpenCode config: %w", err)
	}

	return nil
}

func (o *OpenCode) configDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve user config dir: %w", err)
	}
	return filepath.Join(configDir, "opencode"), nil
}

func ensureObject(root map[string]any, key string) (map[string]any, error) {
	if current, ok := root[key]; ok {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("OpenCode config key %q must be an object", key)
		}
		return object, nil
	}

	object := map[string]any{}
	root[key] = object
	return object, nil
}

func ensureStringInArray(root map[string]any, key string, value string) error {
	if current, ok := root[key]; ok {
		entries, ok := current.([]any)
		if !ok {
			return fmt.Errorf("OpenCode config key %q must be an array", key)
		}

		for _, entry := range entries {
			if s, ok := entry.(string); ok && s == value {
				return nil
			}
		}

		root[key] = append(entries, value)
		return nil
	}

	root[key] = []any{value}
	return nil
}

func extractOpenCodeFiles(destRoot string) error {
	if OpenCodePluginFS == nil {
		return fmt.Errorf("OpenCode plugin files not embedded — this is a build error")
	}

	walkErr := fs.WalkDir(OpenCodePluginFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		dest := filepath.Join(destRoot, path)

		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}

		data, err := fs.ReadFile(OpenCodePluginFS, path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("create %s: %w", filepath.Dir(dest), err)
		}

		if err := os.WriteFile(dest, data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("failed to extract OpenCode plugin: %w", walkErr)
	}

	return nil
}

func writeOpenCodePackageManifest(configDir string) error {
	manifestPath := filepath.Join(configDir, "package.json")
	manifest := map[string]any{
		"name":        "acho-opencode-plugin",
		"private":     true,
		"type":        "module",
		"version":     OpenCodePluginVersion,
		"description": "Acho OpenCode plugin assets installed by acho agent-setup",
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode OpenCode package manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write OpenCode package manifest: %w", err)
	}

	return nil
}
