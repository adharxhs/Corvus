import { createContext, useContext, useEffect, useRef } from "react";
import { useAuth } from "./AuthContext";
import {
  createWebSocketService,
  type ConnectionStatus,
  type WebSocketService,
} from "../websocket/websocket";
import { useState } from "react";
import type { ServerEnvelope } from "../protocol/payloads";
import type { PresenceSnapshotPayload, PresencePayload } from "../types/presence";

interface WebSocketContextValue {
  status: ConnectionStatus;
  presence: Map<string, "online" | "offline">;
  service: WebSocketService;
}

const WebSocketContext = createContext<WebSocketContextValue | null>(null);

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const { token, status: authStatus } = useAuth();
  const serviceRef = useRef(createWebSocketService());
  const [status, setStatus] = useState<ConnectionStatus>("idle");
  const [presence, setPresence] = useState<Map<string, "online" | "offline">>(() => new Map());

  useEffect(() => {
    const service = serviceRef.current;

    if (authStatus === "authenticated" && token) {
      service.connect(token);
    } else {
      service.disconnect();
      setPresence(new Map());
    }

    return () => service.disconnect();
  }, [authStatus, token]);

  useEffect(() => {
    const service = serviceRef.current;
    const onOnline = () => {
      if (authStatus === "authenticated" && token) {
        service.connect(token);
      }
    };
    window.addEventListener("online", onOnline);
    return () => window.removeEventListener("online", onOnline);
  }, [authStatus, token]);

  useEffect(() => {
    const service = serviceRef.current;

    const unsubStatus = service.onStatusChange((s) => {
      setStatus(s);
    });

    const unsubMessage = service.onMessage((envelope: ServerEnvelope) => {
      if (envelope.type === "presence_snapshot") {
        const payload = envelope.payload as PresenceSnapshotPayload;
        const onlineUsers = Array.isArray(payload?.online) ? payload.online : [];
        setPresence((prev) => {
          const next = new Map<string, "online" | "offline">();
          for (const [key] of prev) {
            next.set(key, "offline");
          }
          for (const userId of onlineUsers) {
            if (userId) next.set(userId, "online");
          }
          return next;
        });
      } else if (envelope.type === "presence") {
        const payload = envelope.payload as PresencePayload;
        if (payload?.user_id) {
          setPresence((prev) => {
            const next = new Map(prev);
            next.set(payload.user_id, payload.status === "online" ? "online" : "offline");
            return next;
          });
        }
      }
    });

    return () => {
      unsubStatus();
      unsubMessage();
    };
  }, []);

  const value = {
    status,
    presence,
    service: serviceRef.current,
  };

  return <WebSocketContext.Provider value={value}>{children}</WebSocketContext.Provider>;
}

export function useWebSocket(): WebSocketContextValue {
  const ctx = useContext(WebSocketContext);
  if (!ctx) {
    throw new Error("useWebSocket must be used within WebSocketProvider");
  }
  return ctx;
}
