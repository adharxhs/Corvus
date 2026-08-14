import { createContext, useContext, useMemo } from "react";
import type { ReactNode } from "react";

interface AppContextValue {
  version: string;
}

const AppContext = createContext<AppContextValue>({
  version: "0.1.0",
});

export function AppProvider({ children }: { children: ReactNode }) {
  const value = useMemo(() => ({ version: "0.1.0" }), []);
  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
}

export function useApp(): AppContextValue {
  return useContext(AppContext);
}
