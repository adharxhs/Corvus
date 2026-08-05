package repository

import (
	"database/sql"

	"server/models"
)

// SenderKeyRepository queues encrypted Sender Key distribution messages.
// Ciphertext is stored and returned as opaque bytes.
type SenderKeyRepository interface {
	Enqueue(dist *models.SenderKeyDistribution) error
	FetchByRecipient(recipientID string) ([]models.SenderKeyDistribution, error)
	MarkDelivered(id string) error
	DeleteDelivered() error
}

type senderKeyRepo struct {
	db *sql.DB
}

func (r *senderKeyRepo) Enqueue(d *models.SenderKeyDistribution) error {
	_, err := r.db.Exec(
		`INSERT INTO sender_key_distribution (id, group_id, recipient_id, ciphertext, delivered)
		 VALUES (?, ?, ?, ?, ?)`,
		d.ID, d.GroupID, d.RecipientID, d.Ciphertext, d.Delivered,
	)
	return err
}

func (r *senderKeyRepo) FetchByRecipient(recipientID string) ([]models.SenderKeyDistribution, error) {
	rows, err := r.db.Query(
		`SELECT id, group_id, recipient_id, ciphertext, delivered
		 FROM sender_key_distribution WHERE recipient_id = ? AND delivered = 0 ORDER BY id`,
		recipientID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var distributions []models.SenderKeyDistribution
	for rows.Next() {
		var d models.SenderKeyDistribution
		if err := rows.Scan(&d.ID, &d.GroupID, &d.RecipientID, &d.Ciphertext, &d.Delivered); err != nil {
			return nil, err
		}
		distributions = append(distributions, d)
	}
	return distributions, rows.Err()
}

func (r *senderKeyRepo) MarkDelivered(id string) error {
	res, err := r.db.Exec(`UPDATE sender_key_distribution SET delivered = 1 WHERE id = ?`, id)
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

func (r *senderKeyRepo) DeleteDelivered() error {
	_, err := r.db.Exec(`DELETE FROM sender_key_distribution WHERE delivered = 1`)
	return err
}
