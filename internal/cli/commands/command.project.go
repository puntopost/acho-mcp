package commands

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/puntopost/acho-mcp/internal/cli/config"
)

func init() {
	Register(&project{})
}

var _ Command = (*project)(nil)

type project struct{}

func (c *project) Usage() string {
	return "acho project [status|enable|disable|rename] [--project=NAME]"
}
func (c *project) Description() string { return "Print or manage the detected project name" }
func (c *project) Order() int          { return 50 }
func (c *project) Help() string {
	return `acho project — Print or manage the detected project name

Usage:
  acho project
  acho project status [--project=NAME]
  acho project enable [--project=NAME]
  acho project disable [--project=NAME]
  acho project rename

Without subcommands, prints the project name detected from git remote origin.
If no git remote is available, Acho falls back to the full current directory
path, slugified.

With status, prints whether Acho is enabled for that project.

With enable/disable, updates config.enabled_projects. Acho MCP is disabled by
default for every project until it is explicitly enabled.

With rename, choose an existing stored project and rename its rules, types,
registries and enabled-project config entry to the currently detected project.
`
}

func (c *project) Match(name string) bool {
	return name == "project"
}

func (c *project) Run(args []string) error {
	action, err := parseProjectAction(args)
	if err != nil {
		return err
	}

	projectName, err := resolveProjectName(args)
	if err != nil {
		return err
	}

	if action == "" {
		fmt.Print(projectName)
		return nil
	}

	if action == "status" {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if cfg.IsProjectEnabled(projectName) {
			fmt.Print("enabled")
		} else {
			fmt.Print("disabled")
		}
		return nil
	}

	return config.WithConfigLock(func() error {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		changed := false
		switch action {
		case "enable":
			cfg.EnabledProjects, changed = addProject(cfg.EnabledProjects, projectName)
			if err := saveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			if changed {
				fmt.Printf("Enabled Acho for project %q\n", projectName)
			} else {
				fmt.Printf("Acho was already enabled for project %q\n", projectName)
			}
		case "disable":
			cfg.EnabledProjects, changed = removeProject(cfg.EnabledProjects, projectName)
			if err := saveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			if changed {
				fmt.Printf("Disabled Acho for project %q\n", projectName)
			} else {
				fmt.Printf("Acho was already disabled for project %q\n", projectName)
			}
		default:
			return fmt.Errorf("unsupported action %q", action)
		}

		return nil
	})
}

func parseProjectAction(args []string) (string, error) {
	action := ""
	skipNext := false
	for i, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "--project" {
			if i+1 >= len(args) {
				return "", fmt.Errorf("--project requires a value")
			}
			skipNext = true
			continue
		}
		if strings.HasPrefix(arg, "--project=") {
			if strings.TrimPrefix(arg, "--project=") == "" {
				return "", fmt.Errorf("--project requires a value")
			}
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return "", fmt.Errorf("unexpected flag: %s", arg)
		}
		if action != "" {
			return "", fmt.Errorf("unexpected arguments: %s", strings.Join(args, " "))
		}
		action = arg
	}

	if action == "" || action == "status" || action == "enable" || action == "disable" {
		return action, nil
	}
	return "", fmt.Errorf("unexpected arguments: %s", strings.Join(args, " "))
}

func resolveProjectName(args []string) (string, error) {
	projectName := flagValue(args, "--project")
	if projectName == "" {
		dir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("cannot get working directory: %w", err)
		}
		projectName = detectProject(dir)
	}
	if projectName == "" {
		return "", fmt.Errorf("could not detect project name, use --project=NAME")
	}
	return projectName, nil
}

func addProject(projects []string, project string) ([]string, bool) {
	if slices.Contains(projects, project) {
		return config.NormalizeProjects(projects), false
	}
	return config.NormalizeProjects(append(projects, project)), true
}

func removeProject(projects []string, project string) ([]string, bool) {
	idx := slices.Index(projects, project)
	if idx == -1 {
		return config.NormalizeProjects(projects), false
	}
	updated := append(slices.Clone(projects[:idx]), projects[idx+1:]...)
	return config.NormalizeProjects(updated), true
}
