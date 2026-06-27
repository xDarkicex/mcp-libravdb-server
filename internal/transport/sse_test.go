package transport

import (
	"net/http"
	"testing"
)

func TestNewSSEWriter(t *testing.T) {
	w := NewSSEWriter(nil)
	if w == nil {
		t.Fatal("expected non-nil SSEWriter")
	}
	if w.status != 200 {
		t.Fatalf("expected status 200, got %d", w.status)
	}
	if w.Header() == nil {
		t.Fatal("expected non-nil header")
	}
}

func TestSSEWriter_Header(t *testing.T) {
	w := NewSSEWriter(nil)
	w.Header().Set("Content-Type", "text/event-stream")
	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatal("header not set")
	}
}

func TestSSEWriter_WriteHeader(t *testing.T) {
	w := NewSSEWriter(nil)
	w.WriteHeader(201)
	if w.status != 201 {
		t.Fatalf("expected status 201, got %d", w.status)
	}
}

func TestSSEWriter_Flush(t *testing.T) {
	w := NewSSEWriter(nil)
	// Flush is a no-op, should not panic
	w.Flush()
}

func TestSSEWriter_Flush_NoPanic(t *testing.T) {
	w := NewSSEWriter(nil)
	w.Flush() // should not panic even with nil conn
}

func TestSSEWriter_ImplementsFlusher(t *testing.T) {
	w := NewSSEWriter(nil)
	var _ http.Flusher = w
}

