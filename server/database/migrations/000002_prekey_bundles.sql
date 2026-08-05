CREATE TABLE prekey_bundles (
    user_id TEXT NOT NULL REFERENCES users(id),
    identity_key BLOB NOT NULL,
    signed_prekey BLOB NOT NULL,
    signed_prekey_signature BLOB NOT NULL,
    one_time_prekey BLOB,
    PRIMARY KEY (user_id)
);
