package rule

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
		return nil, fmt.Errorf("failed to migrate rules: %w", err)
	}
	return s, nil
}

func (s *SQLiteRepository) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS rules (
			id           TEXT PRIMARY KEY,
			title        TEXT NOT NULL,
			text         TEXT NOT NULL,
			project      TEXT NOT NULL,
			date         INTEGER NOT NULL,
			deleted      INTEGER NOT NULL DEFAULT 0,
			deleted_date INTEGER
		);

		CREATE INDEX IF NOT EXISTS idx_rule_project ON rules(project);
		CREATE INDEX IF NOT EXISTS idx_rule_deleted ON rules(deleted);
	`)
	return err
}

func (s *SQLiteRepository) Upsert(r Rule) error {
	_, err := s.db.Exec(
		`INSERT INTO rules (id, title, text, project, date)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			text = excluded.text,
			project = excluded.project,
			date = excluded.date
		 WHERE deleted = 0`,
		r.ID, r.Title, r.Text, r.Project, r.Date.Unix(),
	)
	if err != nil {
		return fmt.Errorf("upsert rule %s: %w", r.ID, err)
	}
	return nil
}

func (s *SQLiteRepository) Delete(id string) error {
	result, err := s.db.Exec(
		`UPDATE rules SET deleted = 1, deleted_date = ? WHERE id = ? AND deleted = 0`,
		time.Now().Unix(), id,
	)
	if err != nil {
		return fmt.Errorf("delete rule %s: %w", id, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("rule %s: %w", id, persistence.ErrNotFound)
	}
	return nil
}

func (s *SQLiteRepository) Restore(id string) error {
	result, err := s.db.Exec(
		`UPDATE rules SET deleted = 0, deleted_date = NULL WHERE id = ? AND deleted = 1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("restore rule %s: %w", id, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return restoreRuleStateError(s.db, id)
	}
	return nil
}

func (s *SQLiteRepository) Get(id string) (*Rule, error) {
	return s.getOne(id, true)
}

func (s *SQLiteRepository) GetAny(id string) (*Rule, error) {
	return s.getOne(id, false)
}

func (s *SQLiteRepository) getOne(id string, activeOnly bool) (*Rule, error) {
	q := `SELECT id, title, text, project, date, deleted, deleted_date FROM rules WHERE id = ?`
	if activeOnly {
		q += ` AND deleted = 0`
	}
	row := s.db.QueryRow(q, id)
	r, err := scanRule(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rule %s: %w", id, persistence.ErrNotFound)
		}
		return nil, fmt.Errorf("get rule %s: %w", id, err)
	}
	return r, nil
}

func (s *SQLiteRepository) List(q ListQuery) ([]Rule, error) {
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
		`SELECT id, title, text, project, date, deleted, deleted_date
		 FROM rules
		 WHERE %s
		 ORDER BY id = ? DESC, project = '' DESC, date DESC`,
		strings.Join(where, " AND "),
	)
	args = append(args, JuanRuleID)

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Rule
	for rows.Next() {
		r, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *r)
	}
	return out, nil
}

func (s *SQLiteRepository) Stats() (*Stats, error) {
	st := &Stats{ByProject: make(map[string]persistence.Counts)}
	rows, err := s.db.Query(`SELECT project, deleted FROM rules`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var project string
		var deleted int
		if err := rows.Scan(&project, &deleted); err != nil {
			return nil, err
		}
		key := project
		if key == "" {
			key = "(global)"
		}
		c := st.ByProject[key]
		if deleted != 0 {
			c.Deleted++
			st.TotalDeleted++
		} else {
			c.Active++
			st.TotalActive++
		}
		st.ByProject[key] = c
	}
	return st, nil
}

func (s *SQLiteRepository) RenameProject(oldProject, newProject string) (int, error) {
	result, err := s.db.Exec(`UPDATE rules SET project = ? WHERE project = ? AND deleted = 0`, newProject, oldProject)
	if err != nil {
		return 0, fmt.Errorf("rename project %s -> %s: %w", oldProject, newProject, err)
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanRule(sc scanner) (*Rule, error) {
	var r Rule
	var date int64
	var deleted int
	var delDate sql.NullInt64
	if err := sc.Scan(&r.ID, &r.Title, &r.Text, &r.Project, &date, &deleted, &delDate); err != nil {
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

func (s *SQLiteRepository) PurgeDeleted() (int, error) {
	result, err := s.db.Exec(`DELETE FROM rules WHERE deleted = 1`)
	if err != nil {
		return 0, fmt.Errorf("purge rules: %w", err)
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

func restoreRuleStateError(db *sql.DB, id string) error {
	var deleted int
	err := db.QueryRow(`SELECT deleted FROM rules WHERE id = ?`, id).Scan(&deleted)
	if err == sql.ErrNoRows {
		return fmt.Errorf("rule %s: %w", id, persistence.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("restore rule %s: %w", id, err)
	}
	if deleted == 0 {
		return fmt.Errorf("rule %s is already active: %w", id, persistence.ErrValidation)
	}
	return fmt.Errorf("rule %s: restore failed", id)
}
