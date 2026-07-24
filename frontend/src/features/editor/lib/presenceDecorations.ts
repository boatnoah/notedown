import { RangeSetBuilder, StateEffect, StateEffectType } from '@codemirror/state'
import { Decoration, DecorationSet, WidgetType } from '@codemirror/view'
import type { EditorView } from '@codemirror/view'

import type { Presence } from '../../../lib/protocol'

function clamp(value: number, min: number, max: number) {
  return Math.max(min, Math.min(max, value))
}

class CursorWidget extends WidgetType {
  constructor(
    private readonly color: string,
    private readonly name: string,
    private readonly username: string,
    private readonly pfp: string
  ) {
    super()
  }

  eq(other: CursorWidget) {
    return (
      other instanceof CursorWidget &&
      other.color === this.color &&
      other.name === this.name &&
      other.username === this.username &&
      other.pfp === this.pfp
    )
  }

  toDOM() {
    const pfpColors: Record<string, string> = {
      blue: '#60A5FA',
      green: '#4ADE80',
      red: '#F87171',
      yellow: '#FACC15',
      purple: '#A78BFA',
      orange: '#FB923C',
    }

    const wrapper = document.createElement('span')
    wrapper.style.display = 'inline-flex'
    wrapper.style.alignItems = 'center'
    wrapper.style.gap = '4px'
    wrapper.style.marginLeft = '-1px'

    const caret = document.createElement('span')
    caret.style.borderLeft = `2px solid ${this.color}`
    caret.style.height = '1em'
    caret.style.display = 'inline-block'
    wrapper.appendChild(caret)

    const label = document.createElement('span')
    label.style.display = 'inline-flex'
    label.style.alignItems = 'center'
    label.style.gap = '4px'
    label.style.padding = '1px 6px'
    label.style.borderRadius = '9999px'
    label.style.background = this.color
    label.style.color = '#fff'
    label.style.fontSize = '11px'
    label.style.lineHeight = '1.2'
    label.style.whiteSpace = 'nowrap'

    const avatar = document.createElement('span')
    avatar.style.width = '10px'
    avatar.style.height = '10px'
    avatar.style.borderRadius = '9999px'
    avatar.style.background = pfpColors[this.pfp] ?? '#9CA3AF'
    avatar.style.border = '1px solid rgba(255, 255, 255, 0.8)'
    label.appendChild(avatar)

    const text = document.createElement('span')
    text.textContent = this.name
    label.appendChild(text)

    wrapper.title = `${this.name} (@${this.username})`
    wrapper.appendChild(label)
    return wrapper
  }

  ignoreEvent() {
    return true
  }
}

type DecorationEntry = {
  from: number
  to: number
  value: Decoration
}

export function buildPresenceDecorations(
  docLength: number,
  presences: Map<string, Presence>
): DecorationSet {
  const entries: DecorationEntry[] = []

  const sorted = [...presences.entries()].sort(([, a], [, b]) => {
    const fromA = Math.min(a.anchor, a.head)
    const fromB = Math.min(b.anchor, b.head)
    return fromA - fromB
  })

  sorted.forEach(([, presence]) => {
    const anchor = clamp(presence.anchor, 0, docLength)
    const head = clamp(presence.head, 0, docLength)
    const from = Math.min(anchor, head)
    const to = Math.max(anchor, head)

    if (from !== to) {
      entries.push({
        from,
        to,
        value: Decoration.mark({
          attributes: { style: `background-color:${presence.color}20` },
        }),
      })
    }

    entries.push({
      from: to,
      to,
      value: Decoration.widget({
        widget: new CursorWidget(presence.color, presence.name, presence.username, presence.pfp),
        side: 1,
      }),
    })
  })

  entries.sort((a, b) => a.from - b.from || a.to - b.to)

  const builder = new RangeSetBuilder<Decoration>()
  for (const entry of entries) {
    builder.add(entry.from, entry.to, entry.value)
  }

  return builder.finish()
}

export function applyPresenceDecorations(
  view: EditorView,
  presences: Map<string, Presence>,
  effect: StateEffectType<DecorationSet>
) {
  const deco = buildPresenceDecorations(view.state.doc.length, presences)
  view.dispatch({ effects: effect.of(deco) })
}

export const setRemoteCursors = StateEffect.define<DecorationSet>()
