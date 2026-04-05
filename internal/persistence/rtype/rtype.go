package rtype

import (
	"fmt"
	"regexp"
	"time"

	"github.com/puntopost/acho-mcp/internal/persistence"
)

var nameRe = regexp.MustCompile(`^[a-z][a-z_]*$`)

type RType struct {
	Name        string     `json:"name"`
	Schema      string     `json:"schema"`
	Project     string     `json:"project"` // empty string means global
	Date        time.Time  `json:"date"`
	Deleted     bool       `json:"deleted"`
	DeletedDate *time.Time `json:"deleted_date,omitempty"`
}

func (r *RType) Validate() error {
	if !nameRe.MatchString(r.Name) {
		return fmt.Errorf("invalid name %q (must match ^[a-z][a-z_]*$): %w", r.Name, persistence.ErrValidation)
	}
	if r.Schema == "" {
		return fmt.Errorf("schema is required: %w", persistence.ErrValidation)
	}
	return nil
}

func (r *RType) IsGlobal() bool { return r.Project == "" }

type ListQuery struct {
	Project        string
	Global         bool
	IncludeDeleted bool
	OnlyDeleted    bool
}
