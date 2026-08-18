import { useEffect, useRef } from "react";
import type { ConnectionStatus } from "../websocket/websocket";

export function useConnectionLabel(status: ConnectionStatus, appLabel = "Corvus"): string {
  const wasConnected = useRef(false);

  useEffect(() => {
    if (status === "connected") {
      wasConnected.current = true;
    }
    if (status === "idle") {
      wasConnected.current = false;
    }
  }, [status]);

  if (status === "connected") return appLabel;
  if (status === "idle") return appLabel;
  if (!wasConnected.current) return `${appLabel} (offline)`;
  return "Reconnecting…";
}
