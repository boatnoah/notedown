import { useCallback, useEffect, useRef, type RefObject } from "react";

import { getWebSocketUrl } from "../../../lib/config";
import type { Operation, ServerMessage } from "../../../lib/protocol";
import { parseServerMessage } from "../../../lib/protocol";

type UseCollaborationSessionOptions = {
  documentId: string;
  initialVersion: number;
  socketRef: RefObject<WebSocket | null>;
  isApplyingRemoteRef: RefObject<boolean>;
  onSnapshot: (content: string, version: number) => void;
  onServerMessage: (msg: ServerMessage) => void;
  onConnectionLost: () => void;
};

export function useCollaborationSession({
  documentId,
  initialVersion,
  socketRef,
  isApplyingRemoteRef,
  onSnapshot,
  onServerMessage,
  onConnectionLost,
}: UseCollaborationSessionOptions) {
  const pendingOpsRef = useRef<Operation[]>([]);
  const latestVersionRef = useRef(initialVersion);

  const sendOperation = useCallback(
    (op: Operation) => {
      if (isApplyingRemoteRef.current) {
        return;
      }
      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        pendingOpsRef.current.push(op);
        return;
      }
      socket.send(JSON.stringify({ type: "operation", operation: op }));
    },
    [socketRef],
  );

  const handleServerMessage = useCallback(
    (msg: ServerMessage) => {
      if (msg.type === "snapshot") {
        if (msg.snapshot.version <= latestVersionRef.current) {
          return;
        }
        latestVersionRef.current = msg.snapshot.version;
        onSnapshot(msg.snapshot.content, msg.snapshot.version);
        return;
      }

      if (msg.type === "error") {
        console.error("Server error:", msg.error);
        return;
      }

      onServerMessage(msg);
    },
    [onSnapshot, onServerMessage],
  );

  useEffect(() => {
    latestVersionRef.current = initialVersion;
  }, [initialVersion]);

  useEffect(() => {
    const socket = new WebSocket(getWebSocketUrl(documentId));
    socketRef.current = socket;

    socket.addEventListener("open", () => {
      while (pendingOpsRef.current.length) {
        const next = pendingOpsRef.current.shift();
        if (next) {
          socket.send(JSON.stringify({ type: "operation", operation: next }));
        }
      }
      socket.send(JSON.stringify({ type: "sync" }));
    });

    socket.addEventListener("message", (event) => {
      const msg = parseServerMessage(event.data as string);
      if (!msg) {
        console.error("Invalid WebSocket message");
        return;
      }
      handleServerMessage(msg);
    });

    socket.addEventListener("close", onConnectionLost);
    socket.addEventListener("error", (err) => {
      console.error("WebSocket error:", err);
    });

    return () => {
      socket.close();
      socketRef.current = null;
    };
  }, [documentId, handleServerMessage, onConnectionLost, socketRef]);

  return { sendOperation };
}
