package commands

import (
	"fmt"

	"github.com/puntopost/acho-mcp/internal/cli/term"
)

func init() {
	Register(&typeDelete{})
}

var _ Command = (*typeDelete)(nil)

type typeDelete struct{}

func (c *typeDelete) Match(name string) bool { return name == "types delete" }
func (c *typeDelete) Usage() string          { return "acho types delete <name> [--force]" }
func (c *typeDelete) Description() string    { return "Delete a registry type" }
func (c *typeDelete) Order() int             { return 42 }
func (c *typeDelete) Help() string {
	return `acho types delete — Soft-delete a registry type by name

Usage:
  acho types delete <name> [--force]

Fails if any active registry is using this type, unless --force is passed,
which cascade-deletes every active registry of that type along with the type
itself. Use acho types restore to bring it back.

Examples:
  acho types delete bugfix
  acho types delete obsolete_type --force
`
}

func (c *typeDelete) Run(args []string) error {
	name := positionalArg(args, 0)
	if name == "" {
		return fmt.Errorf("usage: acho types delete <name> [--force]")
	}

	force := hasFlag(args, "--force")

	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	removed, err := d.Types.Delete(name, force)
	if err != nil {
		return err
	}

	if removed > 0 {
		fmt.Printf("%s %s%s%s (and %s%d%s registries)\n",
			term.T.Success("Deleted type"),
			term.T.Bold(), name, term.T.Reset(),
			term.T.Bold(), removed, term.T.Reset())
	} else {
		fmt.Printf("%s %s%s%s\n", term.T.Success("Deleted type"), term.T.Bold(), name, term.T.Reset())
	}
	return nil
}
