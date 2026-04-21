package commands

import (
	"fmt"

	"github.com/puntopost/acho-mcp/internal/cli/term"
)

func init() {
	Register(&ruleDelete{})
}

var _ Command = (*ruleDelete)(nil)

type ruleDelete struct{}

func (c *ruleDelete) Match(name string) bool { return name == "rules delete" }
func (c *ruleDelete) Usage() string          { return "acho rules delete <id>" }
func (c *ruleDelete) Description() string    { return "Delete a rule" }
func (c *ruleDelete) Order() int             { return 40 }
func (c *ruleDelete) Help() string {
	return `acho rules delete — Soft-delete a rule by id

Usage:
  acho rules delete <id>

Marks the rule as deleted. Use acho rules restore to bring it back,
or acho purge to remove soft-deleted entries permanently.

Examples:
  acho rules delete 01K...
`
}

func (c *ruleDelete) Run(args []string) error {
	id := positionalArg(args, 0)
	if id == "" {
		return fmt.Errorf("usage: acho rules delete <id>")
	}

	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	if err := d.Rules.Delete(id); err != nil {
		return err
	}

	fmt.Printf("%s %s\n", term.T.Success("Deleted rule"), term.T.ID(id))
	return nil
}
