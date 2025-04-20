import { EditorView } from "codemirror";
import { markdown } from "@codemirror/lang-markdown";
import { EditorState } from "@codemirror/state";
import { keymap } from "@codemirror/view";
import { insertNewline, defaultKeymap } from "@codemirror/commands";

export const setupEditor = (
  yText: any, // from yjs
  parent: HTMLElement,
  onUpdate: (markdown: string) => void,
) => {
  const state = EditorState.create({
    extensions: [
      markdown(),
      keymap.of([...defaultKeymap, { key: "Enter", run: insertNewline }]),
      EditorView.lineWrapping,
      EditorView.updateListener.of((update) => {
        if (update.docChanged) {
          const text = update.state.doc.toString();
          onUpdate(text);
        }
      }),
    ],
  });

  return new EditorView({
    state,
    parent,
  });
};
