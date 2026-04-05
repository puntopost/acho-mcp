package commands

import (
	"fmt"
	"strconv"

	"github.com/puntopost/acho-mcp/internal/cli/term"
	"github.com/puntopost/acho-mcp/internal/persistence/store"
)

func init() {
	Register(&list{})
}

var _ Command = (*list)(nil)

type list struct{}

func (c *list) Match(name string) bool { return name == "registries list" }
func (c *list) Usage() string          { return "acho registries list [flags]" }
func (c *list) Description() string    { return "List registries" }
func (c *list) Order() int             { return 32 }
func (c *list) Help() string {
	return `acho registries list — List registries

Usage:
  acho registries list [flags]

Filter semantics:
  (no flag)              All registries, across every project (no filter).
  --project              Auto-detected current project + globals.
  --project=NAME         Project NAME + globals.
  --global               Only global registries (project='').
  --project and --global combined is an error.

Deleted filter:
  (no flag)              Show active + soft-deleted.
  --active               Only active (deleted = 0).
  --deleted              Only soft-deleted (deleted = 1).

Other flags:
  --limit=N              Max results (default: from config)
  --offset=N             Skip first N results

Results are sorted by date desc. Content is shown truncated; use
'acho registries get <id>' for full content.

Examples:
  acho registries list                          # everything, all projects
  acho registries list --project                # current project + globals
  acho registries list --project=puntopost      # puntopost + globals
  acho registries list --global                 # only globals
  acho registries list --deleted                # only soft-deleted
`
}

func (c *list) Run(args []string) error {
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

	query := store.ListQuery{}
	switch {
	case onlyDeleted:
		query.OnlyDeleted = true
	case !onlyActive:
		query.IncludeDeleted = true
	}

	switch {
	case global:
		query.Global = true
	case state == "bare":
		query.Project = d.Project
	case state == "value":
		query.Project = value
	}

	if v := flagValue(args, "--limit"); v != "" {
		query.Limit, _ = strconv.Atoi(v)
	}
	if v := flagValue(args, "--offset"); v != "" {
		query.Offset, _ = strconv.Atoi(v)
	}

	results, err := d.Service.List(query)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Printf("%sNo registries found.%s\n", term.T.Muted(), term.T.Reset())
		return nil
	}

	fmt.Printf("%s%sFound %d registries:%s\n\n", term.T.Bold(), term.T.Primary(), len(results), term.T.Reset())
	for _, item := range results {
		printItemPreview(item)
	}
	return nil
}
