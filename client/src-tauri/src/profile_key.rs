use aes_gcm::aead::{Aead, KeyInit};
use aes_gcm::{Aes256Gcm, Nonce};
use base64::{engine::general_purpose::STANDARD as B64, Engine};
use rand::rngs::OsRng;
use rand::RngCore;
use serde::Serialize;
use std::path::Path;
use std::sync::Mutex;

pub struct ProfileKeyStore {
    key: [u8; 32],
}

impl ProfileKeyStore {
    pub fn load_or_create(path: &Path) -> Self {
        if let Ok(raw) = std::fs::read(path) {
            if raw.len() == 32 {
                let mut key = [0u8; 32];
                key.copy_from_slice(&raw);
                return Self { key };
            }
        }
        let mut key = [0u8; 32];
        OsRng.fill_bytes(&mut key);
        let _ = std::fs::write(path, key);
        Self { key }
    }

    pub fn is_ready(&self) -> bool {
        true
    }
}

#[derive(Serialize)]
pub struct EncryptResult {
    pub ciphertext: String,
    pub nonce: String,
}

pub fn encrypt(store: &ProfileKeyStore, plaintext: &[u8]) -> Result<EncryptResult, String> {
    let cipher = Aes256Gcm::new_from_slice(&store.key).map_err(|e| e.to_string())?;
    let mut nonce_bytes = [0u8; 12];
    OsRng.fill_bytes(&mut nonce_bytes);
    let nonce = Nonce::from_slice(&nonce_bytes);
    let ciphertext = cipher.encrypt(nonce, plaintext).map_err(|e| e.to_string())?;
    Ok(EncryptResult {
        ciphertext: B64.encode(ciphertext),
        nonce: B64.encode(nonce_bytes),
    })
}

pub fn decrypt(store: &ProfileKeyStore, ciphertext_b64: &str, nonce_b64: &str) -> Result<Vec<u8>, String> {
    let cipher = Aes256Gcm::new_from_slice(&store.key).map_err(|e| e.to_string())?;
    let ciphertext = B64.decode(ciphertext_b64).map_err(|e| e.to_string())?;
    let nonce_bytes = B64.decode(nonce_b64).map_err(|e| e.to_string())?;
    let nonce = Nonce::from_slice(&nonce_bytes);
    cipher.decrypt(nonce, ciphertext.as_ref()).map_err(|e| e.to_string())
}

pub struct AppState {
    pub profile_key: Mutex<ProfileKeyStore>,
}
