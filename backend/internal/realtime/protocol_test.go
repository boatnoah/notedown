package realtime_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/boatnoah/notedown/internal/ot"
	"github.com/boatnoah/notedown/internal/realtime"
)

// ---------- UnmarshalClient round-trip tests ----------

func TestUnmarshalClient_Operation(t *testing.T) {
	raw := `{"type":"operation","operation":{"id":"op1","clientId":"c1","kind":"insert","offset":5,"length":0,"text":"hello","timestamp":"0001-01-01T00:00:00Z"}}`
	msg, err := realtime.UnmarshalClient([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := msg.(realtime.OperationMsg)
	if !ok {
		t.Fatalf("got %T, want OperationMsg", msg)
	}
	if m.Operation.Kind != "insert" {
		t.Errorf("kind = %q, want %q", m.Operation.Kind, "insert")
	}
	if m.Operation.Offset != 5 {
		t.Errorf("offset = %d, want 5", m.Operation.Offset)
	}
	if m.Operation.Text != "hello" {
		t.Errorf("text = %q, want %q", m.Operation.Text, "hello")
	}
}

func TestUnmarshalClient_Sync(t *testing.T) {
	msg, err := realtime.UnmarshalClient([]byte(`{"type":"sync"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := msg.(realtime.SyncMsg); !ok {
		t.Fatalf("got %T, want SyncMsg", msg)
	}
}

func TestUnmarshalClient_Presence(t *testing.T) {
	raw := `{"type":"presence","presence":{"anchor":3,"head":7}}`
	msg, err := realtime.UnmarshalClient([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := msg.(realtime.PresenceMsg)
	if !ok {
		t.Fatalf("got %T, want PresenceMsg", msg)
	}
	if m.Presence.Anchor != 3 {
		t.Errorf("anchor = %d, want 3", m.Presence.Anchor)
	}
	if m.Presence.Head != 7 {
		t.Errorf("head = %d, want 7", m.Presence.Head)
	}
}

func TestUnmarshalClient_UnknownType(t *testing.T) {
	_, err := realtime.UnmarshalClient([]byte(`{"type":"bogus"}`))
	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}
}

// ---------- MarshalServer round-trip tests ----------

func roundTrip(t *testing.T, msg realtime.ServerMsg) map[string]json.RawMessage {
	t.Helper()
	data, err := realtime.MarshalServer(msg)
	if err != nil {
		t.Fatalf("MarshalServer error: %v", err)
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	return out
}

func jsonString(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("expected JSON string: %v", err)
	}
	return s
}

func TestMarshalServer_Snapshot(t *testing.T) {
	msg := realtime.SnapshotMsg{Snapshot: ot.Snapshot{
		DocumentID: "doc1",
		Version:    3,
		Content:    "hello world",
	}}
	fields := roundTrip(t, msg)
	if got := jsonString(t, fields["type"]); got != "snapshot" {
		t.Errorf("type = %q, want snapshot", got)
	}

	var snap ot.Snapshot
	if err := json.Unmarshal(fields["snapshot"], &snap); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if snap.DocumentID != "doc1" || snap.Version != 3 || snap.Content != "hello world" {
		t.Errorf("snapshot fields mismatch: %+v", snap)
	}
}

func TestMarshalServer_PresenceSnapshot(t *testing.T) {
	presences := map[string]realtime.Presence{
		"u1": {UserID: "u1", Name: "Alice", Color: "#ff0000", Anchor: 1, Head: 2, UpdatedAt: time.Time{}},
	}
	fields := roundTrip(t, realtime.PresenceSnapshotMsg{Presences: presences})
	if got := jsonString(t, fields["type"]); got != "presenceSnapshot" {
		t.Errorf("type = %q, want presenceSnapshot", got)
	}
	var out map[string]realtime.Presence
	if err := json.Unmarshal(fields["presences"], &out); err != nil {
		t.Fatalf("unmarshal presences: %v", err)
	}
	if out["u1"].Name != "Alice" {
		t.Errorf("presence name = %q, want Alice", out["u1"].Name)
	}
}

func TestMarshalServer_PresenceUpdate(t *testing.T) {
	msg := realtime.PresenceUpdateMsg{
		UserID:   "u2",
		Presence: realtime.Presence{UserID: "u2", Anchor: 10, Head: 15},
	}
	fields := roundTrip(t, msg)
	if got := jsonString(t, fields["type"]); got != "presenceUpdate" {
		t.Errorf("type = %q, want presenceUpdate", got)
	}
	if got := jsonString(t, fields["userId"]); got != "u2" {
		t.Errorf("userId = %q, want u2", got)
	}
}

func TestMarshalServer_Error(t *testing.T) {
	fields := roundTrip(t, realtime.ErrorMsg{Error: "something went wrong"})
	if got := jsonString(t, fields["type"]); got != "error" {
		t.Errorf("type = %q, want error", got)
	}
	if got := jsonString(t, fields["error"]); got != "something went wrong" {
		t.Errorf("error = %q, want %q", got, "something went wrong")
	}
}
