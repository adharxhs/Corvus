mod commands;
mod crypto_manager;

use crypto_manager::CryptoManager;

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .manage(CryptoManager::new())
        .invoke_handler(tauri::generate_handler![
            commands::init_crypto,
            commands::has_identity,
            commands::get_identity_key,
            commands::build_prekey_bundle,
            commands::start_session,
            commands::accept_session,
            commands::complete_alice_session,
            commands::encrypt_message,
            commands::decrypt_message,
            commands::create_group_sender_key,
            commands::process_sender_key_distribution,
            commands::encrypt_group_message,
            commands::decrypt_group_message,
            commands::save_crypto_state,
            commands::load_crypto_state,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
