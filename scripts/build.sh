#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
BUILD_DIR="$PROJECT_DIR/.build"
RAW_DIR="$BUILD_DIR/raw"
DIST_DIR="$PROJECT_DIR/dist"
ARCH=${1:-amd64}
VERSION=$(tr -d '[:space:]' < "$PROJECT_DIR/VERSION")
rm -f "$DIST_DIR/zima-display.raw"

case "$ARCH" in
    amd64|arm64) ;;
    *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

rm -rf "$BUILD_DIR"
mkdir -p "$RAW_DIR" "$DIST_DIR" "$RAW_DIR/usr/bin" "$RAW_DIR/usr/share/casaos/www/modules/zima-display"
cp -R "$PROJECT_DIR/packaging/raw/." "$RAW_DIR/"
cp "$PROJECT_DIR/web/index.html" "$PROJECT_DIR/web/styles.css" "$PROJECT_DIR/web/app.js" "$PROJECT_DIR/web/appicon.svg" "$RAW_DIR/usr/share/casaos/www/modules/zima-display/"

echo "Building zima-displayd for linux/$ARCH"
CGO_ENABLED=0 GOOS=linux GOARCH="$ARCH" go build -trimpath -ldflags="-s -w -X main.version=$VERSION" -o "$RAW_DIR/usr/bin/zima-displayd" "$PROJECT_DIR/cmd/zima-displayd"
CGO_ENABLED=0 GOOS=linux GOARCH="$ARCH" go build -trimpath -ldflags="-s -w -X main.version=$VERSION" -o "$RAW_DIR/usr/bin/zima-displayctl" "$PROJECT_DIR/cmd/zima-displayctl"
chmod 0755 "$RAW_DIR/usr/bin/zima-displayd" "$RAW_DIR/usr/bin/zima-displayctl" "$RAW_DIR/usr/libexec/zima-display/start-renderer"
cp "$PROJECT_DIR/scripts/install.sh" "$BUILD_DIR/install.sh"
chmod 0755 "$BUILD_DIR/install.sh"

STAGE_PACKAGE="$DIST_DIR/zima-display-stage.tar.gz"
rm -f "$STAGE_PACKAGE"
COPYFILE_DISABLE=1 tar --no-xattrs -C "$BUILD_DIR" -czf "$STAGE_PACKAGE" raw install.sh
echo "Created $STAGE_PACKAGE"

if [ "$(uname -s)" = Darwin ] || ! command -v mksquashfs >/dev/null 2>&1; then
    echo "Staged raw filesystem at $RAW_DIR"
    echo "Build the .raw package on Linux or ZimaOS for zpkg compatibility."
    exit 0
fi

PACKAGE="$DIST_DIR/zima-display.raw"
rm -f "$PACKAGE"
mksquashfs "$RAW_DIR" "$PACKAGE" -noappend -comp gzip -all-root -no-xattrs -quiet
echo "Created $PACKAGE"
