export interface PresenceSnapshotPayload {
  online: string[];
}

export interface PresencePayload {
  user_id: string;
  status: "online" | "offline";
}

export type PresenceStatus = "online" | "offline" | "unknown";
