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
INSTALL_DIR_EXPLICIT="${TMP_DIR}/bin-explicit"
INSTALL_DIR_LATEST="${TMP_DIR}/bin-latest"
FAKE_PATH_DIR="${TMP_DIR}/fake-path"

mkdir -p "${DOWNLOAD_DIR}" "${INSTALL_DIR_EXPLICIT}" "${INSTALL_DIR_LATEST}" "${FAKE_PATH_DIR}"

go build -o "${TMP_DIR}/skill-dl" "${ROOT_DIR}/cmd/skill-dl"
tar -C "${TMP_DIR}" -czf "${DOWNLOAD_DIR}/${ASSET_NAME}" skill-dl

(
  cd "${DOWNLOAD_DIR}"
  shasum -a 256 "${ASSET_NAME}" > checksums.txt
)

cat > "${FAKE_PATH_DIR}/skill-dl" <<'EOF'
#!/usr/bin/env bash
echo "skill-dl v9.9.9"
EOF
chmod +x "${FAKE_PATH_DIR}/skill-dl"

PATH="${FAKE_PATH_DIR}:$PATH" \
SKILL_DL_RELEASES_URL="file://${TMP_DIR}/releases" \
SKILL_DL_VERSION="${VERSION}" \
SKILL_DL_INSTALL_DIR="${INSTALL_DIR_EXPLICIT}" \
bash "${ROOT_DIR}/install.sh"

test -x "${INSTALL_DIR_EXPLICIT}/skill-dl"
"${INSTALL_DIR_EXPLICIT}/skill-dl" --version | grep -qx "skill-dl v1.3.0"

SERVER_SCRIPT="${TMP_DIR}/server.py"
cat > "${SERVER_SCRIPT}" <<PY
from functools import partial
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
import os

ROOT = Path(os.environ["SKILL_DL_TEST_SERVER_ROOT"])
VERSION = os.environ["SKILL_DL_TEST_VERSION"]

class Handler(SimpleHTTPRequestHandler):
    def do_HEAD(self):
        if self.path == "/releases/latest":
            self.send_response(302)
            self.send_header("Location", f"/releases/tag/{VERSION}")
            self.end_headers()
            return
        if self.path == f"/releases/tag/{VERSION}":
            self.send_response(200)
            self.end_headers()
            return
        super().do_HEAD()

    def do_GET(self):
        if self.path == "/releases/latest":
            self.send_response(302)
            self.send_header("Location", f"/releases/tag/{VERSION}")
            self.end_headers()
            return
        if self.path == f"/releases/tag/{VERSION}":
            payload = b"ok"
            self.send_response(200)
            self.send_header("Content-Type", "text/plain")
            self.send_header("Content-Length", str(len(payload)))
            self.end_headers()
            self.wfile.write(payload)
            return
        super().do_GET()

handler = partial(Handler, directory=str(ROOT))
server = ThreadingHTTPServer(("127.0.0.1", 0), handler)
print(server.server_address[1], flush=True)
server.serve_forever()
PY

PORT_FILE="${TMP_DIR}/server.port"
SKILL_DL_TEST_SERVER_ROOT="${TMP_DIR}" \
SKILL_DL_TEST_VERSION="${VERSION}" \
python3 "${SERVER_SCRIPT}" > "${PORT_FILE}" &
SERVER_PID=$!
trap 'kill "${SERVER_PID}" >/dev/null 2>&1 || true; rm -rf "${TMP_DIR}"' EXIT

for _ in $(seq 1 50); do
  if [[ -s "${PORT_FILE}" ]]; then
    break
  fi
  sleep 0.1
done

PORT="$(cat "${PORT_FILE}")"
PATH="${FAKE_PATH_DIR}:$PATH" \
SKILL_DL_RELEASES_URL="http://127.0.0.1:${PORT}/releases" \
SKILL_DL_INSTALL_DIR="${INSTALL_DIR_LATEST}" \
bash "${ROOT_DIR}/install.sh"

kill "${SERVER_PID}" >/dev/null 2>&1 || true

test -x "${INSTALL_DIR_LATEST}/skill-dl"
"${INSTALL_DIR_LATEST}/skill-dl" --version | grep -qx "skill-dl v1.3.0"

echo "local-install-ok"
