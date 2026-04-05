package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/puntopost/acho-mcp/internal/cli"
	"github.com/puntopost/acho-mcp/internal/cli/config"
	"github.com/puntopost/acho-mcp/internal/cli/term"
)

func init() {
	Register(&configShow{})
}

var _ Command = (*configShow)(nil)

type configShow struct{}

func (c *configShow) Match(name string) bool { return name == "config" }
func (c *configShow) Usage() string          { return "acho config" }
func (c *configShow) Description() string    { return "Create or update configuration" }
func (c *configShow) Order() int             { return 5 }
func (c *configShow) Help() string {
	return `acho config — Create or update configuration

Usage:
  acho config

Interactive setup that creates or updates ~/.acho/config.json with:
  - Database path (default: ~/.acho/acho.db)
  - Default pagination limit (default: 10)
  - Snippet length for list previews (default: 1000 chars)

If a config file already exists, current values are shown as defaults.
Press Enter at each prompt to keep the current value.

Set ACHO_PATH to override the base directory (default: ~/.acho).

Examples:
  acho config
  ACHO_PATH=/tmp/myproject acho config
`
}

func (c *configShow) Run(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(args, " "))
	}

	current := config.Default()
	configPath := configPath()
	if _, err := os.Stat(configPath); err == nil {
		loaded, err := loadConfigFrom(configPath)
		if err == nil {
			current = loaded
		}
	}

	fmt.Printf("%s%sAcho configuration%s\n", term.T.Bold(), term.T.Primary(), term.T.Reset())

	dbPath, err := prompt("Database path", current.DBPath)
	if err != nil {
		return err
	}
	paginationLimit, err := promptInt("Default pagination limit", current.DefaultPaginationLimit)
	if err != nil {
		return err
	}
	snippetLength, err := promptInt("Snippet length", current.SnippetLength)
	if err != nil {
		return err
	}

	if err := config.WithConfigLock(func() error {
		// Re-read under lock so we don't clobber enabled_projects edits that
		// landed while the user was filling the prompts.
		merged := current
		if _, err := os.Stat(configPath); err == nil {
			if loaded, err := loadConfigFrom(configPath); err == nil {
				merged.EnabledProjects = loaded.EnabledProjects
			}
		}
		cfg := config.Config{
			DBPath:                 dbPath,
			DBOptions:              current.DBOptions,
			DefaultPaginationLimit: paginationLimit,
			SnippetLength:          snippetLength,
			EnabledProjects:        merged.EnabledProjects,
		}
		return saveConfig(cfg)
	}); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\n%s to %s%s%s\n", term.T.Success("Config saved"), term.T.Bold(), configPath, term.T.Reset())

	return nil
}

func prompt(label string, defaultValue string) (string, error) {
	fmt.Printf("%s%s%s [%s%s%s]: ", term.T.Primary(), label, term.T.Reset(), term.T.Muted(), defaultValue, term.T.Reset())
	input, err := cli.PromptLine()
	if err != nil {
		return "", err
	}
	if input == "" {
		return defaultValue, nil
	}
	return input, nil
}

func promptInt(label string, defaultValue int) (int, error) {
	fmt.Printf("%s%s%s [%s%d%s]: ", term.T.Primary(), label, term.T.Reset(), term.T.Muted(), defaultValue, term.T.Reset())
	input, err := cli.PromptLine()
	if err != nil {
		return 0, err
	}
	if input == "" {
		return defaultValue, nil
	}
	val, err := strconv.Atoi(input)
	if err != nil {
		fmt.Printf("  %sInvalid number, using default: %d%s\n", term.T.Danger(), defaultValue, term.T.Reset())
		return defaultValue, nil
	}
	return val, nil
}
