import { invoke } from "@tauri-apps/api/core";

export interface PrekeyBundleData {
  identity_key: string;
  signed_prekey: string;
  signed_prekey_signature: string;
  one_time_prekey?: { id: number; public_key: string };
}

export interface X3DHInitPayload {
  sender_identity_key: string;
  ephemeral_key: string;
  one_time_prekey_id?: number;
}

export interface EncryptedPayload {
  header: string;
  ciphertext: string;
  nonce: string;
}

export interface GroupEncryptedPayload {
  group_id: string;
  sender_id: string;
  key_id: number;
  iteration: number;
  ciphertext: string;
  nonce: string;
}

export interface SenderKeyDistributionData {
  group_id: string;
  sender_id: string;
  key_id: number;
  chain_key: string;
  iteration: number;
}

export interface AcceptSessionResult {
  public_key: string;
}

export async function initCrypto(): Promise<PrekeyBundleData> {
  return invoke<PrekeyBundleData>("init_crypto");
}

export async function hasIdentity(): Promise<boolean> {
  return invoke<boolean>("has_identity");
}

export async function getIdentityKey(): Promise<string> {
  return invoke<string>("get_identity_key");
}

export async function buildPrekeyBundle(): Promise<PrekeyBundleData> {
  return invoke<PrekeyBundleData>("build_prekey_bundle");
}

export async function startSession(
  peerId: string,
  identityKey: string,
  signedPrekey: string,
  signedPrekeySignature: string,
  oneTimePrekeyId?: number,
  oneTimePrekeyKey?: string,
): Promise<X3DHInitPayload> {
  return invoke<X3DHInitPayload>("start_session", {
    peerId,
    identityKey,
    signedPrekey,
    signedPrekeySignature,
    oneTimePrekeyId,
    oneTimePrekeyKey,
  });
}

export async function acceptSession(
  senderIdentityKey: string,
  ephemeralKey: string,
  oneTimePrekeyId?: number,
): Promise<AcceptSessionResult> {
  return invoke<AcceptSessionResult>("accept_session", {
    senderIdentityKey,
    ephemeralKey,
    oneTimePrekeyId,
  });
}

export async function completeAliceSession(
  peerId: string,
  bobPublicKey: string,
): Promise<void> {
  return invoke<void>("complete_alice_session", {
    peerId,
    bobPublicKey,
  });
}

export async function encryptMessage(
  peerId: string,
  plaintext: string,
): Promise<EncryptedPayload> {
  return invoke<EncryptedPayload>("encrypt_message", {
    peerId,
    plaintext,
  });
}

export async function decryptMessage(
  peerId: string,
  header: string,
  ciphertext: string,
  nonce: string,
): Promise<string> {
  return invoke<string>("decrypt_message", {
    peerId,
    header,
    ciphertext,
    nonce,
  });
}

export async function createGroupSenderId(
  groupId: string,
  senderId: string,
): Promise<SenderKeyDistributionData> {
  return invoke<SenderKeyDistributionData>("create_group_sender_key", {
    groupId,
    senderId,
  });
}

export async function processSenderKeyDistribution(
  groupId: string,
  senderId: string,
  keyId: number,
  chainKey: string,
  iteration: number,
): Promise<void> {
  return invoke<void>("process_sender_key_distribution", {
    groupId,
    senderId,
    keyId,
    chainKey,
    iteration,
  });
}

export async function encryptGroupMessage(
  groupId: string,
  senderId: string,
  plaintext: string,
): Promise<GroupEncryptedPayload> {
  return invoke<GroupEncryptedPayload>("encrypt_group_message", {
    groupId,
    senderId,
    plaintext,
  });
}

export async function decryptGroupMessage(
  groupId: string,
  senderId: string,
  ciphertext: string,
  nonce: string,
  keyId: number,
  iteration: number,
): Promise<string> {
  return invoke<string>("decrypt_group_message", {
    groupId,
    senderId,
    ciphertext,
    nonce,
    keyId,
    iteration,
  });
}

export async function saveCryptoState(): Promise<void> {
  return invoke<void>("save_crypto_state");
}

export async function loadCryptoState(data: string): Promise<void> {
  return invoke<void>("load_crypto_state", { data });
}
