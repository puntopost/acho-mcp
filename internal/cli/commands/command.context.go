package commands

import (
	"fmt"

	"github.com/puntopost/acho-mcp/internal/cli/term"
	"github.com/puntopost/acho-mcp/internal/persistence/rtype"
	"github.com/puntopost/acho-mcp/internal/persistence/rule"
)

func init() {
	Register(&contextLoad{})
}

var _ Command = (*contextLoad)(nil)

type contextLoad struct{}

func (c *contextLoad) Match(name string) bool { return name == "context" }
func (c *contextLoad) Usage() string          { return "acho context [flags]" }
func (c *contextLoad) Description() string    { return "Show mandatory rules for session context" }
func (c *contextLoad) Order() int             { return 36 }
func (c *contextLoad) Help() string {
	return `acho context — Show mandatory rules for session context

Filter semantics:
  (no flag)              All rules, across every project.
  --project              Auto-detected current project + globals.
  --project=NAME         Project NAME + globals.
  --global               Only global rules (project='').
  --project and --global combined is an error.

Examples:
  acho context --project
  acho context --global
  acho context --project=puntopost-backoffice
`
}

func (c *contextLoad) Run(args []string) error {
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

	project := ""
	switch state {
	case "bare":
		project = d.Project
	case "value":
		project = value
	}

	rules, err := d.Rules.List(project, global, rule.ListQuery{})
	if err != nil {
		return err
	}
	types, err := d.Types.List(project, global, rtype.ListQuery{})
	if err != nil {
		return err
	}

	fmt.Print(renderRuleContext(rules, types))
	return nil
}

func renderRuleContext(rules []rule.Rule, types []rtype.RType) string {
	out := fmt.Sprintf("%s%s==MANDATORY==%s\n\n", term.T.Bold(), term.T.Primary(), term.T.Reset())
	out += fmt.Sprintf("%sRules:%s\n", term.T.Bold(), term.T.Reset())

	if len(rules) == 0 {
		out += fmt.Sprintf("%s(no rules yet)%s\n\n", term.T.Muted(), term.T.Reset())
	} else {
		for _, r := range rules {
			label := "global"
			if !r.IsGlobal() {
				label = fmt.Sprintf("project:%s", r.Project)
			}
			out += fmt.Sprintf("%s[%s]%s %s%s%s\n%s\n\n",
				term.T.Secondary(), label, term.T.Reset(),
				term.T.Bold(), r.Title, term.T.Reset(),
				r.Text,
			)
		}
	}

	out += fmt.Sprintf("%sTypes:%s\n", term.T.Bold(), term.T.Reset())
	if len(types) == 0 {
		out += fmt.Sprintf("%sNo registry types defined. Define at least one with type_create before saving registries.%s\n\n", term.T.Muted(), term.T.Reset())
	} else {
		for _, rt := range types {
			label := "global"
			if !rt.IsGlobal() {
				label = fmt.Sprintf("project:%s", rt.Project)
			}
			out += fmt.Sprintf("%s[%s]%s %s%s%s\n%s\n\n",
				term.T.Secondary(), label, term.T.Reset(),
				term.T.Bold(), rt.Name, term.T.Reset(),
				rt.Schema,
			)
		}
	}

	out += fmt.Sprintf("%sIf any rules above are contradictory, ask the user how to resolve the inconsistency before proceeding.%s\n", term.T.Muted(), term.T.Reset())
	out += fmt.Sprintf("\n%s%s==END==%s\n", term.T.Bold(), term.T.Primary(), term.T.Reset())
	return out
}
