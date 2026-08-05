package services

import (
	"server/repository"
)

// Services aggregates all business-logic services.
type Services struct {
	Users   *UserService
	Groups  *GroupService
	Prekeys *PrekeyService
}

// New returns all services wired to the given repository bundle.
func New(repos *repository.Repository) *Services {
	return &Services{
		Users:   NewUserService(repos.Users),
		Groups:  NewGroupService(repos.Groups),
		Prekeys: NewPrekeyService(repos.Prekeys),
	}
}
