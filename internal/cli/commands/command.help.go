package commands

import (
	"fmt"
	"strings"
)

func init() {
	Register(&help{})
}

var _ Command = (*help)(nil)

type help struct{}

func (c *help) Usage() string       { return "acho --help" }
func (c *help) Description() string { return "Show this help" }
func (c *help) Order() int          { return 100 }
func (c *help) Help() string {
	return "Usage: acho --help\n\nShows the list of available commands.\n"
}

func (c *help) Match(name string) bool {
	return name == "--help" || name == "-h"
}

func (c *help) Run(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(args, " "))
	}

	PrintUsage()

	return nil
}
