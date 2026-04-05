package commands

import (
	"fmt"

	"github.com/puntopost/acho-mcp/internal/cli/term"
)

func init() {
	Register(&purge{})
}

var _ Command = (*purge)(nil)

type purge struct{}

func (c *purge) Match(name string) bool { return name == "purge" }
func (c *purge) Usage() string          { return "acho purge" }
func (c *purge) Description() string {
	return "Hard-delete all soft-deleted rows (registries, types, rules)"
}
func (c *purge) Order() int { return 55 }
func (c *purge) Help() string {
	return `acho purge — Hard-delete every soft-deleted row across all tables

Usage:
  acho purge

Removes all rows with deleted=1 from registries, registry_types and rules.
Runs in order: registries first, then registry_types (safe because any
cascade-deleted registries are gone by then), then rules.

This cannot be undone.
`
}

func (c *purge) Run(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments")
	}

	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	regs, err := d.Service.PurgeDeleted()
	if err != nil {
		return err
	}
	types, err := d.Types.PurgeDeleted()
	if err != nil {
		return err
	}
	rules, err := d.Rules.PurgeDeleted()
	if err != nil {
		return err
	}

	fmt.Printf("%s %sregistries: %s%d%s | types: %s%d%s | rules: %s%d%s\n",
		term.T.Success("Purged"),
		term.T.Muted(),
		term.T.Secondary(), regs, term.T.Muted(),
		term.T.Secondary(), types, term.T.Muted(),
		term.T.Secondary(), rules, term.T.Reset(),
	)
	return nil
}
