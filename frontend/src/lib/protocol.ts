export type Snapshot = {
  documentId: string;
  version: number;
  content: string;
};

export type Operation = {
  kind: "insert" | "delete";
  offset: number;
  length: number;
  text: string;
};

export type DocumentRecord = {
  id: string;
};

export type Presence = {
  userId: string;
  name: string;
  color: string;
  anchor: number;
  head: number;
};

export type SnapshotMessage = {
  type: "snapshot";
  snapshot: Snapshot;
};

export type PresenceSnapshotMessage = {
  type: "presenceSnapshot";
  presences: Record<string, Presence>;
};

export type PresenceUpdateMessage = {
  type: "presenceUpdate";
  userId: string;
  presence: Presence;
};

export type ErrorMessage = {
  type: "error";
  error: string;
};

export type ClientMessage =
  | { type: "operation"; operation: Operation }
  | { type: "sync" }
  | { type: "presence"; presence: { anchor: number; head: number } };

export type ServerMessage =
  | SnapshotMessage
  | PresenceSnapshotMessage
  | PresenceUpdateMessage
  | ErrorMessage;

export function parseServerMessage(data: string): ServerMessage | null {
  try {
    return JSON.parse(data) as ServerMessage;
  } catch {
    return null;
  }
}
