package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ Tool = (*typeCreate)(nil)

func init() {
	RegisterTool(&typeCreate{})
}

type typeCreate struct{}

func (t *typeCreate) Instruction() string {
	return "type_create — define a registry type by name and JSON Schema"
}

type TypeCreateInput struct {
	Name        string `json:"name" jsonschema:"Type name (slug ^[a-z][a-z_]*$, globally unique)"`
	Description string `json:"description" jsonschema:"Short description (max 300 chars) explaining when to use this type"`
	Schema      string `json:"schema" jsonschema:"JSON Schema (draft 2020-12) that content must match"`
	Project     string `json:"project" jsonschema:"\"current\" or \"global\""`
}

type TypeCreateOutput struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func (t *typeCreate) Register(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "type_create",
		Description: "Define a registry type with a JSON Schema. Types are immutable once created; to change, delete and recreate. project must be \"current\" or \"global\".",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TypeCreateInput) (*mcp.CallToolResult, TypeCreateOutput, error) {
		project, err := resolveProject(input.Project, deps.Config.DefaultProject)
		if err != nil {
			return nil, TypeCreateOutput{}, err
		}
		start := logToolStart("type_create", "name", input.Name, "project", project, "description_len", len(input.Description))

		if err := deps.Types.Create(input.Name, input.Description, input.Schema, project, time.Now().UTC()); err != nil {
			logToolError("type_create", start, err, "name", input.Name)
			return nil, TypeCreateOutput{}, fmt.Errorf("type_create failed: %w", err)
		}
		logToolSuccess("type_create", start, "name", input.Name)
		return nil, TypeCreateOutput{Name: input.Name, Message: fmt.Sprintf("Created type %s", input.Name)}, nil
	})
}
