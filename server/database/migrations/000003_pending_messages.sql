CREATE TABLE pending_messages (
    id TEXT PRIMARY KEY,
    recipient_id TEXT NOT NULL REFERENCES users(id),
    ciphertext BLOB NOT NULL,
    delivered INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_pending_messages_recipient_id ON pending_messages(recipient_id);
