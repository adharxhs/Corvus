use crypto::double_ratchet::DoubleRatchetSession;
use crypto::identity::IdentityKeyPair;
use crypto::prekeys::{OneTimePreKeySecret, PreKeyBundle, SignedPreKeySecret};
use crypto::sender_keys::SenderKey;
use crypto::storage::InMemoryStore;
use crypto::x3dh;
use crypto::Store;
use std::collections::HashMap;
use std::sync::Mutex;

pub struct CryptoManager {
    pub store: InMemoryStore,
    sessions: Mutex<HashMap<String, DoubleRatchetSession>>,
    sender_keys: Mutex<HashMap<String, SenderKey>>,
    pending_shared_secrets: Mutex<HashMap<String, [u8; 32]>>,
}

impl CryptoManager {
    pub fn new() -> Self {
        Self {
            store: InMemoryStore::new(),
            sessions: Mutex::new(HashMap::new()),
            sender_keys: Mutex::new(HashMap::new()),
            pending_shared_secrets: Mutex::new(HashMap::new()),
        }
    }

    pub fn has_identity(&self) -> bool {
        self.store.load_identity().ok().flatten().is_some()
    }

    pub fn generate_identity(&self) -> Result<PreKeyBundle, String> {
        let identity = IdentityKeyPair::generate();
        let spk = SignedPreKeySecret::generate(&identity);
        let opks = OneTimePreKeySecret::generate_batch(0, 10);

        let bundle = PreKeyBundle::new(identity.public_bytes(), &spk, Some(&opks[0]));

        self.store
            .save_identity(&identity)
            .map_err(|e| e.to_string())?;
        self.store
            .save_signed_prekey(&spk)
            .map_err(|e| e.to_string())?;
        self.store
            .insert_one_time_prekeys(&opks)
            .map_err(|e| e.to_string())?;

        Ok(bundle)
    }

    pub fn get_identity_public_key(&self) -> Result<[u8; 32], String> {
        self.store
            .load_identity()
            .map_err(|e| e.to_string())?
            .map(|id| id.public_bytes())
            .ok_or_else(|| "No identity key found".to_string())
    }

    pub fn build_prekey_bundle(&self) -> Result<PreKeyBundle, String> {
        let identity = self
            .store
            .load_identity()
            .map_err(|e| e.to_string())?
            .ok_or("No identity key")?;
        let spk = self
            .store
            .load_signed_prekey()
            .map_err(|e| e.to_string())?
            .ok_or("No signed prekey")?;

        let bundle = PreKeyBundle::new(identity.public_bytes(), &spk, None);
        Ok(bundle)
    }

    pub fn initiate_session(
        &self,
        peer_id: &str,
        bundle: &PreKeyBundle,
    ) -> Result<x3dh::X3DHSessionInit, String> {
        let identity = self
            .store
            .load_identity()
            .map_err(|e| e.to_string())?
            .ok_or("No identity key")?;

        let result = x3dh::initiate(&identity, bundle).map_err(|e| e.to_string())?;

        // Store the shared secret temporarily - session will be completed
        // when we receive Bob's ratchet public key via complete_session_as_alice.
        let mut secrets = self.pending_shared_secrets.lock().map_err(|e| e.to_string())?;
        secrets.insert(peer_id.to_string(), result.shared_secret);

        Ok(result.init_payload)
    }

    pub fn complete_session_as_alice(
        &self,
        peer_id: &str,
        bob_dh_pub: [u8; 32],
    ) -> Result<(), String> {
        let shared_secret = {
            let mut secrets = self.pending_shared_secrets.lock().map_err(|e| e.to_string())?;
            secrets.remove(peer_id).ok_or("No pending session for this peer")?
        };

        let session = DoubleRatchetSession::init_alice(shared_secret, bob_dh_pub);
        let mut sessions = self.sessions.lock().map_err(|e| e.to_string())?;
        sessions.insert(peer_id.to_string(), session);
        Ok(())
    }

    pub fn respond_to_session(
        &self,
        init: &x3dh::X3DHSessionInit,
    ) -> Result<([u8; 32], [u8; 32]), String> {
        let identity = self
            .store
            .load_identity()
            .map_err(|e| e.to_string())?
            .ok_or("No identity key")?;
        let spk = self
            .store
            .load_signed_prekey()
            .map_err(|e| e.to_string())?
            .ok_or("No signed prekey")?;

        let opk_secret = if let Some(opk_id) = init.one_time_prekey_id {
            self.store
                .take_one_time_prekey(opk_id)
                .map_err(|e| e.to_string())?
                .map(|k| k.secret())
        } else {
            None
        };

        let shared_secret =
            x3dh::respond(&identity, &spk.secret(), opk_secret.as_ref(), init)
                .map_err(|e| e.to_string())?;

        let (session, bob_pub) = DoubleRatchetSession::init_bob(shared_secret);

        // Store the session for this peer (we'll need the sender identity key to identify)
        // For now, we store with a temporary key and the frontend will need to tell us who this is for
        // Actually, we should store by a unique identifier. Let's use a hash of the sender identity key.
        let peer_key = crypto::util::to_hex(&init.sender_identity_key);
        let mut sessions = self.sessions.lock().map_err(|e| e.to_string())?;
        sessions.insert(peer_key, session);

        Ok((shared_secret, bob_pub))
    }

