package store

import "github.com/puntopost/acho-mcp/internal/persistence"

type Repository interface {
	NextID() string
	Save(registry Registry) error
	Get(id string) (*Registry, error)
	GetAny(id string) (*Registry, error)
	GetByIDs(ids []string) ([]RegistryItem, error)
	Delete(registry Registry) error
	ExportAll() ([]Registry, error)
	Stats() (*Stats, error)
	List(q ListQuery) ([]RegistryItem, error)
	RenameProject(oldProject, newProject string) (int, error)
	PurgeDeleted() (int, error)
}

type Stats struct {
	ByProject    map[string]persistence.Counts
	ByType       map[string]persistence.Counts
	TotalActive  int
	TotalDeleted int
}
