package commands

import (
	"fmt"

	"github.com/puntopost/acho-mcp/internal/cli/term"
)

func init() {
	Register(&deleteCmd{})
}

var _ Command = (*deleteCmd)(nil)

type deleteCmd struct{}

func (c *deleteCmd) Match(name string) bool { return name == "registries delete" }
func (c *deleteCmd) Usage() string          { return "acho registries delete <id>" }
func (c *deleteCmd) Description() string    { return "Delete a registry" }
func (c *deleteCmd) Order() int             { return 34 }
func (c *deleteCmd) Help() string {
	return `acho registries delete — Delete a registry

Usage:
  acho registries delete <id>

Arguments:
  id                     Registry ID to delete (required).

Permanently removes the registry from the database. This cannot be undone.

Examples:
  acho registries delete 01K...
`
}

func (c *deleteCmd) Run(args []string) error {
	id := positionalArg(args, 0)
	if id == "" {
		return fmt.Errorf("usage: acho registries delete <id>")
	}

	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	if err := d.Service.Delete(id); err != nil {
		return err
	}

	fmt.Printf("%s %s\n", term.T.Success("Deleted registry"), term.T.ID(id))
	return nil
}
