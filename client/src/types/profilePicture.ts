export interface ProfilePictureResponse {
  image_data: string;
  version: number;
}

export interface EncryptedProfilePicture {
  ciphertext: string;
  nonce: string;
}
