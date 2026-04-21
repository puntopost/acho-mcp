package commands

import (
	"fmt"

	"github.com/puntopost/acho-mcp/internal/cli/term"
)

func init() {
	Register(&ruleRestore{})
}

var _ Command = (*ruleRestore)(nil)

type ruleRestore struct{}

func (c *ruleRestore) Match(name string) bool { return name == "rules restore" }
func (c *ruleRestore) Usage() string          { return "acho rules restore <id>" }
func (c *ruleRestore) Description() string    { return "Restore a soft-deleted rule" }
func (c *ruleRestore) Order() int             { return 41 }
func (c *ruleRestore) Help() string {
	return `acho rules restore — Restore a soft-deleted rule by id

Usage:
  acho rules restore <id>

Examples:
  acho rules restore 01K...
`
}

func (c *ruleRestore) Run(args []string) error {
	id := positionalArg(args, 0)
	if id == "" {
		return fmt.Errorf("usage: acho rules restore <id>")
	}

	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	if err := d.Rules.Restore(id); err != nil {
		return err
	}

	fmt.Printf("%s %s\n", term.T.Success("Restored rule"), term.T.ID(id))
	return nil
}
