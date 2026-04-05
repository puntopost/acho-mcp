package persistence

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// OpenDB opens a SQLite database connection with the given path and options.
func OpenDB(dbPath, dbOptions string) (*sql.DB, error) {
	dsn := dbPath
	if dbOptions != "" {
		dsn += "?" + dbOptions
	}
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}
