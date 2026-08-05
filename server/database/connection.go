package database

import "database/sql"

// Close closes the database connection.
func Close(db *sql.DB) error {
	return db.Close()
}

// Ping verifies the database connection is alive.
func Ping(db *sql.DB) error {
	return db.Ping()
}
