package commands

import (
	"fmt"

	"github.com/puntopost/acho-mcp/internal/cli/term"
)

func init() {
	Register(&restoreCmd{})
}

var _ Command = (*restoreCmd)(nil)

type restoreCmd struct{}

func (c *restoreCmd) Match(name string) bool { return name == "registries restore" }
func (c *restoreCmd) Usage() string          { return "acho registries restore <id>" }
func (c *restoreCmd) Description() string    { return "Restore a soft-deleted registry" }
func (c *restoreCmd) Order() int             { return 35 }
func (c *restoreCmd) Help() string {
	return `acho registries restore — Restore a soft-deleted registry

Usage:
  acho registries restore <id>

Arguments:
  id                     Registry ID to restore (required).

Examples:
  acho registries restore 01K...
`
}

func (c *restoreCmd) Run(args []string) error {
	id := positionalArg(args, 0)
	if id == "" {
		return fmt.Errorf("usage: acho registries restore <id>")
	}

	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	if err := d.Service.Restore(id); err != nil {
		return err
	}

	fmt.Printf("%s %s\n", term.T.Success("Restored registry"), term.T.ID(id))
	return nil
}
