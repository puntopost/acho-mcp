package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/puntopost/acho-mcp/internal/persistence/store"
)

type StoreService struct {
	repo  store.Repository
	types *RTypeService
}

func NewStoreService(r store.Repository, types *RTypeService) *StoreService {
	return &StoreService{repo: r, types: types}
}

func (s *StoreService) Save(title, content, typ, project string) (string, error) {
	id := s.repo.NextID()
	if err := s.SaveAs(id, title, content, typ, project, time.Now().UTC()); err != nil {
		return "", err
	}
	return id, nil
}

func (s *StoreService) SaveAs(id, title, content, typ, project string, at time.Time) error {
	if typ == "" {
		return fmt.Errorf("save: type is required")
	}
	if err := s.types.ValidateContent(typ, project, content); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	flat, err := contentFlat(content)
	if err != nil {
		return fmt.Errorf("save: %w", err)
	}

	r, err := store.NewRegistry(id, title, content, typ, project, flat, at)
	if err != nil {
		return fmt.Errorf("save: %w", err)
	}
	if err := s.repo.Save(r); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	return nil
}

func (s *StoreService) Update(id string, title, content, typ, project *string) error {
	r, err := s.repo.Get(id)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	if title != nil {
		r.SetTitle(*title)
	}
	if typ != nil {
		r.SetType(*typ)
	}
	if project != nil {
		r.SetProject(*project)
	}
	if content != nil {
		if err := s.types.ValidateContent(r.Type, r.Project, *content); err != nil {
			return fmt.Errorf("update: %w", err)
		}
		flat, err := contentFlat(*content)
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		r.SetContent(*content, flat)
	} else if typ != nil || project != nil {
		if err := s.types.ValidateContent(r.Type, r.Project, r.Content); err != nil {
			return fmt.Errorf("update: %w", err)
		}
	}

	if err := r.Validate(); err != nil {
		return fmt.Errorf("update: %w", err)
	}

	r.IncrementUpdateHits()
	r.SetDate(time.Now().UTC())

	if err := s.repo.Save(*r); err != nil {
		return fmt.Errorf("update: %w", err)
	}
	return nil
}

func (s *StoreService) Delete(id string) error {
	r, err := s.repo.Get(id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	if err := s.repo.Delete(*r); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (s *StoreService) Get(id string) (*store.Registry, error) {
	r, err := s.repo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}
	r.IncrementGetHits()
	s.repo.Save(*r)
	return r, nil
}

func (s *StoreService) GetAny(id string) (*store.Registry, error) {
	r, err := s.repo.GetAny(id)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}
	return r, nil
}

func (s *StoreService) PurgeDeleted() (int, error) {
	n, err := s.repo.PurgeDeleted()
	if err != nil {
		return 0, fmt.Errorf("purge: %w", err)
	}
	return n, nil
}

func (s *StoreService) List(q store.ListQuery) ([]store.RegistryItem, error) {
	items, err := s.repo.List(q)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}
	return items, nil
}

func (s *StoreService) ExportAll() ([]store.Registry, error) {
	regs, err := s.repo.ExportAll()
	if err != nil {
		return nil, fmt.Errorf("export: %w", err)
	}
	return regs, nil
}

func (s *StoreService) Stats() (*store.Stats, error) {
	st, err := s.repo.Stats()
	if err != nil {
		return nil, fmt.Errorf("stats: %w", err)
	}
	return st, nil
}

func (s *StoreService) RenameProject(oldProject, newProject string) (int, error) {
	n, err := s.repo.RenameProject(oldProject, newProject)
	if err != nil {
		return 0, fmt.Errorf("rename project: %w", err)
	}
	return n, nil
}

func contentFlat(content string) (string, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(content), &v); err != nil {
		return "", fmt.Errorf("content is not valid JSON: %v", err)
	}
	var parts []string
	collectValues(v, &parts)
	sort.Slice(parts, func(i, j int) bool { return parts[i] < parts[j] })
	return strings.Join(parts, " "), nil
}

func collectValues(v interface{}, out *[]string) {
	switch x := v.(type) {
	case map[string]interface{}:
		for _, val := range x {
			collectValues(val, out)
		}
	case []interface{}:
		for _, val := range x {
			collectValues(val, out)
		}
	case string:
		if x != "" {
			*out = append(*out, x)
		}
	case float64:
		*out = append(*out, fmt.Sprintf("%v", x))
	case bool:
		*out = append(*out, fmt.Sprintf("%v", x))
	case nil:
	}
}
