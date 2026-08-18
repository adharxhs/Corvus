import { Fragment, useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useChat } from "../contexts/ChatContext";
import { useWebSocket } from "../contexts/WebSocketContext";
import { PageHeader } from "../components/PageHeader";
import { AppAvatar } from "../components/AppAvatar";
import { MessageBubble } from "../components/MessageBubble";
import { MessageInput } from "../components/MessageInput";
import { BottomSheet } from "../components/BottomSheet";
import { GroupAvatar } from "../components/GroupAvatar";
import { GroupProfilePicturePicker } from "../components/GroupProfilePicturePicker";
import { formatMessageDate } from "../utils/format";
import { lookupUserByUsername } from "../services/users";
import { listGroupMembers } from "../services/groups";

export function ChatScreen() {
  const { kind, id } = useParams<{ kind: string; id: string }>();
  const { contacts, groups, conversations, messagesByPeer, selectConversation, clearSelection, sendDirectMessage, sendGroupMessage, inviteToGroup, renameGroup, fetchGroupInfo, usernameFor, resolveUsername, groupPictureVersion } = useChat();
  const { presence } = useWebSocket();
  const navigate = useNavigate();
  const bottomRef = useRef<HTMLDivElement>(null);
  const [groupSettingsOpen, setGroupSettingsOpen] = useState(false);
  const [groupMembers, setGroupMembers] = useState<{ user_id: string; joined_at: number }[]>([]);
  const [memberUsername, setMemberUsername] = useState("");
  const [inviteBusy, setInviteBusy] = useState(false);
  const [inviteResult, setInviteResult] = useState<string | null>(null);
  const [renameValue, setRenameValue] = useState("");
  const [renameBusy, setRenameBusy] = useState(false);
  const [renameResult, setRenameResult] = useState<string | null>(null);

  const peerId = id ?? "";
  const conversationKind = kind === "group" ? "group" : "dm";
  const messages = messagesByPeer[peerId] ?? [];
  const contact = contacts.find((c) => c.id === peerId);
  const group = groups.find((g) => g.id === peerId);
  const conversation = conversations.find((c) => c.peerId === peerId);
  const title = conversationKind === "group"
    ? (group?.name || `Group ${peerId.slice(0, 8)}`)
    : (contact?.username ?? conversation?.title ?? `User ${peerId.slice(0, 8)}`);

  useEffect(() => {
    selectConversation(peerId);
    return () => clearSelection();
  }, [peerId, selectConversation, clearSelection]);

  useEffect(() => {
    if (conversationKind !== "group") return;
    void fetchGroupInfo(peerId).catch(() => {});
  }, [peerId, conversationKind]);

  useEffect(() => {
    if (!groupSettingsOpen || conversationKind !== "group") return;
    let active = true;
    listGroupMembers(peerId).then((members) => {
      if (!active) return;
      setGroupMembers(members);
      for (const m of members) resolveUsername(m.user_id);
    }).catch(() => {});
    return () => { active = false; };
  }, [groupSettingsOpen, conversationKind, peerId, resolveUsername]);

  useEffect(() => {
    if (groupSettingsOpen && group) {
      setRenameValue(group.name || "");
      setRenameResult(null);
    }
  }, [groupSettingsOpen, group]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages.length]);

  function handleSend(text: string) {
    if (conversationKind === "group") {
      sendGroupMessage(group ?? { id: peerId, name: "", members: [] }, text);
    } else {
      const contactInfo = contact ?? { id: peerId, username: title, relationship: "accepted" as const, presence: "unknown" as const };
      sendDirectMessage(contactInfo, text);
    }
  }

  async function handleInviteMember() {
    const username = memberUsername.trim();
    if (!username) return;
    setInviteBusy(true);
    setInviteResult(null);
    try {
      const lookup = await lookupUserByUsername(username);
      await inviteToGroup(peerId, lookup.id);
      setInviteResult("Invitation sent");
      setMemberUsername("");
      const members = await listGroupMembers(peerId);
      setGroupMembers(members);
    } catch (err) {
      setInviteResult(err instanceof Error ? err.message : "Failed to invite");
    } finally {
      setInviteBusy(false);
    }
  }

  async function handleRename() {
    const name = renameValue.trim();
    if (!name || name === (group?.name || "")) {
      setRenameResult(null);
      return;
    }
    setRenameBusy(true);
    setRenameResult(null);
    try {
      await renameGroup(peerId, name);
      setRenameResult("Renamed");
    } catch (err) {
      setRenameResult(err instanceof Error ? err.message : "Failed to rename");
    } finally {
      setRenameBusy(false);
    }
  }

  const avatarNode = conversationKind === "group"
    ? <GroupAvatar name={group?.name || `Group ${peerId.slice(0, 8)}`} groupId={peerId} size={32} pictureVersion={groupPictureVersion[peerId]} />
    : <AppAvatar label={title} userId={peerId} size={32} />;

  return (
    <div className="page chat-page">
      <PageHeader onBack={() => navigate(-1)} title={title} beforeTitle={avatarNode}>
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
            {message.system ? (
              <div className="message-system">{message.content}</div>
            ) : (
              <MessageBubble
                message={message}
                isGroup={conversationKind === "group"}
                senderName={message.senderId ? usernameFor(message.senderId) : undefined}
              />
            )}
          </Fragment>
        ))}
        <div ref={bottomRef} />
      </div>
      <MessageInput disabled={false} onSend={handleSend} />
      <BottomSheet open={groupSettingsOpen} title="Group Settings" onClose={() => { setGroupSettingsOpen(false); setRenameResult(null); setInviteResult(null); setMemberUsername(""); }}>
        <GroupProfilePicturePicker groupId={peerId} pictureVersion={groupPictureVersion[peerId]} />
        <section className="panel">
          <h3>Group Name</h3>
          <div className="sheet-form" style={{ marginBottom: 0 }}>
            <input
              type="text"
              placeholder="Group name"
              maxLength={60}
              value={renameValue}
              onChange={(e) => setRenameValue(e.target.value)}
            />
            <button
              type="button"
              className="primary-button"
              disabled={renameBusy || !renameValue.trim() || renameValue.trim() === (group?.name || "")}
              onClick={() => void handleRename()}
            >
              {renameBusy ? "Saving…" : "Save"}
            </button>
            {renameResult && <p className={renameResult === "Renamed" ? "success-text" : "error-text"}>{renameResult}</p>}
          </div>
        </section>
        <section className="panel">
          <h3>Members ({groupMembers.length})</h3>
          <div className="contact-list">
            {groupMembers.map((m) => {
              const memberName = usernameFor(m.user_id);
              const memberPresence = presence.get(m.user_id);
              return (
                <div key={m.user_id} className="contact-list-item">
                  <AppAvatar label={memberName} userId={m.user_id} size={36} presence={memberPresence} />
                  <span className="contact-main">
                    <span className="contact-name">{memberName}</span>
                  </span>
                </div>
              );
            })}
          </div>
        </section>
        <section className="panel">
          <h3>Add Member</h3>
          <div className="sheet-form">
            <input
              type="text"
              placeholder="Username"
              value={memberUsername}
              onChange={(e) => setMemberUsername(e.target.value)}
              onKeyDown={(e) => { if (e.key === "Enter") void handleInviteMember(); }}
            />
            <button type="button" className="primary-button" disabled={inviteBusy || !memberUsername.trim()} onClick={() => void handleInviteMember()}>
              {inviteBusy ? "Sending…" : "Invite"}
            </button>
            {inviteResult && <p className={inviteResult.startsWith("Invitation") ? "success-text" : "error-text"}>{inviteResult}</p>}
          </div>
        </section>
      </BottomSheet>
    </div>
  );
}
