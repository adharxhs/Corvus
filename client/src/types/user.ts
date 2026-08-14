export interface UserLookup {
  id: string;
  username?: string;
}

export interface Contact extends UserLookup {
  relationship: "pending" | "accepted" | "rejected" | "unknown";
  presence: "online" | "offline" | "unknown";
  profilePictureVersion?: number;
}
