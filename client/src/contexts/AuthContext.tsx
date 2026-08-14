import { createContext, useContext, useEffect, useMemo, useState } from "react";
import type { AuthContextValue } from "../types/auth";
import { loginRequest, registerRequest } from "../services/auth";
import { clearSession, loadSession, saveSession } from "../services/storage";
import { setApiToken, setOnUnauthorized } from "../services/api";

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<AuthContextValue>(() => {
    const session = loadSession();
    if (session) {
      setApiToken(session.token);
      return {
        user: session.user,
        token: session.token,
        status: "authenticated",
        login: login,
        register: register,
        logout: logout,
      };
    }
    return {
      user: null,
      token: null,
      status: "anonymous",
      login: login,
      register: register,
      logout: logout,
    };
  });

  async function login(username: string, password: string) {
    const response = await loginRequest(username, password);
    saveSession(response);
    setApiToken(response.token);
    setState((prev) => ({
      ...prev,
      user: response.user,
      token: response.token,
      status: "authenticated",
    }));
  }

  async function register(username: string, password: string) {
    const response = await registerRequest(username, password);
    saveSession(response);
    setApiToken(response.token);
    setState((prev) => ({
      ...prev,
      user: response.user,
      token: response.token,
      status: "authenticated",
    }));
  }

  function logout() {
    clearSession();
    setApiToken(null);
    setState((prev) => ({
      ...prev,
      user: null,
      token: null,
      status: "anonymous",
    }));
  }

  useEffect(() => {
    setOnUnauthorized(logout);
    return () => setOnUnauthorized(null);
  });

  const value = useMemo<AuthContextValue>(() => {
    return {
      ...state,
      login,
      register,
      logout,
    };
  }, [state]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
}
