import { useNavigate } from "react-router-dom";
import { AuthForm } from "../components/AuthForm";
import { useAuth } from "../contexts/AuthContext";

export function RegisterPage() {
  const { register } = useAuth();
  const navigate = useNavigate();

  async function handleRegister(username: string, password: string) {
    await register(username, password);
    navigate("/chats", { replace: true });
  }

  return (
    <main className="auth-page">
      <section className="auth-card">
        <h1>Create account</h1>
        <p className="muted">Private encrypted messaging</p>
        <AuthForm mode="register" onSubmit={handleRegister} />
        <p className="muted">
          Already have an account? <a onClick={() => navigate("/login")}>Login</a>
        </p>
      </section>
    </main>
  );
}
