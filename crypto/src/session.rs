use crate::prekeys::PreKeyBundle;
use serde::{Deserialize, Serialize};

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct SessionDescriptor {
    pub peer_id: String,
    pub peer_identity_key: [u8; 32],
    pub is_alice: bool,
    pub created_at: u64,
}

impl SessionDescriptor {
    pub fn new(peer_id: String, bundle: &PreKeyBundle, is_alice: bool) -> Self {
        Self {
            peer_id,
            peer_identity_key: bundle.identity_key,
            is_alice,
            created_at: 0,
        }
    }
}
