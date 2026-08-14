import type { LoginResponse, RegisterResponse } from "../types/auth";
import { apiRequest } from "./api";

export function loginRequest(username: string, password: string): Promise<LoginResponse> {
  return apiRequest<LoginResponse>("/login", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  });
}

export function registerRequest(username: string, password: string): Promise<RegisterResponse> {
  return apiRequest<RegisterResponse>("/register", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  });
}

export function changePasswordRequest(currentPassword: string, newPassword: string): Promise<{ status: string }> {
  return apiRequest<{ status: string }>("/user/password", {
    method: "POST",
    body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
  });
}
