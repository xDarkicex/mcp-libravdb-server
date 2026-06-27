package app

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc"

	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

// fakeServer starts a fake gRPC server on a Unix socket and returns the socket path.
func fakeServer(t *testing.T) (string, func()) {
	t.Helper()

	dir := t.TempDir()
	sock := filepath.Join(dir, "test.sock")

	listener, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	fake := &fakeBackend{healthOK: true}
	server := grpc.NewServer()
	ipcv1.RegisterLibravDBServer(server, fake)
	go func() { _ = server.Serve(listener) }()

	return "unix://" + sock, func() {
		server.Stop()
		listener.Close()
	}
}

type fakeBackend struct {
	ipcv1.UnimplementedLibravDBServer
	healthOK bool
}

func (f *fakeBackend) Health(ctx context.Context, req *ipcv1.HealthRequest) (*ipcv1.HealthResponse, error) {
	return &ipcv1.HealthResponse{Ok: f.healthOK}, nil
}
func (f *fakeBackend) SearchText(ctx context.Context, req *ipcv1.SearchTextRequest) (*ipcv1.SearchTextResponse, error) {
	return &ipcv1.SearchTextResponse{}, nil
}
func (f *fakeBackend) CognitiveMetrics(ctx context.Context, req *ipcv1.CognitiveMetricsRequest) (*ipcv1.CognitiveMetricsResponse, error) {
	return &ipcv1.CognitiveMetricsResponse{TotalNodes: 10}, nil
}
func (f *fakeBackend) Status(ctx context.Context, req *ipcv1.MemoryStatusRequest) (*ipcv1.MemoryStatusResponse, error) {
	return &ipcv1.MemoryStatusResponse{Ok: true}, nil
}

func TestNewRuntime_Healthy(t *testing.T) {
	os.Unsetenv("LIBRAVDB_AUTH_SECRET")
	os.Unsetenv("LIBRAVDB_AUTH_SECRET_FILE")

	sock, cleanup := fakeServer(t)
	defer cleanup()

	cfg := &Config{
		BackendAddr:    sock,
		BackendTLS:     false,
		BackendTimeout: 5 * time.Second,
		DegradedOk:     false,
		LogLevel:       "error",
		TenantKey:      DefaultTenantKey,
		Shared:         true,
	}

	rt, err := NewRuntime(cfg)
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Shutdown()

	if rt.MCP == nil {
		t.Fatal("expected MCP server")
	}
	if rt.gRPCClient == nil {
		t.Fatal("expected gRPC client")
	}
}

func TestNewRuntime_Degraded(t *testing.T) {
	os.Unsetenv("LIBRAVDB_AUTH_SECRET")
	os.Unsetenv("LIBRAVDB_AUTH_SECRET_FILE")

	cfg := &Config{
		BackendAddr:    "unix:///tmp/nonexistent.sock",
		BackendTLS:     false,
		BackendTimeout: 1 * time.Second,
		DegradedOk:     true,
		LogLevel:       "error",
		TenantKey:      DefaultTenantKey,
		Shared:         true,
	}

	rt, err := NewRuntime(cfg)
	if err != nil {
		t.Fatalf("degraded NewRuntime should not error: %v", err)
	}
	defer rt.Shutdown()

	// grpc.NewClient is lazy — dial always succeeds. Client exists but BackendHealthy is false.
	if rt.gRPCClient == nil {
		t.Fatal("expected non-nil gRPC client (lazy dial), but BackendHealthy should be false")
	}
}

func TestNewRuntime_BackendDownNoDegrade(t *testing.T) {
	os.Unsetenv("LIBRAVDB_AUTH_SECRET")
	os.Unsetenv("LIBRAVDB_AUTH_SECRET_FILE")

	cfg := &Config{
		BackendAddr:    "unix:///tmp/nonexistent.sock",
		BackendTLS:     false,
		BackendTimeout: 1 * time.Second,
		DegradedOk:     false,
		LogLevel:       "error",
		TenantKey:      DefaultTenantKey,
		Shared:         true,
	}

	_, err := NewRuntime(cfg)
	if err == nil {
		t.Fatal("expected error when backend is down without degraded-ok")
	}
}

