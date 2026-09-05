#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

mkdir -p dist

BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

go build -trimpath -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev) -X main.buildTime=${BUILD_TIME}" -o dist/jetkvm-desktop ./cmd/jetkvm-desktop

echo "Built: dist/jetkvm-desktop ($BUILD_TIME)"
