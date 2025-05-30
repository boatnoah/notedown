import * as Y from "yjs";
import { fromUint8Array, toUint8Array } from "js-base64";
import { marked } from "marked";
import { EditorView } from "codemirror";
import { markdown } from "@codemirror/lang-markdown";
import { EditorState } from "@codemirror/state";
import * as awarenessProtocol from "y-protocols/awareness.js";
import { yCollab } from "y-codemirror.next";
import { keymap } from "@codemirror/view";
import { insertNewline, defaultKeymap } from "@codemirror/commands";

type AwarenessChange = {
  added: number[];
  updated: number[];
  removed: number[];
};

export function initEditor() {
  const root = document.getElementById("app")!;

  // 1) Get or generate room ID & update URL
  const params = new URLSearchParams(window.location.search);
  let room = params.get("room");
  if (!room) {
    room = crypto.randomUUID();
    params.set("room", room);
    window.history.replaceState(
      {},
      "",
      `${window.location.pathname}?${params.toString()}`,
    );
  }

  // 2) Render Share Link UI + Editor/Preview
  root.innerHTML = `
    <div class="share-container">
      <label for="share-link">Share this link:</label>
      <div class="share-controls">
        <input id="share-link" type="text" readonly value="${window.location.href}" />
        <button id="copy-btn">Copy</button>
        <button id="save-btn">Save Link</button>
        <button id="download-btn">Save to Machine</button>
      </div>
    </div>
    <div class="editor-wrapper">
      <div class="editor-container">
        <div id="editor"></div>
      </div>
      <div class="preview-container">
        <div id="preview"></div>
      </div>
    </div>
  `;

  // 3) Wire up Copy button
  document.getElementById("copy-btn")!.addEventListener("click", () => {
    navigator.clipboard
      .writeText(window.location.href)
      .then(() => alert("Link copied to clipboard!"))
      .catch(() => prompt("Copy this URL:", window.location.href));
  });

  // 4) Wire up Save Link button (bookmark helper)
  document.getElementById("save-btn")!.addEventListener("click", () => {
    const url = window.location.href;
    if (navigator.clipboard) {
      navigator.clipboard
        .writeText(url)
        .then(() => alert("URL copied! Now paste into your bookmarks bar."))
        .catch(() => alert(`Here’s the URL:\n${url}`));
    } else {
      window.prompt(
        "Copy this URL and press Ctrl+D (or ⌘+D) to bookmark:",
        url,
      );
    }
  });

  // src/ws.ts
  const apiOrigin = import.meta.env.VITE_WS_URL || window.location.origin;
  const wsScheme = apiOrigin.startsWith("https") ? "wss" : "ws";
  // strip off any path, leaving just host:port
  const host = new URL(apiOrigin).host;
  const socket = new WebSocket(`${wsScheme}://${host}/ws?room=${room}`);

  const ydoc = new Y.Doc();
  const yText = ydoc.getText("markdown");
  const awareness = new awarenessProtocol.Awareness(ydoc);

  const editorEl = document.getElementById("editor")!;
  const previewEl = document.getElementById("preview")!;

  const updatePreview = async (markdownText: string) => {
    previewEl.innerHTML = await marked.parse(markdownText);
  };

  const state = EditorState.create({
    extensions: [
      markdown(),
      keymap.of([...defaultKeymap, { key: "Enter", run: insertNewline }]),
      EditorView.lineWrapping,
      yCollab(yText, awareness, { undoManager: false }),
      EditorView.updateListener.of((update) => {
        if (update.docChanged) {
          updatePreview(update.state.doc.toString());
        }
      }),
    ],
  });

  // create editor view reference for download
  const view = new EditorView({ state, parent: editorEl });

  // 6) Wire up Save to Machine button (download markdown)
  document.getElementById("download-btn")!.addEventListener("click", () => {
    const content = view.state.doc.toString();
    const blob = new Blob([content], { type: "text/markdown" });
    const downloadUrl = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = downloadUrl;
    a.download = `${room}.md`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(downloadUrl);
  });

  // Broadcast document updates
  ydoc.on("update", () => {
    const encoded = fromUint8Array(Y.encodeStateAsUpdate(ydoc));
    socket.send(JSON.stringify({ type: "docContent", content: encoded }));
  });

  // Broadcast awareness changes
  awareness.on("update", ({ added, updated, removed }: AwarenessChange) => {
    const clients = added.concat(updated).concat(removed);
    const encoded = fromUint8Array(
      awarenessProtocol.encodeAwarenessUpdate(awareness, clients),
    );
    socket.send(JSON.stringify({ type: "awarenessContent", content: encoded }));
  });

  socket.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    if (msg.type === "docContent") {
      Y.applyUpdate(ydoc, toUint8Array(msg.content));
    }
    if (msg.type === "awarenessContent") {
      awarenessProtocol.applyAwarenessUpdate(
        awareness,
        toUint8Array(msg.content),
        awareness.clientID,
      );
    }
  };

  socket.onclose = () => {
    console.log("WebSocket closed");
    alert("Connection lost—please refresh the page.");
  };

  socket.onerror = (err) => console.error("WebSocket error:", err);
}
