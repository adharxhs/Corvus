package services

import (
	"database/sql"
	"time"

	"server/models"
	"server/repository"
)

type ProfilePictureService struct {
	pics          repository.ProfilePictureRepository
	relationships repository.RelationshipRepository
}

func NewProfilePictureService(pics repository.ProfilePictureRepository, rels repository.RelationshipRepository) *ProfilePictureService {
	return &ProfilePictureService{pics: pics, relationships: rels}
}

func (s *ProfilePictureService) Upload(userID string, imageData []byte, version int64) error {
	if len(imageData) == 0 {
		return ErrInvalidInput("image_data must not be empty")
	}
	if version <= 0 {
		return ErrInvalidInput("version must be positive")
	}

	existing, err := s.pics.Get(userID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil && existing.Version >= version {
		return ErrConflict("version must be newer than the stored version")
	}

	return s.pics.Upsert(&models.ProfilePicture{
		UserID:    userID,
		ImageData: imageData,
		Version:   version,
		UpdatedAt: time.Now().Unix(),
	})
}

func (s *ProfilePictureService) Get(callerID, userID string) (*models.ProfilePicture, error) {
	pic, err := s.pics.Get(userID)
	if err != nil {
		return nil, mapDBError(err)
	}
	return pic, nil
}
