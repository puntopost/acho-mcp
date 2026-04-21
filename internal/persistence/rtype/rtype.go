package rtype

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/puntopost/acho-mcp/internal/persistence"
)

var nameRe = regexp.MustCompile(`^[a-z][a-z_]*$`)

const MaxDescriptionLength = 300

type RType struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
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
	r.Description = strings.TrimSpace(r.Description)
	if r.Description == "" {
		return fmt.Errorf("description is required: %w", persistence.ErrValidation)
	}
	if len(r.Description) > MaxDescriptionLength {
		return fmt.Errorf("description too long: %d chars (max %d): %w", len(r.Description), MaxDescriptionLength, persistence.ErrValidation)
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
