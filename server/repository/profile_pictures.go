package repository

import (
	"database/sql"

	"server/models"
)

// ProfilePictureRepository stores client-side-encrypted profile pictures.
type ProfilePictureRepository interface {
	Upsert(pic *models.ProfilePicture) error
	Get(userID string) (*models.ProfilePicture, error)
}

type profilePictureRepo struct {
	db *sql.DB
}

func (r *profilePictureRepo) Upsert(pic *models.ProfilePicture) error {
	_, err := r.db.Exec(
		`INSERT INTO profile_pictures (user_id, image_data, version, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(user_id) DO UPDATE SET
			image_data = excluded.image_data,
			version = excluded.version,
			updated_at = excluded.updated_at`,
		pic.UserID, pic.ImageData, pic.Version, pic.UpdatedAt,
	)
	return err
}

func (r *profilePictureRepo) Get(userID string) (*models.ProfilePicture, error) {
	var pic models.ProfilePicture
	err := r.db.QueryRow(
		`SELECT user_id, image_data, version, updated_at
		 FROM profile_pictures WHERE user_id = ?`,
		userID,
	).Scan(&pic.UserID, &pic.ImageData, &pic.Version, &pic.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &pic, nil
}
