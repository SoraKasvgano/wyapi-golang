#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

if [[ "${SKIP_FRONTEND:-}" != "1" ]]; then
  if command -v npm >/dev/null 2>&1; then
    if [[ -d "frontend" ]]; then
      echo "Building frontend..."
      pushd "frontend" >/dev/null
      if [[ ! -d "node_modules" ]]; then
        npm install
      fi
      npm run build
      popd >/dev/null
    fi
  else
    echo "npm not found, skip frontend build."
  fi
fi

mkdir -p dist

LDFLAGS="-s -w"

echo "Building linux/amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="$LDFLAGS" -o dist/wyapi-golang_linux_amd64 ./cmd/server

echo "Building linux/arm64 (armv8)..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="$LDFLAGS" -o dist/wyapi-golang_linux_arm64 ./cmd/server

echo "Building linux/arm (armv7)..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -ldflags="$LDFLAGS" -o dist/wyapi-golang_linux_armv7 ./cmd/server

echo "Building windows/amd64..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="$LDFLAGS" -o dist/wyapi-golang_windows_amd64.exe ./cmd/server

echo "Done. Outputs in ./dist/"
