package tools

import (
	"fmt"

	"github.com/puntopost/acho-mcp/internal/cli"
	"github.com/puntopost/acho-mcp/internal/persistence/store"
)

func formatRegistryFull(r *store.Registry) string {
	project := r.Project
	if project == "" {
		project = "(global)"
	}
	return fmt.Sprintf("%s (%s) — %s\nproject: %s | date: %s\n\n%s",
		r.ID, r.Type, r.Title,
		project,
		r.Date.Format(cli.DateYMD_HM),
		r.Content)
}
