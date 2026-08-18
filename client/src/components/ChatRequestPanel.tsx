import type { RelationshipResponse } from "../types/relationship";
import { AppAvatar } from "./AppAvatar";
import { shortId } from "../utils/format";

interface ChatRequestPanelProps {
  requests: RelationshipResponse[];
  onAccept: (requesterId: string) => Promise<void>;
  onReject: (requesterId: string) => Promise<void>;
  usernameFor?: (id: string) => string;
}

export function ChatRequestPanel({ requests, onAccept, onReject, usernameFor }: ChatRequestPanelProps) {
  if (requests.length === 0) {
    return <p className="muted page-hint">No pending chat requests</p>;
  }

  return (
    <div className="request-list">
      {requests.map((request) => {
        const name = usernameFor ? usernameFor(request.requester_id) : `User ${shortId(request.requester_id)}`;
        return (
          <div key={request.requester_id} className="request-item">
            <AppAvatar label={name} userId={request.requester_id} size={44} />
            <span className="request-main">
              <span className="request-name">{name}</span>
              <span className="request-subtitle">Wants to chat with you</span>
            </span>
            <span className="button-group">
              <button type="button" onClick={() => void onAccept(request.requester_id)}>
                Accept
              </button>
              <button type="button" className="secondary" onClick={() => void onReject(request.requester_id)}>
                Reject
              </button>
            </span>
          </div>
        );
      })}
    </div>
  );
}
