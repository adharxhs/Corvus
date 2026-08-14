import { Fragment, useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useChat } from "../contexts/ChatContext";
import { PageHeader } from "../components/PageHeader";
import { MessageBubble } from "../components/MessageBubble";
import { MessageInput } from "../components/MessageInput";
import { BottomSheet } from "../components/BottomSheet";
import { GroupProfilePicturePicker } from "../components/GroupProfilePicturePicker";
import { formatMessageDate } from "../utils/format";

export function ChatScreen() {
  const { kind, id } = useParams<{ kind: string; id: string }>();
  const { contacts, groups, conversations, messagesByPeer, selectConversation, sendDirectMessage, sendGroupMessage } = useChat();
  const navigate = useNavigate();
  const bottomRef = useRef<HTMLDivElement>(null);
  const [groupSettingsOpen, setGroupSettingsOpen] = useState(false);

  const peerId = id ?? "";
  const conversationKind = kind === "group" ? "group" : "dm";
  const messages = messagesByPeer[peerId] ?? [];
  const contact = contacts.find((c) => c.id === peerId);
  const group = groups.find((g) => g.id === peerId);
  const conversation = conversations.find((c) => c.peerId === peerId);
  const title = conversationKind === "group" ? peerId.slice(0, 20) : (contact?.username ?? conversation?.title ?? peerId.slice(0, 20));

  useEffect(() => { selectConversation(peerId); }, [peerId, selectConversation]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages.length]);

  function handleSend(text: string) {
    if (conversationKind === "group") {
      sendGroupMessage(group ?? { id: peerId, members: [] }, text);
    } else {
      const contactInfo = contact ?? { id: peerId, username: title, relationship: "accepted" as const, presence: "unknown" as const };
      sendDirectMessage(contactInfo, text);
    }
  }

  return (
    <div className="page chat-page">
      <PageHeader onBack={() => navigate(-1)} title={title}>
        {conversationKind === "group" && (
          <button
            type="button"
            className="icon-button"
            onClick={() => setGroupSettingsOpen(true)}
            aria-label="Group settings"
          >
            <svg viewBox="0 0 24 24" width="24" height="24" stroke="currentColor" fill="none" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="1" />
              <circle cx="19" cy="12" r="1" />
              <circle cx="5" cy="12" r="1" />
            </svg>
          </button>
        )}
      </PageHeader>
      <div className="chat-messages">
        {messages.length === 0 && <p className="muted chat-hint">Start the conversation…</p>}
        {messages.reduce<Array<{ key: string; showDate: boolean; message: (typeof messages)[number] }>>(
          (acc, message) => {
            const day = new Date(message.timestamp).toDateString();
            const showDate = acc.length === 0 || day !== new Date(acc[acc.length - 1].message.timestamp).toDateString();
            acc.push({ key: message.id, showDate, message });
            return acc;
          },
          []
        ).map(({ key, showDate, message }) => (
          <Fragment key={key}>
            {showDate && <div className="message-date">{formatMessageDate(message.timestamp)}</div>}
            <MessageBubble message={message} />
          </Fragment>
        ))}
        <div ref={bottomRef} />
      </div>
      <MessageInput disabled={false} onSend={handleSend} />
      <BottomSheet open={groupSettingsOpen} title="Group Settings" onClose={() => setGroupSettingsOpen(false)}>
        <GroupProfilePicturePicker groupId={peerId} />
      </BottomSheet>
    </div>
  );
}
