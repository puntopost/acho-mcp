package rtype

import "github.com/puntopost/acho-mcp/internal/persistence"

type Stats struct {
	ByProject    map[string]persistence.Counts
	TotalActive  int
	TotalDeleted int
}

type Repository interface {
	Create(rt RType) error
	Delete(name string) error
	DeleteCascade(name string) (int, error)
	Get(name string) (*RType, error)
	GetAny(name string) (*RType, error)
	Resolve(name, project string) (*RType, error)
	List(q ListQuery) ([]RType, error)
	Count() (int, error)
	Stats() (*Stats, error)
	CountRegistriesFor(name string) (int, error)
	RenameProject(oldProject, newProject string) (int, error)
	Rename(oldName, newName string) (int, error)
	PurgeDeleted() (int, error)
}
