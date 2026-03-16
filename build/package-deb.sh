#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-0.1.0}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP_NAME="postman-lite"
PKG_ROOT="$ROOT/build/deb/${APP_NAME}_${VERSION}_amd64"
BIN_SRC="$ROOT/deliverables/bin/$APP_NAME"
OUT_DEB="$ROOT/deliverables/${APP_NAME}_${VERSION}_amd64.deb"
ICON_SRC="$ROOT/assets/icon.svg"
DESKTOP_FILE="$PKG_ROOT/usr/share/applications/${APP_NAME}.desktop"

rm -rf "$PKG_ROOT"
mkdir -p \
  "$PKG_ROOT/DEBIAN" \
  "$PKG_ROOT/usr/local/bin" \
  "$PKG_ROOT/usr/share/applications" \
  "$PKG_ROOT/usr/share/icons/hicolor/scalable/apps"

install -m 0755 "$BIN_SRC" "$PKG_ROOT/usr/local/bin/$APP_NAME"
install -m 0644 "$ICON_SRC" "$PKG_ROOT/usr/share/icons/hicolor/scalable/apps/${APP_NAME}.svg"

cat > "$DESKTOP_FILE" <<EOF
[Desktop Entry]
Name=Postman Lite
Comment=Lightweight HTTP API desktop client (local web UI)
Exec=/usr/local/bin/postman-lite
Icon=postman-lite
Terminal=false
Type=Application
Categories=Development;Network;
EOF

cat > "$PKG_ROOT/DEBIAN/control" <<EOF
Package: postman-lite
Version: $VERSION
Section: utils
Priority: optional
Architecture: amd64
Maintainer: clawd <noreply@example.com>
Depends: xdg-utils
Description: Lightweight Postman-style desktop HTTP client written in Go with embedded web UI.
EOF

dpkg-deb --build "$PKG_ROOT" "$OUT_DEB"
echo "Built: $OUT_DEB"
