#!/usr/bin/env bash
# ==============================================================================
# install.sh — Install a prebuilt skill-dl release archive
#
# Usage:
#   sudo -v ; curl https://raw.githubusercontent.com/yigitkonur/cli-skill-downloader/main/install.sh | sudo bash
#   sudo -v ; curl https://raw.githubusercontent.com/yigitkonur/cli-skill-downloader/main/install.sh | sudo bash -s -- v1.3.0
# ==============================================================================
set -euo pipefail

REPO_SLUG="${SKILL_DL_REPO_SLUG:-yigitkonur/cli-skill-downloader}"
RELEASES_URL="${SKILL_DL_RELEASES_URL:-https://github.com/${REPO_SLUG}/releases}"
BINARY="skill-dl"
INSTALL_DIR="${SKILL_DL_INSTALL_DIR:-/usr/local/bin}"
REQUESTED_VERSION="${SKILL_DL_VERSION:-${1:-latest}}"

if [[ $# -gt 1 ]]; then
  echo "usage: install.sh [latest|vX.Y.Z]" >&2
  exit 1
fi

if [[ -t 1 ]]; then
  BOLD='\033[1m'; GREEN='\033[0;32m'; RED='\033[0;31m'
  YELLOW='\033[0;33m'; DIM='\033[2m'; RESET='\033[0m'
else
  BOLD=''; GREEN=''; RED=''; YELLOW=''; DIM=''; RESET=''
fi

info()    { echo -e "${GREEN}▶${RESET} $*"; }
warn()    { echo -e "${YELLOW}!${RESET} $*"; }
error()   { echo -e "${RED}✗${RESET} $*" >&2; }
success() { echo -e "${GREEN}✓${RESET} $*"; }
die()     { error "$*"; exit 1; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "$1 is required but not found."
}

detect_os() {
  case "$(uname -s)" in
    Linux) echo "linux" ;;
    Darwin) echo "darwin" ;;
    *) die "Unsupported OS: $(uname -s). Only Linux and macOS are supported." ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) die "Unsupported architecture: $(uname -m). Only amd64 and arm64 are supported." ;;
  esac
}

resolve_version() {
  if [[ "${REQUESTED_VERSION}" != "latest" ]]; then
    echo "${REQUESTED_VERSION}"
    return
  fi

  if [[ -n "${SKILL_DL_LATEST_VERSION:-}" ]]; then
    echo "${SKILL_DL_LATEST_VERSION}"
    return
  fi

  [[ "${RELEASES_URL}" == file://* ]] && die "Set SKILL_DL_VERSION when using a file:// release source."

  local latest_url
  latest_url="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "${RELEASES_URL}/latest")" || die "Unable to resolve the latest release."
  local latest_version="${latest_url##*/}"
  [[ -n "${latest_version}" ]] || die "Unable to parse the latest release version."
  echo "${latest_version}"
}

current_installed_version() {
  local target="${INSTALL_DIR}/${BINARY}"
  local candidate=""

  if [[ -x "${target}" ]]; then
    candidate="${target}"
  elif command -v "${BINARY}" >/dev/null 2>&1; then
    candidate="$(command -v "${BINARY}")"
  fi

  [[ -z "${candidate}" ]] && return 0
  "${candidate}" --version 2>/dev/null | head -n 1 | sed -n 's/^skill-dl //p'
}

verify_checksum() {
  local asset_path="$1"
  local checksums_path="$2"
  local asset_name
  asset_name="$(basename "${asset_path}")"
  local expected=""

  expected="$(awk -v name="${asset_name}" '$2 == name {print $1}' "${checksums_path}")"
  [[ -n "${expected}" ]] || die "Could not find ${asset_name} in checksums.txt"

  local actual=""
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "${asset_path}" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "${asset_path}" | awk '{print $1}')"
  else
    warn "No checksum tool found (sha256sum/shasum). Skipping checksum verification."
    return
  fi

  [[ "${actual}" == "${expected}" ]] || die "Checksum verification failed for ${asset_name}"
}

require_cmd curl
require_cmd tar
require_cmd install

OS="$(detect_os)"
ARCH="$(detect_arch)"
VERSION="$(resolve_version)"
ASSET_NAME="${BINARY}_${OS}_${ARCH}.tar.gz"
ASSET_URL="${RELEASES_URL}/download/${VERSION}/${ASSET_NAME}"
CHECKSUMS_URL="${RELEASES_URL}/download/${VERSION}/checksums.txt"
TARGET="${INSTALL_DIR}/${BINARY}"

if [[ ! -d "${INSTALL_DIR}" ]]; then
  mkdir -p "${INSTALL_DIR}" 2>/dev/null || true
fi
[[ -w "${INSTALL_DIR}" ]] || die "Install directory ${INSTALL_DIR} is not writable. Use sudo or set SKILL_DL_INSTALL_DIR to a writable path."

INSTALLED_VERSION="$(current_installed_version || true)"
if [[ "${INSTALLED_VERSION}" == "${VERSION}" && "${SKILL_DL_FORCE:-0}" != "1" ]]; then
  success "skill-dl ${VERSION} is already installed."
  exit 0
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT
ARCHIVE_PATH="${TMP_DIR}/${ASSET_NAME}"
CHECKSUMS_PATH="${TMP_DIR}/checksums.txt"

echo ""
echo -e "${BOLD}Installing skill-dl${RESET}"
echo -e "${DIM}Version: ${VERSION}${RESET}"
echo -e "${DIM}Source:  ${ASSET_URL}${RESET}"
echo -e "${DIM}Target:  ${TARGET}${RESET}"
echo ""

info "Downloading release archive..."
curl -fsSL "${ASSET_URL}" -o "${ARCHIVE_PATH}" || die "Download failed for ${ASSET_URL}"

info "Downloading checksums..."
curl -fsSL "${CHECKSUMS_URL}" -o "${CHECKSUMS_PATH}" || die "Download failed for ${CHECKSUMS_URL}"

info "Verifying checksum..."
verify_checksum "${ARCHIVE_PATH}" "${CHECKSUMS_PATH}"

info "Extracting archive..."
tar -xzf "${ARCHIVE_PATH}" -C "${TMP_DIR}" || die "Failed to extract ${ASSET_NAME}"
[[ -x "${TMP_DIR}/${BINARY}" ]] || die "Archive did not contain ${BINARY} at the expected path."

info "Installing to ${INSTALL_DIR}..."
install -m 755 "${TMP_DIR}/${BINARY}" "${TARGET}.new"
mv "${TARGET}.new" "${TARGET}"

INSTALLED_VERSION="$("${TARGET}" --version 2>/dev/null || echo unknown)"

echo ""
success "skill-dl installed successfully! (${INSTALLED_VERSION})"
echo ""
echo -e "${DIM}Get started:${RESET}"
echo "  ${BINARY} https://playbooks.com/skills/mcollina/skills/typescript-magician"
echo "  ${BINARY} --help"
echo ""
