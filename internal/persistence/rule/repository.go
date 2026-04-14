package rule

import "github.com/puntopost/acho-mcp/internal/persistence"

type Repository interface {
	Upsert(rule Rule) error
	Delete(id string) error
	Get(id string) (*Rule, error)
	GetAny(id string) (*Rule, error)
	List(q ListQuery) ([]Rule, error)
	Stats() (*Stats, error)
	RenameProject(oldProject, newProject string) (int, error)
	PurgeDeleted() (int, error)
}

type Stats struct {
	ByProject    map[string]persistence.Counts
	TotalActive  int
	TotalDeleted int
}
