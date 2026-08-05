package database

import (
	"database/sql"
	"embed"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// MigrateDB runs all pending migrations against the given database.
func MigrateDB(db *sql.DB) error {
	return Migrate(db, MigrationsFS)
}
