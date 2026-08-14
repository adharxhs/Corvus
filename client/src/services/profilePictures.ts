import type { ProfilePictureResponse } from "../types/profilePicture";
import { apiRequest } from "./api";

export function getProfilePicture(userId: string): Promise<ProfilePictureResponse> {
  return apiRequest<ProfilePictureResponse>(`/profile-picture/${encodeURIComponent(userId)}`);
}

export function uploadProfilePicture(payload: { ciphertext: string; nonce: string; version: number }): Promise<void> {
  return apiRequest<void>("/profile-picture", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}
