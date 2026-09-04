#!/bin/sh
set -eu

PACKAGE=${1:-./zima-display.raw}
if [ "$(id -u)" -ne 0 ]; then
    exec sudo "$0" "$PACKAGE"
fi
if [ ! -f "$PACKAGE" ]; then
    echo "Package not found: $PACKAGE" >&2
    exit 1
fi

zpkg install --force "$PACKAGE"
systemctl daemon-reload
systemctl enable zima-display.service
systemctl restart zima-display.service
echo "Zima Display installed. Open it from the ZimaOS sidebar."
