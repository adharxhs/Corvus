CREATE TABLE profile_pictures (
    user_id TEXT PRIMARY KEY REFERENCES users(id),
    ciphertext BLOB NOT NULL,
    nonce BLOB NOT NULL,
    version INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
