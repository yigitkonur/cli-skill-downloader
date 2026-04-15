#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if command -v goreleaser >/dev/null 2>&1; then
  GORELEASER_BIN="goreleaser"
else
  GORELEASER_BIN="go run github.com/goreleaser/goreleaser/v2@latest"
fi

cd "${ROOT_DIR}"
rm -rf dist

if [[ "${GORELEASER_BIN}" == goreleaser ]]; then
  goreleaser release --snapshot --clean
else
  go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean
fi
