import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import { useAuth } from "./AuthContext";
import { useWebSocket } from "./WebSocketContext";
import type { Conversation, ChatMessage } from "../types/chat";
import type { Contact } from "../types/user";
import type { Group } from "../types/group";
import type { DirectMessage, GroupMessage } from "../types/message";
import type { RelationshipResponse } from "../types/relationship";
import type { ServerEnvelope, MemberJoinedPayload } from "../protocol/payloads";
import { acceptChatRequest, listChatRequests, rejectChatRequest, sendChatRequest } from "../services/relationships";
import { lookupUserByUsername, getUserById } from "../services/users";
import { createGroup as createGroupRequest, listGroupMembers, inviteToGroup as inviteToGroupRequest, getGroup as getGroupRequest, renameGroup as renameGroupRequest, listMyGroups as listMyGroupsRequest, acceptGroupInvite as acceptGroupInviteRequest, listGroupInvites } from "../services/groups";
import { isNetworkError } from "../services/api";
import { uid } from "../utils/random";
import { clearAllProfilePictures } from "../services/profilePictureCache";
import { clearGroupPictureCache } from "../services/groupProfilePictureCache";
import { CHATS_STORAGE_KEY, CONTACTS_STORAGE_KEY, PENDING_MESSAGES_STORAGE_KEY, USERNAME_CACHE_KEY } from "../utils/constants";
import {
  encryptMessage,
  decryptMessage,
  encryptGroupMessage,
  decryptGroupMessage,
  processSenderKeyDistribution,
  type EncryptedPayload,
  type GroupEncryptedPayload,
} from "../services/crypto";

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
  createGroup: (groupId: string, name: string) => Promise<void>;
  inviteToGroup: (groupId: string, userId: string) => Promise<void>;
  renameGroup: (groupId: string, name: string) => Promise<void>;
  acceptGroupInvite: (groupId: string) => Promise<void>;
  fetchGroupInfo: (groupId: string) => Promise<void>;
  refresh: () => Promise<void>;
  usernameFor: (id: string) => string;
  resolveUsername: (id: string) => void;
  groupNameFor: (groupId: string) => string;
  groupPictureVersion: Record<string, number>;
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

