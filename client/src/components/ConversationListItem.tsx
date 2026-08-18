import type { Conversation } from "../types/chat";
import { formatConversationTime, shortId } from "../utils/format";
import { AppAvatar } from "./AppAvatar";

interface ConversationListItemProps {
  conversation: Conversation;
  active?: boolean;
  presence?: "online" | "offline" | "unknown";
  onClick: () => void;
}

export function ConversationListItem({ conversation, active, presence, onClick }: ConversationListItemProps) {
  const title = conversation.title || (conversation.kind === "group" ? `Group ${shortId(conversation.peerId)}` : `User ${shortId(conversation.peerId)}`);
  return (
    <button type="button" className={`conversation-item ${active ? "active" : ""}`} onClick={onClick}>
      <AppAvatar
        label={title}
        userId={conversation.kind === "dm" ? conversation.peerId : undefined}
        presence={conversation.kind === "dm" ? presence : undefined}
      />
      <span className="conversation-main">
        <span className="conversation-row">
          <strong>{title}</strong>
          <time>{formatConversationTime(conversation.lastTimestamp)}</time>
        </span>
        <span className="conversation-preview">
          {conversation.lastMessage && <span>{conversation.lastMessage}</span>}
        </span>
      </span>
      {conversation.unreadCount > 0 && <span className="unread-badge">{conversation.unreadCount}</span>}
    </button>
  );
}
