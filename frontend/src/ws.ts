import { marked } from "marked";
import { markdown } from "@codemirror/lang-markdown";
import {
  EditorSelection,
  EditorState,
  RangeSetBuilder,
  StateEffect,
  StateEffectType,
  StateField,
} from "@codemirror/state";
import {
  Decoration,
  DecorationSet,
  EditorView,
  WidgetType,
  keymap,
} from "@codemirror/view";
import { defaultKeymap, insertNewline } from "@codemirror/commands";

import { getBackendOrigin, getWebSocketUrl } from "./config";

type Snapshot = {
  documentId: string;
  version: number;
  content: string;
};

type Operation = {
  kind: "insert" | "delete";
  offset: number;
  length: number;
  text: string;
};

type DocumentRecord = {
  id: string;
};

type SnapshotMessage = {
  type: "snapshot";
  snapshot: Snapshot;
};

type Presence = {
  userId: string;
  name: string;
  color: string;
  anchor: number;
  head: number;
};

type PresenceSnapshotMessage = {
  type: "presenceSnapshot";
  presences: Record<string, Presence>;
};

type PresenceUpdateMessage = {
  type: "presenceUpdate";
  userId: string;
  presence: Presence;
};

export async function initEditor() {
  const root = document.getElementById("app");
  if (!root) {
    return;
  }

  root.innerHTML = "<p>Loading editor…</p>";

  try {
    const params = new URLSearchParams(window.location.search);
    let documentId = params.get("room") || params.get("documentId");

    if (!documentId) {
      const doc = await createDocument();
      documentId = doc.id;
      params.set("room", documentId);
      const nextUrl = `${window.location.pathname}?${params.toString()}`;
      window.history.replaceState({}, "", nextUrl);
    }

    if (!documentId) {
      throw new Error("Unable to determine document ID");
    }

    const snapshot = await fetchSnapshot(documentId);
    renderEditor(root, documentId, snapshot);
  } catch (error) {
    root.innerHTML = `<p class="error">Failed to load editor. ${
      error instanceof Error ? error.message : error
    }</p>`;
  }
}

function renderEditor(root: HTMLElement, documentId: string, snapshot: Snapshot) {
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

  const copyButton = document.getElementById("copy-btn");
  copyButton?.addEventListener("click", () => {
    navigator.clipboard
      .writeText(window.location.href)
      .then(() => alert("Link copied to clipboard!"))
      .catch(() => prompt("Copy this URL:", window.location.href));
  });

  const saveButton = document.getElementById("save-btn");
  saveButton?.addEventListener("click", () => {
    const url = window.location.href;
    if (navigator.clipboard) {
      navigator.clipboard
        .writeText(url)
        .then(() => alert("URL copied! Now paste into your bookmarks bar."))
        .catch(() => alert(`Here’s the URL:\n${url}`));
    } else {
      window.prompt("Copy this URL and press Ctrl+D (or ⌘+D) to bookmark:", url);
    }
  });

  const editorEl = document.getElementById("editor");
  const previewEl = document.getElementById("preview");
  if (!editorEl || !previewEl) {
    return;
  }

  const updatePreview = (markdownText: string) => {
    const rendered = marked.parse(markdownText);
    if (rendered instanceof Promise) {
      rendered.then((html) => {
        previewEl.innerHTML = html;
      });
    } else {
      previewEl.innerHTML = rendered;
    }
  };

  const pendingOps: Operation[] = [];
  let socket: WebSocket | null = null;
  let isApplyingRemote = false;
  let latestVersion = snapshot.version;
  const remotePresence = new Map<string, Presence>();

  const setRemoteCursors = StateEffect.define<DecorationSet>();
  const remoteCursorsField = StateField.define<DecorationSet>({
    create() {
      return Decoration.none;
    },
    update(value, tr) {
      for (const e of tr.effects) {
        if (e.is(setRemoteCursors)) {
          return e.value;
        }
      }
      return value.map(tr.changes);
    },
    provide: (f) => EditorView.decorations.from(f),
  });

  const sendOperation = (op: Operation) => {
    if (isApplyingRemote) {
      return;
    }
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      pendingOps.push(op);
      return;
    }
    socket.send(JSON.stringify({ type: "operation", operation: op }));
  };

  let presenceTimer: number | null = null;
  const sendPresence = (anchor: number, head: number) => {
    if (presenceTimer !== null) {
      window.clearTimeout(presenceTimer);
    }
    presenceTimer = window.setTimeout(() => {
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        return;
      }
      socket.send(
        JSON.stringify({
          type: "presence",
          presence: { anchor, head },
        }),
      );
      presenceTimer = null;
    }, 100);
  };

  const state = EditorState.create({
    doc: snapshot.content,
    extensions: [
      markdown(),
      keymap.of([...defaultKeymap, { key: "Enter", run: insertNewline }]),
      EditorView.lineWrapping,
      remoteCursorsField,
      EditorView.updateListener.of((update) => {
        if (update.docChanged && !isApplyingRemote) {
          update.changes.iterChanges((fromA, toA, _fromB, _toB, inserted) => {
            const deletedLength = toA - fromA;
            if (deletedLength > 0) {
              sendOperation({
                kind: "delete",
                offset: fromA,
                length: deletedLength,
                text: "",
              });
            }

            const insertedText = inserted.toString();
            if (insertedText.length > 0) {
              sendOperation({
                kind: "insert",
                offset: fromA,
                length: insertedText.length,
                text: insertedText,
              });
            }
          });
        }

        if (update.docChanged) {
          updatePreview(update.state.doc.toString());
        }

        if (update.selectionSet && !isApplyingRemote) {
          const sel = update.state.selection.main;
          sendPresence(sel.anchor, sel.head);
        }
      }),
    ],
  });

  const view = new EditorView({ state, parent: editorEl });
  updatePreview(snapshot.content);

  const downloadBtn = document.getElementById("download-btn");
  downloadBtn?.addEventListener("click", () => {
    const content = view.state.doc.toString();
    const blob = new Blob([content], { type: "text/markdown" });
    const downloadUrl = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = downloadUrl;
    a.download = `${documentId}.md`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(downloadUrl);
  });

  socket = new WebSocket(getWebSocketUrl(documentId));

  socket.addEventListener("open", () => {
    while (pendingOps.length) {
      const next = pendingOps.shift();
      if (next) {
        socket?.send(JSON.stringify({ type: "operation", operation: next }));
      }
    }
    socket?.send(JSON.stringify({ type: "sync" }));
  });

  socket.addEventListener("message", (event) => {
    try {
      const msg: SnapshotMessage = JSON.parse(event.data);
      if (msg.type === "snapshot") {
        if (msg.snapshot.version <= latestVersion) {
          return;
        }

        latestVersion = msg.snapshot.version;
        applySnapshot(view, msg.snapshot.content, updatePreview);
      }
      const pSnapshot = msg as unknown as PresenceSnapshotMessage;
      if (pSnapshot.type === "presenceSnapshot") {
        remotePresence.clear();
        Object.entries(pSnapshot.presences).forEach(([id, presence]) => {
          remotePresence.set(id, presence);
        });
        applyPresenceDecorations(view, remotePresence, setRemoteCursors);
      }

      const pUpdate = msg as unknown as PresenceUpdateMessage;
      if (pUpdate.type === "presenceUpdate") {
        if (!pUpdate.presence || (!pUpdate.presence.color && !pUpdate.presence.name)) {
          remotePresence.delete(pUpdate.userId);
        } else {
          remotePresence.set(pUpdate.userId, pUpdate.presence);
        }
        applyPresenceDecorations(view, remotePresence, setRemoteCursors);
      }

      if ((msg as any).type === "error") {
        console.error("Server error:", (msg as any).error);
      }
    } catch (err) {
      console.error("Invalid WebSocket message", err);
    }
  });

  socket.addEventListener("close", () => {
    alert("Connection lost—please refresh the page.");
  });

  socket.addEventListener("error", (err) => {
    console.error("WebSocket error:", err);
  });

  window.addEventListener("beforeunload", () => {
    socket?.close();
  });

  function applySnapshot(view: EditorView, content: string, cb: (text: string) => void) {
    if (content === view.state.doc.toString()) {
      return;
    }

    isApplyingRemote = true;
    const anchor = Math.min(
      view.state.selection.main.anchor,
      content.length,
    );

    view.dispatch({
      changes: { from: 0, to: view.state.doc.length, insert: content },
      selection: EditorSelection.cursor(anchor),
    });

    isApplyingRemote = false;
    cb(content);
  }

  applyPresenceDecorations(view, remotePresence, setRemoteCursors);
}

