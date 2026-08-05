package models

type PendingMessage struct {
	ID          string
	RecipientID string
	Ciphertext  []byte
	Delivered   bool
}

type SenderKeyDistribution struct {
	ID          string
	GroupID     string
	RecipientID string
	Ciphertext  []byte
	Delivered   bool
}
