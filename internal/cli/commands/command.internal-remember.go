package commands

import (
	"fmt"

	"github.com/puntopost/acho-mcp/internal/cli/agent"
)

const internalRememberMarkdown = "Check the ==MANDATORY== rules loaded at session start. Some may require you to act before answering this message (e.g., query sql_query, save a registry, apply a convention).\n"

func init() {
	Register(&internalRemember{})
}

type internalRemember struct{}

func (c *internalRemember) Match(name string) bool { return name == "internal remember" }
func (c *internalRemember) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: acho internal remember <agent>")
	}
	a, err := agent.Get(args[0])
	if err != nil {
		return err
	}

	enabled, err := internalRememberProjectEnabled(args)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	fmt.Print(a.FormatRemember(internalRememberMarkdown))
	return nil
}

func (c *internalRemember) Usage() string       { return "" }
func (c *internalRemember) Description() string { return "" }
func (c *internalRemember) Help() string        { return "" }
func (c *internalRemember) Order() int          { return 1001 }
func (c *internalRemember) Hidden() bool        { return true }

func internalRememberProjectEnabled(args []string) (bool, error) {
	cfg, err := loadConfig()
	if err != nil {
		return false, fmt.Errorf("failed to load config: %w", err)
	}

	projectName, err := resolveProjectName(args)
	if err != nil {
		return false, err
	}

	return cfg.IsProjectEnabled(projectName), nil
}
