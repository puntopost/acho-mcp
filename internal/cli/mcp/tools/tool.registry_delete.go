package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ Tool = (*registryDelete)(nil)

func init() {
	RegisterTool(&registryDelete{})
}

type registryDelete struct{}

func (t *registryDelete) Instruction() string {
	return "registry_delete — delete a registry by ID"
}

type RegistryDeleteInput struct {
	ID string `json:"id" jsonschema:"Registry ID to delete"`
}

type RegistryDeleteOutput struct {
	Message string `json:"message"`
}

func (t *registryDelete) Register(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "registry_delete",
		Description: "Delete a registry by ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RegistryDeleteInput) (*mcp.CallToolResult, RegistryDeleteOutput, error) {
		start := logToolStart("registry_delete", "id", input.ID)
		if err := deps.Service.Delete(input.ID); err != nil {
			logToolError("registry_delete", start, err, "id", input.ID)
			return nil, RegistryDeleteOutput{}, fmt.Errorf("registry_delete failed: %w", err)
		}
		logToolSuccess("registry_delete", start, "id", input.ID)
		return nil, RegistryDeleteOutput{Message: fmt.Sprintf("Deleted registry %s", input.ID)}, nil
	})
}
