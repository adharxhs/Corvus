package repository

import (
	"database/sql"
	"errors"

	"server/models"
)

// UserRepository provides access to the users table. It performs no
// authentication or password handling.
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
}

type userRepo struct {
	db *sql.DB
}

func (r *userRepo) Create(u *models.User) error {
	_, err := r.db.Exec(
		`INSERT INTO users (id, username, password_hash, created_at) VALUES (?, ?, ?, ?)`,
		u.ID, u.Username, u.PasswordHash, u.CreatedAt,
	)
	return err
}

func (r *userRepo) GetByID(id string) (*models.User, error) {
	return scanUser(r.db.QueryRow(
		`SELECT id, username, password_hash, created_at FROM users WHERE id = ?`, id,
	))
}

func (r *userRepo) GetByUsername(username string) (*models.User, error) {
	return scanUser(r.db.QueryRow(
		`SELECT id, username, password_hash, created_at FROM users WHERE username = ?`, username,
	))
}

func scanUser(row *sql.Row) (*models.User, error) {
	var u models.User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) Update(u *models.User) error {
	res, err := r.db.Exec(
		`UPDATE users SET username = ?, password_hash = ?, created_at = ? WHERE id = ?`,
		u.Username, u.PasswordHash, u.CreatedAt, u.ID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *userRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
