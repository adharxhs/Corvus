CREATE TABLE group_invites (
    group_id TEXT NOT NULL REFERENCES groups(id),
    user_id TEXT NOT NULL REFERENCES users(id),
    invited_by TEXT NOT NULL REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX idx_group_invites_user ON group_invites(user_id);
