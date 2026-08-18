import type { GroupInviteResponse } from "../types/group";
import { AppAvatar } from "./AppAvatar";
import { GroupAvatar } from "./GroupAvatar";
import { shortId } from "../utils/format";

interface GroupInvitePanelProps {
  invites: GroupInviteResponse[];
  onAccept: (groupId: string) => Promise<void>;
  onReject: (groupId: string) => Promise<void>;
  usernameFor?: (id: string) => string;
  groupNameFor?: (groupId: string) => string;
}

export function GroupInvitePanel({ invites, onAccept, onReject, usernameFor, groupNameFor }: GroupInvitePanelProps) {
  if (invites.length === 0) {
    return <p className="muted page-hint">No pending group invites</p>;
  }

  return (
    <div className="request-list">
      {invites.map((invite) => {
        const inviter = usernameFor ? usernameFor(invite.invited_by) : `User ${shortId(invite.invited_by)}`;
        const groupName = invite.group_name || (groupNameFor ? groupNameFor(invite.group_id) : `Group ${shortId(invite.group_id)}`);
        return (
          <div key={`${invite.group_id}-${invite.invited_by}`} className="group-invite-item">
            <div className="group-invite-group">
              <GroupAvatar name={groupName} groupId={invite.group_id} size={44} />
              <span className="group-invite-group-name">{groupName}</span>
            </div>
            <div className="group-invite-inviter">
              <AppAvatar label={inviter} userId={invite.invited_by} size={32} />
              <span className="group-invite-inviter-text">
                <span className="group-invite-inviter-name">{inviter}</span>
                <span className="group-invite-inviter-action">invited you</span>
              </span>
            </div>
            <div className="group-invite-actions">
              <button type="button" onClick={() => void onAccept(invite.group_id)}>
                Accept
              </button>
              <button type="button" className="secondary" onClick={() => void onReject(invite.group_id)}>
                Decline
              </button>
            </div>
          </div>
        );
      })}
    </div>
  );
}
