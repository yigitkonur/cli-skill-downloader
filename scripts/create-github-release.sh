#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG="${1:-}"

if [[ -z "${TAG}" ]]; then
  echo "Usage: bash scripts/create-github-release.sh <tag>" >&2
  exit 1
fi

if [[ ! "${TAG}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Tag must look like v1.2.3" >&2
  exit 1
fi

VERSION="${TAG#v}"
CHANGELOG_FILE="${ROOT_DIR}/CHANGELOG.md"
NOTES_FILE="$(mktemp)"
trap 'rm -f "${NOTES_FILE}"' EXIT

cd "${ROOT_DIR}"

command -v gh >/dev/null 2>&1 || { echo "gh is required" >&2; exit 1; }
gh auth status >/dev/null

git diff --quiet
git diff --cached --quiet
git ls-files --error-unmatch cmd/skill-dl/main.go >/dev/null

if ! git rev-parse "${TAG}" >/dev/null 2>&1; then
  git tag -a "${TAG}" -m "Release ${TAG}"
fi

if ! git ls-remote --tags origin "${TAG}" | grep -q "${TAG}$"; then
  git push origin "${TAG}"
fi

awk -v version="${VERSION}" '
  $0 ~ "^## \\[" version "\\]" { print; in_section=1; next }
  in_section && /^## \[/ { exit }
  in_section { print }
' "${CHANGELOG_FILE}" > "${NOTES_FILE}"

if [[ ! -s "${NOTES_FILE}" ]]; then
  echo "Could not find CHANGELOG section for ${VERSION}" >&2
  exit 1
fi

go run github.com/goreleaser/goreleaser/v2@latest release --clean --skip=publish --skip=announce --skip=validate

gh release view "${TAG}" >/dev/null 2>&1 && {
  echo "GitHub release ${TAG} already exists" >&2
  exit 1
}

gh release create "${TAG}" \
  dist/skill-dl_linux_amd64.tar.gz \
  dist/skill-dl_linux_arm64.tar.gz \
  dist/skill-dl_darwin_amd64.tar.gz \
  dist/skill-dl_darwin_arm64.tar.gz \
  dist/skill-dl_windows_amd64.zip \
  dist/skill-dl_windows_arm64.zip \
  dist/checksums.txt \
  --title "${TAG}" \
  --notes-file "${NOTES_FILE}"
