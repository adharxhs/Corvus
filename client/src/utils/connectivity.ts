import { useCallback, useEffect, useState } from "react";
import { probeServer } from "../services/api";

export type Reachability = "checking" | "online" | "offline";

export function useServerReachable(enabled = true): { reachable: Reachability; retry: () => void } {
  const [reachable, setReachable] = useState<Reachability>(() =>
    enabled && typeof navigator !== "undefined" && navigator.onLine === false ? "offline" : "checking",
  );

  const probe = useCallback(() => {
    if (!enabled) {
      setReachable("online");
      return;
    }
    setReachable("checking");
    void probeServer().then((ok) => setReachable(ok ? "online" : "offline"));
  }, [enabled]);

  useEffect(() => {
    if (!enabled) {
      setReachable("online");
      return;
    }
    let active = true;
    const run = () => {
      setReachable("checking");
      void probeServer().then((ok) => {
        if (active) setReachable(ok ? "online" : "offline");
      });
    };
    const onOffline = () => {
      if (active) setReachable("offline");
    };
    run();
    window.addEventListener("online", run);
    window.addEventListener("offline", onOffline);
    return () => {
      active = false;
      window.removeEventListener("online", run);
      window.removeEventListener("offline", onOffline);
    };
  }, [enabled]);

  return { reachable, retry: probe };
}
