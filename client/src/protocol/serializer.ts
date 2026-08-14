import type { DirectMessage, GroupMessage, SenderKeyDistribution } from "../types/message";

export function serializeDirectMessage(payload: DirectMessage): string {
  return JSON.stringify(payload);
}

export function serializeGroupMessage(payload: GroupMessage): string {
  return JSON.stringify(payload);
}

export function serializeSenderKeyDistribution(payload: SenderKeyDistribution): string {
  return JSON.stringify(payload);
}

export function serializeProfilePictureUpdated(version: number): string {
  return JSON.stringify({ version });
}
