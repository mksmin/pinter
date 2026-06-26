package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Backup(
	conn *sql.DB,
	destination string,
) error {
	if err := os.MkdirAll(
		filepath.Dir(destination),
		0o700,
	); err != nil {
		return fmt.Errorf(
			"create backup directory: %w",
			err,
		)
	}

	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf(
			"backup already exists: %s",
			destination,
		)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf(
			"check backup destination: %w",
			err,
		)
	}

	escaped := strings.ReplaceAll(
		destination,
		"'",
		"''",
	)
	if _, err := conn.Exec(
		`VACUUM INTO '` + escaped + `'`,
	); err != nil {
		return fmt.Errorf(
			"create database backup: %w",
			err,
		)
	}

	if err := verifyBackup(destination); err != nil {
		_ = os.Remove(destination)
		return err
	}

	return nil
}

func verifyBackup(
	path string,
) error {
	conn, err := sql.Open(
		"sqlite",
		path,
	)
	if err != nil {
		return fmt.Errorf(
			"open database backup: %w",
			err,
		)
	}
	defer conn.Close()

	var result string
	if err := conn.QueryRow(
		`PRAGMA integrity_check`,
	).Scan(&result); err != nil {
		return fmt.Errorf(
			"verify database backup: %w",
			err,
		)
	}
	if result != "ok" {
		return fmt.Errorf(
			"database backup integrity check failed: %s",
			result,
		)
	}
	return nil
}

func BackupFilename(
	fromVersion string,
	toVersion string,
	now time.Time,
) string {
	fromVersion = sanitizeVersion(fromVersion)
	toVersion = sanitizeVersion(toVersion)

	return fmt.Sprintf(
		"pinter-%s-to-%s-%s.sqlite",
		fromVersion,
		toVersion,
		now.UTC().Format("20060102-150405"),
	)
}

func sanitizeVersion(
	value string,
) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(
		value,
		"v",
	)
	if value == "" {
		return "unknown"
	}

	return strings.NewReplacer(
		"/", "-",
		`\`, "-",
		":", "-",
		" ", "-",
	).Replace(value)
}
