import React, { Component, useEffect, type ReactNode } from "react";
import { RouterProvider } from "react-router-dom";
import { AppProvider } from "./contexts/AppContext";
import { ThemeProvider } from "./contexts/ThemeContext";
import { AuthProvider } from "./contexts/AuthContext";
import { WebSocketProvider } from "./contexts/WebSocketContext";
import { ChatProvider } from "./contexts/ChatContext";
import { router } from "./routes";

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error("Uncaught error:", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div style={{ padding: 24, color: "#fff", background: "#0A0D12", minHeight: "100vh" }}>
          <h2>Something went wrong</h2>
          <p style={{ color: "#ff6b7e" }}>{this.state.error?.message ?? "An unexpected error occurred."}</p>
          <button
            type="button"
            onClick={() => {
              this.setState({ hasError: false, error: null });
              window.location.reload();
            }}
            style={{ padding: "8px 16px", borderRadius: 8, background: "#00D9FF", border: 0, fontWeight: "bold", cursor: "pointer" }}
          >
            Reload App
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}

export function App() {
  useEffect(() => {
    const onUnhandledRejection = (event: PromiseRejectionEvent) => {
      console.warn("Unhandled promise rejection:", event.reason);
    };
    window.addEventListener("unhandledrejection", onUnhandledRejection);
    return () => window.removeEventListener("unhandledrejection", onUnhandledRejection);
  }, []);

  return (
    <ErrorBoundary>
      <AppProvider>
        <ThemeProvider>
          <AuthProvider>
            <WebSocketProvider>
              <ChatProvider>
                <RouterProvider router={router} />
              </ChatProvider>
            </WebSocketProvider>
          </AuthProvider>
        </ThemeProvider>
      </AppProvider>
    </ErrorBoundary>
  );
}
