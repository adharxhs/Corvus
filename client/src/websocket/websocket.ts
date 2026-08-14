import { WS_BASE_URL } from "../utils/constants";
import { encodeEnvelope } from "../protocol/envelope";
import { parseServerEnvelope } from "../protocol/parser";
import { validateServerEnvelopeType } from "../protocol/validator";
import type { ServerEnvelope } from "../protocol/payloads";

export type ConnectionStatus = "idle" | "connecting" | "connected" | "disconnected" | "reconnecting" | "error";

type MessageListener = (envelope: ServerEnvelope) => void;
type StatusListener = (status: ConnectionStatus) => void;

export interface WebSocketService {
  connect: (token: string) => void;
  disconnect: () => void;
  send: (type: string, payload: unknown) => boolean;
  onMessage: (listener: MessageListener) => () => void;
  onStatusChange: (listener: StatusListener) => () => void;
}

export function createWebSocketService(): WebSocketService {
  let socket: WebSocket | null = null;
  let currentToken: string | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  let reconnectAttempt = 0;

  const messageListeners = new Set<MessageListener>();
  const statusListeners = new Set<StatusListener>();

  function emitStatus(status: ConnectionStatus) {
    statusListeners.forEach((listener) => listener(status));
  }

  function connect(token: string) {
    currentToken = token;
    reconnectAttempt = 0;
    openSocket(token);
  }

  function openSocket(token: string) {
    if (socket) {
      socket.onclose = null;
      socket.close();
    }

    emitStatus("connecting");
    socket = new WebSocket(`${WS_BASE_URL}/ws?token=${encodeURIComponent(token)}`);
    socket.onopen = () => {
      reconnectAttempt = 0;
      emitStatus("connected");
    };
    socket.onmessage = (event) => {
      const data = event.data;
      if (typeof data !== "string") {
        return;
      }
      try {
        const envelope = parseServerEnvelope(data);
        if (!validateServerEnvelopeType(envelope.type)) {
          return;
        }
        messageListeners.forEach((listener) => listener(envelope as ServerEnvelope));
      } catch {
        // ignore malformed frames
      }
    };
    socket.onclose = () => {
      emitStatus("disconnected");
      scheduleReconnect(token);
    };
    socket.onerror = () => {
      emitStatus("error");
    };
  }

  function scheduleReconnect(token: string) {
    if (!currentToken || currentToken !== token) {
      return;
    }
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
    }
    const delay = Math.min(1000 * 2 ** reconnectAttempt, 30_000);
    reconnectAttempt += 1;
    emitStatus("reconnecting");
    reconnectTimer = setTimeout(() => openSocket(token), delay);
  }

  function disconnect() {
    currentToken = null;
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    if (socket) {
      socket.onclose = null;
      socket.close();
      socket = null;
    }
    emitStatus("idle");
  }

  function send(type: string, payload: unknown) {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return false;
    }
    socket.send(encodeEnvelope({ version: 1, type, payload }));
    return true;
  }

  function onMessage(listener: MessageListener): () => void {
    messageListeners.add(listener);
    return () => messageListeners.delete(listener);
  }

  function onStatusChange(listener: StatusListener): () => void {
    statusListeners.add(listener);
    return () => statusListeners.delete(listener);
  }

  return {
    connect,
    disconnect,
    send,
    onMessage,
    onStatusChange,
  };
}
