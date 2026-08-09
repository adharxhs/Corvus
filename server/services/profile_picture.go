package services

import (
	"database/sql"
	"time"

	"server/models"
	"server/repository"
)

// ProfilePictureService stores/retrieves encrypted profile pictures and
// enforces that callers share an accepted relationship before viewing.
type ProfilePictureService struct {
	pics          repository.ProfilePictureRepository
	relationships repository.RelationshipRepository
}

// NewProfilePictureService returns a ProfilePictureService.
func NewProfilePictureService(pics repository.ProfilePictureRepository, rels repository.RelationshipRepository) *ProfilePictureService {
	return &ProfilePictureService{pics: pics, relationships: rels}
}

// Upload stores a new ciphertext+nonce for the authenticated user. Version must
// be strictly greater than the currently stored version (no re-uploads of the
// same version). No rotation — the client re-encrypts under the same profile
// key and increments the version.
func (s *ProfilePictureService) Upload(userID string, ciphertext, nonce []byte, version int64) error {
	if len(ciphertext) == 0 {
		return ErrInvalidInput("ciphertext must not be empty")
	}
	if len(nonce) != 12 {
		return ErrInvalidInput("nonce must be 12 bytes (AES-GCM)")
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
		UserID:     userID,
		Ciphertext: ciphertext,
		Nonce:      nonce,
		Version:    version,
		UpdatedAt:  time.Now().Unix(),
	})
}

// Get returns the profile picture for userID if the caller shares an accepted
// relationship with them.
func (s *ProfilePictureService) Get(callerID, userID string) (*models.ProfilePicture, error) {
	ok, err := s.relationships.HasAcceptedBetween(callerID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotAccepted("no accepted relationship")
	}
	pic, err := s.pics.Get(userID)
	if err != nil {
		return nil, mapDBError(err)
	}
	return pic, nil
}
