package models

type User struct {
	ID           string
	Username     string
	PasswordHash string
	CreatedAt    int64
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{ID: u.ID, Username: u.Username}
}
