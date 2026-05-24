package realtime

import (
	"encoding/json"
	"fmt"

	"github.com/boatnoah/notedown/internal/ot"
)

// ---------- Client → Server ----------

// ClientMsg is the discriminated union of all client→server message kinds.
// To add a new kind: add a struct with clientMsg(), add a case to UnmarshalClient.
type ClientMsg interface{ clientMsg() }

// OperationMsg carries an OT operation from the client.
type OperationMsg struct {
	Operation ot.Operation `json:"operation"`
}

func (OperationMsg) clientMsg() {}

// SyncMsg requests the current document snapshot.
type SyncMsg struct{}

func (SyncMsg) clientMsg() {}

// PresenceMsg updates the client's cursor position.
type PresenceMsg struct {
	Presence CursorPayload `json:"presence"`
}

func (PresenceMsg) clientMsg() {}

// CursorPayload holds the anchor and head offsets of a client cursor.
type CursorPayload struct {
	Anchor int `json:"anchor"`
	Head   int `json:"head"`
}

// UnmarshalClient decodes a raw WebSocket frame into a typed ClientMsg.
func UnmarshalClient(data []byte) (ClientMsg, error) {
	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}
	if envelope.Type == "" {
		return nil, fmt.Errorf("client message missing required field: type")
	}
	switch envelope.Type {
	case "operation":
		var raw struct {
			Operation *ot.Operation `json:"operation"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, err
		}
		if raw.Operation == nil {
			return nil, fmt.Errorf("operation message missing required field: operation")
		}
		return OperationMsg{Operation: *raw.Operation}, nil
	case "sync":
		return SyncMsg{}, nil
	case "presence":
		var raw struct {
			Presence *CursorPayload `json:"presence"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, err
		}
		if raw.Presence == nil {
			return nil, fmt.Errorf("presence message missing required field: presence")
		}
		return PresenceMsg{Presence: *raw.Presence}, nil
	default:
		return nil, fmt.Errorf("unknown client message type: %q", envelope.Type)
	}
}

// ---------- Server → Client ----------

// ServerMsg is the discriminated union of all server→client message kinds.
// To add a new kind: add a struct with serverMsg(), add a case to MarshalServer.
type ServerMsg interface{ serverMsg() }

// SnapshotMsg delivers the current document state.
type SnapshotMsg struct {
	Snapshot ot.Snapshot `json:"snapshot"`
}

func (SnapshotMsg) serverMsg() {}

// PresenceSnapshotMsg delivers the full presence state on connect.
type PresenceSnapshotMsg struct {
	Presences map[string]Presence `json:"presences"`
}

func (PresenceSnapshotMsg) serverMsg() {}

// PresenceUpdateMsg notifies clients of one user's cursor change.
type PresenceUpdateMsg struct {
	UserID   string   `json:"userId"`
	Presence Presence `json:"presence"`
}

func (PresenceUpdateMsg) serverMsg() {}

// ErrorMsg reports a server-side error to the client.
type ErrorMsg struct {
	Error string `json:"error"`
}

func (ErrorMsg) serverMsg() {}

// MarshalServer encodes a ServerMsg to JSON with the type discriminator injected.
func MarshalServer(msg ServerMsg) ([]byte, error) {
	type T struct {
		Type string `json:"type"`
	}
	switch m := msg.(type) {
	case SnapshotMsg:
		return json.Marshal(struct {
			T
			SnapshotMsg
		}{T{"snapshot"}, m})
	case PresenceSnapshotMsg:
		return json.Marshal(struct {
			T
			PresenceSnapshotMsg
		}{T{"presenceSnapshot"}, m})
	case PresenceUpdateMsg:
		return json.Marshal(struct {
			T
			PresenceUpdateMsg
		}{T{"presenceUpdate"}, m})
	case ErrorMsg:
		return json.Marshal(struct {
			T
			ErrorMsg
		}{T{"error"}, m})
	default:
		return nil, fmt.Errorf("unknown server message type: %T", msg)
	}
}
