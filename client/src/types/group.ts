export interface GroupResponse {
  id: string;
  created_at: number;
}

export interface GroupMemberResponse {
  user_id: string;
  joined_at: number;
}

export interface GroupInviteResponse {
  group_id: string;
  user_id: string;
  invited_by: string;
  created_at: number;
}

export interface Group {
  id: string;
  members: string[];
}

export interface GroupProfilePictureResponse {
  ciphertext: string;
  nonce: string;
  version: number;
}
