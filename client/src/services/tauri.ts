import type { EncryptedProfilePicture } from "../types/profilePicture";
import { invoke } from "@tauri-apps/api/core";

export async function ensureProfileKey(): Promise<boolean> {
  return invoke<boolean>("ensure_profile_key");
}

export async function encryptProfilePicture(bytes: Uint8Array): Promise<EncryptedProfilePicture> {
  const result = await invoke<{ ciphertext: string; nonce: string }>("encrypt_profile_picture", {
    bytes,
  });
  return result;
}

export async function decryptProfilePicture(ciphertextB64: string, nonceB64: string): Promise<Uint8Array> {
  return invoke<Uint8Array>("decrypt_profile_picture", {
    ciphertextB64,
    nonceB64,
  });
}
