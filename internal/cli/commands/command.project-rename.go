package commands

import (
	"fmt"
	"os"
	"sort"

	"github.com/puntopost/acho-mcp/internal/cli"
	"github.com/puntopost/acho-mcp/internal/cli/config"
	"github.com/puntopost/acho-mcp/internal/cli/term"
	"github.com/puntopost/acho-mcp/internal/persistence/rtype"
	"github.com/puntopost/acho-mcp/internal/persistence/rule"
)

func init() {
	Register(&projectRename{})
}

var _ Command = (*projectRename)(nil)

type projectRename struct{}

func (c *projectRename) Match(name string) bool { return name == "project rename" }
func (c *projectRename) Usage() string          { return "acho project rename" }
func (c *projectRename) Description() string    { return "Rename a project to the current one" }
func (c *projectRename) Order() int             { return 51 }
func (c *projectRename) Help() string {
	return `acho project rename — Rename a project to the current one

Usage:
  acho project rename

Lists all projects that have registries. Select one to rename — all its
rules, types and registries will be updated to use the current project
name (detected from git remote or full directory path).

Useful when you moved a directory, added a git remote, or changed the
project slug and old registries are orphaned under the previous name.
`
}

func (c *projectRename) Run(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments")
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get working directory: %w", err)
	}
	currentProject := detectProject(dir)

	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	stats, err := d.Service.Stats()
	if err != nil {
		return err
	}

	// Collect projects across registries, rules and types.
	seen := make(map[string]struct{})
	for p := range stats.ByProject {
		if p == "(global)" {
			continue
		}
		seen[p] = struct{}{}
	}
	rules, err := d.Rules.List("", false, rule.ListQuery{})
	if err != nil {
		return err
	}
	for _, r := range rules {
		if r.Project != "" {
			seen[r.Project] = struct{}{}
		}
	}
	types, err := d.Types.List("", false, rtype.ListQuery{})
	if err != nil {
		return err
	}
	for _, t := range types {
		if t.Project != "" {
			seen[t.Project] = struct{}{}
		}
	}

	if len(seen) == 0 {
		fmt.Printf("%sNo projects found.%s\n", term.T.Muted(), term.T.Reset())
		return nil
	}

	projects := make([]string, 0, len(seen))
	for p := range seen {
		projects = append(projects, p)
	}
	sort.Strings(projects)

	fmt.Printf("%s%sCurrent project:%s %s\n\n", term.T.Bold(), term.T.Primary(), term.T.Reset(), currentProject)
	fmt.Printf("%s%sProjects with data:%s\n", term.T.Bold(), term.T.Primary(), term.T.Reset())
	for i, p := range projects {
		marker := ""
		if p == currentProject {
			marker = fmt.Sprintf(" %s(current)%s", term.T.Muted(), term.T.Reset())
		}
		fmt.Printf("  %s%d.%s %s%s%s%s\n",
			term.T.Secondary(), i+1, term.T.Reset(),
			term.T.Bold(), p, term.T.Reset(),
			marker)
	}

	fmt.Printf("\n%sSelect project to rename (or q to cancel):%s ", term.T.Primary(), term.T.Reset())
	input, err := cli.PromptLine()
	if err != nil {
		return err
	}
	if input == "q" || input == "" {
		return cli.ErrCancelled
	}

	// Try as number
	idx := 0
	fmt.Sscanf(input, "%d", &idx)
	var oldProject string
	if idx >= 1 && idx <= len(projects) {
		oldProject = projects[idx-1]
	} else {
		// Try as name
		for _, p := range projects {
			if p == input {
				oldProject = p
				break
			}
		}
	}

	if oldProject == "" {
		return fmt.Errorf("invalid selection: %s", input)
	}

	if oldProject == currentProject {
		fmt.Printf("%sAlready the current project, nothing to rename.%s\n", term.T.Muted(), term.T.Reset())
		return nil
	}

	var nRules, nTypes, nRegs int
	configUpdated := false

	if err := config.WithConfigLock(func() error {
		var err error
		nRules, err = d.Rules.RenameProject(oldProject, currentProject)
		if err != nil {
			return err
		}
		nTypes, err = d.Types.RenameProject(oldProject, currentProject)
		if err != nil {
			return err
		}
		nRegs, err = d.Service.RenameProject(oldProject, currentProject)
		if err != nil {
			return err
		}

		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if cfg.IsProjectEnabled(oldProject) {
			cfg.EnabledProjects, _ = removeProject(cfg.EnabledProjects, oldProject)
			cfg.EnabledProjects, _ = addProject(cfg.EnabledProjects, currentProject)
			if err := saveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			configUpdated = true
		}
		return nil
	}); err != nil {
		return err
	}

	configNote := ""
	if configUpdated {
		configNote = "; enabled project updated in config"
	}
	fmt.Printf("%s Renamed %s%s%s → %s%s%s (%d rules, %d types, %d registries updated%s)\n",
		term.T.Success("Done!"),
		term.T.Bold(), oldProject, term.T.Reset(),
		term.T.Bold(), currentProject, term.T.Reset(),
		nRules, nTypes, nRegs, configNote)
	return nil
}
