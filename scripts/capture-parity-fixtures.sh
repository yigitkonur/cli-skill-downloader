#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REFERENCE_SCRIPT="${ROOT_DIR}/testdata/reference/skill-dl-bash"
FIXTURES_DIR="${ROOT_DIR}/testdata/parity"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

mkdir -p "${FIXTURES_DIR}"
rm -rf "${FIXTURES_DIR:?}/"*

run_case() {
  local name="$1"
  shift

  local stdout_file="${FIXTURES_DIR}/${name}.stdout"
  local stderr_file="${FIXTURES_DIR}/${name}.stderr"
  local exit_file="${FIXTURES_DIR}/${name}.exit"

  set +e
  (
    cd "${ROOT_DIR}"
    "$@"
  ) >"${stdout_file}" 2>"${stderr_file}"
  local exit_code=$?
  set -e

  printf '%s\n' "${exit_code}" >"${exit_file}"
}

URL_A="https://playbooks.com/skills/mcollina/skills/typescript-magician"
URL_B="https://playbooks.com/skills/nickcrew/claude-cortex/typescript-advanced-patterns"

run_case help "${REFERENCE_SCRIPT}" --help
run_case version "${REFERENCE_SCRIPT}" --version
run_case no-args "${REFERENCE_SCRIPT}"
run_case invalid-source "${REFERENCE_SCRIPT}" definitely-not-a-real-source
run_case search-help "${REFERENCE_SCRIPT}" search --help
run_case search-too-few-keywords "${REFERENCE_SCRIPT}" search typescript react
run_case dry-run-single-url "${REFERENCE_SCRIPT}" "${URL_A}" --dry-run
run_case dry-run-mixed-sources "${REFERENCE_SCRIPT}" examples/typescript-skills.txt "${URL_B}" --dry-run --no-auto-category

cat >"${FIXTURES_DIR}/manifest.txt" <<EOF
help
version
no-args
invalid-source
search-help
search-too-few-keywords
dry-run-single-url
dry-run-mixed-sources
EOF
