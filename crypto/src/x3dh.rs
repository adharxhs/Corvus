use crate::errors::{CryptoError, Result};
use crate::identity::IdentityKeyPair;
use crate::prekeys::PreKeyBundle;
use crate::random::csprng;
use hkdf::Hkdf;
use sha2::Sha256;
use serde::{Deserialize, Serialize};
use x25519_dalek::{PublicKey, StaticSecret};

const X3DH_INFO: &[u8] = b"Corvus X3DH Protocol v1";

#[derive(Clone, Serialize, Deserialize)]
pub struct X3DHSessionInit {
    pub sender_identity_key: [u8; 32],
    pub ephemeral_key: [u8; 32],
    pub one_time_prekey_id: Option<u64>,
}

pub struct X3DHResult {
    pub shared_secret: [u8; 32],
    pub init_payload: X3DHSessionInit,
}

pub fn initiate(
    alice_identity: &IdentityKeyPair,
    bob_bundle: &PreKeyBundle,
) -> Result<X3DHResult> {
    bob_bundle.verify_signature()?;

    let rng = csprng();
    let ephemeral_secret = StaticSecret::random_from_rng(rng);
    let ephemeral_public = PublicKey::from(&ephemeral_secret);

    let alice_ik_dh = alice_identity.to_x25519_secret();
    let bob_spk_pub = PublicKey::from(bob_bundle.signed_prekey);

    let bob_ik_verifying = ed25519_dalek::VerifyingKey::from_bytes(&bob_bundle.identity_key)
        .map_err(|e| CryptoError::Handshake(format!("Invalid Bob identity key: {}", e)))?;
    let bob_ik_pub = PublicKey::from(bob_ik_verifying.to_montgomery().0);

    // DH1 = X25519(IKA_dh, SPKB)
    let dh1 = alice_ik_dh.diffie_hellman(&bob_spk_pub);
    // DH2 = X25519(EKA, IKB_dh)
    let dh2 = ephemeral_secret.diffie_hellman(&bob_ik_pub);
    // DH3 = X25519(EKA, SPKB)
    let dh3 = ephemeral_secret.diffie_hellman(&bob_spk_pub);

    let mut key_material = Vec::with_capacity(32 * 4);
    key_material.extend_from_slice(dh1.as_bytes());
    key_material.extend_from_slice(dh2.as_bytes());
    key_material.extend_from_slice(dh3.as_bytes());

    if let Some((_opk_id, opk_bytes)) = bob_bundle.one_time_prekey {
        let bob_opk_pub = PublicKey::from(opk_bytes);
        let dh4 = ephemeral_secret.diffie_hellman(&bob_opk_pub);
        key_material.extend_from_slice(dh4.as_bytes());
    }

    // DEVIATION from the standard X3DH spec: the spec prefixes the HKDF IKM
    // with 0xFF * 32 as a domain-separation constant (and uses salt = zeroed).
    // Here IKM is DH1‖DH2‖DH3‖(DH4) directly. This is not exploitable with a
    // single fixed curve (X25519) — the constant only disambiguates across
    // curves — but it means this is NOT full X3DH spec compliance. Do not
    // mistake it for such. Responder side mirrors this exact construction.

    let salt = [0u8; 32];
    let h = Hkdf::<Sha256>::new(Some(&salt), &key_material);
    let mut shared_secret = [0u8; 32];
    h.expand(X3DH_INFO, &mut shared_secret)
        .map_err(|_| CryptoError::Handshake("HKDF expansion failed".to_string()))?;

    let init_payload = X3DHSessionInit {
        sender_identity_key: alice_identity.public_bytes(),
        ephemeral_key: ephemeral_public.to_bytes(),
        one_time_prekey_id: bob_bundle.one_time_prekey.map(|(id, _)| id),
    };

    Ok(X3DHResult {
        shared_secret,
        init_payload,
    })
}

