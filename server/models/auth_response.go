package models

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
