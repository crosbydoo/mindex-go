#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if [ ! -f .env ]; then
  echo "ERROR: .env not found. Run: cp .env.example .env && nano .env"
  exit 1
fi

echo "==> Pulling latest image from Docker Hub..."
docker compose pull

echo "==> Starting containers..."
docker compose up -d --remove-orphans

echo "==> Cleaning unused images..."
docker image prune -f

echo "==> Status:"
docker compose ps

echo ""
echo "Done. API should be running on port $(grep '^PORT=' .env | cut -d= -f2 || echo 8080)"
