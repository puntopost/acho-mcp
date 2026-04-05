package rtype

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/puntopost/acho-mcp/internal/persistence"
)

var _ Repository = (*SQLiteRepository)(nil)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) (*SQLiteRepository, error) {
	s := &SQLiteRepository{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate registry_types: %w", err)
	}
	return s, nil
}

func (s *SQLiteRepository) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS registry_types (
			name         TEXT PRIMARY KEY,
			schema       TEXT NOT NULL,
			project      TEXT NOT NULL,
			date         INTEGER NOT NULL,
			deleted      INTEGER NOT NULL DEFAULT 0,
			deleted_date INTEGER,
			CHECK (json_valid(schema))
		);

		CREATE INDEX IF NOT EXISTS idx_rtype_deleted ON registry_types(deleted);
	`)
	return err
}

func (s *SQLiteRepository) Create(rt RType) error {
	var deleted int
	err := s.db.QueryRow(`SELECT deleted FROM registry_types WHERE name = ?`, rt.Name).Scan(&deleted)
	switch {
	case err == sql.ErrNoRows:
		if _, err := s.db.Exec(
			`INSERT INTO registry_types (name, schema, project, date) VALUES (?, ?, ?, ?)`,
			rt.Name, rt.Schema, rt.Project, rt.Date.Unix(),
		); err != nil {
			return fmt.Errorf("create type %s: %w", rt.Name, err)
		}
		return nil
	case err != nil:
		return fmt.Errorf("create type %s: %w", rt.Name, err)
	case deleted == 0:
		return fmt.Errorf("type %s already exists: %w", rt.Name, persistence.ErrValidation)
	default:
		// deleted == 1: resurrect with the new definition.
		if _, err := s.db.Exec(
			`UPDATE registry_types
			 SET schema = ?, project = ?, date = ?, deleted = 0, deleted_date = NULL
			 WHERE name = ?`,
			rt.Schema, rt.Project, rt.Date.Unix(), rt.Name,
		); err != nil {
			return fmt.Errorf("create type %s: %w", rt.Name, err)
		}
		return nil
	}
}

func (s *SQLiteRepository) Delete(name string) error {
	result, err := s.db.Exec(
		`UPDATE registry_types SET deleted = 1, deleted_date = ? WHERE name = ? AND deleted = 0`,
		time.Now().Unix(), name,
	)
	if err != nil {
		return fmt.Errorf("delete type %s: %w", name, err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("type %s: %w", name, persistence.ErrNotFound)
	}
	return nil
}

func (s *SQLiteRepository) DeleteCascade(name string) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := time.Now().Unix()

	res, err := tx.Exec(
		`UPDATE registries SET deleted = 1, deleted_date = ? WHERE type = ? AND deleted = 0`,
		now, name,
	)
	if err != nil {
		return 0, fmt.Errorf("cascade delete registries: %w", err)
	}
	removed, _ := res.RowsAffected()

	res2, err := tx.Exec(
		`UPDATE registry_types SET deleted = 1, deleted_date = ? WHERE name = ? AND deleted = 0`,
		now, name,
	)
	if err != nil {
		return 0, fmt.Errorf("delete type %s: %w", name, err)
	}
	n, _ := res2.RowsAffected()
	if n == 0 {
		return 0, fmt.Errorf("type %s: %w", name, persistence.ErrNotFound)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int(removed), nil
}

func (s *SQLiteRepository) Get(name string) (*RType, error) {
	return s.getOne(name, true)
}

func (s *SQLiteRepository) GetAny(name string) (*RType, error) {
	return s.getOne(name, false)
}

func (s *SQLiteRepository) getOne(name string, activeOnly bool) (*RType, error) {
	q := `SELECT name, schema, project, date, deleted, deleted_date FROM registry_types WHERE name = ?`
	if activeOnly {
		q += ` AND deleted = 0`
	}
	row := s.db.QueryRow(q, name)
	rt, err := scan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("type %s: %w", name, persistence.ErrNotFound)
		}
		return nil, fmt.Errorf("get type %s: %w", name, err)
	}
	return rt, nil
}

// Resolve returns the type applicable for the given project. A type is visible
// if it is global (project=”) or if its project matches.
func (s *SQLiteRepository) Resolve(name, project string) (*RType, error) {
	rt, err := s.Get(name)
	if err != nil {
		return nil, err
	}
	if rt.Project == "" || rt.Project == project {
		return rt, nil
	}
	return nil, fmt.Errorf("type %s not available in project %s (defined for project %s): %w", name, project, rt.Project, persistence.ErrValidation)
}

func (s *SQLiteRepository) List(q ListQuery) ([]RType, error) {
	var where []string
	var args []interface{}

	switch {
	case q.OnlyDeleted:
		where = append(where, "deleted = 1")
	case !q.IncludeDeleted:
		where = append(where, "deleted = 0")
	}

	switch {
	case q.Global:
		where = append(where, "project = ''")
	case q.Project != "":
		where = append(where, "(project = ? OR project = '')")
		args = append(args, q.Project)
	}

	if len(where) == 0 {
		where = append(where, "1=1")
	}

	sql := fmt.Sprintf(
		`SELECT name, schema, project, date, deleted, deleted_date
		 FROM registry_types
		 WHERE %s
		 ORDER BY project = '' DESC, name ASC`,
		strings.Join(where, " AND "),
	)

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RType
	for rows.Next() {
		rt, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *rt)
	}
	return out, nil
}

func (s *SQLiteRepository) Count() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM registry_types WHERE deleted = 0`).Scan(&n)
	return n, err
}

func (s *SQLiteRepository) CountRegistriesFor(name string) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM registries WHERE type = ? AND deleted = 0`, name).Scan(&n)
	return n, err
}

func (s *SQLiteRepository) RenameProject(oldProject, newProject string) (int, error) {
	result, err := s.db.Exec(`UPDATE registry_types SET project = ? WHERE project = ? AND deleted = 0`, newProject, oldProject)
	if err != nil {
		return 0, fmt.Errorf("rename project %s -> %s: %w", oldProject, newProject, err)
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scan(sc scanner) (*RType, error) {
	var rt RType
	var date int64
	var deleted int
	var delDate sql.NullInt64
	if err := sc.Scan(&rt.Name, &rt.Schema, &rt.Project, &date, &deleted, &delDate); err != nil {
		return nil, err
	}
	rt.Date = time.Unix(date, 0).UTC()
	rt.Deleted = deleted != 0
	if delDate.Valid {
		t := time.Unix(delDate.Int64, 0).UTC()
		rt.DeletedDate = &t
	}
	return &rt, nil
}

func (s *SQLiteRepository) PurgeDeleted() (int, error) {
	result, err := s.db.Exec(`DELETE FROM registry_types WHERE deleted = 1`)
	if err != nil {
		return 0, fmt.Errorf("purge types: %w", err)
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}
