import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useChat } from "../contexts/ChatContext";
import { useWebSocket } from "../contexts/WebSocketContext";
import { TitleBar, BottomNav } from "../components/NavHeader";
import { SearchBar } from "../components/SearchBar";
import { ConversationListItem } from "../components/ConversationListItem";
import { Fab } from "../components/Fab";
import { BottomSheet } from "../components/BottomSheet";
import { ContactSearchResults } from "../components/ContactSearchResults";
import { EmptyState } from "../components/EmptyState";

export function ChatsPage() {
  const { presence } = useWebSocket();
  const { conversations, contacts, loading, error, offline, selectConversation, startNewChat } = useChat();
  const [query, setQuery] = useState("");
  const [newChatOpen, setNewChatOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [sheetQuery, setSheetQuery] = useState("");
  const navigate = useNavigate();

  const filtered = query
    ? conversations.filter((c) => c.title.toLowerCase().includes(query.toLowerCase()) || c.peerId.toLowerCase().includes(query.toLowerCase()))
    : conversations;

  const contactsSearch = sheetQuery
    ? contacts.filter((c) => c.username?.toLowerCase().includes(sheetQuery.toLowerCase()) || c.id.toLowerCase().includes(sheetQuery.toLowerCase()))
    : contacts;

  async function handleStartNewChat() {
    if (!sheetQuery.trim()) return;
    setSubmitting(true);
    setFormError(null);
    try {
      await startNewChat(sheetQuery.trim());
      setNewChatOpen(false);
      setSheetQuery("");
    } catch (err) {
      setFormError(err instanceof Error ? err.message : "Failed");
    } finally {
      setSubmitting(false);
    }
  }

  function handleOpenContactChat() {
    setNewChatOpen(false);
  }

  return (
    <div className="page chats-page">
      <TitleBar />
      <div className="chats-search-wrapper">
        <SearchBar query={query} onChange={setQuery} placeholder="Search chats" />
      </div>
      {offline && <div className="offline-banner">Offline — showing saved chats</div>}
      {error && !loading && !offline && <p className="error-text page-error">{error}</p>}
      {loading && <p className="muted page-hint">Loading chats…</p>}
      {filtered.length === 0 && !loading && (
        <EmptyState title="No chats yet" body="Tap the pencil to start a new conversation." />
      )}
      {filtered.length > 0 && (
        <div className="conversation-list">
          {filtered.map((conversation) => (
            <ConversationListItem
              key={conversation.peerId}
              conversation={conversation}
              presence={conversation.kind === "dm" ? presence.get(conversation.peerId) : undefined}
              onClick={() => {
                selectConversation(conversation.peerId);
                navigate(`/chat/${conversation.kind}/${conversation.peerId}`);
              }}
            />
          ))}
        </div>
      )}
      <Fab label="New chat" onClick={() => setNewChatOpen(true)} />
      <BottomSheet open={newChatOpen} title="New chat" onClose={() => { setNewChatOpen(false); setSheetQuery(""); setFormError(null); }}>
        <div className="sheet-search-section">
          <SearchBar query={sheetQuery} onChange={setSheetQuery} placeholder="Enter username or search contacts" />
          {formError && <p className="error-text">{formError}</p>}
          <button type="button" className="primary-button" disabled={submitting || !sheetQuery.trim()} onClick={() => void handleStartNewChat()}>
            {submitting ? "Sending…" : "Send request"}
          </button>
        </div>
        {contactsSearch.length > 0 && (
          <ContactSearchResults contacts={contactsSearch} conversations={conversations} onClickContact={handleOpenContactChat} />
        )}
      </BottomSheet>
      <BottomNav />
    </div>
  );
}
