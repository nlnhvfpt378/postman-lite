#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP_NAME="postman-lite"
VERSION="${VERSION:-0.1.0}"
GO_BIN="${GO_BIN:-/home/node/clawd/.local/go/bin/go}"
OUT_DIR="$ROOT/deliverables"
BUILD_DIR="$ROOT/build/dist"

mkdir -p "$OUT_DIR" "$BUILD_DIR"
rm -rf "$BUILD_DIR/linux_amd64" "$BUILD_DIR/windows_amd64"
mkdir -p "$BUILD_DIR/linux_amd64" "$BUILD_DIR/windows_amd64"

LINUX_BIN="$BUILD_DIR/linux_amd64/$APP_NAME"
WINDOWS_BIN="$BUILD_DIR/windows_amd64/$APP_NAME.exe"
LINUX_ARCHIVE="$OUT_DIR/${APP_NAME}_${VERSION}_linux_amd64.tar.gz"
WINDOWS_ARCHIVE="$OUT_DIR/${APP_NAME}_${VERSION}_windows_amd64.zip"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 "$GO_BIN" build -o "$LINUX_BIN" ./cmd/$APP_NAME
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 "$GO_BIN" build -o "$WINDOWS_BIN" ./cmd/$APP_NAME

cp "$ROOT/README.md" "$BUILD_DIR/linux_amd64/README.md"
cp "$ROOT/README.md" "$BUILD_DIR/windows_amd64/README.md"

cp "$LINUX_BIN" "$OUT_DIR/$APP_NAME"
cp "$WINDOWS_BIN" "$OUT_DIR/$APP_NAME.exe"

tar -C "$BUILD_DIR/linux_amd64" -czf "$LINUX_ARCHIVE" "$APP_NAME" README.md

if command -v zip >/dev/null 2>&1; then
  (
    cd "$BUILD_DIR/windows_amd64"
    zip -q -r "$WINDOWS_ARCHIVE" "$APP_NAME.exe" README.md
  )
else
  python3 - <<PY
import pathlib
import zipfile
root = pathlib.Path(r"$BUILD_DIR/windows_amd64")
out = pathlib.Path(r"$WINDOWS_ARCHIVE")
with zipfile.ZipFile(out, "w", compression=zipfile.ZIP_DEFLATED) as zf:
    for name in ["postman-lite.exe", "README.md"]:
        zf.write(root / name, arcname=name)
PY
fi

echo "Built: $LINUX_ARCHIVE"
echo "Built: $WINDOWS_ARCHIVE"
