#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

VERSION="${VERSION:-$(tr -d '\r\n' < ./cmd/server/VERSION)}"
LDFLAGS="${LDFLAGS:--s -w -X main.Version=${VERSION}}"
OUTPUT="${OUTPUT:-bin/server}"
PKG="${PKG:-./cmd/server}"

mkdir -p "$(dirname "$OUTPUT")"

echo "Building backend binary..."
echo "  pkg: ${PKG}"
echo "  output: ${OUTPUT}"
echo "  version: ${VERSION}"

CGO_ENABLED="${CGO_ENABLED:-0}" go build \
  -ldflags="${LDFLAGS}" \
  -trimpath \
  -o "${OUTPUT}" \
  "${PKG}"

echo "Build complete: ${OUTPUT}"
