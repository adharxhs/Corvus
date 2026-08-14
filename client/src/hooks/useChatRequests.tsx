import { createContext, useContext } from "react";
import { useContacts } from "./useContacts";

const ChatRequestContext = createContext<ReturnType<typeof useContacts> | null>(null);

export function ChatRequestsProvider({ children }: { children: React.ReactNode }) {
  const contacts = useContacts();
  return <ChatRequestContext.Provider value={contacts}>{children}</ChatRequestContext.Provider>;
}

export function useChatRequests() {
  const ctx = useContext(ChatRequestContext);
  if (!ctx) {
    throw new Error("useChatRequests must be used within ChatRequestsProvider");
  }
  return ctx;
}
