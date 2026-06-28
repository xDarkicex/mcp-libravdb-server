#!/bin/bash
set -euo pipefail

VERSION="$1"
DIST_DIR="$2"
OUTPUT="$3"

DARWIN_ARM_SHA=$(shasum -a 256 "$DIST_DIR/mcp-memory-libravdb-darwin-arm64" | awk '{print $1}')
DARWIN_AMD_SHA=$(shasum -a 256 "$DIST_DIR/mcp-memory-libravdb-darwin-amd64" | awk '{print $1}')
LINUX_AMD_SHA=$(shasum -a 256 "$DIST_DIR/mcp-memory-libravdb-linux-amd64" | awk '{print $1}')
LINUX_ARM_SHA=$(shasum -a 256 "$DIST_DIR/mcp-memory-libravdb-linux-arm64" | awk '{print $1}')

cat > "$OUTPUT" <<'HEREDOC_END'
class McpMemoryLibravdb < Formula
  desc "MCP server for the libravdbd cognitive memory kernel"
  homepage "https://github.com/xDarkicex/mcp-libravdb-server"
  version "VERSION_PLACEHOLDER"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/xDarkicex/mcp-libravdb-server/releases/download/vVERSION_PLACEHOLDER/mcp-memory-libravdb-darwin-arm64"
      sha256 "DARWIN_ARM_SHA_PLACEHOLDER"
    else
      url "https://github.com/xDarkicex/mcp-libravdb-server/releases/download/vVERSION_PLACEHOLDER/mcp-memory-libravdb-darwin-amd64"
      sha256 "DARWIN_AMD_SHA_PLACEHOLDER"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/xDarkicex/mcp-libravdb-server/releases/download/vVERSION_PLACEHOLDER/mcp-memory-libravdb-linux-arm64"
      sha256 "LINUX_ARM_SHA_PLACEHOLDER"
    else
      url "https://github.com/xDarkicex/mcp-libravdb-server/releases/download/vVERSION_PLACEHOLDER/mcp-memory-libravdb-linux-amd64"
      sha256 "LINUX_AMD_SHA_PLACEHOLDER"
    end
  end

  def install
    bin.install Dir["*"].first => "mcp-memory-libravdb"
  end

  test do
    system "#{bin}/mcp-memory-libravdb", "--version"
  end
end
HEREDOC_END

sed -i '' "s/VERSION_PLACEHOLDER/$VERSION/g" "$OUTPUT" 2>/dev/null || sed -i "s/VERSION_PLACEHOLDER/$VERSION/g" "$OUTPUT"
sed -i '' "s/DARWIN_ARM_SHA_PLACEHOLDER/$DARWIN_ARM_SHA/" "$OUTPUT" 2>/dev/null || sed -i "s/DARWIN_ARM_SHA_PLACEHOLDER/$DARWIN_ARM_SHA/" "$OUTPUT"
sed -i '' "s/DARWIN_AMD_SHA_PLACEHOLDER/$DARWIN_AMD_SHA/" "$OUTPUT" 2>/dev/null || sed -i "s/DARWIN_AMD_SHA_PLACEHOLDER/$DARWIN_AMD_SHA/" "$OUTPUT"
sed -i '' "s/LINUX_AMD_SHA_PLACEHOLDER/$LINUX_AMD_SHA/" "$OUTPUT" 2>/dev/null || sed -i "s/LINUX_AMD_SHA_PLACEHOLDER/$LINUX_AMD_SHA/" "$OUTPUT"
sed -i '' "s/LINUX_ARM_SHA_PLACEHOLDER/$LINUX_ARM_SHA/" "$OUTPUT" 2>/dev/null || sed -i "s/LINUX_ARM_SHA_PLACEHOLDER/$LINUX_ARM_SHA/" "$OUTPUT"
