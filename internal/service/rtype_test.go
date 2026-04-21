package service

import (
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/puntopost/acho-mcp/internal/persistence"
	"github.com/puntopost/acho-mcp/internal/persistence/rtype"
)

type rtypeRepoStub struct {
	rt                *rtype.RType
	registriesForType int
}

func (s *rtypeRepoStub) Create(rt rtype.RType) error { return nil }
func (s *rtypeRepoStub) Delete(name string) error {
	if s.rt == nil || s.rt.Name != name {
		return persistence.ErrNotFound
	}
	s.rt = nil
	return nil
}
func (s *rtypeRepoStub) DeleteCascade(name string) (int, error) {
	if s.rt == nil || s.rt.Name != name {
		return 0, persistence.ErrNotFound
	}
	removed := s.registriesForType
	s.rt = nil
	s.registriesForType = 0
	return removed, nil
}
func (s *rtypeRepoStub) Get(name string) (*rtype.RType, error) {
	if s.rt == nil || s.rt.Name != name {
		return nil, persistence.ErrNotFound
	}
	return s.rt, nil
}
func (s *rtypeRepoStub) GetAny(name string) (*rtype.RType, error) {
	if s.rt == nil || s.rt.Name != name {
		return nil, persistence.ErrNotFound
	}
	return s.rt, nil
}
func (s *rtypeRepoStub) List(q rtype.ListQuery) ([]rtype.RType, error) { return nil, nil }
func (s *rtypeRepoStub) Count() (int, error)                           { return 0, nil }
func (s *rtypeRepoStub) Stats() (*rtype.Stats, error)                  { return nil, nil }
func (s *rtypeRepoStub) CountRegistriesFor(name string) (int, error) {
	if s.rt == nil || s.rt.Name != name {
		return 0, persistence.ErrNotFound
	}
	return s.registriesForType, nil
}
func (s *rtypeRepoStub) RenameProject(oldProject, newProject string) (int, error) {
	return 0, nil
}
func (s *rtypeRepoStub) Rename(oldName, newName string) (int, error) {
	if s.rt == nil || s.rt.Name != oldName {
		return 0, persistence.ErrNotFound
	}
	s.rt.Name = newName
	return s.registriesForType, nil
}
func (s *rtypeRepoStub) PurgeDeleted() (int, error) { return 0, nil }

func (s *rtypeRepoStub) Resolve(name, project string) (*rtype.RType, error) {
	if s.rt == nil || s.rt.Name != name || s.rt.Project != project {
		return nil, persistence.ErrNotFound
	}
	return s.rt, nil
}

func TestRTypeServiceValidateContentCachesCompiledSchema(t *testing.T) {
	repo := &rtypeRepoStub{rt: &rtype.RType{
		Name:    "note",
		Project: "project-a",
		Schema:  `{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`,
	}}
	svc := NewRTypeService(repo)

	compileCalls := 0
	svc.compileSchema = func(schema string) (*jsonschema.Schema, error) {
		compileCalls++
		return compileSchema(schema)
	}

	content := `{"text":"hola"}`
	for range 2 {
		if err := svc.ValidateContent("note", "project-a", content); err != nil {
			t.Fatalf("ValidateContent() error = %v", err)
		}
	}

	if compileCalls != 1 {
		t.Fatalf("expected schema compilation once, got %d", compileCalls)
	}

	if len(svc.cache) != 1 {
		t.Fatalf("expected one cached schema, got %d", len(svc.cache))
	}
}

func TestRTypeServiceDeleteInvalidatesCompiledSchemaCache(t *testing.T) {
	repo := &rtypeRepoStub{rt: &rtype.RType{
		Name:    "note",
		Project: "project-a",
		Schema:  `{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`,
	}}
	svc := NewRTypeService(repo)

	if err := svc.ValidateContent("note", "project-a", `{"text":"hola"}`); err != nil {
		t.Fatalf("ValidateContent() error = %v", err)
	}

	if len(svc.cache) != 1 {
		t.Fatalf("expected one cached schema before delete, got %d", len(svc.cache))
	}

	if _, err := svc.Delete("note", false); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if len(svc.cache) != 0 {
		t.Fatalf("expected empty cache after delete, got %d", len(svc.cache))
	}
}

func TestRTypeServiceRenameInvalidatesCompiledSchemaCache(t *testing.T) {
	repo := &rtypeRepoStub{rt: &rtype.RType{
		Name:    "note",
		Project: "project-a",
		Schema:  `{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`,
	}}
	svc := NewRTypeService(repo)

	if err := svc.ValidateContent("note", "project-a", `{"text":"hola"}`); err != nil {
		t.Fatalf("ValidateContent() error = %v", err)
	}

	oldKey := compiledSchemaCacheKey(repo.rt)
	if _, ok := svc.cache[oldKey]; !ok {
		t.Fatalf("expected cache entry for original type")
	}

	if _, err := svc.Rename("note", "memo"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}

	if len(svc.cache) != 0 {
		t.Fatalf("expected empty cache after rename, got %d", len(svc.cache))
	}
}
