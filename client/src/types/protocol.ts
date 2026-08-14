export interface ProtocolEnvelope<TPayload = unknown> {
  version: 1;
  type: string;
  payload: TPayload;
}

export interface ProtocolErrorPayload {
  code: string;
  message: string;
}
