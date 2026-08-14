import type { DirectMessage, GroupMessage, SenderKeyDistribution } from "../types/message";
import type { PresencePayload, PresenceSnapshotPayload } from "../types/presence";
import type { ProtocolErrorPayload, ProtocolEnvelope } from "../types/protocol";

export interface GroupProfilePictureUpdatedPayload {
  group_id: string;
  version: number;
}

export type ClientPayload = DirectMessage | GroupMessage | SenderKeyDistribution | ProtocolEnvelope<string>;

export type ServerPayload =
  | DirectMessage
  | GroupMessage
  | SenderKeyDistribution
  | PresenceSnapshotPayload
  | PresencePayload
  | GroupProfilePictureUpdatedPayload
  | ProtocolErrorPayload;

export type ServerEnvelope = ProtocolEnvelope<ServerPayload>;
