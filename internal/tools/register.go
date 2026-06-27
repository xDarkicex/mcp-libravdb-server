package tools

import (
	"log/slog"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

// Deps holds dependencies injected into tool handlers.
type Deps struct {
	Client         ipcv1.LibravDBClient
	Logger         *slog.Logger
	BackendTimeout time.Duration
	BackendHealthy bool
}

// RegisterAll registers all 7 MCP tools on the server.
func RegisterAll(server *mcp.Server, deps *Deps) error {
	registerSearch(server, deps)
	registerAdd(server, deps)
	registerDelete(server, deps)
	registerRecall(server, deps)
	registerGraph(server, deps)
	registerPredict(server, deps)
	registerStats(server, deps)
	return nil
}
