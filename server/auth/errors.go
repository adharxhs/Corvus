package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDuplicateUsername  = errors.New("username already exists")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("invalid token")
)
