package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ Tool = (*ruleDelete)(nil)

func init() {
	RegisterTool(&ruleDelete{})
}

type ruleDelete struct{}

func (t *ruleDelete) Instruction() string {
	return "rule_delete — delete a rule by ID"
}

type RuleDeleteInput struct {
	ID string `json:"id" jsonschema:"Rule ID to delete"`
}

type RuleDeleteOutput struct {
	Message string `json:"message"`
}

func (t *ruleDelete) Register(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rule_delete",
		Description: "Delete a rule by ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RuleDeleteInput) (*mcp.CallToolResult, RuleDeleteOutput, error) {
		start := logToolStart("rule_delete", "id", input.ID)
		if err := deps.Rules.Delete(input.ID); err != nil {
			logToolError("rule_delete", start, err, "id", input.ID)
			return nil, RuleDeleteOutput{}, fmt.Errorf("rule_delete failed: %w", err)
		}
		logToolSuccess("rule_delete", start, "id", input.ID)
		return nil, RuleDeleteOutput{Message: fmt.Sprintf("Deleted rule %s", input.ID)}, nil
	})
}
