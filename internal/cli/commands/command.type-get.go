package commands

import (
	"fmt"

	"github.com/puntopost/acho-mcp/internal/cli"
	"github.com/puntopost/acho-mcp/internal/cli/term"
)

func init() {
	Register(&typeGet{})
}

var _ Command = (*typeGet)(nil)

type typeGet struct{}

func (c *typeGet) Match(name string) bool { return name == "types get" || name == "types detail" }
func (c *typeGet) Usage() string          { return "acho types get <name>" }
func (c *typeGet) Description() string    { return "Get full content of a registry type" }
func (c *typeGet) Order() int             { return 40 }
func (c *typeGet) Help() string {
	return `acho types get — Get full content of a registry type

Usage:
  acho types get <name>
  acho types detail <name>

Arguments:
  name                   Type name (required).

Shows the complete content of a registry type, including description,
schema, project, and date.

Examples:
  acho types get tools
  acho types detail memory_note
`
}

func (c *typeGet) Run(args []string) error {
	name := positionalArg(args, 0)
	if name == "" {
		return fmt.Errorf("usage: acho types get <name>")
	}

	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	rt, err := d.Types.GetAny(name)
	if err != nil {
		return err
	}

	projectLabel := rt.Project
	if projectLabel == "" {
		projectLabel = "(global)"
	}
	deletedTag := ""
	if rt.Deleted {
		d := ""
		if rt.DeletedDate != nil {
			d = " at " + rt.DeletedDate.Format(cli.DateYMD_HM)
		}
		deletedTag = fmt.Sprintf(" %s[DELETED%s]%s", term.T.Danger(), d, term.T.Reset())
	}

	fmt.Printf("%s%s%s%s\n", term.T.Bold(), rt.Name, term.T.Reset(), deletedTag)
	fmt.Printf("%sdescription:%s %s\n", term.T.Muted(), term.T.Reset(), rt.Description)
	fmt.Printf("%sschema:%s %s\n", term.T.Muted(), term.T.Reset(), rt.Schema)
	fmt.Printf("%sproject: %s%s%s | date: %s%s\n\n",
		term.T.Muted(),
		term.T.Secondary(), projectLabel, term.T.Muted(),
		rt.Date.Format(cli.DateYMD_HM),
		term.T.Reset())
	return nil
}
