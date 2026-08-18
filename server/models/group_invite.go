package models

type GroupInviteStatus string

const (
	GroupInvitePending  GroupInviteStatus = "pending"
	GroupInviteAccepted GroupInviteStatus = "accepted"
	GroupInviteRejected GroupInviteStatus = "rejected"
)

// GroupInvite tracks a pending group membership request. The invitee must
// explicitly accept before they become a member.
type GroupInvite struct {
	GroupID   string
	UserID    string
	InvitedBy string
	Status    GroupInviteStatus
	CreatedAt int64
	UpdatedAt int64
}

type GroupInviteResponse struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
	UserID    string `json:"user_id"`
	InvitedBy string `json:"invited_by"`
	CreatedAt int64  `json:"created_at"`
}

type GroupInviteRequest struct {
	UserID string `json:"user_id"`
}
