package types

import "time"

// Document represents metadata about a collaborative note.
type Document struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"ownerId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Session tracks a user's participation within a document room.
type Session struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"documentId"`
	UserID     string    `json:"userId"`
	CreatedAt  time.Time `json:"createdAt"`
}
