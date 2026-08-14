import { useEffect, useRef, useState } from "react";
import type { ConnectionStatus } from "../websocket/websocket";

export function useConnectionLabel(status: ConnectionStatus, appLabel = "Corvus"): string {
  const [showConnecting, setShowConnecting] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (status === "connected") {
      setShowConnecting(false);
      if (timerRef.current !== null) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
      return;
    }

    if (timerRef.current === null) {
      timerRef.current = setTimeout(() => {
        setShowConnecting(true);
      }, 30_000);
    }
  }, [status]);

  useEffect(() => {
    return () => {
      if (timerRef.current !== null) clearTimeout(timerRef.current);
    };
  }, []);

  return showConnecting ? "Connecting" : appLabel;
}
