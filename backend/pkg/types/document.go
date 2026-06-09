package types

import "time"

// ShareMode controls who can access a document via its share link.
type ShareMode string

const (
	// ShareModePrivate restricts access to the document owner.
	ShareModePrivate ShareMode = "private"
	// ShareModeRead lets any authenticated user with the link view the document.
	ShareModeRead ShareMode = "read"
	// ShareModeEdit lets any authenticated user with the link edit the document.
	ShareModeEdit ShareMode = "edit"
)

// Valid reports whether the value is one of the known share modes.
func (m ShareMode) Valid() bool {
	switch m {
	case ShareModePrivate, ShareModeRead, ShareModeEdit:
		return true
	}
	return false
}

// Document represents metadata about a collaborative note.
type Document struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"ownerId"`
	Title     string    `json:"title"`
	ShareMode ShareMode `json:"shareMode"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CanRead reports whether the given user may view the document.
func (d *Document) CanRead(userID string) bool {
	if userID == "" {
		return false
	}
	if d.OwnerID == userID {
		return true
	}
	return d.ShareMode == ShareModeRead || d.ShareMode == ShareModeEdit
}

// CanEdit reports whether the given user may modify the document.
func (d *Document) CanEdit(userID string) bool {
	if userID == "" {
		return false
	}
	if d.OwnerID == userID {
		return true
	}
	return d.ShareMode == ShareModeEdit
}

// Session tracks a user's participation within a document room.
type Session struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"documentId"`
	UserID     string    `json:"userId"`
	CreatedAt  time.Time `json:"createdAt"`
}
