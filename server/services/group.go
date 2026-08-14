package services

import (
	"database/sql"
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
func (s *GroupService) CreateGroup(groupID, creatorID string) (*models.GroupResponse, error) {
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
	if err := s.groups.AddMember(&models.GroupMember{
		GroupID:  groupID,
		UserID:   creatorID,
		JoinedAt: g.CreatedAt,
	}); err != nil {
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

// ListInvites returns pending group invites directed at the given user.
func (s *GroupService) ListInvites(userID string) ([]models.GroupInviteResponse, error) {
	invs, err := s.invites.ListByUser(userID, models.GroupInvitePending)
	if err != nil {
		return nil, err
	}
	resp := make([]models.GroupInviteResponse, 0, len(invs))
	for i := range invs {
		resp = append(resp, models.GroupInviteResponse{
			GroupID:   invs[i].GroupID,
			UserID:    invs[i].UserID,
			InvitedBy: invs[i].InvitedBy,
			CreatedAt: invs[i].CreatedAt,
		})
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

// Leave removes the user from the group unilaterally.
func (s *GroupService) Leave(groupID, userID string) error {
	err := s.groups.RemoveMember(groupID, userID)
	if err == sql.ErrNoRows {
		return ErrNotFound("not a member of the group")
	}
	return err
}

// UploadGroupPicture stores a new encrypted group profile picture. The caller
// must be a member of the group. Version must be strictly greater than the
// currently stored version.
func (s *GroupService) UploadGroupPicture(groupID, userID string, ciphertext, nonce []byte, version int64) error {
	if len(ciphertext) == 0 {
		return ErrInvalidInput("ciphertext must not be empty")
	}
	if len(nonce) != 12 {
		return ErrInvalidInput("nonce must be 12 bytes (AES-GCM)")
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
		GroupID:    groupID,
		Ciphertext: ciphertext,
		Nonce:      nonce,
		Version:    version,
		UpdatedAt:  time.Now().Unix(),
	})
}

// GetGroupPicture returns the encrypted group profile picture for the given
// group if the caller is a member.
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
