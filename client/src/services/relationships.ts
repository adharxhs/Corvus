import type { RelationshipResponse } from "../types/relationship";
import { apiRequest } from "./api";

export function sendChatRequest(recipientId: string): Promise<RelationshipResponse> {
  return apiRequest<RelationshipResponse>("/chat-request", {
    method: "POST",
    body: JSON.stringify({ recipient_id: recipientId }),
  });
}

export function listChatRequests(): Promise<RelationshipResponse[]> {
  return apiRequest<RelationshipResponse[]>("/chat-requests");
}

export function acceptChatRequest(requesterId: string): Promise<RelationshipResponse> {
  return apiRequest<RelationshipResponse>(`/chat-request/${encodeURIComponent(requesterId)}/accept`, { method: "POST" });
}

export function rejectChatRequest(requesterId: string): Promise<RelationshipResponse> {
  return apiRequest<RelationshipResponse>(`/chat-request/${encodeURIComponent(requesterId)}/reject`, { method: "POST" });
}
