package commands

import (
	"fmt"

	"github.com/puntopost/acho-mcp/internal/cli"
	"github.com/puntopost/acho-mcp/internal/cli/agent"
	"github.com/puntopost/acho-mcp/internal/cli/term"
)

func init() {
	Register(&agentSetup{})
}

var _ Command = (*agentSetup)(nil)

type agentSetup struct{}

func (c *agentSetup) Match(name string) bool { return name == "agent-setup" }
func (c *agentSetup) Usage() string          { return "acho agent-setup [<agent>]" }
func (c *agentSetup) Description() string    { return "Register acho in an AI agent" }
func (c *agentSetup) Order() int             { return 20 }
func (c *agentSetup) Help() string {
	return `acho agent-setup — Register acho as MCP server in an AI agent

Usage:
  acho agent-setup <agent>      Register directly
  acho agent-setup              Interactive mode (prompts for agent)

Available agents: ` + agent.Names() + `

Registers acho as an MCP server in the selected agent's configuration.
Safe to run multiple times — checks if already registered first.

Examples:
  acho agent-setup claude
  acho agent-setup
`
}

func (c *agentSetup) Run(args []string) error {
	name := positionalArg(args, 0)

	if name == "" {
		var err error
		name, err = c.promptAgent()
		if err != nil {
			return err
		}
		if name == "" {
			return fmt.Errorf("no agent selected")
		}
	}

	a, err := agent.Get(name)
	if err != nil {
		return err
	}

	if err := a.Setup(); err != nil {
		return err
	}

	fmt.Printf("%s acho registered in %s%s%s\n",
		term.T.Success("Done!"), term.T.Bold(), a.Name(), term.T.Reset())
	return nil
}

func (c *agentSetup) promptAgent() (string, error) {
	agents := agent.All()
	if len(agents) == 0 {
		return "", nil
	}

	fmt.Printf("%s%sAvailable agents:%s\n", term.T.Bold(), term.T.Primary(), term.T.Reset())
	for i, a := range agents {
		fmt.Printf("  %s%d.%s %s%s%s — %s\n",
			term.T.Secondary(), i+1, term.T.Reset(),
			term.T.Bold(), a.Name(), term.T.Reset(),
			a.Description())
	}
	fmt.Printf("%sSelect:%s ", term.T.Primary(), term.T.Reset())

	input, err := cli.PromptLine()
	if err != nil {
		return "", err
	}

	// Try as number first
	idx := 0
	fmt.Sscanf(input, "%d", &idx)
	if idx >= 1 && idx <= len(agents) {
		return agents[idx-1].Name(), nil
	}

	// Try as name
	if input != "" {
		return input, nil
	}

	return "", nil
}
