package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/puntopost/acho-mcp/internal/cli/term"
	"github.com/puntopost/acho-mcp/internal/persistence"
)

func init() {
	Register(&stats{})
}

var _ Command = (*stats)(nil)

type stats struct{}

func (c *stats) Match(name string) bool { return name == "stats" }
func (c *stats) Usage() string          { return "acho stats" }
func (c *stats) Description() string    { return "Show memory statistics" }
func (c *stats) Order() int             { return 37 }
func (c *stats) Help() string {
	return `acho stats — Show memory statistics

Usage:
  acho stats

No arguments. Shows active/deleted counts for registries, rules and types,
broken down by project (and by type for registries).

Examples:
  acho stats
`
}

func (c *stats) Run(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(args, " "))
	}

	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	regStats, err := d.Service.Stats()
	if err != nil {
		return err
	}
	ruleStats, err := d.Rules.Stats()
	if err != nil {
		return err
	}
	typeStats, err := d.Types.Stats()
	if err != nil {
		return err
	}

	box := term.NewBox(60)
	box.Blank()
	box.Title("Acho Stats")
	box.Separator()

	section(box, "Registries", regStats.TotalActive, regStats.TotalDeleted,
		regStats.ByProject, regStats.ByType)

	section(box, "Rules", ruleStats.TotalActive, ruleStats.TotalDeleted,
		ruleStats.ByProject, nil)

	section(box, "Types", typeStats.TotalActive, typeStats.TotalDeleted,
		typeStats.ByProject, nil)

	box.Blank()
	fmt.Print(box.String())
	return nil
}

func section(box *term.Box, name string, active, deleted int, byProject, byType map[string]persistence.Counts) {
	box.Blank()
	box.Section(fmt.Sprintf("%s (%s active / %s deleted)", name,
		term.T.Secondary()+term.T.Bold()+strconv.Itoa(active)+term.T.Reset(),
		term.T.Danger()+term.T.Bold()+strconv.Itoa(deleted)+term.T.Reset(),
	))

	if len(byProject) > 0 {
		box.Blank()
		box.Section("By project")
		drawBuckets(box, byProject)
	}
	if len(byType) > 0 {
		box.Blank()
		box.Section("By type")
		drawBuckets(box, byType)
	}
}

func drawBuckets(box *term.Box, m map[string]persistence.Counts) {
	pairs := sortCountsDesc(m)
	maxTotal := 0
	maxLabel := 0
	maxCount := 0
	for _, p := range pairs {
		if p.c.Total() > maxTotal {
			maxTotal = p.c.Total()
		}
		if len(p.key) > maxLabel {
			maxLabel = len(p.key)
		}
		if p.c.Active > maxCount {
			maxCount = p.c.Active
		}
		if p.c.Deleted > maxCount {
			maxCount = p.c.Deleted
		}
	}
	countW := len(strconv.Itoa(maxCount))
	if countW < 2 {
		countW = 2
	}
	for _, p := range pairs {
		box.SplitBar(p.key, p.c.Active, p.c.Deleted, maxTotal, maxLabel, countW)
	}
}

type countsKV struct {
	key string
	c   persistence.Counts
}

func sortCountsDesc(m map[string]persistence.Counts) []countsKV {
	out := make([]countsKV, 0, len(m))
	for k, v := range m {
		out = append(out, countsKV{k, v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].c.Total() != out[j].c.Total() {
			return out[i].c.Total() > out[j].c.Total()
		}
		return out[i].key < out[j].key
	})
	return out
}
