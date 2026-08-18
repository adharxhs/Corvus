use base64::{engine::general_purpose::STANDARD as BASE64, Engine};
use crypto::double_ratchet::EncryptedMessage;
use crypto::prekeys::PreKeyBundle;
use crypto::sender_keys::{GroupEncryptedMessage, SenderKeyDistributionMessage};
use tauri::State;

use crate::crypto_manager::CryptoManager;

fn b64_encode(data: &[u8]) -> String {
    BASE64.encode(data)
}

fn b64_decode(s: &str) -> Result<Vec<u8>, String> {
    BASE64.decode(s).map_err(|e| e.to_string())
}

#[tauri::command]
pub fn init_crypto(state: State<'_, CryptoManager>) -> Result<serde_json::Value, String> {
    let bundle = state.generate_identity()?;
    Ok(serde_json::json!({
        "identity_key": b64_encode(&bundle.identity_key),
        "signed_prekey": b64_encode(&bundle.signed_prekey),
        "signed_prekey_signature": b64_encode(&bundle.signed_prekey_signature),
        "one_time_prekey": bundle.one_time_prekey.map(|(id, bytes)| {
            serde_json::json!({ "id": id, "public_key": b64_encode(&bytes) })
        }),
    }))
}

#[tauri::command]
pub fn has_identity(state: State<'_, CryptoManager>) -> Result<bool, String> {
    Ok(state.has_identity())
}

#[tauri::command]
pub fn get_identity_key(state: State<'_, CryptoManager>) -> Result<String, String> {
    let key = state.get_identity_public_key()?;
    Ok(b64_encode(&key))
}

#[tauri::command]
pub fn build_prekey_bundle(state: State<'_, CryptoManager>) -> Result<serde_json::Value, String> {
    let bundle = state.build_prekey_bundle()?;
    Ok(serde_json::json!({
        "identity_key": b64_encode(&bundle.identity_key),
        "signed_prekey": b64_encode(&bundle.signed_prekey),
        "signed_prekey_signature": b64_encode(&bundle.signed_prekey_signature),
    }))
}

#[tauri::command]
pub fn start_session(
    state: State<'_, CryptoManager>,
    peer_id: String,
    identity_key: String,
    signed_prekey: String,
    signed_prekey_signature: String,
    one_time_prekey_id: Option<u64>,
    one_time_prekey_key: Option<String>,
) -> Result<serde_json::Value, String> {
    let ik = b64_decode(&identity_key)?;
    let spk = b64_decode(&signed_prekey)?;
    let sig = b64_decode(&signed_prekey_signature)?;

    let ik_arr: [u8; 32] = ik.try_into().map_err(|_| "identity_key must be 32 bytes")?;
    let spk_arr: [u8; 32] = spk.try_into().map_err(|_| "signed_prekey must be 32 bytes")?;
    let sig_arr: [u8; 64] = sig.try_into().map_err(|_| "signed_prekey_signature must be 64 bytes")?;

    let opk = match (one_time_prekey_id, one_time_prekey_key) {
        (Some(id), Some(key)) => {
            let k = b64_decode(&key)?;
            let k_arr: [u8; 32] = k.try_into().map_err(|_| "one_time_prekey must be 32 bytes")?;
            Some((id, k_arr))
        }
        _ => None,
    };

    let bundle = PreKeyBundle {
        identity_key: ik_arr,
        signed_prekey: spk_arr,
        signed_prekey_signature: sig_arr,
        one_time_prekey: opk,
    };

    let init = state.initiate_session(&peer_id, &bundle)?;

    Ok(serde_json::json!({
        "sender_identity_key": b64_encode(&init.sender_identity_key),
        "ephemeral_key": b64_encode(&init.ephemeral_key),
        "one_time_prekey_id": init.one_time_prekey_id,
    }))
}

#[tauri::command]
pub fn accept_session(
    state: State<'_, CryptoManager>,
    sender_identity_key: String,
    ephemeral_key: String,
    one_time_prekey_id: Option<u64>,
) -> Result<serde_json::Value, String> {
    let ik = b64_decode(&sender_identity_key)?;
    let ek = b64_decode(&ephemeral_key)?;

    let ik_arr: [u8; 32] = ik.try_into().map_err(|_| "sender_identity_key must be 32 bytes")?;
    let ek_arr: [u8; 32] = ek.try_into().map_err(|_| "ephemeral_key must be 32 bytes")?;

    let init = crypto::x3dh::X3DHSessionInit {
        sender_identity_key: ik_arr,
        ephemeral_key: ek_arr,
        one_time_prekey_id,
    };

    let (_shared_secret, bob_pub) = state.respond_to_session(&init)?;

    Ok(serde_json::json!({
        "public_key": b64_encode(&bob_pub),
    }))
}

