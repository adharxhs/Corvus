CREATE TABLE group_profile_pictures (
    group_id TEXT PRIMARY KEY REFERENCES groups(id),
    image_data BLOB NOT NULL,
    version INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
