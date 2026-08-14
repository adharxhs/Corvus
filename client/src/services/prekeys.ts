export interface PrekeyBundleRequest {
  identity_key: string;
  signed_prekey: string;
  signed_prekey_signature: string;
  one_time_prekey?: string;
}

export interface PrekeyBundleResponse {
  user_id: string;
  identity_key: string;
  signed_prekey: string;
  signed_prekey_signature: string;
  one_time_prekey?: string;
}

export function uploadPrekeyBundle(bundle: PrekeyBundleRequest): Promise<void> {
  return apiRequest<void>("/prekey", {
    method: "POST",
    body: JSON.stringify(bundle),
  });
}

export function fetchPrekeyBundle(userId: string): Promise<PrekeyBundleResponse> {
  return apiRequest<PrekeyBundleResponse>(`/prekey/${encodeURIComponent(userId)}`);
}

import { apiRequest } from "./api";
