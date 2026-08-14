export interface DirectMessage {
  recipient_id: string;
  sender_id?: string;
  content: string;
  timestamp?: number;
}

export interface GroupMessage {
  group_id: string;
  sender_id?: string;
  content: string;
  timestamp?: number;
}

export interface SenderKeyDistribution {
  group_id: string;
  content: string;
}
