package models

type RelationshipStatus string

const (
	RelationshipPending  RelationshipStatus = "pending"
	RelationshipAccepted RelationshipStatus = "accepted"
	RelationshipRejected RelationshipStatus = "rejected"
)

// Relationship tracks a chat-request state machine between two users:
// pending -> accepted | rejected. Accept is bidirectional and permanent.
type Relationship struct {
	RequesterID string
	RecipientID string
	Status      RelationshipStatus
	CreatedAt   int64
	UpdatedAt   int64
}

type RelationshipResponse struct {
	RequesterID string `json:"requester_id"`
	RecipientID string `json:"recipient_id"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type ChatRequestRequest struct {
	RecipientID string `json:"recipient_id"`
}
