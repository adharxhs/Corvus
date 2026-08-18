export interface GroupResponse {
  id: string;
  name: string;
  created_at: number;
}

export interface GroupMemberResponse {
  user_id: string;
  joined_at: number;
}

export interface GroupInviteResponse {
  group_id: string;
  group_name: string;
  user_id: string;
  invited_by: string;
  created_at: number;
}

export interface Group {
  id: string;
  name: string;
  members: string[];
}

export interface GroupProfilePictureResponse {
  image_data: string;
  version: number;
}
