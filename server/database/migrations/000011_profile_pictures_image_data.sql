-- Convert profile picture storage from client-side encrypted (ciphertext + nonce)
-- to the new image_data format. The old AES-GCM ciphertext cannot be migrated
-- (client encryption format changed), so the tables are rebuilt empty.
DROP TABLE IF EXISTS group_profile_pictures;
DROP TABLE IF EXISTS profile_pictures;

CREATE TABLE profile_pictures (
    user_id TEXT PRIMARY KEY REFERENCES users(id),
    image_data BLOB NOT NULL,
    version INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE group_profile_pictures (
    group_id TEXT PRIMARY KEY REFERENCES groups(id),
    image_data BLOB NOT NULL,
    version INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
