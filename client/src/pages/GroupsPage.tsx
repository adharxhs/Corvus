import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useChat } from "../contexts/ChatContext";
import { TitleBar, BottomNav } from "../components/NavHeader";
import { SearchBar } from "../components/SearchBar";
import { GroupList } from "../components/GroupList";
import { GroupInvitePanel } from "../components/GroupInvitePanel";
import { FilterTabs } from "../components/FilterTabs";
import { Fab } from "../components/Fab";
import { BottomSheet } from "../components/BottomSheet";
import { EmptyState } from "../components/EmptyState";
import { uid } from "../utils/random";
import { listGroupInvites, rejectGroupInvite as apiRejectGroupInvite } from "../services/groups";
import type { GroupInviteResponse } from "../types/group";

type GroupTab = "all" | "unread" | "requests";

export function GroupsPage() {
  const { groups, conversations, createGroup, selectConversation, acceptGroupInvite, usernameFor, groupNameFor, groupPictureVersion } = useChat();
  const navigate = useNavigate();
  const [query, setQuery] = useState("");
  const [tab, setTab] = useState<GroupTab>("all");
  const [createOpen, setCreateOpen] = useState(false);
  const [newGroupName, setNewGroupName] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [groupInvites, setGroupInvites] = useState<GroupInviteResponse[]>([]);

  useEffect(() => {
    let active = true;
    listGroupInvites()
      .then((invites) => { if (active) setGroupInvites(invites); })
      .catch(() => {});
    return () => { active = false; };
  }, []);

  const filtered = query
    ? groups.filter((g) => (g.name || g.id).toLowerCase().includes(query.toLowerCase()))
    : groups;

  const unreadCount = groups.reduce(
    (sum, g) => sum + (conversations.find((c) => c.peerId === g.id)?.unreadCount ?? 0),
    0,
  );

  const visibleGroups = tab === "unread" ? filtered.filter((g) => (conversations.find((c) => c.peerId === g.id)?.unreadCount ?? 0) > 0) : filtered;

  async function handleCreateGroup() {
    const groupId = uid().replace(/-/g, "").slice(0, 20);
    const name = newGroupName.trim() || "New Group";
    setBusy(true);
    setError(null);
    try {
      await createGroup(groupId, name);
      setCreateOpen(false);
      setNewGroupName("");
      selectConversation(groupId);
      navigate(`/chat/group/${groupId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create group");
    } finally {
      setBusy(false);
    }
  }

  async function handleAcceptInvite(groupId: string) {
    try {
      await acceptGroupInvite(groupId);
      setGroupInvites((prev) => prev.filter((i) => i.group_id !== groupId));
    } catch {
      // ignore
    }
  }

  async function handleRejectInvite(groupId: string) {
    try {
      await apiRejectGroupInvite(groupId);
      setGroupInvites((prev) => prev.filter((i) => i.group_id !== groupId));
    } catch {
      // ignore
    }
  }

  return (
    <div className="page groups-page">
      <TitleBar />
      <div className="chats-search-wrapper">
        <SearchBar query={query} onChange={setQuery} placeholder="Search groups" />
        <FilterTabs
          active={tab}
          onChange={setTab}
          options={[
            { key: "all", label: "All" },
            { key: "unread", label: "Unread", count: unreadCount },
            { key: "requests", label: "Requests", count: groupInvites.length },
          ]}
        />
      </div>
      {tab === "requests" ? (
        <GroupInvitePanel invites={groupInvites} onAccept={handleAcceptInvite} onReject={handleRejectInvite} usernameFor={usernameFor} groupNameFor={groupNameFor} />
      ) : (
        <>
          {visibleGroups.length === 0 && (
            <EmptyState
              title={tab === "unread" ? "No unread groups" : "No groups yet"}
              body={tab === "unread" ? "You're all caught up." : "Tap the pencil to create a group."}
            />
          )}
          {visibleGroups.length > 0 && (
            <GroupList
              groups={visibleGroups}
              conversations={conversations}
              onCreateGroup={() => setCreateOpen(true)}
              onOpenGroup={(group) => {
                selectConversation(group.id);
                navigate(`/chat/group/${group.id}`);
              }}
              groupPictureVersion={groupPictureVersion}
            />
          )}
        </>
      )}
      <Fab label="New group" onClick={() => setCreateOpen(true)} />
      <BottomSheet open={createOpen} title="Create group" onClose={() => { setCreateOpen(false); setError(null); setNewGroupName(""); }}>
        <div className="sheet-form">
          <input
            type="text"
            placeholder="Group name"
            maxLength={60}
            value={newGroupName}
            onChange={(e) => setNewGroupName(e.target.value)}
          />
          {error && <p className="error-text">{error}</p>}
          <button type="button" className="primary-button" disabled={busy} onClick={() => void handleCreateGroup()}>
            {busy ? "Creating…" : "Create group"}
          </button>
        </div>
      </BottomSheet>
      <BottomNav />
    </div>
  );
}
