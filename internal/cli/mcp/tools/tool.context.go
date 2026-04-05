package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/puntopost/acho-mcp/internal/persistence/rtype"
	"github.com/puntopost/acho-mcp/internal/persistence/rule"
)

var _ Tool = (*contextLoad)(nil)

func init() {
	RegisterTool(&contextLoad{})
}

type contextLoad struct{}

func (t *contextLoad) Instruction() string {
	return "context — load mandatory rules for the current session"
}

type ContextInput struct{}

type ContextOutput struct {
	Context string `json:"context"`
}

func (t *contextLoad) Register(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "context",
		Description: "Get mandatory rules for the session. Call at the start of every session. Takes no arguments.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ ContextInput) (*mcp.CallToolResult, ContextOutput, error) {
		project := deps.Config.DefaultProject
		start := logToolStart("context", "project", project)

		rules, err := deps.Rules.List(project, false, rule.ListQuery{})
		if err != nil {
			logToolError("context", start, err, "project", project)
			return nil, ContextOutput{}, fmt.Errorf("context failed: %w", err)
		}

		types, err := deps.Types.List(project, false, rtype.ListQuery{})
		if err != nil {
			logToolError("context", start, err, "project", project)
			return nil, ContextOutput{}, fmt.Errorf("context failed: %w", err)
		}

		logToolSuccess("context", start, "project", project, "rules", len(rules), "types", len(types))
		return nil, ContextOutput{Context: renderMandatoryBlock(rules, types)}, nil
	})
}

func renderMandatoryBlock(rules []rule.Rule, types []rtype.RType) string {
	var b strings.Builder
	b.WriteString("==MANDATORY==\n\n")
	b.WriteString("Rules:\n")

	if len(rules) == 0 {
		b.WriteString("(no rules yet)\n\n")
	} else {
		for _, r := range rules {
			label := "global"
			if !r.IsGlobal() {
				label = fmt.Sprintf("project:%s", r.Project)
			}
			fmt.Fprintf(&b, "[%s] %s (id: %s)\n%s\n\n", label, r.Title, r.ID, r.Text)
		}
	}

	b.WriteString("Types:\n")
	if len(types) == 0 {
		b.WriteString("No registry types defined. Ask the user to define at least one with type_create before saving registries.\n\n")
	} else {
		for _, rt := range types {
			label := "global"
			if !rt.IsGlobal() {
				label = fmt.Sprintf("project:%s", rt.Project)
			}
			fmt.Fprintf(&b, "[%s] %s\n%s\n\n", label, rt.Name, rt.Schema)
		}
	}

	b.WriteString("If any rules above are contradictory, ask the user how to resolve the inconsistency before proceeding.\n\n")
	b.WriteString("==END==\n")
	return b.String()
}
