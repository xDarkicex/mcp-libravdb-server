# mcp-memory-libravdb

[![Go Reference](https://pkg.go.dev/badge/github.com/xDarkicex/mcp-libravdb-server.svg)](https://pkg.go.dev/github.com/xDarkicex/mcp-libravdb-server)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8.svg?style=flat-square)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/xDarkicex/mcp-libravdb-server)](https://goreportcard.com/report/github.com/xDarkicex/mcp-libravdb-server)
[![Test](https://img.shields.io/github/actions/workflow/status/xDarkicex/mcp-libravdb-server/go.yml?branch=main&style=flat-square)](https://github.com/xDarkicex/mcp-libravdb-server/actions/workflows/go.yml)
[![Coverage](https://img.shields.io/endpoint?style=flat-square&url=https://raw.githubusercontent.com/xDarkicex/mcp-libravdb-server/coverage/coverage.json?v=86a262f)](https://github.com/xDarkicex/mcp-libravdb-server/actions/workflows/go.yml)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)

---

**The official MCP server for the libravdbd memory kernel cognitive system.**

Give any MCP-compatible AI client — Claude Code, Codex, Cursor, Cline, Zed, Aider — persistent cognitive memory. Semantic search, deontic gating, causal graph traversal, and workspace-aware multi-tenant isolation. One protocol translation layer over gRPC. Zero memory intelligence reimplemented.

---

- [Quick Start](#quick-start)
- [Install](#install)
- [Configuration](#configuration)
- [CLI Reference](#cli-reference)
- [MCP Tools](#mcp-tools)
- [Client Setup](#client-setup)
- [Architecture](#architecture)
- [Development](#development)
- [License](#license)

## Quick Start

```bash
# 1. Start the daemon
libravdbd

# 2. Start the MCP server
mcp-memory-libravdb stdio

# 3. Connect your client
claude mcp add libravdb-memory -- mcp-memory-libravdb stdio
```

## Install

**macOS (Homebrew):**
```bash
brew install xDarkicex/homebrew-mcp-libravdb/mcp-memory-libravdb
```

**Linux (APT):**
```bash
curl -fsSL https://xDarkicex.github.io/apt-mcp-libravdb/gpg.key | sudo gpg --dearmor -o /usr/share/keyrings/mcp-libravdb.gpg
echo "deb [signed-by=/usr/share/keyrings/mcp-libravdb.gpg] https://xDarkicex.github.io/apt-mcp-libravdb stable main" | sudo tee /etc/apt/sources.list.d/mcp-libravdb.list
sudo apt update && sudo apt install mcp-memory-libravdb
```

**Go install:**
```bash
go install github.com/xDarkicex/mcp-libravdb-server/cmd/mcp-memory-libravdb@latest
```

**Build from source:**
```bash
git clone https://github.com/xDarkicex/mcp-libravdb-server
cd mcp-libravdb-server
make build
./bin/mcp-memory-libravdb --version
```

**Docker:**
```bash
docker build -t mcp-memory-libravdb .
docker run mcp-memory-libravdb --version
```

## Configuration

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--backend-addr` | `unix://~/.libravdbd/run/libravdb.sock` | gRPC backend address |
| `--tenant-key` | `libravdb-mcp-server` | Daemon tenant key for data isolation |
| `--backend-timeout` | `5s` | Per-call gRPC timeout |
| `--shared` | `false` | Disable per-workspace isolation |
| `--workspace` | (auto: CWD hash) | Explicit workspace name |
| `--degraded-ok` | `false` | Start even if daemon is unavailable |
| `--log-level` | `info` | debug, info, warn, error |
| `--auth-token` | — | Bearer token for HTTP auth |
| `--http-port` | `8082` | HTTP server port |
| `--http-host` | `127.0.0.1` | HTTP bind address |
| `--http-expose` | `false` | Bind to non-localhost |

### Environment Variables

All flags map to env vars with `LIBRAVDB_` prefix:

| Variable | Maps to |
|----------|---------|
| `LIBRAVDBD_ADDR` | `--backend-addr` |
| `LIBRAVDB_TENANT_KEY` | `--tenant-key` |
| `LIBRAVDB_AUTH_TOKEN` | `--auth-token` |
| `LIBRAVDB_AUTH_SECRET` | HMAC auth secret for gRPC |
| `LIBRAVDB_AUTH_SECRET_FILE` | Path to auth secret file |

### Workspace Isolation

By default, each project directory gets its own tenant key:

```
~/repos/foo → libravdb-mcp-server:a1b2c3d4
~/repos/bar → libravdb-mcp-server:e5f6g7h8
```

Override with `--shared` (single tenant) or `--workspace <name>` (named).

## CLI Reference

```
mcp-memory-libravdb stdio              Start MCP server over stdio
mcp-memory-libravdb http               Start MCP server over Streamable HTTP
mcp-memory-libravdb status             Daemon health + tenant list
mcp-memory-libravdb status -v          + MCP server process info
mcp-memory-libravdb status --json      Machine-readable JSON
mcp-memory-libravdb workspaces         Focused tenant list
mcp-memory-libravdb tenant list        All open tenants
mcp-memory-libravdb tenant inspect <k> Tenant details
mcp-memory-libravdb tenant evict <k>   Unload tenant
```

## MCP Tools

| Tool | Description | Daemon RPC |
|------|-------------|------------|
| `memory.search` | Semantic search across collections | `SearchText` |
| `memory.add` | Add a memory record | `InsertText` |
| `memory.delete` | Delete by ID | `Delete` |
| `memory.recall` | Search + gating enrichment | `SearchText` + `GatingScalar` |
| `memory.graph` | Walk causal graph from a record | `ExpandSummary` |
| `memory.predict` | Predict next relevant memories | `SearchText` + `ExpandSummary` |
| `memory.stats` | Node counts, kind/tier breakdowns | `CognitiveMetrics` + `Status` |

## Client Setup

<summary>Claude Code</summary>

```bash
claude mcp add libravdb-memory -- mcp-memory-libravdb stdio
```

Or `.mcp.json`:

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

<summary>Claude Desktop</summary>

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

<summary>Codex</summary>

```toml
[mcp_servers.libravdb-memory]
command = "mcp-memory-libravdb"
args = ["stdio"]
```

<summary>Cursor</summary>

`~/.cursor/mcp.json`:

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

<summary>Cline</summary>

Settings → MCP Servers → add stdio: `mcp-memory-libravdb stdio`

<summary>Zed</summary>

`~/.config/zed/settings.json`:

```json
{
  "context_servers": {
    "libravdb-memory": {
      "command": { "path": "mcp-memory-libravdb", "args": ["stdio"] }
    }
  }
}
```

## Architecture

```
Claude Code / Codex / Cursor / Cline / Zed / Aider
    │
    │  MCP over stdio or Streamable HTTP (nanite/sse, off-heap)
    │
    └──→ mcp-memory-libravdb
              │
              │  gRPC over Unix socket (HMAC auth + tenant isolation)
              │
              └──→ libravdbd
                        ├── Cognitive classification (6 kinds)
                        ├── Deontic two-pass gating (H/R/DNL + P/A/DTech)
                        ├── TopoRegistry (directed causal graph)
                        ├── Workspace IAM (multi-tenant, shared + private collections)
                        └── Embedding cache (ONNX / remote API)
```

The daemon does the intelligence. This server translates protocols.

## Use Cases

**Code agents with shared memory.** Claude Code, Codex, and Aider in the same project all see each other's decisions, constraints, and architecture choices through the shared workspace graph.

**Chat agents with read access.** OpenClaw or Hermes connect directly to the daemon with a chat tenant, getting read-only access to the code agent's workspace — answering "what did Claude build last week?" with full context.

**Per-project isolation.** Each project directory gets its own tenant automatically via CWD hashing. No configuration needed. Explicitly share with `--shared`.

## Development

```bash
make          # lint + test-race + coverage + build
make test     # go test ./...
make build    # build binary
make coverage # test with coverage report
```

## Ecosystem

- [libravdb.com](https://libravdb.com) — official website
- [OpenClaw memory plugin](https://github.com/xDarkicex/openclaw-memory-libravdb) — TypeScript MCP client
- [Hermes memory plugin](https://github.com/xDarkicex/hermes-memory-libravdb) — Python MCP client
- [Discord](https://discord.gg/x4cu4RA2p) — community

## License

MIT © 2026 xDarkicex
