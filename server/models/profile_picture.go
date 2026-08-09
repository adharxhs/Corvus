package models

// ProfilePicture stores only client-side-encrypted bytes (ciphertext + nonce).
// The symmetric profile key never reaches the server.
type ProfilePicture struct {
	UserID     string
	Ciphertext []byte
	Nonce      []byte
	Version    int64
	UpdatedAt  int64
}

type ProfilePictureRequest struct {
	Ciphertext string `json:"ciphertext"`
	Nonce      string `json:"nonce"`
	Version    int64  `json:"version"`
}

type ProfilePictureResponse struct {
	Ciphertext string `json:"ciphertext"`
	Nonce      string `json:"nonce"`
	Version    int64  `json:"version"`
}
