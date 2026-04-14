package mcp

import (
	"database/sql"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/puntopost/acho-mcp/internal/cli/mcp/tools"
	"github.com/puntopost/acho-mcp/internal/service"
)

const instructionsHeader = `Acho persistent memory. The plugin hook injects a ==MANDATORY== block at session start with the user's rules and types — follow it. If you do not see that block, the plugin is not installed: tell the user to run 'acho agent-setup'.`

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
