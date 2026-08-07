use crate::errors::{CryptoError, Result};
use crate::identity::IdentityKeyPair;
use crate::random::csprng;
use serde::{Deserialize, Serialize};
use x25519_dalek::{PublicKey, StaticSecret};

mod serde_64_bytes {
    use serde::{de::Error, Deserialize, Deserializer, Serializer};
    use serde_bytes::ByteBuf;

    pub fn serialize<S>(bytes: &[u8; 64], serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        serde_bytes::serialize(bytes.as_slice(), serializer)
    }

    pub fn deserialize<'de, D>(deserializer: D) -> Result<[u8; 64], D::Error>
    where
        D: Deserializer<'de>,
    {
        let buf = ByteBuf::deserialize(deserializer)?;
        buf.as_slice()
            .try_into()
            .map_err(|_| Error::custom("expected 64 bytes for signature"))
    }
}

#[derive(Clone, Serialize, Deserialize)]
pub struct SignedPreKeySecret {
    secret_bytes: [u8; 32],
    public_bytes: [u8; 32],
    #[serde(with = "serde_64_bytes")]
    signature_bytes: [u8; 64],
}

impl SignedPreKeySecret {
    pub fn generate(identity: &IdentityKeyPair) -> Self {
        let rng = csprng();
        let secret = StaticSecret::random_from_rng(rng);
        let public = PublicKey::from(&secret);
        let pub_bytes = public.to_bytes();
        let sig = identity.sign(&pub_bytes);
        Self {
            secret_bytes: secret.to_bytes(),
            public_bytes: pub_bytes,
            signature_bytes: sig.to_bytes(),
        }
    }

    pub fn secret(&self) -> StaticSecret {
        StaticSecret::from(self.secret_bytes)
    }

    pub fn public_bytes(&self) -> [u8; 32] {
        self.public_bytes
    }

    pub fn signature_bytes(&self) -> [u8; 64] {
        self.signature_bytes
    }
}

#[derive(Clone, Serialize, Deserialize)]
pub struct OneTimePreKeySecret {
    pub id: u64,
    secret_bytes: [u8; 32],
    public_bytes: [u8; 32],
}

impl OneTimePreKeySecret {
    pub fn generate(id: u64) -> Self {
        let rng = csprng();
        let secret = StaticSecret::random_from_rng(rng);
        let public = PublicKey::from(&secret);
        Self {
            id,
            secret_bytes: secret.to_bytes(),
            public_bytes: public.to_bytes(),
        }
    }

    pub fn generate_batch(start_id: u64, count: usize) -> Vec<Self> {
        (0..count)
            .map(|i| Self::generate(start_id + i as u64))
            .collect()
    }

    pub fn secret(&self) -> StaticSecret {
        StaticSecret::from(self.secret_bytes)
    }

    pub fn public_bytes(&self) -> [u8; 32] {
        self.public_bytes
    }
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub struct PreKeyBundle {
    pub identity_key: [u8; 32],
    pub signed_prekey: [u8; 32],
    #[serde(with = "serde_64_bytes")]
    pub signed_prekey_signature: [u8; 64],
    pub one_time_prekey: Option<(u64, [u8; 32])>,
}

impl PreKeyBundle {
    pub fn new(
        identity_key: [u8; 32],
        signed_prekey: &SignedPreKeySecret,
        one_time_prekey: Option<&OneTimePreKeySecret>,
    ) -> Self {
        Self {
            identity_key,
            signed_prekey: signed_prekey.public_bytes(),
            signed_prekey_signature: signed_prekey.signature_bytes(),
            one_time_prekey: one_time_prekey.map(|opk| (opk.id, opk.public_bytes())),
        }
    }

    pub fn verify_signature(&self) -> Result<()> {
        IdentityKeyPair::verify(
            &self.identity_key,
            &self.signed_prekey,
            &self.signed_prekey_signature,
        )
        .map_err(|e| CryptoError::Prekey(format!("Prekey bundle signature invalid: {}", e)))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn prekey_generation_and_bundle_verify() {
        let identity = IdentityKeyPair::generate();
        let spk = SignedPreKeySecret::generate(&identity);
        let opk_batch = OneTimePreKeySecret::generate_batch(1, 5);
        assert_eq!(opk_batch.len(), 5);

        let bundle = PreKeyBundle::new(identity.public_bytes(), &spk, Some(&opk_batch[0]));
        assert!(bundle.verify_signature().is_ok());

        let mut tampered = bundle.clone();
        tampered.signed_prekey[0] ^= 0xFF;
        assert!(tampered.verify_signature().is_err());
    }
}
