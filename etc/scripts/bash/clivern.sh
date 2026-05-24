#!/usr/bin/env bash
set -euo pipefail

HOST="clivern.com"
DURATION_SECONDS=60

for ((i = 1; i <= DURATION_SECONDS; i++)); do
  echo "tick ${i}/${DURATION_SECONDS} host=${HOST}"
  sleep 1
done

echo "Done."
