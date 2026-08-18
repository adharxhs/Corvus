package services

import (
	"database/sql"
	"strings"
	"time"

	"server/models"
	"server/repository"
)

// GroupService manages groups, membership, group invites, and group profile pictures.
type GroupService struct {
	groups              repository.GroupRepository
	invites             repository.InviteRepository
	relationships       repository.RelationshipRepository
	groupProfilePictures repository.GroupProfilePictureRepository
}

// NewGroupService returns a GroupService.
func NewGroupService(groups repository.GroupRepository, invites repository.InviteRepository, relationships repository.RelationshipRepository, groupProfilePictures repository.GroupProfilePictureRepository) *GroupService {
	return &GroupService{
		groups:              groups,
		invites:             invites,
		relationships:       relationships,
		groupProfilePictures: groupProfilePictures,
	}
}

// CreateGroup creates a new group and auto-adds the creator as the first member.
func (s *GroupService) CreateGroup(groupID, name, creatorID string) (*models.GroupResponse, error) {
	if groupID == "" {
		return nil, ErrInvalidInput("group_id is required")
	}
	if len([]rune(name)) > 60 {
		return nil, ErrInvalidInput("name must be 60 characters or fewer")
	}
	if strings.TrimSpace(name) == "" {
		name = "New Group"
	}
	g := &models.Group{
		ID:        groupID,
		Name:      strings.TrimSpace(name),
		CreatedAt: time.Now().Unix(),
	}
	if err := s.groups.Create(g); err != nil {
		return nil, err
	}
	if err := s.groups.AddMember(&models.GroupMember{
		GroupID:  groupID,
		UserID:   creatorID,
		JoinedAt: g.CreatedAt,
	}); err != nil {
		return nil, err
	}
	return &models.GroupResponse{ID: g.ID, Name: g.Name, CreatedAt: g.CreatedAt}, nil
}

// GetGroup returns basic group info (id, name, created_at).
func (s *GroupService) GetGroup(groupID string) (*models.GroupResponse, error) {
	g, err := s.groups.GetByID(groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound("group not found")
		}
		return nil, err
	}
	return &models.GroupResponse{ID: g.ID, Name: g.Name, CreatedAt: g.CreatedAt}, nil
}

// RenameGroup updates a group's name. The caller must be a member.
func (s *GroupService) RenameGroup(groupID, userID, name string) error {
	isMember, err := s.groups.IsMember(groupID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotMember("not a member of the group")
	}
	trimmed := strings.TrimSpace(name)
	if len([]rune(trimmed)) > 60 {
		return ErrInvalidInput("name must be 60 characters or fewer")
	}
	if trimmed == "" {
		return ErrInvalidInput("name must not be empty")
	}
	return s.groups.Rename(groupID, trimmed)
}

// AddMember adds a user to a group.
func (s *GroupService) AddMember(groupID, userID string) error {
	m := &models.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		JoinedAt: time.Now().Unix(),
	}
	return s.groups.AddMember(m)
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

// Invite sends a pending group invite. The inviter must be a member of the
// group and must have an accepted personal-chat relationship with the invitee.
func (s *GroupService) Invite(groupID, inviterID, userID string) error {
	isMember, err := s.groups.IsMember(groupID, inviterID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotMember("inviter is not a member of the group")
	}
	accepted, err := s.relationships.HasAcceptedBetween(inviterID, userID)
	if err != nil {
		return err
	}
	if !accepted {
		return ErrNotAccepted("no accepted relationship with invitee")
	}

	now := time.Now().Unix()
	inv, err := s.invites.Get(groupID, userID)
	if err == nil {
		switch inv.Status {
		case models.GroupInvitePending:
			return ErrConflict("invite already pending")
		case models.GroupInviteAccepted:
			return ErrConflict("user is already a member")
		case models.GroupInviteRejected:
			return s.invites.UpdateStatus(groupID, userID, models.GroupInvitePending)
		}
	}
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	inv = &models.GroupInvite{
		GroupID:   groupID,
		UserID:    userID,
		InvitedBy: inviterID,
		Status:    models.GroupInvitePending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.invites.Create(inv)
}

// ListByUser returns all groups the given user is a member of.
func (s *GroupService) ListByUser(userID string) ([]models.GroupResponse, error) {
	groups, err := s.groups.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	resp := make([]models.GroupResponse, len(groups))
	for i, g := range groups {
		resp[i] = models.GroupResponse{ID: g.ID, Name: g.Name, CreatedAt: g.CreatedAt}
	}
	return resp, nil
}

// ListInvites returns pending group invites directed at the given user,
// enriched with group name and inviter username.
func (s *GroupService) ListInvites(userID string) ([]models.GroupInviteResponse, error) {
	invs, err := s.invites.ListByUser(userID, models.GroupInvitePending)
	if err != nil {
		return nil, err
	}
	resp := make([]models.GroupInviteResponse, 0, len(invs))
	for i := range invs {
		r := models.GroupInviteResponse{
			GroupID:   invs[i].GroupID,
			UserID:    invs[i].UserID,
			InvitedBy: invs[i].InvitedBy,
			CreatedAt: invs[i].CreatedAt,
		}
		if g, err := s.groups.GetByID(invs[i].GroupID); err == nil {
			r.GroupName = g.Name
		}
		resp = append(resp, r)
	}
	return resp, nil
}

// AcceptInvite adds the invitee to the group and marks the invite accepted.
func (s *GroupService) AcceptInvite(groupID, userID string) error {
	inv, err := s.invites.Get(groupID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound("invite not found")
		}
		return err
	}
	if inv.Status != models.GroupInvitePending {
		return ErrConflict("invite is not pending")
	}
	if err := s.groups.AddMember(&models.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		JoinedAt: time.Now().Unix(),
	}); err != nil {
		return err
	}
	return s.invites.UpdateStatus(groupID, userID, models.GroupInviteAccepted)
}

// RejectInvite marks the invite rejected. The invitee does not join the group.
func (s *GroupService) RejectInvite(groupID, userID string) error {
	inv, err := s.invites.Get(groupID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound("invite not found")
		}
		return err
	}
	if inv.Status != models.GroupInvitePending {
		return ErrConflict("invite is not pending")
	}
	return s.invites.UpdateStatus(groupID, userID, models.GroupInviteRejected)
}

// Leave removes the user from the group unilaterally.
func (s *GroupService) Leave(groupID, userID string) error {
	err := s.groups.RemoveMember(groupID, userID)
	if err == sql.ErrNoRows {
		return ErrNotFound("not a member of the group")
	}
	return err
}

func (s *GroupService) UploadGroupPicture(groupID, userID string, imageData []byte, version int64) error {
	if len(imageData) == 0 {
		return ErrInvalidInput("image_data must not be empty")
	}
	if version <= 0 {
		return ErrInvalidInput("version must be positive")
	}

	isMember, err := s.groups.IsMember(groupID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotMember("caller is not a member of the group")
	}

	existing, err := s.groupProfilePictures.Get(groupID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil && existing.Version >= version {
		return ErrConflict("version must be newer than the stored version")
	}

	return s.groupProfilePictures.Upsert(&models.GroupProfilePicture{
		GroupID:   groupID,
		ImageData: imageData,
		Version:   version,
		UpdatedAt: time.Now().Unix(),
	})
}

func (s *GroupService) GetGroupPicture(groupID, userID string) (*models.GroupProfilePicture, error) {
	isMember, err := s.groups.IsMember(groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotMember("caller is not a member of the group")
	}

	pic, err := s.groupProfilePictures.Get(groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound("no profile picture for this group")
		}
		return nil, err
	}
	return pic, nil
}