    pub fn encrypt_direct(
        &self,
        peer_id: &str,
        plaintext: &[u8],
    ) -> Result<crypto::double_ratchet::EncryptedMessage, String> {
        let mut sessions = self.sessions.lock().map_err(|e| e.to_string())?;
        let session = sessions
            .get_mut(peer_id)
            .ok_or_else(|| format!("No session for peer {}", peer_id))?;
        Ok(session.encrypt(plaintext, b""))
    }

    pub fn decrypt_direct(
        &self,
        peer_id: &str,
        msg: &crypto::double_ratchet::EncryptedMessage,
    ) -> Result<Vec<u8>, String> {
        let mut sessions = self.sessions.lock().map_err(|e| e.to_string())?;
        let session = sessions
            .get_mut(peer_id)
            .ok_or_else(|| format!("No session for peer {}", peer_id))?;
        session.decrypt(msg, b"").map_err(|e| e.to_string())
    }

    pub fn create_sender_key(
        &self,
        group_id: &str,
        sender_id: &str,
    ) -> Result<crypto::sender_keys::SenderKey, String> {
        let key = SenderKey::generate(group_id, sender_id, 1);
        let mut sender_keys = self.sender_keys.lock().map_err(|e| e.to_string())?;
        let k = format!("{}:{}", group_id, sender_id);
        sender_keys.insert(k, key.clone());
        Ok(key)
    }

    pub fn save_sender_key_from_distribution(
        &self,
        key: &SenderKey,
    ) -> Result<(), String> {
        let mut sender_keys = self.sender_keys.lock().map_err(|e| e.to_string())?;
        let k = format!("{}:{}", key.group_id, key.sender_id);
        sender_keys.insert(k, key.clone());
        Ok(())
    }

    pub fn encrypt_group(
        &self,
        group_id: &str,
        sender_id: &str,
        plaintext: &[u8],
    ) -> Result<crypto::sender_keys::GroupEncryptedMessage, String> {
        let mut sender_keys = self.sender_keys.lock().map_err(|e| e.to_string())?;
        let k = format!("{}:{}", group_id, sender_id);
        let key = sender_keys
            .get_mut(&k)
            .ok_or_else(|| format!("No sender key for {}:{}", group_id, sender_id))?;
        crypto::sender_keys::encrypt_group_message(key, plaintext, b"")
            .map_err(|e| e.to_string())
    }

    pub fn decrypt_group(
        &self,
        group_id: &str,
        sender_id: &str,
        msg: &crypto::sender_keys::GroupEncryptedMessage,
    ) -> Result<Vec<u8>, String> {
        let mut sender_keys = self.sender_keys.lock().map_err(|e| e.to_string())?;
        let k = format!("{}:{}", group_id, sender_id);
        let key = sender_keys
            .get_mut(&k)
            .ok_or_else(|| format!("No sender key for {}:{}", group_id, sender_id))?;
        crypto::sender_keys::decrypt_group_message(key, msg, b"")
            .map_err(|e| e.to_string())
    }

    pub fn export_store(&self) -> Result<Vec<u8>, String> {
        let identity = self
            .store
            .load_identity()
            .map_err(|e| e.to_string())?
            .ok_or("No identity")?;
        let spk = self
            .store
            .load_signed_prekey()
            .map_err(|e| e.to_string())?
            .ok_or("No signed prekey")?;

        let data = serde_json::json!({
            "identity_signing": identity.signing_bytes(),
            "identity_public": identity.public_bytes(),
            "signed_prekey": spk,
        });

        Ok(data.to_string().into_bytes())
    }

    pub fn import_store(&self, bytes: &[u8]) -> Result<(), String> {
        let data: serde_json::Value =
            serde_json::from_slice(bytes).map_err(|e| e.to_string())?;

        let signing: [u8; 32] = serde_json::from_value(
            data["identity_signing"].clone(),
        )
        .map_err(|e| e.to_string())?;
        let identity =
            IdentityKeyPair::from_signing_bytes(&signing).map_err(|e| e.to_string())?;

        self.store
            .save_identity(&identity)
            .map_err(|e| e.to_string())?;

        if let Some(spk_val) = data.get("signed_prekey") {
            let spk: SignedPreKeySecret =
                serde_json::from_value(spk_val.clone()).map_err(|e| e.to_string())?;
            self.store
                .save_signed_prekey(&spk)
                .map_err(|e| e.to_string())?;
        }

        Ok(())
    }
}

impl Default for CryptoManager {
    fn default() -> Self {
        Self::new()
    }
}
