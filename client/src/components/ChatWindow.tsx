import type { DirectMessage } from "../types/message";

interface ChatWindowProps {
  messages: DirectMessage[];
}

export function ChatWindow({ messages }: ChatWindowProps) {
  return (
    <section className="panel chat-window">
      <div className="message-list">
        {messages.length === 0 ? <p className="muted">No messages yet</p> : null}
        {messages.map((message, index) => (
          <div key={`${message.recipient_id}-${index}`} className="message-bubble">
            <span className="muted">{message.recipient_id}</span>
            <p>{message.content}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
