import type { GroupInviteResponse, GroupMemberResponse, GroupProfilePictureResponse, GroupResponse } from "../types/group";
import { apiRequest } from "./api";

export function createGroup(groupId: string, name: string): Promise<GroupResponse> {
  return apiRequest<GroupResponse>("/groups", {
    method: "POST",
    body: JSON.stringify({ group_id: groupId, name }),
  });
}

export function listMyGroups(): Promise<GroupResponse[]> {
  return apiRequest<GroupResponse[]>("/groups/mine");
}

export function getGroup(groupId: string): Promise<GroupResponse> {
  return apiRequest<GroupResponse>(`/groups/${encodeURIComponent(groupId)}`);
}

export function renameGroup(groupId: string, name: string): Promise<void> {
  return apiRequest<void>(`/groups/${encodeURIComponent(groupId)}/name`, {
    method: "PUT",
    body: JSON.stringify({ name }),
  });
}

export function listGroupInvites(): Promise<GroupInviteResponse[]> {
  return apiRequest<GroupInviteResponse[]>("/groups/invites");
}

export function listGroupMembers(groupId: string): Promise<GroupMemberResponse[]> {
  return apiRequest<GroupMemberResponse[]>(`/groups/${encodeURIComponent(groupId)}/members`);
}

export function inviteToGroup(groupId: string, userId: string): Promise<void> {
  return apiRequest<void>(`/groups/${encodeURIComponent(groupId)}/invite`, {
    method: "POST",
    body: JSON.stringify({ user_id: userId }),
  });
}

export function acceptGroupInvite(groupId: string): Promise<{ status: string }> {
  return apiRequest<{ status: string }>(`/groups/${encodeURIComponent(groupId)}/invite/accept`, { method: "POST" });
}

export function rejectGroupInvite(groupId: string): Promise<void> {
  return apiRequest<void>(`/groups/${encodeURIComponent(groupId)}/invite/reject`, { method: "POST" });
}

export function leaveGroup(groupId: string): Promise<void> {
  return apiRequest<void>(`/groups/${encodeURIComponent(groupId)}/member`, { method: "DELETE" });
}

export function uploadGroupProfilePicture(groupId: string, payload: { image_data: string; version: number }): Promise<void> {
  return apiRequest<void>(`/groups/${encodeURIComponent(groupId)}/profile-picture`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function getGroupProfilePicture(groupId: string): Promise<GroupProfilePictureResponse> {
  return apiRequest<GroupProfilePictureResponse>(`/groups/${encodeURIComponent(groupId)}/profile-picture`);
}
