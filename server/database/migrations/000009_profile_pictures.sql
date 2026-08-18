CREATE TABLE profile_pictures (
    user_id TEXT PRIMARY KEY REFERENCES users(id),
    image_data BLOB NOT NULL,
    version INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
