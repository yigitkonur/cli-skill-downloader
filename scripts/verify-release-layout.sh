#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${1:-${ROOT_DIR}/dist}"

expected_assets=(
  "skill-dl_linux_amd64.tar.gz"
  "skill-dl_linux_arm64.tar.gz"
  "skill-dl_darwin_amd64.tar.gz"
  "skill-dl_darwin_arm64.tar.gz"
  "skill-dl_windows_amd64.zip"
  "skill-dl_windows_arm64.zip"
  "checksums.txt"
)

for asset in "${expected_assets[@]}"; do
  test -f "${DIST_DIR}/${asset}"
done

for asset in "${expected_assets[@]}"; do
  [[ "${asset}" == checksums.txt ]] && continue
  case "${asset}" in
    *.tar.gz)
      tar -tzf "${DIST_DIR}/${asset}" | grep -qx "skill-dl"
      ;;
    *.zip)
      python3 - <<'PY' "${DIST_DIR}/${asset}"
import sys, zipfile
with zipfile.ZipFile(sys.argv[1]) as zf:
    names = set(zf.namelist())
    if "skill-dl.exe" not in names:
        raise SystemExit(1)
PY
      ;;
  esac
done

echo "release-layout-ok"
