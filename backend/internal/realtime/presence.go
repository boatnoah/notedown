package realtime

import (
	"hash/fnv"
	"sync"
	"time"
)

type Presence struct {
	UserID    string    `json:"userId"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	Anchor    int       `json:"anchor"`
	Head      int       `json:"head"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PresenceStore keeps per-document presence state with simple TTL cleanup.
type PresenceStore struct {
	mu      sync.RWMutex
	entries map[string]map[string]Presence // docID -> userID -> Presence
	maxIdle time.Duration
}

func NewPresenceStore(maxIdle time.Duration) *PresenceStore {
	return &PresenceStore{
		entries: make(map[string]map[string]Presence),
		maxIdle: maxIdle,
	}
}

func (p *PresenceStore) Update(documentID string, presence Presence) {
	p.mu.Lock()
	defer p.mu.Unlock()
	presence.UpdatedAt = time.Now().UTC()
	if _, ok := p.entries[documentID]; !ok {
		p.entries[documentID] = make(map[string]Presence)
	}
	p.entries[documentID][presence.UserID] = presence
}

func (p *PresenceStore) Remove(documentID, userID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if doc, ok := p.entries[documentID]; ok {
		delete(doc, userID)
		if len(doc) == 0 {
			delete(p.entries, documentID)
		}
	}
}

func (p *PresenceStore) Snapshot(documentID string) map[string]Presence {
	p.mu.RLock()
	defer p.mu.RUnlock()
	src := p.entries[documentID]
	result := make(map[string]Presence, len(src))
	for k, v := range src {
		result[k] = v
	}
	return result
}

// Prune removes idle presences older than maxIdle.
func (p *PresenceStore) Prune(documentID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if doc, ok := p.entries[documentID]; ok {
		cutoff := time.Now().UTC().Add(-p.maxIdle)
		for k, v := range doc {
			if v.UpdatedAt.Before(cutoff) {
				delete(doc, k)
			}
		}
		if len(doc) == 0 {
			delete(p.entries, documentID)
		}
	}
}

func assignColor(userID string) string {
	palette := []string{
		"#ff6b6b", "#4dabf7", "#ffd43b", "#b197fc", "#63e6be",
		"#ffa94d", "#74c0fc", "#f783ac", "#82c91e", "#fab005",
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(userID))
	return palette[int(h.Sum32())%len(palette)]
}
