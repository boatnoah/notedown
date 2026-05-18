import { markdown } from '@codemirror/lang-markdown'
import { EditorSelection, EditorState, StateField } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, keymap } from '@codemirror/view'
import { defaultKeymap, insertNewline } from '@codemirror/commands'
import { useCallback, useEffect, useRef, useState } from 'react'

import type { Snapshot } from '../../../lib/protocol'
import { applyPresenceDecorations, setRemoteCursors } from '../lib/presenceDecorations'
import { useCollaborationSession } from '../hooks/useCollaborationSession'
import { usePresence } from '../hooks/usePresence'
import { MarkdownPreview } from './MarkdownPreview'
import { ShareBar } from './ShareBar'

type EditorProps = {
  documentId: string
  initialSnapshot: Snapshot
}

const remoteCursorsField = StateField.define<DecorationSet>({
  create() {
    return Decoration.none
  },
  update(value, tr) {
    for (const e of tr.effects) {
      if (e.is(setRemoteCursors)) {
        return e.value
      }
    }
    return value.map(tr.changes)
  },
  provide: (f) => EditorView.decorations.from(f),
})

export function Editor({ documentId, initialSnapshot }: EditorProps) {
  const editorRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)
  const socketRef = useRef<WebSocket | null>(null)
  const isApplyingRemoteRef = useRef(false)
  const [previewMarkdown, setPreviewMarkdown] = useState(initialSnapshot.content)

  const { remotePresence, sendPresence, handlePresenceMessage } = usePresence(socketRef)

  const applyRemoteSnapshot = useCallback((content: string) => {
    const view = viewRef.current
    if (!view || content === view.state.doc.toString()) {
      return
    }

    isApplyingRemoteRef.current = true
    const anchor = Math.min(view.state.selection.main.anchor, content.length)
    view.dispatch({
      changes: { from: 0, to: view.state.doc.length, insert: content },
      selection: EditorSelection.cursor(anchor),
    })
    isApplyingRemoteRef.current = false
    setPreviewMarkdown(content)
  }, [])

  const onConnectionLost = useCallback(() => {
    alert('Connection lost—please refresh the page.')
  }, [])

  const { sendOperation } = useCollaborationSession({
    documentId,
    initialVersion: initialSnapshot.version,
    socketRef,
    isApplyingRemoteRef,
    onSnapshot: applyRemoteSnapshot,
    onServerMessage: handlePresenceMessage,
    onConnectionLost,
  })

  useEffect(() => {
    const view = viewRef.current
    if (view) {
      applyPresenceDecorations(view, remotePresence, setRemoteCursors)
    }
  }, [remotePresence])

  useEffect(() => {
    const parent = editorRef.current
    if (!parent) {
      return
    }

    const state = EditorState.create({
      doc: initialSnapshot.content,
      extensions: [
        markdown(),
        keymap.of([...defaultKeymap, { key: 'Enter', run: insertNewline }]),
        EditorView.lineWrapping,
        remoteCursorsField,
        EditorView.updateListener.of((update) => {
          if (update.docChanged && !isApplyingRemoteRef.current) {
            update.changes.iterChanges((fromA, toA, _fromB, _toB, inserted) => {
              const deletedLength = toA - fromA
              if (deletedLength > 0) {
                sendOperation({
                  kind: 'delete',
                  offset: fromA,
                  length: deletedLength,
                  text: '',
                })
              }
              const insertedText = inserted.toString()
              if (insertedText.length > 0) {
                sendOperation({
                  kind: 'insert',
                  offset: fromA,
                  length: insertedText.length,
                  text: insertedText,
                })
              }
            })
          }

          if (update.docChanged) {
            setPreviewMarkdown(update.state.doc.toString())
          }

          if (update.selectionSet && !isApplyingRemoteRef.current) {
            const sel = update.state.selection.main
            sendPresence(sel.anchor, sel.head)
          }
        }),
      ],
    })

    const view = new EditorView({ state, parent })
    viewRef.current = view
    applyPresenceDecorations(view, remotePresence, setRemoteCursors)

    return () => {
      view.destroy()
      viewRef.current = null
    }
  }, [documentId])

  const downloadMarkdown = () => {
    const view = viewRef.current
    if (!view) {
      return
    }
    const content = view.state.doc.toString()
    const blob = new Blob([content], { type: 'text/markdown' })
    const downloadUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = downloadUrl
    a.download = `${documentId}.md`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(downloadUrl)
  }

  return (
    <>
      <ShareBar onDownload={downloadMarkdown} />
      <div className="editor-wrapper">
        <div className="editor-container">
          <div ref={editorRef} id="editor" />
        </div>
        <div className="preview-container">
          <MarkdownPreview markdown={previewMarkdown} />
        </div>
      </div>
    </>
  )
}
