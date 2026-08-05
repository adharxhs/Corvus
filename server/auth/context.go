package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyUsername contextKey = "username"
)

// ContextWithUser stores user identity in the request context.
func ContextWithUser(r *http.Request, userID, username string) *http.Request {
	ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
	ctx = context.WithValue(ctx, ContextKeyUsername, username)
	return r.WithContext(ctx)
}

// UserIDFromContext retrieves the authenticated user ID.
func UserIDFromContext(r *http.Request) (string, bool) {
	v := r.Context().Value(ContextKeyUserID)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// UsernameFromContext retrieves the authenticated username.
func UsernameFromContext(r *http.Request) (string, bool) {
	v := r.Context().Value(ContextKeyUsername)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// ExtractBearerToken pulls the token from the Authorization header.
func ExtractBearerToken(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return "", false
	}
	return strings.TrimPrefix(h, "Bearer "), true
}
