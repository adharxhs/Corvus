import type { ProtocolEnvelope } from "../types/protocol";

export const CURRENT_VERSION = 1;

export function createEnvelope<TPayload>(type: string, payload: TPayload): ProtocolEnvelope<TPayload> {
  return {
    version: CURRENT_VERSION,
    type,
    payload,
  };
}

export function encodeEnvelope<TPayload>(envelope: ProtocolEnvelope<TPayload>): string {
  return JSON.stringify(envelope);
}

export function decodeEnvelope(data: string): ProtocolEnvelope {
  const parsed: unknown = JSON.parse(data);
  if (!isProtocolEnvelope(parsed)) {
    throw new Error("malformed envelope");
  }
  return parsed;
}

function isProtocolEnvelope(value: unknown): value is ProtocolEnvelope {
  if (typeof value !== "object" || value === null) {
    return false;
  }
  const record = value as Record<string, unknown>;
  return record.version === CURRENT_VERSION && typeof record.type === "string";
}
