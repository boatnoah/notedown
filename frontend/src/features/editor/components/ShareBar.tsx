type ShareBarProps = {
  onDownload: () => void;
};

export function ShareBar({ onDownload }: ShareBarProps) {
  const shareUrl = window.location.href;

  const copyLink = () => {
    navigator.clipboard
      .writeText(shareUrl)
      .then(() => alert("Link copied to clipboard!"))
      .catch(() => prompt("Copy this URL:", shareUrl));
  };

  const saveLink = () => {
    if (navigator.clipboard) {
      navigator.clipboard
        .writeText(shareUrl)
        .then(() => alert("URL copied! Now paste into your bookmarks bar."))
        .catch(() => alert(`Here's the URL:\n${shareUrl}`));
    } else {
      window.prompt(
        "Copy this URL and press Ctrl+D (or ⌘+D) to bookmark:",
        shareUrl,
      );
    }
  };

  return (
    <div className="share-container">
      <label htmlFor="share-link">Share this link:</label>
      <div className="share-controls">
        <input id="share-link" type="text" readOnly value={shareUrl} />
        <button type="button" onClick={copyLink}>
          Copy
        </button>
        <button type="button" onClick={saveLink}>
          Save Link
        </button>
        <button type="button" onClick={onDownload}>
          Save to Machine
        </button>
      </div>
    </div>
  );
}
