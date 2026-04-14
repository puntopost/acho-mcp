package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ Tool = (*typeRename)(nil)

func init() {
	RegisterTool(&typeRename{})
}

type typeRename struct{}

func (t *typeRename) Instruction() string {
	return "type_rename — rename a type (identity only; schema and scope unchanged)"
}

type TypeRenameInput struct {
	OldName string `json:"old_name" jsonschema:"Current type name"`
	NewName string `json:"new_name" jsonschema:"New type name (slug ^[a-z][a-z_]*$; must not exist, even soft-deleted)"`
}

type TypeRenameOutput struct {
	Message           string `json:"message"`
	RegistriesUpdated int    `json:"registries_updated"`
}

func (t *typeRename) Register(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "type_rename",
		Description: "Rename a registry type. Strict: fails if new_name already exists (including soft-deleted) or old_name is missing/deleted. Updates the type of all registries using it, including soft-deleted ones. Does not touch rules or registry content that mention the old name as free text.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TypeRenameInput) (*mcp.CallToolResult, TypeRenameOutput, error) {
		start := logToolStart("type_rename", "old_name", input.OldName, "new_name", input.NewName)

		updated, err := deps.Types.Rename(input.OldName, input.NewName)
		if err != nil {
			logToolError("type_rename", start, err, "old_name", input.OldName, "new_name", input.NewName)
			return nil, TypeRenameOutput{}, fmt.Errorf("type_rename failed: %w", err)
		}
		logToolSuccess("type_rename", start, "old_name", input.OldName, "new_name", input.NewName, "registries_updated", updated)
		return nil, TypeRenameOutput{
			Message:           fmt.Sprintf("Renamed type %s -> %s (%d registries updated)", input.OldName, input.NewName, updated),
			RegistriesUpdated: updated,
		}, nil
	})
}
