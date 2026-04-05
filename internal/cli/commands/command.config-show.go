package commands

import (
	"fmt"
	"slices"
	"strings"

	"github.com/puntopost/acho-mcp/internal/cli/term"
)

func init() {
	Register(&configShowCmd{})
}

var _ Command = (*configShowCmd)(nil)

type configShowCmd struct{}

func (c *configShowCmd) Match(name string) bool { return name == "config show" }
func (c *configShowCmd) Usage() string          { return "acho config show" }
func (c *configShowCmd) Description() string    { return "Show current configuration" }
func (c *configShowCmd) Order() int             { return 4 }
func (c *configShowCmd) Help() string {
	return `acho config show — Show current configuration

Usage:
  acho config show

Prints the active configuration values from ~/.acho/config.json.

Examples:
  acho config show
`
}

func (c *configShowCmd) Run(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(args, " "))
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	configPath := configPath()

	fmt.Printf("\n%s%sAcho configuration%s (%s)\n\n", term.T.Bold(), term.T.Primary(), term.T.Reset(), configPath)

	printField("Database path", cfg.DBPath)
	printField("DB options", cfg.DBOptions)
	printField("Pagination limit", fmt.Sprintf("%d", cfg.DefaultPaginationLimit))
	printField("Snippet length", fmt.Sprintf("%d", cfg.SnippetLength))
	if len(cfg.EnabledProjects) == 0 {
		printField("Enabled projects", "(none)")
	} else {
		projects := slices.Clone(cfg.EnabledProjects)
		slices.Sort(projects)
		printField("Enabled projects", strings.Join(projects, ", "))
	}

	fmt.Println()
	return nil
}

func printField(label, value string) {
	fmt.Printf("  %s%-20s%s %s\n", term.T.Primary(), label, term.T.Reset(), value)
}
