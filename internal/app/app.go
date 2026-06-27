package app

import (
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewMCP creates and configures the MCP server with capabilities.
func NewMCP(logger *slog.Logger, backendHealthy bool) *mcp.Server {
	impl := &mcp.Implementation{
		Name:    "mcp-memory-libravdb",
		Title:   "libravdb Memory Server",
		Version: "1.0.0",
	}

	instructions := `All memory operations require a collection name.
Use memory://collections to discover available collections.
Use memory.search to find relevant memories before adding new ones.

Cognitive kinds: identity, constraint, decision, fact, preference, episode.
Tiers: hard (durable), soft (decayable), variant (alternative perspectives).`

	return mcp.NewServer(impl, &mcp.ServerOptions{
		Instructions: instructions,
		HasTools:     true,
		HasResources: true,
		Logger:       logger,
	})
}
