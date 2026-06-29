# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-06-29

### Added

- `memory.add` now accepts an optional `metadata` field for attaching agent identity, model, and source provenance to memory records.
- `.mcp.json` configuration file for Claude Code MCP server setup.

### Changed

- Search and recall results are now wrapped in structured response types (`SearchResponse`, `RecallResponse`) for MCP SDK compatibility.
- Removed `--shared` flag from default configuration to enable per-workspace tenant isolation via CWD hash.

## [0.1.0] - 2026-06-24

### Added

- Initial release of `mcp-memory-libravdb`, an MCP server wrapping libravdbd's cognitive memory system.
- MCP tools: `memory.add`, `memory.search`, `memory.recall`, `memory.delete`, `memory.graph`, `memory.stats`, `memory.predict`.
- MCP resources: `memory://collections`, `memory://kinds`, `memory://graph/:containerTag`.
- Stdio and HTTP+SSE transport with nanite zero-allocation router.
- gRPC client to libravdbd with tenant routing, HMAC auth, and retry interceptor.
- Per-workspace tenant isolation via SHA-256 CWD hash.
- Release workflow: multi-platform Go builds, Homebrew tap, APT repository, `.deb` packaging.
- Test coverage at 90% across tools, gRPC, resources, and app packages.
- MIT license.

[0.1.1]: https://github.com/xDarkicex/mcp-libravdb-server/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/xDarkicex/mcp-libravdb-server/releases/tag/v0.1.0
