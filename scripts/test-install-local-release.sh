#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

VERSION="v9.9.9"

case "$(uname -s)" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *) echo "unsupported test os" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "unsupported test arch" >&2; exit 1 ;;
esac

ASSET_NAME="skill-dl_${os}_${arch}.tar.gz"
DOWNLOAD_DIR="${TMP_DIR}/releases/download/${VERSION}"
INSTALL_DIR="${TMP_DIR}/bin"

mkdir -p "${DOWNLOAD_DIR}" "${INSTALL_DIR}"

go build -o "${TMP_DIR}/skill-dl" "${ROOT_DIR}/cmd/skill-dl"
tar -C "${TMP_DIR}" -czf "${DOWNLOAD_DIR}/${ASSET_NAME}" skill-dl

(
  cd "${DOWNLOAD_DIR}"
  shasum -a 256 "${ASSET_NAME}" > checksums.txt
)

SKILL_DL_RELEASES_URL="file://${TMP_DIR}/releases" \
SKILL_DL_VERSION="${VERSION}" \
SKILL_DL_INSTALL_DIR="${INSTALL_DIR}" \
bash "${ROOT_DIR}/install.sh"

test -x "${INSTALL_DIR}/skill-dl"
"${INSTALL_DIR}/skill-dl" --version | grep -qx "skill-dl v1.3.0"

echo "local-install-ok"
