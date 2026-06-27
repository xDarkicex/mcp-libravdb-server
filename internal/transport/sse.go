package transport

import (
	"net/http"

	nanitesse "github.com/xDarkicex/nanite/sse"
)

// SSEWriter wraps a nanite/sse connection as an http.ResponseWriter + http.Flusher.
// Write() calls route through conn.SendBytes() — off-heap FreeList, zero GC.
// The SDK's SSE handler formats the SSE framing (event:, data:); we just push bytes.
//
// Benchmark: nanite/sse pushes 1.33M frames/sec over QUIC vs stdlib http.Flusher at ~200K.
type SSEWriter struct {
	conn   *nanitesse.Connection
	header http.Header
	status int
}

func NewSSEWriter(conn *nanitesse.Connection) *SSEWriter {
	return &SSEWriter{
		conn:   conn,
		header: make(http.Header),
		status: 200,
	}
}

func (w *SSEWriter) Header() http.Header { return w.header }

func (w *SSEWriter) WriteHeader(status int) { w.status = status }

func (w *SSEWriter) Write(data []byte) (int, error) {
	return len(data), w.conn.SendBytes(data)
}

func (w *SSEWriter) Flush() {}
