import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import { useAuth } from "./AuthContext";
import { useWebSocket } from "./WebSocketContext";
import type { Conversation, ChatMessage } from "../types/chat";
import type { Contact } from "../types/user";
import type { Group } from "../types/group";
import type { DirectMessage, GroupMessage } from "../types/message";
import type { RelationshipResponse } from "../types/relationship";
import type { ServerEnvelope } from "../protocol/payloads";
import { acceptChatRequest, listChatRequests, rejectChatRequest, sendChatRequest } from "../services/relationships";
import { lookupUserByUsername, getUserById } from "../services/users";
import { createGroup as createGroupRequest, listGroupMembers } from "../services/groups";
import { isNetworkError } from "../services/api";
import { uid } from "../utils/random";
import { CHATS_STORAGE_KEY, CONTACTS_STORAGE_KEY, PENDING_MESSAGES_STORAGE_KEY, USERNAME_CACHE_KEY } from "../utils/constants";

interface ChatContextValue {
  conversations: Conversation[];
  contacts: Contact[];
  groups: Group[];
  messagesByPeer: Record<string, ChatMessage[]>;
  pendingIncoming: RelationshipResponse[];
  selectedPeerId: string | null;
  selectedConversation: Conversation | null;
  loading: boolean;
  error: string | null;
  offline: boolean;
  selectConversation: (peerId: string) => void;
  clearSelection: () => void;
  sendDirectMessage: (contact: Contact, content: string) => void;
  sendGroupMessage: (group: Group, content: string) => void;
  startNewChat: (username: string) => Promise<void>;
  acceptRequest: (requesterId: string) => Promise<void>;
  rejectRequest: (requesterId: string) => Promise<void>;
  createGroup: (groupId: string) => Promise<void>;
  refresh: () => Promise<void>;
}

interface PersistedChats {
  conversations: Conversation[];
  messagesByPeer: Record<string, ChatMessage[]>;
  groups: Group[];
}

interface PendingMessage {
  id: string;
  peerId: string;
  kind: "dm" | "group";
  type: "message" | "group_message";
  payload: DirectMessage | GroupMessage;
  createdAt: number;
}

const ChatContext = createContext<ChatContextValue | null>(null);

function loadJson<T>(key: string, fallback: T): T {
  const raw = localStorage.getItem(key);
  if (!raw) return fallback;
  try {
    return JSON.parse(raw) as T;
  } catch {
    localStorage.removeItem(key);
    return fallback;
  }
}

function conversationTitle(peerId: string, contacts: Contact[], kind: "dm" | "group", usernameCache?: Map<string, string>) {
  if (kind === "group") return peerId;
  const contact = contacts.find((c) => c.id === peerId);
  if (contact?.username) return contact.username;
  if (usernameCache?.has(peerId)) return usernameCache.get(peerId)!;
  return peerId;
}

function sortConversations(items: Conversation[]) {
  return [...items].sort((a, b) => b.lastTimestamp - a.lastTimestamp || a.title.localeCompare(b.title));
}

