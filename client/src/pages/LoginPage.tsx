import { useNavigate } from "react-router-dom";
import { AuthForm } from "../components/AuthForm";
import { useAuth } from "../contexts/AuthContext";

export function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();

  async function handleLogin(username: string, password: string) {
    await login(username, password);
    navigate("/chats", { replace: true });
  }

  return (
    <main className="auth-page">
      <section className="auth-card">
        <h1>Corvus</h1>
        <p className="muted">Secure encrypted messaging</p>
        <AuthForm mode="login" onSubmit={handleLogin} />
        <p className="muted">
          No account? <a onClick={() => navigate("/register")}>Register</a>
        </p>
      </section>
    </main>
  );
}
