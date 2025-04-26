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
import * as random from "lib0/random";

// src/editor.ts

export function initEditor() {
  const root = document.getElementById("app")!;
  root.innerHTML = `
    <div class="editor-wrapper">
      <div class="editor-container">
        <div id="editor"></div>
      </div>
      <div class="preview-container">
        <div id="preview"></div>
      </div>
    </div> 
  `;

  const protocol = location.protocol === "https:" ? "wss" : "ws";
  const url = `${protocol}://${location.host}/ws`;
  const socket = new WebSocket(url);

  const ydoc = new Y.Doc();
  const yText = ydoc.getText("markdown");
  const awareness = new awarenessProtocol.Awareness(ydoc);

  const usercolors = [
    { color: "#30bced", light: "#30bced33" },
    { color: "#6eeb83", light: "#6eeb8333" },
    { color: "#ffbc42", light: "#ffbc4233" },
    { color: "#ecd444", light: "#ecd44433" },
    { color: "#ee6352", light: "#ee635233" },
    { color: "#9ac2c9", light: "#9ac2c933" },
    { color: "#8acb88", light: "#8acb8833" },
    { color: "#1be7ff", light: "#1be7ff33" },
  ];

  const userColor = usercolors[random.uint32() % usercolors.length];

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
          const text = update.state.doc.toString();
          updatePreview(text);
        }
      }),
    ],
  });

  const view = new EditorView({
    state,
    parent: editorEl,
  });

  ydoc.on("update", () => {
    const documentState = Y.encodeStateAsUpdate(ydoc); // is a Uint8Array
    const base64Encoded = fromUint8Array(documentState);

    const doc = {
      type: "docContent",
      content: base64Encoded,
    };

    socket.send(JSON.stringify(doc));
  });

  awareness.on("update", ({ added, updated, removed }) => {
    const changedClients = added.concat(updated).concat(removed);
    const documentState = awarenessProtocol.encodeAwarenessUpdate(
      awareness,
      changedClients,
    ); // is a Uint8Array
    const base64Encoded = fromUint8Array(documentState);

    const aw = {
      type: "awarenessContent",
      content: base64Encoded,
    };

    socket.send(JSON.stringify(aw));
  });
  socket.onopen = () => {
    console.log("Connected to the WebSocket server.");
  };

  socket.onmessage = (event) => {
    const metaData = JSON.parse(event.data);

    if (metaData.type === "docContent") {
      const binaryEncoded = toUint8Array(metaData.content);
      Y.applyUpdate(ydoc, binaryEncoded);
    }

    if (metaData.type === "awarenessContent") {
      const binaryEncoded = toUint8Array(metaData.content);
      awarenessProtocol.applyAwarenessUpdate(
        awareness,
        binaryEncoded,
        awareness.clientID,
      );
    }
  };

  socket.onclose = () => {
    console.log("WebSocket connection closed.");
    EditorState.readOnly.of(true);
    alert("refresh please");
  };

  socket.onerror = (error) => {
    console.error("WebSocket error:", error);
  };
}
