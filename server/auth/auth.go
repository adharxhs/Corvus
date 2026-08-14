package auth

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"server/models"
	"server/repository"
)

// Service provides authentication operations: registration, login, and
// password verification. It coordinates the user repository with password
// hashing and JWT issuance.
type Service struct {
	users   repository.UserRepository
	jwt     *JWTManager
	clock   func() time.Time
}

// NewService returns an auth Service wired to the given repository and JWT
// configuration.
func NewService(users repository.UserRepository, secret string, exp time.Duration) *Service {
	return &Service{
		users: users,
		jwt:   NewJWTManager(secret, exp),
		clock: time.Now,
	}
}

// JWT returns the underlying JWT manager for use by middleware.
func (s *Service) JWT() *JWTManager { return s.jwt }

// Register creates a new user. Returns ErrDuplicateUsername if the name is taken.
func (s *Service) Register(username, password string) (*models.AuthResponse, error) {
	username = trimSpace(username)
	if len(username) < 3 || len(username) > 64 {
		return nil, ErrInvalidRequest
	}
	if len(password) < 8 || len(password) > 256 {
		return nil, ErrInvalidRequest
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: hash,
		CreatedAt:    s.clock().Unix(),
	}

	if err := s.users.Create(user); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateUsername
		}
		return nil, err
	}

	token, err := s.jwt.Issue(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{Token: token, User: user.ToResponse()}, nil
}

// Login verifies credentials and returns a JWT on success.
func (s *Service) Login(username, password string) (*models.LoginResponse, error) {
	username = trimSpace(username)
	user, err := s.users.GetByUsername(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	ok, err := VerifyPassword(password, user.PasswordHash)
	if err != nil || !ok {
		return nil, ErrInvalidCredentials
	}

	token, err := s.jwt.Issue(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// ChangePassword updates a user's password after verifying their current password.
func (s *Service) ChangePassword(userID, currentPassword, newPassword string) error {
	if len(newPassword) < 8 || len(newPassword) > 256 {
		return ErrInvalidRequest
	}

	user, err := s.users.GetByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return err
	}

	ok, err := VerifyPassword(currentPassword, user.PasswordHash)
	if err != nil || !ok {
		return ErrInvalidCredentials
	}

	newHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	return s.users.Update(user)
}

func trimSpace(s string) string {
	for len(s) > 0 && s[0] == ' ' {
		s = s[1:]
	}
	for len(s) > 0 && s[len(s)-1] == ' ' {
		s = s[:len(s)-1]
	}
	return s
}

func isUniqueViolation(err error) bool {
	return err != nil && (errors.Is(err, sql.ErrNoRows) == false) &&
		contains(err.Error(), "UNIQUE constraint failed")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && containsSubstr(s, sub)
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
