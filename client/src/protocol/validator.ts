import type { ProtocolEnvelope } from "../types/protocol";

const SUPPORTED_SERVER_TYPES = new Set([
  "message",
  "group_message",
  "sender_key_distribution",
  "profile_picture_updated",
  "group_profile_picture_updated",
  "chat_request_updated",
  "member_joined",
  "presence_snapshot",
  "presence",
  "error",
]);

const SUPPORTED_CLIENT_TYPES = new Set([
  "message",
  "group_message",
  "sender_key_distribution",
  "profile_picture_updated",
]);

export function validateClientEnvelope(envelope: ProtocolEnvelope): boolean {
  return SUPPORTED_CLIENT_TYPES.has(envelope.type);
}

export function validateServerEnvelopeType(type: string): boolean {
  return SUPPORTED_SERVER_TYPES.has(type);
}
