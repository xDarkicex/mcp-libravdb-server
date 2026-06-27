# mcp-memory-libravdb

MCP server exposing libravdbd's cognitive memory system to any MCP-compatible AI client — Claude Code, Claude Desktop, Codex, Cursor, Cline, Zed, and Aider.

Thin protocol translation layer. All memory intelligence (cognitive classification, deontic analysis, embedding, graph topology, causal reasoning) lives in the daemon. This server translates MCP JSON-RPC to gRPC and back.

## Quick Start

```bash
# 1. Start the daemon
libravdbd

# 2. Start the MCP server (stdio mode)
mcp-memory-libravdb stdio

# 3. Add to your client
claude mcp add libravdb-memory -- mcp-memory-libravdb stdio
```

## Install

**Go install:**
```bash
go install github.com/xDarkicex/MCP-memory-libravdb/cmd/mcp-memory-libravdb@latest
```

**Binary:**
```bash
git clone https://github.com/xDarkicex/MCP-memory-libravdb
cd MCP-memory-libravdb
make build
./bin/mcp-memory-libravdb --version
```

**Docker:**
```bash
docker build -t mcp-memory-libravdb .
docker run mcp-memory-libravdb --version
```

## Configure

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--backend-addr` | `unix://~/.libravdbd/run/libravdb.sock` | gRPC backend address |
| `--backend-tls` | `false` | Enable TLS for gRPC |
| `--backend-timeout` | `5s` | Per-call gRPC timeout |
| `--tenant-key` | `libravdb-mcp-server` | Daemon tenant key for data isolation |
| `--shared` | `false` | Disable per-workspace isolation |
| `--workspace` | (auto: CWD hash) | Explicit workspace name |
| `--degraded-ok` | `false` | Start even if daemon is down |
| `--log-level` | `info` | debug, info, warn, error |
| `--auth-token` | — | Bearer token for HTTP auth (when `--http-expose`) |
| `--http-port` | `8082` | HTTP server port |
| `--http-host` | `127.0.0.1` | HTTP bind address |
| `--http-expose` | `false` | Bind to non-localhost interfaces |

### Environment variables

All flags map to env vars with `LIBRAVDB_` prefix and underscores. Example: `--backend-addr` → `LIBRAVDB_BACKEND_ADDR`.

| Variable | Maps to |
|----------|---------|
| `LIBRAVDBD_ADDR` | `--backend-addr` |
| `LIBRAVDB_TENANT_KEY` | `--tenant-key` |
| `LIBRAVDB_AUTH_TOKEN` | `--auth-token` |
| `LIBRAVDB_AUTH_SECRET` | HMAC auth secret (gRPC) |
| `LIBRAVDB_AUTH_SECRET_FILE` | Path to auth secret file |
| `LOG_LEVEL` | `--log-level` |

### Workspace isolation

By default, each project directory gets its own tenant key by hashing the CWD:

```
~/repos/foo → tenant: libravdb-mcp-server:a1b2c3d4
~/repos/bar → tenant: libravdb-mcp-server:e5f6g7h8
```

Share a tenant across projects:
```bash
mcp-memory-libravdb stdio --shared          # all projects share one tenant
mcp-memory-libravdb stdio --workspace myproj # explicit named workspace
```

## Client Setup

### Claude Code
```bash
claude mcp add libravdb-memory -- mcp-memory-libravdb stdio
```
Or in `.mcp.json`:
```json
{
  "mcpServers": {
    "libravdb-memory": {
      "type": "stdio",
      "command": "mcp-memory-libravdb",
      "args": ["stdio"],
      "env": { "LIBRAVDB_TENANT_KEY": "my-project" }
    }
  }
}
```

### Claude Desktop
Add to `claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "libravdb-memory": {
      "type": "stdio",
      "command": "mcp-memory-libravdb",
      "args": ["stdio", "--tenant-key", "my-project"]
    }
  }
}
```

### Codex
```toml
# ~/.codex/config.toml
[mcp_servers.libravdb-memory]
command = "mcp-memory-libravdb"
args = ["stdio"]
```

### Cursor
Add to `~/.cursor/mcp.json`:
```json
{
  "mcpServers": {
    "libravdb-memory": {
      "type": "stdio",
      "command": "mcp-memory-libravdb",
      "args": ["stdio"]
    }
  }
}
```

### Cline (VS Code)
In Cline settings → MCP Servers → add stdio server with command `mcp-memory-libravdb stdio`.

### Zed
Add to `~/.config/zed/settings.json`:
```json
{
  "context_servers": {
    "libravdb-memory": {
      "command": { "path": "mcp-memory-libravdb", "args": ["stdio"] }
    }
  }
}
```

## CLI Commands

```
mcp-memory-libravdb stdio              # MCP server over stdio
mcp-memory-libravdb http               # MCP server over Streamable HTTP
mcp-memory-libravdb status             # Daemon health + tenant list
mcp-memory-libravdb status -v          # + MCP server process info (RSS, heap)
mcp-memory-libravdb status --json      # Machine-readable JSON
mcp-memory-libravdb workspaces         # Focused tenant list
mcp-memory-libravdb tenant list        # All open tenants
mcp-memory-libravdb tenant inspect <k> # Tenant details
mcp-memory-libravdb tenant evict <k>   # Unload tenant from daemon memory
```

## MCP Tools

| Tool | Description | Daemon RPC |
|------|-------------|------------|
| `memory.search` | Semantic search across collections | `SearchText` / `SearchTextCollections` |
| `memory.add` | Add a memory record (auto-UUID) | `InsertText` |
| `memory.delete` | Delete by ID (single-ID only) | `Delete` |
| `memory.recall` | Search + gating enrichment + kind filter | `SearchText` + `GatingScalar` |
| `memory.graph` | Walk causal graph from a record | `ExpandSummary` (graph mode) |
| `memory.predict` | Predict next relevant memories | `SearchText` + `ExpandSummary` |
| `memory.stats` | Node counts, kind/tier breakdowns | `CognitiveMetrics` + `Status` |

## Architecture

```
MCP Client (Claude Code / Codex / Cursor / ...)
    │
    │  MCP over stdio or Streamable HTTP
    │
    └──→ mcp-memory-libravdb
              │
              │  gRPC over Unix socket (HMAC + tenant key)
              │
              └──→ libravdbd
                        ├── Cognitive classification (identity/constraint/decision/fact/preference/episode)
                        ├── Deontic two-pass gating
                        ├── Signal bitmask
                        ├── TopoRegistry (directed causal graph)
                        ├── Embedding cache (ONNX / remote API)
                        └── Multi-tenant workspace IAM
```

The daemon does the math. The MCP server translates protocols.

## Development

```bash
make          # lint + test-race + coverage + build
make test     # go test ./...
make build    # build binary
make lint     # golangci-lint
make coverage # test with coverage report
make clean    # remove bin/
```

## License

MIT
