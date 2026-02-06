#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")"

dc() {
  if command -v docker-compose >/dev/null 2>&1; then
    docker-compose "$@"
  else
    docker compose "$@"
  fi
}

BIN_FILE="${BIN_FILE:-wyapi-golang_linux_amd64}"
IMAGE_NAME="${IMAGE_NAME:-wyapi-golang:latest}"

if [ ! -f "dist/${BIN_FILE}" ]; then
  echo "dist/${BIN_FILE} not found. Run ./build.sh first."
  exit 1
fi

echo "[1/4] docker compose down"
dc down --remove-orphans

echo "[2/4] docker image rm ${IMAGE_NAME}"
docker image rm "${IMAGE_NAME}" 2>/dev/null || true

echo "[3/4] docker build (disable BuildKit; use daemon registry mirrors)"
DOCKER_BUILDKIT=0 docker build --build-arg "BIN_FILE=${BIN_FILE}" -t "${IMAGE_NAME}" .

echo "[4/4] docker compose up -d"
dc up -d

echo "Done."
