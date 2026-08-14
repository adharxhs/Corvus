import { NavLink } from "react-router-dom";
import { StatusIndicator } from "./StatusIndicator";

interface SidebarProps {
  connectionStatus: string;
}

export function Sidebar({ connectionStatus }: SidebarProps) {
  return (
    <aside className="sidebar">
      <div className="brand">Corvus</div>
      <nav className="nav-list">
        <NavLink to="/chat">Chat</NavLink>
        <NavLink to="/settings">Settings</NavLink>
      </nav>
      <StatusIndicator status={connectionStatus} />
    </aside>
  );
}
