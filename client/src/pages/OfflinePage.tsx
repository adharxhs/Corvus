import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useServerReachable } from "../utils/connectivity";

interface OfflinePageProps {
  onRetry?: () => void;
}

export function OfflinePage({ onRetry }: OfflinePageProps) {
  const { reachable, retry } = useServerReachable(!onRetry);
  const navigate = useNavigate();

  useEffect(() => {
    if (!onRetry && reachable === "online") {
      navigate("/login", { replace: true });
    }
  }, [onRetry, reachable, navigate]);

  return (
    <main className="auth-page">
      <section className="auth-card">
        <h1>You're offline</h1>
        <p className="muted">
          Corvus can't reach the server right now. Connect to the internet and try again to sign in.
        </p>
        <div className="offline-actions">
          <button
            type="button"
            className="primary-button"
            onClick={() => (onRetry ? onRetry() : retry())}
          >
            Retry
          </button>
        </div>
      </section>
    </main>
  );
}
