package mcp

import (
	"database/sql"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/puntopost/acho-mcp/internal/cli/mcp/tools"
	"github.com/puntopost/acho-mcp/internal/service"
)

const instructionsHeader = `Acho is persistent memory for coding agents. It survives across sessions and compactions.

INTERFACES:
- MCP tools are for agents. Use MCP semantics and defaults when working as an agent.
- CLI commands are for humans. Do not infer MCP behavior from CLI help or defaults.

RULES:
- Call context at the start of every session to load mandatory rules.
- Before saving a registry, ensure its type exists (use type_create if needed).
- Use rule_create / rule_update / rule_delete when the user asks to add, change, or remove a rule.
- Use sql_query for reads (search/list/filter/aggregate). It exposes v_registries
  and v_types, pre-filtered to the current project; rules are not queryable via SQL
  (they come via context).
- Use MCP tools, not CLI commands, unless the user explicitly asks for a CLI command.`

type Config struct {
	DefaultProject string
}

func NewServer(cfg Config, db *sql.DB, svc *service.StoreService, rules *service.RuleService, types *service.RTypeService, version string) *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{Name: "acho", Version: version},
		&mcp.ServerOptions{Instructions: instructionsHeader},
	)

	tools.RegisterAll(server, tools.Deps{
		Config:  tools.Config{DefaultProject: cfg.DefaultProject},
		DB:      db,
		Service: svc,
		Rules:   rules,
		Types:   types,
	})

	return server
}
