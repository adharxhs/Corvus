package models

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
