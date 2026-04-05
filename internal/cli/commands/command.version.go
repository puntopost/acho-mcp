package commands

import (
	"fmt"
	"strings"

	"github.com/puntopost/acho-mcp/internal/cli/term"
)

func init() {
	Register(&version{})
}

var _ Command = (*version)(nil)

type version struct{}

func (c *version) Usage() string       { return "acho --version" }
func (c *version) Description() string { return "Show version" }
func (c *version) Order() int          { return 90 }
func (c *version) Help() string {
	return "Usage: acho --version\n\nShows the current acho version.\n"
}

func (c *version) Match(name string) bool {
	return name == "--version" || name == "-v"
}

func (c *version) Run(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(args, " "))
	}

	fmt.Printf("%s%sacho%s %s%s%s%s\n", term.T.Bold(), term.T.Primary(), term.T.Reset(), term.T.Bold(), term.T.Secondary(), Version, term.T.Reset())

	return nil
}
