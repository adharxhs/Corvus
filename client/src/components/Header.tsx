import { useAuth } from "../contexts/AuthContext";

export function Header() {
  const { user } = useAuth();
  return (
    <header className="app-header">
      <span className="brand">Corvus</span>
      <span className="spacer" />
      <span className="muted">{user?.username ?? "Guest"}</span>
    </header>
  );
}
