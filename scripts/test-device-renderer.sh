#!/bin/sh
set -eu

SOCKET=${ZIMA_DISPLAY_SOCKET:-/run/zima-display/mpv.sock}
CONFIG=${ZIMA_DISPLAY_CONFIG:-/DATA/AppData/zima-display/config.json}
API_BASE=${ZIMA_DISPLAY_API_BASE:-http://127.0.0.1/zima-display/api}
EXPECTED_BACKEND=${ZIMA_DISPLAY_EXPECT_BACKEND:-drm}
STABILITY_SECONDS=${ZIMA_DISPLAY_STABILITY_SECONDS:-60}
EXPECTED_VERSION=${1:-}

fail() {
    echo "FAIL: $*" >&2
    exit 1
}

require_command() {
    command -v "$1" >/dev/null 2>&1 || fail "$1 is required"
}

mpv_ipc_ready() {
    python3 - "$SOCKET" <<'PY'
import json
import socket
import sys

client = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
client.settimeout(2)
client.connect(sys.argv[1])
client.sendall(b'{"command":["get_property","idle-active"],"request_id":1}\n')
response = b""
while not response.endswith(b"\n"):
    chunk = client.recv(4096)
    if not chunk:
        raise SystemExit(1)
    response += chunk
payload = json.loads(response)
raise SystemExit(0 if payload.get("error") == "success" else 1)
PY
}

wait_for_renderer() {
    remaining=20
    while [ "$remaining" -gt 0 ]; do
        if mpv_ipc_ready >/dev/null 2>&1; then
            return 0
        fi
        remaining=$((remaining - 1))
        sleep 1
    done
    return 1
}

api_mode() {
    python3 - "$API_BASE/action" "$1" <<'PY'
import json
import sys
import urllib.request

payload = json.dumps({"action": "mode", "mode": sys.argv[2]}).encode()
request = urllib.request.Request(sys.argv[1], data=payload, headers={"Content-Type": "application/json"})
with urllib.request.urlopen(request, timeout=30) as response:
    if response.status != 200:
        raise SystemExit(1)
PY
}

status_renderer_ready() {
    python3 - "$API_BASE/status" <<'PY'
import json
import sys
import urllib.request

with urllib.request.urlopen(sys.argv[1], timeout=5) as response:
    payload = json.load(response)
raise SystemExit(0 if payload.get("player", {}).get("renderer_ready") is True else 1)
PY
}

api_health_version() {
    python3 - "$API_BASE/health" "$EXPECTED_VERSION" <<'PY'
import json
import sys
import urllib.request

with urllib.request.urlopen(sys.argv[1], timeout=5) as response:
    payload = json.load(response)
raise SystemExit(0 if payload.get("status") == "ok" and payload.get("version") == sys.argv[2] else 1)
PY
}

find_renderer_process() {
    RENDERER_PID=
    RENDERER_COMMAND=
    for candidate in $(pgrep -x mpv || true); do
        [ -r "/proc/$candidate/cmdline" ] || continue
        command_line=$(tr '\000' ' ' < "/proc/$candidate/cmdline")
        case "$command_line" in
            *"--input-ipc-server=$SOCKET"*)
                RENDERER_PID=$candidate
                RENDERER_COMMAND=$command_line
                return 0
                ;;
        esac
    done
    return 1
}

require_command systemctl
require_command python3
require_command pgrep
require_command stat
require_command tr
[ "$(id -u)" -eq 0 ] || fail "run this test as root on the ZimaOS device"
[ -n "$EXPECTED_VERSION" ] || fail "pass the candidate version as the first argument"
[ -r "$CONFIG" ] || fail "cannot read $CONFIG"

backend=$(python3 - "$CONFIG" <<'PY'
import json
import sys
with open(sys.argv[1], encoding="utf-8") as config_file:
    print(json.load(config_file).get("renderer", {}).get("backend", ""))
PY
)
[ "$backend" = "$EXPECTED_BACKEND" ] || fail "renderer backend is $backend; expected $EXPECTED_BACKEND"

