package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/puntopost/acho-mcp/internal/persistence"
)

var _ Repository = (*SQLiteRepository)(nil)

type SQLiteRepository struct {
	db            *sql.DB
	snippetLength int
}

func NewSQLiteRepository(db *sql.DB, snippetLength int) (*SQLiteRepository, error) {
	s := &SQLiteRepository{db: db, snippetLength: snippetLength}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	return s, nil
}

func (s *SQLiteRepository) NextID() string {
	return ulid.Make().String()
}

func (s *SQLiteRepository) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS registries (
			id           TEXT PRIMARY KEY,
			type         TEXT    NOT NULL,
			title        TEXT    NOT NULL,
			content      TEXT    NOT NULL,
			content_flat TEXT    NOT NULL,
			project      TEXT    NOT NULL,
			date         INTEGER NOT NULL,
			deleted      INTEGER NOT NULL DEFAULT 0,
			deleted_date INTEGER,
			CHECK (json_valid(content))
		);

		CREATE INDEX IF NOT EXISTS idx_reg_project ON registries(project);
		CREATE INDEX IF NOT EXISTS idx_reg_type    ON registries(type);
		CREATE INDEX IF NOT EXISTS idx_reg_date    ON registries(date DESC);
		CREATE INDEX IF NOT EXISTS idx_reg_deleted ON registries(deleted);

		CREATE VIRTUAL TABLE IF NOT EXISTS registry_fts USING fts5(
			title,
			content_flat,
			type,
			project,
			content='registries',
			content_rowid='rowid'
		);

		CREATE TRIGGER IF NOT EXISTS reg_fts_insert AFTER INSERT ON registries BEGIN
			INSERT INTO registry_fts(rowid, title, content_flat, type, project)
			SELECT new.rowid, new.title, new.content_flat, new.type, new.project
			WHERE new.deleted = 0;
		END;

		CREATE TRIGGER IF NOT EXISTS reg_fts_delete AFTER DELETE ON registries
		WHEN old.deleted = 0 BEGIN
			INSERT INTO registry_fts(registry_fts, rowid, title, content_flat, type, project)
			VALUES ('delete', old.rowid, old.title, old.content_flat, old.type, old.project);
		END;

		CREATE TRIGGER IF NOT EXISTS reg_fts_update AFTER UPDATE ON registries BEGIN
			INSERT INTO registry_fts(registry_fts, rowid, title, content_flat, type, project)
			SELECT 'delete', old.rowid, old.title, old.content_flat, old.type, old.project
			WHERE old.deleted = 0;
			INSERT INTO registry_fts(rowid, title, content_flat, type, project)
			SELECT new.rowid, new.title, new.content_flat, new.type, new.project
			WHERE new.deleted = 0;
		END;
	`)
	return err
}

func (s *SQLiteRepository) Save(r Registry) error {
	_, err := s.db.Exec(
		`INSERT INTO registries (id, type, title, content, content_flat, project, date)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
			type = excluded.type,
			title = excluded.title,
			content = excluded.content,
			content_flat = excluded.content_flat,
			project = excluded.project,
			date = excluded.date
		 WHERE deleted = 0`,
		r.ID, r.Type, r.Title, r.Content, r.ContentFlat, r.Project,
		r.Date.Unix(),
	)
	if err != nil {
		return fmt.Errorf("save registry %s: %w", r.ID, err)
	}
	return nil
}

func (s *SQLiteRepository) Get(id string) (*Registry, error) {
	return s.getOne(id, true)
}

func (s *SQLiteRepository) GetAny(id string) (*Registry, error) {
	return s.getOne(id, false)
}

func (s *SQLiteRepository) getOne(id string, activeOnly bool) (*Registry, error) {
	q := `SELECT id, type, title, content, content_flat, project, date, deleted, deleted_date
		  FROM registries WHERE id = ?`
	if activeOnly {
		q += ` AND deleted = 0`
	}
	row := s.db.QueryRow(q, id)
	r, err := scanRegistry(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("registry %s: %w", id, persistence.ErrNotFound)
		}
		return nil, fmt.Errorf("get registry %s: %w", id, err)
	}
	return r, nil
}

func (s *SQLiteRepository) Delete(r Registry) error {
	result, err := s.db.Exec(
		`UPDATE registries SET deleted = 1, deleted_date = ? WHERE id = ? AND deleted = 0`,
		time.Now().Unix(), r.ID,
	)
	if err != nil {
		return fmt.Errorf("delete registry %s: %w", r.ID, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("registry %s: %w", r.ID, persistence.ErrNotFound)
	}
	return nil
}

func (s *SQLiteRepository) Restore(id string) error {
	result, err := s.db.Exec(
		`UPDATE registries SET deleted = 0, deleted_date = NULL WHERE id = ? AND deleted = 1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("restore registry %s: %w", id, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return restoreRegistryStateError(s.db, id)
	}
	return nil
}

