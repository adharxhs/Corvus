export interface ProfilePictureResponse {
  ciphertext: string;
  nonce: string;
  version: number;
}

export interface EncryptedProfilePicture {
  ciphertext: string;
  nonce: string;
}

export interface CachedProfilePicture {
  userId: string;
  version: number;
  url: string;
}
