package services

import (
	"database/sql"
	"time"

	"server/models"
	"server/repository"
)

// ErrCooldownActive is returned when a rejected chat request is re-sent
// before the cooldown has elapsed.
type ErrCooldownActive struct {
	RetryAfter time.Duration
}

func (e ErrCooldownActive) Error() string {
	return "chat request re-send is on cooldown"
}

// RelationshipService manages the chat-request lifecycle between users.
type RelationshipService struct {
	users     repository.UserRepository
	rels      repository.RelationshipRepository
	cooldown  time.Duration
}

// NewRelationshipService returns a RelationshipService backed by the given repos.
func NewRelationshipService(users repository.UserRepository, rels repository.RelationshipRepository, cooldown time.Duration) *RelationshipService {
	return &RelationshipService{
		users:    users,
		rels:     rels,
		cooldown: cooldown,
	}
}

// SendRequest initiates or re-arms a pending chat request from requester to
// recipient. Rejects are silent; re-arming after rejection is cooldown-gated.
func (s *RelationshipService) SendRequest(requesterID, recipientID string) (*models.RelationshipResponse, error) {
	if requesterID == recipientID {
		return nil, ErrInvalidInput("cannot send a chat request to yourself")
	}
	if _, err := s.users.GetByID(recipientID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound("recipient not found")
		}
		return nil, err
	}

	now := time.Now().Unix()
	rel, err := s.rels.Get(requesterID, recipientID)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		rel = &models.Relationship{
			RequesterID: requesterID,
			RecipientID: recipientID,
			Status:      models.RelationshipPending,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := s.rels.Create(rel); err != nil {
			return nil, err
		}
		return toRelationshipResponse(rel), nil
	}

	switch rel.Status {
	case models.RelationshipAccepted:
		return nil, ErrConflict("already connected")
	case models.RelationshipPending:
		return nil, ErrConflict("request already pending")
	case models.RelationshipRejected:
		elapsed := time.Since(time.Unix(rel.UpdatedAt, 0))
		if elapsed < s.cooldown {
			return nil, ErrCooldownActive{RetryAfter: s.cooldown - elapsed}
		}
		if err := s.rels.UpdateStatus(requesterID, recipientID, models.RelationshipPending); err != nil {
			return nil, err
		}
		rel.Status = models.RelationshipPending
		rel.UpdatedAt = time.Now().Unix()
		return toRelationshipResponse(rel), nil
	default:
		return nil, ErrInvalidInput("unknown relationship status")
	}
}

// ListPending returns all incoming pending chat requests for the given user.
func (s *RelationshipService) ListPending(recipientID string) ([]models.RelationshipResponse, error) {
	rels, err := s.rels.ListByRecipient(recipientID, models.RelationshipPending)
	if err != nil {
		return nil, err
	}
	resp := make([]models.RelationshipResponse, 0, len(rels))
	for i := range rels {
		resp = append(resp, *toRelationshipResponse(&rels[i]))
	}
	return resp, nil
}

// Respond handles the recipient's accept/reject action on an incoming request.
// Reject is silent; accept is bidirectional and permanent.
func (s *RelationshipService) Respond(recipientID, requesterID string, status models.RelationshipStatus) (*models.RelationshipResponse, error) {
	if status != models.RelationshipAccepted && status != models.RelationshipRejected {
		return nil, ErrInvalidInput("status must be accepted or rejected")
	}
	rel, err := s.rels.Get(requesterID, recipientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound("request not found")
		}
		return nil, err
	}
	if rel.RecipientID != recipientID {
		return nil, ErrNotFound("request not found")
	}
	if rel.Status != models.RelationshipPending {
		return nil, ErrConflict("request is not pending")
	}
	if err := s.rels.UpdateStatus(requesterID, recipientID, status); err != nil {
		return nil, err
	}
	rel.Status = status
	rel.UpdatedAt = time.Now().Unix()
	return toRelationshipResponse(rel), nil
}

func toRelationshipResponse(rel *models.Relationship) *models.RelationshipResponse {
	return &models.RelationshipResponse{
		RequesterID: rel.RequesterID,
		RecipientID: rel.RecipientID,
		Status:      string(rel.Status),
		CreatedAt:   rel.CreatedAt,
		UpdatedAt:   rel.UpdatedAt,
	}
}
