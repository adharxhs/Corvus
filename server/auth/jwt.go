package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager handles creation and validation of JWT tokens.
type JWTManager struct {
	secret     []byte
	expiration time.Duration
}

// NewJWTManager returns a manager that signs tokens with the given secret.
func NewJWTManager(secret string, expiration time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), expiration: expiration}
}

// Issue creates a signed JWT for the given user.
func (m *JWTManager) Issue(userID, username string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Validate parses and verifies a token string, returning its claims.
func (m *JWTManager) Validate(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	if !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}
