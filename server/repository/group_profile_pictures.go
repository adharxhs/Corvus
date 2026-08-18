package repository

import (
	"database/sql"

	"server/models"
)

type GroupProfilePictureRepository interface {
	Upsert(pic *models.GroupProfilePicture) error
	Get(groupID string) (*models.GroupProfilePicture, error)
}

type groupProfilePictureRepo struct {
	db *sql.DB
}

func (r *groupProfilePictureRepo) Upsert(pic *models.GroupProfilePicture) error {
	_, err := r.db.Exec(
		`INSERT INTO group_profile_pictures (group_id, image_data, version, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(group_id) DO UPDATE SET
			image_data = excluded.image_data,
			version = excluded.version,
			updated_at = excluded.updated_at`,
		pic.GroupID, pic.ImageData, pic.Version, pic.UpdatedAt,
	)
	return err
}

func (r *groupProfilePictureRepo) Get(groupID string) (*models.GroupProfilePicture, error) {
	var pic models.GroupProfilePicture
	err := r.db.QueryRow(
		`SELECT group_id, image_data, version, updated_at
		 FROM group_profile_pictures WHERE group_id = ?`,
		groupID,
	).Scan(&pic.GroupID, &pic.ImageData, &pic.Version, &pic.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &pic, nil
}
