use aes_gcm::aead::{Aead, KeyInit};
use aes_gcm::{Aes256Gcm, Nonce};
use crate::errors::{CryptoError, Result};
use crate::random::csprng;
use hkdf::Hkdf;
use rand_core::RngCore;
use serde::{Deserialize, Serialize};
use sha2::Sha256;

const SENDER_KEY_INFO: &[u8] = b"Corvus SenderKey Chain Step";

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct SenderKey {
    pub group_id: String,
    pub sender_id: String,
    pub key_id: u32,
    pub chain_key: [u8; 32],
    pub iteration: u32,
}

impl SenderKey {
    pub fn generate(group_id: &str, sender_id: &str, key_id: u32) -> Self {
        let mut rng = csprng();
        let mut chain_key = [0u8; 32];
        rng.fill_bytes(&mut chain_key);
        Self {
            group_id: group_id.to_string(),
            sender_id: sender_id.to_string(),
            key_id,
            chain_key,
            iteration: 0,
        }
    }

    pub fn step(&mut self) -> Result<([u8; 32], [u8; 12])> {
        let h = Hkdf::<Sha256>::new(Some(&self.chain_key), SENDER_KEY_INFO);
        let mut okm = [0u8; 76];
        h.expand(b"SenderKey Derive", &mut okm)
            .map_err(|_| CryptoError::SenderKey("HKDF step failed".to_string()))?;

        let (next_chain, rest) = okm.split_at(32);
        let (message_key, nonce_bytes) = rest.split_at(32);

        let mut mk = [0u8; 32];
        mk.copy_from_slice(message_key);
        let mut nc = [0u8; 12];
        nc.copy_from_slice(nonce_bytes);

        self.chain_key.copy_from_slice(next_chain);
        self.iteration += 1;

        Ok((mk, nc))
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct SenderKeyDistributionMessage {
    pub group_id: String,
    pub sender_id: String,
    pub key_id: u32,
    pub chain_key: [u8; 32],
    pub iteration: u32,
}

impl SenderKeyDistributionMessage {
    pub fn from_sender_key(key: &SenderKey) -> Self {
        Self {
            group_id: key.group_id.clone(),
            sender_id: key.sender_id.clone(),
            key_id: key.key_id,
            chain_key: key.chain_key,
            iteration: key.iteration,
        }
    }

    pub fn to_sender_key(&self) -> SenderKey {
        SenderKey {
            group_id: self.group_id.clone(),
            sender_id: self.sender_id.clone(),
            key_id: self.key_id,
            chain_key: self.chain_key,
            iteration: self.iteration,
        }
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct GroupEncryptedMessage {
    pub group_id: String,
    pub sender_id: String,
    pub key_id: u32,
    pub iteration: u32,
    pub ciphertext: Vec<u8>,
    pub nonce: [u8; 12],
}

pub fn encrypt_group_message(
    sender_key: &mut SenderKey,
    plaintext: &[u8],
    associated_data: &[u8],
) -> Result<GroupEncryptedMessage> {
    let iteration = sender_key.iteration;
    let key_id = sender_key.key_id;
    let group_id = sender_key.group_id.clone();
    let sender_id = sender_key.sender_id.clone();

    let (mk, nonce_bytes) = sender_key.step()?;

    let cipher = Aes256Gcm::new_from_slice(&mk)
        .map_err(|e| CryptoError::SenderKey(format!("Cipher init error: {}", e)))?;
    let nonce = Nonce::from_slice(&nonce_bytes);

    let ciphertext = cipher
        .encrypt(nonce, plaintext)
        .map_err(|e| CryptoError::SenderKey(format!("Group encryption failed: {}", e)))?;

    let _ = associated_data;

    Ok(GroupEncryptedMessage {
        group_id,
        sender_id,
        key_id,
        iteration,
        ciphertext,
        nonce: nonce_bytes,
    })
}

pub fn decrypt_group_message(
    sender_key: &mut SenderKey,
    msg: &GroupEncryptedMessage,
    associated_data: &[u8],
) -> Result<Vec<u8>> {
    if sender_key.key_id != msg.key_id {
        return Err(CryptoError::SenderKey(format!(
            "Key ID mismatch: expected {}, got {}",
            sender_key.key_id, msg.key_id
        )));
    }

    while sender_key.iteration < msg.iteration {
        let _ = sender_key.step()?;
    }

    if sender_key.iteration != msg.iteration {
        return Err(CryptoError::SenderKey(
            "Sender key iteration out of sync".to_string(),
        ));
    }

    let (mk, nonce_bytes) = sender_key.step()?;

    let cipher = Aes256Gcm::new_from_slice(&mk)
        .map_err(|e| CryptoError::SenderKey(format!("Cipher init error: {}", e)))?;
    let nonce = Nonce::from_slice(&nonce_bytes);

    let plaintext = cipher
        .decrypt(nonce, msg.ciphertext.as_slice())
        .map_err(|e| CryptoError::SenderKey(format!("Group decryption failed: {}", e)))?;

    let _ = associated_data;

    Ok(plaintext)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn sender_key_encrypt_decrypt_flow() {
        let mut alice_key = SenderKey::generate("group-1", "alice", 1);
        let dist = SenderKeyDistributionMessage::from_sender_key(&alice_key);

        let mut bob_received_key = dist.to_sender_key();

        let ad = b"group context ad";
        let plaintext = b"Hello Group!";

        let enc_msg = encrypt_group_message(&mut alice_key, plaintext, ad).unwrap();
        let dec_plaintext = decrypt_group_message(&mut bob_received_key, &enc_msg, ad).unwrap();

        assert_eq!(plaintext.to_vec(), dec_plaintext);
    }

    #[test]
    fn sender_key_rotation() {
        let mut alice_v1 = SenderKey::generate("group-1", "alice", 1);
        let _enc1 = encrypt_group_message(&mut alice_v1, b"msg 1", b"").unwrap();

        let mut alice_v2 = SenderKey::generate("group-1", "alice", 2);
        assert_ne!(alice_v1.key_id, alice_v2.key_id);

        let mut bob_key_v1 = SenderKeyDistributionMessage::from_sender_key(&alice_v1).to_sender_key();
        let enc2 = encrypt_group_message(&mut alice_v2, b"msg 2 after rotation", b"").unwrap();

        assert!(decrypt_group_message(&mut bob_key_v1, &enc2, b"").is_err());
    }
}
