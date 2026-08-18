import type { Conversation } from "../types/chat";
import type { Group } from "../types/group";
import { shortId } from "../utils/format";
import { GroupAvatar } from "./GroupAvatar";

interface GroupListProps {
  groups: Group[];
  conversations: Conversation[];
  onOpenGroup: (group: Group) => void;
  onCreateGroup: () => void;
  groupPictureVersion?: Record<string, number>;
}

export function GroupList({ groups, conversations, onOpenGroup, onCreateGroup, groupPictureVersion }: GroupListProps) {
  return (
    <div className="panel contact-list">
      <div className="contact-list-header">
        <h2>Groups</h2>
        <button type="button" className="icon-button" onClick={onCreateGroup} aria-label="Create group">
          <svg viewBox="0 0 24 24" width="20" height="20" stroke="currentColor" fill="none" strokeWidth="2">
            <path d="M12 5v14M5 12h14" />
          </svg>
        </button>
      </div>
      {groups.length === 0 && <p className="sheet-hint">No groups yet</p>}
      {groups.map((group) => {
        const conversation = conversations.find((c) => c.peerId === group.id);
        const unreadCount = conversation?.unreadCount ?? 0;
        const displayName = group.name || `Group ${shortId(group.id)}`;
        return (
          <button key={group.id} type="button" className="contact-list-item" onClick={() => onOpenGroup(group)}>
            <GroupAvatar name={displayName} groupId={group.id} size={40} pictureVersion={groupPictureVersion?.[group.id]} />
            <span className="contact-main">
              <span className="contact-name">{displayName}</span>
              {conversation && <span className="contact-last">{conversation.lastMessage}</span>}
            </span>
            {unreadCount > 0 && <span className="unread-badge">{unreadCount}</span>}
          </button>
        );
      })}
    </div>
  );
}
