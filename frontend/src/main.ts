import * as Y from "yjs";
import { fromUint8Array, toUint8Array } from "js-base64";
import { setupEditor } from "./editor";
import { marked } from "marked";

const socket = new WebSocket("ws://localhost:3000/ws");

const ydoc = new Y.Doc();
const yText = ydoc.getText("markdown");

const editorEl = document.getElementById("editor")!;
const previewEl = document.getElementById("preview")!;

const updatePreview = async (markdownText: string) => {
  previewEl.innerHTML = await marked.parse(markdownText);
};

setupEditor(yText, editorEl, updatePreview);

ydoc.on("update", (update, origin) => {
  const documentState = Y.encodeStateAsUpdate(ydoc); // is a Uint8Array
  const base64Encoded = fromUint8Array(documentState);

  socket.send(base64Encoded);
});

socket.onopen = () => {
  console.log("Connected to the WebSocket server.");
};

socket.onmessage = (event) => {
  const binaryEncoded = toUint8Array(event.data);

  Y.applyUpdate(ydoc, binaryEncoded);
};

socket.onclose = () => {
  console.log("WebSocket connection closed.");
};

socket.onerror = (error) => {
  console.error("WebSocket error:", error);
};
