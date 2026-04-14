package persistence

const (
	DefaultLimit        = 10
	TruncationIndicator = "<...>"
)

// Counts holds active/deleted tallies for a single bucket (project, type, …).
type Counts struct {
	Active  int
	Deleted int
}

func (c Counts) Total() int { return c.Active + c.Deleted }
