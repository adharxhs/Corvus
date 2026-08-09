use crate::errors::{CryptoError, Result};
use double_ratchet_2::header::Header;
use double_ratchet_2::ratchet::Ratchet;
use serde::{Deserialize, Serialize};
use std::panic::AssertUnwindSafe;
use x25519_dalek::{PublicKey, StaticSecret};

#[derive(Clone, Serialize, Deserialize)]
pub struct EncryptedMessage {
    pub header_bytes: Vec<u8>,
    pub ciphertext: Vec<u8>,
    pub nonce: [u8; 12],
}

pub struct DoubleRatchetSession {
    inner: Ratchet<StaticSecret>,
}

impl DoubleRatchetSession {
    pub fn init_alice(shared_secret: [u8; 32], bob_dh_pub: [u8; 32]) -> Self {
        let bob_public = PublicKey::from(bob_dh_pub);
        let ratchet = Ratchet::<StaticSecret>::init_alice(shared_secret, bob_public);
        Self { inner: ratchet }
    }

    pub fn init_bob(shared_secret: [u8; 32]) -> (Self, [u8; 32]) {
        let (ratchet, public_key) = Ratchet::<StaticSecret>::init_bob(shared_secret);
        (Self { inner: ratchet }, public_key.to_bytes())
    }

    pub fn encrypt(&mut self, plaintext: &[u8], associated_data: &[u8]) -> EncryptedMessage {
        let (header, ciphertext, nonce) = self.inner.ratchet_encrypt(plaintext, associated_data);
        let header_bytes = header.concat(b"");
        EncryptedMessage {
            header_bytes,
            ciphertext,
            nonce,
        }
    }

    pub fn decrypt(&mut self, msg: &EncryptedMessage, associated_data: &[u8]) -> Result<Vec<u8>> {
        let header = Header::<PublicKey>::from(msg.header_bytes.as_slice());

        let ratchet_clone = self.export_bincode()?;

        // Upstream double-ratchet-2 has no fallible decrypt path: ratchet_decrypt
        // panics (.unwrap()) on authentication failure / invalid header. We run it
        // under catch_unwind against a CLONE and only commit the new state to self
        // on success, so a tampered message neither panics nor corrupts state.
        // Requires the release profile to keep panic = "unwind" (see Cargo.toml) —
        // do not "simplify" this by removing the wrapper assuming decrypt is
        // normally fallible.
        let result = std::panic::catch_unwind(AssertUnwindSafe(|| {
            let mut temp_ratchet = Self::import_bincode(&ratchet_clone).unwrap();
            let decrypted = temp_ratchet.inner.ratchet_decrypt(
                &header,
                &msg.ciphertext,
                &msg.nonce,
                associated_data,
            );
            let new_bytes = temp_ratchet.export_bincode().unwrap();
            (decrypted, new_bytes)
        }));

        match result {
            Ok((decrypted, new_ratchet_bytes)) => {
                let updated = Self::import_bincode(&new_ratchet_bytes)?;
                self.inner = updated.inner;
                Ok(decrypted)
            }
            Err(_) => Err(CryptoError::Ratchet(
                "Decryption failed or header authentication error".to_string(),
            )),
        }
    }

    pub fn export_bincode(&self) -> Result<Vec<u8>> {
        bincode::serialize(&self.inner)
            .map_err(|e| CryptoError::Serialization(format!("Ratchet export failed: {}", e)))
    }

    pub fn import_bincode(bytes: &[u8]) -> Result<Self> {
        let inner: Ratchet<StaticSecret> = bincode::deserialize(bytes)
            .map_err(|e| CryptoError::Serialization(format!("Ratchet import failed: {}", e)))?;
        Ok(Self { inner })
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn double_ratchet_roundtrip() {
        let sk = [7u8; 32];
        let (mut bob, bob_pub) = DoubleRatchetSession::init_bob(sk);
        let mut alice = DoubleRatchetSession::init_alice(sk, bob_pub);

        let ad = b"Protocol Envelope AD";
        let msg1 = alice.encrypt(b"Hello Bob from Alice!", ad);

        let dec1 = bob.decrypt(&msg1, ad).unwrap();
        assert_eq!(dec1, b"Hello Bob from Alice!");

        let msg2 = bob.encrypt(b"Hello Alice back from Bob!", ad);
        let dec2 = alice.decrypt(&msg2, ad).unwrap();
        assert_eq!(dec2, b"Hello Alice back from Bob!");
    }

    #[test]
    fn double_ratchet_out_of_order_recovery() {
        let sk = [11u8; 32];
        let (mut bob, bob_pub) = DoubleRatchetSession::init_bob(sk);
        let mut alice = DoubleRatchetSession::init_alice(sk, bob_pub);

        let ad = b"AD";
        let m1 = alice.encrypt(b"First message", ad);
        let m2 = alice.encrypt(b"Second message", ad);

        let d2 = bob.decrypt(&m2, ad).unwrap();
        assert_eq!(d2, b"Second message");

        let d1 = bob.decrypt(&m1, ad).unwrap();
        assert_eq!(d1, b"First message");
    }

    #[test]
    fn double_ratchet_tampered_ciphertext_rejected() {
        let sk = [13u8; 32];
        let (mut bob, bob_pub) = DoubleRatchetSession::init_bob(sk);
        let mut alice = DoubleRatchetSession::init_alice(sk, bob_pub);

        let ad = b"AD";
        let mut m1 = alice.encrypt(b"Secret data", ad);
        m1.ciphertext[0] ^= 0xFF;

        let bob_state_before = bob.export_bincode().unwrap();
        assert!(bob.decrypt(&m1, ad).is_err());
        let bob_state_after = bob.export_bincode().unwrap();

        assert_eq!(bob_state_before, bob_state_after);
    }
}
