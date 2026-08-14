import { useCallback, useEffect, useMemo, useState } from "react";
import type { Contact } from "../types/user";
import type { RelationshipResponse } from "../types/relationship";
import { getUserById, lookupUserByUsername } from "../services/users";
import { sendChatRequest, acceptChatRequest, rejectChatRequest, listChatRequests } from "../services/relationships";
import { useWebSocket } from "../contexts/WebSocketContext";
import { useAuth } from "../contexts/AuthContext";

export function useContacts() {
  const { user } = useAuth();
  const { presence } = useWebSocket();
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [pendingIncoming, setPendingIncoming] = useState<RelationshipResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!user) {
      setContacts([]);
      setPendingIncoming([]);
      return;
    }
    let active = true;
    setLoading(true);
    listChatRequests()
      .then(async (requests) => {
        if (!active) {
          return;
        }
        setPendingIncoming(requests);
        const resolved = await Promise.all(
          requests.map(async (rel) => {
            const otherId = rel.recipient_id === user.id ? rel.requester_id : rel.recipient_id;
            let username: string | undefined;
            try {
              const u = await getUserById(otherId);
              username = u.username;
            } catch {
              // ignore
            }
            return {
              id: otherId,
              username,
              relationship: rel.status,
              presence: presence.get(otherId) ?? ("unknown" as const),
            };
          }),
        );
        if (active) {
          setContacts(resolved);
        }
      })
      .catch(() => {
        if (active) {
          setError("Failed to load contacts");
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false);
        }
      });
    return () => {
      active = false;
    };
  }, [user, presence]);

  const startNewChat = useCallback(async (username: string): Promise<RelationshipResponse> => {
    const lookup = await lookupUserByUsername(username);
    return sendChatRequest(lookup.id);
  }, []);

  const acceptRequest = useCallback(async (requesterId: string): Promise<RelationshipResponse> => {
    return acceptChatRequest(requesterId);
  }, []);

  const rejectRequest = useCallback(async (requesterId: string): Promise<RelationshipResponse> => {
    return rejectChatRequest(requesterId);
  }, []);

  return useMemo(
    () => ({
      contacts,
      pendingIncoming,
      loading,
      error,
      startNewChat,
      acceptRequest,
      rejectRequest,
    }),
    [contacts, pendingIncoming, loading, error, startNewChat, acceptRequest, rejectRequest],
  );
}
