package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

type migration struct {
	version string
	name    string
	sql     string
}

// Migrate reads SQL migration files from the given filesystem and applies
// them in order. It uses a schema_migrations table to track applied versions.
func Migrate(db *sql.DB, migrationsFS fs.FS) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at INTEGER NOT NULL
	)`); err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	applied, err := appliedVersions(db)
	if err != nil {
		return fmt.Errorf("reading applied versions: %w", err)
	}

	entries, err := readMigrationFiles(migrationsFS)
	if err != nil {
		return fmt.Errorf("reading migration files: %w", err)
	}

	for _, m := range entries {
		if applied[m.version] {
			continue
		}
		if err := applyMigration(db, m); err != nil {
			return fmt.Errorf("applying migration %s: %w", m.name, err)
		}
	}

	return nil
}

func appliedVersions(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions[v] = true
	}
	return versions, rows.Err()
}

func readMigrationFiles(fsys fs.FS) ([]migration, error) {
	var files []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".sql") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	var migrations []migration
	for _, f := range files {
		data, err := fs.ReadFile(fsys, f)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", f, err)
		}
		base := filepath.Base(f)
		version := strings.SplitN(base, "_", 2)[0]
		migrations = append(migrations, migration{
			version: version,
			name:    base,
			sql:     string(data),
		})
	}
	return migrations, nil
}

func applyMigration(db *sql.DB, m migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(m.sql); err != nil {
		return err
	}

	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (version, applied_at) VALUES (?, strftime('%s','now'))`,
		m.version,
	); err != nil {
		return err
	}

	return tx.Commit()
}
