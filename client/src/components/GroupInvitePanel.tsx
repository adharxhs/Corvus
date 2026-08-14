import type { GroupInviteResponse } from "../types/group";

interface GroupInvitePanelProps {
  invites: GroupInviteResponse[];
  onAccept: (groupId: string) => Promise<void>;
}

export function GroupInvitePanel({ invites, onAccept }: GroupInvitePanelProps) {
  if (invites.length === 0) {
    return null;
  }

  return (
    <section className="panel">
      <h2>Group Invites</h2>
      <div className="list">
        {invites.map((invite) => (
          <div key={`${invite.group_id}-${invite.invited_by}`} className="list-item">
            <span>{invite.group_id}</span>
            <button type="button" onClick={() => onAccept(invite.group_id)}>
              Accept
            </button>
          </div>
        ))}
      </div>
    </section>
  );
}
