import type { User } from "../types/auth";
import { apiRequest } from "./api";

export function lookupUserByUsername(username: string): Promise<{ id: string }> {
  return apiRequest<{ id: string }>(`/users/by-username/${encodeURIComponent(username)}`);
}

export function getUserById(id: string): Promise<User> {
  return apiRequest<User>(`/users/${encodeURIComponent(id)}`);
}
