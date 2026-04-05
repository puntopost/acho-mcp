package service

import (
	"fmt"
	"time"

	"github.com/puntopost/acho-mcp/internal/persistence/rule"
)

type RuleService struct {
	repo rule.Repository
}

func NewRuleService(r rule.Repository) *RuleService {
	return &RuleService{repo: r}
}

func (s *RuleService) Upsert(id, title, text, project string, at time.Time) error {
	r := rule.Rule{
		ID:      id,
		Title:   title,
		Text:    text,
		Project: project,
		Date:    at,
	}
	if err := r.Validate(); err != nil {
		return fmt.Errorf("upsert: %w", err)
	}
	if err := s.repo.Upsert(r); err != nil {
		return fmt.Errorf("upsert: %w", err)
	}
	return nil
}

func (s *RuleService) Delete(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (s *RuleService) Get(id string) (*rule.Rule, error) {
	r, err := s.repo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}
	return r, nil
}

func (s *RuleService) GetAny(id string) (*rule.Rule, error) {
	r, err := s.repo.GetAny(id)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}
	return r, nil
}

func (s *RuleService) PurgeDeleted() (int, error) {
	n, err := s.repo.PurgeDeleted()
	if err != nil {
		return 0, fmt.Errorf("purge: %w", err)
	}
	return n, nil
}

// List returns rules. If global is true, only rules with project=” are
// returned; otherwise rules visible in the given project (own + global).
// When project is empty and global is false, returns everything (used by export).
func (s *RuleService) List(project string, global bool, q rule.ListQuery) ([]rule.Rule, error) {
	q.Project = project
	q.Global = global
	rules, err := s.repo.List(q)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}
	return rules, nil
}

func (s *RuleService) RenameProject(oldProject, newProject string) (int, error) {
	n, err := s.repo.RenameProject(oldProject, newProject)
	if err != nil {
		return 0, fmt.Errorf("rename project: %w", err)
	}
	return n, nil
}
