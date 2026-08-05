package auth

import "github.com/golang-jwt/jwt/v5"

// Claims extends standard JWT claims with application-specific fields.
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
