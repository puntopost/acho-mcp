// Package app wires the Acho services from a given configuration.
package app

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/puntopost/acho-mcp/internal/cli/config"
	"github.com/puntopost/acho-mcp/internal/persistence"
	"github.com/puntopost/acho-mcp/internal/persistence/rtype"
	"github.com/puntopost/acho-mcp/internal/persistence/rule"
	"github.com/puntopost/acho-mcp/internal/persistence/store"
	"github.com/puntopost/acho-mcp/internal/service"
)

type Deps struct {
	DB       *sql.DB
	Repo     store.Repository
	RuleRepo rule.Repository
	TypeRepo rtype.Repository
	Service  *service.StoreService
	Rules    *service.RuleService
	Types    *service.RTypeService
	Project  string
}

func Build(cfg *config.Config, project string) (*Deps, error) {
	if cfg == nil {
		return nil, fmt.Errorf("app: config is required")
	}

	if cfg.DBPath != ":memory:" {
		dbDir := filepath.Dir(cfg.DBPath)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, fmt.Errorf("create database directory %s: %w", dbDir, err)
		}
	}

	db, err := persistence.OpenDB(cfg.DBPath, cfg.DBOptions)
	if err != nil {
		return nil, err
	}

	repo, err := store.NewSQLiteRepository(db, cfg.SnippetLength)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("init repository: %w", err)
	}

	ruleRepo, err := rule.NewSQLiteRepository(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("init rule repository: %w", err)
	}

	typeRepo, err := rtype.NewSQLiteRepository(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("init type repository: %w", err)
	}

	types := service.NewRTypeService(typeRepo)
	svc := service.NewStoreService(repo, types)
	rules := service.NewRuleService(ruleRepo)

	if project != "" {
		if err := createProjectViews(db, project); err != nil {
			db.Close()
			return nil, fmt.Errorf("init project views: %w", err)
		}
	}

	return &Deps{
		DB:       db,
		Repo:     repo,
		RuleRepo: ruleRepo,
		TypeRepo: typeRepo,
		Service:  svc,
		Rules:    rules,
		Types:    types,
		Project:  project,
	}, nil
}

func (d *Deps) Close() {
	d.DB.Close()
}

// createProjectViews creates temp views that pre-filter registries and types
// by the current project. The sql_query MCP tool exposes only these views and
// blocks access to the raw tables, so the agent cannot accidentally read data
// from other projects. Views are temp so they disappear with the connection.
func createProjectViews(db *sql.DB, project string) error {
	// project is only interpolated in a safe way: SQLite TEMP views cannot be
	// parameterized, so we inline it. The project name comes from trusted
	// sources (--project flag or git/cwd detection), but we still escape '
	// to be defensive.
	p := escapeSQLString(project)
	stmts := []string{
		`DROP VIEW IF EXISTS temp.v_registries`,
		`DROP VIEW IF EXISTS temp.v_types`,
		`DROP VIEW IF EXISTS temp.v_rules`,
		fmt.Sprintf(`CREATE TEMP VIEW v_registries AS
			SELECT rowid, id, type, title, content, content_flat, project,
			       search_hits, get_hits, update_hits, date
			FROM registries
			WHERE deleted = 0 AND (project = '%s' OR project = '')`, p),
		fmt.Sprintf(`CREATE TEMP VIEW v_types AS
			SELECT name, schema, project, date
			FROM registry_types
			WHERE deleted = 0 AND (project = '%s' OR project = '')`, p),
		fmt.Sprintf(`CREATE TEMP VIEW v_rules AS
			SELECT rowid, id, title, text, project, date
			FROM rules
			WHERE deleted = 0 AND (project = '%s' OR project = '')`, p),
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

func escapeSQLString(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\'' {
			out = append(out, '\'', '\'')
			continue
		}
		out = append(out, s[i])
	}
	return string(out)
}
