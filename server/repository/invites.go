package repository

import (
	"database/sql"

	"server/models"
)

// InviteRepository manages group membership invites.
type InviteRepository interface {
	Create(inv *models.GroupInvite) error
	Get(groupID, userID string) (*models.GroupInvite, error)
	ListByUser(userID string, status models.GroupInviteStatus) ([]models.GroupInvite, error)
	UpdateStatus(groupID, userID string, status models.GroupInviteStatus) error
}

type inviteRepo struct {
	db *sql.DB
}

func (r *inviteRepo) Create(inv *models.GroupInvite) error {
	_, err := r.db.Exec(
		`INSERT INTO group_invites (group_id, user_id, invited_by, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		inv.GroupID, inv.UserID, inv.InvitedBy, inv.Status, inv.CreatedAt, inv.UpdatedAt,
	)
	return err
}

func (r *inviteRepo) Get(groupID, userID string) (*models.GroupInvite, error) {
	var inv models.GroupInvite
	err := r.db.QueryRow(
		`SELECT group_id, user_id, invited_by, status, created_at, updated_at
		 FROM group_invites WHERE group_id = ? AND user_id = ?`,
		groupID, userID,
	).Scan(&inv.GroupID, &inv.UserID, &inv.InvitedBy, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *inviteRepo) ListByUser(userID string, status models.GroupInviteStatus) ([]models.GroupInvite, error) {
	rows, err := r.db.Query(
		`SELECT group_id, user_id, invited_by, status, created_at, updated_at
		 FROM group_invites WHERE user_id = ? AND status = ? ORDER BY created_at DESC`,
		userID, status,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []models.GroupInvite
	for rows.Next() {
		var inv models.GroupInvite
		if err := rows.Scan(&inv.GroupID, &inv.UserID, &inv.InvitedBy, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		invites = append(invites, inv)
	}
	return invites, rows.Err()
}

func (r *inviteRepo) UpdateStatus(groupID, userID string, status models.GroupInviteStatus) error {
	res, err := r.db.Exec(
		`UPDATE group_invites SET status = ?, updated_at = strftime('%s','now')
		 WHERE group_id = ? AND user_id = ?`,
		status, groupID, userID,
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