#[tauri::command]
pub fn complete_alice_session(
    state: State<'_, CryptoManager>,
    peer_id: String,
    bob_public_key: String,
) -> Result<(), String> {
    let bk = b64_decode(&bob_public_key)?;
    let bk_arr: [u8; 32] = bk.try_into().map_err(|_| "bob_public_key must be 32 bytes")?;

    state.complete_session_as_alice(&peer_id, bk_arr)
}

#[tauri::command]
pub fn encrypt_message(
    state: State<'_, CryptoManager>,
    peer_id: String,
    plaintext: String,
) -> Result<serde_json::Value, String> {
    let msg = state.encrypt_direct(&peer_id, plaintext.as_bytes())?;
    Ok(encrypted_message_to_json(&msg))
}

#[tauri::command]
pub fn decrypt_message(
    state: State<'_, CryptoManager>,
    peer_id: String,
    header: String,
    ciphertext: String,
    nonce: String,
) -> Result<String, String> {
    let msg = EncryptedMessage {
        header_bytes: b64_decode(&header)?,
        ciphertext: b64_decode(&ciphertext)?,
        nonce: b64_decode(&nonce)?
            .try_into()
            .map_err(|_| "nonce must be 12 bytes")?,
    };
    let plaintext = state.decrypt_direct(&peer_id, &msg)?;
    String::from_utf8(plaintext).map_err(|e| e.to_string())
}

#[tauri::command]
pub fn create_group_sender_key(
    state: State<'_, CryptoManager>,
    group_id: String,
    sender_id: String,
) -> Result<serde_json::Value, String> {
    let key = state.create_sender_key(&group_id, &sender_id)?;
    let dist = SenderKeyDistributionMessage::from_sender_key(&key);
    Ok(serde_json::json!({
        "group_id": dist.group_id,
        "sender_id": dist.sender_id,
        "key_id": dist.key_id,
        "chain_key": b64_encode(&dist.chain_key),
        "iteration": dist.iteration,
    }))
}

#[tauri::command]
pub fn process_sender_key_distribution(
    state: State<'_, CryptoManager>,
    group_id: String,
    sender_id: String,
    key_id: u32,
    chain_key: String,
    iteration: u32,
) -> Result<(), String> {
    let ck = b64_decode(&chain_key)?;
    let ck_arr: [u8; 32] = ck.try_into().map_err(|_| "chain_key must be 32 bytes")?;

    let dist = SenderKeyDistributionMessage {
        group_id,
        sender_id,
        key_id,
        chain_key: ck_arr,
        iteration,
    };

    let key = dist.to_sender_key();
    state.save_sender_key_from_distribution(&key)
}

#[tauri::command]
pub fn encrypt_group_message(
    state: State<'_, CryptoManager>,
    group_id: String,
    sender_id: String,
    plaintext: String,
) -> Result<serde_json::Value, String> {
    let msg = state.encrypt_group(&group_id, &sender_id, plaintext.as_bytes())?;
    Ok(group_encrypted_message_to_json(&msg))
}

#[tauri::command]
pub fn decrypt_group_message(
    state: State<'_, CryptoManager>,
    group_id: String,
    sender_id: String,
    ciphertext: String,
    nonce: String,
    key_id: u32,
    iteration: u32,
) -> Result<String, String> {
    let nonce_arr: [u8; 12] = b64_decode(&nonce)?
        .try_into()
        .map_err(|_| "nonce must be 12 bytes")?;

    let msg = GroupEncryptedMessage {
        group_id: group_id.clone(),
        sender_id: sender_id.clone(),
        key_id,
        iteration,
        ciphertext: b64_decode(&ciphertext)?,
        nonce: nonce_arr,
    };
    let plaintext = state.decrypt_group(&group_id, &sender_id, &msg)?;
    String::from_utf8(plaintext).map_err(|e| e.to_string())
}

#[tauri::command]
pub fn save_crypto_state(state: State<'_, CryptoManager>) -> Result<(), String> {
    let _bytes = state.export_store()?;
    // Store in app data directory via Tauri would be ideal,
    // but for now we return the bytes and let TS handle persistence
    // through localStorage. A future improvement would use tauri::api::path.
    Ok(())
}

#[tauri::command]
pub fn load_crypto_state(state: State<'_, CryptoManager>, data: String) -> Result<(), String> {
    let bytes = data.as_bytes();
    state.import_store(bytes)
}

fn encrypted_message_to_json(msg: &EncryptedMessage) -> serde_json::Value {
    serde_json::json!({
        "header": b64_encode(&msg.header_bytes),
        "ciphertext": b64_encode(&msg.ciphertext),
        "nonce": b64_encode(&msg.nonce),
    })
}

fn group_encrypted_message_to_json(msg: &GroupEncryptedMessage) -> serde_json::Value {
    serde_json::json!({
        "group_id": msg.group_id,
        "sender_id": msg.sender_id,
        "key_id": msg.key_id,
        "iteration": msg.iteration,
        "ciphertext": b64_encode(&msg.ciphertext),
        "nonce": b64_encode(&msg.nonce),
    })
}
