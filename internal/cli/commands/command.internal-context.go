package commands

import (
	"fmt"
	"strings"

	"github.com/puntopost/acho-mcp/internal/cli/agent"
	"github.com/puntopost/acho-mcp/internal/persistence/rtype"
	"github.com/puntopost/acho-mcp/internal/persistence/rule"
)

func init() {
	Register(&internalContext{})
}

type internalContext struct{}

func (c *internalContext) Match(name string) bool { return name == "internal context" }
func (c *internalContext) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: acho internal context <agent>")
	}
	a, err := agent.Get(args[0])
	if err != nil {
		return err
	}

	enabled, err := internalContextProjectEnabled(args)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	rules, err := d.Rules.List(d.Project, false, rule.ListQuery{})
	if err != nil {
		return err
	}
	types, err := d.Types.List(d.Project, false, rtype.ListQuery{})
	if err != nil {
		return err
	}

	fmt.Print(a.FormatContext(renderInternalContextMarkdown(rules, types)))
	return nil
}
func (c *internalContext) Usage() string       { return "" }
func (c *internalContext) Description() string { return "" }
func (c *internalContext) Help() string        { return "" }
func (c *internalContext) Order() int          { return 1000 }
func (c *internalContext) Hidden() bool        { return true }

func internalContextProjectEnabled(args []string) (bool, error) {
	cfg, err := loadConfig()
	if err != nil {
		return false, fmt.Errorf("failed to load config: %w", err)
	}

	projectName, err := resolveProjectName(args)
	if err != nil {
		return false, err
	}

	return cfg.IsProjectEnabled(projectName), nil
}

func renderInternalContextMarkdown(rules []rule.Rule, types []rtype.RType) string {
	var b strings.Builder
	b.WriteString("# Acho Persistent Memory\n\n")
	b.WriteString("Persistent memory across sessions and compactions. Three kinds of objects:\n\n")
	b.WriteString("- **Rules** — free-text mandatory instructions.\n")
	b.WriteString("- **Types** — user-defined JSON Schemas. Nothing can be saved until a matching type exists.\n")
	b.WriteString("- **Registries** — JSON objects that must validate against their type's schema.\n\n")
	b.WriteString("Tool names, parameters and descriptions are served by the MCP protocol. Read them there, do not guess.\n\n")
	b.WriteString("## Reading with sql_query\n\n")
	b.WriteString("Use the `sql_query` tool to read. See its description for the allowed views, columns, FTS index and examples.\n\n")

	b.WriteString("==MANDATORY==\n\n")
	b.WriteString("## Rules\n")
	if len(rules) == 0 {
		b.WriteString("- No rules defined. Ask the user to define at least one with rule_create before proceeding.\n\n")
	} else {
		for _, r := range rules {
			label := "global"
			if !r.IsGlobal() {
				label = "project:" + r.Project
			}
			fmt.Fprintf(&b, "- [%s] %s (id: %s)\n%s\n\n", label, r.Title, r.ID, r.Text)
		}
	}
	b.WriteString("Follow all the rules strictly. If any rules above are contradictory, ask the user how to resolve the inconsistency before proceeding.\n\n")

	b.WriteString("## Registry types\n")
	if len(types) == 0 {
		b.WriteString("- No registry types defined. Ask the user to define at least one with type_create before saving registries.\n\n")
	} else {
		for _, rt := range types {
			label := "global"
			if !rt.IsGlobal() {
				label = "project:" + rt.Project
			}
			fmt.Fprintf(&b, "- [%s] %s - %s\n", label, rt.Name, rt.Description)
		}
		b.WriteString("\nUse `sql_query` on `v_types` to read the full JSON Schema before creating or updating registries.\n\n")
	}
	b.WriteString("- Do not invent type schemas silently — always get approval before calling `type_create`.\n\n")

	b.WriteString("## Registries\n")
	b.WriteString("- One idea per registry. Title = searchable keywords, not a sentence. Content = concise declarative JSON matching the schema.\n\n")

	b.WriteString("## Other requirements\n")
	b.WriteString("- Never run the `acho` CLI via Bash unless the user explicitly asks.\n")
	b.WriteString("- If `rule`, `registry`, `type`, or `project` are ambiguous, prefer the current repository meaning unless the user clearly refers to the Acho plugin.\n\n")
	b.WriteString("==END==\n\n")
	return b.String()
}
