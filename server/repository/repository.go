package repository

import "database/sql"

// Repository bundles the per-table repositories backed by a single *sql.DB.
type Repository struct {
	db         *sql.DB
	Users      UserRepository
	Prekeys    PrekeyRepository
	Messages   MessageRepository
	Groups     GroupRepository
	SenderKeys SenderKeyRepository
}

// New returns a Repository wired to the given database handle.
func New(db *sql.DB) *Repository {
	return &Repository{
		db:         db,
		Users:      &userRepo{db: db},
		Prekeys:    &prekeyRepo{db: db},
		Messages:   &messageRepo{db: db},
		Groups:     &groupRepo{db: db},
		SenderKeys: &senderKeyRepo{db: db},
	}
}

// WithTx runs fn inside a transaction, rolling back on any error.
func (r *Repository) WithTx(fn func(tx *sql.Tx) error) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
