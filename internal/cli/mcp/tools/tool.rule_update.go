package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ Tool = (*ruleUpdate)(nil)

func init() {
	RegisterTool(&ruleUpdate{})
}

type ruleUpdate struct{}

func (t *ruleUpdate) Instruction() string {
	return "rule_update — update an existing rule (only provided fields change)"
}

type RuleUpdateInput struct {
	ID      string `json:"id" jsonschema:"Rule ID to update"`
	Title   string `json:"title,omitempty" jsonschema:"New title (omit to leave unchanged)"`
	Text    string `json:"text,omitempty" jsonschema:"New text (omit to leave unchanged)"`
	Project string `json:"project,omitempty" jsonschema:"\"current\" or \"global\" to move the rule; omit to leave unchanged"`
}

type RuleUpdateOutput struct {
	Message   string `json:"message"`
	Mandatory string `json:"mandatory"`
}

func (t *ruleUpdate) Register(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rule_update",
		Description: "Update an existing rule by ID. Only provided fields are changed. Fails if the id does not exist.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RuleUpdateInput) (*mcp.CallToolResult, RuleUpdateOutput, error) {
		start := logToolStart("rule_update",
			"id", input.ID,
			"title_set", input.Title != "",
			"text_set", input.Text != "",
			"project_set", input.Project != "",
		)

		existing, err := deps.Rules.Get(input.ID)
		if err != nil {
			logToolError("rule_update", start, err, "id", input.ID)
			return nil, RuleUpdateOutput{}, fmt.Errorf("rule_update failed: %w", err)
		}

		title := existing.Title
		if input.Title != "" {
			title = input.Title
		}
		text := existing.Text
		if input.Text != "" {
			text = input.Text
		}
		project := existing.Project
		if input.Project != "" {
			p, err := resolveProject(input.Project, deps.Config.DefaultProject)
			if err != nil {
				return nil, RuleUpdateOutput{}, err
			}
			project = p
		}

		if err := deps.Rules.Upsert(input.ID, title, text, project, time.Now().UTC()); err != nil {
			logToolError("rule_update", start, err, "id", input.ID)
			return nil, RuleUpdateOutput{}, fmt.Errorf("rule_update failed: %w", err)
		}
		logToolSuccess("rule_update", start, "id", input.ID)
		return nil, RuleUpdateOutput{
			Message: fmt.Sprintf("Updated rule %s", input.ID),
			Mandatory: fmt.Sprintf(
				"==MANDATORY==\nThis rule has been updated in the current session. Stop following the previous version of this rule and follow this new version instead:\n- %s (id: %s)\n%s\n==END==",
				title,
				input.ID,
				text,
			),
		}, nil
	})
}
