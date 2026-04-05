package rule

import (
	"fmt"
	"time"

	"github.com/puntopost/acho-mcp/internal/persistence"
)

const MaxTextLength = 1000

const JuanRuleID = "00000000000000000000000000"

type Rule struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Text        string     `json:"text"`
	Project     string     `json:"project"` // empty string means global
	Date        time.Time  `json:"date"`
	Deleted     bool       `json:"deleted"`
	DeletedDate *time.Time `json:"deleted_date,omitempty"`
}

func (r *Rule) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("id is required: %w", persistence.ErrValidation)
	}
	if r.Title == "" {
		return fmt.Errorf("title is required: %w", persistence.ErrValidation)
	}
	if r.Text == "" {
		return fmt.Errorf("text is required: %w", persistence.ErrValidation)
	}
	if len(r.Text) > MaxTextLength {
		return fmt.Errorf("text too long: %d chars (max %d): %w", len(r.Text), MaxTextLength, persistence.ErrValidation)
	}
	return nil
}

// IsGlobal returns true if this rule applies to every project.
func (r *Rule) IsGlobal() bool { return r.Project == "" }

type ListQuery struct {
	Project        string // empty + Global=false means "no filter" (used by export)
	Global         bool   // when true, return only rules with project=""
	IncludeDeleted bool   // include both active and soft-deleted
	OnlyDeleted    bool   // return only soft-deleted (overrides IncludeDeleted)
}
