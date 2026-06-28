#!/bin/bash
set -euo pipefail

VERSION="$1"
ARCH="$2"
BINARY="$3"
OUTDIR="$4"

PKG_NAME=mcp-memory-libravdb
PKG_DIR="${OUTDIR}/${PKG_NAME}_${VERSION}_${ARCH}"
BIN_DIR="${PKG_DIR}/usr/local/bin"
DEBIAN_DIR="${PKG_DIR}/DEBIAN"

mkdir -p "${BIN_DIR}" "${DEBIAN_DIR}"
cp "${BINARY}" "${BIN_DIR}/mcp-memory-libravdb"
chmod 755 "${BIN_DIR}/mcp-memory-libravdb"

INSTALLED_SIZE=$(du -sk "${PKG_DIR}" | cut -f1)

cat > "${DEBIAN_DIR}/control" <<EOF
Package: ${PKG_NAME}
Version: ${VERSION}
Architecture: ${ARCH}
Maintainer: xDarkicex
Installed-Size: ${INSTALLED_SIZE}
Section: utils
Priority: optional
Homepage: https://github.com/xDarkicex/mcp-libravdb-server
Description: MCP server for the libravdbd cognitive memory kernel
 Exposes libravdbd's cognitive memory system to any MCP-compatible AI client.
EOF

DEB_NAME="${PKG_NAME}_${VERSION}_${ARCH}.deb"
dpkg-deb --build "${PKG_DIR}" "${OUTDIR}/${DEB_NAME}"
shasum -a 256 "${OUTDIR}/${DEB_NAME}" > "${OUTDIR}/${DEB_NAME}.sha256"
