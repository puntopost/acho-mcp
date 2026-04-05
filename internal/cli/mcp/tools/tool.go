package tools

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/puntopost/acho-mcp/internal/service"
)

type Config struct {
	DefaultProject string
}

type Deps struct {
	Config  Config
	DB      *sql.DB
	Service *service.StoreService
	Rules   *service.RuleService
	Types   *service.RTypeService
}

// Tool defines the interface for MCP tools.
type Tool interface {
	Instruction() string
	Register(server *mcp.Server, deps Deps)
}

var registry []Tool

func RegisterTool(t Tool) {
	registry = append(registry, t)
}

func RegisterAll(server *mcp.Server, deps Deps) {
	for _, t := range registry {
		t.Register(server, deps)
	}
}

func Instructions() string {
	var lines []string
	for _, t := range registry {
		lines = append(lines, fmt.Sprintf("  %s", t.Instruction()))
	}
	return strings.Join(lines, "\n")
}
