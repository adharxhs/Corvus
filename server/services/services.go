package services

import (
	"time"

	"server/repository"
)

// Services aggregates all business-logic services.
type Services struct {
	Users           *UserService
	Groups          *GroupService
	Prekeys         *PrekeyService
	Relationships   *RelationshipService
	ProfilePictures *ProfilePictureService
}

// New returns all services wired to the given repository bundle.
// chatRequestCooldown controls the minimum gap between repeated requests to
// the same recipient after a rejection (default: 24h, configurable via
// CHAT_REQUEST_COOLDOWN).
func New(repos *repository.Repository, chatRequestCooldown time.Duration) *Services {
	return &Services{
		Users:         NewUserService(repos.Users),
		Groups:        NewGroupService(repos.Groups, repos.Invites, repos.Relationships),
		Prekeys:       NewPrekeyService(repos.Prekeys),
		Relationships: NewRelationshipService(repos.Users, repos.Relationships, chatRequestCooldown),
		ProfilePictures: NewProfilePictureService(repos.ProfilePictures, repos.Relationships),
	}
}
