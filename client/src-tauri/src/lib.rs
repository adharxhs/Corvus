mod profile_key;

use profile_key::{AppState, ProfileKeyStore};
use std::sync::Mutex;
use tauri::Manager;

#[tauri::command]
fn ensure_profile_key(state: tauri::State<AppState>) -> bool {
    state.profile_key.lock().expect("profile key mutex poisoned").is_ready()
}

#[tauri::command]
fn encrypt_profile_picture(
    state: tauri::State<AppState>,
    bytes: Vec<u8>,
) -> Result<profile_key::EncryptResult, String> {
    let store = state.profile_key.lock().expect("profile key mutex poisoned");
    profile_key::encrypt(&store, &bytes)
}

#[tauri::command]
fn decrypt_profile_picture(
    state: tauri::State<AppState>,
    ciphertext_b64: String,
    nonce_b64: String,
) -> Result<Vec<u8>, String> {
    let store = state.profile_key.lock().expect("profile key mutex poisoned");
    profile_key::decrypt(&store, &ciphertext_b64, &nonce_b64)
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .setup(|app| {
            let dir = app.path().app_data_dir()?;
            std::fs::create_dir_all(&dir)?;
            let store = ProfileKeyStore::load_or_create(&dir.join("profile_key.bin"));
            app.manage(AppState {
                profile_key: Mutex::new(store),
            });
            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            ensure_profile_key,
            encrypt_profile_picture,
            decrypt_profile_picture
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
