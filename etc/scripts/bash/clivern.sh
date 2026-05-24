#!/usr/bin/env bash
set -euo pipefail

HOST="clivern.com"
DURATION_SECONDS=60
ping -c "${DURATION_SECONDS}" "${HOST}"
echo "Done."
