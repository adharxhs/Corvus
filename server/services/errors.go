package services

import (
	"database/sql"
)

// ErrNotFound is returned when a requested entity does not exist.
type ErrNotFound string

func (e ErrNotFound) Error() string { return string(e) }

// ErrConflict is returned when a uniqueness constraint is violated.
type ErrConflict string

func (e ErrConflict) Error() string { return string(e) }

// ErrInvalidInput is returned when input validation fails.
type ErrInvalidInput string

func (e ErrInvalidInput) Error() string { return string(e) }

func mapDBError(err error) error {
	if err == nil {
		return nil
	}
	if sql.ErrNoRows == err {
		return ErrNotFound("not found")
	}
	return err
}
