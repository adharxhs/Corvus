import type { DirectMessage, GroupMessage, SenderKeyDistribution } from "../types/message";
import type { PresencePayload, PresenceSnapshotPayload } from "../types/presence";
import type { ProtocolErrorPayload, ProtocolEnvelope } from "../types/protocol";

export interface GroupProfilePictureUpdatedPayload {
  group_id: string;
  version: number;
}

export interface ChatRequestUpdatedPayload {
  requester_id: string;
  recipient_id: string;
  status: "pending" | "accepted" | "rejected";
}

export interface MemberJoinedPayload {
  group_id: string;
  user_id: string;
  username: string;
}

export type ClientPayload = DirectMessage | GroupMessage | SenderKeyDistribution | ProtocolEnvelope<string>;

export type ServerPayload =
  | DirectMessage
  | GroupMessage
  | SenderKeyDistribution
  | PresenceSnapshotPayload
  | PresencePayload
  | GroupProfilePictureUpdatedPayload
  | ChatRequestUpdatedPayload
  | MemberJoinedPayload
  | ProtocolErrorPayload;

export type ServerEnvelope = ProtocolEnvelope<ServerPayload>;
