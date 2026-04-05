package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ Tool = (*typeDelete)(nil)

func init() {
	RegisterTool(&typeDelete{})
}

type typeDelete struct{}

func (t *typeDelete) Instruction() string {
	return "type_delete — delete a type (force=true to cascade-delete its registries)"
}

type TypeDeleteInput struct {
	Name  string `json:"name" jsonschema:"Type name to delete"`
	Force bool   `json:"force,omitempty" jsonschema:"Required when registries use this type; cascades and deletes them"`
}

type TypeDeleteOutput struct {
	Message           string `json:"message"`
	RegistriesRemoved int    `json:"registries_removed"`
}

func (t *typeDelete) Register(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "type_delete",
		Description: "Delete a type. Fails if registries use it unless force=true (which cascade-deletes them).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TypeDeleteInput) (*mcp.CallToolResult, TypeDeleteOutput, error) {
		start := logToolStart("type_delete", "name", input.Name, "force", input.Force)
		removed, err := deps.Types.Delete(input.Name, input.Force)
		if err != nil {
			logToolError("type_delete", start, err, "name", input.Name)
			return nil, TypeDeleteOutput{}, fmt.Errorf("type_delete failed: %w", err)
		}
		logToolSuccess("type_delete", start, "name", input.Name, "registries_removed", removed)
		msg := fmt.Sprintf("Deleted type %s", input.Name)
		if removed > 0 {
			msg = fmt.Sprintf("Deleted type %s and %d registries", input.Name, removed)
		}
		return nil, TypeDeleteOutput{Message: msg, RegistriesRemoved: removed}, nil
	})
}
