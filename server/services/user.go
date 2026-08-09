package services

import (
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

// GetByUsername returns a user by exact username match.
func (s *UserService) GetByUsername(username string) (*models.UserResponse, error) {
	u, err := s.users.GetByUsername(username)
	if err != nil {
		return nil, mapDBError(err)
	}
	resp := u.ToResponse()
	return &resp, nil
}
