package repository

import (
	"database/sql"

	"server/models"
)

// RelationshipRepository manages the chat-request relationship table.
type RelationshipRepository interface {
	Create(rel *models.Relationship) error
	Get(requesterID, recipientID string) (*models.Relationship, error)
	ListByRecipient(recipientID string, status models.RelationshipStatus) ([]models.Relationship, error)
	UpdateStatus(requesterID, recipientID string, status models.RelationshipStatus) error
	HasAcceptedBetween(a, b string) (bool, error)
	AcceptedPeers(userID string) ([]string, error)
}

type relationshipRepo struct {
	db *sql.DB
}

func (r *relationshipRepo) Create(rel *models.Relationship) error {
	_, err := r.db.Exec(
		`INSERT INTO relationships (requester_id, recipient_id, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		rel.RequesterID, rel.RecipientID, rel.Status, rel.CreatedAt, rel.UpdatedAt,
	)
	return err
}

func (r *relationshipRepo) Get(requesterID, recipientID string) (*models.Relationship, error) {
	var rel models.Relationship
	err := r.db.QueryRow(
		`SELECT requester_id, recipient_id, status, created_at, updated_at
		 FROM relationships WHERE requester_id = ? AND recipient_id = ?`,
		requesterID, recipientID,
	).Scan(&rel.RequesterID, &rel.RecipientID, &rel.Status, &rel.CreatedAt, &rel.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

func (r *relationshipRepo) ListByRecipient(recipientID string, status models.RelationshipStatus) ([]models.Relationship, error) {
	rows, err := r.db.Query(
		`SELECT requester_id, recipient_id, status, created_at, updated_at
		 FROM relationships WHERE recipient_id = ? AND status = ? ORDER BY created_at DESC`,
		recipientID, status,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rels []models.Relationship
	for rows.Next() {
		var rel models.Relationship
		if err := rows.Scan(&rel.RequesterID, &rel.RecipientID, &rel.Status, &rel.CreatedAt, &rel.UpdatedAt); err != nil {
			return nil, err
		}
		rels = append(rels, rel)
	}
	return rels, rows.Err()
}

func (r *relationshipRepo) UpdateStatus(requesterID, recipientID string, status models.RelationshipStatus) error {
	res, err := r.db.Exec(
		`UPDATE relationships SET status = ?, updated_at = strftime('%s','now')
		 WHERE requester_id = ? AND recipient_id = ?`,
		status, requesterID, recipientID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *relationshipRepo) HasAcceptedBetween(a, b string) (bool, error) {
	var one int
	err := r.db.QueryRow(
		`SELECT 1 FROM relationships
		 WHERE ((requester_id = ? AND recipient_id = ?) OR (requester_id = ? AND recipient_id = ?))
		   AND status = 'accepted'
		 LIMIT 1`,
		a, b, b, a,
	).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *relationshipRepo) AcceptedPeers(userID string) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT requester_id FROM relationships WHERE recipient_id = ? AND status = 'accepted'
		 UNION
		 SELECT recipient_id FROM relationships WHERE requester_id = ? AND status = 'accepted'`,
		userID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		peers = append(peers, id)
	}
	return peers, rows.Err()
}
