package repository

import (
	"database/sql"
	"errors"

	"server/models"
)

// GroupRepository manages group metadata and membership.
type GroupRepository interface {
	Create(group *models.Group) error
	GetByID(id string) (*models.Group, error)
	AddMember(member *models.GroupMember) error
	RemoveMember(groupID, userID string) error
	ListMembers(groupID string) ([]models.GroupMember, error)
}

type groupRepo struct {
	db *sql.DB
}

func (r *groupRepo) Create(g *models.Group) error {
	_, err := r.db.Exec(
		`INSERT INTO groups (id, created_at) VALUES (?, ?)`,
		g.ID, g.CreatedAt,
	)
	return err
}

func (r *groupRepo) GetByID(id string) (*models.Group, error) {
	var g models.Group
	err := r.db.QueryRow(
		`SELECT id, created_at FROM groups WHERE id = ?`, id,
	).Scan(&g.ID, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &g, nil
}

func (r *groupRepo) AddMember(m *models.GroupMember) error {
	_, err := r.db.Exec(
		`INSERT INTO group_members (group_id, user_id, joined_at) VALUES (?, ?, ?)`,
		m.GroupID, m.UserID, m.JoinedAt,
	)
	return err
}

func (r *groupRepo) RemoveMember(groupID, userID string) error {
	res, err := r.db.Exec(
		`DELETE FROM group_members WHERE group_id = ? AND user_id = ?`, groupID, userID,
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

func (r *groupRepo) ListMembers(groupID string) ([]models.GroupMember, error) {
	rows, err := r.db.Query(
		`SELECT group_id, user_id, joined_at FROM group_members WHERE group_id = ? ORDER BY joined_at`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.GroupMember
	for rows.Next() {
		var m models.GroupMember
		if err := rows.Scan(&m.GroupID, &m.UserID, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}
