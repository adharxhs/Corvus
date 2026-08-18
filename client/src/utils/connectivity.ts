import { useCallback, useEffect, useState } from "react";
import { probeServer } from "../services/api";

export type Reachability = "checking" | "online" | "offline";

export function useServerReachable(enabled = true): { reachable: Reachability; retry: () => void } {
  const [reachable, setReachable] = useState<Reachability>("checking");

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
    run();
    window.addEventListener("online", run);
    return () => {
      active = false;
      window.removeEventListener("online", run);
    };
  }, [enabled]);

  return { reachable, retry: probe };
}
