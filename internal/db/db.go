package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

func Open(
	path string,
) (
	*sql.DB,
	error,
) {
	if err := os.MkdirAll(
		filepath.Dir(path),
		0o700,
	); err != nil {
		return nil, fmt.Errorf(
			"create data directory: %w",
			err,
		)
	}

	conn, err := sql.Open(
		"sqlite",
		path,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"open sqlite: %w",
			err,
		)
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
	if _, err := conn.Exec(
		`PRAGMA foreign_keys = ON`,
	); err != nil {
		return fmt.Errorf(
			"enable foreign keys: %w",
			err,
		)
	}

	if _, err := conn.Exec(
		`PRAGMA journal_mode = WAL`,
	); err != nil {
		return fmt.Errorf(
			"enable WAL mode: %w",
			err,
		)
	}

	if _, err := conn.Exec(
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		)`,
	); err != nil {
		return fmt.Errorf(
			"create schema migrations: %w",
			err,
		)
	}

	currentVersion, err := schemaVersion(conn)
	if err != nil {
		return err
	}

	latestVersion := latestSchemaVersion()
	if currentVersion > latestVersion {
		return fmt.Errorf(
			"database schema version %d is newer than supported version %d",
			currentVersion,
			latestVersion,
		)
	}
	for _, item := range migrations {
		if item.version <= currentVersion {
			continue
		}
		if err := applyMigration(
			conn,
			item,
		); err != nil {
			return err
		}
	}
	return nil
}

func schemaVersion(
	conn *sql.DB,
) (
	int,
	error,
) {
	var version int
	if err := conn.QueryRow(
		`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`,
	).Scan(&version); err != nil {
		return 0, fmt.Errorf(
			"read schema version: %w",
			err,
		)
	}
	return version, nil
}

func latestSchemaVersion() int {
	if len(migrations) == 0 {
		return 0
	}
	return migrations[len(migrations)-1].version
}

func applyMigration(
	conn *sql.DB,
	item migration,
) error {
	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf(
			"begin transaction %d: %w",
			item.version,
			err,
		)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	for _, statement := range item.statements {
		if _, err := tx.Exec(statement); err != nil {
			return fmt.Errorf(
				"apply migration %d: %w",
				item.version,
				err,
			)
		}
	}

	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`,
		item.version,
		time.Now().UTC().Format(time.RFC3339Nano),
	); err != nil {
		return fmt.Errorf(
			"record migration %d: %w",
			item.version,
			err,
		)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(
			"commit transaction %d: %w",
			item.version,
			err,
		)
	}

	committed = true
	return nil
}
