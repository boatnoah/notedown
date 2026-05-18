import { RangeSetBuilder, StateEffect, StateEffectType } from "@codemirror/state";
import { Decoration, DecorationSet, WidgetType } from "@codemirror/view";
import type { EditorView } from "@codemirror/view";

import type { Presence } from "../../../lib/protocol";

function clamp(value: number, min: number, max: number) {
  return Math.max(min, Math.min(max, value));
}

class CursorWidget extends WidgetType {
  constructor(
    private readonly color: string,
    private readonly label: string,
  ) {
    super();
  }

  eq(other: CursorWidget) {
    return (
      other instanceof CursorWidget &&
      other.color === this.color &&
      other.label === this.label
    );
  }

  toDOM() {
    const el = document.createElement("span");
    el.style.borderLeft = `2px solid ${this.color}`;
    el.style.marginLeft = "-1px";
    el.style.paddingLeft = "1px";
    el.style.height = "1em";
    el.style.display = "inline-block";
    el.title = this.label;
    return el;
  }

  ignoreEvent() {
    return true;
  }
}

type DecorationEntry = {
  from: number;
  to: number;
  value: Decoration;
};

export function buildPresenceDecorations(
  docLength: number,
  presences: Map<string, Presence>,
): DecorationSet {
  const entries: DecorationEntry[] = [];

  const sorted = [...presences.entries()].sort(([, a], [, b]) => {
    const fromA = Math.min(a.anchor, a.head);
    const fromB = Math.min(b.anchor, b.head);
    return fromA - fromB;
  });

  sorted.forEach(([userId, presence]) => {
    const anchor = clamp(presence.anchor, 0, docLength);
    const head = clamp(presence.head, 0, docLength);
    const from = Math.min(anchor, head);
    const to = Math.max(anchor, head);

    if (from !== to) {
      entries.push({
        from,
        to,
        value: Decoration.mark({
          attributes: { style: `background-color:${presence.color}20` },
        }),
      });
    }

    entries.push({
      from: to,
      to,
      value: Decoration.widget({
        widget: new CursorWidget(presence.color, presence.name || userId),
        side: 1,
      }),
    });
  });

  entries.sort((a, b) => a.from - b.from || a.to - b.to);

  const builder = new RangeSetBuilder<Decoration>();
  for (const entry of entries) {
    builder.add(entry.from, entry.to, entry.value);
  }

  return builder.finish();
}

export function applyPresenceDecorations(
  view: EditorView,
  presences: Map<string, Presence>,
  effect: StateEffectType<DecorationSet>,
) {
  const deco = buildPresenceDecorations(view.state.doc.length, presences);
  view.dispatch({ effects: effect.of(deco) });
}

export const setRemoteCursors = StateEffect.define<DecorationSet>();
