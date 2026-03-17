#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP_NAME="postman-lite"
VERSION="${VERSION:-0.3.0}"
GO_BIN="${GO_BIN:-go}"
FYNE_BIN="${FYNE_BIN:-fyne}"
APP_ID="${APP_ID:-io.github.nlnhvfpt378.postmanlite}"
OUT_DIR="$ROOT/deliverables"
BUILD_DIR="$ROOT/build/dist"
ICON_PATH="$ROOT/assets/icon.png"

mkdir -p "$OUT_DIR" "$BUILD_DIR"
rm -rf \
  "$BUILD_DIR/linux_amd64" \
  "$BUILD_DIR/windows_amd64" \
  "$BUILD_DIR/darwin_amd64" \
  "$BUILD_DIR/darwin_arm64"
mkdir -p \
  "$BUILD_DIR/linux_amd64" \
  "$BUILD_DIR/windows_amd64" \
  "$BUILD_DIR/darwin_amd64" \
  "$BUILD_DIR/darwin_arm64"

LINUX_BIN="$BUILD_DIR/linux_amd64/$APP_NAME"
WINDOWS_BIN="$BUILD_DIR/windows_amd64/$APP_NAME.exe"
DARWIN_AMD64_APP="$BUILD_DIR/darwin_amd64/$APP_NAME.app"
DARWIN_ARM64_APP="$BUILD_DIR/darwin_arm64/$APP_NAME.app"
LINUX_ARCHIVE="$OUT_DIR/${APP_NAME}_${VERSION}_linux_amd64.tar.gz"
WINDOWS_ARCHIVE="$OUT_DIR/${APP_NAME}_${VERSION}_windows_amd64.zip"
DARWIN_AMD64_ARCHIVE="$OUT_DIR/${APP_NAME}_${VERSION}_darwin_amd64.tar.gz"
DARWIN_ARM64_ARCHIVE="$OUT_DIR/${APP_NAME}_${VERSION}_darwin_arm64.tar.gz"

CGO_ENABLED=1 GOOS=linux GOARCH=amd64 "$GO_BIN" build -o "$LINUX_BIN" ./cmd/$APP_NAME
CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 "$GO_BIN" build -ldflags='-H windowsgui' -o "$WINDOWS_BIN" ./cmd/$APP_NAME

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

if command -v "$FYNE_BIN" >/dev/null 2>&1; then
  for arch in amd64 arm64; do
    APP_DIR="$BUILD_DIR/darwin_${arch}"
    (
      cd "$ROOT"
      CGO_ENABLED=1 GOOS=darwin GOARCH="$arch" "$FYNE_BIN" package \
        --os darwin \
        --release \
        --sourceDir ./cmd/$APP_NAME \
        --name "Postman Lite" \
        --appID "$APP_ID" \
        --appVersion "$VERSION" \
        --icon "$ICON_PATH"
    )
    mv "$ROOT/Postman Lite.app" "$APP_DIR/$APP_NAME.app"
    cp "$ROOT/README.md" "$APP_DIR/README.md"
    tar -C "$APP_DIR" -czf "$OUT_DIR/${APP_NAME}_${VERSION}_darwin_${arch}.tar.gz" "$APP_NAME.app" README.md
  done
else
  echo "warning: fyne CLI not found, macOS app packaging skipped" >&2
fi

echo "Built: $LINUX_ARCHIVE"
echo "Built: $WINDOWS_ARCHIVE"
if [[ -f "$DARWIN_AMD64_ARCHIVE" ]]; then echo "Built: $DARWIN_AMD64_ARCHIVE"; fi
if [[ -f "$DARWIN_ARM64_ARCHIVE" ]]; then echo "Built: $DARWIN_ARM64_ARCHIVE"; fi
