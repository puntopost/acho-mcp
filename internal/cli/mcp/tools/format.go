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
	return fmt.Sprintf("%s (%s) — %s\nproject: %s | date: %s\nsearches: %d | gets: %d | updates: %d\n\n%s",
		r.ID, r.Type, r.Title,
		project,
		r.Date.Format(cli.DateYMD_HM),
		r.SearchHits, r.GetHits, r.UpdateHits,
		r.Content)
}
