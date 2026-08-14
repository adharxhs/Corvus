import { useMemo, useState } from "react";
import type { DirectMessage } from "../types/message";

export function useMessages() {
  const [messages, setMessages] = useState<DirectMessage[]>([]);

  return useMemo(
    () => ({
      messages,
      addMessage: (message: DirectMessage) => setMessages((prev) => [...prev, message]),
      clearMessages: () => setMessages([]),
    }),
    [messages],
  );
}
