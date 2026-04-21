package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/puntopost/acho-mcp/internal/cli"
	"github.com/puntopost/acho-mcp/internal/cli/term"
	"github.com/puntopost/acho-mcp/internal/persistence/rtype"
)

func init() {
	Register(&typeList{})
}

var jsonKeyLineRe = regexp.MustCompile(`^(\s*)"([^"]+)":(.*)$`)

var _ Command = (*typeList)(nil)

type typeList struct{}

func (c *typeList) Match(name string) bool { return name == "types list" }
func (c *typeList) Usage() string          { return "acho types list [flags]" }
func (c *typeList) Description() string    { return "List registry types" }
func (c *typeList) Order() int             { return 39 }
func (c *typeList) Help() string {
	return `acho types list — List registry types

Filter semantics:
  (no flag)              All types, across every project.
  --project              Auto-detected current project + globals.
  --project=NAME         Project NAME + globals.
  --global               Only global types (project='').
  --project and --global combined is an error.

Deleted filter:
  (no flag)              Show active + soft-deleted.
  --active               Only active (deleted = 0).
  --deleted              Only soft-deleted (deleted = 1).

Examples:
  acho types list
  acho types list --project
  acho types list --global
  acho types list --deleted
`
}

func (c *typeList) Run(args []string) error {
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

	q := rtype.ListQuery{}
	switch {
	case onlyDeleted:
		q.OnlyDeleted = true
	case !onlyActive:
		q.IncludeDeleted = true
	}

	types, err := d.Types.List(project, global, q)
	if err != nil {
		return err
	}

	if len(types) == 0 {
		fmt.Printf("%sNo types defined.%s\n", term.T.Muted(), term.T.Reset())
		return nil
	}

	fmt.Printf("%s%sFound %d types:%s\n\n", term.T.Bold(), term.T.Primary(), len(types), term.T.Reset())
	for _, t := range types {
		printType(t)
	}
	return nil
}

func printType(t rtype.RType) {
	label := "global"
	if !t.IsGlobal() {
		label = fmt.Sprintf("project:%s", t.Project)
	}
	deletedTag := ""
	if t.Deleted {
		d := ""
		if t.DeletedDate != nil {
			d = " at " + t.DeletedDate.Format(cli.DateYMD_HM)
		}
		deletedTag = fmt.Sprintf(" %s[DELETED%s]%s", term.T.Danger(), d, term.T.Reset())
	}
	fmt.Printf("%s%s%s (%s%s%s)%s\n  %sdescription:%s %s\n  %sschema:%s\n%s\n  %sdate: %s%s\n\n",
		term.T.Bold(), t.Name, term.T.Reset(),
		term.T.Primary(), label, term.T.Reset(), deletedTag,
		term.T.Muted(), term.T.Reset(), t.Description,
		term.T.Muted(), term.T.Reset(), indentBlock(formatJSON(t.Schema), "    "),
		term.T.Muted(), t.Date.Format(cli.DateYMD_HM), term.T.Reset(),
	)
}

func formatJSON(raw string) string {
	var out bytes.Buffer
	if err := json.Indent(&out, []byte(raw), "", "  "); err != nil {
		return raw
	}
	return colorizeJSONKeys(out.String())
}

func indentBlock(s, prefix string) string {
	if s == "" {
		return prefix
	}
	lines := bytes.Split([]byte(s), []byte("\n"))
	out := make([]byte, 0, len(s)+len(lines)*len(prefix))
	for i, line := range lines {
		if i > 0 {
			out = append(out, '\n')
		}
		out = append(out, prefix...)
		out = append(out, line...)
	}
	return string(out)
}

func colorizeJSONKeys(s string) string {
	lines := bytes.Split([]byte(s), []byte("\n"))
	out := make([]byte, 0, len(s)+len(lines)*16)
	for i, line := range lines {
		if i > 0 {
			out = append(out, '\n')
		}
		m := jsonKeyLineRe.FindSubmatch(line)
		if m == nil {
			out = append(out, line...)
			continue
		}
		out = append(out, m[1]...)
		out = append(out, term.T.Secondary()...)
		out = append(out, '"')
		out = append(out, m[2]...)
		out = append(out, '"', ':')
		out = append(out, term.T.Reset()...)
		out = append(out, m[3]...)
	}
	return string(out)
}
