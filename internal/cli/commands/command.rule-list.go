package commands

import (
	"fmt"

	"github.com/puntopost/acho-mcp/internal/cli"
	"github.com/puntopost/acho-mcp/internal/cli/term"
	"github.com/puntopost/acho-mcp/internal/persistence/rule"
)

func init() {
	Register(&ruleList{})
}

var _ Command = (*ruleList)(nil)

type ruleList struct{}

func (c *ruleList) Match(name string) bool { return name == "rules list" }
func (c *ruleList) Usage() string          { return "acho rules list [flags]" }
func (c *ruleList) Description() string    { return "List rules" }
func (c *ruleList) Order() int             { return 38 }
func (c *ruleList) Help() string {
	return `acho rules list — List rules

Filter semantics:
  (no flag)              All rules, across every project.
  --project              Auto-detected current project + globals.
  --project=NAME         Project NAME + globals.
  --global               Only global rules (project='').
  --project and --global combined is an error.

Deleted filter:
  (no flag)              Show active + soft-deleted.
  --active               Only active (deleted = 0).
  --deleted              Only soft-deleted (deleted = 1).

Examples:
  acho rules list
  acho rules list --project
  acho rules list --global
  acho rules list --deleted
`
}

func (c *ruleList) Run(args []string) error {
	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	state, value := projectFlag(args)
	global := hasFlag(args, "--global")
	if global && state != "absent" {
		return fmt.Errorf("--project and --global are mutually exclusive")
	}

	onlyActive := hasFlag(args, "--active")
	onlyDeleted := hasFlag(args, "--deleted")
	if onlyActive && onlyDeleted {
		return fmt.Errorf("--active and --deleted are mutually exclusive")
	}

	project := ""
	switch state {
	case "bare":
		project = d.Project
	case "value":
		project = value
	}

	q := rule.ListQuery{}
	switch {
	case onlyDeleted:
		q.OnlyDeleted = true
	case !onlyActive:
		q.IncludeDeleted = true
	}

	rules, err := d.Rules.List(project, global, q)
	if err != nil {
		return err
	}

	if len(rules) == 0 {
		fmt.Printf("%sNo rules found.%s\n", term.T.Muted(), term.T.Reset())
		return nil
	}

	fmt.Printf("%s%sFound %d rules:%s\n\n", term.T.Bold(), term.T.Primary(), len(rules), term.T.Reset())
	for _, r := range rules {
		printRule(r)
	}
	return nil
}

func printRule(r rule.Rule) {
	label := "global"
	if !r.IsGlobal() {
		label = fmt.Sprintf("project:%s", r.Project)
	}
	deletedTag := ""
	if r.Deleted {
		d := ""
		if r.DeletedDate != nil {
			d = " at " + r.DeletedDate.Format(cli.DateYMD_HM)
		}
		deletedTag = fmt.Sprintf(" %s[DELETED%s]%s", term.T.Danger(), d, term.T.Reset())
	}
	fmt.Printf("%s (%s%s%s) — %s%s%s%s\n  %s\n  %sdate: %s%s\n\n",
		term.T.ID(r.ID),
		term.T.Primary(), label, term.T.Reset(),
		term.T.Bold(), r.Title, term.T.Reset(), deletedTag,
		r.Text,
		term.T.Muted(), r.Date.Format(cli.DateYMD_HM), term.T.Reset(),
	)
}
