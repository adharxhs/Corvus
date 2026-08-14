import type { ChatMessage } from "../types/chat";
import { formatMessageTime } from "../utils/format";

interface MessageBubbleProps {
  message: ChatMessage;
}

function StatusIcon({ status }: { status?: "pending" | "sent" }) {
  if (!status || status === "sent") {
    return (
      <svg className="msg-status-icon msg-status-sent" viewBox="0 0 16 16" width="14" height="14">
        <path d="M1.5 8.5l3 3 7-7" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
    );
  }
  return (
    <svg className="msg-status-icon msg-status-pending" viewBox="0 0 16 16" width="14" height="14">
      <circle cx="8" cy="8" r="6.5" fill="none" stroke="currentColor" strokeWidth="1.2" />
      <path d="M8 4.5v4l2.5 1.5" fill="none" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export function MessageBubble({ message }: MessageBubbleProps) {
  return (
    <div className={`message-row ${message.direction}`}>
      <div className="message-bubble">
        <p>{message.content}</p>
        <time>
          {formatMessageTime(message.timestamp)}
          {message.direction === "out" && <StatusIcon status={message.status} />}
        </time>
      </div>
    </div>
  );
}
