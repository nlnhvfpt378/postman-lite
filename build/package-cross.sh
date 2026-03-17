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

TARGETS="${TARGETS:-linux windows darwin_amd64 darwin_arm64}"

has_target() {
  case " $TARGETS " in
    *" $1 "*) return 0 ;;
    *) return 1 ;;
  esac
}

mkdir -p "$OUT_DIR" "$BUILD_DIR"
rm -rf \
  "$BUILD_DIR/linux_amd64" \
  "$BUILD_DIR/windows_amd64" \
  "$BUILD_DIR/darwin_amd64" \
  "$BUILD_DIR/darwin_arm64"

LINUX_ARCHIVE="$OUT_DIR/${APP_NAME}_${VERSION}_linux_amd64.tar.gz"
WINDOWS_ARCHIVE="$OUT_DIR/${APP_NAME}_${VERSION}_windows_amd64.zip"
DARWIN_AMD64_ARCHIVE="$OUT_DIR/${APP_NAME}_${VERSION}_darwin_amd64.tar.gz"
DARWIN_ARM64_ARCHIVE="$OUT_DIR/${APP_NAME}_${VERSION}_darwin_arm64.tar.gz"

if has_target linux; then
  mkdir -p "$BUILD_DIR/linux_amd64"
  LINUX_BIN="$BUILD_DIR/linux_amd64/$APP_NAME"
  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 "$GO_BIN" build -o "$LINUX_BIN" ./cmd/$APP_NAME
  cp "$ROOT/README.md" "$BUILD_DIR/linux_amd64/README.md"
  cp "$LINUX_BIN" "$OUT_DIR/$APP_NAME"
  tar -C "$BUILD_DIR/linux_amd64" -czf "$LINUX_ARCHIVE" "$APP_NAME" README.md
fi

if has_target windows; then
  mkdir -p "$BUILD_DIR/windows_amd64"
  WINDOWS_BIN="$BUILD_DIR/windows_amd64/$APP_NAME.exe"
  CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 "$GO_BIN" build -ldflags='-H windowsgui' -o "$WINDOWS_BIN" ./cmd/$APP_NAME
  cp "$ROOT/README.md" "$BUILD_DIR/windows_amd64/README.md"
  cp "$WINDOWS_BIN" "$OUT_DIR/$APP_NAME.exe"

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
fi

if command -v "$FYNE_BIN" >/dev/null 2>&1; then
  for target in darwin_amd64 darwin_arm64; do
    if ! has_target "$target"; then
      continue
    fi
    arch="${target#darwin_}"
    APP_DIR="$BUILD_DIR/$target"
    mkdir -p "$APP_DIR"
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
  if has_target darwin_amd64 || has_target darwin_arm64; then
    echo "warning: fyne CLI not found, macOS app packaging skipped" >&2
  fi
fi

if [[ -f "$LINUX_ARCHIVE" ]]; then echo "Built: $LINUX_ARCHIVE"; fi
if [[ -f "$WINDOWS_ARCHIVE" ]]; then echo "Built: $WINDOWS_ARCHIVE"; fi
if [[ -f "$DARWIN_AMD64_ARCHIVE" ]]; then echo "Built: $DARWIN_AMD64_ARCHIVE"; fi
if [[ -f "$DARWIN_ARM64_ARCHIVE" ]]; then echo "Built: $DARWIN_ARM64_ARCHIVE"; fi
