export type ChatKind = "dm" | "group";

export interface Conversation {
  peerId: string;
  kind: ChatKind;
  title: string;
  lastMessage: string;
  lastTimestamp: number;
  unreadCount: number;
}

export interface ChatMessage {
  id: string;
  peerId: string;
  content: string;
  direction: "in" | "out";
  timestamp: number;
  status?: "pending" | "sent";
}
