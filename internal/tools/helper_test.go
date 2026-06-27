package tools

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/xDarkicex/mcp-libravdb-server/internal/testutil"
	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

func setupTest(t *testing.T) (*testutil.FakeBackend, ipcv1.LibravDBClient, func()) {
	t.Helper()
	return testutil.NewFakeBackend(t)
}

func startServer(t *testing.T, grpcClient ipcv1.LibravDBClient) *mcp.ClientSession {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	deps := &Deps{
		Client:         grpcClient,
		Logger:         logger,
		BackendTimeout: 5 * time.Second,
		BackendHealthy: grpcClient != nil,
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.0"}, &mcp.ServerOptions{
		HasTools: true,
	})
	if err := RegisterAll(server, deps); err != nil {
		t.Fatalf("register tools: %v", err)
	}

	clientTrans, serverTrans := mcp.NewInMemoryTransports()
	go func() {
		if _, err := server.Connect(t.Context(), serverTrans, nil); err != nil {
			// connection closed after test
		}
	}()

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, err := mcpClient.Connect(t.Context(), clientTrans, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}