async function createDocument(): Promise<DocumentRecord> {
  const response = await fetch(`${getBackendOrigin()}/documents`, {
    method: "POST",
  });

  if (!response.ok) {
    const body = await safeText(response);
    throw new Error(
      `Failed to create document (${response.status}): ${body ?? ""}`,
    );
  }

  return response.json();
}

async function fetchSnapshot(documentId: string): Promise<Snapshot> {
  const response = await fetch(`${getBackendOrigin()}/documents/${documentId}`, {
    method: "GET",
  });

  if (!response.ok) {
    const body = await safeText(response);
    throw new Error(
      `Failed to fetch document (${response.status}): ${body ?? ""}`,
    );
  }

  return response.json();
}

function applyPresenceDecorations(
  view: EditorView,
  presences: Map<string, Presence>,
  effect: StateEffectType<DecorationSet>,
) {
  const deco = buildPresenceDecorations(view.state.doc.length, presences);
  view.dispatch({ effects: effect.of(deco) });
}

function buildPresenceDecorations(
  docLength: number,
  presences: Map<string, Presence>,
): DecorationSet {
  const builder = new RangeSetBuilder<Decoration>();

  presences.forEach((presence, userId) => {
    const anchor = clamp(presence.anchor, 0, docLength);
    const head = clamp(presence.head, 0, docLength);
    const from = Math.min(anchor, head);
    const to = Math.max(anchor, head);

    if (from !== to) {
      builder.add(
        from,
        to,
        Decoration.mark({
          attributes: { style: `background-color:${presence.color}20` },
        }),
      );
    }

    const caret = Decoration.widget({
      widget: new (class extends WidgetType {
        toDOM() {
          const el = document.createElement("span");
          el.style.borderLeft = `2px solid ${presence.color}`;
          el.style.marginLeft = "-1px";
          el.style.paddingLeft = "1px";
          el.style.height = "1em";
          el.style.display = "inline-block";
          el.title = presence.name || userId;
          return el;
        }
        ignoreEvent() {
          return true;
        }
      })(),
      side: 1,
    });

    builder.add(to, to, caret);
  });

  return builder.finish();
}

function clamp(value: number, min: number, max: number) {
  return Math.max(min, Math.min(max, value));
}

async function safeText(response: Response): Promise<string | null> {
  try {
    return await response.text();
  } catch {
    return null;
  }
}
