package store

import (
	"fmt"
	"time"

	"github.com/puntopost/acho-mcp/internal/persistence"
)

type Registry struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	ContentFlat string     `json:"-"`
	Project     string     `json:"project"` // empty string means global
	SearchHits  int        `json:"search_hits"`
	GetHits     int        `json:"get_hits"`
	UpdateHits  int        `json:"update_hits"`
	Date        time.Time  `json:"date"`
	Deleted     bool       `json:"deleted"`
	DeletedDate *time.Time `json:"deleted_date,omitempty"`
}

type RegistryItem struct {
	ID            string
	Type          string
	Title         string
	Content       string
	ContentLength int
	Project       string
	SearchHits    int
	GetHits       int
	UpdateHits    int
	Date          time.Time
	Deleted       bool
	DeletedDate   *time.Time
}

func NewRegistry(id, title, content, typ, project, contentFlat string, now time.Time) (Registry, error) {
	r := Registry{
		ID:          id,
		Type:        typ,
		Title:       title,
		Content:     content,
		ContentFlat: contentFlat,
		Project:     project,
		Date:        now,
	}
	if err := r.Validate(); err != nil {
		return Registry{}, err
	}
	return r, nil
}

func (r *Registry) Validate() error {
	if r.Title == "" {
		return fmt.Errorf("title is required: %w", persistence.ErrValidation)
	}
	if r.Content == "" {
		return fmt.Errorf("content is required: %w", persistence.ErrValidation)
	}
	if r.Type == "" {
		return fmt.Errorf("type is required: %w", persistence.ErrValidation)
	}
	return nil
}

func (r *Registry) IsGlobal() bool { return r.Project == "" }

func (r *Registry) SetTitle(title string)     { r.Title = title }
func (r *Registry) SetContent(c, flat string) { r.Content = c; r.ContentFlat = flat }
func (r *Registry) SetProject(project string) { r.Project = project }
func (r *Registry) SetType(typ string)        { r.Type = typ }
func (r *Registry) SetDate(t time.Time)       { r.Date = t }

func (r *Registry) IncrementSearchHits() { r.SearchHits++ }
func (r *Registry) IncrementGetHits()    { r.GetHits++ }
func (r *Registry) IncrementUpdateHits() { r.UpdateHits++ }
