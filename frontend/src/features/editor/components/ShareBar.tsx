type ShareBarProps = {
  onDownload: () => void
}

export function ShareBar({ onDownload }: ShareBarProps) {
  const shareUrl = window.location.href

  const copyLink = () => {
    navigator.clipboard
      .writeText(shareUrl)
      .then(() => alert('Link copied to clipboard!'))
      .catch(() => prompt('Copy this URL:', shareUrl))
  }

  const shareNative = async () => {
    if (navigator.share) {
      try {
        await navigator.share({ title: 'Notedown', url: shareUrl })
      } catch (err) {
        if (err instanceof Error && err.name !== 'AbortError') {
          copyLink()
        }
      }
      return
    }
    copyLink()
  }

  return (
    <div className="share-container">
      <label htmlFor="share-link">Share this link:</label>
      <div className="share-controls">
        <input id="share-link" type="text" readOnly value={shareUrl} />
        <button type="button" onClick={copyLink}>
          Copy
        </button>
        {typeof navigator.share === 'function' && (
          <button type="button" onClick={shareNative}>
            Share
          </button>
        )}
        <button type="button" onClick={onDownload}>
          Save to Machine
        </button>
      </div>
    </div>
  )
}
