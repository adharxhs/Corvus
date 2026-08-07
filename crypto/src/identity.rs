use crate::errors::{CryptoError, Result};
use crate::random::csprng;
use ed25519_dalek::{Signature, SigningKey, VerifyingKey};
use serde::{Deserialize, Serialize};

#[derive(Clone, Serialize, Deserialize)]
pub struct IdentityKeyPair {
    signing_bytes: [u8; 32],
    verifying_bytes: [u8; 32],
}

impl IdentityKeyPair {
    pub fn generate() -> Self {
        let mut rng = csprng();
        let signing = SigningKey::generate(&mut rng);
        let verifying = signing.verifying_key();
        Self {
            signing_bytes: signing.to_bytes(),
            verifying_bytes: verifying.to_bytes(),
        }
    }

    pub fn from_signing_bytes(bytes: &[u8; 32]) -> Result<Self> {
        let signing = SigningKey::from_bytes(bytes);
        let verifying = signing.verifying_key();
        Ok(Self {
            signing_bytes: *bytes,
            verifying_bytes: verifying.to_bytes(),
        })
    }

    pub fn signing_key(&self) -> SigningKey {
        SigningKey::from_bytes(&self.signing_bytes)
    }

    pub fn verifying_key(&self) -> VerifyingKey {
        VerifyingKey::from_bytes(&self.verifying_bytes).expect("Valid verifying key bytes")
    }

    pub fn public_bytes(&self) -> [u8; 32] {
        self.verifying_bytes
    }

    pub fn signing_bytes(&self) -> [u8; 32] {
        self.signing_bytes
    }

    pub fn sign(&self, message: &[u8]) -> Signature {
        use ed25519_dalek::Signer;
        self.signing_key().sign(message)
    }

    pub fn verify(verifying_bytes: &[u8; 32], message: &[u8], signature_bytes: &[u8; 64]) -> Result<()> {
        let verifying = VerifyingKey::from_bytes(verifying_bytes)
            .map_err(|e| CryptoError::Identity(format!("Invalid verifying key: {}", e)))?;
        let signature = Signature::from_bytes(signature_bytes);
        verifying
            .verify_strict(message, &signature)
            .map_err(|e| CryptoError::Identity(format!("Signature verification failed: {}", e)))
    }

    /// Converts the Ed25519 signing key into the corresponding X25519 static secret scalar.
    pub fn to_x25519_secret(&self) -> x25519_dalek::StaticSecret {
        let scalar_bytes = self.signing_key().to_scalar_bytes();
        x25519_dalek::StaticSecret::from(scalar_bytes)
    }

    /// Converts the Ed25519 verifying key into the corresponding X25519 public key.
    pub fn to_x25519_public(&self) -> x25519_dalek::PublicKey {
        let mont = self.verifying_key().to_montgomery();
        x25519_dalek::PublicKey::from(mont.0)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn identity_generate_sign_verify() {
        let id = IdentityKeyPair::generate();
        let msg = b"Corvus Identity Test Payload";
        let sig = id.sign(msg);
        assert!(IdentityKeyPair::verify(&id.public_bytes(), msg, &sig.to_bytes()).is_ok());
        assert!(IdentityKeyPair::verify(&id.public_bytes(), b"tampered", &sig.to_bytes()).is_err());
    }

    #[test]
    fn identity_montgomery_conversion() {
        let id = IdentityKeyPair::generate();
        let secret = id.to_x25519_secret();
        let pubkey = id.to_x25519_public();
        let derived_pub = x25519_dalek::PublicKey::from(&secret);
        assert_eq!(pubkey.as_bytes(), derived_pub.as_bytes());
    }
}
