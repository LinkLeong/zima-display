#!/bin/sh
set -eu

# /etc must carry this helper: /usr disappears when the .raw is not merged.
# The optional root is used by the isolated boot regression tests.
ROOT=${ZIMA_DISPLAY_BOOT_ROOT:-}
DAEMON="$ROOT/usr/bin/zima-displayd"
LINK="$ROOT/run/extensions/zima-display.raw"

if ! mountpoint -q "$ROOT/DATA"; then
    echo 'zima-display boot: persistent data is not mounted yet' >&2
    exit 1
fi

if [ ! -x "$DAEMON" ]; then
    IMAGE=
    for candidate in \
        "$ROOT/DATA/.extensions/zima-display.raw" \
        "$ROOT/var/lib/casaos_data/.extensions/zima-display.raw"; do
        if [ -r "$candidate" ]; then
            IMAGE=$candidate
            break
        fi
    done
    if [ -z "$IMAGE" ]; then
        echo 'zima-display boot: the persistent package is not available yet' >&2
        exit 1
    fi
    if [ -L "$ROOT/run/extensions" ]; then
        echo 'zima-display boot: refusing a symlinked extension directory' >&2
        exit 1
    fi
    mkdir -p "$ROOT/run/extensions"
    if [ -e "$LINK" ] || [ -L "$LINK" ]; then
        if [ "$(readlink -f "$LINK")" != "$(readlink -f "$IMAGE")" ]; then
            echo 'zima-display boot: refusing to replace a different extension image' >&2
            exit 1
        fi
    else
        ln -s "$IMAGE" "$LINK"
    fi
    echo 'zima-display boot: recovering the missing system extension registration'
    systemd-sysext refresh
fi

if [ ! -x "$DAEMON" ]; then
    echo 'zima-display boot: the extension did not expose zima-displayd' >&2
    exit 1
fi

if ! systemctl is-active --quiet zima-display.service; then
    # Also clear an earlier cached "unit not found" after another process
    # merged the image before this check.
    systemctl daemon-reload
    systemctl start zima-display.service
fi
