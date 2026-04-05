package commands

import (
	"fmt"

	"github.com/puntopost/acho-mcp/internal/cli"
	"github.com/puntopost/acho-mcp/internal/cli/term"
)

func init() {
	Register(&get{})
}

var _ Command = (*get)(nil)

type get struct{}

func (c *get) Match(name string) bool { return name == "registries get" }
func (c *get) Usage() string          { return "acho registries get <id>" }
func (c *get) Description() string    { return "Get full content of a registry" }
func (c *get) Order() int             { return 33 }
func (c *get) Help() string {
	return `acho registries get — Get full content of a registry

Usage:
  acho registries get <id>

Arguments:
  id                     Registry ID (required). Shown in search and list results.

Shows the complete content of a registry, including metadata:
type, project, and hit counters (searches, gets, updates).

Each call increments the get_hits counter for the registry.

Examples:
  acho registries get 01K...
`
}

func (c *get) Run(args []string) error {
	id := positionalArg(args, 0)
	if id == "" {
		return fmt.Errorf("usage: acho registries get <id>")
	}

	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	r, err := d.Service.GetAny(id)
	if err != nil {
		return err
	}

	projectLabel := r.Project
	if projectLabel == "" {
		projectLabel = "(global)"
	}
	deletedTag := ""
	if r.Deleted {
		d := ""
		if r.DeletedDate != nil {
			d = " at " + r.DeletedDate.Format(cli.DateYMD_HM)
		}
		deletedTag = fmt.Sprintf(" %s[DELETED%s]%s", term.T.Danger(), d, term.T.Reset())
	}
	fmt.Printf("%s (%s%s%s) — %s%s%s%s\n",
		term.T.ID(r.ID), term.T.Primary(), r.Type, term.T.Reset(),
		term.T.Bold(), r.Title, term.T.Reset(), deletedTag)
	fmt.Printf("%sproject: %s%s%s | date: %s%s\n",
		term.T.Muted(),
		term.T.Secondary(), projectLabel, term.T.Muted(),
		r.Date.Format(cli.DateYMD_HM),
		term.T.Reset())
	fmt.Printf("%ssearches: %s%d%s | gets: %s%d%s | updates: %s%d%s\n\n",
		term.T.Muted(),
		term.T.Secondary(), r.SearchHits, term.T.Muted(),
		term.T.Secondary(), r.GetHits, term.T.Muted(),
		term.T.Secondary(), r.UpdateHits, term.T.Reset())
	fmt.Println(r.Content)
	return nil
}
