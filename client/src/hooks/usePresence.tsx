import { createContext, useContext, useMemo, useState } from "react";

export type PresenceMap = Map<string, "online" | "offline">;

interface PresenceContextValue {
  presence: PresenceMap;
  setUserStatus: (userId: string, status: "online" | "offline") => void;
  setSnapshot: (online: string[]) => void;
}

const PresenceContext = createContext<PresenceContextValue | null>(null);

export function PresenceProvider({ children }: { children: React.ReactNode }) {
  const [presence, setPresence] = useState<PresenceMap>(() => new Map());

  const value = useMemo<PresenceContextValue>(
    () => ({
      presence,
      setUserStatus: (userId, status) => {
        setPresence((prev) => {
          const next = new Map(prev);
          next.set(userId, status);
          return next;
        });
      },
      setSnapshot: (online) => {
        setPresence((prev) => {
          const next = new Map(prev);
          for (const [key] of next) {
            next.set(key, "offline");
          }
          for (const userId of online) {
            next.set(userId, "online");
          }
          return next;
        });
      },
    }),
    [presence],
  );

  return <PresenceContext.Provider value={value}>{children}</PresenceContext.Provider>;
}

export function usePresence(): PresenceContextValue {
  const ctx = useContext(PresenceContext);
  if (!ctx) {
    throw new Error("usePresence must be used within PresenceProvider");
  }
  return ctx;
}
