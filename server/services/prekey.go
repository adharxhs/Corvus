package services

import (
	"database/sql"

	"server/models"
	"server/repository"
)

// PrekeyService manages X3DH prekey bundle storage.
type PrekeyService struct {
	prekeys repository.PrekeyRepository
}

// NewPrekeyService returns a PrekeyService.
func NewPrekeyService(prekeys repository.PrekeyRepository) *PrekeyService {
	return &PrekeyService{prekeys: prekeys}
}

// Upsert stores or replaces a user's prekey bundle.
func (s *PrekeyService) Upsert(bundle *models.PrekeyBundle) error {
	return s.prekeys.Upsert(bundle)
}

// Get retrieves a user's prekey bundle.
func (s *PrekeyService) Get(userID string) (*models.PrekeyBundle, error) {
	b, err := s.prekeys.GetByUserID(userID)
	if err != nil {
		return nil, mapDBError(err)
	}
	return b, nil
}

// Delete removes a user's prekey bundle.
func (s *PrekeyService) Delete(userID string) error {
	err := s.prekeys.Delete(userID)
	if err == sql.ErrNoRows {
		return ErrNotFound("prekey bundle not found")
	}
	return err
}
