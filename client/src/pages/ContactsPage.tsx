import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useChat } from "../contexts/ChatContext";
import { TitleBar, BottomNav } from "../components/NavHeader";
import { SearchBar } from "../components/SearchBar";
import { ContactList } from "../components/ContactList";
import { Fab } from "../components/Fab";
import { BottomSheet } from "../components/BottomSheet";
import { NewChatForm } from "../components/NewChatForm";
import { EmptyState } from "../components/EmptyState";
import { useAuth } from "../contexts/AuthContext";

export function ContactsPage() {
  const { user } = useAuth();
  const { contacts, conversations, pendingIncoming, acceptRequest, rejectRequest, startNewChat, selectConversation, usernameFor } = useChat();
  const navigate = useNavigate();
  const [query, setQuery] = useState("");
  const [newChatOpen, setNewChatOpen] = useState(false);

  const filtered = query
    ? contacts.filter((c) => c.username?.toLowerCase().includes(query.toLowerCase()) || c.id.toLowerCase().includes(query.toLowerCase()))
    : contacts;

  return (
    <div className="page contacts-page">
      <TitleBar />
      <div className="chats-search-wrapper">
        <SearchBar query={query} onChange={setQuery} placeholder="Search contacts" />
      </div>
      {pendingIncoming.length > 0 && (
        <div className="panel pending-requests">
          <h3>Pending requests</h3>
          {pendingIncoming.map((request) => {
            const otherId = request.recipient_id === user?.id ? request.requester_id : request.recipient_id;
            return (
              <div key={request.requester_id} className="pending-row">
                <span>{usernameFor(otherId)}</span>
                <span className="button-group">
                  <button type="button" onClick={() => void acceptRequest(request.requester_id)}>Accept</button>
                  <button type="button" className="secondary" onClick={() => void rejectRequest(request.requester_id)}>Reject</button>
                </span>
              </div>
            );
          })}
        </div>
      )}
      {filtered.length === 0 && pendingIncoming.length === 0 && (
        <EmptyState title="No contacts yet" body="Add someone to start a conversation." />
      )}
      <ContactList
        contacts={filtered}
        conversations={conversations}
        onClickContact={(contact) => {
          selectConversation(contact.id);
          navigate(`/chat/dm/${contact.id}`);
        }}
        onSendRequest={() => setNewChatOpen(true)}
      />
      <Fab label="Add contact" onClick={() => setNewChatOpen(true)} />
      <BottomSheet open={newChatOpen} title="Add contact" onClose={() => setNewChatOpen(false)}>
        <NewChatForm onSubmit={async (username) => { await startNewChat(username); setNewChatOpen(false); }} />
      </BottomSheet>
      <BottomNav />
    </div>
  );
}
