package database

import (
	"database/sql"
	"strings"

	_ "modernc.org/sqlite"
)

const driverName = "sqlite"

// Open opens a connection to the SQLite database at the given path.
// Foreign key enforcement is enabled on every connection.
func Open(path string) (*sql.DB, error) {
	dsn := path
	if !strings.HasPrefix(path, "file:") {
		dsn = "file:" + path
	}
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	dsn = dsn + sep + "_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
