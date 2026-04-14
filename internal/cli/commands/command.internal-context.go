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

	body := renderInternalContextMarkdown(renderInternalMandatoryBlock(rules, types))

	fmt.Print(a.FormatContext(body))
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

func renderInternalMandatoryBlock(rules []rule.Rule, types []rtype.RType) string {
	var b strings.Builder
	b.WriteString("==MANDATORY==\n\n")
	b.WriteString("Rules:\n")

	if len(rules) == 0 {
		b.WriteString("(no rules yet)\n\n")
	} else {
		for _, r := range rules {
			label := "global"
			if !r.IsGlobal() {
				label = "project:" + r.Project
			}
			fmt.Fprintf(&b, "[%s] %s (id: %s)\n%s\n\n", label, r.Title, r.ID, r.Text)
		}
	}

	b.WriteString("Types:\n")
	if len(types) == 0 {
		b.WriteString("No registry types defined. Ask the user to define at least one with type_create before saving registries.\n\n")
	} else {
		for _, rt := range types {
			label := "global"
			if !rt.IsGlobal() {
				label = "project:" + rt.Project
			}
			fmt.Fprintf(&b, "[%s] %s\n%s\n\n", label, rt.Name, rt.Schema)
		}
	}

	b.WriteString("If any rules above are contradictory, ask the user how to resolve the inconsistency before proceeding.\n\n")
	b.WriteString("==END==\n")
	return b.String()
}

func renderInternalContextMarkdown(mandatory string) string {
	var b strings.Builder
	b.WriteString("# Acho Persistent Memory\n\n")
	b.WriteString("Persistent memory across sessions and compactions. Three kinds of objects:\n\n")
	b.WriteString("- **Rules** — free-text mandatory instructions. Delivered below in the ==MANDATORY== block.\n")
	b.WriteString("- **Types** — user-defined JSON Schemas. Nothing can be saved until a matching type exists.\n")
	b.WriteString("- **Registries** — JSON objects that must validate against their type's schema.\n\n")
	b.WriteString("Tool names, parameters and descriptions are served by the MCP protocol (`tools/list`). Read them there, do not guess.\n\n")
	b.WriteString(mandatory)
	b.WriteString("\n")
	b.WriteString("## Onboarding — first use on a project\n\n")
	b.WriteString("If no types are listed above, the user has not set them up yet. Ask what kinds of things they want to store and propose a schema for each. Do not invent schemas silently — always get approval before calling `type_create`.\n\n")
	b.WriteString("## When to save — proactively\n\n")
	b.WriteString("Save immediately after: user preferences or corrections, architecture decisions, bug fixes with non-obvious cause, or discoveries the user would want to recall later.\n\n")
	b.WriteString("## When NOT to save\n\n")
	b.WriteString("Do not save: trivial current-turn info, facts already in code or git history, intermediate steps, reasoning chains, or unverified assumptions.\n\n")
	b.WriteString("## Writing registries\n\n")
	b.WriteString("One idea per registry. Title = searchable keywords, not a sentence. Content = concise declarative JSON matching the schema.\n\n")
	b.WriteString("## Reading with sql_query\n\n")
	b.WriteString("Exposed objects (pre-filtered to the current project): `v_registries`, `v_types`, `v_rules`, and the FTS5 index `registry_fts` (join by `rowid` to `v_registries`; `bm25()` / `snippet()` available). The full schema, columns and examples are in the `sql_query` tool description. Raw tables and writes are blocked.\n\n")
	b.WriteString("## MANDATORY\n\n")
	b.WriteString("- Follow every rule in the ==MANDATORY== block. If rules contradict, ask the user how to resolve.\n")
	b.WriteString("- Never run the `acho` CLI via Bash unless the user explicitly asks.\n")
	b.WriteString("- If `rule`, `registry`, `type`, or `project` are ambiguous, prefer the current repository meaning unless the user clearly refers to the Acho plugin.\n")
	return b.String()
}