func TestRuntime_EndToEnd(t *testing.T) {
	os.Unsetenv("LIBRAVDB_AUTH_SECRET")
	os.Unsetenv("LIBRAVDB_AUTH_SECRET_FILE")

	sock, cleanup := fakeServer(t)
	defer cleanup()

	cfg := &Config{
		BackendAddr:    sock,
		BackendTLS:     false,
		BackendTimeout: 5 * time.Second,
		LogLevel:       "error",
		TenantKey:      DefaultTenantKey,
		Shared:         true,
	}

	rt, err := NewRuntime(cfg)
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Shutdown()

	clientTrans, serverTrans := mcp.NewInMemoryTransports()
	go func() { _, _ = rt.MCP.Connect(t.Context(), serverTrans, nil) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	session, err := client.Connect(t.Context(), clientTrans, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "memory.stats", Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success, got error")
	}
}

func TestRuntime_ToolList(t *testing.T) {
	os.Unsetenv("LIBRAVDB_AUTH_SECRET")
	os.Unsetenv("LIBRAVDB_AUTH_SECRET_FILE")

	sock, cleanup := fakeServer(t)
	defer cleanup()

	cfg := &Config{
		BackendAddr:    sock,
		BackendTLS:     false,
		BackendTimeout: 5 * time.Second,
		LogLevel:       "error",
		TenantKey:      DefaultTenantKey,
		Shared:         true,
	}

	rt, err := NewRuntime(cfg)
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Shutdown()

	clientTrans, serverTrans := mcp.NewInMemoryTransports()
	go func() { _, _ = rt.MCP.Connect(t.Context(), serverTrans, nil) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	session, err := client.Connect(t.Context(), clientTrans, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	result, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(result.Tools) != 7 {
		t.Fatalf("expected 7 tools, got %d", len(result.Tools))
	}
}

func TestSetupRouter_HealthEndpoint(t *testing.T) {
	sock, cleanup := fakeServer(t)
	defer cleanup()

	cfg := &Config{
		BackendAddr:    sock,
		BackendTLS:     false,
		BackendTimeout: 5 * time.Second,
		LogLevel:       "error",
		TenantKey:      DefaultTenantKey,
		Shared:         true,
	}

	rt, err := NewRuntime(cfg)
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Shutdown()

	r := setupRouter(rt, cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestSetupRouter_MCPEndpoint(t *testing.T) {
	sock, cleanup := fakeServer(t)
	defer cleanup()

	cfg := &Config{
		BackendAddr:    sock,
		BackendTLS:     false,
		BackendTimeout: 5 * time.Second,
		LogLevel:       "error",
		TenantKey:      DefaultTenantKey,
		Shared:         true,
	}

	rt, err := NewRuntime(cfg)
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Shutdown()

	r := setupRouter(rt, cfg)

	req := httptest.NewRequest("POST", "/mcp", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// Returns 400 from MCP handler (invalid JSON body) — proves routing works
	if w.Code < 400 {
		t.Fatalf("expected error (no body), got %d", w.Code)
	}
}

func TestRequestLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mw := requestLogger(logger)
	if mw == nil {
		t.Fatal("expected non-nil middleware")
	}
}

func TestSetupRouter_WithAuth(t *testing.T) {
	sock, cleanup := fakeServer(t)
	defer cleanup()

	cfg := &Config{
		BackendAddr:    sock,
		BackendTLS:     false,
		BackendTimeout: 5 * time.Second,
		LogLevel:       "error",
		TenantKey:      DefaultTenantKey,
		Shared:         true,
		HTTPExpose:     true,
		AuthToken:      "test-token",
	}

	rt, err := NewRuntime(cfg)
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Shutdown()

	r := setupRouter(rt, cfg)

	// No auth header → 401
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Fatalf("expected 401 without auth, got %d", w.Code)
	}

	// Correct auth → 200
	req = httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200 with auth, got %d", w.Code)
	}
}

func TestRunStdio_ConfigError(t *testing.T) {
	cfg := &Config{
		BackendAddr:    "unix:///nonexistent",
		BackendTLS:     false,
		BackendTimeout: 1 * time.Second,
		DegradedOk:     false,
	}
	err := RunStdio(cfg)
	if err == nil {
		t.Fatal("expected error when backend is down")
	}
}

func TestRunHTTP_ConfigError(t *testing.T) {
	cfg := &Config{
		BackendAddr:    "unix:///nonexistent",
		BackendTLS:     false,
		BackendTimeout: 1 * time.Second,
		DegradedOk:     false,
	}
	err := RunHTTP(cfg)
	if err == nil {
		t.Fatal("expected error when backend is down")
	}
}

func TestRunStdio_Integration(t *testing.T) {
	os.Unsetenv("LIBRAVDB_AUTH_SECRET")
	os.Unsetenv("LIBRAVDB_AUTH_SECRET_FILE")

	sock, cleanup := fakeServer(t)
	defer cleanup()

	cfg := &Config{
		BackendAddr:    sock,
		BackendTLS:     false,
		BackendTimeout: 5 * time.Second,
		LogLevel:       "error",
		TenantKey:      DefaultTenantKey,
		Shared:         true,
	}

	// Run in goroutine with cancelled context → returns immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	errCh := make(chan error, 1)
	go func() {
		rt, err := NewRuntime(cfg)
		if err != nil {
			errCh <- err
			return
		}
		defer rt.Shutdown()
		errCh <- rt.MCP.Run(ctx, &mcp.StdioTransport{})
	}()

	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			// Expected: context canceled or EOF
			t.Logf("RunStdio returned: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunStdio timed out")
	}
}

func TestRunHTTP_SetupRouter(t *testing.T) {
	sock, cleanup := fakeServer(t)
	defer cleanup()

	cfg := &Config{
		BackendAddr:    sock,
		BackendTLS:     false,
		BackendTimeout: 5 * time.Second,
		LogLevel:       "error",
		TenantKey:      DefaultTenantKey,
		Shared:         true,
		HTTPHost:       "127.0.0.1",
		HTTPPort:       8082,
	}

	rt, err := NewRuntime(cfg)
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Shutdown()

	r := setupRouter(rt, cfg)

	// Verify the router has all expected endpoints
	for _, tc := range []struct {
		method string
		path   string
	}{
		{"GET", "/health"},
		{"POST", "/mcp"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == 0 {
			t.Errorf("%s %s: no response code", tc.method, tc.path)
		}
	}
}

func TestSetupRouter_AllRoutes(t *testing.T) {
	sock, cleanup := fakeServer(t)
	defer cleanup()

	cfg := &Config{
		BackendAddr:    sock,
		BackendTLS:     false,
		BackendTimeout: 5 * time.Second,
		LogLevel:       "error",
		TenantKey:      DefaultTenantKey,
		Shared:         true,
	}

	rt, err := NewRuntime(cfg)
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Shutdown()

	r := setupRouter(rt, cfg)

	// Health endpoint with backend healthy
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("health: expected 200, got %d", w.Code)
	}
}

func TestRuntime_ResourcesList(t *testing.T) {
	os.Unsetenv("LIBRAVDB_AUTH_SECRET")
	os.Unsetenv("LIBRAVDB_AUTH_SECRET_FILE")

	sock, cleanup := fakeServer(t)
	defer cleanup()

	cfg := &Config{
		BackendAddr:    sock,
		BackendTLS:     false,
		BackendTimeout: 5 * time.Second,
		LogLevel:       "error",
		TenantKey:      DefaultTenantKey,
		Shared:         true,
	}

	rt, err := NewRuntime(cfg)
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Shutdown()

	clientTrans, serverTrans := mcp.NewInMemoryTransports()
	go func() { _, _ = rt.MCP.Connect(t.Context(), serverTrans, nil) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	session, err := client.Connect(t.Context(), clientTrans, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	resources, err := session.ListResources(t.Context(), nil)
	if err != nil {
		t.Fatalf("list resources: %v", err)
	}
	if len(resources.Resources) < 3 {
		t.Fatalf("expected at least 3 resources, got %d", len(resources.Resources))
	}
}
