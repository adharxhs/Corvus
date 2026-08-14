import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useChat } from "../contexts/ChatContext";
import { TitleBar, BottomNav } from "../components/NavHeader";
import { SearchBar } from "../components/SearchBar";
import { GroupList } from "../components/GroupList";
import { Fab } from "../components/Fab";
import { BottomSheet } from "../components/BottomSheet";
import { EmptyState } from "../components/EmptyState";
import { uid } from "../utils/random";

export function GroupsPage() {
  const { groups, conversations, createGroup, selectConversation } = useChat();
  const navigate = useNavigate();
  const [query, setQuery] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const filtered = query
    ? groups.filter((g) => g.id.toLowerCase().includes(query.toLowerCase()))
    : groups;

  async function handleCreateGroup() {
    const groupId = uid().replace(/-/g, "").slice(0, 20);
    setBusy(true);
    setError(null);
    try {
      await createGroup(groupId);
      setCreateOpen(false);
      selectConversation(groupId);
      navigate(`/chat/group/${groupId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create group");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="page groups-page">
      <TitleBar />
      <div className="chats-search-wrapper">
        <SearchBar query={query} onChange={setQuery} placeholder="Search groups" />
      </div>
      {filtered.length === 0 && (
        <EmptyState title="No groups yet" body="Tap the pencil to create a group." />
      )}
      <GroupList
        groups={filtered}
        conversations={conversations}
        onCreateGroup={() => setCreateOpen(true)}
        onOpenGroup={(group) => {
          selectConversation(group.id);
          navigate(`/chat/group/${group.id}`);
        }}
      />
      <Fab label="New group" onClick={() => setCreateOpen(true)} />
      <BottomSheet open={createOpen} title="Create group" onClose={() => { setCreateOpen(false); setError(null); }}>
        <div className="sheet-form">
          <p className="sheet-hint">Group ID: generated automatically</p>
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
