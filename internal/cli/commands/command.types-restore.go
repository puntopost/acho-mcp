package commands

import (
	"fmt"

	"github.com/puntopost/acho-mcp/internal/cli/term"
)

func init() {
	Register(&typeRestore{})
}

var _ Command = (*typeRestore)(nil)

type typeRestore struct{}

func (c *typeRestore) Match(name string) bool { return name == "types restore" }
func (c *typeRestore) Usage() string          { return "acho types restore <name> [--force]" }
func (c *typeRestore) Description() string    { return "Restore a soft-deleted registry type" }
func (c *typeRestore) Order() int             { return 43 }
func (c *typeRestore) Help() string {
	return `acho types restore — Restore a soft-deleted registry type by name

Usage:
  acho types restore <name> [--force]

If deleted registries still use this type, restore fails unless --force is
passed, which cascade-restores every deleted registry of that type along with
the type itself.

Examples:
  acho types restore bugfix
  acho types restore obsolete_type --force
`
}

func (c *typeRestore) Run(args []string) error {
	name := positionalArg(args, 0)
	if name == "" {
		return fmt.Errorf("usage: acho types restore <name> [--force]")
	}

	force := hasFlag(args, "--force")

	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	restored, err := d.Types.Restore(name, force)
	if err != nil {
		return err
	}

	if restored > 0 {
		fmt.Printf("%s %s%s%s (and %s%d%s registries)\n",
			term.T.Success("Restored type"),
			term.T.Bold(), name, term.T.Reset(),
			term.T.Bold(), restored, term.T.Reset())
	} else {
		fmt.Printf("%s %s%s%s\n", term.T.Success("Restored type"), term.T.Bold(), name, term.T.Reset())
	}
	return nil
}
