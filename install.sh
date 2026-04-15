#!/usr/bin/env bash
# ==============================================================================
# install.sh — Build and install the Go rewrite of skill-dl from a shallow clone
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/yigitkonur/cli-skill-downloader/main/install.sh | bash
# ==============================================================================
set -euo pipefail

REPO_URL="https://github.com/yigitkonur/cli-skill-downloader.git"
BINARY="skill-dl"

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

OS="$(uname -s)"
if [[ "$OS" != "Linux" && "$OS" != "Darwin" ]]; then
  die "Unsupported OS: $OS. Only Linux and macOS are supported."
fi

command -v git >/dev/null 2>&1 || die "git is required but not found. Install git first."
command -v go >/dev/null 2>&1 || die "Go is required but not found. Install Go 1.26+ first."

INSTALL_DIR=""
NEED_SUDO=false

if [[ -w "/usr/local/bin" ]]; then
  INSTALL_DIR="/usr/local/bin"
elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
  INSTALL_DIR="/usr/local/bin"
  NEED_SUDO=true
else
  INSTALL_DIR="${HOME}/.local/bin"
  mkdir -p "$INSTALL_DIR"
  warn "No write access to /usr/local/bin, installing to ~/.local/bin"
  if [[ ":$PATH:" != *":${HOME}/.local/bin:"* ]]; then
    warn "Add ~/.local/bin to your PATH:"
    warn "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc  # or ~/.zshrc"
  fi
fi

TARGET="${INSTALL_DIR}/${BINARY}"
TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT

echo ""
echo -e "${BOLD}Installing skill-dl${RESET}"
echo -e "${DIM}Repo: ${REPO_URL}${RESET}"
echo -e "${DIM}  To: ${TARGET}${RESET}"
echo ""

info "Cloning repository..."
git clone --depth 1 --quiet "${REPO_URL}" "${TMP_ROOT}/repo" || die "Clone failed. Check your internet connection."

info "Building Go binary..."
(
  cd "${TMP_ROOT}/repo"
  go build -o "${TMP_ROOT}/${BINARY}" ./cmd/skill-dl
) || die "Go build failed. Install Go 1.26+ and try again."

chmod +x "${TMP_ROOT}/${BINARY}"

if [[ "$NEED_SUDO" == "true" ]]; then
  info "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${TMP_ROOT}/${BINARY}" "${TARGET}"
  sudo chmod +x "${TARGET}"
else
  info "Installing to ${INSTALL_DIR}..."
  mv "${TMP_ROOT}/${BINARY}" "${TARGET}"
  chmod +x "${TARGET}"
fi

version="$("${TARGET}" --version 2>/dev/null || echo "unknown")"

echo ""
success "skill-dl installed successfully! (${version})"
echo ""
echo -e "${DIM}Get started:${RESET}"
echo "  skill-dl https://playbooks.com/skills/mcollina/skills/typescript-magician"
echo "  skill-dl --help"
echo ""

if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
  warn "Installed to ${TARGET}, but ${INSTALL_DIR} is not currently in PATH."
fi
