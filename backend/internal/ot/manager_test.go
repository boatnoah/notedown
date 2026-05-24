package ot

import (
	"testing"
)

func initDoc(t *testing.T, m *Manager, id string) {
	t.Helper()
	m.InitDocument(id)
}

func apply(t *testing.T, m *Manager, id string, op Operation) Snapshot {
	t.Helper()
	snap, _, err := m.Apply(id, op)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	return snap
}

// TestApplyDirect verifies basic same-version operation application.
func TestApplyDirect(t *testing.T) {
	m := NewManager()
	initDoc(t, m, "doc1")

	snap := apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 0, Offset: 0, Text: "hello"})
	if snap.Content != "hello" {
		t.Fatalf("want %q, got %q", "hello", snap.Content)
	}
	if snap.Version != 1 {
		t.Fatalf("want version 1, got %d", snap.Version)
	}

	snap = apply(t, m, "doc1", Operation{Kind: OperationDelete, ClientVersion: 1, Offset: 0, Length: 3})
	if snap.Content != "lo" {
		t.Fatalf("want %q, got %q", "lo", snap.Content)
	}
}

// TestConcurrentInsertsSamePosition verifies deterministic ordering when two
// clients insert at the same position concurrently.
func TestConcurrentInsertsSamePosition(t *testing.T) {
	m := NewManager()
	initDoc(t, m, "doc1")

	// Client A at v0: insert "A" at 0 — arrives and is applied first.
	apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 0, Offset: 0, Text: "A"})

	// Client B also at v0: insert "B" at 0 — arrives after A.
	// Should be transformed: "B" ends up after "A".
	snap := apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 0, Offset: 0, Text: "B"})

	if snap.Content != "AB" {
		t.Fatalf("want %q, got %q", "AB", snap.Content)
	}
}

// TestConcurrentInsertsBeforeAndAfter verifies offset adjustment for inserts
// at different positions.
func TestConcurrentInsertsBeforeAndAfter(t *testing.T) {
	m := NewManager()
	initDoc(t, m, "doc1")

	// Server content: "hello" (v1 after this).
	apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 0, Offset: 0, Text: "hello"})

	// Client A at v1: insert "!" at 5 — applied first.
	apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 1, Offset: 5, Text: "!"})
	// Content: "hello!", v=2

	// Client B at v1: insert ">" at 0. Arrives after A's insert.
	// Offset 0 is before A's insert at 5, so no shift needed.
	snap := apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 1, Offset: 0, Text: ">"})
	// Expected: ">" + "hello" + "!" = ">hello!"
	if snap.Content != ">hello!" {
		t.Fatalf("want %q, got %q", ">hello!", snap.Content)
	}

	m2 := NewManager()
	initDoc(t, m2, "doc2")
	apply(t, m2, "doc2", Operation{Kind: OperationInsert, ClientVersion: 0, Offset: 0, Text: "hello"})

	// Client A at v1: insert ">" at 0 — applied first.
	apply(t, m2, "doc2", Operation{Kind: OperationInsert, ClientVersion: 1, Offset: 0, Text: ">"})
	// Content: ">hello", v=2

	// Client B at v1: insert "!" at 5. Since A inserted at 0 (before 5), shift right.
	snap = apply(t, m2, "doc2", Operation{Kind: OperationInsert, ClientVersion: 1, Offset: 5, Text: "!"})
	// Expected: ">hello!" (! shifted from pos 5 to pos 6)
	if snap.Content != ">hello!" {
		t.Fatalf("want %q, got %q", ">hello!", snap.Content)
	}
}

// TestConcurrentDeletesOverlap verifies that overlapping concurrent deletes
// do not panic and produce consistent state.
func TestConcurrentDeletesOverlap(t *testing.T) {
	m := NewManager()
	initDoc(t, m, "doc1")

	apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 0, Offset: 0, Text: "abcdef"})
	// v=1, content="abcdef"

	// Client A at v1: delete [1,4) → "bcd".
	apply(t, m, "doc1", Operation{Kind: OperationDelete, ClientVersion: 1, Offset: 1, Length: 3})
	// v=2, content="aef"

	// Client B at v1: delete [2,5) → "cde". Overlaps with A's [1,4).
	// After transform: A deleted [1,4). Overlap with [2,5) is [2,4) (len 2).
	// leftShift = min(4,2)-1 = 2-1 = 1. Overlap = 2.
	// result.Offset = 2-1=1, result.Length = 3-2=1 → delete [1,2) → "e" in "aef"
	snap := apply(t, m, "doc1", Operation{Kind: OperationDelete, ClientVersion: 1, Offset: 2, Length: 3})
	// Expected: "af"
	if snap.Content != "af" {
		t.Fatalf("want %q, got %q", "af", snap.Content)
	}
}

// TestConcurrentDeletesNoOverlap verifies that non-overlapping concurrent
// deletes both apply correctly with proper offset adjustment.
func TestConcurrentDeletesNoOverlap(t *testing.T) {
	m := NewManager()
	initDoc(t, m, "doc1")

	apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 0, Offset: 0, Text: "abcdef"})
	// v=1, content="abcdef"

	// Client A at v1: delete [0,2) → "ab".
	apply(t, m, "doc1", Operation{Kind: OperationDelete, ClientVersion: 1, Offset: 0, Length: 2})
	// v=2, content="cdef"

	// Client B at v1: delete [4,6) → "ef". A's delete is before B's range; shift left by 2.
	snap := apply(t, m, "doc1", Operation{Kind: OperationDelete, ClientVersion: 1, Offset: 4, Length: 2})
	// Expected: "cd"
	if snap.Content != "cd" {
		t.Fatalf("want %q, got %q", "cd", snap.Content)
	}
}

