package app

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"
)

const DefaultTimeout = 5 * time.Second
const DefaultTenantKey = "libravdb-mcp-server"

// Config holds all server configuration, populated from viper (env + flags).
type Config struct {
	BackendAddr    string
	BackendTLS     bool
	BackendTimeout time.Duration
	DegradedOk     bool
	LogLevel       string
	TenantKey      string

	AuthToken string // bearer token for HTTP auth (MCP_AUTH_TOKEN)

	HTTPPort   int
	HTTPHost   string
	HTTPExpose bool

	// Workspace isolation
	Workspace string // explicit workspace name (appended to tenant key)
	Shared    bool   // disable per-workspace isolation
}

// ResolveTenantKey returns the effective tenant key with workspace isolation applied.
// Default: "libravdb-mcp-server:<cwd-hash>" (isolated per project).
// With --shared: "libravdb-mcp-server" (all workspaces share).
// With --workspace foo: "libravdb-mcp-server:foo" (explicit named workspace).
// With --tenant-key custom + --shared: "custom" (explicit override).
func (c *Config) ResolveTenantKey() string {
	base := c.TenantKey
	if base == "" {
		base = DefaultTenantKey
	}

	if c.Shared {
		return base
	}

	if c.Workspace != "" {
		return base + ":" + c.Workspace
	}

	return base + ":" + hashCWD()
}

func hashCWD() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	h := sha256.Sum256([]byte(cwd))
	return fmt.Sprintf("%x", h[:4]) // 8 hex chars, enough for project isolation
}
