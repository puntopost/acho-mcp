package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/oklog/ulid/v2"
)

var _ Tool = (*ruleCreate)(nil)

func init() {
	RegisterTool(&ruleCreate{})
}

type ruleCreate struct{}

func (t *ruleCreate) Instruction() string {
	return "rule_create — create a new rule (mandatory instruction injected via context)"
}

type RuleCreateInput struct {
	Title   string `json:"title" jsonschema:"Short title"`
	Text    string `json:"text" jsonschema:"Rule text (max 1000 characters)"`
	Project string `json:"project" jsonschema:"\"current\" (applies only to this project) or \"global\" (applies everywhere)"`
}

type RuleCreateOutput struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Mandatory string `json:"mandatory"`
}

func (t *ruleCreate) Register(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rule_create",
		Description: "Create a new rule. Rules are served by context at the start of every session as mandatory instructions. project must be \"current\" or \"global\".",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RuleCreateInput) (*mcp.CallToolResult, RuleCreateOutput, error) {
		project, err := resolveProject(input.Project, deps.Config.DefaultProject)
		if err != nil {
			return nil, RuleCreateOutput{}, err
		}
		start := logToolStart("rule_create", "title", input.Title, "project", project, "text_len", len(input.Text))

		id := ulid.Make().String()
		if err := deps.Rules.Upsert(id, input.Title, input.Text, project, time.Now().UTC()); err != nil {
			logToolError("rule_create", start, err, "id", id)
			return nil, RuleCreateOutput{}, fmt.Errorf("rule_create failed: %w", err)
		}
		logToolSuccess("rule_create", start, "id", id)
		return nil, RuleCreateOutput{
			ID:        id,
			Message:   fmt.Sprintf("Created rule %s", id),
			Mandatory: fmt.Sprintf("==MANDATORY==\nFollow this newly created rule immediately in the current session:\n- %s (id: %s)\n%s\n==END==", input.Title, id, input.Text),
		}, nil
	})
}
