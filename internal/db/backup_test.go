package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBackupCopiesCurrentDatabase(
	t *testing.T,
) {
	dir := testBackupDir(t)
	sourcePath := filepath.Join(
		dir,
		"source.sqlite",
	)
	backupPath := filepath.Join(
		dir,
		"backup.sqlite",
	)

	source, err := Open(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(
		func() {
			_ = source.Close()
		},
	)

	insertTestHost(
		t,
		source,
		"one",
		"first host",
	)

	if err := Backup(
		source,
		backupPath,
	); err != nil {
		t.Fatal(err)
	}

	backup, err := sql.Open(
		"sqlite",
		backupPath,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(
		func() {
			_ = backup.Close()
		})

	if got := testHostCount(
		t,
		backup,
	); got != 1 {
		t.Fatalf(
			"backup host count = %d; want 1",
			got,
		)
	}

	insertTestHost(
		t,
		source,
		"two",
		"added after backup",
	)

	if got := testHostCount(
		t,
		backup,
	); got != 1 {
		t.Fatalf(
			"backup changed after source write: got %d; want 1",
			got,
		)
	}

}

func TestBackupRejectsExistingFile(
	t *testing.T,
) {
	dir := testBackupDir(t)
	sourcePath := filepath.Join(
		dir,
		"source.sqlite",
	)
	backupPath := filepath.Join(
		dir,
		"backup.sqlite",
	)

	source, err := Open(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = source.Close()
	})

	if err := os.WriteFile(
		backupPath,
		[]byte("keep this file"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	err = Backup(
		source,
		backupPath,
	)
	if err == nil {
		t.Fatal(
			"Backup() error = nil, want existing file error")
	}
}

func TestBackupFilename(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.June,
		26,
		12,
		34,
		56,
		0,
		time.UTC,
	)

	got := BackupFilename(
		"v0.1.1",
		"v0.2.0",
		now,
	)
	want := "pinter-0.1.1-to-0.2.0-20260626-123456.sqlite"

	if got != want {
		t.Fatalf(
			"BackupFilename() = %q, want %q",
			got,
			want,
		)
	}
}

func testBackupDir(
	t *testing.T,
) string {
	t.Helper()

	cacheDir, err := filepath.Abs(
		filepath.Join(
			"..",
			"..",
			".cache",
		))
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(
		cacheDir,
		0o700,
	); err != nil {
		t.Fatal(err)
	}

	dir, err := os.MkdirTemp(
		cacheDir,
		"pinter-backup-test-",
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	return dir

}

func insertTestHost(
	t *testing.T,
	conn *sql.DB,
	alias string,
	notes string,
) {
	t.Helper()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := conn.Exec(
		`INSERT INTO hosts (
			id,
			alias,
			hostname,
			port,
			username,
			identity_file,
			notes,
			favorite,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"hst_"+alias,
		alias,
		"127.0.0.1",
		22,
		"",
		"",
		notes,
		0,
		now,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func testHostCount(
	t *testing.T,
	conn *sql.DB,
) int {
	t.Helper()

	var count int
	if err := conn.QueryRow(
		`SELECT COUNT(*) FROM hosts`,
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
