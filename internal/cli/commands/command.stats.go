package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/puntopost/acho-mcp/internal/cli/term"
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

No arguments. Shows totals grouped by project and type.

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

	stats, err := d.Service.Stats()
	if err != nil {
		return err
	}

	box := term.NewBox(52)
	box.Blank()
	box.Title("Acho Stats")
	box.Separator()

	// Registries
	box.Blank()
	box.Section(fmt.Sprintf("Registries (%d total)", stats.Total))

	if len(stats.ByProject) > 0 {
		box.Blank()
		box.Section("By project")
		maxVal := maxCount(stats.ByProject)
		labelW := maxLabelWidth(stats.ByProject)
		for _, kv := range sortDesc(stats.ByProject) {
			box.Bar(kv.key, kv.val, maxVal, labelW)
		}
	}

	if len(stats.ByType) > 0 {
		box.Blank()
		box.Section("By type")
		maxVal := maxCount(stats.ByType)
		labelW := maxLabelWidth(stats.ByType)
		for _, kv := range sortDesc(stats.ByType) {
			box.Bar(kv.key, kv.val, maxVal, labelW)
		}
	}

	box.Blank()
	fmt.Print(box.String())
	return nil
}

type kv struct {
	key string
	val int
}

func sortDesc(m map[string]int) []kv {
	pairs := make([]kv, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].val > pairs[j].val
	})
	return pairs
}

func maxCount(m map[string]int) int {
	max := 0
	for _, v := range m {
		if v > max {
			max = v
		}
	}
	return max
}

func maxLabelWidth(m map[string]int) int {
	max := 0
	for k := range m {
		if len(k) > max {
			max = len(k)
		}
	}
	return max
}
