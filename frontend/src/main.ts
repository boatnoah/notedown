import * as Y from "yjs";
import { WebsocketProvider } from "y-websocket";
import { setupEditor } from "./editor";
import { marked } from "marked";

const socket = new WebSocket("ws://localhost:3000/ws");

const ydoc = new Y.Doc();
const yText = ydoc.getText("markdown");
const provider = new WebsocketProvider("ws://localhost:3000", "ws", ydoc);

const editorEl = document.getElementById("editor")!;
const previewEl = document.getElementById("preview")!;

const updatePreview = (markdownText: string) => {
  previewEl.innerHTML = marked.parse(markdownText);
};

setupEditor(yText, editorEl, updatePreview);

socket.onopen = () => {
  console.log("Connected to the WebSocket server.");
};

socket.onmessage = (event) => {
  console.log(event);
};

socket.onclose = () => {
  console.log("WebSocket connection closed.");
  editorEl.blur();
};

socket.onerror = (error) => {
  console.error("WebSocket error:", error);
};
