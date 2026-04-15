#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${1:-${ROOT_DIR}/dist}"

expected_assets=(
  "skill-dl_linux_amd64.tar.gz"
  "skill-dl_linux_arm64.tar.gz"
  "skill-dl_darwin_amd64.tar.gz"
  "skill-dl_darwin_arm64.tar.gz"
  "checksums.txt"
)

for asset in "${expected_assets[@]}"; do
  test -f "${DIST_DIR}/${asset}"
done

for asset in "${expected_assets[@]}"; do
  [[ "${asset}" == checksums.txt ]] && continue
  tar -tzf "${DIST_DIR}/${asset}" | grep -qx "skill-dl"
done

echo "release-layout-ok"
