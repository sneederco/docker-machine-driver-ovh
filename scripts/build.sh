#!/usr/bin/env bash
set -euo pipefail

ARCH="${1:-amd64}"
OUT_DIR="${2:-dist}"

export GO111MODULE=off
export CGO_ENABLED=0
export GOOS=linux
export GOARCH="$ARCH"

mkdir -p "$OUT_DIR"
OUT_BIN="$OUT_DIR/docker-machine-driver-ovh-linux-${GOARCH}"

# Build with reproducibility-focused flags.
go build -trimpath -ldflags='-s -w -buildid=' -o "$OUT_BIN" .

if command -v sha256sum >/dev/null 2>&1; then
  sha256sum "$OUT_BIN" > "$OUT_BIN.sha256"
elif command -v shasum >/dev/null 2>&1; then
  shasum -a 256 "$OUT_BIN" > "$OUT_BIN.sha256"
else
  echo "No SHA256 tool available (sha256sum/shasum)" >&2
  exit 1
fi

echo "Built: $OUT_BIN"
echo "Checksum: $OUT_BIN.sha256"
