import { RangeSetBuilder, StateEffect, StateEffectType } from "@codemirror/state";
import { Decoration, DecorationSet, WidgetType } from "@codemirror/view";
import type { EditorView } from "@codemirror/view";

import type { Presence } from "../../../lib/protocol";

function clamp(value: number, min: number, max: number) {
  return Math.max(min, Math.min(max, value));
}

export function buildPresenceDecorations(
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

export function applyPresenceDecorations(
  view: EditorView,
  presences: Map<string, Presence>,
  effect: StateEffectType<DecorationSet>,
) {
  const deco = buildPresenceDecorations(view.state.doc.length, presences);
  view.dispatch({ effects: effect.of(deco) });
}

export const setRemoteCursors = StateEffect.define<DecorationSet>();
