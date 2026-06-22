package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Open(
	path string,
) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := migrate(conn); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

func migrate(
	conn *sql.DB,
) error {
	statements := []string{
		`PRAGMA foreign_keys = ON`,
		`PRAGMA journal_mode = WAL`,
		`CREATE TABLE IF NOT EXISTS hosts (
			id TEXT PRIMARY KEY,
			alias TEXT NOT NULL UNIQUE,
			hostname TEXT NOT NULL,
			port INTEGER NOT NULL DEFAULT 22,
			username TEXT NOT NULL,
			identity_file TEXT NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			favorite INTEGER NOT NULL DEFAULT 0 CHECK (favorite IN (0, 1)),
			last_connected_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_hosts_alias ON hosts(alias)`,
		`CREATE INDEX IF NOT EXISTS idx_hosts_last_connected ON hosts(last_connected_at DESC)`,
		`CREATE TABLE IF NOT EXISTS connection_history (
			id TEXT PRIMARY KEY,
			host_id TEXT NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
			alias_snapshot TEXT NOT NULL,
			command TEXT NOT NULL,
			started_at TEXT NOT NULL,
			exit_status INTEGER,
			terminal_app TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_connection_history_host ON connection_history(host_id, started_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_connection_history_started ON connection_history(started_at DESC)`,
	}

	for _, stmt := range statements {
		if _, err := conn.Exec(stmt); err != nil {
			return fmt.Errorf("migrate sqlite: %w", err)
		}
	}
	return nil
}
