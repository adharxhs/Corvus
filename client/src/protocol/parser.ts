import type { ProtocolEnvelope } from "../types/protocol";
import type { ServerPayload } from "./payloads";

export function parseServerEnvelope(data: unknown): ProtocolEnvelope<ServerPayload> {
  if (typeof data !== "string") {
    throw new Error("expected string");
  }
  const parsed: unknown = JSON.parse(data);
  if (!isValidServerEnvelope(parsed)) {
    throw new Error("malformed server envelope");
  }
  return parsed;
}

function isValidServerEnvelope(value: unknown): value is ProtocolEnvelope<ServerPayload> {
  if (typeof value !== "object" || value === null) {
    return false;
  }
  const record = value as Record<string, unknown>;
  return record.version === 1 && typeof record.type === "string" && "payload" in record;
}
