import type { Conversation } from "../types/chat";
import type { Contact } from "../types/user";
import { shortId } from "../utils/format";

interface ContactListProps {
  contacts: Contact[];
  conversations: Conversation[];
  onClickContact: (contact: Contact) => void;
  onSendRequest: (username: string) => void;
}

export function ContactList({ contacts, conversations, onClickContact, onSendRequest }: ContactListProps) {
  return (
    <div className="panel contact-list">
      <div className="contact-list-header">
        <h2>Contacts</h2>
        <button type="button" className="icon-button" onClick={() => onSendRequest("")} aria-label="Add contact">
          <svg viewBox="0 0 24 24" width="20" height="20" stroke="currentColor" fill="none" strokeWidth="2">
            <path d="M12 5v14M5 12h14" />
          </svg>
        </button>
      </div>
      {contacts.length === 0 && <p className="sheet-hint">No contacts yet</p>}
      {contacts.map((contact) => {
        const conversation = conversations.find((c) => c.peerId === contact.id);
        const displayName = contact.username || `User ${shortId(contact.id)}`;
        return (
          <button key={contact.id} type="button" className="contact-list-item" onClick={() => onClickContact(contact)}>
            <span className="contact-avatar">{displayName.slice(0, 1).toUpperCase()}</span>
            <span className="contact-main">
              <span className="contact-name">{displayName}</span>
              {conversation && <span className="contact-last">{conversation.lastMessage}</span>}
            </span>
            <span className={`presence-dot presence-${contact.presence}`} />
          </button>
        );
      })}
    </div>
  );
}
