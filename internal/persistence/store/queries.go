package store

type ListQuery struct {
	Project        string
	Global         bool
	Limit          int
	Offset         int
	IncludeDeleted bool
	OnlyDeleted    bool
}
