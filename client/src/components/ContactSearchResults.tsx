import type { Conversation } from "../types/chat";
import type { Contact } from "../types/user";
import { formatConversationTime, shortId } from "../utils/format";

interface ContactSearchResultsProps {
  contacts: Contact[];
  conversations: Conversation[];
  onClickContact: (contact: Contact) => void;
}

export function ContactSearchResults({ contacts, conversations, onClickContact }: ContactSearchResultsProps) {
  if (contacts.length === 0) return <p className="sheet-hint">No contacts found</p>;

  return (
    <ul className="contact-search-list">
      {contacts.map((contact) => {
        const conversation = conversations.find((c) => c.peerId === contact.id);
        const displayName = contact.username || `User ${shortId(contact.id)}`;
        return (
          <li key={contact.id} className="contact-search-item">
            <button type="button" className="contact-search-button" onClick={() => onClickContact(contact)}>
              <span className="contact-avatar">{displayName.slice(0, 1).toUpperCase()}</span>
              <span className="contact-main">
                <span className="contact-name">{displayName}</span>
                {conversation && <span className="contact-last">{conversation.lastMessage}</span>}
              </span>
              {conversation && <time>{formatConversationTime(conversation.lastTimestamp)}</time>}
            </button>
          </li>
        );
      })}
    </ul>
  );
}
