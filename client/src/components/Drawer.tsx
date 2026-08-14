import { useEffect, useRef } from "react";

interface DrawerProps {
  open: boolean;
  onClose: () => void;
  username?: string;
  usernameId?: string;
  version: string;
  connectionStatus: string;
  onNavigate: (path: string) => void;
  onLogout: () => void;
}

export function Drawer({ open, onClose, username, usernameId, version, connectionStatus, onNavigate, onLogout }: DrawerProps) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const handleClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        onClose();
      }
    };
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("mousedown", handleClick);
    document.addEventListener("keydown", handleKey);
    return () => {
      document.removeEventListener("mousedown", handleClick);
      document.removeEventListener("keydown", handleKey);
    };
  }, [open, onClose]);

  return (
    <div className={`drawer-backdrop ${open ? "open" : ""}`} aria-hidden={!open}>
      <div className={`drawer ${open ? "open" : ""}`} ref={ref}>
        <div className="drawer-header">
          <AppAvatar name={username ?? "User"} size={64} />
          <div className="drawer-identity">
            <div className="drawer-name">{username ?? "Guest"}</div>
            {usernameId && <div className="drawer-sub">{usernameId}</div>}
          </div>
        </div>
        <nav className="drawer-menu">
          <button type="button" onClick={() => { onNavigate("/chats"); onClose(); }}>Chats</button>
          <button type="button" onClick={() => { onNavigate("/contacts"); onClose(); }}>Contacts</button>
          <button type="button" onClick={() => { onNavigate("/groups"); onClose(); }}>Groups</button>
          <button type="button" onClick={() => { onNavigate("/settings"); onClose(); }}>Settings</button>
        </nav>
        <div className="drawer-footer">
          <div className={`drawer-status status-${connectionStatus}`}>{connectionStatus}</div>
          <div className="drawer-version">Corvus {version}</div>
          <button type="button" className="drawer-logout" onClick={onLogout}>Logout</button>
        </div>
      </div>
    </div>
  );
}

function AppAvatar({ name, size = 64 }: { name: string; size?: number }) {
  const initials = name.trim().split(/\s+/).slice(0, 2).map((p) => p.charAt(0)).join("").toUpperCase();
  return (
    <div className="avatar" style={{ width: size, height: size, fontSize: Math.round(size * 0.38) }}>
      <span>{initials || "U"}</span>
    </div>
  );
}