function conversationTitle(peerId: string, contacts: Contact[], kind: "dm" | "group", groups?: Group[], usernameCache?: Map<string, string>) {
  if (kind === "group") {
    const group = groups?.find((g) => g.id === peerId);
    if (group?.name) return group.name;
    return `Group ${peerId.slice(0, 8)}`;
  }
  const contact = contacts.find((c) => c.id === peerId);
  if (contact?.username) return contact.username;
  if (usernameCache?.has(peerId)) return usernameCache.get(peerId)!;
  return `User ${peerId.slice(0, 8)}`;
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
  const [groupPictureVersion, setGroupPictureVersion] = useState<Record<string, number>>({});
  const bootedRef = useRef(false);
  const usernameCacheRef = useRef(usernameCache);
  usernameCacheRef.current = usernameCache;

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

  const upsertConversation = useCallback((peerId: string, kind: "dm" | "group", lastMessage = "", titleOverride?: string) => {
    setConversations((prev) => {
      const title = titleOverride || conversationTitle(peerId, contacts, kind, groups, usernameCache);
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
  }, [contacts, groups, usernameCache]);

  const addMessage = useCallback((message: ChatMessage, kind: "dm" | "group") => {
    setMessagesByPeer((prev) => ({ ...prev, [message.peerId]: [...(prev[message.peerId] ?? []), message] }));
    setConversations((prev) => {
      const title = conversationTitle(message.peerId, contacts, kind, groups, usernameCache);
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
  }, [contacts, selectedPeerId, groups, usernameCache]);

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

      // Load user's groups from the server.
      try {
        const myGroups = await listMyGroupsRequest();
        setGroups((prev) => {
          const existingMap = new Map(prev.map((g) => [g.id, g]));
          const merged: Group[] = myGroups.map((g) => {
            const existing = existingMap.get(g.id);
            return { id: g.id, name: g.name, members: existing?.members ?? [] };
          });
          return merged;
        });
        for (const g of myGroups) {
          upsertConversation(g.id, "group", "", g.name);
        }
      } catch {
        // ignore
      }

      // Resolve usernames for group invite inviters.
      try {
        const invites = await listGroupInvites();
        for (const inv of invites) {
          if (usernameCacheRef.current.has(inv.invited_by)) continue;
          try {
            const u = await getUserById(inv.invited_by);
            if (u && u.username) {
              setUsernameCache((prev) => {
                if (prev.has(inv.invited_by)) return prev;
                const next = new Map(prev);
                next.set(inv.invited_by, u.username);
                return next;
              });
            }
          } catch {
            // ignore
          }
        }
      } catch {
        // ignore
      }

      // Resolve usernames for pending incoming requests (Fix 4: display username instead of ID).
      const pendingIds = requests.filter((r) => r && r.status === "pending" && r.recipient_id === user.id).map((r) => r.requester_id);
      for (const otherId of pendingIds) {
        if (usernameCacheRef.current.has(otherId)) continue;
        try {
          const u = await getUserById(otherId);
          if (u && u.username) {
            setUsernameCache((prev) => {
              const next = new Map(prev);
              next.set(otherId, u.username);
              return next;
            });
          }
        } catch {
          // Ignore
        }
      }
    } catch (err) {
      if (!isNetworkError(err)) {
        setError(err instanceof Error ? err.message : "Failed to refresh chats");
      }
    } finally {
      setLoading(false);
      bootedRef.current = true;
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
    if (!bootedRef.current) return;
    setConversations((prev) => {
      const existingPeerIds = new Set(prev.map((c) => c.peerId));
      const newDms: Conversation[] = contacts
        .filter((contact) => !existingPeerIds.has(contact.id))
        .map((contact) => ({
          peerId: contact.id,
          kind: "dm",
          title: contact.username ?? `User ${contact.id.slice(0, 8)}`,
          lastMessage: "",
          lastTimestamp: Date.now(),
          unreadCount: 0,
        }));
      const updated = prev.map((c) => ({
        ...c,
        title: conversationTitle(c.peerId, contacts, c.kind, groups, usernameCache),
      }));
      return sortConversations([...updated, ...newDms]);
    });
  }, [user, userContactsKey, contacts, groups, usernameCache]);

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

  const selectConversation = useCallback((peerId: string) => {
    setSelectedPeerId(peerId);
    setConversations((prev) => {
      const target = prev.find((c) => c.peerId === peerId);
      if (!target || target.unreadCount === 0) return prev;
      return prev.map((c) => c.peerId === peerId ? { ...c, unreadCount: 0 } : c);
    });
  }, []);

  const sendDirectMessage = useCallback(async (contact: Contact, content: string) => {
    let encryptedContent = content;
    try {
      const encrypted = await encryptMessage(contact.id, content);
      encryptedContent = JSON.stringify(encrypted);
    } catch {
      // If encryption fails (no session), send plaintext for now
      // In production, this should fail and prompt session setup
    }

    const payload: DirectMessage = { recipient_id: contact.id, content: encryptedContent };
    const timestamp = Date.now();
    const id = uid();
    const sent = service.send("message", payload);
    if (sent) setError(null);
    addMessage({ id, peerId: contact.id, senderId: user?.id, content, direction: "out", timestamp, status: sent ? "sent" : "pending" }, "dm");
    if (!sent) {
      enqueuePending({ id, peerId: contact.id, kind: "dm", type: "message", payload, createdAt: timestamp });
    }
  }, [addMessage, enqueuePending, service, user]);

  const sendGroupMessage = useCallback(async (group: Group, content: string) => {
    let encryptedContent = content;
    try {
      const encrypted = await encryptGroupMessage(group.id, user?.id ?? "", content);
      encryptedContent = JSON.stringify(encrypted);
    } catch {
      // If encryption fails (no sender key), send plaintext for now
    }

    const payload: GroupMessage = { group_id: group.id, content: encryptedContent };
    const timestamp = Date.now();
    const id = uid();
    const sent = service.send("group_message", payload);
    if (sent) setError(null);
    addMessage({ id, peerId: group.id, senderId: user?.id, content, direction: "out", timestamp, status: sent ? "sent" : "pending" }, "group");
    if (!sent) {
      enqueuePending({ id, peerId: group.id, kind: "group", type: "group_message", payload, createdAt: timestamp });
    }
  }, [addMessage, enqueuePending, service, user]);

  const startNewChat = useCallback(async (username: string) => {
    const lookup = await lookupUserByUsername(username);
    try {
      await sendChatRequest(lookup.id);
      const contact: Contact = { id: lookup.id, username, relationship: "pending", presence: presence.get(lookup.id) ?? "unknown" };
      setContacts((prev) => prev.some((c) => c.id === lookup.id) ? prev : [...prev, contact]);
      upsertConversation(lookup.id, "dm", "Chat request sent");
    } catch (err: unknown) {
      // If already connected or request already pending, still surface the conversation.
      const isConflict = err instanceof Error && (err.message.includes("already connected") || err.message.includes("request already pending"));
      if (isConflict) {
        const contact: Contact = { id: lookup.id, username, relationship: "accepted", presence: presence.get(lookup.id) ?? "unknown" };
        setContacts((prev) => prev.some((c) => c.id === lookup.id) ? prev : [...prev, contact]);
        upsertConversation(lookup.id, "dm");
        return;
      }
      throw err;
    }
  }, [presence, upsertConversation]);

  const acceptRequest = useCallback(async (requesterId: string) => {
    await acceptChatRequest(requesterId);
    await refresh();
    upsertConversation(requesterId, "dm", "Chat request accepted");
  }, [refresh, upsertConversation]);

  const rejectRequest = useCallback(async (requesterId: string) => {
    await rejectChatRequest(requesterId);
    setPendingIncoming((prev) => prev.filter((r) => r.requester_id !== requesterId));
  }, []);

  const createGroup = useCallback(async (groupId: string, name: string) => {
    await createGroupRequest(groupId, name);
    const members = await listGroupMembers(groupId);
    const group = { id: groupId, name, members: members.map((m) => m.user_id) };
    setGroups((prev) => prev.some((g) => g.id === groupId) ? prev : [...prev, group]);
    upsertConversation(groupId, "group", "Group created", name);
  }, [upsertConversation]);

  const inviteToGroupMember = useCallback(async (groupId: string, userId: string) => {
    await inviteToGroupRequest(groupId, userId);
  }, []);

  const renameGroup = useCallback(async (groupId: string, name: string) => {
    await renameGroupRequest(groupId, name);
    setGroups((prev) => prev.map((g) => g.id === groupId ? { ...g, name } : g));
    upsertConversation(groupId, "group", "", name);
  }, [upsertConversation]);

  const acceptGroupInvite = useCallback(async (groupId: string) => {
    await acceptGroupInviteRequest(groupId);
    try {
      const info = await getGroupRequest(groupId);
      const members = await listGroupMembers(groupId);
      const group: Group = { id: info.id, name: info.name, members: members.map((m) => m.user_id) };
      setGroups((prev) => prev.some((g) => g.id === groupId) ? prev.map((g) => g.id === groupId ? group : g) : [...prev, group]);
      upsertConversation(groupId, "group", "", info.name);
    } catch {
      // Group will appear on next refresh
    }
  }, [upsertConversation]);

  const fetchGroupInfo = useCallback(async (groupId: string) => {
    const info = await getGroupRequest(groupId);
    const members = await listGroupMembers(groupId);
    setGroups((prev) => {
      const existing = prev.find((g) => g.id === groupId);
      const updated: Group = { id: info.id, name: info.name, members: members.map((m) => m.user_id) };
      if (existing) {
        return prev.map((g) => g.id === groupId ? updated : g);
      }
      return [...prev, updated];
    });
    upsertConversation(groupId, "group", "", info.name);
  }, [upsertConversation]);

  const usernameFor = useCallback((id: string): string => {
    const contact = contacts.find((c) => c.id === id);
    if (contact?.username) return contact.username;
    if (usernameCache.has(id)) return usernameCache.get(id)!;
    return `User ${id.slice(0, 8)}`;
  }, [contacts, usernameCache]);

  const groupNameFor = useCallback((groupId: string): string => {
    const group = groups.find((g) => g.id === groupId);
    if (group?.name) return group.name;
    return `Group ${groupId.slice(0, 8)}`;
  }, [groups]);

  const resolveUsername = useCallback((id: string) => {
    if (!id) return;
    if (contacts.some((c) => c.id === id && c.username)) return;
    if (usernameCache.has(id)) return;
    let active = true;
    getUserById(id)
      .then((u) => {
        if (active && u && u.username) {
          setUsernameCache((prev) => {
            if (prev.has(id)) return prev;
            const next = new Map(prev);
            next.set(id, u.username);
            return next;
          });
        }
      })
      .catch(() => {});
    return () => { active = false; };
  }, [contacts, usernameCache]);

  const clearSelection = useCallback(() => setSelectedPeerId(null), []);

  const addMessageRef = useRef(addMessage);
  addMessageRef.current = addMessage;
  const upsertConversationRef = useRef(upsertConversation);
  upsertConversationRef.current = upsertConversation;
  const usernameForRef = useRef(usernameFor);
  usernameForRef.current = usernameFor;
  const refreshRef = useRef(refresh);
  refreshRef.current = refresh;
  const resolveUsernameRef = useRef(resolveUsername);
  resolveUsernameRef.current = resolveUsername;

  useEffect(() => {
    const unsubscribe = service.onMessage(async (envelope: ServerEnvelope) => {
      if (envelope.type === "message") {
        const payload = envelope.payload as DirectMessage;
        const peerId = payload.sender_id ?? payload.recipient_id;
        if (payload.sender_id) resolveUsernameRef.current(payload.sender_id);

        let content = payload.content;
        try {
          const encrypted: EncryptedPayload = JSON.parse(content);
          if (encrypted.header && encrypted.ciphertext && encrypted.nonce) {
            content = await decryptMessage(
              payload.sender_id ?? "",
              encrypted.header,
              encrypted.ciphertext,
              encrypted.nonce,
            );
          }
        } catch {
          // Not encrypted or decryption failed, use as-is
        }

        addMessageRef.current({ id: uid(), peerId, senderId: payload.sender_id, content, direction: "in", timestamp: Date.now() }, "dm");
      }
      if (envelope.type === "group_message") {
        const payload = envelope.payload as GroupMessage;
        if (payload.sender_id) resolveUsernameRef.current(payload.sender_id);

        let content = payload.content;
        try {
          const encrypted: GroupEncryptedPayload = JSON.parse(content);
          if (encrypted.ciphertext && encrypted.nonce) {
            content = await decryptGroupMessage(
              payload.group_id,
              payload.sender_id ?? "",
              encrypted.ciphertext,
              encrypted.nonce,
              encrypted.key_id,
              encrypted.iteration,
            );
          }
        } catch {
          // Not encrypted or decryption failed, use as-is
        }

        addMessageRef.current({ id: uid(), peerId: payload.group_id, senderId: payload.sender_id, content, direction: "in", timestamp: Date.now() }, "group");
      }
      if (envelope.type === "sender_key_distribution") {
        const payload = envelope.payload as unknown as { group_id: string; sender_id: string; key_id: number; chain_key: string; iteration: number };
        try {
          await processSenderKeyDistribution(
            payload.group_id,
            payload.sender_id,
            payload.key_id,
            payload.chain_key,
            payload.iteration,
          );
        } catch {
          // Failed to process sender key
        }
      }
      if (envelope.type === "group_profile_picture_updated") {
        const payload = envelope.payload as { group_id: string; version: number };
        upsertConversationRef.current(payload.group_id, "group", "Group picture updated");
        setGroupPictureVersion((prev) => ({ ...prev, [payload.group_id]: (prev[payload.group_id] ?? 0) + 1 }));
        clearGroupPictureCache(payload.group_id);
      }
      if (envelope.type === "profile_picture_updated") {
        clearAllProfilePictures();
      }
      if (envelope.type === "member_joined") {
        const payload = envelope.payload as MemberJoinedPayload;
        const username = payload.username || usernameForRef.current(payload.user_id);
        addMessageRef.current({
          id: uid(),
          peerId: payload.group_id,
          senderId: undefined,
          content: `${username} joined the group`,
          direction: "in",
          timestamp: Date.now(),
          system: true,
        }, "group");
      }
      if (envelope.type === "chat_request_updated") {
        const payload = envelope.payload as { requester_id: string; recipient_id: string; status: string };
        if (payload.status === "accepted") {
          const otherId = payload.requester_id === user?.id ? payload.recipient_id : payload.requester_id;
          if (otherId) upsertConversationRef.current(otherId, "dm", "Chat request accepted");
          void refreshRef.current();
        } else if (payload.status === "pending" && payload.recipient_id === user?.id) {
          void refreshRef.current();
        }
      }
      if (envelope.type === "error") {
        setError("Message failed");
      }
    });
    return unsubscribe;
  }, [service, user]);

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
    clearSelection,
    sendDirectMessage,
    sendGroupMessage,
    startNewChat,
    acceptRequest,
    rejectRequest,
    createGroup,
    inviteToGroup: inviteToGroupMember,
    renameGroup,
    acceptGroupInvite,
    fetchGroupInfo,
    refresh,
    usernameFor,
    resolveUsername,
    groupNameFor,
    groupPictureVersion,
  }), [conversations, contacts, groups, messagesByPeer, pendingIncoming, selectedPeerId, selectedConversation, loading, error, offline, selectConversation, clearSelection, sendDirectMessage, sendGroupMessage, startNewChat, acceptRequest, rejectRequest, createGroup, inviteToGroupMember, renameGroup, acceptGroupInvite, fetchGroupInfo, refresh, usernameFor, resolveUsername, groupNameFor, groupPictureVersion]);

  return <ChatContext.Provider value={value}>{children}</ChatContext.Provider>;
}

export function useChat(): ChatContextValue {
  const ctx = useContext(ChatContext);
  if (!ctx) {
    throw new Error("useChat must be used within ChatProvider");
  }
  return ctx;
}
