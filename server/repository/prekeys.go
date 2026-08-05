package repository

import (
	"database/sql"
	"errors"

	"server/models"
)

// PrekeyRepository stores X3DH public key bundles as opaque bytes.
type PrekeyRepository interface {
	Upsert(bundle *models.PrekeyBundle) error
	GetByUserID(userID string) (*models.PrekeyBundle, error)
	Delete(userID string) error
}

type prekeyRepo struct {
	db *sql.DB
}

func (r *prekeyRepo) Upsert(b *models.PrekeyBundle) error {
	_, err := r.db.Exec(
		`INSERT INTO prekey_bundles (user_id, identity_key, signed_prekey, signed_prekey_signature, one_time_prekey)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(user_id) DO UPDATE SET
			identity_key = excluded.identity_key,
			signed_prekey = excluded.signed_prekey,
			signed_prekey_signature = excluded.signed_prekey_signature,
			one_time_prekey = excluded.one_time_prekey`,
		b.UserID, b.IdentityKey, b.SignedPrekey, b.SignedPrekeySignature, b.OneTimePrekey,
	)
	return err
}

func (r *prekeyRepo) GetByUserID(userID string) (*models.PrekeyBundle, error) {
	var b models.PrekeyBundle
	err := r.db.QueryRow(
		`SELECT user_id, identity_key, signed_prekey, signed_prekey_signature, one_time_prekey
		 FROM prekey_bundles WHERE user_id = ?`, userID,
	).Scan(&b.UserID, &b.IdentityKey, &b.SignedPrekey, &b.SignedPrekeySignature, &b.OneTimePrekey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &b, nil
}

func (r *prekeyRepo) Delete(userID string) error {
	res, err := r.db.Exec(`DELETE FROM prekey_bundles WHERE user_id = ?`, userID)
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
