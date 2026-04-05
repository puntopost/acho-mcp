package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ Tool = (*registryUpdate)(nil)

func init() {
	RegisterTool(&registryUpdate{})
}

type registryUpdate struct{}

func (t *registryUpdate) Instruction() string {
	return "registry_update — update an existing registry by ID (only provided fields change)"
}

type RegistryUpdateInput struct {
	ID      string `json:"id" jsonschema:"Registry ID to update"`
	Title   string `json:"title,omitempty" jsonschema:"New title (omit to leave unchanged)"`
	Content string `json:"content,omitempty" jsonschema:"New content: JSON object matching the type's schema (omit to leave unchanged)"`
	Type    string `json:"type,omitempty" jsonschema:"New type name (omit to leave unchanged)"`
	Project string `json:"project,omitempty" jsonschema:"\"current\" or \"global\" to move the registry; omit to leave unchanged"`
}

type RegistryUpdateOutput struct {
	Message string `json:"message"`
}

func (t *registryUpdate) Register(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "registry_update",
		Description: "Update an existing registry by ID. Only provided fields are changed.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RegistryUpdateInput) (*mcp.CallToolResult, RegistryUpdateOutput, error) {
		start := logToolStart("registry_update",
			"id", input.ID,
			"title_set", input.Title != "",
			"content_len", len(input.Content),
			"type", input.Type,
			"project_set", input.Project != "",
		)
		var title, content, typ, project *string
		if input.Title != "" {
			title = &input.Title
		}
		if input.Content != "" {
			content = &input.Content
		}
		if input.Type != "" {
			typ = &input.Type
		}
		if input.Project != "" {
			p, err := resolveProject(input.Project, deps.Config.DefaultProject)
			if err != nil {
				return nil, RegistryUpdateOutput{}, err
			}
			project = &p
		}

		if err := deps.Service.Update(input.ID, title, content, typ, project); err != nil {
			logToolError("registry_update", start, err, "id", input.ID)
			return nil, RegistryUpdateOutput{}, fmt.Errorf("registry_update failed: %w", err)
		}
		logToolSuccess("registry_update", start, "id", input.ID)
		return nil, RegistryUpdateOutput{Message: fmt.Sprintf("Updated registry %s", input.ID)}, nil
	})
}