installed_version=$(/usr/bin/zima-displayd --version)
[ "$installed_version" = "$EXPECTED_VERSION" ] || fail "installed version is $installed_version; expected $EXPECTED_VERSION"
[ "$(stat -c '%a' "$CONFIG")" = 600 ] || fail "$CONFIG must have mode 600"
echo "Testing zima-displayd $installed_version with backend $backend"
systemctl restart zima-display.service
wait_for_renderer || fail "mpv IPC was not ready within 20 seconds after service restart"
systemctl is-active --quiet zima-display.service || fail "zima-display.service is not active"
systemctl is-active --quiet zima-display-renderer.service || fail "zima-display-renderer.service is not active"
api_health_version || fail "health endpoint did not report candidate version $EXPECTED_VERSION"

for cycle in 1 2 3; do
    systemctl restart zima-display-renderer.service
    wait_for_renderer || fail "renderer restart cycle $cycle did not become IPC-ready within 20 seconds"
done

for mode in dashboard canvas clock black dashboard; do
    api_mode "$mode" || fail "mode request failed: $mode"
    wait_for_renderer || fail "mpv IPC was not ready after mode request: $mode"
    status_renderer_ready || fail "API did not report renderer_ready after mode request: $mode"
done

api_mode terminal || fail "terminal mode request failed"
remaining=10
while systemctl is-active --quiet zima-display-renderer.service && [ "$remaining" -gt 0 ]; do
    remaining=$((remaining - 1))
    sleep 1
done
systemctl is-active --quiet zima-display-renderer.service && fail "renderer remained active in terminal mode"
systemctl is-active --quiet getty@tty1.service || fail "getty@tty1.service is not active in terminal mode"
# Welcome uses Type=idle and may acquire tty1 after systemctl reports it started.
# Immediately taking the screen back reproduces the boot-time console race.
welcome_state=$(systemctl show zimaos-welcome@tty1.service --property=LoadState --value)
if [ "$welcome_state" = loaded ]; then
    systemctl start zimaos-welcome@tty1.service || fail "cannot start ZimaOS Welcome for the console handoff regression"
    systemctl is-active --quiet zimaos-welcome@tty1.service || fail "ZimaOS Welcome did not start"
fi
api_mode dashboard || fail "dashboard request failed after terminal mode"
wait_for_renderer || fail "renderer did not recover after terminal mode"
for console in getty@tty1.service zimaos-welcome@tty1.service; do
    if systemctl is-active --quiet "$console"; then
        fail "$console is still competing for tty1 with the renderer"
    fi
done
handoff_pid=$(systemctl show zima-display-renderer.service --property=MainPID --value)
sleep 6
[ "$(systemctl show zima-display-renderer.service --property=MainPID --value)" = "$handoff_pid" ] || fail "renderer exited after the delayed console handoff"
systemctl is-active --quiet zima-display-renderer.service || fail "renderer stopped after the delayed console handoff"
status_renderer_ready || fail "API did not report renderer_ready after the delayed console handoff"
echo "Exclusive tty1 console handoff passed"

find_renderer_process || fail "cannot find the mpv renderer process"
kill -9 "$RENDERER_PID"
api_mode dashboard || fail "dashboard request did not recover the killed mpv process"
wait_for_renderer || fail "mpv IPC did not recover after fault injection"
status_renderer_ready || fail "API did not report renderer_ready after fault injection"

if [ -e /run/zima-display/software-renderer ]; then
    find_renderer_process || fail "cannot inspect the live mpv command"
    echo "$RENDERER_COMMAND" | grep -q -- '--vo=drm' || fail "fallback marker exists but mpv is not using --vo=drm"
    echo "$RENDERER_COMMAND" | grep -q -- '--hwdec=no' || fail "fallback marker exists but mpv is not using --hwdec=no"
    if echo "$RENDERER_COMMAND" | grep -q -- '--vo=gpu-next'; then
        fail "fallback marker exists but mpv is still using gpu-next"
    fi
    echo "Software DRM fallback is active"
else
    echo "GPU renderer remained healthy; no software fallback marker was created"
fi

elapsed=0
while [ "$elapsed" -lt "$STABILITY_SECONDS" ]; do
    sleep 5
    elapsed=$((elapsed + 5))
    mpv_ipc_ready || fail "mpv IPC stopped responding during the stability window at ${elapsed}s"
done

echo "PASS: renderer recovery and ${STABILITY_SECONDS}s stability checks completed"
