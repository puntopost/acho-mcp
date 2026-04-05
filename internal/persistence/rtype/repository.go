package rtype

type Repository interface {
	Create(rt RType) error
	Delete(name string) error
	DeleteCascade(name string) (int, error)
	Get(name string) (*RType, error)
	GetAny(name string) (*RType, error)
	Resolve(name, project string) (*RType, error)
	List(q ListQuery) ([]RType, error)
	Count() (int, error)
	CountRegistriesFor(name string) (int, error)
	RenameProject(oldProject, newProject string) (int, error)
	PurgeDeleted() (int, error)
}
