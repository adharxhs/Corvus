import { useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import { useWebSocket } from "../contexts/WebSocketContext";
import { useChat } from "../contexts/ChatContext";
import { useConnectionLabel } from "../hooks/useConnectionLabel";
import { AppAvatar } from "./AppAvatar";

export function TitleBar() {
  const { status } = useWebSocket();
  const label = useConnectionLabel(status);

  return (
    <header className="title-bar">
      <h1 className="title-bar-name">{label}</h1>
      <span className="spacer" />
    </header>
  );
}

export function BottomNav() {
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAuth();
  const { conversations } = useChat();

  const active = location.pathname.startsWith("/groups")
    ? "groups"
    : location.pathname.startsWith("/settings")
    ? "settings"
    : "chats";

  const chatsUnread = conversations
    .filter((c) => c.kind === "dm")
    .reduce((sum, c) => sum + c.unreadCount, 0);
  const groupsUnread = conversations
    .filter((c) => c.kind === "group")
    .reduce((sum, c) => sum + c.unreadCount, 0);

  return (
    <nav className="bottom-nav">
      <button
        type="button"
        className={`bottom-nav-item ${active === "chats" ? "active" : ""}`}
        onClick={() => navigate("/chats")}
      >
        <span className="bottom-nav-icon-wrapper">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
          </svg>
          {chatsUnread > 0 && <span className="bottom-nav-badge">{chatsUnread > 99 ? "99+" : chatsUnread}</span>}
        </span>
        <span>Chats</span>
      </button>
      <button
        type="button"
        className={`bottom-nav-item ${active === "groups" ? "active" : ""}`}
        onClick={() => navigate("/groups")}
      >
        <span className="bottom-nav-icon-wrapper">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
            <circle cx="9" cy="7" r="4" />
            <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
            <path d="M16 3.13a4 4 0 0 1 0 7.75" />
          </svg>
          {groupsUnread > 0 && <span className="bottom-nav-badge">{groupsUnread > 99 ? "99+" : groupsUnread}</span>}
        </span>
        <span>Groups</span>
      </button>
      <button
        type="button"
        className={`bottom-nav-item ${active === "settings" ? "active" : ""}`}
        onClick={() => navigate("/settings")}
      >
        <AppAvatar label={user?.username || "Me"} userId={user?.id} size={24} />
        <span>Settings</span>
      </button>
    </nav>
  );
}
