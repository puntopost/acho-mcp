package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ Tool = (*registryGet)(nil)

func init() {
	RegisterTool(&registryGet{})
}

type registryGet struct{}

func (t *registryGet) Instruction() string {
	return "registry_get — get the full content of a registry by ID"
}

type RegistryGetInput struct {
	ID string `json:"id" jsonschema:"Registry ID"`
}

type RegistryGetOutput struct {
	Registry string `json:"registry"`
}

func (t *registryGet) Register(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "registry_get",
		Description: "Get the full content of a registry by ID. Increments get_hits.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RegistryGetInput) (*mcp.CallToolResult, RegistryGetOutput, error) {
		start := logToolStart("registry_get", "id", input.ID)
		r, err := deps.Service.Get(input.ID)
		if err != nil {
			logToolError("registry_get", start, err, "id", input.ID)
			return nil, RegistryGetOutput{}, fmt.Errorf("registry_get failed: %w", err)
		}
		logToolSuccess("registry_get", start, "id", input.ID, "type", r.Type, "project", r.Project)
		return nil, RegistryGetOutput{Registry: formatRegistryFull(r)}, nil
	})
}
