import type { ProfilePictureResponse } from "../types/profilePicture";
import { apiRequest } from "./api";

export function getProfilePicture(userId: string): Promise<ProfilePictureResponse> {
  return apiRequest<ProfilePictureResponse>(`/profile-picture/${encodeURIComponent(userId)}`);
}

export function uploadProfilePicture(payload: { image_data: string; version: number }): Promise<void> {
  return apiRequest<void>("/profile-picture", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function compressImage(file: File, maxWidth = 400, maxHeight = 400, quality = 0.8): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      const img = new Image();
      img.onload = () => {
        const canvas = document.createElement("canvas");
        let { width, height } = img;

        if (width > maxWidth || height > maxHeight) {
          const aspectRatio = width / height;
          if (width > height) {
            width = maxWidth;
            height = width / aspectRatio;
          } else {
            height = maxHeight;
            width = height * aspectRatio;
          }
        }

        canvas.width = width;
        canvas.height = height;
        const ctx = canvas.getContext("2d");
        if (!ctx) {
          reject(new Error("Canvas context not available"));
          return;
        }
        ctx.drawImage(img, 0, 0, width, height);
        const base64 = canvas.toDataURL("image/jpeg", quality).split(",")[1];
        resolve(base64);
      };
      img.onerror = () => reject(new Error("Failed to load image"));
      img.src = e.target?.result as string;
    };
    reader.onerror = () => reject(new Error("Failed to read file"));
    reader.readAsDataURL(file);
  });
}
