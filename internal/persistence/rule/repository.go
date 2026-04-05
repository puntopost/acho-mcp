package rule

type Repository interface {
	Upsert(rule Rule) error
	Delete(id string) error
	Get(id string) (*Rule, error)
	GetAny(id string) (*Rule, error)
	List(q ListQuery) ([]Rule, error)
	RenameProject(oldProject, newProject string) (int, error)
	PurgeDeleted() (int, error)
}
