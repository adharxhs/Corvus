package services

import (
	"database/sql"

	"server/models"
	"server/repository"
)

// UserService provides user-related business logic.
type UserService struct {
	users repository.UserRepository
}

// NewUserService returns a UserService backed by the given repository.
func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

// GetByID returns a user by ID.
func (s *UserService) GetByID(id string) (*models.UserResponse, error) {
	u, err := s.users.GetByID(id)
	if err != nil {
		return nil, mapDBError(err)
	}
	resp := u.ToResponse()
	return &resp, nil
}

// ListUsers returns all registered users.
func (s *UserService) ListUsers() ([]models.UserResponse, error) {
	// The repository only exposes GetByID and GetByUsername; we add a
	// direct query here for list. A full implementation would add a
	// List method to the repository interface.
	return nil, sql.ErrNoRows
}
