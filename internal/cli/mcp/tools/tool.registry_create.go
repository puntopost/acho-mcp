package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ Tool = (*registryCreate)(nil)

func init() {
	RegisterTool(&registryCreate{})
}

type registryCreate struct{}

func (t *registryCreate) Instruction() string {
	return "registry_create — create a new registry (content must validate against the type's schema)"
}

type RegistryCreateInput struct {
	Title   string `json:"title" jsonschema:"Short searchable title"`
	Content string `json:"content" jsonschema:"Content as a JSON object matching the type's schema"`
	Type    string `json:"type" jsonschema:"Name of an existing registry type"`
	Project string `json:"project" jsonschema:"\"current\" or \"global\""`
}

type RegistryCreateOutput struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func (t *registryCreate) Register(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "registry_create",
		Description: "Create a new registry. Call PROACTIVELY after decisions, bug fixes, discoveries. project must be \"current\" or \"global\".",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RegistryCreateInput) (*mcp.CallToolResult, RegistryCreateOutput, error) {
		project, err := resolveProject(input.Project, deps.Config.DefaultProject)
		if err != nil {
			return nil, RegistryCreateOutput{}, err
		}
		start := logToolStart("registry_create",
			"title", input.Title,
			"type", input.Type,
			"project", project,
			"content_len", len(input.Content),
		)

		id, err := deps.Service.Save(input.Title, input.Content, input.Type, project)
		if err != nil {
			logToolError("registry_create", start, err, "title", input.Title, "project", project)
			return nil, RegistryCreateOutput{}, fmt.Errorf("registry_create failed: %w", err)
		}
		logToolSuccess("registry_create", start, "id", id, "project", project)
		return nil, RegistryCreateOutput{ID: id, Message: fmt.Sprintf("Created registry %s", id)}, nil
	})
}
