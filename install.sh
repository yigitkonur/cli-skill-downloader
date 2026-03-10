#!/usr/bin/env bash
# ==============================================================================
# install.sh — Install skill-dl
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/yigitkonur/cli-skill-downloader/main/install.sh | bash
# ==============================================================================
set -euo pipefail

REPO="yigitkonur/cli-skill-downloader"
BRANCH="main"
BINARY="skill-dl"
RAW_URL="https://raw.githubusercontent.com/${REPO}/${BRANCH}/${BINARY}"

# Colors
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

# Check OS
OS="$(uname -s)"
if [[ "$OS" != "Linux" && "$OS" != "Darwin" ]]; then
  die "Unsupported OS: $OS. Only Linux and macOS are supported."
fi

# Check bash version (need 4+ for associative arrays)
if [[ "${BASH_VERSINFO[0]}" -lt 4 ]]; then
  if [[ "$OS" == "Darwin" ]]; then
    die "skill-dl requires Bash 4+. macOS ships Bash 3.\nFix: brew install bash"
  else
    die "skill-dl requires Bash 4+. Current: ${BASH_VERSION}"
  fi
fi

# Check for git
command -v git >/dev/null 2>&1 || die "git is required but not found. Install git first."

# Check for download tool
if command -v curl >/dev/null 2>&1; then
  DOWNLOADER="curl"
elif command -v wget >/dev/null 2>&1; then
  DOWNLOADER="wget"
else
  die "curl or wget is required but neither was found."
fi

# Determine install directory
INSTALL_DIR=""
NEED_SUDO=false

if [[ -w "/usr/local/bin" ]]; then
  INSTALL_DIR="/usr/local/bin"
elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
  INSTALL_DIR="/usr/local/bin"
  NEED_SUDO=true
else
  # Fallback to user-local
  INSTALL_DIR="${HOME}/.local/bin"
  mkdir -p "$INSTALL_DIR"
  warn "No write access to /usr/local/bin, installing to ~/.local/bin"
  if [[ ":$PATH:" != *":${HOME}/.local/bin:"* ]]; then
    warn "Add ~/.local/bin to your PATH:"
    warn "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc  # or ~/.zshrc"
  fi
fi

TARGET="${INSTALL_DIR}/${BINARY}"

echo ""
echo -e "${BOLD}Installing skill-dl${RESET}"
echo -e "${DIM}From: ${RAW_URL}${RESET}"
echo -e "${DIM}  To: ${TARGET}${RESET}"
echo ""

# Download
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

info "Downloading..."
if [[ "$DOWNLOADER" == "curl" ]]; then
  curl -fsSL "$RAW_URL" -o "$TMP" || die "Download failed. Check your internet connection."
else
  wget -qO "$TMP" "$RAW_URL" || die "Download failed. Check your internet connection."
fi

# Validate it looks like a bash script
head -1 "$TMP" | grep -q "bash" || die "Downloaded file does not look like a shell script. Aborting."

chmod +x "$TMP"

# Install
if [[ "$NEED_SUDO" == "true" ]]; then
  info "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "$TMP" "$TARGET"
  sudo chmod +x "$TARGET"
else
  info "Installing to ${INSTALL_DIR}..."
  mv "$TMP" "$TARGET"
  chmod +x "$TARGET"
fi

# Verify
if command -v "$BINARY" >/dev/null 2>&1; then
  version=$("$BINARY" --version 2>/dev/null || echo "unknown")
  echo ""
  success "skill-dl installed successfully! (${version})"
  echo ""
  echo -e "${DIM}Get started:${RESET}"
  echo "  skill-dl https://playbooks.com/skills/mcollina/skills/typescript-magician"
  echo "  skill-dl --help"
  echo ""
else
  echo ""
  warn "Installed to ${TARGET}, but 'skill-dl' is not in your PATH."
  warn "You can run it directly: ${TARGET}"
  echo ""
fi
