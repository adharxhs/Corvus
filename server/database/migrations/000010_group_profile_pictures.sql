CREATE TABLE group_profile_pictures (
    group_id TEXT PRIMARY KEY REFERENCES groups(id),
    ciphertext BLOB NOT NULL,
    nonce BLOB NOT NULL,
    version INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