// TestConcurrentDeleteEntirelyCoversOther verifies that when one delete entirely
// covers another concurrent delete the result is a no-op (length=0) that doesn't panic.
func TestConcurrentDeleteEntirelyCoversOther(t *testing.T) {
	m := NewManager()
	initDoc(t, m, "doc1")

	apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 0, Offset: 0, Text: "abcdef"})

	// Client A at v1: delete entire string [0,6).
	apply(t, m, "doc1", Operation{Kind: OperationDelete, ClientVersion: 1, Offset: 0, Length: 6})
	// v=2, content=""

	// Client B at v1: delete [2,4) — already gone. Should be a safe no-op.
	snap := apply(t, m, "doc1", Operation{Kind: OperationDelete, ClientVersion: 1, Offset: 2, Length: 2})
	if snap.Content != "" {
		t.Fatalf("want empty, got %q", snap.Content)
	}
}

// TestMultiHopTransform verifies a client that is multiple versions behind is
// correctly transformed against the entire concurrent history.
func TestMultiHopTransform(t *testing.T) {
	m := NewManager()
	initDoc(t, m, "doc1")

	// Build up 3 versions.
	apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 0, Offset: 0, Text: "a"}) // v1: "a"
	apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 1, Offset: 1, Text: "b"}) // v2: "ab"
	apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 2, Offset: 2, Text: "c"}) // v3: "abc"

	// Client at v0 (saw empty document) inserts "X" at 0.
	// Must be transformed against all 3 ops:
	//   against {Insert,0,"a"}: offset 0==0, shift right → 1
	//   against {Insert,1,"b"}: offset 1==1, shift right → 2
	//   against {Insert,2,"c"}: offset 2==2, shift right → 3
	snap := apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 0, Offset: 0, Text: "X"})
	if snap.Content != "abcX" {
		t.Fatalf("want %q, got %q", "abcX", snap.Content)
	}
}

// TestSnapshotUnchanged verifies that Snapshot continues to return the correct
// document state (existing behaviour).
func TestSnapshotUnchanged(t *testing.T) {
	m := NewManager()
	initDoc(t, m, "doc1")

	apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 0, Offset: 0, Text: "hi"})

	snap, err := m.Snapshot("doc1")
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}
	if snap.Content != "hi" || snap.Version != 1 {
		t.Fatalf("unexpected snapshot: %+v", snap)
	}
}

// TestInvalidClientVersion verifies that an out-of-range client version is rejected.
func TestInvalidClientVersion(t *testing.T) {
	m := NewManager()
	initDoc(t, m, "doc1")

	_, _, err := m.Apply("doc1", Operation{Kind: OperationInsert, ClientVersion: 99, Offset: 0, Text: "x"})
	if err == nil {
		t.Fatal("expected error for future client version, got nil")
	}
}

// TestApplyReturnsCanonicalOp verifies that Apply returns the post-transform
// operation so callers can persist it for correct replay via ApplyDirect.
func TestApplyReturnsCanonicalOp(t *testing.T) {
	m := NewManager()
	initDoc(t, m, "doc1")

	// Apply a first op to advance the server to v1.
	_, canon0, _ := m.Apply("doc1", Operation{Kind: OperationInsert, ClientVersion: 0, Offset: 0, Text: "A"})

	// A concurrent op at v0 is transformed: Insert@0 "B" → Insert@1 "B".
	_, canon1, err := m.Apply("doc1", Operation{Kind: OperationInsert, ClientVersion: 0, Offset: 0, Text: "B"})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if canon1.Offset != 1 {
		t.Fatalf("canonical offset want 1, got %d", canon1.Offset)
	}

	// Replay both canonical ops with ApplyDirect into a fresh manager.
	m2 := NewManager()
	initDoc(t, m2, "doc2")
	if _, err := m2.ApplyDirect("doc2", canon0); err != nil {
		t.Fatalf("ApplyDirect canon0: %v", err)
	}
	if _, err := m2.ApplyDirect("doc2", canon1); err != nil {
		t.Fatalf("ApplyDirect canon1: %v", err)
	}
	snap, _ := m2.Snapshot("doc2")
	if snap.Content != "AB" {
		t.Fatalf("replay want %q, got %q", "AB", snap.Content)
	}
}

// TestConcurrentDeleteInsert verifies delete vs insert transformation.
func TestConcurrentDeleteInsert(t *testing.T) {
	m := NewManager()
	initDoc(t, m, "doc1")

	apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 0, Offset: 0, Text: "abcde"})
	// v=1, content="abcde"

	// Client A at v1: delete [1,3) → "bc". Applied first.
	apply(t, m, "doc1", Operation{Kind: OperationDelete, ClientVersion: 1, Offset: 1, Length: 2})
	// v=2, content="ade"

	// Client B at v1: insert "X" at 4 (after 'abcde'). A deleted [1,3) before pos 4.
	// Transform: aEnd(3) <= op.Offset(4) → shift left by 2 → offset=2.
	snap := apply(t, m, "doc1", Operation{Kind: OperationInsert, ClientVersion: 1, Offset: 4, Text: "X"})
	// Expected: "adXe" → wait: "ade" insert at 2 → "adXe"
	if snap.Content != "adXe" {
		t.Fatalf("want %q, got %q", "adXe", snap.Content)
	}
}
