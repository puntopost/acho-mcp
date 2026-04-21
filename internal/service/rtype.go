package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/puntopost/acho-mcp/internal/persistence"
	"github.com/puntopost/acho-mcp/internal/persistence/rtype"
)

type RTypeService struct {
	repo          rtype.Repository
	compileSchema func(schema string) (*jsonschema.Schema, error)
	cacheMu       sync.RWMutex
	cache         map[string]*jsonschema.Schema
}

func NewRTypeService(r rtype.Repository) *RTypeService {
	return &RTypeService{repo: r, compileSchema: compileSchema}
}

func (s *RTypeService) Create(name, description, schema, project string, at time.Time) error {
	rt := rtype.RType{
		Name:        name,
		Description: description,
		Schema:      schema,
		Project:     project,
		Date:        at,
	}
	if err := rt.Validate(); err != nil {
		return fmt.Errorf("create type: %w", err)
	}
	if _, err := compileSchema(schema); err != nil {
		return fmt.Errorf("create type: invalid JSON Schema: %w: %v", persistence.ErrValidation, err)
	}
	if err := s.repo.Create(rt); err != nil {
		return fmt.Errorf("create type: %w", err)
	}
	return nil
}

func (s *RTypeService) Delete(name string, force bool) (int, error) {
	rt, err := s.repo.Get(name)
	if err != nil {
		return 0, fmt.Errorf("delete type: %w", err)
	}

	n, err := s.repo.CountRegistriesFor(name)
	if err != nil {
		return 0, fmt.Errorf("delete type: %w", err)
	}
	if n > 0 && !force {
		return 0, fmt.Errorf("delete type: %d registries use %q, pass force=true to cascade: %w", n, name, persistence.ErrValidation)
	}
	if n > 0 {
		removed, err := s.repo.DeleteCascade(name)
		if err != nil {
			return 0, fmt.Errorf("delete type: %w", err)
		}
		s.invalidateCompiledSchema(rt)
		return removed, nil
	}
	if err := s.repo.Delete(name); err != nil {
		return 0, fmt.Errorf("delete type: %w", err)
	}
	s.invalidateCompiledSchema(rt)
	return 0, nil
}

func (s *RTypeService) Restore(name string, force bool) (int, error) {
	rt, err := s.repo.GetAny(name)
	if err != nil {
		return 0, fmt.Errorf("restore type: %w", err)
	}
	if !rt.Deleted {
		return 0, fmt.Errorf("restore type: type %s is already active: %w", name, persistence.ErrValidation)
	}

	n, err := s.repo.CountDeletedRegistriesFor(name)
	if err != nil {
		return 0, fmt.Errorf("restore type: %w", err)
	}
	if n > 0 && !force {
		return 0, fmt.Errorf("restore type: %d deleted registries use %q, pass force=true to cascade restore: %w", n, name, persistence.ErrValidation)
	}
	if n > 0 {
		restored, err := s.repo.RestoreCascade(name)
		if err != nil {
			return 0, fmt.Errorf("restore type: %w", err)
		}
		return restored, nil
	}
	if err := s.repo.Restore(name); err != nil {
		return 0, fmt.Errorf("restore type: %w", err)
	}
	return 0, nil
}

func (s *RTypeService) List(project string, global bool, q rtype.ListQuery) ([]rtype.RType, error) {
	q.Project = project
	q.Global = global
	out, err := s.repo.List(q)
	if err != nil {
		return nil, fmt.Errorf("list types: %w", err)
	}
	return out, nil
}

func (s *RTypeService) GetAny(name string) (*rtype.RType, error) {
	rt, err := s.repo.GetAny(name)
	if err != nil {
		return nil, fmt.Errorf("get type: %w", err)
	}
	return rt, nil
}

func (s *RTypeService) PurgeDeleted() (int, error) {
	n, err := s.repo.PurgeDeleted()
	if err != nil {
		return 0, fmt.Errorf("purge: %w", err)
	}
	return n, nil
}

func (s *RTypeService) Resolve(name, project string) (*rtype.RType, error) {
	return s.repo.Resolve(name, project)
}

func (s *RTypeService) Count() (int, error) {
	return s.repo.Count()
}

func (s *RTypeService) Stats() (*rtype.Stats, error) {
	st, err := s.repo.Stats()
	if err != nil {
		return nil, fmt.Errorf("type stats: %w", err)
	}
	return st, nil
}

func (s *RTypeService) RenameProject(oldProject, newProject string) (int, error) {
	n, err := s.repo.RenameProject(oldProject, newProject)
	if err != nil {
		return 0, fmt.Errorf("rename project: %w", err)
	}
	return n, nil
}

func (s *RTypeService) Rename(oldName, newName string) (int, error) {
	rt, err := s.repo.Get(oldName)
	if err != nil {
		return 0, fmt.Errorf("rename type: %w", err)
	}
	rtBeforeRename := *rt

	n, err := s.repo.Rename(oldName, newName)
	if err != nil {
		return 0, fmt.Errorf("rename type: %w", err)
	}
	s.invalidateCompiledSchema(&rtBeforeRename)
	return n, nil
}

func (s *RTypeService) ValidateContent(name, project, content string) error {
	rt, err := s.repo.Resolve(name, project)
	if err != nil {
		return err
	}
	return s.validateAgainst(rt, content)
}

func compileSchema(schema string) (*jsonschema.Schema, error) {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", strings.NewReader(schema)); err != nil {
		return nil, err
	}
	return compiler.Compile("schema.json")
}

func (s *RTypeService) validateAgainst(rt *rtype.RType, content string) error {
	sch, err := s.compiledSchema(rt)
	if err != nil {
		return fmt.Errorf("schema invalid: %w", err)
	}
	var v interface{}
	if err := json.Unmarshal([]byte(content), &v); err != nil {
		return fmt.Errorf("content is not valid JSON: %w: %v", persistence.ErrValidation, err)
	}
	if err := sch.Validate(v); err != nil {
		return fmt.Errorf("content does not match schema: %w: %v", persistence.ErrValidation, err)
	}
	return nil
}

func (s *RTypeService) compiledSchema(rt *rtype.RType) (*jsonschema.Schema, error) {
	key := compiledSchemaCacheKey(rt)

	s.cacheMu.RLock()
	if sch, ok := s.cache[key]; ok {
		s.cacheMu.RUnlock()
		return sch, nil
	}
	s.cacheMu.RUnlock()

	sch, err := s.compileSchema(rt.Schema)
	if err != nil {
		return nil, err
	}

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if s.cache == nil {
		s.cache = make(map[string]*jsonschema.Schema)
	}
	if existing, ok := s.cache[key]; ok {
		return existing, nil
	}
	s.cache[key] = sch
	return sch, nil
}

func (s *RTypeService) invalidateCompiledSchema(rt *rtype.RType) {
	if rt == nil {
		return
	}

	key := compiledSchemaCacheKey(rt)
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	delete(s.cache, key)
}

func compiledSchemaCacheKey(rt *rtype.RType) string {
	return rt.Name + "\x00" + rt.Project + "\x00" + rt.Schema
}