func (s *SQLiteRepository) ExportAll() ([]Registry, error) {
	rows, err := s.db.Query(
		`SELECT id, type, title, content, content_flat, project, date, deleted, deleted_date
		 FROM registries WHERE deleted = 0 ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var registries []Registry
	for rows.Next() {
		r, err := scanRegistry(rows)
		if err != nil {
			return nil, err
		}
		registries = append(registries, *r)
	}
	return registries, nil
}

func (s *SQLiteRepository) Stats() (*Stats, error) {
	st := &Stats{
		ByProject: make(map[string]persistence.Counts),
		ByType:    make(map[string]persistence.Counts),
	}

	rows, err := s.db.Query(`SELECT project, type, deleted FROM registries`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var project, typ string
		var deleted int
		if err := rows.Scan(&project, &typ, &deleted); err != nil {
			return nil, err
		}
		key := project
		if key == "" {
			key = "(global)"
		}
		pc := st.ByProject[key]
		tc := st.ByType[typ]
		if deleted != 0 {
			pc.Deleted++
			tc.Deleted++
			st.TotalDeleted++
		} else {
			pc.Active++
			tc.Active++
			st.TotalActive++
		}
		st.ByProject[key] = pc
		st.ByType[typ] = tc
	}
	return st, nil
}

func (s *SQLiteRepository) List(q ListQuery) ([]RegistryItem, error) {
	if q.Limit <= 0 {
		q.Limit = persistence.DefaultLimit
	}

	contentCol := fmt.Sprintf("substr(r.content, 1, %d)", s.snippetLength)

	var where []string
	var args []interface{}

	switch {
	case q.OnlyDeleted:
		where = append(where, "r.deleted = 1")
	case !q.IncludeDeleted:
		where = append(where, "r.deleted = 0")
	}

	switch {
	case q.Global:
		where = append(where, "r.project = ''")
	case q.Project != "":
		where = append(where, "(r.project = ? OR r.project = '')")
		args = append(args, q.Project)
	}

	if len(where) == 0 {
		where = append(where, "1=1")
	}

	sql := fmt.Sprintf(
		`SELECT r.id, r.type, r.title, %s, length(r.content), r.project, r.date, r.deleted, r.deleted_date
		 FROM registries r
		 WHERE %s
		 ORDER BY r.date DESC
		 LIMIT ? OFFSET ?`,
		contentCol, strings.Join(where, " AND "),
	)
	args = append(args, q.Limit, q.Offset)

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanItems(rows)
}

func (s *SQLiteRepository) PurgeDeleted() (int, error) {
	result, err := s.db.Exec(`DELETE FROM registries WHERE deleted = 1`)
	if err != nil {
		return 0, fmt.Errorf("purge registries: %w", err)
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

func (s *SQLiteRepository) RenameProject(oldProject, newProject string) (int, error) {
	result, err := s.db.Exec(`UPDATE registries SET project = ? WHERE project = ? AND deleted = 0`, newProject, oldProject)
	if err != nil {
		return 0, fmt.Errorf("rename project %s -> %s: %w", oldProject, newProject, err)
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanRegistry(s scanner) (*Registry, error) {
	var r Registry
	var date int64
	var deleted int
	var delDate sql.NullInt64

	err := s.Scan(&r.ID, &r.Type, &r.Title, &r.Content, &r.ContentFlat, &r.Project,
		&date, &deleted, &delDate)
	if err != nil {
		return nil, err
	}
	r.Date = time.Unix(date, 0).UTC()
	r.Deleted = deleted != 0
	if delDate.Valid {
		t := time.Unix(delDate.Int64, 0).UTC()
		r.DeletedDate = &t
	}
	return &r, nil
}

func (s *SQLiteRepository) GetByIDs(ids []string) ([]RegistryItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	contentCol := fmt.Sprintf("substr(r.content, 1, %d)", s.snippetLength)

	query := fmt.Sprintf(
		`SELECT r.id, r.type, r.title, %s, length(r.content), r.project, r.date, r.deleted, r.deleted_date
		 FROM registries r
		 WHERE r.id IN (%s) AND r.deleted = 0`,
		contentCol, strings.Join(placeholders, ","),
	)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanItems(rows)
}

func scanItems(rows *sql.Rows) ([]RegistryItem, error) {
	var items []RegistryItem
	for rows.Next() {
		var item RegistryItem
		var date int64
		var deleted int
		var delDate sql.NullInt64

		err := rows.Scan(&item.ID, &item.Type, &item.Title, &item.Content, &item.ContentLength,
			&item.Project, &date, &deleted, &delDate)
		if err != nil {
			return nil, err
		}
		item.Date = time.Unix(date, 0).UTC()
		item.Deleted = deleted != 0
		if delDate.Valid {
			t := time.Unix(delDate.Int64, 0).UTC()
			item.DeletedDate = &t
		}

		if item.ContentLength > len(item.Content) &&
			!strings.HasSuffix(item.Content, persistence.TruncationIndicator) {
			item.Content += persistence.TruncationIndicator
		}
		items = append(items, item)
	}
	return items, nil
}

func restoreRegistryStateError(db *sql.DB, id string) error {
	var deleted int
	err := db.QueryRow(`SELECT deleted FROM registries WHERE id = ?`, id).Scan(&deleted)
	if err == sql.ErrNoRows {
		return fmt.Errorf("registry %s: %w", id, persistence.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("restore registry %s: %w", id, err)
	}
	if deleted == 0 {
		return fmt.Errorf("registry %s is already active: %w", id, persistence.ErrValidation)
	}
	return fmt.Errorf("registry %s: restore failed", id)
}
