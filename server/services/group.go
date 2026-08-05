package services

import (
	"time"

	"server/models"
	"server/repository"
)

// GroupService manages groups and membership.
type GroupService struct {
	groups repository.GroupRepository
}

// NewGroupService returns a GroupService.
func NewGroupService(groups repository.GroupRepository) *GroupService {
	return &GroupService{groups: groups}
}

// CreateGroup creates a new group with the given ID.
func (s *GroupService) CreateGroup(groupID string) (*models.GroupResponse, error) {
	if groupID == "" {
		return nil, ErrInvalidInput("group_id is required")
	}
	g := &models.Group{
		ID:        groupID,
		CreatedAt: time.Now().Unix(),
	}
	if err := s.groups.Create(g); err != nil {
		return nil, err
	}
	return &models.GroupResponse{ID: g.ID, CreatedAt: g.CreatedAt}, nil
}

// AddMember adds a user to a group.
func (s *GroupService) AddMember(groupID, userID string) error {
	m := &models.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		JoinedAt: time.Now().Unix(),
	}
	if err := s.groups.AddMember(m); err != nil {
		return err
	}
	return nil
}

// ListMembers returns all members of a group.
func (s *GroupService) ListMembers(groupID string) ([]models.GroupMemberResponse, error) {
	members, err := s.groups.ListMembers(groupID)
	if err != nil {
		return nil, err
	}
	resp := make([]models.GroupMemberResponse, len(members))
	for i, m := range members {
		resp[i] = models.GroupMemberResponse{
			UserID:   m.UserID,
			JoinedAt: m.JoinedAt,
		}
	}
	return resp, nil
}
