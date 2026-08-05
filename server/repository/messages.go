package repository

import (
	"database/sql"

	"server/models"
)

// MessageRepository queues encrypted pending messages. Ciphertext is stored
// and returned as opaque bytes.
type MessageRepository interface {
	Insert(msg *models.PendingMessage) error
	FetchByRecipient(recipientID string) ([]models.PendingMessage, error)
	MarkDelivered(id string) error
	DeleteDelivered() error
}

type messageRepo struct {
	db *sql.DB
}

func (r *messageRepo) Insert(m *models.PendingMessage) error {
	_, err := r.db.Exec(
		`INSERT INTO pending_messages (id, recipient_id, ciphertext, delivered) VALUES (?, ?, ?, ?)`,
		m.ID, m.RecipientID, m.Ciphertext, m.Delivered,
	)
	return err
}

func (r *messageRepo) FetchByRecipient(recipientID string) ([]models.PendingMessage, error) {
	rows, err := r.db.Query(
		`SELECT id, recipient_id, ciphertext, delivered
		 FROM pending_messages WHERE recipient_id = ? AND delivered = 0 ORDER BY id`,
		recipientID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.PendingMessage
	for rows.Next() {
		var m models.PendingMessage
		if err := rows.Scan(&m.ID, &m.RecipientID, &m.Ciphertext, &m.Delivered); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (r *messageRepo) MarkDelivered(id string) error {
	res, err := r.db.Exec(`UPDATE pending_messages SET delivered = 1 WHERE id = ?`, id)
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

func (r *messageRepo) DeleteDelivered() error {
	_, err := r.db.Exec(`DELETE FROM pending_messages WHERE delivered = 1`)
	return err
}
