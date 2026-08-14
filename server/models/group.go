package models

type Group struct {
	ID        string
	CreatedAt int64
}

type GroupResponse struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"created_at"`
}

type GroupMember struct {
	GroupID  string
	UserID   string
	JoinedAt int64
}

type GroupMemberResponse struct {
	UserID   string `json:"user_id"`
	JoinedAt int64  `json:"joined_at"`
}

type CreateGroupRequest struct {
	GroupID string `json:"group_id"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id"`
}

type GroupProfilePicture struct {
	GroupID    string
	Ciphertext []byte
	Nonce      []byte
	Version    int64
	UpdatedAt  int64
}

type GroupProfilePictureRequest struct {
	Ciphertext string `json:"ciphertext"`
	Nonce      string `json:"nonce"`
	Version    int64  `json:"version"`
}

type GroupProfilePictureResponse struct {
	Ciphertext string `json:"ciphertext"`
	Nonce      string `json:"nonce"`
	Version    int64  `json:"version"`
}
