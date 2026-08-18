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
	Rename(id, name string) error
	AddMember(member *models.GroupMember) error
	RemoveMember(groupID, userID string) error
	ListMembers(groupID string) ([]models.GroupMember, error)
	ListByUser(userID string) ([]models.Group, error)
	IsMember(groupID, userID string) (bool, error)
}

type groupRepo struct {
	db *sql.DB
}

func (r *groupRepo) Create(g *models.Group) error {
	_, err := r.db.Exec(
		`INSERT INTO groups (id, name, created_at) VALUES (?, ?, ?)`,
		g.ID, g.Name, g.CreatedAt,
	)
	return err
}

func (r *groupRepo) GetByID(id string) (*models.Group, error) {
	var g models.Group
	err := r.db.QueryRow(
		`SELECT id, name, created_at FROM groups WHERE id = ?`, id,
	).Scan(&g.ID, &g.Name, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &g, nil
}

func (r *groupRepo) Rename(id, name string) error {
	_, err := r.db.Exec(
		`UPDATE groups SET name = ? WHERE id = ?`, name, id,
	)
	return err
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

func (r *groupRepo) ListByUser(userID string) ([]models.Group, error) {
	rows, err := r.db.Query(
		`SELECT g.id, g.name, g.created_at
		 FROM groups g
		 INNER JOIN group_members gm ON g.id = gm.group_id
		 WHERE gm.user_id = ?
		 ORDER BY g.name`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var g models.Group
		if err := rows.Scan(&g.ID, &g.Name, &g.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *groupRepo) IsMember(groupID, userID string) (bool, error) {
	var one int
	err := r.db.QueryRow(
		`SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?`,
		groupID, userID,
	).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