pub fn respond(
    bob_identity: &IdentityKeyPair,
    bob_spk_secret: &StaticSecret,
    bob_opk_secret: Option<&StaticSecret>,
    init: &X3DHSessionInit,
) -> Result<[u8; 32]> {
    let alice_ik_verifying = ed25519_dalek::VerifyingKey::from_bytes(&init.sender_identity_key)
        .map_err(|e| CryptoError::Handshake(format!("Invalid Alice identity key: {}", e)))?;
    let alice_ik_pub = PublicKey::from(alice_ik_verifying.to_montgomery().0);

    let alice_ek_pub = PublicKey::from(init.ephemeral_key);
    let bob_spk_dh = bob_spk_secret;
    let bob_ik_dh = bob_identity.to_x25519_secret();

    // DH1 = X25519(SPKB, IKA)
    let dh1 = bob_spk_dh.diffie_hellman(&alice_ik_pub);
    // DH2 = X25519(IKB, EKA)
    let dh2 = bob_ik_dh.diffie_hellman(&alice_ek_pub);
    // DH3 = X25519(SPKB, EKA)
    let dh3 = bob_spk_dh.diffie_hellman(&alice_ek_pub);

    let mut key_material = Vec::with_capacity(32 * 4);
    key_material.extend_from_slice(dh1.as_bytes());
    key_material.extend_from_slice(dh2.as_bytes());
    key_material.extend_from_slice(dh3.as_bytes());

    if let Some(opk) = bob_opk_secret {
        let dh4 = opk.diffie_hellman(&alice_ek_pub);
        key_material.extend_from_slice(dh4.as_bytes());
    }

    let salt = [0u8; 32];
    let h = Hkdf::<Sha256>::new(Some(&salt), &key_material);
    let mut shared_secret = [0u8; 32];
    h.expand(X3DH_INFO, &mut shared_secret)
        .map_err(|_| CryptoError::Handshake("HKDF expansion failed".to_string()))?;

    Ok(shared_secret)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::prekeys::{OneTimePreKeySecret, SignedPreKeySecret};

    #[test]
    fn x3dh_handshake_with_opk() {
        let alice_identity = IdentityKeyPair::generate();
        let bob_identity = IdentityKeyPair::generate();

        let bob_spk = SignedPreKeySecret::generate(&bob_identity);
        let bob_opk = OneTimePreKeySecret::generate(42);

        let bob_bundle = PreKeyBundle::new(
            bob_identity.public_bytes(),
            &bob_spk,
            Some(&bob_opk),
        );

        let alice_res = initiate(&alice_identity, &bob_bundle).unwrap();

        let bob_secret = respond(
            &bob_identity,
            &bob_spk.secret(),
            Some(&bob_opk.secret()),
            &alice_res.init_payload,
        )
        .unwrap();

        assert_eq!(alice_res.shared_secret, bob_secret);
    }

    #[test]
    fn x3dh_handshake_without_opk() {
        let alice_identity = IdentityKeyPair::generate();
        let bob_identity = IdentityKeyPair::generate();

        let bob_spk = SignedPreKeySecret::generate(&bob_identity);
        let bob_bundle = PreKeyBundle::new(bob_identity.public_bytes(), &bob_spk, None);

        let alice_res = initiate(&alice_identity, &bob_bundle).unwrap();

        let bob_secret = respond(
            &bob_identity,
            &bob_spk.secret(),
            None,
            &alice_res.init_payload,
        )
        .unwrap();

        assert_eq!(alice_res.shared_secret, bob_secret);
    }

    #[test]
    fn x3dh_invalid_signature_aborts() {
        let alice_identity = IdentityKeyPair::generate();
        let bob_identity = IdentityKeyPair::generate();
        let bob_spk = SignedPreKeySecret::generate(&bob_identity);
        let mut tampered = PreKeyBundle::new(bob_identity.public_bytes(), &bob_spk, None);
        tampered.signed_prekey[0] ^= 0xFF;

        assert!(initiate(&alice_identity, &tampered).is_err());
    }
}
