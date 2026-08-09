use crate::errors::Result;
use crate::identity::IdentityKeyPair;
use crate::prekeys::{OneTimePreKeySecret, SignedPreKeySecret};
use crate::sender_keys::SenderKey;
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

pub trait Store: Send + Sync {
    fn load_identity(&self) -> Result<Option<IdentityKeyPair>>;
    fn save_identity(&self, key: &IdentityKeyPair) -> Result<()>;

    fn load_signed_prekey(&self) -> Result<Option<SignedPreKeySecret>>;
    fn save_signed_prekey(&self, key: &SignedPreKeySecret) -> Result<()>;

    fn take_one_time_prekey(&self, id: u64) -> Result<Option<OneTimePreKeySecret>>;
    fn insert_one_time_prekeys(&self, keys: &[OneTimePreKeySecret]) -> Result<()>;

    fn load_ratchet_bytes(&self, peer_id: &str) -> Result<Option<Vec<u8>>>;
    fn save_ratchet_bytes(&self, peer_id: &str, bytes: &[u8]) -> Result<()>;

    fn load_sender_key(&self, group_id: &str, sender_id: &str) -> Result<Option<SenderKey>>;
    fn save_sender_key(
        &self,
        group_id: &str,
        sender_id: &str,
        key: &SenderKey,
    ) -> Result<()>;
}

#[derive(Default, Clone)]
pub struct InMemoryStore {
    inner: Arc<Mutex<InMemoryStoreData>>,
}

#[derive(Default)]
struct InMemoryStoreData {
    identity: Option<IdentityKeyPair>,
    signed_prekey: Option<SignedPreKeySecret>,
    one_time_prekeys: HashMap<u64, OneTimePreKeySecret>,
    ratchet_bytes: HashMap<String, Vec<u8>>,
    sender_keys: HashMap<String, SenderKey>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self::default()
    }
}

impl Store for InMemoryStore {
    fn load_identity(&self) -> Result<Option<IdentityKeyPair>> {
        let guard = self.lock();
        Ok(guard.identity.clone())
    }

    fn save_identity(&self, key: &IdentityKeyPair) -> Result<()> {
        let mut guard = self.lock();
        guard.identity = Some(key.clone());
        Ok(())
    }

    fn load_signed_prekey(&self) -> Result<Option<SignedPreKeySecret>> {
        let guard = self.lock();
        Ok(guard.signed_prekey.clone())
    }

    fn save_signed_prekey(&self, key: &SignedPreKeySecret) -> Result<()> {
        let mut guard = self.lock();
        guard.signed_prekey = Some(key.clone());
        Ok(())
    }

    fn take_one_time_prekey(&self, id: u64) -> Result<Option<OneTimePreKeySecret>> {
        let mut guard = self.lock();
        Ok(guard.one_time_prekeys.remove(&id))
    }

    fn insert_one_time_prekeys(&self, keys: &[OneTimePreKeySecret]) -> Result<()> {
        let mut guard = self.lock();
        for k in keys {
            guard.one_time_prekeys.insert(k.id, k.clone());
        }
        Ok(())
    }

    fn load_ratchet_bytes(&self, peer_id: &str) -> Result<Option<Vec<u8>>> {
        let guard = self.lock();
        Ok(guard.ratchet_bytes.get(peer_id).cloned())
    }

    fn save_ratchet_bytes(&self, peer_id: &str, bytes: &[u8]) -> Result<()> {
        let mut guard = self.lock();
        guard
            .ratchet_bytes
            .insert(peer_id.to_string(), bytes.to_vec());
        Ok(())
    }

    fn load_sender_key(&self, group_id: &str, sender_id: &str) -> Result<Option<SenderKey>> {
        let guard = self.lock();
        let key = format!("{}:{}", group_id, sender_id);
        Ok(guard.sender_keys.get(&key).cloned())
    }

    fn save_sender_key(
        &self,
        group_id: &str,
        sender_id: &str,
        key: &SenderKey,
    ) -> Result<()> {
        let mut guard = self.lock();
        let k = format!("{}:{}", group_id, sender_id);
        guard.sender_keys.insert(k, key.clone());
        Ok(())
    }
}

impl InMemoryStore {
    // A poisoned mutex means some holder panicked mid-update. Recover the inner
    // data (instead of .unwrap() panicking everywhere downstream) — the store
    // stays usable and the panic is surfaced by whoever returned an error.
    fn lock(&self) -> std::sync::MutexGuard<'_, InMemoryStoreData> {
        self.inner
            .lock()
            .unwrap_or_else(|e| e.into_inner())
    }
}
