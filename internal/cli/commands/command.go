package commands

import (
	"fmt"
	"sort"

	"github.com/puntopost/acho-mcp/internal/cli/term"
)

var Version string

type Command interface {
	Match(name string) bool
	Run(args []string) error
	Usage() string
	Description() string
	Help() string
	Order() int
}

var registry []Command

func Register(cmd Command) {
	registry = append(registry, cmd)
	sort.Slice(registry, func(i, j int) bool {
		return registry[i].Order() < registry[j].Order()
	})
}

func All() []Command {
	return registry
}

func PrintUsage() {
	fmt.Print(term.Banner())
	fmt.Printf("%s%sUsage:%s\n", term.T.Bold(), term.T.Primary(), term.T.Reset())
	for _, cmd := range registry {
		fmt.Printf("  %s%-28s%s %s\n", term.T.Secondary(), cmd.Usage(), term.T.Reset(), cmd.Description())
	}
	fmt.Printf("\n%s%sGlobal flags:%s\n", term.T.Bold(), term.T.Primary(), term.T.Reset())
	fmt.Printf("  %s--config=PATH%s                Path to config file (default: ~/.acho/config.json)\n", term.T.Secondary(), term.T.Reset())
	fmt.Printf("\n%s%sEnvironment:%s\n", term.T.Bold(), term.T.Primary(), term.T.Reset())
	fmt.Printf("  %sACHO_PATH%s                    Override base directory (default: ~/.acho)\n", term.T.Secondary(), term.T.Reset())
}
