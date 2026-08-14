import type { RelationshipResponse } from "../types/relationship";

interface ChatRequestPanelProps {
  requests: RelationshipResponse[];
  onAccept: (requesterId: string) => Promise<void>;
  onReject: (requesterId: string) => Promise<void>;
}

export function ChatRequestPanel({ requests, onAccept, onReject }: ChatRequestPanelProps) {
  if (requests.length === 0) {
    return null;
  }

  return (
    <section className="panel">
      <h2>Chat Requests</h2>
      <div className="list">
        {requests.map((request) => (
          <div key={request.requester_id} className="list-item">
            <span>{request.requester_id}</span>
            <span className="button-group">
              <button type="button" onClick={() => onAccept(request.requester_id)}>
                Accept
              </button>
              <button type="button" className="secondary" onClick={() => onReject(request.requester_id)}>
                Reject
              </button>
            </span>
          </div>
        ))}
      </div>
    </section>
  );
}
