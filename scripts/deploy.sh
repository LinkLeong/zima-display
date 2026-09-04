#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
TARGET=${1:-${ZIMA_DISPLAY_TARGET:-}}
REMOTE_STAGE=/tmp/zima-display-stage.tar.gz

if [ -z "$TARGET" ]; then
    echo "Usage: $0 user@zimaos-host" >&2
    echo "Or set ZIMA_DISPLAY_TARGET." >&2
    exit 2
fi

"$SCRIPT_DIR/build.sh" amd64
ssh "$TARGET" "umask 077; cat > '$REMOTE_STAGE'" < "$PROJECT_DIR/dist/zima-display-stage.tar.gz"

ssh -t "$TARGET" "set -eu
deploy_dir=\$(mktemp -d /tmp/zima-display.XXXXXX)
trap 'rm -rf \"\$deploy_dir\"' EXIT
tar -xzf '$REMOTE_STAGE' -C \"\$deploy_dir\"
command -v mksquashfs >/dev/null 2>&1 || { echo 'mksquashfs is required on the ZimaOS host' >&2; exit 1; }
mksquashfs \"\$deploy_dir/raw\" \"\$deploy_dir/zima-display.raw\" -noappend -comp gzip -all-root -no-xattrs -quiet
sudo sh \"\$deploy_dir/install.sh\" \"\$deploy_dir/zima-display.raw\"
"
