CREATE TABLE relationships (
    requester_id TEXT NOT NULL REFERENCES users(id),
    recipient_id TEXT NOT NULL REFERENCES users(id),
    status TEXT NOT NULL CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (requester_id, recipient_id)
);

CREATE INDEX idx_relationships_recipient ON relationships(recipient_id);
CREATE INDEX idx_relationships_requester ON relationships(requester_id);
