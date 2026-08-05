CREATE TABLE sender_key_distribution (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL REFERENCES groups(id),
    recipient_id TEXT NOT NULL REFERENCES users(id),
    ciphertext BLOB NOT NULL,
    delivered INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_sender_key_distribution_recipient_id ON sender_key_distribution(recipient_id);
CREATE INDEX idx_sender_key_distribution_group_id ON sender_key_distribution(group_id);