export function ChatProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();
  const { service, presence, status } = useWebSocket();

  const userChatsKey = user ? `${CHATS_STORAGE_KEY}_${user.id}` : CHATS_STORAGE_KEY;
  const userContactsKey = user ? `${CONTACTS_STORAGE_KEY}_${user.id}` : CONTACTS_STORAGE_KEY;
  const userPendingKey = user ? `${PENDING_MESSAGES_STORAGE_KEY}_${user.id}` : PENDING_MESSAGES_STORAGE_KEY;

  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [messagesByPeer, setMessagesByPeer] = useState<Record<string, ChatMessage[]>>({});
  const [pendingMessages, setPendingMessages] = useState<PendingMessage[]>([]);
  const pendingRef = useRef<PendingMessage[]>([]);
  const [pendingIncoming, setPendingIncoming] = useState<RelationshipResponse[]>([]);
  const [selectedPeerId, setSelectedPeerId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const offline = status === "disconnected" || status === "error" || status === "reconnecting";
  const [usernameCache, setUsernameCache] = useState<Map<string, string>>(() => {
    try {
      const raw = localStorage.getItem(USERNAME_CACHE_KEY);
      return raw ? new Map(Object.entries(JSON.parse(raw))) : new Map();
    } catch { return new Map(); }
  });

  useEffect(() => {
    if (!user) {
      setConversations([]);
      setContacts([]);
      setGroups([]);
      setMessagesByPeer({});
      setPendingMessages([]);
      setPendingIncoming([]);
      return;
    }
    const persisted = loadJson<PersistedChats>(userChatsKey, { conversations: [], messagesByPeer: {}, groups: [] });
    const storedContacts = loadJson<Contact[]>(userContactsKey, []);
    const storedPending = loadJson<PendingMessage[]>(userPendingKey, []);
    setConversations(persisted.conversations);
    setContacts(storedContacts);
    setGroups(persisted.groups);
    setMessagesByPeer(persisted.messagesByPeer);
    setPendingMessages(storedPending);
  }, [user, userChatsKey, userContactsKey, userPendingKey]);

  const selectedConversation = conversations.find((c) => c.peerId === selectedPeerId) ?? null;

  useEffect(() => {
    pendingRef.current = pendingMessages;
  }, [pendingMessages]);

  const markMessagesSent = useCallback((ids: string[]) => {
    setMessagesByPeer((prev) => {
      const idSet = new Set(ids);
      let changed = false;
      const next: Record<string, ChatMessage[]> = {};
      for (const [peerId, messages] of Object.entries(prev)) {
        next[peerId] = messages.map((m) => {
          if (m.direction === "out" && m.status !== "sent" && idSet.has(m.id)) {
            changed = true;
            return { ...m, status: "sent" };
          }
          return m;
        });
      }
      return changed ? next : prev;
    });
  }, []);

  const enqueuePending = useCallback((item: PendingMessage) => {
    setPendingMessages((prev) => (prev.some((p) => p.id === item.id) ? prev : [...prev, item]));
  }, []);

  const flushPending = useCallback(() => {
    const queued = pendingRef.current;
    if (queued.length === 0) return;
    const sentIds: string[] = [];
    for (const item of queued) {
      if (service.send(item.type, item.payload)) {
        sentIds.push(item.id);
      }
    }
    if (sentIds.length === 0) return;
    setPendingMessages((prev) => prev.filter((p) => !sentIds.includes(p.id)));
    markMessagesSent(sentIds);
  }, [service, markMessagesSent]);

  const upsertConversation = useCallback((peerId: string, kind: "dm" | "group", lastMessage = "") => {
    setConversations((prev) => {
      const title = conversationTitle(peerId, contacts, kind, usernameCache);
      const existing = prev.find((c) => c.peerId === peerId);
      const next = existing
        ? prev.map((c) =>
            c.peerId === peerId
              ? { ...c, title, kind, lastMessage: lastMessage || c.lastMessage, lastTimestamp: Date.now() }
              : c,
          )
        : [...prev, { peerId, kind, title, lastMessage, lastTimestamp: Date.now(), unreadCount: 0 }];
      return sortConversations(next);
    });
  }, [contacts, usernameCache]);

  const addMessage = useCallback((message: ChatMessage, kind: "dm" | "group") => {
    setMessagesByPeer((prev) => ({ ...prev, [message.peerId]: [...(prev[message.peerId] ?? []), message] }));
    setConversations((prev) => {
      const title = conversationTitle(message.peerId, contacts, kind, usernameCache);
      const existing = prev.find((c) => c.peerId === message.peerId);
      const unreadCount = message.direction === "in" && selectedPeerId !== message.peerId ? (existing?.unreadCount ?? 0) + 1 : 0;
      const item: Conversation = {
        peerId: message.peerId,
        kind,
        title,
        lastMessage: message.content,
        lastTimestamp: message.timestamp,
        unreadCount,
      };
      return sortConversations(existing ? prev.map((c) => (c.peerId === message.peerId ? item : c)) : [...prev, item]);
    });
  }, [contacts, selectedPeerId, usernameCache]);

  const refresh = useCallback(async () => {
    if (!user) return;
    setLoading(true);
    setError(null);
    try {
      const rawRequests = await listChatRequests();
      const requests = Array.isArray(rawRequests) ? rawRequests : [];
      setPendingIncoming(requests.filter((r) => r && r.status === "pending" && r.recipient_id === user.id));
      const accepted = requests.filter((r) => r && r.status === "accepted");
      const resolved = await Promise.all(accepted.map(async (rel) => {
        const otherId = rel.requester_id === user.id ? rel.recipient_id : rel.requester_id;
        try {
          const resolvedUser = await getUserById(otherId);
          return { id: otherId, username: resolvedUser.username, relationship: "accepted" as const, presence: "unknown" as const };
        } catch {
          return { id: otherId, relationship: "accepted" as const, presence: "unknown" as const };
        }
      }));
      setContacts(resolved);
    } catch (err) {
      if (!isNetworkError(err)) {
        setError(err instanceof Error ? err.message : "Failed to refresh chats");
      }
    } finally {
      setLoading(false);
    }
  }, [user]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const prevStatusRef = useRef(status);

  useEffect(() => {
    if (status === "connected" && prevStatusRef.current !== "connected") {
      flushPending();
      void refresh();
    }
    prevStatusRef.current = status;
  }, [status, flushPending, refresh]);

  useEffect(() => {
    setContacts((prev) => prev.map((contact) => ({ ...contact, presence: presence.get(contact.id) ?? "unknown" })));
  }, [presence]);

  useEffect(() => {
    if (!user) return;
    localStorage.setItem(userChatsKey, JSON.stringify({ conversations, messagesByPeer, groups }));
  }, [user, userChatsKey, conversations, messagesByPeer, groups]);

  useEffect(() => {
    if (!user) return;
    localStorage.setItem(userPendingKey, JSON.stringify(pendingMessages));
  }, [user, userPendingKey, pendingMessages]);

  useEffect(() => {
    if (!user) return;
    localStorage.setItem(userContactsKey, JSON.stringify(contacts));
    setConversations((prev) => {
      const existingPeerIds = new Set(prev.map((c) => c.peerId));
      const newDms: Conversation[] = contacts
        .filter((contact) => !existingPeerIds.has(contact.id))
        .map((contact) => ({
          peerId: contact.id,
          kind: "dm",
          title: contact.username ?? contact.id,
          lastMessage: "",
          lastTimestamp: Date.now(),
          unreadCount: 0,
        }));
      const updated = prev.map((c) => ({
        ...c,
        title: conversationTitle(c.peerId, contacts, c.kind, usernameCache),
      }));
      return sortConversations([...updated, ...newDms]);
    });
  }, [user, userContactsKey, contacts, usernameCache]);

  useEffect(() => {
    if (!user) return;
    const unresolved = conversations.filter(
      (c) => c.kind === "dm" && c.title === c.peerId && !usernameCache.has(c.peerId)
    );
    if (unresolved.length === 0) return;
    let active = true;
    void (async () => {
      for (const conv of unresolved) {
        try {
          const u = await getUserById(conv.peerId);
          if (active && u && u.username) {
            setUsernameCache((prev) => {
              const next = new Map(prev);
              next.set(conv.peerId, u.username);
              return next;
            });
          }
        } catch {
          // ignore
        }
      }
    })();
    return () => { active = false; };
  }, [user, conversations, usernameCache]);

  useEffect(() => {
    const obj: Record<string, string> = {};
    usernameCache.forEach((v, k) => { obj[k] = v; });
    localStorage.setItem(USERNAME_CACHE_KEY, JSON.stringify(obj));
  }, [usernameCache]);

  useEffect(() => {
    const unsubscribe = service.onMessage((envelope: ServerEnvelope) => {
      if (envelope.type === "message") {
        const payload = envelope.payload as DirectMessage;
        const peerId = payload.sender_id ?? payload.recipient_id;
        addMessage({ id: uid(), peerId, content: payload.content, direction: "in", timestamp: Date.now() }, "dm");
      }
      if (envelope.type === "group_message") {
        const payload = envelope.payload as GroupMessage;
        addMessage({ id: uid(), peerId: payload.group_id, content: payload.content, direction: "in", timestamp: Date.now() }, "group");
      }
      if (envelope.type === "group_profile_picture_updated") {
        const payload = envelope.payload as { group_id: string; version: number };
        upsertConversation(payload.group_id, "group", "Group picture updated");
      }
      if (envelope.type === "error") {
        setError("Message failed");
      }
    });
    return unsubscribe;
  }, [addMessage, service]);

  function selectConversation(peerId: string) {
    setSelectedPeerId(peerId);
    setConversations((prev) => prev.map((c) => c.peerId === peerId ? { ...c, unreadCount: 0 } : c));
  }

  function sendDirectMessage(contact: Contact, content: string) {
    const payload: DirectMessage = { recipient_id: contact.id, content };
    const timestamp = Date.now();
    const id = uid();
    const sent = service.send("message", payload);
    addMessage({ id, peerId: contact.id, content, direction: "out", timestamp, status: sent ? "sent" : "pending" }, "dm");
    if (!sent) {
      enqueuePending({ id, peerId: contact.id, kind: "dm", type: "message", payload, createdAt: timestamp });
    }
  }

  function sendGroupMessage(group: Group, content: string) {
    const payload: GroupMessage = { group_id: group.id, content };
    const timestamp = Date.now();
    const id = uid();
    const sent = service.send("group_message", payload);
    addMessage({ id, peerId: group.id, content, direction: "out", timestamp, status: sent ? "sent" : "pending" }, "group");
    if (!sent) {
      enqueuePending({ id, peerId: group.id, kind: "group", type: "group_message", payload, createdAt: timestamp });
    }
  }

  async function startNewChat(username: string) {
    const lookup = await lookupUserByUsername(username);
    await sendChatRequest(lookup.id);
    const contact: Contact = { id: lookup.id, username, relationship: "pending", presence: presence.get(lookup.id) ?? "unknown" };
    setContacts((prev) => prev.some((c) => c.id === lookup.id) ? prev : [...prev, contact]);
    upsertConversation(lookup.id, "dm", "Chat request sent");
  }

  async function acceptRequest(requesterId: string) {
    await acceptChatRequest(requesterId);
    await refresh();
    upsertConversation(requesterId, "dm", "Chat request accepted");
  }

  async function rejectRequest(requesterId: string) {
    await rejectChatRequest(requesterId);
    setPendingIncoming((prev) => prev.filter((r) => r.requester_id !== requesterId));
  }

  async function createGroup(groupId: string) {
    await createGroupRequest(groupId);
    const members = await listGroupMembers(groupId);
    const group = { id: groupId, members: members.map((m) => m.user_id) };
    setGroups((prev) => prev.some((g) => g.id === groupId) ? prev : [...prev, group]);
    upsertConversation(groupId, "group", "Group created");
  }

  const value = useMemo<ChatContextValue>(() => ({
    conversations,
    contacts,
    groups,
    messagesByPeer,
    pendingIncoming,
    selectedPeerId,
    selectedConversation,
    loading,
    error,
    offline,
    selectConversation,
    clearSelection: () => setSelectedPeerId(null),
    sendDirectMessage,
    sendGroupMessage,
    startNewChat,
    acceptRequest,
    rejectRequest,
    createGroup,
    refresh,
  }), [conversations, contacts, groups, messagesByPeer, pendingIncoming, selectedPeerId, selectedConversation, loading, error, offline, refresh]);

  return <ChatContext.Provider value={value}>{children}</ChatContext.Provider>;
}

export function useChat(): ChatContextValue {
  const ctx = useContext(ChatContext);
  if (!ctx) {
    throw new Error("useChat must be used within ChatProvider");
  }
  return ctx;
}
