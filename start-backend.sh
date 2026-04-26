#!/bin/bash
set -e

# Backend startup script for image-play
# Usage: ./start-backend.sh [api|worker|all]

BACKEND_DIR="$(cd "$(dirname "$0")/backend" && pwd)"
CONFIG_PATH="${BACKEND_DIR}/config.yaml"

cd "${BACKEND_DIR}"

# Ensure config file exists
if [ ! -f "${CONFIG_PATH}" ]; then
    echo "[error] config.yaml not found at ${CONFIG_PATH}"
    echo "[hint] copy backend/config.yaml and fill in your values"
    exit 1
fi

CMD="${1:-api}"

case "${CMD}" in
    api)
        echo "[start-backend] starting API server..."
        echo "[start-backend] config: ${CONFIG_PATH}"
        CONFIG_PATH="${CONFIG_PATH}" go run ./cmd/api
        ;;
    worker)
        echo "[start-backend] starting worker..."
        echo "[start-backend] config: ${CONFIG_PATH}"
        CONFIG_PATH="${CONFIG_PATH}" go run ./cmd/worker
        ;;
    all)
        echo "[start-backend] starting API server and worker..."
        echo "[start-backend] config: ${CONFIG_PATH}"
        CONFIG_PATH="${CONFIG_PATH}" go run ./cmd/api &
        API_PID=$!
        CONFIG_PATH="${CONFIG_PATH}" go run ./cmd/worker &
        WORKER_PID=$!
        echo "[start-backend] api pid: ${API_PID}, worker pid: ${WORKER_PID}"
        wait
        ;;
    *)
        echo "Usage: $0 [api|worker|all]"
        echo "  api    - start API server only (default)"
        echo "  worker - start worker only"
        echo "  all    - start both api and worker"
        exit 1
        ;;
esac
