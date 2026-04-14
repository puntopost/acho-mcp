package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/puntopost/acho-mcp/internal/app"
	achomcp "github.com/puntopost/acho-mcp/internal/cli/mcp"
)

func init() {
	Register(&mcpStart{})
}

var _ Command = (*mcpStart)(nil)

type mcpStart struct{}

func (c *mcpStart) Usage() string       { return "acho mcp" }
func (c *mcpStart) Description() string { return "Start MCP server (stdio)" }
func (c *mcpStart) Order() int          { return 10 }
func (c *mcpStart) Help() string {
	return `acho mcp — Start MCP server (stdio transport)

Usage:
  acho mcp

The project name is always auto-detected from:
  1. Git remote origin (repo name)
  2. Full current directory path
Always normalized to lowercase slug (a-z, 0-9, hyphens).

Starts acho as an MCP server communicating via stdin/stdout (NDJSON).
This is what Claude Code, Cursor, Codex, and other MCP clients call.

Register with an agent using: acho agent-setup <claude|opencode>

Examples:
  acho mcp
`
}

func (c *mcpStart) Match(name string) bool {
	return name == "mcp"
}

func (c *mcpStart) Run(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %v", args)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get working directory: %w", err)
	}
	projectName := detectProject(cwd)
	if projectName == "" {
		return fmt.Errorf("could not detect project name from current directory")
	}

	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if !cfg.IsProjectEnabled(projectName) {
		slog.Info("mcp started disabled", "project", projectName, "version", Version)
		server := gomcp.NewServer(
			&gomcp.Implementation{Name: "acho", Version: Version},
			&gomcp.ServerOptions{Instructions: fmt.Sprintf(
				"Acho is disabled for project %q. No tools are exposed. Run `acho project enable --project=%s` and reconnect the MCP to activate.",
				projectName, projectName,
			)},
		)
		err = server.Run(ctx, &gomcp.StdioTransport{})
		slog.Info("mcp stopped")
		return err
	}

	d, err := app.Build(&cfg, projectName)
	if err != nil {
		return err
	}
	defer d.Close()

	server := achomcp.NewServer(achomcp.Config{
		DefaultProject: projectName,
	}, d.DB, d.Service, d.Rules, d.Types, Version)

	slog.Info("mcp started", "project", projectName, "version", Version)

	err = server.Run(ctx, &gomcp.StdioTransport{})

	slog.Info("mcp stopped")

	return err
}
