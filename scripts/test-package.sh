#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
PACKAGE="$PROJECT_DIR/dist/zima-display.raw"
STAGE_PACKAGE="$PROJECT_DIR/dist/zima-display-stage.tar.gz"
VERSION=$(tr -d '[:space:]' < "$PROJECT_DIR/VERSION")
TEST_DIR=$(mktemp -d "${TMPDIR:-/tmp}/zima-display-package.XXXXXX")
trap 'rm -rf "$TEST_DIR"' EXIT INT TERM

command -v unsquashfs >/dev/null 2>&1 || { echo "unsquashfs is required" >&2; exit 1; }
[ -f "$PACKAGE" ] || { echo "missing $PACKAGE; run ./scripts/build.sh amd64 first" >&2; exit 1; }
[ -f "$STAGE_PACKAGE" ] || { echo "missing $STAGE_PACKAGE; run ./scripts/build.sh amd64 first" >&2; exit 1; }

unsquashfs -quiet -d "$TEST_DIR/raw" "$PACKAGE"

for path in \
    usr/bin/zima-displayd \
    usr/bin/zima-displayctl \
    usr/libexec/zima-display/start-renderer \
    usr/lib/systemd/system/zima-display.service \
    usr/lib/systemd/system/zima-display-renderer.service \
    usr/share/casaos/www/modules/zima-display/index.html; do
    [ -f "$TEST_DIR/raw/$path" ] || { echo "package is missing $path" >&2; exit 1; }
done

[ -x "$TEST_DIR/raw/usr/bin/zima-displayd" ]
[ -x "$TEST_DIR/raw/usr/bin/zima-displayctl" ]
[ -x "$TEST_DIR/raw/usr/libexec/zima-display/start-renderer" ]
cmp "$PROJECT_DIR/packaging/raw/usr/libexec/zima-display/start-renderer" "$TEST_DIR/raw/usr/libexec/zima-display/start-renderer"
cmp "$PROJECT_DIR/packaging/raw/usr/lib/systemd/system/zima-display-renderer.service" "$TEST_DIR/raw/usr/lib/systemd/system/zima-display-renderer.service"
grep -qx 'TTYVHangup=no' "$TEST_DIR/raw/usr/lib/systemd/system/zima-display-renderer.service" || {
    echo "renderer must not hang up its own HDMI terminal" >&2
    exit 1
}
for console in getty@tty1.service zimaos-welcome@tty1.service; do
    for directive in Conflicts After; do
        if ! sed -n "s/^$directive=//p" "$TEST_DIR/raw/usr/lib/systemd/system/zima-display-renderer.service" | tr ' ' '\n' | grep -Fxq "$console"; then
            echo "renderer must have $directive=$console for exclusive, ordered tty1 ownership" >&2
            exit 1
        fi
    done
done

case "$(uname -s)" in
    Linux)
        [ "$("$TEST_DIR/raw/usr/bin/zima-displayd" --version)" = "$VERSION" ]
        [ "$("$TEST_DIR/raw/usr/bin/zima-displayctl" --version)" = "$VERSION" ]
        ;;
    *)
        strings "$TEST_DIR/raw/usr/bin/zima-displayd" | grep -Fqx "$VERSION"
        strings "$TEST_DIR/raw/usr/bin/zima-displayctl" | grep -Fqx "$VERSION"
        ;;
esac

tar -tzf "$STAGE_PACKAGE" | grep -Fqx './raw/usr/libexec/zima-display/start-renderer' || \
    tar -tzf "$STAGE_PACKAGE" | grep -Fqx 'raw/usr/libexec/zima-display/start-renderer'

echo "package test passed for version $VERSION"
