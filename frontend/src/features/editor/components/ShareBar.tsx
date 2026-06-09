import { useState } from 'react'

import type { DocumentMeta, ShareMode } from '../../../lib/protocol'
import { useUpdateShareMode } from '../hooks/useUpdateShareMode'

type ShareBarProps = {
  meta: DocumentMeta
  isOwner: boolean
  readOnly: boolean
  onDownload: () => void
}

const SHARE_MODES: { value: ShareMode; label: string; description: string }[] = [
  { value: 'private', label: 'Private', description: 'Only you can access' },
  { value: 'read', label: 'Can view', description: 'Anyone with the link can view' },
  { value: 'edit', label: 'Can edit', description: 'Anyone with the link can edit' },
]

export function ShareBar({ meta, isOwner, readOnly, onDownload }: ShareBarProps) {
  return (
    <div className="share-container" style={styles.bar}>
      {isOwner ? (
        <SharePanel meta={meta} />
      ) : (
        readOnly && (
          <span style={styles.readOnlyBadge} title="The owner shared this document as view-only.">
            View only
          </span>
        )
      )}
      <button type="button" onClick={onDownload}>
        Save to Machine
      </button>
    </div>
  )
}

// SharePanel is the owner-only control for the document's share mode plus a
// copy-link button. Non-owners never see it; the server enforces ownership
// regardless.
function SharePanel({ meta }: { meta: DocumentMeta }) {
  const { mutate: setShareMode, isPending, error } = useUpdateShareMode(meta.id)
  const [copied, setCopied] = useState(false)

  const shareUrl = window.location.href

  const copyLink = () => {
    navigator.clipboard
      .writeText(shareUrl)
      .then(() => {
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
      })
      .catch(() => prompt('Copy this URL:', shareUrl))
  }

  const active = SHARE_MODES.find((m) => m.value === meta.shareMode)

  return (
    <div className="share-controls" style={styles.panel}>
      <span style={styles.label}>Share:</span>
      <div role="radiogroup" aria-label="Share mode" style={styles.toggleGroup}>
        {SHARE_MODES.map((mode) => (
          <button
            key={mode.value}
            type="button"
            role="radio"
            aria-checked={meta.shareMode === mode.value}
            title={mode.description}
            disabled={isPending}
            onClick={() => {
              if (meta.shareMode !== mode.value) setShareMode(mode.value)
            }}
            style={{
              ...styles.toggleButton,
              ...(meta.shareMode === mode.value ? styles.toggleButtonActive : {}),
            }}
          >
            {mode.label}
          </button>
        ))}
      </div>
      {meta.shareMode !== 'private' && (
        <button type="button" onClick={copyLink}>
          {copied ? 'Copied!' : 'Copy link'}
        </button>
      )}
      {active && meta.shareMode !== 'private' && (
        <span style={styles.hint}>{active.description}</span>
      )}
      {error && (
        <span style={styles.error}>
          {error instanceof Error ? error.message : 'Failed to update sharing'}
        </span>
      )}
    </div>
  )
}

const styles = {
  bar: {
    display: 'flex',
    alignItems: 'center',
    gap: '0.75rem',
    padding: '0.5rem 1rem',
    flexWrap: 'wrap' as const,
  },
  panel: {
    display: 'flex',
    alignItems: 'center',
    gap: '0.5rem',
    flexWrap: 'wrap' as const,
  },
  label: {
    fontWeight: 600,
    fontSize: '0.875rem',
  },
  toggleGroup: {
    display: 'inline-flex',
    border: '1px solid #d1d5db',
    borderRadius: '6px',
    overflow: 'hidden',
  },
  toggleButton: {
    padding: '0.35rem 0.75rem',
    border: 'none',
    background: 'transparent',
    cursor: 'pointer',
    fontSize: '0.875rem',
  },
  toggleButtonActive: {
    background: '#2563eb',
    color: '#fff',
  },
  hint: {
    fontSize: '0.8rem',
    color: '#6b7280',
  },
  readOnlyBadge: {
    padding: '0.25rem 0.6rem',
    borderRadius: '9999px',
    background: '#f3f4f6',
    color: '#374151',
    fontSize: '0.8rem',
    fontWeight: 600,
  },
  error: {
    fontSize: '0.8rem',
    color: '#dc2626',
  },
} as const
