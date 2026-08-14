import { PROFILE_PICTURE_MAX_BYTES } from "./constants";

export function validateUsername(username: string): string | null {
  const trimmed = username.trim();
  if (trimmed.length < 3) {
    return "Username must be at least 3 characters";
  }
  if (trimmed.length > 64) {
    return "Username must be at most 64 characters";
  }
  return null;
}

export function validatePassword(password: string): string | null {
  if (password.length < 8) {
    return "Password must be at least 8 characters";
  }
  if (password.length > 256) {
    return "Password must be at most 256 characters";
  }
  return null;
}

export function validateProfilePicture(file: File): string | null {
  if (!file.type.startsWith("image/")) {
    return "Choose an image file";
  }
  if (file.size > PROFILE_PICTURE_MAX_BYTES) {
    return "Profile picture must be 2 MB or smaller";
  }
  return null;
}
