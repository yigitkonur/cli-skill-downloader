#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REFERENCE_SCRIPT="${ROOT_DIR}/testdata/reference/skill-dl-bash"
BUILD_DIR="$(mktemp -d)"
trap 'rm -rf "${BUILD_DIR}"' EXIT

GO_BINARY="${BUILD_DIR}/skill-dl.go"
REFERENCE_BINARY="${BUILD_DIR}/skill-dl"

go build -o "${GO_BINARY}" "${ROOT_DIR}/cmd/skill-dl"
cp "${REFERENCE_SCRIPT}" "${REFERENCE_BINARY}"
chmod +x "${REFERENCE_BINARY}"

check_case() {
  local name="$1"
  shift

  local go_stdout="${BUILD_DIR}/${name}.go.stdout"
  local go_stderr="${BUILD_DIR}/${name}.go.stderr"
  local ref_stdout="${BUILD_DIR}/${name}.ref.stdout"
  local ref_stderr="${BUILD_DIR}/${name}.ref.stderr"

  set +e
  "${GO_BINARY}" "$@" >"${go_stdout}" 2>"${go_stderr}"
  local go_exit=$?
  "${REFERENCE_BINARY}" "$@" >"${ref_stdout}" 2>"${ref_stderr}"
  local ref_exit=$?
  set -e

  diff -u "${ref_stdout}" "${go_stdout}" >/dev/null
  diff -u "${ref_stderr}" "${go_stderr}" >/dev/null
  test "${go_exit}" = "${ref_exit}"
}

check_case help --help
check_case version --version
check_case no-args
check_case invalid-source definitely-not-a-real-source
check_case search-help search --help
check_case search-too-few-keywords search typescript react
check_case dry-run-single-url https://playbooks.com/skills/mcollina/skills/typescript-magician --dry-run
check_case dry-run-mixed-sources "${ROOT_DIR}/examples/typescript-skills.txt" https://playbooks.com/skills/nickcrew/claude-cortex/typescript-advanced-patterns --dry-run --no-auto-category

echo "reference-parity-ok"
