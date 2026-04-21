package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentSetupOpenCodeInstallsPluginAndConfig(t *testing.T) {
	configHome := t.TempDir()
	achoPath := t.TempDir()

	cmd := exec.Command(binary, "agent-setup", "opencode")
	cmd.Env = append(os.Environ(),
		"ACHO_PATH="+achoPath,
		"XDG_CONFIG_HOME="+configHome,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("agent-setup opencode failed: %v: %s", err, output)
	}

	configPath := filepath.Join(configHome, "opencode", "opencode.json")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read opencode config: %v", err)
	}
	config := string(configData)

	for _, want := range []string{
		`"$schema": "https://opencode.ai/config.json"`,
		`"acho"`,
		`"command": [`,
		`"mcp"`,
	} {
		if !strings.Contains(config, want) {
			t.Fatalf("expected config to contain %q, got %s", want, config)
		}
	}

	pluginPath := filepath.Join(configHome, "opencode", "plugins", "acho.js")
	pluginData, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatalf("read OpenCode plugin: %v", err)
	}
	for _, want := range []string{
		`event.type === "session.created"`,
		`event.type === "session.idle"`,
		`acho project status`,
		`acho internal context`,
		`acho internal remember`,
	} {
		if !strings.Contains(string(pluginData), want) {
			t.Fatalf("expected plugin to contain %q, got %s", want, pluginData)
		}
	}

	manifestPath := filepath.Join(configHome, "opencode", "package.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read OpenCode package manifest: %v", err)
	}
	manifest := string(manifestData)
	for _, want := range []string{
		`"name": "acho-opencode-plugin"`,
		`"private": true`,
		`"type": "module"`,
		`"version": "0.5.0"`,
	} {
		if !strings.Contains(manifest, want) {
			t.Fatalf("expected package manifest to contain %q, got %s", want, manifest)
		}
	}
}
