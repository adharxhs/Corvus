export interface RelationshipResponse {
  requester_id: string;
  recipient_id: string;
  status: "pending" | "accepted" | "rejected";
  created_at: number;
  updated_at: number;
}
